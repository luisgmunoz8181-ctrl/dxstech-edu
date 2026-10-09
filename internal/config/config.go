package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
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
}

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

	return &Config{
		AppEnv:      env,
		Port:        port,
		Host:        host,
		DataDir:     dataDir,
		JWTSecret:   jwtSecret,
		AppURL:      strings.TrimRight(appURL, "/"),
		CORSOrigins: corsOrigins,
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

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.AppEnv) == "development"
}

func (c *Config) PublicURL() string {
	return "http://localhost:" + c.Port
}

func (c *Config) BindAddress() string {
	return c.Host + ":" + c.Port
}
