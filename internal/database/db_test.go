package database_test

import (
	"os"
	"strings"
	"testing"

	"dxstech-edu/internal/database"

	"golang.org/x/crypto/bcrypt"
)

func TestInitDBWithoutDemoCreatesSingleBootstrapSuperadmin(t *testing.T) {
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", "")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "")
	dir, _ := os.MkdirTemp("", "dxstech_db_boot_*")
	defer os.RemoveAll(dir)

	db, err := database.InitDB(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var n int
	_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&n)
	if n != 1 {
		t.Fatalf("se esperaba 1 usuario inicial, hay %d", n)
	}
	var hash string
	var mustChange int
	_ = db.QueryRow("SELECT password_hash, must_change_password FROM users").Scan(&hash, &mustChange)
	if mustChange != 1 {
		t.Errorf("el superadmin inicial debe forzar cambio de contraseña")
	}
	for _, known := range []string{"Admin1234*", "Student1234*", "dxstech2026"} {
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(known)) == nil {
			t.Errorf("el superadmin inicial no debe usar la contraseña conocida %q", known)
		}
	}
	_ = db.QueryRow("SELECT COUNT(*) FROM courses").Scan(&n)
	if n != 0 {
		t.Errorf("sin seedDemo no debe haber cursos de demostración, hay %d", n)
	}
}

func TestBootstrapSuperadminUsesEnvCredentials(t *testing.T) {
	t.Setenv("BOOTSTRAP_ADMIN_EMAIL", " Root@Example.com ")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "S3cret-Passw0rd!")
	dir, _ := os.MkdirTemp("", "dxstech_db_boot_env_*")
	defer os.RemoveAll(dir)

	db, err := database.InitDB(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var email, hash string
	_ = db.QueryRow("SELECT email, password_hash FROM users").Scan(&email, &hash)
	if email != strings.ToLower("root@example.com") {
		t.Errorf("email inesperado: %q", email)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte("S3cret-Passw0rd!")) != nil {
		t.Errorf("la contraseña del entorno no se aplicó")
	}
}
