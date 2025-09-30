import { Tag } from './tag';
import { Frequency } from './frequency';

export interface Task {
  id: string;
  name: string;
  description?: string;
  completed: boolean;
  priority?: number;
  frequency_id?: string;
  frequency?: Frequency;
  tags: Tag[];
  created_at?: string;
  updated_at?: string;
  // Dynamic edit properties
  editing?: boolean;
  editName?: string;
  editDescription?: string;
  editPriority?: string;
  editFrequencyId?: string;
  editSelectedTags?: { [key: string]: boolean };
}
