package courses_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/courses"

	"github.com/gin-gonic/gin"
)

const secSecret = "test-secret-key-0123456789-0123456789"

func secToken(t *testing.T, id, role string) string {
	t.Helper()
	tok, err := auth.GenerateToken(&auth.User{ID: id, Email: id + "@t.local", Role: role, FirstName: "N", LastName: id}, secSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func newCourseRouter(t *testing.T) (*gin.Engine, string, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, dir, cleanup := setupTestDB(t)
	r := gin.New()
	r.Use(auth.Authenticate(secSecret))
	courses.NewHandler(courses.NewService(db, dir)).RegisterRoutes(r.Group("/api/courses"))
	return r, dir, cleanup
}

func upload(r *gin.Engine, token, filename string, content []byte) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", filename)
	_, _ = fw.Write(content)
	_ = mw.Close()
	req, _ := http.NewRequest("POST", "/api/courses/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUploadValidatesRealContent(t *testing.T) {
	r, dir, cleanup := newCourseRouter(t)
	defer cleanup()
	admin := secToken(t, "usr-admin-01", "ADMINISTRADOR")

	pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 64)...)
	png := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte("x"), 64)...)

	w := upload(r, admin, "Mi Archivo (1).PDF", pdf)
	if w.Code != http.StatusOK {
		t.Fatalf("PDF válido: esperado 200, obtuvo %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		URL string `json:"url"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if !regexp.MustCompile(`^/uploads/[0-9a-f]{32}\.pdf$`).MatchString(resp.URL) {
		t.Errorf("el nombre guardado debe ser aleatorio y no depender del cliente: %q", resp.URL)
	}
	if _, err := os.Stat(filepath.Join(dir, "uploads", filepath.Base(resp.URL))); err != nil {
		t.Errorf("el archivo debería existir en disco: %v", err)
	}

	if w := upload(r, admin, "foto.png", png); w.Code != http.StatusOK {
		t.Errorf("PNG válido: esperado 200, obtuvo %d", w.Code)
	}

	bad := []struct {
		name    string
		file    string
		content []byte
	}{
		{"HTML disfrazado de PDF", "x.pdf", []byte("<html><script>alert(1)</script></html>")},
		{"HTML disfrazado de PNG", "x.png", []byte("<script>alert(1)</script>")},
		{"PDF con extensión PNG", "x.png", pdf},
		{"SVG no permitido", "x.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"/>`)},
		{"HTML directo", "x.html", []byte("<html></html>")},
		{"vacío", "x.pdf", nil},
	}
	for _, tc := range bad {
		if w := upload(r, admin, tc.file, tc.content); w.Code != http.StatusBadRequest {
			t.Errorf("%s: esperado 400, obtuvo %d", tc.name, w.Code)
		}
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "uploads"))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".html") || strings.HasSuffix(e.Name(), ".svg") {
			t.Errorf("no debe quedar basura en disco: %s", e.Name())
		}
	}

	// Un estudiante no puede subir archivos.
	if w := upload(r, secToken(t, "usr-student-01", "ESTUDIANTE"), "a.pdf", pdf); w.Code != http.StatusForbidden {
		t.Errorf("estudiante subiendo: esperado 403, obtuvo %d", w.Code)
	}
}

func TestUploadRejectsOversizedBody(t *testing.T) {
	r, _, cleanup := newCourseRouter(t)
	defer cleanup()
	big := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 27*1024*1024)...)
	if w := upload(r, secToken(t, "usr-admin-01", "ADMINISTRADOR"), "big.pdf", big); w.Code != http.StatusBadRequest {
		t.Errorf("archivo > 25 MB: esperado 400, obtuvo %d", w.Code)
	}
}

func TestForumTutorAndReviewsRequireEnrollment(t *testing.T) {
	r, _, cleanup := newCourseRouter(t)
	defer cleanup()

	enrolled := secToken(t, "usr-student-01", "ESTUDIANTE") // matriculado en crs-ai-101 por el seed demo
	outsider := secToken(t, "usr-nobody", "ESTUDIANTE")
	admin := secToken(t, "usr-admin-01", "ADMINISTRADOR")

	call := func(method, path, token string, body string) int {
		req, _ := http.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	for _, p := range []string{
		"/api/courses/crs-ai-101/discussions",
		"/api/courses/crs-ai-101/tutor-context",
	} {
		if c := call("GET", p, "", ""); c != http.StatusUnauthorized {
			t.Errorf("GET %s anónimo: esperado 401, obtuvo %d", p, c)
		}
		if c := call("GET", p, outsider, ""); c != http.StatusForbidden {
			t.Errorf("GET %s sin matrícula: esperado 403, obtuvo %d", p, c)
		}
		if c := call("GET", p, enrolled, ""); c != http.StatusOK {
			t.Errorf("GET %s matriculado: esperado 200, obtuvo %d", p, c)
		}
		if c := call("GET", p, admin, ""); c != http.StatusOK {
			t.Errorf("GET %s admin: esperado 200, obtuvo %d", p, c)
		}
	}

	if c := call("POST", "/api/courses/crs-ai-101/discussions", outsider, `{"message":"hola"}`); c != http.StatusForbidden {
		t.Errorf("POST foro sin matrícula: esperado 403, obtuvo %d", c)
	}
	if c := call("POST", "/api/courses/crs-ai-101/reviews", outsider, `{"rating":5,"comment":"x"}`); c != http.StatusForbidden {
		t.Errorf("POST reseña sin matrícula: esperado 403, obtuvo %d", c)
	}
	if c := call("POST", "/api/courses/crs-ai-101/discussions", enrolled, `{"message":"hola"}`); c != http.StatusCreated {
		t.Errorf("POST foro matriculado: esperado 201, obtuvo %d", c)
	}
	// Las reseñas siguen siendo de lectura pública (catálogo).
	if c := call("GET", "/api/courses/crs-ai-101/reviews", "", ""); c != http.StatusOK {
		t.Errorf("GET reseñas público: esperado 200, obtuvo %d", c)
	}
}

func TestForumThreadsAndReplies(t *testing.T) {
	db, dir, cleanup := setupTestDB(t)
	defer cleanup()
	svc := courses.NewService(db, dir)
	ctx := context.Background()

	root, err := svc.CreateDiscussion(ctx, "crs-ai-101", "lsn-ai-01", "usr-student-01", "Carlos", "ESTUDIANTE", "¿Cómo funciona un LLM?", "Duda sobre LLM", "")
	if err != nil {
		t.Fatal(err)
	}
	if root.Title != "Duda sobre LLM" || root.ParentID != "" {
		t.Fatalf("hilo mal creado: %+v", root)
	}

	r1, err := svc.CreateDiscussion(ctx, "crs-ai-101", "", "usr-admin-01", "Admin", "ADMINISTRADOR", "Con transformers.", "ignorado", root.ID)
	if err != nil {
		t.Fatal(err)
	}
	if r1.ParentID != root.ID || r1.Title != "" || r1.LessonID != "lsn-ai-01" {
		t.Fatalf("la respuesta debe heredar hilo y lección y no llevar título: %+v", r1)
	}
	// Responder a una respuesta se anida en el hilo raíz.
	r2, err := svc.CreateDiscussion(ctx, "crs-ai-101", "", "usr-student-01", "Carlos", "ESTUDIANTE", "Gracias", "", r1.ID)
	if err != nil || r2.ParentID != root.ID {
		t.Fatalf("respuesta anidada: %+v %v", r2, err)
	}

	// Otro hilo en otra lección no aparece al filtrar por lsn-ai-01, pero uno general sí.
	if _, err := svc.CreateDiscussion(ctx, "crs-ai-101", "lsn-ai-02", "usr-student-01", "Carlos", "ESTUDIANTE", "Otra lección", "Otro", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateDiscussion(ctx, "crs-ai-101", "", "usr-student-01", "Carlos", "ESTUDIANTE", "General", "General", ""); err != nil {
		t.Fatal(err)
	}

	threads, err := svc.GetDiscussions(ctx, "crs-ai-101", "lsn-ai-01")
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 2 {
		t.Fatalf("hilos para lsn-ai-01 = %d, esperados 2 (el de la lección y el general)", len(threads))
	}
	if len(threads[0].Replies) != 2 || threads[0].Replies[0].Message != "Con transformers." {
		t.Fatalf("respuestas anidadas incorrectas: %+v", threads[0].Replies)
	}
	if all, _ := svc.GetDiscussions(ctx, "crs-ai-101", ""); len(all) != 3 {
		t.Errorf("sin filtro deben verse los 3 hilos, hay %d", len(all))
	}

	// Validaciones.
	if _, err := svc.CreateDiscussion(ctx, "crs-ai-101", "", "u", "n", "ESTUDIANTE", "x", "", "no-existe"); err == nil {
		t.Error("responder a un hilo inexistente debe fallar")
	}
	if _, err := svc.CreateDiscussion(ctx, "crs-dev-201", "", "u", "n", "ESTUDIANTE", "x", "", root.ID); err == nil {
		t.Error("no se puede responder a un hilo de otro curso")
	}
	if _, err := svc.CreateDiscussion(ctx, "crs-ai-101", "", "u", "n", "ESTUDIANTE", strings.Repeat("a", 5001), "", ""); err == nil {
		t.Error("un mensaje demasiado largo debe rechazarse")
	}
	if _, err := svc.CreateDiscussion(ctx, "crs-ai-101", "", "u", "n", "ESTUDIANTE", "ok", strings.Repeat("t", 151), ""); err == nil {
		t.Error("un título demasiado largo debe rechazarse")
	}
}
