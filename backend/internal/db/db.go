// Package db provides SQLite database operations for BugIt.
package db

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver - no CGO required
	"github.com/unrealsolutions/bugit/internal/models"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps the SQLite database connection.
type DB struct {
	conn *sql.DB
}

// Open opens the SQLite database and initializes the schema.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Set connection pool settings - allow multiple readers
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return db, nil
}

// OpenWithSchema opens the database with an external schema file.
func OpenWithSchema(dbPath, schemaPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	return &DB{conn: conn}, nil
}

func (db *DB) initSchema() error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	_, err = db.conn.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// CheckHealth verifies database connectivity.
func (db *DB) CheckHealth() error {
	var result int
	return db.conn.QueryRow("SELECT 1").Scan(&result)
}

// InsertBundle inserts a new repro bundle.
// Returns the bundle_id if successful, or existing bundle_id if content_hash exists.
func (db *DB) InsertBundle(bundle *models.ReproBundle) (string, bool, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return "", false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Check if bundle already exists by content hash
	var existingID string
	err = tx.QueryRow(
		"SELECT bundle_id FROM repro_bundles WHERE content_hash = ?",
		bundle.ContentHash,
	).Scan(&existingID)

	if err == nil {
		// Bundle already exists
		return existingID, true, nil
	} else if err != sql.ErrNoRows {
		return "", false, fmt.Errorf("check existing: %w", err)
	}

	// Insert new bundle
	metadataJSON := ""
	if bundle.Metadata != nil {
		metadataJSON = string(bundle.Metadata)
	}

	_, err = tx.Exec(`
		INSERT INTO repro_bundles (
			bundle_id, content_hash, schema_version, build_id, map_name,
			platform, rvr_version, bundle_timestamp, metadata_json,
			size_bytes, artifact_count, storage_path
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		bundle.BundleID,
		bundle.ContentHash,
		bundle.SchemaVersion,
		bundle.BuildID,
		bundle.MapName,
		bundle.Platform,
		bundle.RVRVersion,
		bundle.BundleTimestamp.Format(time.RFC3339),
		metadataJSON,
		bundle.SizeBytes,
		bundle.ArtifactCount,
		bundle.StoragePath,
	)
	if err != nil {
		return "", false, fmt.Errorf("insert bundle: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", false, fmt.Errorf("commit: %w", err)
	}

	return bundle.BundleID, false, nil
}

// InsertArtifact inserts an artifact for a bundle.
func (db *DB) InsertArtifact(artifact *models.Artifact) error {
	_, err := db.conn.Exec(`
		INSERT INTO artifacts (
			artifact_id, bundle_id, filename, artifact_type,
			mime_type, size_bytes, storage_path, checksum
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		artifact.ArtifactID,
		artifact.BundleID,
		artifact.Filename,
		artifact.ArtifactType,
		artifact.MimeType,
		artifact.SizeBytes,
		artifact.StoragePath,
		artifact.Checksum,
	)
	return err
}

// GetBundle retrieves a bundle by ID with all related data.
func (db *DB) GetBundle(bundleID string) (*models.ReproBundle, error) {
	bundle := &models.ReproBundle{}
	var metadataJSON sql.NullString
	var bundleTimestamp string
	var createdAt string

	err := db.conn.QueryRow(`
		SELECT id, bundle_id, content_hash, schema_version, build_id, map_name,
		       platform, rvr_version, bundle_timestamp, metadata_json,
		       size_bytes, artifact_count, storage_path, created_at
		FROM repro_bundles WHERE bundle_id = ?`, bundleID,
	).Scan(
		&bundle.ID,
		&bundle.BundleID,
		&bundle.ContentHash,
		&bundle.SchemaVersion,
		&bundle.BuildID,
		&bundle.MapName,
		&bundle.Platform,
		&bundle.RVRVersion,
		&bundleTimestamp,
		&metadataJSON,
		&bundle.SizeBytes,
		&bundle.ArtifactCount,
		&bundle.StoragePath,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query bundle: %w", err)
	}

	bundle.BundleTimestamp, _ = time.Parse(time.RFC3339, bundleTimestamp)
	bundle.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if metadataJSON.Valid {
		bundle.Metadata = json.RawMessage(metadataJSON.String)
	}

	// Load artifacts
	bundle.Artifacts, err = db.GetArtifacts(bundleID)
	if err != nil {
		return nil, fmt.Errorf("get artifacts: %w", err)
	}

	// Load tags
	bundle.Tags, err = db.GetTags(bundleID)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}

	// Load notes
	bundle.Notes, err = db.GetNotes(bundleID)
	if err != nil {
		return nil, fmt.Errorf("get notes: %w", err)
	}

	return bundle, nil
}

// GetArtifacts retrieves all artifacts for a bundle.
func (db *DB) GetArtifacts(bundleID string) ([]models.Artifact, error) {
	rows, err := db.conn.Query(`
		SELECT artifact_id, filename, artifact_type, mime_type,
		       size_bytes, storage_path, checksum, created_at
		FROM artifacts WHERE bundle_id = ?
		ORDER BY filename`, bundleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []models.Artifact
	for rows.Next() {
		var a models.Artifact
		var mimeType, checksum sql.NullString
		var createdAt string

		err := rows.Scan(
			&a.ArtifactID,
			&a.Filename,
			&a.ArtifactType,
			&mimeType,
			&a.SizeBytes,
			&a.StoragePath,
			&checksum,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		a.BundleID = bundleID
		a.MimeType = mimeType.String
		a.Checksum = checksum.String
		a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		artifacts = append(artifacts, a)
	}

	return artifacts, rows.Err()
}

// GetArtifact retrieves a single artifact by ID.
func (db *DB) GetArtifact(artifactID string) (*models.Artifact, error) {
	var a models.Artifact
	var mimeType, checksum sql.NullString
	var createdAt string

	err := db.conn.QueryRow(`
		SELECT artifact_id, bundle_id, filename, artifact_type, mime_type,
		       size_bytes, storage_path, checksum, created_at
		FROM artifacts WHERE artifact_id = ?`, artifactID,
	).Scan(
		&a.ArtifactID,
		&a.BundleID,
		&a.Filename,
		&a.ArtifactType,
		&mimeType,
		&a.SizeBytes,
		&a.StoragePath,
		&checksum,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	a.MimeType = mimeType.String
	a.Checksum = checksum.String
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return &a, nil
}

// GetTags retrieves all tags for a bundle.
func (db *DB) GetTags(bundleID string) ([]string, error) {
	rows, err := db.conn.Query(
		"SELECT tag FROM tags WHERE bundle_id = ? ORDER BY tag",
		bundleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// AddTag adds a tag to a bundle.
func (db *DB) AddTag(bundleID, tag string) error {
	_, err := db.conn.Exec(
		"INSERT OR IGNORE INTO tags (bundle_id, tag) VALUES (?, ?)",
		bundleID, tag,
	)
	return err
}

// GetNotes retrieves all notes for a bundle.
func (db *DB) GetNotes(bundleID string) ([]models.QANote, error) {
	rows, err := db.conn.Query(`
		SELECT note_id, author, content, created_at
		FROM qa_notes WHERE bundle_id = ?
		ORDER BY created_at`, bundleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []models.QANote
	for rows.Next() {
		var n models.QANote
		var createdAt string

		if err := rows.Scan(&n.NoteID, &n.Author, &n.Content, &createdAt); err != nil {
			return nil, err
		}

		n.BundleID = bundleID
		n.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		notes = append(notes, n)
	}

	return notes, rows.Err()
}

// AddNote adds a note to a bundle.
func (db *DB) AddNote(bundleID string, note *models.QANote) error {
	_, err := db.conn.Exec(`
		INSERT INTO qa_notes (note_id, bundle_id, author, content)
		VALUES (?, ?, ?, ?)`,
		note.NoteID, bundleID, note.Author, note.Content,
	)
	return err
}

// ListBundles lists bundles with optional filtering.
func (db *DB) ListBundles(query *models.BundleListQuery) (*models.BundleListResult, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}

	if query.BuildID != "" {
		conditions = append(conditions, "build_id = ?")
		args = append(args, query.BuildID)
	}
	if query.MapName != "" {
		conditions = append(conditions, "map_name = ?")
		args = append(args, query.MapName)
	}
	if query.Platform != "" {
		conditions = append(conditions, "platform = ?")
		args = append(args, query.Platform)
	}
	if query.Since != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, query.Since.Format(time.RFC3339))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countSQL := "SELECT COUNT(*) FROM repro_bundles " + whereClause
	var total int
	if err := db.conn.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count bundles: %w", err)
	}

	// Apply defaults
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	// Query bundles
	querySQL := fmt.Sprintf(`
		SELECT bundle_id, content_hash, schema_version, build_id, map_name,
		       platform, rvr_version, bundle_timestamp, size_bytes,
		       artifact_count, created_at
		FROM repro_bundles %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, whereClause)

	args = append(args, limit, query.Offset)

	rows, err := db.conn.Query(querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query bundles: %w", err)
	}
	defer rows.Close()

	var bundles []models.ReproBundle
	for rows.Next() {
		var b models.ReproBundle
		var bundleTimestamp, createdAt string

		err := rows.Scan(
			&b.BundleID,
			&b.ContentHash,
			&b.SchemaVersion,
			&b.BuildID,
			&b.MapName,
			&b.Platform,
			&b.RVRVersion,
			&bundleTimestamp,
			&b.SizeBytes,
			&b.ArtifactCount,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan bundle: %w", err)
		}

		b.BundleTimestamp, _ = time.Parse(time.RFC3339, bundleTimestamp)
		b.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

		// Load tags for each bundle
		b.Tags, _ = db.GetTags(b.BundleID)

		bundles = append(bundles, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bundles: %w", err)
	}

	return &models.BundleListResult{
		Bundles: bundles,
		Total:   total,
		Limit:   limit,
		Offset:  query.Offset,
	}, nil
}
