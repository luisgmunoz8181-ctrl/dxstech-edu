import { Toast } from '../../components/toast.js';
import { Loading } from '../../components/loading.js';
import { esc, safeUrl } from '../../utils/escape.js';

// Aula virtual: lección activa, progreso, matrícula y certificado.

export const classroomMethods = {
  renderClassroomView() {
    const c = this.selectedCourse;
    const l = this.selectedLesson;
    const enr = this.courseEnrollment;
    const user = window.router?.currentUser;
    const isAdmin = user && ['SUPERADMIN', 'ADMINISTRADOR'].includes(user.role);
    const progressPercent = enr ? Math.round(enr.progressPercent) : 0;
    const isCompleted = enr && enr.status === 'completed';

    return `
      <div class="space-y-6 max-w-7xl mx-auto">
        <!-- Classroom Navigation Header -->
        <div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div class="flex items-center gap-3">
            <button id="back-to-catalog-btn" class="p-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 transition-colors cursor-pointer" title="Volver al catálogo">
              <i data-lucide="arrow-left" class="w-5 h-5"></i>
            </button>
            <div>
              <div class="flex items-center gap-2">
                <span class="px-2 py-0.5 rounded-md text-[10px] font-bold bg-indigo-50 text-indigo-700 border border-indigo-200">
                  ${esc(c.category)}
                </span>
                <span class="text-xs font-mono font-bold text-slate-400">${esc(c.code)}</span>
                ${isCompleted ? `
                  <span class="px-2 py-0.5 rounded-md text-[10px] font-bold bg-emerald-100 text-emerald-800 border border-emerald-300 flex items-center gap-1">
                    <i data-lucide="award" class="w-3 h-3 text-emerald-600"></i> Completado
                  </span>
                ` : ''}
              </div>
              <h3 class="text-lg font-black text-slate-900 leading-tight mt-0.5">${esc(c.title)}</h3>
            </div>
          </div>

          <div class="flex items-center gap-3 w-full sm:w-auto justify-between sm:justify-end">
            <!-- Progress Bar / Enroll Button / Certificate Button -->
            ${user ? (enr ? `
              <div class="flex items-center gap-2.5">
                <div class="w-28 sm:w-40 bg-slate-100 rounded-full h-2 overflow-hidden border border-slate-200">
                  <div class="bg-gradient-to-r from-indigo-500 to-emerald-500 h-2 rounded-full transition-all duration-500" style="width: ${progressPercent}%"></div>
                </div>
                <span class="text-xs font-bold text-slate-700 font-mono">${progressPercent}%</span>
              </div>
              ${(isCompleted || progressPercent >= 100) ? `
                <button id="view-cert-btn" class="px-3.5 py-2 rounded-xl text-xs font-bold text-white bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 transition-all shadow-md shadow-amber-200 flex items-center gap-1.5 cursor-pointer animate-pulse" title="Ver y descargar certificado">
                  <i data-lucide="award" class="w-4 h-4 text-white"></i>
                  <span>🎓 Certificado</span>
                </button>
              ` : ''}
            ` : `
              <button id="enroll-course-header-btn" class="px-4 py-2 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-700 transition-all shadow-xs flex items-center gap-1.5 cursor-pointer">
                <i data-lucide="user-plus" class="w-3.5 h-3.5"></i>
                <span>Inscribirme en este Curso</span>
              </button>
            `) : ''}

            <!-- Community & AI Tools -->
            <button id="open-tutor-modal-btn" class="px-3 py-2 rounded-xl text-xs font-bold text-white bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 transition-all shadow-xs flex items-center gap-1.5 cursor-pointer" title="Consultar al Tutor IA">
              <i data-lucide="bot" class="w-4 h-4"></i>
              <span class="hidden sm:inline">Tutor IA</span>
            </button>

            <button id="open-discussions-modal-btn" class="px-3 py-2 rounded-xl border border-slate-200 text-slate-700 hover:bg-slate-50 text-xs font-semibold flex items-center gap-1.5 transition-colors cursor-pointer" title="Foro de Dudas y Preguntas">
              <i data-lucide="message-square" class="w-4 h-4 text-indigo-600"></i>
              <span class="hidden md:inline">Foro</span>
            </button>

            <button id="open-reviews-modal-btn" class="px-3 py-2 rounded-xl border border-slate-200 text-slate-700 hover:bg-slate-50 text-xs font-semibold flex items-center gap-1.5 transition-colors cursor-pointer" title="Calificaciones y Reseñas">
              <i data-lucide="star" class="w-4 h-4 text-amber-500 fill-amber-400"></i>
              <span class="hidden md:inline">Opiniones</span>
            </button>

            ${isAdmin ? `
              <button id="view-students-btn" class="px-3.5 py-2 rounded-xl border border-slate-200 text-slate-700 hover:bg-slate-50 text-xs font-semibold flex items-center gap-1.5 transition-colors cursor-pointer" title="Ver estudiantes matriculados">
                <i data-lucide="users" class="w-3.5 h-3.5 text-indigo-600"></i>
                <span class="hidden md:inline">Alumnos</span>
              </button>
              <button id="add-module-btn" class="px-3.5 py-2 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-all shadow-xs flex items-center gap-1.5 cursor-pointer">
                <i data-lucide="plus" class="w-3.5 h-3.5"></i>
                <span>Nuevo Módulo</span>
              </button>
              <button id="edit-course-btn" class="p-2 rounded-xl border border-slate-200 text-slate-600 hover:bg-slate-50 transition-colors cursor-pointer" title="Editar detalles del curso">
                <i data-lucide="settings" class="w-4 h-4"></i>
              </button>
            ` : ''}
          </div>
        </div>

        <!-- Main Classroom Layout (Sidebar Modules + Content Viewer) -->
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          <!-- Syllabus Sidebar (Modules & Lessons list) -->
          <div class="lg:col-span-4 bg-white rounded-3xl border border-slate-200/80 shadow-xs p-5 space-y-4 max-h-[85vh] overflow-y-auto">
            <div class="flex items-center justify-between pb-3 border-b border-slate-100">
              <div class="flex items-center gap-2">
                <i data-lucide="book-marked" class="w-4 h-4 text-indigo-600"></i>
                <h4 class="text-xs font-bold uppercase tracking-wider text-slate-700">Temario del Programa</h4>
              </div>
              <span class="text-[11px] font-semibold text-slate-400">${c.modules?.length || 0} módulos</span>
            </div>

            <div class="space-y-4">
              ${(c.modules && c.modules.length > 0) ? c.modules.map((m, mIdx) => `
                <div class="rounded-2xl border border-slate-200/70 overflow-hidden bg-slate-50/50">
                  <div class="p-3 bg-slate-100/70 border-b border-slate-200/60 flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <span class="w-5 h-5 rounded-md bg-white text-indigo-700 text-[10px] font-bold flex items-center justify-center shadow-2xs border border-slate-200">
                        ${mIdx + 1}
                      </span>
                      <h5 class="text-xs font-bold text-slate-800 leading-tight">${esc(m.title)}</h5>
                    </div>

                    ${isAdmin ? `
                      <div class="flex items-center gap-1">
                        <button data-add-lesson-mod="${esc(m.id)}" class="p-1 rounded-md text-slate-400 hover:text-indigo-600 hover:bg-white transition-colors cursor-pointer" title="Añadir Lección">
                          <i data-lucide="plus" class="w-3.5 h-3.5"></i>
                        </button>
                        <button data-edit-mod="${esc(m.id)}" class="p-1 rounded-md text-slate-400 hover:text-indigo-600 hover:bg-white transition-colors cursor-pointer" title="Editar Módulo">
                          <i data-lucide="edit-3" class="w-3.5 h-3.5"></i>
                        </button>
                        <button data-del-mod="${esc(m.id)}" class="p-1 rounded-md text-slate-400 hover:text-rose-600 hover:bg-white transition-colors cursor-pointer" title="Eliminar Módulo">
                          <i data-lucide="trash-2" class="w-3.5 h-3.5"></i>
                        </button>
                      </div>
                    ` : ''}
                  </div>

                  <!-- Lessons in module -->
                  <div class="p-1.5 space-y-1">
                    ${(m.lessons && m.lessons.length > 0) ? m.lessons.map(lsn => {
                      const isCurrent = l && l.id === lsn.id;
                      const isCompleted = this.completedLessons.includes(lsn.id);
                      const iconName = {
                        youtube: 'video',
                        mp4: 'film',
                        pdf: 'file-text',
                        pptx: 'presentation',
                        text: 'align-left',
                        quiz: 'help-circle'
                      }[lsn.contentType] || 'file';

                      return `
                        <div class="group flex items-center justify-between p-2 rounded-xl text-xs transition-all cursor-pointer ${isCurrent ? 'bg-indigo-600 text-white font-bold shadow-xs' : 'text-slate-700 hover:bg-slate-200/60 font-medium'}">
                          <div data-select-lesson="${esc(lsn.id)}" class="flex-1 flex items-center gap-2.5 truncate">
                            ${isCompleted ? `
                              <i data-lucide="check-circle-2" class="w-4 h-4 shrink-0 ${isCurrent ? 'text-emerald-300' : 'text-emerald-500'}"></i>
                            ` : `
                              <i data-lucide="${iconName}" class="w-4 h-4 shrink-0 ${isCurrent ? 'text-indigo-200' : 'text-slate-400'}"></i>
                            `}
                            <span class="truncate">${esc(lsn.title)}</span>
                          </div>

                          <div class="flex items-center gap-1.5 shrink-0 ml-2">
                            ${lsn.isFreePreview ? `
                              <span class="text-[9px] px-1.5 py-0.5 rounded-md font-bold uppercase ${isCurrent ? 'bg-indigo-700 text-indigo-100' : 'bg-emerald-50 text-emerald-700 border border-emerald-200'}">
                                Gratis
                              </span>
                            ` : ''}
                            <span class="text-[10px] ${isCurrent ? 'text-indigo-200' : 'text-slate-400'}">${lsn.durationMinutes}m</span>

                            ${isAdmin ? `
                              <button data-edit-lesson="${esc(lsn.id)}" class="p-1 rounded text-slate-400 hover:text-indigo-600 hover:bg-white transition-colors hidden group-hover:block cursor-pointer" title="Editar lección">
                                <i data-lucide="edit-2" class="w-3 h-3"></i>
                              </button>
                              <button data-del-lesson="${esc(lsn.id)}" class="p-1 rounded text-slate-400 hover:text-rose-600 hover:bg-white transition-colors hidden group-hover:block cursor-pointer" title="Eliminar lección">
                                <i data-lucide="trash" class="w-3 h-3"></i>
                              </button>
                            ` : ''}
                          </div>
                        </div>
                      `;
                    }).join('') : `
                      <p class="text-[11px] text-slate-400 text-center py-3 italic">Sin lecciones añadidas aún</p>
                    `}
                  </div>
                </div>
              `).join('') : `
                <div class="text-center py-10 text-slate-400">
                  <i data-lucide="folder-plus" class="w-8 h-8 text-slate-300 mx-auto mb-2"></i>
                  <p class="text-xs">No hay módulos creados para este curso.</p>
                </div>
              `}
            </div>
          </div>

          <!-- Content Viewer Panel -->
          <div class="lg:col-span-8 bg-white rounded-3xl border border-slate-200/80 shadow-xs p-6 space-y-6">
            ${l ? this.renderLessonContent(l) : `
              <div class="py-20 text-center space-y-4">
                <div class="w-16 h-16 rounded-3xl bg-indigo-50 text-indigo-600 flex items-center justify-center mx-auto shadow-inner">
                  <i data-lucide="play-circle" class="w-8 h-8"></i>
                </div>
                <div>
                  <h4 class="text-base font-bold text-slate-800">Bienvenido al Aula de Aprendizaje</h4>
                  <p class="text-xs text-slate-400 max-w-sm mx-auto mt-1">
                    Selecciona una lección del temario lateral para reproducir videos, estudiar documentos PDF o leer los contenidos teóricos.
                  </p>
                </div>
              </div>
            `}
          </div>
        </div>
      </div>
    `;
  },

  renderLessonContent(l) {
    const isCompleted = this.completedLessons.includes(l.id);
    const nextLesson = this.getNextLesson(l.id);
    const user = window.router?.currentUser;

    let viewerHTML = '';

    if (l.contentType === 'youtube') {
      const videoId = this.extractYouTubeId(l.contentUrl);
      if (videoId) {
        viewerHTML = `
          <div class="aspect-video w-full rounded-2xl overflow-hidden shadow-lg border border-slate-200 bg-black">
            <iframe 
              src="https://www.youtube.com/embed/${esc(videoId)}?rel=0" 
              class="w-full h-full" 
              title="${esc(l.title)}"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" 
              allowfullscreen>
            </iframe>
          </div>
        `;
      } else {
        viewerHTML = `<div class="p-6 text-xs text-amber-800 bg-amber-50 rounded-2xl border border-amber-200">Enlace de YouTube no reconocido: ${safeUrl(l.contentUrl)}</div>`;
      }
    } else if (l.contentType === 'mp4') {
      viewerHTML = `
        <div class="aspect-video w-full rounded-2xl overflow-hidden shadow-lg border border-slate-200 bg-black">
          <video src="${safeUrl(l.contentUrl)}" controls class="w-full h-full" preload="metadata"></video>
        </div>
      `;
    } else if (l.contentType === 'pdf') {
      viewerHTML = `
        <div class="space-y-3">
          <div class="flex items-center justify-between bg-slate-50 p-3 rounded-2xl border border-slate-200 text-xs">
            <span class="font-semibold text-slate-700 flex items-center gap-1.5">
              <i data-lucide="file-text" class="w-4 h-4 text-rose-500"></i> Documento Oficial PDF
            </span>
            <div class="flex items-center gap-2">
              <a href="${safeUrl(l.contentUrl)}" target="_blank" class="px-3 py-1.5 rounded-xl bg-white border border-slate-200 font-bold text-slate-700 hover:bg-slate-100 transition-colors flex items-center gap-1">
                <i data-lucide="external-link" class="w-3.5 h-3.5"></i> Abrir en nueva ventana
              </a>
              <a href="${safeUrl(l.contentUrl)}" download class="px-3 py-1.5 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 transition-colors flex items-center gap-1 shadow-xs">
                <i data-lucide="download" class="w-3.5 h-3.5"></i> Descargar
              </a>
            </div>
          </div>
          <div class="w-full h-[650px] rounded-2xl border border-slate-200 overflow-hidden shadow-sm">
            <iframe src="${safeUrl(l.contentUrl)}" class="w-full h-full"></iframe>
          </div>
        </div>
      `;
    } else if (l.contentType === 'pptx') {
      const officeViewerURL = `https://view.officeapps.live.com/op/view.aspx?src=${encodeURIComponent(l.contentUrl)}`;
      viewerHTML = `
        <div class="space-y-3">
          <div class="flex items-center justify-between bg-slate-50 p-3 rounded-2xl border border-slate-200 text-xs">
            <span class="font-semibold text-slate-700 flex items-center gap-1.5">
              <i data-lucide="presentation" class="w-4 h-4 text-amber-500"></i> Presentación de Diapositivas PPTX
            </span>
            <a href="${safeUrl(l.contentUrl)}" download class="px-3 py-1.5 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 transition-colors flex items-center gap-1 shadow-xs">
              <i data-lucide="download" class="w-3.5 h-3.5"></i> Descargar Diapositivas
            </a>
          </div>
          <div class="w-full h-[600px] rounded-2xl border border-slate-200 overflow-hidden shadow-sm">
            <iframe src="${officeViewerURL}" class="w-full h-full"></iframe>
          </div>
        </div>
      `;
    } else if (l.contentType === 'quiz') {
      viewerHTML = `
        <div id="lesson-quiz-player" class="bg-slate-50/70 rounded-3xl border border-slate-200 p-6 sm:p-8 space-y-6">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-200/80 pb-4">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-2xl bg-indigo-100 text-indigo-700 flex items-center justify-center font-bold shrink-0">
                <i data-lucide="help-circle" class="w-6 h-6"></i>
              </div>
              <div>
                <span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-indigo-50 text-indigo-700 border border-indigo-200 uppercase tracking-wider">
                  Evaluación Interactiva IA
                </span>
                <h4 class="text-base font-bold text-slate-900 mt-0.5">Test de Conocimientos y Validación Pedagógica</h4>
              </div>
            </div>

            <div id="lesson-quiz-status-pill">
              ${isCompleted ? `
                <span class="px-3 py-1 rounded-full text-xs font-bold bg-emerald-100 text-emerald-800 border border-emerald-300 flex items-center gap-1.5 shadow-2xs">
                  <i data-lucide="check-circle" class="w-3.5 h-3.5 text-emerald-600"></i> Aprobado (>= 70%)
                </span>
              ` : `
                <span class="px-3 py-1 rounded-full text-xs font-bold bg-amber-100 text-amber-800 border border-amber-300 flex items-center gap-1.5 shadow-2xs">
                  <i data-lucide="award" class="w-3.5 h-3.5 text-amber-600"></i> Mínimo 70% para Aprobar
                </span>
              `}
            </div>
          </div>

          <div id="lesson-quiz-container" class="space-y-6">
            <div class="py-12 text-center text-slate-400">
              <div class="w-8 h-8 border-3 border-indigo-600/30 border-t-indigo-600 rounded-full animate-spin mx-auto mb-3"></div>
              <p class="text-xs font-medium">Cargando reactivos de evaluación con Gemini...</p>
            </div>
          </div>
        </div>
      `;
    } else {
      // Text / Guide content
      viewerHTML = `
        <div class="prose prose-slate max-w-none text-slate-700 text-sm leading-relaxed p-6 bg-slate-50/70 rounded-2xl border border-slate-200">
          ${l.contentBody ? esc(l.contentBody).replace(/\n/g, '<br>') : '<p class="italic text-slate-400">Sin contenido de lectura añadido.</p>'}
        </div>
      `;
    }

    return `
      <div class="space-y-6">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-100">
          <div>
            <div class="flex items-center gap-2 mb-1">
              <span class="px-2.5 py-0.5 rounded-md text-[10px] font-bold bg-indigo-50 text-indigo-700 uppercase tracking-wider">
                Lección Activa
              </span>
              <span class="text-xs text-slate-400 font-semibold">• ${l.durationMinutes} minutos estimados</span>
            </div>
            <h3 class="text-xl font-black text-slate-900">${esc(l.title)}</h3>
            ${l.description ? `<p class="text-xs text-slate-500 mt-1">${esc(l.description)}</p>` : ''}
          </div>

          <!-- Progress Action Buttons -->
          ${user ? `
            <div class="flex items-center gap-2 shrink-0">
              ${l.contentType === 'quiz' ? `
                <div class="px-3.5 py-2 rounded-xl text-xs font-bold ${isCompleted ? 'bg-emerald-50 text-emerald-700 border border-emerald-300' : 'bg-amber-50 text-amber-700 border border-amber-300'} flex items-center gap-1.5 shadow-2xs">
                  <i data-lucide="${isCompleted ? 'check-circle-2' : 'award'}" class="w-4 h-4"></i>
                  <span>${isCompleted ? 'Evaluación Aprobada ✓' : 'Aprobación Requerida (≥ 70%)'}</span>
                </div>
              ` : `
                <button id="toggle-lesson-progress-btn" class="px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 cursor-pointer shadow-xs ${isCompleted ? 'bg-emerald-50 text-emerald-700 border border-emerald-300 hover:bg-emerald-100' : 'bg-indigo-600 text-white hover:bg-indigo-700 shadow-indigo-100'}">
                  <i data-lucide="${isCompleted ? 'check-circle-2' : 'check'}" class="w-4 h-4"></i>
                  <span>${isCompleted ? 'Lección Completada ✓' : 'Marcar como Completada'}</span>
                </button>
              `}

              ${nextLesson ? `
                <button id="next-lesson-btn" class="px-3.5 py-2 rounded-xl text-xs font-bold bg-slate-100 text-slate-700 hover:bg-slate-200 transition-colors flex items-center gap-1 cursor-pointer">
                  <span>Siguiente</span>
                  <i data-lucide="chevron-right" class="w-4 h-4"></i>
                </button>
              ` : ''}
            </div>
          ` : ''}
        </div>

        ${viewerHTML}
      </div>
    `;
  },

  getNextLesson(currentLessonId) {
    if (!this.selectedCourse) return null;
    let allLessons = [];
    for (const mod of this.selectedCourse.modules || []) {
      for (const lsn of mod.lessons || []) {
        allLessons.push(lsn);
      }
    }
    const currentIndex = allLessons.findIndex(l => l.id === currentLessonId);
    if (currentIndex >= 0 && currentIndex < allLessons.length - 1) {
      return allLessons[currentIndex + 1];
    }
    return null;
  },

  async openCourse(courseId) {
    try {
      const res = await fetch(`/api/courses/${courseId}`);
      if (!res.ok) throw new Error('No se pudo acceder al curso');

      this.selectedCourse = await res.json();

      // If user is logged in, fetch enrollment status & progress
      const user = window.router?.currentUser;
      if (user) {
        try {
          const resProg = await fetch(`/api/enrollments/courses/${courseId}/progress`);
          if (resProg.ok) {
            const dataProg = await resProg.json();
            this.courseEnrollment = dataProg.enrollment;
            this.completedLessons = dataProg.completedLessons || [];
          }
        } catch {
          this.courseEnrollment = null;
          this.completedLessons = [];
        }
      }

      // Resume from last accessed lesson or select first
      if (this.courseEnrollment?.lastLessonId) {
        this.selectLessonById(this.courseEnrollment.lastLessonId, false);
      }
      if (!this.selectedLesson) {
        this.selectedLesson = this.selectedCourse.modules?.[0]?.lessons?.[0] || null;
      }

      window.router?.navigate('courses');
    } catch (e) {
      Toast.error(e.message);
    }
  },

  selectLessonById(lessonId, shouldNavigate = true) {
    if (!this.selectedCourse) return;
    for (const mod of this.selectedCourse.modules || []) {
      for (const lsn of mod.lessons || []) {
        if (lsn.id === lessonId) {
          this.selectedLesson = lsn;
          if (shouldNavigate) {
            window.router?.navigate('courses');
          }
          return;
        }
      }
    }
  },

  async enrollInCourse(courseId) {
    const user = window.router?.currentUser;
    if (!user) {
      Toast.info('Inicia sesión para matricularte en este curso.');
      window.router?.navigate('login');
      return;
    }

    try {
      const res = await fetch(`/api/enrollments/courses/${courseId}`, { method: 'POST' });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error);

      Toast.success('¡Te has matriculado con éxito en el curso!');
      await this.fetchEnrollmentsAndStats();
      await this.openCourse(courseId);
    } catch (e) {
      Toast.error(e.message);
    }
  },

  async toggleLessonProgress(lessonId, completed) {
    const user = window.router?.currentUser;
    if (!user) {
      Toast.info('Inicia sesión para registrar tu avance académico.');
      return;
    }

    try {
      const res = await fetch('/api/enrollments/progress', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          courseId: this.selectedCourse.id,
          lessonId,
          completed,
        }),
      });

      const updatedEnr = await res.json();
      if (!res.ok) throw new Error(updatedEnr.error);

      this.courseEnrollment = updatedEnr;
      if (completed) {
        if (!this.completedLessons.includes(lessonId)) {
          this.completedLessons.push(lessonId);
        }
        Toast.success('¡Lección completada!');
        if (updatedEnr.status === 'completed' || updatedEnr.progressPercent >= 100) {
          Toast.success('🎉 ¡Felicidades! Has completado el 100% del programa.');
          setTimeout(() => {
            this.openCertificateModal(this.selectedCourse.id);
          }, 600);
        }
      } else {
        this.completedLessons = this.completedLessons.filter(id => id !== lessonId);
        Toast.info('Progreso actualizado');
      }

      await this.fetchEnrollmentsAndStats();
      window.router?.navigate('courses');
    } catch (e) {
      Toast.error(e.message);
    }
  },

  async openCertificateModal(courseId) {
    Loading.show('Consultando certificado oficial...');
    try {
      const res = await fetch(`/api/certificates/course/${courseId}`);
      const cert = await res.json();
      Loading.hide();

      if (!res.ok) {
        Toast.error(cert.error || 'Aún no tienes certificado para este curso.');
        return;
      }

      const modalHtml = `
        <div id="cert-modal-backdrop" class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4">
          <div class="bg-white rounded-3xl max-w-xl w-full p-6 sm:p-8 border border-amber-200 shadow-2xl relative space-y-6 text-center animate-in fade-in zoom-in-95 duration-200">
            <button id="close-cert-modal" class="absolute top-5 right-5 text-slate-400 hover:text-slate-600 p-2 rounded-xl hover:bg-slate-100 transition-colors cursor-pointer">
              <i data-lucide="x" class="w-5 h-5"></i>
            </button>

            <div class="w-20 h-20 rounded-3xl bg-amber-50 text-amber-600 flex items-center justify-center mx-auto shadow-inner border border-amber-200">
              <i data-lucide="award" class="w-10 h-10"></i>
            </div>

            <div>
              <span class="inline-block px-3 py-1 rounded-full text-[11px] font-bold bg-amber-100 text-amber-800 border border-amber-300 mb-2">
                🏆 ¡Felicitaciones! Certificación Oficial Obtenida
              </span>
              <h3 class="text-xl font-black text-slate-900">${esc(cert.studentName)}</h3>
              <p class="text-xs text-slate-500 mt-1">ha completado y aprobado satisfactoriamente los requisitos académicos de:</p>
              <p class="text-sm font-bold text-indigo-700 mt-1">« ${esc(cert.courseTitle)} »</p>
            </div>

            <div class="bg-slate-50 rounded-2xl p-4 border border-slate-200/80 text-left space-y-2 text-xs">
              <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                <span class="text-slate-500 font-medium">Código de Registro Único:</span>
                <span class="font-mono font-bold text-slate-900 bg-white px-2 py-0.5 rounded border border-slate-200">${esc(cert.id)}</span>
              </div>
              <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                <span class="text-slate-500 font-medium">Intensidad Académica:</span>
                <span class="font-semibold text-slate-800">${cert.durationHours || 10} Horas</span>
              </div>
              <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                <span class="text-slate-500 font-medium">Fecha de Emisión:</span>
                <span class="font-semibold text-slate-800">${esc(cert.issueDate)}</span>
              </div>
              <div class="flex justify-between items-center py-1">
                <span class="text-slate-500 font-medium">Docente / Director:</span>
                <span class="font-semibold text-indigo-600">${esc(cert.instructorName || 'DxSTech Edu')}</span>
              </div>
            </div>

            <div class="pt-2 flex flex-wrap items-center justify-center gap-3">
              <a href="/api/certificates/${esc(cert.id)}/pdf" target="_blank" download="Certificado_${esc(cert.id)}.pdf" class="px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-all shadow-md shadow-indigo-100 flex items-center gap-2 cursor-pointer">
                <i data-lucide="download" class="w-4 h-4"></i>
                <span>Descargar Diploma Oficial (PDF)</span>
              </a>
              <a href="#verify/${esc(cert.id)}" id="go-verify-link" class="px-4 py-2.5 rounded-xl text-xs font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center gap-1.5 cursor-pointer">
                <i data-lucide="shield-check" class="w-4 h-4 text-emerald-600"></i>
                <span>Validar en Línea</span>
              </a>
            </div>
          </div>
        </div>
      `;

      document.body.insertAdjacentHTML('beforeend', modalHtml);
      if (window.lucide) window.lucide.createIcons();

      const modal = document.getElementById('cert-modal-backdrop');
      document.getElementById('close-cert-modal')?.addEventListener('click', () => modal?.remove());
      document.getElementById('go-verify-link')?.addEventListener('click', () => {
        modal?.remove();
        if (window.router) window.router.renderVerificationScreen(cert.id);
      });
      modal?.addEventListener('click', (e) => {
        if (e.target === modal) modal.remove();
      });
    } catch (err) {
      Loading.hide();
      Toast.error('Error cargando certificado: ' + err.message);
    }
  },

  extractYouTubeId(url) {
    if (!url) return null;
    const regExp = /^.*(youtu.be\/|v\/|u\/\w\/|embed\/|watch\?v=|\&v=)([^#\&\?]*).*/;
    const match = url.match(regExp);
    return (match && match[2].length === 11) ? match[2] : null;
  },
};
