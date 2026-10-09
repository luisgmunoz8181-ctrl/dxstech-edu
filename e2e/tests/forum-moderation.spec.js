const { test, expect, loginApi, loginUi } = require('./fixtures');

async function openForum(page) {
  await page.goto('/#courses');
  await page.locator('[data-open-course="crs-ai-101"]').first().click();
  await page.click('#open-discussions-modal-btn');
  await expect(page.locator('#modal-container')).toBeVisible();
}

test.describe('Foro: eliminar y reportar desde la interfaz', () => {
  test('el autor elimina su publicación con doble confirmación', async ({ page }) => {
    await loginUi(page, 'student');
    await openForum(page);

    const text = `Para borrar ${Date.now()}`;
    await page.fill('#disc-new-title', 'Duda');
    await page.fill('#disc-new-message', text);
    await page.click('#disc-post-btn');
    const post = page.locator('#disc-items-container > div').filter({ hasText: text });
    await expect(post).toBeVisible();

    const del = post.locator('[data-delete-post]');
    await del.click();
    await expect(del).toContainText('¿Seguro?');
    await expect(post).toBeVisible(); // el primer clic no borra
    await del.click();
    await expect(page.getByText('Publicación eliminada').first()).toBeVisible();
    await expect(page.locator('#disc-items-container').getByText(text)).toHaveCount(0);
    expect(page.problems).toEqual([]);
  });

  test('un estudiante reporta la publicación de otro, pero no la propia', async ({ page, browser, baseURL }) => {
    const admin = await browser.newContext({ baseURL });
    await loginApi(admin.request, 'admin');
    const text = `Publicación del docente ${Date.now()}`;
    await admin.request.post('/api/courses/crs-ai-101/discussions', { data: { message: text, title: 'Aviso' } });
    await admin.close();

    await loginUi(page, 'student');
    await openForum(page);
    const post = page.locator('#disc-items-container > div').filter({ hasText: text });
    await expect(post.locator('[data-delete-post]')).toHaveCount(0); // un estudiante no puede borrar la ajena
    await post.locator('[data-report-post]').click();
    await post.locator('input[id^="report-reason-"]').fill('Prueba e2e');
    await post.locator('[data-send-report]').click();
    await expect(page.getByText(/reporte fue enviado/i).first()).toBeVisible();

    // Reportar dos veces es un conflicto claro.
    await post.locator('[data-report-post]').click(); // el formulario se cerró tras enviar: se reabre
    await post.locator('[data-send-report]').click();
    await expect(page.getByText(/ya reportaste/i).first()).toBeVisible();
    expect(page.problems).toEqual([]);
  });
});
