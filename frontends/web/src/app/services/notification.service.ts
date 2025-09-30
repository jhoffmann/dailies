import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';

export interface Toast {
  id: string;
  message: string;
  type: 'success' | 'info' | 'warning' | 'error';
  visible: boolean;
  timeout?: number;
}

@Injectable({
  providedIn: 'root',
})
export class NotificationService {
  private notificationsSubject = new Subject<Toast>();
  public notifications$ = this.notificationsSubject.asObservable();

  constructor() {}

  public showSuccess(message: string, timeout = 5000): void {
    this.showNotification(message, 'success', timeout);
  }

  public showError(message: string, timeout = 8000): void {
    this.showNotification(message, 'error', timeout);
  }

  public showWarning(message: string, timeout = 6000): void {
    this.showNotification(message, 'warning', timeout);
  }

  public showInfo(message: string, timeout = 5000): void {
    this.showNotification(message, 'info', timeout);
  }

  private showNotification(message: string, type: Toast['type'], timeout: number): void {
    const toast: Toast = {
      id: Date.now().toString() + Math.random().toString(36).substr(2, 9),
      message,
      type,
      visible: true,
      timeout,
    };

    this.notificationsSubject.next(toast);
  }
}
