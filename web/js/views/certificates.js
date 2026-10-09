import { Toast } from '../components/toast.js';
import { tabsMethods } from './certificates/tabs.js';
import { templateEditorMethods } from './certificates/template-editor.js';
import { issuedMethods } from './certificates/issued.js';

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
    activeTab: 'my-certs',
    myCertificates: [],
  },

  // Event listener references for proper cleanup
  cleanupFns: [],

  render() {
    const user = window.router?.currentUser;
    const isAdmin = user && ['SUPERADMIN', 'ADMINISTRADOR'].includes(user.role);

    return `
      <div class="space-y-6 max-w-6xl mx-auto">
        <!-- Top Toolbar / Tabs -->
        <div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-xs flex flex-col sm:flex-row items-center justify-between gap-4">
          <div class="flex items-center gap-1.5 p-1 bg-slate-100 rounded-xl w-full sm:w-auto">
            <button id="cert-tab-my" class="flex-1 sm:flex-none px-4 py-2 rounded-lg text-xs font-bold transition-all cursor-pointer ${this.state.activeTab === 'my-certs' ? 'bg-white text-indigo-700 shadow-2xs' : 'text-slate-600 hover:text-slate-900'}">
              <i data-lucide="award" class="w-3.5 h-3.5 inline mr-1 text-amber-500"></i>
              <span>Mis Certificados</span>
            </button>
            ${isAdmin ? `
              <button id="cert-tab-gen" class="flex-1 sm:flex-none px-4 py-2 rounded-lg text-xs font-bold transition-all cursor-pointer ${this.state.activeTab === 'generator' ? 'bg-white text-indigo-700 shadow-2xs' : 'text-slate-600 hover:text-slate-900'}">
                <i data-lucide="layout-template" class="w-3.5 h-3.5 inline mr-1 text-indigo-600"></i>
                <span>Generador Masivo</span>
              </button>
            ` : ''}
            <button id="cert-tab-verify" class="flex-1 sm:flex-none px-4 py-2 rounded-lg text-xs font-bold transition-all cursor-pointer ${this.state.activeTab === 'verify' ? 'bg-white text-indigo-700 shadow-2xs' : 'text-slate-600 hover:text-slate-900'}">
              <i data-lucide="shield-check" class="w-3.5 h-3.5 inline mr-1 text-emerald-600"></i>
              <span>Validar Registro</span>
            </button>
          </div>
          <div class="text-[11px] text-slate-400 font-medium">
            <span>Acreditaciones oficiales emitidas por </span><strong class="text-indigo-600">DxSTech Edu</strong>
          </div>
        </div>

        ${this.state.activeTab === 'my-certs' ? this.renderMyCertificatesTab() : (this.state.activeTab === 'generator' ? this.renderGeneratorTab() : this.renderVerifyTab())}
      </div>
    `;
  },

  mount() {
    this.bindTabEvents();
    if (this.state.activeTab === 'my-certs') {
      this.loadMyCertificates();
    } else if (this.state.activeTab === 'generator') {
      this.bindEvents();
      this.updateStudentCount();
      if (this.state.templateDataUrl) {
        this.drawCanvas();
      }
      this.loadIssuedCertificates();
    } else if (this.state.activeTab === 'verify') {
      this.bindVerifyEvents();
    }
  },

  bindTabEvents() {
    document.getElementById('cert-tab-my')?.addEventListener('click', () => {
      this.state.activeTab = 'my-certs';
      window.router?.navigate('certificates');
    });
    document.getElementById('cert-tab-gen')?.addEventListener('click', () => {
      this.state.activeTab = 'generator';
      window.router?.navigate('certificates');
    });
    document.getElementById('cert-tab-verify')?.addEventListener('click', () => {
      this.state.activeTab = 'verify';
      window.router?.navigate('certificates');
    });
  },

  bindVerifyEvents() {
    const input = document.getElementById('verify-code-input');
    const btn = document.getElementById('verify-code-btn');
    if (btn && input) {
      btn.addEventListener('click', () => {
        const code = input.value.trim();
        if (!code) {
          Toast.warning('Ingresa un código de certificado para validar.');
          return;
        }
        window.router?.renderVerificationScreen(code);
      });
      input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
          btn.click();
        }
      });
    }
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

  destroy() {
    this.cleanupFns.forEach(fn => fn());
    this.cleanupFns = [];
  },
};

// Las vistas grandes se dividen por responsabilidad; los métodos se mezclan en la vista
// y siguen usando `this`, por lo que su comportamiento no cambia.
Object.assign(CertificatesView, tabsMethods, templateEditorMethods, issuedMethods);
