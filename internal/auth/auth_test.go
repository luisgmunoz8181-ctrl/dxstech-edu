package auth_test

import (
	"context"
	"os"
	"testing"

	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"
)

func TestAuthFlow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dxstech-auth-test-*")
	if err != nil {
		t.Fatalf("Error creando tmpDir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := database.InitDB(tmpDir, true)
	if err != nil {
		t.Fatalf("Error inicializando DB de test: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		AppEnv:    "development",
		JWTSecret: "test-secret-key-12345",
	}

	svc := auth.NewService(db, cfg)
	ctx := context.Background()

	// 1. Test Superadmin login
	loginReq := auth.LoginRequest{
		Email:      "superadmin@dxstech.edu",
		Password:   "Admin1234*",
		RememberMe: false,
	}

	resp, err := svc.Login(ctx, loginReq, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Fallo en login de Superadmin: %v", err)
	}

	if resp.User.Role != "SUPERADMIN" {
		t.Errorf("Rol esperado SUPERADMIN, recibido %s", resp.User.Role)
	}

	if resp.Token == "" {
		t.Error("Token JWT no debe estar vacío")
	}

	// 2. Test GetMe
	me, err := svc.GetMe(ctx, resp.User.ID)
	if err != nil {
		t.Fatalf("Fallo en GetMe: %v", err)
	}
	if me.Email != "superadmin@dxstech.edu" {
		t.Errorf("Email esperado superadmin@dxstech.edu, recibido %s", me.Email)
	}

	// 3. Test Invalid password
	badReq := auth.LoginRequest{
		Email:    "superadmin@dxstech.edu",
		Password: "WrongPassword999*",
	}
	_, err = svc.Login(ctx, badReq, "127.0.0.1", "test-agent")
	if err == nil {
		t.Error("Login con contraseña incorrecta debió fallar")
	}

	// 4. Test Student login
	studentReq := auth.LoginRequest{
		Email:    "estudiante@dxstech.edu",
		Password: "Student1234*",
	}
	studentResp, err := svc.Login(ctx, studentReq, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Fallo en login de Estudiante: %v", err)
	}
	if studentResp.User.Role != "ESTUDIANTE" {
		t.Errorf("Rol esperado ESTUDIANTE, recibido %s", studentResp.User.Role)
	}

	// 5. Test Password Reset request
	rawToken, err := svc.ForgotPassword(ctx, "estudiante@dxstech.edu", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Fallo en ForgotPassword: %v", err)
	}
	if rawToken == "" {
		t.Fatal("Token de recuperación esperado")
	}

	// 6. Test Reset Password
	err = svc.ResetPassword(ctx, rawToken, "NewPass2026*", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Fallo en ResetPassword: %v", err)
	}

	// 7. Verify login with new password
	newLoginReq := auth.LoginRequest{
		Email:    "estudiante@dxstech.edu",
		Password: "NewPass2026*",
	}
	_, err = svc.Login(ctx, newLoginReq, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Fallo login con nueva contraseña: %v", err)
	}
}
