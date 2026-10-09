package main

import (
	"os"
	"testing"

	"dxstech-edu/internal/database"

	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	database.BcryptCost = bcrypt.MinCost
	os.Exit(m.Run())
}
