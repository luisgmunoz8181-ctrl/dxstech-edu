package quizzes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"dxstech-edu/internal/ai"
	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/database"
	"dxstech-edu/internal/quizzes"

	"github.com/gin-gonic/gin"
)

type mockCompleter struct {
	calledWithLesson string
}

func (m *mockCompleter) CompleteLesson(ctx context.Context, userID, courseID, lessonID string) error {
	m.calledWithLesson = lessonID
	return nil
}

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "dxstech_quizzes_test_*")
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

func TestQuizSubmissionAndAutoCompletion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup := setupTestDB(t)
	defer cleanup()

	completer := &mockCompleter{}
	gemini := ai.NewGeminiClient()
	svc := quizzes.NewService(db, gemini, completer)

	// 1. Insert a test quiz
	quizID := "qz-test-101"
	questions := []ai.QuizQuestion{
		{
			Question:      "¿Qué es un LLM?",
			Options:       []string{"Modelo de Lenguaje Grande", "Algoritmo lineal", "Base de datos", "Compilador"},
			CorrectAnswer: "Modelo de Lenguaje Grande",
		},
		{
			Question:      "¿Qué técnica permite dar memoria a un agente?",
			Options:       []string{"Persistencia y Embeddings", "Reinicio constante", "Desactivar contexto", "None"},
			CorrectAnswer: "Persistencia y Embeddings",
		},
	}
	qJSON, _ := json.Marshal(questions)
	courseID := "crs-ai-101"
	lessonID := "lsn-ai-06"

	_, err := db.Exec(`
		INSERT INTO quizzes (id, title, notes, question_count, questions_json, course_id, lesson_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, quizID, "Quiz de Prueba", "Apuntes de prueba", 2, string(qJSON), courseID, lessonID)
	if err != nil {
		t.Fatalf("Error insertando quiz: %v", err)
	}

	r := gin.New()
	r.Use(auth.Authenticate(testSecret))
	svc.RegisterRoutes(r.Group("/api/quizzes"))
	studentToken := tokenFor(t, "usr-student-01", "ESTUDIANTE")

	// 2. Test GetByLesson
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/quizzes/lesson/"+lessonID, nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for GetByLesson, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Test Submit with Passing score (100%)
	subReq := quizzes.SubmitQuizRequest{
		CourseID: courseID,
		LessonID: lessonID,
		Answers: map[string]string{
			"0": "Modelo de Lenguaje Grande",
			"1": "Persistencia y Embeddings",
		},
	}
	subJSON, _ := json.Marshal(subReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/quizzes/%s/submit", quizID), bytes.NewReader(subJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+studentToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for Submit, got %d: %s", w.Code, w.Body.String())
	}

	var resp quizzes.SubmitQuizResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Error parsing submit response: %v", err)
	}

	if resp.Score != 100.0 || !resp.Passed {
		t.Errorf("Expected 100%% score and passed true, got score %.1f passed %v", resp.Score, resp.Passed)
	}
	if !resp.LessonMarked {
		t.Errorf("Expected lessonMarked true")
	}
	if completer.calledWithLesson != lessonID {
		t.Errorf("Expected completer called with lesson %s, got %s", lessonID, completer.calledWithLesson)
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
