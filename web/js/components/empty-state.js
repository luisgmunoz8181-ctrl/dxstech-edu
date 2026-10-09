import { esc } from '../utils/escape.js';
// Empty state component

export const EmptyState = {
  render({
    icon = 'inbox',
    title = 'No hay elementos disponibles',
    description = 'No se han registrado datos todavía.',
    actionText = '',
    actionId = '',
  } = {}) {
    return `
      <div class="flex flex-col items-center justify-center p-12 text-center rounded-2xl border-2 border-dashed border-slate-200/80 bg-white/50 my-4">
        <div class="w-12 h-12 rounded-2xl bg-slate-100 text-slate-400 flex items-center justify-center mb-3">
          <i data-lucide="${icon}" class="w-6 h-6"></i>
        </div>
        <h4 class="text-sm font-bold text-slate-700 mb-1">${esc(title)}</h4>
        <p class="text-xs text-slate-400 max-w-sm mb-4 leading-relaxed">${esc(description)}</p>
        ${actionText ? `
          <button id="${esc(actionId)}" class="px-4 py-2 rounded-xl text-xs font-semibold text-indigo-600 bg-indigo-50 hover:bg-indigo-100 transition-colors">
            ${actionText}
          </button>
        ` : ''}
      </div>
    `;
  }
};
