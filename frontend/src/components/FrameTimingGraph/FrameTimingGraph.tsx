import { useRef, useEffect, useMemo, useCallback } from 'react';
import { useTime } from '../../context/TimeContext';
import type { FrameSample, FrameSummary } from '../../types';
import styles from './FrameTimingGraph.module.css';

interface FrameTimingGraphProps {
  samples: FrameSample[];
  summary: FrameSummary;
}

const MIN_FPS = 0;
const MAX_FPS = 120;
const PADDING = { top: 20, right: 50, bottom: 24, left: 10 };

export function FrameTimingGraph({ samples, summary }: FrameTimingGraphProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const { currentTimeMs, durationMs, seek } = useTime();

  // Downsample for performance if needed
  const displaySamples = useMemo(() => {
    if (samples.length <= 1000) return samples;
    const step = Math.ceil(samples.length / 1000);
    return samples.filter((_, i) => i % step === 0);
  }, [samples]);

  // Resize canvas to container
  useEffect(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;

    const resizeObserver = new ResizeObserver(() => {
      const { width, height } = container.getBoundingClientRect();
      canvas.width = width * window.devicePixelRatio;
      canvas.height = height * window.devicePixelRatio;
      canvas.style.width = `${width}px`;
      canvas.style.height = `${height}px`;
    });

    resizeObserver.observe(container);
    return () => resizeObserver.disconnect();
  }, []);

  // Draw graph
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    const dpr = window.devicePixelRatio;
    const width = canvas.width / dpr;
    const height = canvas.height / dpr;
    
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    
    const graphWidth = width - PADDING.left - PADDING.right;
    const graphHeight = height - PADDING.top - PADDING.bottom;

    // Clear
    ctx.fillStyle = '#1a1a2e';
    ctx.fillRect(0, 0, width, height);

    if (durationMs <= 0 || displaySamples.length === 0) {
      ctx.fillStyle = '#666';
      ctx.font = '12px Inter, sans-serif';
      ctx.fillText('No frame timing data', width / 2 - 60, height / 2);
      return;
    }

    // Draw reference lines
    ctx.strokeStyle = '#333';
    ctx.setLineDash([4, 4]);
    ctx.lineWidth = 1;
    
    // 60 FPS line
    const y60 = PADDING.top + graphHeight * (1 - (60 - MIN_FPS) / (MAX_FPS - MIN_FPS));
    ctx.beginPath();
    ctx.moveTo(PADDING.left, y60);
    ctx.lineTo(width - PADDING.right, y60);
    ctx.stroke();
    
    ctx.fillStyle = '#666';
    ctx.font = '10px JetBrains Mono, monospace';
    ctx.fillText('60fps', width - PADDING.right + 4, y60 + 3);
    
    // 30 FPS line
    const y30 = PADDING.top + graphHeight * (1 - (30 - MIN_FPS) / (MAX_FPS - MIN_FPS));
    ctx.beginPath();
    ctx.moveTo(PADDING.left, y30);
    ctx.lineTo(width - PADDING.right, y30);
    ctx.stroke();
    ctx.fillText('30fps', width - PADDING.right + 4, y30 + 3);

    ctx.setLineDash([]);

    // Draw problem areas (FPS < 30) as background
    ctx.fillStyle = 'rgba(239, 68, 68, 0.15)';
    displaySamples.forEach((sample) => {
      if (sample.fps < 30) {
        const x = PADDING.left + (sample.timestampMs / durationMs) * graphWidth;
        const barWidth = Math.max(2, graphWidth / displaySamples.length);
        ctx.fillRect(x - barWidth / 2, PADDING.top, barWidth, graphHeight);
      }
    });

    // Draw FPS line
    ctx.strokeStyle = '#4ade80';
    ctx.lineWidth = 1.5;
    ctx.beginPath();
    
    displaySamples.forEach((sample, i) => {
      const x = PADDING.left + (sample.timestampMs / durationMs) * graphWidth;
      const fps = Math.min(MAX_FPS, Math.max(MIN_FPS, sample.fps));
      const y = PADDING.top + graphHeight * (1 - (fps - MIN_FPS) / (MAX_FPS - MIN_FPS));
      
      if (i === 0) {
        ctx.moveTo(x, y);
      } else {
        ctx.lineTo(x, y);
      }
    });
    
    ctx.stroke();

    // Draw playhead
    const playheadX = PADDING.left + (currentTimeMs / durationMs) * graphWidth;
    ctx.strokeStyle = '#fff';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(playheadX, PADDING.top);
    ctx.lineTo(playheadX, height - PADDING.bottom);
    ctx.stroke();

    // Playhead indicator
    ctx.fillStyle = '#fff';
    ctx.beginPath();
    ctx.moveTo(playheadX - 4, PADDING.top);
    ctx.lineTo(playheadX + 4, PADDING.top);
    ctx.lineTo(playheadX, PADDING.top + 6);
    ctx.closePath();
    ctx.fill();

    // Draw time axis
    ctx.fillStyle = '#666';
    ctx.font = '10px JetBrains Mono, monospace';
    const tickCount = Math.min(10, Math.floor(durationMs / 5000) + 1);
    for (let i = 0; i <= tickCount; i++) {
      const t = (i / tickCount) * durationMs;
      const x = PADDING.left + (t / durationMs) * graphWidth;
      ctx.fillText(`${Math.round(t / 1000)}s`, x - 8, height - 6);
    }

  }, [displaySamples, currentTimeMs, durationMs]);

  const handleClick = useCallback((e: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current;
    if (!canvas || durationMs <= 0) return;
    
    const rect = canvas.getBoundingClientRect();
    const x = e.clientX - rect.left - PADDING.left;
    const graphWidth = rect.width - PADDING.left - PADDING.right;
    
    const ratio = Math.max(0, Math.min(1, x / graphWidth));
    seek(ratio * durationMs);
  }, [durationMs, seek]);

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <span className={styles.title}>FRAME TIMING</span>
        <div className={styles.summary}>
          <span className={styles.stat} title="Average frames per second across all samples">
            Avg: <strong>{summary.avgFps.toFixed(0)}</strong> FPS
          </span>
          <span className={styles.stat} title="Lowest recorded frames per second">
            Min: <strong className={summary.minFps < 30 ? styles.bad : ''}>
              {summary.minFps.toFixed(0)}
            </strong> FPS
          </span>
          <span className={styles.stat} title="Number of frames that dropped below 30 FPS">
            Stutters: <strong className={summary.stutterCount > 0 ? styles.bad : ''}>
              {summary.stutterCount}
            </strong>
          </span>
        </div>
      </div>
      <div ref={containerRef} className={styles.canvasContainer}>
        <canvas
          ref={canvasRef}
          className={styles.canvas}
          onClick={handleClick}
        />
      </div>
    </div>
  );
}
