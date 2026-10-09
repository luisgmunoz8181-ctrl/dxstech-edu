package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func tempDir(t *testing.T) string {
	t.Helper()
	d, err := os.MkdirTemp("", "dxstech_mig_*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(d) })
	return d
}

func appliedVersions(t *testing.T, db *DB) []int {
	t.Helper()
	rows, err := db.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var vs []int
	for rows.Next() {
		var v int
		_ = rows.Scan(&v)
		vs = append(vs, v)
	}
	return vs
}

func TestMigrationsAreVersionedAndIdempotent(t *testing.T) {
	dir := tempDir(t)
	db, err := InitDB(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := appliedVersions(t, db); len(got) != len(migrations) {
		t.Fatalf("versiones aplicadas = %v, se esperaban %d", got, len(migrations))
	}
	db.Close()

	// Reabrir no debe reaplicar nada ni fallar.
	db, err = InitDB(dir, false)
	if err != nil {
		t.Fatalf("reabrir la base falló: %v", err)
	}
	defer db.Close()
	if got := appliedVersions(t, db); len(got) != len(migrations) {
		t.Fatalf("tras reabrir, versiones = %v", got)
	}
}

// Las bases creadas antes del versionado (sin schema_migrations y sin token_version)
// deben adoptarse conservando sus datos.
func TestMigrationsAdoptLegacyDatabase(t *testing.T) {
	dir := tempDir(t)
	raw, err := sql.Open("sqlite", filepath.Join(dir, "dxstech.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(baselineSchema); err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`INSERT INTO users (id, first_name, last_name, email, password_hash, role_id) VALUES ('u1','A','B','a@b.co','x',3)`); err != nil {
		t.Fatal(err)
	}
	raw.Close()

	db, err := InitDB(dir, false)
	if err != nil {
		t.Fatalf("migrar base legada: %v", err)
	}
	defer db.Close()

	var tv int
	if err := db.QueryRow(`SELECT token_version FROM users WHERE id = 'u1'`).Scan(&tv); err != nil {
		t.Fatalf("token_version debería existir y conservar al usuario: %v", err)
	}
	if err := db.QueryRow(`SELECT course_id FROM quizzes LIMIT 1`).Scan(new(string)); err != nil && err != sql.ErrNoRows {
		t.Fatalf("columnas históricas de quizzes no migradas: %v", err)
	}
}

func TestKeyQueriesUseIndexes(t *testing.T) {
	db, err := InitDB(tempDir(t), false)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	queries := map[string]string{
		"lecciones por curso":        `SELECT * FROM lessons WHERE course_id = 'c1'`,
		"módulos por curso":          `SELECT * FROM course_modules WHERE course_id = 'c1' ORDER BY order_index`,
		"matrículas por curso":       `SELECT * FROM enrollments WHERE course_id = 'c1' AND status != 'dropped'`,
		"progreso por curso/usuario": `SELECT COUNT(*) FROM lesson_progress WHERE user_id = 'u1' AND course_id = 'c1'`,
		"progreso por curso":         `SELECT user_id, COUNT(*) FROM lesson_progress WHERE course_id = 'c1' GROUP BY user_id`,
		"foro por curso":             `SELECT * FROM course_discussions WHERE course_id = 'c1' ORDER BY created_at`,
		"quiz por lección":           `SELECT * FROM quizzes WHERE lesson_id = 'l1'`,
		"intentos por quiz/usuario":  `SELECT * FROM quiz_submissions WHERE quiz_id = 'q1' AND user_id = 'u1' ORDER BY submitted_at DESC LIMIT 1`,
		"auditoría por usuario":      `SELECT * FROM audit_logs WHERE user_id = 'u1' ORDER BY created_at DESC`,
		"reseteo por token":          `SELECT * FROM password_resets WHERE token_hash = 'h'`,
		"certificados recientes":     `SELECT * FROM issued_certificates ORDER BY created_at DESC LIMIT 100`,
		"monitoreo reciente":         `SELECT * FROM enrollments ORDER BY last_accessed_at DESC LIMIT 50`,
	}
	for name, q := range queries {
		rows, err := db.Query("EXPLAIN QUERY PLAN " + q)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var plan []string
		for rows.Next() {
			var id, parent, notused int
			var detail string
			_ = rows.Scan(&id, &parent, &notused, &detail)
			plan = append(plan, detail)
		}
		rows.Close()
		for _, d := range plan {
			// "SCAN <tabla>" a secas es un recorrido completo; con USING ... INDEX es aceptable.
			if strings.HasPrefix(d, "SCAN ") && !strings.Contains(d, "USING") {
				t.Errorf("%s hace un recorrido completo de tabla: %s", name, strings.Join(plan, " | "))
			}
		}
	}
}

func TestForeignKeysCascadeOnCourseDelete(t *testing.T) {
	db, err := InitDB(tempDir(t), true)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	count := func(table string) int {
		var n int
		_ = db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE course_id = 'crs-ai-101'`, table)).Scan(&n)
		return n
	}
	if count("course_modules") == 0 || count("lessons") == 0 || count("enrollments") == 0 {
		t.Fatal("el seed demo debería tener módulos, lecciones y matrículas")
	}
	if _, err := db.Exec(`DELETE FROM courses WHERE id = 'crs-ai-101'`); err != nil {
		t.Fatal(err)
	}
	for _, tbl := range []string{"course_modules", "lessons", "enrollments", "lesson_progress"} {
		if n := count(tbl); n != 0 {
			t.Errorf("%s quedó con %d filas huérfanas tras borrar el curso", tbl, n)
		}
	}
}

// Lecturas y escrituras concurrentes no deben producir "database is locked".
func TestConcurrentReadsAndWrites(t *testing.T) {
	db, err := InitDB(tempDir(t), true)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var wg sync.WaitGroup
	errs := make(chan error, 200)
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 25; i++ {
				_, err := db.Exec(`INSERT INTO audit_logs (user_id, action, resource) VALUES (?, 'T', 'r')`, fmt.Sprintf("w%d", w))
				if err != nil {
					errs <- fmt.Errorf("escritura: %w", err)
				}
			}
		}(w)
	}
	for r := 0; r < 8; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 25; i++ {
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM enrollments e JOIN courses c ON c.id = e.course_id`).Scan(&n); err != nil {
					errs <- fmt.Errorf("lectura: %w", err)
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action = 'T'`).Scan(&n)
	if n != 100 {
		t.Errorf("se esperaban 100 escrituras, hay %d", n)
	}
}
