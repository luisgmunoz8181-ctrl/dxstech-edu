const { test, expect, ACCOUNTS, loginApi } = require('./fixtures');

async function loginViaUi(page, { email, password }) {
  await page.goto('/#login');
  await page.fill('#login-email', email);
  await page.fill('#login-password', password);
  await page.click('#login-submit-btn');
}

test.describe('Autenticación y permisos de la interfaz', () => {
  test('un estudiante inicia sesión y no ve los módulos administrativos', async ({ page }) => {
    await loginViaUi(page, ACCOUNTS.student);
    await expect(page.locator('#view-title')).toContainText('Cursos');
    await expect(page.getByText('Carlos Estudiante').first()).toBeVisible();

    for (const nav of ['dashboard', 'users', 'quizzes', 'whatsapp']) {
      await expect(page.locator(`.nav-btn[data-nav="${nav}"]`)).toBeHidden();
    }
    await expect(page.locator('.nav-btn[data-nav="courses"]')).toBeVisible();
    expect(page.problems).toEqual([]);
  });

  test('navegar por URL a un módulo administrativo redirige a Cursos', async ({ page }) => {
    await loginViaUi(page, ACCOUNTS.student);
    await expect(page.locator('#view-title')).toContainText('Cursos');
    await page.evaluate(() => window.router.navigate('users'));
    await expect(page.locator('#view-title')).toContainText('Cursos');
  });

  test('un administrador ve todos los módulos', async ({ page }) => {
    await loginViaUi(page, ACCOUNTS.admin);
    await expect(page.locator('#view-title')).toContainText('Dashboard');
    for (const nav of ['dashboard', 'users', 'quizzes', 'whatsapp']) {
      await expect(page.locator(`.nav-btn[data-nav="${nav}"]`)).toBeVisible();
    }
  });

  test('credenciales incorrectas muestran un error y no inician sesión', async ({ page }) => {
    await loginViaUi(page, { email: ACCOUNTS.student.email, password: 'incorrecta' });
    await expect(page.getByText(/incorrectos/i).first()).toBeVisible();
    await expect(page.locator('#login-form')).toBeVisible();
  });

  test('la API rechaza endpoints de administración a un estudiante', async ({ request }) => {
    await loginApi(request, 'student');
    for (const path of ['/api/admin/audit', '/api/admin/metrics/overview', '/api/quizzes', '/api/certificates/issued']) {
      expect((await request.get(path)).status(), path).toBe(403);
    }
  });

  test('cerrar sesión revoca el acceso', async ({ page, request }) => {
    await loginViaUi(page, ACCOUNTS.student);
    await expect(page.locator('#view-title')).toContainText('Cursos');
    await page.evaluate(() => fetch('/api/auth/logout', { method: 'POST' }));
    const me = await page.evaluate(async () => (await fetch('/api/auth/me')).status);
    expect(me).toBe(401);
  });
});
