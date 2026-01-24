import { api } from './client';
import type { QANote } from '../types';

export interface CreateNoteRequest {
  author: string;
  content: string;
}

// Add a note to a bundle
export async function createNote(bundleId: string, data: CreateNoteRequest): Promise<QANote> {
  return api.post<QANote>(`/repro-bundles/${bundleId}/notes`, data);
}

// Legacy aliases for compatibility
export type Note = QANote;
export type GetNotesResponse = { notes: QANote[] };
export type UpdateNoteRequest = { content: string };

// Get notes for a bundle (notes are included in bundle detail response)
export async function getNotes(bundleId: string): Promise<GetNotesResponse> {
  // Notes are included in the bundle detail, fetch the bundle
  const response = await api.get<{ qa_notes?: QANote[] }>(`/repro-bundles/${bundleId}`);
  return { notes: response.qa_notes || [] };
}

// Note: Backend doesn't support updating or deleting notes
export async function updateNote(
  _bundleId: string, 
  _noteId: string, 
  _data: UpdateNoteRequest
): Promise<QANote> {
  throw new Error('Note update not supported by backend');
}

export async function deleteNote(_bundleId: string, _noteId: string): Promise<void> {
  throw new Error('Note deletion not supported by backend');
}
