# BugIt - Forensic QA Dashboard

A video-first forensic QA dashboard for investigating immutable repro bundles produced by [Runtime Video Recorder (RVR)](https://unrealsolutions.ltd/products/runtime-video-recorder).

## Features

- **Video Playback** - Scrub through gameplay recordings with frame-accurate seeking
- **Input Timeline** - Visualize keyboard, mouse, and gamepad inputs synced to video
  - KBD track: Shows key press durations as colored segments
  - Mouse track: Shows click events with button identification
  - Gamepad track: Shows button presses
- **Frame Timing Graph** - Identify FPS drops and stutters
- **Log Panel** - Filter and search engine logs aligned to video time
- **Time Synchronization** - All views stay in sync with video playhead
- **QA Notes** - Add timestamped notes to recordings
- **Artifact Downloads** - Download any bundle artifact (video, logs, JSON data)

## Quick Start

```bash
# Install dependencies
npm install

# Start development server (proxies API to backend on port 8080)
npm run dev

# Build for production
npm run build
```

The dev server runs on `http://localhost:3000` and proxies `/api/*` requests to the backend.

## Requirements

- Node.js 18+
- BugIt backend running on port 8080 (see `../backend/README.md`)

## Tech Stack

- React 18 + TypeScript
- Vite (dev server with API proxy)
- React Router v6
- TanStack Query (React Query)
- CSS Modules

## Project Structure

```
src/
├── api/          # API client and type definitions
├── components/   # Reusable UI components
│   ├── InputTimeline/   # Keyboard/mouse/gamepad visualization
│   ├── VideoPlayer/     # HTML5 video with custom controls
│   ├── ReproCard/       # Bundle list item card
│   └── ...
├── context/      # React context (TimeContext for video sync)
├── pages/        # Route pages
│   ├── ReproListPage    # Bundle list with filters
│   └── ReproViewerPage  # Full bundle viewer
├── types/        # TypeScript type definitions
└── utils/        # Helper functions (time formatting, etc.)
```

## Backend API

The frontend expects the BugIt Go backend at `http://localhost:8080`. In development, Vite proxies `/api/*` requests automatically.

### API Endpoints Used

- `GET /api/repro-bundles` - List bundles with pagination
- `GET /api/repro-bundles/:id` - Get bundle details with artifacts
- `GET /api/repro-bundles/:id/artifacts/:artifact_id` - Download artifact
- `POST /api/repro-bundles/:id/tags` - Add tags to bundle
- `POST /api/repro-bundles/:id/notes` - Add QA note
- `GET /api/health` - Health check

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Space` / `K` | Play/Pause |
| `←` / `→` | Seek ±5s |
| `Shift+←/→` | Seek ±1s |
| `J` / `L` | Seek ±10s |
| `,` / `.` | Previous/Next frame |

## Development

```bash
# Start backend first (in ../backend)
go run ./cmd/bugit serve --port 8080 --data-dir ./data

# Then start frontend dev server
npm run dev

# Run linter
npm run lint

# Preview production build
npm run preview
```

## Configuration

The Vite dev server proxy is configured in `vite.config.ts`:

```typescript
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

For production, configure your reverse proxy (nginx, etc.) to route `/api/*` to the backend.

## License

**Non-Commercial Use Only**

For commercial licensing, contact: **business@unrealsolutions.com**

Copyright (c) Unreal Solutions Ltd.
