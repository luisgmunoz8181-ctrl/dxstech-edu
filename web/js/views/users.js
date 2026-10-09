import { Toast } from '../components/toast.js';
import { Modal } from '../components/modal.js';
import { Loading } from '../components/loading.js';
import { esc } from '../utils/escape.js';

export const UsersView = {
  users: [],
  search: '',
  roleFilter: 0,

  render() {
    return `
      <div class="space-y-6 max-w-6xl mx-auto">
        <!-- Header Banner -->
        <div class="bg-gradient-to-r from-slate-900 to-indigo-950 rounded-3xl p-6 text-white shadow-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <div class="flex items-center gap-2">
              <span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-indigo-500/30 text-indigo-200 border border-indigo-400/30 uppercase tracking-wider">
                Módulo Administrativo
              </span>
            </div>
            <h3 class="text-xl font-bold mt-1">Gestión Integral de Usuarios</h3>
            <p class="text-xs text-slate-300 mt-1 max-w-xl">
              Crea cuentas, asigna roles basados en permisos (RBAC) y administra el acceso de administradores y estudiantes.
            </p>
          </div>

          <button id="create-user-btn" class="px-4 py-2.5 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-500 transition-all shadow-md shadow-indigo-900/50 flex items-center gap-2 shrink-0">
            <i data-lucide="user-plus" class="w-4 h-4"></i>
            <span>Nuevo Usuario</span>
          </button>
        </div>

        <!-- Filter & Search Bar -->
        <div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-xs flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="relative w-full sm:w-72">
            <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
              <i data-lucide="search" class="w-4 h-4"></i>
            </span>
            <input type="text" id="user-search-input" value="${esc(this.search)}" placeholder="Buscar por nombre, correo..." class="w-full text-xs rounded-xl border border-slate-200 pl-9 pr-3.5 py-2 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
          </div>

          <div class="flex items-center gap-2 w-full sm:w-auto">
            <label class="text-xs font-semibold text-slate-500 whitespace-nowrap">Filtrar por rol:</label>
            <select id="user-role-filter" class="text-xs rounded-xl border border-slate-200 px-3 py-2 bg-slate-50 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
              <option value="0" ${this.roleFilter === 0 ? 'selected' : ''}>Todos los roles</option>
              <option value="1" ${this.roleFilter === 1 ? 'selected' : ''}>Superadmin</option>
              <option value="2" ${this.roleFilter === 2 ? 'selected' : ''}>Administrador</option>
              <option value="3" ${this.roleFilter === 3 ? 'selected' : ''}>Estudiante</option>
            </select>
          </div>
        </div>

        <!-- Users Table Container -->
        <div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-50 border-b border-slate-200/80 text-slate-500 uppercase tracking-wider text-[10px] font-semibold">
                <tr>
                  <th class="py-3 px-5">Usuario</th>
                  <th class="py-3 px-4">Rol RBAC</th>
                  <th class="py-3 px-4">Estado</th>
                  <th class="py-3 px-4">Empresa / Cargo</th>
                  <th class="py-3 px-4">Último Acceso</th>
                  <th class="py-3 px-5 text-right">Acciones</th>
                </tr>
              </thead>
              <tbody id="users-tbody" class="divide-y divide-slate-100">
                <tr>
                  <td colspan="6" class="py-12 text-center text-slate-400">Cargando directorio de usuarios...</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    `;
  },

  async mount() {
    this.bindEvents();
    await this.fetchUsers();
    if (window.lucide) window.lucide.createIcons();
  },

  bindEvents() {
    // Search Input
    const searchInput = document.getElementById('user-search-input');
    if (searchInput) {
      let timeout;
      searchInput.addEventListener('input', (e) => {
        clearTimeout(timeout);
        this.search = e.target.value;
        timeout = setTimeout(() => this.fetchUsers(), 300);
      });
    }

    // Role Filter
    const roleSelect = document.getElementById('user-role-filter');
    if (roleSelect) {
      roleSelect.addEventListener('change', (e) => {
        this.roleFilter = parseInt(e.target.value, 10);
        this.fetchUsers();
      });
    }

    // Create User Button
    const createBtn = document.getElementById('create-user-btn');
    if (createBtn) {
      createBtn.addEventListener('click', () => this.showCreateUserModal());
    }
  },

  async fetchUsers() {
    try {
      const params = new URLSearchParams();
      if (this.roleFilter > 0) params.append('roleId', this.roleFilter);
      if (this.search) params.append('search', this.search);

      const res = await fetch(`/api/auth/users?${params.toString()}`);
      if (!res.ok) throw new Error('Error al consultar usuarios');

      this.users = await res.json();
      this.renderTableRows();
    } catch (err) {
      Toast.error(err.message);
    }
  },

  renderTableRows() {
    const tbody = document.getElementById('users-tbody');
    if (!tbody) return;

    if (!this.users || this.users.length === 0) {
      tbody.innerHTML = `
        <tr>
          <td colspan="6" class="py-12 text-center text-slate-400">
            <div class="flex flex-col items-center gap-2">
              <i data-lucide="users" class="w-8 h-8 text-slate-300"></i>
              <p>No se encontraron usuarios con los filtros aplicados</p>
            </div>
          </td>
        </tr>
      `;
      if (window.lucide) window.lucide.createIcons();
      return;
    }

    const rows = this.users.map(u => {
      const roleBadge = {
        SUPERADMIN: '<span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-purple-50 text-purple-700 border border-purple-200">Superadmin</span>',
        ADMINISTRADOR: '<span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-indigo-50 text-indigo-700 border border-indigo-200">Administrador</span>',
        ESTUDIANTE: '<span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">Estudiante</span>',
      }[u.role] || `<span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-slate-100 text-slate-700">${esc(u.role)}</span>`;

      const statusBadge = u.status === 'active'
        ? '<span class="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-700"><span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>Activo</span>'
        : '<span class="inline-flex items-center gap-1 text-[11px] font-semibold text-rose-700"><span class="w-1.5 h-1.5 rounded-full bg-rose-500"></span>Inactivo</span>';

      const lastLoginText = u.lastLogin ? new Date(u.lastLogin).toLocaleDateString() : 'Nunca';

      return `
        <tr class="hover:bg-slate-50/70 transition-colors">
          <td class="py-3 px-5">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-full bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center text-xs">
                ${esc(u.firstName.charAt(0))}${esc(u.lastName.charAt(0))}
              </div>
              <div>
                <p class="font-bold text-slate-800">${esc(u.firstName)} ${esc(u.lastName)}</p>
                <p class="text-[11px] text-slate-400">${esc(u.email)}</p>
              </div>
            </div>
          </td>
          <td class="py-3 px-4">${roleBadge}</td>
          <td class="py-3 px-4">${statusBadge}</td>
          <td class="py-3 px-4 text-slate-600">
            <p class="font-medium">${esc(u.company || '—')}</p>
            <p class="text-[10px] text-slate-400">${esc(u.jobTitle || '')}</p>
          </td>
          <td class="py-3 px-4 text-slate-500 text-[11px]">${lastLoginText}</td>
          <td class="py-3 px-5 text-right">
            <button data-toggle-id="${esc(u.id)}" class="px-2.5 py-1 rounded-lg text-[11px] font-semibold ${u.status === 'active' ? 'text-amber-700 bg-amber-50 hover:bg-amber-100' : 'text-emerald-700 bg-emerald-50 hover:bg-emerald-100'} transition-colors">
              ${u.status === 'active' ? 'Desactivar' : 'Activar'}
            </button>
          </td>
        </tr>
      `;
    }).join('');

    tbody.innerHTML = rows;

    // Bind Toggle Status Handlers
    tbody.querySelectorAll('[data-toggle-id]').forEach(btn => {
      btn.addEventListener('click', () => {
        const id = btn.dataset.toggleId;
        this.toggleUserStatus(id);
      });
    });

    if (window.lucide) window.lucide.createIcons();
  },

  async toggleUserStatus(id) {
    try {
      const res = await fetch(`/api/auth/users/${id}/status`, { method: 'PATCH' });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error);

      Toast.success(`Estado de usuario actualizado: ${data.status}`);
      await this.fetchUsers();
    } catch (e) {
      Toast.error(e.message);
    }
  },

  showCreateUserModal() {
    Modal.show({
      title: 'Crear Nuevo Usuario',
      confirmText: 'Crear Usuario',
      cancelText: 'Cancelar',
      showCancel: true,
      content: `
        <form id="modal-create-user-form" class="space-y-3 text-left">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-[11px] font-semibold text-slate-700 mb-1">Nombres *</label>
              <input type="text" id="new-user-first" required placeholder="Carlos" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2">
            </div>
            <div>
              <label class="block text-[11px] font-semibold text-slate-700 mb-1">Apellidos *</label>
              <input type="text" id="new-user-last" required placeholder="Mendoza" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2">
            </div>
          </div>

          <div>
            <label class="block text-[11px] font-semibold text-slate-700 mb-1">Correo Electrónico *</label>
            <input type="email" id="new-user-email" required placeholder="carlos@ejemplo.com" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2">
          </div>

          <div>
            <label class="block text-[11px] font-semibold text-slate-700 mb-1">Contraseña Inicial *</label>
            <input type="password" id="new-user-pass" required minlength="8" placeholder="Mínimo 8 caracteres (Mayús, minús, num)" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2">
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-[11px] font-semibold text-slate-700 mb-1">Rol en el LMS *</label>
              <select id="new-user-role" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2 bg-white">
                <option value="3">Estudiante</option>
                <option value="2">Administrador</option>
                <option value="1">Superadmin</option>
              </select>
            </div>
            <div>
              <label class="block text-[11px] font-semibold text-slate-700 mb-1">Identificación / DNI</label>
              <input type="text" id="new-user-iden" placeholder="1020304050" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2">
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-[11px] font-semibold text-slate-700 mb-1">Empresa / Institución</label>
              <input type="text" id="new-user-comp" placeholder="DxSTech Corp" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2">
            </div>
            <div>
              <label class="block text-[11px] font-semibold text-slate-700 mb-1">Cargo</label>
              <input type="text" id="new-user-job" placeholder="Analista de Datos" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2">
            </div>
          </div>
        </form>
      `,
      onConfirm: async () => {
        const first = document.getElementById('new-user-first').value.trim();
        const last = document.getElementById('new-user-last').value.trim();
        const email = document.getElementById('new-user-email').value.trim();
        const password = document.getElementById('new-user-pass').value;
        const roleId = parseInt(document.getElementById('new-user-role').value, 10);
        const identification = document.getElementById('new-user-iden').value.trim();
        const company = document.getElementById('new-user-comp').value.trim();
        const jobTitle = document.getElementById('new-user-job').value.trim();

        if (!first || !last || !email || !password) {
          Toast.warning('Complete todos los campos obligatorios');
          return;
        }

        try {
          const res = await fetch('/api/auth/users', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              firstName: first,
              lastName: last,
              email,
              password,
              roleId,
              identification,
              company,
              jobTitle,
            }),
          });

          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          Toast.success(`Usuario creado exitosamente: ${data.email}`);
          await this.fetchUsers();
        } catch (e) {
          Toast.error(e.message);
        }
      }
    });
  }
};
