import { Toast } from '../../components/toast.js';
import { Modal } from '../../components/modal.js';
import { Loading } from '../../components/loading.js';
import { esc } from '../../utils/escape.js';

// Evaluaciones por lección generadas con IA.

export const lessonQuizMethods = {
  async generateQuizForLesson(lessonId, apiKey) {
    Loading.show('Gemini está creando las preguntas pedagógicas para la lección...');
    try {
      const res = await fetch('/api/quizzes/generate-for-lesson', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Gemini-API-Key': apiKey || '',
        },
        body: JSON.stringify({ lessonId, count: 5 }),
      });
      const data = await res.json();
      Loading.hide();
      if (!res.ok) throw new Error(data.error);

      Toast.success('¡Evaluación con IA generada exitosamente!');
      Modal.close();
      await this.openCourse(this.selectedCourse.id);
      this.selectLessonById(lessonId);
    } catch (err) {
      Loading.hide();
      Toast.error('Error generando evaluación: ' + err.message);
    }
  },

  async loadLessonQuiz(lesson) {
    const container = document.getElementById('lesson-quiz-container');
    if (!container) return;

    try {
      const res = await fetch(`/api/quizzes/lesson/${lesson.id}`);
      const user = window.router?.currentUser;
      const isAdmin = user && ['SUPERADMIN', 'ADMINISTRADOR'].includes(user.role);

      if (!res.ok) {
        if (isAdmin) {
          container.innerHTML = `
            <div class="p-8 text-center bg-white rounded-2xl border border-dashed border-purple-300 space-y-4">
              <div class="w-12 h-12 rounded-2xl bg-purple-50 text-purple-600 flex items-center justify-center mx-auto">
                <i data-lucide="sparkles" class="w-6 h-6"></i>
              </div>
              <div>
                <h4 class="text-sm font-bold text-slate-800">Esta lección aún no tiene preguntas generadas</h4>
                <p class="text-xs text-slate-500 max-w-md mx-auto mt-1">
                  Puedes formular automáticamente preguntas pedagógicas de opción múltiple con Google Gemini basadas en los temas de esta lección.
                </p>
              </div>
              <button id="btn-lesson-autogen-now" class="px-5 py-2.5 rounded-xl bg-purple-600 hover:bg-purple-700 text-white font-bold text-xs flex items-center gap-2 mx-auto cursor-pointer shadow-md shadow-purple-100 transition-all">
                <i data-lucide="bot" class="w-4 h-4"></i>
                <span>Autogenerar Evaluación con IA Ahora</span>
              </button>
            </div>
          `;
          if (window.lucide) window.lucide.createIcons();
          document.getElementById('btn-lesson-autogen-now')?.addEventListener('click', () => {
            let apiKey = localStorage.getItem('dxstech_gemini_api_key');
            if (!apiKey) {
              Modal.promptGeminiKey({
                onSaved: (key) => this.generateQuizForLesson(lesson.id, key)
              });
              return;
            }
            this.generateQuizForLesson(lesson.id, apiKey);
          });
        } else {
          container.innerHTML = `
            <div class="p-8 text-center bg-white rounded-2xl border border-slate-200">
              <i data-lucide="clock" class="w-8 h-8 text-slate-300 mx-auto mb-2"></i>
              <h4 class="text-sm font-bold text-slate-700">Evaluación en Preparación</h4>
              <p class="text-xs text-slate-400 mt-1">El docente o tutor está formulando los reactivos de esta lección.</p>
            </div>
          `;
          if (window.lucide) window.lucide.createIcons();
        }
        return;
      }

      const quiz = await res.json();
      this.activeLessonQuiz = quiz;

      if (!quiz.questions || quiz.questions.length === 0) {
        container.innerHTML = `<p class="text-xs text-slate-400 text-center py-6">No hay preguntas registradas en este examen.</p>`;
        return;
      }

      const isPassed = quiz.userPassed || this.completedLessons.includes(lesson.id);

      let html = `
        <div class="space-y-5">
          ${isPassed ? `
            <div class="p-4 bg-emerald-50 border border-emerald-200 rounded-2xl text-xs text-emerald-800 flex items-center gap-3">
              <div class="w-8 h-8 rounded-xl bg-emerald-100 text-emerald-700 flex items-center justify-center font-bold shrink-0">
                <i data-lucide="award" class="w-4 h-4"></i>
              </div>
              <div>
                <p class="font-bold">¡Has completado y aprobado satisfactoriamente esta evaluación!</p>
                <p class="text-[11px] text-emerald-700 mt-0.5">Calificación obtenida: ${quiz.latestScore ? Math.round(quiz.latestScore) : 100}%. Puedes presentarla nuevamente si deseas repasar.</p>
              </div>
            </div>
          ` : ''}

          <div class="space-y-4">
            ${quiz.questions.map((q, idx) => `
              <div class="p-4 sm:p-5 bg-white rounded-2xl border border-slate-200 shadow-2xs space-y-3 quiz-q-card transition-all" data-q-idx="${idx}">
                <div class="flex items-start gap-2.5">
                  <span class="w-6 h-6 rounded-lg bg-indigo-50 text-indigo-700 font-bold text-xs flex items-center justify-center shrink-0 border border-indigo-100">
                    ${idx + 1}
                  </span>
                  <p class="font-bold text-slate-800 text-xs sm:text-sm leading-relaxed">${esc(q.question)}</p>
                </div>

                <div class="space-y-2 pt-1 pl-8">
                  ${q.options.map((opt, optIdx) => `
                    <label class="flex items-center gap-3 p-2.5 rounded-xl border border-slate-200/80 hover:bg-indigo-50/50 hover:border-indigo-200 cursor-pointer transition-colors text-xs text-slate-700 font-medium opt-row">
                      <input type="radio" name="lesson-quiz-q-${idx}" value="${opt.replace(/"/g, '&quot;')}" class="w-4 h-4 text-indigo-600 focus:ring-indigo-500">
                      <span>${esc(opt)}</span>
                    </label>
                  `).join('')}
                </div>

                <div id="quiz-feedback-${idx}" class="hidden pt-2 pl-8 text-xs font-semibold"></div>
              </div>
            `).join('')}
          </div>

          <div class="pt-4 flex flex-col sm:flex-row items-center justify-between gap-3 border-t border-slate-200/80">
            <p id="quiz-submit-summary" class="text-xs text-slate-500 font-medium">
              Porcentaje mínimo requerido para aprobar y registrar avance: <strong>70%</strong>
            </p>
            <button id="submit-lesson-quiz-btn" class="w-full sm:w-auto px-6 py-3 rounded-2xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-all shadow-md shadow-indigo-100 flex items-center justify-center gap-2 cursor-pointer">
              <i data-lucide="check-circle" class="w-4 h-4"></i>
              <span>Enviar y Calificar Evaluación</span>
            </button>
          </div>
        </div>
      `;

      container.innerHTML = html;
      if (window.lucide) window.lucide.createIcons();

      // Bind submit
      document.getElementById('submit-lesson-quiz-btn')?.addEventListener('click', () => {
        this.submitLessonQuiz(quiz, lesson);
      });
    } catch (err) {
      container.innerHTML = `<div class="p-6 text-center text-xs text-rose-600 bg-rose-50 rounded-2xl">${esc(err.message)}</div>`;
    }
  },

  async submitLessonQuiz(quiz, lesson) {
    const answers = {};
    let missing = false;

    quiz.questions.forEach((q, idx) => {
      const selected = document.querySelector(`input[name="lesson-quiz-q-${idx}"]:checked`);
      if (!selected) {
        missing = true;
      } else {
        answers[idx.toString()] = selected.value;
      }
    });

    if (missing) {
      Toast.warning('Por favor responde todas las preguntas antes de enviar la evaluación.');
      return;
    }

    Loading.show('Calificando evaluación interactiva con IA...');
    try {
      const res = await fetch(`/api/quizzes/${quiz.id}/submit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          courseId: this.selectedCourse.id,
          lessonId: lesson.id,
          answers,
        }),
      });
      const data = await res.json();
      Loading.hide();

      if (!res.ok) throw new Error(data.error);

      // Render results feedback on cards
      if (data.feedback) {
        data.feedback.forEach(item => {
          const card = document.querySelector(`.quiz-q-card[data-q-idx="${item.index}"]`);
          const feedbackDiv = document.getElementById(`quiz-feedback-${item.index}`);
          if (!card || !feedbackDiv) return;

          feedbackDiv.classList.remove('hidden');
          if (item.isCorrect) {
            card.classList.remove('border-slate-200', 'border-rose-300');
            card.classList.add('border-emerald-300', 'bg-emerald-50/20');
            feedbackDiv.innerHTML = `<span class="text-emerald-700 flex items-center gap-1.5"><i data-lucide="check" class="w-3.5 h-3.5"></i> ¡Respuesta Correcta!</span>`;
          } else {
            card.classList.remove('border-slate-200', 'border-emerald-300');
            card.classList.add('border-rose-300', 'bg-rose-50/20');
            feedbackDiv.innerHTML = `<span class="text-rose-700 flex items-center gap-1.5"><i data-lucide="x" class="w-3.5 h-3.5"></i> Incorrecto. Respuesta correcta: <strong>${esc(item.correctAnswer)}</strong></span>`;
          }
        });
      }

      if (window.lucide) window.lucide.createIcons();

      if (data.passed) {
        Toast.success(`¡Felicitaciones! Aprobaste con ${Math.round(data.score)}% (${data.correct}/${data.total})`);
        if (!this.completedLessons.includes(lesson.id)) {
          this.completedLessons.push(lesson.id);
        }

        // Update pill
        const pill = document.getElementById('lesson-quiz-status-pill');
        if (pill) {
          pill.innerHTML = `
            <span class="px-3 py-1 rounded-full text-xs font-bold bg-emerald-100 text-emerald-800 border border-emerald-300 flex items-center gap-1.5 shadow-2xs">
              <i data-lucide="check-circle" class="w-3.5 h-3.5 text-emerald-600"></i> Aprobado (${Math.round(data.score)}%)
            </span>
          `;
        }

        // Refresh enrollment and course status
        await this.fetchEnrollmentsAndStats();
        if (this.selectedCourse) {
          await this.openCourse(this.selectedCourse.id);
          this.selectLessonById(lesson.id);
        }
      } else {
        Toast.warning(`Calificación: ${Math.round(data.score)}%. Se requiere mínimo 70% para aprobar. Puedes revisar y reintentar.`);
      }
    } catch (err) {
      Loading.hide();
      Toast.error('Error al calificar: ' + err.message);
    }
  },
};
