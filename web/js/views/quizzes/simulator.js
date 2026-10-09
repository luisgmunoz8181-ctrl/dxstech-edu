import { Toast } from '../../components/toast.js';
import { Loading } from '../../components/loading.js';

// Simulador de examen: preguntas, respuestas, sonido y resultado.

export const simulatorMethods = {
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
};
