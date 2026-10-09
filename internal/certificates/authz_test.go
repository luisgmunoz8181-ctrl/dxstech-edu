package certificates_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/certificates"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

const authzSecret = "test-secret-key-0123456789-0123456789"

func TestCertificateRoutesAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir, _ := os.MkdirTemp("", "dxstech_cert_authz_*")
	defer os.RemoveAll(dir)
	db, err := database.InitDB(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	r := gin.New()
	r.Use(auth.Authenticate(authzSecret))
	certificates.NewService(db).RegisterRoutes(r.Group("/api/certificates"))

	token := func(role string) string {
		tok, _ := auth.GenerateToken(&auth.User{ID: "u-" + role, Role: role}, authzSecret, time.Hour)
		return tok
	}
	call := func(method, path, tok string) int {
		req, _ := http.NewRequest(method, path, nil)
		if tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	for _, rt := range []struct{ method, path string }{
		{"POST", "/api/certificates/generate"},
		{"GET", "/api/certificates/issued"},
	} {
		if c := call(rt.method, rt.path, ""); c != http.StatusUnauthorized {
			t.Errorf("%s %s anónimo: esperado 401, obtuvo %d", rt.method, rt.path, c)
		}
		if c := call(rt.method, rt.path, token("ESTUDIANTE")); c != http.StatusForbidden {
			t.Errorf("%s %s estudiante: esperado 403, obtuvo %d", rt.method, rt.path, c)
		}
	}
	if c := call("GET", "/api/certificates/issued", token("ADMINISTRADOR")); c != http.StatusOK {
		t.Errorf("issued como admin: esperado 200, obtuvo %d", c)
	}
	if c := call("GET", "/api/certificates/my-certificates", ""); c != http.StatusUnauthorized {
		t.Errorf("my-certificates anónimo: esperado 401, obtuvo %d", c)
	}
	// La verificación por QR sigue siendo pública (404 = no existe, pero sin exigir sesión).
	if c := call("GET", "/api/certificates/verify/DXS-0000-0000", ""); c == http.StatusUnauthorized || c == http.StatusForbidden {
		t.Errorf("verify debe ser público, obtuvo %d", c)
	}
}

func newCertEnv(t *testing.T) (*gin.Engine, *certificates.Service, string, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir, _ := os.MkdirTemp("", "dxstech_cert_pdf_*")
	db, err := database.InitDB(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	svc := certificates.NewService(db)
	r := gin.New()
	r.Use(auth.Authenticate(authzSecret))
	svc.RegisterRoutes(r.Group("/api/certificates"))

	cert, err := svc.IssueCourseCertificate(context.Background(), "usr-student-01", "crs-ai-101", "Carlos Estudiante", "Curso de IA", 8, "Docente", "https://edu.example.com")
	if err != nil {
		t.Fatal(err)
	}
	return r, svc, cert.ID, func() { db.Close(); os.RemoveAll(dir) }
}

func bearer(role, id string) string {
	tok, _ := auth.GenerateToken(&auth.User{ID: id, Role: role}, authzSecret, time.Hour)
	return tok
}

func get(r *gin.Engine, path, token string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCertificatePDFIsOnlyForOwnerOrAdmin(t *testing.T) {
	r, _, id, cleanup := newCertEnv(t)
	defer cleanup()

	for _, path := range []string{"/api/certificates/" + id + "/pdf", "/api/certificates/download/" + id} {
		if w := get(r, path, ""); w.Code != http.StatusUnauthorized {
			t.Errorf("%s anónimo: esperado 401, obtuvo %d", path, w.Code)
		}
		if w := get(r, path, bearer("ESTUDIANTE", "usr-otro")); w.Code != http.StatusForbidden {
			t.Errorf("%s otro estudiante: esperado 403, obtuvo %d", path, w.Code)
		}
		if w := get(r, path, bearer("ESTUDIANTE", "usr-student-01")); w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/pdf" {
			t.Errorf("%s titular: esperado 200 PDF, obtuvo %d (%s)", path, w.Code, w.Header().Get("Content-Type"))
		}
		if w := get(r, path, bearer("ADMINISTRADOR", "usr-admin-01")); w.Code != http.StatusOK {
			t.Errorf("%s admin: esperado 200, obtuvo %d", path, w.Code)
		}
	}
}

func TestPublicVerificationExposesOnlyMinimalData(t *testing.T) {
	r, _, id, cleanup := newCertEnv(t)
	defer cleanup()

	w := get(r, "/api/certificates/verify/"+id, "")
	if w.Code != http.StatusOK {
		t.Fatalf("verify: %d", w.Code)
	}
	body := w.Body.String()
	for _, leaked := range []string{"usr-student-01", "crs-ai-101", "userId", "courseId", "qrCodeUrl"} {
		if strings.Contains(body, leaked) {
			t.Errorf("la verificación pública no debe exponer %q: %s", leaked, body)
		}
	}
	for _, needed := range []string{"Carlos Estudiante", "Curso de IA", id} {
		if !strings.Contains(body, needed) {
			t.Errorf("la verificación pública debe incluir %q: %s", needed, body)
		}
	}
	if !regexp.MustCompile(`^DXS-\d{4}-[0-9A-F]{12}$`).MatchString(id) {
		t.Errorf("formato de código inesperado (se esperan 48 bits): %s", id)
	}
}

func TestVerificationEnumerationIsRateLimited(t *testing.T) {
	r, _, id, cleanup := newCertEnv(t)
	defer cleanup()

	var last int
	for i := 0; i < 40; i++ {
		last = get(r, fmt.Sprintf("/api/certificates/verify/DXS-2026-%012X", i), "").Code
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("tras 30 códigos inexistentes la IP debe limitarse (429), obtuvo %d", last)
	}
	// Mientras está limitada, ni un código válido responde.
	if w := get(r, "/api/certificates/verify/"+id, ""); w.Code != http.StatusTooManyRequests {
		t.Errorf("IP limitada: esperado 429, obtuvo %d", w.Code)
	}
}
