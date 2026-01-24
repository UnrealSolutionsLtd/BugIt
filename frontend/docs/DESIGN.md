# BugIt - Forensic QA Dashboard Design

## Overview

BugIt is an on-premises forensic QA dashboard for visualizing immutable repro bundles produced by Runtime Video Recorder (RVR). Designed for QA engineers and developers investigating gameplay issues.

---

## 1. UI Wireframe Descriptions

### Screen 1: Repro List (`/`)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  BugIt                                              [Refresh] [Settings ⚙]  │
├─────────────────────────────────────────────────────────────────────────────┤
│  Filters:                                                                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │ Build ▼  │ │Platform▼ │ │  Map ▼   │ │ Date ▼   │ │ Search tags...   │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────────────┘  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ ▶ [Thumb]  BUG-1234: Player falls through floor                       │  │
│  │            Build: 2.4.1-rc2  |  Win64  |  Map: L_CityCenter           │  │
│  │            2026-01-21 14:32  |  Duration: 00:47  |  Tags: physics     │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ ▶ [Thumb]  BUG-1235: Animation stuck in T-pose                        │  │
│  │            Build: 2.4.1-rc2  |  PS5  |  Map: L_Tutorial               │  │
│  │            2026-01-21 13:15  |  Duration: 01:23  |  Tags: animation   │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  Showing 24 of 156 repros                              [< 1 2 3 4 5 ... >]  │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Key Elements:**
- Header with app name, refresh, settings
- Filter bar with dropdowns and search
- Repro cards with thumbnail, title, metadata
- Pagination controls

---

### Screen 2: Repro Viewer (`/repro/:id`)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ← Back to List    BUG-1234: Player falls through floor         [Export ↓]  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌────────────────────────────────────────────────┐ ┌─────────────────────┐ │
│  │                                                │ │ METADATA            │ │
│  │                                                │ │                     │ │
│  │              VIDEO PLAYER                      │ │ Build: 2.4.1-rc2    │ │
│  │                                                │ │ Platform: Win64     │ │
│  │              [Gameplay footage]                │ │ Map: L_CityCenter   │ │
│  │                                                │ │ Date: 2026-01-21    │ │
│  │                                                │ │ Duration: 00:47     │ │
│  │                                                │ │ Reporter: jsmith    │ │
│  │                                                │ │                     │ │
│  ├────────────────────────────────────────────────┤ │ TAGS                │ │
│  │  ▶ ▐▐  ■   ────●────────────────  00:23/00:47  │ │ physics, collision  │ │
│  └────────────────────────────────────────────────┘ │ player-movement     │ │
│                                                     │                     │ │
│  ┌────────────────────────────────────────────────┐ │ SYSTEM INFO         │ │
│  │ INPUT TIMELINE                                 │ │ GPU: RTX 4080       │ │
│  │ ─────────────────────────────────────────────  │ │ CPU: i9-13900K      │ │
│  │ KBD  ▓▓  ▓▓▓  ▓     ▓▓▓▓▓  ▓▓   ▓  ▓         │ │ RAM: 32GB           │ │
│  │ Mouse───○────○─────○──○○───────○──────○───────│ │ OS: Win 11 23H2     │ │
│  │ Pad  ────────────────────────────────────────  │ │                     │ │
│  │ 0s        10s       20s       30s       47s    │ └─────────────────────┘ │
│  └────────────────────────────────────────────────┘                         │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │ FRAME TIMING                                           FPS │ ms        │ │
│  │ 60fps ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │ │
│  │       ████████████████████████████████████▄▂▁▅██████████████████████  │ │
│  │ 30fps ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │ │
│  │       │   FPS drop here!                                              │ │
│  │       0s        10s       20s       30s       47s                     │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │ LOGS                                                    [Filter: All▼] │ │
│  │ ─────────────────────────────────────────────────────────────────────  │ │
│  │ 00:22.340 [Warning] PhysicsEngine: Collision overlap detected          │ │
│  │ 00:22.341 [Error]   CharacterMovement: Invalid floor result            │ │
│  │ 00:22.350 [Warning] PhysicsEngine: Penetration correction failed       │ │
│  │ 00:23.010 [Error]   CharacterMovement: Fell out of world               │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Key Elements:**
- Video player with standard controls
- Synchronized scrubber across all timelines
- Input timeline showing keyboard, mouse, gamepad activity
- Frame timing graph with FPS visualization
- Scrollable log panel with severity filtering
- Metadata sidebar

---

### Screen 3: QA Notes Panel (Within Repro Viewer)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  QA NOTES                                              [+ Add Note at 00:23]│
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ 📍 00:22 - jsmith                                     2026-01-21 14:45│  │
│  │ ─────────────────────────────────────────────────────────────────────  │  │
│  │ Player started jumping on the stairs. Notice the **collision gaps**   │  │
│  │ between the steps. This seems related to BUG-1189.                    │  │
│  │                                                           [Edit] [🗑] │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ 📍 00:23 - jsmith                                     2026-01-21 14:46│  │
│  │ ─────────────────────────────────────────────────────────────────────  │  │
│  │ **THE BUG OCCURS HERE**                                               │  │
│  │                                                                        │  │
│  │ Player falls through floor at exact frame when `Fell out of world`    │  │
│  │ error appears. Look at the penetration correction failure in logs.    │  │
│  │                                                           [Edit] [🗑] │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Component Hierarchy

```
App
├── Layout
│   ├── Header
│   │   ├── Logo
│   │   ├── Navigation
│   │   └── GlobalActions (Refresh, Settings)
│   └── Main (outlet)
│
├── Pages
│   ├── ReproListPage
│   │   ├── FilterBar
│   │   │   ├── FilterDropdown (Build)
│   │   │   ├── FilterDropdown (Platform)
│   │   │   ├── FilterDropdown (Map)
│   │   │   ├── DateRangePicker
│   │   │   └── SearchInput
│   │   ├── ReproList
│   │   │   └── ReproCard[]
│   │   │       ├── Thumbnail
│   │   │       ├── ReproTitle
│   │   │       ├── MetadataBadges
│   │   │       └── TagList
│   │   └── Pagination
│   │
│   └── ReproViewerPage
│       ├── ViewerHeader
│       │   ├── BackButton
│       │   ├── ReproTitle
│       │   └── ExportButton
│       ├── MainContent
│       │   ├── VideoSection
│       │   │   ├── VideoPlayer
│       │   │   │   ├── VideoCanvas
│       │   │   │   └── VideoControls
│       │   │   │       ├── PlayPauseButton
│       │   │   │       ├── TimeDisplay
│       │   │   │       ├── Scrubber
│       │   │   │       ├── VolumeControl
│       │   │   │       └── FullscreenButton
│       │   │   └── InputTimeline
│       │   │       ├── TimelineRuler
│       │   │       ├── KeyboardTrack
│       │   │       ├── MouseTrack
│       │   │       └── GamepadTrack
│       │   ├── FrameTimingGraph
│       │   │   ├── GraphCanvas
│       │   │   ├── FPSLine
│       │   │   ├── FrameTimeLine
│       │   │   └── Annotations
│       │   └── LogPanel
│       │       ├── LogFilter
│       │       ├── LogList
│       │       │   └── LogEntry[]
│       │       └── LogTimeline (mini)
│       └── Sidebar
│           ├── MetadataPanel
│           │   ├── BuildInfo
│           │   ├── SystemInfo
│           │   └── TagList
│           └── NotesPanel
│               ├── NotesList
│               │   └── NoteCard[]
│               └── AddNoteForm
│
└── Shared Components
    ├── Timeline (base component)
    ├── TimeSync (context provider)
    ├── Tooltip
    ├── Badge
    ├── Button
    ├── Dropdown
    ├── Modal
    └── MarkdownRenderer
```

---

## 3. Timeline Alignment Strategy

### Core Concept: Unified Time Context

All timeline-based components subscribe to a shared time state. When any component changes the current time (video scrub, click on log, click on graph), all others update.

```typescript
// TimeContext.ts
interface TimeState {
  currentTimeMs: number;        // Current playback position
  durationMs: number;           // Total duration
  isPlaying: boolean;
  playbackRate: number;
  
  // Selection for zoom/detail
  selectionStartMs: number | null;
  selectionEndMs: number | null;
}

interface TimeActions {
  seek: (timeMs: number) => void;
  play: () => void;
  pause: () => void;
  setPlaybackRate: (rate: number) => void;
  setSelection: (start: number, end: number) => void;
  clearSelection: () => void;
}
```

### Synchronization Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                         TimeContext Provider                         │
│                                                                      │
│  currentTimeMs: 23450                                                │
│  durationMs: 47000                                                   │
│  isPlaying: false                                                    │
└──────────────────────────────┬───────────────────────────────────────┘
                               │
       ┌───────────────────────┼───────────────────────┐
       │                       │                       │
       ▼                       ▼                       ▼
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│ VideoPlayer │         │InputTimeline│         │FrameGraph   │
│             │         │             │         │             │
│ Seeks video │         │ Highlights  │         │ Draws       │
│ to 23.45s   │         │ current pos │         │ playhead    │
└─────────────┘         └─────────────┘         └─────────────┘
       │                       │                       │
       │ onTimeUpdate          │ onClick               │ onClick
       │                       │                       │
       └───────────────────────┴───────────────────────┘
                               │
                               ▼
                     dispatch(seek(newTimeMs))
```

### Data Alignment Requirements

All timeline data uses milliseconds from video start (0ms = first frame).

```typescript
// Input event alignment
interface InputEvent {
  timestampMs: number;
  type: 'keydown' | 'keyup' | 'mousedown' | 'mouseup' | 'mousemove' | 'gamepad';
  data: KeyboardData | MouseData | GamepadData;
}

// Log alignment
interface LogEntry {
  timestampMs: number;
  level: 'verbose' | 'log' | 'warning' | 'error';
  category: string;
  message: string;
}

// Frame timing alignment
interface FrameSample {
  timestampMs: number;
  frameTimeMs: number;
  fps: number;
}
```

### Efficient Rendering Strategy

```typescript
// Binary search for visible range
function getVisibleEvents(
  events: InputEvent[],
  viewStartMs: number,
  viewEndMs: number
): InputEvent[] {
  const startIdx = binarySearchGte(events, viewStartMs, e => e.timestampMs);
  const endIdx = binarySearchLte(events, viewEndMs, e => e.timestampMs);
  return events.slice(startIdx, endIdx + 1);
}

// Virtualization for logs
function useVirtualizedLogs(
  logs: LogEntry[],
  containerHeight: number,
  rowHeight: number
) {
  const visibleCount = Math.ceil(containerHeight / rowHeight);
  const [scrollTop, setScrollTop] = useState(0);
  const startIndex = Math.floor(scrollTop / rowHeight);
  
  return {
    visibleLogs: logs.slice(startIndex, startIndex + visibleCount + 2),
    totalHeight: logs.length * rowHeight,
    offsetY: startIndex * rowHeight,
  };
}
```

---

## 4. Data Contracts with Backend

### API Endpoints

```typescript
// Base URL: http://localhost:3001/api/v1

// ─────────────────────────────────────────────────────────────────────
// REPRO LIST
// ─────────────────────────────────────────────────────────────────────

// GET /repros
// Query params: build, platform, map, dateFrom, dateTo, search, page, limit
interface GetReprosResponse {
  repros: ReproSummary[];
  total: number;
  page: number;
  pageSize: number;
}

interface ReproSummary {
  id: string;                    // UUID
  title: string;                 // "Player falls through floor"
  thumbnailUrl: string;          // "/repros/{id}/thumbnail.jpg"
  build: string;                 // "2.4.1-rc2"
  platform: Platform;            // "Win64" | "PS5" | "XSX" | "Switch" | ...
  map: string;                   // "L_CityCenter"
  createdAt: string;             // ISO 8601
  durationMs: number;
  tags: string[];
  reporter: string;
}

// ─────────────────────────────────────────────────────────────────────
// REPRO DETAIL
// ─────────────────────────────────────────────────────────────────────

// GET /repros/:id
interface GetReproResponse {
  id: string;
  title: string;
  description: string;
  build: string;
  platform: Platform;
  map: string;
  createdAt: string;
  durationMs: number;
  tags: string[];
  reporter: string;
  
  // System info captured at recording time
  systemInfo: {
    os: string;
    osVersion: string;
    cpu: string;
    gpu: string;
    ramGB: number;
    gpuDriverVersion: string;
  };
  
  // Media URLs
  videoUrl: string;              // "/repros/{id}/video.mp4"
  thumbnailUrl: string;
  
  // Bundle info
  bundleVersion: string;         // RVR bundle format version
  bundleHash: string;            // SHA256 for integrity
}

// ─────────────────────────────────────────────────────────────────────
// INPUT EVENTS
// ─────────────────────────────────────────────────────────────────────

// GET /repros/:id/inputs
interface GetInputsResponse {
  keyboard: KeyboardEvent[];
  mouse: MouseEvent[];
  gamepad: GamepadEvent[];
}

interface KeyboardEvent {
  timestampMs: number;
  type: 'down' | 'up';
  key: string;                   // "W", "Space", "Shift", etc.
  keyCode: number;
}

interface MouseEvent {
  timestampMs: number;
  type: 'down' | 'up' | 'move' | 'wheel';
  button?: number;               // 0=left, 1=middle, 2=right
  x: number;
  y: number;
  deltaX?: number;
  deltaY?: number;
}

interface GamepadEvent {
  timestampMs: number;
  type: 'button' | 'axis';
  index: number;                 // Button or axis index
  value: number;                 // 0-1 for buttons, -1 to 1 for axes
}

// ─────────────────────────────────────────────────────────────────────
// FRAME TIMING
// ─────────────────────────────────────────────────────────────────────

// GET /repros/:id/frames
interface GetFramesResponse {
  samples: FrameSample[];
  summary: {
    avgFps: number;
    minFps: number;
    maxFps: number;
    p99FrameTimeMs: number;
    stutterCount: number;        // Frames > 2x avg frame time
  };
}

interface FrameSample {
  timestampMs: number;
  frameTimeMs: number;
  gameThreadMs: number;
  renderThreadMs: number;
  gpuTimeMs: number;
}

// ─────────────────────────────────────────────────────────────────────
// LOGS
// ─────────────────────────────────────────────────────────────────────

// GET /repros/:id/logs
// Query params: level (comma-separated), category, search
interface GetLogsResponse {
  logs: LogEntry[];
  categories: string[];          // All unique categories for filtering
}

interface LogEntry {
  timestampMs: number;
  level: 'verbose' | 'log' | 'warning' | 'error';
  category: string;              // "PhysicsEngine", "CharacterMovement", etc.
  message: string;
  file?: string;                 // Source file if available
  line?: number;
}

// ─────────────────────────────────────────────────────────────────────
// QA NOTES
// ─────────────────────────────────────────────────────────────────────

// GET /repros/:id/notes
interface GetNotesResponse {
  notes: Note[];
}

interface Note {
  id: string;
  timestampMs: number;           // Video timestamp this note refers to
  author: string;
  content: string;               // Markdown
  createdAt: string;
  updatedAt: string;
}

// POST /repros/:id/notes
interface CreateNoteRequest {
  timestampMs: number;
  author: string;
  content: string;
}

// PUT /repros/:id/notes/:noteId
interface UpdateNoteRequest {
  content: string;
}

// DELETE /repros/:id/notes/:noteId

// ─────────────────────────────────────────────────────────────────────
// FILTER OPTIONS
// ─────────────────────────────────────────────────────────────────────

// GET /filters
interface GetFiltersResponse {
  builds: string[];              // ["2.4.1-rc2", "2.4.1-rc1", "2.4.0", ...]
  platforms: Platform[];
  maps: string[];
  tags: string[];
}

type Platform = 'Win64' | 'Linux' | 'Mac' | 'PS5' | 'XSX' | 'Switch' | 'iOS' | 'Android';
```

### WebSocket Events (Optional, for live updates)

```typescript
// ws://localhost:3001/ws

// Server -> Client
interface NewReproEvent {
  type: 'repro:new';
  repro: ReproSummary;
}

interface NoteAddedEvent {
  type: 'note:added';
  reproId: string;
  note: Note;
}
```

---

## 5. Example React Components

### TimeContext Provider

```tsx
// src/context/TimeContext.tsx
import { createContext, useContext, useState, useCallback, useRef, ReactNode } from 'react';

interface TimeState {
  currentTimeMs: number;
  durationMs: number;
  isPlaying: boolean;
  playbackRate: number;
}

interface TimeContextValue extends TimeState {
  seek: (timeMs: number) => void;
  play: () => void;
  pause: () => void;
  toggle: () => void;
  setPlaybackRate: (rate: number) => void;
  setDuration: (durationMs: number) => void;
  registerVideoRef: (video: HTMLVideoElement | null) => void;
}

const TimeContext = createContext<TimeContextValue | null>(null);

export function TimeProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<TimeState>({
    currentTimeMs: 0,
    durationMs: 0,
    isPlaying: false,
    playbackRate: 1,
  });
  
  const videoRef = useRef<HTMLVideoElement | null>(null);

  const seek = useCallback((timeMs: number) => {
    const clampedTime = Math.max(0, Math.min(timeMs, state.durationMs));
    setState(s => ({ ...s, currentTimeMs: clampedTime }));
    
    if (videoRef.current) {
      videoRef.current.currentTime = clampedTime / 1000;
    }
  }, [state.durationMs]);

  const play = useCallback(() => {
    setState(s => ({ ...s, isPlaying: true }));
    videoRef.current?.play();
  }, []);

  const pause = useCallback(() => {
    setState(s => ({ ...s, isPlaying: false }));
    videoRef.current?.pause();
  }, []);

  const toggle = useCallback(() => {
    if (state.isPlaying) {
      pause();
    } else {
      play();
    }
  }, [state.isPlaying, play, pause]);

  const setPlaybackRate = useCallback((rate: number) => {
    setState(s => ({ ...s, playbackRate: rate }));
    if (videoRef.current) {
      videoRef.current.playbackRate = rate;
    }
  }, []);

  const setDuration = useCallback((durationMs: number) => {
    setState(s => ({ ...s, durationMs }));
  }, []);

  const registerVideoRef = useCallback((video: HTMLVideoElement | null) => {
    videoRef.current = video;
  }, []);

  return (
    <TimeContext.Provider value={{
      ...state,
      seek,
      play,
      pause,
      toggle,
      setPlaybackRate,
      setDuration,
      registerVideoRef,
    }}>
      {children}
    </TimeContext.Provider>
  );
}

export function useTime() {
  const context = useContext(TimeContext);
  if (!context) {
    throw new Error('useTime must be used within TimeProvider');
  }
  return context;
}
```

### Video Player Component

```tsx
// src/components/VideoPlayer/VideoPlayer.tsx
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
    setDuration,
    registerVideoRef,
  } = useTime();

  useEffect(() => {
    registerVideoRef(videoRef.current);
    return () => registerVideoRef(null);
  }, [registerVideoRef]);

  const handleLoadedMetadata = () => {
    if (videoRef.current) {
      setDuration(videoRef.current.duration * 1000);
    }
  };

  const handleTimeUpdate = () => {
    if (videoRef.current && !document.hidden) {
      const newTimeMs = videoRef.current.currentTime * 1000;
      // Only update if difference is significant (avoid feedback loops)
      if (Math.abs(newTimeMs - currentTimeMs) > 50) {
        seek(newTimeMs);
      }
    }
  };

  const handleScrub = (e: React.ChangeEvent<HTMLInputElement>) => {
    seek(Number(e.target.value));
  };

  return (
    <div className={styles.container}>
      <video
        ref={videoRef}
        src={src}
        poster={poster}
        className={styles.video}
        onLoadedMetadata={handleLoadedMetadata}
        onTimeUpdate={handleTimeUpdate}
        onClick={toggle}
      />
      
      <div className={styles.controls}>
        <button 
          className={styles.playButton} 
          onClick={toggle}
          aria-label={isPlaying ? 'Pause' : 'Play'}
        >
          {isPlaying ? '⏸' : '▶'}
        </button>
        
        <span className={styles.time}>
          {formatTime(currentTimeMs)} / {formatTime(durationMs)}
        </span>
        
        <input
          type="range"
          className={styles.scrubber}
          min={0}
          max={durationMs}
          value={currentTimeMs}
          onChange={handleScrub}
        />
        
        <select
          className={styles.rate}
          value={playbackRate}
          onChange={(e) => setPlaybackRate(Number(e.target.value))}
        >
          <option value={0.25}>0.25x</option>
          <option value={0.5}>0.5x</option>
          <option value={1}>1x</option>
          <option value={2}>2x</option>
        </select>
      </div>
    </div>
  );
}
```

### Input Timeline Component

```tsx
// src/components/InputTimeline/InputTimeline.tsx
import { useMemo } from 'react';
import { useTime } from '../../context/TimeContext';
import type { KeyboardEvent, MouseEvent, GamepadEvent } from '../../types/inputs';
import styles from './InputTimeline.module.css';

interface InputTimelineProps {
  keyboard: KeyboardEvent[];
  mouse: MouseEvent[];
  gamepad: GamepadEvent[];
}

export function InputTimeline({ keyboard, mouse, gamepad }: InputTimelineProps) {
  const { currentTimeMs, durationMs, seek } = useTime();

  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const ratio = (e.clientX - rect.left) / rect.width;
    seek(ratio * durationMs);
  };

  // Convert events to visual segments
  const keyboardSegments = useMemo(() => 
    computeKeyboardSegments(keyboard, durationMs),
    [keyboard, durationMs]
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
            {keyboardSegments.map((seg, i) => (
              <div
                key={i}
                className={styles.segment}
                style={{
                  left: `${(seg.startMs / durationMs) * 100}%`,
                  width: `${((seg.endMs - seg.startMs) / durationMs) * 100}%`,
                }}
                title={seg.keys.join(' + ')}
              />
            ))}
          </div>
        </div>
        
        {/* Mouse Track */}
        <div className={styles.track}>
          <span className={styles.label}>Mouse</span>
          <div className={styles.trackContent}>
            {mouse
              .filter(m => m.type === 'down')
              .map((m, i) => (
                <div
                  key={i}
                  className={styles.mouseClick}
                  style={{ left: `${(m.timestampMs / durationMs) * 100}%` }}
                  title={`Button ${m.button}`}
                />
              ))}
          </div>
        </div>
        
        {/* Gamepad Track */}
        <div className={styles.track}>
          <span className={styles.label}>Pad</span>
          <div className={styles.trackContent}>
            {gamepad.length === 0 && (
              <span className={styles.noData}>No gamepad input</span>
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

// Helper functions
function computeKeyboardSegments(
  events: KeyboardEvent[],
  durationMs: number
): Array<{ startMs: number; endMs: number; keys: string[] }> {
  const segments: Array<{ startMs: number; endMs: number; keys: string[] }> = [];
  const activeKeys = new Map<string, number>(); // key -> startTime
  
  for (const event of events) {
    if (event.type === 'down' && !activeKeys.has(event.key)) {
      activeKeys.set(event.key, event.timestampMs);
    } else if (event.type === 'up' && activeKeys.has(event.key)) {
      const startMs = activeKeys.get(event.key)!;
      segments.push({
        startMs,
        endMs: event.timestampMs,
        keys: [event.key],
      });
      activeKeys.delete(event.key);
    }
  }
  
  // Close any still-active keys at end of recording
  for (const [key, startMs] of activeKeys) {
    segments.push({ startMs, endMs: durationMs, keys: [key] });
  }
  
  return segments;
}

function getTickInterval(durationMs: number): number {
  if (durationMs <= 10000) return 1000;      // 1s ticks for < 10s
  if (durationMs <= 60000) return 5000;      // 5s ticks for < 1min
  if (durationMs <= 300000) return 30000;    // 30s ticks for < 5min
  return 60000;                               // 1min ticks otherwise
}
```

### Log Panel Component

```tsx
// src/components/LogPanel/LogPanel.tsx
import { useState, useMemo, useCallback } from 'react';
import { useTime } from '../../context/TimeContext';
import { formatTimeMs } from '../../utils/time';
import type { LogEntry } from '../../types/logs';
import styles from './LogPanel.module.css';

interface LogPanelProps {
  logs: LogEntry[];
  categories: string[];
}

type LogLevel = 'verbose' | 'log' | 'warning' | 'error';

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
    let closest = 0;
    let minDiff = Infinity;
    for (let i = 0; i < filteredLogs.length; i++) {
      const diff = Math.abs(filteredLogs[i].timestampMs - currentTimeMs);
      if (diff < minDiff) {
        minDiff = diff;
        closest = i;
      }
    }
    return closest;
  }, [filteredLogs, currentTimeMs]);

  const handleLogClick = useCallback((timestampMs: number) => {
    seek(timestampMs);
  }, [seek]);

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <span>LOGS</span>
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
          >
            <option value="all">All Categories</option>
            {categories.map(cat => (
              <option key={cat} value={cat}>{cat}</option>
            ))}
          </select>
        </div>
      </div>
      
      <div className={styles.logList}>
        {filteredLogs.map((log, index) => (
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
            <span className={styles.level}>
              {LEVEL_ICONS[log.level]}
            </span>
            <span className={styles.category}>[{log.category}]</span>
            <span className={styles.message}>{log.message}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
```

### Frame Timing Graph Component

```tsx
// src/components/FrameTimingGraph/FrameTimingGraph.tsx
import { useRef, useEffect, useMemo } from 'react';
import { useTime } from '../../context/TimeContext';
import type { FrameSample } from '../../types/frames';
import styles from './FrameTimingGraph.module.css';

interface FrameTimingGraphProps {
  samples: FrameSample[];
  summary: {
    avgFps: number;
    minFps: number;
    maxFps: number;
  };
}

const TARGET_FPS = 60;
const MIN_FPS = 0;
const MAX_FPS = 120;

export function FrameTimingGraph({ samples, summary }: FrameTimingGraphProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const { currentTimeMs, durationMs, seek } = useTime();

  // Downsample for performance if needed
  const displaySamples = useMemo(() => {
    if (samples.length <= 1000) return samples;
    const step = Math.ceil(samples.length / 1000);
    return samples.filter((_, i) => i % step === 0);
  }, [samples]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    const { width, height } = canvas;
    const padding = { top: 20, right: 50, bottom: 30, left: 10 };
    const graphWidth = width - padding.left - padding.right;
    const graphHeight = height - padding.top - padding.bottom;

    // Clear
    ctx.fillStyle = '#1a1a2e';
    ctx.fillRect(0, 0, width, height);

    // Draw reference lines
    ctx.strokeStyle = '#333';
    ctx.setLineDash([5, 5]);
    
    // 60 FPS line
    const y60 = padding.top + graphHeight * (1 - (60 - MIN_FPS) / (MAX_FPS - MIN_FPS));
    ctx.beginPath();
    ctx.moveTo(padding.left, y60);
    ctx.lineTo(width - padding.right, y60);
    ctx.stroke();
    ctx.fillStyle = '#666';
    ctx.fillText('60fps', width - padding.right + 5, y60 + 4);
    
    // 30 FPS line
    const y30 = padding.top + graphHeight * (1 - (30 - MIN_FPS) / (MAX_FPS - MIN_FPS));
    ctx.beginPath();
    ctx.moveTo(padding.left, y30);
    ctx.lineTo(width - padding.right, y30);
    ctx.stroke();
    ctx.fillText('30fps', width - padding.right + 5, y30 + 4);

    ctx.setLineDash([]);

    // Draw FPS line
    ctx.strokeStyle = '#4ade80';
    ctx.lineWidth = 1.5;
    ctx.beginPath();
    
    displaySamples.forEach((sample, i) => {
      const x = padding.left + (sample.timestampMs / durationMs) * graphWidth;
      const fps = Math.min(MAX_FPS, Math.max(MIN_FPS, sample.fps));
      const y = padding.top + graphHeight * (1 - (fps - MIN_FPS) / (MAX_FPS - MIN_FPS));
      
      if (i === 0) {
        ctx.moveTo(x, y);
      } else {
        ctx.lineTo(x, y);
      }
    });
    
    ctx.stroke();

    // Draw problem areas (FPS drops)
    ctx.fillStyle = 'rgba(239, 68, 68, 0.3)';
    displaySamples.forEach((sample) => {
      if (sample.fps < 30) {
        const x = padding.left + (sample.timestampMs / durationMs) * graphWidth;
        const barWidth = Math.max(2, graphWidth / displaySamples.length);
        ctx.fillRect(x - barWidth / 2, padding.top, barWidth, graphHeight);
      }
    });

    // Draw playhead
    const playheadX = padding.left + (currentTimeMs / durationMs) * graphWidth;
    ctx.strokeStyle = '#fff';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(playheadX, padding.top);
    ctx.lineTo(playheadX, height - padding.bottom);
    ctx.stroke();

  }, [displaySamples, currentTimeMs, durationMs]);

  const handleClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const rect = canvas.getBoundingClientRect();
    const padding = { left: 10, right: 50 };
    const graphWidth = canvas.width - padding.left - padding.right;
    
    const x = e.clientX - rect.left - padding.left;
    const ratio = Math.max(0, Math.min(1, x / graphWidth));
    seek(ratio * durationMs);
  };

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <span>FRAME TIMING</span>
        <div className={styles.summary}>
          <span>Avg: {summary.avgFps.toFixed(0)} FPS</span>
          <span>Min: {summary.minFps.toFixed(0)} FPS</span>
          <span>Max: {summary.maxFps.toFixed(0)} FPS</span>
        </div>
      </div>
      <canvas
        ref={canvasRef}
        width={800}
        height={150}
        className={styles.canvas}
        onClick={handleClick}
      />
    </div>
  );
}
```

### Repro Card Component

```tsx
// src/components/ReproCard/ReproCard.tsx
import { Link } from 'react-router-dom';
import { formatRelativeTime, formatDuration } from '../../utils/time';
import type { ReproSummary } from '../../types/repro';
import styles from './ReproCard.module.css';

interface ReproCardProps {
  repro: ReproSummary;
}

export function ReproCard({ repro }: ReproCardProps) {
  return (
    <Link to={`/repro/${repro.id}`} className={styles.card}>
      <div className={styles.thumbnail}>
        <img 
          src={repro.thumbnailUrl} 
          alt="" 
          loading="lazy"
        />
        <span className={styles.duration}>
          {formatDuration(repro.durationMs)}
        </span>
      </div>
      
      <div className={styles.content}>
        <h3 className={styles.title}>{repro.title}</h3>
        
        <div className={styles.meta}>
          <span className={styles.build}>{repro.build}</span>
          <span className={styles.separator}>|</span>
          <span className={styles.platform}>{repro.platform}</span>
          <span className={styles.separator}>|</span>
          <span className={styles.map}>{repro.map}</span>
        </div>
        
        <div className={styles.footer}>
          <span className={styles.date}>
            {formatRelativeTime(repro.createdAt)}
          </span>
          <div className={styles.tags}>
            {repro.tags.slice(0, 3).map(tag => (
              <span key={tag} className={styles.tag}>{tag}</span>
            ))}
            {repro.tags.length > 3 && (
              <span className={styles.moreTag}>+{repro.tags.length - 3}</span>
            )}
          </div>
        </div>
      </div>
    </Link>
  );
}
```

### Utility Functions

```typescript
// src/utils/time.ts

export function formatTime(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}

export function formatTimeMs(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  const millis = Math.floor(ms % 1000);
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}.${millis.toString().padStart(3, '0')}`;
}

export function formatDuration(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  if (totalSeconds < 60) {
    return `0:${totalSeconds.toString().padStart(2, '0')}`;
  }
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, '0')}`;
}

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
  
  return date.toLocaleDateString();
}
```

---

## 6. Project Structure

```
bugit/
├── public/
│   └── favicon.ico
├── src/
│   ├── api/
│   │   ├── client.ts              # Fetch wrapper
│   │   ├── repros.ts              # Repro API calls
│   │   ├── inputs.ts              # Input events API
│   │   ├── logs.ts                # Logs API
│   │   └── notes.ts               # Notes CRUD
│   ├── components/
│   │   ├── Layout/
│   │   │   ├── Header.tsx
│   │   │   ├── Layout.tsx
│   │   │   └── Layout.module.css
│   │   ├── VideoPlayer/
│   │   │   ├── VideoPlayer.tsx
│   │   │   └── VideoPlayer.module.css
│   │   ├── InputTimeline/
│   │   │   ├── InputTimeline.tsx
│   │   │   └── InputTimeline.module.css
│   │   ├── FrameTimingGraph/
│   │   │   ├── FrameTimingGraph.tsx
│   │   │   └── FrameTimingGraph.module.css
│   │   ├── LogPanel/
│   │   │   ├── LogPanel.tsx
│   │   │   └── LogPanel.module.css
│   │   ├── NotesPanel/
│   │   │   ├── NotesPanel.tsx
│   │   │   ├── NoteCard.tsx
│   │   │   └── NotesPanel.module.css
│   │   ├── ReproCard/
│   │   │   ├── ReproCard.tsx
│   │   │   └── ReproCard.module.css
│   │   ├── FilterBar/
│   │   │   ├── FilterBar.tsx
│   │   │   └── FilterBar.module.css
│   │   └── shared/
│   │       ├── Button.tsx
│   │       ├── Dropdown.tsx
│   │       ├── Badge.tsx
│   │       └── Tooltip.tsx
│   ├── context/
│   │   └── TimeContext.tsx
│   ├── hooks/
│   │   ├── useRepros.ts
│   │   ├── useRepro.ts
│   │   ├── useInputs.ts
│   │   ├── useLogs.ts
│   │   └── useNotes.ts
│   ├── pages/
│   │   ├── ReproListPage.tsx
│   │   └── ReproViewerPage.tsx
│   ├── types/
│   │   ├── repro.ts
│   │   ├── inputs.ts
│   │   ├── logs.ts
│   │   ├── frames.ts
│   │   └── notes.ts
│   ├── utils/
│   │   ├── time.ts
│   │   └── search.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── README.md
```

---

## 7. CSS Design Tokens

```css
/* src/index.css */
:root {
  /* Colors - Dark theme optimized for long viewing sessions */
  --bg-primary: #0f0f1a;
  --bg-secondary: #1a1a2e;
  --bg-tertiary: #252542;
  
  --text-primary: #e4e4e7;
  --text-secondary: #a1a1aa;
  --text-muted: #71717a;
  
  --accent-primary: #6366f1;    /* Indigo */
  --accent-success: #4ade80;    /* Green - good FPS */
  --accent-warning: #fbbf24;    /* Yellow - warnings */
  --accent-error: #ef4444;      /* Red - errors */
  
  --border-color: #27273f;
  
  /* Typography */
  --font-mono: 'JetBrains Mono', 'Fira Code', monospace;
  --font-sans: 'Inter', -apple-system, sans-serif;
  
  /* Spacing */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  
  /* Borders */
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: var(--font-sans);
  background: var(--bg-primary);
  color: var(--text-primary);
  line-height: 1.5;
}

code, .mono {
  font-family: var(--font-mono);
}
```

---

## 8. Next Steps

1. **Initialize project**: `npm create vite@latest bugit -- --template react-ts`
2. **Install dependencies**: `react-router-dom`, `@tanstack/react-query`, `react-markdown`
3. **Build mock backend**: JSON files or simple Express server for development
4. **Implement TimeContext** first - it's the foundation
5. **Build VideoPlayer + InputTimeline** as first milestone
6. **Add FrameTimingGraph and LogPanel** with sync
7. **Build ReproList page**
8. **Add NotesPanel with CRUD**

---

## Appendix: Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Space` | Play/Pause |
| `←` / `→` | Seek ±5s |
| `Shift+←/→` | Seek ±1s |
| `J` / `L` | Seek ±10s |
| `K` | Pause |
| `,` / `.` | Previous/Next frame |
| `M` | Mute |
| `F` | Fullscreen |
| `Esc` | Exit fullscreen / Close modal |
| `N` | Add note at current time |
