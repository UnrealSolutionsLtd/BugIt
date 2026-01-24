// Package api provides the HTTP API for BugIt.
package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/unrealsolutions/bugit/internal/db"
	"github.com/unrealsolutions/bugit/internal/ingest"
	"github.com/unrealsolutions/bugit/internal/models"
	"github.com/unrealsolutions/bugit/internal/storage"
)

// Server is the HTTP API server.
type Server struct {
	db       *db.DB
	storage  *storage.Storage
	ingester *ingest.Ingester
	version  string
	logger   *slog.Logger
}

// Config holds server configuration.
type Config struct {
	Port        int
	DataDir     string
	Version     string
	MaxUploadMB int
}

// NewServer creates a new API server.
func NewServer(database *db.DB, store *storage.Storage, version string) *Server {
	return &Server{
		db:       database,
		storage:  store,
		ingester: ingest.New(database, store),
		version:  version,
		logger:   slog.Default(),
	}
}

// Handler returns the HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/health", s.handleHealth)

	// Repro bundles
	mux.HandleFunc("POST /api/repro-bundles", s.handleIngestBundle)
	mux.HandleFunc("GET /api/repro-bundles", s.handleListBundles)
	mux.HandleFunc("GET /api/repro-bundles/{bundle_id}", s.handleGetBundle)
	mux.HandleFunc("GET /api/repro-bundles/{bundle_id}/artifacts/{artifact_id}", s.handleGetArtifact)
	mux.HandleFunc("POST /api/repro-bundles/{bundle_id}/tags", s.handleAddTags)
	mux.HandleFunc("POST /api/repro-bundles/{bundle_id}/notes", s.handleAddNote)

	// Wrap with middleware
	return s.loggingMiddleware(mux)
}

// loggingMiddleware logs all requests.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// handleHealth handles GET /api/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := models.HealthStatus{
		Status:   "ok",
		Version:  s.version,
		Database: "ok",
		Storage:  "ok",
	}

	if err := s.db.CheckHealth(); err != nil {
		status.Status = "degraded"
		status.Database = "error"
	}

	if err := s.storage.CheckHealth(); err != nil {
		status.Status = "degraded"
		status.Storage = "error"
	}

	s.writeJSON(w, http.StatusOK, status)
}

// handleIngestBundle handles POST /api/repro-bundles
func (s *Server) handleIngestBundle(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	var result *ingest.IngestResult
	var err error

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle multipart upload
		if err := r.ParseMultipartForm(500 << 20); err != nil { // 500MB max
			s.writeError(w, http.StatusBadRequest, &models.APIError{
				Code:    models.ErrCodeInvalidZip,
				Message: "failed to parse multipart form",
			})
			return
		}

		// Try ZIP file first (field name "file")
		file, _, zipErr := r.FormFile("file")
		if zipErr == nil {
			defer file.Close()
			result, err = s.ingester.IngestFromReader(file, r.ContentLength)
		} else {
			// No "file" field - try direct multipart files
			// This supports Unreal Engine uploads with individual files
			files := make(map[string][]byte)
			
			for name, headers := range r.MultipartForm.File {
				if len(headers) == 0 {
					continue
				}
				f, openErr := headers[0].Open()
				if openErr != nil {
					continue
				}
				data, readErr := io.ReadAll(f)
				f.Close()
				if readErr != nil {
					continue
				}
				// Use the form field name as the filename
				files[name] = data
			}
			
			if len(files) == 0 {
				s.writeError(w, http.StatusBadRequest, &models.APIError{
					Code:    models.ErrCodeInvalidZip,
					Message: "no files found in multipart form",
				})
				return
			}
			
			result, err = s.ingester.IngestFromFiles(files)
		}
	} else {
		// Handle raw ZIP upload
		result, err = s.ingester.IngestFromReader(r.Body, r.ContentLength)
	}

	if err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			status := http.StatusBadRequest
			if apiErr.Code == models.ErrCodeStorageError || apiErr.Code == models.ErrCodeDatabaseError {
				status = http.StatusInternalServerError
			}
			s.writeError(w, status, apiErr)
		} else {
			s.writeError(w, http.StatusInternalServerError, &models.APIError{
				Code:    models.ErrCodeStorageError,
				Message: err.Error(),
			})
		}
		return
	}

	status := http.StatusCreated
	if result.Status == "already_exists" {
		status = http.StatusOK
	}

	s.writeJSON(w, status, result)
}

// handleListBundles handles GET /api/repro-bundles
func (s *Server) handleListBundles(w http.ResponseWriter, r *http.Request) {
	query := &models.BundleListQuery{
		BuildID:  r.URL.Query().Get("build_id"),
		MapName:  r.URL.Query().Get("map_name"),
		Platform: r.URL.Query().Get("platform"),
	}

	if since := r.URL.Query().Get("since"); since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err == nil {
			query.Since = &t
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		query.Limit, _ = strconv.Atoi(limit)
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		query.Offset, _ = strconv.Atoi(offset)
	}

	result, err := s.db.ListBundles(query)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, &models.APIError{
			Code:    models.ErrCodeDatabaseError,
			Message: err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

// handleGetBundle handles GET /api/repro-bundles/{bundle_id}
func (s *Server) handleGetBundle(w http.ResponseWriter, r *http.Request) {
	bundleID := r.PathValue("bundle_id")
	if bundleID == "" {
		s.writeError(w, http.StatusBadRequest, &models.APIError{
			Code:    models.ErrCodeBundleNotFound,
			Message: "bundle_id required",
		})
		return
	}

	bundle, err := s.db.GetBundle(bundleID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, &models.APIError{
			Code:    models.ErrCodeDatabaseError,
			Message: err.Error(),
		})
		return
	}

	if bundle == nil {
		s.writeError(w, http.StatusNotFound, &models.APIError{
			Code:    models.ErrCodeBundleNotFound,
			Message: "bundle not found: " + bundleID,
		})
		return
	}

	s.writeJSON(w, http.StatusOK, bundle)
}

// handleGetArtifact handles GET /api/repro-bundles/{bundle_id}/artifacts/{artifact_id}
func (s *Server) handleGetArtifact(w http.ResponseWriter, r *http.Request) {
	bundleID := r.PathValue("bundle_id")
	artifactID := r.PathValue("artifact_id")

	// Get bundle to find storage path
	bundle, err := s.db.GetBundle(bundleID)
	if err != nil || bundle == nil {
		s.writeError(w, http.StatusNotFound, &models.APIError{
			Code:    models.ErrCodeBundleNotFound,
			Message: "bundle not found: " + bundleID,
		})
		return
	}

	// Get artifact
	artifact, err := s.db.GetArtifact(artifactID)
	if err != nil || artifact == nil {
		s.writeError(w, http.StatusNotFound, &models.APIError{
			Code:    models.ErrCodeArtifactNotFound,
			Message: "artifact not found: " + artifactID,
		})
		return
	}

	// Verify artifact belongs to bundle
	if artifact.BundleID != bundleID {
		s.writeError(w, http.StatusNotFound, &models.APIError{
			Code:    models.ErrCodeArtifactNotFound,
			Message: "artifact not found in bundle",
		})
		return
	}

	// Get file path
	filePath := s.storage.ArtifactPath(bundle.StoragePath, artifact.StoragePath)
	s.logger.Info("serving artifact", "path", filePath, "bundle_storage", bundle.StoragePath, "artifact_storage", artifact.StoragePath)

	// Open file
	f, err := os.Open(filePath)
	if err != nil {
		s.logger.Error("failed to open artifact", "path", filePath, "error", err)
		s.writeError(w, http.StatusInternalServerError, &models.APIError{
			Code:    models.ErrCodeStorageError,
			Message: fmt.Sprintf("failed to open artifact file: %v (path: %s)", err, filePath),
		})
		return
	}
	defer f.Close()

	// Set content type
	contentType := artifact.MimeType
	if contentType == "" {
		contentType = getMimeType(artifact.Filename)
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", artifact.Filename))
	w.Header().Set("Content-Length", strconv.FormatInt(artifact.SizeBytes, 10))

	io.Copy(w, f)
}

// handleAddTags handles POST /api/repro-bundles/{bundle_id}/tags
func (s *Server) handleAddTags(w http.ResponseWriter, r *http.Request) {
	bundleID := r.PathValue("bundle_id")

	var req struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, &models.APIError{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	// Verify bundle exists
	bundle, _ := s.db.GetBundle(bundleID)
	if bundle == nil {
		s.writeError(w, http.StatusNotFound, &models.APIError{
			Code:    models.ErrCodeBundleNotFound,
			Message: "bundle not found: " + bundleID,
		})
		return
	}

	for _, tag := range req.Tags {
		if err := s.db.AddTag(bundleID, tag); err != nil {
			s.logger.Error("failed to add tag", "bundle_id", bundleID, "tag", tag, "error", err)
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleAddNote handles POST /api/repro-bundles/{bundle_id}/notes
func (s *Server) handleAddNote(w http.ResponseWriter, r *http.Request) {
	bundleID := r.PathValue("bundle_id")

	var req struct {
		Author  string `json:"author"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, &models.APIError{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	if req.Author == "" || req.Content == "" {
		s.writeError(w, http.StatusBadRequest, &models.APIError{
			Code:    "INVALID_REQUEST",
			Message: "author and content are required",
		})
		return
	}

	// Verify bundle exists
	bundle, _ := s.db.GetBundle(bundleID)
	if bundle == nil {
		s.writeError(w, http.StatusNotFound, &models.APIError{
			Code:    models.ErrCodeBundleNotFound,
			Message: "bundle not found: " + bundleID,
		})
		return
	}

	note := &models.QANote{
		NoteID:  "note_" + generateID(8),
		Author:  req.Author,
		Content: req.Content,
	}

	if err := s.db.AddNote(bundleID, note); err != nil {
		s.writeError(w, http.StatusInternalServerError, &models.APIError{
			Code:    models.ErrCodeDatabaseError,
			Message: err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusCreated, note)
}

// writeJSON writes a JSON response.
func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes an error response.
func (s *Server) writeError(w http.ResponseWriter, status int, err *models.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": err})
}

// getMimeType returns MIME type based on file extension.
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".log", ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".dmp":
		return "application/octet-stream"
	default:
		return "application/octet-stream"
	}
}

// generateID generates a random hex ID.
func generateID(length int) string {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%x", time.Now().UnixNano())[:length]
	}
	return fmt.Sprintf("%x", b)[:length]
}
