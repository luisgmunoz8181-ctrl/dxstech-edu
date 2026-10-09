// Pestañas: mis certificados, verificación y generador (plantillas HTML).

export const tabsMethods = {
  renderMyCertificatesTab() {
    return `
      <div class="space-y-6">
        <div class="bg-gradient-to-r from-amber-500 to-amber-700 rounded-3xl p-6 text-white shadow-lg shadow-amber-100 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <span class="inline-block px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-white/20 text-white mb-2">
              🎓 Acreditación Profesional Digital
            </span>
            <h3 class="text-lg font-bold">Mis Diplomas y Certificaciones Obtenidas</h3>
            <p class="text-xs text-amber-100 mt-1 max-w-xl">
              Aquí puedes consultar y descargar en formato PDF de alta fidelidad todos los certificados oficiales obtenidos tras culminar satisfactoriamente tus cursos en DxSTech Edu.
            </p>
          </div>
          <a href="#courses" class="px-4 py-2.5 rounded-xl text-xs font-bold bg-white text-amber-800 hover:bg-amber-50 transition-all shadow-xs flex items-center gap-1.5 shrink-0">
            <i data-lucide="book-open" class="w-4 h-4 text-amber-600"></i>
            <span>Ir a Mis Cursos</span>
          </a>
        </div>

        <div id="my-certs-container" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div class="col-span-full py-16 text-center text-slate-400">
            <div class="w-8 h-8 border-3 border-amber-600/30 border-t-amber-600 rounded-full animate-spin mx-auto mb-3"></div>
            <p class="text-xs font-medium">Consultando tus certificaciones oficiales...</p>
          </div>
        </div>
      </div>
    `;
  },

  renderVerifyTab() {
    return `
      <div class="max-w-2xl mx-auto py-6 space-y-6">
        <div class="bg-white rounded-3xl p-8 border border-slate-200 shadow-sm text-center space-y-6">
          <div class="w-16 h-16 rounded-3xl bg-emerald-50 text-emerald-600 flex items-center justify-center mx-auto shadow-inner border border-emerald-100">
            <i data-lucide="shield-check" class="w-8 h-8"></i>
          </div>

          <div>
            <h3 class="text-xl font-bold text-slate-900">Validación Pública de Certificados</h3>
            <p class="text-xs text-slate-500 mt-1 max-w-md mx-auto">
              Ingresa el código único de registro (ejemplo: <code class="font-mono font-bold text-slate-700 bg-slate-100 px-1.5 py-0.5 rounded">DXS-2026-XXXX</code>) para verificar en tiempo real la autenticidad académica del documento.
            </p>
          </div>

          <div class="flex items-center gap-2 max-w-md mx-auto">
            <input type="text" id="verify-code-input" placeholder="DXS-2026-XXXX..." class="flex-1 px-4 py-2.5 rounded-xl border border-slate-200 text-xs font-mono font-bold focus:ring-2 focus:ring-emerald-500 focus:outline-hidden text-center uppercase tracking-wider">
            <button id="verify-code-btn" class="px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-700 transition-all shadow-xs flex items-center gap-1.5 cursor-pointer">
              <i data-lucide="search" class="w-3.5 h-3.5"></i>
              <span>Validar</span>
            </button>
          </div>

          <div class="bg-slate-50 rounded-2xl p-4 border border-slate-200/80 text-left text-xs text-slate-600 space-y-2">
            <p class="font-semibold text-slate-800 flex items-center gap-1.5">
              <i data-lucide="info" class="w-4 h-4 text-emerald-600"></i>
              Sobre la Verificación Oficial:
            </p>
            <p class="text-[11px] text-slate-500 leading-relaxed">
              Cada certificado emitido por DxSTech Edu cuenta con una firma digital determinística y un código QR único indeleble. El sistema contrasta el registro directamente con la base de datos oficial para certificar estudiante, curso, intensidad horaria e instructor.
            </p>
          </div>
        </div>
      </div>
    `;
  },

  renderGeneratorTab() {
    return `
      <div class="space-y-6">
        <!-- Top Banner / Description -->
        <div class="bg-gradient-to-r from-indigo-600 to-indigo-800 rounded-2xl p-6 text-white shadow-lg shadow-indigo-100 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h3 class="text-lg font-bold">Generador Profesional de Certificados</h3>
            <p class="text-xs text-indigo-100 mt-1 max-w-xl">
              Carga tu plantilla en cualquier formato, ubica el nombre interactivamente con guías milimétricas y genera cientos de PDFs de alta fidelidad empaquetados en un archivo ZIP.
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <button id="use-sample-template-btn" class="flex items-center gap-1.5 text-xs bg-white text-indigo-700 px-3.5 py-2 rounded-xl font-bold hover:bg-indigo-50 transition-all shadow-sm">
              <i data-lucide="sparkles" class="w-4 h-4 text-indigo-600"></i>
              <span>Cargar Plantilla Demo</span>
            </button>
            <div class="flex items-center gap-2 text-xs bg-white/10 backdrop-blur-md px-3 py-2 rounded-xl border border-white/20">
              <i data-lucide="shield-check" class="w-4 h-4 text-emerald-300"></i>
              <span>WebP ~0.82</span>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <!-- Left Column: Visual Canvas Editor -->
          <div class="lg:col-span-7 space-y-5">
            <!-- Upload Area & Image Preview -->
            <div class="bg-white rounded-2xl p-5 border border-slate-200/80 shadow-sm space-y-4">
              <div class="flex items-center justify-between">
                <h4 class="text-sm font-bold text-slate-800 flex items-center gap-2">
                  <i data-lucide="image" class="w-4 h-4 text-indigo-600"></i>
                  1. Plantilla del Certificado
                </h4>
                <div id="image-stats" class="hidden text-[11px] text-slate-500 bg-slate-100 px-2.5 py-1 rounded-lg"></div>
              </div>

              <!-- Drag & Drop / Paste Target -->
              <div id="drop-zone" class="relative group border-2 border-dashed border-slate-300 hover:border-indigo-500 rounded-2xl p-6 text-center transition-all cursor-pointer bg-slate-50/50 hover:bg-indigo-50/30">
                <input type="file" id="file-input" accept="image/png,image/jpeg,image/webp,image/jpg" class="hidden">
                <div id="drop-zone-prompt" class="flex flex-col items-center gap-2">
                  <div class="w-12 h-12 rounded-xl bg-white shadow-sm border border-slate-200 flex items-center justify-center text-indigo-600 group-hover:scale-105 transition-transform">
                    <i data-lucide="upload-cloud" class="w-6 h-6"></i>
                  </div>
                  <div>
                    <p class="text-xs font-semibold text-slate-700">Arrastra tu plantilla aquí o <span class="text-indigo-600 underline">haz clic para explorar</span></p>
                    <p class="text-[11px] text-slate-400 mt-0.5">Soporta PNG, JPG o WebP • Puedes pegar directamente con <kbd class="px-1.5 py-0.5 bg-slate-200 rounded text-[10px] text-slate-600 font-mono">Ctrl + V</kbd></p>
                  </div>
                </div>

                <!-- Canvas Preview Container -->
                <div id="canvas-wrapper" class="hidden flex-col items-center mt-2">
                  <div class="relative inline-block border border-slate-300 rounded-xl overflow-hidden shadow-md max-w-full bg-slate-100">
                    <canvas id="cert-canvas" class="cursor-crosshair max-h-[380px] w-auto object-contain"></canvas>
                  </div>
                  <p class="text-[11px] text-slate-500 mt-2 flex items-center gap-1.5">
                    <i data-lucide="mouse-pointer-click" class="w-3.5 h-3.5 text-indigo-500"></i>
                    Haz clic o arrastra sobre el lienzo para posicionar el nombre
                  </p>
                </div>
              </div>

              <!-- Interactive Placement & Typography Controls -->
              <div id="canvas-controls" class="space-y-3 pt-3 border-t border-slate-100">
                <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
                  <!-- Posición X -->
                  <div>
                    <label class="block text-[11px] font-semibold text-slate-600 mb-1">Posición X (%)</label>
                    <input type="number" id="ctrl-x" min="0" max="100" step="0.5" value="50" class="w-full px-2.5 py-1.5 bg-slate-50 border border-slate-200 rounded-lg text-xs text-slate-800 focus:ring-1 focus:ring-indigo-500">
                  </div>
                  <!-- Posición Y -->
                  <div>
                    <label class="block text-[11px] font-semibold text-slate-600 mb-1">Posición Y (%)</label>
                    <input type="number" id="ctrl-y" min="0" max="100" step="0.5" value="52" class="w-full px-2.5 py-1.5 bg-slate-50 border border-slate-200 rounded-lg text-xs text-slate-800 focus:ring-1 focus:ring-indigo-500">
                  </div>
                  <!-- Tamaño Fuente -->
                  <div>
                    <label class="block text-[11px] font-semibold text-slate-600 mb-1">Tamaño Fuente</label>
                    <input type="number" id="ctrl-size" min="12" max="90" step="1" value="32" class="w-full px-2.5 py-1.5 bg-slate-50 border border-slate-200 rounded-lg text-xs text-slate-800 focus:ring-1 focus:ring-indigo-500">
                  </div>
                  <!-- Color Fuente -->
                  <div>
                    <label class="block text-[11px] font-semibold text-slate-600 mb-1">Color de Texto</label>
                    <div class="flex items-center gap-1.5">
                      <input type="color" id="ctrl-color" value="#1e293b" class="w-8 h-7 rounded border border-slate-200 cursor-pointer p-0.5 bg-white">
                      <span id="color-label" class="text-[11px] font-mono text-slate-500">#1e293b</span>
                    </div>
                  </div>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
                  <!-- Tipografía -->
                  <div>
                    <label class="block text-[11px] font-semibold text-slate-600 mb-1">Tipografía</label>
                    <select id="ctrl-font" class="w-full px-2.5 py-1.5 bg-slate-50 border border-slate-200 rounded-lg text-xs text-slate-800 focus:ring-1 focus:ring-indigo-500">
                      <option value="Helvetica" selected>Helvetica / Arial (Moderna)</option>
                      <option value="Times">Times New Roman (Clásica Formal)</option>
                      <option value="Courier">Courier (Monospaciada Técnica)</option>
                    </select>
                  </div>

                  <!-- Alineación -->
                  <div>
                    <label class="block text-[11px] font-semibold text-slate-600 mb-1">Alineación Horizontal</label>
                    <div class="grid grid-cols-3 gap-1 bg-slate-100 p-0.5 rounded-lg text-xs">
                      <button type="button" data-align="left" class="align-btn py-1 rounded-md text-slate-600 hover:text-slate-900 flex justify-center">
                        <i data-lucide="align-left" class="w-4 h-4"></i>
                      </button>
                      <button type="button" data-align="center" class="align-btn py-1 rounded-md bg-white shadow-xs font-bold text-indigo-600 flex justify-center">
                        <i data-lucide="align-center" class="w-4 h-4"></i>
                      </button>
                      <button type="button" data-align="right" class="align-btn py-1 rounded-md text-slate-600 hover:text-slate-900 flex justify-center">
                        <i data-lucide="align-right" class="w-4 h-4"></i>
                      </button>
                    </div>
                  </div>
                </div>

                <!-- QR Verification Settings -->
                <div class="pt-3 border-t border-slate-100 space-y-2.5">
                  <div class="flex items-center justify-between">
                    <label class="flex items-center gap-2 cursor-pointer">
                      <input type="checkbox" id="ctrl-include-qr" checked class="rounded text-indigo-600 focus:ring-indigo-500">
                      <span class="text-xs font-bold text-slate-800 flex items-center gap-1.5">
                        <i data-lucide="qr-code" class="w-4 h-4 text-indigo-600"></i>
                        Código QR de Verificación Oficial
                      </span>
                    </label>
                    <span class="text-[10px] text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full font-semibold border border-emerald-200">
                      Antifraude Activo
                    </span>
                  </div>

                  <div id="qr-settings-panel" class="space-y-2 bg-indigo-50/40 p-3 rounded-xl border border-indigo-100/80">
                    <div>
                      <label class="block text-[11px] font-semibold text-slate-600 mb-0.5">Nombre del Programa / Diplomado</label>
                      <input type="text" id="ctrl-course-title" value="Diplomado en Inteligencia Artificial & Desarrollo Web" class="w-full px-2.5 py-1.5 bg-white border border-slate-200 rounded-lg text-xs text-slate-800 focus:ring-1 focus:ring-indigo-500 font-medium">
                    </div>
                    <div class="grid grid-cols-3 gap-2">
                      <div>
                        <label class="block text-[10px] font-semibold text-slate-600 mb-0.5">Posición X (%)</label>
                        <input type="number" id="ctrl-qr-x" min="0" max="100" step="0.5" value="88" class="w-full px-2 py-1 bg-white border border-slate-200 rounded-lg text-xs text-slate-800">
                      </div>
                      <div>
                        <label class="block text-[10px] font-semibold text-slate-600 mb-0.5">Posición Y (%)</label>
                        <input type="number" id="ctrl-qr-y" min="0" max="100" step="0.5" value="82" class="w-full px-2 py-1 bg-white border border-slate-200 rounded-lg text-xs text-slate-800">
                      </div>
                      <div>
                        <label class="block text-[10px] font-semibold text-slate-600 mb-0.5">Tamaño (mm)</label>
                        <input type="number" id="ctrl-qr-size" min="10" max="45" step="1" value="22" class="w-full px-2 py-1 bg-white border border-slate-200 rounded-lg text-xs text-slate-800">
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Right Column: Student List & Actions -->
          <div class="lg:col-span-5 space-y-5">
            <div class="bg-white rounded-2xl p-5 border border-slate-200/80 shadow-sm space-y-4">
              <div class="flex items-center justify-between">
                <h4 class="text-sm font-bold text-slate-800 flex items-center gap-2">
                  <i data-lucide="users" class="w-4 h-4 text-indigo-600"></i>
                  2. Lista de Estudiantes
                </h4>
                <span id="student-count-badge" class="text-[11px] font-semibold text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded-full">
                  0 alumnos
                </span>
              </div>

              <!-- Tabs: Manual Text or Excel Import -->
              <div class="flex border-b border-slate-200 text-xs font-semibold">
                <button id="tab-text-btn" class="py-2 px-3 border-b-2 border-indigo-600 text-indigo-600 transition-colors">
                  Texto Línea a Línea
                </button>
                <button id="tab-excel-btn" class="py-2 px-3 border-b-2 border-transparent text-slate-500 hover:text-slate-700 transition-colors flex items-center gap-1.5">
                  <i data-lucide="file-spreadsheet" class="w-3.5 h-3.5 text-emerald-600"></i>
                  Importar Excel
                </button>
              </div>

              <!-- Manual Text Input -->
              <div id="tab-text-content" class="space-y-2">
                <textarea id="students-textarea" rows="8" placeholder="Ingresa un nombre de estudiante por línea...&#10;Ejemplo:&#10;Juan Camilo Pérez&#10;María Fernanda Gómez&#10;Carlos Andrés Mendoza" class="w-full p-3 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-sans leading-relaxed resize-none"></textarea>
                <div class="flex items-center justify-between">
                  <p class="text-[11px] text-slate-400">Pega directamente la lista desde un documento o bloc de notas.</p>
                  <button id="load-sample-students-btn" class="text-[11px] text-indigo-600 hover:underline font-medium">Cargar 3 alumnos demo</button>
                </div>
              </div>

              <!-- Excel Upload Tab Content -->
              <div id="tab-excel-content" class="hidden space-y-3">
                <div id="excel-drop-zone" class="border-2 border-dashed border-emerald-300 hover:border-emerald-500 rounded-xl p-4 text-center cursor-pointer bg-emerald-50/30 transition-all">
                  <input type="file" id="excel-file-input" accept=".xlsx,.xls,.csv" class="hidden">
                  <i data-lucide="sheet" class="w-7 h-7 text-emerald-600 mx-auto mb-1.5"></i>
                  <p class="text-xs font-semibold text-slate-700">Arrastra tu archivo Excel (.xlsx / .xls) o haz clic</p>
                  <p class="text-[10px] text-slate-400 mt-0.5">Detección automática inteligente de columnas</p>
                </div>

                <!-- Column selector when Excel parsed -->
                <div id="excel-column-picker" class="hidden space-y-1.5 bg-slate-50 p-3 rounded-xl border border-slate-200">
                  <label class="block text-[11px] font-semibold text-slate-700">Columna de nombres detectada:</label>
                  <select id="excel-column-select" class="w-full px-2.5 py-1.5 bg-white border border-slate-200 rounded-lg text-xs text-slate-800 focus:ring-1 focus:ring-indigo-500"></select>
                </div>
              </div>

              <!-- Generation Actions -->
              <div class="pt-3 border-t border-slate-100 space-y-2">
                <button id="generate-btn" class="w-full py-3 px-4 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 shadow-md shadow-indigo-200 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed">
                  <i data-lucide="download" class="w-4 h-4"></i>
                  <span>Generar Certificados en ZIP</span>
                </button>
                <button id="preview-single-btn" class="w-full py-2 px-4 rounded-xl text-xs font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center justify-center gap-1.5">
                  <i data-lucide="eye" class="w-3.5 h-3.5 text-slate-500"></i>
                  <span>Descargar 1 PDF de Prueba</span>
                </button>
                <p class="text-[11px] text-slate-400 text-center">
                  Los PDFs se compilan en Go en alta resolución vectorial.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Bottom: Issued Certificates Registry -->
        <div class="bg-white rounded-2xl p-6 border border-slate-200/80 shadow-sm space-y-4">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <h4 class="text-sm font-bold text-slate-800 flex items-center gap-2">
                <i data-lucide="shield-check" class="w-4 h-4 text-emerald-600"></i>
                Registro Oficial de Certificados Emitidos (Auditoría Antifraude)
              </h4>
              <p class="text-xs text-slate-500 mt-0.5">
                Historial de certificados generados con código QR y clave criptográfica en SQLite.
              </p>
            </div>
            <div class="flex items-center gap-2">
              <input type="text" id="search-issued-input" placeholder="Buscar alumno o código..." class="px-3 py-1.5 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 w-44 sm:w-56">
              <button id="export-issued-csv-btn" class="flex items-center gap-1.5 text-xs text-slate-600 hover:text-emerald-700 font-semibold px-3 py-1.5 rounded-lg bg-slate-50 hover:bg-emerald-50 transition-colors shrink-0" title="Descargar registro de certificados en archivo CSV">
                <i data-lucide="download" class="w-3.5 h-3.5 text-emerald-600"></i>
                <span>CSV</span>
              </button>
              <button id="refresh-issued-btn" class="flex items-center gap-1.5 text-xs text-slate-600 hover:text-indigo-600 font-semibold px-3 py-1.5 rounded-lg bg-slate-50 hover:bg-indigo-50 transition-colors shrink-0">
                <i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i>
                <span>Actualizar</span>
              </button>
            </div>
          </div>

          <div id="issued-certificates-container" class="space-y-2">
            <div class="py-6 text-center text-xs text-slate-400">
              Cargando historial de certificados...
            </div>
          </div>
        </div>
      </div>
    `;
  },
};
