// Package ingest handles repro bundle ingestion.
package ingest

import (
	"archive/zip"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/unrealsolutions/bugit/internal/db"
	"github.com/unrealsolutions/bugit/internal/models"
	"github.com/unrealsolutions/bugit/internal/storage"
)

// Ingester processes repro bundle uploads.
type Ingester struct {
	db      *db.DB
	storage *storage.Storage
}

// New creates a new Ingester.
func New(database *db.DB, store *storage.Storage) *Ingester {
	return &Ingester{
		db:      database,
		storage: store,
	}
}

// IngestResult contains the outcome of ingestion.
type IngestResult struct {
	BundleID      string `json:"bundle_id"`
	Status        string `json:"status"` // "ingested" or "already_exists"
	ArtifactCount int    `json:"artifact_count"`
	CreatedAt     string `json:"created_at"`
}

// IngestZipFile ingests a repro bundle from a ZIP file path.
func (i *Ingester) IngestZipFile(zipPath string) (*IngestResult, error) {
	// Generate unique upload ID
	uploadID := generateID(8)

	// Create temp directory for extraction
	tmpDir, err := i.storage.CreateTempDir(uploadID)
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	// Ensure cleanup on failure
	success := false
	defer func() {
		if !success {
			i.storage.RemoveTempDir(tmpDir)
		}
	}()

	// Compute content hash of ZIP
	contentHash, err := storage.HashFile(zipPath)
	if err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeInvalidZip,
			Message: fmt.Sprintf("failed to hash zip: %v", err),
		}
	}

	// Extract ZIP
	if err := extractZip(zipPath, tmpDir); err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeInvalidZip,
			Message: fmt.Sprintf("failed to extract zip: %v", err),
		}
	}

	// Parse and validate manifest
	manifest, err := parseManifest(filepath.Join(tmpDir, "manifest.json"))
	if err != nil {
		return nil, err
	}

	if err := manifest.Validate(); err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeInvalidManifest,
			Message: err.Error(),
		}
	}

	// Calculate total size
	totalSize, err := storage.DirSize(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("calculate size: %w", err)
	}

	// Generate bundle ID
	bundleID := "rb_" + generateID(8)

	// Create bundle record
	bundle := &models.ReproBundle{
		BundleID:        bundleID,
		ContentHash:     contentHash,
		SchemaVersion:   manifest.SchemaVersion,
		BuildID:         manifest.BuildID,
		MapName:         manifest.MapName,
		Platform:        manifest.Platform,
		RVRVersion:      manifest.RVRVersion,
		BundleTimestamp: manifest.Timestamp,
		Metadata:        manifest.Metadata,
		SizeBytes:       totalSize,
		ArtifactCount:   len(manifest.Artifacts),
	}

	// Insert bundle (handles idempotency via content hash)
	existingID, alreadyExists, err := i.db.InsertBundle(bundle)
	if err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeDatabaseError,
			Message: fmt.Sprintf("insert bundle: %v", err),
		}
	}

	if alreadyExists {
		// Bundle already exists, clean up and return existing
		i.storage.RemoveTempDir(tmpDir)
		return &IngestResult{
			BundleID:      existingID,
			Status:        "already_exists",
			ArtifactCount: len(manifest.Artifacts),
		}, nil
	}

	// Move to permanent storage
	storagePath, err := i.storage.MoveToBundles(tmpDir, bundleID)
	if err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeStorageError,
			Message: fmt.Sprintf("move to storage: %v", err),
		}
	}

	// Update bundle with storage path (would need additional DB method)
	// For simplicity, we set it before insert - this is handled by updating the model

	// Insert artifacts
	for _, ma := range manifest.Artifacts {
		artifactPath := filepath.Join(i.storage.BundlePath(storagePath), ma.Filename)
		size, _ := storage.FileSize(artifactPath)

		artifact := &models.Artifact{
			ArtifactID:   "art_" + generateID(8),
			BundleID:     bundleID,
			Filename:     ma.Filename,
			ArtifactType: normalizeArtifactType(ma.Type),
			MimeType:     ma.MimeType,
			SizeBytes:    size,
			StoragePath:  ma.Filename,
		}

		if err := i.db.InsertArtifact(artifact); err != nil {
			// Log but don't fail - bundle is already stored
			fmt.Printf("warning: failed to insert artifact %s: %v\n", ma.Filename, err)
		}
	}

	success = true
	return &IngestResult{
		BundleID:      bundleID,
		Status:        "ingested",
		ArtifactCount: len(manifest.Artifacts),
	}, nil
}

// IngestFromReader ingests a repro bundle from a reader (for HTTP uploads).
func (i *Ingester) IngestFromReader(r io.Reader, contentLength int64) (*IngestResult, error) {
	// Generate unique upload ID
	uploadID := generateID(8)

	// Create temp directory
	tmpDir, err := i.storage.CreateTempDir(uploadID)
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	// Ensure cleanup on failure
	success := false
	defer func() {
		if !success {
			i.storage.RemoveTempDir(tmpDir)
		}
	}()

	// Write uploaded file to temp location
	zipPath := filepath.Join(tmpDir, "upload.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}

	// Hash while writing
	h := sha256.New()
	written, err := io.Copy(io.MultiWriter(f, h), r)
	f.Close()
	if err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeInvalidZip,
			Message: fmt.Sprintf("failed to read upload: %v", err),
		}
	}

	contentHash := "sha256:" + hex.EncodeToString(h.Sum(nil))

	// Extract to a subdirectory
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, fmt.Errorf("create extract dir: %w", err)
	}

	if err := extractZip(zipPath, extractDir); err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeInvalidZip,
			Message: fmt.Sprintf("failed to extract zip: %v", err),
		}
	}

	// Remove the zip file to save space
	os.Remove(zipPath)

	// Parse and validate manifest
	manifest, err := parseManifest(filepath.Join(extractDir, "manifest.json"))
	if err != nil {
		return nil, err
	}

	if err := manifest.Validate(); err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeInvalidManifest,
			Message: err.Error(),
		}
	}

	// Generate bundle ID
	bundleID := "rb_" + generateID(8)

	// Create bundle record
	bundle := &models.ReproBundle{
		BundleID:        bundleID,
		ContentHash:     contentHash,
		SchemaVersion:   manifest.SchemaVersion,
		BuildID:         manifest.BuildID,
		MapName:         manifest.MapName,
		Platform:        manifest.Platform,
		RVRVersion:      manifest.RVRVersion,
		BundleTimestamp: manifest.Timestamp,
		Metadata:        manifest.Metadata,
		SizeBytes:       written,
		ArtifactCount:   len(manifest.Artifacts),
	}

	// Insert bundle (handles idempotency)
	existingID, alreadyExists, err := i.db.InsertBundle(bundle)
	if err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeDatabaseError,
			Message: fmt.Sprintf("insert bundle: %v", err),
		}
	}

	if alreadyExists {
		i.storage.RemoveTempDir(tmpDir)
		return &IngestResult{
			BundleID:      existingID,
			Status:        "already_exists",
			ArtifactCount: len(manifest.Artifacts),
		}, nil
	}

	// Move extracted files to permanent storage
	storagePath, err := i.storage.MoveToBundles(extractDir, bundleID)
	if err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeStorageError,
			Message: fmt.Sprintf("move to storage: %v", err),
		}
	}

	// Clean up remaining temp dir
	i.storage.RemoveTempDir(tmpDir)

	// Insert artifacts
	for _, ma := range manifest.Artifacts {
		artifactPath := filepath.Join(i.storage.BundlePath(storagePath), ma.Filename)
		size, _ := storage.FileSize(artifactPath)

		artifact := &models.Artifact{
			ArtifactID:   "art_" + generateID(8),
			BundleID:     bundleID,
			Filename:     ma.Filename,
			ArtifactType: normalizeArtifactType(ma.Type),
			MimeType:     ma.MimeType,
			SizeBytes:    size,
			StoragePath:  ma.Filename,
		}

		if err := i.db.InsertArtifact(artifact); err != nil {
			fmt.Printf("warning: failed to insert artifact %s: %v\n", ma.Filename, err)
		}
	}

	success = true
	return &IngestResult{
		BundleID:      bundleID,
		Status:        "ingested",
		ArtifactCount: len(manifest.Artifacts),
	}, nil
}

// IngestFromFiles ingests a repro bundle from individual files (for direct multipart uploads).
// This supports Unreal Engine uploads that send files individually rather than as a ZIP.
func (i *Ingester) IngestFromFiles(files map[string][]byte) (*IngestResult, error) {
	// Generate unique upload ID
	uploadID := generateID(8)

	// Create temp directory
	tmpDir, err := i.storage.CreateTempDir(uploadID)
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	// Ensure cleanup on failure
	success := false
	defer func() {
		if !success {
			i.storage.RemoveTempDir(tmpDir)
		}
	}()

	// Compute content hash from all files
	h := sha256.New()
	var totalSize int64
	
	// Write all files to temp directory and compute hash
	for filename, data := range files {
		// Security: prevent path traversal
		if strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\") {
			// Flatten any path - just use base filename
			filename = filepath.Base(filename)
		}
		
		destPath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return nil, fmt.Errorf("write file %s: %w", filename, err)
		}
		
		h.Write(data)
		totalSize += int64(len(data))
	}
	
	contentHash := "sha256:" + hex.EncodeToString(h.Sum(nil))

	// Parse and validate manifest
	manifest, err := parseManifest(filepath.Join(tmpDir, "manifest.json"))
	if err != nil {
		return nil, err
	}

	if err := manifest.Validate(); err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeInvalidManifest,
			Message: err.Error(),
		}
	}

	// Generate bundle ID
	bundleID := "rb_" + generateID(8)

	// Move to permanent storage FIRST (so we have the storage path for the bundle record)
	storagePath, err := i.storage.MoveToBundles(tmpDir, bundleID)
	if err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeStorageError,
			Message: fmt.Sprintf("move to storage: %v", err),
		}
	}

	// Create bundle record with storage path
	bundle := &models.ReproBundle{
		BundleID:        bundleID,
		ContentHash:     contentHash,
		SchemaVersion:   manifest.SchemaVersion,
		BuildID:         manifest.BuildID,
		MapName:         manifest.MapName,
		Platform:        manifest.Platform,
		RVRVersion:      manifest.RVRVersion,
		BundleTimestamp: manifest.Timestamp,
		Metadata:        manifest.Metadata,
		SizeBytes:       totalSize,
		ArtifactCount:   len(manifest.Artifacts),
		StoragePath:     storagePath,
	}

	// Insert bundle (handles idempotency)
	existingID, alreadyExists, err := i.db.InsertBundle(bundle)
	if err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeDatabaseError,
			Message: fmt.Sprintf("insert bundle: %v", err),
		}
	}

	if alreadyExists {
		// Files already moved, but bundle exists - this shouldn't happen with content hash check
		// but handle it gracefully
		return &IngestResult{
			BundleID:      existingID,
			Status:        "already_exists",
			ArtifactCount: len(manifest.Artifacts),
		}, nil
	}

	// Insert artifacts
	for _, ma := range manifest.Artifacts {
		artifactPath := filepath.Join(i.storage.BundlePath(storagePath), ma.Filename)
		size, _ := storage.FileSize(artifactPath)

		artifact := &models.Artifact{
			ArtifactID:   "art_" + generateID(8),
			BundleID:     bundleID,
			Filename:     ma.Filename,
			ArtifactType: normalizeArtifactType(ma.Type),
			MimeType:     ma.MimeType,
			SizeBytes:    size,
			StoragePath:  ma.Filename,
		}

		if err := i.db.InsertArtifact(artifact); err != nil {
			fmt.Printf("warning: failed to insert artifact %s: %v\n", ma.Filename, err)
		}
	}

	success = true
	return &IngestResult{
		BundleID:      bundleID,
		Status:        "ingested",
		ArtifactCount: len(manifest.Artifacts),
	}, nil
}

// parseManifest reads and parses manifest.json.
func parseManifest(path string) (*models.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeInvalidManifest,
			Message: "manifest.json not found or unreadable",
		}
	}

	var manifest models.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, &models.APIError{
			Code:    models.ErrCodeInvalidManifest,
			Message: fmt.Sprintf("invalid JSON in manifest.json: %v", err),
		}
	}

	return &manifest, nil
}

// extractZip extracts a ZIP file to destination directory.
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		// Security: prevent path traversal
		destPath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, f.Mode())
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("create dir for %s: %w", f.Name, err)
		}

		// Extract file
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open %s in zip: %w", f.Name, err)
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("create %s: %w", destPath, err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("extract %s: %w", f.Name, err)
		}
	}

	return nil
}

// generateID generates a random hex ID of specified length.
func generateID(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// normalizeArtifactType maps manifest types to DB enum values.
func normalizeArtifactType(t string) string {
	switch strings.ToLower(t) {
	case "video":
		return "video"
	case "log":
		return "log"
	case "screenshot":
		return "screenshot"
	case "crash_dump", "crashdump", "dump":
		return "crash_dump"
	case "thumbnail", "thumb":
		return "thumbnail"
	default:
		return "other"
	}
}
