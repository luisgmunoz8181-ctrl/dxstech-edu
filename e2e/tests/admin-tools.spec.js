const { test, expect, loginApi, loginUi } = require('./fixtures');

test.describe('Administración: auditoría, moderación y copias', () => {
  test('el módulo solo existe para administradores y las copias solo para el superadministrador', async ({ page }) => {
    await loginUi(page, 'student');
    await expect(page.locator('.nav-btn[data-nav="admin"]')).toBeHidden();
    await page.evaluate(() => window.router.navigate('admin'));
    await expect(page.locator('#view-title')).toContainText('Cursos'); // redirigido

    await page.evaluate(() => fetch('/api/auth/logout', { method: 'POST' }));
    await loginUi(page, 'admin');
    await page.click('.nav-btn[data-nav="admin"]');
    await expect(page.locator('#view-title')).toContainText('Administración');
    await expect(page.locator('#adm-tab-audit')).toBeVisible();
    await expect(page.locator('#adm-tab-moderation')).toBeVisible();
    await expect(page.locator('#adm-tab-backups')).toHaveCount(0); // ADMINISTRADOR no gestiona copias
  });

  test('la auditoría muestra y filtra las acciones administrativas', async ({ page, browser, baseURL }) => {
    const admin = await browser.newContext({ baseURL });
    await loginApi(admin.request, 'admin');
    const code = `AUD-${Date.now()}`;
    await admin.request.post('/api/courses', { data: { title: 'Curso para auditar', code, description: 'd', category: 'Tecnología', level: 'Principiante', durationHours: 1 } });
    await admin.close();

    await loginUi(page, 'admin');
    await page.click('.nav-btn[data-nav="admin"]');
    await page.selectOption('#audit-action', 'COURSE_CREATE');
    await page.fill('#audit-search', code);
    await page.click('#audit-apply');
    const rows = page.locator('#audit-rows tr');
    await expect(rows).toHaveCount(1);
    await expect(rows.first()).toContainText(code);
    await expect(rows.first()).toContainText('COURSE_CREATE');
    await expect(rows.first()).toContainText('Admin DxSTech');

    await page.fill('#audit-search', 'texto-que-no-existe-xyz');
    await page.click('#audit-apply');
    await expect(page.getByText('Sin eventos para este filtro.')).toBeVisible();
    expect(page.problems).toEqual([]);
  });

  test('copias de seguridad: crear, listar y descargar (superadministrador)', async ({ page }) => {
    await loginUi(page, 'superadmin');
    await page.click('.nav-btn[data-nav="admin"]');
    await page.click('#adm-tab-backups');
    await expect(page.getByText('Aún no hay copias.')).toBeVisible();

    await page.click('#backup-create');
    await expect(page.getByText(/Copia creada: dxstech-backup-/)).toBeVisible();
    const link = page.locator('a[href^="/api/admin/backups/dxstech-backup-"]');
    await expect(link).toHaveCount(1);

    const [download] = await Promise.all([page.waitForEvent('download'), link.click()]);
    expect(download.suggestedFilename()).toMatch(/^dxstech-backup-\d{8}-\d{6}\.tar\.gz$/);
    const stream = await download.createReadStream();
    const first = await new Promise((resolve) => stream.once('data', (b) => resolve(b)));
    expect([first[0], first[1]]).toEqual([0x1f, 0x8b]); // cabecera gzip
    expect(page.problems).toEqual([]);
  });

  test('moderación: un reporte del foro se revisa y la publicación se elimina', async ({ page, browser, baseURL }) => {
    const student = await browser.newContext({ baseURL });
    const admin = await browser.newContext({ baseURL });
    await loginApi(student.request, 'student');
    await loginApi(admin.request, 'admin');

    const message = `Mensaje a moderar ${Date.now()}`;
    const posted = await student.request.post('/api/courses/crs-ai-101/discussions', { data: { message, title: 'Duda' } });
    const post = await posted.json();
    // El administrador (con acceso al curso) reporta la publicación del estudiante.
    const rep = await admin.request.post(`/api/courses/crs-ai-101/discussions/${post.id}/report`, { data: { reason: 'Prueba de moderación' } });
    expect(rep.status()).toBe(201);

    await loginUi(page, 'admin');
    await page.click('.nav-btn[data-nav="admin"]');
    await page.click('#adm-tab-moderation');
    const card = page.locator('[data-report]').filter({ hasText: message });
    await expect(card).toBeVisible();
    await expect(card).toContainText('Prueba de moderación');
    await card.locator('[data-resolve="remove"]').click();
    await expect(page.getByText('Publicación eliminada').first()).toBeVisible();
    await expect(page.locator('[data-report]').filter({ hasText: message })).toHaveCount(0); // ya no está pendiente

    // La publicación desapareció del foro y el historial la conserva.
    const forum = await (await student.request.get('/api/courses/crs-ai-101/discussions')).json();
    expect(forum.find((d) => d.id === post.id)).toBeUndefined();
    await page.selectOption('#mod-status', 'all');
    await expect(page.locator('[data-report]').filter({ hasText: message })).toContainText('publicación eliminada');
    await student.close();
    await admin.close();
  });
});
