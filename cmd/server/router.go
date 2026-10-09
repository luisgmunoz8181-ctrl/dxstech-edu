package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"dxstech-edu/internal/ai"
	"dxstech-edu/internal/analytics"
	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/backup"
	"dxstech-edu/internal/certificates"
	"dxstech-edu/internal/config"
	"dxstech-edu/internal/courses"
	"dxstech-edu/internal/database"
	"dxstech-edu/internal/enrollments"
	"dxstech-edu/internal/quizzes"
	"dxstech-edu/internal/whatsapp"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type quizEnrollmentCompleter struct {
	enrollSvc *enrollments.Service
}

func (q *quizEnrollmentCompleter) CompleteLesson(ctx context.Context, userID, courseID, lessonID string) error {
	_, err := q.enrollSvc.ToggleLessonProgress(ctx, userID, courseID, lessonID, true)
	return err
}

// newRouter construye el servidor HTTP completo (middlewares, API y SPA). Está
// separado de main para poder probarlo de extremo a extremo.
func newRouter(cfg *config.Config, db *database.DB, backupSvc *backup.Service) (*gin.Engine, error) {
	geminiClient := ai.NewGeminiClient()

	authService := auth.NewService(db, cfg)
	authHandler := auth.NewHandler(authService, !cfg.IsDevelopment())
	courseService := courses.NewService(db, cfg.DataDir)
	courseHandler := courses.NewHandler(courseService)
	certService := certificates.NewService(db)
	enrollService := enrollments.NewService(db, certService)
	enrollHandler := enrollments.NewHandler(enrollService)
	quizService := quizzes.NewService(db, geminiClient, &quizEnrollmentCompleter{enrollSvc: enrollService})
	waService := whatsapp.NewService(db, geminiClient, cfg)
	analyticsHandler := analytics.NewHandler(analytics.NewService(db))

	r := gin.New()
	// Solo se confía en X-Forwarded-For si la petición llega desde un proxy
	// declarado; así un cliente no puede falsear su IP para evadir el rate limit.
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, fmt.Errorf("TRUSTED_PROXIES inválido: %w", err)
	}
	r.Use(requestIDMiddleware())
	r.Use(recoveryMiddleware())
	r.Use(accessLogMiddleware())
	r.Use(securityHeadersMiddleware())
	r.Use(cacheControlMiddleware())
	r.Use(requestDeadlinesMiddleware())
	// Compresión gzip de JSON/JS/CSS/HTML/CSV. Se excluyen los binarios ya
	// comprimidos o servidos por rangos (video, PDF, PPTX, ZIP, imágenes).
	r.Use(gzip.Gzip(gzip.DefaultCompression,
		gzip.WithExcludedPaths([]string{"/uploads/", "/api/certificates/", "/api/admin/backups/"}),
		gzip.WithExcludedExtensions([]string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".pdf", ".pptx", ".ppt", ".mp4", ".webm", ".zip", ".woff", ".woff2"}),
	))
	r.Use(corsMiddleware(cfg.AllowedOrigins()))
	r.Use(auth.Authenticate(cfg.JWTSecret, auth.NewSessionLookup(db)))

	// Set max multipart memory (25 MB)
	r.MaxMultipartMemory = 25 << 20

	// Rutas amigables de verificación pública de certificados
	r.GET("/verify/:id", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/#verify/"+c.Param("id"))
	})
	r.GET("/validar/:id", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/#verify/"+c.Param("id"))
	})

	api := r.Group("/api")
	{
		api.GET("/health", healthHandler(db, cfg))

		authHandler.RegisterRoutes(api.Group("/auth"))
		courseHandler.RegisterRoutes(api.Group("/courses"))
		enrollHandler.RegisterRoutes(api.Group("/enrollments"))
		certService.RegisterRoutes(api.Group("/certificates"))
		quizService.RegisterRoutes(api.Group("/quizzes"))
		waService.RegisterRoutes(api.Group("/whatsapp"))
		analyticsHandler.RegisterRoutes(api.Group("/admin"))
		backup.NewHandler(backupSvc, db).RegisterRoutes(api.Group("/admin/backups"))
	}

	// Serve Frontend Static SPA & Uploads
	uploadsDir := filepath.Join(cfg.DataDir, "uploads")
	_ = os.MkdirAll(uploadsDir, 0755)
	// nosniff (global) + extensiones/magic bytes validados al subir evitan que un archivo se interprete como HTML/JS.
	r.Static("/uploads", uploadsDir)
	r.Static("/js", "./web/js")
	r.Static("/css", "./web/css")
	r.Static("/vendor", "./web/vendor")
	r.StaticFile("/favicon.ico", "./web/favicon.ico")

	// SPA Fallback: any non-API route serves index.html
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint no encontrado"})
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.File("./web/index.html")
	})

	return r, nil
}
