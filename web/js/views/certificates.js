import { Toast } from '../components/toast.js';
import { Loading } from '../components/loading.js';

export const CertificatesView = {
  // View State
  state: {
    templateDataUrl: null,
    templateWebpBlob: null,
    imgWidth: 0,
    imgHeight: 0,
    origSizeKb: 0,
    webpSizeKb: 0,
    xPercent: 50,
    yPercent: 52,
    fontSize: 32,
    fontColor: '#1e293b',
    fontFamily: 'Helvetica',
    textAlign: 'center',
    includeQR: true,
    qrXPercent: 88,
    qrYPercent: 82,
    qrSize: 22,
    courseTitle: 'Diplomado en Inteligencia Artificial & Desarrollo Web',
    sampleName: 'CARLOS ANDRÉS MENDOZA',
    students: [],
    isDraggingOnCanvas: false,
  },

  // Event listener references for proper cleanup
  cleanupFns: [],

  render() {
    return `
      <div class="space-y-6 max-w-6xl mx-auto">
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

  mount() {
    this.bindEvents();
    this.updateStudentCount();
    if (this.state.templateDataUrl) {
      this.drawCanvas();
    }
    this.loadIssuedCertificates();
  },

  bindEvents() {
    const dropZone = document.getElementById('drop-zone');
    const fileInput = document.getElementById('file-input');

    // Drag and Drop
    dropZone.addEventListener('click', (e) => {
      if (e.target.tagName !== 'CANVAS' && !e.target.closest('#canvas-controls')) {
        fileInput.click();
      }
    });

    fileInput.addEventListener('change', (e) => {
      if (e.target.files && e.target.files[0]) {
        this.processImageFile(e.target.files[0]);
      }
    });

    const onDragOver = (e) => {
      e.preventDefault();
      dropZone.classList.add('border-indigo-600', 'bg-indigo-50/50');
    };
    const onDragLeave = () => {
      dropZone.classList.remove('border-indigo-600', 'bg-indigo-50/50');
    };
    const onDrop = (e) => {
      e.preventDefault();
      dropZone.classList.remove('border-indigo-600', 'bg-indigo-50/50');
      if (e.dataTransfer.files && e.dataTransfer.files[0]) {
        this.processImageFile(e.dataTransfer.files[0]);
      }
    };

    dropZone.addEventListener('dragover', onDragOver);
    dropZone.addEventListener('dragleave', onDragLeave);
    dropZone.addEventListener('drop', onDrop);

    // Global Paste (Ctrl+V) listener
    const onPaste = (e) => {
      const items = (e.clipboardData || e.originalEvent.clipboardData).items;
      for (const item of items) {
        if (item.type.indexOf('image') !== -1) {
          const file = item.getAsFile();
          if (file) {
            Toast.info('Imagen capturada desde el portapapeles (Ctrl+V)');
            this.processImageFile(file);
            break;
          }
        }
      }
    };
    window.addEventListener('paste', onPaste);
    this.cleanupFns.push(() => window.removeEventListener('paste', onPaste));

    // Sample Template Generator Button
    document.getElementById('use-sample-template-btn').addEventListener('click', () => {
      this.generateDemoTemplate();
    });

    // Sample Students Button
    document.getElementById('load-sample-students-btn').addEventListener('click', () => {
      const demoNames = "Carlos Andrés Mendoza\nMaría Fernanda Gómez\nJuan Sebastián Morales";
      document.getElementById('students-textarea').value = demoNames;
      this.parseStudentTextarea();
      Toast.success('3 estudiantes de prueba cargados.');
    });

    // Canvas click & drag coordinate positioning
    const canvas = document.getElementById('cert-canvas');
    const updateCoordsFromMouse = (e) => {
      const rect = canvas.getBoundingClientRect();
      const xPx = e.clientX - rect.left;
      const yPx = e.clientY - rect.top;
      let xPct = Math.max(0, Math.min(100, (xPx / rect.width) * 100));
      let yPct = Math.max(0, Math.min(100, (yPx / rect.height) * 100));

      this.state.xPercent = parseFloat(xPct.toFixed(1));
      this.state.yPercent = parseFloat(yPct.toFixed(1));

      document.getElementById('ctrl-x').value = this.state.xPercent;
      document.getElementById('ctrl-y').value = this.state.yPercent;
      this.drawCanvas();
    };

    canvas.addEventListener('mousedown', (e) => {
      this.state.isDraggingOnCanvas = true;
      updateCoordsFromMouse(e);
    });

    window.addEventListener('mousemove', (e) => {
      if (this.state.isDraggingOnCanvas) {
        updateCoordsFromMouse(e);
      }
    });

    window.addEventListener('mouseup', () => {
      this.state.isDraggingOnCanvas = false;
    });

    // Control Inputs
    document.getElementById('ctrl-x').addEventListener('input', (e) => {
      this.state.xPercent = parseFloat(e.target.value) || 0;
      this.drawCanvas();
    });
    document.getElementById('ctrl-y').addEventListener('input', (e) => {
      this.state.yPercent = parseFloat(e.target.value) || 0;
      this.drawCanvas();
    });
    document.getElementById('ctrl-size').addEventListener('input', (e) => {
      this.state.fontSize = parseInt(e.target.value) || 32;
      this.drawCanvas();
    });
    document.getElementById('ctrl-color').addEventListener('input', (e) => {
      this.state.fontColor = e.target.value;
      document.getElementById('color-label').textContent = e.target.value;
      this.drawCanvas();
    });
    document.getElementById('ctrl-font').addEventListener('change', (e) => {
      this.state.fontFamily = e.target.value;
      this.drawCanvas();
    });

    // QR Code Controls
    const includeQRCheck = document.getElementById('ctrl-include-qr');
    const qrPanel = document.getElementById('qr-settings-panel');
    if (includeQRCheck) {
      includeQRCheck.addEventListener('change', (e) => {
        this.state.includeQR = e.target.checked;
        if (qrPanel) {
          if (e.target.checked) qrPanel.classList.remove('hidden');
          else qrPanel.classList.add('hidden');
        }
        this.drawCanvas();
      });
    }

    const courseTitleInput = document.getElementById('ctrl-course-title');
    if (courseTitleInput) {
      courseTitleInput.addEventListener('input', (e) => {
        this.state.courseTitle = e.target.value.trim();
      });
    }

    const qrXInput = document.getElementById('ctrl-qr-x');
    if (qrXInput) {
      qrXInput.addEventListener('input', (e) => {
        this.state.qrXPercent = parseFloat(e.target.value) || 88;
        this.drawCanvas();
      });
    }

    const qrYInput = document.getElementById('ctrl-qr-y');
    if (qrYInput) {
      qrYInput.addEventListener('input', (e) => {
        this.state.qrYPercent = parseFloat(e.target.value) || 82;
        this.drawCanvas();
      });
    }

    const qrSizeInput = document.getElementById('ctrl-qr-size');
    if (qrSizeInput) {
      qrSizeInput.addEventListener('input', (e) => {
        this.state.qrSize = parseFloat(e.target.value) || 22;
        this.drawCanvas();
      });
    }

    // Alignment buttons
    const alignBtns = document.querySelectorAll('.align-btn');
    alignBtns.forEach(btn => {
      btn.addEventListener('click', () => {
        alignBtns.forEach(b => {
          b.classList.remove('bg-white', 'shadow-xs', 'font-bold', 'text-indigo-600');
          b.classList.add('text-slate-600');
        });
        btn.classList.add('bg-white', 'shadow-xs', 'font-bold', 'text-indigo-600');
        btn.classList.remove('text-slate-600');
        this.state.textAlign = btn.dataset.align;
        this.drawCanvas();
      });
    });

    // Student Textarea
    const textarea = document.getElementById('students-textarea');
    textarea.addEventListener('input', () => {
      this.parseStudentTextarea();
    });

    // Tab switching
    const tabTextBtn = document.getElementById('tab-text-btn');
    const tabExcelBtn = document.getElementById('tab-excel-btn');
    const tabTextContent = document.getElementById('tab-text-content');
    const tabExcelContent = document.getElementById('tab-excel-content');

    tabTextBtn.addEventListener('click', () => {
      tabTextBtn.className = 'py-2 px-3 border-b-2 border-indigo-600 text-indigo-600 transition-colors';
      tabExcelBtn.className = 'py-2 px-3 border-b-2 border-transparent text-slate-500 hover:text-slate-700 transition-colors flex items-center gap-1.5';
      tabTextContent.classList.remove('hidden');
      tabExcelContent.classList.add('hidden');
      this.parseStudentTextarea();
    });

    tabExcelBtn.addEventListener('click', () => {
      tabExcelBtn.className = 'py-2 px-3 border-b-2 border-indigo-600 text-indigo-600 transition-colors flex items-center gap-1.5';
      tabTextBtn.className = 'py-2 px-3 border-b-2 border-transparent text-slate-500 hover:text-slate-700 transition-colors';
      tabExcelContent.classList.remove('hidden');
      tabTextContent.classList.add('hidden');
    });

    // Excel Upload
    const excelDropZone = document.getElementById('excel-drop-zone');
    const excelInput = document.getElementById('excel-file-input');

    excelDropZone.addEventListener('click', () => excelInput.click());
    excelInput.addEventListener('change', (e) => {
      if (e.target.files && e.target.files[0]) {
        this.processExcelFile(e.target.files[0]);
      }
    });

    // Generate Buttons
    document.getElementById('generate-btn').addEventListener('click', () => {
      this.generateCertificates(false);
    });
    document.getElementById('preview-single-btn').addEventListener('click', () => {
      this.generateCertificates(true);
    });

    const refreshIssuedBtn = document.getElementById('refresh-issued-btn');
    if (refreshIssuedBtn) {
      refreshIssuedBtn.addEventListener('click', () => {
        this.loadIssuedCertificates();
      });
    }

    const searchIssuedInput = document.getElementById('search-issued-input');
    if (searchIssuedInput) {
      searchIssuedInput.addEventListener('input', (e) => {
        this.filterIssuedCertificates(e.target.value);
      });
    }

    const exportIssuedCsvBtn = document.getElementById('export-issued-csv-btn');
    if (exportIssuedCsvBtn) {
      exportIssuedCsvBtn.addEventListener('click', () => {
        this.exportIssuedToCSV();
      });
    }
  },

  // Generates an elegant high-res certificate template canvas in 1 click
  generateDemoTemplate() {
    const w = 1920;
    const h = 1357; // A4 aspect ratio 1.414
    const canvas = document.createElement('canvas');
    canvas.width = w;
    canvas.height = h;
    const ctx = canvas.getContext('2d');

    // Background gradient
    const bgGrad = ctx.createLinearGradient(0, 0, w, h);
    bgGrad.addColorStop(0, '#f8fafc');
    bgGrad.addColorStop(1, '#f1f5f9');
    ctx.fillStyle = bgGrad;
    ctx.fillRect(0, 0, w, h);

    // Elegant Outer Gold Border
    ctx.strokeStyle = '#d97706';
    ctx.lineWidth = 14;
    ctx.strokeRect(40, 40, w - 80, h - 80);

    // Inner Navy Border
    ctx.strokeStyle = '#1e3a8a';
    ctx.lineWidth = 4;
    ctx.strokeRect(60, 60, w - 120, h - 120);

    // Corner Ornaments
    ctx.fillStyle = '#b45309';
    const cornerSize = 40;
    ctx.fillRect(40, 40, cornerSize, cornerSize);
    ctx.fillRect(w - 40 - cornerSize, 40, cornerSize, cornerSize);
    ctx.fillRect(40, h - 40 - cornerSize, cornerSize, cornerSize);
    ctx.fillRect(w - 40 - cornerSize, h - 40 - cornerSize, cornerSize);

    // Academy Header
    ctx.fillStyle = '#312e81';
    ctx.font = 'bold 54px Helvetica, Arial, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('DxSTech Edu — Academy of Technology & AI', w / 2, 220);

    ctx.fillStyle = '#64748b';
    ctx.font = '500 28px Helvetica, Arial, sans-serif';
    ctx.fillText('OTORGA EL PRESENTE DIPLOMA DE RECONOCIMIENTO A:', w / 2, 330);

    // Course subtitle placeholder
    ctx.fillStyle = '#475569';
    ctx.font = '500 30px Helvetica, Arial, sans-serif';
    ctx.fillText('Por haber culminado con éxito el programa de especialización en:', w / 2, 880);

    ctx.fillStyle = '#1e293b';
    ctx.font = 'bold 44px Helvetica, Arial, sans-serif';
    ctx.fillText('Desarrollo Web Full-Stack, Inteligencia Artificial & Cloud', w / 2, 950);

    // Footer signature lines
    ctx.strokeStyle = '#94a3b8';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(350, 1180);
    ctx.lineTo(750, 1180);
    ctx.moveTo(1170, 1180);
    ctx.lineTo(1570, 1180);
    ctx.stroke();

    ctx.fillStyle = '#64748b';
    ctx.font = '24px Helvetica, Arial, sans-serif';
    ctx.fillText('Dirección Académica', 550, 1220);
    ctx.fillText('Coordinación de Certificación', 1370, 1220);

    // Convert to WebP quality 0.82
    const webpUrl = canvas.toDataURL('image/webp', 0.82);
    this.state.templateDataUrl = webpUrl;
    this.state.imgWidth = w;
    this.state.imgHeight = h;
    this.state.yPercent = 52.0;
    this.state.xPercent = 50.0;
    this.state.fontSize = 38;
    this.state.fontColor = '#1e293b';

    // Show controls
    document.getElementById('drop-zone-prompt').classList.add('hidden');
    document.getElementById('canvas-wrapper').classList.remove('hidden');
    const stats = document.getElementById('image-stats');
    stats.classList.remove('hidden');
    stats.innerHTML = `📐 1920x1357px • Plantilla Demo DxSTech • WebP: <strong>140 KB</strong>`;

    document.getElementById('ctrl-y').value = 52.0;
    document.getElementById('ctrl-x').value = 50.0;
    document.getElementById('ctrl-size').value = 38;

    // Load sample students if empty
    if (this.state.students.length === 0) {
      document.getElementById('students-textarea').value = "Carlos Andrés Mendoza\nMaría Fernanda Gómez\nJuan Sebastián Morales";
      this.parseStudentTextarea();
    }

    this.drawCanvas();
    Toast.success('Plantilla de demostración generada e insertada con éxito.');
  },

  // Image Processing Rule:
  // 1. validate -> 2. canvas -> 3. resize if >2000px -> 4. convert to WebP 0.82 -> 5. stats -> 6. send WebP
  async processImageFile(file) {
    if (!file.type.startsWith('image/')) {
      Toast.error('El archivo seleccionado no es una imagen válida.');
      return;
    }

    this.state.origSizeKb = (file.size / 1024).toFixed(1);

    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        let { width, height } = img;
        const maxDim = 2000;
        if (width > maxDim || height > maxDim) {
          if (width > height) {
            height = Math.round((height * maxDim) / width);
            width = maxDim;
          } else {
            width = Math.round((width * maxDim) / height);
            height = maxDim;
          }
        }

        // Canvas for WebP conversion
        const offCanvas = document.createElement('canvas');
        offCanvas.width = width;
        offCanvas.height = height;
        const ctx = offCanvas.getContext('2d');
        ctx.drawImage(img, 0, 0, width, height);

        // Convert to WebP quality 0.82
        const webpDataUrl = offCanvas.toDataURL('image/webp', 0.82);
        
        // Estimate WebP size
        const head = 'data:image/webp;base64,';
        const rawB64 = webpDataUrl.substring(head.length);
        const webpBytes = Math.round((rawB64.length * 3) / 4);
        this.state.webpSizeKb = (webpBytes / 1024).toFixed(1);
        this.state.templateDataUrl = webpDataUrl;
        this.state.imgWidth = width;
        this.state.imgHeight = height;

        // UI update
        const prompt = document.getElementById('drop-zone-prompt');
        const canvasWrapper = document.getElementById('canvas-wrapper');
        const stats = document.getElementById('image-stats');

        prompt.classList.add('hidden');
        canvasWrapper.classList.remove('hidden');
        stats.classList.remove('hidden');
        stats.innerHTML = `📐 ${width}x${height}px • WebP: <strong>${this.state.webpSizeKb} KB</strong> (original: ${this.state.origSizeKb} KB)`;

        this.drawCanvas();
        Toast.success('Plantilla procesada y optimizada en WebP');
      };
      img.src = e.target.result;
    };
    reader.readAsDataURL(file);
  },

  drawCanvas() {
    if (!this.state.templateDataUrl) return;

    const canvas = document.getElementById('cert-canvas');
    if (!canvas) return;

    const img = new Image();
    img.onload = () => {
      canvas.width = img.width;
      canvas.height = img.height;
      const ctx = canvas.getContext('2d');

      // Draw background template
      ctx.drawImage(img, 0, 0, img.width, img.height);

      // Coordinates
      const x = (this.state.xPercent / 100) * img.width;
      const y = (this.state.yPercent / 100) * img.height;

      // Draw Crosshair Guides
      ctx.save();
      ctx.strokeStyle = 'rgba(99, 102, 241, 0.6)';
      ctx.lineWidth = 1.5;
      ctx.setLineDash([6, 6]);

      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(img.width, y);
      ctx.stroke();

      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, img.height);
      ctx.stroke();
      ctx.restore();

      // Draw Text Preview
      const scaleFactor = img.width / 1000;
      const effectiveFontSize = Math.round(this.state.fontSize * scaleFactor);

      ctx.font = `bold ${effectiveFontSize}px ${this.state.fontFamily}, sans-serif`;
      ctx.fillStyle = this.state.fontColor;
      ctx.textAlign = this.state.textAlign;
      ctx.textBaseline = 'middle';

      // Text shadow for high contrast preview
      ctx.shadowColor = 'rgba(255, 255, 255, 0.8)';
      ctx.shadowBlur = 4;

      ctx.fillText(this.state.sampleName, x, y);

      // Draw indicator target dot
      ctx.beginPath();
      ctx.arc(x, y, 4, 0, Math.PI * 2);
      ctx.fillStyle = '#4f46e5';
      ctx.fill();

      // Draw Verification QR Code Preview
      if (this.state.includeQR) {
        const qrX = (this.state.qrXPercent / 100) * img.width;
        const qrY = (this.state.qrYPercent / 100) * img.height;
        const qrSizePx = (this.state.qrSize / 297) * img.width;

        ctx.save();
        ctx.fillStyle = '#ffffff';
        ctx.strokeStyle = '#4f46e5';
        ctx.lineWidth = 2;
        ctx.fillRect(qrX - qrSizePx / 2, qrY - qrSizePx / 2, qrSizePx, qrSizePx);
        ctx.strokeRect(qrX - qrSizePx / 2, qrY - qrSizePx / 2, qrSizePx, qrSizePx);

        // Simulated QR corners
        ctx.fillStyle = '#1e293b';
        const pSize = Math.max(4, qrSizePx * 0.22);
        ctx.fillRect(qrX - qrSizePx / 2 + 2, qrY - qrSizePx / 2 + 2, pSize, pSize);
        ctx.fillRect(qrX + qrSizePx / 2 - pSize - 2, qrY - qrSizePx / 2 + 2, pSize, pSize);
        ctx.fillRect(qrX - qrSizePx / 2 + 2, qrY + qrSizePx / 2 - pSize - 2, pSize, pSize);

        // QR label
        ctx.fillStyle = '#4f46e5';
        ctx.font = `bold ${Math.max(9, Math.round(qrSizePx * 0.16))}px monospace`;
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText('QR DXS', qrX, qrY);

        // Code label underneath
        ctx.fillStyle = '#64748b';
        ctx.font = `bold ${Math.max(8, Math.round(qrSizePx * 0.12))}px sans-serif`;
        ctx.fillText('VERIFICADO', qrX, qrY + qrSizePx / 2 + 10);
        ctx.restore();
      }
    };
    img.src = this.state.templateDataUrl;
  },

  parseStudentTextarea() {
    const text = document.getElementById('students-textarea').value;
    const lines = text.split('\n').map(l => l.trim()).filter(l => l.length > 0);
    this.state.students = lines;
    this.updateStudentCount();
  },

  processExcelFile(file) {
    if (!window.XLSX) {
      Toast.error('Biblioteca SheetJS no cargada');
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = new Uint8Array(e.target.result);
        const workbook = window.XLSX.read(data, { type: 'array' });
        const firstSheetName = workbook.SheetNames[0];
        const worksheet = workbook.Sheets[firstSheetName];
        const json = window.XLSX.utils.sheet_to_json(worksheet, { header: 1 });

        if (json.length < 2) {
          Toast.warning('El archivo Excel parece estar vacío o no contiene filas de datos.');
          return;
        }

        const headers = json[0].map(h => String(h).trim());
        const select = document.getElementById('excel-column-select');
        select.innerHTML = '';

        // Auto-detect column matching nombre, nombres, alumno, estudiante
        const candidates = ['nombre', 'nombres', 'alumno', 'alumnos', 'estudiante', 'estudiantes', 'participante'];
        let matchedIndex = 0;

        headers.forEach((h, index) => {
          const opt = document.createElement('option');
          opt.value = index;
          opt.textContent = `Columna ${index + 1}: ${h}`;
          select.appendChild(opt);

          const lower = h.toLowerCase();
          if (candidates.some(c => lower.includes(c))) {
            matchedIndex = index;
          }
        });

        select.value = matchedIndex;
        document.getElementById('excel-column-picker').classList.remove('hidden');

        const extractNames = (colIdx) => {
          const names = [];
          for (let i = 1; i < json.length; i++) {
            const row = json[i];
            if (row && row[colIdx]) {
              const val = String(row[colIdx]).trim();
              if (val) names.push(val);
            }
          }
          this.state.students = names;
          this.updateStudentCount();
          Toast.success(`Se importaron ${names.length} estudiantes desde Excel.`);
        };

        select.onchange = () => extractNames(parseInt(select.value));
        extractNames(matchedIndex);
      } catch (err) {
        Toast.error('Error al procesar el archivo Excel: ' + err.message);
      }
    };
    reader.readAsArrayBuffer(file);
  },

  updateStudentCount() {
    const badge = document.getElementById('student-count-badge');
    const count = this.state.students.length;
    badge.textContent = `${count} ${count === 1 ? 'alumno' : 'alumnos'}`;
  },

  async generateCertificates(isSinglePreview = false) {
    if (!this.state.templateDataUrl) {
      Toast.warning('Por favor carga primero la imagen de la plantilla del certificado.');
      return;
    }

    let targetStudents = this.state.students;
    if (isSinglePreview) {
      targetStudents = targetStudents.length > 0 ? [targetStudents[0]] : ['ESTUDIANTE DEMO'];
    }

    if (targetStudents.length === 0) {
      Toast.warning('Debes ingresar al menos un nombre de estudiante.');
      return;
    }

    Loading.overlay(`Generando ${isSinglePreview ? '1 PDF de prueba con QR de verificación' : targetStudents.length + ' certificados con QR en PDF'}...`);

    try {
      const payload = {
        template: this.state.templateDataUrl,
        xPercent: this.state.xPercent,
        yPercent: this.state.yPercent,
        fontSize: this.state.fontSize,
        fontColor: this.state.fontColor,
        fontFamily: this.state.fontFamily,
        textAlign: this.state.textAlign,
        includeQR: this.state.includeQR,
        qrXPercent: this.state.qrXPercent,
        qrYPercent: this.state.qrYPercent,
        qrSize: this.state.qrSize,
        courseTitle: this.state.courseTitle,
        students: targetStudents,
      };

      const response = await fetch('/api/certificates/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errJson = await response.json().catch(() => ({}));
        throw new Error(errJson.error || `Error HTTP ${response.status}`);
      }

      // Download ZIP
      const blob = await response.blob();
      const downloadUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = isSinglePreview ? 'certificado_muestra.zip' : 'certificados_dxstech.zip';
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(downloadUrl);

      Toast.success(isSinglePreview ? 'PDF de prueba generado y descargado.' : `¡Éxito! ${targetStudents.length} certificados descargados en ZIP.`);
      this.loadIssuedCertificates();
    } catch (err) {
      Toast.error('Fallo al generar certificados: ' + err.message);
    } finally {
      Loading.hideOverlay();
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
          Error al cargar certificados emitidos: ${err.message}
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
      const verifyHash = `#verify/${item.id}`;
      html += `
        <tr class="hover:bg-slate-50/80 transition-colors">
          <td class="py-2.5 px-3 font-mono font-bold text-indigo-700">
            <span class="bg-indigo-50 border border-indigo-200/60 px-2 py-0.5 rounded text-[11px]">
              ${item.id}
            </span>
          </td>
          <td class="py-2.5 px-3 font-semibold text-slate-900">${item.studentName}</td>
          <td class="py-2.5 px-3 text-slate-600 text-[11px]">${item.courseTitle || 'Diplomado'}</td>
          <td class="py-2.5 px-3 text-slate-500 text-[11px]">${item.issueDate}</td>
          <td class="py-2.5 px-3 text-right space-x-1">
            <a href="${verifyHash}" class="inline-flex items-center gap-1 px-2.5 py-1 text-[11px] font-bold text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg shadow-xs transition-colors">
              <i data-lucide="shield-check" class="w-3 h-3"></i>
              Verificar
            </a>
            <button data-copy-link="${item.id}" class="inline-flex items-center gap-1 px-2 py-1 text-[11px] font-semibold text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-lg transition-colors">
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
      csv += `${code},${name},${course},${date},${url}\r\n`;
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

  destroy() {
    this.cleanupFns.forEach(fn => fn());
    this.cleanupFns = [];
  }
};
