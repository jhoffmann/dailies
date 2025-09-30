import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormsModule, FormBuilder, FormGroup } from '@angular/forms';
import { Subject, takeUntil } from 'rxjs';
import cronstrue from 'cronstrue';

import { ApiService } from '../../services/api.service';
import { WebSocketService, WebSocketEvent } from '../../services/websocket.service';
import { Frequency } from '../../models/frequency';
import { TimezoneInfo } from '../../models/timezone-info';

@Component({
  selector: 'app-frequencies',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, FormsModule],
  templateUrl: './frequencies.component.html',
  styleUrl: './frequencies.component.css',
})
export class FrequenciesComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();

  frequencies: Frequency[] = [];
  addFrequencyForm: FormGroup;
  timezoneInfo: TimezoneInfo | null = null;

  constructor(
    private apiService: ApiService,
    private wsService: WebSocketService,
    private fb: FormBuilder,
  ) {
    this.addFrequencyForm = this.fb.group({
      name: [''],
      period: [''],
    });
  }

  ngOnInit() {
    this.loadFrequencies();
    this.loadTimezone();
    this.setupWebSocketSubscriptions();
  }

  private setupWebSocketSubscriptions() {
    // Listen for WebSocket events to refresh frequency data automatically
    this.wsService.events$.pipe(takeUntil(this.destroy$)).subscribe((event: WebSocketEvent) => {
      switch (event.type) {
        case 'frequency_create':
        case 'frequency_update':
        case 'frequency_delete':
          this.loadFrequencies();
          break;
      }
    });
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
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

  loadTimezone() {
    this.apiService
      .getTimezone()
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (timezone) => {
          this.timezoneInfo = timezone;
        },
        error: (error) => {
          console.error('Error loading timezone:', error);
        },
      });
  }

  getHumanReadableCron(cronExpression: string): string {
    try {
      return cronstrue.toString(cronExpression);
    } catch (error) {
      return 'Invalid cron expression';
    }
  }

  addFrequency() {
    const formValue = this.addFrequencyForm.value;
    if (!formValue.name || !formValue.period) return;

    const frequencyData: any = {
      name: formValue.name,
      period: formValue.period,
    };

    this.apiService
      .createFrequency(frequencyData)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.loadFrequencies();
          this.addFrequencyForm.reset();
        },
        error: (error) => {
          console.error('Error adding frequency:', error);
          alert('Error adding frequency: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }

  editFrequency(frequency: Frequency) {
    frequency.editing = true;
    frequency.editName = frequency.name;
    frequency.editPeriod = frequency.period;
  }

  saveFrequency(frequency: Frequency) {
    const updateData: any = {
      name: frequency.editName,
      period: frequency.editPeriod,
    };

    this.apiService
      .updateFrequency(frequency.id, updateData)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (updatedFrequency) => {
          frequency.name = frequency.editName!;
          frequency.period = frequency.editPeriod!;
          frequency.reset = frequency.editPeriod!;
          frequency.editing = false;
          this.apiService.notifyFrequenciesChanged();
        },
        error: (error) => {
          console.error('Error updating frequency:', error);
          alert('Error updating frequency: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }

  cancelEdit(frequency: Frequency) {
    frequency.editing = false;
    frequency.editName = undefined;
    frequency.editPeriod = undefined;
  }

  deleteFrequency(frequencyId: string) {
    if (!confirm('Are you sure you want to delete this frequency?')) return;

    this.apiService
      .deleteFrequency(frequencyId)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.loadFrequencies();
          this.apiService.notifyFrequenciesChanged();
        },
        error: (error) => {
          console.error('Error deleting frequency:', error);
          alert('Error deleting frequency: ' + (error.error?.error || 'Unknown error'));
        },
      });
  }
}
