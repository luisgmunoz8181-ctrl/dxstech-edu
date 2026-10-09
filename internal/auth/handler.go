package auth

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc    *Service
	isProd bool
}

func NewHandler(svc *Service, isProd bool) *Handler {
	return &Handler{
		svc:    svc,
		isProd: isProd,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	limiter := NewRateLimiter(15, 1*time.Minute)

	// Public routes
	r.POST("/login", limiter.Middleware(), h.HandleLogin)
	r.POST("/logout", h.HandleLogout)
	r.POST("/forgot-password", limiter.Middleware(), h.HandleForgotPassword)
	r.POST("/reset-password", limiter.Middleware(), h.HandleResetPassword)

	// Protected routes (any authenticated user)
	protected := r.Group("")
	protected.Use(RequireAuth())
	{
		protected.GET("/me", h.HandleGetMe)
		protected.PUT("/change-password", h.HandleChangePassword)
	}

	// Administrative routes (Superadmin & Administrador)
	adminGroup := r.Group("/users")
	adminGroup.Use(RequireAuth())
	adminGroup.Use(RequireRole("SUPERADMIN", "ADMINISTRADOR"))
	{
		adminGroup.GET("", h.HandleListUsers)
		adminGroup.POST("", h.HandleCreateUser)
		adminGroup.PUT("/:id", h.HandleUpdateUser)
		adminGroup.PATCH("/:id/status", h.HandleToggleUserStatus)
		adminGroup.POST("/:id/unlock", h.HandleUnlockUser)
	}
}

func (h *Handler) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar correo y contraseña"})
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		var locked *LockedError
		if errors.As(err, &locked) {
			c.Header("Retry-After", strconv.Itoa(int(time.Until(locked.Until).Seconds())+1))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error(), "code": "ACCOUNT_LOCKED"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	maxAge := 86400 // 24 hours
	if req.RememberMe {
		maxAge = 86400 * 30 // 30 days
	}

	SetAuthCookie(c, resp.Token, maxAge, h.isProd)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleLogout(c *gin.Context) {
	ClearAuthCookie(c, h.isProd)
	c.JSON(http.StatusOK, gin.H{"message": "Sesión cerrada correctamente"})
}

func (h *Handler) HandleGetMe(c *gin.Context) {
	userID := c.GetString("userID")
	user, err := h.svc.GetMe(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user":          user,
	})
}

func (h *Handler) HandleChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos"})
		return
	}

	userID := c.GetString("userID")
	err := h.svc.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		var locked *LockedError
		if errors.As(err, &locked) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error(), "code": "ACCOUNT_LOCKED"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// El cambio invalida los tokens anteriores: se emite uno nuevo para esta sesión.
	if token, err := h.svc.IssueToken(c.Request.Context(), userID, 24*time.Hour); err == nil {
		SetAuthCookie(c, token, 86400, h.isProd)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contraseña actualizada con éxito"})
}

func (h *Handler) HandleForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar un correo electrónico válido"})
		return
	}

	rawToken, err := h.svc.ForgotPassword(c.Request.Context(), req.Email, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error procesando solicitud"})
		return
	}

	// Always return uniform friendly message to prevent enumeration
	resp := gin.H{
		"message": "Si el correo está registrado en la plataforma, recibirás instrucciones para restablecer tu contraseña.",
	}

	// In development, return token in payload to facilitate offline testing
	if !h.isProd && rawToken != "" {
		resp["devToken"] = rawToken
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token y nueva contraseña son obligatorios"})
		return
	}

	err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contraseña restablecida exitosamente. Ya puedes iniciar sesión con tu nueva contraseña."})
}

func (h *Handler) HandleListUsers(c *gin.Context) {
	roleStr := c.Query("roleId")
	roleID := 0
	if roleStr != "" {
		roleID, _ = strconv.Atoi(roleStr)
	}
	search := c.Query("search")

	users, err := h.svc.ListUsers(c.Request.Context(), roleID, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error consultando usuarios: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) HandleCreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de usuario incompletos o inválidos: " + err.Error()})
		return
	}

	adminID := c.GetString("userID")
	user, err := h.svc.CreateUser(c.Request.Context(), adminID, req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) HandleUpdateUser(c *gin.Context) {
	targetID := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de usuario inválidos: " + err.Error()})
		return
	}

	adminID := c.GetString("userID")
	user, err := h.svc.UpdateUser(c.Request.Context(), adminID, targetID, req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) HandleToggleUserStatus(c *gin.Context) {
	targetID := c.Param("id")
	adminID := c.GetString("userID")

	user, err := h.svc.ToggleUserStatus(c.Request.Context(), adminID, targetID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) HandleUnlockUser(c *gin.Context) {
	adminID := c.GetString("userID")
	user, err := h.svc.UnlockUser(c.Request.Context(), adminID, c.Param("id"), c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
