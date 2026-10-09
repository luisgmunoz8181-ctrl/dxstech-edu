const { test, expect, XSS, loginApi } = require('./fixtures');

test.describe('XSS: los datos de usuarios se muestran como texto, nunca como HTML', () => {
  test('foro, reseñas y catálogo con cargas maliciosas', async ({ browser, baseURL }) => {
    const admin = await browser.newContext({ baseURL });
    const student = await browser.newContext({ baseURL });
    await loginApi(admin.request, 'admin');
    await loginApi(student.request, 'student');

    // Curso con título malicioso, publicado.
    const code = `XSS-${Date.now()}`;
    const created = await admin.request.post('/api/courses', {
      data: { title: `${XSS} Curso`, code, description: XSS, shortDescription: XSS, category: XSS, instructorName: XSS, level: 'Principiante', durationHours: 1 },
    });
    expect(created.status()).toBe(201);
    const course = await created.json();
    await admin.request.patch(`/api/courses/${course.id}/status`, { data: { status: 'published' } });

    // Mensajes del estudiante con HTML.
    const forum = await student.request.post('/api/courses/crs-ai-101/discussions', { data: { message: `${XSS} hola foro` } });
    expect(forum.status()).toBe(201);
    const review = await student.request.post('/api/courses/crs-ai-101/reviews', { data: { rating: 5, comment: `${XSS} excelente` } });
    expect(review.status()).toBe(201);

    for (const [name, ctx] of [['estudiante', student], ['administrador', admin]]) {
      const page = await ctx.newPage();
      const dialogs = [];
      page.on('dialog', (d) => { dialogs.push(d.message()); d.dismiss(); });
      await ctx.route(/^https?:\/\/(?!localhost)/, (r) => r.abort());

      await page.goto('/#courses');
      await expect(page.getByText('Curso').first()).toBeVisible();
      // La carga útil se ve literalmente como texto en el catálogo.
      await expect(page.getByText('<img src=x onerror=', { exact: false }).first()).toBeVisible();

      if (name === 'estudiante') {
        await page.locator('[data-open-course="crs-ai-101"]').first().click();
        await page.click('#open-discussions-modal-btn');
        await expect(page.getByText('hola foro')).toBeVisible();
        await expect(page.locator('#modal-container img[src="x"]')).toHaveCount(0);
        await page.click('#modal-confirm-btn');
        await expect(page.locator('#modal-container')).toBeHidden();
        await page.click('#open-reviews-modal-btn');
        await expect(page.getByText('excelente')).toBeVisible();
        await expect(page.locator('img[src="x"]')).toHaveCount(0);
      }

      expect(await page.evaluate(() => window.__xss || 0), `XSS ejecutado (${name})`).toBe(0);
      expect(dialogs).toEqual([]);
      await page.close();
    }

    await admin.close();
    await student.close();
  });

  test('la verificación pública no refleja HTML del código de la URL', async ({ page }) => {
    await page.goto('/#verify/' + encodeURIComponent('<img src=x onerror=window.__xss=1>'));
    await page.waitForTimeout(1200);
    expect(await page.evaluate(() => window.__xss || 0)).toBe(0);
    await expect(page.locator('img[src="x"]')).toHaveCount(0);
  });
});

test.describe('Sin dependencias externas en ejecución', () => {
  test('Tailwind, Lucide y SheetJS se cargan desde el propio servidor', async ({ page }) => {
    const external = [];
    page.on('request', (r) => {
      if (r.resourceType() === 'image') return; // las miniaturas de cursos demo son datos, no código
      const u = new URL(r.url());
      if (u.hostname !== 'localhost' && !/fonts\.(googleapis|gstatic)\.com/.test(u.hostname)) external.push(r.url());
    });
    await page.goto('/#courses');
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(800);
    expect(external).toEqual([]);
    expect(await page.evaluate(() => typeof window.lucide)).toBe('object');
    expect(await page.evaluate(() => typeof window.XLSX)).toBe('object');
    // Si el CSS compilado cargó, el botón de navegación tiene estilos de Tailwind.
    const display = await page.evaluate(() => getComputedStyle(document.querySelector('.nav-btn[data-nav="courses"]')).display);
    expect(display).toBe('flex');
  });
});
