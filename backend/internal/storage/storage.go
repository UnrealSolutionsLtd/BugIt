// Package storage handles filesystem operations for BugIt.
package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Storage manages the filesystem layout for repro bundles.
type Storage struct {
	dataDir    string
	bundlesDir string
	tmpDir     string
}

// New creates a new Storage instance.
func New(dataDir string) (*Storage, error) {
	s := &Storage{
		dataDir:    dataDir,
		bundlesDir: filepath.Join(dataDir, "bundles"),
		tmpDir:     filepath.Join(dataDir, "tmp"),
	}

	// Create directories
	for _, dir := range []string{s.dataDir, s.bundlesDir, s.tmpDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return s, nil
}

// DataDir returns the data directory path.
func (s *Storage) DataDir() string {
	return s.dataDir
}

// DBPath returns the SQLite database path.
func (s *Storage) DBPath() string {
	return filepath.Join(s.dataDir, "bugit.db")
}

// CreateTempDir creates a unique temporary directory for upload staging.
func (s *Storage) CreateTempDir(uploadID string) (string, error) {
	dir := filepath.Join(s.tmpDir, "upload_"+uploadID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	return dir, nil
}

// RemoveTempDir removes a temporary upload directory.
func (s *Storage) RemoveTempDir(dir string) error {
	return os.RemoveAll(dir)
}

// MoveToBundles atomically moves a directory from tmp to bundles.
func (s *Storage) MoveToBundles(srcDir, bundleID string) (string, error) {
	// Use first 8 chars of bundle_id for directory name (after "rb_" prefix)
	dirName := bundleID
	if len(dirName) > 11 {
		dirName = dirName[:11] // "rb_" + 8 chars
	}

	destDir := filepath.Join(s.bundlesDir, dirName)

	// Ensure bundles directory exists (may have been deleted)
	if err := os.MkdirAll(s.bundlesDir, 0755); err != nil {
		return "", fmt.Errorf("ensure bundles dir: %w", err)
	}

	// Check if destination already exists
	if _, err := os.Stat(destDir); err == nil {
		return "", fmt.Errorf("destination already exists: %s", destDir)
	}

	// Atomic rename (works on same filesystem)
	if err := os.Rename(srcDir, destDir); err != nil {
		return "", fmt.Errorf("rename to bundles: %w", err)
	}

	// Return relative path from data dir
	relPath, _ := filepath.Rel(s.dataDir, destDir)
	return relPath, nil
}

// BundlePath returns the absolute path to a bundle directory.
func (s *Storage) BundlePath(storagePath string) string {
	return filepath.Join(s.dataDir, storagePath)
}

// ArtifactPath returns the absolute path to an artifact file.
func (s *Storage) ArtifactPath(bundleStoragePath, artifactPath string) string {
	return filepath.Join(s.dataDir, bundleStoragePath, artifactPath)
}

// PurgeAllBundles removes all bundle directories from storage.
func (s *Storage) PurgeAllBundles() error {
	if err := os.RemoveAll(s.bundlesDir); err != nil {
		return fmt.Errorf("remove bundles dir: %w", err)
	}
	// Recreate the empty directory
	if err := os.MkdirAll(s.bundlesDir, 0755); err != nil {
		return fmt.Errorf("recreate bundles dir: %w", err)
	}
	return nil
}

// CheckHealth verifies storage is accessible.
func (s *Storage) CheckHealth() error {
	// Check data directory is writable
	testFile := filepath.Join(s.tmpDir, ".health_check")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("storage not writable: %w", err)
	}
	f.Close()
	os.Remove(testFile)
	return nil
}

// CleanupOldTempDirs removes temp directories older than maxAge.
func (s *Storage) CleanupOldTempDirs(maxAge time.Duration) (int, error) {
	entries, err := os.ReadDir(s.tmpDir)
	if err != nil {
		return 0, fmt.Errorf("read tmp dir: %w", err)
	}

	removed := 0
	cutoff := time.Now().Add(-maxAge)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			path := filepath.Join(s.tmpDir, entry.Name())
			if err := os.RemoveAll(path); err == nil {
				removed++
			}
		}
	}

	return removed, nil
}

// HashFile computes SHA256 hash of a file.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// HashReader computes SHA256 hash from a reader.
func HashReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// FileSize returns the size of a file in bytes.
func FileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// DirSize calculates the total size of all files in a directory.
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
