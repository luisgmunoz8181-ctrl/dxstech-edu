package enrollments_test

import (
	"context"
	"os"
	"testing"

	"dxstech-edu/internal/certificates"
	"dxstech-edu/internal/database"
	"dxstech-edu/internal/enrollments"
)

func setupTestDB(t *testing.T) (*database.DB, string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "dxstech_enr_test_*")
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

func TestEnrollmentAndProgressFlow(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	certSvc := certificates.NewService(db)
	svc := enrollments.NewService(db, certSvc)
	ctx := context.Background()

	// 1. Initial seeded enrollment check (usr-student-01 enrolled in crs-ai-101 with 20% progress)
	studentID := "usr-student-01"
	courseID := "crs-ai-101"

	myCourses, err := svc.GetStudentEnrollments(ctx, studentID)
	if err != nil {
		t.Fatalf("Error fetching student enrollments: %v", err)
	}
	if len(myCourses) != 1 {
		t.Fatalf("Expected 1 seeded enrollment, got %d", len(myCourses))
	}
	if myCourses[0].CourseID != courseID {
		t.Errorf("Expected course %s, got %s", courseID, myCourses[0].CourseID)
	}
	if myCourses[0].ProgressPercent != 20.0 {
		t.Errorf("Expected 20%% progress, got %.1f%%", myCourses[0].ProgressPercent)
	}

	// 2. Fetch enrollment status & completed lessons
	enr, completedLessons, err := svc.GetCourseEnrollmentStatus(ctx, studentID, courseID)
	if err != nil {
		t.Fatalf("Error getting course status: %v", err)
	}
	if enr == nil || len(completedLessons) != 1 {
		t.Fatalf("Expected 1 completed lesson, got %d", len(completedLessons))
	}

	// 3. Mark Lesson 2 as completed
	updatedEnr, err := svc.ToggleLessonProgress(ctx, studentID, courseID, "lsn-ai-02", true)
	if err != nil {
		t.Fatalf("Error marking lesson completed: %v", err)
	}
	// 2 out of 5 lessons completed = 40%
	if updatedEnr.ProgressPercent != 40.0 {
		t.Errorf("Expected 40%% progress, got %.1f%%", updatedEnr.ProgressPercent)
	}

	// 4. Mark remaining lessons 3, 4, 5 as completed to reach 100%
	_, _ = svc.ToggleLessonProgress(ctx, studentID, courseID, "lsn-ai-03", true)
	_, _ = svc.ToggleLessonProgress(ctx, studentID, courseID, "lsn-ai-04", true)
	finalEnr, err := svc.ToggleLessonProgress(ctx, studentID, courseID, "lsn-ai-05", true)
	if err != nil {
		t.Fatalf("Error completing all lessons: %v", err)
	}
	if finalEnr.ProgressPercent != 100.0 {
		t.Errorf("Expected 100%% progress, got %.1f%%", finalEnr.ProgressPercent)
	}
	if finalEnr.Status != "completed" || finalEnr.CompletedAt == nil {
		t.Errorf("Enrollment should be marked as completed: %+v", finalEnr)
	}

	// 5. Verify Student Stats
	stats, err := svc.GetStudentStats(ctx, studentID)
	if err != nil {
		t.Fatalf("Error fetching stats: %v", err)
	}
	if stats.Completed != 1 || stats.TotalEnrolled != 1 {
		t.Errorf("Expected 1 completed course in stats, got: %+v", stats)
	}

	// 6. Admin list course students
	adminStudents, err := svc.ListCourseStudents(ctx, courseID)
	if err != nil {
		t.Fatalf("Error listing course students: %v", err)
	}
	if len(adminStudents) != 1 || adminStudents[0].StudentEmail != "estudiante@dxstech.edu" {
		t.Errorf("Unexpected course students list: %+v", adminStudents)
	}

	// 7. Verify Phase 4: Automatic Certificate Issuance on 100% completion
	cert, err := certSvc.GetCertificateByUserAndCourse(ctx, studentID, courseID)
	if err != nil || cert == nil {
		t.Fatalf("Expected automatic certificate to be issued upon 100%% course completion: %v", err)
	}
	if cert.StudentName != "Carlos Estudiante" {
		t.Errorf("Expected student name 'Carlos Estudiante', got '%s'", cert.StudentName)
	}
	if cert.CourseTitle == "" {
		t.Errorf("Certificate missing course title")
	}

	// Verify Official PDF Generation
	pdfBytes, err := certSvc.GenerateOfficialPDF(cert)
	if err != nil {
		t.Fatalf("Error generating official certificate PDF: %v", err)
	}
	if len(pdfBytes) < 1000 {
		t.Errorf("Generated PDF seems too small (%d bytes)", len(pdfBytes))
	}
}
