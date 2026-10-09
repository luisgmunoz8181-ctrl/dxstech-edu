const path = require('path');
const fs = require('fs');
const { test, expect } = require('./fixtures');

const fixture = (name) => fs.readFileSync(path.join(__dirname, '..', 'fixtures', name)).toString('base64');

// Ejecuta readSpreadsheetRows dentro de la página (módulo ES real, DOMParser real).
async function readRows(page, name, base64, mime = 'application/octet-stream') {
  return page.evaluate(async ({ name, base64, mime }) => {
    const { readSpreadsheetRows } = await import('/js/utils/spreadsheet.js');
    const bytes = Uint8Array.from(atob(base64), (c) => c.charCodeAt(0));
    try {
      return { rows: await readSpreadsheetRows(new File([bytes], name, { type: mime })) };
    } catch (e) {
      return { error: e.message };
    }
  }, { name, base64, mime });
}

test.describe('Lector de Excel/CSV (reemplazo de SheetJS)', () => {
  test.beforeEach(async ({ page }) => { await page.goto('/#courses'); });

  test('lee un .xlsx con acentos, números largos, comillas y filas vacías', async ({ page }) => {
    const { rows, error } = await readRows(page, 'alumnos.xlsx', fixture('alumnos.xlsx'));
    expect(error).toBeUndefined();
    expect(rows).toEqual([
      ['Nombre completo', 'Correo', 'Teléfono'],
      ['María José Peña', 'maria@example.com', '573001234567'],
      ['José Ñandú', 'jose@example.com', '+57 300 765 4321'],
      ['Ana "la Dra." Gómez', 'ana@example.com', '3109876543'],
    ]);
  });

  test('solo usa la primera hoja', async ({ page }) => {
    const { rows } = await readRows(page, 'dos-hojas.xlsx', fixture('dos-hojas.xlsx'));
    expect(rows).toEqual([['Nombre'], ['Solo en la primera hoja']]);
  });

  test('lee un .csv con punto y coma, comillas, BOM y líneas vacías', async ({ page }) => {
    const { rows } = await readRows(page, 'alumnos.csv', fixture('alumnos.csv'), 'text/csv');
    expect(rows).toEqual([
      ['Nombre', 'Correo'],
      ['Pérez; Juan', 'juan@example.com'],
      ['Con "comillas"', 'c@example.com'],
      ['Sin correo', ''],
    ]);
  });

  test('rechaza .xls antiguo, formatos desconocidos y archivos dañados con mensajes claros', async ({ page }) => {
    expect((await readRows(page, 'viejo.xls', 'AAAA')).error).toMatch(/\.xls.*no es compatible/i);
    expect((await readRows(page, 'archivo.pdf', 'JVBERi0xLjQ=')).error).toMatch(/formato no compatible/i);
    expect((await readRows(page, 'roto.xlsx', 'UEsDBAoAAAAAAAAAAAAAAAAAAAAAAAAA')).error).toMatch(/no se pudo leer|dañado|válido/i);
  });

  test('rechaza una "zip bomb" y archivos de más de 5 MB', async ({ page }) => {
    const result = await page.evaluate(async () => {
      const { zipSync } = await import('/vendor/fflate-0.8.2.browser.js');
      const { readSpreadsheetRows } = await import('/js/utils/spreadsheet.js');
      // ~35 MB descomprimidos que ocupan pocos KB comprimidos.
      const bomb = zipSync({
        'xl/workbook.xml': new TextEncoder().encode('<workbook/>'),
        'xl/worksheets/sheet1.xml': new Uint8Array(35 * 1024 * 1024),
      }, { level: 9 });
      const out = { bombSize: bomb.length };
      try { await readSpreadsheetRows(new File([bomb], 'bomba.xlsx')); out.bomb = 'aceptada'; } catch (e) { out.bomb = e.message; }
      try { await readSpreadsheetRows(new File([new Uint8Array(6 * 1024 * 1024)], 'grande.csv')); out.big = 'aceptado'; } catch (e) { out.big = e.message; }
      return out;
    });
    expect(result.bombSize).toBeLessThan(5 * 1024 * 1024);
    expect(result.bomb).toMatch(/demasiado grande/i);
    expect(result.big).toMatch(/supera el máximo/i);
  });

  test('un libro con claves tipo __proto__ no contamina Object.prototype', async ({ page }) => {
    const polluted = await page.evaluate(async () => {
      const { parseCsv } = await import('/js/utils/spreadsheet.js');
      parseCsv('__proto__,constructor\npolluted,yes\n');
      return ({}).polluted === 'yes';
    });
    expect(polluted).toBe(false);
  });
});

test.describe('Importación en la interfaz', () => {
  test('la lista de alumnos del generador de certificados se importa desde un .xlsx', async ({ page }) => {
    await page.goto('/#login');
    await page.fill('#login-email', 'admin@dxstech.edu');
    await page.fill('#login-password', 'Admin1234*');
    await page.click('#login-submit-btn');
    await expect(page.locator('#view-title')).toContainText('Dashboard');
    await page.click('.nav-btn[data-nav="certificates"]');
    await expect(page.locator('#view-title')).toContainText(/Certificad/i);
    await page.click('#cert-tab-gen');
    await page.setInputFiles('#excel-file-input', path.join(__dirname, '..', 'fixtures', 'alumnos.xlsx'));
    await expect(page.getByText(/Se importaron 3 estudiantes/)).toBeVisible();
    await expect(page.locator('#excel-column-select option')).toHaveCount(3);
    expect(page.problems).toEqual([]);
  });

  test('los destinatarios de WhatsApp se importan desde un .xlsx y un .csv', async ({ page }) => {
    await page.goto('/#login');
    await page.fill('#login-email', 'admin@dxstech.edu');
    await page.fill('#login-password', 'Admin1234*');
    await page.click('#login-submit-btn');
    await expect(page.locator('#view-title')).toContainText('Dashboard');
    await page.click('.nav-btn[data-nav="whatsapp"]');
    await page.click('#wa-tab-send-btn');
    await expect(page.locator('#recipients-input')).toBeVisible();

    await page.setInputFiles('#wa-recipients-file-input', path.join(__dirname, '..', 'fixtures', 'alumnos.xlsx'));
    await expect(page.getByText(/Se importaron 3 números/)).toBeVisible();
    const value = await page.inputValue('#recipients-input');
    expect(value.split('\n')).toEqual(['573001234567', '+573007654321', '3109876543']);
    expect(page.problems).toEqual([]);
  });
});
