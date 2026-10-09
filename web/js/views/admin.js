import { Toast } from '../components/toast.js';
import { Loading } from '../components/loading.js';
import { esc } from '../utils/escape.js';

// Administración: registro de auditoría, moderación del foro y copias de seguridad.
const AUDIT_ACTIONS = [
  ['', 'Todas las acciones'],
  ['LOGIN_SUCCESS', 'Ingresos'], ['LOGIN_FAILED', 'Ingresos fallidos'], ['ACCOUNT_LOCKED', 'Cuentas bloqueadas'], ['UNLOCK_USER', 'Desbloqueos'],
  ['CREATE_USER', 'Usuarios creados'], ['UPDATE_USER', 'Usuarios modificados'], ['TOGGLE_USER_STATUS', 'Activar/desactivar usuario'],
  ['COURSE_CREATE', 'Cursos creados'], ['COURSE_UPDATE', 'Cursos modificados'], ['COURSE_DELETE', 'Cursos eliminados'], ['COURSE_STATUS_CHANGE', 'Cambios de estado de curso'],
  ['LESSON_DELETE', 'Lecciones eliminadas'], ['MODULE_DELETE', 'Módulos eliminados'], ['FILE_UPLOAD', 'Archivos subidos'], ['UPLOADS_CLEANUP', 'Limpieza de archivos'],
  ['CERTIFICATE_ISSUED', 'Certificados emitidos'], ['CERTIFICATES_GENERATE_BULK', 'Certificados masivos'],
  ['QUIZ_GENERATE', 'Evaluaciones generadas'], ['QUIZ_DELETE', 'Evaluaciones eliminadas'],
  ['ENROLLMENT_ADMIN_ENROLL', 'Matrículas administrativas'], ['ENROLLMENT_ADMIN_UNENROLL', 'Bajas administrativas'],
  ['WHATSAPP_BULK_SEND', 'Envíos de WhatsApp'], ['FORUM_POST_DELETE', 'Publicaciones eliminadas'], ['FORUM_POST_REPORT', 'Reportes del foro'], ['FORUM_REPORT_RESOLVE', 'Moderación del foro'],
  ['BACKUP_CREATE', 'Copias creadas'], ['BACKUP_DOWNLOAD', 'Copias descargadas'],
];
const PAGE_SIZE = 25;

function formatBytes(n) {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / 1024 / 1024).toFixed(1)} MB`;
}

export const AdminView = {
  tab: 'audit',
  audit: { action: '', search: '', offset: 0, rows: [] },

  get isSuperadmin() {
    return window.router?.currentUser?.role === 'SUPERADMIN';
  },

  render() {
    const tabBtn = (id, label, icon) => `
      <button id="adm-tab-${id}" data-tab="${id}" class="px-4 py-2.5 text-xs font-bold border-b-2 flex items-center gap-2 transition-colors ${this.tab === id ? 'border-indigo-600 text-indigo-700' : 'border-transparent text-slate-500 hover:text-slate-700'}">
        <i data-lucide="${icon}" class="w-4 h-4"></i>${label}
      </button>`;
    return `
      <div class="max-w-7xl mx-auto space-y-5">
        <div class="flex border-b border-slate-200 bg-white rounded-t-2xl px-2">
          ${tabBtn('audit', 'Auditoría', 'clipboard-list')}
          ${tabBtn('moderation', 'Moderación del foro', 'flag')}
          ${this.isSuperadmin ? tabBtn('backups', 'Copias de seguridad', 'database-backup') : ''}
        </div>
        <div id="adm-content"></div>
      </div>`;
  },

  mount() {
    document.querySelectorAll('[data-tab]').forEach((btn) => {
      btn.addEventListener('click', () => this.switchTab(btn.dataset.tab));
    });
    this.switchTab(this.tab);
  },

  switchTab(tab) {
    this.tab = tab;
    document.querySelectorAll('[data-tab]').forEach((btn) => {
      const active = btn.dataset.tab === tab;
      btn.className = `px-4 py-2.5 text-xs font-bold border-b-2 flex items-center gap-2 transition-colors ${active ? 'border-indigo-600 text-indigo-700' : 'border-transparent text-slate-500 hover:text-slate-700'}`;
    });
    if (tab === 'audit') this.loadAudit();
    else if (tab === 'moderation') this.loadModeration();
    else if (tab === 'backups') this.loadBackups();
  },

  // ------------------------------------------------------------ Auditoría
  async loadAudit() {
    const c = document.getElementById('adm-content');
    if (!c) return;
    const { action, search, offset } = this.audit;
    c.innerHTML = Loading.spinner('Consultando auditoría...');
    try {
      const qs = new URLSearchParams({ limit: PAGE_SIZE, offset });
      if (action) qs.set('action', action);
      if (search) qs.set('search', search);
      const res = await fetch(`/api/admin/audit?${qs}`);
      if (!res.ok) throw new Error((await res.json()).error || 'No se pudo consultar la auditoría');
      this.audit.rows = await res.json();
      this.renderAudit();
    } catch (e) {
      c.innerHTML = `<p class="p-6 text-xs text-rose-600">${esc(e.message)}</p>`;
    }
  },

  renderAudit() {
    const c = document.getElementById('adm-content');
    const { action, search, offset, rows } = this.audit;
    c.innerHTML = `
      <div class="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-4 space-y-4">
        <div class="flex flex-wrap items-end gap-3">
          <div>
            <label class="block text-[11px] font-semibold text-slate-500 mb-1">Acción</label>
            <select id="audit-action" class="text-xs rounded-xl border border-slate-200 px-3 py-2 bg-white">
              ${AUDIT_ACTIONS.map(([v, l]) => `<option value="${esc(v)}" ${v === action ? 'selected' : ''}>${esc(l)}</option>`).join('')}
            </select>
          </div>
          <div class="flex-1 min-w-48">
            <label class="block text-[11px] font-semibold text-slate-500 mb-1">Buscar en el detalle</label>
            <input id="audit-search" type="text" value="${esc(search)}" placeholder="Ej. nombre de curso, correo, código..." class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2">
          </div>
          <button id="audit-apply" class="px-4 py-2 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700">Filtrar</button>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="text-[11px] uppercase text-slate-400 border-b border-slate-100">
              <tr><th class="py-2 pr-3">Fecha</th><th class="py-2 pr-3">Usuario</th><th class="py-2 pr-3">Acción</th><th class="py-2 pr-3">Detalle</th><th class="py-2 pr-3">IP</th><th class="py-2">Petición</th></tr>
            </thead>
            <tbody id="audit-rows" class="divide-y divide-slate-100">
              ${rows.length === 0 ? '<tr><td colspan="6" class="py-8 text-center text-slate-400">Sin eventos para este filtro.</td></tr>' : rows.map((r) => `
                <tr class="align-top">
                  <td class="py-2 pr-3 whitespace-nowrap text-slate-500">${esc(new Date(r.createdAt).toLocaleString())}</td>
                  <td class="py-2 pr-3">${esc(r.userName || r.userId || '—')}</td>
                  <td class="py-2 pr-3"><span class="font-mono text-[10px] bg-slate-100 px-1.5 py-0.5 rounded">${esc(r.action)}</span></td>
                  <td class="py-2 pr-3 text-slate-700">${esc(r.details)}</td>
                  <td class="py-2 pr-3 font-mono text-[10px] text-slate-400">${esc(r.ipAddress)}</td>
                  <td class="py-2 font-mono text-[10px] text-slate-400" title="${esc(r.requestId)}">${esc((r.requestId || '').slice(0, 8))}</td>
                </tr>`).join('')}
            </tbody>
          </table>
        </div>

        <div class="flex items-center justify-between text-[11px] text-slate-500">
          <span>Mostrando ${rows.length ? offset + 1 : 0}–${offset + rows.length}</span>
          <div class="flex gap-2">
            <button id="audit-prev" ${offset === 0 ? 'disabled' : ''} class="px-3 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 disabled:opacity-40 font-semibold">← Anteriores</button>
            <button id="audit-next" ${rows.length < PAGE_SIZE ? 'disabled' : ''} class="px-3 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 disabled:opacity-40 font-semibold">Siguientes →</button>
          </div>
        </div>
      </div>`;

    const apply = () => {
      this.audit.action = document.getElementById('audit-action').value;
      this.audit.search = document.getElementById('audit-search').value.trim();
      this.audit.offset = 0;
      this.loadAudit();
    };
    document.getElementById('audit-apply').addEventListener('click', apply);
    document.getElementById('audit-search').addEventListener('keydown', (e) => { if (e.key === 'Enter') apply(); });
    document.getElementById('audit-prev').addEventListener('click', () => { this.audit.offset = Math.max(0, offset - PAGE_SIZE); this.loadAudit(); });
    document.getElementById('audit-next').addEventListener('click', () => { this.audit.offset = offset + PAGE_SIZE; this.loadAudit(); });
    if (window.lucide) window.lucide.createIcons();
  },

  // ------------------------------------------------------------ Moderación
  async loadModeration(status = 'open') {
    const c = document.getElementById('adm-content');
    if (!c) return;
    c.innerHTML = Loading.spinner('Consultando reportes...');
    try {
      const res = await fetch(`/api/admin/forum/reports?status=${encodeURIComponent(status)}`);
      if (!res.ok) throw new Error((await res.json()).error || 'No se pudieron consultar los reportes');
      const reports = await res.json();
      const label = { open: 'Pendiente', dismissed: 'Descartado', removed: 'Eliminada' };

      c.innerHTML = `
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <p class="text-xs text-slate-500">${reports.length} reporte(s) ${status === 'open' ? 'pendiente(s)' : 'en total'}</p>
            <select id="mod-status" class="text-xs rounded-xl border border-slate-200 px-3 py-2 bg-white">
              <option value="open" ${status === 'open' ? 'selected' : ''}>Solo pendientes</option>
              <option value="all" ${status === 'all' ? 'selected' : ''}>Todos (historial)</option>
            </select>
          </div>
          ${reports.length === 0 ? '<div class="bg-white rounded-2xl border border-slate-200/80 p-10 text-center text-xs text-slate-400">No hay reportes por revisar. 🎉</div>' : reports.map((r) => `
            <div class="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-4 space-y-2" data-report="${esc(r.id)}">
              <div class="flex flex-wrap items-center justify-between gap-2 text-[11px] text-slate-500">
                <span><strong class="text-slate-700">${esc(r.courseTitle || r.courseId)}</strong> · reportado por ${esc(r.reporterName || 'un usuario')} · ${esc(new Date(r.createdAt).toLocaleString())}</span>
                <span class="px-2 py-0.5 rounded-full font-bold ${r.status === 'open' ? 'bg-amber-100 text-amber-800' : 'bg-slate-100 text-slate-600'}">${esc(label[r.status] || r.status)}</span>
              </div>
              <p class="text-[11px] text-slate-500">Motivo: <em>${esc(r.reason)}</em></p>
              <div class="p-3 rounded-xl bg-slate-50 border border-slate-200/70 text-xs">
                <p class="font-bold text-slate-700 mb-1">${esc(r.postAuthorName || 'Autor')} ${r.postExists ? '' : '<span class="text-[10px] font-normal text-rose-600">(publicación eliminada)</span>'}</p>
                <p class="text-slate-600 whitespace-pre-line">${esc(r.postMessage)}</p>
              </div>
              ${r.status === 'open' ? `
                <div class="flex justify-end gap-2 pt-1">
                  <button data-resolve="dismiss" data-id="${esc(r.id)}" class="px-3 py-1.5 rounded-xl text-[11px] font-bold text-slate-600 bg-slate-100 hover:bg-slate-200">Descartar reporte</button>
                  <button data-resolve="remove" data-id="${esc(r.id)}" class="px-3 py-1.5 rounded-xl text-[11px] font-bold text-white bg-rose-600 hover:bg-rose-700">Eliminar publicación</button>
                </div>` : (r.resolvedBy ? `<p class="text-[10px] text-slate-400">Resuelto por ${esc(r.resolvedBy)}</p>` : '')}
            </div>`).join('')}
        </div>`;

      document.getElementById('mod-status').addEventListener('change', (e) => this.loadModeration(e.target.value));
      c.querySelectorAll('[data-resolve]').forEach((btn) => {
        btn.addEventListener('click', async () => {
          try {
            const r = await fetch(`/api/admin/forum/reports/${encodeURIComponent(btn.dataset.id)}/resolve`, {
              method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ action: btn.dataset.resolve }),
            });
            const data = await r.json();
            if (!r.ok) throw new Error(data.error);
            Toast.success(btn.dataset.resolve === 'remove' ? 'Publicación eliminada' : 'Reporte descartado');
            this.loadModeration(status);
          } catch (e) {
            Toast.error(e.message);
          }
        });
      });
    } catch (e) {
      c.innerHTML = `<p class="p-6 text-xs text-rose-600">${esc(e.message)}</p>`;
    }
  },

  // ------------------------------------------------------------ Copias de seguridad
  async loadBackups() {
    const c = document.getElementById('adm-content');
    if (!c) return;
    c.innerHTML = Loading.spinner('Consultando copias...');
    try {
      const res = await fetch('/api/admin/backups');
      if (!res.ok) throw new Error((await res.json()).error || 'No se pudieron listar las copias');
      const { backups, retention } = await res.json();

      c.innerHTML = `
        <div class="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-5 space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="text-sm font-bold text-slate-800">Copias de seguridad</h3>
              <p class="text-[11px] text-slate-500">Incluyen la base de datos y los archivos subidos. Se crean automáticamente y se conservan las últimas ${esc(retention)}.</p>
            </div>
            <button id="backup-create" class="px-4 py-2 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 flex items-center gap-2">
              <i data-lucide="database-backup" class="w-4 h-4"></i> Crear copia ahora
            </button>
          </div>

          <table class="w-full text-left text-xs">
            <thead class="text-[11px] uppercase text-slate-400 border-b border-slate-100"><tr><th class="py-2">Archivo</th><th class="py-2">Fecha (UTC)</th><th class="py-2">Tamaño</th><th></th></tr></thead>
            <tbody class="divide-y divide-slate-100">
              ${backups.length === 0 ? '<tr><td colspan="4" class="py-8 text-center text-slate-400">Aún no hay copias.</td></tr>' : backups.map((b) => `
                <tr>
                  <td class="py-2 font-mono text-[11px]">${esc(b.name)}</td>
                  <td class="py-2 text-slate-500">${esc(new Date(b.createdAt).toLocaleString())}</td>
                  <td class="py-2 text-slate-500">${esc(formatBytes(b.sizeBytes))}</td>
                  <td class="py-2 text-right"><a href="/api/admin/backups/${encodeURIComponent(b.name)}" download class="text-indigo-600 hover:text-indigo-800 font-bold">Descargar</a></td>
                </tr>`).join('')}
            </tbody>
          </table>

          <div class="p-3 rounded-xl bg-amber-50 border border-amber-200 text-[11px] text-amber-900 space-y-1">
            <p><strong>Importante:</strong> las copias se guardan en el mismo servidor. Descarga periódicamente una y guárdala en otro lugar (si el disco se pierde, se pierden también estas copias). Contienen datos personales y contraseñas cifradas: trátalas como información confidencial.</p>
            <p><strong>Restaurar:</strong> con el servidor detenido, ejecuta <code class="font-mono">dxstech-server -restore ruta/copia.tar.gz -yes</code>. Los datos actuales se conservan en una carpeta <code class="font-mono">pre-restore-…</code>.</p>
          </div>
        </div>`;

      document.getElementById('backup-create').addEventListener('click', async (e) => {
        e.currentTarget.disabled = true;
        Loading.show('Creando copia de seguridad...');
        try {
          const r = await fetch('/api/admin/backups', { method: 'POST' });
          const data = await r.json();
          if (!r.ok) throw new Error(data.error);
          Toast.success(`Copia creada: ${data.name}`);
        } catch (err) {
          Toast.error(err.message);
        } finally {
          Loading.hide();
          this.loadBackups();
        }
      });
      if (window.lucide) window.lucide.createIcons();
    } catch (e) {
      c.innerHTML = `<p class="p-6 text-xs text-rose-600">${esc(e.message)}</p>`;
    }
  },

  destroy() {},
};
