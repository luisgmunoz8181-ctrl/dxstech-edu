package auth_test

import (
	"os"
	"testing"

	"dxstech-edu/internal/database"

	"golang.org/x/crypto/bcrypt"
)

// Los tests crean muchos hashes; el costo mínimo de bcrypt los acelera sin
// cambiar la lógica (producción usa bcrypt.DefaultCost).
func TestMain(m *testing.M) {
	database.BcryptCost = bcrypt.MinCost
	os.Exit(m.Run())
}
