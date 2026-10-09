package courses

import (
	"net/http"
	"time"

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
	// Public / All Authenticated users
	r.GET("", h.HandleListCourses)
	r.GET("/:id", h.HandleGetCourse)
	r.GET("/:id/discussions", auth.RequireAuth(), h.HandleGetDiscussions)
	r.POST("/:id/discussions", auth.RequireAuth(), h.HandleCreateDiscussion)
	r.GET("/:id/reviews", h.HandleGetReviews)
	r.POST("/:id/reviews", auth.RequireAuth(), h.HandleCreateReview)
	r.GET("/:id/tutor-context", auth.RequireAuth(), h.HandleGetTutorContext)

	// Admin Only Routes (Superadmin & Administrador)
	admin := r.Group("")
	admin.Use(auth.RequireAuth())
	admin.Use(auth.RequireRole("SUPERADMIN", "ADMINISTRADOR"))
	{
		admin.POST("", h.HandleCreateCourse)
		admin.PUT("/:id", h.HandleUpdateCourse)
		admin.POST("/:id/duplicate", h.HandleDuplicateCourse)
		admin.PATCH("/:id/status", h.HandleChangeStatus)
		admin.DELETE("/:id", h.HandleDeleteCourse)

		// Modules
		admin.POST("/:id/modules", h.HandleCreateModule)
		admin.PUT("/modules/:moduleId", h.HandleUpdateModule)
		admin.DELETE("/modules/:moduleId", h.HandleDeleteModule)

		// Lessons
		admin.POST("/modules/:moduleId/lessons", h.HandleCreateLesson)
		admin.PUT("/lessons/:lessonId", h.HandleUpdateLesson)
		admin.DELETE("/lessons/:lessonId", h.HandleDeleteLesson)

		// File Upload
		admin.POST("/upload", h.HandleUpload)
		admin.POST("/uploads/cleanup", h.HandleCleanupUploads)
	}
}

func (h *Handler) HandleListCourses(c *gin.Context) {
	role := c.GetString("userRole")
	search := c.Query("search")
	category := c.Query("category")
	status := c.Query("status")

	courses, err := h.svc.ListCourses(c.Request.Context(), role, search, category, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, courses)
}

func (h *Handler) HandleGetCourse(c *gin.Context) {
	id := c.Param("id")
	role := c.GetString("userRole")

	course, err := h.svc.GetCourse(c.Request.Context(), id, role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, course)
}

func (h *Handler) HandleCreateCourse(c *gin.Context) {
	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos: " + err.Error()})
		return
	}

	userID := c.GetString("userID")
	course, err := h.svc.CreateCourse(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.CourseCreate, "courses", "Curso creado: %s (%s)", course.Code, course.ID)
	c.JSON(http.StatusCreated, course)
}

func (h *Handler) HandleUpdateCourse(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos: " + err.Error()})
		return
	}

	course, err := h.svc.UpdateCourse(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.CourseUpdate, "courses", "Curso actualizado: %s (%s)", course.Code, course.ID)
	c.JSON(http.StatusOK, course)
}

func (h *Handler) HandleDuplicateCourse(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	course, err := h.svc.DuplicateCourse(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.CourseDuplicate, "courses", "Curso %s duplicado como %s", id, course.ID)
	c.JSON(http.StatusCreated, course)
}

func (h *Handler) HandleChangeStatus(c *gin.Context) {
	id := c.Param("id")
	var req ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe especificar un estado válido ('draft', 'published', 'archived')"})
		return
	}

	course, err := h.svc.ChangeCourseStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.CourseStatus, "courses", "Curso %s (%s) pasó a estado %s", course.Code, course.ID, req.Status)
	c.JSON(http.StatusOK, course)
}

func (h *Handler) HandleDeleteCourse(c *gin.Context) {
	id := c.Param("id")
	var code, title string
	_ = h.svc.db.QueryRowContext(c.Request.Context(), `SELECT code, title FROM courses WHERE id = ?`, id).Scan(&code, &title)
	if err := h.svc.DeleteCourse(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.CourseDelete, "courses", "Curso eliminado: %s \"%s\" (%s)", code, title, id)
	c.JSON(http.StatusOK, gin.H{"message": "Curso eliminado correctamente"})
}

// Module Handlers

func (h *Handler) HandleCreateModule(c *gin.Context) {
	courseID := c.Param("id")
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos: " + err.Error()})
		return
	}

	mod, err := h.svc.CreateModule(c.Request.Context(), courseID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.ModuleCreate, "course_modules", "Módulo %s creado en el curso %s", mod.ID, courseID)
	c.JSON(http.StatusCreated, mod)
}

func (h *Handler) HandleUpdateModule(c *gin.Context) {
	moduleID := c.Param("moduleId")
	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos: " + err.Error()})
		return
	}

	mod, err := h.svc.UpdateModule(c.Request.Context(), moduleID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.ModuleUpdate, "course_modules", "Módulo actualizado: %s", moduleID)
	c.JSON(http.StatusOK, mod)
}

func (h *Handler) HandleDeleteModule(c *gin.Context) {
	moduleID := c.Param("moduleId")
	if err := h.svc.DeleteModule(c.Request.Context(), moduleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.ModuleDelete, "course_modules", "Módulo eliminado: %s", moduleID)
	c.JSON(http.StatusOK, gin.H{"message": "Módulo eliminado correctamente"})
}

// Lesson Handlers

func (h *Handler) HandleCreateLesson(c *gin.Context) {
	moduleID := c.Param("moduleId")
	var req CreateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos: " + err.Error()})
		return
	}

	lesson, err := h.svc.CreateLesson(c.Request.Context(), moduleID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.LessonCreate, "lessons", "Lección %s creada en el módulo %s", lesson.ID, moduleID)
	c.JSON(http.StatusCreated, lesson)
}

func (h *Handler) HandleUpdateLesson(c *gin.Context) {
	lessonID := c.Param("lessonId")
	var req UpdateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos: " + err.Error()})
		return
	}

	lesson, err := h.svc.UpdateLesson(c.Request.Context(), lessonID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.LessonUpdate, "lessons", "Lección actualizada: %s", lessonID)
	c.JSON(http.StatusOK, lesson)
}

func (h *Handler) HandleDeleteLesson(c *gin.Context) {
	lessonID := c.Param("lessonId")
	if err := h.svc.DeleteLesson(c.Request.Context(), lessonID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.LessonDelete, "lessons", "Lección eliminada: %s", lessonID)
	c.JSON(http.StatusOK, gin.H{"message": "Lección eliminada correctamente"})
}

// Upload Handler

func (h *Handler) HandleUpload(c *gin.Context) {
	// Corta la lectura del cuerpo más allá del límite (+1 MB de margen multipart).
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes+1<<20)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe adjuntar un archivo en el campo 'file' (máx. 25 MB)"})
		return
	}

	url, err := h.svc.SaveUpload(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audit.Record(h.svc.db, c, audit.FileUpload, "uploads", "Archivo subido: %s (%d bytes) como %s", file.Filename, file.Size, url)
	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// Discussions Handlers

// requireCourseAccess responde 403 si el usuario no es admin ni está matriculado.
func (h *Handler) requireCourseAccess(c *gin.Context, courseID string) bool {
	if h.svc.db.UserCanAccessCourse(c.Request.Context(), c.GetString("userID"), c.GetString("userRole"), courseID) {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "Debes estar matriculado en este curso para acceder a esta función"})
	return false
}

func (h *Handler) HandleGetDiscussions(c *gin.Context) {
	courseID := c.Param("id")
	if !h.requireCourseAccess(c, courseID) {
		return
	}
	lessonID := c.Query("lessonId")

	list, err := h.svc.GetDiscussions(c.Request.Context(), courseID, lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) HandleCreateDiscussion(c *gin.Context) {
	courseID := c.Param("id")
	if !h.requireCourseAccess(c, courseID) {
		return
	}
	userID := c.GetString("userID")
	userName := c.GetString("userName")
	userRole := c.GetString("userRole")

	var req CreateDiscussionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El mensaje es requerido"})
		return
	}

	d, err := h.svc.CreateDiscussion(c.Request.Context(), courseID, req.LessonID, userID, userName, userRole, req.Message, req.Title, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

// Reviews Handlers

func (h *Handler) HandleGetReviews(c *gin.Context) {
	courseID := c.Param("id")

	summary, err := h.svc.GetReviews(c.Request.Context(), courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) HandleCreateReview(c *gin.Context) {
	courseID := c.Param("id")
	if !h.requireCourseAccess(c, courseID) {
		return
	}
	userID := c.GetString("userID")
	userName := c.GetString("userName")

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Calificación válida (1 a 5) requerida"})
		return
	}

	rev, err := h.svc.CreateReview(c.Request.Context(), courseID, userID, userName, req.Rating, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rev)
}

func (h *Handler) HandleGetTutorContext(c *gin.Context) {
	courseID := c.Param("id")
	if !h.requireCourseAccess(c, courseID) {
		return
	}
	tutorCtx, err := h.svc.GetCourseSummaryContext(c.Request.Context(), courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"courseId": courseID,
		"context":  tutorCtx,
	})
}

// HandleCleanupUploads detecta (o, con ?apply=true, elimina) los archivos de
// /uploads que ningún curso o lección referencia. Por defecto solo informa.
// Ignora los archivos modificados hace menos de una hora.
func (h *Handler) HandleCleanupUploads(c *gin.Context) {
	apply := c.Query("apply") == "true"
	orphans, err := h.svc.CleanOrphanUploads(c.Request.Context(), apply, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo revisar el directorio de archivos"})
		return
	}
	var total int64
	for _, o := range orphans {
		total += o.SizeBytes
	}
	if apply {
		audit.Record(h.svc.db, c, audit.UploadsCleanup, "uploads", "Limpieza de archivos huérfanos: %d archivos, %d bytes", len(orphans), total)
	}
	c.JSON(http.StatusOK, gin.H{"dryRun": !apply, "count": len(orphans), "totalBytes": total, "files": orphans})
}
