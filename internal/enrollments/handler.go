package enrollments

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"dxstech-edu/internal/audit"
	"dxstech-edu/internal/auth"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Protected routes for any authenticated user (Student or Admin)
	r.Use(auth.RequireAuth())
	{
		r.POST("/courses/:courseId", h.HandleEnrollCourse)
		r.GET("/my-courses", h.HandleGetMyCourses)
		r.GET("/courses/:courseId/progress", h.HandleGetCourseProgress)
		r.POST("/progress", h.HandleToggleLessonProgress)
		r.GET("/stats", h.HandleGetStudentStats)

		// Admin only routes
		admin := r.Group("/admin")
		admin.Use(auth.RequireRole("SUPERADMIN", "ADMINISTRADOR"))
		{
			admin.GET("/course/:courseId/students", h.HandleListCourseStudents)
			admin.POST("/enroll", h.HandleAdminEnrollStudent)
			admin.DELETE("/enroll/:id", h.HandleAdminUnenroll)
		}
	}
}

func (h *Handler) HandleEnrollCourse(c *gin.Context) {
	courseID := c.Param("courseId")
	userID := c.GetString("userID")

	enr, err := h.svc.EnrollStudent(c.Request.Context(), userID, courseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, enr)
}

func (h *Handler) HandleGetMyCourses(c *gin.Context) {
	userID := c.GetString("userID")

	courses, err := h.svc.GetStudentEnrollments(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, courses)
}

func (h *Handler) HandleGetCourseProgress(c *gin.Context) {
	courseID := c.Param("courseId")
	userID := c.GetString("userID")

	enr, completedLessons, err := h.svc.GetCourseEnrollmentStatus(c.Request.Context(), userID, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enrollment":       enr,
		"completedLessons": completedLessons,
	})
}

func (h *Handler) HandleToggleLessonProgress(c *gin.Context) {
	var req ToggleProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos"})
		return
	}

	userID := c.GetString("userID")
	enr, err := h.svc.ToggleLessonProgress(c.Request.Context(), userID, req.CourseID, req.LessonID, req.Completed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, enr)
}

func (h *Handler) HandleGetStudentStats(c *gin.Context) {
	userID := c.GetString("userID")

	stats, err := h.svc.GetStudentStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) HandleListCourseStudents(c *gin.Context) {
	courseID := c.Param("courseId")

	students, err := h.svc.ListCourseStudents(c.Request.Context(), courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, students)
}

func (h *Handler) HandleAdminEnrollStudent(c *gin.Context) {
	var req AdminEnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe especificar el estudiante y el curso"})
		return
	}

	enr, err := h.svc.EnrollStudent(c.Request.Context(), req.UserID, req.CourseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.EnrollAdmin, "enrollments", "Matrícula administrativa: usuario %s en el curso %s", req.UserID, req.CourseID)
	c.JSON(http.StatusCreated, enr)
}

func (h *Handler) HandleAdminUnenroll(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.AdminUnenrollStudent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.UnenrollAdmin, "enrollments", "Matrícula %s dada de baja por un administrador", id)
	c.JSON(http.StatusOK, gin.H{"message": "Estudiante desmatriculado correctamente"})
}
