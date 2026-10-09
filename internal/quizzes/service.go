package quizzes

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"dxstech-edu/internal/ai"
	"dxstech-edu/internal/audit"
	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

type LessonCompleter interface {
	CompleteLesson(ctx context.Context, userID, courseID, lessonID string) error
}

type GenerateRequest struct {
	Title    string `json:"title" binding:"required"`
	Notes    string `json:"notes" binding:"required"`
	Count    int    `json:"count"`
	APIKey   string `json:"apiKey"`
	CourseID string `json:"courseId"`
	LessonID string `json:"lessonId"`
}

type GenerateForLessonRequest struct {
	LessonID string `json:"lessonId" binding:"required"`
	Count    int    `json:"count"`
	APIKey   string `json:"apiKey"`
}

type QuizSummary struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Notes         string    `json:"notes"`
	QuestionCount int       `json:"questionCount"`
	CourseID      string    `json:"courseId,omitempty"`
	LessonID      string    `json:"lessonId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type QuizDetail struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Notes         string            `json:"notes"`
	QuestionCount int               `json:"questionCount"`
	CourseID      string            `json:"courseId,omitempty"`
	LessonID      string            `json:"lessonId,omitempty"`
	Questions     []ai.QuizQuestion `json:"questions"`
	CreatedAt     time.Time         `json:"createdAt"`
	UserPassed    bool              `json:"userPassed,omitempty"`
	LatestScore   *float64          `json:"latestScore,omitempty"`
}

type SubmitQuizRequest struct {
	CourseID string            `json:"courseId"`
	LessonID string            `json:"lessonId"`
	Answers  map[string]string `json:"answers" binding:"required"` // questionIndex ("0", "1", ...) -> selected option
}

type QuestionFeedback struct {
	Index          int    `json:"index"`
	Question       string `json:"question"`
	SelectedAnswer string `json:"selectedAnswer"`
	CorrectAnswer  string `json:"correctAnswer"`
	IsCorrect      bool   `json:"isCorrect"`
}

type SubmitQuizResponse struct {
	SubmissionID string             `json:"submissionId"`
	QuizID       string             `json:"quizId"`
	Total        int                `json:"total"`
	Correct      int                `json:"correct"`
	Score        float64            `json:"score"`
	Passed       bool               `json:"passed"`
	LessonMarked bool               `json:"lessonMarked"`
	Feedback     []QuestionFeedback `json:"feedback"`
}

type Service struct {
	db        *database.DB
	gemini    *ai.GeminiClient
	completer LessonCompleter
}

func NewService(db *database.DB, gemini *ai.GeminiClient, completer LessonCompleter) *Service {
	return &Service{
		db:        db,
		gemini:    gemini,
		completer: completer,
	}
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	admin := auth.RequireRole("SUPERADMIN", "ADMINISTRADOR")

	// Gestión (solo administradores)
	r.POST("/generate", admin, s.Generate)
	r.POST("/generate-for-lesson", admin, s.GenerateForLesson)
	r.GET("", admin, s.List)
	r.DELETE("/:id", admin, s.Delete)

	// Consumo por estudiantes (requiere sesión)
	r.GET("/lesson/:lessonId", auth.RequireAuth(), s.GetByLesson)
	r.POST("/:id/submit", auth.RequireAuth(), s.Submit)
	r.GET("/:id", auth.RequireAuth(), s.GetByID)
}

func isAdminRole(c *gin.Context) bool {
	role := c.GetString("userRole")
	return role == "SUPERADMIN" || role == "ADMINISTRADOR"
}

// redactAnswers elimina la clave de respuesta de las preguntas para que los
// estudiantes no puedan leerla antes de enviar su intento. El resultado
// correcto solo se devuelve en la retroalimentación de Submit.
func redactAnswers(questions []ai.QuizQuestion) []ai.QuizQuestion {
	out := make([]ai.QuizQuestion, len(questions))
	for i, q := range questions {
		q.CorrectAnswer = ""
		out[i] = q
	}
	return out
}

func (s *Service) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe completar el título y el material para generar la evaluación"})
		return
	}

	apiKey := req.APIKey
	if strings.TrimSpace(apiKey) == "" {
		apiKey = c.GetHeader("X-Gemini-API-Key")
	}
	if strings.TrimSpace(apiKey) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Se requiere tu Gemini API Key para generar la evaluación. Por favor configúrala en la pestaña Configuración."})
		return
	}

	if req.Count < 3 || req.Count > 15 {
		req.Count = 5
	}

	questions, err := s.gemini.GenerateQuiz(c.Request.Context(), apiKey, req.Title, req.Notes, req.Count)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Fallo al generar preguntas con Gemini: " + err.Error()})
		return
	}

	questionsJSON, err := json.Marshal(questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno serializando evaluación"})
		return
	}

	idBytes := make([]byte, 8)
	_, _ = rand.Read(idBytes)
	quizID := hex.EncodeToString(idBytes)

	now := time.Now()
	_, err = s.db.Exec(`
		INSERT INTO quizzes (id, title, notes, question_count, questions_json, course_id, lesson_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, quizID, req.Title, req.Notes, len(questions), string(questionsJSON), req.CourseID, req.LessonID, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error guardando la evaluación en la base de datos"})
		return
	}

	// If linked to a lesson, update lesson
	if strings.TrimSpace(req.LessonID) != "" {
		_, _ = s.db.Exec(`UPDATE lessons SET quiz_id = ?, content_type = 'quiz', updated_at = ? WHERE id = ?`, quizID, now, req.LessonID)
	}

	audit.Record(s.db, c, audit.QuizGenerate, "quizzes", "Evaluación generada con IA: %s (%s, %d preguntas)", req.Title, quizID, len(questions))
	c.JSON(http.StatusCreated, QuizDetail{
		ID:            quizID,
		Title:         req.Title,
		Notes:         req.Notes,
		QuestionCount: len(questions),
		CourseID:      req.CourseID,
		LessonID:      req.LessonID,
		Questions:     questions,
		CreatedAt:     now,
	})
}

func (s *Service) GenerateForLesson(c *gin.Context) {
	var req GenerateForLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe especificar la lección (lessonId)"})
		return
	}

	apiKey := req.APIKey
	if strings.TrimSpace(apiKey) == "" {
		apiKey = c.GetHeader("X-Gemini-API-Key")
	}
	if strings.TrimSpace(apiKey) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Se requiere tu Gemini API Key para generar la evaluación pedagógica."})
		return
	}

	// Fetch lesson details
	var title, desc, body, courseID string
	err := s.db.QueryRowContext(c.Request.Context(), `
		SELECT title, description, content_body, course_id
		FROM lessons WHERE id = ?
	`, req.LessonID).Scan(&title, &desc, &body, &courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lección no encontrada"})
		return
	}

	material := desc
	if strings.TrimSpace(body) != "" {
		material += "\n" + body
	}
	if len(strings.TrimSpace(material)) < 15 {
		material = fmt.Sprintf("Evaluación sobre los conceptos fundamentales de la lección: %s", title)
	}

	count := req.Count
	if count < 3 || count > 10 {
		count = 4
	}

	questions, err := s.gemini.GenerateQuiz(c.Request.Context(), apiKey, title, material, count)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error de Gemini: " + err.Error()})
		return
	}

	questionsJSON, err := json.Marshal(questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al serializar examen"})
		return
	}

	idBytes := make([]byte, 8)
	_, _ = rand.Read(idBytes)
	quizID := fmt.Sprintf("qz-%s", hex.EncodeToString(idBytes))
	now := time.Now()

	_, err = s.db.ExecContext(c.Request.Context(), `
		INSERT INTO quizzes (id, title, notes, question_count, questions_json, course_id, lesson_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, quizID, "Evaluación: "+title, material, len(questions), string(questionsJSON), courseID, req.LessonID, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar quiz"})
		return
	}

	// Link quiz to lesson
	_, _ = s.db.ExecContext(c.Request.Context(), `
		UPDATE lessons SET quiz_id = ?, content_type = 'quiz', updated_at = ? WHERE id = ?
	`, quizID, now, req.LessonID)

	audit.Record(s.db, c, audit.QuizGenerate, "quizzes", "Evaluación generada con IA para la lección %s (%s, %d preguntas)", req.LessonID, quizID, len(questions))
	c.JSON(http.StatusCreated, QuizDetail{
		ID:            quizID,
		Title:         "Evaluación: " + title,
		Notes:         material,
		QuestionCount: len(questions),
		CourseID:      courseID,
		LessonID:      req.LessonID,
		Questions:     questions,
		CreatedAt:     now,
	})
}

func (s *Service) Submit(c *gin.Context) {
	quizID := c.Param("id")
	userID := c.GetString("userID")

	var req SubmitQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe enviar sus respuestas para ser evaluadas"})
		return
	}

	var questionsJSON, courseID, lessonID string
	err := s.db.QueryRowContext(c.Request.Context(), `
		SELECT questions_json, coalesce(course_id, ''), coalesce(lesson_id, '')
		FROM quizzes WHERE id = ?
	`, quizID).Scan(&questionsJSON, &courseID, &lessonID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evaluación no encontrada"})
		return
	}

	if req.CourseID != "" {
		courseID = req.CourseID
	}
	if req.LessonID != "" {
		lessonID = req.LessonID
	}

	var questions []ai.QuizQuestion
	if err := json.Unmarshal([]byte(questionsJSON), &questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error leyendo preguntas de la evaluación"})
		return
	}

	feedback := make([]QuestionFeedback, 0, len(questions))
	correctCount := 0

	for i, q := range questions {
		key := fmt.Sprintf("%d", i)
		selected := strings.TrimSpace(req.Answers[key])
		isCorrect := strings.EqualFold(selected, strings.TrimSpace(q.CorrectAnswer))
		if isCorrect {
			correctCount++
		}
		feedback = append(feedback, QuestionFeedback{
			Index:          i + 1,
			Question:       q.Question,
			SelectedAnswer: selected,
			CorrectAnswer:  q.CorrectAnswer,
			IsCorrect:      isCorrect,
		})
	}

	total := len(questions)
	score := 0.0
	if total > 0 {
		score = (float64(correctCount) / float64(total)) * 100.0
	}
	passed := score >= 70.0

	// Store submission
	idBytes := make([]byte, 8)
	_, _ = rand.Read(idBytes)
	subID := fmt.Sprintf("sub-%s", hex.EncodeToString(idBytes))
	answersJSON, _ := json.Marshal(req.Answers)

	passedInt := 0
	if passed {
		passedInt = 1
	}

	_, _ = s.db.ExecContext(c.Request.Context(), `
		INSERT INTO quiz_submissions (id, quiz_id, user_id, course_id, lesson_id, score, passed, answers_json, submitted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, subID, quizID, userID, courseID, lessonID, score, passedInt, string(answersJSON), time.Now())

	lessonMarked := false
	if passed && lessonID != "" && courseID != "" && userID != "" && s.completer != nil {
		if err := s.completer.CompleteLesson(c.Request.Context(), userID, courseID, lessonID); err == nil {
			lessonMarked = true
		}
	}

	c.JSON(http.StatusOK, SubmitQuizResponse{
		SubmissionID: subID,
		QuizID:       quizID,
		Total:        total,
		Correct:      correctCount,
		Score:        score,
		Passed:       passed,
		LessonMarked: lessonMarked,
		Feedback:     feedback,
	})
}

func (s *Service) GetByLesson(c *gin.Context) {
	lessonID := c.Param("lessonId")
	userID := c.GetString("userID")

	var q QuizDetail
	var questionsJSON string
	err := s.db.QueryRowContext(c.Request.Context(), `
		SELECT q.id, q.title, q.notes, q.question_count, q.questions_json, coalesce(q.course_id, ''), coalesce(q.lesson_id, ''), q.created_at
		FROM quizzes q
		WHERE q.lesson_id = ? OR q.id = (SELECT quiz_id FROM lessons WHERE id = ?)
		ORDER BY q.created_at DESC LIMIT 1
	`, lessonID, lessonID).Scan(&q.ID, &q.Title, &q.Notes, &q.QuestionCount, &questionsJSON, &q.CourseID, &q.LessonID, &q.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No hay evaluación asociada a esta lección"})
		return
	}

	if err := json.Unmarshal([]byte(questionsJSON), &q.Questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deserializando preguntas"})
		return
	}

	// Check if user has previously submitted
	if userID != "" {
		var latestScore sql.NullFloat64
		var passedInt int
		err = s.db.QueryRowContext(c.Request.Context(), `
			SELECT score, passed FROM quiz_submissions
			WHERE quiz_id = ? AND user_id = ?
			ORDER BY submitted_at DESC LIMIT 1
		`, q.ID, userID).Scan(&latestScore, &passedInt)
		if err == nil {
			if latestScore.Valid {
				sc := latestScore.Float64
				q.LatestScore = &sc
			}
			q.UserPassed = passedInt == 1
		}
	}

	if !isAdminRole(c) {
		q.Questions = redactAnswers(q.Questions)
	}

	c.JSON(http.StatusOK, q)
}

func (s *Service) List(c *gin.Context) {
	rows, err := s.db.Query(`
		SELECT id, title, notes, question_count, coalesce(course_id, ''), coalesce(lesson_id, ''), created_at
		FROM quizzes
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error consultando evaluaciones"})
		return
	}
	defer rows.Close()

	list := make([]QuizSummary, 0)
	for rows.Next() {
		var q QuizSummary
		if err := rows.Scan(&q.ID, &q.Title, &q.Notes, &q.QuestionCount, &q.CourseID, &q.LessonID, &q.CreatedAt); err == nil {
			list = append(list, q)
		}
	}

	c.JSON(http.StatusOK, list)
}

func (s *Service) GetByID(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var q QuizDetail
	var questionsJSON string
	err := s.db.QueryRow(`
		SELECT id, title, notes, question_count, questions_json, coalesce(course_id, ''), coalesce(lesson_id, ''), created_at
		FROM quizzes
		WHERE id = ?
	`, id).Scan(&q.ID, &q.Title, &q.Notes, &q.QuestionCount, &questionsJSON, &q.CourseID, &q.LessonID, &q.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evaluación no encontrada"})
		return
	}

	if err := json.Unmarshal([]byte(questionsJSON), &q.Questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al leer preguntas del quiz"})
		return
	}

	if userID != "" {
		var latestScore sql.NullFloat64
		var passedInt int
		_ = s.db.QueryRow(`
			SELECT score, passed FROM quiz_submissions
			WHERE quiz_id = ? AND user_id = ?
			ORDER BY submitted_at DESC LIMIT 1
		`, id, userID).Scan(&latestScore, &passedInt)
		if latestScore.Valid {
			sc := latestScore.Float64
			q.LatestScore = &sc
		}
		q.UserPassed = passedInt == 1
	}

	if !isAdminRole(c) {
		q.Questions = redactAnswers(q.Questions)
	}

	c.JSON(http.StatusOK, q)
}

func (s *Service) Delete(c *gin.Context) {
	id := c.Param("id")

	res, err := s.db.Exec(`DELETE FROM quizzes WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando evaluación de la base de datos"})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evaluación no encontrada"})
		return
	}

	audit.Record(s.db, c, audit.QuizDelete, "quizzes", "Evaluación eliminada: %s", id)
	c.JSON(http.StatusOK, gin.H{"message": "Evaluación eliminada correctamente", "id": id})
}
