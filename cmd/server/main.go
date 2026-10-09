package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	devFlag := flag.Bool("dev", false, "ejecuta en modo desarrollo (equivale a APP_ENV=development)")
	flag.Parse()

	cfg := config.Load()
	if *devFlag {
		cfg.AppEnv = "development"
	}
	setupLogging(cfg)

	if err := cfg.EnsureJWTSecret(); err != nil {
		log.Fatalf("❌ Configuración de seguridad inválida: %v", err)
	}

	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	db, err := database.InitDB(cfg.DataDir, cfg.IsDevelopment())
	if err != nil {
		log.Fatalf("❌ Error crítico inicializando SQLite: %v", err)
	}
	defer db.Close()

	r, err := newRouter(cfg, db)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

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
