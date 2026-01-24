-- BugIt SQLite Schema
-- Version: 1.0
-- 
-- Design principles:
-- - Immutable bundles (no UPDATE on core fields)
-- - Indexed for common query patterns
-- - Foreign keys enforced
-- - Timestamps in ISO8601 format

PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

--------------------------------------------------------------------------------
-- repro_bundles: Core table for ingested repro bundles
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS repro_bundles (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    bundle_id       TEXT NOT NULL UNIQUE,           -- External ID: rb_<8chars>
    content_hash    TEXT NOT NULL UNIQUE,           -- sha256:<hex> for idempotency
    schema_version  TEXT NOT NULL,                  -- Manifest schema version
    build_id        TEXT NOT NULL,                  -- Game build identifier
    map_name        TEXT,                           -- Unreal map path
    platform        TEXT NOT NULL,                  -- Win64, Linux, Android, iOS, etc.
    rvr_version     TEXT,                           -- Runtime Video Recorder version
    bundle_timestamp TEXT NOT NULL,                 -- When bundle was created (from manifest)
    metadata_json   TEXT,                           -- Full metadata blob from manifest
    size_bytes      INTEGER NOT NULL DEFAULT 0,     -- Total bundle size
    artifact_count  INTEGER NOT NULL DEFAULT 0,     -- Number of artifacts
    storage_path    TEXT NOT NULL,                  -- Relative path to bundle directory
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    
    -- Constraints
    CHECK (bundle_id LIKE 'rb_%'),
    CHECK (schema_version != ''),
    CHECK (build_id != '')
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_bundles_build_id ON repro_bundles(build_id);
CREATE INDEX IF NOT EXISTS idx_bundles_map_name ON repro_bundles(map_name);
CREATE INDEX IF NOT EXISTS idx_bundles_platform ON repro_bundles(platform);
CREATE INDEX IF NOT EXISTS idx_bundles_created_at ON repro_bundles(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bundles_content_hash ON repro_bundles(content_hash);

--------------------------------------------------------------------------------
-- artifacts: Individual files within a bundle
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS artifacts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    artifact_id     TEXT NOT NULL UNIQUE,           -- External ID: art_<8chars>
    bundle_id       TEXT NOT NULL,                  -- Parent bundle
    filename        TEXT NOT NULL,                  -- Original filename
    artifact_type   TEXT NOT NULL,                  -- video, log, screenshot, crash_dump, other
    mime_type       TEXT,                           -- MIME type
    size_bytes      INTEGER NOT NULL DEFAULT 0,     -- File size
    storage_path    TEXT NOT NULL,                  -- Relative path within bundle dir
    checksum        TEXT,                           -- Optional SHA256 of artifact
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    
    FOREIGN KEY (bundle_id) REFERENCES repro_bundles(bundle_id) ON DELETE CASCADE,
    
    CHECK (artifact_id LIKE 'art_%'),
    CHECK (artifact_type IN ('video', 'log', 'screenshot', 'crash_dump', 'thumbnail', 'other'))
);

CREATE INDEX IF NOT EXISTS idx_artifacts_bundle_id ON artifacts(bundle_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_type ON artifacts(artifact_type);

--------------------------------------------------------------------------------
-- tags: Labels attached to bundles
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tags (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    bundle_id       TEXT NOT NULL,
    tag             TEXT NOT NULL,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    
    FOREIGN KEY (bundle_id) REFERENCES repro_bundles(bundle_id) ON DELETE CASCADE,
    
    UNIQUE(bundle_id, tag),
    CHECK (tag != '' AND length(tag) <= 64)
);

CREATE INDEX IF NOT EXISTS idx_tags_bundle_id ON tags(bundle_id);
CREATE INDEX IF NOT EXISTS idx_tags_tag ON tags(tag);

--------------------------------------------------------------------------------
-- qa_notes: Notes added by QA testers
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS qa_notes (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id         TEXT NOT NULL UNIQUE,           -- External ID: note_<8chars>
    bundle_id       TEXT NOT NULL,
    author          TEXT NOT NULL,                  -- QA tester identifier
    content         TEXT NOT NULL,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    
    FOREIGN KEY (bundle_id) REFERENCES repro_bundles(bundle_id) ON DELETE CASCADE,
    
    CHECK (note_id LIKE 'note_%'),
    CHECK (author != ''),
    CHECK (content != '')
);

CREATE INDEX IF NOT EXISTS idx_notes_bundle_id ON qa_notes(bundle_id);
CREATE INDEX IF NOT EXISTS idx_notes_author ON qa_notes(author);

--------------------------------------------------------------------------------
-- schema_migrations: Track applied migrations
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS schema_migrations (
    version         INTEGER PRIMARY KEY,
    applied_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- Record this schema version
INSERT OR IGNORE INTO schema_migrations (version) VALUES (1);

--------------------------------------------------------------------------------
-- Views for common queries
--------------------------------------------------------------------------------

-- Bundle summary view with tag list
CREATE VIEW IF NOT EXISTS v_bundle_summary AS
SELECT 
    b.bundle_id,
    b.build_id,
    b.map_name,
    b.platform,
    b.created_at,
    b.artifact_count,
    b.size_bytes,
    GROUP_CONCAT(t.tag, ',') as tags
FROM repro_bundles b
LEFT JOIN tags t ON b.bundle_id = t.bundle_id
GROUP BY b.bundle_id;

-- Recent bundles (last 24 hours)
CREATE VIEW IF NOT EXISTS v_recent_bundles AS
SELECT * FROM repro_bundles
WHERE created_at >= datetime('now', '-24 hours')
ORDER BY created_at DESC;
