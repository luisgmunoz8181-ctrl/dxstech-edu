package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func InitDB(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "dxstech.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Optimize SQLite performance and concurrent readers
	if _, err := db.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA synchronous=NORMAL;
		PRAGMA busy_timeout=5000;
	`); err != nil {
		// Non-fatal if WAL is not supported in some environments
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return &DB{DB: db}, nil
}

func runMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS quizzes (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		notes TEXT,
		question_count INTEGER NOT NULL,
		questions_json TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS whatsapp_config (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		is_active INTEGER NOT NULL DEFAULT 1,
		knowledge_base TEXT NOT NULL DEFAULT '',
		strict_mode INTEGER NOT NULL DEFAULT 1,
		respond_groups INTEGER NOT NULL DEFAULT 0,
		provider TEXT NOT NULL DEFAULT 'Simulador / Baileys QR',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS whatsapp_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		recipient TEXT NOT NULL,
		message TEXT NOT NULL,
		status TEXT NOT NULL,
		error_detail TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS issued_certificates (
		id TEXT PRIMARY KEY,
		student_name TEXT NOT NULL,
		course_title TEXT NOT NULL,
		issue_date TEXT NOT NULL,
		qr_code_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	INSERT OR IGNORE INTO whatsapp_config (id, is_active, knowledge_base, strict_mode, respond_groups, provider)
	VALUES (1, 1, 'DxSTech Edu es una academia digital líder en tecnología, programación e inteligencia artificial. Ofrecemos cursos prácticos con proyectos reales, tutoría personalizada y certificados digitales verificados.', 1, 0, 'Simulador / Baileys QR');
	`

	_, err := db.Exec(schema)
	return err
}
