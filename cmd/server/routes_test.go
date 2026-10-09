package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"

	"dxstech-edu/internal/backup"
	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

type access int

const (
	public     access = iota // sin sesión
	authed                   // cualquier usuario con sesión
	enrolled                 // con sesión Y matrícula en el curso (o rol administrador)
	admin                    // SUPERADMIN o ADMINISTRADOR
	superadmin               // solo SUPERADMIN (datos de toda la plataforma)
)

// routeAccess clasifica TODAS las rutas de la API. Si se añade una ruta nueva
// sin declararla aquí, TestEveryAPIRouteIsClassified falla: obliga a decidir
// explícitamente quién puede usarla.
var routeAccess = map[string]access{
	"GET /api/health": public,

	"POST /api/auth/login":             public,
	"POST /api/auth/logout":            public,
	"POST /api/auth/forgot-password":   public,
	"POST /api/auth/reset-password":    public,
	"GET /api/auth/me":                 authed,
	"PUT /api/auth/change-password":    authed,
	"GET /api/auth/users":              admin,
	"POST /api/auth/users":             admin,
	"PUT /api/auth/users/:id":          admin,
	"PATCH /api/auth/users/:id/status": admin,
	"POST /api/auth/users/:id/unlock":  admin,

	"GET /api/courses":                            public,
	"GET /api/courses/:id":                        public,
	"GET /api/courses/:id/reviews":                public,
	"GET /api/courses/:id/discussions":            enrolled,
	"POST /api/courses/:id/discussions":           enrolled,
	"POST /api/courses/:id/reviews":               enrolled,
	"GET /api/courses/:id/tutor-context":          enrolled,
	"POST /api/courses":                           admin,
	"PUT /api/courses/:id":                        admin,
	"POST /api/courses/:id/duplicate":             admin,
	"PATCH /api/courses/:id/status":               admin,
	"DELETE /api/courses/:id":                     admin,
	"POST /api/courses/:id/modules":               admin,
	"PUT /api/courses/modules/:moduleId":          admin,
	"DELETE /api/courses/modules/:moduleId":       admin,
	"POST /api/courses/modules/:moduleId/lessons": admin,
	"PUT /api/courses/lessons/:lessonId":          admin,
	"DELETE /api/courses/lessons/:lessonId":       admin,
	"POST /api/courses/upload":                    admin,
	"POST /api/courses/uploads/cleanup":           admin,

	"POST /api/enrollments/courses/:courseId":              authed,
	"GET /api/enrollments/my-courses":                      authed,
	"GET /api/enrollments/courses/:courseId/progress":      authed,
	"POST /api/enrollments/progress":                       authed,
	"GET /api/enrollments/stats":                           authed,
	"GET /api/enrollments/admin/course/:courseId/students": admin,
	"POST /api/enrollments/admin/enroll":                   admin,
	"DELETE /api/enrollments/admin/enroll/:id":             admin,

	// La verificación por código/QR es pública (respuesta mínima); el PDF es solo del titular o admin.
	"GET /api/certificates/verify/:id":       public,
	"GET /api/certificates/:id/pdf":          authed, // además: solo el titular o un administrador
	"GET /api/certificates/download/:id":     authed,
	"GET /api/certificates/my-certificates":  authed,
	"GET /api/certificates/course/:courseId": authed,
	"POST /api/certificates/generate":        admin,
	"GET /api/certificates/issued":           admin,

	"GET /api/quizzes/lesson/:lessonId":     authed,
	"GET /api/quizzes/:id":                  authed,
	"POST /api/quizzes/:id/submit":          authed,
	"POST /api/quizzes/generate":            admin,
	"POST /api/quizzes/generate-for-lesson": admin,
	"GET /api/quizzes":                      admin,
	"DELETE /api/quizzes/:id":               admin,

	"POST /api/whatsapp/ask-tutor":     authed,
	"GET /api/whatsapp/status":         admin,
	"POST /api/whatsapp/config":        admin,
	"POST /api/whatsapp/sync-courses":  admin,
	"POST /api/whatsapp/send":          admin,
	"POST /api/whatsapp/simulate-chat": admin,
	"GET /api/whatsapp/logs":           admin,
	"DELETE /api/whatsapp/logs":        admin,

	"GET /api/admin/metrics/overview":             admin,
	"GET /api/admin/metrics/courses":              admin,
	"GET /api/admin/metrics/timeline/users":       admin,
	"GET /api/admin/metrics/timeline/enrollments": admin,
	"GET /api/admin/monitoring/students":          admin,
	"GET /api/admin/reports/enrollments.csv":      admin,
	"GET /api/admin/reports/certificates.csv":     admin,
	"GET /api/admin/audit":                        admin,
	"GET /api/admin/backups":                      superadmin,
	"POST /api/admin/backups":                     superadmin,
	"GET /api/admin/backups/:name":                superadmin,
}

type routeEnv struct {
	r       *gin.Engine
	student string
	admin   string
	super   string
}

func newRouteEnv(t *testing.T) *routeEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil))) // sin ruido del access log
	dir, _ := os.MkdirTemp("", "dxstech_routes_*")
	t.Cleanup(func() { os.RemoveAll(dir) })
	db, err := database.InitDB(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	// Producción: el gateway de WhatsApp exige rol de administrador.
	cfg := &config.Config{AppEnv: "production", JWTSecret: strings.Repeat("s", 40), DataDir: dir, TrustedProxies: nil}
	r, err := newRouter(cfg, db, backup.New(db, dir, dir+"/backups", 3))
	if err != nil {
		t.Fatal(err)
	}
	e := &routeEnv{r: r}
	e.student = e.login(t, "estudiante@dxstech.edu", "Student1234*")
	e.admin = e.login(t, "admin@dxstech.edu", "Admin1234*")
	e.super = e.login(t, "superadmin@dxstech.edu", "Admin1234*")
	return e
}

func (e *routeEnv) login(t *testing.T, email, pass string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": pass})
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	var out struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if out.Token == "" {
		t.Fatalf("login de %s falló: %d %s", email, w.Code, w.Body.String())
	}
	return out.Token
}

func (e *routeEnv) call(method, path, token string) int {
	req, _ := http.NewRequest(method, path, strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	return w.Code
}

func concretePath(pattern string) string {
	parts := strings.Split(pattern, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, ":") {
			parts[i] = "x"
		}
	}
	return strings.Join(parts, "/")
}

func TestEveryAPIRouteIsClassified(t *testing.T) {
	e := newRouteEnv(t)
	seen := map[string]bool{}
	var missing []string
	for _, rt := range e.r.Routes() {
		if !strings.HasPrefix(rt.Path, "/api/") {
			continue
		}
		key := rt.Method + " " + rt.Path
		seen[key] = true
		if _, ok := routeAccess[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("rutas de la API sin clasificar en routeAccess (decide quién puede usarlas):\n  %s", strings.Join(missing, "\n  "))
	}
	for key := range routeAccess {
		if !seen[key] {
			t.Errorf("routeAccess declara una ruta inexistente: %s", key)
		}
	}
}

func TestRouteAccessMatrix(t *testing.T) {
	e := newRouteEnv(t)

	keys := make([]string, 0, len(routeAccess))
	for k := range routeAccess {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	denied := func(code int) bool { return code == http.StatusUnauthorized || code == http.StatusForbidden }

	for _, key := range keys {
		parts := strings.SplitN(key, " ", 2)
		method, path := parts[0], concretePath(parts[1])
		want := routeAccess[key]

		anon := e.call(method, path, "")
		student := e.call(method, path, e.student)
		adm := e.call(method, path, e.admin)
		sup := e.call(method, path, e.super)

		switch want {
		case public:
			if denied(anon) {
				t.Errorf("%s (public): un anónimo recibió %d", key, anon)
			}
		case authed:
			if anon != http.StatusUnauthorized {
				t.Errorf("%s (authed): anónimo esperado 401, obtuvo %d", key, anon)
			}
			if denied(student) {
				t.Errorf("%s (authed): un estudiante con sesión recibió %d", key, student)
			}
		case enrolled:
			if anon != http.StatusUnauthorized {
				t.Errorf("%s (enrolled): anónimo esperado 401, obtuvo %d", key, anon)
			}
			// El curso "x" no existe: un estudiante sin matrícula debe recibir 403.
			if student != http.StatusForbidden {
				t.Errorf("%s (enrolled): estudiante sin matrícula esperado 403, obtuvo %d", key, student)
			}
		case superadmin:
			if anon != http.StatusUnauthorized {
				t.Errorf("%s (superadmin): anónimo esperado 401, obtuvo %d", key, anon)
			}
			if student != http.StatusForbidden || adm != http.StatusForbidden {
				t.Errorf("%s (superadmin): estudiante y administrador esperaban 403, obtuvieron %d y %d", key, student, adm)
			}
			if denied(sup) {
				t.Errorf("%s (superadmin): el superadministrador recibió %d", key, sup)
			}
		case admin:
			if anon != http.StatusUnauthorized {
				t.Errorf("%s (admin): anónimo esperado 401, obtuvo %d", key, anon)
			}
			if student != http.StatusForbidden {
				t.Errorf("%s (admin): estudiante esperado 403, obtuvo %d", key, student)
			}
		}
		if want != superadmin && denied(adm) && key != "GET /api/auth/me" {
			t.Errorf("%s: un administrador recibió %d", key, adm)
		}
	}
}
