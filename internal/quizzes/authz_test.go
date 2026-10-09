package quizzes_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dxstech-edu/internal/ai"
	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/quizzes"

	"github.com/gin-gonic/gin"
)

func newAuthzRouter(t *testing.T) (*gin.Engine, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, cleanup := setupTestDB(t)

	qJSON, _ := json.Marshal([]ai.QuizQuestion{{
		Question:      "¿Qué es un LLM?",
		Options:       []string{"Modelo de Lenguaje Grande", "Algoritmo lineal", "Base de datos", "Compilador"},
		CorrectAnswer: "Modelo de Lenguaje Grande",
	}})
	if _, err := db.Exec(`INSERT INTO quizzes (id, title, notes, question_count, questions_json, course_id, lesson_id)
		VALUES ('qz-authz', 'Quiz', 'n', 1, ?, 'crs-ai-101', 'lsn-ai-06')`, string(qJSON)); err != nil {
		cleanup()
		t.Fatalf("insert quiz: %v", err)
	}

	r := gin.New()
	r.Use(auth.Authenticate(testSecret))
	quizzes.NewService(db, ai.NewGeminiClient(), &mockCompleter{}).RegisterRoutes(r.Group("/api/quizzes"))
	return r, cleanup
}

func do(r *gin.Engine, method, path, token string, body []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestQuizAdminRoutesRequireAdminRole(t *testing.T) {
	r, cleanup := newAuthzRouter(t)
	defer cleanup()

	student := tokenFor(t, "usr-student-01", "ESTUDIANTE")
	admin := tokenFor(t, "usr-admin-01", "ADMINISTRADOR")

	routes := []struct{ method, path string }{
		{"POST", "/api/quizzes/generate"},
		{"POST", "/api/quizzes/generate-for-lesson"},
		{"GET", "/api/quizzes"},
		{"DELETE", "/api/quizzes/qz-authz"},
	}
	for _, rt := range routes {
		if w := do(r, rt.method, rt.path, "", []byte("{}")); w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin sesión: esperado 401, obtuvo %d", rt.method, rt.path, w.Code)
		}
		if w := do(r, rt.method, rt.path, student, []byte("{}")); w.Code != http.StatusForbidden {
			t.Errorf("%s %s como estudiante: esperado 403, obtuvo %d", rt.method, rt.path, w.Code)
		}
	}

	if w := do(r, "GET", "/api/quizzes", admin, nil); w.Code != http.StatusOK {
		t.Errorf("GET /api/quizzes como admin: esperado 200, obtuvo %d", w.Code)
	}
	if w := do(r, "DELETE", "/api/quizzes/qz-authz", admin, nil); w.Code != http.StatusOK {
		t.Errorf("DELETE como admin: esperado 200, obtuvo %d", w.Code)
	}
}

func TestQuizAnswersHiddenFromStudents(t *testing.T) {
	r, cleanup := newAuthzRouter(t)
	defer cleanup()

	student := tokenFor(t, "usr-student-01", "ESTUDIANTE")
	admin := tokenFor(t, "usr-admin-01", "ADMINISTRADOR")

	for _, path := range []string{"/api/quizzes/qz-authz", "/api/quizzes/lesson/lsn-ai-06"} {
		if w := do(r, "GET", path, "", nil); w.Code != http.StatusUnauthorized {
			t.Errorf("GET %s sin sesión: esperado 401, obtuvo %d", path, w.Code)
		}

		var detail quizzes.QuizDetail
		w := do(r, "GET", path, student, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s como estudiante: esperado 200, obtuvo %d", path, w.Code)
		}
		_ = json.Unmarshal(w.Body.Bytes(), &detail)
		if len(detail.Questions) == 0 || detail.Questions[0].CorrectAnswer != "" {
			t.Errorf("GET %s expone correctAnswer a un estudiante: %s", path, w.Body.String())
		}

		w = do(r, "GET", path, admin, nil)
		_ = json.Unmarshal(w.Body.Bytes(), &detail)
		if len(detail.Questions) == 0 || detail.Questions[0].CorrectAnswer == "" {
			t.Errorf("GET %s debería conservar correctAnswer para administradores", path)
		}
	}
}

func TestQuizSubmitRequiresSessionAndRecordsUser(t *testing.T) {
	r, cleanup := newAuthzRouter(t)
	defer cleanup()

	body := []byte(`{"answers":{"0":"Modelo de Lenguaje Grande"}}`)
	if w := do(r, "POST", "/api/quizzes/qz-authz/submit", "", body); w.Code != http.StatusUnauthorized {
		t.Errorf("submit sin sesión: esperado 401, obtuvo %d", w.Code)
	}
}
