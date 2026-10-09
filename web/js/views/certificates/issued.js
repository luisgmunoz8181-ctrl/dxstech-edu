import { Toast } from '../../components/toast.js';
import { esc } from '../../utils/escape.js';

// Certificados emitidos: listado, filtro y exportación CSV.

export const issuedMethods = {
  async loadMyCertificates() {
    const container = document.getElementById('my-certs-container');
    if (!container) return;
    try {
      const res = await fetch('/api/certificates/my-certificates');
      if (!res.ok) {
        container.innerHTML = `
          <div class="col-span-full py-12 text-center text-slate-400 bg-white rounded-3xl border border-slate-200">
            <i data-lucide="lock" class="w-10 h-10 mx-auto mb-2 text-slate-300"></i>
            <h4 class="text-sm font-bold text-slate-700">Inicia sesión</h4>
            <p class="text-xs text-slate-400 mt-1">Debes iniciar sesión para consultar tus certificados oficiales emitidos.</p>
          </div>
        `;
        if (window.lucide) window.lucide.createIcons();
        return;
      }

      const certs = await res.json();
      this.state.myCertificates = certs || [];

      if (!certs || certs.length === 0) {
        container.innerHTML = `
          <div class="col-span-full py-16 text-center text-slate-400 bg-white rounded-3xl border border-slate-200">
            <div class="w-16 h-16 rounded-3xl bg-amber-50 text-amber-500 flex items-center justify-center mx-auto mb-3 shadow-inner border border-amber-100">
              <i data-lucide="award" class="w-8 h-8"></i>
            </div>
            <h4 class="text-base font-bold text-slate-800">Aún no tienes certificados expedidos</h4>
            <p class="text-xs text-slate-500 mt-1 max-w-md mx-auto">
              Completa el 100% de las lecciones de cualquiera de nuestros cursos en el catálogo para que tu certificado oficial se genere de forma inmediata y automática con código QR.
            </p>
            <div class="mt-4">
              <a href="#courses" class="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-colors shadow-xs">
                <i data-lucide="book-open" class="w-4 h-4"></i>
                <span>Explorar Cursos Disponibles</span>
              </a>
            </div>
          </div>
        `;
        if (window.lucide) window.lucide.createIcons();
        return;
      }

      container.innerHTML = certs.map(c => `
        <div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs hover:shadow-lg transition-all p-6 flex flex-col justify-between group">
          <div class="space-y-4">
            <div class="flex items-start justify-between gap-3">
              <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-amber-400 to-amber-600 text-white flex items-center justify-center shadow-md shadow-amber-200">
                <i data-lucide="award" class="w-6 h-6"></i>
              </div>
              <span class="px-2.5 py-1 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">
                Oficial Verificado
              </span>
            </div>

            <div>
              <span class="text-[11px] font-mono font-bold text-slate-400">${esc(c.id)}</span>
              <h4 class="text-base font-bold text-slate-900 leading-snug mt-1 group-hover:text-indigo-600 transition-colors">
                ${esc(c.courseTitle)}
              </h4>
              <p class="text-xs text-slate-500 mt-1">Acreditado a: <strong class="text-slate-700 font-semibold">${esc(c.studentName)}</strong></p>
            </div>

            <div class="bg-slate-50 rounded-xl p-3 border border-slate-200/60 text-[11px] space-y-1.5 text-slate-600">
              <div class="flex justify-between">
                <span class="text-slate-400">Intensidad:</span>
                <span class="font-semibold text-slate-800">${c.durationHours || 10} Horas</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-400">Fecha de Emisión:</span>
                <span class="font-semibold text-slate-800">${esc(c.issueDate)}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-400">Docente / Director:</span>
                <span class="font-semibold text-indigo-600">${esc(c.instructorName || 'DxSTech Edu')}</span>
              </div>
            </div>
          </div>

          <div class="pt-4 border-t border-slate-100 flex flex-wrap items-center justify-between gap-2 mt-4">
            <a href="/api/certificates/${esc(c.id)}/pdf" target="_blank" download="Certificado_${esc(c.id)}.pdf" class="flex-1 py-2 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-colors shadow-xs flex items-center justify-center gap-1.5 cursor-pointer">
              <i data-lucide="download" class="w-3.5 h-3.5"></i>
              <span>Descargar PDF</span>
            </a>
            <a href="#verify/${esc(c.id)}" class="px-3 py-2 rounded-xl text-xs font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center gap-1 cursor-pointer" title="Verificar autenticidad">
              <i data-lucide="shield-check" class="w-3.5 h-3.5 text-emerald-600"></i>
            </a>
          </div>
        </div>
      `).join('');

      if (window.lucide) window.lucide.createIcons();
    } catch (e) {
      container.innerHTML = `<div class="p-6 text-xs text-rose-600 bg-rose-50 rounded-2xl col-span-full">Error cargando certificados: ${esc(e.message)}</div>`;
    }
  },

  async loadIssuedCertificates() {
    const container = document.getElementById('issued-certificates-container');
    if (!container) return;

    try {
      const res = await fetch('/api/certificates/issued');
      if (!res.ok) throw new Error('Error al consultar base de datos');
      const list = await res.json();
      this.state.issuedCertificates = list || [];
      this.renderIssuedTable(this.state.issuedCertificates);
    } catch (err) {
      container.innerHTML = `
        <div class="p-3 text-xs text-rose-600 bg-rose-50 rounded-xl">
          Error al cargar certificados emitidos: ${esc(err.message)}
        </div>
      `;
    }
  },

  filterIssuedCertificates(query) {
    if (!this.state.issuedCertificates) return;
    const q = (query || '').toLowerCase().trim();
    if (!q) {
      this.renderIssuedTable(this.state.issuedCertificates);
      return;
    }

    const filtered = this.state.issuedCertificates.filter(item => {
      const name = (item.studentName || '').toLowerCase();
      const id = (item.id || '').toLowerCase();
      const course = (item.courseTitle || '').toLowerCase();
      return name.includes(q) || id.includes(q) || course.includes(q);
    });

    this.renderIssuedTable(filtered, true);
  },

  renderIssuedTable(list, isFiltered = false) {
    const container = document.getElementById('issued-certificates-container');
    if (!container) return;

    if (!list || list.length === 0) {
      container.innerHTML = `
        <div class="py-8 text-center bg-slate-50/50 rounded-xl border border-dashed border-slate-200">
          <i data-lucide="award" class="w-8 h-8 text-slate-300 mx-auto mb-2"></i>
          <p class="text-xs font-semibold text-slate-600">${isFiltered ? 'No se encontraron certificados coincidentes' : 'Aún no se han emitido certificados con QR'}</p>
          <p class="text-[11px] text-slate-400 mt-0.5">${isFiltered ? 'Prueba con otro término de búsqueda o limpia el filtro.' : 'Al generar certificados con el código QR activo, quedarán registrados aquí para auditoría.'}</p>
        </div>
      `;
      if (window.lucide) window.lucide.createIcons();
      return;
    }

    let html = `
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs text-slate-700">
          <thead class="bg-slate-50 text-[10px] uppercase tracking-wider text-slate-500 font-semibold border-b border-slate-200">
            <tr>
              <th class="py-2.5 px-3">Código Oficial</th>
              <th class="py-2.5 px-3">Estudiante</th>
              <th class="py-2.5 px-3">Programa / Diplomado</th>
              <th class="py-2.5 px-3">Fecha de Emisión</th>
              <th class="py-2.5 px-3 text-right">Acciones</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
    `;

    list.slice(0, 50).forEach(item => {
      const verifyHash = `#verify/${esc(item.id)}`;
      html += `
        <tr class="hover:bg-slate-50/80 transition-colors">
          <td class="py-2.5 px-3 font-mono font-bold text-indigo-700">
            <span class="bg-indigo-50 border border-indigo-200/60 px-2 py-0.5 rounded text-[11px]">
              ${esc(item.id)}
            </span>
          </td>
          <td class="py-2.5 px-3 font-semibold text-slate-900">${esc(item.studentName)}</td>
          <td class="py-2.5 px-3 text-slate-600 text-[11px]">${esc(item.courseTitle || 'Diplomado')}</td>
          <td class="py-2.5 px-3 text-slate-500 text-[11px]">${esc(item.issueDate)}</td>
          <td class="py-2.5 px-3 text-right space-x-1">
            <a href="${esc(verifyHash)}" class="inline-flex items-center gap-1 px-2.5 py-1 text-[11px] font-bold text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg shadow-xs transition-colors">
              <i data-lucide="shield-check" class="w-3 h-3"></i>
              Verificar
            </a>
            <button data-copy-link="${esc(item.id)}" class="inline-flex items-center gap-1 px-2 py-1 text-[11px] font-semibold text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors">
              <i data-lucide="copy" class="w-3 h-3"></i>
              Copiar
            </button>
          </td>
        </tr>
      `;
    });

    html += `
          </tbody>
        </table>
      </div>
    `;

    container.innerHTML = html;
    if (window.lucide) window.lucide.createIcons();

    // Bind copy buttons
    container.querySelectorAll('[data-copy-link]').forEach(btn => {
      btn.addEventListener('click', () => {
        const certId = btn.dataset.copyLink;
        const fullUrl = `${window.location.origin}/#verify/${certId}`;
        navigator.clipboard.writeText(fullUrl).then(() => {
          Toast.success(`Enlace copiado al portapapeles: ${certId}`);
        }).catch(() => {
          Toast.info(`Enlace: ${fullUrl}`);
        });
      });
    });
  },

  exportIssuedToCSV() {
    const list = this.state.issuedCertificates || [];
    if (list.length === 0) {
      Toast.warning('No hay certificados emitidos para exportar.');
      return;
    }

    let csv = 'Codigo,Estudiante,Curso,Fecha_Emision,URL_Verificacion\r\n';
    list.forEach(c => {
      const code = `"${(c.id || '').replace(/"/g, '""')}"`;
      const name = `"${(c.studentName || '').replace(/"/g, '""')}"`;
      const course = `"${(c.courseTitle || '').replace(/"/g, '""')}"`;
      const date = `"${(c.issueDate || '').replace(/"/g, '""')}"`;
      const url = `"${window.location.origin}/#verify/${c.id}"`;
      csv += `${esc(code)},${esc(name)},${course},${esc(date)},${url}\r\n`;
    });

    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const u = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = u;
    a.download = `certificados_emitidos_dxstech_${new Date().toISOString().slice(0, 10)}.csv`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(u);
    Toast.success('Listado de certificados emitidos descargado en CSV.');
  },
};
