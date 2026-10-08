import { Toast } from '../components/toast.js';
import { Modal } from '../components/modal.js';
import { Loading } from '../components/loading.js';

export const WhatsAppView = {
  state: {
    activeTab: 'config', // 'config' | 'send'
    botConfig: {
      isActive: true,
      knowledgeBase: '',
      strictMode: true,
      respondGroups: false,
      provider: 'Simulador / Baileys QR',
    },
    statusData: null,
    pollIntervalId: null,
    chatMessages: [
      { sender: 'bot', text: '👋 ¡Hola! Soy el asistente virtual oficial de DxSTech Edu. ¿En qué te puedo asesorar hoy?' }
    ],
    isBotTyping: false,
  },

  cleanupFns: [],

  render() {
    return `
      <div class="space-y-6 max-w-6xl mx-auto">
        <!-- Top Banner with Free Dev Mode Notice -->
        <div class="bg-gradient-to-r from-emerald-600 to-teal-700 rounded-2xl p-6 text-white shadow-lg shadow-emerald-100 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <div class="flex items-center gap-2 mb-1">
              <span class="text-[10px] font-bold uppercase tracking-wider bg-white/20 text-white px-2 py-0.5 rounded-full">
                Acceso Libre en Desarrollo Local
              </span>
              <span class="text-[10px] font-medium text-emerald-100">Sin login requerido</span>
            </div>
            <h3 class="text-lg font-bold">WhatsApp Gateway + Chatbot IA</h3>
            <p class="text-xs text-emerald-100 mt-0.5 max-w-xl">
              Configura las respuestas automáticas de tu academia con base de conocimiento estricta y programa envíos masivos con control de tasa de entrega.
            </p>
          </div>

          <!-- Status Indicator Widget -->
          <div id="wa-status-badge" class="flex items-center gap-3 bg-white/10 backdrop-blur-md px-4 py-2 rounded-xl border border-white/20 text-xs">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-300 animate-pulse"></span>
            <div>
              <p class="font-bold leading-tight" id="status-conn-label">Conectado al Gateway</p>
              <p class="text-[10px] text-emerald-100" id="status-phone-label">+57 (300) 890-DXSTECH</p>
            </div>
          </div>
        </div>

        <!-- Tab Selection -->
        <div class="flex border-b border-slate-200 text-xs font-semibold">
          <button id="wa-tab-config-btn" class="py-3 px-5 border-b-2 border-emerald-600 text-emerald-700 transition-colors flex items-center gap-2">
            <i data-lucide="bot" class="w-4 h-4"></i>
            <span>Configurar Asistente IA</span>
          </button>
          <button id="wa-tab-send-btn" class="py-3 px-5 border-b-2 border-transparent text-slate-500 hover:text-slate-700 transition-colors flex items-center gap-2">
            <i data-lucide="send" class="w-4 h-4"></i>
            <span>Gestor de Envíos Masivos</span>
          </button>
        </div>

        <!-- TAB 1: Configurar IA -->
        <div id="wa-tab-config-content" class="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <!-- Left: Configuration Form -->
          <div class="lg:col-span-7 bg-white rounded-2xl p-6 border border-slate-200/80 shadow-sm space-y-5">
            <div class="flex items-center justify-between pb-3 border-b border-slate-100">
              <div>
                <h4 class="text-sm font-bold text-slate-800">Parámetros del Chatbot</h4>
                <p class="text-xs text-slate-400">Personaliza la base de conocimiento y el comportamiento del bot</p>
              </div>
              <!-- Toggle Active -->
              <label class="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" id="bot-active-toggle" class="sr-only peer" checked>
                <div class="w-11 h-6 bg-slate-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-emerald-600"></div>
                <span class="ml-2 text-xs font-semibold text-slate-700">Bot Activo</span>
              </label>
            </div>

            <!-- Knowledge Base -->
            <div class="space-y-1.5">
              <div class="flex items-center justify-between">
                <label class="block text-xs font-semibold text-slate-700">Base de Conocimiento Institucional</label>
                <button id="reset-kb-btn" class="text-[11px] text-emerald-600 hover:underline font-medium">Restablecer datos base</button>
              </div>
              <textarea id="knowledge-base-input" rows="7" placeholder="Escribe aquí toda la información oficial de la institución:&#10;- Horarios de atención: Lunes a Viernes de 8:00 AM a 6:00 PM&#10;- Cursos disponibles: Desarrollo Web, IA Aplicada, Certificaciones Cloud&#10;- Métodos de pago y políticas de reembolso...&#10;- Contacto de admisiones..." class="w-full p-3 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-emerald-500 font-sans leading-relaxed resize-none"></textarea>
            </div>

            <!-- Strict Mode & Groups -->
            <div class="space-y-3 pt-2 border-t border-slate-100">
              <label class="flex items-start gap-3 p-3 rounded-xl bg-slate-50 border border-slate-200/80 cursor-pointer hover:bg-slate-100/60 transition-colors">
                <input type="checkbox" id="strict-mode-toggle" class="mt-0.5 rounded text-emerald-600 focus:ring-emerald-500" checked>
                <div>
                  <p class="text-xs font-bold text-slate-800">Modo Estricto de Veracidad (Recomendado)</p>
                  <p class="text-[11px] text-slate-500 leading-relaxed">
                    Si la pregunta del usuario no está cubierta en la base de conocimiento, el bot responderá estrictamente: <em>"Lo siento, solo puedo responder preguntas sobre la base de conocimiento proporcionada."</em> Evita alucinaciones.
                  </p>
                </div>
              </label>

              <label class="flex items-start gap-3 p-3 rounded-xl bg-slate-50 border border-slate-200/80 cursor-pointer hover:bg-slate-100/60 transition-colors">
                <input type="checkbox" id="groups-toggle" class="mt-0.5 rounded text-emerald-600 focus:ring-emerald-500">
                <div>
                  <p class="text-xs font-bold text-slate-800">Responder en Grupos de WhatsApp</p>
                  <p class="text-[11px] text-slate-500">Permite que el bot responda cuando es mencionado en grupos académicos.</p>
                </div>
              </label>
            </div>

            <!-- Save Config Button -->
            <div class="pt-3 border-t border-slate-100 flex justify-end">
              <button id="save-wa-config-btn" class="px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-700 shadow-md shadow-emerald-200 transition-all flex items-center gap-2">
                <i data-lucide="save" class="w-4 h-4"></i>
                <span>Guardar Configuración del Bot</span>
              </button>
            </div>
          </div>

          <!-- Right: Interactive Live Chatbot Simulator -->
          <div class="lg:col-span-5 bg-white rounded-2xl border border-slate-200/80 shadow-sm flex flex-col h-[560px] overflow-hidden">
            <!-- Chat Simulator Header -->
            <div class="p-4 bg-emerald-50/80 border-b border-emerald-100 flex items-center justify-between">
              <div class="flex items-center gap-2.5">
                <div class="w-8 h-8 rounded-full bg-emerald-600 text-white flex items-center justify-center font-bold text-xs shadow-xs">
                  <i data-lucide="bot" class="w-4 h-4"></i>
                </div>
                <div>
                  <h5 class="text-xs font-bold text-slate-800">Simulador de WhatsApp en Vivo</h5>
                  <p class="text-[10px] text-emerald-700 font-medium">Validación con Gemini + Base de Conocimiento</p>
                </div>
              </div>
              <button id="clear-chat-btn" class="text-[10px] text-slate-400 hover:text-slate-600">Limpiar</button>
            </div>

            <!-- Quick Suggestion Chips -->
            <div class="px-3 py-2 bg-slate-100/60 border-b border-slate-100 flex items-center gap-1.5 overflow-x-auto text-[10px]">
              <span class="text-slate-400 font-semibold shrink-0">Probar:</span>
              <button class="chip-btn px-2.5 py-1 rounded-full bg-white hover:bg-emerald-50 text-slate-700 border border-slate-200 shrink-0 font-medium transition-colors" data-msg="¿Qué cursos ofrecen y en qué modalidad?">
                ¿Qué cursos ofrecen?
              </button>
              <button class="chip-btn px-2.5 py-1 rounded-full bg-white hover:bg-emerald-50 text-slate-700 border border-slate-200 shrink-0 font-medium transition-colors" data-msg="¿Entregan certificado al finalizar?">
                ¿Dan certificado?
              </button>
              <button class="chip-btn px-2.5 py-1 rounded-full bg-white hover:bg-amber-50 text-amber-700 border border-amber-200 shrink-0 font-medium transition-colors" data-msg="¿Cuánto cuesta un viaje a la luna en cohete?">
                Probar Modo Estricto 🚀
              </button>
            </div>

            <!-- Chat Messages Flow -->
            <div id="chat-messages-container" class="flex-1 p-4 overflow-y-auto space-y-3 bg-slate-50/50">
              <!-- Messages will be injected here -->
            </div>

            <!-- Chat Input Box -->
            <div class="p-3 bg-white border-t border-slate-100">
              <form id="chat-form" class="flex items-center gap-2">
                <input type="text" id="chat-user-input" placeholder="Escribe un mensaje de prueba..." class="flex-1 px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-1 focus:ring-emerald-500">
                <button type="submit" class="w-9 h-9 rounded-xl bg-emerald-600 text-white flex items-center justify-center hover:bg-emerald-700 transition-colors shrink-0">
                  <i data-lucide="send" class="w-4 h-4"></i>
                </button>
              </form>
            </div>
          </div>
        </div>

        <!-- TAB 2: Gestor de Envíos Masivos -->
        <div id="wa-tab-send-content" class="hidden space-y-6">
          <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
            <!-- Left: Bulk Sender Form -->
            <div class="lg:col-span-6 bg-white rounded-2xl p-6 border border-slate-200/80 shadow-sm space-y-4">
              <h4 class="text-sm font-bold text-slate-800 flex items-center gap-2">
                <i data-lucide="send" class="w-4 h-4 text-emerald-600"></i>
                Programar Notificaciones Masivas
              </h4>

              <!-- Recipients -->
              <div>
                <label class="block text-xs font-semibold text-slate-700 mb-1">Destinatarios (Números con código de país)</label>
                <textarea id="recipients-input" rows="4" placeholder="+573001234567&#10;+573109876543&#10;+14155552671" class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-emerald-500 font-mono resize-none"></textarea>
                <div class="flex items-center justify-between mt-1">
                  <p class="text-[10px] text-slate-400">Un número por línea o separados por comas.</p>
                  <div class="flex items-center gap-2">
                    <input type="file" id="wa-recipients-file-input" accept=".xlsx,.xls,.csv" class="hidden">
                    <button type="button" id="import-wa-excel-btn" class="text-[10px] text-slate-600 hover:text-emerald-700 font-semibold flex items-center gap-1 hover:underline">
                      <i data-lucide="file-spreadsheet" class="w-3 h-3 text-emerald-600"></i>
                      <span>Importar Excel/CSV</span>
                    </button>
                    <span class="text-slate-300">|</span>
                    <button type="button" id="load-sample-recipients-btn" class="text-[10px] text-emerald-600 hover:underline font-semibold">3 Demo</button>
                  </div>
                </div>
              </div>

              <!-- Message Body -->
              <div>
                <div class="flex items-center justify-between mb-1">
                  <label class="block text-xs font-semibold text-slate-700">Mensaje de Notificación</label>
                  <div class="flex items-center gap-1.5 text-[10px]">
                    <span class="text-slate-400">Plantillas:</span>
                    <button type="button" id="tmpl-cert-btn" class="text-emerald-600 hover:underline font-semibold">Certificado</button>
                    <span class="text-slate-300">•</span>
                    <button type="button" id="tmpl-exam-btn" class="text-emerald-600 hover:underline font-semibold">Evaluación</button>
                    <span class="text-slate-300">•</span>
                    <button type="button" id="tmpl-welcome-btn" class="text-emerald-600 hover:underline font-semibold">Bienvenida</button>
                  </div>
                </div>
                <textarea id="bulk-message-input" rows="4" placeholder="Estimado estudiante, tu certificado del curso ya se encuentra disponible para descarga en el portal oficial de DxSTech Edu." class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-emerald-500 resize-none"></textarea>
              </div>

              <!-- Rate limit slider -->
              <div>
                <div class="flex items-center justify-between mb-1">
                  <label class="text-xs font-semibold text-slate-700">Rate Limit (Demora entre envíos)</label>
                  <span id="delay-label" class="text-xs font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-md">2 segundos</span>
                </div>
                <input type="range" id="delay-slider" min="1" max="10" value="2" step="1" class="w-full accent-emerald-600 cursor-pointer">
                <p class="text-[10px] text-slate-400">Previene bloqueos o rate limits de la API de WhatsApp.</p>
              </div>

              <!-- Consent acknowledgement -->
              <label class="flex items-start gap-2.5 p-3 rounded-xl bg-amber-50/70 border border-amber-200/80 text-xs text-amber-900 cursor-pointer">
                <input type="checkbox" id="consent-checkbox" class="mt-0.5 rounded text-amber-600 focus:ring-amber-500" checked>
                <span class="leading-relaxed">
                  Confirmo que los destinatarios han dado su consentimiento para recibir comunicaciones educativas de conformidad con las políticas anti-spam.
                </span>
              </label>

              <button id="start-bulk-send-btn" class="w-full py-3 px-4 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-700 shadow-md shadow-emerald-200 transition-all flex items-center justify-center gap-2">
                <i data-lucide="play" class="w-4 h-4"></i>
                <span>Iniciar Envío en Cola</span>
              </button>
            </div>

            <!-- Right: Queue Status & Activity Log -->
            <div class="lg:col-span-6 space-y-5">
              <!-- Live Queue Metrics Card -->
              <div class="grid grid-cols-2 gap-4">
                <div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-sm text-center">
                  <span class="text-[10px] font-semibold text-slate-400 uppercase">Mensajes en Cola</span>
                  <p id="metric-queue" class="text-2xl font-black text-indigo-600 mt-1">0</p>
                </div>
                <div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-sm text-center">
                  <span class="text-[10px] font-semibold text-slate-400 uppercase">Enviados Hoy</span>
                  <p id="metric-sent" class="text-2xl font-black text-emerald-600 mt-1">0</p>
                </div>
              </div>

              <!-- Real-time Bitácora Table -->
              <div class="bg-white rounded-2xl p-5 border border-slate-200/80 shadow-sm space-y-3">
                <div class="flex items-center justify-between">
                  <h4 class="text-xs font-bold text-slate-800 flex items-center gap-2">
                    <i data-lucide="list-checks" class="w-4 h-4 text-emerald-600"></i>
                    Bitácora Individual de Mensajes
                  </h4>
                  <div class="flex items-center gap-2 text-[11px]">
                    <button type="button" id="export-logs-csv-btn" class="text-slate-600 hover:text-emerald-700 font-semibold flex items-center gap-1 hover:underline">
                      <i data-lucide="download" class="w-3 h-3"></i>
                      <span>Exportar CSV</span>
                    </button>
                    <span class="text-slate-300">|</span>
                    <button type="button" id="clear-logs-btn" class="text-rose-500 hover:text-rose-700 font-semibold flex items-center gap-1 hover:underline">
                      <i data-lucide="trash-2" class="w-3 h-3"></i>
                      <span>Vaciar</span>
                    </button>
                  </div>
                </div>

                <div id="logs-container" class="max-h-[300px] overflow-y-auto space-y-2">
                  <div class="p-4 text-center text-slate-400 text-xs">Cargando registros...</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    `;
  },

  mount() {
    this.bindEvents();
    this.loadConfig();
    this.renderChatMessages();
    this.startPolling();
  },

  bindEvents() {
    // Tabs
    const tabConfigBtn = document.getElementById('wa-tab-config-btn');
    const tabSendBtn = document.getElementById('wa-tab-send-btn');
    const tabConfigContent = document.getElementById('wa-tab-config-content');
    const tabSendContent = document.getElementById('wa-tab-send-content');

    tabConfigBtn.addEventListener('click', () => {
      this.state.activeTab = 'config';
      tabConfigBtn.className = 'py-3 px-5 border-b-2 border-emerald-600 text-emerald-700 transition-colors flex items-center gap-2';
      tabSendBtn.className = 'py-3 px-5 border-b-2 border-transparent text-slate-500 hover:text-slate-700 transition-colors flex items-center gap-2';
      tabConfigContent.classList.remove('hidden');
      tabSendContent.classList.add('hidden');
    });

    tabSendBtn.addEventListener('click', () => {
      this.state.activeTab = 'send';
      tabSendBtn.className = 'py-3 px-5 border-b-2 border-emerald-600 text-emerald-700 transition-colors flex items-center gap-2';
      tabConfigBtn.className = 'py-3 px-5 border-b-2 border-transparent text-slate-500 hover:text-slate-700 transition-colors flex items-center gap-2';
      tabSendContent.classList.remove('hidden');
      tabConfigContent.classList.add('hidden');
      this.fetchLogs();
    });

    // Reset KB
    document.getElementById('reset-kb-btn').addEventListener('click', () => {
      document.getElementById('knowledge-base-input').value = 
`DxSTech Edu es una academia digital de alta tecnología especializada en Inteligencia Artificial, Desarrollo Web Moderno y Cloud Computing.
Cursos oficiales disponibles:
1. Diplomado Full-Stack Developer con Go y JavaScript.
2. Certificación en Inteligencia Artificial Generativa y Gemini API.
3. Arquitectura Cloud y Microservicios.
Horarios de atención académica: Lunes a Viernes de 8:00 AM a 6:00 PM (hora de Colombia).
Metodología: 100% online con clases prácticas en vivo, tutoría personalizada y proyectos para portafolio profesional.
Certificados: Al aprobar con 60% o más, se emite un certificado digital en PDF verificable con código único.`;
      Toast.info('Base de conocimiento restablecida con datos institucionales completos.');
    });

    // Save Bot Config
    document.getElementById('save-wa-config-btn').addEventListener('click', () => {
      this.saveConfig();
    });

    // Sample recipients button
    document.getElementById('load-sample-recipients-btn').addEventListener('click', () => {
      document.getElementById('recipients-input').value = "+573008901234\n+573155554321\n+573209876543";
      document.getElementById('bulk-message-input').value = "Estimado alumno, tu certificado del Diplomado DxSTech Edu ya está disponible para descarga.";
      Toast.info('Destinatarios y mensaje de prueba cargados.');
    });

    // Excel/CSV import for recipients
    const importWaExcelBtn = document.getElementById('import-wa-excel-btn');
    const waFileInput = document.getElementById('wa-recipients-file-input');
    if (importWaExcelBtn && waFileInput) {
      importWaExcelBtn.addEventListener('click', () => waFileInput.click());
      waFileInput.addEventListener('change', (e) => {
        if (e.target.files && e.target.files[0]) {
          this.processRecipientsExcelFile(e.target.files[0]);
        }
      });
    }

    // Quick Templates Buttons
    document.getElementById('tmpl-cert-btn')?.addEventListener('click', () => {
      document.getElementById('bulk-message-input').value = '🎓 Estimado/a estudiante, tu certificado oficial ya se encuentra disponible para descarga con código de verificación QR en la plataforma DxSTech Edu.';
      Toast.info('Plantilla de certificado cargada.');
    });
    document.getElementById('tmpl-exam-btn')?.addEventListener('click', () => {
      document.getElementById('bulk-message-input').value = '📝 Te recordamos que la evaluación pedagógica del módulo ya está habilitada. Puedes realizarla desde el simulador interactivo en línea.';
      Toast.info('Plantilla de evaluación cargada.');
    });
    document.getElementById('tmpl-welcome-btn')?.addEventListener('click', () => {
      document.getElementById('bulk-message-input').value = '👋 ¡Bienvenido/a al programa de formación de DxSTech Edu! Nuestro asistente virtual con IA está disponible 24/7 para resolver tus dudas.';
      Toast.info('Plantilla de bienvenida cargada.');
    });

    // Delay slider
    const slider = document.getElementById('delay-slider');
    const delayLabel = document.getElementById('delay-label');
    slider.addEventListener('input', (e) => {
      delayLabel.textContent = `${e.target.value} ${e.target.value === '1' ? 'segundo' : 'segundos'}`;
    });

    // Bulk Send
    document.getElementById('start-bulk-send-btn').addEventListener('click', () => {
      this.startBulkSend();
    });

    // Suggestion chips
    document.querySelectorAll('.chip-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const msg = btn.dataset.msg;
        document.getElementById('chat-user-input').value = msg;
        this.handleUserChatMessage();
      });
    });

    // Chat Simulator Form
    document.getElementById('chat-form').addEventListener('submit', (e) => {
      e.preventDefault();
      this.handleUserChatMessage();
    });

    // Clear chat
    document.getElementById('clear-chat-btn').addEventListener('click', () => {
      this.state.chatMessages = [
        { sender: 'bot', text: '👋 ¡Hola! Soy el asistente virtual oficial de DxSTech Edu. ¿En qué te puedo asesorar hoy?' }
      ];
      this.renderChatMessages();
    });

    // Logs Actions: CSV Export & Clear
    document.getElementById('export-logs-csv-btn')?.addEventListener('click', () => {
      this.exportLogsToCSV();
    });

    document.getElementById('clear-logs-btn')?.addEventListener('click', () => {
      this.confirmClearLogs();
    });
  },

  async loadConfig() {
    try {
      const res = await fetch('/api/whatsapp/status');
      if (!res.ok) return;
      const data = await res.json();
      this.state.statusData = data;

      if (data.config) {
        this.state.botConfig = data.config;
        document.getElementById('bot-active-toggle').checked = data.config.isActive;
        document.getElementById('knowledge-base-input').value = data.config.knowledgeBase || '';
        document.getElementById('strict-mode-toggle').checked = data.config.strictMode;
        document.getElementById('groups-toggle').checked = data.config.respondGroups;
      }

      this.updateStatusUI(data);
    } catch (e) {
      console.error('Error cargando estado WhatsApp:', e);
    }
  },

  async saveConfig() {
    const isActive = document.getElementById('bot-active-toggle').checked;
    const knowledgeBase = document.getElementById('knowledge-base-input').value.trim();
    const strictMode = document.getElementById('strict-mode-toggle').checked;
    const respondGroups = document.getElementById('groups-toggle').checked;

    Loading.overlay('Guardando configuración de WhatsApp...');
    try {
      const res = await fetch('/api/whatsapp/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          isActive,
          knowledgeBase,
          strictMode,
          respondGroups,
          provider: 'Simulador / Baileys Gateway',
        }),
      });

      if (!res.ok) throw new Error('Error guardando en el servidor');
      Toast.success('Configuración del Asistente Virtual guardada.');
    } catch (e) {
      Toast.error('Fallo al guardar: ' + e.message);
    } finally {
      Loading.hideOverlay();
    }
  },

  async handleUserChatMessage() {
    const input = document.getElementById('chat-user-input');
    const msg = input.value.trim();
    if (!msg) return;

    input.value = '';
    this.state.chatMessages.push({ sender: 'user', text: msg });
    this.state.isBotTyping = true;
    this.renderChatMessages();

    // Check Gemini API Key
    const apiKey = localStorage.getItem('dxstech_gemini_api_key');
    if (!apiKey) {
      this.state.isBotTyping = false;
      this.state.chatMessages.push({
        sender: 'bot',
        text: '⚠️ Para que pueda responder con Inteligencia Artificial, configura tu Gemini API Key en Ajustes.',
      });
      this.renderChatMessages();
      Modal.promptGeminiKey();
      return;
    }

    // Call chat simulator endpoint
    try {
      const res = await fetch('/api/whatsapp/simulate-chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Gemini-API-Key': apiKey,
        },
        body: JSON.stringify({ message: msg }),
      });

      const data = await res.json();
      this.state.isBotTyping = false;
      this.state.chatMessages.push({
        sender: 'bot',
        text: data.reply || 'Lo siento, no pude procesar la consulta.',
      });
    } catch (e) {
      this.state.isBotTyping = false;
      this.state.chatMessages.push({
        sender: 'bot',
        text: '⚠️ Error de comunicación con la IA: ' + e.message,
      });
    }
    this.renderChatMessages();
  },

  renderChatMessages() {
    const container = document.getElementById('chat-messages-container');
    if (!container) return;

    let html = this.state.chatMessages.map(m => {
      const isBot = m.sender === 'bot';
      return `
        <div class="flex items-start gap-2.5 ${isBot ? '' : 'justify-end'}">
          ${isBot ? `
            <div class="w-6 h-6 rounded-full bg-emerald-600 text-white flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5">
              <i data-lucide="bot" class="w-3.5 h-3.5"></i>
            </div>
          ` : ''}
          <div class="max-w-[82%] rounded-2xl px-3.5 py-2.5 text-xs leading-relaxed ${
            isBot
              ? 'bg-white text-slate-800 border border-slate-200/80 shadow-xs'
              : 'bg-emerald-600 text-white shadow-xs'
          }">
            ${this.escapeHtml(m.text)}
          </div>
        </div>
      `;
    }).join('');

    // Typing bubble
    if (this.state.isBotTyping) {
      html += `
        <div class="flex items-start gap-2.5">
          <div class="w-6 h-6 rounded-full bg-emerald-600 text-white flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5">
            <i data-lucide="bot" class="w-3.5 h-3.5"></i>
          </div>
          <div class="bg-white border border-slate-200/80 px-3.5 py-2 rounded-2xl flex items-center gap-1 shadow-xs">
            <span class="w-1.5 h-1.5 rounded-full bg-slate-400 animate-bounce"></span>
            <span class="w-1.5 h-1.5 rounded-full bg-slate-400 animate-bounce" style="animation-delay: 0.15s"></span>
            <span class="w-1.5 h-1.5 rounded-full bg-slate-400 animate-bounce" style="animation-delay: 0.3s"></span>
          </div>
        </div>
      `;
    }

    container.innerHTML = html;
    if (window.lucide) window.lucide.createIcons();
    container.scrollTop = container.scrollHeight;
  },

  async startBulkSend() {
    const recText = document.getElementById('recipients-input').value;
    const msg = document.getElementById('bulk-message-input').value.trim();
    const delay = parseInt(document.getElementById('delay-slider').value) || 2;
    const consent = document.getElementById('consent-checkbox').checked;

    const recipients = recText
      .split(/[\n,]/)
      .map(r => r.trim())
      .filter(r => r.length > 0);

    if (recipients.length === 0) {
      Toast.warning('Ingresa al menos un número de teléfono destinatario.');
      return;
    }
    if (!msg) {
      Toast.warning('El mensaje a enviar no puede estar vacío.');
      return;
    }
    if (!consent) {
      Toast.warning('Debes confirmar el consentimiento de los destinatarios para continuar.');
      return;
    }

    try {
      const res = await fetch('/api/whatsapp/send', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          recipients,
          message: msg,
          delaySec: delay,
          consentAck: true,
        }),
      });

      if (!res.ok) throw new Error('Error al programar envío');
      const data = await res.json();
      Toast.success(data.message);
      this.fetchLogs();
    } catch (e) {
      Toast.error('Fallo al programar envío: ' + e.message);
    }
  },

  startPolling() {
    this.pollIntervalId = setInterval(async () => {
      try {
        const res = await fetch('/api/whatsapp/status');
        if (res.ok) {
          const data = await res.json();
          this.updateStatusUI(data);
        }
        if (this.state.activeTab === 'send') {
          this.fetchLogs();
        }
      } catch (e) {
        // Silent poll error
      }
    }, 3000);

    this.cleanupFns.push(() => {
      if (this.pollIntervalId) {
        clearInterval(this.pollIntervalId);
        this.pollIntervalId = null;
      }
    });
  },

  updateStatusUI(data) {
    const queueElem = document.getElementById('metric-queue');
    const sentElem = document.getElementById('metric-sent');
    if (queueElem) queueElem.textContent = data.pendingQueue || 0;
    if (sentElem) sentElem.textContent = data.totalSentToday || 0;
  },

  async fetchLogs() {
    const container = document.getElementById('logs-container');
    if (!container) return;

    try {
      const res = await fetch('/api/whatsapp/logs');
      if (!res.ok) return;
      const logs = await res.json();

      if (logs.length === 0) {
        this.state.latestLogs = [];
        container.innerHTML = '<div class="p-6 text-center text-xs text-slate-400">No hay envíos registrados recientemente.</div>';
        return;
      }

      this.state.latestLogs = logs;
      container.innerHTML = logs.map(l => {
        const isSuccess = l.status === 'Entregado';
        const dateStr = new Date(l.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        return `
          <div class="p-2.5 rounded-xl border border-slate-100 bg-slate-50/50 flex items-center justify-between text-xs">
            <div class="space-y-0.5">
              <span class="font-mono font-bold text-slate-800">${this.escapeHtml(l.recipient)}</span>
              <p class="text-[11px] text-slate-500 truncate max-w-xs">${this.escapeHtml(l.message)}</p>
            </div>
            <div class="text-right">
              <span class="inline-flex items-center gap-1 font-semibold text-[10px] px-2 py-0.5 rounded-full ${
                isSuccess ? 'bg-emerald-50 text-emerald-700' : 'bg-rose-50 text-rose-700'
              }">
                <span class="w-1.5 h-1.5 rounded-full ${isSuccess ? 'bg-emerald-500' : 'bg-rose-500'}"></span>
                ${l.status}
              </span>
              <p class="text-[9px] text-slate-400 mt-0.5">${dateStr}</p>
            </div>
          </div>
        `;
      }).join('');
    } catch (e) {
      console.error(e);
    }
  },

  exportLogsToCSV() {
    const logs = this.state.latestLogs || [];
    if (logs.length === 0) {
      Toast.warning('No hay registros de envíos en la bitácora para exportar.');
      return;
    }

    let csvContent = 'ID,Destinatario,Mensaje,Estado,Detalle,Fecha\r\n';
    logs.forEach(l => {
      const id = l.id || '';
      const recipient = `"${(l.recipient || '').replace(/"/g, '""')}"`;
      const message = `"${(l.message || '').replace(/"/g, '""')}"`;
      const status = `"${(l.status || '').replace(/"/g, '""')}"`;
      const errorDetail = `"${(l.errorDetail || '').replace(/"/g, '""')}"`;
      const date = `"${new Date(l.createdAt).toLocaleString('es-CO')}"`;
      csvContent += `${id},${recipient},${message},${status},${errorDetail},${date}\r\n`;
    });

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `bitacora_whatsapp_${new Date().toISOString().slice(0, 10)}.csv`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
    Toast.success('Bitácora descargada en formato CSV.');
  },

  confirmClearLogs() {
    Modal.show({
      title: '¿Vaciar Bitácora de Envíos?',
      content: 'Esta acción eliminará todos los registros históricos de mensajes enviados de la base de datos.',
      confirmText: 'Sí, Vaciar',
      cancelText: 'Cancelar',
      showCancel: true,
      icon: 'trash-2',
      iconColor: 'text-rose-600 bg-rose-50',
      onConfirm: async () => {
        try {
          const res = await fetch('/api/whatsapp/logs', { method: 'DELETE' });
          if (!res.ok) throw new Error('Error al vaciar registros');
          Toast.success('Bitácora vaciada correctamente.');
          await this.fetchLogs();
        } catch (e) {
          Toast.error(e.message);
        }
      }
    });
  },

  processRecipientsExcelFile(file) {
    if (!window.XLSX) {
      Toast.error('Biblioteca SheetJS no cargada en el navegador.');
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

        if (!json || json.length < 2) {
          Toast.warning('El archivo Excel o CSV no contiene suficientes filas de datos.');
          return;
        }

        const headers = json[0].map(h => String(h || '').trim().toLowerCase());
        const phoneKeywords = ['telefono', 'teléfono', 'celular', 'phone', 'whatsapp', 'tel', 'movil', 'móvil', 'numero', 'número'];
        let matchedCol = -1;

        headers.forEach((h, idx) => {
          if (phoneKeywords.some(k => h.includes(k)) && matchedCol === -1) {
            matchedCol = idx;
          }
        });

        const numbers = [];
        for (let i = 1; i < json.length; i++) {
          const row = json[i];
          if (!row) continue;

          let rawVal = '';
          if (matchedCol !== -1 && row[matchedCol]) {
            rawVal = String(row[matchedCol]).trim();
          } else {
            for (let c = 0; c < row.length; c++) {
              const val = String(row[c] || '').trim();
              if (val.replace(/[^0-9]/g, '').length >= 7) {
                rawVal = val;
                break;
              }
            }
          }

          if (rawVal) {
            const cleaned = rawVal.replace(/[^0-9+]/g, '');
            if (cleaned.length >= 7) {
              numbers.push(cleaned);
            }
          }
        }

        if (numbers.length === 0) {
          Toast.warning('No se detectaron columnas con números de teléfono válidos en el archivo.');
          return;
        }

        const input = document.getElementById('recipients-input');
        const currentText = input.value.trim();
        const combined = currentText ? `${currentText}\n${numbers.join('\n')}` : numbers.join('\n');
        input.value = combined;
        Toast.success(`Se importaron ${numbers.length} números de teléfono desde el archivo.`);
      } catch (err) {
        Toast.error('Error al procesar archivo: ' + err.message);
      }
    };
    reader.readAsArrayBuffer(file);
  },

  escapeHtml(str) {
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  },

  destroy() {
    this.cleanupFns.forEach(fn => fn());
    this.cleanupFns = [];
    if (this.pollIntervalId) {
      clearInterval(this.pollIntervalId);
      this.pollIntervalId = null;
    }
  }
};
