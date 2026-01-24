import { useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getRepro, getInputs, getLogs, getFrames, getArtifactUrl } from '../api';
import { TimeProvider } from '../context/TimeContext';
import { VideoPlayer } from '../components/VideoPlayer';
import { InputTimeline } from '../components/InputTimeline';
import { FrameTimingGraph } from '../components/FrameTimingGraph';
import { LogPanel } from '../components/LogPanel';
import { formatBytes } from '../utils/time';
import styles from './ReproViewerPage.module.css';

export function ReproViewerPage() {
  const { id } = useParams<{ id: string }>();

  // Fetch all repro data in parallel
  const { data: repro, isLoading: loadingRepro } = useQuery({
    queryKey: ['repro', id],
    queryFn: () => getRepro(id!),
    enabled: !!id,
  });

  const { data: inputs } = useQuery({
    queryKey: ['inputs', id],
    queryFn: () => getInputs(id!),
    enabled: !!id,
  });

  const { data: logs } = useQuery({
    queryKey: ['logs', id],
    queryFn: () => getLogs(id!),
    enabled: !!id,
  });

  const { data: frames } = useQuery({
    queryKey: ['frames', id],
    queryFn: () => getFrames(id!),
    enabled: !!id,
  });

  // Find video artifact for playback
  const videoArtifact = useMemo(() => {
    return repro?.artifacts?.find(a => a.type === 'video');
  }, [repro?.artifacts]);

  const videoUrl = useMemo(() => {
    if (videoArtifact && repro) {
      return getArtifactUrl(repro.bundle_id, videoArtifact.artifact_id);
    }
    return undefined;
  }, [videoArtifact, repro]);

  // Extract display map name
  const displayMap = repro?.map_name?.replace(/^UEDPIE_\d+_/, '') || 'Unknown';

  if (loadingRepro) {
    return (
      <div className={styles.loading}>
        Loading repro...
      </div>
    );
  }

  if (!repro) {
    return (
      <div className={styles.error}>
        <h2>Repro not found</h2>
        <Link to="/">Back to list</Link>
      </div>
    );
  }

  const tags = repro.tags || [];

  return (
    <TimeProvider>
      <div className={styles.container}>
        {/* Header */}
        <header className={styles.header}>
          <Link to="/" className={styles.backBtn}>
            ← Back
          </Link>
          <h1 className={styles.title}>{repro.bundle_id}</h1>
          <button className={styles.exportBtn}>
            Export
          </button>
        </header>

        {/* Main content */}
        <div className={styles.content}>
          {/* Left column: Video and timelines */}
          <div className={styles.mainColumn}>
            {videoUrl ? (
              <VideoPlayer src={videoUrl} />
            ) : (
              <div className={styles.noVideo}>No video available</div>
            )}

            {inputs && (
              <InputTimeline
                keyboard={inputs.keyboard}
                mouse={inputs.mouse}
                gamepad={inputs.gamepad}
              />
            )}

            {frames && (
              <FrameTimingGraph
                samples={frames.samples}
                summary={frames.summary}
              />
            )}

            {logs && (
              <LogPanel
                logs={logs.logs}
                categories={logs.categories}
              />
            )}
          </div>

          {/* Right column: Metadata sidebar */}
          <aside className={styles.sidebar}>
            <section className={styles.section}>
              <h3 className={styles.sectionTitle}>Metadata</h3>
              <dl className={styles.metaList}>
                <dt>Build</dt>
                <dd className={styles.monospace}>{repro.build_id}</dd>
                
                <dt>Platform</dt>
                <dd>{repro.platform}</dd>
                
                <dt>Map</dt>
                <dd className={styles.monospace}>{displayMap}</dd>
                
                <dt>Date</dt>
                <dd>{new Date(repro.created_at).toLocaleString()}</dd>
                
                <dt>Size</dt>
                <dd>{formatBytes(repro.size_bytes)}</dd>
                
                <dt>Artifacts</dt>
                <dd>{repro.artifact_count} files</dd>
              </dl>
            </section>

            {tags.length > 0 && (
              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Tags</h3>
                <div className={styles.tags}>
                  {tags.map(tag => (
                    <span key={tag} className={styles.tag}>{tag}</span>
                  ))}
                </div>
              </section>
            )}

            <section className={styles.section}>
              <h3 className={styles.sectionTitle}>Bundle</h3>
              <dl className={styles.metaList}>
                <dt>Schema</dt>
                <dd className={styles.monospace}>{repro.schema_version}</dd>
                
                <dt>Hash</dt>
                <dd className={`${styles.monospace} ${styles.hash}`}>
                  {repro.content_hash.substring(0, 20)}...
                </dd>

                {repro.rvr_version && (
                  <>
                    <dt>RVR Version</dt>
                    <dd className={styles.monospace}>{repro.rvr_version}</dd>
                  </>
                )}
              </dl>
            </section>

            {repro.artifacts && repro.artifacts.length > 0 && (
              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>Artifacts</h3>
                <ul className={styles.artifactList}>
                  {repro.artifacts.map(artifact => (
                    <li key={artifact.artifact_id} className={styles.artifact}>
                      <a 
                        href={getArtifactUrl(repro.bundle_id, artifact.artifact_id)}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        {artifact.filename}
                      </a>
                      <span className={styles.artifactSize}>
                        {formatBytes(artifact.size_bytes)}
                      </span>
                    </li>
                  ))}
                </ul>
              </section>
            )}

            {repro.qa_notes && repro.qa_notes.length > 0 && (
              <section className={styles.section}>
                <h3 className={styles.sectionTitle}>QA Notes</h3>
                {repro.qa_notes.map(note => (
                  <div key={note.note_id} className={styles.note}>
                    <p className={styles.noteContent}>{note.content}</p>
                    <span className={styles.noteAuthor}>— {note.author}</span>
                  </div>
                ))}
              </section>
            )}
          </aside>
        </div>

        {/* Keyboard shortcuts help */}
        <footer className={styles.footer}>
          <span>Space: Play/Pause</span>
          <span>←/→: Seek ±5s</span>
          <span>J/L: ±10s</span>
          <span>,/.: Frame step</span>
        </footer>
      </div>
    </TimeProvider>
  );
}
