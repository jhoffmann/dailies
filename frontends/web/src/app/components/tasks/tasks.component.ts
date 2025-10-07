import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import {
  ReactiveFormsModule,
  FormsModule,
  FormBuilder,
  FormGroup,
  FormArray,
} from '@angular/forms';
import { Subject, takeUntil } from 'rxjs';

import { ApiService } from '../../services/api.service';
import { FilterStorageService } from '../../services/filter-storage.service';
import { WebSocketService, WebSocketEvent } from '../../services/websocket.service';
import { Task } from '../../models/task';
import { Tag } from '../../models/tag';
import { Frequency } from '../../models/frequency';
import { Timer } from '../../models/timer';
import { TaskFilters } from '../../models/task-filters';

@Component({
  selector: 'app-tasks',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, FormsModule],
  templateUrl: './tasks.component.html',
  styleUrl: './tasks.component.css',
})
export class TasksComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  private timerRefreshInterval?: number;

  tasks: Task[] = [];
  filteredTasks: Task[] = [];
  frequencies: Frequency[] = [];
  tags: Tag[] = [];
  timers: Timer[] = [];

  addTaskForm: FormGroup;
  filters: TaskFilters = {};

  constructor(
    private apiService: ApiService,
    private filterStorage: FilterStorageService,
    private wsService: WebSocketService,
    private fb: FormBuilder,
    private sanitizer: DomSanitizer,
  ) {
    this.addTaskForm = this.fb.group({
      name: [''],
      description: [''],
      priority: [''],
      frequency_id: [''],
      selectedTags: this.fb.array([]),
    });
  }

  ngOnInit() {
    this.loadStaticData();
    this.initializeFilters();
    this.setupEventSubscriptions();
    this.startTimerRefresh();
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();

    if (this.timerRefreshInterval) {
      clearInterval(this.timerRefreshInterval);
    }
  }

  private initializeFilters() {
    this.filterStorage.filters$.pipe(takeUntil(this.destroy$)).subscribe((filters) => {
      this.filters = filters;
      this.loadTasks();
    });
  }

  private loadData() {
    this.loadTasks();
    this.loadFrequencies();
    this.loadTags();
    this.loadTimers();
  }

  private loadStaticData() {
    this.loadFrequencies();
    this.loadTags();
    this.loadTimers();
  }

  private setupEventSubscriptions() {
    this.apiService.tagsChanged$.pipe(takeUntil(this.destroy$)).subscribe(() => this.loadTags());

    this.apiService.frequenciesChanged$
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => this.loadFrequencies());

    // Listen for WebSocket events to refresh data automatically
    this.wsService.events$.pipe(takeUntil(this.destroy$)).subscribe((event: WebSocketEvent) => {
      switch (event.type) {
        case 'task_reset':
        case 'task_create':
        case 'task_update':
        case 'task_delete':
          this.loadTasks();
          break;
        case 'tag_create':
        case 'tag_update':
        case 'tag_delete':
          this.loadTags();
          break;
        case 'frequency_create':
        case 'frequency_update':
        case 'frequency_delete':
          this.loadFrequencies();
          break;
      }
    });
  }

  private startTimerRefresh() {
    this.timerRefreshInterval = window.setInterval(() => {
      this.loadTimers();
    }, 60000);
  }

  loadTasks() {
    const params: any = {};
    if (this.filters.name) params.name = this.filters.name;
    if (this.filters.completed !== undefined && this.filters.completed !== '') {
      params.completed = this.filters.completed;
    }
    if (this.filters.tag) params.tag_ids = this.filters.tag;
    if (this.filters.sort) params.sort = this.filters.sort;

    this.apiService
      .getTasks(params)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (tasks) => {
          this.tasks = tasks || [];
          this.updateFilteredTasks();
        },
        error: (error) => {
          console.error('Error loading tasks:', error);
          this.tasks = [];
        },
      });
  }

  loadFrequencies() {
    this.apiService
      .getFrequencies()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (frequencies) => {
          this.frequencies = frequencies || [];
        },
        error: (error) => {
          console.error('Error loading frequencies:', error);
          this.frequencies = [];
        },
      });
  }

  loadTags() {
    this.apiService
      .getTags()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (tags) => {
          this.tags = tags || [];
          this.updateTagCheckboxes();
        },
        error: (error) => {
          console.error('Error loading tags:', error);
          this.tags = [];
        },
      });
  }

  loadTimers() {
    this.apiService
      .getTimers()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (timers) => {
          this.timers = timers || [];
        },
        error: (error) => {
          console.error('Error loading timers:', error);
          this.timers = [];
        },
      });
  }

  private updateTagCheckboxes() {
    const tagArray = this.addTaskForm.get('selectedTags') as FormArray;

    // If the array length doesn't match tags length, rebuild it
    if (tagArray.length !== this.tags.length) {
      tagArray.clear();
      this.tags.forEach(() => {
        tagArray.push(this.fb.control(false));
      });
    } else {
      // Otherwise just reset all values to false
      tagArray.controls.forEach((control) => {
        control.setValue(false);
      });
    }
  }

  private updateFilteredTasks() {
    this.filteredTasks = this.tasks.filter((task) => {
      if (
        this.filters.name &&
        task.name.toLowerCase().indexOf(this.filters.name.toLowerCase()) === -1
      ) {
        return false;
      }

      if (this.filters.completed !== undefined && this.filters.completed !== '') {
        if (this.filters.completed === 'true' && !task.completed) return false;
        if (this.filters.completed === 'false' && task.completed) return false;
      }

      if (this.filters.tag) {
        const hasTag = task.tags && task.tags.some((tag) => tag.id === this.filters.tag);
        if (!hasTag) return false;
      }

      return true;
    });
  }

  updateFilters(key: keyof TaskFilters, value: any) {
    this.filters = { ...this.filters, [key]: value };
    this.filterStorage.updateFilters(this.filters);
  }

  addTask() {
    const formValue = this.addTaskForm.value;
    if (!formValue.name) return;

    const selectedTagIds: string[] = [];
    const tagArray = this.addTaskForm.get('selectedTags') as FormArray;

    // Get the selected tag IDs from the FormArray
    tagArray.controls.forEach((control, index) => {
      if (control.value === true) {
        selectedTagIds.push(this.tags[index].id);
      }
    });

    const taskData: any = {
      name: formValue.name,
      description: formValue.description || '',
      priority: formValue.priority ? parseInt(formValue.priority) : undefined,
      frequency_id: formValue.frequency_id || undefined,
      tag_ids: selectedTagIds,
    };

    this.apiService
      .createTask(taskData)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.loadTasks();
          // Reset the entire form including the FormArray
          this.addTaskForm.reset();
          // Reinitialize tag checkboxes to false
          const tagArray = this.addTaskForm.get('selectedTags') as FormArray;
          tagArray.controls.forEach((control) => {
            control.setValue(false);
          });
        },
        error: (error) => {
          console.error('Error adding task:', error);
          alert('Error adding task: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }

  toggleCompleted(task: Task) {
    this.apiService
      .updateTask(task.id, { completed: task.completed })
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.loadTasks();
        },
        error: (error) => {
          console.error('Error updating task:', error);
          task.completed = !task.completed; // Revert on error
          alert('Error updating task: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }

  editTask(task: Task) {
    task.editing = true;
    task.editName = task.name;
    task.editDescription = task.description;
    task.editPriority = task.priority ? task.priority.toString() : '';
    task.editFrequencyId = task.frequency ? task.frequency.id : '';
    task.editSelectedTags = {};

    if (task.tags && task.tags.length > 0) {
      task.tags.forEach((tag) => {
        task.editSelectedTags![tag.id] = true;
      });
    }
  }

  saveTask(task: Task) {
    const selectedTagIds: string[] = [];

    if (task.editSelectedTags) {
      Object.keys(task.editSelectedTags).forEach((tagId) => {
        if (task.editSelectedTags![tagId]) {
          selectedTagIds.push(tagId);
        }
      });
    }

    const updateData: any = {
      name: task.editName,
      description: task.editDescription,
      priority: task.editPriority ? parseInt(task.editPriority) : undefined,
      frequency_id: task.editFrequencyId || '',
      tag_ids: selectedTagIds,
    };

    this.apiService
      .updateTask(task.id, updateData)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.loadTasks();
          task.editing = false;
        },
        error: (error) => {
          console.error('Error updating task:', error);
          alert('Error updating task: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }

  cancelEdit(task: Task) {
    task.editing = false;
    task.editName = undefined;
    task.editDescription = undefined;
    task.editPriority = undefined;
    task.editFrequencyId = undefined;
    task.editSelectedTags = undefined;
  }

  deleteTask(taskId: string) {
    if (!confirm('Are you sure you want to delete this task?')) return;

    this.apiService
      .deleteTask(taskId)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.loadTasks();
        },
        error: (error) => {
          console.error('Error deleting task:', error);
          alert('Error deleting task: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }

  get selectedTagsArray(): FormArray {
    return this.addTaskForm.get('selectedTags') as FormArray;
  }

  convertLinksToSafeHtml(text: string): SafeHtml {
    if (!text) return '';

    // Convert URLs to links with target="_blank"
    const urlRegex = /(https?:\/\/[^\s]+)/g;
    const htmlWithLinks = text.replace(
      urlRegex,
      '<a href="$1" target="_blank" rel="noopener noreferrer">$1</a>',
    );

    return this.sanitizer.bypassSecurityTrustHtml(htmlWithLinks);
  }
}
