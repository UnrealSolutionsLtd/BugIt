import type { GetLogsResponse, LogLevel } from '../types';
import { getBundle, getArtifactUrl } from './repros';

export interface LogFilters {
  level?: LogLevel[];
  category?: string;
  search?: string;
}

// Get logs for a bundle by fetching the logs.txt artifact
export async function getLogs(
  bundleId: string, 
  filters: LogFilters = {}
): Promise<GetLogsResponse> {
  const bundle = await getBundle(bundleId);
  
  // Find logs artifact
  const logsArtifact = bundle.artifacts?.find(
    a => a.filename === 'logs.txt' || a.type === 'log'
  );
  
  if (!logsArtifact) {
    return { logs: [], categories: [] };
  }
  
  // Fetch the artifact content
  const url = getArtifactUrl(bundleId, logsArtifact.artifact_id);
  const response = await fetch(url);
  
  if (!response.ok) {
    throw new Error(`Failed to fetch logs: ${response.statusText}`);
  }
  
  const text = await response.text();
  
  // Parse log format: [FrameNumber|TimestampMs|Verbosity] Category: Message
  const logRegex = /^\[(\d+)\|([0-9.]+)\|(\w+)\]\s*(\w+):\s*(.*)$/;
  const categories = new Set<string>();
  
  const logs = text.split('\n')
    .filter(line => line.trim())
    .map(line => {
      const match = line.match(logRegex);
      if (match) {
        const category = match[4];
        categories.add(category);
        return {
          timestampMs: parseFloat(match[2]),
          level: match[3].toLowerCase() as LogLevel,
          category,
          message: match[5],
        };
      }
      // Fallback for unstructured logs
      return {
        timestampMs: 0,
        level: 'log' as LogLevel,
        category: 'Unknown',
        message: line,
      };
    })
    .filter(log => {
      // Apply filters
      if (filters.level?.length && !filters.level.includes(log.level)) {
        return false;
      }
      if (filters.category && log.category !== filters.category) {
        return false;
      }
      if (filters.search && !log.message.toLowerCase().includes(filters.search.toLowerCase())) {
        return false;
      }
      return true;
    });
  
  return { logs, categories: [...categories] };
}
