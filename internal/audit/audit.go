// Package audit registra acciones administrativas y sensibles en audit_logs.
package audit

import (
	"fmt"
	"log/slog"

	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

// Acciones auditadas (además de las de autenticación, definidas en el paquete auth).
const (
	CourseCreate     = "COURSE_CREATE"
	CourseUpdate     = "COURSE_UPDATE"
	CourseDuplicate  = "COURSE_DUPLICATE"
	CourseStatus     = "COURSE_STATUS_CHANGE"
	CourseDelete     = "COURSE_DELETE"
	ModuleCreate     = "MODULE_CREATE"
	ModuleUpdate     = "MODULE_UPDATE"
	ModuleDelete     = "MODULE_DELETE"
	LessonCreate     = "LESSON_CREATE"
	LessonUpdate     = "LESSON_UPDATE"
	LessonDelete     = "LESSON_DELETE"
	FileUpload       = "FILE_UPLOAD"
	UploadsCleanup   = "UPLOADS_CLEANUP"
	CertGenerateBulk = "CERTIFICATES_GENERATE_BULK"
	CertIssued       = "CERTIFICATE_ISSUED"
	QuizGenerate     = "QUIZ_GENERATE"
	QuizDelete       = "QUIZ_DELETE"
	EnrollAdmin      = "ENROLLMENT_ADMIN_ENROLL"
	UnenrollAdmin    = "ENROLLMENT_ADMIN_UNENROLL"
	WhatsAppSend     = "WHATSAPP_BULK_SEND"
	WhatsAppConfig   = "WHATSAPP_CONFIG_UPDATE"
	WhatsAppSync     = "WHATSAPP_SYNC_COURSES"
	WhatsAppLogsWipe = "WHATSAPP_LOGS_CLEAR"
	BackupCreate     = "BACKUP_CREATE"
	BackupDownload   = "BACKUP_DOWNLOAD"
	ForumDelete      = "FORUM_POST_DELETE"
	ForumReport      = "FORUM_POST_REPORT"
	ForumModerate    = "FORUM_REPORT_RESOLVE"
)

// Entry describe un evento de auditoría.
type Entry struct {
	UserID    string
	Action    string
	Resource  string
	IP        string
	UserAgent string
	RequestID string
	Details   string
}

// Log inserta el evento. Es de mejor esfuerzo: un fallo de auditoría nunca debe
// romper la operación de negocio, pero sí queda registrado en el log.
func Log(db *database.DB, e Entry) {
	var uid any
	if e.UserID != "" {
		uid = e.UserID
	}
	var rid any
	if e.RequestID != "" {
		rid = e.RequestID
	}
	if _, err := db.Exec(`
		INSERT INTO audit_logs (user_id, action, resource, ip_address, user_agent, details, request_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uid, e.Action, e.Resource, e.IP, e.UserAgent, e.Details, rid); err != nil {
		slog.Error("no se pudo registrar la auditoría", "action", e.Action, "error", err)
	}
}

// Record registra un evento tomando usuario, IP, agente y request-id del contexto HTTP.
func Record(db *database.DB, c *gin.Context, action, resource, format string, args ...any) {
	Log(db, Entry{
		UserID:    c.GetString("userID"),
		Action:    action,
		Resource:  resource,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		RequestID: c.GetString("requestID"),
		Details:   fmt.Sprintf(format, args...),
	})
}
