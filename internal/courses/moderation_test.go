package courses_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/courses"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

type modEnv struct {
	svc *courses.Service
	db  *database.DB
	r   *gin.Engine
}

func newModEnv(t *testing.T) (*modEnv, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, dir, cleanup := setupTestDB(t)
	svc := courses.NewService(db, dir)
	h := courses.NewHandler(svc)
	r := gin.New()
	r.Use(auth.Authenticate(secSecret))
	h.RegisterRoutes(r.Group("/api/courses"))
	h.RegisterModerationRoutes(r.Group("/api/admin/forum"))
	return &modEnv{svc: svc, db: db, r: r}, cleanup
}

func (e *modEnv) call(method, path, token, body string) (int, []byte) {
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	return w.Code, w.Body.Bytes()
}

func (e *modEnv) post(t *testing.T, token, msg, parent string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"message": msg, "title": "t", "parentId": parent})
	code, resp := e.call("POST", "/api/courses/crs-ai-101/discussions", token, string(body))
	if code != http.StatusCreated {
		t.Fatalf("publicar: %d %s", code, resp)
	}
	var d struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(resp, &d)
	return d.ID
}

func TestAuthorsAndAdminsCanDeleteForumPosts(t *testing.T) {
	e, cleanup := newModEnv(t)
	defer cleanup()
	student := secToken(t, "usr-student-01", "ESTUDIANTE")
	admin := secToken(t, "usr-admin-01", "ADMINISTRADOR")

	root := e.post(t, student, "mi duda", "")
	reply := e.post(t, admin, "respuesta del docente", root)

	// Otro estudiante matriculado no puede borrar la publicación ajena.
	if _, err := e.svc.GetDiscussions(context.Background(), "crs-ai-101", ""); err != nil {
		t.Fatal(err)
	}
	other := secToken(t, "usr-otro", "ESTUDIANTE")
	if code, _ := e.call("DELETE", "/api/courses/crs-ai-101/discussions/"+root, other, ""); code != http.StatusForbidden {
		t.Errorf("otro usuario borrando: esperado 403, obtuvo %d", code)
	}
	if code, _ := e.call("DELETE", "/api/courses/crs-ai-101/discussions/"+root, "", ""); code != http.StatusUnauthorized {
		t.Errorf("anónimo borrando: esperado 401, obtuvo %d", code)
	}
	if code, _ := e.call("DELETE", "/api/courses/crs-ai-101/discussions/no-existe", student, ""); code != http.StatusNotFound {
		t.Errorf("publicación inexistente: esperado 404, obtuvo %d", code)
	}
	// Una publicación no se borra a través de otro curso.
	if code, _ := e.call("DELETE", "/api/courses/crs-dev-201/discussions/"+root, admin, ""); code != http.StatusNotFound {
		t.Errorf("borrar vía otro curso: esperado 404, obtuvo %d", code)
	}

	// El autor borra su hilo: se van también las respuestas.
	if code, body := e.call("DELETE", "/api/courses/crs-ai-101/discussions/"+root, student, ""); code != http.StatusOK {
		t.Fatalf("autor borrando: %d %s", code, body)
	}
	threads, _ := e.svc.GetDiscussions(context.Background(), "crs-ai-101", "")
	if len(threads) != 0 {
		t.Errorf("el hilo y sus respuestas debían desaparecer, quedan %d", len(threads))
	}
	_ = reply

	// Un admin puede borrar publicaciones ajenas.
	p := e.post(t, student, "otra", "")
	if code, _ := e.call("DELETE", "/api/courses/crs-ai-101/discussions/"+p, admin, ""); code != http.StatusOK {
		t.Errorf("admin borrando publicación ajena: esperado 200, obtuvo %d", code)
	}
}

func TestReportAndModerationQueue(t *testing.T) {
	e, cleanup := newModEnv(t)
	defer cleanup()
	student := secToken(t, "usr-student-01", "ESTUDIANTE")
	admin := secToken(t, "usr-admin-01", "ADMINISTRADOR")

	// Un segundo usuario matriculado reporta.
	if _, err := e.svc.GetDiscussions(context.Background(), "crs-ai-101", ""); err != nil {
		t.Fatal(err)
	}
	dbExec := func(q string, args ...any) {
		if _, err := e.db.Exec(q, args...); err != nil {
			t.Fatal(err)
		}
	}
	dbExec(`INSERT INTO enrollments (id, user_id, course_id) VALUES ('enr-rep', 'usr-admin-01', 'crs-ai-101')`)
	reporter := secToken(t, "usr-estudiante-2", "ESTUDIANTE")
	dbExec(`INSERT INTO users (id, first_name, last_name, email, password_hash, role_id) VALUES ('usr-estudiante-2','Laura','Reportera','l@x.co','x',3)`)
	dbExec(`INSERT INTO enrollments (id, user_id, course_id) VALUES ('enr-rep2', 'usr-estudiante-2', 'crs-ai-101')`)

	post := e.post(t, student, "mensaje ofensivo <b>x</b>", "")
	url := "/api/courses/crs-ai-101/discussions/" + post + "/report"

	if code, _ := e.call("POST", url, student, `{"reason":"x"}`); code != http.StatusBadRequest {
		t.Errorf("reportar la propia publicación: esperado 400, obtuvo %d", code)
	}
	if code, _ := e.call("POST", url, secToken(t, "usr-nobody", "ESTUDIANTE"), `{}`); code != http.StatusForbidden {
		t.Errorf("reportar sin matrícula: esperado 403, obtuvo %d", code)
	}
	if code, body := e.call("POST", url, reporter, `{"reason":"Lenguaje ofensivo"}`); code != http.StatusCreated {
		t.Fatalf("reportar: %d %s", code, body)
	}
	if code, _ := e.call("POST", url, reporter, `{}`); code != http.StatusConflict {
		t.Errorf("reporte duplicado: esperado 409, obtuvo %d", code)
	}
	if code, _ := e.call("POST", "/api/courses/crs-ai-101/discussions/no-existe/report", reporter, `{}`); code != http.StatusNotFound {
		t.Errorf("reportar inexistente: esperado 404, obtuvo %d", code)
	}

	// Cola de moderación: solo administradores.
	if code, _ := e.call("GET", "/api/admin/forum/reports", student, ""); code != http.StatusForbidden {
		t.Errorf("estudiante viendo la cola: esperado 403, obtuvo %d", code)
	}
	code, body := e.call("GET", "/api/admin/forum/reports", admin, "")
	var reports []courses.ForumReport
	_ = json.Unmarshal(body, &reports)
	if code != 200 || len(reports) != 1 {
		t.Fatalf("cola: %d %s", code, body)
	}
	r := reports[0]
	if r.Status != "open" || !r.PostExists || r.PostAuthorName == "" && r.PostAuthorID != "usr-student-01" || r.Reason != "Lenguaje ofensivo" || r.CourseTitle == "" {
		t.Errorf("reporte incompleto: %+v", r)
	}

	// Descartar: la publicación se queda.
	if code, _ := e.call("POST", "/api/admin/forum/reports/"+r.ID+"/resolve", admin, `{"action":"dismiss"}`); code != 200 {
		t.Fatalf("descartar: %d", code)
	}
	if th, _ := e.svc.GetDiscussions(context.Background(), "crs-ai-101", ""); len(th) != 1 {
		t.Error("descartar no debe eliminar la publicación")
	}
	if code, _ := e.call("POST", "/api/admin/forum/reports/"+r.ID+"/resolve", admin, `{"action":"remove"}`); code != http.StatusBadRequest {
		t.Errorf("resolver dos veces: esperado 400, obtuvo %d", code)
	}
	if code, _ := e.call("GET", "/api/admin/forum/reports", admin, ""); code != 200 {
		t.Fatal("listar")
	}
	_, open := e.call("GET", "/api/admin/forum/reports", admin, "")
	if strings.Contains(string(open), r.ID) {
		t.Error("un reporte resuelto no debe aparecer en la cola abierta")
	}

	// Segundo reporte sobre otra publicación -> eliminar: se borra con sus respuestas, el historial se conserva.
	post2 := e.post(t, student, "spam", "")
	e.post(t, admin, "respuesta", post2)
	e.call("POST", "/api/courses/crs-ai-101/discussions/"+post2+"/report", reporter, `{"reason":"spam"}`)
	_, body = e.call("GET", "/api/admin/forum/reports", admin, "")
	_ = json.Unmarshal(body, &reports)
	if code, _ := e.call("POST", "/api/admin/forum/reports/"+reports[0].ID+"/resolve", admin, `{"action":"remove"}`); code != 200 {
		t.Fatalf("eliminar: %d", code)
	}
	if th, _ := e.svc.GetDiscussions(context.Background(), "crs-ai-101", ""); len(th) != 1 {
		t.Errorf("solo debía quedar la publicación descartada, hay %d", len(th))
	}
	_, body = e.call("GET", "/api/admin/forum/reports?status=all", admin, "")
	_ = json.Unmarshal(body, &reports)
	removed := 0
	for _, rp := range reports {
		if rp.Status == "removed" {
			removed++
			if rp.PostMessage != "spam" || rp.PostExists || rp.ResolvedBy == "" {
				t.Errorf("el historial debe conservar el mensaje y quién resolvió: %+v", rp)
			}
		}
	}
	if removed != 1 {
		t.Errorf("reportes eliminados en el historial = %d, esperado 1", removed)
	}
	if code, _ := e.call("POST", "/api/admin/forum/reports/x/resolve", admin, `{"action":"borrar-todo"}`); code != http.StatusBadRequest {
		t.Errorf("acción inválida: esperado 400, obtuvo %d", code)
	}
}
