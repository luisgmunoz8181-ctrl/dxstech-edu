package courses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dxstech-edu/internal/database"
)

type Service struct {
	db        *database.DB
	uploadDir string
}

func NewService(db *database.DB, dataDir string) *Service {
	uploadDir := filepath.Join(dataDir, "uploads")
	_ = os.MkdirAll(uploadDir, 0755)

	return &Service{
		db:        db,
		uploadDir: uploadDir,
	}
}

func (s *Service) ListCourses(ctx context.Context, role, search, category, status string) ([]Course, error) {
	isAdmin := role == "SUPERADMIN" || role == "ADMINISTRADOR"

	var query strings.Builder
	query.WriteString(`
		SELECT 
			c.id, c.title, c.code, c.slug, c.short_description, c.description,
			c.thumbnail_url, c.category, c.instructor_name, c.duration_hours,
			c.level, c.status, c.published_at, c.requirements, c.learning_objectives,
			c.created_by, c.created_at, c.updated_at,
			(SELECT COUNT(*) FROM course_modules m WHERE m.course_id = c.id) as modules_count,
			(SELECT COUNT(*) FROM lessons l WHERE l.course_id = c.id) as lessons_count
		FROM courses c
		WHERE 1=1
	`)

	var args []any

	if !isAdmin {
		query.WriteString(` AND c.status = 'published'`)
	} else if strings.TrimSpace(status) != "" {
		query.WriteString(` AND c.status = ?`)
		args = append(args, strings.TrimSpace(status))
	}

	if strings.TrimSpace(category) != "" {
		query.WriteString(` AND c.category = ?`)
		args = append(args, strings.TrimSpace(category))
	}

	if strings.TrimSpace(search) != "" {
		searchTerm := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query.WriteString(` AND (lower(c.title) LIKE ? OR lower(c.description) LIKE ? OR lower(c.code) LIKE ?)`)
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	query.WriteString(` ORDER BY c.created_at DESC`)

	rows, err := s.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("error al consultar catálogo de cursos: %w", err)
	}
	defer rows.Close()

	var result []Course
	for rows.Next() {
		var c Course
		var pubAt sql.NullTime
		var createdBy sql.NullString

		err := rows.Scan(
			&c.ID, &c.Title, &c.Code, &c.Slug, &c.ShortDescription, &c.Description,
			&c.ThumbnailURL, &c.Category, &c.InstructorName, &c.DurationHours,
			&c.Level, &c.Status, &pubAt, &c.Requirements, &c.LearningObjectives,
			&createdBy, &c.CreatedAt, &c.UpdatedAt,
			&c.ModulesCount, &c.LessonsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("error al mapear curso: %w", err)
		}

		if pubAt.Valid {
			c.PublishedAt = &pubAt.Time
		}
		if createdBy.Valid {
			c.CreatedBy = createdBy.String
		}

		result = append(result, c)
	}

	if result == nil {
		result = []Course{}
	}

	return result, nil
}

func (s *Service) GetCourse(ctx context.Context, idOrSlug string, role string) (*Course, error) {
	isAdmin := role == "SUPERADMIN" || role == "ADMINISTRADOR"

	var c Course
	var pubAt sql.NullTime
	var createdBy sql.NullString

	query := `
		SELECT 
			c.id, c.title, c.code, c.slug, c.short_description, c.description,
			c.thumbnail_url, c.category, c.instructor_name, c.duration_hours,
			c.level, c.status, c.published_at, c.requirements, c.learning_objectives,
			c.created_by, c.created_at, c.updated_at,
			(SELECT COUNT(*) FROM course_modules m WHERE m.course_id = c.id) as modules_count,
			(SELECT COUNT(*) FROM lessons l WHERE l.course_id = c.id) as lessons_count
		FROM courses c
		WHERE c.id = ? OR c.slug = ?
	`
	err := s.db.QueryRowContext(ctx, query, idOrSlug, idOrSlug).Scan(
		&c.ID, &c.Title, &c.Code, &c.Slug, &c.ShortDescription, &c.Description,
		&c.ThumbnailURL, &c.Category, &c.InstructorName, &c.DurationHours,
		&c.Level, &c.Status, &pubAt, &c.Requirements, &c.LearningObjectives,
		&createdBy, &c.CreatedAt, &c.UpdatedAt,
		&c.ModulesCount, &c.LessonsCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("curso no encontrado")
		}
		return nil, fmt.Errorf("error consultando curso: %w", err)
	}

	if !isAdmin && c.Status != "published" {
		return nil, errors.New("este curso no se encuentra disponible")
	}

	if pubAt.Valid {
		c.PublishedAt = &pubAt.Time
	}
	if createdBy.Valid {
		c.CreatedBy = createdBy.String
	}

	// Fetch Modules
	modRows, err := s.db.QueryContext(ctx, `
		SELECT id, course_id, title, description, order_index, created_at, updated_at
		FROM course_modules
		WHERE course_id = ?
		ORDER BY order_index ASC, created_at ASC
	`, c.ID)
	if err != nil {
		return nil, fmt.Errorf("error consultando módulos: %w", err)
	}
	defer modRows.Close()

	modulesMap := make(map[string]*CourseModule)
	var moduleOrder []string

	for modRows.Next() {
		var m CourseModule
		if err := modRows.Scan(&m.ID, &m.CourseID, &m.Title, &m.Description, &m.OrderIndex, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		m.Lessons = []Lesson{}
		modulesMap[m.ID] = &m
		moduleOrder = append(moduleOrder, m.ID)
	}

	// Fetch Lessons
	lessonRows, err := s.db.QueryContext(ctx, `
		SELECT id, module_id, course_id, title, description, content_type,
		       content_url, content_body, duration_minutes, order_index,
		       is_free_preview, created_at, updated_at
		FROM lessons
		WHERE course_id = ?
		ORDER BY order_index ASC, created_at ASC
	`, c.ID)
	if err != nil {
		return nil, fmt.Errorf("error consultando lecciones: %w", err)
	}
	defer lessonRows.Close()

	for lessonRows.Next() {
		var l Lesson
		var isFree int
		err := lessonRows.Scan(
			&l.ID, &l.ModuleID, &l.CourseID, &l.Title, &l.Description,
			&l.ContentType, &l.ContentURL, &l.ContentBody, &l.DurationMinutes,
			&l.OrderIndex, &isFree, &l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		l.IsFreePreview = isFree == 1

		if mod, exists := modulesMap[l.ModuleID]; exists {
			mod.Lessons = append(mod.Lessons, l)
		}
	}

	c.Modules = make([]CourseModule, 0, len(moduleOrder))
	for _, modID := range moduleOrder {
		c.Modules = append(c.Modules, *modulesMap[modID])
	}

	return &c, nil
}

func (s *Service) CreateCourse(ctx context.Context, req CreateCourseRequest, userID string) (*Course, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("el título del curso es obligatorio")
	}
	if strings.TrimSpace(req.Code) == "" {
		return nil, errors.New("el código del curso es obligatorio")
	}

	id := fmt.Sprintf("crs-%d", time.Now().UnixNano())
	slug := GenerateSlug(req.Title)

	// Verify code uniqueness
	var existingID string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM courses WHERE code = ?", req.Code).Scan(&existingID)
	if err == nil {
		return nil, fmt.Errorf("ya existe un curso con el código '%s'", req.Code)
	}

	// Ensure unique slug
	var slugCount int
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM courses WHERE slug = ?", slug).Scan(&slugCount)
	if slugCount > 0 {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix()%10000)
	}

	now := time.Now()
	category := req.Category
	if category == "" {
		category = "Tecnología"
	}
	level := req.Level
	if level == "" {
		level = "Principiante"
	}
	duration := req.DurationHours
	if duration <= 0 {
		duration = 1.0
	}
	instructor := req.InstructorName
	if instructor == "" {
		instructor = "Equipo DxSTech"
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO courses (
			id, title, code, slug, short_description, description,
			thumbnail_url, category, instructor_name, duration_hours,
			level, status, published_at, requirements, learning_objectives,
			created_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'draft', NULL, ?, ?, ?, ?, ?)
	`,
		id, req.Title, req.Code, slug, req.ShortDescription, req.Description,
		req.ThumbnailURL, category, instructor, duration,
		level, req.Requirements, req.LearningObjectives,
		userID, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("error al guardar curso: %w", err)
	}

	return s.GetCourse(ctx, id, "ADMINISTRADOR")
}

func (s *Service) UpdateCourse(ctx context.Context, id string, req UpdateCourseRequest) (*Course, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("el título del curso es obligatorio")
	}
	if strings.TrimSpace(req.Code) == "" {
		return nil, errors.New("el código del curso es obligatorio")
	}

	// Verify code uniqueness on another course
	var otherID string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM courses WHERE code = ? AND id != ?", req.Code, id).Scan(&otherID)
	if err == nil {
		return nil, fmt.Errorf("ya existe otro curso con el código '%s'", req.Code)
	}

	status := req.Status
	if status != "draft" && status != "published" && status != "archived" {
		status = "draft"
	}

	now := time.Now()
	var pubAt *time.Time
	if status == "published" {
		var currentPub sql.NullTime
		_ = s.db.QueryRowContext(ctx, "SELECT published_at FROM courses WHERE id = ?", id).Scan(&currentPub)
		if currentPub.Valid {
			pubAt = &currentPub.Time
		} else {
			pubAt = &now
		}
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE courses SET
			title = ?, code = ?, short_description = ?, description = ?,
			thumbnail_url = ?, category = ?, instructor_name = ?,
			duration_hours = ?, level = ?, status = ?, published_at = ?,
			requirements = ?, learning_objectives = ?, updated_at = ?
		WHERE id = ?
	`,
		req.Title, req.Code, req.ShortDescription, req.Description,
		req.ThumbnailURL, req.Category, req.InstructorName,
		req.DurationHours, req.Level, status, pubAt,
		req.Requirements, req.LearningObjectives, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar curso: %w", err)
	}

	return s.GetCourse(ctx, id, "ADMINISTRADOR")
}

func (s *Service) DuplicateCourse(ctx context.Context, id string, userID string) (*Course, error) {
	original, err := s.GetCourse(ctx, id, "ADMINISTRADOR")
	if err != nil {
		return nil, err
	}

	newID := fmt.Sprintf("crs-%d", time.Now().UnixNano())
	newTitle := fmt.Sprintf("Copia de %s", original.Title)
	newCode := fmt.Sprintf("%s-COPY-%d", original.Code, time.Now().Unix()%1000)
	newSlug := fmt.Sprintf("%s-copia-%d", original.Slug, time.Now().Unix()%1000)
	now := time.Now()

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO courses (
			id, title, code, slug, short_description, description,
			thumbnail_url, category, instructor_name, duration_hours,
			level, status, published_at, requirements, learning_objectives,
			created_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'draft', NULL, ?, ?, ?, ?, ?)
	`,
		newID, newTitle, newCode, newSlug, original.ShortDescription, original.Description,
		original.ThumbnailURL, original.Category, original.InstructorName, original.DurationHours,
		original.Level, original.Requirements, original.LearningObjectives,
		userID, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("error duplicando curso: %w", err)
	}

	// Duplicate modules and lessons
	for mIdx, mod := range original.Modules {
		newModID := fmt.Sprintf("mod-%d-%d", time.Now().UnixNano(), mIdx)
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO course_modules (id, course_id, title, description, order_index, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, newModID, newID, mod.Title, mod.Description, mod.OrderIndex, now, now)
		if err != nil {
			return nil, fmt.Errorf("error duplicando módulo: %w", err)
		}

		for lIdx, lsn := range mod.Lessons {
			newLsnID := fmt.Sprintf("lsn-%d-%d", time.Now().UnixNano(), lIdx)
			isFreeInt := 0
			if lsn.IsFreePreview {
				isFreeInt = 1
			}
			_, err = s.db.ExecContext(ctx, `
				INSERT INTO lessons (
					id, module_id, course_id, title, description, content_type,
					content_url, content_body, duration_minutes, order_index,
					is_free_preview, created_at, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				newLsnID, newModID, newID, lsn.Title, lsn.Description, lsn.ContentType,
				lsn.ContentURL, lsn.ContentBody, lsn.DurationMinutes, lsn.OrderIndex,
				isFreeInt, now, now,
			)
			if err != nil {
				return nil, fmt.Errorf("error duplicando lección: %w", err)
			}
		}
	}

	return s.GetCourse(ctx, newID, "ADMINISTRADOR")
}

func (s *Service) ChangeCourseStatus(ctx context.Context, id string, status string) (*Course, error) {
	if status != "draft" && status != "published" && status != "archived" {
		return nil, errors.New("estado inválido: debe ser 'draft', 'published' o 'archived'")
	}

	now := time.Now()
	var pubAt *time.Time
	if status == "published" {
		var currentPub sql.NullTime
		_ = s.db.QueryRowContext(ctx, "SELECT published_at FROM courses WHERE id = ?", id).Scan(&currentPub)
		if currentPub.Valid {
			pubAt = &currentPub.Time
		} else {
			pubAt = &now
		}
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE courses SET status = ?, published_at = ?, updated_at = ? WHERE id = ?
	`, status, pubAt, now, id)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("curso no encontrado")
	}

	return s.GetCourse(ctx, id, "ADMINISTRADOR")
}

func (s *Service) DeleteCourse(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM courses WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error al eliminar curso: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("curso no encontrado")
	}
	return nil
}

// Module Management

func (s *Service) CreateModule(ctx context.Context, courseID string, req CreateModuleRequest) (*CourseModule, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("el título del módulo es obligatorio")
	}

	// Verify course exists
	var dummy string
	if err := s.db.QueryRowContext(ctx, "SELECT id FROM courses WHERE id = ?", courseID).Scan(&dummy); err != nil {
		return nil, errors.New("curso no encontrado")
	}

	id := fmt.Sprintf("mod-%d", time.Now().UnixNano())
	now := time.Now()

	orderIndex := req.OrderIndex
	if orderIndex <= 0 {
		var maxOrder sql.NullInt64
		_ = s.db.QueryRowContext(ctx, "SELECT MAX(order_index) FROM course_modules WHERE course_id = ?", courseID).Scan(&maxOrder)
		orderIndex = int(maxOrder.Int64) + 1
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO course_modules (id, course_id, title, description, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, courseID, req.Title, req.Description, orderIndex, now, now)
	if err != nil {
		return nil, fmt.Errorf("error creando módulo: %w", err)
	}

	return &CourseModule{
		ID:          id,
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
		OrderIndex:  orderIndex,
		Lessons:     []Lesson{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Service) UpdateModule(ctx context.Context, moduleID string, req UpdateModuleRequest) (*CourseModule, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("el título del módulo es obligatorio")
	}

	now := time.Now()
	res, err := s.db.ExecContext(ctx, `
		UPDATE course_modules SET title = ?, description = ?, order_index = ?, updated_at = ?
		WHERE id = ?
	`, req.Title, req.Description, req.OrderIndex, now, moduleID)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("módulo no encontrado")
	}

	var m CourseModule
	err = s.db.QueryRowContext(ctx, `SELECT id, course_id, title, description, order_index, created_at, updated_at FROM course_modules WHERE id = ?`, moduleID).Scan(
		&m.ID, &m.CourseID, &m.Title, &m.Description, &m.OrderIndex, &m.CreatedAt, &m.UpdatedAt,
	)
	return &m, err
}

func (s *Service) DeleteModule(ctx context.Context, moduleID string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM course_modules WHERE id = ?", moduleID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("módulo no encontrado")
	}
	return nil
}

// Lesson Management

func (s *Service) CreateLesson(ctx context.Context, moduleID string, req CreateLessonRequest) (*Lesson, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("el título de la lección es obligatorio")
	}
	if err := ValidateContentType(req.ContentType); err != nil {
		return nil, err
	}

	// Find CourseID from module
	var courseID string
	err := s.db.QueryRowContext(ctx, "SELECT course_id FROM course_modules WHERE id = ?", moduleID).Scan(&courseID)
	if err != nil {
		return nil, errors.New("módulo no encontrado")
	}

	id := fmt.Sprintf("lsn-%d", time.Now().UnixNano())
	now := time.Now()

	orderIndex := req.OrderIndex
	if orderIndex <= 0 {
		var maxOrder sql.NullInt64
		_ = s.db.QueryRowContext(ctx, "SELECT MAX(order_index) FROM lessons WHERE module_id = ?", moduleID).Scan(&maxOrder)
		orderIndex = int(maxOrder.Int64) + 1
	}

	duration := req.DurationMinutes
	if duration <= 0 {
		duration = 10
	}

	isFreeInt := 0
	if req.IsFreePreview {
		isFreeInt = 1
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO lessons (
			id, module_id, course_id, title, description, content_type,
			content_url, content_body, duration_minutes, order_index,
			is_free_preview, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		id, moduleID, courseID, req.Title, req.Description, req.ContentType,
		req.ContentURL, req.ContentBody, duration, orderIndex,
		isFreeInt, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("error creando lección: %w", err)
	}

	return &Lesson{
		ID:              id,
		ModuleID:        moduleID,
		CourseID:        courseID,
		Title:           req.Title,
		Description:     req.Description,
		ContentType:     req.ContentType,
		ContentURL:      req.ContentURL,
		ContentBody:     req.ContentBody,
		DurationMinutes: duration,
		OrderIndex:      orderIndex,
		IsFreePreview:   req.IsFreePreview,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (s *Service) UpdateLesson(ctx context.Context, lessonID string, req UpdateLessonRequest) (*Lesson, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("el título de la lección es obligatorio")
	}
	if err := ValidateContentType(req.ContentType); err != nil {
		return nil, err
	}

	duration := req.DurationMinutes
	if duration <= 0 {
		duration = 10
	}

	isFreeInt := 0
	if req.IsFreePreview {
		isFreeInt = 1
	}

	now := time.Now()
	res, err := s.db.ExecContext(ctx, `
		UPDATE lessons SET
			title = ?, description = ?, content_type = ?, content_url = ?,
			content_body = ?, duration_minutes = ?, order_index = ?,
			is_free_preview = ?, updated_at = ?
		WHERE id = ?
	`,
		req.Title, req.Description, req.ContentType, req.ContentURL,
		req.ContentBody, duration, req.OrderIndex,
		isFreeInt, now, lessonID,
	)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("lección no encontrada")
	}

	var l Lesson
	var freeFlag int
	err = s.db.QueryRowContext(ctx, `
		SELECT id, module_id, course_id, title, description, content_type,
		       content_url, content_body, duration_minutes, order_index,
		       is_free_preview, created_at, updated_at
		FROM lessons WHERE id = ?
	`, lessonID).Scan(
		&l.ID, &l.ModuleID, &l.CourseID, &l.Title, &l.Description,
		&l.ContentType, &l.ContentURL, &l.ContentBody, &l.DurationMinutes,
		&l.OrderIndex, &freeFlag, &l.CreatedAt, &l.UpdatedAt,
	)
	l.IsFreePreview = freeFlag == 1
	return &l, err
}

func (s *Service) DeleteLesson(ctx context.Context, lessonID string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM lessons WHERE id = ?", lessonID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("lección no encontrada")
	}
	return nil
}

// File Upload Handler

var allowedExts = map[string]bool{
	".pdf":  true,
	".pptx": true,
	".ppt":  true,
	".mp4":  true,
	".webm": true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".svg":  true,
}

func (s *Service) SaveUpload(fileHeader *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExts[ext] {
		return "", fmt.Errorf("formato no permitido (%s). Se admiten: PDF, PPTX, MP4, WEBM, PNG, JPG, WEBP", ext)
	}

	// Maximum upload size: 25MB
	if fileHeader.Size > 25*1024*1024 {
		return "", errors.New("el archivo excede el tamaño máximo permitido de 25 MB")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("error al leer archivo cargado: %w", err)
	}
	defer src.Close()

	// Safe filename
	cleanBase := filepath.Base(fileHeader.Filename)
	cleanBase = strings.ReplaceAll(cleanBase, " ", "_")
	uniqueName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), cleanBase)
	destPath := filepath.Join(s.uploadDir, uniqueName)

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("error al guardar archivo en el servidor: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("error al escribir contenido: %w", err)
	}

	return "/uploads/" + uniqueName, nil
}
