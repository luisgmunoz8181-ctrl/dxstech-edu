package courses

import (
	"errors"
	"net/http"

	"dxstech-edu/internal/audit"
	"dxstech-edu/internal/auth"

	"github.com/gin-gonic/gin"
)

// RegisterModerationRoutes registra la cola de moderación del foro (solo administradores).
func (h *Handler) RegisterModerationRoutes(r *gin.RouterGroup) {
	r.Use(auth.RequireAuth(), auth.RequireRole("SUPERADMIN", "ADMINISTRADOR"))
	r.GET("/reports", h.HandleListForumReports)
	r.POST("/reports/:reportId/resolve", h.HandleResolveForumReport)
}

func moderationStatus(err error) int {
	switch {
	case errors.Is(err, ErrDiscussionNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrNotAllowed):
		return http.StatusForbidden
	case errors.Is(err, ErrAlreadyReported):
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

func (h *Handler) HandleDeleteDiscussion(c *gin.Context) {
	courseID, id := c.Param("id"), c.Param("discussionId")
	err := h.svc.DeleteDiscussion(c.Request.Context(), courseID, id, c.GetString("userID"), c.GetString("userRole"))
	if err != nil {
		c.JSON(moderationStatus(err), gin.H{"error": err.Error()})
		return
	}
	audit.Record(h.svc.db, c, audit.ForumDelete, "course_discussions", "Publicación %s eliminada del curso %s", id, courseID)
	c.JSON(http.StatusOK, gin.H{"message": "Publicación eliminada"})
}

func (h *Handler) HandleReportDiscussion(c *gin.Context) {
	courseID, id := c.Param("id"), c.Param("discussionId")
	if !h.requireCourseAccess(c, courseID) {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	err := h.svc.ReportDiscussion(c.Request.Context(), courseID, id, c.GetString("userID"), c.GetString("userName"), req.Reason)
	if err != nil {
		c.JSON(moderationStatus(err), gin.H{"error": err.Error()})
		return
	}
	audit.Record(h.svc.db, c, audit.ForumReport, "discussion_reports", "Publicación %s reportada en el curso %s", id, courseID)
	c.JSON(http.StatusCreated, gin.H{"message": "Gracias, el reporte fue enviado a los moderadores"})
}

func (h *Handler) HandleListForumReports(c *gin.Context) {
	list, err := h.svc.ListForumReports(c.Request.Context(), c.DefaultQuery("status", "open"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron consultar los reportes"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) HandleResolveForumReport(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Indique la acción (dismiss o remove)"})
		return
	}
	report, err := h.svc.ResolveForumReport(c.Request.Context(), c.Param("reportId"), req.Action, c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	audit.Record(h.svc.db, c, audit.ForumModerate, "discussion_reports", "Reporte %s resuelto: %s", c.Param("reportId"), req.Action)
	c.JSON(http.StatusOK, report)
}
