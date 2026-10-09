package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
)

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type User struct {
	ID                 string     `json:"id"`
	FirstName          string     `json:"firstName"`
	LastName           string     `json:"lastName"`
	Email              string     `json:"email"`
	PasswordHash       string     `json:"-"`
	RoleID             int        `json:"roleId"`
	Role               string     `json:"role"`
	Status             string     `json:"status"` // 'active', 'inactive'
	EmailVerified      bool       `json:"emailVerified"`
	MustChangePassword bool       `json:"mustChangePassword"`
	Identification     string     `json:"identification,omitempty"`
	Company            string     `json:"company,omitempty"`
	JobTitle           string     `json:"jobTitle,omitempty"`
	LastLogin          *time.Time `json:"lastLogin,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	PasswordChangedAt  *time.Time `json:"passwordChangedAt,omitempty"`
	TokenVersion       int        `json:"-"`
}

type LoginRequest struct {
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"rememberMe"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type CreateUserRequest struct {
	FirstName      string `json:"firstName" binding:"required"`
	LastName       string `json:"lastName" binding:"required"`
	Email          string `json:"email" binding:"required"`
	Password       string `json:"password" binding:"required"`
	RoleID         int    `json:"roleId" binding:"required"`
	Identification string `json:"identification"`
	Company        string `json:"company"`
	JobTitle       string `json:"jobTitle"`
}

type UpdateUserRequest struct {
	FirstName      string `json:"firstName" binding:"required"`
	LastName       string `json:"lastName" binding:"required"`
	Email          string `json:"email" binding:"required"`
	RoleID         int    `json:"roleId" binding:"required"`
	Status         string `json:"status" binding:"required"`
	Identification string `json:"identification"`
	Company        string `json:"company"`
	JobTitle       string `json:"jobTitle"`
}

func NormalizeAndValidateEmail(email string) (string, error) {
	clean := strings.ToLower(strings.TrimSpace(email))
	if clean == "" {
		return "", errors.New("el correo electrónico no puede estar vacío")
	}

	addr, err := mail.ParseAddress(clean)
	if err != nil {
		return "", errors.New("formato de correo electrónico inválido")
	}

	// Basic RFC regex sanity
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	if !re.MatchString(addr.Address) {
		return "", errors.New("formato de correo electrónico inválido")
	}

	return addr.Address, nil
}

const (
	minPasswordLength = 10
	// bcrypt solo procesa los primeros 72 bytes; más allá se truncaría en silencio.
	maxPasswordBytes = 72
)

// commonPasswords son contraseñas triviales que cumplirían las reglas de
// composición pero se adivinan de inmediato (se comparan en minúsculas).
var commonPasswords = map[string]bool{
	"password1": true, "password12": true, "password123": true, "password1234": true,
	"contraseña1": true, "contraseña123": true, "contrasena123": true, "qwerty12345": true,
	"qwertyuiop1": true, "admin12345": true, "admin1234*": true, "admin123456": true,
	"student1234*": true, "estudiante123": true, "welcome123": true, "bienvenido1": true,
	"letmein1234": true, "iloveyou123": true, "abc1234567": true, "1234567890a": true,
	"dxstech2026": true, "dxstech1234": true, "dxstech12345": true, "123456789a": true,
}

// ValidatePasswordPolicy valida la contraseña. Opcionalmente recibe datos
// personales del usuario (correo, nombre, apellido) que no pueden formar parte
// de ella.
func ValidatePasswordPolicy(password string, personal ...string) error {
	if len([]rune(password)) < minPasswordLength {
		return fmt.Errorf("la contraseña debe tener al menos %d caracteres", minPasswordLength)
	}
	if len(password) > maxPasswordBytes {
		return fmt.Errorf("la contraseña no puede superar los %d bytes", maxPasswordBytes)
	}

	var hasUpper, hasLower, hasNumber bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasNumber = true
		}
	}
	if !hasUpper || !hasLower || !hasNumber {
		return errors.New("la contraseña debe incluir al menos una letra mayúscula, una minúscula y un número")
	}

	lower := strings.ToLower(password)
	if commonPasswords[lower] {
		return errors.New("esa contraseña es demasiado común; elige una más difícil de adivinar")
	}
	for _, p := range personal {
		// Para un correo se compara solo la parte local (antes de la @).
		if at := strings.Index(p, "@"); at > 0 {
			p = p[:at]
		}
		p = strings.ToLower(strings.TrimSpace(p))
		if len([]rune(p)) >= 4 && strings.Contains(lower, p) {
			return errors.New("la contraseña no debe contener tu nombre ni tu correo")
		}
	}

	return nil
}
