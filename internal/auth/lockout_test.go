package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dxstech-edu/internal/auth"
)

func TestPasswordPolicy(t *testing.T) {
	cases := []struct {
		name, pass string
		personal   []string
		ok         bool
	}{
		{"válida", "ClaveSegura#2026", nil, true},
		{"muy corta", "Abc123xyz", nil, false}, // 9 caracteres
		{"10 exactos", "Abcdefgh12", nil, true},
		{"sin mayúscula", "clavesegura2026", nil, false},
		{"sin minúscula", "CLAVESEGURA2026", nil, false},
		{"sin número", "ClaveSeguraAbc", nil, false},
		{"común", "Password123", nil, false},
		{"semilla conocida", "Admin1234*", nil, false},
		{"contiene el correo", "Maria.Lopez2026", []string{"maria.lopez@dxstech.edu"}, false},
		{"contiene el nombre", "CarlosFuerte123", []string{"x@y.co", "Carlos", "Mena"}, false},
		{"nombre corto no cuenta", "ClaveSegura2026", []string{"x@y.co", "Al"}, true},
		{"demasiado larga para bcrypt", strings.Repeat("Aa1", 30), nil, false},
	}
	for _, tc := range cases {
		err := auth.ValidatePasswordPolicy(tc.pass, tc.personal...)
		if (err == nil) != tc.ok {
			t.Errorf("%s: ok=%v, err=%v", tc.name, tc.ok, err)
		}
	}
}

func (e *sessionEnv) tryLogin(t *testing.T, email, password string) (int, map[string]any) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func TestAccountLocksAfterRepeatedFailures(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()
	const email, good = "estudiante@dxstech.edu", "Student1234*"

	// 4 fallos no bloquean; un acierto reinicia el contador (el límite por IP es de 15/min, por eso se acotan los intentos).
	for i := 0; i < 4; i++ {
		if code, _ := e.tryLogin(t, email, "incorrecta"); code != 401 {
			t.Fatalf("intento %d: esperado 401, obtuvo %d", i+1, code)
		}
	}
	if code, _ := e.tryLogin(t, email, good); code != 200 {
		t.Fatalf("el acceso correcto antes del límite debe funcionar, obtuvo %d", code)
	}
	// 5 fallos seguidos bloquean; incluso la contraseña correcta queda rechazada.
	for i := 0; i < 5; i++ {
		e.tryLogin(t, email, "incorrecta")
	}
	code, body := e.tryLogin(t, email, good)
	if code != http.StatusTooManyRequests || body["code"] != "ACCOUNT_LOCKED" {
		t.Fatalf("cuenta bloqueada: esperado 429/ACCOUNT_LOCKED, obtuvo %d %v", code, body)
	}

	// Un administrador puede desbloquearla.
	if _, err := e.svc.UnlockUser(context.Background(), "usr-admin-01", "usr-student-01", "", ""); err != nil {
		t.Fatal(err)
	}
	if code, _ := e.tryLogin(t, email, good); code != 200 {
		t.Fatalf("tras desbloquear debe poder ingresar, obtuvo %d", code)
	}
}

func TestLockoutDoesNotAffectOtherAccountsOrUnknownEmails(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()
	for i := 0; i < 8; i++ {
		if code, _ := e.tryLogin(t, "noexiste@dxstech.edu", "x"); code != 401 {
			t.Fatalf("correo inexistente debe responder siempre 401, obtuvo %d", code)
		}
	}
	for i := 0; i < 5; i++ {
		e.tryLogin(t, "estudiante@dxstech.edu", "incorrecta")
	}
	if code, _ := e.tryLogin(t, "admin@dxstech.edu", "Admin1234*"); code != 200 {
		t.Fatalf("bloquear a un usuario no debe afectar a otros, obtuvo %d", code)
	}
}

func TestPasswordChangeFailuresAlsoLock(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if err := e.svc.ChangePassword(ctx, "usr-student-01", "mala", "NuevaClave#2026", "", ""); err == nil {
			t.Fatal("debía fallar con la contraseña actual incorrecta")
		}
	}
	err := e.svc.ChangePassword(ctx, "usr-student-01", "Student1234*", "NuevaClave#2026", "", "")
	var locked *auth.LockedError
	if err == nil || !strings.Contains(err.Error(), "bloqueada") {
		t.Fatalf("tras 5 fallos el cambio de contraseña debe bloquearse, obtuvo %v (%T)", err, locked)
	}
}
