# ==============================================================================
# Stage 1: Build static Go binary
# ==============================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Install build dependencies
RUN apk add --no-cache ca-certificates tzdata git

# Cache module dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Compile completely static CGO-free binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -extldflags '-static'" \
    -trimpath \
    -o /src/dxstech-server ./cmd/server

# ==============================================================================
# Stage 2: Minimal, secure production runtime
# ==============================================================================
FROM alpine:3.21

WORKDIR /app

# Install SSL root certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Create directory for persistent SQLite storage
RUN mkdir -p /app/data

# Copy compiled binary from builder
COPY --from=builder /src/dxstech-server /app/dxstech-server

# Copy SPA frontend static assets (HTML, JavaScript ES Modules, Assets)
COPY web/ /app/web/

# Configure production environment
ENV APP_ENV=production \
    PORT=3000 \
    HOST=0.0.0.0 \
    DATA_DIR=/app/data

# Persistent storage volume for SQLite
VOLUME ["/app/data"]

# Expose internal HTTP service port
EXPOSE 3000

# Healthcheck endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:3000/api/health || exit 1

# Execute server
CMD ["/app/dxstech-server"]
