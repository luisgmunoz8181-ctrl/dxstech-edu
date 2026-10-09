import { esc } from '../utils/escape.js';
// Loading and Skeleton component

export const Loading = {
  spinner(text = 'Cargando...') {
    return `
      <div class="flex flex-col items-center justify-center p-12 text-slate-500 gap-3">
        <div class="w-8 h-8 border-3 border-indigo-600/30 border-t-indigo-600 rounded-full animate-spin"></div>
        <span class="text-xs font-medium">${esc(text)}</span>
      </div>
    `;
  },

  overlay(text = 'Procesando...') {
    const existing = document.getElementById('global-loading-overlay');
    if (existing) existing.remove();

    const overlay = document.createElement('div');
    overlay.id = 'global-loading-overlay';
    overlay.className = 'fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm flex items-center justify-center p-4 transition-opacity';
    overlay.innerHTML = `
      <div class="bg-white rounded-2xl shadow-2xl p-6 flex items-center gap-4 border border-slate-100 max-w-sm w-full">
        <div class="w-7 h-7 border-3 border-indigo-600/30 border-t-indigo-600 rounded-full animate-spin shrink-0"></div>
        <div>
          <p class="text-xs font-bold text-slate-800">${esc(text)}</p>
          <p class="text-[11px] text-slate-400">Por favor, espera un momento...</p>
        </div>
      </div>
    `;
    document.body.appendChild(overlay);
  },

  hideOverlay() {
    const overlay = document.getElementById('global-loading-overlay');
    if (overlay) overlay.remove();
  },

  // Alias usados por las vistas (courses.js, etc.): antes no existían y lanzaban TypeError.
  show(text) {
    this.overlay(text);
  },

  hide() {
    this.hideOverlay();
  },

  skeletonCard() {
    return `
      <div class="animate-pulse bg-white rounded-2xl p-6 border border-slate-200/80 shadow-sm space-y-4">
        <div class="h-4 bg-slate-200 rounded w-1/3"></div>
        <div class="space-y-2">
          <div class="h-3 bg-slate-100 rounded w-full"></div>
          <div class="h-3 bg-slate-100 rounded w-5/6"></div>
        </div>
        <div class="h-8 bg-slate-100 rounded-xl w-1/4"></div>
      </div>
    `;
  }
};
