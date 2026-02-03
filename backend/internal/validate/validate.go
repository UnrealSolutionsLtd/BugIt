// Package validate provides bundle consistency validation.
package validate

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ValidationResult contains the outcome of bundle validation.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
	Stats    *BundleStats      `json:"stats,omitempty"`
}

// ValidationError describes a single validation issue.
type ValidationError struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Got     string `json:"got,omitempty"`
	Want    string `json:"want,omitempty"`
}

// BundleStats contains computed statistics about the bundle.
type BundleStats struct {
	// From manifest
	ManifestDurationSec  float64 `json:"manifestDurationSec"`
	ManifestTotalFrames  int     `json:"manifestTotalFrames"`
	
	// From timing.json (frame contexts - now 1:1 with video frames)
	TimingFrameCount     int     `json:"timingFrameCount"`
	TimingFirstTimestamp float64 `json:"timingFirstTimestampMs"`
	TimingLastTimestamp  float64 `json:"timingLastTimestampMs"`
	TimingDurationMs     float64 `json:"timingDurationMs"`
	
	// From inputs.json
	InputEventCount      int     `json:"inputEventCount"`
	KeyboardEventCount   int     `json:"keyboardEventCount"`
	MouseEventCount      int     `json:"mouseEventCount"`
	InputFirstTimestamp  float64 `json:"inputFirstTimestampMs"`
	InputLastTimestamp   float64 `json:"inputLastTimestampMs"`
	
	// Computed
	VideoFPS             float64 `json:"videoFps"`             // Video frame rate (manifest frames / duration)
	DurationMismatchMs   float64 `json:"durationMismatchMs"`
}

// Manifest represents the manifest.json structure for validation.
type Manifest struct {
	SchemaVersion    string          `json:"schemaVersion"`
	BundleID         string          `json:"bundleId"`
	DurationSeconds  float64         `json:"durationSeconds"`
	TotalFrames      int             `json:"totalFrames"`
	SessionInfo      *SessionInfo    `json:"sessionInfo"`
	Artifacts        json.RawMessage `json:"artifacts"`
}

type SessionInfo struct {
	MapName   string `json:"mapName"`
	TargetFPS int    `json:"targetFps"`
}

// TimingData represents timing.json structure.
type TimingData struct {
	SchemaVersion string       `json:"schemaVersion"`
	Frames        []FrameEntry `json:"frames"`
}

type FrameEntry struct {
	VideoFrameIndex int     `json:"videoFrameIndex"`
	TimestampMs     float64 `json:"timestampMs"`
	IsPaused        bool    `json:"isPaused,omitempty"`
}

// InputData represents inputs.json structure.
type InputData struct {
	SchemaVersion string       `json:"schemaVersion"`
	TotalEvents   int          `json:"totalEvents"`
	Events        []InputEvent `json:"events"`
}

type InputEvent struct {
	TimestampMs float64 `json:"timestampMs"`
	InputType   string  `json:"inputType"`
	KeyName     string  `json:"keyName,omitempty"`
	KeyCode     int     `json:"keyCode,omitempty"`
}

// ValidateBundle validates a bundle directory for internal consistency.
func ValidateBundle(bundlePath string) *ValidationResult {
	result := &ValidationResult{
		Valid: true,
		Stats: &BundleStats{},
	}
	
	// Load manifest
	manifest, err := loadManifest(bundlePath)
	if err != nil {
		result.addError("MANIFEST_LOAD", "", err.Error())
		result.Valid = false
		return result
	}
	
	// Load timing.json
	timing, err := loadTiming(bundlePath)
	if err != nil {
		result.addWarning("TIMING_LOAD", "", err.Error())
	}
	
	// Load inputs.json
	inputs, err := loadInputs(bundlePath)
	if err != nil {
		result.addWarning("INPUTS_LOAD", "", err.Error())
	}
	
	// Populate stats from manifest
	result.Stats.ManifestDurationSec = manifest.DurationSeconds
	result.Stats.ManifestTotalFrames = manifest.TotalFrames
	
	// Validate manifest internal consistency
	validateManifestInternal(result, manifest)
	
	// Validate timing.json if available
	if timing != nil {
		validateTiming(result, timing)
		validateManifestVsTiming(result, manifest, timing)
	}
	
	// Validate inputs.json if available
	if inputs != nil {
		validateInputs(result, inputs, manifest)
	}
	
	return result
}

func (r *ValidationResult) addError(code, field, message string) {
	r.Errors = append(r.Errors, ValidationError{
		Code:    code,
		Field:   field,
		Message: message,
	})
	r.Valid = false
}

func (r *ValidationResult) addErrorWithValues(code, field, message, got, want string) {
	r.Errors = append(r.Errors, ValidationError{
		Code:    code,
		Field:   field,
		Message: message,
		Got:     got,
		Want:    want,
	})
	r.Valid = false
}

func (r *ValidationResult) addWarning(code, field, message string) {
	r.Warnings = append(r.Warnings, ValidationError{
		Code:    code,
		Field:   field,
		Message: message,
	})
}

func (r *ValidationResult) addWarningWithValues(code, field, message, got, want string) {
	r.Warnings = append(r.Warnings, ValidationError{
		Code:    code,
		Field:   field,
		Message: message,
		Got:     got,
		Want:    want,
	})
}

func loadManifest(bundlePath string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(bundlePath, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("read manifest.json: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest.json: %w", err)
	}
	return &m, nil
}

func loadTiming(bundlePath string) (*TimingData, error) {
	data, err := os.ReadFile(filepath.Join(bundlePath, "timing.json"))
	if err != nil {
		return nil, fmt.Errorf("read timing.json: %w", err)
	}
	var t TimingData
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parse timing.json: %w", err)
	}
	return &t, nil
}

func loadInputs(bundlePath string) (*InputData, error) {
	data, err := os.ReadFile(filepath.Join(bundlePath, "inputs.json"))
	if err != nil {
		return nil, fmt.Errorf("read inputs.json: %w", err)
	}
	var i InputData
	if err := json.Unmarshal(data, &i); err != nil {
		return nil, fmt.Errorf("parse inputs.json: %w", err)
	}
	return &i, nil
}

func validateManifestInternal(r *ValidationResult, m *Manifest) {
	// Duration must be positive
	if m.DurationSeconds <= 0 {
		r.addError("MANIFEST_DURATION", "durationSeconds", 
			fmt.Sprintf("must be positive, got %.3f", m.DurationSeconds))
	}
	
	// TotalFrames must be positive
	if m.TotalFrames <= 0 {
		r.addError("MANIFEST_FRAMES", "totalFrames", 
			fmt.Sprintf("must be positive, got %d", m.TotalFrames))
	}
	
	// Check reasonable FPS (1-240)
	if m.DurationSeconds > 0 && m.TotalFrames > 0 {
		fps := float64(m.TotalFrames) / m.DurationSeconds
		if fps < 1 || fps > 240 {
			r.addWarning("MANIFEST_FPS", "", 
				fmt.Sprintf("unusual FPS: %.1f (frames=%d, duration=%.2fs)", fps, m.TotalFrames, m.DurationSeconds))
		}
	}
	
}

func validateTiming(r *ValidationResult, t *TimingData) {
	if len(t.Frames) == 0 {
		r.addError("TIMING_EMPTY", "frames", "no frames in timing.json")
		return
	}
	
	r.Stats.TimingFrameCount = len(t.Frames)
	r.Stats.TimingFirstTimestamp = t.Frames[0].TimestampMs
	r.Stats.TimingLastTimestamp = t.Frames[len(t.Frames)-1].TimestampMs
	r.Stats.TimingDurationMs = r.Stats.TimingLastTimestamp - r.Stats.TimingFirstTimestamp
	
	// Check videoFrameIndex is sequential starting from 0
	for i, frame := range t.Frames {
		if frame.VideoFrameIndex != i {
			r.addErrorWithValues("TIMING_INDEX_MISMATCH", "videoFrameIndex",
				"videoFrameIndex should be sequential",
				fmt.Sprintf("frame[%d].videoFrameIndex = %d", i, frame.VideoFrameIndex),
				fmt.Sprintf("%d", i))
			break // Only report first mismatch
		}
	}
	
	// First timestamp should be near 0 (normalized)
	if t.Frames[0].TimestampMs > 100 {
		r.addWarning("TIMING_NOT_NORMALIZED", "timestampMs",
			fmt.Sprintf("first frame timestamp is %.1fms (expected near 0 if normalized)", t.Frames[0].TimestampMs))
	}
	
	// Check timestamps are monotonically increasing
	for i := 1; i < len(t.Frames); i++ {
		if t.Frames[i].TimestampMs < t.Frames[i-1].TimestampMs {
			r.addError("TIMING_NON_MONOTONIC", "timestampMs",
				fmt.Sprintf("timestamp decreased at frame %d: %.1fms -> %.1fms", 
					i, t.Frames[i-1].TimestampMs, t.Frames[i].TimestampMs))
			break
		}
	}
}

func validateManifestVsTiming(r *ValidationResult, m *Manifest, t *TimingData) {
	// Frame contexts should now be 1:1 with video frames
	// Both should have the same count
	
	if m.TotalFrames != len(t.Frames) {
		r.addErrorWithValues("FRAME_COUNT_MISMATCH", "totalFrames",
			"manifest totalFrames should match timing.json frame count (1:1 with video)",
			fmt.Sprintf("%d", m.TotalFrames),
			fmt.Sprintf("%d", len(t.Frames)))
	}
	
	// Calculate video FPS
	if m.DurationSeconds > 0 {
		videoFPS := float64(m.TotalFrames) / m.DurationSeconds
		r.Stats.VideoFPS = videoFPS
		
		// Video FPS should be reasonable (15-60)
		if videoFPS < 10 || videoFPS > 120 {
			r.addWarning("VIDEO_FPS_UNUSUAL", "",
				fmt.Sprintf("video FPS (%.1f) seems unusual", videoFPS))
		}
	}
	
	// Duration should approximately match
	timingDurationSec := r.Stats.TimingDurationMs / 1000.0
	durationDiff := math.Abs(m.DurationSeconds - timingDurationSec)
	r.Stats.DurationMismatchMs = durationDiff * 1000
	
	// Allow 100ms tolerance
	if durationDiff > 0.1 {
		r.addWarningWithValues("DURATION_MISMATCH", "durationSeconds",
			fmt.Sprintf("manifest duration differs from timing.json by %.1fms", durationDiff*1000),
			fmt.Sprintf("%.3fs", m.DurationSeconds),
			fmt.Sprintf("%.3fs", timingDurationSec))
	}
}

func validateInputs(r *ValidationResult, inputs *InputData, m *Manifest) {
	r.Stats.InputEventCount = len(inputs.Events)
	
	if len(inputs.Events) == 0 {
		r.addWarning("INPUTS_EMPTY", "", "no input events recorded")
		return
	}
	
	// Count by type and find timestamp range
	keyboardCount := 0
	mouseCount := 0
	minTs := math.MaxFloat64
	maxTs := 0.0
	
	keyDowns := make(map[string]float64) // key -> timestamp of last KeyDown
	
	for _, event := range inputs.Events {
		if event.TimestampMs < minTs {
			minTs = event.TimestampMs
		}
		if event.TimestampMs > maxTs {
			maxTs = event.TimestampMs
		}
		
		switch {
		case strings.HasPrefix(event.InputType, "Key"):
			keyboardCount++
			if event.InputType == "KeyDown" {
				keyDowns[event.KeyName] = event.TimestampMs
			} else if event.InputType == "KeyUp" {
				delete(keyDowns, event.KeyName)
			}
		case strings.HasPrefix(event.InputType, "Mouse"):
			mouseCount++
		}
	}
	
	r.Stats.KeyboardEventCount = keyboardCount
	r.Stats.MouseEventCount = mouseCount
	r.Stats.InputFirstTimestamp = minTs
	r.Stats.InputLastTimestamp = maxTs
	
	// Check all timestamps are within video duration
	videoDurationMs := m.DurationSeconds * 1000
	outOfRange := 0
	for _, event := range inputs.Events {
		if event.TimestampMs < 0 || event.TimestampMs > videoDurationMs+100 { // 100ms tolerance
			outOfRange++
		}
	}
	if outOfRange > 0 {
		r.addWarning("INPUTS_OUT_OF_RANGE", "timestampMs",
			fmt.Sprintf("%d input events outside video duration [0, %.1fms]", outOfRange, videoDurationMs))
	}
	
	// Report unmatched KeyDowns (missing KeyUp)
	if len(keyDowns) > 0 {
		keys := make([]string, 0, len(keyDowns))
		for k := range keyDowns {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		r.addWarning("INPUTS_UNMATCHED_KEYDOWN", "",
			fmt.Sprintf("KeyDown without KeyUp: %v", keys))
	}
}

// BundleSummary contains a human-readable summary of bundle contents.
type BundleSummary struct {
	BundleID    string          `json:"bundleId"`
	MapName     string          `json:"mapName"`
	Duration    float64         `json:"durationSeconds"`
	VideoFrames int             `json:"videoFrames"`
	VideoFPS    float64         `json:"videoFps"`
	KeyPresses  []KeyPressSummary `json:"keyPresses,omitempty"`
	MouseClicks []MouseClickSummary `json:"mouseClicks,omitempty"`
}

type KeyPressSummary struct {
	Key       string  `json:"key"`
	StartMs   float64 `json:"startMs"`
	EndMs     float64 `json:"endMs,omitempty"`
	DurationMs float64 `json:"durationMs,omitempty"`
}

type MouseClickSummary struct {
	Button    string  `json:"button"`
	TimestampMs float64 `json:"timestampMs"`
	Position  [2]float64 `json:"position,omitempty"`
}

// SummarizeBundle returns a human-readable summary of bundle contents.
func SummarizeBundle(bundlePath string) (*BundleSummary, error) {
	summary := &BundleSummary{}
	
	// Load manifest
	manifest, err := loadManifest(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}
	
	summary.BundleID = manifest.BundleID
	summary.Duration = manifest.DurationSeconds
	summary.VideoFrames = manifest.TotalFrames
	if manifest.DurationSeconds > 0 {
		summary.VideoFPS = float64(manifest.TotalFrames) / manifest.DurationSeconds
	}
	if manifest.SessionInfo != nil {
		summary.MapName = manifest.SessionInfo.MapName
	}
	
	// Load inputs
	inputs, err := loadInputs(bundlePath)
	if err == nil && inputs != nil {
		// Track key down events to pair with key up
		keyDownTimes := make(map[string]float64)
		
		for _, event := range inputs.Events {
			switch event.InputType {
			case "KeyDown":
				keyDownTimes[event.KeyName] = event.TimestampMs
			case "KeyUp":
				if startMs, ok := keyDownTimes[event.KeyName]; ok {
					// Normal case: KeyDown was seen
					summary.KeyPresses = append(summary.KeyPresses, KeyPressSummary{
						Key:        event.KeyName,
						StartMs:    startMs,
						EndMs:      event.TimestampMs,
						DurationMs: event.TimestampMs - startMs,
					})
					delete(keyDownTimes, event.KeyName)
				} else {
					// Orphaned KeyUp: key was held before capture started
					// Mark with StartMs = -1 to indicate "held from start"
					summary.KeyPresses = append(summary.KeyPresses, KeyPressSummary{
						Key:        event.KeyName,
						StartMs:    -1, // Indicates "held from start"
						EndMs:      event.TimestampMs,
						DurationMs: event.TimestampMs, // Duration from start of capture
					})
				}
			case "MouseButtonDown":
				summary.MouseClicks = append(summary.MouseClicks, MouseClickSummary{
					Button:      event.KeyName,
					TimestampMs: event.TimestampMs,
				})
			}
		}
		
		// Add any unpaired key downs (held at end of recording)
		for key, startMs := range keyDownTimes {
			summary.KeyPresses = append(summary.KeyPresses, KeyPressSummary{
				Key:     key,
				StartMs: startMs,
				// No EndMs - key was held at end
			})
		}
		
		// Sort key presses by start time (orphaned keys with -1 go first)
		sort.Slice(summary.KeyPresses, func(i, j int) bool {
			// Treat -1 (held from start) as 0 for sorting
			si := summary.KeyPresses[i].StartMs
			sj := summary.KeyPresses[j].StartMs
			if si < 0 { si = 0 }
			if sj < 0 { sj = 0 }
			return si < sj
		})
	}
	
	return summary, nil
}

// FormatSummary returns a human-readable text summary.
func FormatSummary(s *BundleSummary) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("=== Bundle Summary: %s ===\n\n", s.BundleID))
	
	if s.MapName != "" {
		sb.WriteString(fmt.Sprintf("Map: %s\n", s.MapName))
	}
	sb.WriteString(fmt.Sprintf("Duration: %.2fs (%.1fms)\n", s.Duration, s.Duration*1000))
	sb.WriteString(fmt.Sprintf("Video: %d frames @ %.1f FPS\n", s.VideoFrames, s.VideoFPS))
	sb.WriteString(fmt.Sprintf("Timeline: 0ms → %.0fms\n", s.Duration*1000))
	
	if len(s.KeyPresses) > 0 {
		sb.WriteString(fmt.Sprintf("\n=== Key Presses (%d) ===\n", len(s.KeyPresses)))
		for _, kp := range s.KeyPresses {
			if kp.StartMs < 0 {
				// Orphaned KeyUp - key was held before capture started
				sb.WriteString(fmt.Sprintf("  [     0ms - %6.0fms] %s (held from start)\n", 
					kp.EndMs, kp.Key))
			} else if kp.EndMs > 0 {
				// Normal key press with both KeyDown and KeyUp
				sb.WriteString(fmt.Sprintf("  [%6.0fms - %6.0fms] %s (held %.0fms)\n", 
					kp.StartMs, kp.EndMs, kp.Key, kp.DurationMs))
			} else {
				// Orphaned KeyDown - key was held at end of capture
				sb.WriteString(fmt.Sprintf("  [%6.0fms - ???    ] %s (held at end)\n", 
					kp.StartMs, kp.Key))
			}
		}
	} else {
		sb.WriteString("\nNo keyboard input recorded.\n")
	}
	
	if len(s.MouseClicks) > 0 {
		sb.WriteString(fmt.Sprintf("\n=== Mouse Clicks (%d) ===\n", len(s.MouseClicks)))
		for _, mc := range s.MouseClicks {
			sb.WriteString(fmt.Sprintf("  [%6.0fms] %s\n", mc.TimestampMs, mc.Button))
		}
	}
	
	return sb.String()
}

// FormatResult returns a human-readable validation report.
func FormatResult(r *ValidationResult) string {
	var sb strings.Builder
	
	if r.Valid {
		sb.WriteString("✓ Bundle is VALID\n\n")
	} else {
		sb.WriteString("✗ Bundle is INVALID\n\n")
	}
	
	// Stats
	sb.WriteString("=== Bundle Statistics ===\n")
	if r.Stats != nil {
		sb.WriteString(fmt.Sprintf("Manifest:\n"))
		sb.WriteString(fmt.Sprintf("  Duration: %.3fs\n", r.Stats.ManifestDurationSec))
		sb.WriteString(fmt.Sprintf("  Video frames: %d\n", r.Stats.ManifestTotalFrames))
		sb.WriteString(fmt.Sprintf("  Video FPS: %.1f\n", r.Stats.VideoFPS))
		
		if r.Stats.TimingFrameCount > 0 {
			sb.WriteString(fmt.Sprintf("\nTiming.json (1:1 with video frames):\n"))
			sb.WriteString(fmt.Sprintf("  Frame contexts: %d\n", r.Stats.TimingFrameCount))
			sb.WriteString(fmt.Sprintf("  Timestamp range: %.1fms - %.1fms\n", 
				r.Stats.TimingFirstTimestamp, r.Stats.TimingLastTimestamp))
		}
		
		if r.Stats.InputEventCount > 0 {
			sb.WriteString(fmt.Sprintf("\nInputs.json:\n"))
			sb.WriteString(fmt.Sprintf("  Total events: %d\n", r.Stats.InputEventCount))
			sb.WriteString(fmt.Sprintf("  Keyboard: %d, Mouse: %d\n", r.Stats.KeyboardEventCount, r.Stats.MouseEventCount))
			sb.WriteString(fmt.Sprintf("  Timestamp range: %.1fms - %.1fms\n", r.Stats.InputFirstTimestamp, r.Stats.InputLastTimestamp))
		}
		
		if r.Stats.DurationMismatchMs > 0 {
			sb.WriteString(fmt.Sprintf("\nDuration mismatch: %.1fms\n", r.Stats.DurationMismatchMs))
		}
	}
	
	// Errors
	if len(r.Errors) > 0 {
		sb.WriteString("\n=== ERRORS ===\n")
		for _, e := range r.Errors {
			sb.WriteString(fmt.Sprintf("  [%s] %s\n", e.Code, e.Message))
			if e.Got != "" || e.Want != "" {
				sb.WriteString(fmt.Sprintf("    Got:  %s\n", e.Got))
				sb.WriteString(fmt.Sprintf("    Want: %s\n", e.Want))
			}
		}
	}
	
	// Warnings
	if len(r.Warnings) > 0 {
		sb.WriteString("\n=== WARNINGS ===\n")
		for _, w := range r.Warnings {
			sb.WriteString(fmt.Sprintf("  [%s] %s\n", w.Code, w.Message))
			if w.Got != "" || w.Want != "" {
				sb.WriteString(fmt.Sprintf("    Got:  %s\n", w.Got))
				sb.WriteString(fmt.Sprintf("    Want: %s\n", w.Want))
			}
		}
	}
	
	return sb.String()
}
