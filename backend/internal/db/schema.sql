-- BugIt SQLite Schema (embedded version)
-- Version: 1.0

PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

--------------------------------------------------------------------------------
-- repro_bundles: Core table for ingested repro bundles
--------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS repro_bundles (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    bundle_id       TEXT NOT NULL UNIQUE,
    content_hash    TEXT NOT NULL UNIQUE,
    schema_version  TEXT NOT NULL,
    build_id        TEXT NOT NULL,
    map_name        TEXT,
    platform        TEXT NOT NULL,
    rvr_version     TEXT,
    bundle_timestamp TEXT NOT NULL,
    metadata_json   TEXT,
    size_bytes      INTEGER NOT NULL DEFAULT 0,
    artifact_count  INTEGER NOT NULL DEFAULT 0,
    storage_path    TEXT NOT NULL,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    
    CHECK (bundle_id LIKE 'rb_%'),
    CHECK (schema_version != ''),
    CHECK (build_id != '')
);

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
    artifact_id     TEXT NOT NULL UNIQUE,
    bundle_id       TEXT NOT NULL,
    filename        TEXT NOT NULL,
    artifact_type   TEXT NOT NULL,
    mime_type       TEXT,
    size_bytes      INTEGER NOT NULL DEFAULT 0,
    storage_path    TEXT NOT NULL,
    checksum        TEXT,
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
    note_id         TEXT NOT NULL UNIQUE,
    bundle_id       TEXT NOT NULL,
    author          TEXT NOT NULL,
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

INSERT OR IGNORE INTO schema_migrations (version) VALUES (1);
