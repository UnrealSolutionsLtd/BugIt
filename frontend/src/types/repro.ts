// Platform types matching backend validation
export type Platform = 
  | 'Win64' 
  | 'Linux' 
  | 'Mac' 
  | 'Android' 
  | 'iOS' 
  | 'Other';

// Artifact within a bundle
export interface Artifact {
  artifact_id: string;
  filename: string;
  type: 'video' | 'log' | 'screenshot' | 'crash_dump' | 'thumbnail' | 'other';
  mime_type?: string;
  size_bytes: number;
  checksum?: string;
  created_at: string;
}

// QA Note on a bundle
export interface QANote {
  note_id: string;
  author: string;
  content: string;
  created_at: string;
}

// Repro bundle from backend
export interface ReproBundle {
  bundle_id: string;
  content_hash: string;
  schema_version: string;
  build_id: string;
  map_name?: string;
  platform: Platform;
  rvr_version?: string;
  bundle_timestamp: string;
  metadata?: Record<string, unknown>;
  size_bytes: number;
  artifact_count: number;
  created_at: string;
  // Populated on detail queries
  artifacts?: Artifact[];
  tags?: string[];
  qa_notes?: QANote[];
}

// List response from GET /api/repro-bundles
export interface GetBundlesResponse {
  bundles: ReproBundle[];
  total: number;
  limit: number;
  offset: number;
}

// Query params for listing bundles
export interface BundleFilters {
  build_id?: string;
  map_name?: string;
  platform?: Platform;
  since?: string;
  limit?: number;
  offset?: number;
}

// Legacy aliases for compatibility with existing components
export type ReproSummary = ReproBundle;
export type ReproDetail = ReproBundle;
export type GetReprosResponse = {
  repros: ReproBundle[];
  total: number;
  page: number;
  pageSize: number;
};

export interface GetFiltersResponse {
  builds: string[];
  platforms: Platform[];
  maps: string[];
  tags: string[];
}

export interface ReproFilters {
  build?: string;
  platform?: Platform;
  map?: string;
  dateFrom?: string;
  dateTo?: string;
  search?: string;
  page?: number;
  limit?: number;
}
