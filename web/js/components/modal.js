// Modal dialog component

class ModalManager {
  constructor() {
    this.container = document.getElementById('modal-container');
  }

  getContainer() {
    if (!this.container) {
      this.container = document.getElementById('modal-container');
    }
    return this.container;
  }

  show({
    title = 'Atención',
    content = '',
    confirmText = 'Aceptar',
    cancelText = 'Cancelar',
    showCancel = false,
    icon = 'info',
    iconColor = 'text-indigo-600 bg-indigo-50',
    maxWidth = 'max-w-md',
    onConfirm = () => {},
    onCancel = () => {},
  }) {
    const container = this.getContainer();
    if (!container) return;

    container.innerHTML = `
      <div class="bg-white rounded-2xl shadow-2xl border border-slate-200/80 ${maxWidth} w-full p-6 transform transition-all duration-200 scale-95 opacity-0 modal-card max-h-[90vh] flex flex-col">
        <div class="flex items-start gap-4">
          <div class="w-10 h-10 rounded-xl ${iconColor} flex items-center justify-center shrink-0">
            <i data-lucide="${icon}" class="w-5 h-5"></i>
          </div>
          <div class="flex-1">
            <h3 class="text-base font-bold text-slate-900 mb-1">${this.escapeHtml(title)}</h3>
            <div class="text-sm text-slate-600 leading-relaxed">${content}</div>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-end gap-3">
          ${showCancel ? `
            <button id="modal-cancel-btn" class="px-4 py-2 rounded-xl text-xs font-semibold text-slate-600 hover:bg-slate-100 transition-colors">
              ${this.escapeHtml(cancelText)}
            </button>
          ` : ''}
          <button id="modal-confirm-btn" class="px-5 py-2 rounded-xl text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-700 shadow-sm shadow-indigo-200 transition-all">
            ${this.escapeHtml(confirmText)}
          </button>
        </div>
      </div>
    `;

    container.classList.remove('hidden');
    if (window.lucide) window.lucide.createIcons();

    const card = container.querySelector('.modal-card');
    requestAnimationFrame(() => {
      card.classList.remove('scale-95', 'opacity-0');
      card.classList.add('scale-100', 'opacity-100');
    });

    const confirmBtn = container.querySelector('#modal-confirm-btn');
    const cancelBtn = container.querySelector('#modal-cancel-btn');

    const close = () => {
      card.classList.remove('scale-100', 'opacity-100');
      card.classList.add('scale-95', 'opacity-0');
      setTimeout(() => {
        container.classList.add('hidden');
        container.innerHTML = '';
      }, 150);
    };

    confirmBtn.addEventListener('click', () => {
      close();
      onConfirm();
    });

    if (cancelBtn) {
      cancelBtn.addEventListener('click', () => {
        close();
        onCancel();
      });
    }

    // Dismiss on backdrop click
    container.onclick = (e) => {
      if (e.target === container) {
        close();
        onCancel();
      }
    };
  }

  promptGeminiKey({ onSaved = () => {}, onCancel = () => {} } = {}) {
    const container = this.getContainer();
    if (!container) return;

    container.innerHTML = `
      <div class="bg-white rounded-2xl shadow-2xl border border-slate-200/80 max-w-lg w-full p-6 transform transition-all duration-200 scale-95 opacity-0 modal-card">
        <div class="flex items-start gap-4">
          <div class="w-11 h-11 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center shrink-0">
            <i data-lucide="sparkles" class="w-6 h-6"></i>
          </div>
          <div class="flex-1">
            <h3 class="text-base font-bold text-slate-900 mb-1">Configura tu Gemini API Key</h3>
            <p class="text-xs text-slate-500 leading-relaxed mb-4">
              DxSTech Edu requiere tu propia clave de Google Gemini para generar evaluaciones y potenciar el Asistente Virtual. La clave se almacena <strong class="text-indigo-600">únicamente en tu navegador</strong> y nunca en nuestros servidores.
            </p>

            <div class="space-y-2">
              <label class="block text-xs font-semibold text-slate-700">Tu API Key (Google AI Studio)</label>
              <div class="relative">
                <input type="password" id="modal-gemini-input" placeholder="AIzaSy..." class="w-full pl-3 pr-10 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono">
                <button type="button" id="modal-gemini-toggle" class="absolute right-2.5 top-2.5 text-slate-400 hover:text-slate-600">
                  <i data-lucide="eye" class="w-4 h-4"></i>
                </button>
              </div>
              <p class="text-[11px] text-slate-400">
                ¿No tienes una? Consíguela gratis en <a href="https://aistudio.google.com/app/apikey" target="_blank" class="text-indigo-600 font-medium hover:underline">Google AI Studio</a>.
              </p>
            </div>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-end gap-3 pt-3 border-t border-slate-100">
          <button id="modal-gemini-cancel" class="px-4 py-2 rounded-xl text-xs font-semibold text-slate-600 hover:bg-slate-100 transition-colors">
            Cancelar
          </button>
          <button id="modal-gemini-save" class="px-5 py-2 rounded-xl text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-700 shadow-sm shadow-indigo-200 transition-all flex items-center gap-1.5">
            <i data-lucide="check" class="w-3.5 h-3.5"></i>
            Guardar en Navegador
          </button>
        </div>
      </div>
    `;

    container.classList.remove('hidden');
    if (window.lucide) window.lucide.createIcons();

    const card = container.querySelector('.modal-card');
    requestAnimationFrame(() => {
      card.classList.remove('scale-95', 'opacity-0');
      card.classList.add('scale-100', 'opacity-100');
    });

    const input = container.querySelector('#modal-gemini-input');
    const toggle = container.querySelector('#modal-gemini-toggle');
    const saveBtn = container.querySelector('#modal-gemini-save');
    const cancelBtn = container.querySelector('#modal-gemini-cancel');

    // Pre-fill existing key if available
    const existing = localStorage.getItem('dxstech_gemini_api_key');
    if (existing) input.value = existing;

    toggle.addEventListener('click', () => {
      const isPwd = input.type === 'password';
      input.type = isPwd ? 'text' : 'password';
      toggle.innerHTML = `<i data-lucide="${isPwd ? 'eye-off' : 'eye'}" class="w-4 h-4"></i>`;
      if (window.lucide) window.lucide.createIcons();
    });

    const close = () => {
      card.classList.remove('scale-100', 'opacity-100');
      card.classList.add('scale-95', 'opacity-0');
      setTimeout(() => {
        container.classList.add('hidden');
        container.innerHTML = '';
      }, 150);
    };

    saveBtn.addEventListener('click', () => {
      const val = input.value.trim();
      if (!val) {
        input.focus();
        input.classList.add('ring-2', 'ring-rose-500');
        return;
      }
      localStorage.setItem('dxstech_gemini_api_key', val);
      window.dispatchEvent(new CustomEvent('gemini-key-changed', { detail: { key: val } }));
      close();
      onSaved(val);
    });

    cancelBtn.addEventListener('click', () => {
      close();
      onCancel();
    });

    container.onclick = (e) => {
      if (e.target === container) {
        close();
        onCancel();
      }
    };
  }

  escapeHtml(str) {
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }
}

export const Modal = new ModalManager();
