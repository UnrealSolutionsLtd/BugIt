import { useEffect, useRef } from 'react';
import { useTime } from '../../context/TimeContext';
import { formatTime } from '../../utils/time';
import styles from './VideoPlayer.module.css';

interface VideoPlayerProps {
  src: string;
  poster?: string;
}

export function VideoPlayer({ src, poster }: VideoPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const {
    currentTimeMs,
    durationMs,
    isPlaying,
    playbackRate,
    seek,
    toggle,
    setPlaybackRate,
    registerVideoRef,
  } = useTime();

  useEffect(() => {
    const video = videoRef.current;
    if (video) {
      const cleanup = registerVideoRef(video);
      return () => {
        registerVideoRef(null);
        if (cleanup) cleanup();
      };
    }
  }, [registerVideoRef]);

  const handleScrub = (e: React.ChangeEvent<HTMLInputElement>) => {
    seek(Number(e.target.value));
  };

  const playheadPercent = durationMs > 0 ? (currentTimeMs / durationMs) * 100 : 0;

  return (
    <div className={styles.container}>
      <div className={styles.videoWrapper}>
        <video
          ref={videoRef}
          src={src}
          poster={poster}
          className={styles.video}
          onClick={toggle}
          playsInline
        />
        
        {/* Play overlay */}
        {!isPlaying && (
          <button className={styles.playOverlay} onClick={toggle} aria-label="Play">
            <svg viewBox="0 0 24 24" fill="currentColor" width="64" height="64">
              <path d="M8 5v14l11-7z" />
            </svg>
          </button>
        )}
      </div>
      
      <div className={styles.controls}>
        <button 
          className={styles.playButton} 
          onClick={toggle}
          aria-label={isPlaying ? 'Pause' : 'Play'}
        >
          {isPlaying ? (
            <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
              <path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" />
            </svg>
          ) : (
            <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
              <path d="M8 5v14l11-7z" />
            </svg>
          )}
        </button>
        
        <span className={styles.time}>
          {formatTime(currentTimeMs)} / {formatTime(durationMs)}
        </span>
        
        <div className={styles.scrubberContainer}>
          <div 
            className={styles.scrubberProgress} 
            style={{ width: `${playheadPercent}%` }}
          />
          <input
            type="range"
            className={styles.scrubber}
            min={0}
            max={durationMs || 100}
            value={currentTimeMs}
            onChange={handleScrub}
          />
        </div>
        
        <select
          className={styles.rate}
          value={playbackRate}
          onChange={(e) => setPlaybackRate(Number(e.target.value))}
        >
          <option value={0.25}>0.25x</option>
          <option value={0.5}>0.5x</option>
          <option value={1}>1x</option>
          <option value={1.5}>1.5x</option>
          <option value={2}>2x</option>
        </select>
      </div>
    </div>
  );
}
