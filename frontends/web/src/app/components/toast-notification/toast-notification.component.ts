import { Component, OnInit, OnDestroy } from '@angular/core';
import { trigger, state, style, transition, animate } from '@angular/animations';
import { Subscription } from 'rxjs';
import { WebSocketService, WebSocketEvent } from '../../services/websocket.service';
import { NotificationService } from '../../services/notification.service';
import { AudioService } from '../../services/audio.service';
import { CommonModule } from '@angular/common';

export interface Toast {
  id: string;
  message: string;
  type: 'success' | 'info' | 'warning' | 'error';
  visible: boolean;
  timeout?: number;
}

@Component({
  selector: 'app-toast-notification',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="toast-container">
      @for (toast of toasts; track toast.id) {
        <div
          class="toast toast-{{ toast.type }}"
          [@slideInOut]="toast.visible ? 'in' : 'out'"
          (@slideInOut.done)="onAnimationDone($event, toast.id)"
        >
          <div class="toast-content">
            <span class="toast-icon">{{ getIcon(toast.type) }}</span>
            <span class="toast-message">{{ toast.message }}</span>
          </div>
        </div>
      }
    </div>
  `,
  styleUrls: ['./toast-notification.component.css'],
  animations: [
    trigger('slideInOut', [
      state(
        'in',
        style({
          transform: 'translateX(0)',
          opacity: 1,
        }),
      ),
      state(
        'out',
        style({
          transform: 'translateX(-100%)',
          opacity: 0,
        }),
      ),
      transition('out => in', [animate('300ms ease-out')]),
      transition('in => out', [animate('300ms ease-in')]),
    ]),
  ],
})
export class ToastNotificationComponent implements OnInit, OnDestroy {
  toasts: Toast[] = [];
  private subscriptions: Subscription[] = [];

  constructor(
    private wsService: WebSocketService,
    private notificationService: NotificationService,
    private audioService: AudioService,
  ) {}

  ngOnInit(): void {
    // Listen for WebSocket events
    this.subscriptions.push(
      this.wsService.events$.subscribe((event: WebSocketEvent) => {
        this.handleWebSocketEvent(event);
      }),
    );

    // Listen for manual notifications
    this.subscriptions.push(
      this.notificationService.notifications$.subscribe((toast: Toast) => {
        this.showToast(toast);
      }),
    );
  }

  ngOnDestroy(): void {
    this.subscriptions.forEach((sub) => sub.unsubscribe());
  }

  private handleWebSocketEvent(event: WebSocketEvent): void {
    let message = '';
    let type: Toast['type'] = 'info';

    switch (event.type) {
      case 'task_reset':
        message = `Task "${(event.data as any).name}" was reset automatically`;
        type = 'info';
        break;
      case 'task_create':
        message = `Task "${(event.data as any).name}" was created`;
        type = 'success';
        break;
      case 'task_update':
        message = `Task "${(event.data as any).name}" was updated`;
        type = 'info';
        break;
      case 'task_delete':
        message = `Task "${(event.data as any).name}" was deleted`;
        type = 'warning';
        break;
      case 'tag_create':
        message = `Tag "${(event.data as any).name}" was created`;
        type = 'success';
        break;
      case 'tag_update':
        message = `Tag "${(event.data as any).name}" was updated`;
        type = 'info';
        break;
      case 'tag_delete':
        message = `Tag "${(event.data as any).name}" was deleted`;
        type = 'warning';
        break;
      case 'frequency_create':
        message = `Frequency "${(event.data as any).name}" was created`;
        type = 'success';
        break;
      case 'frequency_update':
        message = `Frequency "${(event.data as any).name}" was updated`;
        type = 'info';
        break;
      case 'frequency_delete':
        message = `Frequency "${(event.data as any).name}" was deleted`;
        type = 'warning';
        break;
    }

    if (message) {
      this.showToast({
        id: Date.now().toString(),
        message,
        type,
        visible: true,
        timeout: 10000, // 10 seconds
      });

      // Play notification sound
      this.audioService.playNotification();
    }
  }

  private showToast(toast: Toast): void {
    // Add the toast to the array
    this.toasts.push({ ...toast, visible: true });

    // Auto-hide after timeout
    if (toast.timeout) {
      setTimeout(() => {
        this.hideToast(toast.id);
      }, toast.timeout);
    }
  }

  private hideToast(id: string): void {
    const toast = this.toasts.find((t) => t.id === id);
    if (toast) {
      toast.visible = false;
    }
  }

  onAnimationDone(event: any, toastId: string): void {
    if (event.toState === 'out') {
      // Remove the toast from the array after slide-out animation
      this.toasts = this.toasts.filter((t) => t.id !== toastId);
    }
  }

  getIcon(type: Toast['type']): string {
    switch (type) {
      case 'success':
        return '✓';
      case 'error':
        return '✗';
      case 'warning':
        return '⚠';
      case 'info':
        return 'ℹ';
      default:
        return 'ℹ';
    }
  }
}
