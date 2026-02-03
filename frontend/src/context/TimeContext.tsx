import { 
  createContext, 
  useContext, 
  useState, 
  useCallback, 
  useRef, 
  useEffect,
  type ReactNode 
} from 'react';

interface TimeState {
  currentTimeMs: number;
  durationMs: number;
  isPlaying: boolean;
  playbackRate: number;
}

interface TimeContextValue extends TimeState {
  seek: (timeMs: number) => void;
  seekRelative: (deltaMs: number) => void;
  play: () => void;
  pause: () => void;
  toggle: () => void;
  setPlaybackRate: (rate: number) => void;
  setDuration: (durationMs: number) => void;
  registerVideoRef: (video: HTMLVideoElement | null) => (() => void) | undefined;
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
  const lastSyncTime = useRef<number>(0);
  const lastUpdateTime = useRef<number>(0);

  const seek = useCallback((timeMs: number) => {
    setState(s => {
      const clampedTime = Math.max(0, Math.min(timeMs, s.durationMs));
      return { ...s, currentTimeMs: clampedTime };
    });
    
    if (videoRef.current) {
      videoRef.current.currentTime = timeMs / 1000;
    }
    lastSyncTime.current = Date.now();
  }, []);

  const seekRelative = useCallback((deltaMs: number) => {
    setState(s => {
      const newTime = Math.max(0, Math.min(s.currentTimeMs + deltaMs, s.durationMs));
      if (videoRef.current) {
        videoRef.current.currentTime = newTime / 1000;
      }
      return { ...s, currentTimeMs: newTime };
    });
  }, []);

  const play = useCallback(() => {
    setState(s => ({ ...s, isPlaying: true }));
    videoRef.current?.play();
  }, []);

  const pause = useCallback(() => {
    setState(s => ({ ...s, isPlaying: false }));
    videoRef.current?.pause();
  }, []);

  const toggle = useCallback(() => {
    setState(s => {
      if (s.isPlaying) {
        videoRef.current?.pause();
      } else {
        videoRef.current?.play();
      }
      return { ...s, isPlaying: !s.isPlaying };
    });
  }, []);

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
    
    if (video) {
      // Sync initial state
      video.playbackRate = state.playbackRate;
      
      // Handle video time updates
      const handleTimeUpdate = () => {
        const now = Date.now();
        
        // Skip if we just programmatically seeked (avoid feedback loop)
        if (now - lastSyncTime.current < 100) return;
        
        // Throttle to ~15fps max to reduce re-renders (every 66ms)
        if (now - lastUpdateTime.current < 66) return;
        lastUpdateTime.current = now;
        
        const newTimeMs = video.currentTime * 1000;
        setState(s => ({ ...s, currentTimeMs: newTimeMs }));
      };
      
      const handlePlay = () => setState(s => ({ ...s, isPlaying: true }));
      const handlePause = () => setState(s => ({ ...s, isPlaying: false }));
      const handleLoadedMetadata = () => {
        // Only set duration from video if not already set externally
        setState(s => {
          if (s.durationMs > 0) {
            // Duration already set from data - don't override
            return s;
          }
          return { ...s, durationMs: video.duration * 1000 };
        });
      };
      
      video.addEventListener('timeupdate', handleTimeUpdate);
      video.addEventListener('play', handlePlay);
      video.addEventListener('pause', handlePause);
      video.addEventListener('loadedmetadata', handleLoadedMetadata);
      
      return () => {
        video.removeEventListener('timeupdate', handleTimeUpdate);
        video.removeEventListener('play', handlePlay);
        video.removeEventListener('pause', handlePause);
        video.removeEventListener('loadedmetadata', handleLoadedMetadata);
      };
    }
  }, [state.playbackRate]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if typing in input
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }
      
      switch (e.key) {
        case ' ':
        case 'k':
          e.preventDefault();
          toggle();
          break;
        case 'ArrowLeft':
          e.preventDefault();
          seekRelative(e.shiftKey ? -1000 : -5000);
          break;
        case 'ArrowRight':
          e.preventDefault();
          seekRelative(e.shiftKey ? 1000 : 5000);
          break;
        case 'j':
          seekRelative(-10000);
          break;
        case 'l':
          seekRelative(10000);
          break;
        case ',':
          // Previous frame (~33ms for 30fps)
          seekRelative(-33);
          break;
        case '.':
          // Next frame
          seekRelative(33);
          break;
      }
    };
    
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [toggle, seekRelative]);

  return (
    <TimeContext.Provider value={{
      ...state,
      seek,
      seekRelative,
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
