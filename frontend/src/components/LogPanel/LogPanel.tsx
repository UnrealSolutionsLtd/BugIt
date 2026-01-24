import { useState, useMemo, useCallback, useRef, useEffect } from 'react';
import { useTime } from '../../context/TimeContext';
import { formatTimeMs, binarySearchLte } from '../../utils/time';
import type { LogEntry, LogLevel } from '../../types';
import styles from './LogPanel.module.css';

interface LogPanelProps {
  logs: LogEntry[];
  categories: string[];
}

const LEVEL_ICONS: Record<LogLevel, string> = {
  verbose: '💬',
  log: 'ℹ️',
  warning: '⚠️',
  error: '❌',
};

export function LogPanel({ logs, categories }: LogPanelProps) {
  const { currentTimeMs, seek } = useTime();
  const [levelFilter, setLevelFilter] = useState<LogLevel | 'all'>('all');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const [autoScroll, setAutoScroll] = useState(true);
  
  const listRef = useRef<HTMLDivElement>(null);

  const filteredLogs = useMemo(() => {
    return logs.filter(log => {
      if (levelFilter !== 'all' && log.level !== levelFilter) return false;
      if (categoryFilter !== 'all' && log.category !== categoryFilter) return false;
      if (searchTerm && !log.message.toLowerCase().includes(searchTerm.toLowerCase())) {
        return false;
      }
      return true;
    });
  }, [logs, levelFilter, categoryFilter, searchTerm]);

  // Find the log entry closest to current time
  const currentLogIndex = useMemo(() => {
    if (filteredLogs.length === 0) return -1;
    return binarySearchLte(filteredLogs, currentTimeMs);
  }, [filteredLogs, currentTimeMs]);

  // Auto-scroll to current log
  useEffect(() => {
    if (autoScroll && listRef.current && currentLogIndex >= 0) {
      const item = listRef.current.children[currentLogIndex] as HTMLElement;
      if (item) {
        item.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
      }
    }
  }, [currentLogIndex, autoScroll]);

  const handleLogClick = useCallback((timestampMs: number) => {
    seek(timestampMs);
  }, [seek]);

  const handleScroll = useCallback(() => {
    // Disable auto-scroll when user scrolls manually
    setAutoScroll(false);
  }, []);

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <span className={styles.title}>LOGS</span>
        <div className={styles.filters}>
          <input
            type="text"
            placeholder="Search..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className={styles.search}
          />
          <select
            value={levelFilter}
            onChange={(e) => setLevelFilter(e.target.value as LogLevel | 'all')}
            className={styles.select}
          >
            <option value="all">All Levels</option>
            <option value="error">Errors</option>
            <option value="warning">Warnings</option>
            <option value="log">Logs</option>
            <option value="verbose">Verbose</option>
          </select>
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
            className={styles.select}
          >
            <option value="all">All Categories</option>
            {categories.map(cat => (
              <option key={cat} value={cat}>{cat}</option>
            ))}
          </select>
          <button
            className={`${styles.autoScrollBtn} ${autoScroll ? styles.active : ''}`}
            onClick={() => setAutoScroll(!autoScroll)}
            title="Auto-scroll to current time"
          >
            ↓
          </button>
        </div>
      </div>
      
      <div 
        ref={listRef} 
        className={styles.logList}
        onScroll={handleScroll}
      >
        {filteredLogs.length === 0 ? (
          <div className={styles.empty}>No logs match filters</div>
        ) : (
          filteredLogs.map((log, index) => (
            <div
              key={`${log.timestampMs}-${index}`}
              className={`${styles.logEntry} ${styles[log.level]} ${
                index === currentLogIndex ? styles.current : ''
              }`}
              onClick={() => handleLogClick(log.timestampMs)}
            >
              <span className={styles.timestamp}>
                {formatTimeMs(log.timestampMs)}
              </span>
              <span className={styles.level} title={log.level}>
                {LEVEL_ICONS[log.level]}
              </span>
              <span className={styles.category}>[{log.category}]</span>
              <span className={styles.message}>{log.message}</span>
            </div>
          ))
        )}
      </div>
      
      <div className={styles.footer}>
        {filteredLogs.length} of {logs.length} entries
      </div>
    </div>
  );
}
