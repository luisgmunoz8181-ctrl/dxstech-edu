package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"

	gzipmw "github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func TestHealthReflectsDatabaseState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir, _ := os.MkdirTemp("", "dxstech_health_*")
	defer os.RemoveAll(dir)
	db, err := database.InitDB(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.GET("/api/health", healthHandler(db, &config.Config{AppEnv: "production"}))
	get := func() (int, string) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/health", nil)
		r.ServeHTTP(w, req)
		return w.Code, w.Body.String()
	}

	if code, body := get(); code != 200 || !strings.Contains(body, `"database":"ok"`) {
		t.Fatalf("base sana: esperado 200/ok, obtuvo %d %s", code, body)
	}
	db.Close()
	code, body := get()
	if code != http.StatusServiceUnavailable || !strings.Contains(body, `"database":"error"`) || !strings.Contains(body, "degraded") {
		t.Fatalf("base caída: esperado 503/degraded, obtuvo %d %s", code, body)
	}
	if strings.Contains(strings.ToLower(body), "closed") {
		t.Errorf("el health no debe filtrar detalles internos del error: %s", body)
	}
}

func TestCacheControlPolicies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(cacheControlMiddleware())
	ok := func(c *gin.Context) { c.String(200, "x") }
	r.GET("/api/courses", ok)
	r.GET("/api/certificates/x/pdf", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=86400") // el handler tiene la última palabra
		c.String(200, "x")
	})
	r.GET("/vendor/lucide-0.469.0.min.js", ok)
	r.GET("/uploads/abc.pdf", ok)
	r.GET("/js/main.js", ok)
	r.GET("/css/tailwind.css", ok)

	want := map[string]string{
		"/api/courses":                  "no-store",
		"/api/certificates/x/pdf":       "public, max-age=86400",
		"/vendor/lucide-0.469.0.min.js": "public, max-age=31536000, immutable",
		"/uploads/abc.pdf":              "private, max-age=31536000, immutable",
		"/js/main.js":                   "no-cache",
		"/css/tailwind.css":             "no-cache",
	}
	for path, cc := range want {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", path, nil)
		r.ServeHTTP(w, req)
		if got := w.Header().Get("Cache-Control"); got != cc {
			t.Errorf("%s: Cache-Control = %q, esperado %q", path, got, cc)
		}
	}
}

func TestGzipCompressesTextButNotBinaries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gzipmw.Gzip(gzipmw.DefaultCompression,
		gzipmw.WithExcludedPaths([]string{"/uploads/", "/api/certificates/"}),
		gzipmw.WithExcludedExtensions([]string{".png", ".pdf", ".mp4"}),
	))
	big := strings.Repeat(`{"title":"curso","description":"texto repetido"},`, 200)
	r.GET("/api/courses", func(c *gin.Context) { c.Data(200, "application/json", []byte(big)) })
	r.GET("/uploads/video.mp4", func(c *gin.Context) { c.Data(200, "video/mp4", []byte(big)) })
	r.GET("/api/certificates/x/pdf", func(c *gin.Context) { c.Data(200, "application/pdf", []byte(big)) })

	do := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", path, nil)
		req.Header.Set("Accept-Encoding", "gzip")
		r.ServeHTTP(w, req)
		return w
	}

	w := do("/api/courses")
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("el JSON debe comprimirse")
	}
	zr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	plain, _ := io.ReadAll(zr)
	if string(plain) != big || w.Body.Len() >= len(big)/2 {
		t.Errorf("la compresión debe ser íntegra y efectiva (%d -> %d)", len(big), w.Body.Len())
	}

	for _, p := range []string{"/uploads/video.mp4", "/api/certificates/x/pdf"} {
		if w := do(p); w.Header().Get("Content-Encoding") != "" {
			t.Errorf("%s no debe comprimirse", p)
		}
	}
}

func TestSlowRoutesGetLongerDeadlinesWithoutBreakingRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(requestDeadlinesMiddleware())
	r.POST("/api/courses/upload", func(c *gin.Context) { c.String(200, "ok") })
	r.GET("/api/admin/reports/enrollments.csv", func(c *gin.Context) { c.String(200, "ok") })
	r.GET("/api/health", func(c *gin.Context) { c.String(200, "ok") })

	for _, rt := range []struct{ m, p string }{
		{"POST", "/api/courses/upload"},
		{"GET", "/api/admin/reports/enrollments.csv"},
		{"GET", "/api/health"},
	} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(rt.m, rt.p, nil)
		r.ServeHTTP(w, req) // ResponseRecorder no soporta deadlines: debe ignorarse sin fallar
		if w.Code != 200 {
			t.Errorf("%s %s: %d", rt.m, rt.p, w.Code)
		}
	}
}

// Con un servidor HTTP real, la ruta de subida conserva su conexión más allá del
// WriteTimeout corto del servidor, mientras que una ruta normal lo respeta.
func TestExtendedWriteDeadlineSurvivesShortServerTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(requestDeadlinesMiddleware())
	slow := func(c *gin.Context) {
		// Simula trabajo lento pero menor al plazo ampliado.
		sleep(600)
		c.String(200, "done")
	}
	r.POST("/api/quizzes/generate", slow)
	r.POST("/api/normal", slow)

	srv := httptest.NewUnstartedServer(r)
	srv.Config.WriteTimeout = 200 * 1e6 // 200 ms
	srv.Start()
	defer srv.Close()

	call := func(path string) error {
		resp, err := http.Post(srv.URL+path, "application/json", strings.NewReader("{}"))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil || string(b) != "done" {
			return io.ErrUnexpectedEOF
		}
		return nil
	}
	if err := call("/api/quizzes/generate"); err != nil {
		t.Errorf("la ruta lenta debería completar gracias al plazo ampliado: %v", err)
	}
	if err := call("/api/normal"); err == nil {
		t.Errorf("una ruta normal debería cortarse con el WriteTimeout corto del servidor")
	}
}

func sleep(ms int) { time.Sleep(time.Duration(ms) * time.Millisecond) }
