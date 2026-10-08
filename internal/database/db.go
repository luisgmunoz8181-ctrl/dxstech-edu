package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
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
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite file databases require serialized write access to avoid SQLITE_BUSY / database is locked.
	db.SetMaxOpenConns(1)

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return &DB{DB: db}, nil
}

func runMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		description TEXT NOT NULL
	);

	INSERT OR IGNORE INTO roles (id, name, description) VALUES
		(1, 'SUPERADMIN', 'Acceso y control total de la plataforma'),
		(2, 'ADMINISTRADOR', 'Gestión de usuarios, cursos, contenidos y reportes'),
		(3, 'ESTUDIANTE', 'Acceso a cursos asignados, aula virtual y certificados');

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role_id INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		email_verified INTEGER NOT NULL DEFAULT 1,
		must_change_password INTEGER NOT NULL DEFAULT 0,
		identification TEXT DEFAULT '',
		company TEXT DEFAULT '',
		job_title TEXT DEFAULT '',
		last_login DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		password_changed_at DATETIME,
		FOREIGN KEY (role_id) REFERENCES roles(id)
	);

	CREATE TABLE IF NOT EXISTS password_resets (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		token_hash TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		used INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		action TEXT NOT NULL,
		resource TEXT NOT NULL,
		ip_address TEXT,
		user_agent TEXT,
		details TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

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

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	return seedDefaultUsers(db)
}

func seedDefaultUsers(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	// Hashes de contraseñas de desarrollo para testing inmediato
	// Admin1234* para Superadmin y Administrador
	adminHash, err := bcrypt.GenerateFromPassword([]byte("Admin1234*"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Student1234* para Estudiante
	studentHash, err := bcrypt.GenerateFromPassword([]byte("Student1234*"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare(`
		INSERT INTO users (id, first_name, last_name, email, password_hash, role_id, status, email_verified, must_change_password)
		VALUES (?, ?, ?, ?, ?, ?, 'active', 1, 0)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// 1. Superadmin
	if _, err := stmt.Exec("usr-superadmin-01", "Super", "Admin", "superadmin@dxstech.edu", string(adminHash), 1); err != nil {
		return err
	}

	// 2. Administrador
	if _, err := stmt.Exec("usr-admin-01", "Admin", "DxSTech", "admin@dxstech.edu", string(adminHash), 2); err != nil {
		return err
	}

	// 3. Estudiante
	if _, err := stmt.Exec("usr-student-01", "Carlos", "Estudiante", "estudiante@dxstech.edu", string(studentHash), 3); err != nil {
		return err
	}

	return nil
}
