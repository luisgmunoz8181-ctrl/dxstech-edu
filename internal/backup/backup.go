// Package backup crea y restaura copias de seguridad consistentes de la base de
// datos SQLite y de los archivos subidos.
//
// Cada copia es un .tar.gz con:
//   - dxstech.db       instantánea consistente (VACUUM INTO, sin detener la app)
//   - uploads/...      archivos subidos
//   - MANIFEST.json    fecha, versión del esquema, conteos y SHA-256 de la base
//
// La restauración (Restore) es OFFLINE: se ejecuta con el servidor detenido
// (dxstech-server --restore <archivo>) y nunca borra lo existente: lo mueve a
// una carpeta pre-restore-<fecha>.
package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"dxstech-edu/internal/database"
)

const (
	dbEntry       = "dxstech.db"
	uploadsPrefix = "uploads/"
	manifestEntry = "MANIFEST.json"
	filePrefix    = "dxstech-backup-"
	fileSuffix    = ".tar.gz"
)

// namePattern restringe los nombres de archivo de copia a los que genera este paquete.
var namePattern = regexp.MustCompile(`^dxstech-backup-\d{8}-\d{6}\.tar\.gz$`)

// Manifest describe el contenido de una copia.
type Manifest struct {
	CreatedAt     time.Time `json:"createdAt"`
	SchemaVersion int       `json:"schemaVersion"`
	DBSHA256      string    `json:"dbSha256"`
	DBBytes       int64     `json:"dbBytes"`
	UploadFiles   int       `json:"uploadFiles"`
	UploadBytes   int64     `json:"uploadBytes"`
}

// Info es una copia disponible en el directorio de respaldos.
type Info struct {
	Name      string    `json:"name"`
	SizeBytes int64     `json:"sizeBytes"`
	CreatedAt time.Time `json:"createdAt"`
}

// Service gestiona las copias de un directorio de datos.
type Service struct {
	db        *database.DB
	dataDir   string
	backupDir string
	retention int

	firstCheckDelay time.Duration // espera antes de la primera comprobación del programador

	mu sync.Mutex // una copia a la vez
}

// New crea el servicio. retention es el número de copias que se conservan (mínimo 1).
func New(db *database.DB, dataDir, backupDir string, retention int) *Service {
	if retention < 1 {
		retention = 1
	}
	return &Service{db: db, dataDir: dataDir, backupDir: backupDir, retention: retention, firstCheckDelay: time.Minute}
}

// WithFirstCheckDelay ajusta la espera inicial del programador (útil en tests).
func (s *Service) WithFirstCheckDelay(d time.Duration) *Service {
	s.firstCheckDelay = d
	return s
}

// Dir devuelve el directorio donde se guardan las copias.
func (s *Service) Dir() string { return s.backupDir }

// Create genera una copia nueva y aplica la retención.
func (s *Service) Create(ctx context.Context) (*Info, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.backupDir, 0o700); err != nil {
		return nil, fmt.Errorf("no se pudo crear el directorio de respaldos: %w", err)
	}

	tmpDir, err := os.MkdirTemp(s.backupDir, ".tmp-snapshot-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	// 1. Instantánea consistente de la base (no bloquea a los escritores en modo WAL).
	snapshot := filepath.Join(tmpDir, dbEntry)
	if _, err := s.db.ExecContext(ctx, "VACUUM INTO '"+strings.ReplaceAll(snapshot, "'", "''")+"'"); err != nil {
		return nil, fmt.Errorf("no se pudo copiar la base de datos: %w", err)
	}

	sum, size, err := hashFile(snapshot)
	if err != nil {
		return nil, err
	}
	var schema int
	_ = s.db.QueryRowContext(ctx, `SELECT coalesce(MAX(version), 0) FROM schema_migrations`).Scan(&schema)

	// 2. Archivo .tar.gz (se escribe con extensión temporal y se renombra al final).
	now := time.Now().UTC()
	name := filePrefix + now.Format("20060102-150405") + fileSuffix
	finalPath := filepath.Join(s.backupDir, name)
	tmpPath := finalPath + ".part"

	out, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear la copia: %w", err)
	}
	cleanup := func() { out.Close(); os.Remove(tmpPath) }

	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)

	if err := addFile(tw, snapshot, dbEntry); err != nil {
		cleanup()
		return nil, err
	}

	m := Manifest{CreatedAt: now, SchemaVersion: schema, DBSHA256: sum, DBBytes: size}
	uploadsDir := filepath.Join(s.dataDir, "uploads")
	entries, _ := os.ReadDir(uploadsDir)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || !info.Mode().IsRegular() { // se omiten carpetas y enlaces simbólicos
			continue
		}
		if err := addFile(tw, filepath.Join(uploadsDir, e.Name()), uploadsPrefix+e.Name()); err != nil {
			cleanup()
			return nil, err
		}
		m.UploadFiles++
		m.UploadBytes += info.Size()
	}

	manifest, _ := json.MarshalIndent(m, "", "  ")
	if err := tw.WriteHeader(&tar.Header{Name: manifestEntry, Mode: 0o600, Size: int64(len(manifest)), ModTime: now}); err != nil {
		cleanup()
		return nil, err
	}
	if _, err := tw.Write(manifest); err != nil {
		cleanup()
		return nil, err
	}

	if err := tw.Close(); err != nil {
		cleanup()
		return nil, err
	}
	if err := gz.Close(); err != nil {
		cleanup()
		return nil, err
	}
	if err := out.Sync(); err != nil {
		cleanup()
		return nil, err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmpPath)
		return nil, err
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return nil, err
	}

	s.prune()

	st, err := os.Stat(finalPath)
	if err != nil {
		return nil, err
	}
	return &Info{Name: name, SizeBytes: st.Size(), CreatedAt: now}, nil
}

// List devuelve las copias existentes, de la más reciente a la más antigua.
func (s *Service) List() ([]Info, error) {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Info{}, nil
		}
		return nil, err
	}
	list := make([]Info, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !namePattern.MatchString(e.Name()) {
			continue
		}
		st, err := e.Info()
		if err != nil {
			continue
		}
		list = append(list, Info{Name: e.Name(), SizeBytes: st.Size(), CreatedAt: parseTime(e.Name(), st.ModTime())})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name > list[j].Name })
	return list, nil
}

// Path devuelve la ruta de una copia validando el nombre (sin rutas ni caracteres raros).
func (s *Service) Path(name string) (string, error) {
	if !namePattern.MatchString(name) {
		return "", errors.New("nombre de copia inválido")
	}
	p := filepath.Join(s.backupDir, name)
	if _, err := os.Stat(p); err != nil {
		return "", errors.New("la copia no existe")
	}
	return p, nil
}

// prune elimina las copias más antiguas que exceden la retención.
func (s *Service) prune() {
	list, err := s.List()
	if err != nil || len(list) <= s.retention {
		return
	}
	for _, old := range list[s.retention:] {
		if err := os.Remove(filepath.Join(s.backupDir, old.Name)); err != nil {
			slog.Warn("no se pudo eliminar una copia antigua", "name", old.Name, "error", err)
		}
	}
}

// StartScheduler crea una copia cada `interval` (comprobando cada hora si la última
// ya es más antigua que el intervalo, de modo que reiniciar el servidor no genera
// copias de más). Termina cuando se cancela ctx.
func (s *Service) StartScheduler(ctx context.Context, interval time.Duration) {
	check := func() {
		list, _ := s.List()
		if len(list) > 0 && time.Since(list[0].CreatedAt) < interval {
			return
		}
		info, err := s.Create(ctx)
		if err != nil {
			slog.Error("falló la copia de seguridad programada", "error", err)
			return
		}
		slog.Info("copia de seguridad creada", "name", info.Name, "bytes", info.SizeBytes)
	}

	go func() {
		// Primera comprobación poco después del arranque, para no competir con la migración.
		select {
		case <-time.After(s.firstCheckDelay):
		case <-ctx.Done():
			return
		}
		check()
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				check()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// ---------------------------------------------------------------- restauración

// Restore restaura una copia en dataDir. Debe ejecutarse con el servidor detenido.
// Lo que existía (base de datos y uploads) se mueve a dataDir/pre-restore-<fecha>/
// en lugar de borrarse.
func Restore(archivePath, dataDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("no se pudo abrir la copia: %w", err)
	}
	defer f.Close()

	// Se extrae primero a una carpeta temporal y se valida antes de tocar nada.
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(dataDir, ".restore-stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)

	m, err := extract(f, stage)
	if err != nil {
		return err
	}

	// Integridad: hash del manifiesto y PRAGMA integrity_check de SQLite.
	sum, _, err := hashFile(filepath.Join(stage, dbEntry))
	if err != nil {
		return err
	}
	if sum != m.DBSHA256 {
		return errors.New("la copia está dañada o fue modificada: el hash de la base de datos no coincide")
	}
	if err := integrityCheck(filepath.Join(stage, dbEntry)); err != nil {
		return err
	}

	// Se aparta lo existente (nunca se borra).
	backupOld := filepath.Join(dataDir, "pre-restore-"+time.Now().UTC().Format("20060102-150405"))
	moved := false
	for _, name := range []string{dbEntry, dbEntry + "-wal", dbEntry + "-shm", "uploads"} {
		src := filepath.Join(dataDir, name)
		if _, err := os.Lstat(src); err != nil {
			continue
		}
		if err := os.MkdirAll(backupOld, 0o755); err != nil {
			return err
		}
		if err := os.Rename(src, filepath.Join(backupOld, name)); err != nil {
			return fmt.Errorf("no se pudo apartar %s: %w", name, err)
		}
		moved = true
	}

	if err := os.Rename(filepath.Join(stage, dbEntry), filepath.Join(dataDir, dbEntry)); err != nil {
		return err
	}
	uploads := filepath.Join(stage, "uploads")
	if _, err := os.Stat(uploads); err == nil {
		if err := os.Rename(uploads, filepath.Join(dataDir, "uploads")); err != nil {
			return err
		}
	} else if err := os.MkdirAll(filepath.Join(dataDir, "uploads"), 0o755); err != nil {
		return err
	}
	if moved {
		slog.Info("datos anteriores conservados", "path", backupOld)
	}
	return nil
}

// extract descomprime el archivo en dir aceptando SOLO las entradas esperadas
// (base de datos, uploads planos y manifiesto): bloquea rutas con .., absolutas,
// enlaces y cualquier otro nombre.
func extract(r io.Reader, dir string) (*Manifest, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.New("el archivo no es una copia válida (.tar.gz)")
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	var manifest *Manifest
	haveDB := false
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("copia dañada: %w", err)
		}
		if h.Typeflag != tar.TypeReg {
			return nil, fmt.Errorf("entrada no permitida en la copia: %q", h.Name)
		}

		switch {
		case h.Name == manifestEntry:
			var m Manifest
			if err := json.NewDecoder(io.LimitReader(tr, 1<<20)).Decode(&m); err != nil {
				return nil, errors.New("manifiesto inválido")
			}
			manifest = &m
		case h.Name == dbEntry:
			if err := writeEntry(tr, filepath.Join(dir, dbEntry)); err != nil {
				return nil, err
			}
			haveDB = true
		case strings.HasPrefix(h.Name, uploadsPrefix):
			base := strings.TrimPrefix(h.Name, uploadsPrefix)
			if base == "" || base != filepath.Base(base) || strings.ContainsAny(base, `\/`) || base == "." || base == ".." {
				return nil, fmt.Errorf("entrada no permitida en la copia: %q", h.Name)
			}
			if err := os.MkdirAll(filepath.Join(dir, "uploads"), 0o755); err != nil {
				return nil, err
			}
			if err := writeEntry(tr, filepath.Join(dir, "uploads", base)); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("entrada no permitida en la copia: %q", h.Name)
		}
	}
	if manifest == nil || !haveDB {
		return nil, errors.New("la copia está incompleta (falta la base de datos o el manifiesto)")
	}
	return manifest, nil
}

func writeEntry(r io.Reader, dst string) error {
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func integrityCheck(path string) error {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()
	var res string
	if err := db.QueryRow(`PRAGMA integrity_check`).Scan(&res); err != nil {
		return fmt.Errorf("no se pudo verificar la base restaurada: %w", err)
	}
	if res != "ok" {
		return fmt.Errorf("la base de datos de la copia está corrupta: %s", res)
	}
	return nil
}

// ---------------------------------------------------------------- utilidades

func addFile(tw *tar.Writer, src, name string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: st.Size(), ModTime: st.ModTime()}); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

func hashFile(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func parseTime(name string, fallback time.Time) time.Time {
	ts := strings.TrimSuffix(strings.TrimPrefix(name, filePrefix), fileSuffix)
	if t, err := time.ParseInLocation("20060102-150405", ts, time.UTC); err == nil {
		return t
	}
	return fallback
}
