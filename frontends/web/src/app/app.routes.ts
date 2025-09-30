import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: '/tasks', pathMatch: 'full' },
  {
    path: 'tasks',
    loadComponent: () => import('./components/tasks/tasks.component').then((c) => c.TasksComponent),
  },
  {
    path: 'tags',
    loadComponent: () => import('./components/tags/tags.component').then((c) => c.TagsComponent),
  },
  {
    path: 'frequencies',
    loadComponent: () =>
      import('./components/frequencies/frequencies.component').then((c) => c.FrequenciesComponent),
  },
  { path: '**', redirectTo: '/tasks' },
];
