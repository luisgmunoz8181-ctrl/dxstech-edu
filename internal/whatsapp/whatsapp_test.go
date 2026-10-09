package whatsapp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"dxstech-edu/internal/ai"
	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"
	"dxstech-edu/internal/whatsapp"

	"github.com/gin-gonic/gin"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "dxstech_wa_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	db, err := database.InitDB(tempDir, true)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to init db: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return db, cleanup
}

func TestWhatsAppCourseSyncAndTutor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.Config{AppEnv: "development"}
	gemini := ai.NewGeminiClient()
	svc := whatsapp.NewService(db, gemini, cfg)

	r := gin.New()
	r.Use(auth.Authenticate(testSecret))
	svc.RegisterRoutes(r.Group("/api/whatsapp"))

	// 1. Test Sync Courses
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/whatsapp/sync-courses", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 for sync courses, got %d: %s", w.Code, w.Body.String())
	}

	var syncResp struct {
		SyncedCourses int    `json:"syncedCourses"`
		KnowledgeBase string `json:"knowledgeBase"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &syncResp)
	if syncResp.SyncedCourses < 1 {
		t.Errorf("Expected at least 1 synced course from seeded courses, got %d", syncResp.SyncedCourses)
	}

	// 2. Test Ask Tutor (without API key returns helpful fallback prompt)
	tutorReq := map[string]string{
		"courseId": "crs-ai-101",
		"question": "¿Qué temas se cubren en este curso?",
	}
	body, _ := json.Marshal(tutorReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/whatsapp/ask-tutor", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenFor(t, "usr-student-01", "ESTUDIANTE"))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 for ask tutor, got %d", w.Code)
	}

	var tutorResp struct {
		Reply string `json:"reply"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &tutorResp)
	if tutorResp.Reply == "" {
		t.Errorf("Expected non-empty reply from tutor")
	}
}

const testSecret = "test-secret-key-0123456789-0123456789"

func tokenFor(t *testing.T, userID, role string) string {
	t.Helper()
	tok, err := auth.GenerateToken(&auth.User{ID: userID, Email: userID + "@test.local", Role: role}, testSecret, time.Hour)
	if err != nil {
		t.Fatalf("no se pudo generar token: %v", err)
	}
	return tok
}

func TestWhatsAppProductionRequiresAdminSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc := whatsapp.NewService(db, ai.NewGeminiClient(), &config.Config{AppEnv: "production"})
	r := gin.New()
	r.Use(auth.Authenticate(testSecret))
	svc.RegisterRoutes(r.Group("/api/whatsapp"))

	call := func(method, path string, mutate func(*http.Request)) int {
		req, _ := http.NewRequest(method, path, bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		if mutate != nil {
			mutate(req)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}
	bearer := func(role string) func(*http.Request) {
		return func(req *http.Request) { req.Header.Set("Authorization", "Bearer "+tokenFor(t, "u-"+role, role)) }
	}

	adminRoutes := []struct{ method, path string }{
		{"GET", "/api/whatsapp/status"},
		{"POST", "/api/whatsapp/config"},
		{"POST", "/api/whatsapp/send"},
		{"GET", "/api/whatsapp/logs"},
		{"DELETE", "/api/whatsapp/logs"},
	}
	for _, rt := range adminRoutes {
		if code := call(rt.method, rt.path, nil); code != http.StatusUnauthorized {
			t.Errorf("%s %s anónimo: esperado 401, obtuvo %d", rt.method, rt.path, code)
		}
		// La antigua contraseña hardcodeada ya no concede acceso.
		legacy := func(req *http.Request) { req.Header.Set("X-Admin-Password", "dxstech2026") }
		if code := call(rt.method, rt.path, legacy); code != http.StatusUnauthorized {
			t.Errorf("%s %s con contraseña heredada: esperado 401, obtuvo %d", rt.method, rt.path, code)
		}
		if code := call(rt.method, rt.path, bearer("ESTUDIANTE")); code != http.StatusForbidden {
			t.Errorf("%s %s como estudiante: esperado 403, obtuvo %d", rt.method, rt.path, code)
		}
	}
	if code := call("GET", "/api/whatsapp/status", bearer("ADMINISTRADOR")); code != http.StatusOK {
		t.Errorf("status como admin: esperado 200, obtuvo %d", code)
	}
	if code := call("POST", "/api/whatsapp/ask-tutor", nil); code != http.StatusUnauthorized {
		t.Errorf("ask-tutor anónimo: esperado 401, obtuvo %d", code)
	}
}
