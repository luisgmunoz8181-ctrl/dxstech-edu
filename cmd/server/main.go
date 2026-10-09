package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"dxstech-edu/internal/ai"
	"dxstech-edu/internal/analytics"
	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/certificates"
	"dxstech-edu/internal/config"
	"dxstech-edu/internal/courses"
	"dxstech-edu/internal/database"
	"dxstech-edu/internal/enrollments"
	"dxstech-edu/internal/quizzes"
	"dxstech-edu/internal/whatsapp"

	"github.com/gin-gonic/gin"
)

type quizEnrollmentCompleter struct {
	enrollSvc *enrollments.Service
}

func (q *quizEnrollmentCompleter) CompleteLesson(ctx context.Context, userID, courseID, lessonID string) error {
	_, err := q.enrollSvc.ToggleLessonProgress(ctx, userID, courseID, lessonID, true)
	return err
}

func main() {
	devFlag := flag.Bool("dev", false, "ejecuta en modo desarrollo (equivale a APP_ENV=development)")
	flag.Parse()

	cfg := config.Load()
	if *devFlag {
		cfg.AppEnv = "development"
	}
	if err := cfg.EnsureJWTSecret(); err != nil {
		log.Fatalf("❌ Configuración de seguridad inválida: %v", err)
	}

	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize Database
	db, err := database.InitDB(cfg.DataDir, cfg.IsDevelopment())
	if err != nil {
		log.Fatalf("❌ Error crítico inicializando SQLite: %v", err)
	}
	defer db.Close()

	// Initialize AI Client
	geminiClient := ai.NewGeminiClient()

	// Initialize Feature Services
	authService := auth.NewService(db, cfg)
	authHandler := auth.NewHandler(authService, !cfg.IsDevelopment())
	courseService := courses.NewService(db, cfg.DataDir)
	courseHandler := courses.NewHandler(courseService)
	certService := certificates.NewService(db)
	enrollService := enrollments.NewService(db, certService)
	enrollHandler := enrollments.NewHandler(enrollService)
	quizCompleter := &quizEnrollmentCompleter{enrollSvc: enrollService}
	quizService := quizzes.NewService(db, geminiClient, quizCompleter)
	waService := whatsapp.NewService(db, geminiClient, cfg)
	analyticsService := analytics.NewService(db)
	analyticsHandler := analytics.NewHandler(analyticsService)

	// Router setup
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(safeLoggerMiddleware())
	r.Use(corsMiddleware(cfg.AllowedOrigins()))
	r.Use(auth.Authenticate(cfg.JWTSecret))

	// Set max multipart memory (25 MB)
	r.MaxMultipartMemory = 25 << 20

	// Rutas amigables de verificación pública de certificados
	r.GET("/verify/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(http.StatusFound, "/#verify/"+id)
	})
	r.GET("/validar/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.Redirect(http.StatusFound, "/#verify/"+id)
	})

	// API Routes Group
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"app":       "DxSTech Edu",
				"version":   "1.0.0",
				"env":       cfg.AppEnv,
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})

		authHandler.RegisterRoutes(api.Group("/auth"))
		courseHandler.RegisterRoutes(api.Group("/courses"))
		enrollHandler.RegisterRoutes(api.Group("/enrollments"))
		certService.RegisterRoutes(api.Group("/certificates"))
		quizService.RegisterRoutes(api.Group("/quizzes"))
		waService.RegisterRoutes(api.Group("/whatsapp"))
		analyticsHandler.RegisterRoutes(api.Group("/admin"))
	}

	// Serve Frontend Static SPA & Uploads
	uploadsDir := filepath.Join(cfg.DataDir, "uploads")
	_ = os.MkdirAll(uploadsDir, 0755)
	r.Static("/uploads", uploadsDir)
	r.Static("/js", "./web/js")
	r.StaticFile("/favicon.ico", "./web/favicon.ico")

	// SPA Fallback: any non-API route serves index.html
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint no encontrado"})
			return
		}
		c.File("./web/index.html")
	})

	srv := &http.Server{
		Addr:         cfg.BindAddress(),
		Handler:      r,
		ReadTimeout:  90 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	go func() {
		fmt.Println("==================================================")
		fmt.Println("🎓 DxSTech Edu — Plataforma Educativa Premium")
		fmt.Printf("🌱 Entorno: %s\n", cfg.AppEnv)
		fmt.Printf("🌐 Servidor listo: %s\n", cfg.PublicURL())
		fmt.Println("==================================================")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Error ejecutando el servidor HTTP: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Apagando el servidor con gracia...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Error forzado en apagado: %v", err)
	}

	log.Println("✅ Servidor detenido correctamente.")
}

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

func safeLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		// Never log API keys or sensitive data
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// Clean logger output
		log.Printf("[GIN] %d | %13v | %15s | %-7s %s",
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}
