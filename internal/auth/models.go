package auth

import (
	"errors"
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

func ValidatePasswordPolicy(password string) error {
	if len(password) < 8 {
		return errors.New("la contraseña debe tener al menos 8 caracteres")
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

	return nil
}
