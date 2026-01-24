/**
 * Format milliseconds as MM:SS
 */
export function formatTime(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}

/**
 * Format milliseconds as MM:SS.mmm
 */
export function formatTimeMs(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  const millis = Math.floor(ms % 1000);
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}.${millis.toString().padStart(3, '0')}`;
}

/**
 * Format milliseconds as compact duration (M:SS or H:MM:SS)
 */
export function formatDuration(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  
  if (totalSeconds < 60) {
    return `0:${totalSeconds.toString().padStart(2, '0')}`;
  }
  
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  
  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
  }
  
  return `${minutes}:${seconds.toString().padStart(2, '0')}`;
}

/**
 * Format bytes as human-readable size
 */
export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

/**
 * Format ISO date string as relative time
 */
export function formatRelativeTime(isoDate: string): string {
  const date = new Date(isoDate);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);
  
  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  
  return date.toLocaleDateString(undefined, { 
    month: 'short', 
    day: 'numeric',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined 
  });
}

/**
 * Calculate appropriate tick interval for timeline ruler
 */
export function getTickInterval(durationMs: number): number {
  if (durationMs <= 10000) return 1000;        // 1s ticks for < 10s
  if (durationMs <= 30000) return 2000;        // 2s ticks for < 30s
  if (durationMs <= 60000) return 5000;        // 5s ticks for < 1min
  if (durationMs <= 180000) return 10000;      // 10s ticks for < 3min
  if (durationMs <= 300000) return 30000;      // 30s ticks for < 5min
  if (durationMs <= 600000) return 60000;      // 1min ticks for < 10min
  return 120000;                                // 2min ticks otherwise
}

/**
 * Binary search for first index where arr[i].timestampMs >= target
 */
export function binarySearchGte<T extends { timestampMs: number }>(
  arr: T[],
  targetMs: number
): number {
  let lo = 0;
  let hi = arr.length;
  
  while (lo < hi) {
    const mid = (lo + hi) >>> 1;
    if (arr[mid].timestampMs < targetMs) {
      lo = mid + 1;
    } else {
      hi = mid;
    }
  }
  
  return lo;
}

/**
 * Binary search for last index where arr[i].timestampMs <= target
 */
export function binarySearchLte<T extends { timestampMs: number }>(
  arr: T[],
  targetMs: number
): number {
  let lo = 0;
  let hi = arr.length;
  
  while (lo < hi) {
    const mid = (lo + hi) >>> 1;
    if (arr[mid].timestampMs <= targetMs) {
      lo = mid + 1;
    } else {
      hi = mid;
    }
  }
  
  return lo - 1;
}
