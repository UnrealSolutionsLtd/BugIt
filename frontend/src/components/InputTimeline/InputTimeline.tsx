import { useMemo } from 'react';
import { useTime } from '../../context/TimeContext';
import { getTickInterval } from '../../utils/time';
import type { KeyboardEvent, MouseEvent, GamepadEvent } from '../../types';
import styles from './InputTimeline.module.css';

interface InputTimelineProps {
  keyboard: KeyboardEvent[];
  mouse: MouseEvent[];
  gamepad: GamepadEvent[];
}

// A keyboard segment represents a contiguous period where keys were held
interface KeyboardSegment {
  startMs: number;
  endMs: number;
  keys: string[];
}

export function InputTimeline({ keyboard, mouse, gamepad }: InputTimelineProps) {
  const { currentTimeMs, durationMs, seek } = useTime();

  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const ratio = (e.clientX - rect.left) / rect.width;
    seek(ratio * durationMs);
  };

  // Convert keyboard events to visual segments
  const keyboardSegments = useMemo(() => 
    computeKeyboardSegments(keyboard, durationMs),
    [keyboard, durationMs]
  );

  // Filter mouse clicks
  const mouseClicks = useMemo(() => 
    mouse.filter(m => m.type === 'down'),
    [mouse]
  );

  // Filter gamepad button presses
  const gamepadPresses = useMemo(() =>
    gamepad.filter(g => g.type === 'button' && g.value > 0.5),
    [gamepad]
  );

  const playheadPosition = durationMs > 0 
    ? (currentTimeMs / durationMs) * 100 
    : 0;

  return (
    <div className={styles.container}>
      <div className={styles.header}>INPUT TIMELINE</div>
      
      <div className={styles.tracks} onClick={handleClick}>
        {/* Playhead */}
        <div 
          className={styles.playhead} 
          style={{ left: `${playheadPosition}%` }}
        />
        
        {/* Keyboard Track */}
        <div className={styles.track}>
          <span className={styles.label}>KBD</span>
          <div className={styles.trackContent}>
            {keyboardSegments.length === 0 ? (
              <span className={styles.noData}>No keyboard input</span>
            ) : (
              keyboardSegments.map((seg, i) => {
                // Handle segments that start before timeline (negative timestamps)
                const clampedStart = Math.max(0, seg.startMs);
                const clampedEnd = Math.min(durationMs, seg.endMs);
                // Skip segments entirely before timeline
                if (clampedEnd <= 0) return null;
                const leftPercent = (clampedStart / durationMs) * 100;
                const widthPercent = Math.max(0.5, ((clampedEnd - clampedStart) / durationMs) * 100);
                return (
                  <div
                    key={i}
                    className={styles.segment}
                    style={{
                      left: `${leftPercent}%`,
                      width: `${widthPercent}%`,
                    }}
                    title={seg.keys.join(' + ')}
                  />
                );
              })
            )}
          </div>
        </div>
        
        {/* Mouse Track */}
        <div className={styles.track}>
          <span className={styles.label}>Mouse</span>
          <div className={styles.trackContent}>
            {mouseClicks.map((m, i) => (
              <div
                key={i}
                className={`${styles.mouseClick} ${styles[`button${m.button ?? 0}`]}`}
                style={{ left: `${(m.timestampMs / durationMs) * 100}%` }}
                title={`Button ${m.button ?? 0} at (${m.x}, ${m.y})`}
              />
            ))}
          </div>
        </div>
        
        {/* Gamepad Track */}
        <div className={styles.track}>
          <span className={styles.label}>Pad</span>
          <div className={styles.trackContent}>
            {gamepadPresses.length === 0 ? (
              <span className={styles.noData}>No gamepad input</span>
            ) : (
              gamepadPresses.map((g, i) => (
                <div
                  key={i}
                  className={styles.gamepadPress}
                  style={{ left: `${(g.timestampMs / durationMs) * 100}%` }}
                  title={`Button ${g.index}`}
                />
              ))
            )}
          </div>
        </div>
        
        {/* Time Ruler */}
        <TimeRuler durationMs={durationMs} />
      </div>
    </div>
  );
}

function TimeRuler({ durationMs }: { durationMs: number }) {
  const ticks = useMemo(() => {
    if (durationMs <= 0) return [];
    
    const interval = getTickInterval(durationMs);
    const result = [];
    for (let t = 0; t <= durationMs; t += interval) {
      result.push({
        timeMs: t,
        position: (t / durationMs) * 100,
      });
    }
    return result;
  }, [durationMs]);

  return (
    <div className={styles.ruler}>
      {ticks.map(tick => (
        <span
          key={tick.timeMs}
          className={styles.tick}
          style={{ left: `${tick.position}%` }}
        >
          {Math.round(tick.timeMs / 1000)}s
        </span>
      ))}
    </div>
  );
}

// Convert keyboard down/up events into visual segments
function computeKeyboardSegments(
  events: KeyboardEvent[],
  durationMs: number
): KeyboardSegment[] {
  const segments: KeyboardSegment[] = [];
  const activeKeys = new Map<string, number>(); // key -> startTime
  
  for (const event of events) {
    if (event.type === 'down' && !activeKeys.has(event.key)) {
      activeKeys.set(event.key, event.timestampMs);
    } else if (event.type === 'up') {
      if (activeKeys.has(event.key)) {
        // Normal case: KeyDown was seen, create segment
        const startMs = activeKeys.get(event.key)!;
        segments.push({
          startMs,
          endMs: event.timestampMs,
          keys: [event.key],
        });
        activeKeys.delete(event.key);
      } else {
        // Orphaned KeyUp: key was held before capture started
        // Treat as if KeyDown happened at time 0
        segments.push({
          startMs: 0,
          endMs: event.timestampMs,
          keys: [event.key],
        });
      }
    }
  }
  
  // Close any still-active keys at end of recording (orphaned KeyDown)
  for (const [key, startMs] of activeKeys) {
    segments.push({
      startMs,
      endMs: durationMs,
      keys: [key],
    });
  }
  
  // Sort by start time
  segments.sort((a, b) => a.startMs - b.startMs);
  
  // Merge overlapping segments
  return mergeOverlappingSegments(segments);
}

// Merge segments that overlap in time
function mergeOverlappingSegments(segments: KeyboardSegment[]): KeyboardSegment[] {
  if (segments.length === 0) return [];
  
  const merged: KeyboardSegment[] = [];
  let current = { ...segments[0], keys: [...segments[0].keys] };
  
  for (let i = 1; i < segments.length; i++) {
    const seg = segments[i];
    
    // Check if segments overlap
    if (seg.startMs <= current.endMs) {
      // Extend current segment and add keys
      current.endMs = Math.max(current.endMs, seg.endMs);
      if (!current.keys.includes(seg.keys[0])) {
        current.keys.push(seg.keys[0]);
      }
    } else {
      // No overlap, push current and start new
      merged.push(current);
      current = { ...seg, keys: [...seg.keys] };
    }
  }
  
  merged.push(current);
  return merged;
}
