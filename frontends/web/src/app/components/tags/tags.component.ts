import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormsModule, FormBuilder, FormGroup } from '@angular/forms';
import { Subject, takeUntil } from 'rxjs';

import { ApiService } from '../../services/api.service';
import { WebSocketService, WebSocketEvent } from '../../services/websocket.service';
import { Tag } from '../../models/tag';

@Component({
  selector: 'app-tags',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, FormsModule],
  templateUrl: './tags.component.html',
  styleUrl: './tags.component.css',
})
export class TagsComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();

  tags: Tag[] = [];
  addTagForm: FormGroup;

  constructor(
    private apiService: ApiService,
    private wsService: WebSocketService,
    private fb: FormBuilder,
  ) {
    this.addTagForm = this.fb.group({
      name: [''],
      color: ['#3498db'],
    });
  }

  ngOnInit() {
    this.loadTags();
    this.setupWebSocketSubscriptions();
  }

  private setupWebSocketSubscriptions() {
    // Listen for WebSocket events to refresh tag data automatically
    this.wsService.events$.pipe(takeUntil(this.destroy$)).subscribe((event: WebSocketEvent) => {
      switch (event.type) {
        case 'tag_create':
        case 'tag_update':
        case 'tag_delete':
          this.loadTags();
          break;
      }
    });
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  loadTags() {
    this.apiService
      .getTags()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (tags) => {
          this.tags = tags || [];
        },
        error: (error) => {
          console.error('Error loading tags:', error);
          this.tags = [];
        },
      });
  }

  addTag() {
    const formValue = this.addTagForm.value;
    if (!formValue.name) return;

    const tagData: Partial<Tag> = {
      name: formValue.name,
      color: formValue.color,
    };

    this.apiService
      .createTag(tagData)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.loadTags();
          this.addTagForm.reset();
          this.addTagForm.patchValue({ color: '#3498db' });
        },
        error: (error) => {
          console.error('Error adding tag:', error);
          alert('Error adding tag: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }

  editTag(tag: Tag) {
    tag.editing = true;
    tag.editName = tag.name;
    tag.editColor = tag.color;
  }

  saveTag(tag: Tag) {
    const updateData: Partial<Tag> = {
      name: tag.editName,
      color: tag.editColor,
    };

    this.apiService
      .updateTag(tag.id, updateData)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          tag.name = tag.editName!;
          tag.color = tag.editColor!;
          tag.editing = false;
          this.apiService.notifyTagsChanged();
        },
        error: (error) => {
          console.error('Error updating tag:', error);
          alert('Error updating tag: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }

  cancelEdit(tag: Tag) {
    tag.editing = false;
    tag.editName = undefined;
    tag.editColor = undefined;
  }

  deleteTag(tagId: string) {
    if (!confirm('Are you sure you want to delete this tag?')) return;

    this.apiService
      .deleteTag(tagId)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.loadTags();
          this.apiService.notifyTagsChanged();
        },
        error: (error) => {
          console.error('Error deleting tag:', error);
          alert('Error deleting tag: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }
}
