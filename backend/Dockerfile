# BugIt Dockerfile
# Multi-stage build for minimal production image

# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO for SQLite
ARG VERSION=dev
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o bugit \
    ./cmd/bugit

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates sqlite-libs

# Create non-root user
RUN adduser -D -g '' bugit
USER bugit

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bugit /app/bugit

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/api/health || exit 1

# Default command
ENTRYPOINT ["/app/bugit"]
CMD ["serve", "--port", "8080", "--data-dir", "/app/data"]
