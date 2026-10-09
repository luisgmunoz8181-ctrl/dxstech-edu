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
