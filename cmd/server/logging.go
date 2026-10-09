package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"time"

	"dxstech-edu/internal/config"

	"github.com/gin-gonic/gin"
)

// setupLogging configura slog como logger por defecto: JSON en producción (apto
// para agregadores de logs) y texto legible en desarrollo. Redirige también el
// paquete log estándar, por lo que los log.Printf existentes salen con el mismo formato.
func setupLogging(cfg *config.Config) {
	var h slog.Handler
	if cfg.IsDevelopment() {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(h))
}

const requestIDHeader = "X-Request-ID"

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)

// requestIDMiddleware asigna un identificador a cada petición (respetando uno
// válido enviado por el proxy) y lo devuelve en la cabecera X-Request-ID.
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if !validRequestID.MatchString(id) {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set("requestID", id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}

// accessLogMiddleware registra cada petición en una línea estructurada con
// request-id y usuario autenticado. Nunca se registra la query string (puede
// contener tokens o datos personales) ni cabeceras como X-Gemini-API-Key.
func accessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		// Los healthchecks correctos (cada 30 s) solo añaden ruido.
		if c.Request.URL.Path == "/api/health" && status == http.StatusOK {
			return
		}

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		attrs := []any{
			"request_id", c.GetString("requestID"),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", time.Since(start).Milliseconds(),
			"bytes", c.Writer.Size(),
			"ip", c.ClientIP(),
		}
		if route := c.FullPath(); route != "" {
			attrs = append(attrs, "route", route)
		}
		if uid := c.GetString("userID"); uid != "" {
			attrs = append(attrs, "user_id", uid, "role", c.GetString("userRole"))
		}
		slog.Log(c.Request.Context(), level, "http_request", attrs...)
	}
}

// recoveryMiddleware captura los panics, los registra con su stack y el
// request-id, y responde 500 sin filtrar detalles internos.
func recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, err any) {
		slog.Error("panic recuperado",
			"request_id", c.GetString("requestID"),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"panic", err,
			"stack", string(debug.Stack()),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Error interno del servidor",
			"request_id": c.GetString("requestID"),
		})
	})
}
