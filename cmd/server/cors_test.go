package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSOnlyAllowsConfiguredOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(corsMiddleware([]string{"https://edu.example.com"}))
	r.GET("/ping", func(c *gin.Context) { c.String(200, "ok") })

	get := func(origin string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest("GET", "/ping", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	if w := get("https://edu.example.com"); w.Header().Get("Access-Control-Allow-Origin") != "https://edu.example.com" ||
		w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("origen autorizado debe reflejarse con credenciales: %v", w.Header())
	}
	if w := get("https://evil.example.net"); w.Header().Get("Access-Control-Allow-Origin") != "" ||
		w.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Errorf("origen no autorizado no debe recibir cabeceras CORS: %v", w.Header())
	}
	if w := get(""); w.Header().Get("Access-Control-Allow-Origin") == "*" {
		t.Errorf("nunca debe usarse comodín")
	}
}

func TestClientIPIgnoresSpoofedForwardedForFromUntrustedPeer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	build := func(trusted []string) *gin.Engine {
		r := gin.New()
		_ = r.SetTrustedProxies(trusted)
		r.GET("/ip", func(c *gin.Context) { c.String(200, c.ClientIP()) })
		return r
	}
	ask := func(r *gin.Engine, remote string) string {
		req, _ := http.NewRequest("GET", "/ip", nil)
		req.RemoteAddr = remote
		req.Header.Set("X-Forwarded-For", "6.6.6.6")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Body.String()
	}

	// Cliente directo de internet (no es proxy de confianza): el header se ignora.
	if got := ask(build([]string{"10.0.0.0/8"}), "203.0.113.9:5555"); got != "203.0.113.9" {
		t.Errorf("IP falseable desde un peer no confiable: %q", got)
	}
	// Detrás de un proxy privado declarado: se respeta el reenvío.
	if got := ask(build([]string{"10.0.0.0/8"}), "10.1.2.3:5555"); got != "6.6.6.6" {
		t.Errorf("con proxy de confianza debe usarse X-Forwarded-For, obtuvo %q", got)
	}
}
