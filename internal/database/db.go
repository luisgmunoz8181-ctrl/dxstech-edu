package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// BcryptCost es el costo de bcrypt para generar hashes de contraseña. En
// producción es el valor por defecto; los tests lo reducen para ir más rápido.
var BcryptCost = bcrypt.DefaultCost

// Credenciales de las cuentas de demostración que siembra el modo desarrollo. Son
// públicas (están en el código y el README): cualquier cuenta que todavía las use en
// producción debe rotarse con `dxstech-server -rotate-demo-users`.
const (
	DemoAdminPassword   = "Admin1234*"
	DemoStudentPassword = "Student1234*"
)

// DemoUser describe una cuenta de demostración y la contraseña pública con que se siembra.
type DemoUser struct {
	Email    string
	Password string
}

// DemoUsers son las cuentas que siembra seedDefaultUsers.
var DemoUsers = []DemoUser{
	{"superadmin@dxstech.edu", DemoAdminPassword},
	{"admin@dxstech.edu", DemoAdminPassword},
	{"estudiante@dxstech.edu", DemoStudentPassword},
}

// maxOpenConns es el tamaño del pool de conexiones SQLite (WAL).
const maxOpenConns = 8

type DB struct {
	*sql.DB
}

// OpenExisting abre una base de datos que ya existe en dataDir (sin crearla, sembrar datos ni
// migrar), para las herramientas de línea de comandos. Falla con un mensaje claro si no existe.
func OpenExisting(dataDir string) (*DB, error) {
	dbPath := filepath.Join(dataDir, "dxstech.db")
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("no se encontró la base de datos en %s (revisa DATA_DIR)", dbPath)
	}
	dsn := fmt.Sprintf("%s?_txlock=immediate&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	// Una base creada por una versión anterior debe migrarse antes de usarla con esta herramienta.
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migración de la base de datos: %w", err)
	}
	return &DB{DB: db}, nil
}

// InitDB abre la base de datos y aplica las migraciones.
//
// seedDemo controla si se siembran usuarios, cursos y matrículas de
// demostración (con credenciales conocidas). Solo debe ser true en desarrollo
// y tests. Con seedDemo=false, si la base está vacía se crea un único
// SUPERADMIN inicial (ver bootstrapSuperadmin).
func InitDB(dataDir string, seedDemo bool) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "dxstech.db")
	// WAL permite varios lectores concurrentes con un escritor. _txlock=immediate
	// hace que las transacciones tomen el bloqueo de escritura al inicio y
	// esperen (busy_timeout) en vez de fallar al promoverlo. foreign_keys activa
	// los ON DELETE CASCADE declarados en el esquema (SQLite los ignora por defecto).
	dsn := fmt.Sprintf("%s?_txlock=immediate&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Pool de conexiones: las lecturas (analytics, catálogo, reportes CSV) ya no
	// quedan detrás de una única conexión; las escrituras se serializan en SQLite.
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxOpenConns)

	if err := runMigrations(db, seedDemo); err != nil {
		db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return &DB{DB: db}, nil
}

func runMigrations(db *sql.DB, seedDemo bool) error {
	if err := migrate(db); err != nil {
		return err
	}

	if !seedDemo {
		return bootstrapSuperadmin(db)
	}

	if err := seedDefaultUsers(db); err != nil {
		return err
	}

	if err := seedDemoCourses(db); err != nil {
		return err
	}

	return seedDemoEnrollments(db)
}

// bootstrapSuperadmin crea el primer SUPERADMIN cuando no existe ningún usuario.
//
// Credenciales: BOOTSTRAP_ADMIN_EMAIL (por defecto superadmin@dxstech.edu) y
// BOOTSTRAP_ADMIN_PASSWORD. Si no se define contraseña, se genera una aleatoria
// que se imprime una única vez en el log. En ambos casos el usuario queda con
// must_change_password=1.
func bootstrapSuperadmin(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	email := strings.ToLower(strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_EMAIL")))
	if email == "" {
		email = "superadmin@dxstech.edu"
	}

	password := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")
	generated := false
	if password == "" {
		b := make([]byte, 18)
		if _, err := rand.Read(b); err != nil {
			return err
		}
		password = base64.RawURLEncoding.EncodeToString(b)
		generated = true
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO users (id, first_name, last_name, email, password_hash, role_id, status, email_verified, must_change_password)
		VALUES ('usr-superadmin-01', 'Super', 'Admin', ?, ?, 1, 'active', 1, 1)
	`, email, string(hash))
	if err != nil {
		return err
	}

	log.Printf("🔐 Se creó el SUPERADMIN inicial: %s", email)
	if generated {
		log.Printf("🔐 Contraseña temporal (se muestra una sola vez, cámbiela al ingresar): %s", password)
	}
	return nil
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
	adminHash, err := bcrypt.GenerateFromPassword([]byte(DemoAdminPassword), BcryptCost)
	if err != nil {
		return err
	}

	// Student1234* para Estudiante
	studentHash, err := bcrypt.GenerateFromPassword([]byte(DemoStudentPassword), BcryptCost)
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

// UserCanAccessCourse indica si el usuario puede acceder al contenido
// interactivo de un curso (foro, tutor, reseñas): administradores siempre;
// estudiantes solo si están matriculados (y no han abandonado el curso).
func (d *DB) UserCanAccessCourse(ctx context.Context, userID, role, courseID string) bool {
	if role == "SUPERADMIN" || role == "ADMINISTRADOR" {
		return true
	}
	if userID == "" || courseID == "" {
		return false
	}
	var n int
	err := d.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM enrollments WHERE user_id = ? AND course_id = ? AND status != 'dropped'`,
		userID, courseID).Scan(&n)
	return err == nil && n > 0
}
