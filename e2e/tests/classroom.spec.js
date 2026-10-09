const { test, expect, ACCOUNTS } = require('./fixtures');

async function loginUi(page, who) {
  await page.goto('/#login');
  await page.fill('#login-email', ACCOUNTS[who].email);
  await page.fill('#login-password', ACCOUNTS[who].password);
  await page.click('#login-submit-btn');
  await expect(page.locator('#view-title')).not.toContainText('Iniciar');
}

test.describe('Aula virtual', () => {
  test.beforeEach(async ({ page }) => {
    await loginUi(page, 'student');
    await page.goto('/#courses');
    await page.locator('[data-open-course="crs-ai-101"]').first().click();
    await expect(page.locator('#open-discussions-modal-btn')).toBeVisible();
  });

  test('la lección de video incrusta YouTube (regresión: contentUrl)', async ({ page }) => {
    await expect(page.locator('iframe[src^="https://www.youtube.com/embed/"]').first()).toBeVisible();
    expect(page.problems).toEqual([]);
  });

  test('las lecciones de PDF y de texto se renderizan', async ({ page }) => {
    await page.getByText('Guía de Arquitectura').first().click();
    await expect(page.locator('iframe, embed, object, a[href$=".pdf"]').first()).toBeAttached();
    await page.getByText('Principios de Context Engine').first().click();
    await expect(page.locator('#app-content')).toContainText(/Context Engine/i);
    expect(page.problems).toEqual([]);
  });

  test('marcar una lección como completada actualiza el progreso', async ({ page }) => {
    await page.getByText('Guía de Arquitectura').first().click();
    const btn = page.locator('#toggle-lesson-progress-btn');
    await expect(btn).toBeVisible();
    const before = await page.locator('#app-content').innerText();
    await btn.click();
    await expect(page.locator('#app-content')).not.toHaveText(before);
    expect(page.problems).toEqual([]);
  });

  test('el foro, las reseñas y el tutor IA abren sin errores', async ({ page }) => {
    await page.click('#open-discussions-modal-btn');
    await expect(page.locator('#modal-container')).toBeVisible();
    await page.click('#modal-confirm-btn');
    await expect(page.locator('#modal-container')).toBeHidden();
    await page.click('#open-reviews-modal-btn');
    await expect(page.locator('#modal-container')).toBeVisible();
    await page.click('#modal-confirm-btn');
    await expect(page.locator('#modal-container')).toBeHidden();
    await page.click('#open-tutor-modal-btn');
    await expect(page.locator('#modal-container')).toBeVisible();
    expect(page.problems).toEqual([]);
  });

  test('publicar en el foro funciona y se muestra', async ({ page }) => {
    await page.click('#open-discussions-modal-btn');
    await page.fill('#disc-new-title', 'Duda de prueba');
    await page.fill('#disc-new-message', 'Mensaje e2e del foro');
    await page.click('#disc-post-btn');
    await expect(page.getByText('Mensaje e2e del foro')).toBeVisible();
    expect(page.problems).toEqual([]);
  });
});
