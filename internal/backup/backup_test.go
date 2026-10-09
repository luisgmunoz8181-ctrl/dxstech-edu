package backup_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"dxstech-edu/internal/backup"
	"dxstech-edu/internal/database"
)

type env struct {
	db      *database.DB
	dataDir string
	svc     *backup.Service
}

func newEnv(t *testing.T, retention int) *env {
	t.Helper()
	dir, _ := os.MkdirTemp("", "dxstech_backup_*")
	t.Cleanup(func() { os.RemoveAll(dir) })
	db, err := database.InitDB(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	_ = os.MkdirAll(filepath.Join(dir, "uploads"), 0o755)
	return &env{db: db, dataDir: dir, svc: backup.New(db, dir, filepath.Join(dir, "backups"), retention)}
}

func (e *env) upload(t *testing.T, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(e.dataDir, "uploads", name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func count(t *testing.T, path, query string) int {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(query).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestCreateAndRestoreRoundTrip(t *testing.T) {
	e := newEnv(t, 5)
	e.upload(t, "a.pdf", "contenido A")
	e.upload(t, "b.png", "contenido B")

	info, err := e.svc.Create(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(info.Name, "dxstech-backup-") || info.SizeBytes == 0 {
		t.Fatalf("info inesperada: %+v", info)
	}

	// Se destruye el estado actual: usuarios borrados y un archivo eliminado.
	if _, err := e.db.Exec(`DELETE FROM lesson_progress`); err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`DELETE FROM enrollments`); err != nil {
		t.Fatal(err)
	}
	if _, err := e.db.Exec(`DELETE FROM users WHERE id = 'usr-student-01'`); err != nil {
		t.Fatal(err)
	}
	os.Remove(filepath.Join(e.dataDir, "uploads", "a.pdf"))
	e.db.Close()

	path, err := e.svc.Path(info.Name)
	if err != nil {
		t.Fatal(err)
	}
	if err := backup.Restore(path, e.dataDir); err != nil {
		t.Fatalf("restaurar: %v", err)
	}

	restored := filepath.Join(e.dataDir, "dxstech.db")
	if n := count(t, restored, `SELECT COUNT(*) FROM users`); n != 3 {
		t.Errorf("usuarios restaurados = %d, esperados 3", n)
	}
	if n := count(t, restored, `SELECT COUNT(*) FROM enrollments`); n != 1 {
		t.Errorf("matrículas restauradas = %d, esperadas 1", n)
	}
	for name, want := range map[string]string{"a.pdf": "contenido A", "b.png": "contenido B"} {
		got, err := os.ReadFile(filepath.Join(e.dataDir, "uploads", name))
		if err != nil || string(got) != want {
			t.Errorf("archivo %s no restaurado: %q %v", name, got, err)
		}
	}

	// Lo anterior se conserva aparte, no se borra.
	matches, _ := filepath.Glob(filepath.Join(e.dataDir, "pre-restore-*", "dxstech.db"))
	if len(matches) != 1 {
		t.Errorf("la base anterior debe conservarse en pre-restore-*, hay %d", len(matches))
	} else if n := count(t, matches[0], `SELECT COUNT(*) FROM users`); n != 2 {
		t.Errorf("la base apartada debía tener el estado previo (2 usuarios), tiene %d", n)
	}
}

func TestSnapshotIsConsistentWhileWriting(t *testing.T) {
	e := newEnv(t, 5)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
				_, _ = e.db.Exec(`INSERT INTO audit_logs (action, resource) VALUES ('X', 'r')`)
			}
		}
	}()
	info, err := e.svc.Create(context.Background())
	close(stop)
	wg.Wait()
	if err != nil {
		t.Fatalf("la copia debe poder hacerse con escrituras en curso: %v", err)
	}

	e2 := t.TempDir()
	p, _ := e.svc.Path(info.Name)
	if err := backup.Restore(p, e2); err != nil { // Restore ejecuta integrity_check
		t.Fatalf("la instantánea debe ser una base íntegra: %v", err)
	}
}

func TestRetentionKeepsOnlyTheNewestBackups(t *testing.T) {
	e := newEnv(t, 2)
	dir := e.svc.Dir()
	_ = os.MkdirAll(dir, 0o700)
	for _, n := range []string{"dxstech-backup-20240101-000000.tar.gz", "dxstech-backup-20240102-000000.tar.gz", "dxstech-backup-20240103-000000.tar.gz"} {
		_ = os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o600)
	}
	_ = os.WriteFile(filepath.Join(dir, "notas.txt"), []byte("no es una copia"), 0o600)

	if _, err := e.svc.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	list, _ := e.svc.List()
	if len(list) != 2 {
		t.Fatalf("se esperaban 2 copias tras la retención, hay %d: %+v", len(list), list)
	}
	if !strings.HasPrefix(list[1].Name, "dxstech-backup-20240103") {
		t.Errorf("deben conservarse las más recientes: %+v", list)
	}
	if _, err := os.Stat(filepath.Join(dir, "notas.txt")); err != nil {
		t.Error("la retención no debe tocar archivos ajenos")
	}
	if list[0].CreatedAt.Before(list[1].CreatedAt) {
		t.Error("el listado debe ir de la más reciente a la más antigua")
	}
}

func TestPathRejectsTraversalAndForeignNames(t *testing.T) {
	e := newEnv(t, 2)
	if _, err := e.svc.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"../dxstech.db", "..%2Fx", "/etc/passwd", "dxstech-backup-x.tar.gz", "dxstech-backup-20240101-000000.tar.gz/../../x", "dxstech.db", ""} {
		if _, err := e.svc.Path(bad); err == nil {
			t.Errorf("Path(%q) debería rechazarse", bad)
		}
	}
}

func makeArchive(t *testing.T, entries map[string]string, typeflag byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "evil.tar.gz")
	f, _ := os.Create(path)
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, body := range entries {
		_ = tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: typeflag})
		_, _ = tw.Write([]byte(body))
	}
	_ = tw.Close()
	_ = gz.Close()
	_ = f.Close()
	return path
}

func TestRestoreBlocksMaliciousArchives(t *testing.T) {
	manifest, _ := json.Marshal(backup.Manifest{DBSHA256: "x"})
	cases := map[string]map[string]string{
		"traversal en uploads": {"MANIFEST.json": string(manifest), "dxstech.db": "x", "uploads/../../evil.txt": "pwn"},
		"ruta absoluta":        {"MANIFEST.json": string(manifest), "dxstech.db": "x", "/etc/cron.d/evil": "pwn"},
		"subcarpeta":           {"MANIFEST.json": string(manifest), "dxstech.db": "x", "uploads/sub/dir.txt": "x"},
		"nombre desconocido":   {"MANIFEST.json": string(manifest), "dxstech.db": "x", "otra-cosa.sh": "x"},
		"sin base de datos":    {"MANIFEST.json": string(manifest)},
		"sin manifiesto":       {"dxstech.db": "x"},
	}
	for name, entries := range cases {
		dataDir := t.TempDir()
		if err := backup.Restore(makeArchive(t, entries, tar.TypeReg), dataDir); err == nil {
			t.Errorf("%s: debe rechazarse", name)
		}
		if _, err := os.Stat(filepath.Join(dataDir, "..", "evil.txt")); err == nil {
			t.Errorf("%s: se escribió fuera de DATA_DIR", name)
		}
		if _, err := os.Stat(filepath.Join(dataDir, "dxstech.db")); err == nil {
			t.Errorf("%s: no debe quedar una base a medio restaurar", name)
		}
	}
	// Enlaces simbólicos dentro de la copia.
	if err := backup.Restore(makeArchive(t, map[string]string{"uploads/link": "/etc/passwd"}, tar.TypeSymlink), t.TempDir()); err == nil {
		t.Error("los enlaces simbólicos deben rechazarse")
	}
}

func TestRestoreDetectsTamperedDatabase(t *testing.T) {
	e := newEnv(t, 2)
	info, err := e.svc.Create(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	p, _ := e.svc.Path(info.Name)

	// Se rehace la copia cambiando un byte de la base pero conservando el manifiesto.
	src, _ := os.Open(p)
	gz, _ := gzip.NewReader(src)
	tr := tar.NewReader(gz)
	tampered := filepath.Join(t.TempDir(), "tampered.tar.gz")
	out, _ := os.Create(tampered)
	ogz := gzip.NewWriter(out)
	tw := tar.NewWriter(ogz)
	for {
		h, err := tr.Next()
		if err != nil {
			break
		}
		body, _ := io.ReadAll(tr)
		if h.Name == "dxstech.db" {
			body[len(body)-1] ^= 0xFF
		}
		_ = tw.WriteHeader(&tar.Header{Name: h.Name, Mode: 0o644, Size: h.Size})
		_, _ = tw.Write(body)
	}
	_ = tw.Close()
	_ = ogz.Close()
	_ = out.Close()
	_ = src.Close()

	err = backup.Restore(tampered, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "hash") {
		t.Fatalf("una copia alterada debe detectarse por su hash, obtuvo: %v", err)
	}
}

func TestSchedulerCreatesABackupOnlyWhenTheLastOneIsOld(t *testing.T) {
	e := newEnv(t, 5)
	svc := e.svc.WithFirstCheckDelay(30 * time.Millisecond)

	// Con una copia reciente no se crea otra.
	if _, err := svc.Create(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	svc.StartScheduler(ctx, 24*time.Hour)
	time.Sleep(400 * time.Millisecond)
	cancel()
	if list, _ := svc.List(); len(list) != 1 {
		t.Fatalf("con una copia reciente no debe crearse otra, hay %d", len(list))
	}

	// Con la última copia antigua (renombrada a hace 3 días) sí se crea una nueva.
	list, _ := svc.List()
	old := "dxstech-backup-" + time.Now().UTC().Add(-72*time.Hour).Format("20060102-150405") + ".tar.gz"
	if err := os.Rename(filepath.Join(svc.Dir(), list[0].Name), filepath.Join(svc.Dir(), old)); err != nil {
		t.Fatal(err)
	}
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	svc.StartScheduler(ctx2, 24*time.Hour)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if l, _ := svc.List(); len(l) == 2 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("el programador debía crear una copia nueva al ser la última antigua")
}
