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

# Install SSL root certificates, timezone data and su-exec (cesión de privilegios)
RUN apk add --no-cache ca-certificates tzdata su-exec

# Usuario sin privilegios que ejecuta la aplicación
RUN addgroup -S -g 10001 app && adduser -S -H -u 10001 -G app -h /app/data app

# /app (binario y frontend) pertenece a root y es de solo lectura para la aplicación;
# solo /app/data (SQLite y archivos subidos) es escribible por el usuario `app`.
RUN mkdir -p /app/data && chown root:root /app && chmod 755 /app && chown app:app /app/data

# Copy compiled binary from builder
COPY --from=builder /src/dxstech-server /app/dxstech-server

# Copy SPA frontend static assets (HTML, JavaScript ES Modules, Assets)
COPY web/ /app/web/

# Entrypoint: corrige permisos del volumen y arranca como usuario no-root
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Configure production environment
ENV APP_ENV=production \
    PORT=3000 \
    HOST=0.0.0.0 \
    DATA_DIR=/app/data

# Persistent storage volume for SQLite
VOLUME ["/app/data"]

# Expose internal HTTP service port
EXPOSE 3000

# Healthcheck endpoint (responde 503 si la base de datos no está disponible)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:${PORT:-3000}/api/health || exit 1

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["/app/dxstech-server"]
