package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db  *database.DB
	cfg *config.Config
}

const (
	// Tras maxFailedAttempts contraseñas incorrectas seguidas, la cuenta se
	// bloquea lockDuration (además del límite por IP).
	maxFailedAttempts = 5
	lockDuration      = 15 * time.Minute
)

// LockedError indica que la cuenta está bloqueada temporalmente.
type LockedError struct{ Until time.Time }

func (e *LockedError) Error() string {
	mins := int(time.Until(e.Until).Minutes()) + 1
	return fmt.Sprintf("cuenta bloqueada temporalmente por intentos fallidos. Intenta de nuevo en %d min o restablece tu contraseña", mins)
}

// dummyHash permite igualar el tiempo de respuesta cuando el correo no existe,
// evitando enumerar usuarios por diferencia de latencia.
var (
	dummyHash     []byte
	dummyHashOnce sync.Once
)

func timingEqualizer(password string) {
	dummyHashOnce.Do(func() {
		dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dxstech-timing-equalizer"), database.BcryptCost)
	})
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
}

func NewService(db *database.DB, cfg *config.Config) *Service {
	return &Service{
		db:  db,
		cfg: cfg,
	}
}

func (s *Service) Login(ctx context.Context, req LoginRequest, clientIP, userAgent string) (*LoginResponse, error) {
	email, err := NormalizeAndValidateEmail(req.Email)
	if err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.email, u.password_hash, u.role_id,
		       r.name as role_name, u.status, u.email_verified, u.must_change_password,
		       u.identification, u.company, u.job_title, u.last_login, u.created_at, u.updated_at,
		       u.token_version, u.locked_until
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE lower(u.email) = ?
	`, email)

	var u User
	var lastLogin, lockedUntil sql.NullTime
	var iden, comp, job sql.NullString

	err = row.Scan(
		&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.RoleID,
		&u.Role, &u.Status, &u.EmailVerified, &u.MustChangePassword,
		&iden, &comp, &job, &lastLogin, &u.CreatedAt, &u.UpdatedAt,
		&u.TokenVersion, &lockedUntil,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			timingEqualizer(req.Password)
			s.LogAudit(nil, "LOGIN_FAILED", "auth", clientIP, userAgent, fmt.Sprintf("Usuario no encontrado: %s", email))
			return nil, errors.New("correo electrónico o contraseña incorrectos")
		}
		return nil, fmt.Errorf("error de base de datos durante autenticación: %w", err)
	}

	if iden.Valid {
		u.Identification = iden.String
	}
	if comp.Valid {
		u.Company = comp.String
	}
	if job.Valid {
		u.JobTitle = job.String
	}
	if lastLogin.Valid {
		u.LastLogin = &lastLogin.Time
	}

	// Cuenta bloqueada: se rechaza antes de comprobar la contraseña para que
	// seguir adivinando no tenga efecto durante el bloqueo.
	if lockedUntil.Valid && lockedUntil.Time.After(time.Now()) {
		s.LogAudit(&u.ID, "LOGIN_BLOCKED", "auth", clientIP, userAgent, "Cuenta bloqueada temporalmente")
		return nil, &LockedError{Until: lockedUntil.Time}
	}

	// Verify password hash
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		s.LogAudit(&u.ID, "LOGIN_FAILED", "auth", clientIP, userAgent, "Contraseña incorrecta")
		s.registerFailedAttempt(ctx, u.ID, clientIP, userAgent)
		return nil, errors.New("correo electrónico o contraseña incorrectos")
	}

	// Verify active status
	if u.Status != "active" {
		s.LogAudit(&u.ID, "LOGIN_BLOCKED", "auth", clientIP, userAgent, "Usuario desactivado")
		return nil, errors.New("su cuenta se encuentra inactiva. Contacte al administrador de DxSTech Edu")
	}

	// Calculate token lifetime: 30 days if remember me, 24 hours otherwise
	duration := 24 * time.Hour
	if req.RememberMe {
		duration = 30 * 24 * time.Hour
	}

	token, err := GenerateToken(&u, s.cfg.JWTSecret, duration)
	if err != nil {
		return nil, fmt.Errorf("error generando sesión segura: %w", err)
	}

	// Update last_login
	now := time.Now()
	_, _ = s.db.ExecContext(ctx, `UPDATE users SET last_login = ?, failed_login_attempts = 0, locked_until = NULL WHERE id = ?`, now, u.ID)
	u.LastLogin = &now

	s.LogAudit(&u.ID, "LOGIN_SUCCESS", "auth", clientIP, userAgent, fmt.Sprintf("Inicio de sesión exitoso con rol %s", u.Role))

	return &LoginResponse{
		Token: token,
		User:  u,
	}, nil
}

func (s *Service) GetMe(ctx context.Context, userID string) (*User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.first_name, u.last_name, u.email, u.role_id,
		       r.name as role_name, u.status, u.email_verified, u.must_change_password,
		       u.identification, u.company, u.job_title, u.last_login, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = ?
	`, userID)

	var u User
	var lastLogin sql.NullTime
	var iden, comp, job sql.NullString

	err := row.Scan(
		&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.RoleID,
		&u.Role, &u.Status, &u.EmailVerified, &u.MustChangePassword,
		&iden, &comp, &job, &lastLogin, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		return nil, errors.New("usuario no encontrado")
	}

	if iden.Valid {
		u.Identification = iden.String
	}
	if comp.Valid {
		u.Company = comp.String
	}
	if job.Valid {
		u.JobTitle = job.String
	}
	if lastLogin.Valid {
		u.LastLogin = &lastLogin.Time
	}

	return &u, nil
}

// registerFailedAttempt suma un intento fallido y bloquea la cuenta al llegar al límite.
func (s *Service) registerFailedAttempt(ctx context.Context, userID, clientIP, userAgent string) {
	var attempts int
	if err := s.db.QueryRowContext(ctx, `
		UPDATE users SET failed_login_attempts = failed_login_attempts + 1 WHERE id = ?
		RETURNING failed_login_attempts`, userID).Scan(&attempts); err != nil {
		return
	}
	if attempts >= maxFailedAttempts {
		until := time.Now().Add(lockDuration)
		_, _ = s.db.ExecContext(ctx, `UPDATE users SET failed_login_attempts = 0, locked_until = ? WHERE id = ?`, until, userID)
		s.LogAudit(&userID, "ACCOUNT_LOCKED", "auth", clientIP, userAgent, fmt.Sprintf("Bloqueo de %s tras %d intentos fallidos", lockDuration, maxFailedAttempts))
	}
}

// UnlockUser levanta el bloqueo temporal de una cuenta (acción de administrador).
func (s *Service) UnlockUser(ctx context.Context, adminID, targetID, clientIP, userAgent string) (*User, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE id = ?`, targetID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("usuario no encontrado")
	}
	s.LogAudit(&adminID, "UNLOCK_USER", "users", clientIP, userAgent, fmt.Sprintf("Cuenta desbloqueada: %s", targetID))
	return s.GetMe(ctx, targetID)
}

// IssueToken emite un token nuevo para un usuario ya autenticado (p. ej. tras
// cambiar la contraseña, que invalida los tokens anteriores).
func (s *Service) IssueToken(ctx context.Context, userID string, duration time.Duration) (string, error) {
	u, err := s.GetMe(ctx, userID)
	if err != nil {
		return "", err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT token_version FROM users WHERE id = ?`, userID).Scan(&u.TokenVersion); err != nil {
		return "", err
	}
	return GenerateToken(u, s.cfg.JWTSecret, duration)
}

func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword, clientIP, userAgent string) error {
	var currentHash, email, firstName, lastName string
	var lockedUntil sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT password_hash, email, first_name, last_name, locked_until FROM users WHERE id = ?`, userID).
		Scan(&currentHash, &email, &firstName, &lastName, &lockedUntil)
	if err != nil {
		return errors.New("usuario no encontrado")
	}
	if lockedUntil.Valid && lockedUntil.Time.After(time.Now()) {
		return &LockedError{Until: lockedUntil.Time}
	}

	if err := ValidatePasswordPolicy(newPassword, email, firstName, lastName); err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword)); err != nil {
		s.LogAudit(&userID, "CHANGE_PASSWORD_FAILED", "users", clientIP, userAgent, "Contraseña actual incorrecta")
		// Una sesión robada no puede adivinar la contraseña actual sin límite.
		s.registerFailedAttempt(ctx, userID, clientIP, userAgent)
		return errors.New("la contraseña actual no es correcta")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), database.BcryptCost)
	if err != nil {
		return errors.New("error procesando nueva contraseña")
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx, `
		UPDATE users 
		SET password_hash = ?, password_changed_at = ?, must_change_password = 0, updated_at = ?,
		    token_version = token_version + 1, failed_login_attempts = 0, locked_until = NULL
		WHERE id = ?
	`, string(newHash), now, now, userID)
	if err != nil {
		return fmt.Errorf("error actualizando contraseña: %w", err)
	}

	s.LogAudit(&userID, "CHANGE_PASSWORD_SUCCESS", "users", clientIP, userAgent, "Contraseña cambiada exitosamente")
	return nil
}

func (s *Service) ForgotPassword(ctx context.Context, email, clientIP, userAgent string) (string, error) {
	normEmail, err := NormalizeAndValidateEmail(email)
	if err != nil {
		return "", nil // Security rule: always succeed silently to prevent user enumeration
	}

	var userID string
	err = s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE lower(email) = ? AND status = 'active'`, normEmail).Scan(&userID)
	if err != nil {
		s.LogAudit(nil, "FORGOT_PASSWORD_REQUEST", "auth", clientIP, userAgent, fmt.Sprintf("Solicitud para correo inexistente: %s", normEmail))
		return "", nil
	}

	// Generate secure 32-byte token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("error generando token de recuperación")
	}
	rawToken := hex.EncodeToString(b)

	// Store SHA-256 hash of token
	tokenHash := sha256Hex(rawToken)
	expiresAt := time.Now().Add(1 * time.Hour)
	resetID := fmt.Sprintf("rst-%d", time.Now().UnixNano())

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO password_resets (id, email, token_hash, expires_at, used)
		VALUES (?, ?, ?, ?, 0)
	`, resetID, normEmail, tokenHash, expiresAt)
	if err != nil {
		return "", fmt.Errorf("error guardando solicitud de recuperación: %w", err)
	}

	s.LogAudit(&userID, "FORGOT_PASSWORD_GENERATED", "auth", clientIP, userAgent, fmt.Sprintf("Token de recuperación generado para %s", normEmail))

	// Returns the raw token (in development it can be previewed or sent via email)
	return rawToken, nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword, clientIP, userAgent string) error {
	if err := ValidatePasswordPolicy(newPassword); err != nil {
		return err
	}

	tokenHash := sha256Hex(rawToken)
	row := s.db.QueryRowContext(ctx, `
		SELECT id, email, expires_at, used
		FROM password_resets
		WHERE token_hash = ? AND used = 0
	`, tokenHash)

	var resetID, email string
	var expiresAt time.Time
	var used int

	err := row.Scan(&resetID, &email, &expiresAt, &used)
	if err != nil {
		return errors.New("el enlace de restablecimiento es inválido o ya ha sido utilizado")
	}

	if time.Now().After(expiresAt) {
		return errors.New("el enlace de restablecimiento ha expirado. Solicita uno nuevo")
	}

	var firstName, lastName string
	_ = s.db.QueryRowContext(ctx, `SELECT first_name, last_name FROM users WHERE lower(email) = ?`, email).Scan(&firstName, &lastName)
	if err := ValidatePasswordPolicy(newPassword, email, firstName, lastName); err != nil {
		return err
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), database.BcryptCost)
	if err != nil {
		return errors.New("error procesando contraseña")
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx, `
		UPDATE users 
		SET password_hash = ?, password_changed_at = ?, must_change_password = 0, updated_at = ?,
		    token_version = token_version + 1, failed_login_attempts = 0, locked_until = NULL
		WHERE lower(email) = ?
	`, string(newHash), now, now, email)
	if err != nil {
		return fmt.Errorf("error restableciendo contraseña: %w", err)
	}

	// Invalidate token
	_, _ = s.db.ExecContext(ctx, `UPDATE password_resets SET used = 1 WHERE id = ?`, resetID)

	s.LogAudit(nil, "RESET_PASSWORD_SUCCESS", "auth", clientIP, userAgent, fmt.Sprintf("Contraseña restablecida exitosamente para %s", email))
	return nil
}

func (s *Service) ListUsers(ctx context.Context, roleFilter int, search string) ([]User, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email, u.role_id,
		       r.name as role_name, u.status, u.email_verified, u.must_change_password,
		       u.identification, u.company, u.job_title, u.last_login, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE 1=1
	`
	var args []interface{}

	if roleFilter > 0 {
		query += " AND u.role_id = ?"
		args = append(args, roleFilter)
	}

	if strings.TrimSpace(search) != "" {
		sPattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query += " AND (lower(u.first_name) LIKE ? OR lower(u.last_name) LIKE ? OR lower(u.email) LIKE ? OR lower(u.identification) LIKE ?)"
		args = append(args, sPattern, sPattern, sPattern, sPattern)
	}

	query += " ORDER BY u.created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var lastLogin sql.NullTime
		var iden, comp, job sql.NullString

		if err := rows.Scan(
			&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.RoleID,
			&u.Role, &u.Status, &u.EmailVerified, &u.MustChangePassword,
			&iden, &comp, &job, &lastLogin, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			continue
		}

		if iden.Valid {
			u.Identification = iden.String
		}
		if comp.Valid {
			u.Company = comp.String
		}
		if job.Valid {
			u.JobTitle = job.String
		}
		if lastLogin.Valid {
			u.LastLogin = &lastLogin.Time
		}

		users = append(users, u)
	}

	return users, nil
}

func (s *Service) CreateUser(ctx context.Context, adminID string, req CreateUserRequest, clientIP, userAgent string) (*User, error) {
	email, err := NormalizeAndValidateEmail(req.Email)
	if err != nil {
		return nil, err
	}

	if err := ValidatePasswordPolicy(req.Password, email, req.FirstName, req.LastName); err != nil {
		return nil, err
	}

	// Check duplicates
	var exists int
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE lower(email) = ?`, email).Scan(&exists)
	if err == nil && exists > 0 {
		return nil, errors.New("ya existe un usuario registrado con este correo electrónico")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), database.BcryptCost)
	if err != nil {
		return nil, errors.New("error encriptando contraseña")
	}

	userID := fmt.Sprintf("usr-%d", time.Now().UnixNano())
	now := time.Now()

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, role_id, status, email_verified, must_change_password, identification, company, job_title, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'active', 1, 1, ?, ?, ?, ?, ?)
	`, userID, strings.TrimSpace(req.FirstName), strings.TrimSpace(req.LastName), email, string(hash), req.RoleID, req.Identification, req.Company, req.JobTitle, now, now)

	if err != nil {
		return nil, fmt.Errorf("error guardando usuario: %w", err)
	}

	s.LogAudit(&adminID, "CREATE_USER", "users", clientIP, userAgent, fmt.Sprintf("Usuario creado: %s (%s)", email, userID))
	return s.GetMe(ctx, userID)
}

func (s *Service) UpdateUser(ctx context.Context, adminID, targetID string, req UpdateUserRequest, clientIP, userAgent string) (*User, error) {
	email, err := NormalizeAndValidateEmail(req.Email)
	if err != nil {
		return nil, err
	}

	// Check duplicate email for other users
	var count int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE lower(email) = ? AND id != ?`, email, targetID).Scan(&count)
	if count > 0 {
		return nil, errors.New("otro usuario ya está utilizando este correo electrónico")
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx, `
		UPDATE users
		SET first_name = ?, last_name = ?, email = ?, role_id = ?, status = ?,
		    identification = ?, company = ?, job_title = ?, updated_at = ?,
		    token_version = token_version + CASE WHEN role_id != ? OR status != ? THEN 1 ELSE 0 END
		WHERE id = ?
	`, strings.TrimSpace(req.FirstName), strings.TrimSpace(req.LastName), email, req.RoleID, req.Status, req.Identification, req.Company, req.JobTitle, now, req.RoleID, req.Status, targetID)

	if err != nil {
		return nil, fmt.Errorf("error actualizando usuario: %w", err)
	}

	s.LogAudit(&adminID, "UPDATE_USER", "users", clientIP, userAgent, fmt.Sprintf("Usuario modificado: %s", targetID))
	return s.GetMe(ctx, targetID)
}

func (s *Service) ToggleUserStatus(ctx context.Context, adminID, targetID, clientIP, userAgent string) (*User, error) {
	if adminID == targetID {
		return nil, errors.New("no puedes desactivar tu propia cuenta administrativa")
	}

	var currentStatus string
	err := s.db.QueryRowContext(ctx, `SELECT status FROM users WHERE id = ?`, targetID).Scan(&currentStatus)
	if err != nil {
		return nil, errors.New("usuario no encontrado")
	}

	newStatus := "inactive"
	if currentStatus == "inactive" {
		newStatus = "active"
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx, `UPDATE users SET status = ?, updated_at = ?, token_version = token_version + 1 WHERE id = ?`, newStatus, now, targetID)
	if err != nil {
		return nil, err
	}

	s.LogAudit(&adminID, "TOGGLE_USER_STATUS", "users", clientIP, userAgent, fmt.Sprintf("Estado de usuario %s cambiado a %s", targetID, newStatus))
	return s.GetMe(ctx, targetID)
}

func (s *Service) LogAudit(userID *string, action, resource, ip, ua, details string) {
	_, _ = s.db.Exec(`
		INSERT INTO audit_logs (user_id, action, resource, ip_address, user_agent, details)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, action, resource, ip, ua, details)
}

func sha256Hex(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}
