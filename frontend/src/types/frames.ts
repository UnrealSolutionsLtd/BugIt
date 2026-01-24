export interface FrameSample {
  timestampMs: number;
  frameTimeMs: number;
  fps: number;
  gameThreadMs?: number;
  renderThreadMs?: number;
  gpuTimeMs?: number;
}

export interface FrameSummary {
  avgFps: number;
  minFps: number;
  maxFps: number;
  p99FrameTimeMs: number;
  stutterCount: number;
}

export interface GetFramesResponse {
  samples: FrameSample[];
  summary: FrameSummary;
}
