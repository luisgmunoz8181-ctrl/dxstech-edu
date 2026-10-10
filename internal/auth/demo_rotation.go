package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"dxstech-edu/internal/audit"
	"dxstech-edu/internal/database"

	"golang.org/x/crypto/bcrypt"
)

// Resultados posibles para cada cuenta demo.
const (
	DemoNotFound    = "no existe"
	DemoAlreadySafe = "ya no usa la contraseña demo"
	DemoInactive    = "ya estaba desactivada"
	DemoRotated     = "contraseña rotada"
	DemoDeactivated = "desactivada"
	DemoWouldRotate = "se rotaría"
	DemoWouldDeact  = "se desactivaría"
)

// DemoRotationOptions controla RotateDemoUsers.
type DemoRotationOptions struct {
	// Apply=false es una simulación: informa qué cuentas se tocarían sin escribir nada.
	Apply bool
	// Deactivate desactiva las cuentas demo en lugar de darles una contraseña nueva. Una
	// cuenta SUPERADMIN solo se desactiva si queda otro SUPERADMIN activo; si no, se rota,
	// para no dejar la plataforma sin administrador.
	Deactivate bool
}

// DemoRotationResult es el resultado para una cuenta demo.
type DemoRotationResult struct {
	Email    string
	Role     string
	Outcome  string
	Password string // solo con Apply y rotación: se muestra una única vez
}

// RotateDemoUsers neutraliza las cuentas de demostración que todavía usan su contraseña
// pública conocida. Solo toca una cuenta si su contraseña SIGUE siendo la demo: una
// cuenta cuya contraseña ya se cambió (o que alguien reutilizó) no se modifica.
//
// Rotar = contraseña aleatoria que cumple la política, cambio obligatorio en el primer
// ingreso y todas las sesiones previas revocadas. La operación es atómica.
func RotateDemoUsers(ctx context.Context, db *database.DB, opts DemoRotationOptions) ([]DemoRotationResult, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	type account struct {
		id, hash, role string
		active         bool
	}
	found := map[string]*account{}
	results := make([]DemoRotationResult, 0, len(database.DemoUsers))

	for _, demo := range database.DemoUsers {
		var a account
		var status string
		err := tx.QueryRowContext(ctx, `
			SELECT u.id, u.password_hash, r.name, u.status
			FROM users u JOIN roles r ON r.id = u.role_id
			WHERE lower(u.email) = ?`, strings.ToLower(demo.Email)).Scan(&a.id, &a.hash, &a.role, &status)
		if errors.Is(err, sql.ErrNoRows) {
			results = append(results, DemoRotationResult{Email: demo.Email, Outcome: DemoNotFound})
			continue
		}
		if err != nil {
			return nil, err
		}
		a.active = status == "active"
		res := DemoRotationResult{Email: demo.Email, Role: a.role}

		if bcrypt.CompareHashAndPassword([]byte(a.hash), []byte(demo.Password)) != nil {
			res.Outcome = DemoAlreadySafe
		} else if !a.active {
			// Una cuenta desactivada no puede iniciar sesión: no se reactiva ni se toca.
			res.Outcome = DemoInactive
		} else {
			found[demo.Email] = &a
			res.Outcome = "pendiente"
		}
		results = append(results, res)
	}

	// ¿Queda algún SUPERADMIN activo si se desactivan las cuentas demo? Si no, se rotan.
	deactivating := map[string]bool{}
	for _, a := range found {
		deactivating[a.id] = opts.Deactivate && a.role != "SUPERADMIN"
	}
	if opts.Deactivate {
		var others int
		err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM users u JOIN roles r ON r.id = u.role_id
			WHERE r.name = 'SUPERADMIN' AND u.status = 'active'
			  AND lower(u.email) NOT IN (`+demoEmailPlaceholders()+`)`, demoEmailArgs()...).Scan(&others)
		if err != nil {
			return nil, err
		}
		if others > 0 { // hay un SUPERADMIN propio: el demo también puede desactivarse
			for _, a := range found {
				if a.role == "SUPERADMIN" {
					deactivating[a.id] = true
				}
			}
		}
	}

	for i := range results {
		a := found[results[i].Email]
		if a == nil {
			continue
		}
		deact := deactivating[a.id]
		if !opts.Apply {
			results[i].Outcome = DemoWouldRotate
			if deact {
				results[i].Outcome = DemoWouldDeact
			}
			continue
		}

		if deact {
			if _, err := tx.ExecContext(ctx, `
				UPDATE users SET status = 'inactive', token_version = token_version + 1, updated_at = CURRENT_TIMESTAMP
				WHERE id = ?`, a.id); err != nil {
				return nil, err
			}
			results[i].Outcome = DemoDeactivated
			continue
		}

		pass, err := generateStrongPassword()
		if err != nil {
			return nil, err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), database.BcryptCost)
		if err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE users SET password_hash = ?, must_change_password = 1,
			    failed_login_attempts = 0, locked_until = NULL,
			    password_changed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP,
			    token_version = token_version + 1
			WHERE id = ?`, string(hash), a.id); err != nil {
			return nil, err
		}
		results[i].Outcome = DemoRotated
		results[i].Password = pass
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if opts.Apply {
		var touched []string
		for _, r := range results {
			if r.Outcome == DemoRotated || r.Outcome == DemoDeactivated {
				touched = append(touched, r.Email+": "+r.Outcome)
			}
		}
		if len(touched) > 0 { // sin contraseñas ni datos sensibles en la auditoría
			audit.Log(db, audit.Entry{Action: audit.DemoUsersRotated, Resource: "users", Details: "Cuentas demo neutralizadas desde la línea de comandos: " + strings.Join(touched, "; ")})
		}
	}
	return results, nil
}

func demoEmailPlaceholders() string {
	return strings.TrimSuffix(strings.Repeat("?,", len(database.DemoUsers)), ",")
}

func demoEmailArgs() []any {
	args := make([]any, len(database.DemoUsers))
	for i, d := range database.DemoUsers {
		args[i] = strings.ToLower(d.Email)
	}
	return args
}

const (
	pwUpper  = "ABCDEFGHJKLMNPQRSTUVWXYZ" // sin I ni O (ambiguas al copiarla a mano)
	pwLower  = "abcdefghijkmnpqrstuvwxyz" // sin l ni o
	pwDigit  = "23456789"
	pwLength = 20
)

// generateStrongPassword crea una contraseña aleatoria (criptográficamente segura) que
// cumple la política: mayúscula, minúscula y número garantizados, sin caracteres ambiguos.
func generateStrongPassword() (string, error) {
	all := pwUpper + pwLower + pwDigit
	pick := func(set string) (byte, error) {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
		if err != nil {
			return 0, err
		}
		return set[n.Int64()], nil
	}

	for attempt := 0; attempt < 10; attempt++ {
		b := make([]byte, pwLength)
		for i := range b {
			c, err := pick(all)
			if err != nil {
				return "", err
			}
			b[i] = c
		}
		// Se fuerzan las tres clases en posiciones aleatorias.
		for _, set := range []string{pwUpper, pwLower, pwDigit} {
			pos, err := rand.Int(rand.Reader, big.NewInt(pwLength))
			if err != nil {
				return "", err
			}
			c, err := pick(set)
			if err != nil {
				return "", err
			}
			b[pos.Int64()] = c
		}
		if p := string(b); ValidatePasswordPolicy(p) == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no se pudo generar una contraseña válida")
}
