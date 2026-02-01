import { api } from './client';
import type { 
  ReproBundle,
  GetBundlesResponse,
  BundleFilters,
  GetReprosResponse,
  ReproFilters,
  GetFiltersResponse,
} from '../types';

// Get list of bundles from backend
export async function getBundles(filters: BundleFilters = {}): Promise<GetBundlesResponse> {
  return api.get<GetBundlesResponse>('/repro-bundles', {
    build_id: filters.build_id,
    map_name: filters.map_name,
    platform: filters.platform,
    since: filters.since,
    limit: filters.limit,
    offset: filters.offset,
  });
}

// Get single bundle by ID
export async function getBundle(bundleId: string): Promise<ReproBundle> {
  return api.get<ReproBundle>(`/repro-bundles/${bundleId}`);
}

// Legacy wrapper - adapts backend response to old frontend format
export async function getRepros(filters: ReproFilters = {}): Promise<GetReprosResponse> {
  const limit = filters.limit || 20;
  const page = filters.page || 1;
  const offset = (page - 1) * limit;

  const response = await getBundles({
    build_id: filters.build,
    map_name: filters.map,
    platform: filters.platform,
    limit,
    offset,
  });

  return {
    repros: response.bundles,
    total: response.total,
    page,
    pageSize: limit,
  };
}

// Legacy wrapper for single bundle
export async function getRepro(id: string): Promise<ReproBundle> {
  return getBundle(id);
}

// Get available filter options
// Note: Backend doesn't have a dedicated filters endpoint, 
// so this returns empty arrays - could be computed from bundle list
export async function getFilters(): Promise<GetFiltersResponse> {
  // Try to fetch bundles and extract unique values
  try {
    const response = await getBundles({ limit: 1000 });
    const builds = [...new Set(response.bundles.map(b => b.build_id))];
    const platforms = [...new Set(response.bundles.map(b => b.platform))];
    const maps = [...new Set(response.bundles.map(b => b.map_name).filter(Boolean))] as string[];
    const tags = [...new Set(response.bundles.flatMap(b => b.tags || []))];
    
    return { builds, platforms, maps, tags };
  } catch {
    return { builds: [], platforms: [], maps: [], tags: [] };
  }
}

// Upload bundle response
export interface UploadBundleResponse {
  bundle_id: string;
  status: 'ingested' | 'already_exists';
  artifact_count: number;
  created_at?: string;
}

// Upload a new bundle (ZIP file)
export async function uploadBundle(file: File): Promise<UploadBundleResponse> {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('/api/repro-bundles', {
    method: 'POST',
    body: formData,
    // Note: Don't set Content-Type header - browser will set it with boundary
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: { message: 'Upload failed' } }));
    throw new Error(errorData.error?.message || `Upload failed with status ${response.status}`);
  }

  return response.json();
}

// Add tags to a bundle
export async function addTags(bundleId: string, tags: string[]): Promise<void> {
  await api.post(`/repro-bundles/${bundleId}/tags`, { tags });
}

// Purge all bundles from the database
export interface PurgeResponse {
  status: string;
  bundles_purged: number;
}

export async function purgeAllBundles(): Promise<PurgeResponse> {
  return api.delete<PurgeResponse>('/repro-bundles');
}

// Get artifact URL for a bundle
export function getArtifactUrl(bundleId: string, artifactId: string): string {
  return `/api/repro-bundles/${bundleId}/artifacts/${artifactId}`;
}
