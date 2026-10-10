package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"dxstech-edu/internal/config"
	"dxstech-edu/internal/database"
)

func rotateEnv(t *testing.T, seedDemo bool) *config.Config {
	t.Helper()
	dir, _ := os.MkdirTemp("", "dxstech_rotate_*")
	t.Cleanup(func() { os.RemoveAll(dir) })
	db, err := database.InitDB(dir, seedDemo)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	return &config.Config{DataDir: dir}
}

func TestRotateDemoUsersCommand(t *testing.T) {
	cfg := rotateEnv(t, true)

	// Sin -yes: simulación, código de salida 2 y no imprime contraseñas.
	var out bytes.Buffer
	if code := runRotateDemoUsers(cfg, false, false, &out); code != 2 {
		t.Fatalf("simulación: código %d, esperado 2\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "se rotaría") || strings.Contains(out.String(), "Contraseñas nuevas") {
		t.Errorf("la simulación debe listar lo pendiente sin generar contraseñas:\n%s", out.String())
	}

	// Con -yes: rota, imprime una contraseña por cuenta y sale con 0.
	out.Reset()
	if code := runRotateDemoUsers(cfg, true, false, &out); code != 0 {
		t.Fatalf("aplicar: código %d\n%s", code, out.String())
	}
	text := out.String()
	for _, email := range []string{"superadmin@dxstech.edu", "admin@dxstech.edu", "estudiante@dxstech.edu"} {
		if !strings.Contains(text, email) {
			t.Errorf("falta %s en la salida", email)
		}
	}
	if !strings.Contains(text, "UNA sola vez") || strings.Count(text, "contraseña rotada") != 3 {
		t.Errorf("salida inesperada:\n%s", text)
	}

	// Segunda ejecución: nada que hacer y código 0.
	out.Reset()
	if code := runRotateDemoUsers(cfg, false, false, &out); code != 0 || !strings.Contains(out.String(), "Nada que hacer") {
		t.Errorf("segunda ejecución: código %d\n%s", code, out.String())
	}
}

func TestRotateDemoUsersCommandErrors(t *testing.T) {
	// Sin base de datos: error claro y sin crear ninguna.
	dir, _ := os.MkdirTemp("", "dxstech_rotate_empty_*")
	defer os.RemoveAll(dir)
	var out bytes.Buffer
	if code := runRotateDemoUsers(&config.Config{DataDir: dir}, true, false, &out); code != 1 {
		t.Fatalf("sin base: código %d, esperado 1", code)
	}
	if !strings.Contains(out.String(), "DATA_DIR") {
		t.Errorf("el error debe orientar sobre DATA_DIR:\n%s", out.String())
	}
	if _, err := os.Stat(dir + "/dxstech.db"); err == nil {
		t.Error("la herramienta no debe crear una base de datos nueva")
	}

	// Base sin cuentas demo (producción normal): nada que hacer.
	cfg := rotateEnv(t, false)
	out.Reset()
	if code := runRotateDemoUsers(cfg, true, false, &out); code != 0 || !strings.Contains(out.String(), "no existe") {
		t.Errorf("sin cuentas demo: código %d\n%s", code, out.String())
	}
}
