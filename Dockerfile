# BugIt - Combined Frontend + Backend Docker Build
# Multi-stage build for minimal production image

# ============================================================
# Stage 1: Build Frontend
# ============================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

# Install dependencies first (cache layer)
COPY frontend/package*.json ./
RUN npm ci --no-audit --no-fund

# Copy source and build
COPY frontend/ ./
RUN npm run build

# ============================================================
# Stage 2: Build Backend
# ============================================================
FROM golang:1.22-alpine AS backend-builder

# Install build dependencies for CGO (SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files and download dependencies (cache layer)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./

# Build with CGO for SQLite support
ARG VERSION=dev
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o bugit \
    ./cmd/bugit

# ============================================================
# Stage 3: Runtime Image
# ============================================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs wget

# Create non-root user for security
RUN addgroup -g 1000 bugit && \
    adduser -D -u 1000 -G bugit bugit

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/bugit /app/bugit

# Copy frontend static files
COPY --from=frontend-builder /frontend/dist /app/static

# Create data directory with correct permissions
RUN mkdir -p /app/data && chown -R bugit:bugit /app/data

# Switch to non-root user
USER bugit

# Expose HTTP port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/api/health || exit 1

# Labels for container metadata
LABEL org.opencontainers.image.title="BugIt" \
      org.opencontainers.image.description="Forensic QA Dashboard for Runtime Video Recorder" \
      org.opencontainers.image.vendor="Unreal Solutions Ltd"

# Default command - serve with frontend static files
ENTRYPOINT ["/app/bugit"]
CMD ["serve", "--port", "8080", "--data-dir", "/app/data", "--static-dir", "/app/static"]
