// Package models defines the data structures for BugIt.
package models

import (
	"encoding/json"
	"strings"
	"time"
)

// ReproBundle represents an ingested repro bundle.
type ReproBundle struct {
	ID              int64           `json:"-"`
	BundleID        string          `json:"bundle_id"`
	ContentHash     string          `json:"content_hash"`
	SchemaVersion   string          `json:"schema_version"`
	BuildID         string          `json:"build_id"`
	MapName         string          `json:"map_name,omitempty"`
	Platform        string          `json:"platform"`
	RVRVersion      string          `json:"rvr_version,omitempty"`
	BundleTimestamp time.Time       `json:"bundle_timestamp"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
	SizeBytes       int64           `json:"size_bytes"`
	ArtifactCount   int             `json:"artifact_count"`
	StoragePath     string          `json:"-"`
	CreatedAt       time.Time       `json:"created_at"`

	// Populated on detail queries
	Artifacts []Artifact `json:"artifacts,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Notes     []QANote   `json:"qa_notes,omitempty"`
}

// Artifact represents a file within a repro bundle.
type Artifact struct {
	ID           int64     `json:"-"`
	ArtifactID   string    `json:"artifact_id"`
	BundleID     string    `json:"-"`
	Filename     string    `json:"filename"`
	ArtifactType string    `json:"type"`
	MimeType     string    `json:"mime_type,omitempty"`
	SizeBytes    int64     `json:"size_bytes"`
	StoragePath  string    `json:"-"`
	Checksum     string    `json:"checksum,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Tag represents a label on a bundle.
type Tag struct {
	ID        int64     `json:"-"`
	BundleID  string    `json:"-"`
	Tag       string    `json:"tag"`
	CreatedAt time.Time `json:"created_at"`
}

// QANote represents a note added by a QA tester.
type QANote struct {
	ID        int64     `json:"-"`
	NoteID    string    `json:"note_id"`
	BundleID  string    `json:"-"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Manifest represents the manifest.json structure in a repro bundle.
// Matches Unreal's nested structure with camelCase field names.
type Manifest struct {
	SchemaVersion      string             `json:"schemaVersion"`
	BundleID           string             `json:"bundleId,omitempty"`
	ContentHash        string             `json:"contentHash,omitempty"`
	ReportTimestampUtc int64              `json:"reportTimestampUtc,omitempty"` // Unix ms from Unreal
	Timestamp          time.Time          `json:"-"`                            // Derived from ReportTimestampUtc
	Notes              string             `json:"notes,omitempty"`
	BuildInfo          *ManifestBuildInfo `json:"buildInfo,omitempty"`
	SessionInfo        *ManifestSession   `json:"sessionInfo,omitempty"`
	HardwareInfo       *ManifestHardware  `json:"hardwareInfo,omitempty"`
	Artifacts          []ManifestArtifact `json:"-"` // Custom unmarshal
	RawArtifacts       json.RawMessage    `json:"artifacts"`
	Metadata           json.RawMessage    `json:"metadata,omitempty"`

	// Derived fields for DB storage (populated after parsing)
	BuildID    string `json:"-"`
	MapName    string `json:"-"`
	Platform   string `json:"-"`
	RVRVersion string `json:"-"`
}

// ManifestBuildInfo contains build information
type ManifestBuildInfo struct {
	BuildID        string `json:"buildId"`
	CommitHash     string `json:"commitHash,omitempty"`
	Branch         string `json:"branch,omitempty"`
	BuildConfig    string `json:"buildConfig,omitempty"`
	EngineVersion  string `json:"engineVersion,omitempty"`
	ProjectName    string `json:"projectName,omitempty"`
	ProjectVersion string `json:"projectVersion,omitempty"`
	RVRVersion     string `json:"rvrVersion,omitempty"`
}

// ManifestSession contains session information
type ManifestSession struct {
	SessionID    string `json:"sessionId,omitempty"`
	MapName      string `json:"mapName,omitempty"`
	GameModeName string `json:"gameModeName,omitempty"`
	TesterName   string `json:"testerName,omitempty"`
	TestCaseName string `json:"testCaseName,omitempty"`
}

// ManifestHardware contains hardware information
type ManifestHardware struct {
	Platform  string `json:"platform"`
	OSVersion string `json:"osVersion,omitempty"`
	CPUBrand  string `json:"cpuBrand,omitempty"`
	GPUBrand  string `json:"gpuBrand,omitempty"`
	RHIName   string `json:"rhiName,omitempty"`
	DeviceID  string `json:"deviceId,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling to handle Unreal's manifest format
func (m *Manifest) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion
	type ManifestAlias Manifest
	aux := &struct {
		*ManifestAlias
	}{
		ManifestAlias: (*ManifestAlias)(m),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Convert Unix milliseconds to time.Time
	if m.ReportTimestampUtc > 0 {
		m.Timestamp = time.UnixMilli(m.ReportTimestampUtc)
	}

	// Populate derived fields from nested structures
	if m.BuildInfo != nil {
		if m.BuildInfo.BuildID != "" {
			m.BuildID = m.BuildInfo.BuildID
		}
		m.RVRVersion = m.BuildInfo.RVRVersion
	}
	if m.SessionInfo != nil {
		m.MapName = m.SessionInfo.MapName
	}
	if m.HardwareInfo != nil {
		m.Platform = m.HardwareInfo.Platform
	}

	// Fallback: use bundleId as buildId if buildInfo.buildId is empty
	if m.BuildID == "" && m.BundleID != "" {
		m.BuildID = m.BundleID
	}

	// Now parse artifacts - can be string[] or ManifestArtifact[]
	if len(m.RawArtifacts) > 0 {
		// Try array of objects first
		var artifacts []ManifestArtifact
		if err := json.Unmarshal(m.RawArtifacts, &artifacts); err == nil {
			m.Artifacts = artifacts
			return nil
		}

		// Try array of strings (Unreal's format)
		var filenames []string
		if err := json.Unmarshal(m.RawArtifacts, &filenames); err == nil {
			m.Artifacts = make([]ManifestArtifact, len(filenames))
			for i, filename := range filenames {
				m.Artifacts[i] = ManifestArtifact{
					Filename: filename,
					Type:     guessArtifactType(filename),
					MimeType: guessMimeType(filename),
				}
			}
			return nil
		}
	}

	return nil
}

// ManifestArtifact describes an artifact in the manifest.
type ManifestArtifact struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	MimeType string `json:"mime_type,omitempty"`
}

// guessArtifactType infers artifact type from filename
func guessArtifactType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".mp4"), strings.HasSuffix(filename, ".webm"):
		return "video"
	case strings.HasSuffix(filename, ".txt"), strings.HasSuffix(filename, ".log"):
		return "log"
	case strings.HasSuffix(filename, ".jpg"), strings.HasSuffix(filename, ".png"):
		if strings.Contains(filename, "thumbnail") {
			return "thumbnail"
		}
		return "screenshot"
	case strings.HasSuffix(filename, ".json"):
		return "other"
	case strings.HasSuffix(filename, ".dmp"):
		return "crash_dump"
	default:
		return "other"
	}
}

// guessMimeType infers MIME type from filename
func guessMimeType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".mp4"):
		return "video/mp4"
	case strings.HasSuffix(filename, ".webm"):
		return "video/webm"
	case strings.HasSuffix(filename, ".jpg"), strings.HasSuffix(filename, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(filename, ".png"):
		return "image/png"
	case strings.HasSuffix(filename, ".json"):
		return "application/json"
	case strings.HasSuffix(filename, ".txt"), strings.HasSuffix(filename, ".log"):
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// Validate checks the manifest for required fields.
func (m *Manifest) Validate() error {
	if m.SchemaVersion == "" {
		return &ValidationError{Field: "schemaVersion", Message: "required"}
	}
	// Accept 1.0, 1.0.0, or any 1.x version
	if !strings.HasPrefix(m.SchemaVersion, "1.") && m.SchemaVersion != "1" {
		return &ValidationError{Field: "schemaVersion", Message: "unsupported version: " + m.SchemaVersion}
	}
	// BuildID can come from nested buildInfo or direct field
	if m.BuildID == "" {
		// Try to use bundleId as fallback
		if m.BundleID != "" {
			m.BuildID = m.BundleID
		} else {
			return &ValidationError{Field: "buildId", Message: "required (in buildInfo or as bundleId)"}
		}
	}
	// Platform can come from nested hardwareInfo - default to "Other" if missing
	if m.Platform == "" {
		m.Platform = "Other"
	}
	// Timestamp is optional - use current time if not provided
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}
	return nil
}

// ValidationError represents a manifest validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Field + ": " + e.Message
}

// BundleListQuery defines parameters for listing bundles.
type BundleListQuery struct {
	BuildID  string
	MapName  string
	Platform string
	Since    *time.Time
	Limit    int
	Offset   int
}

// BundleListResult contains paginated bundle results.
type BundleListResult struct {
	Bundles []ReproBundle `json:"bundles"`
	Total   int           `json:"total"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
}

// IngestResult represents the result of ingesting a bundle.
type IngestResult struct {
	BundleID      string `json:"bundle_id"`
	Status        string `json:"status"` // "ingested" or "already_exists"
	ArtifactCount int    `json:"artifact_count"`
	CreatedAt     string `json:"created_at"`
}

// HealthStatus represents the health check response.
type HealthStatus struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Database string `json:"database"`
	Storage  string `json:"storage"`
}

// APIError represents an error response.
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Code + ": " + e.Message
}

// NewAPIError creates a new API error.
func NewAPIError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

// WithDetails adds details to an API error.
func (e *APIError) WithDetails(key string, value interface{}) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Common error codes
const (
	ErrCodeInvalidManifest   = "INVALID_MANIFEST"
	ErrCodeUnsupportedSchema = "UNSUPPORTED_SCHEMA"
	ErrCodeInvalidZip        = "INVALID_ZIP"
	ErrCodeBundleNotFound    = "BUNDLE_NOT_FOUND"
	ErrCodeArtifactNotFound  = "ARTIFACT_NOT_FOUND"
	ErrCodeStorageError      = "STORAGE_ERROR"
	ErrCodeDatabaseError     = "DATABASE_ERROR"
)
