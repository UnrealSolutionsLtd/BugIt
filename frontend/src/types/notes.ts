// Re-export QANote from repro types for compatibility
export type { QANote } from './repro';

// Legacy Note type for compatibility
export interface Note {
  id: string;
  timestampMs?: number;
  author: string;
  content: string;
  createdAt: string;
  updatedAt?: string;
}

export interface CreateNoteRequest {
  author: string;
  content: string;
}

export interface UpdateNoteRequest {
  content: string;
}

export interface GetNotesResponse {
  notes: Note[];
}
