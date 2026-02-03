import type { GetFramesResponse, FrameSample } from '../types';
import { getBundle, getArtifactUrl } from './repros';

// Get frame timing data for a bundle by fetching the timing.json artifact
export async function getFrames(bundleId: string): Promise<GetFramesResponse> {
  const bundle = await getBundle(bundleId);
  
  // Find timing artifact
  const timingArtifact = bundle.artifacts?.find(
    a => a.filename === 'timing.json'
  );
  
  if (!timingArtifact) {
    return { 
      samples: [], 
      summary: { avgFps: 0, minFps: 0, maxFps: 0, p99FrameTimeMs: 0, stutterCount: 0 } 
    };
  }
  
  // Fetch the artifact content
  const url = getArtifactUrl(bundleId, timingArtifact.artifact_id);
  const response = await fetch(url);
  
  if (!response.ok) {
    throw new Error(`Failed to fetch frames: ${response.statusText}`);
  }
  
  const data = await response.json();
  
  let samples: FrameSample[] = [];
  
  // Transform backend format to frontend format
  if (data.frames && data.frames.length > 0) {
    const targetFps = data.targetFps || 30;
    const targetFrameTimeMs = 1000 / targetFps;
    
    // Calculate frame timing from timestamps (deltaTimeSeconds was removed from plugin output)
    samples = data.frames.map((frame: { 
      timestampMs: number; 
      videoFrameIndex: number;
      deltaTimeSeconds?: number; // Optional for backwards compatibility
    }, index: number, arr: typeof data.frames) => {
      // Use deltaTimeSeconds if present (legacy), otherwise derive from timestamps
      let frameTimeMs: number;
      if (frame.deltaTimeSeconds !== undefined) {
        frameTimeMs = frame.deltaTimeSeconds * 1000;
      } else if (index > 0) {
        // Calculate delta from previous frame's timestamp
        frameTimeMs = frame.timestampMs - arr[index - 1].timestampMs;
      } else {
        // First frame: use target frame time
        frameTimeMs = targetFrameTimeMs;
      }
      
      const fps = frameTimeMs > 0 ? 1000 / frameTimeMs : targetFps;
      return {
        timestampMs: frame.timestampMs,
        frameTimeMs,
        fps,
      };
    });
  } else if (data.samples) {
    samples = data.samples;
  }
  
  // Timestamps are pre-normalized by the Unreal plugin (relative to video start)
  
  // Calculate summary statistics
  const fpsList = samples.map(s => s.fps);
  const summary = {
    avgFps: fpsList.length > 0 ? fpsList.reduce((a, b) => a + b, 0) / fpsList.length : 0,
    minFps: fpsList.length > 0 ? Math.min(...fpsList) : 0,
    maxFps: fpsList.length > 0 ? Math.max(...fpsList) : 0,
    p99FrameTimeMs: fpsList.length > 0 ? 1000 / Math.min(...fpsList) : 0,
    stutterCount: samples.filter(s => s.fps < 30).length,
  };
  
  return { samples, summary };
}
