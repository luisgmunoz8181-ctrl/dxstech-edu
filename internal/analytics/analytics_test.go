package analytics_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"dxstech-edu/internal/analytics"
	"dxstech-edu/internal/certificates"
	"dxstech-edu/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "dxstech_analytics_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	db, err := database.InitDB(tempDir)
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

func TestAnalyticsService(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	svc := analytics.NewService(db)
	certSvc := certificates.NewService(db)

	// Iniciar validando overview en base limpia
	overview, err := svc.GetOverviewMetrics(ctx)
	if err != nil {
		t.Fatalf("Error obteniendo overview: %v", err)
	}
	// Initial seed usually has 3 default users (superadmin, admin, demo student)
	if overview.TotalUsers < 1 {
		t.Errorf("Se esperaban usuarios semilla, total: %d", overview.TotalUsers)
	}

	// 1. Insertar un curso de prueba
	courseID := "crs-analytics-101"
	_, err = db.ExecContext(ctx, `
		INSERT INTO courses (id, title, code, slug, short_description, description, category, instructor_name, duration_hours, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, courseID, "Inteligencia Artificial para Directivos", "IA-DIR", "ia-para-directivos", "Resumen", "Detalle", "Tecnología", "Prof. Alan Turing", 12.0, "published", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("Error insertando curso: %v", err)
	}

	// 2. Obtener un estudiante de la semilla
	var studentID string
	err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = 'estudiante@dxstech.edu'").Scan(&studentID)
	if err != nil {
		t.Fatalf("Error obteniendo estudiante semilla: %v", err)
	}

	// 3. Insertar matrícula
	enrollID := "enr-analytics-101"
	now := time.Now()
	_, err = db.ExecContext(ctx, `
		INSERT INTO enrollments (id, user_id, course_id, status, progress_percent, enrolled_at, completed_at, last_accessed_at)
		VALUES (?, ?, ?, 'completed', 100.0, ?, ?, ?)
	`, enrollID, studentID, courseID, now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("Error creando matricula: %v", err)
	}

	// 4. Emitir un certificado
	_, err = certSvc.IssueCourseCertificate(ctx, studentID, courseID, "Carlos Estudiante", "Inteligencia Artificial para Directivos", 12.0, "Prof. Alan Turing", "https://dxstech-edu.onrender.com")
	if err != nil {
		t.Fatalf("Error emitiendo certificado: %v", err)
	}

	// 5. Probar Overview Metrics actualizado
	overview2, err := svc.GetOverviewMetrics(ctx)
	if err != nil {
		t.Fatalf("Error en overview2: %v", err)
	}
	if overview2.CompletedEnrollments < 1 {
		t.Errorf("Se esperaba al menos 1 matricula completada, got %d", overview2.CompletedEnrollments)
	}
	if overview2.TotalCertificates < 1 {
		t.Errorf("Se esperaba al menos 1 certificado emitido, got %d", overview2.TotalCertificates)
	}
	if overview2.TotalHoursDelivered < 12.0 {
		t.Errorf("Se esperaban al menos 12.0 horas entregadas, got %.1f", overview2.TotalHoursDelivered)
	}

	// 6. Probar Course Metrics
	courseMetrics, err := svc.GetCourseMetrics(ctx, "")
	if err != nil {
		t.Fatalf("Error obteniendo course metrics: %v", err)
	}
	if len(courseMetrics) == 0 {
		t.Fatalf("Se esperaban metricas de curso, got 0")
	}
	found := false
	for _, cm := range courseMetrics {
		if cm.CourseID == courseID {
			found = true
			if cm.TotalStudents != 1 {
				t.Errorf("Esperado 1 estudiante en el curso, got %d", cm.TotalStudents)
			}
			if cm.CompletedStudents != 1 {
				t.Errorf("Esperado 1 estudiante completado, got %d", cm.CompletedStudents)
			}
			if cm.CompletionRate != 100.0 {
				t.Errorf("Esperado 100%% completion rate, got %.1f", cm.CompletionRate)
			}
		}
	}
	if !found {
		t.Errorf("Curso %s no encontrado en métricas", courseID)
	}

	// 7. Probar Timelines
	regTimeline, err := svc.GetUserRegistrationTimeline(ctx)
	if err != nil {
		t.Fatalf("Error en timeline de usuarios: %v", err)
	}
	if len(regTimeline) == 0 {
		t.Errorf("Timeline de registros vacía")
	}

	enrTimeline, err := svc.GetEnrollmentsTimeline(ctx)
	if err != nil {
		t.Fatalf("Error en timeline de matrículas: %v", err)
	}
	if len(enrTimeline) == 0 {
		t.Errorf("Timeline de matrículas vacía")
	}

	// 8. Probar Monitoreo de Alumnos
	students, err := svc.GetStudentMonitoring(ctx, analytics.FilterOptions{
		CourseID: courseID,
	})
	if err != nil {
		t.Fatalf("Error en monitoreo de alumnos: %v", err)
	}
	if len(students) != 1 {
		t.Fatalf("Esperado 1 registro en monitoreo, got %d", len(students))
	}
	if students[0].StudentEmail != "estudiante@dxstech.edu" {
		t.Errorf("Email inesperado: %s", students[0].StudentEmail)
	}
	if students[0].CertificateID == "" {
		t.Errorf("Se esperaba que certificateId no estuviera vacío para matrícula completada")
	}

	// 9. Probar Exportación CSV de Matrículas
	csvEnrollments, err := svc.ExportEnrollmentsCSV(ctx, analytics.FilterOptions{})
	if err != nil {
		t.Fatalf("Error exportando CSV de matrículas: %v", err)
	}
	// Verificar BOM UTF-8 (\xEF\xBB\xBF)
	if len(csvEnrollments) < 3 || csvEnrollments[0] != 0xEF || csvEnrollments[1] != 0xBB || csvEnrollments[2] != 0xBF {
		t.Errorf("Falta BOM UTF-8 en CSV de matrículas")
	}
	csvStr := string(csvEnrollments)
	if !strings.Contains(csvStr, "estudiante@dxstech.edu") {
		t.Errorf("CSV no contiene estudiante esperado: %s", csvStr)
	}
	if !strings.Contains(csvStr, "IA-DIR") {
		t.Errorf("CSV no contiene código de curso esperado: %s", csvStr)
	}

	// 10. Probar Exportación CSV de Certificados
	csvCerts, err := svc.ExportCertificatesCSV(ctx)
	if err != nil {
		t.Fatalf("Error exportando CSV de certificados: %v", err)
	}
	if len(csvCerts) < 3 || csvCerts[0] != 0xEF || csvCerts[1] != 0xBB || csvCerts[2] != 0xBF {
		t.Errorf("Falta BOM UTF-8 en CSV de certificados")
	}
	if !strings.Contains(string(csvCerts), "Carlos Estudiante") {
		t.Errorf("CSV de certificados no contiene el nombre esperado")
	}
}
