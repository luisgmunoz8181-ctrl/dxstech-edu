import { Toast } from '../components/toast.js';
import { Modal } from '../components/modal.js';
import { Loading } from '../components/loading.js';

export const SettingsView = {
  cleanupFns: [],

  render() {
    return `
      <div class="space-y-6 max-w-4xl mx-auto">
        <!-- Top Banner -->
        <div class="bg-gradient-to-r from-slate-800 to-slate-900 rounded-2xl p-6 text-white shadow-lg shadow-slate-200 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h3 class="text-lg font-bold">Configuración de Seguridad y Gemini API</h3>
            <p class="text-xs text-slate-300 mt-1 max-w-xl">
              Administra tu clave privada de Inteligencia Artificial para habilitar las evaluaciones automáticas y el asistente virtual en tiempo real.
            </p>
          </div>
          <div class="flex items-center gap-2 text-xs bg-white/10 px-3 py-1.5 rounded-xl border border-white/10 font-mono">
            <span>BYOK: Bring Your Own Key</span>
          </div>
        </div>

        <!-- Gemini API Key Management Card -->
        <div class="bg-white rounded-2xl p-6 border border-slate-200/80 shadow-sm space-y-5">
          <div class="flex items-center justify-between pb-3 border-b border-slate-100">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold">
                <i data-lucide="key" class="w-5 h-5"></i>
              </div>
              <div>
                <h4 class="text-sm font-bold text-slate-800">Tu Google Gemini API Key</h4>
                <p class="text-xs text-slate-400">Esta clave se almacena de forma exclusiva en tu navegador</p>
              </div>
            </div>
            <span id="current-key-badge" class="text-xs font-semibold px-2.5 py-1 rounded-full bg-slate-100 text-slate-600">
              Verificando...
            </span>
          </div>

          <!-- Key Input Box -->
          <div class="space-y-2">
            <label class="block text-xs font-semibold text-slate-700">Introduce tu API Key personal</label>
            <div class="relative">
              <input type="password" id="gemini-key-input" placeholder="AIzaSy..." class="w-full pl-3 pr-12 py-3 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono tracking-wide">
              <button type="button" id="toggle-key-visibility" class="absolute right-3 top-3 text-slate-400 hover:text-slate-600 p-0.5">
                <i data-lucide="eye" class="w-4 h-4" id="eye-icon"></i>
              </button>
            </div>
            <p class="text-[11px] text-slate-400">
              ¿No tienes una Gemini API Key? Es 100% gratuita para desarrollo en <a href="https://aistudio.google.com/app/apikey" target="_blank" class="text-indigo-600 font-medium hover:underline inline-flex items-center gap-0.5">Google AI Studio <i data-lucide="external-link" class="w-3 h-3"></i></a>.
            </p>
          </div>

          <!-- Action Buttons -->
          <div class="flex flex-wrap items-center gap-3 pt-2">
            <button id="save-key-btn" class="px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 shadow-md shadow-indigo-200 transition-all flex items-center gap-2">
              <i data-lucide="save" class="w-4 h-4"></i>
              <span>Guardar en este Navegador</span>
            </button>

            <button id="test-key-btn" class="px-4 py-2.5 rounded-xl text-xs font-bold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center gap-2">
              <i data-lucide="sparkles" class="w-4 h-4 text-purple-600"></i>
              <span>Probar Conexión</span>
            </button>

            <button id="delete-key-btn" class="px-4 py-2.5 rounded-xl text-xs font-bold text-rose-600 hover:bg-rose-50 border border-rose-200 transition-colors flex items-center gap-2 ml-auto">
              <i data-lucide="trash-2" class="w-4 h-4"></i>
              <span>Eliminar Clave</span>
            </button>
          </div>
        </div>

        <!-- Privacy & Security Guarantee Card -->
        <div class="bg-indigo-50/50 rounded-2xl p-6 border border-indigo-100 space-y-3">
          <div class="flex items-center gap-2 text-indigo-900 font-bold text-xs uppercase tracking-wider">
            <i data-lucide="shield-check" class="w-4 h-4 text-indigo-600"></i>
            Garantía de Privacidad y Arquitectura Segura
          </div>
          <p class="text-xs text-slate-600 leading-relaxed">
            DxSTech Edu sigue una política estricta de aislamiento de credenciales:
          </p>
          <ul class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs text-slate-700 pt-1">
            <li class="flex items-start gap-2 bg-white/80 p-2.5 rounded-xl border border-indigo-100">
              <i data-lucide="check" class="w-4 h-4 text-emerald-600 shrink-0 mt-0.5"></i>
              <span><strong>Sin almacenamiento en servidor:</strong> Tu clave nunca se guarda en SQLite, archivos ni base de datos.</span>
            </li>
            <li class="flex items-start gap-2 bg-white/80 p-2.5 rounded-xl border border-indigo-100">
              <i data-lucide="check" class="w-4 h-4 text-emerald-600 shrink-0 mt-0.5"></i>
              <span><strong>Sin persistencia en variables de entorno:</strong> No existe una clave global compartida en .env.</span>
            </li>
            <li class="flex items-start gap-2 bg-white/80 p-2.5 rounded-xl border border-indigo-100">
              <i data-lucide="check" class="w-4 h-4 text-emerald-600 shrink-0 mt-0.5"></i>
              <span><strong>Uso efímero en memoria:</strong> Se envía temporalmente con cada solicitud de IA y se descarta al instante.</span>
            </li>
            <li class="flex items-start gap-2 bg-white/80 p-2.5 rounded-xl border border-indigo-100">
              <i data-lucide="check" class="w-4 h-4 text-emerald-600 shrink-0 mt-0.5"></i>
              <span><strong>Logs totalmente sanitizados:</strong> Los registros de red del servidor nunca exponen tu clave.</span>
            </li>
          </ul>
        </div>
      </div>
    `;
  },

  mount() {
    this.bindEvents();
    this.refreshKeyStatus();
  },

  bindEvents() {
    const input = document.getElementById('gemini-key-input');
    const toggleBtn = document.getElementById('toggle-key-visibility');
    const saveBtn = document.getElementById('save-key-btn');
    const testBtn = document.getElementById('test-key-btn');
    const deleteBtn = document.getElementById('delete-key-btn');

    // Toggle Eye Visibility
    toggleBtn.addEventListener('click', () => {
      const isPwd = input.type === 'password';
      input.type = isPwd ? 'text' : 'password';
      toggleBtn.innerHTML = `<i data-lucide="${isPwd ? 'eye-off' : 'eye'}" class="w-4 h-4"></i>`;
      if (window.lucide) window.lucide.createIcons();
    });

    // Save
    saveBtn.addEventListener('click', () => {
      const key = input.value.trim();
      if (!key) {
        Toast.warning('Por favor introduce una clave válida antes de guardar.');
        input.focus();
        return;
      }
      localStorage.setItem('dxstech_gemini_api_key', key);
      window.dispatchEvent(new CustomEvent('gemini-key-changed', { detail: { key } }));
      this.refreshKeyStatus();
      Toast.success('Gemini API Key guardada exitosamente en este navegador.');
    });

    // Test Key
    testBtn.addEventListener('click', async () => {
      const key = input.value.trim() || localStorage.getItem('dxstech_gemini_api_key');
      if (!key) {
        Toast.warning('Ingresa tu clave para realizar la prueba de conexión.');
        return;
      }

      Loading.overlay('Validando clave con la API de Google Gemini...');
      try {
        const res = await fetch(`https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=${key}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            contents: [{ parts: [{ text: 'Responde únicamente con la palabra "OK".' }] }]
          })
        });

        if (!res.ok) {
          const err = await res.json().catch(() => ({}));
          throw new Error(err.error?.message || `HTTP ${res.status}`);
        }

        Toast.success('✅ ¡Conexión con Gemini verificada exitosamente!');
      } catch (err) {
        Toast.error('Fallo en la prueba de Gemini: ' + err.message);
      } finally {
        Loading.hideOverlay();
      }
    });

    // Delete
    deleteBtn.addEventListener('click', () => {
      Modal.show({
        title: '¿Eliminar Gemini API Key?',
        content: 'La clave se borrará de tu almacenamiento local. Las funciones de generación de evaluaciones requerirán que la ingreses de nuevo.',
        confirmText: 'Sí, eliminar',
        cancelText: 'Cancelar',
        showCancel: true,
        icon: 'trash-2',
        iconColor: 'text-rose-600 bg-rose-50',
        onConfirm: () => {
          localStorage.removeItem('dxstech_gemini_api_key');
          input.value = '';
          window.dispatchEvent(new CustomEvent('gemini-key-changed', { detail: { key: null } }));
          this.refreshKeyStatus();
          Toast.info('Gemini API Key eliminada del navegador.');
        }
      });
    });
  },

  refreshKeyStatus() {
    const input = document.getElementById('gemini-key-input');
    const badge = document.getElementById('current-key-badge');
    const key = localStorage.getItem('dxstech_gemini_api_key');

    if (key) {
      input.value = key;
      badge.textContent = '🟢 Clave Configurada';
      badge.className = 'text-xs font-semibold px-2.5 py-1 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200';
    } else {
      input.value = '';
      badge.textContent = '🔴 Sin Clave';
      badge.className = 'text-xs font-semibold px-2.5 py-1 rounded-full bg-amber-50 text-amber-700 border border-amber-200';
    }
  },

  destroy() {
    this.cleanupFns.forEach(fn => fn());
    this.cleanupFns = [];
  }
};
