package certificates_test

import (
	"context"
	"os"
	"testing"

	"dxstech-edu/internal/certificates"
	"dxstech-edu/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "dxstech_cert_test_*")
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

	return db, cleanup
}

func TestCertificateIssuanceAndPDF(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	svc := certificates.NewService(db)
	ctx := context.Background()

	userID := "usr-test-01"
	courseID := "crs-ai-101"
	studentName := "María José Peña Cárdenas"
	courseTitle := "Especialización en Agentes Autónomos & LLMs"
	durationHours := 15.5
	instructorName := "Dr. Alejandro Gómez"
	baseURL := "https://dxstech-edu.onrender.com"

	// 1. Issue certificate
	cert, err := svc.IssueCourseCertificate(ctx, userID, courseID, studentName, courseTitle, durationHours, instructorName, baseURL)
	if err != nil {
		t.Fatalf("Failed to issue certificate: %v", err)
	}

	if cert == nil || cert.ID == "" {
		t.Fatalf("Expected valid certificate, got nil or empty ID")
	}

	if cert.StudentName != studentName {
		t.Errorf("Expected student name %s, got %s", studentName, cert.StudentName)
	}

	// 2. Fetch by ID
	fetched, err := svc.GetCertificateByID(ctx, cert.ID)
	if err != nil {
		t.Fatalf("Failed to fetch certificate by ID: %v", err)
	}
	if fetched.ID != cert.ID || fetched.CourseTitle != courseTitle {
		t.Errorf("Fetched certificate mismatch: %+v", fetched)
	}

	// 3. Fetch by User and Course (idempotency check)
	byCourse, err := svc.GetCertificateByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		t.Fatalf("Failed to fetch certificate by user and course: %v", err)
	}
	if byCourse.ID != cert.ID {
		t.Errorf("Expected same certificate ID, got %s vs %s", byCourse.ID, cert.ID)
	}

	// 4. Re-issuing should return the existing certificate without error
	reissued, err := svc.IssueCourseCertificate(ctx, userID, courseID, studentName, courseTitle, durationHours, instructorName, baseURL)
	if err != nil {
		t.Fatalf("Re-issuance failed: %v", err)
	}
	if reissued.ID != cert.ID {
		t.Errorf("Re-issued certificate ID should match original: %s != %s", reissued.ID, cert.ID)
	}

	// 5. List User Certificates
	userCerts, err := svc.ListUserCertificates(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to list user certificates: %v", err)
	}
	if len(userCerts) != 1 {
		t.Errorf("Expected 1 certificate in user list, got %d", len(userCerts))
	}

	// 6. Generate Official PDF
	pdfBytes, err := svc.GenerateOfficialPDF(cert)
	if err != nil {
		t.Fatalf("Failed to generate official PDF: %v", err)
	}

	if len(pdfBytes) < 2000 {
		t.Errorf("PDF length too small (%d bytes), likely corrupt", len(pdfBytes))
	}

	// Verify PDF Header signature (%PDF-1.)
	if string(pdfBytes[:5]) != "%PDF-" {
		t.Errorf("Generated file does not have valid PDF magic header")
	}
}
