package config_test

import (
	"strings"
	"testing"

	"dxstech-edu/internal/config"
)

func TestLoadDefaultsToProduction(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("JWT_SECRET", "")
	cfg := config.Load()
	if cfg.IsDevelopment() {
		t.Fatalf("sin APP_ENV se esperaba producción, obtuvo %q", cfg.AppEnv)
	}
	if cfg.JWTSecret != "" {
		t.Fatalf("no debe existir un JWT_SECRET por defecto")
	}
}

func TestEnsureJWTSecretProduction(t *testing.T) {
	cases := []struct {
		name, secret string
		wantErr      bool
	}{
		{"vacío", "", true},
		{"corto", "abc123", true},
		{"válido", strings.Repeat("a", config.MinJWTSecretLength), false},
	}
	for _, tc := range cases {
		cfg := &config.Config{AppEnv: "production", JWTSecret: tc.secret}
		if err := cfg.EnsureJWTSecret(); (err != nil) != tc.wantErr {
			t.Errorf("%s: err=%v, wantErr=%v", tc.name, err, tc.wantErr)
		}
	}
}

func TestEnsureJWTSecretDevelopmentGeneratesRandom(t *testing.T) {
	a := &config.Config{AppEnv: "development"}
	b := &config.Config{AppEnv: "development"}
	if err := a.EnsureJWTSecret(); err != nil {
		t.Fatal(err)
	}
	_ = b.EnsureJWTSecret()
	if len(a.JWTSecret) < config.MinJWTSecretLength || a.JWTSecret == b.JWTSecret {
		t.Fatalf("se esperaba un secreto aleatorio y único en desarrollo")
	}
}

func TestAllowedOrigins(t *testing.T) {
	t.Setenv("APP_URL", "https://edu.example.com/")
	t.Setenv("CORS_ORIGINS", "https://a.example.com/, ,https://b.example.com")
	got := config.Load().AllowedOrigins()
	want := []string{"https://edu.example.com", "https://a.example.com", "https://b.example.com"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("AllowedOrigins = %v, quería %v", got, want)
	}
}

func TestTrustedProxiesConfig(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "")
	if got := config.Load().TrustedProxies; len(got) == 0 {
		t.Fatalf("por defecto deben confiarse los rangos privados")
	}
	t.Setenv("TRUSTED_PROXIES", "none")
	if got := config.Load().TrustedProxies; len(got) != 0 {
		t.Fatalf("'none' no debe confiar en ningún proxy, obtuvo %v", got)
	}
	t.Setenv("TRUSTED_PROXIES", "203.0.113.0/24, 198.51.100.7")
	if got := config.Load().TrustedProxies; len(got) != 2 {
		t.Fatalf("lista personalizada mal parseada: %v", got)
	}
}
