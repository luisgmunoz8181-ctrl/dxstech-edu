package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

// corsMiddleware solo habilita CORS con credenciales para los orígenes
// explícitamente autorizados. La SPA se sirve desde el mismo origen que la API,
// por lo que las peticiones same-origin no necesitan estas cabeceras.
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Add("Vary", "Origin")
		}

		if origin != "" && allowed[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Gemini-API-Key")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// securityHeadersMiddleware añade cabeceras defensivas a todas las respuestas.
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// healthHandler comprueba que la base de datos responda; con 503 los
// healthchecks de Docker/Coolify/Render reinician o retiran la instancia.
func healthHandler(db *database.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		dbStatus, code := "ok", http.StatusOK
		var one int
		if err := db.PingContext(ctx); err != nil || db.QueryRowContext(ctx, "SELECT 1").Scan(&one) != nil {
			dbStatus, code = "error", http.StatusServiceUnavailable
		}

		status := "ok"
		if code != http.StatusOK {
			status = "degraded"
		}
		c.JSON(code, gin.H{
			"status":    status,
			"database":  dbStatus,
			"app":       "DxSTech Edu",
			"version":   "1.0.0",
			"env":       cfg.AppEnv,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// cacheControlMiddleware fija políticas de caché por tipo de recurso. Los
// handlers pueden sobrescribirla (p. ej. los PDF de certificados).
func cacheControlMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		switch {
		case strings.HasPrefix(p, "/api/"):
			c.Header("Cache-Control", "no-store")
		case strings.HasPrefix(p, "/vendor/"):
			// Archivos con versión en el nombre: no cambian nunca.
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		case strings.HasPrefix(p, "/uploads/"):
			// Nombres aleatorios únicos: el contenido de una URL nunca cambia.
			c.Header("Cache-Control", "private, max-age=31536000, immutable")
		case strings.HasPrefix(p, "/js/"), strings.HasPrefix(p, "/css/"):
			// Sin hash en el nombre: siempre se revalida (304 por Last-Modified).
			c.Header("Cache-Control", "no-cache")
		}
		c.Next()
	}
}

// slowRoute amplía los plazos de lectura/escritura de una ruta concreta.
type slowRoute struct {
	method, path string
	read, write  time.Duration
}

var slowRoutes = []slowRoute{
	{"POST", "/api/courses/upload", 5 * time.Minute, time.Minute},
	{"POST", "/api/quizzes/generate", time.Minute, 4 * time.Minute},
	{"POST", "/api/quizzes/generate-for-lesson", time.Minute, 4 * time.Minute},
	{"POST", "/api/whatsapp/ask-tutor", time.Minute, 4 * time.Minute},
	{"POST", "/api/whatsapp/simulate-chat", time.Minute, 4 * time.Minute},
	{"POST", "/api/certificates/generate", 2 * time.Minute, 3 * time.Minute},
}

// requestDeadlinesMiddleware aplica plazos más largos solo a las rutas lentas
// (subidas, IA, generación masiva de PDF y reportes); el resto conserva los
// timeouts cortos del servidor.
func requestDeadlinesMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		read, write := time.Duration(0), time.Duration(0)
		for _, r := range slowRoutes {
			if c.Request.Method == r.method && c.Request.URL.Path == r.path {
				read, write = r.read, r.write
				break
			}
		}
		if read == 0 && strings.HasPrefix(c.Request.URL.Path, "/api/admin/reports/") {
			read, write = 30*time.Second, 2*time.Minute
		}
		if read > 0 {
			rc := http.NewResponseController(c.Writer)
			now := time.Now()
			_ = rc.SetReadDeadline(now.Add(read))
			_ = rc.SetWriteDeadline(now.Add(write))
		}
		c.Next()
	}
}
