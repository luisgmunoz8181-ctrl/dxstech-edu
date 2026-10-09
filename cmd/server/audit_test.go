package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func (e *routeEnv) do(method, path, token, contentType string, body []byte) (int, []byte) {
	req, _ := http.NewRequest(method, path, bytes.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	return w.Code, w.Body.Bytes()
}

type auditRow struct {
	UserID, UserName, Action, Resource, RequestID, Details string
}

func (e *routeEnv) audit(t *testing.T, query string) []auditRow {
	t.Helper()
	code, body := e.do("GET", "/api/admin/audit?"+query, e.admin, "", nil)
	if code != 200 {
		t.Fatalf("GET audit: %d %s", code, body)
	}
	var rows []auditRow
	if err := json.Unmarshal(body, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestAdminActionsAreAudited(t *testing.T) {
	e := newRouteEnv(t)
	jsonCT := "application/json"

	// Crear curso -> módulo -> subir archivo -> eliminar curso.
	code, body := e.do("POST", "/api/courses", e.admin, jsonCT, []byte(`{"title":"Curso de auditoría","code":"AUD-1","description":"d","category":"Tecnología","level":"Principiante","durationHours":1}`))
	if code != 201 {
		t.Fatalf("crear curso: %d %s", code, body)
	}
	var course struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &course)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "doc.pdf")
	_, _ = fw.Write([]byte("%PDF-1.7 contenido"))
	_ = mw.Close()
	if code, b := e.do("POST", "/api/courses/upload", e.admin, mw.FormDataContentType(), buf.Bytes()); code != 200 {
		t.Fatalf("subida: %d %s", code, b)
	}
	if code, b := e.do("POST", "/api/enrollments/admin/enroll", e.admin, jsonCT, []byte(`{"userId":"usr-student-01","courseId":"`+course.ID+`"}`)); code != 201 {
		t.Fatalf("matrícula admin: %d %s", code, b)
	}
	if code, b := e.do("DELETE", "/api/courses/"+course.ID, e.admin, "", nil); code != 200 {
		t.Fatalf("borrar curso: %d %s", code, b)
	}
	e.do("POST", "/api/whatsapp/sync-courses", e.admin, jsonCT, []byte("{}"))

	for action, mustContain := range map[string]string{
		"COURSE_CREATE":           "AUD-1",
		"FILE_UPLOAD":             "doc.pdf",
		"ENROLLMENT_ADMIN_ENROLL": course.ID,
		"COURSE_DELETE":           "AUD-1",
		"WHATSAPP_SYNC_COURSES":   "cursos",
	} {
		rows := e.audit(t, "action="+action)
		if len(rows) != 1 {
			t.Errorf("%s: se esperaba 1 evento, hay %d", action, len(rows))
			continue
		}
		r := rows[0]
		if r.UserID != "usr-admin-01" || r.UserName == "" || r.RequestID == "" || !strings.Contains(r.Details, mustContain) {
			t.Errorf("%s: evento incompleto: %+v", action, r)
		}
	}

	// Filtros y paginación.
	if rows := e.audit(t, "userId=usr-admin-01&limit=2"); len(rows) != 2 {
		t.Errorf("limit=2 devolvió %d filas", len(rows))
	}
	if rows := e.audit(t, "search=AUD-1"); len(rows) < 2 {
		t.Errorf("la búsqueda por texto debería encontrar creación y borrado, encontró %d", len(rows))
	}

	// Un estudiante no puede leer la auditoría.
	if code, _ := e.do("GET", "/api/admin/audit", e.student, "", nil); code != http.StatusForbidden {
		t.Errorf("estudiante leyendo auditoría: esperado 403, obtuvo %d", code)
	}
}
