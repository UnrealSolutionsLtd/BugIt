# BugIt - Forensic QA Dashboard Backend

On-premises Go backend for ingesting and querying immutable repro bundles from [Runtime Video Recorder (RVR)](https://unrealsolutions.ltd/products/runtime-video-recorder).

## Features

- **Zero-CGO SQLite** - Pure Go SQLite driver, no C compiler needed
- **Multipart Upload Support** - Accept bundles from Unreal Engine via HTTP multipart
- **ZIP Upload Support** - Also accept standard ZIP file uploads
- **Idempotent Ingestion** - Duplicate bundles detected via SHA256 content hash
- **Artifact Streaming** - Serve video/log files directly to frontend
- **QA Annotations** - Add tags and timestamped notes to bundles

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    Unreal Engine Games                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  QA Tester  │  │  QA Tester  │  │  QA Tester  │             │
│  │  + RVR      │  │  + RVR      │  │  + RVR      │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
│         │                │                │                     │
│         └────────────────┼────────────────┘                     │
│                          │ POST /api/repro-bundles              │
│                          │ (multipart/form-data)                │
│                          ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                      BugIt Server                           ││
│  │  ┌───────────┐  ┌───────────────┐  ┌───────────────────┐   ││
│  │  │ HTTP API  │──│   Ingester    │──│ Artifact Storage  │   ││
│  │  │  (8080)   │  │               │  │   (filesystem)    │   ││
│  │  └───────────┘  └───────┬───────┘  └───────────────────┘   ││
│  │                         │                                   ││
│  │                    ┌────┴────┐                              ││
│  │                    │ SQLite  │                              ││
│  │                    │  (WAL)  │                              ││
│  │                    └─────────┘                              ││
│  └─────────────────────────────────────────────────────────────┘│
│                          │                                      │
│                          ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │              BugIt Frontend (React)                         ││
│  │  - Video playback with input timeline                       ││
│  │  - Frame timing graphs                                      ││
│  │  - Log viewer synced to video                               ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

```bash
# Build
go build -o bugit ./cmd/bugit

# Run server (Windows)
.\bugit.exe serve --port 8080 --data-dir ./data

# Run server (Linux/Mac)
./bugit serve --port 8080 --data-dir ./data

# Or run directly without building
go run ./cmd/bugit serve --port 8080 --data-dir ./data
```

The server will:
1. Create `./data` directory if it doesn't exist
2. Initialize SQLite database with WAL mode
3. Listen on `http://localhost:8080`

## API Design

### POST /api/repro-bundles

Ingest a new repro bundle. Supports two upload formats:

**Option 1: Multipart Form Data (Unreal Engine)**

Used by the RVR plugin's BugIt SDK for direct uploads from game.

```
Content-Type: multipart/form-data

Parts:
- manifest.json (required)
- video.mp4
- inputs.json
- timing.json
- logs.txt
- hardware.json
- ... any other artifacts
```

**Option 2: ZIP File Upload**

For manual uploads or CI/CD pipelines.

```
Content-Type: application/zip
Body: repro_bundle.zip
```

**Response:**
```json
{
  "bundle_id": "rb_a1b2c3d4",
  "status": "ingested",
  "artifact_count": 5
}
```

**Idempotency:** Bundles are identified by SHA256 hash of contents. Re-uploading the same bundle returns the existing bundle_id with status `"already_exists"`.

### GET /api/repro-bundles

List repro bundles with filtering.

**Query Parameters:**
- `build_id` - Filter by build ID
- `map_name` - Filter by map name
- `platform` - Filter by platform (Win64, Linux, Android, iOS)
- `since` - ISO8601 timestamp
- `limit` - Max results (default: 50, max: 500)
- `offset` - Pagination offset

**Response:**
```json
{
  "bundles": [
    {
      "bundle_id": "rb_a1b2c3d4e5f6",
      "build_id": "MyGame-1.2.3+456",
      "map_name": "/Game/Maps/Level01",
      "platform": "Win64",
      "created_at": "2026-01-21T10:30:00Z",
      "artifact_count": 5,
      "tags": ["crash", "multiplayer"]
    }
  ],
  "total": 142,
  "limit": 50,
  "offset": 0
}
```

### GET /api/repro-bundles/:bundle_id

Get full bundle details.

**Response:**
```json
{
  "bundle_id": "rb_a1b2c3d4e5f6",
  "content_hash": "sha256:abc123...",
  "build_id": "MyGame-1.2.3+456",
  "map_name": "/Game/Maps/Level01",
  "platform": "Win64",
  "schema_version": "1.0",
  "created_at": "2026-01-21T10:30:00Z",
  "metadata": {
    "player_position": {"x": 100, "y": 200, "z": 50},
    "game_time": 3600.5
  },
  "artifacts": [
    {
      "artifact_id": "art_111",
      "type": "video",
      "filename": "replay.mp4",
      "size_bytes": 52428800,
      "mime_type": "video/mp4"
    },
    {
      "artifact_id": "art_222",
      "type": "log",
      "filename": "game.log",
      "size_bytes": 1024000,
      "mime_type": "text/plain"
    }
  ],
  "tags": ["crash", "multiplayer"],
  "qa_notes": [
    {
      "note_id": "note_001",
      "author": "qa_john",
      "content": "Reproducible 3/5 times",
      "created_at": "2026-01-21T11:00:00Z"
    }
  ]
}
```

### GET /api/repro-bundles/:bundle_id/artifacts/:artifact_id

Download a specific artifact.

**Response:** Raw file with appropriate Content-Type header.

### POST /api/repro-bundles/:bundle_id/tags

Add tags to a bundle.

**Request:**
```json
{
  "tags": ["crash", "priority-high"]
}
```

### POST /api/repro-bundles/:bundle_id/notes

Add a QA note.

**Request:**
```json
{
  "author": "qa_john",
  "content": "Confirmed reproducible on build 457"
}
```

### GET /api/health

Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "database": "ok",
  "storage": "ok"
}
```

---

## Repro Bundle Schema

Expected ZIP structure:

```
repro_bundle.zip
├── manifest.json       # Required: bundle metadata
├── replay.mp4          # Video recording from RVR
├── replay_thumb.jpg    # Optional thumbnail
├── game.log            # Game log file
├── crash.dmp           # Optional crash dump
└── screenshots/        # Optional screenshot directory
    ├── 001.png
    └── 002.png
```

### manifest.json Schema (v1.0)

The manifest can be in either format - the backend normalizes both:

**Unreal Engine Format (from RVR BugIt SDK):**

```json
{
  "schemaVersion": "1.0.0",
  "bundleId": "rb_abc12345",
  "reportTimestampUtc": 1737672000000,
  "mapName": "/Game/Maps/Level01",
  "platform": "WindowsEditor",
  "buildInfo": {
    "buildId": "MyGame++UE5+Release-5.5",
    "rvrVersion": "2.1.0"
  },
  "sessionInfo": {
    "sessionId": "session_xyz",
    "durationMs": 30000
  },
  "hardwareInfo": {
    "cpu": "AMD Ryzen 9",
    "gpu": "NVIDIA RTX 4090",
    "ram": "64GB"
  },
  "artifacts": [
    "video.mp4",
    "inputs.json",
    "timing.json",
    "logs.txt",
    "hardware.json"
  ],
  "customData": {}
}
```

**Standard Format:**

```json
{
  "schema_version": "1.0",
  "build_id": "MyGame-1.2.3+456",
  "map_name": "/Game/Maps/Level01",
  "platform": "Win64",
  "timestamp": "2026-01-21T10:25:00Z",
  "rvr_version": "2.1.0",
  "metadata": {
    "player_position": {"x": 100, "y": 200, "z": 50}
  },
  "artifacts": [
    {
      "filename": "replay.mp4",
      "type": "video",
      "mime_type": "video/mp4"
    }
  ]
}
```

**Schema Version Compatibility:**
- Accepts `1.0`, `1.0.0`, `1.1`, etc. (any version starting with `1.`)
- Platform field accepts any string (e.g., `Win64`, `WindowsEditor`, `Android`, etc.)

---

## Filesystem Layout

```
data/
├── bugit.db                           # SQLite database
├── bundles/
│   ├── rb_a1b2c3d4/
│   │   ├── manifest.json
│   │   ├── replay.mp4
│   │   ├── game.log
│   │   └── screenshots/
│   │       └── 001.png
│   └── rb_e5f6g7h8/
│       └── ...
└── tmp/                               # Temporary upload staging
    └── upload_<uuid>/
```

**Design Decisions:**

1. **Bundle directories use truncated ID** - First 8 chars of bundle_id for directory name, balancing uniqueness with path length
2. **Flat bundle storage** - No date-based partitioning to simplify backup/restore
3. **Artifacts stored as-is** - No renaming or restructuring to maintain forensic integrity
4. **Staging directory** - Uploads written to tmp/ first, then atomically moved on success

---

## Concurrency Model

```
┌────────────────────────────────────────────────────────┐
│                   Upload Request                        │
│                         │                               │
│                         ▼                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 1. Generate upload_id (UUID)                     │  │
│  │ 2. Stream to tmp/upload_<uuid>/                  │  │
│  │    (No locks needed - unique directory)          │  │
│  └──────────────────────────────────────────────────┘  │
│                         │                               │
│                         ▼                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 3. Extract & validate manifest.json              │  │
│  │ 4. Compute SHA256 of entire ZIP                  │  │
│  └──────────────────────────────────────────────────┘  │
│                         │                               │
│                         ▼                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 5. BEGIN TRANSACTION                             │  │
│  │ 6. Check if content_hash exists                  │  │
│  │    - If exists: return existing bundle_id        │  │
│  │    - If new: INSERT bundle row                   │  │
│  │ 7. COMMIT                                        │  │
│  └──────────────────────────────────────────────────┘  │
│                         │                               │
│                         ▼                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 8. If new bundle:                                │  │
│  │    - os.Rename(tmp/upload_xxx, bundles/rb_xxx)   │  │
│  │    (Atomic on same filesystem)                   │  │
│  │ 9. Cleanup tmp/ on success or failure            │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────┘
```

**Key guarantees:**

1. **No file-level locks** - Each upload writes to unique tmp directory
2. **SQLite handles DB concurrency** - WAL mode + IMMEDIATE transactions
3. **Idempotency via content hash** - SHA256 checked inside transaction
4. **Atomic directory placement** - `os.Rename` is atomic on same filesystem
5. **Cleanup on failure** - tmp directories removed if ingestion fails

---

## Error Handling Strategy

### HTTP Error Responses

All errors return JSON:

```json
{
  "error": {
    "code": "INVALID_MANIFEST",
    "message": "manifest.json missing required field: build_id",
    "details": {
      "field": "build_id",
      "location": "manifest.json"
    }
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_MANIFEST` | 400 | manifest.json malformed or missing required fields |
| `UNSUPPORTED_SCHEMA` | 400 | Schema version not supported |
| `INVALID_ZIP` | 400 | ZIP file corrupted or unreadable |
| `BUNDLE_NOT_FOUND` | 404 | Bundle ID does not exist |
| `ARTIFACT_NOT_FOUND` | 404 | Artifact ID does not exist |
| `STORAGE_ERROR` | 500 | Filesystem operation failed |
| `DATABASE_ERROR` | 500 | SQLite operation failed |

### Logging

All operations logged with structured JSON:

```json
{
  "level": "info",
  "ts": "2026-01-21T10:30:00Z",
  "msg": "bundle_ingested",
  "bundle_id": "rb_a1b2c3d4",
  "content_hash": "sha256:abc...",
  "size_bytes": 52428800,
  "duration_ms": 1523
}
```

### Recovery

- **Orphaned tmp directories**: Cleaned up on server start (>1 hour old)
- **Partial uploads**: tmp directory deleted on connection close
- **Database corruption**: SQLite integrity check on startup

---

## CLI Reference

### bugit serve

Start the HTTP server.

```bash
bugit serve [flags]

Flags:
  --port int          HTTP port (default 8080)
  --data-dir string   Data directory path (default "./data")
  --log-level string  Log level: debug, info, warn, error (default "info")
```

### bugit ingest

Ingest a repro bundle from local file.

```bash
bugit ingest <path-to-zip> [flags]

Flags:
  --data-dir string   Data directory path (default "./data")
```

### bugit list

List ingested bundles.

```bash
bugit list [flags]

Flags:
  --data-dir string   Data directory path (default "./data")
  --build-id string   Filter by build ID
  --platform string   Filter by platform
  --limit int         Max results (default 20)
  --json              Output as JSON
```

### bugit inspect

Show bundle details.

```bash
bugit inspect <bundle_id> [flags]

Flags:
  --data-dir string   Data directory path (default "./data")
  --json              Output as JSON
```

---

## Configuration

Environment variables (all optional):

| Variable | Default | Description |
|----------|---------|-------------|
| `BUGIT_PORT` | 8080 | HTTP server port |
| `BUGIT_DATA_DIR` | ./data | Data directory path |
| `BUGIT_LOG_LEVEL` | info | Logging verbosity |
| `BUGIT_MAX_UPLOAD_MB` | 500 | Maximum upload size in MB |

---

## Building

```bash
# Development build
go build -o bugit ./cmd/bugit

# Production build with version
go build -ldflags="-X main.Version=1.0.0" -o bugit ./cmd/bugit

# Docker build
docker build -t bugit:latest .
```

## License

**Non-Commercial Use Only**

For commercial licensing, contact: **business@unrealsolutions.com**

Copyright (c) Unreal Solutions Ltd.
