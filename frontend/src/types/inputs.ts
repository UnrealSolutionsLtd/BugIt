export interface KeyboardEvent {
  timestampMs: number;
  type: 'down' | 'up';
  key: string;
  keyCode: number;
}

export interface MouseEvent {
  timestampMs: number;
  type: 'down' | 'up' | 'move' | 'wheel';
  button?: number;
  x: number;
  y: number;
  deltaX?: number;
  deltaY?: number;
}

export interface GamepadEvent {
  timestampMs: number;
  type: 'button' | 'axis';
  index: number;
  value: number;
}

export interface GetInputsResponse {
  keyboard: KeyboardEvent[];
  mouse: MouseEvent[];
  gamepad: GamepadEvent[];
}

export interface KeyboardSegment {
  startMs: number;
  endMs: number;
  keys: string[];
}
