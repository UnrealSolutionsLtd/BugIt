export type LogLevel = 'verbose' | 'log' | 'warning' | 'error';

export interface LogEntry {
  timestampMs: number;
  level: LogLevel;
  category: string;
  message: string;
  file?: string;
  line?: number;
}

export interface GetLogsResponse {
  logs: LogEntry[];
  categories: string[];
}
