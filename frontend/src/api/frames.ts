import type { GetFramesResponse } from '../types';
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
  
  // Transform backend format to frontend format
  if (data.frames) {
    const targetFps = data.targetFps || 30;
    
    const samples = data.frames.map((frame: { 
      timestampMs: number; 
      deltaTimeSeconds: number;
      videoFrameIndex: number;
    }) => {
      const frameTimeMs = frame.deltaTimeSeconds * 1000;
      const fps = frameTimeMs > 0 ? 1000 / frameTimeMs : targetFps;
      return {
        timestampMs: frame.timestampMs,
        frameTimeMs,
        fps,
      };
    });
    
    const fpsList = samples.map((s: { fps: number }) => s.fps);
    const summary = {
      avgFps: fpsList.length > 0 ? fpsList.reduce((a: number, b: number) => a + b, 0) / fpsList.length : 0,
      minFps: fpsList.length > 0 ? Math.min(...fpsList) : 0,
      maxFps: fpsList.length > 0 ? Math.max(...fpsList) : 0,
      p99FrameTimeMs: fpsList.length > 0 ? 1000 / Math.min(...fpsList) : 0,
      stutterCount: samples.filter((s: { fps: number }) => s.fps < 30).length,
    };
    
    return { samples, summary };
  }
  
  // Already in frontend format
  return data;
}
