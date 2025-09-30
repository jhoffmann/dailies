import { Task } from './task';

export interface Frequency {
  id: string;
  name: string;
  period: string;
  reset: string;
  tasks?: Task[];
  created_at?: string;
  updated_at?: string;
  // Dynamic edit properties
  editing?: boolean;
  editName?: string;
  editPeriod?: string;
}
