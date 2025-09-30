import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';
import { TaskFilters } from '../models/task-filters';

@Injectable({
  providedIn: 'root',
})
export class FilterStorageService {
  private readonly STORAGE_KEY = 'dailies-task-filters';
  private filtersSubject = new BehaviorSubject<TaskFilters>(this.loadFiltersFromStorage());

  public filters$ = this.filtersSubject.asObservable();

  constructor() {}

  getFilters(): TaskFilters {
    return this.filtersSubject.value;
  }

  updateFilters(filters: TaskFilters): void {
    this.filtersSubject.next(filters);
    this.saveFiltersToStorage(filters);
  }

  private loadFiltersFromStorage(): TaskFilters {
    try {
      const savedFilters = localStorage.getItem(this.STORAGE_KEY);
      if (savedFilters) {
        return JSON.parse(savedFilters);
      }
    } catch (e) {
      console.error('Error loading filters from localStorage:', e);
    }
    return {};
  }

  private saveFiltersToStorage(filters: TaskFilters): void {
    try {
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(filters));
    } catch (e) {
      console.error('Error saving filters to localStorage:', e);
    }
  }
}
