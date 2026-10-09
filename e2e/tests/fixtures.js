const base = require('@playwright/test');

const ACCOUNTS = {
  student: { email: 'estudiante@dxstech.edu', password: 'Student1234*' },
  admin: { email: 'admin@dxstech.edu', password: 'Admin1234*' },
  superadmin: { email: 'superadmin@dxstech.edu', password: 'Admin1234*' },
};

const XSS = '<img src=x onerror="window.__xss=(window.__xss||0)+1">';

/** Inicia sesión por API (la cookie HttpOnly queda en el contexto). */
async function loginApi(request, who) {
  const acc = typeof who === 'string' ? ACCOUNTS[who] : who;
  const res = await request.post('/api/auth/login', { data: acc });
  if (!res.ok()) throw new Error(`login ${acc.email}: ${res.status()} ${await res.text()}`);
  return res.json();
}

const test = base.test.extend({
  // Sin dependencias externas (Google Fonts, imágenes de Unsplash): pruebas deterministas y rápidas.
  context: async ({ context }, use) => {
    await context.route(/^https?:\/\/(?!localhost)/, (route) => route.abort());
    await use(context);
  },
  // Registra errores de JavaScript y diálogos inesperados de cada página.
  page: async ({ page }, use) => {
    const problems = [];
    page.on('pageerror', (e) => problems.push(`pageerror: ${e.message}`));
    page.on('dialog', (d) => { problems.push(`dialog: ${d.message()}`); d.dismiss(); });
    page.problems = problems;
    await use(page);
  },
});

/** Inicia sesión por la interfaz y espera a que cargue la vista inicial. */
async function loginUi(page, who) {
  const acc = typeof who === 'string' ? ACCOUNTS[who] : who;
  await page.goto('/#login');
  await page.fill('#login-email', acc.email);
  await page.fill('#login-password', acc.password);
  await page.click('#login-submit-btn');
  await base.expect(page.locator('#login-form')).toBeHidden();
}

module.exports = { test, expect: base.expect, ACCOUNTS, XSS, loginApi, loginUi };
