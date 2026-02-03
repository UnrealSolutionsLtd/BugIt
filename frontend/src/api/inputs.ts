import type { GetInputsResponse, KeyboardEvent, MouseEvent, GamepadEvent } from '../types';
import { getBundle, getArtifactUrl } from './repros';

// Get inputs for a bundle by fetching the inputs.json artifact
export async function getInputs(bundleId: string): Promise<GetInputsResponse> {
  const bundle = await getBundle(bundleId);
  
  // Find inputs artifact
  const inputsArtifact = bundle.artifacts?.find(
    a => a.filename === 'inputs.json' || a.type === 'other' && a.filename.includes('input')
  );
  
  if (!inputsArtifact) {
    return { keyboard: [], mouse: [], gamepad: [] };
  }
  
  // Fetch the artifact content
  const url = getArtifactUrl(bundleId, inputsArtifact.artifact_id);
  const response = await fetch(url);
  
  if (!response.ok) {
    throw new Error(`Failed to fetch inputs: ${response.statusText}`);
  }
  
  const data = await response.json();
  
  let keyboard: KeyboardEvent[] = [];
  let mouse: MouseEvent[] = [];
  let gamepad: GamepadEvent[] = [];
  
  // Transform backend format to frontend format if needed
  if (data.events) {
    // Backend format with events array
    keyboard = data.events
      .filter((e: { inputType: string }) => e.inputType === 'KeyDown' || e.inputType === 'KeyUp')
      .map((e: { timestampMs: number; inputType: string; keyName: string; keyCode: number }) => ({
        timestampMs: e.timestampMs,
        type: e.inputType === 'KeyDown' ? 'down' : 'up',
        key: e.keyName,
        keyCode: e.keyCode,
      }));
    
    mouse = data.events
      .filter((e: { inputType: string }) => e.inputType.startsWith('Mouse'))
      .map((e: { timestampMs: number; inputType: string; keyName?: string; screenPosition?: number[] }) => ({
        timestampMs: e.timestampMs,
        type: e.inputType === 'MouseButtonDown' ? 'down' : e.inputType === 'MouseButtonUp' ? 'up' : 'move',
        button: e.keyName === 'LeftMouseButton' ? 0 : e.keyName === 'RightMouseButton' ? 2 : 1,
        x: e.screenPosition?.[0] || 0,
        y: e.screenPosition?.[1] || 0,
      }));
    
    gamepad = data.events
      .filter((e: { inputType: string }) => e.inputType.startsWith('Gamepad'))
      .map((e: { timestampMs: number; inputType: string; keyName: string; axisValue?: number }) => ({
        timestampMs: e.timestampMs,
        type: e.inputType,
        button: e.keyName,
        value: e.axisValue,
      }));
  } else {
    // Already in frontend format
    keyboard = data.keyboard || [];
    mouse = data.mouse || [];
    gamepad = data.gamepad || [];
  }
  
  // Timestamps are pre-normalized by the Unreal plugin (relative to video start)
  return { keyboard, mouse, gamepad };
}
