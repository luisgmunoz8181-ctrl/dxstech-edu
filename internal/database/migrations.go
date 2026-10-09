package database

import (
	"database/sql"
	"fmt"
	"log"
)

// migration es un cambio de esquema versionado. Cada migración se ejecuta una
// sola vez, dentro de una transacción, y queda registrada en schema_migrations.
//
// Reglas para añadir cambios de esquema:
//   - Nunca edite una migración ya publicada; agregue una nueva al final.
//   - Use el siguiente número de versión consecutivo.
//   - Los errores se propagan: ya no se ignoran silenciosamente.
type migration struct {
	version int
	name    string
	up      func(tx *sql.Tx) error
}

var migrations = []migration{
	{1, "baseline_schema", migrateBaseline},
	{2, "users_token_version", func(tx *sql.Tx) error {
		return addColumnIfMissing(tx, "users", "token_version", "INTEGER NOT NULL DEFAULT 0")
	}},
	{3, "performance_indexes", migratePerformanceIndexes},
	{4, "account_lockout_and_audit_request_id", func(tx *sql.Tx) error {
		for _, c := range []struct{ table, column, def string }{
			{"users", "failed_login_attempts", "INTEGER NOT NULL DEFAULT 0"},
			{"users", "locked_until", "DATETIME"},
			{"audit_logs", "request_id", "TEXT"},
		} {
			if err := addColumnIfMissing(tx, c.table, c.column, c.def); err != nil {
				return err
			}
		}
		return nil
	}},
	{5, "forum_titles_and_replies", func(tx *sql.Tx) error {
		for _, c := range []struct{ table, column, def string }{
			{"course_discussions", "title", "TEXT NOT NULL DEFAULT ''"},
			{"course_discussions", "parent_id", "TEXT NOT NULL DEFAULT ''"},
		} {
			if err := addColumnIfMissing(tx, c.table, c.column, c.def); err != nil {
				return err
			}
		}
		_, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_discussions_parent ON course_discussions(course_id, parent_id)`)
		return err
	}},
	{6, "forum_reports", func(tx *sql.Tx) error {
		// Los reportes guardan una copia del mensaje y no tienen clave foránea hacia
		// course_discussions: si un moderador elimina la publicación, el historial se conserva.
		_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS discussion_reports (
			id TEXT PRIMARY KEY,
			discussion_id TEXT NOT NULL,
			course_id TEXT NOT NULL,
			reporter_id TEXT NOT NULL,
			reporter_name TEXT NOT NULL DEFAULT '',
			reason TEXT NOT NULL DEFAULT '',
			post_author_id TEXT NOT NULL DEFAULT '',
			post_author_name TEXT NOT NULL DEFAULT '',
			post_message TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'open',
			resolved_by TEXT,
			resolved_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (discussion_id, reporter_id)
		);
		CREATE INDEX IF NOT EXISTS idx_discussion_reports_status ON discussion_reports(status, created_at);
		CREATE INDEX IF NOT EXISTS idx_discussion_reports_discussion ON discussion_reports(discussion_id);
		`)
		return err
	}},
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return fmt.Errorf("no se pudo crear schema_migrations: %w", err)
	}

	applied := map[int]bool{}
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if err := m.up(tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migración %d (%s): %w", m.version, m.name, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version, name) VALUES (?, ?)`, m.version, m.name); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Printf("🗄️  Migración %d aplicada: %s", m.version, m.name)
	}
	return nil
}

// addColumnIfMissing añade una columna solo si aún no existe (SQLite no admite
// ADD COLUMN IF NOT EXISTS), propagando cualquier error real.
func addColumnIfMissing(tx *sql.Tx, table, column, definition string) error {
	rows, err := tx.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		if name == column {
			found = true
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if found {
		return nil
	}
	_, err = tx.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, definition))
	return err
}

// migrateBaseline reúne el esquema histórico (fases 1-6). Es idempotente, de
// modo que las bases de datos creadas antes del versionado la adoptan sin cambios.
func migrateBaseline(tx *sql.Tx) error {
	if _, err := tx.Exec(baselineSchema); err != nil {
		return err
	}

	legacyColumns := []struct{ table, column, def string }{
		{"issued_certificates", "user_id", "TEXT"},
		{"issued_certificates", "course_id", "TEXT"},
		{"issued_certificates", "duration_hours", "REAL DEFAULT 0.0"},
		{"issued_certificates", "instructor_name", "TEXT DEFAULT 'DxSTech Edu'"},
		{"lessons", "quiz_id", "TEXT DEFAULT ''"},
		{"quizzes", "course_id", "TEXT DEFAULT ''"},
		{"quizzes", "lesson_id", "TEXT DEFAULT ''"},
	}
	for _, c := range legacyColumns {
		if err := addColumnIfMissing(tx, c.table, c.column, c.def); err != nil {
			return fmt.Errorf("%s.%s: %w", c.table, c.column, err)
		}
	}

	_, err := tx.Exec(phase6Schema)
	return err
}

// migratePerformanceIndexes añade los índices que respaldan las consultas más
// frecuentes (filtros por curso/usuario, ordenamientos y agregados).
func migratePerformanceIndexes(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role_id);

	CREATE INDEX IF NOT EXISTS idx_courses_status_created ON courses(status, created_at);
	CREATE INDEX IF NOT EXISTS idx_courses_category ON courses(category);

	CREATE INDEX IF NOT EXISTS idx_modules_course ON course_modules(course_id, order_index);
	CREATE INDEX IF NOT EXISTS idx_lessons_course ON lessons(course_id, order_index);
	CREATE INDEX IF NOT EXISTS idx_lessons_module ON lessons(module_id, order_index);

	CREATE INDEX IF NOT EXISTS idx_enrollments_course_status ON enrollments(course_id, status);
	CREATE INDEX IF NOT EXISTS idx_enrollments_last_accessed ON enrollments(last_accessed_at);
	CREATE INDEX IF NOT EXISTS idx_enrollments_enrolled_at ON enrollments(enrolled_at);
	CREATE INDEX IF NOT EXISTS idx_lesson_progress_course_user ON lesson_progress(course_id, user_id);

	CREATE INDEX IF NOT EXISTS idx_quizzes_lesson ON quizzes(lesson_id);
	CREATE INDEX IF NOT EXISTS idx_quizzes_course ON quizzes(course_id);
	CREATE INDEX IF NOT EXISTS idx_quizzes_created ON quizzes(created_at);
	CREATE INDEX IF NOT EXISTS idx_quiz_submissions_quiz_user ON quiz_submissions(quiz_id, user_id, submitted_at);

	CREATE INDEX IF NOT EXISTS idx_discussions_course ON course_discussions(course_id, created_at);

	CREATE INDEX IF NOT EXISTS idx_certs_created ON issued_certificates(created_at);
	CREATE INDEX IF NOT EXISTS idx_certs_course ON issued_certificates(course_id);

	CREATE INDEX IF NOT EXISTS idx_audit_user_created ON audit_logs(user_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token_hash);
	CREATE INDEX IF NOT EXISTS idx_whatsapp_logs_created ON whatsapp_logs(created_at);
	`)
	return err
}

const baselineSchema = `
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
		user_id TEXT,
		course_id TEXT,
		student_name TEXT NOT NULL,
		course_title TEXT NOT NULL,
		duration_hours REAL NOT NULL DEFAULT 0.0,
		instructor_name TEXT NOT NULL DEFAULT 'DxSTech Edu',
		issue_date TEXT NOT NULL,
		qr_code_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS courses (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		code TEXT UNIQUE NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		short_description TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		thumbnail_url TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT 'Tecnología',
		instructor_name TEXT NOT NULL DEFAULT 'Equipo DxSTech',
		duration_hours REAL NOT NULL DEFAULT 1.0,
		level TEXT NOT NULL DEFAULT 'Principiante',
		status TEXT NOT NULL DEFAULT 'draft',
		published_at DATETIME,
		requirements TEXT NOT NULL DEFAULT '',
		learning_objectives TEXT NOT NULL DEFAULT '',
		created_by TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS course_modules (
		id TEXT PRIMARY KEY,
		course_id TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		order_index INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS lessons (
		id TEXT PRIMARY KEY,
		module_id TEXT NOT NULL,
		course_id TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		content_type TEXT NOT NULL DEFAULT 'text',
		content_url TEXT NOT NULL DEFAULT '',
		content_body TEXT NOT NULL DEFAULT '',
		duration_minutes INTEGER NOT NULL DEFAULT 10,
		order_index INTEGER NOT NULL DEFAULT 0,
		is_free_preview INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (module_id) REFERENCES course_modules(id) ON DELETE CASCADE,
		FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS enrollments (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		course_id TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		progress_percent REAL NOT NULL DEFAULT 0.0,
		enrolled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		last_accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_lesson_id TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
		UNIQUE (user_id, course_id)
	);

	CREATE TABLE IF NOT EXISTS lesson_progress (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		course_id TEXT NOT NULL,
		lesson_id TEXT NOT NULL,
		completed INTEGER NOT NULL DEFAULT 1,
		completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
		FOREIGN KEY (lesson_id) REFERENCES lessons(id) ON DELETE CASCADE,
		UNIQUE (user_id, lesson_id)
	);

	INSERT OR IGNORE INTO whatsapp_config (id, is_active, knowledge_base, strict_mode, respond_groups, provider)
	VALUES (1, 1, 'DxSTech Edu es una academia digital líder en tecnología, programación e inteligencia artificial. Ofrecemos cursos prácticos con proyectos reales, tutoría personalizada y certificados digitales verificados.', 1, 0, 'Simulador / Baileys QR');
	`

const phase6Schema = `
	CREATE TABLE IF NOT EXISTS quiz_submissions (
		id TEXT PRIMARY KEY,
		quiz_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		course_id TEXT NOT NULL DEFAULT '',
		lesson_id TEXT NOT NULL DEFAULT '',
		score REAL NOT NULL,
		passed INTEGER NOT NULL DEFAULT 0,
		answers_json TEXT NOT NULL DEFAULT '',
		submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS course_discussions (
		id TEXT PRIMARY KEY,
		course_id TEXT NOT NULL,
		lesson_id TEXT NOT NULL DEFAULT '',
		user_id TEXT NOT NULL,
		user_name TEXT NOT NULL,
		user_role TEXT NOT NULL DEFAULT 'ESTUDIANTE',
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS course_reviews (
		id TEXT PRIMARY KEY,
		course_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		user_name TEXT NOT NULL,
		rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
		comment TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(course_id, user_id)
	);
	`
