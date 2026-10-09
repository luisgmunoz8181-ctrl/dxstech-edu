package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

type sessionEnv struct {
	db  *database.DB
	svc *auth.Service
	r   *gin.Engine
	cfg *config.Config
}

func newSessionEnv(t *testing.T) (*sessionEnv, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir, _ := os.MkdirTemp("", "dxstech_sess_*")
	db, err := database.InitDB(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{AppEnv: "development", JWTSecret: "test-secret-key-0123456789-0123456789"}
	svc := auth.NewService(db, cfg)

	r := gin.New()
	r.Use(auth.Authenticate(cfg.JWTSecret, auth.NewSessionLookup(db)))
	auth.NewHandler(svc, false).RegisterRoutes(r.Group("/api/auth"))
	r.GET("/api/ping", auth.RequireAuth(), func(c *gin.Context) { c.String(200, "pong") })

	return &sessionEnv{db: db, svc: svc, r: r, cfg: cfg}, func() { db.Close(); os.RemoveAll(dir) }
}

func (e *sessionEnv) login(t *testing.T, email, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	var resp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if w.Code != 200 || resp.Token == "" {
		t.Fatalf("login falló (%d): %s", w.Code, w.Body.String())
	}
	return resp.Token
}

func (e *sessionEnv) do(method, path, token string) (int, string) {
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	return w.Code, w.Body.String()
}

func TestTokenRevokedAfterPasswordChange(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()

	tok := e.login(t, "estudiante@dxstech.edu", "Student1234*")
	if c, _ := e.do("GET", "/api/ping", tok); c != 200 {
		t.Fatalf("token recién emitido debe ser válido, obtuvo %d", c)
	}

	if err := e.svc.ChangePassword(context.Background(), "usr-student-01", "Student1234*", "NuevaClave#2026", "", ""); err != nil {
		t.Fatal(err)
	}
	if c, _ := e.do("GET", "/api/ping", tok); c != 401 {
		t.Errorf("token anterior al cambio de contraseña debe ser rechazado, obtuvo %d", c)
	}

	fresh, err := e.svc.IssueToken(context.Background(), "usr-student-01", 3600e9)
	if err != nil {
		t.Fatal(err)
	}
	if c, _ := e.do("GET", "/api/ping", fresh); c != 200 {
		t.Errorf("el token reemitido debe ser válido, obtuvo %d", c)
	}
}

func TestTokenRevokedWhenUserDeactivatedOrRoleChanged(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()
	ctx := context.Background()

	studentTok := e.login(t, "estudiante@dxstech.edu", "Student1234*")
	if _, err := e.svc.ToggleUserStatus(ctx, "usr-admin-01", "usr-student-01", "", ""); err != nil {
		t.Fatal(err)
	}
	if c, _ := e.do("GET", "/api/ping", studentTok); c != 401 {
		t.Errorf("usuario desactivado: esperado 401, obtuvo %d", c)
	}

	adminTok := e.login(t, "admin@dxstech.edu", "Admin1234*")
	_, err := e.svc.UpdateUser(ctx, "usr-superadmin-01", "usr-admin-01", auth.UpdateUserRequest{
		FirstName: "Admin", LastName: "DxSTech", Email: "admin@dxstech.edu", RoleID: 3, Status: "active",
	}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if c, _ := e.do("GET", "/api/ping", adminTok); c != 401 {
		t.Errorf("rol degradado: el token con el rol anterior debe caducar, obtuvo %d", c)
	}
}

func TestMustChangePasswordBlocksEverythingButPasswordChange(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()

	tok := e.login(t, "estudiante@dxstech.edu", "Student1234*")
	if _, err := e.db.Exec(`UPDATE users SET must_change_password = 1 WHERE id = 'usr-student-01'`); err != nil {
		t.Fatal(err)
	}

	c, body := e.do("GET", "/api/ping", tok)
	if c != 403 || !bytes.Contains([]byte(body), []byte("PASSWORD_CHANGE_REQUIRED")) {
		t.Errorf("con contraseña temporal debe bloquear: %d %s", c, body)
	}
	if c, _ := e.do("GET", "/api/auth/me", tok); c != 200 {
		t.Errorf("/auth/me debe seguir disponible, obtuvo %d", c)
	}

	if err := e.svc.ChangePassword(context.Background(), "usr-student-01", "Student1234*", "NuevaClave#2026", "", ""); err != nil {
		t.Fatal(err)
	}
	fresh, _ := e.svc.IssueToken(context.Background(), "usr-student-01", 3600e9)
	if c, _ := e.do("GET", "/api/ping", fresh); c != 200 {
		t.Errorf("tras cambiar la contraseña debe desbloquearse, obtuvo %d", c)
	}
}
