# BugIt Deployment Guide

On-premises deployment of the BugIt forensic QA dashboard.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           QA Network (Trusted)                           │
│                                                                          │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                     │
│   │  QA Tester  │  │  QA Tester  │  │  Developer  │                     │
│   │  (Browser)  │  │  (Browser)  │  │  (Browser)  │                     │
│   └──────┬──────┘  └──────┬──────┘  └──────┬──────┘                     │
│          │                │                │                             │
│          └────────────────┼────────────────┘                             │
│                           │                                              │
│                           ▼                                              │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                       BugIt Server                               │   │
│   │                     (Single Container)                           │   │
│   │                                                                  │   │
│   │   ┌────────────────┐    ┌────────────────────────────────────┐  │   │
│   │   │   Go Backend   │    │           Static Files             │  │   │
│   │   │                │    │                                    │  │   │
│   │   │  /api/*        │    │  /              → index.html       │  │   │
│   │   │  /api/health   │    │  /assets/*      → JS/CSS           │  │   │
│   │   │                │    │  /repro/:id     → index.html (SPA) │  │   │
│   │   └───────┬────────┘    └────────────────────────────────────┘  │   │
│   │           │                                                      │   │
│   │           ▼                                                      │   │
│   │   ┌────────────────┐    ┌────────────────────────────────────┐  │   │
│   │   │    SQLite      │    │        Artifact Storage            │  │   │
│   │   │   (metadata)   │    │   (videos, logs, screenshots)      │  │   │
│   │   └────────────────┘    └────────────────────────────────────┘  │   │
│   │                                                                  │   │
│   │   Volume: /app/data                                              │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                           │                                              │
│                           │ Port 8080                                    │
│                           │                                              │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                    Game Build Machines                           │   │
│   │                                                                  │   │
│   │   POST /api/repro-bundles ← Automated upload from RVR SDK       │   │
│   │                                                                  │   │
│   └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Deployment Options

### Option 1: Docker (Recommended)

Single container bundles frontend + backend. Simplest production deployment.

```bash
# Build the container
docker build -t bugit:latest .

# Run with persistent data
docker run -d \
  --name bugit \
  -p 8080:8080 \
  -v bugit-data:/app/data \
  --restart unless-stopped \
  bugit:latest
```

**Access:** `http://<server-ip>:8080`

### Option 2: Docker Compose

For environments that prefer compose:

```bash
cd /opt/bugit
docker compose up -d
```

### Option 3: Binary + Reverse Proxy

For environments where Docker isn't available:

```bash
# Build from source
cd backend && go build -o bugit ./cmd/bugit
cd frontend && npm run build

# Run server (serves both API and static files)
./bugit serve \
  --port 8080 \
  --data-dir /var/lib/bugit \
  --static-dir ../frontend/dist
```

Then configure nginx/caddy as reverse proxy if needed.

---

## Quick Start (Docker)

```bash
# 1. Clone and build
git clone <repo> /opt/bugit
cd /opt/bugit
docker build -t bugit:latest .

# 2. Create data directory (optional - Docker volume works too)
mkdir -p /var/lib/bugit

# 3. Run
docker run -d \
  --name bugit \
  -p 8080:8080 \
  -v /var/lib/bugit:/app/data \
  --restart unless-stopped \
  bugit:latest

# 4. Verify
curl http://localhost:8080/api/health
```

---

## Build Process

### Frontend Build

```bash
cd frontend
npm ci
npm run build
# Output: frontend/dist/
```

### Backend Build

```bash
cd backend
go build -ldflags="-s -w -X main.Version=$(git describe --tags)" \
  -o bugit ./cmd/bugit
```

### Combined Docker Build

The Dockerfile handles both:

```dockerfile
# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend-builder
# ... (compiles Go binary)

# Stage 3: Runtime
FROM alpine:3.19
COPY --from=frontend-builder /frontend/dist /app/static
COPY --from=backend-builder /app/bugit /app/bugit
# Go binary serves static files from /app/static
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BUGIT_PORT` | 8080 | HTTP server port |
| `BUGIT_DATA_DIR` | /app/data | Data directory |
| `BUGIT_STATIC_DIR` | /app/static | Frontend static files |
| `BUGIT_LOG_LEVEL` | info | debug, info, warn, error |
| `BUGIT_MAX_UPLOAD_MB` | 500 | Max upload size (MB) |

### Data Directory Structure

```
/app/data/
├── bugit.db              # SQLite database
├── bundles/              # Repro bundle artifacts
│   ├── rb_a1b2c3d4/
│   │   ├── manifest.json
│   │   ├── replay.mp4
│   │   └── game.log
│   └── rb_e5f6g7h8/
│       └── ...
└── tmp/                  # Upload staging
```

---

## Network Requirements

### Ports

| Port | Protocol | Direction | Purpose |
|------|----------|-----------|---------|
| 8080 | TCP | Inbound | Web UI + API |

### Firewall Rules

BugIt should only be accessible from the internal QA network:

```bash
# Example: iptables
iptables -A INPUT -p tcp --dport 8080 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP

# Example: UFW
ufw allow from 10.0.0.0/8 to any port 8080
ufw deny 8080
```

### No Authentication Required

BugIt assumes a trusted network. Do NOT expose to the internet.

If you need external access:
1. Use a VPN
2. Or add a reverse proxy with authentication (nginx + htpasswd, Caddy + basicauth)

---

## Storage Sizing

### Estimate Formula

```
Storage = (avg_bundle_size_mb) × (bundles_per_day) × (retention_days)
```

### Example

| Scenario | Bundle Size | Bundles/Day | Retention | Storage Needed |
|----------|-------------|-------------|-----------|----------------|
| Small QA team | 50 MB | 10 | 90 days | ~45 GB |
| Medium QA team | 100 MB | 50 | 90 days | ~450 GB |
| Large QA team | 100 MB | 200 | 30 days | ~600 GB |

### Recommendations

- **Minimum:** 100 GB SSD
- **Recommended:** 500 GB - 1 TB SSD
- **Database:** SQLite file stays small (~100 MB for 10K bundles)

---

## Backup & Recovery

### Backup Strategy

```bash
# Stop container (ensures consistent SQLite backup)
docker stop bugit

# Backup data directory
tar -czvf bugit-backup-$(date +%Y%m%d).tar.gz /var/lib/bugit

# Restart
docker start bugit
```

### Automated Backup Script

```bash
#!/bin/bash
# /etc/cron.daily/bugit-backup

BACKUP_DIR=/backups/bugit
DATA_DIR=/var/lib/bugit
RETENTION_DAYS=7

# Create backup
docker stop bugit
tar -czvf "$BACKUP_DIR/bugit-$(date +%Y%m%d).tar.gz" "$DATA_DIR"
docker start bugit

# Cleanup old backups
find "$BACKUP_DIR" -name "bugit-*.tar.gz" -mtime +$RETENTION_DAYS -delete
```

### Recovery

```bash
# Stop container
docker stop bugit

# Restore data
rm -rf /var/lib/bugit/*
tar -xzvf bugit-backup-20260121.tar.gz -C /

# Restart
docker start bugit
```

---

## Monitoring

### Health Check

```bash
# HTTP health endpoint
curl http://localhost:8080/api/health

# Response
{
  "status": "ok",
  "version": "1.2.0",
  "database": "ok",
  "storage": "ok",
  "disk_free_gb": 423.5
}
```

### Docker Health

The container includes a healthcheck:

```bash
docker inspect --format='{{.State.Health.Status}}' bugit
# healthy | unhealthy | starting
```

### Prometheus Metrics (Future)

Planned endpoint: `GET /api/metrics`

---

## Upgrading

### Standard Upgrade

```bash
# Pull latest code
cd /opt/bugit
git pull

# Rebuild
docker build -t bugit:latest .

# Replace container
docker stop bugit
docker rm bugit
docker run -d \
  --name bugit \
  -p 8080:8080 \
  -v /var/lib/bugit:/app/data \
  --restart unless-stopped \
  bugit:latest
```

### Zero-Downtime (Optional)

For teams that need continuous availability:

```bash
# Build new image
docker build -t bugit:new .

# Start new container on different port
docker run -d --name bugit-new -p 8081:8080 -v /var/lib/bugit:/app/data bugit:new

# Test new version
curl http://localhost:8081/api/health

# Switch traffic (update load balancer or iptables)
# Then cleanup old container
docker stop bugit && docker rm bugit
docker rename bugit-new bugit
```

---

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker logs bugit

# Common issues:
# - Port already in use: Change -p flag
# - Permission denied on data dir: Check volume ownership
```

### Database Locked

SQLite may report "database is locked" under high load:

```bash
# Check for stuck processes
docker exec bugit fuser /app/data/bugit.db

# Restart container to clear locks
docker restart bugit
```

### Slow Uploads

Large video files may timeout:

1. Increase nginx/proxy timeout if using reverse proxy
2. Check disk I/O: `docker exec bugit iostat -x 1`
3. Consider SSD storage for data directory

### Disk Full

```bash
# Check disk usage
docker exec bugit df -h /app/data

# Find large bundles
docker exec bugit du -sh /app/data/bundles/* | sort -h | tail -20

# Delete old bundles via API or CLI
docker exec bugit ./bugit delete rb_old123
```

---

## Security Checklist

- [ ] BugIt only accessible from internal network
- [ ] Firewall rules block external access
- [ ] Data directory on dedicated partition (prevents system disk full)
- [ ] Regular backups configured
- [ ] Container runs as non-root user (default in Dockerfile)
- [ ] No sensitive data in environment variables (none required)

---

## Integration with RVR

### Game Integration

Configure Runtime Video Recorder to upload bundles on bug report:

```cpp
// In your game code
UBugItCaptureSubsystem* BugIt = GetGameInstance()->GetSubsystem<UBugItCaptureSubsystem>();
BugIt->SetEndpointURL(TEXT("http://bugit-server:8080/api/repro-bundles"));
BugIt->CaptureAndUpload();
```

### CI/CD Integration

Upload test artifacts from automated builds:

```bash
# Option 1: ZIP file upload
curl -X POST \
  -H "Content-Type: application/zip" \
  --data-binary @./test_artifacts/repro_bundle.zip \
  http://bugit-server:8080/api/repro-bundles

# Option 2: Multipart form data (same as Unreal SDK)
curl -X POST \
  -F "manifest.json=@./bundle/manifest.json" \
  -F "video.mp4=@./bundle/video.mp4" \
  -F "logs.txt=@./bundle/logs.txt" \
  http://bugit-server:8080/api/repro-bundles
```

---

## Support & License

**Non-Commercial Use Only**

This software is provided for non-commercial purposes. For commercial licensing, contact: **business@unrealsolutions.com**

For technical issues, see the [GitHub Issues](https://github.com/unrealsolutions/bugit/issues) or contact support@unrealsolutions.com.

Copyright (c) Unreal Solutions Ltd.
