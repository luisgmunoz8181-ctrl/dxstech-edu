const { test, expect, ACCOUNTS, loginApi } = require('./fixtures');

test.describe('Contraseña temporal y bloqueo de cuenta', () => {
  test('un usuario creado por un admin debe cambiar su contraseña antes de usar la plataforma', async ({ page, browser, baseURL }) => {
    const admin = await browser.newContext({ baseURL });
    await loginApi(admin.request, 'admin');
    const email = `usuario.temporal.${Date.now()}@dxstech.edu`;
    const created = await admin.request.post('/api/auth/users', {
      data: { firstName: 'Nuevo', lastName: 'Alumno', email, password: 'ClaveInicial#2026', roleId: 3 },
    });
    expect(created.status()).toBe(201);

    await page.goto('/#login');
    await page.fill('#login-email', email);
    await page.fill('#login-password', 'ClaveInicial#2026');
    await page.click('#login-submit-btn');

    // El router lo lleva al perfil y bloquea el resto de vistas.
    await expect(page.locator('#view-title')).toContainText(/perfil/i);
    await page.evaluate(() => window.router.navigate('courses'));
    await expect(page.locator('#view-title')).toContainText(/perfil/i);
    // El servidor también lo bloquea.
    expect(await page.evaluate(async () => (await fetch('/api/enrollments/my-courses')).status)).toBe(403);

    // Una contraseña débil se rechaza; una válida desbloquea la cuenta.
    const change = (current, next) => page.evaluate(async ([c, n]) => {
      const r = await fetch('/api/auth/change-password', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ currentPassword: c, newPassword: n }) });
      return r.status;
    }, [current, next]);
    expect(await change('ClaveInicial#2026', 'corta1A')).toBe(400);
    expect(await change('ClaveInicial#2026', 'NuevaClaveSegura#77')).toBe(200);
    expect(await page.evaluate(async () => (await fetch('/api/enrollments/my-courses')).status)).toBe(200);
    await admin.close();
  });

  test('5 contraseñas incorrectas bloquean la cuenta y el mensaje es claro', async ({ page, browser, baseURL }) => {
    const admin = await browser.newContext({ baseURL });
    await loginApi(admin.request, 'admin');
    const email = `lock${Date.now()}@dxstech.edu`;
    await admin.request.post('/api/auth/users', {
      data: { firstName: 'Bloqueo', lastName: 'Prueba', email, password: 'ClaveInicial#2026', roleId: 3 },
    });

    const attempt = (password) => page.evaluate(async ([e, p]) => {
      const r = await fetch('/api/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email: e, password: p }) });
      return { status: r.status, body: await r.json() };
    }, [email, password]);

    await page.goto('/');
    for (let i = 0; i < 5; i++) expect((await attempt('mala')).status).toBe(401);
    const locked = await attempt('ClaveInicial#2026'); // ni la correcta entra
    expect(locked.status).toBe(429);
    expect(locked.body.code).toBe('ACCOUNT_LOCKED');

    // En la interfaz se ve el aviso.
    await page.goto('/#login');
    await page.fill('#login-email', email);
    await page.fill('#login-password', 'ClaveInicial#2026');
    await page.click('#login-submit-btn');
    await expect(page.getByText(/bloqueada temporalmente|demasiados intentos/i).first()).toBeVisible();

    // Un administrador la desbloquea.
    const users = await (await admin.request.get(`/api/auth/users?search=${encodeURIComponent(email)}`)).json();
    const target = (users.users || users).find((u) => u.email === email);
    expect((await admin.request.post(`/api/auth/users/${target.id}/unlock`)).status()).toBe(200);
    expect((await attempt('ClaveInicial#2026')).status).toBe(200);
    await admin.close();
  });
});
