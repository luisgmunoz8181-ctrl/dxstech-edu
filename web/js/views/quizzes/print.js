import { Toast } from '../../components/toast.js';
import { Modal } from '../../components/modal.js';
import { Loading } from '../../components/loading.js';
import { esc } from '../../utils/escape.js';

// Hoja imprimible y confirmación de borrado.

export const printMethods = {
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
};
