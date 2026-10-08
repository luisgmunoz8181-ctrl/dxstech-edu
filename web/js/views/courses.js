import { Toast } from '../components/toast.js';
import { Modal } from '../components/modal.js';
import { Loading } from '../components/loading.js';

export const CoursesView = {
  courses: [],
  myEnrollments: [],
  studentStats: null,
  activeTab: 'all', // 'all' | 'my-courses'
  selectedCourse: null,
  selectedLesson: null,
  courseEnrollment: null,
  completedLessons: [],
  search: '',
  categoryFilter: 'all',
  statusFilter: 'all',

  render() {
    if (this.selectedCourse) {
      return this.renderClassroomView();
    }
    return this.renderCatalogView();
  },

  renderCatalogView() {
    const user = window.router?.currentUser;
    const isAdmin = user && ['SUPERADMIN', 'ADMINISTRADOR'].includes(user.role);
    const isStudent = user && user.role === 'ESTUDIANTE';
    const stats = this.studentStats;

    return `
      <div class="space-y-6 max-w-7xl mx-auto">
        <!-- Hero Header -->
        <div class="bg-gradient-to-r from-indigo-900 via-indigo-950 to-slate-900 rounded-3xl p-6 sm:p-8 text-white shadow-xl flex flex-col md:flex-row items-start md:items-center justify-between gap-6 relative overflow-hidden">
          <div class="absolute -right-10 -bottom-10 w-60 h-60 bg-indigo-500/10 rounded-full blur-3xl pointer-events-none"></div>
          <div>
            <div class="flex items-center gap-2 mb-2">
              <span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-indigo-500/30 text-indigo-200 border border-indigo-400/30 uppercase tracking-wider">
                Academia Digital & LMS
              </span>
              <span class="text-xs text-indigo-300">• Aula Virtual DxSTech</span>
            </div>
            <h3 class="text-2xl sm:text-3xl font-black tracking-tight">Catálogo de Cursos & Aulas Virtuales</h3>
            <p class="text-xs sm:text-sm text-slate-300 mt-2 max-w-2xl leading-relaxed">
              Explora programas de capacitación por módulos y lecciones. Accede a documentos PDF, presentaciones PPTX, videos interactivos y registra tu avance académico.
            </p>
          </div>

          ${isAdmin ? `
            <button id="create-course-btn" class="px-5 py-3 rounded-2xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-500 transition-all shadow-lg shadow-indigo-600/30 flex items-center gap-2.5 shrink-0 cursor-pointer">
              <i data-lucide="plus-circle" class="w-4 h-4"></i>
              <span>Nuevo Curso</span>
            </button>
          ` : ''}
        </div>

        <!-- Student Stats Ribbon (If authenticated and has stats) -->
        ${(user && stats) ? `
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
            <div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-2xs flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold">
                <i data-lucide="book-marked" class="w-5 h-5"></i>
              </div>
              <div>
                <p class="text-[11px] font-semibold text-slate-400 leading-none">Matriculados</p>
                <p class="text-lg font-black text-slate-900 mt-1">${stats.totalEnrolled}</p>
              </div>
            </div>

            <div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-2xs flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold">
                <i data-lucide="clock" class="w-5 h-5"></i>
              </div>
              <div>
                <p class="text-[11px] font-semibold text-slate-400 leading-none">En Progreso</p>
                <p class="text-lg font-black text-slate-900 mt-1">${stats.inProgress}</p>
              </div>
            </div>

            <div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-2xs flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold">
                <i data-lucide="award" class="w-5 h-5"></i>
              </div>
              <div>
                <p class="text-[11px] font-semibold text-slate-400 leading-none">Completados</p>
                <p class="text-lg font-black text-slate-900 mt-1">${stats.completed}</p>
              </div>
            </div>

            <div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-2xs flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-purple-50 text-purple-600 flex items-center justify-center font-bold">
                <i data-lucide="trending-up" class="w-5 h-5"></i>
              </div>
              <div>
                <p class="text-[11px] font-semibold text-slate-400 leading-none">Avance Promedio</p>
                <p class="text-lg font-black text-slate-900 mt-1">${Math.round(stats.averageProgress)}%</p>
              </div>
            </div>
          </div>
        ` : ''}

        <!-- Tabs & Filters Toolbar -->
        <div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-xs flex flex-col md:flex-row items-center justify-between gap-4">
          <!-- View Tabs (if user is authenticated) -->
          <div class="flex items-center gap-1.5 p-1 bg-slate-100 rounded-xl w-full md:w-auto">
            <button id="tab-all-courses" class="flex-1 md:flex-none px-4 py-2 rounded-lg text-xs font-bold transition-all cursor-pointer ${this.activeTab === 'all' ? 'bg-white text-indigo-700 shadow-2xs' : 'text-slate-600 hover:text-slate-900'}">
              Todos los Cursos (${this.courses.length})
            </button>
            ${user ? `
              <button id="tab-my-courses" class="flex-1 md:flex-none px-4 py-2 rounded-lg text-xs font-bold transition-all cursor-pointer ${this.activeTab === 'my-courses' ? 'bg-white text-indigo-700 shadow-2xs' : 'text-slate-600 hover:text-slate-900'}">
                Mis Cursos (${this.myEnrollments.length})
              </button>
            ` : ''}
          </div>

          <div class="flex flex-wrap items-center gap-3 w-full md:w-auto">
            <!-- Search bar -->
            <div class="relative w-full sm:w-64">
              <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                <i data-lucide="search" class="w-4 h-4"></i>
              </span>
              <input type="text" id="course-search-input" value="${this.search}" placeholder="Buscar curso o código..." class="w-full text-xs rounded-xl border border-slate-200 pl-9 pr-3.5 py-2 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
            </div>

            <!-- Category filter -->
            <select id="course-category-filter" class="text-xs rounded-xl border border-slate-200 px-3 py-2 bg-slate-50 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
              <option value="all" ${this.categoryFilter === 'all' ? 'selected' : ''}>Todas las categorías</option>
              <option value="Inteligencia Artificial" ${this.categoryFilter === 'Inteligencia Artificial' ? 'selected' : ''}>Inteligencia Artificial</option>
              <option value="Desarrollo Web" ${this.categoryFilter === 'Desarrollo Web' ? 'selected' : ''}>Desarrollo Web</option>
              <option value="Ciberseguridad" ${this.categoryFilter === 'Ciberseguridad' ? 'selected' : ''}>Ciberseguridad</option>
              <option value="Tecnología" ${this.categoryFilter === 'Tecnología' ? 'selected' : ''}>Tecnología</option>
            </select>

            ${isAdmin ? `
              <select id="course-status-filter" class="text-xs rounded-xl border border-slate-200 px-3 py-2 bg-slate-50 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
                <option value="all" ${this.statusFilter === 'all' ? 'selected' : ''}>Todos los estados</option>
                <option value="published" ${this.statusFilter === 'published' ? 'selected' : ''}>Publicados</option>
                <option value="draft" ${this.statusFilter === 'draft' ? 'selected' : ''}>Borradores</option>
                <option value="archived" ${this.statusFilter === 'archived' ? 'selected' : ''}>Archivados</option>
              </select>
            ` : ''}
          </div>
        </div>

        <!-- Courses Cards Grid -->
        <div id="courses-grid" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div class="col-span-full py-16 text-center text-slate-400">
            <div class="w-8 h-8 border-3 border-indigo-600/30 border-t-indigo-600 rounded-full animate-spin mx-auto mb-3"></div>
            <p class="text-xs font-medium">Cargando catálogo educativo de DxSTech Edu...</p>
          </div>
        </div>
      </div>
    `;
  },

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
                  ${c.category}
                </span>
                <span class="text-xs font-mono font-bold text-slate-400">${c.code}</span>
                ${isCompleted ? `
                  <span class="px-2 py-0.5 rounded-md text-[10px] font-bold bg-emerald-100 text-emerald-800 border border-emerald-300 flex items-center gap-1">
                    <i data-lucide="award" class="w-3 h-3 text-emerald-600"></i> Completado
                  </span>
                ` : ''}
              </div>
              <h3 class="text-lg font-black text-slate-900 leading-tight mt-0.5">${c.title}</h3>
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
                      <h5 class="text-xs font-bold text-slate-800 leading-tight">${m.title}</h5>
                    </div>

                    ${isAdmin ? `
                      <div class="flex items-center gap-1">
                        <button data-add-lesson-mod="${m.id}" class="p-1 rounded-md text-slate-400 hover:text-indigo-600 hover:bg-white transition-colors cursor-pointer" title="Añadir Lección">
                          <i data-lucide="plus" class="w-3.5 h-3.5"></i>
                        </button>
                        <button data-edit-mod="${m.id}" class="p-1 rounded-md text-slate-400 hover:text-indigo-600 hover:bg-white transition-colors cursor-pointer" title="Editar Módulo">
                          <i data-lucide="edit-3" class="w-3.5 h-3.5"></i>
                        </button>
                        <button data-del-mod="${m.id}" class="p-1 rounded-md text-slate-400 hover:text-rose-600 hover:bg-white transition-colors cursor-pointer" title="Eliminar Módulo">
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
                        text: 'align-left'
                      }[lsn.contentType] || 'file';

                      return `
                        <div class="group flex items-center justify-between p-2 rounded-xl text-xs transition-all cursor-pointer ${isCurrent ? 'bg-indigo-600 text-white font-bold shadow-xs' : 'text-slate-700 hover:bg-slate-200/60 font-medium'}">
                          <div data-select-lesson="${lsn.id}" class="flex-1 flex items-center gap-2.5 truncate">
                            ${isCompleted ? `
                              <i data-lucide="check-circle-2" class="w-4 h-4 shrink-0 ${isCurrent ? 'text-emerald-300' : 'text-emerald-500'}"></i>
                            ` : `
                              <i data-lucide="${iconName}" class="w-4 h-4 shrink-0 ${isCurrent ? 'text-indigo-200' : 'text-slate-400'}"></i>
                            `}
                            <span class="truncate">${lsn.title}</span>
                          </div>

                          <div class="flex items-center gap-1.5 shrink-0 ml-2">
                            ${lsn.isFreePreview ? `
                              <span class="text-[9px] px-1.5 py-0.5 rounded-md font-bold uppercase ${isCurrent ? 'bg-indigo-700 text-indigo-100' : 'bg-emerald-50 text-emerald-700 border border-emerald-200'}">
                                Gratis
                              </span>
                            ` : ''}
                            <span class="text-[10px] ${isCurrent ? 'text-indigo-200' : 'text-slate-400'}">${lsn.durationMinutes}m</span>

                            ${isAdmin ? `
                              <button data-edit-lesson="${lsn.id}" class="p-1 rounded text-slate-400 hover:text-indigo-600 hover:bg-white transition-colors hidden group-hover:block cursor-pointer" title="Editar lección">
                                <i data-lucide="edit-2" class="w-3 h-3"></i>
                              </button>
                              <button data-del-lesson="${lsn.id}" class="p-1 rounded text-slate-400 hover:text-rose-600 hover:bg-white transition-colors hidden group-hover:block cursor-pointer" title="Eliminar lección">
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
      const videoId = this.extractYouTubeId(l.contentURL);
      if (videoId) {
        viewerHTML = `
          <div class="aspect-video w-full rounded-2xl overflow-hidden shadow-lg border border-slate-200 bg-black">
            <iframe 
              src="https://www.youtube.com/embed/${videoId}?rel=0" 
              class="w-full h-full" 
              title="${l.title}"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" 
              allowfullscreen>
            </iframe>
          </div>
        `;
      } else {
        viewerHTML = `<div class="p-6 text-xs text-amber-800 bg-amber-50 rounded-2xl border border-amber-200">Enlace de YouTube no reconocido: ${l.contentURL}</div>`;
      }
    } else if (l.contentType === 'mp4') {
      viewerHTML = `
        <div class="aspect-video w-full rounded-2xl overflow-hidden shadow-lg border border-slate-200 bg-black">
          <video src="${l.contentURL}" controls class="w-full h-full" preload="metadata"></video>
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
              <a href="${l.contentURL}" target="_blank" class="px-3 py-1.5 rounded-xl bg-white border border-slate-200 font-bold text-slate-700 hover:bg-slate-100 transition-colors flex items-center gap-1">
                <i data-lucide="external-link" class="w-3.5 h-3.5"></i> Abrir en nueva ventana
              </a>
              <a href="${l.contentURL}" download class="px-3 py-1.5 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 transition-colors flex items-center gap-1 shadow-xs">
                <i data-lucide="download" class="w-3.5 h-3.5"></i> Descargar
              </a>
            </div>
          </div>
          <div class="w-full h-[650px] rounded-2xl border border-slate-200 overflow-hidden shadow-sm">
            <iframe src="${l.contentURL}" class="w-full h-full"></iframe>
          </div>
        </div>
      `;
    } else if (l.contentType === 'pptx') {
      const officeViewerURL = `https://view.officeapps.live.com/op/view.aspx?src=${encodeURIComponent(l.contentURL)}`;
      viewerHTML = `
        <div class="space-y-3">
          <div class="flex items-center justify-between bg-slate-50 p-3 rounded-2xl border border-slate-200 text-xs">
            <span class="font-semibold text-slate-700 flex items-center gap-1.5">
              <i data-lucide="presentation" class="w-4 h-4 text-amber-500"></i> Presentación de Diapositivas PPTX
            </span>
            <a href="${l.contentURL}" download class="px-3 py-1.5 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 transition-colors flex items-center gap-1 shadow-xs">
              <i data-lucide="download" class="w-3.5 h-3.5"></i> Descargar Diapositivas
            </a>
          </div>
          <div class="w-full h-[600px] rounded-2xl border border-slate-200 overflow-hidden shadow-sm">
            <iframe src="${officeViewerURL}" class="w-full h-full"></iframe>
          </div>
        </div>
      `;
    } else {
      // Text / Guide content
      viewerHTML = `
        <div class="prose prose-slate max-w-none text-slate-700 text-sm leading-relaxed p-6 bg-slate-50/70 rounded-2xl border border-slate-200">
          ${l.contentBody ? l.contentBody.replace(/\n/g, '<br>') : '<p class="italic text-slate-400">Sin contenido de lectura añadido.</p>'}
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
            <h3 class="text-xl font-black text-slate-900">${l.title}</h3>
            ${l.description ? `<p class="text-xs text-slate-500 mt-1">${l.description}</p>` : ''}
          </div>

          <!-- Progress Action Buttons -->
          ${user ? `
            <div class="flex items-center gap-2 shrink-0">
              <button id="toggle-lesson-progress-btn" class="px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 cursor-pointer shadow-xs ${isCompleted ? 'bg-emerald-50 text-emerald-700 border border-emerald-300 hover:bg-emerald-100' : 'bg-indigo-600 text-white hover:bg-indigo-700 shadow-indigo-100'}">
                <i data-lucide="${isCompleted ? 'check-circle-2' : 'check'}" class="w-4 h-4"></i>
                <span>${isCompleted ? 'Lección Completada ✓' : 'Marcar como Completada'}</span>
              </button>

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

  async mount() {
    this.bindEvents();
    await this.fetchEnrollmentsAndStats();
    if (!this.selectedCourse) {
      await this.fetchCourses();
    }
    if (window.lucide) window.lucide.createIcons();
  },

  async fetchEnrollmentsAndStats() {
    const user = window.router?.currentUser;
    if (!user) {
      this.myEnrollments = [];
      this.studentStats = null;
      return;
    }

    try {
      const [resEnr, resStats] = await Promise.all([
        fetch('/api/enrollments/my-courses'),
        fetch('/api/enrollments/stats'),
      ]);

      if (resEnr.ok) {
        this.myEnrollments = await resEnr.json();
      }
      if (resStats.ok) {
        this.studentStats = await resStats.json();
      }
    } catch {
      // Ignorar errores no críticos de red
    }
  },

  bindEvents() {
    // Search input
    const searchInput = document.getElementById('course-search-input');
    if (searchInput) {
      searchInput.addEventListener('input', (e) => {
        this.search = e.target.value;
        this.renderCoursesGrid();
      });
    }

    // Category filter
    const catSelect = document.getElementById('course-category-filter');
    if (catSelect) {
      catSelect.addEventListener('change', (e) => {
        this.categoryFilter = e.target.value;
        this.fetchCourses();
      });
    }

    // Status filter
    const statusSelect = document.getElementById('course-status-filter');
    if (statusSelect) {
      statusSelect.addEventListener('change', (e) => {
        this.statusFilter = e.target.value;
        this.fetchCourses();
      });
    }

    // Tabs
    const tabAll = document.getElementById('tab-all-courses');
    if (tabAll) {
      tabAll.addEventListener('click', () => {
        this.activeTab = 'all';
        window.router?.navigate('courses');
      });
    }

    const tabMy = document.getElementById('tab-my-courses');
    if (tabMy) {
      tabMy.addEventListener('click', () => {
        this.activeTab = 'my-courses';
        window.router?.navigate('courses');
      });
    }

    // Create course button
    const createBtn = document.getElementById('create-course-btn');
    if (createBtn) {
      createBtn.addEventListener('click', () => this.showCourseModal());
    }

    // Back to catalog button
    const backBtn = document.getElementById('back-to-catalog-btn');
    if (backBtn) {
      backBtn.addEventListener('click', () => {
        this.selectedCourse = null;
        this.selectedLesson = null;
        this.courseEnrollment = null;
        this.completedLessons = [];
        window.router?.navigate('courses');
      });
    }

    // View Certificate button from classroom header
    const viewCertBtn = document.getElementById('view-cert-btn');
    if (viewCertBtn && this.selectedCourse) {
      viewCertBtn.addEventListener('click', () => {
        this.openCertificateModal(this.selectedCourse.id);
      });
    }

    // Enroll from classroom header
    const enrollHeaderBtn = document.getElementById('enroll-course-header-btn');
    if (enrollHeaderBtn) {
      enrollHeaderBtn.addEventListener('click', () => {
        this.enrollInCourse(this.selectedCourse.id);
      });
    }

    // Add module button
    const addModBtn = document.getElementById('add-module-btn');
    if (addModBtn) {
      addModBtn.addEventListener('click', () => this.showModuleModal(this.selectedCourse.id));
    }

    // View enrolled students button (Admin)
    const viewStudentsBtn = document.getElementById('view-students-btn');
    if (viewStudentsBtn) {
      viewStudentsBtn.addEventListener('click', () => this.showEnrolledStudentsModal(this.selectedCourse.id));
    }

    // Edit course details button
    const editCourseBtn = document.getElementById('edit-course-btn');
    if (editCourseBtn) {
      editCourseBtn.addEventListener('click', () => this.showCourseModal(this.selectedCourse));
    }

    // Toggle Lesson Progress Button
    const toggleProgBtn = document.getElementById('toggle-lesson-progress-btn');
    if (toggleProgBtn && this.selectedLesson) {
      toggleProgBtn.addEventListener('click', () => {
        const isCompleted = this.completedLessons.includes(this.selectedLesson.id);
        this.toggleLessonProgress(this.selectedLesson.id, !isCompleted);
      });
    }

    // Next Lesson Button
    const nextLessonBtn = document.getElementById('next-lesson-btn');
    if (nextLessonBtn && this.selectedLesson) {
      nextLessonBtn.addEventListener('click', () => {
        const next = this.getNextLesson(this.selectedLesson.id);
        if (next) {
          this.selectedLesson = next;
          window.router?.navigate('courses');
        }
      });
    }

    // Classroom sidebar delegated clicks
    document.querySelectorAll('[data-select-lesson]').forEach(el => {
      el.addEventListener('click', () => {
        const lessonId = el.dataset.selectLesson;
        this.selectLessonById(lessonId);
      });
    });

    document.querySelectorAll('[data-add-lesson-mod]').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.showLessonModal(btn.dataset.addLessonMod);
      });
    });

    document.querySelectorAll('[data-edit-mod]').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const modId = btn.dataset.editMod;
        const mod = this.selectedCourse.modules.find(m => m.id === modId);
        if (mod) this.showModuleModal(this.selectedCourse.id, mod);
      });
    });

    document.querySelectorAll('[data-del-mod]').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.deleteModule(btn.dataset.delMod);
      });
    });

    document.querySelectorAll('[data-edit-lesson]').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const lessonId = btn.dataset.editLesson;
        this.findAndEditLesson(lessonId);
      });
    });

    document.querySelectorAll('[data-del-lesson]').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.deleteLesson(btn.dataset.delLesson);
      });
    });
  },

  async fetchCourses() {
    try {
      const params = new URLSearchParams();
      if (this.search) params.append('search', this.search);
      if (this.categoryFilter !== 'all') params.append('category', this.categoryFilter);
      if (this.statusFilter !== 'all') params.append('status', this.statusFilter);

      const res = await fetch(`/api/courses?${params.toString()}`);
      if (!res.ok) throw new Error('Error cargando catálogo');

      this.courses = await res.json();
      this.renderCoursesGrid();
    } catch (e) {
      Toast.error(e.message);
    }
  },

  renderCoursesGrid() {
    const grid = document.getElementById('courses-grid');
    if (!grid) return;

    let coursesToDisplay = this.courses;

    // Filter by activeTab
    if (this.activeTab === 'my-courses') {
      const enrolledCourseIDs = this.myEnrollments.map(e => e.courseId);
      coursesToDisplay = this.courses.filter(c => enrolledCourseIDs.includes(c.id));
    }

    // Filter by search locally if needed
    if (this.search.trim()) {
      const q = this.search.toLowerCase().trim();
      coursesToDisplay = coursesToDisplay.filter(c => 
        c.title.toLowerCase().includes(q) ||
        c.code.toLowerCase().includes(q) ||
        (c.description && c.description.toLowerCase().includes(q))
      );
    }

    if (!coursesToDisplay || coursesToDisplay.length === 0) {
      grid.innerHTML = `
        <div class="col-span-full py-16 text-center text-slate-400 bg-white rounded-3xl border border-slate-200">
          <i data-lucide="book-x" class="w-10 h-10 mx-auto mb-2 text-slate-300"></i>
          <h4 class="text-sm font-bold text-slate-700">No se encontraron cursos</h4>
          <p class="text-xs text-slate-400 mt-1">
            ${this.activeTab === 'my-courses' ? 'Aún no te has matriculado en ningún curso.' : 'Ajusta los filtros de búsqueda o categoría.'}
          </p>
        </div>
      `;
      if (window.lucide) window.lucide.createIcons();
      return;
    }

    const user = window.router?.currentUser;
    const isAdmin = user && ['SUPERADMIN', 'ADMINISTRADOR'].includes(user.role);

    grid.innerHTML = coursesToDisplay.map(c => {
      const enr = this.myEnrollments.find(e => e.courseId === c.id);
      const isEnrolled = !!enr;
      const progressPercent = enr ? Math.round(enr.progressPercent) : 0;
      const isCompleted = enr && enr.status === 'completed';

      const statusBadge = {
        published: '<span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">Publicado</span>',
        draft: '<span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200">Borrador</span>',
        archived: '<span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-100 text-slate-600">Archivado</span>',
      }[c.status] || '';

      const fallbackThumb = 'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=800&q=80';
      const thumb = c.thumbnailUrl || fallbackThumb;

      return `
        <div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs hover:shadow-lg transition-all overflow-hidden flex flex-col group">
          <!-- Thumbnail Image -->
          <div class="relative h-44 bg-slate-100 overflow-hidden">
            <img src="${thumb}" alt="${c.title}" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" onerror="this.src='${fallbackThumb}'">
            <div class="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent"></div>
            
            <div class="absolute top-3 left-3 flex flex-wrap gap-1.5">
              <span class="px-2.5 py-1 rounded-full text-[10px] font-bold bg-white/90 backdrop-blur-md text-slate-800 shadow-2xs">
                ${c.category}
              </span>
              ${isAdmin ? statusBadge : ''}
              ${isEnrolled ? `
                <span class="px-2 py-0.5 rounded-full text-[10px] font-bold ${isCompleted ? 'bg-emerald-500 text-white' : 'bg-indigo-600 text-white'} shadow-xs">
                  ${isCompleted ? '✓ Completado' : `${progressPercent}% avance`}
                </span>
              ` : ''}
            </div>

            <div class="absolute bottom-3 left-3 right-3 flex items-center justify-between text-white text-[11px] font-semibold">
              <span class="flex items-center gap-1 drop-shadow-sm">
                <i data-lucide="clock" class="w-3.5 h-3.5"></i> ${c.durationHours}h
              </span>
              <span class="flex items-center gap-1 drop-shadow-sm font-mono text-[10px] bg-black/40 px-2 py-0.5 rounded-full">
                ${c.code}
              </span>
            </div>
          </div>

          <!-- Body Info -->
          <div class="p-5 flex-1 flex flex-col justify-between space-y-4">
            <div>
              <div class="flex items-center gap-2 text-slate-400 text-[11px] mb-1">
                <i data-lucide="user" class="w-3.5 h-3.5"></i>
                <span>${c.instructorName || 'Docente Asignado'}</span>
                <span>•</span>
                <span class="font-medium text-indigo-600">${c.level}</span>
              </div>

              <h4 class="text-base font-bold text-slate-900 leading-snug group-hover:text-indigo-600 transition-colors">
                ${c.title}
              </h4>
              <p class="text-xs text-slate-500 mt-2 line-clamp-2 leading-relaxed">
                ${c.shortDescription || c.description || 'Sin descripción disponible.'}
              </p>
            </div>

            <!-- Enrollment Progress Bar if enrolled -->
            ${isEnrolled ? `
              <div class="space-y-1">
                <div class="flex justify-between text-[11px] font-semibold">
                  <span class="text-slate-400">Progreso del curso:</span>
                  <span class="${isCompleted ? 'text-emerald-600 font-bold' : 'text-indigo-600 font-bold'}">${progressPercent}%</span>
                </div>
                <div class="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                  <div class="${isCompleted ? 'bg-emerald-500' : 'bg-indigo-600'} h-1.5 rounded-full transition-all" style="width: ${progressPercent}%"></div>
                </div>
              </div>
            ` : ''}

            <!-- Footer Stats & Actions -->
            <div class="pt-3 border-t border-slate-100 flex items-center justify-between">
              <div class="flex items-center gap-3 text-xs text-slate-400 font-semibold">
                <span title="Módulos" class="flex items-center gap-1"><i data-lucide="layers" class="w-3.5 h-3.5"></i> ${c.modulesCount}</span>
                <span title="Lecciones" class="flex items-center gap-1"><i data-lucide="file-video" class="w-3.5 h-3.5"></i> ${c.lessonsCount}</span>
              </div>

              <div class="flex items-center gap-1.5">
                ${isEnrolled && isCompleted ? `
                  <button data-course-cert="${c.id}" class="p-2 rounded-xl text-amber-600 bg-amber-50 hover:bg-amber-100 border border-amber-200 transition-colors flex items-center justify-center shadow-2xs cursor-pointer" title="Descargar Certificado Oficial">
                    <i data-lucide="award" class="w-4 h-4"></i>
                  </button>
                ` : ''}
                ${isEnrolled ? `
                  <button data-open-course="${c.id}" class="px-3.5 py-1.5 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-colors flex items-center gap-1 shadow-xs cursor-pointer">
                    <span>Continuar</span>
                    <i data-lucide="play" class="w-3 h-3"></i>
                  </button>
                ` : `
                  <button data-open-course="${c.id}" class="px-3 py-1.5 rounded-xl text-xs font-bold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center gap-1 cursor-pointer">
                    <span>${isAdmin ? 'Gestionar' : 'Explorar'}</span>
                    <i data-lucide="arrow-right" class="w-3 h-3"></i>
                  </button>
                `}

                ${isAdmin ? `
                  <button data-course-menu="${c.id}" class="p-1.5 rounded-xl text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors cursor-pointer" title="Opciones Administrativas">
                    <i data-lucide="more-vertical" class="w-4 h-4"></i>
                  </button>
                ` : ''}
              </div>
            </div>
          </div>
        </div>
      `;
    }).join('');

    // Bind card buttons
    grid.querySelectorAll('[data-course-cert]').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.openCertificateModal(btn.dataset.courseCert);
      });
    });

    grid.querySelectorAll('[data-open-course]').forEach(btn => {
      btn.addEventListener('click', () => {
        this.openCourse(btn.dataset.openCourse);
      });
    });

    grid.querySelectorAll('[data-course-menu]').forEach(btn => {
      btn.addEventListener('click', () => {
        this.showCourseActionMenu(btn.dataset.courseMenu);
      });
    });

    if (window.lucide) window.lucide.createIcons();
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
              <h3 class="text-xl font-black text-slate-900">${cert.studentName}</h3>
              <p class="text-xs text-slate-500 mt-1">ha completado y aprobado satisfactoriamente los requisitos académicos de:</p>
              <p class="text-sm font-bold text-indigo-700 mt-1">« ${cert.courseTitle} »</p>
            </div>

            <div class="bg-slate-50 rounded-2xl p-4 border border-slate-200/80 text-left space-y-2 text-xs">
              <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                <span class="text-slate-500 font-medium">Código de Registro Único:</span>
                <span class="font-mono font-bold text-slate-900 bg-white px-2 py-0.5 rounded border border-slate-200">${cert.id}</span>
              </div>
              <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                <span class="text-slate-500 font-medium">Intensidad Académica:</span>
                <span class="font-semibold text-slate-800">${cert.durationHours || 10} Horas</span>
              </div>
              <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                <span class="text-slate-500 font-medium">Fecha de Emisión:</span>
                <span class="font-semibold text-slate-800">${cert.issueDate}</span>
              </div>
              <div class="flex justify-between items-center py-1">
                <span class="text-slate-500 font-medium">Docente / Director:</span>
                <span class="font-semibold text-indigo-600">${cert.instructorName || 'DxSTech Edu'}</span>
              </div>
            </div>

            <div class="pt-2 flex flex-wrap items-center justify-center gap-3">
              <a href="/api/certificates/${cert.id}/pdf" target="_blank" download="Certificado_${cert.id}.pdf" class="px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-all shadow-md shadow-indigo-100 flex items-center gap-2 cursor-pointer">
                <i data-lucide="download" class="w-4 h-4"></i>
                <span>Descargar Diploma Oficial (PDF)</span>
              </a>
              <a href="#verify/${cert.id}" id="go-verify-link" class="px-4 py-2.5 rounded-xl text-xs font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center gap-1.5 cursor-pointer">
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

  showEnrolledStudentsModal(courseId) {
    Modal.show({
      title: 'Estudiantes Matriculados',
      confirmText: 'Cerrar',
      showCancel: false,
      content: `
        <div class="space-y-4 text-left text-xs">
          <div class="flex items-center justify-between pb-2 border-b border-slate-100">
            <span class="font-bold text-slate-800">Alumnos inscritos en el programa</span>
            <span id="students-modal-count" class="text-slate-400 font-semibold">Cargando...</span>
          </div>

          <div id="students-modal-list" class="max-h-60 overflow-y-auto space-y-2">
            <div class="py-6 text-center text-slate-400">Cargando directorio de estudiantes...</div>
          </div>
        </div>
      `,
      onConfirm: () => {}
    });

    // Fetch students list
    setTimeout(async () => {
      try {
        const res = await fetch(`/api/enrollments/admin/course/${courseId}/students`);
        const list = await res.json();
        const container = document.getElementById('students-modal-list');
        const countSpan = document.getElementById('students-modal-count');

        if (!container) return;

        if (countSpan) countSpan.textContent = `${list.length} alumnos`;

        if (!list || list.length === 0) {
          container.innerHTML = `<p class="py-6 text-center text-slate-400 italic">No hay estudiantes matriculados aún.</p>`;
          return;
        }

        container.innerHTML = list.map(e => `
          <div class="p-3 bg-slate-50 rounded-2xl border border-slate-200/80 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-xl bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center text-xs">
                ${(e.studentName || 'U').charAt(0)}
              </div>
              <div>
                <p class="font-bold text-slate-800">${e.studentName}</p>
                <p class="text-[11px] text-slate-400">${e.studentEmail}</p>
              </div>
            </div>

            <div class="text-right">
              <span class="inline-block px-2 py-0.5 rounded-full text-[10px] font-bold ${e.status === 'completed' ? 'bg-emerald-100 text-emerald-800' : 'bg-indigo-100 text-indigo-800'}">
                ${Math.round(e.progressPercent)}% avance
              </span>
              <p class="text-[10px] text-slate-400 mt-0.5">Inscrito: ${new Date(e.enrolledAt).toLocaleDateString()}</p>
            </div>
          </div>
        `).join('');
      } catch (err) {
        const container = document.getElementById('students-modal-list');
        if (container) container.innerHTML = `<p class="text-rose-500 py-4 text-center">${err.message}</p>`;
      }
    }, 50);
  },

  showCourseActionMenu(courseId) {
    const course = this.courses.find(c => c.id === courseId);
    if (!course) return;

    Modal.show({
      title: `Opciones de Curso: ${course.code}`,
      confirmText: 'Cerrar',
      showCancel: false,
      content: `
        <div class="space-y-2 text-left text-xs">
          <p class="font-bold text-slate-800 text-sm mb-3">${course.title}</p>
          
          <button id="modal-action-students" class="w-full text-left p-2.5 rounded-xl hover:bg-slate-100 flex items-center gap-2.5 text-slate-700 font-semibold transition-colors cursor-pointer">
            <i data-lucide="users" class="w-4 h-4 text-indigo-600"></i> Ver Estudiantes Matriculados
          </button>

          <button id="modal-action-edit" class="w-full text-left p-2.5 rounded-xl hover:bg-slate-100 flex items-center gap-2.5 text-slate-700 font-semibold transition-colors cursor-pointer">
            <i data-lucide="edit" class="w-4 h-4 text-blue-600"></i> Editar Información Básica
          </button>

          <button id="modal-action-duplicate" class="w-full text-left p-2.5 rounded-xl hover:bg-slate-100 flex items-center gap-2.5 text-slate-700 font-semibold transition-colors cursor-pointer">
            <i data-lucide="copy" class="w-4 h-4 text-purple-600"></i> Duplicar Curso Completo
          </button>

          ${course.status === 'published' ? `
            <button id="modal-action-unpublish" class="w-full text-left p-2.5 rounded-xl hover:bg-slate-100 flex items-center gap-2.5 text-amber-700 font-semibold transition-colors cursor-pointer">
              <i data-lucide="eye-off" class="w-4 h-4 text-amber-600"></i> Pasar a Borrador (Despublicar)
            </button>
          ` : `
            <button id="modal-action-publish" class="w-full text-left p-2.5 rounded-xl hover:bg-slate-100 flex items-center gap-2.5 text-emerald-700 font-semibold transition-colors cursor-pointer">
              <i data-lucide="check-circle" class="w-4 h-4 text-emerald-600"></i> Publicar en Catálogo
            </button>
          `}

          <button id="modal-action-delete" class="w-full text-left p-2.5 rounded-xl hover:bg-rose-50 flex items-center gap-2.5 text-rose-700 font-semibold transition-colors cursor-pointer">
            <i data-lucide="trash-2" class="w-4 h-4 text-rose-600"></i> Eliminar Curso
          </button>
        </div>
      `,
      onConfirm: () => {}
    });

    // Bind action buttons inside modal
    setTimeout(() => {
      document.getElementById('modal-action-students')?.addEventListener('click', () => {
        Modal.close();
        this.showEnrolledStudentsModal(course.id);
      });
      document.getElementById('modal-action-edit')?.addEventListener('click', () => {
        Modal.close();
        this.showCourseModal(course);
      });
      document.getElementById('modal-action-duplicate')?.addEventListener('click', () => {
        Modal.close();
        this.duplicateCourse(course.id);
      });
      document.getElementById('modal-action-publish')?.addEventListener('click', () => {
        Modal.close();
        this.changeStatus(course.id, 'published');
      });
      document.getElementById('modal-action-unpublish')?.addEventListener('click', () => {
        Modal.close();
        this.changeStatus(course.id, 'draft');
      });
      document.getElementById('modal-action-delete')?.addEventListener('click', () => {
        Modal.close();
        this.deleteCourse(course.id);
      });
      if (window.lucide) window.lucide.createIcons();
    }, 50);
  },

  showCourseModal(course = null) {
    const isEdit = !!course;
    Modal.show({
      title: isEdit ? 'Editar Curso' : 'Crear Nuevo Curso',
      confirmText: isEdit ? 'Guardar Cambios' : 'Crear Curso',
      showCancel: true,
      cancelText: 'Cancelar',
      content: `
        <form id="course-form" class="space-y-3.5 text-left text-xs">
          <div class="grid grid-cols-3 gap-3">
            <div class="col-span-2">
              <label class="block font-semibold text-slate-700 mb-1">Título del Curso *</label>
              <input type="text" id="course-title" required value="${course?.title || ''}" placeholder="Ej. Arquitectura de Microservicios con Go" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
            </div>
            <div>
              <label class="block font-semibold text-slate-700 mb-1">Código Único *</label>
              <input type="text" id="course-code" required value="${course?.code || ''}" placeholder="DXS-GO-201" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs font-mono">
            </div>
          </div>

          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block font-semibold text-slate-700 mb-1">Categoría</label>
              <input type="text" id="course-category" value="${course?.category || 'Tecnología'}" placeholder="Inteligencia Artificial" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
            </div>
            <div>
              <label class="block font-semibold text-slate-700 mb-1">Nivel</label>
              <select id="course-level" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs bg-white">
                <option value="Principiante" ${course?.level === 'Principiante' ? 'selected' : ''}>Principiante</option>
                <option value="Intermedio" ${course?.level === 'Intermedio' ? 'selected' : ''}>Intermedio</option>
                <option value="Avanzado" ${course?.level === 'Avanzado' ? 'selected' : ''}>Avanzado</option>
              </select>
            </div>
            <div>
              <label class="block font-semibold text-slate-700 mb-1">Duración (Horas)</label>
              <input type="number" step="0.5" id="course-duration" value="${course?.durationHours || 2.0}" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
            </div>
          </div>

          <div>
            <label class="block font-semibold text-slate-700 mb-1">Instructor</label>
            <input type="text" id="course-instructor" value="${course?.instructorName || 'Equipo DxSTech'}" placeholder="Dr. Alexander Gómez" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
          </div>

          <div>
            <label class="block font-semibold text-slate-700 mb-1">Imagen de Portada (URL o Subir Archivo)</label>
            <div class="flex gap-2">
              <input type="text" id="course-thumb" value="${course?.thumbnailUrl || ''}" placeholder="https://ejemplo.com/imagen.jpg" class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-xs">
              <label class="px-3 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs flex items-center gap-1 cursor-pointer transition-colors shrink-0">
                <i data-lucide="upload" class="w-3.5 h-3.5"></i>
                <span>Subir</span>
                <input type="file" id="course-thumb-file" accept="image/*" class="hidden">
              </label>
            </div>
          </div>

          <div>
            <label class="block font-semibold text-slate-700 mb-1">Descripción Corta</label>
            <input type="text" id="course-short-desc" value="${course?.shortDescription || ''}" placeholder="Resumen conciso en 1-2 líneas" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
          </div>

          <div>
            <label class="block font-semibold text-slate-700 mb-1">Descripción Completa</label>
            <textarea id="course-desc" rows="3" placeholder="Detalles de la capacitación..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">${course?.description || ''}</textarea>
          </div>
        </form>
      `,
      onConfirm: async () => {
        const title = document.getElementById('course-title').value.trim();
        const code = document.getElementById('course-code').value.trim();
        const category = document.getElementById('course-category').value.trim();
        const level = document.getElementById('course-level').value;
        const durationHours = parseFloat(document.getElementById('course-duration').value) || 1.0;
        const instructorName = document.getElementById('course-instructor').value.trim();
        const thumbnailUrl = document.getElementById('course-thumb').value.trim();
        const shortDescription = document.getElementById('course-short-desc').value.trim();
        const description = document.getElementById('course-desc').value.trim();

        if (!title || !code) {
          Toast.warning('Título y Código son obligatorios');
          return;
        }

        try {
          const payload = {
            title, code, category, level, durationHours,
            instructorName, thumbnailUrl, shortDescription, description,
            requirements: course?.requirements || '',
            learningObjectives: course?.learningObjectives || '',
            status: course?.status || 'draft',
          };

          const url = isEdit ? `/api/courses/${course.id}` : '/api/courses';
          const method = isEdit ? 'PUT' : 'POST';

          const res = await fetch(url, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
          });
          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          Toast.success(`Curso ${isEdit ? 'actualizado' : 'creado'} con éxito`);
          await this.fetchCourses();
          if (this.selectedCourse && this.selectedCourse.id === data.id) {
            this.selectedCourse = data;
            window.router?.navigate('courses');
          }
        } catch (e) {
          Toast.error(e.message);
        }
      }
    });

    // Bind Image Upload Helper
    setTimeout(() => {
      document.getElementById('course-thumb-file')?.addEventListener('change', async (e) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const formData = new FormData();
        formData.append('file', file);

        try {
          Toast.info('Subiendo imagen de portada...');
          const res = await fetch('/api/courses/upload', {
            method: 'POST',
            body: formData,
          });
          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          document.getElementById('course-thumb').value = data.url;
          Toast.success('Imagen subida correctamente');
        } catch (err) {
          Toast.error(err.message);
        }
      });
      if (window.lucide) window.lucide.createIcons();
    }, 50);
  },

  showModuleModal(courseId, mod = null) {
    const isEdit = !!mod;
    Modal.show({
      title: isEdit ? 'Editar Módulo' : 'Nuevo Módulo Temático',
      confirmText: isEdit ? 'Guardar Módulo' : 'Crear Módulo',
      showCancel: true,
      cancelText: 'Cancelar',
      content: `
        <form class="space-y-3 text-left text-xs">
          <div>
            <label class="block font-semibold text-slate-700 mb-1">Título del Módulo *</label>
            <input type="text" id="mod-title" required value="${mod?.title || ''}" placeholder="Ej. Módulo 1: Fundamentos y Conceptos Básicos" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
          </div>
          <div>
            <label class="block font-semibold text-slate-700 mb-1">Descripción Breve</label>
            <textarea id="mod-desc" rows="2" placeholder="Resumen del contenido del módulo..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">${mod?.description || ''}</textarea>
          </div>
          <div>
            <label class="block font-semibold text-slate-700 mb-1">Posición / Orden</label>
            <input type="number" id="mod-order" value="${mod?.orderIndex || 1}" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
          </div>
        </form>
      `,
      onConfirm: async () => {
        const title = document.getElementById('mod-title').value.trim();
        const description = document.getElementById('mod-desc').value.trim();
        const orderIndex = parseInt(document.getElementById('mod-order').value, 10) || 1;

        if (!title) {
          Toast.warning('El título del módulo es obligatorio');
          return;
        }

        try {
          const url = isEdit ? `/api/courses/modules/${mod.id}` : `/api/courses/${courseId}/modules`;
          const method = isEdit ? 'PUT' : 'POST';

          const res = await fetch(url, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, description, orderIndex }),
          });
          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          Toast.success(`Módulo ${isEdit ? 'actualizado' : 'creado'} exitosamente`);
          await this.openCourse(courseId);
        } catch (e) {
          Toast.error(e.message);
        }
      }
    });
  },

  showLessonModal(moduleId, lesson = null) {
    const isEdit = !!lesson;
    Modal.show({
      title: isEdit ? 'Editar Lección' : 'Nueva Lección de Aprendizaje',
      confirmText: isEdit ? 'Guardar Lección' : 'Crear Lección',
      showCancel: true,
      cancelText: 'Cancelar',
      content: `
        <form class="space-y-3.5 text-left text-xs">
          <div>
            <label class="block font-semibold text-slate-700 mb-1">Título de la Lección *</label>
            <input type="text" id="lesson-title" required value="${lesson?.title || ''}" placeholder="Ej. Introducción y Arquitectura General" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-semibold text-slate-700 mb-1">Tipo de Contenido *</label>
              <select id="lesson-type" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs bg-white font-semibold">
                <option value="youtube" ${lesson?.contentType === 'youtube' ? 'selected' : ''}>🎥 Video de YouTube</option>
                <option value="pdf" ${lesson?.contentType === 'pdf' ? 'selected' : ''}>📄 Documento PDF</option>
                <option value="pptx" ${lesson?.contentType === 'pptx' ? 'selected' : ''}>📊 Presentación PPTX</option>
                <option value="mp4" ${lesson?.contentType === 'mp4' ? 'selected' : ''}>🎬 Video MP4</option>
                <option value="text" ${lesson?.contentType === 'text' ? 'selected' : ''}>📝 Texto / Guía Teórica</option>
              </select>
            </div>
            <div>
              <label class="block font-semibold text-slate-700 mb-1">Duración (Minutos)</label>
              <input type="number" id="lesson-duration" value="${lesson?.durationMinutes || 15}" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
            </div>
          </div>

          <!-- URL / Upload File Container -->
          <div id="lesson-url-box">
            <label class="block font-semibold text-slate-700 mb-1">URL o Archivo Adjunto</label>
            <div class="flex gap-2">
              <input type="text" id="lesson-url" value="${lesson?.contentURL || ''}" placeholder="https://www.youtube.com/watch?v=... o archivo subido" class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-xs">
              <label id="lesson-upload-btn-label" class="px-3 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs flex items-center gap-1 cursor-pointer transition-colors shrink-0">
                <i data-lucide="upload" class="w-3.5 h-3.5"></i>
                <span>Subir</span>
                <input type="file" id="lesson-file-input" class="hidden">
              </label>
            </div>
          </div>

          <!-- Rich Text Body for text types -->
          <div id="lesson-body-box">
            <label class="block font-semibold text-slate-700 mb-1">Contenido de Lectura (Markdown / Texto)</label>
            <textarea id="lesson-body" rows="4" placeholder="Escribe aquí las instrucciones, apuntes o guías de estudio..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">${lesson?.contentBody || ''}</textarea>
          </div>

          <div>
            <label class="block font-semibold text-slate-700 mb-1">Descripción Breve de la Lección</label>
            <input type="text" id="lesson-desc" value="${lesson?.description || ''}" placeholder="Orientaciones para el estudiante..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
          </div>

          <div class="flex items-center gap-2 pt-1">
            <input type="checkbox" id="lesson-free" ${lesson?.isFreePreview ? 'checked' : ''} class="w-4 h-4 rounded text-indigo-600 focus:ring-indigo-500">
            <label for="lesson-free" class="text-xs font-semibold text-slate-700 cursor-pointer">
              Permitir como Vista Previa Gratuita (Free Preview)
            </label>
          </div>
        </form>
      `,
      onConfirm: async () => {
        const title = document.getElementById('lesson-title').value.trim();
        const contentType = document.getElementById('lesson-type').value;
        const durationMinutes = parseInt(document.getElementById('lesson-duration').value, 10) || 10;
        const contentUrl = document.getElementById('lesson-url').value.trim();
        const contentBody = document.getElementById('lesson-body').value.trim();
        const description = document.getElementById('lesson-desc').value.trim();
        const isFreePreview = document.getElementById('lesson-free').checked;

        if (!title) {
          Toast.warning('El título de la lección es obligatorio');
          return;
        }

        try {
          const payload = {
            title, contentType, durationMinutes, contentUrl, contentBody, description, isFreePreview,
            orderIndex: lesson?.orderIndex || 1,
          };

          const url = isEdit ? `/api/courses/lessons/${lesson.id}` : `/api/courses/modules/${moduleId}/lessons`;
          const method = isEdit ? 'PUT' : 'POST';

          const res = await fetch(url, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
          });
          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          Toast.success(`Lección ${isEdit ? 'actualizada' : 'creada'} exitosamente`);
          await this.openCourse(this.selectedCourse.id);
          this.selectedLesson = data;
          window.router?.navigate('courses');
        } catch (e) {
          Toast.error(e.message);
        }
      }
    });

    // File upload handler inside Lesson Modal
    setTimeout(() => {
      document.getElementById('lesson-file-input')?.addEventListener('change', async (e) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const formData = new FormData();
        formData.append('file', file);

        try {
          Toast.info(`Subiendo archivo: ${file.name}...`);
          const res = await fetch('/api/courses/upload', {
            method: 'POST',
            body: formData,
          });
          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          document.getElementById('lesson-url').value = data.url;
          Toast.success('Archivo subido al servidor correctamente');
        } catch (err) {
          Toast.error(err.message);
        }
      });
      if (window.lucide) window.lucide.createIcons();
    }, 50);
  },

  findAndEditLesson(lessonId) {
    if (!this.selectedCourse) return;
    for (const mod of this.selectedCourse.modules || []) {
      for (const lsn of mod.lessons || []) {
        if (lsn.id === lessonId) {
          this.showLessonModal(mod.id, lsn);
          return;
        }
      }
    }
  },

  async deleteModule(moduleId) {
    if (!confirm('¿Estás seguro de eliminar este módulo y todas sus lecciones?')) return;
    try {
      const res = await fetch(`/api/courses/modules/${moduleId}`, { method: 'DELETE' });
      if (!res.ok) throw new Error('Error al eliminar módulo');
      Toast.info('Módulo eliminado');
      await this.openCourse(this.selectedCourse.id);
    } catch (e) {
      Toast.error(e.message);
    }
  },

  async deleteLesson(lessonId) {
    if (!confirm('¿Estás seguro de eliminar esta lección?')) return;
    try {
      const res = await fetch(`/api/courses/lessons/${lessonId}`, { method: 'DELETE' });
      if (!res.ok) throw new Error('Error al eliminar lección');
      Toast.info('Lección eliminada');
      if (this.selectedLesson?.id === lessonId) this.selectedLesson = null;
      await this.openCourse(this.selectedCourse.id);
    } catch (e) {
      Toast.error(e.message);
    }
  },

  async duplicateCourse(courseId) {
    try {
      Toast.info('Duplicando programa educativo...');
      const res = await fetch(`/api/courses/${courseId}/duplicate`, { method: 'POST' });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error);

      Toast.success(`Curso duplicado con éxito: ${data.code}`);
      await this.fetchCourses();
    } catch (e) {
      Toast.error(e.message);
    }
  },

  async changeStatus(courseId, newStatus) {
    try {
      const res = await fetch(`/api/courses/${courseId}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error);

      Toast.success(`Estado actualizado a: ${newStatus}`);
      await this.fetchCourses();
    } catch (e) {
      Toast.error(e.message);
    }
  },

  async deleteCourse(courseId) {
    if (!confirm('¿Deseas eliminar definitivamente este curso? Esta acción no se puede deshacer.')) return;
    try {
      const res = await fetch(`/api/courses/${courseId}`, { method: 'DELETE' });
      if (!res.ok) throw new Error('No se pudo eliminar el curso');

      Toast.info('Curso eliminado correctamente');
      this.selectedCourse = null;
      this.selectedLesson = null;
      await this.fetchCourses();
    } catch (e) {
      Toast.error(e.message);
    }
  }
};
