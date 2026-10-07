// Toast notification component

class ToastManager {
  constructor() {
    this.container = document.getElementById('toast-container');
  }

  show(message, type = 'info', duration = 4000) {
    if (!this.container) {
      this.container = document.getElementById('toast-container');
      if (!this.container) return;
    }

    const toast = document.createElement('div');
    toast.className = `pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-xl shadow-lg border text-sm font-medium transition-all duration-300 transform translate-x-8 opacity-0 max-w-md ${this.getTypeStyles(type)}`;

    const iconName = this.getTypeIcon(type);
    toast.innerHTML = `
      <i data-lucide="${iconName}" class="w-4 h-4 shrink-0"></i>
      <span class="flex-1 text-xs leading-relaxed">${this.escapeHtml(message)}</span>
      <button class="text-current opacity-60 hover:opacity-100 transition-opacity p-0.5 rounded">
        <i data-lucide="x" class="w-3.5 h-3.5"></i>
      </button>
    `;

    this.container.appendChild(toast);
    if (window.lucide) window.lucide.createIcons();

    // Trigger enter animation
    requestAnimationFrame(() => {
      toast.classList.remove('translate-x-8', 'opacity-0');
      toast.classList.add('translate-x-0', 'opacity-100');
    });

    const closeBtn = toast.querySelector('button');
    closeBtn.addEventListener('click', () => this.dismiss(toast));

    if (duration > 0) {
      setTimeout(() => this.dismiss(toast), duration);
    }
  }

  dismiss(toast) {
    toast.classList.add('opacity-0', 'translate-x-8');
    setTimeout(() => {
      if (toast.parentElement) toast.parentElement.removeChild(toast);
    }, 300);
  }

  success(msg, duration) {
    this.show(msg, 'success', duration);
  }

  error(msg, duration) {
    this.show(msg, 'error', duration);
  }

  warning(msg, duration) {
    this.show(msg, 'warning', duration);
  }

  info(msg, duration) {
    this.show(msg, 'info', duration);
  }

  getTypeStyles(type) {
    switch (type) {
      case 'success':
        return 'bg-emerald-50 border-emerald-200 text-emerald-800';
      case 'error':
        return 'bg-rose-50 border-rose-200 text-rose-800';
      case 'warning':
        return 'bg-amber-50 border-amber-200 text-amber-800';
      default:
        return 'bg-slate-900 border-slate-800 text-white';
    }
  }

  getTypeIcon(type) {
    switch (type) {
      case 'success':
        return 'check-circle-2';
      case 'error':
        return 'alert-circle';
      case 'warning':
        return 'alert-triangle';
      default:
        return 'info';
    }
  }

  escapeHtml(str) {
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }
}

export const Toast = new ToastManager();
