const { test, expect, ACCOUNTS } = require('./fixtures');

test.describe('Administración de cursos (regresión: Modal.close)', () => {
  test('crear un curso desde la interfaz lo guarda, cierra el modal y lo lista', async ({ page }) => {
    await page.goto('/#login');
    await page.fill('#login-email', ACCOUNTS.admin.email);
    await page.fill('#login-password', ACCOUNTS.admin.password);
    await page.click('#login-submit-btn');
    await expect(page.locator('#view-title')).toContainText('Dashboard');
    await page.click('.nav-btn[data-nav="courses"]');
    await expect(page.locator('#view-title')).toContainText('Cursos');

    const code = `E2E-${Date.now()}`;
    await page.click('#create-course-btn');
    await page.fill('#course-title', 'Curso creado por e2e');
    await page.fill('#course-code', code);
    await page.fill('#course-short-desc', 'Descripción corta');
    await page.click('#modal-confirm-btn');

    // El modal se cierra y no aparece "... is not a function".
    await expect(page.locator('#modal-container')).toBeHidden();
    await expect(page.getByText('is not a function')).toHaveCount(0);
    await expect(page.getByText(code).first()).toBeVisible();
    expect(page.problems).toEqual([]);
  });
});
