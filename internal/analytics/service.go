package analytics

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"dxstech-edu/internal/database"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetOverviewMetrics(ctx context.Context) (*OverviewMetrics, error) {
	m := &OverviewMetrics{}

	// 1. Users KPIs
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&m.TotalUsers)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&m.ActiveUsers)
	m.InactiveUsers = m.TotalUsers - m.ActiveUsers
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users u JOIN roles r ON u.role_id = r.id WHERE r.name = 'ESTUDIANTE'").Scan(&m.StudentsCount)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users u JOIN roles r ON u.role_id = r.id WHERE r.name IN ('ADMINISTRADOR', 'SUPERADMIN')").Scan(&m.InstructorsCount)

	// 2. Courses KPIs
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM courses").Scan(&m.TotalCourses)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM courses WHERE status = 'published'").Scan(&m.PublishedCourses)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM courses WHERE status = 'draft'").Scan(&m.DraftCourses)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM courses WHERE status = 'archived'").Scan(&m.ArchivedCourses)

	// 3. Enrollments KPIs
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM enrollments").Scan(&m.TotalEnrollments)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM enrollments WHERE status = 'active'").Scan(&m.ActiveEnrollments)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM enrollments WHERE status = 'completed'").Scan(&m.CompletedEnrollments)

	var avgProg sql.NullFloat64
	_ = s.db.QueryRowContext(ctx, "SELECT AVG(progress_percent) FROM enrollments").Scan(&avgProg)
	if avgProg.Valid {
		m.AverageProgress = avgProg.Float64
	}

	if m.TotalEnrollments > 0 {
		m.CompletionRate = (float64(m.CompletedEnrollments) / float64(m.TotalEnrollments)) * 100.0
	}

	// 4. Certificates KPIs
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM issued_certificates").Scan(&m.TotalCertificates)

	var totalHours sql.NullFloat64
	_ = s.db.QueryRowContext(ctx, `
		SELECT SUM(c.duration_hours)
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.status = 'completed'
	`).Scan(&totalHours)
	if totalHours.Valid {
		m.TotalHoursDelivered = totalHours.Float64
	}

	return m, nil
}

func (s *Service) GetCourseMetrics(ctx context.Context, category string) ([]CourseMetric, error) {
	category = strings.TrimSpace(category)
	query := `
		SELECT
			c.id, c.code, c.title, c.category, c.status, c.duration_hours,
			COUNT(e.id) as total_students,
			SUM(CASE WHEN e.status = 'active' THEN 1 ELSE 0 END) as active_students,
			SUM(CASE WHEN e.status = 'completed' THEN 1 ELSE 0 END) as completed_students,
			coalesce(AVG(e.progress_percent), 0.0) as avg_progress
		FROM courses c
		LEFT JOIN enrollments e ON c.id = e.course_id
	`
	var args []any
	if category != "" && strings.ToLower(category) != "all" {
		query += " WHERE c.category = ?"
		args = append(args, category)
	}
	query += " GROUP BY c.id ORDER BY total_students DESC, c.created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error consultando métricas por curso: %w", err)
	}
	defer rows.Close()

	metrics := make([]CourseMetric, 0)
	for rows.Next() {
		var cm CourseMetric
		var act, comp sql.NullInt64
		var avg sql.NullFloat64
		err := rows.Scan(&cm.CourseID, &cm.CourseCode, &cm.CourseTitle, &cm.Category, &cm.Status, &cm.DurationHours, &cm.TotalStudents, &act, &comp, &avg)
		if err == nil {
			if act.Valid {
				cm.ActiveStudents = int(act.Int64)
			}
			if comp.Valid {
				cm.CompletedStudents = int(comp.Int64)
			}
			if avg.Valid {
				cm.AverageProgress = avg.Float64
			}
			if cm.TotalStudents > 0 {
				cm.CompletionRate = (float64(cm.CompletedStudents) / float64(cm.TotalStudents)) * 100.0
			}
			metrics = append(metrics, cm)
		}
	}

	return metrics, nil
}

func (s *Service) GetUserRegistrationTimeline(ctx context.Context) ([]TimelinePoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT substr(created_at, 1, 7) as period, COUNT(*) as count
		FROM users
		WHERE created_at IS NOT NULL AND length(created_at) >= 7
		GROUP BY period
		ORDER BY period ASC
		LIMIT 12
	`)
	if err != nil {
		return nil, fmt.Errorf("error consultando historial de registro: %w", err)
	}
	defer rows.Close()

	points := make([]TimelinePoint, 0)
	for rows.Next() {
		var p TimelinePoint
		if err := rows.Scan(&p.Period, &p.Count); err == nil {
			points = append(points, p)
		}
	}
	return points, nil
}

func (s *Service) GetEnrollmentsTimeline(ctx context.Context) ([]TimelinePoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT substr(enrolled_at, 1, 7) as period, COUNT(*) as count
		FROM enrollments
		WHERE enrolled_at IS NOT NULL AND length(enrolled_at) >= 7
		GROUP BY period
		ORDER BY period ASC
		LIMIT 12
	`)
	if err != nil {
		return nil, fmt.Errorf("error consultando historial de matriculas: %w", err)
	}
	defer rows.Close()

	points := make([]TimelinePoint, 0)
	for rows.Next() {
		var p TimelinePoint
		if err := rows.Scan(&p.Period, &p.Count); err == nil {
			points = append(points, p)
		}
	}
	return points, nil
}

func (s *Service) GetStudentMonitoring(ctx context.Context, opts FilterOptions) ([]StudentMonitoringRow, error) {
	var conditions []string
	var args []any

	if strings.TrimSpace(opts.CourseID) != "" && opts.CourseID != "all" {
		conditions = append(conditions, "e.course_id = ?")
		args = append(args, strings.TrimSpace(opts.CourseID))
	}
	if strings.TrimSpace(opts.Category) != "" && opts.Category != "all" {
		conditions = append(conditions, "c.category = ?")
		args = append(args, strings.TrimSpace(opts.Category))
	}
	if strings.TrimSpace(opts.Status) != "" && opts.Status != "all" {
		conditions = append(conditions, "e.status = ?")
		args = append(args, strings.TrimSpace(opts.Status))
	}
	if strings.TrimSpace(opts.StartDate) != "" {
		conditions = append(conditions, "date(e.enrolled_at) >= date(?)")
		args = append(args, strings.TrimSpace(opts.StartDate))
	}
	if strings.TrimSpace(opts.EndDate) != "" {
		conditions = append(conditions, "date(e.enrolled_at) <= date(?)")
		args = append(args, strings.TrimSpace(opts.EndDate))
	}
	if strings.TrimSpace(opts.Search) != "" {
		term := "%" + strings.TrimSpace(opts.Search) + "%"
		conditions = append(conditions, "(u.first_name LIKE ? OR u.last_name LIKE ? OR u.email LIKE ? OR c.title LIKE ? OR c.code LIKE ?)")
		args = append(args, term, term, term, term, term)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := opts.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT
			e.id, e.user_id, u.first_name, u.last_name, u.email,
			c.id, c.code, c.title, c.category,
			e.progress_percent, e.status,
			coalesce(cert.id, ''),
			e.enrolled_at, e.completed_at, e.last_accessed_at
		FROM enrollments e
		JOIN users u ON e.user_id = u.id
		JOIN courses c ON e.course_id = c.id
		LEFT JOIN issued_certificates cert ON cert.user_id = e.user_id AND cert.course_id = e.course_id
		%s
		ORDER BY e.last_accessed_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error consultando monitoreo de estudiantes: %w", err)
	}
	defer rows.Close()

	results := make([]StudentMonitoringRow, 0)
	for rows.Next() {
		var r StudentMonitoringRow
		var fn, ln, certID string
		var compAt *time.Time
		err := rows.Scan(
			&r.EnrollmentID, &r.UserID, &fn, &ln, &r.StudentEmail,
			&r.CourseID, &r.CourseCode, &r.CourseTitle, &r.Category,
			&r.ProgressPercent, &r.Status,
			&certID,
			&r.EnrolledAt, &compAt, &r.LastAccessedAt,
		)
		if err == nil {
			r.StudentName = strings.TrimSpace(fn + " " + ln)
			r.CertificateID = certID
			r.CompletedAt = compAt
			results = append(results, r)
		}
	}

	return results, nil
}

func (s *Service) ExportEnrollmentsCSV(ctx context.Context, opts FilterOptions) ([]byte, error) {
	opts.Limit = 1000 // Export up to 1000 rows
	opts.Offset = 0
	rows, err := s.GetStudentMonitoring(ctx, opts)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	// UTF-8 BOM for perfect Microsoft Excel rendering
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)
	header := []string{
		"ID Matricula",
		"Estudiante",
		"Correo Electronico",
		"Codigo Curso",
		"Nombre Curso",
		"Categoria",
		"Progreso (%)",
		"Estado",
		"Certificado ID",
		"Fecha Inscripcion",
		"Fecha Finalizacion",
		"Ultimo Acceso",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, r := range rows {
		compStr := "Pendiente"
		if r.CompletedAt != nil {
			compStr = r.CompletedAt.Format("2006-01-02 15:04")
		}
		record := []string{
			r.EnrollmentID,
			r.StudentName,
			r.StudentEmail,
			r.CourseCode,
			r.CourseTitle,
			r.Category,
			fmt.Sprintf("%.1f", r.ProgressPercent),
			r.Status,
			r.CertificateID,
			r.EnrolledAt.Format("2006-01-02 15:04"),
			compStr,
			r.LastAccessedAt.Format("2006-01-02 15:04"),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *Service) ExportCertificatesCSV(ctx context.Context) ([]byte, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, student_name, course_title, duration_hours, instructor_name, issue_date, qr_code_url, created_at
		FROM issued_certificates
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buf := new(bytes.Buffer)
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)
	header := []string{
		"ID Registro",
		"Nombre Estudiante",
		"Curso Acreditado",
		"Horas Lectivas",
		"Instructor / Emisor",
		"Fecha de Emision",
		"URL Validacion QR",
		"Fecha de Registro",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for rows.Next() {
		var id, name, title, inst, issueDate, qrURL string
		var hours sql.NullFloat64
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &title, &hours, &inst, &issueDate, &qrURL, &createdAt); err == nil {
			hVal := 0.0
			if hours.Valid {
				hVal = hours.Float64
			}
			record := []string{
				id,
				name,
				title,
				fmt.Sprintf("%.1f", hVal),
				inst,
				issueDate,
				qrURL,
				createdAt.Format("2006-01-02 15:04:05"),
			}
			_ = writer.Write(record)
		}
	}

	writer.Flush()
	return buf.Bytes(), nil
}
