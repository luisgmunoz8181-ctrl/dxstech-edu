package quizzes

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"dxstech-edu/internal/ai"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

type GenerateRequest struct {
	Title  string `json:"title" binding:"required"`
	Notes  string `json:"notes" binding:"required"`
	Count  int    `json:"count"`
	APIKey string `json:"apiKey"`
}

type QuizSummary struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Notes         string    `json:"notes"`
	QuestionCount int       `json:"questionCount"`
	CreatedAt     time.Time `json:"createdAt"`
}

type QuizDetail struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Notes         string            `json:"notes"`
	QuestionCount int               `json:"questionCount"`
	Questions     []ai.QuizQuestion `json:"questions"`
	CreatedAt     time.Time         `json:"createdAt"`
}

type Service struct {
	db     *database.DB
	gemini *ai.GeminiClient
}

func NewService(db *database.DB, gemini *ai.GeminiClient) *Service {
	return &Service{
		db:     db,
		gemini: gemini,
	}
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/generate", s.Generate)
	r.GET("", s.List)
	r.GET("/:id", s.GetByID)
	r.DELETE("/:id", s.Delete)
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

	// Request questions from Gemini
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

	// Generate ID
	idBytes := make([]byte, 8)
	_, _ = rand.Read(idBytes)
	quizID := hex.EncodeToString(idBytes)

	now := time.Now()
	_, err = s.db.Exec(`
		INSERT INTO quizzes (id, title, notes, question_count, questions_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, quizID, req.Title, req.Notes, len(questions), string(questionsJSON), now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error guardando la evaluación en la base de datos"})
		return
	}

	c.JSON(http.StatusCreated, QuizDetail{
		ID:            quizID,
		Title:         req.Title,
		Notes:         req.Notes,
		QuestionCount: len(questions),
		Questions:     questions,
		CreatedAt:     now,
	})
}

func (s *Service) List(c *gin.Context) {
	rows, err := s.db.Query(`
		SELECT id, title, notes, question_count, created_at
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
		if err := rows.Scan(&q.ID, &q.Title, &q.Notes, &q.QuestionCount, &q.CreatedAt); err == nil {
			list = append(list, q)
		}
	}

	c.JSON(http.StatusOK, list)
}

func (s *Service) GetByID(c *gin.Context) {
	id := c.Param("id")

	var q QuizDetail
	var questionsJSON string
	err := s.db.QueryRow(`
		SELECT id, title, notes, question_count, questions_json, created_at
		FROM quizzes
		WHERE id = ?
	`, id).Scan(&q.ID, &q.Title, &q.Notes, &q.QuestionCount, &questionsJSON, &q.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evaluación no encontrada"})
		return
	}

	if err := json.Unmarshal([]byte(questionsJSON), &q.Questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al leer preguntas del quiz"})
		return
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

	c.JSON(http.StatusOK, gin.H{"message": "Evaluación eliminada correctamente", "id": id})
}

