package analytics

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogRow es un evento del registro de auditoría con el nombre del usuario.
type AuditLogRow struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	IPAddress string    `json:"ipAddress"`
	RequestID string    `json:"requestId"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"createdAt"`
}

// AuditFilter acota la consulta del registro de auditoría.
type AuditFilter struct {
	Action string
	UserID string
	Search string
	Limit  int
	Offset int
}

func (s *Service) ListAuditLogs(ctx context.Context, f AuditFilter) ([]AuditLogRow, error) {
	var where []string
	var args []any
	if f.Action != "" {
		where = append(where, "a.action = ?")
		args = append(args, f.Action)
	}
	if f.UserID != "" {
		where = append(where, "a.user_id = ?")
		args = append(args, f.UserID)
	}
	if f.Search != "" {
		where = append(where, "(a.details LIKE ? OR a.resource LIKE ?)")
		term := "%" + f.Search + "%"
		args = append(args, term, term)
	}
	clause := ""
	if len(where) > 0 {
		clause = "WHERE " + strings.Join(where, " AND ")
	}
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	args = append(args, f.Limit, f.Offset)

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT a.id, coalesce(a.user_id, ''), coalesce(u.first_name || ' ' || u.last_name, ''),
		       a.action, a.resource, coalesce(a.ip_address, ''), coalesce(a.request_id, ''),
		       coalesce(a.details, ''), a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.user_id
		%s
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT ? OFFSET ?`, clause), args...)
	if err != nil {
		return nil, fmt.Errorf("error consultando la auditoría: %w", err)
	}
	defer rows.Close()

	out := make([]AuditLogRow, 0)
	for rows.Next() {
		var r AuditLogRow
		if err := rows.Scan(&r.ID, &r.UserID, &r.UserName, &r.Action, &r.Resource, &r.IPAddress, &r.RequestID, &r.Details, &r.CreatedAt); err == nil {
			out = append(out, r)
		}
	}
	return out, nil
}

func (h *Handler) HandleAuditLogs(c *gin.Context) {
	f := AuditFilter{
		Action: c.Query("action"),
		UserID: c.Query("userId"),
		Search: c.Query("search"),
	}
	if l, err := strconv.Atoi(c.Query("limit")); err == nil {
		f.Limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset")); err == nil {
		f.Offset = o
	}
	rows, err := h.svc.ListAuditLogs(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo consultar la auditoría"})
		return
	}
	c.JSON(http.StatusOK, rows)
}
