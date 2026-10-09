import { Toast } from '../components/toast.js';
import { esc } from '../utils/escape.js';

export const ProfileView = {
  render() {
    const user = window.router?.currentUser || {
      firstName: 'Usuario',
      lastName: '',
      email: 'cargando...',
      role: 'ESTUDIANTE',
    };

    const roleBadgeColor = {
      SUPERADMIN: 'bg-purple-100 text-purple-800 border-purple-200',
      ADMINISTRADOR: 'bg-indigo-100 text-indigo-800 border-indigo-200',
      ESTUDIANTE: 'bg-emerald-100 text-emerald-800 border-emerald-200',
    }[user.role] || 'bg-slate-100 text-slate-800 border-slate-200';

    return `
      <div class="max-w-4xl mx-auto space-y-6">
        <!-- Profile Header Card -->
        <div class="bg-white rounded-3xl p-6 md:p-8 border border-slate-200/80 shadow-sm flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
          <div class="flex items-center gap-4">
            <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-indigo-500 to-indigo-700 flex items-center justify-center text-white text-2xl font-bold shadow-lg shadow-indigo-100">
              ${esc(user.firstName.charAt(0))}${esc(user.lastName.charAt(0))}
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-xl font-bold text-slate-900">${esc(user.firstName)} ${esc(user.lastName)}</h3>
                <span class="px-2.5 py-0.5 rounded-full text-[11px] font-bold border ${roleBadgeColor}">
                  ${esc(user.role)}
                </span>
              </div>
              <p class="text-xs text-slate-500 mt-0.5">${esc(user.email)}</p>
              <p class="text-[11px] text-slate-400 mt-1">
                Estado: <span class="font-semibold text-emerald-600">Activo</span> • ID: <code class="font-mono text-[10px] bg-slate-100 px-1 py-0.5 rounded">${esc(user.id || 'N/A')}</code>
              </p>
            </div>
          </div>

          <button id="profile-logout-btn" class="px-4 py-2 rounded-xl text-xs font-bold text-rose-700 bg-rose-50 hover:bg-rose-100 border border-rose-200 transition-colors flex items-center gap-1.5">
            <i data-lucide="log-out" class="w-4 h-4"></i>
            <span>Cerrar Sesión</span>
          </button>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Account Details -->
          <div class="bg-white rounded-3xl p-6 border border-slate-200/80 shadow-sm space-y-4">
            <div class="flex items-center gap-2 pb-3 border-b border-slate-100">
              <i data-lucide="user-check" class="w-4 h-4 text-indigo-600"></i>
              <h4 class="text-sm font-bold text-slate-800">Datos Personales & Institucionales</h4>
            </div>

            <div class="space-y-3 text-xs">
              <div class="flex justify-between py-1.5 border-b border-slate-50">
                <span class="text-slate-400">Identificación / DNI:</span>
                <span class="font-semibold text-slate-700">${esc(user.identification || 'No registrada')}</span>
              </div>
              <div class="flex justify-between py-1.5 border-b border-slate-50">
                <span class="text-slate-400">Empresa / Organización:</span>
                <span class="font-semibold text-slate-700">${esc(user.company || 'DxSTech Academy')}</span>
              </div>
              <div class="flex justify-between py-1.5 border-b border-slate-50">
                <span class="text-slate-400">Cargo / Posición:</span>
                <span class="font-semibold text-slate-700">${esc(user.jobTitle || 'Miembro Activo')}</span>
              </div>
              <div class="flex justify-between py-1.5">
                <span class="text-slate-400">Último Inicio de Sesión:</span>
                <span class="font-semibold text-slate-700">${user.lastLogin ? new Date(user.lastLogin).toLocaleString() : 'Sesión Actual'}</span>
              </div>
            </div>
          </div>

          <!-- Change Password Form -->
          <div class="bg-white rounded-3xl p-6 border border-slate-200/80 shadow-sm space-y-4">
            <div class="flex items-center gap-2 pb-3 border-b border-slate-100">
              <i data-lucide="shield" class="w-4 h-4 text-indigo-600"></i>
              <h4 class="text-sm font-bold text-slate-800">Actualizar Contraseña</h4>
            </div>

            <form id="change-password-form" class="space-y-3">
              <div>
                <label class="block text-[11px] font-semibold text-slate-600 mb-1">Contraseña Actual</label>
                <input type="password" id="current-pass" required placeholder="••••••••" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
              </div>

              <div>
                <label class="block text-[11px] font-semibold text-slate-600 mb-1">Nueva Contraseña</label>
                <input type="password" id="new-pass" required minlength="8" placeholder="Mínimo 8 caracteres (Mayús, minús, número)" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
              </div>

              <div>
                <label class="block text-[11px] font-semibold text-slate-600 mb-1">Confirmar Nueva Contraseña</label>
                <input type="password" id="confirm-pass" required minlength="8" placeholder="Repite la nueva contraseña" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
              </div>

              <button type="submit" class="w-full py-2.5 px-4 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-all shadow-md shadow-indigo-100 flex items-center justify-center gap-1.5 mt-2">
                <i data-lucide="save" class="w-3.5 h-3.5"></i>
                <span>Guardar Nueva Contraseña</span>
              </button>
            </form>
          </div>
        </div>
      </div>
    `;
  },

  mount() {
    // Logout Handler
    const logoutBtn = document.getElementById('profile-logout-btn');
    if (logoutBtn) {
      logoutBtn.addEventListener('click', async () => {
        try {
          await fetch('/api/auth/logout', { method: 'POST' });
          Toast.info('Sesión cerrada correctamente');
          if (window.router) {
            window.router.currentUser = null;
            window.router.updateSessionUI();
            window.location.hash = 'login';
          }
        } catch (e) {
          Toast.error(e.message);
        }
      });
    }

    // Change Password Form
    const changeForm = document.getElementById('change-password-form');
    if (changeForm) {
      changeForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const currentPassword = document.getElementById('current-pass').value;
        const newPassword = document.getElementById('new-pass').value;
        const confirmPassword = document.getElementById('confirm-pass').value;

        if (newPassword !== confirmPassword) {
          Toast.warning('La nueva contraseña y su confirmación no coinciden');
          return;
        }

        try {
          const res = await fetch('/api/auth/change-password', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ currentPassword, newPassword }),
          });

          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          Toast.success('Contraseña actualizada con éxito');
          changeForm.reset();
          if (window.router?.currentUser) window.router.currentUser.mustChangePassword = false;
        } catch (err) {
          Toast.error(err.message);
        }
      });
    }

    if (window.lucide) window.lucide.createIcons();
  }
};
