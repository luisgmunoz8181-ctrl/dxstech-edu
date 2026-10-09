// Utilidades de saneamiento para plantillas que se asignan con innerHTML.
// Todo texto que provenga de la API, de la URL o del usuario debe pasar por
// esc() (texto/atributos) o safeUrl() (src/href) antes de interpolarse.

const ESCAPES = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;', '`': '&#96;' };

export function esc(value) {
  if (value === null || value === undefined) return '';
  return String(value).replace(/[&<>"'`]/g, (ch) => ESCAPES[ch]);
}

// Admite rutas relativas del mismo origen y URLs http(s); descarta
// javascript:, data:, vbscript:, etc. El resultado ya viene escapado.
export function safeUrl(value) {
  const raw = String(value ?? '').trim();
  if (!raw) return '';
  if (/^\/(?!\/)/.test(raw)) return esc(raw);
  try {
    const u = new URL(raw);
    if (u.protocol === 'http:' || u.protocol === 'https:') return esc(raw);
  } catch (_) { /* URL inválida */ }
  return '';
}
