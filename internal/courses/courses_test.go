package courses_test

import (
	"context"
	"os"
	"testing"

	"dxstech-edu/internal/courses"
	"dxstech-edu/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "dxstech_courses_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	db, err := database.InitDB(tempDir, true)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to init db: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return db, tempDir, cleanup
}

func TestCourseLifecycle(t *testing.T) {
	db, tempDir, cleanup := setupTestDB(t)
	defer cleanup()

	svc := courses.NewService(db, tempDir)
	ctx := context.Background()

	// 1. Initial list should include seeded courses (1 published, 1 draft)
	publicCourses, err := svc.ListCourses(ctx, "ESTUDIANTE", "", "", "")
	if err != nil {
		t.Fatalf("Error listing public courses: %v", err)
	}
	if len(publicCourses) != 1 {
		t.Errorf("Expected 1 published course for student, got %d", len(publicCourses))
	}

	allCourses, err := svc.ListCourses(ctx, "ADMINISTRADOR", "", "", "")
	if err != nil {
		t.Fatalf("Error listing all courses: %v", err)
	}
	if len(allCourses) != 2 {
		t.Errorf("Expected 2 courses for admin, got %d", len(allCourses))
	}

	// 2. Create new course
	createReq := courses.CreateCourseRequest{
		Title:              "Ciberseguridad Ofensiva & Hacking Ético",
		Code:               "DXS-SEC-301",
		ShortDescription:   "Aprende análisis de vulnerabilidades y pentesting.",
		Description:        "Curso intensivo con laboratorios prácticos.",
		Category:           "Ciberseguridad",
		InstructorName:     "Ing. Marcos Paz",
		DurationHours:      15.0,
		Level:              "Avanzado",
		Requirements:       "Redes y Linux básico",
		LearningObjectives: "Identificar fallos; Ejecutar auditorías; Mitigar riesgos",
	}

	newCourse, err := svc.CreateCourse(ctx, createReq, "usr-admin-01")
	if err != nil {
		t.Fatalf("Error creating course: %v", err)
	}
	if newCourse.ID == "" || newCourse.Status != "draft" {
		t.Errorf("Invalid created course: %+v", newCourse)
	}

	// 3. Add module to course
	modReq := courses.CreateModuleRequest{
		Title:       "Módulo 1: Reconocimiento y Escaneo de Puertos",
		Description: "Técnicas pasivas y activas con nmap.",
		OrderIndex:  1,
	}
	mod, err := svc.CreateModule(ctx, newCourse.ID, modReq)
	if err != nil {
		t.Fatalf("Error creating module: %v", err)
	}
	if mod.ID == "" || mod.CourseID != newCourse.ID {
		t.Errorf("Invalid module: %+v", mod)
	}

	// 4. Add lesson to module
	lsnReq := courses.CreateLessonRequest{
		Title:           "Escaneo con Nmap y Detección de Versiones",
		Description:     "Comandos fundamentales y evasión de firewalls.",
		ContentType:     "youtube",
		ContentURL:      "https://www.youtube.com/watch?v=4t4kBkBn9Sk",
		DurationMinutes: 25,
		OrderIndex:      1,
		IsFreePreview:   true,
	}
	lsn, err := svc.CreateLesson(ctx, mod.ID, lsnReq)
	if err != nil {
		t.Fatalf("Error creating lesson: %v", err)
	}
	if lsn.ID == "" || lsn.ModuleID != mod.ID {
		t.Errorf("Invalid lesson: %+v", lsn)
	}

	// 5. Publish course
	publishedCourse, err := svc.ChangeCourseStatus(ctx, newCourse.ID, "published")
	if err != nil {
		t.Fatalf("Error publishing course: %v", err)
	}
	if publishedCourse.Status != "published" || publishedCourse.PublishedAt == nil {
		t.Errorf("Course status should be published with valid date: %+v", publishedCourse)
	}

	// 6. Duplicate course
	dupCourse, err := svc.DuplicateCourse(ctx, newCourse.ID, "usr-admin-01")
	if err != nil {
		t.Fatalf("Error duplicating course: %v", err)
	}
	if len(dupCourse.Modules) != 1 || len(dupCourse.Modules[0].Lessons) != 1 {
		t.Errorf("Duplicated course should have 1 module and 1 lesson, got: %+v", dupCourse)
	}

	// 7. Delete duplicated course
	err = svc.DeleteCourse(ctx, dupCourse.ID)
	if err != nil {
		t.Fatalf("Error deleting duplicated course: %v", err)
	}

	// 8. Test Discussions / Forum
	disc, err := svc.CreateDiscussion(ctx, newCourse.ID, lsn.ID, "usr-student-01", "Carlos Estudiante", "ESTUDIANTE", "¿Cuáles son las herramientas recomendadas para escaneo de puertos?")
	if err != nil {
		t.Fatalf("Error creating discussion: %v", err)
	}
	if disc.ID == "" || disc.Message == "" {
		t.Errorf("Invalid discussion created: %+v", disc)
	}

	discs, err := svc.GetDiscussions(ctx, newCourse.ID, lsn.ID)
	if err != nil || len(discs) == 0 {
		t.Fatalf("Expected discussions list, got error: %v, count: %d", err, len(discs))
	}

	// 9. Test Reviews / Ratings
	rev, err := svc.CreateReview(ctx, newCourse.ID, "usr-student-01", "Carlos Estudiante", 5, "Excelente contenido y muy práctico.")
	if err != nil {
		t.Fatalf("Error creating review: %v", err)
	}
	if rev.Rating != 5 {
		t.Errorf("Expected 5 star rating, got %d", rev.Rating)
	}

	revSummary, err := svc.GetReviews(ctx, newCourse.ID)
	if err != nil || revSummary.TotalReviews != 1 || revSummary.AverageRating != 5.0 {
		t.Errorf("Invalid review summary: %+v", revSummary)
	}

	// 10. Test Course Summary Context
	summaryCtx, err := svc.GetCourseSummaryContext(ctx, newCourse.ID)
	if err != nil || summaryCtx == "" {
		t.Errorf("Expected summary context, got error: %v, text: %s", err, summaryCtx)
	}
}
