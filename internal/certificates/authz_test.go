package certificates_test

import (
	"net/http"
	"net/http/httptest"
	"os"
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
