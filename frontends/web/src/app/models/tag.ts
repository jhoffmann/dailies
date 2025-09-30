import { Task } from './task';

export interface Tag {
  id: string;
  name: string;
  color: string;
  tasks?: Task[];
  created_at?: string;
  updated_at?: string;
  // Dynamic edit properties
  editing?: boolean;
  editName?: string;
  editColor?: string;
}
