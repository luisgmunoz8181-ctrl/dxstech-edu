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
	// Solo se confía en X-Forwarded-For si la petición llega desde un proxy
	// declarado; así un cliente no puede falsear su IP para evadir el rate limit.
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatalf("❌ TRUSTED_PROXIES inválido: %v", err)
	}
	r.Use(gin.Recovery())
	r.Use(securityHeadersMiddleware())
	r.Use(cacheControlMiddleware())
	r.Use(requestDeadlinesMiddleware())
	// Compresión gzip de JSON/JS/CSS/HTML/CSV. Se excluyen los binarios ya
	// comprimidos o servidos por rangos (video, PDF, PPTX, ZIP, imágenes).
	r.Use(gzip.Gzip(gzip.DefaultCompression,
		gzip.WithExcludedPaths([]string{"/uploads/", "/api/certificates/"}),
		gzip.WithExcludedExtensions([]string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".pdf", ".pptx", ".ppt", ".mp4", ".webm", ".zip", ".woff", ".woff2"}),
	))
	r.Use(safeLoggerMiddleware())
	r.Use(corsMiddleware(cfg.AllowedOrigins()))
	r.Use(auth.Authenticate(cfg.JWTSecret, auth.NewSessionLookup(db)))

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
		api.GET("/health", healthHandler(db, cfg))

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

	// Timeouts por defecto cortos (frente a slowloris); las rutas lentas
	// (subidas, IA, PDFs masivos, reportes) amplían su plazo en requestDeadlinesMiddleware.
	srv := &http.Server{
		Addr:              cfg.BindAddress(),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
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
