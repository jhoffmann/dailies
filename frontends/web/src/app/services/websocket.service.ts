import { Injectable } from '@angular/core';
import { BehaviorSubject, Subject } from 'rxjs';
import { Task } from '../models/task';
import { Tag } from '../models/tag';
import { Frequency } from '../models/frequency';
import { environment } from '../../environments/environment';

export interface WebSocketEvent {
  type:
    | 'task_reset'
    | 'task_update'
    | 'task_create'
    | 'task_delete'
    | 'tag_update'
    | 'tag_create'
    | 'tag_delete'
    | 'frequency_update'
    | 'frequency_create'
    | 'frequency_delete';
  data: Task | Tag | Frequency;
}

@Injectable({
  providedIn: 'root',
})
export class WebSocketService {
  private socket: WebSocket | null = null;
  private reconnectInterval = 5000;
  private maxReconnectAttempts = 10;
  private reconnectAttempts = 0;

  private eventsSubject = new Subject<WebSocketEvent>();
  private connectionSubject = new BehaviorSubject<boolean>(false);

  public events$ = this.eventsSubject.asObservable();
  public connected$ = this.connectionSubject.asObservable();

  constructor() {
    this.connect();
  }

  private connect(): void {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = environment.apiPort.toString();
    const wsUrl = `${protocol}//${host}:${port}/ws`;

    try {
      this.socket = new WebSocket(wsUrl);

      this.socket.onopen = () => {
        console.log('WebSocket connected');
        this.connectionSubject.next(true);
        this.reconnectAttempts = 0;
      };

      this.socket.onmessage = (event) => {
        try {
          const data: WebSocketEvent = JSON.parse(event.data);
          this.eventsSubject.next(data);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      this.socket.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        this.connectionSubject.next(false);
        this.socket = null;
        this.scheduleReconnect();
      };

      this.socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.connectionSubject.next(false);
      };
    } catch (error) {
      console.error('Error creating WebSocket connection:', error);
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(
        `Attempting to reconnect in ${this.reconnectInterval}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`,
      );

      setTimeout(() => {
        this.connect();
      }, this.reconnectInterval);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  public disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.connectionSubject.next(false);
  }

  public isConnected(): boolean {
    return this.socket !== null && this.socket.readyState === WebSocket.OPEN;
  }
}
