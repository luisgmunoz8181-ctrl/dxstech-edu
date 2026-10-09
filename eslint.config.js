// Configuración mínima de ESLint para el frontend (web/js): detecta identificadores
// sin definir (imports olvidados, métodos movidos entre módulos) y claves duplicadas.
const browser = ['window','document','fetch','localStorage','sessionStorage','console','setTimeout','clearTimeout','setInterval','clearInterval','FormData','URL','URLSearchParams','Blob','Image','FileReader','navigator','location','history','requestAnimationFrame','cancelAnimationFrame','Audio','AudioContext','webkitAudioContext','XLSX','alert','confirm','prompt','Event','CustomEvent','HTMLElement','Element','Node','MutationObserver','IntersectionObserver','ResizeObserver','AbortController','Promise','Intl','encodeURIComponent','decodeURIComponent','btoa','atob','TextEncoder','TextDecoder','structuredClone','getComputedStyle','matchMedia','performance','crypto','CSS','DOMParser','XMLSerializer','Uint8Array','ArrayBuffer','Response','Headers','Request','File','FileList','KeyboardEvent','MouseEvent','DragEvent','ClipboardEvent','DataTransfer','Path2D','OffscreenCanvas','createImageBitmap'];
module.exports = [{
  files: ['web/js/**/*.js'],
  languageOptions: { ecmaVersion: 2023, sourceType: 'module', globals: Object.fromEntries(browser.map(g => [g, 'readonly'])) },
  rules: { 'no-undef': 'error', 'no-unused-vars': ['warn', { args: 'none', caughtErrors: 'none' }], 'no-dupe-keys': 'error', 'no-redeclare': 'error' },
}];
