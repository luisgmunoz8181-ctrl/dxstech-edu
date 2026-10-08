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

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	if err := seedDefaultUsers(db); err != nil {
		return err
	}

	if err := seedDemoCourses(db); err != nil {
		return err
	}

	return seedDemoEnrollments(db)
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

func seedDemoCourses(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM courses").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// 1. Curso 1: Publicado
	_, err = db.Exec(`
		INSERT INTO courses (
			id, title, code, slug, short_description, description, thumbnail_url,
			category, instructor_name, duration_hours, level, status, published_at,
			requirements, learning_objectives, created_by
		) VALUES (
			'crs-ai-101',
			'Inducción a la Inteligencia Artificial & Agentes Autónomos',
			'DXS-AI-101',
			'induccion-inteligencia-artificial-agentes',
			'Aprende los fundamentos de modelos generativos, prompting avanzado y diseño de agentes inteligentes autónomos.',
			'Este curso profesional introduce las bases teórico-prácticas para comprender el funcionamiento de los modelos de lenguaje modernos (LLMs), arquitecturas transformer, técnicas de ingeniería de contexto y construcción de flujos de trabajo autónomos.',
			'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=1200&q=80',
			'Inteligencia Artificial',
			'Dr. Alexander Gómez',
			8.5,
			'Principiante',
			'published',
			CURRENT_TIMESTAMP,
			'Conocimientos básicos de computación e interés en herramientas de inteligencia artificial.',
			'Dominar la terminología de modelos generativos; Aplicar técnicas de prompting estructurado; Comprender la arquitectura de agentes multi-herramienta.',
			'usr-superadmin-01'
		)
	`)
	if err != nil {
		return err
	}

	// Módulos para Curso 1
	_, err = db.Exec(`
		INSERT INTO course_modules (id, course_id, title, description, order_index) VALUES
		('mod-ai-01', 'crs-ai-101', 'Módulo 1: Fundamentos de Modelos Generativos', 'Introducción al estado del arte, arquitectura y conceptos clave de la IA moderna.', 1),
		('mod-ai-02', 'crs-ai-101', 'Módulo 2: Construcción de Agentes Autónomos', 'Diseño e implementación de sistemas que razonan, usan herramientas y ejecutan tareas complejas.', 2)
	`)
	if err != nil {
		return err
	}

	// Lecciones para Curso 1
	_, err = db.Exec(`
		INSERT INTO lessons (id, module_id, course_id, title, description, content_type, content_url, content_body, duration_minutes, order_index, is_free_preview) VALUES
		('lsn-ai-01', 'mod-ai-01', 'crs-ai-101', 'Bienvenida y Hoja de Ruta del Curso', 'Visión general de las tecnologías que dominaremos en este programa.', 'youtube', 'https://www.youtube.com/watch?v=aircAruvnKk', '', 15, 1, 1),
		('lsn-ai-02', 'mod-ai-01', 'crs-ai-101', 'Guía de Arquitectura de LLMs y Transformers', 'Lectura académica sobre el mecanismo de auto-atención en transformers.', 'pdf', 'https://arxiv.org/pdf/1706.03762.pdf', '', 25, 2, 1),
		('lsn-ai-03', 'mod-ai-01', 'crs-ai-101', 'Principios de Context Engineering y Prompting', 'Guía práctica para estructurar prompts con roles, restricciones y formato JSON.', 'text', '', '### Principios de Context Engineering\n\nEl prompting moderno no se trata solo de hacer preguntas, sino de estructurar el contexto del modelo:\n\n1. **Rol y Propósito:** Define quién es el modelo y qué objetivo persigue.\n2. **Instrucciones Claras:** Usa delimitadores y reglas explícitas.\n3. **Ejemplos (Few-shot):** Proporciona entradas y salidas de referencia.\n4. **Restricciones de Salida:** Exige formatos estructurados como JSON o Markdown.', 20, 3, 0),
		('lsn-ai-04', 'mod-ai-02', 'crs-ai-101', 'Patrones de Agentes: ReAct, Planning y Memoria', 'Sesión magistral sobre el estado de la técnica y diseño de agentes.', 'youtube', 'https://www.youtube.com/watch?v=sal78ACtGTc', '', 35, 1, 0),
		('lsn-ai-05', 'mod-ai-02', 'crs-ai-101', 'Presentación Ejecutiva: Casos de Uso Empresariales', 'Diapositivas descargables y visualizables de arquitectura empresarial.', 'pptx', 'https://view.officeapps.live.com/op/view.aspx?src=https://scholar.harvard.edu/files/torman/files/sample.pptx', '', 20, 2, 0)
	`)
	if err != nil {
		return err
	}

	// 2. Curso 2: Borrador (Draft)
	_, err = db.Exec(`
		INSERT INTO courses (
			id, title, code, slug, short_description, description, thumbnail_url,
			category, instructor_name, duration_hours, level, status, published_at,
			requirements, learning_objectives, created_by
		) VALUES (
			'crs-dev-201',
			'Desarrollo Web Full-Stack Moderno con Go y Tailwind',
			'DXS-DEV-201',
			'desarrollo-web-fullstack-go-tailwind',
			'Crea aplicaciones web robustas, rápidas y escalables con backend en Go y frontend responsivo.',
			'En este curso aprenderás a estructurar un servidor HTTP en Go con Gin, SQLite/PostgreSQL, middleware JWT, y construir una SPA moderna sin sobrecarga de dependencias utilizando Tailwind CSS.',
			'https://images.unsplash.com/photo-1555066931-4365d14bab8c?auto=format&fit=crop&w=1200&q=80',
			'Desarrollo Web',
			'Ing. Sofía Morales',
			12.0,
			'Intermedio',
			'draft',
			NULL,
			'Conocimientos básicos de programación en cualquier lenguaje.',
			'Aprender la sintaxis y concurrencia de Go; Construir APIs REST seguras; Diseñar interfaces responsivas con Tailwind CSS.',
			'usr-superadmin-01'
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO course_modules (id, course_id, title, description, order_index) VALUES
		('mod-dev-01', 'crs-dev-201', 'Módulo 1: Introducción a Go y Concurrencia', 'Sintaxis básica, estructuras, interfaces y goroutines.', 1)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO lessons (id, module_id, course_id, title, description, content_type, content_url, content_body, duration_minutes, order_index, is_free_preview) VALUES
		('lsn-dev-01', 'mod-dev-01', 'crs-dev-201', 'Instalación y Tu Primer Servidor en Go', 'Configuración del entorno y primeros pasos.', 'text', '', 'Bienvenido al curso de desarrollo full-stack. En esta primera sesión configuraremos Go 1.25 y exploraremos el paquete net/http.', 15, 1, 1)
	`)

	return err
}

func seedDemoEnrollments(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM enrollments").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// Matricular estudiante demo en curso de IA
	_, err = db.Exec(`
		INSERT INTO enrollments (
			id, user_id, course_id, status, progress_percent,
			enrolled_at, last_accessed_at, last_lesson_id
		) VALUES (
			'enr-demo-student-01',
			'usr-student-01',
			'crs-ai-101',
			'active',
			20.0,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP,
			'lsn-ai-01'
		)
	`)
	if err != nil {
		return err
	}

	// Marcar lección 1 como completada
	_, err = db.Exec(`
		INSERT INTO lesson_progress (id, user_id, course_id, lesson_id, completed, completed_at)
		VALUES ('prg-demo-01', 'usr-student-01', 'crs-ai-101', 'lsn-ai-01', 1, CURRENT_TIMESTAMP)
	`)

	return err
}
