const { test, expect, loginApi, loginUi } = require('./fixtures');

test('un administrador ve las cuentas bloqueadas y las desbloquea desde Usuarios', async ({ page, browser, baseURL }) => {
  const admin = await browser.newContext({ baseURL });
  await loginApi(admin.request, 'admin');
  const email = `bloqueado.${Date.now()}@dxstech.edu`;
  await admin.request.post('/api/auth/users', { data: { firstName: 'Cuenta', lastName: 'Bloqueable', email, password: 'ClaveInicial#2026', roleId: 3 } });
  await admin.close();

  // 5 fallos seguidos bloquean la cuenta.
  await page.goto('/');
  for (let i = 0; i < 5; i++) {
    await page.evaluate(async (e) => fetch('/api/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email: e, password: 'mala' }) }), email);
  }

  await loginUi(page, 'admin');
  await page.click('.nav-btn[data-nav="users"]');
  await page.fill('#user-search-input', email);
  const row = page.locator('tbody tr').filter({ hasText: email });
  await expect(row).toBeVisible();
  await expect(row).toContainText('Bloqueada');

  await row.locator('[data-unlock-id]').click();
  await expect(page.getByText('Cuenta desbloqueada').first()).toBeVisible();
  await expect(page.locator('tbody tr').filter({ hasText: email })).not.toContainText('Bloqueada');
  await expect(page.locator('tbody tr').filter({ hasText: email }).locator('[data-unlock-id]')).toHaveCount(0);
  expect(page.problems).toEqual([]);
});
