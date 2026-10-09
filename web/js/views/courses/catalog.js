import { Modal } from '../../components/modal.js';
import { esc, safeUrl } from '../../utils/escape.js';

// Catálogo: tarjetas de curso, rejilla y menú de acciones.

export const catalogMethods = {
  renderCatalogView() {
    const user = window.router?.currentUser;
    const isAdmin = user && ['SUPERADMIN', 'ADMINISTRADOR'].includes(user.role);
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
              <input type="text" id="course-search-input" value="${esc(this.search)}" placeholder="Buscar curso o código..." class="w-full text-xs rounded-xl border border-slate-200 pl-9 pr-3.5 py-2 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
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
            <img src="${safeUrl(thumb)}" alt="${esc(c.title)}" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" onerror="this.src='${fallbackThumb}'">
            <div class="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent"></div>
            
            <div class="absolute top-3 left-3 flex flex-wrap gap-1.5">
              <span class="px-2.5 py-1 rounded-full text-[10px] font-bold bg-white/90 backdrop-blur-md text-slate-800 shadow-2xs">
                ${esc(c.category)}
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
                ${esc(c.code)}
              </span>
            </div>
          </div>

          <!-- Body Info -->
          <div class="p-5 flex-1 flex flex-col justify-between space-y-4">
            <div>
              <div class="flex items-center gap-2 text-slate-400 text-[11px] mb-1">
                <i data-lucide="user" class="w-3.5 h-3.5"></i>
                <span>${esc(c.instructorName || 'Docente Asignado')}</span>
                <span>•</span>
                <span class="font-medium text-indigo-600">${esc(c.level)}</span>
              </div>

              <h4 class="text-base font-bold text-slate-900 leading-snug group-hover:text-indigo-600 transition-colors">
                ${esc(c.title)}
              </h4>
              <p class="text-xs text-slate-500 mt-2 line-clamp-2 leading-relaxed">
                ${esc(c.shortDescription || c.description || 'Sin descripción disponible.')}
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
                  <button data-course-cert="${esc(c.id)}" class="p-2 rounded-xl text-amber-600 bg-amber-50 hover:bg-amber-100 border border-amber-200 transition-colors flex items-center justify-center shadow-2xs cursor-pointer" title="Descargar Certificado Oficial">
                    <i data-lucide="award" class="w-4 h-4"></i>
                  </button>
                ` : ''}
                ${isEnrolled ? `
                  <button data-open-course="${esc(c.id)}" class="px-3.5 py-1.5 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-colors flex items-center gap-1 shadow-xs cursor-pointer">
                    <span>Continuar</span>
                    <i data-lucide="play" class="w-3 h-3"></i>
                  </button>
                ` : `
                  <button data-open-course="${esc(c.id)}" class="px-3 py-1.5 rounded-xl text-xs font-bold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center gap-1 cursor-pointer">
                    <span>${isAdmin ? 'Gestionar' : 'Explorar'}</span>
                    <i data-lucide="arrow-right" class="w-3 h-3"></i>
                  </button>
                `}

                ${isAdmin ? `
                  <button data-course-menu="${esc(c.id)}" class="p-1.5 rounded-xl text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors cursor-pointer" title="Opciones Administrativas">
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

  showCourseActionMenu(courseId) {
    const course = this.courses.find(c => c.id === courseId);
    if (!course) return;

    Modal.show({
      title: `Opciones de Curso: ${course.code}`,
      confirmText: 'Cerrar',
      showCancel: false,
      content: `
        <div class="space-y-2 text-left text-xs">
          <p class="font-bold text-slate-800 text-sm mb-3">${esc(course.title)}</p>
          
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
};
