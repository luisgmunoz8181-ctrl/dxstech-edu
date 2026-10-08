package analytics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"dxstech-edu/internal/auth"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Todos los endpoints de administración requieren rol SUPERADMIN o ADMINISTRADOR
	r.Use(auth.RequireAuth())
	r.Use(auth.RequireRole("SUPERADMIN", "ADMINISTRADOR"))
	{
		r.GET("/metrics/overview", h.HandleOverview)
		r.GET("/metrics/courses", h.HandleCourseMetrics)
		r.GET("/metrics/timeline/users", h.HandleUsersTimeline)
		r.GET("/metrics/timeline/enrollments", h.HandleEnrollmentsTimeline)
		r.GET("/monitoring/students", h.HandleStudentMonitoring)
		r.GET("/reports/enrollments.csv", h.HandleExportEnrollmentsCSV)
		r.GET("/reports/certificates.csv", h.HandleExportCertificatesCSV)
	}
}

func (h *Handler) HandleOverview(c *gin.Context) {
	metrics, err := h.svc.GetOverviewMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) HandleCourseMetrics(c *gin.Context) {
	category := c.Query("category")
	metrics, err := h.svc.GetCourseMetrics(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) HandleUsersTimeline(c *gin.Context) {
	timeline, err := h.svc.GetUserRegistrationTimeline(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timeline)
}

func (h *Handler) HandleEnrollmentsTimeline(c *gin.Context) {
	timeline, err := h.svc.GetEnrollmentsTimeline(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timeline)
}

func (h *Handler) HandleStudentMonitoring(c *gin.Context) {
	opts := FilterOptions{
		CourseID:  c.Query("courseId"),
		Category:  c.Query("category"),
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}

	if l, err := strconv.Atoi(c.Query("limit")); err == nil {
		opts.Limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset")); err == nil {
		opts.Offset = o
	}

	results, err := h.svc.GetStudentMonitoring(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *Handler) HandleExportEnrollmentsCSV(c *gin.Context) {
	opts := FilterOptions{
		CourseID:  c.Query("courseId"),
		Category:  c.Query("category"),
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}

	csvBytes, err := h.svc.ExportEnrollmentsCSV(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generando reporte CSV: " + err.Error()})
		return
	}

	filename := fmt.Sprintf("reporte_matriculas_dxstech_%s.csv", time.Now().Format("20060102_1504"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}

func (h *Handler) HandleExportCertificatesCSV(c *gin.Context) {
	csvBytes, err := h.svc.ExportCertificatesCSV(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generando reporte de certificados: " + err.Error()})
		return
	}

	filename := fmt.Sprintf("reporte_certificados_dxstech_%s.csv", time.Now().Format("20060102_1504"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}
