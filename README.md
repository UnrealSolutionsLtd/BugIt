# BugIt - QA Dashboard for Unreal Engine Games

**Reduce back-and-forth between QA testers and developers.** BugIt is an on-premises dashboard that turns bug reports into undeniable video evidence - complete with inputs, logs, and frame timings - so developers can reproduce issues on the first try.

- **Built for Unreal Engine** - Native integration with UE5 games
- **QA-first workflow** - Video evidence eliminates "works on my machine" disputes
- **Powered by [RVR](https://unrealsolutions.com)** - In-engine capture creates repro bundles automatically. Works for any game

![BugIt Dashboard](docs/images/screenshot.png)

## Features

- **Video-first investigation** - Scrub through gameplay recordings with synced timelines
- **Input visualization** - See keyboard, mouse, and gamepad inputs aligned to video
- **Frame timing analysis** - Identify FPS drops and performance stutters
- **Log correlation** - Filter and search engine logs synced to video time
- **QA notes** - Add timestamped markdown notes to recordings
- **Immutable bundles** - SHA256 verified repro bundles for forensic integrity

## Architecture

```
BugIt/
├── frontend/       # React + TypeScript + Vite
├── backend/        # Go + SQLite
├── Dockerfile      # Combined multi-stage build
├── docker-compose.yml
└── DEPLOYMENT.md   # Full deployment guide
```

Single container serves both frontend static files and backend API.

## Quick Start

### Docker (Recommended)

```bash
# Build and run
docker compose up -d

# Access dashboard
open http://localhost:8080
```

### Development

```bash
# Terminal 1: Backend
cd backend
go run ./cmd/bugit serve --port 8080 --data-dir ./data

# Terminal 2: Frontend
cd frontend
npm install
npm run dev
```

Frontend dev server proxies `/api/*` to backend on port 8080.

## Deployment

See [DEPLOYMENT.md](./DEPLOYMENT.md) for:

- Production Docker deployment
- Network/firewall configuration
- Backup & recovery
- Storage sizing
- Monitoring

## API Overview

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/repro-bundles` | POST | Upload new repro bundle |
| `/api/repro-bundles` | GET | List bundles with filters |
| `/api/repro-bundles/:id` | GET | Get bundle details |
| `/api/repro-bundles/:id/artifacts/:aid` | GET | Download artifact |
| `/api/repro-bundles/:id/notes` | POST | Add QA note |

See [backend/README.md](./backend/README.md) for full API documentation.

## Integration

### Game Integration (UE5)

```cpp
UBugItCaptureSubsystem* BugIt = GetGameInstance()->GetSubsystem<UBugItCaptureSubsystem>();
BugIt->SetEndpointURL(TEXT("http://bugit-server:8080/api/repro-bundles"));
BugIt->CaptureAndUpload();
```

### CI/CD Integration

```bash
curl -X POST \
  -F "bundle=@./repro_bundle.zip" \
  http://bugit-server:8080/api/repro-bundles
```

## Security

BugIt is designed for **trusted internal networks only**:

- No authentication (assumes network-level access control)
- No encryption at rest (use full-disk encryption if needed)
- Block external access via firewall

Do NOT expose to the public internet.

## License

**Non-Commercial Use Only**

This software is provided for non-commercial purposes only. You may use, copy, and modify this software for personal projects, educational purposes, and internal evaluation.

For commercial licensing, please contact: **business@unrealsolutions.com**

Copyright (c) Unreal Solutions Ltd. All rights reserved.
