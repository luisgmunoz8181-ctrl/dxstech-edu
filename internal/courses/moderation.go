package courses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrDiscussionNotFound = errors.New("la publicación no existe")
	ErrNotAllowed         = errors.New("no tienes permiso para realizar esta acción")
	ErrAlreadyReported    = errors.New("ya reportaste esta publicación")
)

const maxReportReason = 300

// ForumReport es un reporte de una publicación del foro.
type ForumReport struct {
	ID             string     `json:"id"`
	DiscussionID   string     `json:"discussionId"`
	CourseID       string     `json:"courseId"`
	CourseTitle    string     `json:"courseTitle"`
	ReporterName   string     `json:"reporterName"`
	Reason         string     `json:"reason"`
	PostAuthorID   string     `json:"postAuthorId"`
	PostAuthorName string     `json:"postAuthorName"`
	PostMessage    string     `json:"postMessage"`
	PostExists     bool       `json:"postExists"`
	Status         string     `json:"status"` // open | dismissed | removed
	ResolvedBy     string     `json:"resolvedBy,omitempty"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type discussionRow struct {
	userID, userName, message, parentID string
}

func (s *Service) getDiscussion(ctx context.Context, courseID, id string) (*discussionRow, error) {
	var d discussionRow
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, user_name, message, parent_id
		FROM course_discussions WHERE id = ? AND course_id = ?`, id, courseID).
		Scan(&d.userID, &d.userName, &d.message, &d.parentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDiscussionNotFound
	}
	return &d, err
}

// DeleteDiscussion elimina una publicación (y sus respuestas si es un hilo). Pueden
// hacerlo su autor y los administradores.
func (s *Service) DeleteDiscussion(ctx context.Context, courseID, discussionID, userID, role string) error {
	d, err := s.getDiscussion(ctx, courseID, discussionID)
	if err != nil {
		return err
	}
	isAdmin := role == "SUPERADMIN" || role == "ADMINISTRADOR"
	if !isAdmin && d.userID != userID {
		return ErrNotAllowed
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM course_discussions WHERE (id = ? OR parent_id = ?) AND course_id = ?`, discussionID, discussionID, courseID)
	return err
}

// ReportDiscussion registra el reporte de una publicación ajena.
func (s *Service) ReportDiscussion(ctx context.Context, courseID, discussionID, reporterID, reporterName, reason string) error {
	d, err := s.getDiscussion(ctx, courseID, discussionID)
	if err != nil {
		return err
	}
	if d.userID == reporterID {
		return errors.New("no puedes reportar tu propia publicación")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "Contenido inapropiado"
	}
	if len([]rune(reason)) > maxReportReason {
		return fmt.Errorf("el motivo no puede superar los %d caracteres", maxReportReason)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO discussion_reports (id, discussion_id, course_id, reporter_id, reporter_name, reason,
		                                post_author_id, post_author_name, post_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fmt.Sprintf("rep-%d", time.Now().UnixNano()), discussionID, courseID, reporterID, reporterName, reason,
		d.userID, d.userName, d.message)
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		return ErrAlreadyReported
	}
	return err
}

// ListForumReports devuelve los reportes (status "open" por defecto, o "all").
func (s *Service) ListForumReports(ctx context.Context, status string) ([]ForumReport, error) {
	query := `
		SELECT r.id, r.discussion_id, r.course_id, coalesce(c.title, ''), r.reporter_name, r.reason,
		       r.post_author_id, r.post_author_name, r.post_message,
		       EXISTS(SELECT 1 FROM course_discussions d WHERE d.id = r.discussion_id),
		       r.status, coalesce(u.first_name || ' ' || u.last_name, ''), r.resolved_at, r.created_at
		FROM discussion_reports r
		LEFT JOIN courses c ON c.id = r.course_id
		LEFT JOIN users u ON u.id = r.resolved_by`
	var args []any
	if status != "all" {
		query += ` WHERE r.status = ?`
		if status == "" {
			status = "open"
		}
		args = append(args, status)
	}
	query += ` ORDER BY r.created_at DESC, r.id DESC LIMIT 200`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ForumReport, 0)
	for rows.Next() {
		var r ForumReport
		var resolvedAt sql.NullTime
		if err := rows.Scan(&r.ID, &r.DiscussionID, &r.CourseID, &r.CourseTitle, &r.ReporterName, &r.Reason,
			&r.PostAuthorID, &r.PostAuthorName, &r.PostMessage, &r.PostExists, &r.Status, &r.ResolvedBy, &resolvedAt, &r.CreatedAt); err != nil {
			continue
		}
		if resolvedAt.Valid {
			r.ResolvedAt = &resolvedAt.Time
		}
		out = append(out, r)
	}
	return out, nil
}

// ResolveForumReport cierra todos los reportes abiertos de la misma publicación:
// "dismiss" los descarta y "remove" elimina la publicación (y sus respuestas).
func (s *Service) ResolveForumReport(ctx context.Context, reportID, action, adminID string) (*ForumReport, error) {
	if action != "dismiss" && action != "remove" {
		return nil, errors.New("acción inválida (use dismiss o remove)")
	}
	var discussionID, courseID, status string
	err := s.db.QueryRowContext(ctx, `SELECT discussion_id, course_id, status FROM discussion_reports WHERE id = ?`, reportID).
		Scan(&discussionID, &courseID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("el reporte no existe")
	}
	if err != nil {
		return nil, err
	}
	if status != "open" {
		return nil, errors.New("el reporte ya fue resuelto")
	}

	newStatus := "dismissed"
	if action == "remove" {
		newStatus = "removed"
		if _, err := s.db.ExecContext(ctx, `DELETE FROM course_discussions WHERE (id = ? OR parent_id = ?) AND course_id = ?`, discussionID, discussionID, courseID); err != nil {
			return nil, err
		}
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE discussion_reports SET status = ?, resolved_by = ?, resolved_at = ?
		WHERE discussion_id = ? AND status = 'open'`, newStatus, adminID, time.Now(), discussionID); err != nil {
		return nil, err
	}

	all, err := s.ListForumReports(ctx, "all")
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == reportID {
			return &all[i], nil
		}
	}
	return nil, nil
}
