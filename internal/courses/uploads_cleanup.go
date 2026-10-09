package courses

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const uploadsURLPrefix = "/uploads/"

// uploadName devuelve el nombre de archivo de una URL "/uploads/<archivo>" o ""
// si la URL no apunta a un archivo local válido (externa, vacía o con rutas).
func uploadName(url string) string {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, uploadsURLPrefix) {
		return ""
	}
	name := strings.TrimPrefix(url, uploadsURLPrefix)
	if name == "" || name != filepath.Base(name) || strings.ContainsAny(name, `\/`) || name == "." || name == ".." {
		return ""
	}
	return name
}

// isUploadReferenced indica si algún registro sigue usando el archivo: como URL
// de lección o miniatura de curso, o enlazado dentro del texto de una lección
// (los cursos duplicados comparten archivos).
func (s *Service) isUploadReferenced(ctx context.Context, name string) bool {
	url := uploadsURLPrefix + name
	var n int
	err := s.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM lessons WHERE content_url = ?) +
			(SELECT COUNT(*) FROM courses WHERE thumbnail_url = ?) +
			(SELECT COUNT(*) FROM lessons WHERE instr(content_body, ?) > 0 OR instr(description, ?) > 0)
	`, url, url, name, name).Scan(&n)
	// Ante un error de consulta se conserva el archivo.
	return err != nil || n > 0
}

// removeUnreferencedUploads elimina de disco los archivos subidos que ya no
// referencia ningún registro. Es de mejor esfuerzo: los fallos solo se registran.
func (s *Service) removeUnreferencedUploads(ctx context.Context, urls ...string) {
	seen := map[string]bool{}
	for _, u := range urls {
		name := uploadName(u)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		if s.isUploadReferenced(ctx, name) {
			continue
		}
		if err := os.Remove(filepath.Join(s.uploadDir, name)); err != nil && !os.IsNotExist(err) {
			log.Printf("⚠️  No se pudo eliminar el archivo huérfano %s: %v", name, err)
		}
	}
}

// queryStrings ejecuta una consulta que devuelve una columna de texto.
func (s *Service) queryStrings(ctx context.Context, query string, args ...any) []string {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var v string
		if rows.Scan(&v) == nil {
			out = append(out, v)
		}
	}
	return out
}

// OrphanUpload describe un archivo de /uploads sin referencias.
type OrphanUpload struct {
	Name      string    `json:"name"`
	SizeBytes int64     `json:"sizeBytes"`
	ModTime   time.Time `json:"modTime"`
}

// CleanOrphanUploads recorre el directorio de subidas y detecta (o elimina, si
// apply es true) los archivos sin referencias. Ignora los modificados dentro de
// minAge: un archivo se sube antes de guardar la lección que lo usará.
func (s *Service) CleanOrphanUploads(ctx context.Context, apply bool, minAge time.Duration) ([]OrphanUpload, error) {
	entries, err := os.ReadDir(s.uploadDir)
	if err != nil {
		return nil, err
	}
	orphans := []OrphanUpload{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil || time.Since(info.ModTime()) < minAge {
			continue
		}
		if s.isUploadReferenced(ctx, e.Name()) {
			continue
		}
		orphans = append(orphans, OrphanUpload{Name: e.Name(), SizeBytes: info.Size(), ModTime: info.ModTime()})
		if apply {
			if err := os.Remove(filepath.Join(s.uploadDir, e.Name())); err != nil && !os.IsNotExist(err) {
				return orphans, err
			}
		}
	}
	return orphans, nil
}
