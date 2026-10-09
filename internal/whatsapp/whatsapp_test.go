package whatsapp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"dxstech-edu/internal/ai"
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

	db, err := database.InitDB(tempDir)
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
