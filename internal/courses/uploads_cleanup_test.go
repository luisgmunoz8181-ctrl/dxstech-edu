package courses_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"dxstech-edu/internal/courses"
	"dxstech-edu/internal/database"
)

func writeUpload(t *testing.T, dir, name string) string {
	t.Helper()
	p := filepath.Join(dir, "uploads", name)
	if err := os.WriteFile(p, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func mustExec(t *testing.T, db *database.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAndReplaceRemoveOnlyUnreferencedFiles(t *testing.T) {
	db, dir, cleanup := setupTestDB(t)
	defer cleanup()
	svc := courses.NewService(db, dir)
	ctx := context.Background()

	shared := writeUpload(t, dir, "shared.pdf")
	solo := writeUpload(t, dir, "solo.pdf")
	mustExec(t, db, `UPDATE lessons SET content_url = '/uploads/shared.pdf' WHERE id IN ('lsn-ai-02','lsn-ai-03')`)
	mustExec(t, db, `UPDATE lessons SET content_url = '/uploads/solo.pdf' WHERE id = 'lsn-ai-01'`)

	// Dos lecciones comparten archivo (p. ej. curso duplicado): se conserva hasta la última.
	if err := svc.DeleteLesson(ctx, "lsn-ai-02"); err != nil {
		t.Fatal(err)
	}
	if !exists(shared) {
		t.Error("el archivo compartido no debe borrarse mientras otra lección lo use")
	}
	if err := svc.DeleteLesson(ctx, "lsn-ai-03"); err != nil {
		t.Fatal(err)
	}
	if exists(shared) {
		t.Error("el archivo sin referencias debería haberse eliminado")
	}

	// Reemplazar el archivo de una lección elimina el anterior.
	_, err := svc.UpdateLesson(ctx, "lsn-ai-01", courses.UpdateLessonRequest{
		Title: "Bienvenida", ContentType: "pdf", ContentURL: "https://example.com/otro.pdf", DurationMinutes: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if exists(solo) {
		t.Error("el archivo reemplazado debería eliminarse")
	}

	// Una URL con '..' o externa nunca debe permitir borrar fuera de /uploads.
	outside := filepath.Join(dir, "outside.txt")
	_ = os.WriteFile(outside, []byte("x"), 0644)
	mustExec(t, db, `UPDATE lessons SET content_url = '/uploads/../outside.txt' WHERE id = 'lsn-ai-05'`)
	if err := svc.DeleteLesson(ctx, "lsn-ai-05"); err != nil {
		t.Fatal(err)
	}
	if !exists(outside) {
		t.Error("una URL con '..' no debe permitir borrar fuera de /uploads")
	}
}

func TestDeleteModuleAndCourseRemoveTheirFiles(t *testing.T) {
	db, dir, cleanup := setupTestDB(t)
	defer cleanup()
	svc := courses.NewService(db, dir)
	ctx := context.Background()

	modFile := writeUpload(t, dir, "module.mp4")
	thumb := writeUpload(t, dir, "thumb.png")
	lessonFile := writeUpload(t, dir, "lesson.pdf")
	var moduleID string
	_ = db.QueryRow(`SELECT module_id FROM lessons WHERE id = 'lsn-ai-04'`).Scan(&moduleID)
	mustExec(t, db, `UPDATE lessons SET content_url = '/uploads/module.mp4' WHERE id = 'lsn-ai-04'`)
	mustExec(t, db, `UPDATE lessons SET content_url = '/uploads/lesson.pdf' WHERE id = 'lsn-ai-01'`)
	mustExec(t, db, `UPDATE courses SET thumbnail_url = '/uploads/thumb.png' WHERE id = 'crs-ai-101'`)

	if err := svc.DeleteModule(ctx, moduleID); err != nil {
		t.Fatal(err)
	}
	if exists(modFile) {
		t.Error("al borrar el módulo debe eliminarse el archivo de sus lecciones")
	}
	if !exists(thumb) || !exists(lessonFile) {
		t.Error("el borrado de un módulo no debe tocar archivos de otros módulos ni la miniatura")
	}

	if err := svc.DeleteCourse(ctx, "crs-ai-101"); err != nil {
		t.Fatal(err)
	}
	if exists(thumb) || exists(lessonFile) {
		t.Error("al borrar el curso deben eliminarse la miniatura y los archivos de sus lecciones")
	}
}

func TestCleanOrphanUploadsDryRunApplyAndAgeGuard(t *testing.T) {
	db, dir, cleanup := setupTestDB(t)
	defer cleanup()
	svc := courses.NewService(db, dir)
	ctx := context.Background()

	old := writeUpload(t, dir, "old-orphan.pdf")
	fresh := writeUpload(t, dir, "fresh-orphan.pdf") // recién subido: aún puede estar por guardarse
	used := writeUpload(t, dir, "used.pdf")
	embedded := writeUpload(t, dir, "embedded.png")
	longAgo := time.Now().Add(-48 * time.Hour)
	for _, p := range []string{old, used, embedded} {
		_ = os.Chtimes(p, longAgo, longAgo)
	}
	mustExec(t, db, `UPDATE lessons SET content_url = '/uploads/used.pdf' WHERE id = 'lsn-ai-01'`)
	mustExec(t, db, `UPDATE lessons SET content_body = 'imagen: /uploads/embedded.png' WHERE id = 'lsn-ai-03'`)

	orphans, err := svc.CleanOrphanUploads(ctx, false, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if len(orphans) != 1 || orphans[0].Name != "old-orphan.pdf" {
		t.Fatalf("huérfanos detectados = %+v, se esperaba solo old-orphan.pdf", orphans)
	}
	if !exists(old) {
		t.Fatal("el modo simulación no debe borrar nada")
	}

	if _, err := svc.CleanOrphanUploads(ctx, true, time.Hour); err != nil {
		t.Fatal(err)
	}
	if exists(old) {
		t.Error("con apply debe eliminarse el huérfano antiguo")
	}
	if !exists(fresh) || !exists(used) || !exists(embedded) {
		t.Error("no deben borrarse archivos recientes, referenciados ni enlazados en texto")
	}
}
