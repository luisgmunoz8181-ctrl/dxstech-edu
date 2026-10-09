import { Toast } from '../components/toast.js';
import { Modal } from '../components/modal.js';
import { Loading } from '../components/loading.js';
import { EmptyState } from '../components/empty-state.js';
import { esc } from '../utils/escape.js';

export const QuizzesView = {
  state: {
    quizzesList: [],
    activeQuiz: null,
    currentQuestionIndex: 0,
    selectedOption: null,
    score: 0,
    userAnswers: [],
    isReviewMode: false,
    soundEnabled: true,
  },

  cleanupFns: [],

  // Native Web Audio Synthesizer for feedback effects without external files
  playChime(isCorrect) {
    if (this.state.soundEnabled === false) return;
    try {
      const audioCtx = new (window.AudioContext || window.webkitAudioContext)();
      const osc = audioCtx.createOscillator();
      const gain = audioCtx.createGain();
      osc.connect(gain);
      gain.connect(audioCtx.destination);

      if (isCorrect) {
        osc.type = 'sine';
        osc.frequency.setValueAtTime(587.33, audioCtx.currentTime); // D5
        osc.frequency.exponentialRampToValueAtTime(880, audioCtx.currentTime + 0.15); // A5
        gain.gain.setValueAtTime(0.12, audioCtx.currentTime);
        gain.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.3);
        osc.start();
        osc.stop(audioCtx.currentTime + 0.3);
      } else {
        osc.type = 'triangle';
        osc.frequency.setValueAtTime(320, audioCtx.currentTime);
        osc.frequency.exponentialRampToValueAtTime(220, audioCtx.currentTime + 0.2);
        gain.gain.setValueAtTime(0.12, audioCtx.currentTime);
        gain.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.25);
        osc.start();
        osc.stop(audioCtx.currentTime + 0.25);
      }
    } catch (e) {
      // AudioContext policy might require user interaction or be disabled; safe ignore
    }
  },

  render() {
    return `
      <div class="space-y-6 max-w-6xl mx-auto">
        <!-- Header Banner -->
        <div class="bg-gradient-to-r from-purple-600 to-indigo-700 rounded-2xl p-6 text-white shadow-lg shadow-purple-100 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h3 class="text-lg font-bold">Evaluaciones Inteligentes con Gemini</h3>
            <p class="text-xs text-purple-100 mt-1 max-w-xl">
              Genera exámenes pedagógicos de opción múltiple a partir de cualquier tema o material de clase, y pruébalos en un simulador interactivo en tiempo real.
            </p>
          </div>
          <div class="flex items-center gap-2">
            <button id="load-sample-quiz-btn" class="flex items-center gap-1.5 text-xs bg-white/20 hover:bg-white/30 text-white px-3.5 py-2.5 rounded-xl font-semibold transition-colors backdrop-blur-md border border-white/20">
              <i data-lucide="book-open" class="w-4 h-4"></i>
              <span>Cargar Tema Demo</span>
            </button>
            <button id="quick-create-btn" class="flex items-center gap-2 text-xs bg-white text-indigo-700 px-4 py-2.5 rounded-xl font-bold hover:bg-purple-50 transition-colors shadow-sm">
              <i data-lucide="plus" class="w-4 h-4"></i>
              <span>Nueva Evaluación</span>
            </button>
          </div>
        </div>

        <!-- Main Workspace: Form & Simulator Container -->
        <div id="quiz-workspace" class="space-y-6">
          <!-- Quiz Creation Form Card -->
          <div id="quiz-create-card" class="bg-white rounded-2xl p-6 border border-slate-200/80 shadow-sm space-y-5">
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-bold text-slate-800 flex items-center gap-2">
                <i data-lucide="sparkles" class="w-4 h-4 text-purple-600"></i>
                Diseñar Nueva Evaluación
              </h4>
              <span class="text-[11px] text-purple-600 bg-purple-50 px-2.5 py-1 rounded-lg font-medium border border-purple-100">
                Google Gemini Flash 2.5
              </span>
            </div>

            <div class="space-y-4">
              <div>
                <label class="block text-xs font-semibold text-slate-700 mb-1">Título o Tema de la Evaluación</label>
                <input type="text" id="quiz-title-input" placeholder="Ej: Fundamentos de Arquitectura de Software y APIs REST en Go" class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-purple-500 font-medium">
              </div>

              <!-- Link to LMS Course (Optional) -->
              <div>
                <label class="block text-xs font-semibold text-slate-700 mb-1">Vincular a Curso del LMS (Opcional)</label>
                <select id="quiz-course-select" class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-purple-500 font-medium">
                  <option value="">Ninguno / Evaluación General</option>
                </select>
              </div>

              <!-- Question count slider (3-15) -->
              <div>
                <div class="flex items-center justify-between mb-1.5">
                  <label class="text-xs font-semibold text-slate-700">Cantidad de Preguntas</label>
                  <span id="slider-count-badge" class="text-xs font-bold text-purple-700 bg-purple-50 px-2.5 py-0.5 rounded-lg border border-purple-200">5 preguntas</span>
                </div>
                <input type="range" id="quiz-count-slider" min="3" max="15" value="5" step="1" class="w-full accent-purple-600 cursor-pointer">
                <div class="flex justify-between text-[10px] text-slate-400 mt-1">
                  <span>Mínimo 3</span>
                  <span>5 rápido</span>
                  <span>10 estándar</span>
                  <span>Máximo 15</span>
                </div>
              </div>

              <!-- Notes / Study Material -->
              <div>
                <label class="block text-xs font-semibold text-slate-700 mb-1">Material de Estudio / Apuntes del Profesor</label>
                <textarea id="quiz-notes-input" rows="6" placeholder="Pega aquí el contenido de la clase, resumen, artículos o apuntes que Gemini debe analizar para formular las preguntas pedagógicas..." class="w-full p-3 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-purple-500 font-sans leading-relaxed resize-none"></textarea>
              </div>

              <div class="flex items-center justify-end gap-3 pt-2">
                <button id="generate-quiz-btn" class="py-2.5 px-6 rounded-xl text-xs font-bold text-white bg-purple-600 hover:bg-purple-700 shadow-md shadow-purple-200 transition-all flex items-center gap-2">
                  <i data-lucide="bot" class="w-4 h-4"></i>
                  <span>Generar Evaluación con Gemini</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Interactive Simulator Container (Initially Hidden) -->
          <div id="quiz-simulator-card" class="hidden bg-white rounded-2xl p-6 border border-slate-200/80 shadow-md space-y-6">
            <!-- Simulator Top Bar -->
            <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between border-b border-slate-100 pb-4 gap-3">
              <div>
                <span class="text-[11px] font-semibold text-purple-600 uppercase tracking-wider">Simulador de Evaluación</span>
                <h3 id="sim-quiz-title" class="text-base font-bold text-slate-900"></h3>
              </div>
              <div class="flex items-center gap-2">
                <button id="sim-sound-toggle-btn" class="text-xs text-purple-700 bg-purple-50 hover:bg-purple-100 font-medium flex items-center gap-1.5 px-3 py-1.5 rounded-lg transition-colors" title="Activar o silenciar sonido">
                  <i data-lucide="volume-2" class="w-3.5 h-3.5" id="sound-icon"></i>
                  <span id="sound-label">Sonido</span>
                </button>
                <button id="sim-exit-btn" class="text-xs text-slate-500 hover:text-slate-800 font-medium flex items-center gap-1.5 px-3 py-1.5 rounded-lg hover:bg-slate-100 transition-colors">
                  <i data-lucide="x" class="w-4 h-4"></i>
                  <span>Salir</span>
                </button>
              </div>
            </div>

            <!-- Progress Bar -->
            <div>
              <div class="flex items-center justify-between text-xs font-medium text-slate-500 mb-1.5">
                <span id="sim-progress-text">Pregunta 1 de 5</span>
                <span id="sim-progress-pct">20%</span>
              </div>
              <div class="w-full h-2 bg-slate-100 rounded-full overflow-hidden">
                <div id="sim-progress-bar" class="h-full bg-purple-600 transition-all duration-300 w-1/5"></div>
              </div>
            </div>

            <!-- Active Question Body -->
            <div id="sim-question-area" class="space-y-4">
              <div class="p-4 rounded-xl bg-purple-50/50 border border-purple-100">
                <p id="sim-question-text" class="text-sm font-semibold text-slate-800 leading-relaxed"></p>
              </div>

              <!-- Options Grid -->
              <div id="sim-options-list" class="grid grid-cols-1 gap-2.5">
                <!-- Injected options -->
              </div>
            </div>

            <!-- Simulator Footer & Navigation -->
            <div class="flex items-center justify-between pt-4 border-t border-slate-100">
              <span id="sim-feedback-badge" class="text-xs font-semibold"></span>
              <button id="sim-next-btn" class="hidden px-5 py-2 rounded-xl text-xs font-bold text-white bg-purple-600 hover:bg-purple-700 shadow-sm transition-all flex items-center gap-1.5">
                <span>Siguiente Pregunta</span>
                <i data-lucide="arrow-right" class="w-4 h-4"></i>
              </button>
            </div>
          </div>

          <!-- Simulator Final Score Card (Initially Hidden) -->
          <div id="quiz-result-card" class="hidden bg-white rounded-2xl p-8 border border-slate-200/80 shadow-md text-center space-y-6">
            <div id="result-icon-container" class="w-16 h-16 rounded-2xl mx-auto flex items-center justify-center"></div>
            <div>
              <h3 id="result-title" class="text-xl font-bold text-slate-800"></h3>
              <p id="result-desc" class="text-xs text-slate-500 mt-1 max-w-md mx-auto"></p>
            </div>

            <!-- Score Display -->
            <div class="inline-flex items-center gap-4 p-4 rounded-2xl bg-slate-50 border border-slate-200/80">
              <div class="text-center px-4 border-r border-slate-200">
                <p class="text-[11px] font-semibold text-slate-400">Puntaje Obtenido</p>
                <p id="result-score" class="text-2xl font-extrabold text-slate-900"></p>
              </div>
              <div class="text-center px-4">
                <p class="text-[11px] font-semibold text-slate-400">Estado</p>
                <p id="result-badge" class="text-sm font-bold mt-1"></p>
              </div>
            </div>

            <!-- Answers Review Accordion -->
            <div id="result-answers-review" class="max-w-xl mx-auto text-left space-y-3 pt-3">
              <h5 class="text-xs font-bold text-slate-700 uppercase tracking-wider">Desglose de Respuestas:</h5>
              <div id="result-answers-list" class="space-y-2"></div>
            </div>

            <div class="flex items-center justify-center gap-3 pt-4 border-t border-slate-100">
              <button id="result-retry-btn" class="px-5 py-2.5 rounded-xl text-xs font-bold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center gap-2">
                <i data-lucide="rotate-ccw" class="w-4 h-4"></i>
                <span>Reintentar Examen</span>
              </button>
              <button id="result-back-btn" class="px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-purple-600 hover:bg-purple-700 transition-colors flex items-center gap-2">
                <i data-lucide="check" class="w-4 h-4"></i>
                <span>Volver a Mis Evaluaciones</span>
              </button>
            </div>
          </div>

          <!-- Saved Quizzes History List -->
          <div class="bg-white rounded-2xl p-6 border border-slate-200/80 shadow-sm space-y-4">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <h4 class="text-sm font-bold text-slate-800 flex items-center gap-2">
                <i data-lucide="history" class="w-4 h-4 text-purple-600"></i>
                Evaluaciones Guardadas en SQLite
              </h4>
              <div class="flex items-center gap-2">
                <input type="text" id="search-quizzes-input" placeholder="Buscar evaluación..." class="px-3 py-1.5 text-xs bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-purple-500 w-44 sm:w-56">
                <button id="refresh-quizzes-btn" class="flex items-center gap-1.5 text-xs text-slate-600 hover:text-purple-600 font-semibold px-3 py-1.5 rounded-lg bg-slate-50 hover:bg-purple-50 transition-colors shrink-0">
                  <i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i>
                  <span>Actualizar</span>
                </button>
              </div>
            </div>

            <div id="quizzes-table-container">
              ${Loading.spinner('Cargando evaluaciones...')}
            </div>
          </div>
        </div>
      </div>
    `;
  },

  mount() {
    this.bindEvents();
    this.fetchQuizzes();
    this.loadCourseOptions();
  },

  async loadCourseOptions() {
    try {
      const res = await fetch('/api/courses');
      if (!res.ok) return;
      const courses = await res.json();
      const select = document.getElementById('quiz-course-select');
      if (!select) return;

      courses.forEach(c => {
        const opt = document.createElement('option');
        opt.value = c.id;
        opt.textContent = `${esc(c.title)} (${esc(c.code)})`;
        select.appendChild(opt);
      });
    } catch {
      // Safe ignore
    }
  },

  bindEvents() {
    // Slider
    const slider = document.getElementById('quiz-count-slider');
    const badge = document.getElementById('slider-count-badge');
    slider.addEventListener('input', (e) => {
      badge.textContent = `${e.target.value} preguntas`;
    });

    // Demo Topic Loader
    document.getElementById('load-sample-quiz-btn').addEventListener('click', () => {
      document.getElementById('quiz-title-input').value = 'Fundamentos de REST APIs, Microservicios y Concurrencia en Go';
      document.getElementById('quiz-notes-input').value = 
`Las APIs REST son un estilo arquitectónico basado en HTTP que utiliza verbos estándar (GET, POST, PUT, DELETE) y recursos representados mediante URLs uniformes. En Go, frameworks como Gin proporcionan enrutamiento de alta velocidad y middleware eficiente.
La concurrencia en Go se basa en Goroutines (hilos livianos multiplexados por el runtime de Go) y Canales (Channels) para la comunicación sin memoria compartida (patrón CSP).
Para persistencia, SQLite en modo WAL (Write-Ahead Logging) permite lecturas concurrentes sin bloquear escrituras.
Los códigos de respuesta HTTP esenciales son: 200 OK, 201 Created, 400 Bad Request, 401 Unauthorized, 403 Forbidden y 404 Not Found.`;
      document.getElementById('quiz-count-slider').value = 5;
      badge.textContent = '5 preguntas';
      Toast.info('Tema y material de estudio de ejemplo cargados en el formulario.');
      document.getElementById('quiz-create-card').classList.remove('hidden');
      document.getElementById('quiz-create-card').scrollIntoView({ behavior: 'smooth' });
    });

    // Generate Button
    document.getElementById('generate-quiz-btn').addEventListener('click', () => {
      this.generateQuiz();
    });

    // Quick create button scroll to form
    document.getElementById('quick-create-btn').addEventListener('click', () => {
      document.getElementById('quiz-create-card').classList.remove('hidden');
      document.getElementById('quiz-simulator-card').classList.add('hidden');
      document.getElementById('quiz-result-card').classList.add('hidden');
      document.getElementById('quiz-title-input').focus();
    });

    // Simulator Exit
    document.getElementById('sim-exit-btn').addEventListener('click', () => {
      this.exitSimulator();
    });

    // Sound toggle
    document.getElementById('sim-sound-toggle-btn')?.addEventListener('click', () => {
      this.state.soundEnabled = !this.state.soundEnabled;
      const icon = document.getElementById('sound-icon');
      const label = document.getElementById('sound-label');
      const btn = document.getElementById('sim-sound-toggle-btn');
      if (this.state.soundEnabled) {
        icon?.setAttribute('data-lucide', 'volume-2');
        if (label) label.textContent = 'Sonido';
        btn?.classList.remove('text-slate-500', 'bg-slate-100');
        btn?.classList.add('text-purple-700', 'bg-purple-50');
        Toast.info('Efectos de sonido activados');
      } else {
        icon?.setAttribute('data-lucide', 'volume-x');
        if (label) label.textContent = 'Silenciado';
        btn?.classList.remove('text-purple-700', 'bg-purple-50');
        btn?.classList.add('text-slate-500', 'bg-slate-100');
        Toast.info('Efectos de sonido silenciados');
      }
      if (window.lucide) window.lucide.createIcons();
    });

    // Simulator Next
    document.getElementById('sim-next-btn').addEventListener('click', () => {
      this.nextQuestion();
    });

    // Result buttons
    document.getElementById('result-retry-btn').addEventListener('click', () => {
      if (this.state.activeQuiz) {
        this.startSimulator(this.state.activeQuiz);
      }
    });

    document.getElementById('result-back-btn').addEventListener('click', () => {
      this.exitSimulator();
    });

    document.getElementById('refresh-quizzes-btn')?.addEventListener('click', () => {
      this.fetchQuizzes();
    });

    document.getElementById('search-quizzes-input')?.addEventListener('input', (e) => {
      this.filterQuizzes(e.target.value);
    });
  },

  async generateQuiz() {
    const title = document.getElementById('quiz-title-input').value.trim();
    const notes = document.getElementById('quiz-notes-input').value.trim();
    const count = parseInt(document.getElementById('quiz-count-slider').value) || 5;

    if (!title) {
      Toast.warning('Por favor ingresa un título para la evaluación.');
      document.getElementById('quiz-title-input').focus();
      return;
    }
    if (!notes) {
      Toast.warning('Debes ingresar el material de estudio o apuntes.');
      document.getElementById('quiz-notes-input').focus();
      return;
    }

    // Check Gemini API Key in browser
    let apiKey = localStorage.getItem('dxstech_gemini_api_key');
    if (!apiKey) {
      Modal.promptGeminiKey({
        onSaved: (key) => {
          this.callGenerateQuizAPI(title, notes, count, key);
        }
      });
      return;
    }

    await this.callGenerateQuizAPI(title, notes, count, apiKey);
  },

  async callGenerateQuizAPI(title, notes, count, apiKey) {
    Loading.overlay(`Gemini está formulando ${count} preguntas pedagógicas...`);

    try {
      const courseId = document.getElementById('quiz-course-select')?.value || '';
      const response = await fetch('/api/quizzes/generate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Gemini-API-Key': apiKey,
        },
        body: JSON.stringify({ title, notes, count, courseId }),
      });

      if (!response.ok) {
        const err = await response.json().catch(() => ({}));
        throw new Error(err.error || `Error HTTP ${response.status}`);
      }

      const newQuiz = await response.json();
      Toast.success('¡Evaluación generada con éxito por Gemini!');

      // Reset form
      document.getElementById('quiz-title-input').value = '';
      document.getElementById('quiz-notes-input').value = '';

      // Refresh list and start simulator immediately
      await this.fetchQuizzes();
      this.startSimulator(newQuiz);
    } catch (err) {
      Toast.error('Fallo en generación: ' + err.message);
    } finally {
      Loading.hideOverlay();
    }
  },

  async fetchQuizzes() {
    const container = document.getElementById('quizzes-table-container');
    try {
      const response = await fetch('/api/quizzes');
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const list = await response.json();
      this.state.quizzesList = list;
      this.renderQuizzesTable(list);
    } catch (err) {
      container.innerHTML = `<div class="p-4 text-xs text-rose-600 bg-rose-50 rounded-xl">Error cargando evaluaciones: ${esc(err.message)}</div>`;
    }
  },

  filterQuizzes(query) {
    if (!this.state.quizzesList) return;
    const q = (query || '').toLowerCase().trim();
    if (!q) {
      this.renderQuizzesTable(this.state.quizzesList);
      return;
    }
    const filtered = this.state.quizzesList.filter(item => {
      const title = (item.title || '').toLowerCase();
      return title.includes(q);
    });
    this.renderQuizzesTable(filtered, true);
  },

  renderQuizzesTable(quizzes, isFiltered = false) {
    const container = document.getElementById('quizzes-table-container');
    if (!quizzes || quizzes.length === 0) {
      if (isFiltered) {
        container.innerHTML = `
          <div class="py-8 text-center bg-slate-50/50 rounded-xl border border-dashed border-slate-200">
            <i data-lucide="search" class="w-8 h-8 text-slate-300 mx-auto mb-2"></i>
            <p class="text-xs font-semibold text-slate-600">No se encontraron evaluaciones coincidentes</p>
            <p class="text-[11px] text-slate-400 mt-0.5">Prueba con otro término de búsqueda o limpia el filtro.</p>
          </div>
        `;
      } else {
        container.innerHTML = EmptyState.render({
          icon: 'book-open',
          title: 'No hay evaluaciones creadas',
          description: 'Formula tu primera evaluación con IA utilizando el formulario superior.',
        });
      }
      if (window.lucide) window.lucide.createIcons();
      return;
    }

    let html = `
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200/80 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
              <th class="py-3 px-4">Título</th>
              <th class="py-3 px-4">Preguntas</th>
              <th class="py-3 px-4">Fecha</th>
              <th class="py-3 px-4 text-right">Acción</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs text-slate-700">
    `;

    quizzes.forEach(q => {
      const dateStr = new Date(q.createdAt).toLocaleDateString('es-CO', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });

      html += `
        <tr class="hover:bg-slate-50/80 transition-colors">
          <td class="py-3 px-4 font-semibold text-slate-800 max-w-xs truncate">
            ${this.escapeHtml(q.title)}
            ${q.courseId ? `<span class="inline-block bg-indigo-50 text-indigo-700 font-semibold px-1.5 py-0.2 rounded text-[10px] ml-1.5 border border-indigo-200">🎓 Curso</span>` : ''}
          </td>
          <td class="py-3 px-4">
            <span class="bg-purple-50 text-purple-700 font-semibold px-2 py-0.5 rounded-md text-[11px]">
              ${q.questionCount} preguntas
            </span>
          </td>
          <td class="py-3 px-4 text-slate-400">${dateStr}</td>
          <td class="py-3 px-4 text-right space-x-1.5">
            <button data-quiz-id="${esc(q.id)}" class="take-quiz-btn inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-indigo-50 hover:bg-indigo-100 text-indigo-700 font-bold text-xs transition-colors">
              <i data-lucide="play" class="w-3.5 h-3.5"></i>
              <span>Simular</span>
            </button>
            <button data-print-id="${esc(q.id)}" class="print-quiz-btn inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold text-xs transition-colors" title="Imprimir hoja de examen para clase">
              <i data-lucide="printer" class="w-3.5 h-3.5 text-slate-500"></i>
              <span>Imprimir</span>
            </button>
            <button data-delete-id="${esc(q.id)}" class="delete-quiz-btn inline-flex items-center gap-1 px-2 py-1.5 rounded-lg text-rose-500 hover:bg-rose-50 transition-colors" title="Eliminar examen">
              <i data-lucide="trash-2" class="w-3.5 h-3.5"></i>
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

    container.querySelectorAll('.take-quiz-btn').forEach(btn => {
      btn.addEventListener('click', async () => {
        const id = btn.dataset.quizId;
        await this.loadAndStartQuiz(id);
      });
    });

    container.querySelectorAll('.print-quiz-btn').forEach(btn => {
      btn.addEventListener('click', async () => {
        const id = btn.dataset.printId;
        await this.openPrintModal(id);
      });
    });

    container.querySelectorAll('.delete-quiz-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const id = btn.dataset.deleteId;
        this.confirmDeleteQuiz(id);
      });
    });
  },

  async loadAndStartQuiz(quizId) {
    Loading.overlay('Cargando examen...');
    try {
      const res = await fetch(`/api/quizzes/${quizId}`);
      if (!res.ok) throw new Error('Evaluación no encontrada');
      const quiz = await res.json();
      this.startSimulator(quiz);
    } catch (err) {
      Toast.error(err.message);
    } finally {
      Loading.hideOverlay();
    }
  },

  async openPrintModal(quizId) {
    Loading.overlay('Preparando formato de impresión...');
    try {
      const res = await fetch(`/api/quizzes/${quizId}`);
      if (!res.ok) throw new Error('Evaluación no encontrada');
      const quiz = await res.json();

      let questionsHtml = '';
      quiz.questions.forEach((q, idx) => {
        questionsHtml += `
          <div class="mb-4 text-xs">
            <p class="font-bold text-slate-800 mb-1.5">${idx + 1}. ${this.escapeHtml(q.question)}</p>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-1.5 pl-2 text-slate-700">
              ${q.options.map((opt, oIdx) => `
                <div class="flex items-start gap-1.5">
                  <span class="inline-block w-4 h-4 rounded border border-slate-400 text-[10px] text-center font-bold font-mono shrink-0">${String.fromCharCode(65 + oIdx)}</span>
                  <span>${this.escapeHtml(opt)}</span>
                </div>
              `).join('')}
            </div>
          </div>
        `;
      });

      const modalContent = `
        <div class="space-y-4">
          <div class="flex items-center justify-between pb-3 border-b border-slate-200">
            <div>
              <p class="text-[11px] font-bold text-purple-600 uppercase tracking-wider">DxSTech Edu — Hoja de Evaluación</p>
              <h4 class="text-sm font-bold text-slate-900">${this.escapeHtml(quiz.title)}</h4>
            </div>
            <span class="text-[11px] text-slate-500 bg-slate-100 px-2 py-0.5 rounded font-medium">${quiz.questionCount} reactivos</span>
          </div>

          <div class="p-3 bg-slate-50 rounded-xl border border-slate-200 text-xs text-slate-600 grid grid-cols-2 gap-2">
            <div><strong>Alumno:</strong> ____________________________________</div>
            <div><strong>Fecha:</strong> ________________</div>
            <div><strong>Curso:</strong> ____________________________________</div>
            <div><strong>Nota:</strong> _____ / 100</div>
          </div>

          <div class="overflow-y-auto max-h-[45vh] pr-2 space-y-2 border-y border-slate-100 py-3">
            ${questionsHtml}
          </div>

          <div class="flex items-center justify-between pt-1">
            <button id="print-now-btn" class="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-xl text-xs font-bold transition-colors flex items-center gap-1.5 shadow-sm">
              <i data-lucide="printer" class="w-3.5 h-3.5"></i>
              <span>Imprimir / Guardar PDF</span>
            </button>
            <button id="copy-exam-text-btn" class="px-3.5 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl text-xs font-semibold transition-colors flex items-center gap-1.5">
              <i data-lucide="copy" class="w-3.5 h-3.5 text-slate-500"></i>
              <span>Copiar Texto</span>
            </button>
          </div>
        </div>
      `;

      Modal.show({
        title: 'Hoja de Evaluación para Imprimir',
        content: modalContent,
        confirmText: 'Cerrar',
        showCancel: false,
        icon: 'printer',
        iconColor: 'text-purple-600 bg-purple-50',
        maxWidth: 'max-w-2xl',
      });

      // Bind print now
      document.getElementById('print-now-btn')?.addEventListener('click', () => {
        const printWindow = window.open('', '_blank');
        if (!printWindow) {
          window.print();
          return;
        }

        printWindow.document.write(`
          <!DOCTYPE html>
          <html>
          <head>
            <title>${this.escapeHtml(quiz.title)} - Evaluación</title>
            <style>
              body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; padding: 25px; color: #1e293b; max-width: 800px; margin: 0 auto; line-height: 1.5; font-size: 13px; }
              .header { border-bottom: 2px solid #334155; padding-bottom: 12px; margin-bottom: 16px; }
              .header h1 { margin: 0 0 4px 0; font-size: 18px; color: #0f172a; }
              .header p { margin: 0; font-size: 12px; color: #64748b; }
              .fields { display: flex; justify-content: space-between; border: 1px solid #cbd5e1; border-radius: 6px; padding: 10px 14px; margin-bottom: 20px; font-size: 12px; }
              .q-block { margin-bottom: 16px; page-break-inside: avoid; }
              .q-title { font-weight: bold; margin-bottom: 6px; }
              .opts { margin-left: 12px; }
              .opt-item { margin-bottom: 4px; display: flex; align-items: center; }
              .opt-box { width: 14px; height: 14px; border: 1px solid #475569; display: inline-block; margin-right: 8px; border-radius: 2px; }
              @media print { button { display: none; } }
            </style>
          </head>
          <body>
            <div class="header">
              <h1>DxSTech Edu — Evaluación</h1>
              <p>${this.escapeHtml(quiz.title)}</p>
            </div>
            <div class="fields">
              <div>Estudiante: _____________________________________</div>
              <div>Fecha: _______________</div>
              <div>Calificación: _______</div>
            </div>
            ${quiz.questions.map((q, idx) => `
              <div class="q-block">
                <div class="q-title">${idx + 1}. ${this.escapeHtml(q.question)}</div>
                <div class="opts">
                  ${q.options.map((opt, oIdx) => `
                    <div class="opt-item">
                      <span class="opt-box"></span>
                      <strong>${String.fromCharCode(65 + oIdx)})</strong>&nbsp;${this.escapeHtml(opt)}
                    </div>
                  `).join('')}
                </div>
              </div>
            `).join('')}
            <script>
              window.onload = function() { window.print(); };
            <\/script>
          </body>
          </html>
        `);
        printWindow.document.close();
      });

      // Bind copy text
      document.getElementById('copy-exam-text-btn')?.addEventListener('click', () => {
        let plainText = `DxSTech Edu — Evaluación\n${esc(quiz.title)}\nEstudiante: __________________  Fecha: ___________\n\n`;
        quiz.questions.forEach((q, idx) => {
          plainText += `${idx + 1}. ${esc(q.question)}\n`;
          q.options.forEach((opt, oIdx) => {
            plainText += `   [ ] ${String.fromCharCode(65 + oIdx)}) ${esc(opt)}\n`;
          });
          plainText += '\n';
        });

        navigator.clipboard.writeText(plainText).then(() => {
          Toast.success('Cuestionario copiado como texto al portapapeles.');
        });
      });

    } catch (err) {
      Toast.error(err.message);
    } finally {
      Loading.hideOverlay();
    }
  },

  confirmDeleteQuiz(quizId) {
    Modal.show({
      title: '¿Eliminar Evaluación?',
      content: 'Esta acción borrará de forma permanente el examen y su historial de preguntas de la base de datos SQLite.',
      confirmText: 'Sí, Eliminar',
      cancelText: 'Cancelar',
      showCancel: true,
      icon: 'trash-2',
      iconColor: 'text-rose-600 bg-rose-50',
      onConfirm: async () => {
        try {
          const res = await fetch(`/api/quizzes/${quizId}`, { method: 'DELETE' });
          if (!res.ok) throw new Error('Error al eliminar');
          Toast.success('Evaluación eliminada correctamente.');
          await this.fetchQuizzes();
        } catch (err) {
          Toast.error(err.message);
        }
      }
    });
  },

  startSimulator(quiz) {
    this.state.activeQuiz = quiz;
    this.state.currentQuestionIndex = 0;
    this.state.score = 0;
    this.state.userAnswers = [];
    this.state.isReviewMode = false;

    document.getElementById('quiz-create-card').classList.add('hidden');
    document.getElementById('quiz-result-card').classList.add('hidden');
    const simCard = document.getElementById('quiz-simulator-card');
    simCard.classList.remove('hidden');

    document.getElementById('sim-quiz-title').textContent = quiz.title;
    this.renderSimulatorQuestion();

    simCard.scrollIntoView({ behavior: 'smooth' });
  },

  renderSimulatorQuestion() {
    const quiz = this.state.activeQuiz;
    const idx = this.state.currentQuestionIndex;
    const total = quiz.questions.length;
    const q = quiz.questions[idx];

    // Update Progress
    const pct = Math.round(((idx + 1) / total) * 100);
    document.getElementById('sim-progress-text').textContent = `Pregunta ${idx + 1} de ${total}`;
    document.getElementById('sim-progress-pct').textContent = `${pct}%`;
    document.getElementById('sim-progress-bar').style.width = `${pct}%`;

    // Reset feedback
    const feedback = document.getElementById('sim-feedback-badge');
    feedback.textContent = '';
    feedback.className = 'text-xs font-semibold';
    document.getElementById('sim-next-btn').classList.add('hidden');

    // Question Text
    document.getElementById('sim-question-text').textContent = q.question;

    // Options
    const optionsContainer = document.getElementById('sim-options-list');
    optionsContainer.innerHTML = '';

    const letters = ['A', 'B', 'C', 'D'];
    q.options.forEach((optText, optIdx) => {
      const letter = letters[optIdx] || String.fromCharCode(65 + optIdx);
      const btn = document.createElement('button');
      btn.className = 'quiz-opt-btn w-full text-left p-3.5 rounded-xl border border-slate-200 bg-white hover:bg-slate-50/80 transition-all flex items-start gap-3 text-xs font-medium text-slate-700 shadow-xs';
      btn.innerHTML = `
        <span class="opt-letter w-6 h-6 rounded-lg bg-slate-100 text-slate-600 font-bold flex items-center justify-center shrink-0 text-xs">
          ${letter}
        </span>
        <span class="flex-1 leading-relaxed">${this.escapeHtml(optText)}</span>
      `;

      btn.addEventListener('click', () => {
        this.selectAnswer(optText, q.correctAnswer, btn);
      });

      optionsContainer.appendChild(btn);
    });
  },

  selectAnswer(selected, correct, selectedBtn) {
    const allBtns = document.querySelectorAll('.quiz-opt-btn');
    allBtns.forEach(b => {
      b.disabled = true;
      b.classList.remove('hover:bg-slate-50/80');
    });

    const isCorrect = selected.trim() === correct.trim();
    this.playChime(isCorrect);

    if (isCorrect) {
      this.state.score++;
      selectedBtn.classList.remove('bg-white', 'border-slate-200');
      selectedBtn.classList.add('bg-emerald-50', 'border-emerald-300', 'text-emerald-900', 'ring-1', 'ring-emerald-400');
      selectedBtn.querySelector('.opt-letter').className = 'opt-letter w-6 h-6 rounded-lg bg-emerald-600 text-white font-bold flex items-center justify-center shrink-0 text-xs';

      const feedback = document.getElementById('sim-feedback-badge');
      feedback.textContent = '✨ ¡Correcto! Excelente respuesta.';
      feedback.className = 'text-xs font-bold text-emerald-600 flex items-center gap-1.5';
    } else {
      selectedBtn.classList.remove('bg-white', 'border-slate-200');
      selectedBtn.classList.add('bg-rose-50', 'border-rose-300', 'text-rose-900', 'ring-1', 'ring-rose-400');
      selectedBtn.querySelector('.opt-letter').className = 'opt-letter w-6 h-6 rounded-lg bg-rose-600 text-white font-bold flex items-center justify-center shrink-0 text-xs';

      // Highlight the correct one
      allBtns.forEach(b => {
        if (b.innerText.includes(correct)) {
          b.classList.add('bg-emerald-50', 'border-emerald-300', 'text-emerald-900');
        }
      });

      const feedback = document.getElementById('sim-feedback-badge');
      feedback.textContent = '❌ Incorrecto. La opción correcta ha sido destacada en verde.';
      feedback.className = 'text-xs font-bold text-rose-600 flex items-center gap-1.5';
    }

    const currentQ = this.state.activeQuiz.questions[this.state.currentQuestionIndex];
    this.state.userAnswers.push({
      question: currentQ.question,
      selected,
      correct,
      isCorrect
    });

    // Show Next Button
    const nextBtn = document.getElementById('sim-next-btn');
    nextBtn.classList.remove('hidden');
    if (this.state.currentQuestionIndex === this.state.activeQuiz.questions.length - 1) {
      nextBtn.querySelector('span').textContent = 'Finalizar Examen';
    } else {
      nextBtn.querySelector('span').textContent = 'Siguiente Pregunta';
    }
  },

  nextQuestion() {
    this.state.currentQuestionIndex++;
    if (this.state.currentQuestionIndex < this.state.activeQuiz.questions.length) {
      this.renderSimulatorQuestion();
    } else {
      this.finishSimulator();
    }
  },

  finishSimulator() {
    document.getElementById('quiz-simulator-card').classList.add('hidden');
    const resultCard = document.getElementById('quiz-result-card');
    resultCard.classList.remove('hidden');

    const total = this.state.activeQuiz.questions.length;
    const finalScore = Math.round((this.state.score / total) * 100);
    const passed = finalScore >= 60;

    const iconContainer = document.getElementById('result-icon-container');
    const title = document.getElementById('result-title');
    const desc = document.getElementById('result-desc');
    const scoreVal = document.getElementById('result-score');
    const badge = document.getElementById('result-badge');

    scoreVal.textContent = `${finalScore} / 100`;

    if (passed) {
      iconContainer.className = 'w-16 h-16 rounded-2xl mx-auto flex items-center justify-center bg-emerald-100 text-emerald-600';
      iconContainer.innerHTML = '<i data-lucide="check-circle-2" class="w-8 h-8"></i>';
      title.textContent = '¡Felicitaciones! Has Aprobado la Evaluación';
      desc.textContent = `Acertaste ${this.state.score} de ${total} preguntas correctamente. Demuestras un sólido entendimiento del tema.`;
      badge.textContent = 'APROBADO (≥ 60%)';
      badge.className = 'text-xs font-bold text-emerald-600 bg-emerald-50 px-3 py-1 rounded-full border border-emerald-200';
    } else {
      iconContainer.className = 'w-16 h-16 rounded-2xl mx-auto flex items-center justify-center bg-amber-100 text-amber-600';
      iconContainer.innerHTML = '<i data-lucide="alert-circle" class="w-8 h-8"></i>';
      title.textContent = 'No se alcanzó el puntaje mínimo de aprobación';
      desc.textContent = `Acertaste ${this.state.score} de ${total} preguntas. Te recomendamos repasar los apuntes e intentarlo nuevamente.`;
      badge.textContent = 'NO APROBADO (< 60%)';
      badge.className = 'text-xs font-bold text-amber-600 bg-amber-50 px-3 py-1 rounded-full border border-amber-200';
    }

    // Render Answer Review List
    const reviewList = document.getElementById('result-answers-list');
    reviewList.innerHTML = this.state.userAnswers.map((ans, i) => `
      <div class="p-3 rounded-xl border text-xs ${ans.isCorrect ? 'bg-emerald-50/50 border-emerald-200' : 'bg-rose-50/50 border-rose-200'}">
        <p class="font-bold text-slate-800">${i + 1}. ${this.escapeHtml(ans.question)}</p>
        <div class="mt-1 space-y-0.5 text-[11px]">
          <p class="${ans.isCorrect ? 'text-emerald-700 font-semibold' : 'text-rose-700'}">
            Tu respuesta: ${this.escapeHtml(ans.selected)} ${ans.isCorrect ? '✓' : '✗'}
          </p>
          ${!ans.isCorrect ? `<p class="text-emerald-700 font-semibold">Respuesta correcta: ${this.escapeHtml(ans.correct)}</p>` : ''}
        </div>
      </div>
    `).join('');

    if (window.lucide) window.lucide.createIcons();
    resultCard.scrollIntoView({ behavior: 'smooth' });
  },

  exitSimulator() {
    this.state.activeQuiz = null;
    document.getElementById('quiz-simulator-card').classList.add('hidden');
    document.getElementById('quiz-result-card').classList.add('hidden');
    document.getElementById('quiz-create-card').classList.remove('hidden');
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
  }
};
