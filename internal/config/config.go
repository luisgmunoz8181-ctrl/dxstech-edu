package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// MinJWTSecretLength es la longitud mínima exigida al secreto JWT en producción.
const MinJWTSecretLength = 32

type Config struct {
	AppEnv    string
	Port      string
	Host      string
	DataDir   string
	JWTSecret string
	AppURL    string

	// CORSOrigins lista los orígenes adicionales autorizados a llamar la API
	// con credenciales. La SPA se sirve desde el mismo origen, por lo que por
	// defecto solo se permite AppURL.
	CORSOrigins []string

	// TrustedProxies son las IP/CIDR de proxies inversos cuyas cabeceras
	// X-Forwarded-* se aceptan. Por defecto, los rangos privados/loopback
	// (Docker, Coolify, Render). Con "none" no se confía en ninguno.
	TrustedProxies []string

	// Copias de seguridad automáticas (SQLite + archivos subidos).
	BackupEnabled   bool
	BackupDir       string
	BackupInterval  time.Duration
	BackupRetention int
}

var defaultTrustedProxies = []string{"127.0.0.1/8", "::1/128", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fc00::/7"}

func Load() *Config {
	env := os.Getenv("APP_ENV")
	if strings.TrimSpace(env) == "" {
		// Seguro por defecto: si no se declara el entorno, se asume producción.
		env = "production"
	}

	port := os.Getenv("PORT")
	if strings.TrimSpace(port) == "" {
		port = "3000"
	}

	host := os.Getenv("HOST")
	if strings.TrimSpace(host) == "" {
		host = "0.0.0.0"
	}

	dataDir := os.Getenv("DATA_DIR")
	if strings.TrimSpace(dataDir) == "" {
		dataDir = "./data"
	}

	// JWT_SECRET no tiene valor por defecto. En desarrollo, main genera uno
	// aleatorio efímero (ver EnsureJWTSecret); en producción es obligatorio.
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	appURL := os.Getenv("APP_URL")
	if strings.TrimSpace(appURL) == "" {
		appURL = "http://localhost:" + port
	}

	var corsOrigins []string
	for _, o := range strings.Split(os.Getenv("CORS_ORIGINS"), ",") {
		if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
			corsOrigins = append(corsOrigins, o)
		}
	}

	trustedProxies := defaultTrustedProxies
	if raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES")); raw != "" {
		trustedProxies = nil
		if strings.ToLower(raw) != "none" {
			for _, p := range strings.Split(raw, ",") {
				if p = strings.TrimSpace(p); p != "" {
					trustedProxies = append(trustedProxies, p)
				}
			}
		}
	}

	backupDir := strings.TrimSpace(os.Getenv("BACKUP_DIR"))
	if backupDir == "" {
		backupDir = filepath.Join(dataDir, "backups")
	}
	backupEnabled := !strings.EqualFold(strings.TrimSpace(os.Getenv("BACKUP_ENABLED")), "false")
	backupHours := atoiDefault(os.Getenv("BACKUP_INTERVAL_HOURS"), 24)
	if backupHours < 1 {
		backupHours = 24
	}
	backupRetention := atoiDefault(os.Getenv("BACKUP_RETENTION"), 7)
	if backupRetention < 1 {
		backupRetention = 7
	}

	return &Config{
		BackupEnabled:   backupEnabled,
		BackupDir:       backupDir,
		BackupInterval:  time.Duration(backupHours) * time.Hour,
		BackupRetention: backupRetention,
		TrustedProxies:  trustedProxies,
		AppEnv:          env,
		Port:            port,
		Host:            host,
		DataDir:         dataDir,
		JWTSecret:       jwtSecret,
		AppURL:          strings.TrimRight(appURL, "/"),
		CORSOrigins:     corsOrigins,
	}
}

// EnsureJWTSecret deja la configuración lista para firmar tokens.
// En producción exige un JWT_SECRET explícito y robusto. En desarrollo, si no
// hay uno, genera un secreto aleatorio efímero (las sesiones se invalidan al
// reiniciar el servidor).
func (c *Config) EnsureJWTSecret() error {
	if c.IsDevelopment() {
		if c.JWTSecret == "" {
			b := make([]byte, 32)
			if _, err := rand.Read(b); err != nil {
				return err
			}
			c.JWTSecret = hex.EncodeToString(b)
		}
		return nil
	}

	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET es obligatorio en producción (genere uno con: openssl rand -hex 32)")
	}
	if len(c.JWTSecret) < MinJWTSecretLength {
		return errors.New("JWT_SECRET debe tener al menos 32 caracteres en producción")
	}
	return nil
}

// AllowedOrigins devuelve los orígenes autorizados para CORS con credenciales.
func (c *Config) AllowedOrigins() []string {
	origins := make([]string, 0, len(c.CORSOrigins)+1)
	if c.AppURL != "" {
		origins = append(origins, c.AppURL)
	}
	return append(origins, c.CORSOrigins...)
}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return n
	}
	return def
}

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.AppEnv) == "development"
}

func (c *Config) PublicURL() string {
	return "http://localhost:" + c.Port
}

func (c *Config) BindAddress() string {
	return c.Host + ":" + c.Port
}
