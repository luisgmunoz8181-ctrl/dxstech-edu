// @ts-check
const { defineConfig } = require('@playwright/test');
const path = require('path');
const os = require('os');

const PORT = process.env.E2E_PORT || '3300';
// Base de datos temporal y limpia en cada ejecución (modo --dev siembra los datos demo).
const DATA_DIR = path.join(process.env.E2E_DATA_DIR || os.tmpdir(), `dxstech-e2e-${Date.now()}`);

module.exports = defineConfig({
  testDir: './tests',
  timeout: 45_000,
  expect: { timeout: 8_000 },
  fullyParallel: false, // comparten una sola base de datos
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: `http://localhost:${PORT}`,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    // Chromium preinstalado en entornos sin descarga de navegadores.
    launchOptions: process.env.PW_CHROMIUM_PATH ? { executablePath: process.env.PW_CHROMIUM_PATH } : {},
  },
  webServer: {
    command: 'go run ./cmd/server --dev',
    cwd: path.resolve(__dirname, '..'),
    url: `http://localhost:${PORT}/api/health`,
    timeout: 180_000,
    reuseExistingServer: false,
    env: { PORT, DATA_DIR, TRUSTED_PROXIES: 'none' },
  },
});
