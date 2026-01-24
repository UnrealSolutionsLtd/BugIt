import { Link } from 'react-router-dom';
import { formatRelativeTime, formatBytes } from '../../utils/time';
import type { ReproBundle } from '../../types';
import styles from './ReproCard.module.css';

interface ReproCardProps {
  repro: ReproBundle;
}

const PLATFORM_ICONS: Record<string, string> = {
  Win64: '🪟',
  Windows: '🪟',
  WindowsEditor: '🪟',
  Linux: '🐧',
  LinuxEditor: '🐧',
  Mac: '🍎',
  MacEditor: '🍎',
  PS5: '🎮',
  XSX: '🎮',
  Switch: '🎮',
  iOS: '📱',
  Android: '🤖',
};

export function ReproCard({ repro }: ReproCardProps) {
  // Extract map name from full path (e.g., "UEDPIE_0_LVL_StackOBot" -> "LVL_StackOBot")
  const displayMap = repro.map_name?.replace(/^UEDPIE_\d+_/, '') || 'Unknown Map';
  const tags = repro.tags || [];

  return (
    <Link to={`/repro/${repro.bundle_id}`} className={styles.card}>
      <div className={styles.thumbnail}>
        <div className={styles.placeholderThumb}>
          <span className={styles.bundleIcon}>📦</span>
          <span className={styles.artifactCount}>{repro.artifact_count} files</span>
        </div>
        <span className={styles.duration}>
          {formatBytes(repro.size_bytes)}
        </span>
      </div>
      
      <div className={styles.content}>
        <h3 className={styles.title}>{repro.bundle_id}</h3>
        
        <div className={styles.meta}>
          <span className={styles.build} title="Build version">
            {repro.build_id}
          </span>
          <span className={styles.separator}>•</span>
          <span className={styles.platform} title={repro.platform}>
            {PLATFORM_ICONS[repro.platform] ?? '💻'} {repro.platform}
          </span>
          <span className={styles.separator}>•</span>
          <span className={styles.map} title="Map name">
            {displayMap}
          </span>
        </div>
        
        <div className={styles.footer}>
          <span className={styles.date}>
            {formatRelativeTime(repro.created_at)}
          </span>
        </div>
        
        {tags.length > 0 && (
          <div className={styles.tags}>
            {tags.slice(0, 4).map(tag => (
              <span key={tag} className={styles.tag}>{tag}</span>
            ))}
            {tags.length > 4 && (
              <span className={styles.moreTag}>+{tags.length - 4}</span>
            )}
          </div>
        )}
      </div>
    </Link>
  );
}
