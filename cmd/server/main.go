package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/backup"
	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	devFlag := flag.Bool("dev", false, "ejecuta en modo desarrollo (equivale a APP_ENV=development)")
	restoreFlag := flag.String("restore", "", "restaura una copia de seguridad (.tar.gz) en DATA_DIR y termina; el servidor debe estar detenido")
	yesFlag := flag.Bool("yes", false, "confirma la operación (-restore o -rotate-demo-users); sin él solo se muestra qué se haría")
	rotateFlag := flag.Bool("rotate-demo-users", false, "neutraliza las cuentas demo (admin/superadmin/estudiante@dxstech.edu) que aún usan su contraseña pública y termina")
	deactivateFlag := flag.Bool("demo-deactivate", false, "con -rotate-demo-users: desactiva las cuentas demo en lugar de darles una contraseña nueva")
	flag.Parse()

	cfg := config.Load()
	if *devFlag {
		cfg.AppEnv = "development"
	}
	setupLogging(cfg)

	if *restoreFlag != "" {
		runRestore(cfg, *restoreFlag, *yesFlag)
		return
	}
	if *rotateFlag {
		os.Exit(runRotateDemoUsers(cfg, *yesFlag, *deactivateFlag, os.Stdout))
	}

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

	backupSvc := backup.New(db, cfg.DataDir, cfg.BackupDir, cfg.BackupRetention)
	ctx, stopBackups := context.WithCancel(context.Background())
	defer stopBackups()
	if cfg.BackupEnabled {
		backupSvc.StartScheduler(ctx, cfg.BackupInterval)
		log.Printf("💾 Copias de seguridad automáticas cada %s (se conservan %d) en %s", cfg.BackupInterval, cfg.BackupRetention, cfg.BackupDir)
	}

	r, err := newRouter(cfg, db, backupSvc)
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

	stopBackups()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Error forzado en apagado: %v", err)
	}

	log.Println("✅ Servidor detenido correctamente.")
}

// runRestore restaura una copia de seguridad en DATA_DIR (modo CLI, servidor detenido).
func runRestore(cfg *config.Config, archive string, confirmed bool) {
	if !confirmed {
		fmt.Printf("Se restaurará %s en %s.\n", archive, cfg.DataDir)
		fmt.Println("El servidor debe estar DETENIDO. La base de datos y los archivos actuales no se borran:")
		fmt.Println("se mueven a una carpeta pre-restore-<fecha> dentro de DATA_DIR.")
		fmt.Println("Vuelve a ejecutar el comando agregando -yes para confirmar.")
		os.Exit(2)
	}
	if err := backup.Restore(archive, cfg.DataDir); err != nil {
		log.Fatalf("❌ No se pudo restaurar la copia: %v", err)
	}
	fmt.Println("✅ Copia restaurada. Ya puedes iniciar el servidor.")
}

// runRotateDemoUsers neutraliza las cuentas de demostración que todavía usan su contraseña
// pública (modo CLI). Devuelve el código de salida: 0 correcto, 1 error, 2 falta confirmar.
// Las contraseñas nuevas se escriben SOLO en out (no en los logs) y se muestran una vez.
func runRotateDemoUsers(cfg *config.Config, confirmed, deactivate bool, out io.Writer) int {
	db, err := database.OpenExisting(cfg.DataDir)
	if err != nil {
		fmt.Fprintf(out, "❌ %v\n", err)
		return 1
	}
	defer db.Close()

	results, err := auth.RotateDemoUsers(context.Background(), db, auth.DemoRotationOptions{Apply: confirmed, Deactivate: deactivate})
	if err != nil {
		fmt.Fprintf(out, "❌ No se pudo completar la operación (no se cambió nada): %v\n", err)
		return 1
	}

	fmt.Fprintf(out, "Cuentas de demostración en %s:\n\n", cfg.DataDir)
	pending := 0
	for _, r := range results {
		fmt.Fprintf(out, "  • %-26s %s\n", r.Email, r.Outcome)
		if r.Outcome == auth.DemoWouldRotate || r.Outcome == auth.DemoWouldDeact {
			pending++
		}
	}
	fmt.Fprintln(out)

	if !confirmed {
		if pending == 0 {
			fmt.Fprintln(out, "✅ Nada que hacer: ninguna cuenta demo usa ya su contraseña pública.")
			return 0
		}
		fmt.Fprintln(out, "Simulación: no se cambió nada. Detén el servidor y vuelve a ejecutar agregando -yes para aplicarlo.")
		return 2
	}

	rotated := 0
	for _, r := range results {
		if r.Outcome != auth.DemoRotated {
			continue
		}
		if rotated == 0 {
			fmt.Fprintln(out, "🔐 Contraseñas nuevas (se muestran UNA sola vez; guárdalas ahora). Cada cuenta deberá cambiarla al ingresar:")
			fmt.Fprintln(out)
		}
		rotated++
		fmt.Fprintf(out, "  %-26s %s\n", r.Email, r.Password)
	}
	if rotated > 0 {
		fmt.Fprintln(out)
	}
	fmt.Fprintln(out, "✅ Listo. Las sesiones anteriores de estas cuentas quedaron cerradas.")
	return 0
}
