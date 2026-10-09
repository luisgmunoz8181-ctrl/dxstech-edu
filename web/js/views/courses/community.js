import { Toast } from '../../components/toast.js';
import { Modal } from '../../components/modal.js';
import { Loading } from '../../components/loading.js';
import { esc } from '../../utils/escape.js';

// Tutor IA, foro de dudas y reseñas del curso.

export const communityMethods = {
  openTutorModal() {
    if (!this.selectedCourse) return;
    const course = this.selectedCourse;
    const lesson = this.selectedLesson;

    Modal.show({
      title: '🤖 Tutor Pedagógico IA',
      confirmText: 'Cerrar',
      showCancel: false,
      content: `
        <div class="space-y-4 text-left text-xs">
          <div class="p-3 bg-purple-50 rounded-2xl border border-purple-200 flex items-center justify-between">
            <div>
              <span class="text-[10px] font-bold text-purple-700 uppercase tracking-wider">Contexto Académico</span>
              <p class="font-bold text-slate-800 text-xs">${esc(course.title)}</p>
              ${lesson ? `<p class="text-[11px] text-purple-600">Lección: ${esc(lesson.title)}</p>` : ''}
            </div>
            <a href="https://wa.me/?text=${encodeURIComponent('Hola DxSTech Edu, tengo una consulta sobre el curso ' + course.title + (lesson ? ' - Lección: ' + lesson.title : '') + ': ')}" target="_blank" class="px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-[11px] flex items-center gap-1 transition-colors shrink-0 shadow-xs" title="Continuar en WhatsApp">
              <i data-lucide="phone" class="w-3.5 h-3.5"></i>
              <span class="hidden sm:inline">WhatsApp</span>
            </a>
          </div>

          <!-- Chat History -->
          <div id="tutor-chat-box" class="h-64 overflow-y-auto space-y-3 p-3 bg-slate-50 rounded-2xl border border-slate-200/80">
            <div class="flex items-start gap-2.5">
              <div class="w-7 h-7 rounded-lg bg-purple-600 text-white flex items-center justify-center shrink-0 font-bold text-xs">
                AI
              </div>
              <div class="bg-white p-3 rounded-2xl border border-slate-200 text-slate-700 leading-relaxed shadow-2xs">
                <p class="font-semibold text-purple-900 mb-0.5">¡Hola! Soy tu Tutor IA de DxSTech Edu.</p>
                <p>Estoy aquí para explicarte conceptos de este curso, darte ejemplos claros o resolver tus dudas antes de presentar las evaluaciones. ¿En qué puedo orientarte hoy?</p>
              </div>
            </div>
          </div>

          <!-- Suggestions -->
          <div class="flex flex-wrap gap-1.5 pt-1">
            <button type="button" class="tutor-prompt-chip text-[11px] px-2.5 py-1 rounded-lg bg-slate-100 hover:bg-purple-100 hover:text-purple-700 text-slate-600 transition-colors font-medium cursor-pointer">
              💡 Explícame con un ejemplo sencillo
            </button>
            <button type="button" class="tutor-prompt-chip text-[11px] px-2.5 py-1 rounded-lg bg-slate-100 hover:bg-purple-100 hover:text-purple-700 text-slate-600 transition-colors font-medium cursor-pointer">
              📌 ¿Cuáles son los puntos clave?
            </button>
            <button type="button" class="tutor-prompt-chip text-[11px] px-2.5 py-1 rounded-lg bg-slate-100 hover:bg-purple-100 hover:text-purple-700 text-slate-600 transition-colors font-medium cursor-pointer">
              🎯 Dame un ejercicio práctico
            </button>
          </div>

          <!-- Prompt input -->
          <div class="flex gap-2 pt-1">
            <input type="text" id="tutor-user-input" placeholder="Escribe tu consulta o duda académica..." class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-purple-500">
            <button id="tutor-send-btn" class="px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-700 text-white font-bold text-xs flex items-center gap-1.5 shadow-xs cursor-pointer transition-colors shrink-0">
              <i data-lucide="send" class="w-3.5 h-3.5"></i>
              <span>Preguntar</span>
            </button>
          </div>
        </div>
      `,
      onConfirm: () => {}
    });

    setTimeout(() => {
      if (window.lucide) window.lucide.createIcons();

      const input = document.getElementById('tutor-user-input');
      const sendBtn = document.getElementById('tutor-send-btn');
      const chatBox = document.getElementById('tutor-chat-box');

      const handleSend = async (questionText) => {
        const q = questionText || input.value.trim();
        if (!q) return;

        // Render user question
        chatBox.insertAdjacentHTML('beforeend', `
          <div class="flex items-start justify-end gap-2.5">
            <div class="bg-purple-600 text-white p-3 rounded-2xl leading-relaxed text-left max-w-[85%] shadow-2xs">
              <p>${esc(q).replace(/\n/g, '<br>')}</p>
            </div>
          </div>
        `);
        input.value = '';
        chatBox.scrollTop = chatBox.scrollHeight;

        // Thinking indicator
        const typingId = 'tutor-typing-' + Date.now();
        chatBox.insertAdjacentHTML('beforeend', `
          <div id="${esc(typingId)}" class="flex items-start gap-2.5">
            <div class="w-7 h-7 rounded-lg bg-purple-600 text-white flex items-center justify-center shrink-0 font-bold text-xs">
              AI
            </div>
            <div class="bg-white p-3 rounded-2xl border border-purple-200 text-purple-700 italic flex items-center gap-2">
              <div class="w-2 h-2 rounded-full bg-purple-600 animate-ping"></div>
              <span>El Tutor IA está analizando el material temático...</span>
            </div>
          </div>
        `);
        chatBox.scrollTop = chatBox.scrollHeight;

        try {
          const res = await fetch('/api/whatsapp/ask-tutor', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              courseId: course.id,
              lessonId: lesson?.id || '',
              question: q,
            }),
          });
          const data = await res.json();
          document.getElementById(typingId)?.remove();

          const answerText = data.answer || 'Disculpa, no pude procesar la respuesta en este momento.';
          chatBox.insertAdjacentHTML('beforeend', `
            <div class="flex items-start gap-2.5">
              <div class="w-7 h-7 rounded-lg bg-purple-600 text-white flex items-center justify-center shrink-0 font-bold text-xs">
                AI
              </div>
              <div class="bg-white p-3.5 rounded-2xl border border-slate-200 text-slate-800 leading-relaxed shadow-2xs max-w-[90%] space-y-1">
                ${esc(answerText).replace(/\n/g, '<br>')}
              </div>
            </div>
          `);
        } catch (err) {
          document.getElementById(typingId)?.remove();
          chatBox.insertAdjacentHTML('beforeend', `
            <div class="text-rose-500 text-xs py-1 text-center">${esc(err.message)}</div>
          `);
        }
        chatBox.scrollTop = chatBox.scrollHeight;
      };

      sendBtn?.addEventListener('click', () => handleSend());
      input?.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') handleSend();
      });

      document.querySelectorAll('.tutor-prompt-chip').forEach(chip => {
        chip.addEventListener('click', () => {
          handleSend(chip.textContent.replace(/^[^\wáéíóúÁÉÍÓÚ¿?]+/, '').trim());
        });
      });
    }, 50);
  },

  async openDiscussionsModal() {
    if (!this.selectedCourse) return;
    const course = this.selectedCourse;
    const lesson = this.selectedLesson;

    Loading.show('Cargando foro de dudas...');
    let discussions = [];
    try {
      const res = await fetch(`/api/courses/${course.id}/discussions`);
      if (res.ok) discussions = await res.json();
    } catch {
      // Ignorar error inicial
    }
    Loading.hide();

    Modal.show({
      title: `💬 Foro de Dudas: ${course.title}`,
      confirmText: 'Cerrar',
      showCancel: false,
      content: `
        <div class="space-y-4 text-left text-xs max-h-[75vh] flex flex-col">
          <!-- New question box -->
          <div class="p-3 bg-slate-50 rounded-2xl border border-slate-200/80 space-y-2.5 shrink-0">
            <p class="font-bold text-slate-800 flex items-center gap-1.5">
              <i data-lucide="plus-circle" class="w-3.5 h-3.5 text-indigo-600"></i> Publicar una Pregunta o Aporte
            </p>
            <input type="text" id="disc-new-title" placeholder="Título resumido de tu duda..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs bg-white">
            <textarea id="disc-new-message" rows="2" placeholder="Explica detalladamente tu inquietud para que tutores y compañeros puedan responderte..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs bg-white"></textarea>
            
            <div class="flex items-center justify-between pt-1">
              <label class="flex items-center gap-1.5 text-slate-600 cursor-pointer">
                <input type="checkbox" id="disc-link-lesson" ${lesson ? 'checked' : ''} class="w-3.5 h-3.5 rounded text-indigo-600">
                <span class="text-[11px]">${lesson ? `Asociar a "${esc(lesson.title)}"` : 'Consulta general del curso'}</span>
              </label>
              <button id="disc-post-btn" class="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs flex items-center gap-1 transition-colors cursor-pointer shadow-xs">
                <i data-lucide="send" class="w-3.5 h-3.5"></i>
                <span>Publicar</span>
              </button>
            </div>
          </div>

          <!-- Discussions feed -->
          <div id="disc-items-container" class="overflow-y-auto space-y-3 flex-1 pr-1">
            ${(!discussions || discussions.length === 0) ? `
              <div class="py-10 text-center text-slate-400">
                <i data-lucide="message-square-dashed" class="w-8 h-8 text-slate-300 mx-auto mb-2"></i>
                <p class="font-semibold text-slate-700">Aún no hay preguntas en este foro.</p>
                <p class="text-[11px] text-slate-400 mt-0.5">Sé el primero en iniciar una conversación pedagógica.</p>
              </div>
            ` : discussions.map(d => `
              <div class="p-3.5 bg-white rounded-2xl border border-slate-200 shadow-2xs space-y-2">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="w-7 h-7 rounded-xl bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center text-[11px]">
                      ${esc((d.userName || 'U').charAt(0))}
                    </div>
                    <div>
                      <span class="font-bold text-slate-800">${esc(d.userName)}</span>
                      <span class="text-[10px] text-indigo-600 bg-indigo-50 px-1.5 py-0.5 rounded font-semibold ml-1">${esc(d.userRole)}</span>
                    </div>
                  </div>
                  <span class="text-[10px] text-slate-400">${new Date(d.createdAt).toLocaleDateString()}</span>
                </div>

                <div>
                  <h5 class="font-bold text-slate-900 text-xs">${esc(d.title || 'Aporte')}</h5>
                  <p class="text-slate-600 text-xs mt-1 leading-relaxed">${esc(d.message).replace(/\n/g, '<br>')}</p>
                </div>

                <!-- Nested replies -->
                ${(d.replies && d.replies.length > 0) ? `
                  <div class="mt-2 pl-4 border-l-2 border-indigo-200 space-y-2 pt-1">
                    ${d.replies.map(r => `
                      <div class="p-2 bg-slate-50 rounded-xl border border-slate-200/60 text-xs">
                        <div class="flex items-center justify-between mb-1">
                          <span class="font-bold text-slate-800">${esc(r.userName)} <span class="text-[10px] text-indigo-600">(${esc(r.userRole)})</span></span>
                          <span class="text-[10px] text-slate-400">${new Date(r.createdAt).toLocaleDateString()}</span>
                        </div>
                        <p class="text-slate-600">${esc(r.message).replace(/\n/g, '<br>')}</p>
                      </div>
                    `).join('')}
                  </div>
                ` : ''}

                <!-- Reply trigger -->
                <div class="pt-1 flex justify-end">
                  <button data-reply-to="${esc(d.id)}" class="text-[11px] font-bold text-indigo-600 hover:text-indigo-800 transition-colors flex items-center gap-1 cursor-pointer">
                    <i data-lucide="corner-down-right" class="w-3 h-3"></i> Responder
                  </button>
                </div>

                <div id="reply-form-${esc(d.id)}" class="hidden pt-2 border-t border-slate-100 space-y-2">
                  <input type="text" id="reply-input-${esc(d.id)}" placeholder="Escribe tu respuesta..." class="w-full rounded-xl border border-slate-200 px-3 py-1.5 text-xs bg-slate-50">
                  <div class="flex justify-end gap-1.5">
                    <button data-send-reply="${esc(d.id)}" class="px-3 py-1 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg font-bold text-[11px] cursor-pointer">Enviar Respuesta</button>
                  </div>
                </div>
              </div>
            `).join('')}
          </div>
        </div>
      `,
      onConfirm: () => {}
    });

    setTimeout(() => {
      if (window.lucide) window.lucide.createIcons();

      // Bind post
      document.getElementById('disc-post-btn')?.addEventListener('click', async () => {
        const title = document.getElementById('disc-new-title').value.trim();
        const message = document.getElementById('disc-new-message').value.trim();
        const linkLesson = document.getElementById('disc-link-lesson').checked;

        if (!title || !message) {
          Toast.warning('Completa el título y mensaje de la duda');
          return;
        }

        try {
          const res = await fetch(`/api/courses/${course.id}/discussions`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              title,
              message,
              lessonId: linkLesson && lesson ? lesson.id : '',
            }),
          });
          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          Toast.success('¡Pregunta publicada en el foro!');
          Modal.close();
          this.openDiscussionsModal();
        } catch (err) {
          Toast.error(err.message);
        }
      });

      // Bind reply toggle
      document.querySelectorAll('[data-reply-to]').forEach(btn => {
        btn.addEventListener('click', () => {
          const id = btn.dataset.replyTo;
          document.getElementById(`reply-form-${esc(id)}`)?.classList.toggle('hidden');
        });
      });

      // Bind send reply
      document.querySelectorAll('[data-send-reply]').forEach(btn => {
        btn.addEventListener('click', async () => {
          const parentId = btn.dataset.sendReply;
          const input = document.getElementById(`reply-input-${esc(parentId)}`);
          const message = input?.value.trim();
          if (!message) return;

          try {
            const res = await fetch(`/api/courses/${course.id}/discussions`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                parentId,
                title: 'Respuesta',
                message,
              }),
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error);

            Toast.success('Respuesta publicada con éxito');
            Modal.close();
            this.openDiscussionsModal();
          } catch (err) {
            Toast.error(err.message);
          }
        });
      });
    }, 50);
  },

  async openReviewsModal() {
    if (!this.selectedCourse) return;
    const course = this.selectedCourse;
    const user = window.router?.currentUser;

    Loading.show('Consultando calificaciones...');
    let summary = { reviews: [], averageRating: 5.0, totalReviews: 0, ratingBreakdown: {} };
    try {
      const res = await fetch(`/api/courses/${course.id}/reviews`);
      if (res.ok) summary = await res.json();
    } catch {
      // Ignorar error inicial
    }
    Loading.hide();

    let selectedRating = 5;

    Modal.show({
      title: `⭐ Opiniones: ${course.title}`,
      confirmText: 'Cerrar',
      showCancel: false,
      content: `
        <div class="space-y-4 text-left text-xs max-h-[75vh] flex flex-col">
          <!-- Summary Header Ribbon -->
          <div class="p-4 bg-gradient-to-r from-amber-500/10 to-indigo-500/10 rounded-2xl border border-amber-200/80 flex items-center justify-between gap-4 shrink-0">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-2xl bg-amber-500 text-white font-black text-lg flex items-center justify-center shadow-md shadow-amber-200">
                ${(summary.averageRating || 5.0).toFixed(1)}
              </div>
              <div>
                <div class="flex text-amber-500 text-sm">
                  ★★★★★
                </div>
                <p class="text-xs font-bold text-slate-800 mt-0.5">${summary.totalReviews} opiniones de estudiantes</p>
              </div>
            </div>
            <span class="text-[11px] font-semibold text-slate-500 bg-white px-3 py-1 rounded-full border border-slate-200">
              DxSTech Verified
            </span>
          </div>

          <!-- Submit Review Box if user is logged in -->
          ${user ? `
            <div class="p-3.5 bg-slate-50 rounded-2xl border border-slate-200 space-y-2.5 shrink-0">
              <div class="flex items-center justify-between">
                <span class="font-bold text-slate-800">Califica este curso</span>
                <div id="star-selector" class="flex gap-1 text-slate-300 text-lg cursor-pointer">
                  <span data-star="1" class="text-amber-400 hover:scale-110 transition-transform">★</span>
                  <span data-star="2" class="text-amber-400 hover:scale-110 transition-transform">★</span>
                  <span data-star="3" class="text-amber-400 hover:scale-110 transition-transform">★</span>
                  <span data-star="4" class="text-amber-400 hover:scale-110 transition-transform">★</span>
                  <span data-star="5" class="text-amber-400 hover:scale-110 transition-transform">★</span>
                </div>
              </div>
              <textarea id="review-comment-input" rows="2" placeholder="Escribe tu opinión sobre el docente, el material y las evaluaciones..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs bg-white"></textarea>
              <div class="flex justify-end">
                <button id="submit-review-btn" class="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-700 text-white font-bold text-xs flex items-center gap-1 transition-colors cursor-pointer shadow-xs">
                  <i data-lucide="star" class="w-3.5 h-3.5"></i>
                  <span>Publicar Opinión</span>
                </button>
              </div>
            </div>
          ` : `
            <p class="text-slate-400 text-center py-2 italic text-[11px]">Inicia sesión para dejar una reseña en este curso.</p>
          `}

          <!-- Reviews List -->
          <div class="overflow-y-auto space-y-3 flex-1 pr-1">
            ${(!summary.reviews || summary.reviews.length === 0) ? `
              <p class="py-8 text-center text-slate-400 italic">No hay opiniones publicadas aún. ¡Sé el primero en calificar!</p>
            ` : summary.reviews.map(r => `
              <div class="p-3 bg-white rounded-2xl border border-slate-200 shadow-2xs space-y-1.5">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="w-6 h-6 rounded-lg bg-amber-100 text-amber-800 font-bold flex items-center justify-center text-[10px]">
                      ${esc((r.userName || 'U').charAt(0))}
                    </div>
                    <span class="font-bold text-slate-800">${esc(r.userName)}</span>
                  </div>
                  <div class="flex text-amber-500 text-xs">
                    ${'★'.repeat(r.rating)}${'☆'.repeat(5 - r.rating)}
                  </div>
                </div>
                ${r.comment ? `<p class="text-slate-600 text-xs leading-relaxed pl-8">${esc(r.comment).replace(/\n/g, '<br>')}</p>` : ''}
                <div class="text-[10px] text-slate-400 text-right">
                  ${new Date(r.createdAt).toLocaleDateString()}
                </div>
              </div>
            `).join('')}
          </div>
        </div>
      `,
      onConfirm: () => {}
    });

    setTimeout(() => {
      if (window.lucide) window.lucide.createIcons();

      // Star selection
      const starSpans = document.querySelectorAll('#star-selector span');
      starSpans.forEach(s => {
        s.addEventListener('click', () => {
          selectedRating = parseInt(s.dataset.star, 10);
          starSpans.forEach(st => {
            const val = parseInt(st.dataset.star, 10);
            if (val <= selectedRating) {
              st.className = 'text-amber-400 hover:scale-110 transition-transform';
            } else {
              st.className = 'text-slate-300 hover:scale-110 transition-transform';
            }
          });
        });
      });

      // Submit review
      document.getElementById('submit-review-btn')?.addEventListener('click', async () => {
        const comment = document.getElementById('review-comment-input')?.value.trim() || '';

        try {
          const res = await fetch(`/api/courses/${course.id}/reviews`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              rating: selectedRating,
              comment,
            }),
          });
          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          Toast.success('¡Gracias por tu opinión! Reseña registrada.');
          Modal.close();
          this.openReviewsModal();
        } catch (err) {
          Toast.error(err.message);
        }
      });
    }, 50);
  },
};
