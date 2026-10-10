package auth_test

import (
	"context"
	"strings"
	"testing"

	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/database"
)

func byEmail(results []auth.DemoRotationResult) map[string]auth.DemoRotationResult {
	m := map[string]auth.DemoRotationResult{}
	for _, r := range results {
		m[r.Email] = r
	}
	return m
}

func TestRotateDemoUsersSimulationChangesNothing(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()

	res, err := auth.RotateDemoUsers(context.Background(), e.db, auth.DemoRotationOptions{Apply: false})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range res {
		if r.Outcome != auth.DemoWouldRotate || r.Password != "" {
			t.Errorf("simulación: %+v", r)
		}
	}
	// Las credenciales demo siguen funcionando: no se tocó nada.
	if code, _ := e.tryLogin(t, "admin@dxstech.edu", "Admin1234*"); code != 200 {
		t.Errorf("tras una simulación el login demo debe seguir funcionando, obtuvo %d", code)
	}
}

func TestRotateDemoUsersReplacesKnownPasswords(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()
	ctx := context.Background()

	// Una sesión abierta antes de la rotación debe quedar revocada.
	oldToken := e.login(t, "admin@dxstech.edu", "Admin1234*")
	if c, _ := e.do("GET", "/api/ping", oldToken); c != 200 {
		t.Fatalf("la sesión previa debe funcionar antes de rotar, obtuvo %d", c)
	}

	res, err := auth.RotateDemoUsers(ctx, e.db, auth.DemoRotationOptions{Apply: true})
	if err != nil {
		t.Fatal(err)
	}
	got := byEmail(res)
	seen := map[string]bool{}
	for _, d := range database.DemoUsers {
		r := got[d.Email]
		if r.Outcome != auth.DemoRotated {
			t.Fatalf("%s: %+v", d.Email, r)
		}
		if auth.ValidatePasswordPolicy(r.Password) != nil || len(r.Password) < 20 {
			t.Errorf("%s: la contraseña generada no cumple la política: %q", d.Email, r.Password)
		}
		if strings.ContainsAny(r.Password, "IlOo01") {
			t.Errorf("%s: no debe contener caracteres ambiguos: %q", d.Email, r.Password)
		}
		if seen[r.Password] {
			t.Errorf("las contraseñas deben ser distintas por cuenta")
		}
		seen[r.Password] = true

		// La contraseña pública ya no entra; la nueva sí.
		if code, _ := e.tryLogin(t, d.Email, d.Password); code != 401 {
			t.Errorf("%s: la contraseña demo debe dejar de funcionar, obtuvo %d", d.Email, code)
		}
		if code, _ := e.tryLogin(t, d.Email, r.Password); code != 200 {
			t.Errorf("%s: la contraseña nueva debe funcionar, obtuvo %d", d.Email, code)
		}
	}

	if c, _ := e.do("GET", "/api/ping", oldToken); c != 401 {
		t.Errorf("la sesión anterior a la rotación debe quedar revocada, obtuvo %d", c)
	}
	// Cambio obligatorio en el primer ingreso.
	newTok := e.login(t, "admin@dxstech.edu", got["admin@dxstech.edu"].Password)
	if c, body := e.do("GET", "/api/ping", newTok); c != 403 || !strings.Contains(body, "PASSWORD_CHANGE_REQUIRED") {
		t.Errorf("tras rotar debe exigirse cambiar la contraseña: %d %s", c, body)
	}

	// Queda constancia en la auditoría, sin contraseñas.
	var details string
	if err := e.db.QueryRow(`SELECT details FROM audit_logs WHERE action = 'DEMO_USERS_ROTATED'`).Scan(&details); err != nil {
		t.Fatalf("falta el registro de auditoría: %v", err)
	}
	for _, r := range res {
		if r.Password != "" && strings.Contains(details, r.Password) {
			t.Error("la auditoría no debe contener contraseñas")
		}
	}
}

func TestRotateDemoUsersIsIdempotentAndRespectsChangedPasswords(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()
	ctx := context.Background()

	// El administrador demo ya cambió su contraseña: no debe tocarse.
	if err := e.svc.ChangePassword(ctx, "usr-admin-01", "Admin1234*", "MiClavePropia#2026", "", ""); err != nil {
		t.Fatal(err)
	}

	first, err := auth.RotateDemoUsers(ctx, e.db, auth.DemoRotationOptions{Apply: true})
	if err != nil {
		t.Fatal(err)
	}
	got := byEmail(first)
	if got["admin@dxstech.edu"].Outcome != auth.DemoAlreadySafe || got["admin@dxstech.edu"].Password != "" {
		t.Errorf("una cuenta con contraseña propia no debe tocarse: %+v", got["admin@dxstech.edu"])
	}
	if code, _ := e.tryLogin(t, "admin@dxstech.edu", "MiClavePropia#2026"); code != 200 {
		t.Errorf("la contraseña propia debe seguir funcionando, obtuvo %d", code)
	}
	if got["superadmin@dxstech.edu"].Outcome != auth.DemoRotated || got["estudiante@dxstech.edu"].Outcome != auth.DemoRotated {
		t.Errorf("las otras dos cuentas sí debían rotarse: %+v", got)
	}

	// Segunda ejecución: no queda nada por hacer y no se escribe auditoría nueva.
	second, err := auth.RotateDemoUsers(ctx, e.db, auth.DemoRotationOptions{Apply: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range second {
		if r.Outcome != auth.DemoAlreadySafe || r.Password != "" {
			t.Errorf("segunda ejecución: %+v", r)
		}
	}
	var n int
	_ = e.db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action = 'DEMO_USERS_ROTATED'`).Scan(&n)
	if n != 1 {
		t.Errorf("solo la primera ejecución debe auditarse, hay %d registros", n)
	}
}

func TestRotateDemoUsersMissingAccountsAndInactiveOnes(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()
	ctx := context.Background()

	if _, err := e.db.Exec(`DELETE FROM users WHERE id = 'usr-student-01'`); err != nil {
		t.Fatal(err)
	}
	// Una cuenta demo desactivada con la contraseña pública no se reactiva ni se modifica.
	if _, err := e.svc.ToggleUserStatus(ctx, "usr-superadmin-01", "usr-admin-01", "", ""); err != nil {
		t.Fatal(err)
	}

	res, err := auth.RotateDemoUsers(ctx, e.db, auth.DemoRotationOptions{Apply: true})
	if err != nil {
		t.Fatal(err)
	}
	got := byEmail(res)
	if got["estudiante@dxstech.edu"].Outcome != auth.DemoNotFound {
		t.Errorf("cuenta inexistente: %+v", got["estudiante@dxstech.edu"])
	}
	if got["admin@dxstech.edu"].Outcome != auth.DemoInactive {
		t.Errorf("cuenta ya desactivada: %+v", got["admin@dxstech.edu"])
	}
	var status string
	_ = e.db.QueryRow(`SELECT status FROM users WHERE id = 'usr-admin-01'`).Scan(&status)
	if status != "inactive" {
		t.Errorf("no debe reactivarse una cuenta desactivada, status=%s", status)
	}
}

func TestRotateDemoUsersDeactivateKeepsAnAdministrator(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()
	ctx := context.Background()

	// Sin otro SUPERADMIN: admin y estudiante se desactivan, pero el superadmin se rota
	// (si no, la plataforma quedaría sin administrador).
	res, err := auth.RotateDemoUsers(ctx, e.db, auth.DemoRotationOptions{Apply: true, Deactivate: true})
	if err != nil {
		t.Fatal(err)
	}
	got := byEmail(res)
	if got["admin@dxstech.edu"].Outcome != auth.DemoDeactivated || got["estudiante@dxstech.edu"].Outcome != auth.DemoDeactivated {
		t.Errorf("admin y estudiante debían desactivarse: %+v", got)
	}
	if got["superadmin@dxstech.edu"].Outcome != auth.DemoRotated || got["superadmin@dxstech.edu"].Password == "" {
		t.Errorf("el único superadmin debe rotarse, no desactivarse: %+v", got["superadmin@dxstech.edu"])
	}
	if code, _ := e.tryLogin(t, "admin@dxstech.edu", "Admin1234*"); code == 200 {
		t.Error("una cuenta desactivada no debe poder ingresar")
	}
	if code, _ := e.tryLogin(t, "superadmin@dxstech.edu", got["superadmin@dxstech.edu"].Password); code != 200 {
		t.Errorf("el superadmin rotado debe poder ingresar, obtuvo %d", code)
	}
}

func TestRotateDemoUsersDeactivatesSuperadminWhenAnotherExists(t *testing.T) {
	e, cleanup := newSessionEnv(t)
	defer cleanup()
	ctx := context.Background()

	if _, err := e.svc.CreateUser(ctx, "usr-superadmin-01", auth.CreateUserRequest{
		FirstName: "Ana", LastName: "Directora", Email: "ana@colegio.edu", Password: "UnaClaveMuyLarga#77", RoleID: 1,
	}, "", ""); err != nil {
		t.Fatal(err)
	}
	res, err := auth.RotateDemoUsers(ctx, e.db, auth.DemoRotationOptions{Apply: true, Deactivate: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range res {
		if r.Outcome != auth.DemoDeactivated {
			t.Errorf("con otro SUPERADMIN propio todas las cuentas demo pueden desactivarse: %+v", r)
		}
	}
	var active int
	_ = e.db.QueryRow(`SELECT COUNT(*) FROM users u JOIN roles r ON r.id = u.role_id WHERE r.name = 'SUPERADMIN' AND u.status = 'active'`).Scan(&active)
	if active != 1 {
		t.Errorf("debe quedar exactamente el SUPERADMIN propio activo, hay %d", active)
	}
}
