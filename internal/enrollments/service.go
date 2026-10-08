package enrollments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"dxstech-edu/internal/certificates"
	"dxstech-edu/internal/database"
)

type CertificateIssuer interface {
	IssueCourseCertificate(ctx context.Context, userID, courseID, studentName, courseTitle string, durationHours float64, instructorName, baseURL string) (*certificates.IssuedCertificate, error)
}

type Service struct {
	db         *database.DB
	certIssuer CertificateIssuer
}

func NewService(db *database.DB, certIssuer CertificateIssuer) *Service {
	return &Service{db: db, certIssuer: certIssuer}
}

func (s *Service) EnrollStudent(ctx context.Context, userID, courseID string) (*Enrollment, error) {
	// Verify course exists
	var status string
	err := s.db.QueryRowContext(ctx, "SELECT status FROM courses WHERE id = ?", courseID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("curso no encontrado")
		}
		return nil, fmt.Errorf("error al verificar curso: %w", err)
	}

	// Check if already enrolled
	var existingID, currentStatus string
	err = s.db.QueryRowContext(ctx, "SELECT id, status FROM enrollments WHERE user_id = ? AND course_id = ?", userID, courseID).Scan(&existingID, &currentStatus)
	if err == nil {
		if currentStatus == "dropped" {
			now := time.Now()
			_, _ = s.db.ExecContext(ctx, "UPDATE enrollments SET status = 'active', last_accessed_at = ? WHERE id = ?", now, existingID)
		}
		return s.getEnrollmentByID(ctx, existingID)
	}

	enrID := fmt.Sprintf("enr-%d", time.Now().UnixNano())
	now := time.Now()

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO enrollments (id, user_id, course_id, status, progress_percent, enrolled_at, last_accessed_at)
		VALUES (?, ?, ?, 'active', 0.0, ?, ?)
	`, enrID, userID, courseID, now, now)
	if err != nil {
		return nil, fmt.Errorf("error al matricular estudiante: %w", err)
	}

	return s.getEnrollmentByID(ctx, enrID)
}

func (s *Service) GetStudentEnrollments(ctx context.Context, userID string) ([]Enrollment, error) {
	query := `
		SELECT 
			e.id, e.user_id, e.course_id, e.status, e.progress_percent,
			e.enrolled_at, e.completed_at, e.last_accessed_at, coalesce(e.last_lesson_id, ''),
			c.title, c.code, c.thumbnail_url, c.category, c.level,
			c.instructor_name, c.duration_hours,
			(SELECT COUNT(*) FROM lesson_progress lp WHERE lp.user_id = e.user_id AND lp.course_id = e.course_id) as completed_count,
			(SELECT COUNT(*) FROM lessons l WHERE l.course_id = e.course_id) as total_count
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.user_id = ? AND e.status != 'dropped'
		ORDER BY e.last_accessed_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error consultando matrículas del estudiante: %w", err)
	}
	defer rows.Close()

	var result []Enrollment
	for rows.Next() {
		var e Enrollment
		var compAt sql.NullTime

		err := rows.Scan(
			&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.ProgressPercent,
			&e.EnrolledAt, &compAt, &e.LastAccessedAt, &e.LastLessonID,
			&e.CourseTitle, &e.CourseCode, &e.CourseThumbnail, &e.CourseCategory, &e.CourseLevel,
			&e.CourseInstructor, &e.CourseDurationHours,
			&e.CompletedLessons, &e.TotalLessons,
		)
		if err != nil {
			return nil, err
		}
		if compAt.Valid {
			e.CompletedAt = &compAt.Time
		}

		result = append(result, e)
	}

	if result == nil {
		result = []Enrollment{}
	}

	return result, nil
}

func (s *Service) GetCourseEnrollmentStatus(ctx context.Context, userID, courseID string) (*Enrollment, []string, error) {
	var enr Enrollment
	var compAt sql.NullTime

	query := `
		SELECT 
			e.id, e.user_id, e.course_id, e.status, e.progress_percent,
			e.enrolled_at, e.completed_at, e.last_accessed_at, coalesce(e.last_lesson_id, ''),
			(SELECT COUNT(*) FROM lesson_progress lp WHERE lp.user_id = e.user_id AND lp.course_id = e.course_id) as completed_count,
			(SELECT COUNT(*) FROM lessons l WHERE l.course_id = e.course_id) as total_count
		FROM enrollments e
		WHERE e.user_id = ? AND e.course_id = ? AND e.status != 'dropped'
	`
	err := s.db.QueryRowContext(ctx, query, userID, courseID).Scan(
		&enr.ID, &enr.UserID, &enr.CourseID, &enr.Status, &enr.ProgressPercent,
		&enr.EnrolledAt, &compAt, &enr.LastAccessedAt, &enr.LastLessonID,
		&enr.CompletedLessons, &enr.TotalLessons,
	)

	var pEnr *Enrollment
	if err == nil {
		if compAt.Valid {
			enr.CompletedAt = &compAt.Time
		}
		pEnr = &enr
	}

	// Fetch completed lesson IDs
	rows, err := s.db.QueryContext(ctx, `
		SELECT lesson_id FROM lesson_progress WHERE user_id = ? AND course_id = ?
	`, userID, courseID)
	if err != nil {
		return pEnr, []string{}, nil
	}
	defer rows.Close()

	var completedLessonIDs []string
	for rows.Next() {
		var lID string
		if err := rows.Scan(&lID); err == nil {
			completedLessonIDs = append(completedLessonIDs, lID)
		}
	}
	if completedLessonIDs == nil {
		completedLessonIDs = []string{}
	}

	return pEnr, completedLessonIDs, nil
}

func (s *Service) ToggleLessonProgress(ctx context.Context, userID, courseID, lessonID string, completed bool) (*Enrollment, error) {
	// Auto-enroll if not enrolled yet
	enr, err := s.EnrollStudent(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	if completed {
		prgID := fmt.Sprintf("prg-%s-%s", userID, lessonID)
		_, err = s.db.ExecContext(ctx, `
			INSERT OR IGNORE INTO lesson_progress (id, user_id, course_id, lesson_id, completed, completed_at)
			VALUES (?, ?, ?, ?, 1, ?)
		`, prgID, userID, courseID, lessonID, now)
	} else {
		_, err = s.db.ExecContext(ctx, `
			DELETE FROM lesson_progress WHERE user_id = ? AND lesson_id = ?
		`, userID, lessonID)
	}
	if err != nil {
		return nil, fmt.Errorf("error actualizando progreso de lección: %w", err)
	}

	// Recalculate Course Progress
	var completedCount, totalCount int
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM lesson_progress WHERE user_id = ? AND course_id = ?", userID, courseID).Scan(&completedCount)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM lessons WHERE course_id = ?", courseID).Scan(&totalCount)

	progressPercent := 0.0
	if totalCount > 0 {
		progressPercent = (float64(completedCount) / float64(totalCount)) * 100.0
		if progressPercent > 100.0 {
			progressPercent = 100.0
		}
	}

	status := "active"
	var compAt *time.Time
	if progressPercent >= 100.0 {
		status = "completed"
		compAt = &now

		if s.certIssuer != nil {
			var firstName, lastName, courseTitle, instructorName string
			var durationHours float64
			_ = s.db.QueryRowContext(ctx, "SELECT first_name, last_name FROM users WHERE id = ?", userID).Scan(&firstName, &lastName)
			_ = s.db.QueryRowContext(ctx, "SELECT title, duration_hours, instructor_name FROM courses WHERE id = ?", courseID).Scan(&courseTitle, &durationHours, &instructorName)

			studentName := strings.TrimSpace(firstName + " " + lastName)
			if studentName == "" {
				studentName = "Estudiante DxSTech"
			}
			if instructorName == "" {
				instructorName = "Dirección Académica DxSTech"
			}

			_, _ = s.certIssuer.IssueCourseCertificate(ctx, userID, courseID, studentName, courseTitle, durationHours, instructorName, "")
		}
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE enrollments SET
			progress_percent = ?,
			status = ?,
			completed_at = ?,
			last_lesson_id = ?,
			last_accessed_at = ?
		WHERE id = ?
	`, progressPercent, status, compAt, lessonID, now, enr.ID)
	if err != nil {
		return nil, fmt.Errorf("error actualizando matrícula: %w", err)
	}

	return s.getEnrollmentByID(ctx, enr.ID)
}

func (s *Service) GetStudentStats(ctx context.Context, userID string) (*StudentStats, error) {
	stats := &StudentStats{}

	_ = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM enrollments WHERE user_id = ? AND status != 'dropped'
	`, userID).Scan(&stats.TotalEnrolled)

	_ = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM enrollments WHERE user_id = ? AND status = 'active'
	`, userID).Scan(&stats.InProgress)

	_ = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM enrollments WHERE user_id = ? AND status = 'completed'
	`, userID).Scan(&stats.Completed)

	_ = s.db.QueryRowContext(ctx, `
		SELECT coalesce(SUM(c.duration_hours), 0.0)
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.user_id = ? AND e.status != 'dropped'
	`, userID).Scan(&stats.TotalHours)

	var avg sql.NullFloat64
	_ = s.db.QueryRowContext(ctx, `
		SELECT AVG(progress_percent) FROM enrollments WHERE user_id = ? AND status != 'dropped'
	`, userID).Scan(&avg)
	if avg.Valid {
		stats.AverageProgress = avg.Float64
	}

	return stats, nil
}

func (s *Service) ListCourseStudents(ctx context.Context, courseID string) ([]Enrollment, error) {
	query := `
		SELECT 
			e.id, e.user_id, e.course_id, e.status, e.progress_percent,
			e.enrolled_at, e.completed_at, e.last_accessed_at, coalesce(e.last_lesson_id, ''),
			u.first_name || ' ' || u.last_name as student_name,
			u.email as student_email,
			(SELECT COUNT(*) FROM lesson_progress lp WHERE lp.user_id = e.user_id AND lp.course_id = e.course_id) as completed_count,
			(SELECT COUNT(*) FROM lessons l WHERE l.course_id = e.course_id) as total_count
		FROM enrollments e
		JOIN users u ON e.user_id = u.id
		WHERE e.course_id = ? AND e.status != 'dropped'
		ORDER BY e.enrolled_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("error consultando estudiantes matriculados: %w", err)
	}
	defer rows.Close()

	var result []Enrollment
	for rows.Next() {
		var e Enrollment
		var compAt sql.NullTime

		err := rows.Scan(
			&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.ProgressPercent,
			&e.EnrolledAt, &compAt, &e.LastAccessedAt, &e.LastLessonID,
			&e.StudentName, &e.StudentEmail,
			&e.CompletedLessons, &e.TotalLessons,
		)
		if err != nil {
			return nil, err
		}
		if compAt.Valid {
			e.CompletedAt = &compAt.Time
		}

		result = append(result, e)
	}

	if result == nil {
		result = []Enrollment{}
	}

	return result, nil
}

func (s *Service) AdminUnenrollStudent(ctx context.Context, enrollmentID string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM enrollments WHERE id = ?", enrollmentID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("matrícula no encontrada")
	}
	return nil
}

func (s *Service) getEnrollmentByID(ctx context.Context, enrID string) (*Enrollment, error) {
	query := `
		SELECT 
			e.id, e.user_id, e.course_id, e.status, e.progress_percent,
			e.enrolled_at, e.completed_at, e.last_accessed_at, coalesce(e.last_lesson_id, ''),
			c.title, c.code, c.thumbnail_url, c.category, c.level,
			c.instructor_name, c.duration_hours,
			(SELECT COUNT(*) FROM lesson_progress lp WHERE lp.user_id = e.user_id AND lp.course_id = e.course_id) as completed_count,
			(SELECT COUNT(*) FROM lessons l WHERE l.course_id = e.course_id) as total_count
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.id = ?
	`

	var e Enrollment
	var compAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, enrID).Scan(
		&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.ProgressPercent,
		&e.EnrolledAt, &compAt, &e.LastAccessedAt, &e.LastLessonID,
		&e.CourseTitle, &e.CourseCode, &e.CourseThumbnail, &e.CourseCategory, &e.CourseLevel,
		&e.CourseInstructor, &e.CourseDurationHours,
		&e.CompletedLessons, &e.TotalLessons,
	)
	if err != nil {
		return nil, err
	}
	if compAt.Valid {
		e.CompletedAt = &compAt.Time
	}

	return &e, nil
}
