import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';
import { Task } from '../models/task';
import { Tag } from '../models/tag';
import { Frequency } from '../models/frequency';
import { Timer } from '../models/timer';
import { TimezoneInfo } from '../models/timezone-info';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class ApiService {
  private readonly API_BASE = `http://${environment.apiHost}:${environment.apiPort}/api`;

  // Event subjects for cross-component communication
  private tagsChangedSubject = new Subject<void>();
  private frequenciesChangedSubject = new Subject<void>();

  public tagsChanged$ = this.tagsChangedSubject.asObservable();
  public frequenciesChanged$ = this.frequenciesChangedSubject.asObservable();

  constructor(private http: HttpClient) {}

  // Generic HTTP methods
  private get<T>(endpoint: string, params?: any): Observable<T> {
    let httpParams = new HttpParams();
    if (params) {
      Object.keys(params).forEach((key) => {
        if (params[key] !== null && params[key] !== undefined && params[key] !== '') {
          httpParams = httpParams.set(key, params[key].toString());
        }
      });
    }
    return this.http.get<T>(`${this.API_BASE}${endpoint}`, { params: httpParams });
  }

  private post<T>(endpoint: string, data: any): Observable<T> {
    return this.http.post<T>(`${this.API_BASE}${endpoint}`, data);
  }

  private put<T>(endpoint: string, data: any): Observable<T> {
    return this.http.put<T>(`${this.API_BASE}${endpoint}`, data);
  }

  private delete<T>(endpoint: string): Observable<T> {
    return this.http.delete<T>(`${this.API_BASE}${endpoint}`);
  }

  // Task API methods
  getTasks(params?: any): Observable<Task[]> {
    return this.get<Task[]>('/tasks', params);
  }

  createTask(task: Partial<Task>): Observable<Task> {
    return this.post<Task>('/tasks', task);
  }

  updateTask(id: string, task: Partial<Task>): Observable<Task> {
    return this.put<Task>(`/tasks/${id}`, task);
  }

  deleteTask(id: string): Observable<void> {
    return this.delete<void>(`/tasks/${id}`);
  }

  // Tag API methods
  getTags(): Observable<Tag[]> {
    return this.get<Tag[]>('/tags');
  }

  createTag(tag: Partial<Tag>): Observable<Tag> {
    return this.post<Tag>('/tags', tag);
  }

  updateTag(id: string, tag: Partial<Tag>): Observable<Tag> {
    return this.put<Tag>(`/tags/${id}`, tag);
  }

  deleteTag(id: string): Observable<void> {
    return this.delete<void>(`/tags/${id}`);
  }

  // Frequency API methods
  getFrequencies(): Observable<Frequency[]> {
    return this.get<Frequency[]>('/frequencies');
  }

  createFrequency(frequency: Partial<Frequency>): Observable<Frequency> {
    return this.post<Frequency>('/frequencies', frequency);
  }

  updateFrequency(id: string, frequency: Partial<Frequency>): Observable<Frequency> {
    return this.put<Frequency>(`/frequencies/${id}`, frequency);
  }

  deleteFrequency(id: string): Observable<void> {
    return this.delete<void>(`/frequencies/${id}`);
  }

  // Timer methods
  getTimers(): Observable<Timer[]> {
    return this.get<Timer[]>('/frequencies/timers');
  }

  // Timezone methods
  getTimezone(): Observable<TimezoneInfo> {
    return this.get<TimezoneInfo>('/timezone');
  }

  // Event notification methods
  notifyTagsChanged(): void {
    this.tagsChangedSubject.next();
  }

  notifyFrequenciesChanged(): void {
    this.frequenciesChangedSubject.next();
  }
}
