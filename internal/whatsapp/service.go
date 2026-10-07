package whatsapp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"dxstech-edu/internal/ai"
	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

type BotConfig struct {
	IsActive      bool   `json:"isActive"`
	KnowledgeBase string `json:"knowledgeBase"`
	StrictMode    bool   `json:"strictMode"`
	RespondGroups bool   `json:"respondGroups"`
	Provider      string `json:"provider"`
}

type BotStatus struct {
	Connected      bool      `json:"connected"`
	Provider       string    `json:"provider"`
	PhoneNumber    string    `json:"phoneNumber"`
	QRCode         string    `json:"qrCode,omitempty"`
	Config         BotConfig `json:"config"`
	PendingQueue   int       `json:"pendingQueue"`
	TotalSentToday int       `json:"totalSentToday"`
	LastHeartbeat  time.Time `json:"lastHeartbeat"`
}

type SendRequest struct {
	Recipients []string `json:"recipients" binding:"required"`
	Message    string   `json:"message" binding:"required"`
	DelaySec   int      `json:"delaySec"`
	ConsentAck bool     `json:"consentAck"`
}

type ChatSimulateRequest struct {
	Message string `json:"message" binding:"required"`
	APIKey  string `json:"apiKey"`
}

type LogItem struct {
	ID          int64     `json:"id"`
	Recipient   string    `json:"recipient"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	ErrorDetail string    `json:"errorDetail,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Service manages WhatsApp bot configuration, simulated/real gateway connection, and message queue
type Service struct {
	db        *database.DB
	gemini    *ai.GeminiClient
	cfg       *config.Config
	mu        sync.RWMutex
	queueMu   sync.Mutex
	queueLen  int
	isSending bool
}

func NewService(db *database.DB, gemini *ai.GeminiClient, cfg *config.Config) *Service {
	return &Service{
		db:     db,
		gemini: gemini,
		cfg:    cfg,
	}
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	r.Use(s.productionAuthMiddleware())
	r.GET("/status", s.GetStatus)
	r.POST("/config", s.UpdateConfig)
	r.POST("/send", s.SendBulkMessages)
	r.POST("/simulate-chat", s.SimulateChat)
	r.GET("/logs", s.GetLogs)
	r.DELETE("/logs", s.ClearLogs)
}

func (s *Service) productionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// IN DEVELOPMENT: completely free access, no auth required!
		if s.cfg == nil || s.cfg.IsDevelopment() {
			c.Next()
			return
		}

		// IN PRODUCTION: require admin authorization
		adminPass := "dxstech2026"
		reqPass := c.GetHeader("X-Admin-Password")
		user, pass, hasBasic := c.Request.BasicAuth()
		if reqPass == adminPass || (hasBasic && (pass == adminPass || user == "admin" && pass == adminPass)) {
			c.Next()
			return
		}

		c.Header("WWW-Authenticate", `Basic realm="DxSTech Edu WhatsApp Admin"`)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Acceso protegido en producción. Ingrese credenciales de administrador para WhatsApp.",
		})
	}
}

func (s *Service) GetConfig() (BotConfig, error) {
	var cfg BotConfig
	var isActive, strictMode, respondGroups int
	err := s.db.QueryRow(`
		SELECT is_active, knowledge_base, strict_mode, respond_groups, provider
		FROM whatsapp_config WHERE id = 1
	`).Scan(&isActive, &cfg.KnowledgeBase, &strictMode, &respondGroups, &cfg.Provider)

	if err != nil {
		return BotConfig{
			IsActive:      true,
			KnowledgeBase: "DxSTech Edu - Formación en Inteligencia Artificial y Programación.",
			StrictMode:    true,
			RespondGroups: false,
			Provider:      "Simulador / Baileys Gateway",
		}, nil
	}

	cfg.IsActive = isActive == 1
	cfg.StrictMode = strictMode == 1
	cfg.RespondGroups = respondGroups == 1
	return cfg, nil
}

func (s *Service) GetStatus(c *gin.Context) {
	cfg, err := s.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cargando configuración"})
		return
	}

	var totalToday int
	_ = s.db.QueryRow(`
		SELECT COUNT(*) FROM whatsapp_logs 
		WHERE date(created_at) = date('now')
	`).Scan(&totalToday)

	s.mu.RLock()
	curQueue := s.queueLen
	s.mu.RUnlock()

	// High quality simulated SVG QR code for visual connection
	qrCodeSVG := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" class="w-full h-full text-slate-900"><path fill="currentColor" d="M10 10h30v30H10zm5 5v20h20V15zm5 5h10v10H20zm40-15h30v30H60zm5 5v20h20V15zm5 5h10v10H70zM10 60h30v30H10zm5 5v20h20V65zm5 5h10v10H20zm40 0h10v10H60zm10 10h10v10H70zm10-10h10v10H80zm0 10h10v10H80zm-20-20h10v10H60zm20 0h10v10H80z"/></svg>`

	status := BotStatus{
		Connected:      true,
		Provider:       cfg.Provider,
		PhoneNumber:    "+57 (300) 890-DXSTECH",
		QRCode:         qrCodeSVG,
		Config:         cfg,
		PendingQueue:   curQueue,
		TotalSentToday: totalToday,
		LastHeartbeat:  time.Now(),
	}

	c.JSON(http.StatusOK, status)
}

func (s *Service) UpdateConfig(c *gin.Context) {
	var req BotConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuración inválida: " + err.Error()})
		return
	}

	isActive := 0
	if req.IsActive {
		isActive = 1
	}
	strictMode := 0
	if req.StrictMode {
		strictMode = 1
	}
	respondGroups := 0
	if req.RespondGroups {
		respondGroups = 1
	}

	_, err := s.db.Exec(`
		UPDATE whatsapp_config
		SET is_active = ?, knowledge_base = ?, strict_mode = ?, respond_groups = ?, provider = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, isActive, req.KnowledgeBase, strictMode, respondGroups, req.Provider)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar la configuración de WhatsApp"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuración de WhatsApp guardada exitosamente", "config": req})
}

func (s *Service) SendBulkMessages(c *gin.Context) {
	var req SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Solicitud de envío inválida: " + err.Error()})
		return
	}

	if len(req.Recipients) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe incluir al menos un destinatario"})
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El mensaje no puede estar vacío"})
		return
	}

	if req.DelaySec < 1 {
		req.DelaySec = 2 // default safe rate-limit
	}

	// Dispatch worker in background to simulate controlled rate-limiting delivery
	go func(recipients []string, message string, delay time.Duration) {
		s.queueMu.Lock()
		s.mu.Lock()
		s.queueLen += len(recipients)
		s.mu.Unlock()
		s.queueMu.Unlock()

		for _, rec := range recipients {
			phone := strings.TrimSpace(rec)
			if phone == "" {
				s.mu.Lock()
				s.queueLen--
				s.mu.Unlock()
				continue
			}

			// Simulate transmission delay respecting rate limits
			time.Sleep(delay)

			status := "Entregado"
			errDetail := ""
			// Simple validation
			if len(phone) < 7 {
				status = "Fallido"
				errDetail = "Número de teléfono incompleto o formato erróneo"
			}

			_, _ = s.db.Exec(`
				INSERT INTO whatsapp_logs (recipient, message, status, error_detail, created_at)
				VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
			`, phone, message, status, errDetail)

			s.mu.Lock()
			if s.queueLen > 0 {
				s.queueLen--
			}
			s.mu.Unlock()
		}
	}(req.Recipients, req.Message, time.Duration(req.DelaySec)*time.Second)

	c.JSON(http.StatusAccepted, gin.H{
		"message":  fmt.Sprintf("%d mensaje(s) programados para envío en cola con rate limit", len(req.Recipients)),
		"inQueue":  len(req.Recipients),
		"delaySec": req.DelaySec,
	})
}

func (s *Service) SimulateChat(c *gin.Context) {
	var req ChatSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mensaje requerido"})
		return
	}

	apiKey := req.APIKey
	if strings.TrimSpace(apiKey) == "" {
		apiKey = c.GetHeader("X-Gemini-API-Key")
	}

	cfg, err := s.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo configuración del bot"})
		return
	}

	if !cfg.IsActive {
		c.JSON(http.StatusOK, gin.H{"reply": "🔴 El Asistente Virtual se encuentra pausado temporalmente."})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	reply, err := s.gemini.AskChatbot(ctx, apiKey, req.Message, cfg.KnowledgeBase, cfg.StrictMode)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"reply": fmt.Sprintf("⚠️ Hubo un error al generar la respuesta con IA: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

func (s *Service) GetLogs(c *gin.Context) {
	rows, err := s.db.Query(`
		SELECT id, recipient, message, status, COALESCE(error_detail, ''), created_at
		FROM whatsapp_logs
		ORDER BY id DESC
		LIMIT 50
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error consultando bitácora de mensajes"})
		return
	}
	defer rows.Close()

	logs := make([]LogItem, 0)
	for rows.Next() {
		var item LogItem
		if err := rows.Scan(&item.ID, &item.Recipient, &item.Message, &item.Status, &item.ErrorDetail, &item.CreatedAt); err == nil {
			logs = append(logs, item)
		}
	}

	c.JSON(http.StatusOK, logs)
}

func (s *Service) ClearLogs(c *gin.Context) {
	_, err := s.db.Exec(`DELETE FROM whatsapp_logs`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al vaciar la bitácora: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bitácora vaciada correctamente"})
}

