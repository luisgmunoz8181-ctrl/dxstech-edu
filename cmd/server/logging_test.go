package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func logLines(buf *bytes.Buffer) []map[string]any {
	var out []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var m map[string]any
		if json.Unmarshal([]byte(l), &m) == nil {
			out = append(out, m)
		}
	}
	return out
}

func TestRequestIDAndStructuredAccessLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	buf := captureLogs(t)

	r := gin.New()
	r.Use(requestIDMiddleware(), accessLogMiddleware())
	r.GET("/api/courses", func(c *gin.Context) {
		c.Set("userID", "usr-42")
		c.Set("userRole", "ESTUDIANTE")
		c.String(200, "ok")
	})
	r.GET("/api/health", func(c *gin.Context) { c.String(200, "ok") })

	do := func(path, rid string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest("GET", path, nil)
		if rid != "" {
			req.Header.Set("X-Request-ID", rid)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	// Un request-id válido del proxy se respeta; uno inválido (inyección) se reemplaza.
	if w := do("/api/courses?token=SECRET123&apiKey=AIza-secret", "proxy-req-0001"); w.Header().Get("X-Request-ID") != "proxy-req-0001" {
		t.Errorf("debe respetar el X-Request-ID válido: %q", w.Header().Get("X-Request-ID"))
	}
	if w := do("/api/courses", "mal id\nx"); w.Header().Get("X-Request-ID") == "" || strings.ContainsAny(w.Header().Get("X-Request-ID"), " \n") {
		t.Errorf("un X-Request-ID inválido debe sustituirse: %q", w.Header().Get("X-Request-ID"))
	}
	do("/api/health", "") // un health correcto no se registra

	raw := buf.String()
	if strings.Contains(raw, "SECRET123") || strings.Contains(raw, "AIza-secret") {
		t.Fatalf("la query string (tokens, claves) no debe aparecer en los logs: %s", raw)
	}
	lines := logLines(buf)
	if len(lines) != 2 {
		t.Fatalf("se esperaban 2 líneas de log (sin el health), hay %d: %s", len(lines), raw)
	}
	first := lines[0]
	for k, want := range map[string]any{"msg": "http_request", "request_id": "proxy-req-0001", "method": "GET", "path": "/api/courses", "status": float64(200), "user_id": "usr-42", "role": "ESTUDIANTE", "route": "/api/courses"} {
		if first[k] != want {
			t.Errorf("campo %s = %v, esperado %v", k, first[k], want)
		}
	}
	if _, ok := first["latency_ms"]; !ok {
		t.Error("falta latency_ms")
	}
}

func TestPanicIsRecoveredAndLoggedWithRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	buf := captureLogs(t)

	r := gin.New()
	r.Use(requestIDMiddleware(), recoveryMiddleware(), accessLogMiddleware())
	r.GET("/boom", func(c *gin.Context) { panic("fallo interno con detalles secretos") })

	req, _ := http.NewRequest("GET", "/boom", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Fatalf("esperado 500, obtuvo %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "secretos") {
		t.Errorf("la respuesta no debe filtrar el detalle del panic: %s", w.Body.String())
	}
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["request_id"] == "" || body["request_id"] != w.Header().Get("X-Request-ID") {
		t.Errorf("la respuesta debe incluir el request_id para soporte: %v", body)
	}
	if !strings.Contains(buf.String(), "panic recuperado") || !strings.Contains(buf.String(), body["request_id"]) {
		t.Errorf("el panic debe registrarse con su request_id: %s", buf.String())
	}
}
