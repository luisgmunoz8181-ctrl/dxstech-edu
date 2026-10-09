import { Toast } from '../components/toast.js';
import { esc } from '../utils/escape.js';

export const DashboardView = {
  overview: null,
  courseMetrics: [],
  userTimeline: [],
  enrollTimeline: [],
  students: [],
  courses: [],
  isLoading: false,
  filters: {
    courseId: '',
    category: '',
    status: 'all',
    search: '',
    startDate: '',
    endDate: '',
  },

  render() {
    return `
      <div class="space-y-6 max-w-7xl mx-auto">
        <!-- Header Banner -->
        <div class="bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 rounded-3xl p-6 md:p-8 text-white shadow-xl flex flex-col lg:flex-row items-start lg:items-center justify-between gap-6 border border-indigo-900/40">
          <div class="space-y-1.5">
            <div class="flex items-center gap-2">
              <span class="px-3 py-1 rounded-full text-[10px] font-bold bg-indigo-500/20 text-indigo-300 border border-indigo-400/30 uppercase tracking-wider flex items-center gap-1.5">
                <i data-lucide="bar-chart-3" class="w-3.5 h-3.5"></i>
                Panel Administrativo LMS
              </span>
              <span class="px-2.5 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-500/20 text-emerald-300 border border-emerald-400/30">
                Tiempo Real
              </span>
            </div>
            <h3 class="text-2xl font-black tracking-tight text-white">Métricas & Monitoreo Académico</h3>
            <p class="text-xs md:text-sm text-slate-300 max-w-2xl leading-relaxed">
              Supervisión integral de estudiantes, efectividad de cursos, retención, diplomas oficiales emitidos y descarga de reportes ejecutivos para auditoría.
            </p>
          </div>

          <!-- Quick Report Download Buttons -->
          <div class="flex flex-wrap items-center gap-3 shrink-0">
            <button id="btn-export-enrollments" class="px-4 py-2.5 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-500 transition-all shadow-md shadow-indigo-900/50 flex items-center gap-2 border border-indigo-500/30">
              <i data-lucide="file-spreadsheet" class="w-4 h-4 text-indigo-200"></i>
              <span>Exportar Matrículas (CSV)</span>
            </button>
            <button id="btn-export-certificates" class="px-4 py-2.5 rounded-xl text-xs font-bold text-slate-200 bg-slate-800/80 hover:bg-slate-700 transition-all border border-slate-700 flex items-center gap-2">
              <i data-lucide="award" class="w-4 h-4 text-amber-400"></i>
              <span>Reporte Certificados (CSV)</span>
            </button>
            <button id="btn-refresh-dashboard" class="p-2.5 rounded-xl text-slate-300 hover:text-white bg-slate-800/80 hover:bg-slate-700 transition-all border border-slate-700" title="Actualizar datos">
              <i data-lucide="refresh-cw" class="w-4 h-4"></i>
            </button>
          </div>
        </div>

        <!-- 1. Executive KPI Cards (Grid 4) -->
        <div id="kpi-cards-container" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-5">
          <!-- KPI 1: Usuarios -->
          <div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs relative overflow-hidden group hover:border-indigo-300 transition-all">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold text-slate-500 uppercase tracking-wider">Comunidad & Alumnos</span>
              <div class="w-10 h-10 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center">
                <i data-lucide="users" class="w-5 h-5"></i>
              </div>
            </div>
            <div class="mt-4">
              <div id="kpi-total-users" class="text-3xl font-extrabold text-slate-900 tracking-tight">--</div>
              <p class="text-xs text-slate-500 mt-1 flex items-center gap-1.5 font-medium">
                <span id="kpi-active-users" class="text-emerald-600 font-semibold">-- activos</span>
                <span>•</span>
                <span id="kpi-students-count" class="text-slate-600">-- estudiantes</span>
              </p>
            </div>
          </div>

          <!-- KPI 2: Cursos -->
          <div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs relative overflow-hidden group hover:border-emerald-300 transition-all">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold text-slate-500 uppercase tracking-wider">Oferta de Cursos</span>
              <div class="w-10 h-10 rounded-2xl bg-emerald-50 text-emerald-600 flex items-center justify-center">
                <i data-lucide="book-open" class="w-5 h-5"></i>
              </div>
            </div>
            <div class="mt-4">
              <div id="kpi-total-courses" class="text-3xl font-extrabold text-slate-900 tracking-tight">--</div>
              <p class="text-xs text-slate-500 mt-1 flex items-center gap-1.5 font-medium">
                <span id="kpi-published-courses" class="text-emerald-600 font-semibold">-- publicados</span>
                <span>•</span>
                <span id="kpi-draft-courses" class="text-amber-600">-- borradores</span>
              </p>
            </div>
          </div>

          <!-- KPI 3: Matrículas & Finalización -->
          <div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs relative overflow-hidden group hover:border-purple-300 transition-all">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold text-slate-500 uppercase tracking-wider">Matrículas & Éxito</span>
              <div class="w-10 h-10 rounded-2xl bg-purple-50 text-purple-600 flex items-center justify-center">
                <i data-lucide="graduation-cap" class="w-5 h-5"></i>
              </div>
            </div>
            <div class="mt-4">
              <div id="kpi-total-enrollments" class="text-3xl font-extrabold text-slate-900 tracking-tight">--</div>
              <p class="text-xs text-slate-500 mt-1 flex items-center gap-1.5 font-medium">
                <span id="kpi-completed-enrollments" class="text-purple-600 font-semibold">-- completadas</span>
                <span>•</span>
                <span id="kpi-completion-rate" class="text-emerald-600 font-bold">--% éxito</span>
              </p>
            </div>
          </div>

          <!-- KPI 4: Certificaciones & Horas -->
          <div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs relative overflow-hidden group hover:border-amber-300 transition-all">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold text-slate-500 uppercase tracking-wider">Diplomas Emitidos</span>
              <div class="w-10 h-10 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center">
                <i data-lucide="award" class="w-5 h-5"></i>
              </div>
            </div>
            <div class="mt-4">
              <div id="kpi-total-certificates" class="text-3xl font-extrabold text-slate-900 tracking-tight">--</div>
              <p class="text-xs text-slate-500 mt-1 flex items-center gap-1.5 font-medium">
                <span id="kpi-total-hours" class="text-amber-700 font-bold">-- hrs</span>
                <span>formación acreditada</span>
              </p>
            </div>
          </div>
        </div>

        <!-- 2. Interactive SVG Charts Section -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Chart 1: Desempeño por Curso -->
          <div class="bg-white rounded-3xl p-6 border border-slate-200/80 shadow-xs flex flex-col justify-between">
            <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4">
              <div>
                <h4 class="text-sm font-bold text-slate-800 flex items-center gap-2">
                  <i data-lucide="layers" class="w-4 h-4 text-indigo-600"></i>
                  Rendimiento y Alumnos por Curso
                </h4>
                <p class="text-[11px] text-slate-400">Total inscritos, completados y tasa promedio de avance</p>
              </div>

              <!-- Category filter for course chart -->
              <select id="chart-category-filter" class="text-xs rounded-xl border border-slate-200 px-3 py-1.5 bg-slate-50 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
                <option value="">Todas las categorías</option>
              </select>
            </div>

            <!-- Chart SVG / Bars Container -->
            <div id="courses-chart-container" class="space-y-4 pt-2">
              <div class="py-12 text-center text-slate-400 text-xs">Cargando métricas de cursos...</div>
            </div>
          </div>

          <!-- Chart 2: Tendencia de Crecimiento & Matrículas -->
          <div class="bg-white rounded-3xl p-6 border border-slate-200/80 shadow-xs flex flex-col justify-between">
            <div class="flex items-center justify-between mb-4">
              <div>
                <h4 class="text-sm font-bold text-slate-800 flex items-center gap-2">
                  <i data-lucide="trending-up" class="w-4 h-4 text-emerald-600"></i>
                  Historial de Crecimiento Mensual
                </h4>
                <p class="text-[11px] text-slate-400">Nuevos usuarios registrados e inscripciones en el tiempo</p>
              </div>

              <div class="flex items-center gap-3 text-[11px] font-semibold">
                <span class="flex items-center gap-1.5 text-indigo-600">
                  <span class="w-2.5 h-2.5 rounded-sm bg-indigo-600"></span> Usuarios
                </span>
                <span class="flex items-center gap-1.5 text-emerald-600">
                  <span class="w-2.5 h-2.5 rounded-sm bg-emerald-500"></span> Matrículas
                </span>
              </div>
            </div>

            <!-- Timeline Chart Container -->
            <div id="timeline-chart-container" class="pt-2">
              <div class="py-12 text-center text-slate-400 text-xs">Cargando gráfico de tendencias...</div>
            </div>
          </div>
        </div>

        <!-- 3. Granular Student Monitoring Table & Live Filters -->
        <div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs overflow-hidden">
          <!-- Filters Toolbar -->
          <div class="p-5 border-b border-slate-100 space-y-4 bg-slate-50/60">
            <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
              <div>
                <h4 class="text-sm font-bold text-slate-800 flex items-center gap-2">
                  <i data-lucide="user-check" class="w-4 h-4 text-indigo-600"></i>
                  Monitoreo de Estudiantes & Progreso
                </h4>
                <p class="text-[11px] text-slate-400">Trazabilidad individual de avance, estatus y certificados conferidos</p>
              </div>

              <button id="btn-clear-filters" class="text-xs font-semibold text-slate-500 hover:text-indigo-600 transition-colors flex items-center gap-1">
                <i data-lucide="x-circle" class="w-3.5 h-3.5"></i>
                Limpiar Filtros
              </button>
            </div>

            <!-- Filter Controls Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
              <!-- Search Input -->
              <div class="relative lg:col-span-1">
                <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
                  <i data-lucide="search" class="w-3.5 h-3.5"></i>
                </span>
                <input type="text" id="filter-search" placeholder="Buscar estudiante o curso..." class="w-full text-xs rounded-xl border border-slate-200 pl-8 pr-3 py-2 bg-white focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
              </div>

              <!-- Course Select -->
              <div>
                <select id="filter-course" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2 bg-white focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
                  <option value="">Todos los cursos</option>
                </select>
              </div>

              <!-- Status Select -->
              <div>
                <select id="filter-status" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2 bg-white focus:ring-2 focus:ring-indigo-500 focus:outline-hidden">
                  <option value="all">Todos los estados</option>
                  <option value="active">En progreso (Activo)</option>
                  <option value="completed">Completado (100%)</option>
                </select>
              </div>

              <!-- Start Date -->
              <div>
                <input type="date" id="filter-start-date" title="Inscrito desde" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2 bg-white focus:ring-2 focus:ring-indigo-500 focus:outline-hidden text-slate-600">
              </div>

              <!-- End Date -->
              <div>
                <input type="date" id="filter-end-date" title="Inscrito hasta" class="w-full text-xs rounded-xl border border-slate-200 px-3 py-2 bg-white focus:ring-2 focus:ring-indigo-500 focus:outline-hidden text-slate-600">
              </div>
            </div>
          </div>

          <!-- Students Table -->
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-50 border-b border-slate-200/80 text-slate-500 uppercase tracking-wider text-[10px] font-semibold">
                <tr>
                  <th class="py-3 px-5">Estudiante</th>
                  <th class="py-3 px-4">Curso Asignado</th>
                  <th class="py-3 px-4">Avance (%)</th>
                  <th class="py-3 px-4">Estado</th>
                  <th class="py-3 px-4">Certificado</th>
                  <th class="py-3 px-4">Fecha Inscripción</th>
                  <th class="py-3 px-4">Último Acceso</th>
                  <th class="py-3 px-5 text-right">Acciones</th>
                </tr>
              </thead>
              <tbody id="monitoring-tbody" class="divide-y divide-slate-100">
                <tr>
                  <td colspan="8" class="py-12 text-center text-slate-400">Cargando registros de estudiantes...</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Table Footer / Count -->
          <div class="p-4 bg-slate-50/70 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500">
            <span id="monitoring-count">Mostrando 0 registros</span>
            <span class="text-[11px] text-slate-400">Datos auditados en tiempo real</span>
          </div>
        </div>
      </div>
    `;
  },

  async mount() {
    this.bindEvents();
    await this.loadAllData();
    if (window.lucide) window.lucide.createIcons();
  },

  bindEvents() {
    // Refresh button
    const refreshBtn = document.getElementById('btn-refresh-dashboard');
    if (refreshBtn) {
      refreshBtn.addEventListener('click', () => {
        refreshBtn.classList.add('animate-spin');
        this.loadAllData().finally(() => {
          setTimeout(() => refreshBtn.classList.remove('animate-spin'), 600);
        });
      });
    }

    // Export Enrollments CSV
    const expEnrollBtn = document.getElementById('btn-export-enrollments');
    if (expEnrollBtn) {
      expEnrollBtn.addEventListener('click', () => this.exportEnrollments());
    }

    // Export Certificates CSV
    const expCertBtn = document.getElementById('btn-export-certificates');
    if (expCertBtn) {
      expCertBtn.addEventListener('click', () => this.exportCertificates());
    }

    // Clear filters
    const clearBtn = document.getElementById('btn-clear-filters');
    if (clearBtn) {
      clearBtn.addEventListener('click', () => {
        this.filters = {
          courseId: '',
          category: '',
          status: 'all',
          search: '',
          startDate: '',
          endDate: '',
        };
        document.getElementById('filter-search').value = '';
        document.getElementById('filter-course').value = '';
        document.getElementById('filter-status').value = 'all';
        document.getElementById('filter-start-date').value = '';
        document.getElementById('filter-end-date').value = '';
        this.fetchMonitoring();
      });
    }

    // Filter Search with Debounce
    const searchInput = document.getElementById('filter-search');
    if (searchInput) {
      let timeout;
      searchInput.addEventListener('input', (e) => {
        clearTimeout(timeout);
        this.filters.search = e.target.value.trim();
        timeout = setTimeout(() => this.fetchMonitoring(), 300);
      });
    }

    // Filter Course
    const courseSelect = document.getElementById('filter-course');
    if (courseSelect) {
      courseSelect.addEventListener('change', (e) => {
        this.filters.courseId = e.target.value;
        this.fetchMonitoring();
      });
    }

    // Filter Status
    const statusSelect = document.getElementById('filter-status');
    if (statusSelect) {
      statusSelect.addEventListener('change', (e) => {
        this.filters.status = e.target.value;
        this.fetchMonitoring();
      });
    }

    // Filter Dates
    const startInput = document.getElementById('filter-start-date');
    if (startInput) {
      startInput.addEventListener('change', (e) => {
        this.filters.startDate = e.target.value;
        this.fetchMonitoring();
      });
    }

    const endInput = document.getElementById('filter-end-date');
    if (endInput) {
      endInput.addEventListener('change', (e) => {
        this.filters.endDate = e.target.value;
        this.fetchMonitoring();
      });
    }

    // Chart Category Filter
    const chartCatFilter = document.getElementById('chart-category-filter');
    if (chartCatFilter) {
      chartCatFilter.addEventListener('change', (e) => {
        this.fetchCourseMetrics(e.target.value);
      });
    }
  },

  async loadAllData() {
    this.isLoading = true;
    try {
      await Promise.all([
        this.fetchOverview(),
        this.fetchCourseMetrics(''),
        this.fetchTimelines(),
        this.fetchCoursesList(),
        this.fetchMonitoring(),
      ]);
    } catch (err) {
      console.error('Error cargando dashboard:', err);
      Toast.error('Error al sincronizar datos del panel administrativo: ' + err.message);
    } finally {
      this.isLoading = false;
      if (window.lucide) window.lucide.createIcons();
    }
  },

  async fetchOverview() {
    const res = await fetch('/api/admin/metrics/overview');
    if (!res.ok) throw new Error('Error consultando resumen general');
    this.overview = await res.json();
    this.renderKPIs();
  },

  renderKPIs() {
    if (!this.overview) return;

    // Users
    document.getElementById('kpi-total-users').textContent = this.overview.totalUsers;
    document.getElementById('kpi-active-users').textContent = `${this.overview.activeUsers} activos`;
    document.getElementById('kpi-students-count').textContent = `${this.overview.studentsCount} alumnos`;

    // Courses
    document.getElementById('kpi-total-courses').textContent = this.overview.totalCourses;
    document.getElementById('kpi-published-courses').textContent = `${this.overview.publishedCourses} publicados`;
    document.getElementById('kpi-draft-courses').textContent = `${this.overview.draftCourses} borradores`;

    // Enrollments
    document.getElementById('kpi-total-enrollments').textContent = this.overview.totalEnrollments;
    document.getElementById('kpi-completed-enrollments').textContent = `${this.overview.completedEnrollments} completadas`;
    document.getElementById('kpi-completion-rate').textContent = `${(this.overview.completionRate || 0).toFixed(1)}% tasa finalización`;

    // Certificates
    document.getElementById('kpi-total-certificates').textContent = this.overview.totalCertificates;
    document.getElementById('kpi-total-hours').textContent = `${(this.overview.totalHoursDelivered || 0).toFixed(0)} hrs`;
  },

  async fetchCourseMetrics(category = '') {
    const url = category ? `/api/admin/metrics/courses?category=${encodeURIComponent(category)}` : '/api/admin/metrics/courses';
    const res = await fetch(url);
    if (!res.ok) throw new Error('Error consultando métricas por curso');
    this.courseMetrics = await res.json();
    this.renderCourseMetricsChart();
  },

  renderCourseMetricsChart() {
    const container = document.getElementById('courses-chart-container');
    if (!container) return;

    if (!this.courseMetrics || this.courseMetrics.length === 0) {
      container.innerHTML = `
        <div class="py-12 text-center text-slate-400 text-xs">
          No hay cursos registrados para esta categoría.
        </div>
      `;
      return;
    }

    // Render horizontal progress bar per course
    const html = this.courseMetrics.slice(0, 6).map(cm => {
      const completionRate = cm.completionRate ? cm.completionRate.toFixed(1) : '0.0';
      const avgProgress = cm.averageProgress ? cm.averageProgress.toFixed(1) : '0.0';
      return `
        <div class="p-3.5 rounded-2xl bg-slate-50/80 border border-slate-100 hover:border-indigo-200 transition-all space-y-2">
          <div class="flex items-center justify-between gap-3">
            <div class="truncate">
              <span class="text-xs font-bold text-slate-800 truncate block">${esc(cm.courseTitle)}</span>
              <span class="text-[10px] text-slate-400 font-medium">${esc(cm.courseCode)} • ${esc(cm.category)} • ${cm.durationHours} hrs</span>
            </div>
            <div class="text-right shrink-0">
              <span class="text-xs font-extrabold text-indigo-700">${cm.totalStudents}</span>
              <span class="text-[10px] text-slate-400 block">${cm.completedStudents} finalizados</span>
            </div>
          </div>

          <!-- Progress Bar Comparison -->
          <div class="space-y-1">
            <div class="flex justify-between text-[10px] text-slate-500 font-medium">
              <span>Avance promedio de la cohorte: <strong class="text-indigo-600">${avgProgress}%</strong></span>
              <span>Completitud: <strong class="text-emerald-600">${completionRate}%</strong></span>
            </div>
            <div class="w-full bg-slate-200/80 rounded-full h-2 overflow-hidden flex">
              <div class="bg-gradient-to-r from-indigo-500 to-indigo-600 h-2 rounded-full transition-all duration-500" style="width: ${Math.min(100, Math.max(0, cm.averageProgress))}%"></div>
            </div>
          </div>
        </div>
      `;
    }).join('');

    container.innerHTML = html;
  },

  async fetchTimelines() {
    const [resUsers, resEnroll] = await Promise.all([
      fetch('/api/admin/metrics/timeline/users'),
      fetch('/api/admin/metrics/timeline/enrollments'),
    ]);

    if (resUsers.ok) this.userTimeline = await resUsers.json();
    if (resEnroll.ok) this.enrollTimeline = await resEnroll.json();

    this.renderTimelineChart();
  },

  renderTimelineChart() {
    const container = document.getElementById('timeline-chart-container');
    if (!container) return;

    // Combine monthly periods
    const periodsSet = new Set();
    this.userTimeline.forEach(p => periodsSet.add(p.period));
    this.enrollTimeline.forEach(p => periodsSet.add(p.period));

    const periods = Array.from(periodsSet).sort();

    if (periods.length === 0) {
      container.innerHTML = `
        <div class="py-12 text-center text-slate-400 text-xs">
          Aún no hay suficiente historial mensual acumulado.
        </div>
      `;
      return;
    }

    // Map counts
    const userMap = new Map(this.userTimeline.map(p => [p.period, p.count]));
    const enrollMap = new Map(this.enrollTimeline.map(p => [p.period, p.count]));

    let maxVal = 1;
    periods.forEach(p => {
      const u = userMap.get(p) || 0;
      const e = enrollMap.get(p) || 0;
      if (u > maxVal) maxVal = u;
      if (e > maxVal) maxVal = e;
    });

    // Render interactive Column Bars
    const barsHtml = periods.map(period => {
      const uCount = userMap.get(period) || 0;
      const eCount = enrollMap.get(period) || 0;
      const uHeight = Math.max(8, Math.round((uCount / maxVal) * 120));
      const eHeight = Math.max(8, Math.round((eCount / maxVal) * 120));

      return `
        <div class="flex-1 flex flex-col items-center gap-1.5 group">
          <div class="h-32 w-full flex items-end justify-center gap-1.5 px-1">
            <!-- Users Bar -->
            <div class="w-1/2 bg-indigo-500 hover:bg-indigo-600 rounded-t-md transition-all relative cursor-pointer" style="height: ${uHeight}px;">
              <div class="absolute -top-7 left-1/2 -translate-x-1/2 bg-slate-900 text-white text-[9px] font-bold px-1.5 py-0.5 rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-10">
                ${uCount} usuarios
              </div>
            </div>
            <!-- Enrollments Bar -->
            <div class="w-1/2 bg-emerald-500 hover:bg-emerald-600 rounded-t-md transition-all relative cursor-pointer" style="height: ${eHeight}px;">
              <div class="absolute -top-7 left-1/2 -translate-x-1/2 bg-slate-900 text-white text-[9px] font-bold px-1.5 py-0.5 rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-10">
                ${eCount} matrículas
              </div>
            </div>
          </div>
          <span class="text-[10px] font-semibold text-slate-500">${period}</span>
        </div>
      `;
    }).join('');

    container.innerHTML = `
      <div class="pt-6 pb-2 px-2 bg-slate-50/70 rounded-2xl border border-slate-100">
        <div class="flex items-end justify-between gap-2 h-36">
          ${barsHtml}
        </div>
      </div>
    `;
  },

  async fetchCoursesList() {
    try {
      const res = await fetch('/api/courses');
      if (res.ok) {
        this.courses = await res.json();
        const courseSelect = document.getElementById('filter-course');
        const catSelect = document.getElementById('chart-category-filter');

        if (courseSelect) {
          const currentVal = courseSelect.value;
          courseSelect.innerHTML = `<option value="">Todos los cursos</option>` +
            this.courses.map(c => `<option value="${esc(c.id)}">${esc(c.title)} (${esc(c.code)})</option>`).join('');
          courseSelect.value = currentVal;
        }

        if (catSelect) {
          const categories = Array.from(new Set(this.courses.map(c => c.category).filter(Boolean)));
          catSelect.innerHTML = `<option value="">Todas las categorías</option>` +
            categories.map(cat => `<option value="${cat}">${cat}</option>`).join('');
        }
      }
    } catch (e) {
      console.warn('Error poblando lista de cursos:', e);
    }
  },

  async fetchMonitoring() {
    try {
      const params = new URLSearchParams();
      if (this.filters.courseId) params.append('courseId', this.filters.courseId);
      if (this.filters.category) params.append('category', this.filters.category);
      if (this.filters.status && this.filters.status !== 'all') params.append('status', this.filters.status);
      if (this.filters.search) params.append('search', this.filters.search);
      if (this.filters.startDate) params.append('startDate', this.filters.startDate);
      if (this.filters.endDate) params.append('endDate', this.filters.endDate);

      const res = await fetch(`/api/admin/monitoring/students?${params.toString()}`);
      if (!res.ok) throw new Error('Error consultando monitoreo de alumnos');

      this.students = await res.json();
      this.renderMonitoringTable();
    } catch (err) {
      Toast.error(err.message);
    }
  },

  renderMonitoringTable() {
    const tbody = document.getElementById('monitoring-tbody');
    const countEl = document.getElementById('monitoring-count');
    if (!tbody) return;

    if (countEl) {
      countEl.textContent = `Mostrando ${this.students.length} matrícula${this.students.length === 1 ? '' : 's'}`;
    }

    if (!this.students || this.students.length === 0) {
      tbody.innerHTML = `
        <tr>
          <td colspan="8" class="py-12 text-center text-slate-400">
            <div class="flex flex-col items-center gap-2">
              <i data-lucide="inbox" class="w-8 h-8 text-slate-300"></i>
              <p>No se encontraron registros de estudiantes con los filtros actuales.</p>
            </div>
          </td>
        </tr>
      `;
      if (window.lucide) window.lucide.createIcons();
      return;
    }

    const rows = this.students.map(r => {
      const initials = r.studentName ? r.studentName.split(' ').map(n => n.charAt(0)).slice(0, 2).join('').toUpperCase() : 'ES';
      const isCompleted = r.status === 'completed';
      const progress = (r.progressPercent || 0).toFixed(0);

      const statusBadge = isCompleted
        ? `<span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">Completado</span>`
        : `<span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-indigo-50 text-indigo-700 border border-indigo-200">En Curso</span>`;

      const certBadge = r.certificateId
        ? `<a href="/api/certificates/${esc(r.certificateId)}/pdf" target="_blank" class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[10px] font-bold bg-amber-50 text-amber-800 border border-amber-300 hover:bg-amber-100 transition-colors" title="Descargar diploma oficial">
            <i data-lucide="award" class="w-3 h-3 text-amber-600"></i>
            <span>${esc(r.certificateId)}</span>
          </a>`
        : `<span class="text-[10px] text-slate-400 italic">Pendiente (100%)</span>`;

      const enrollDate = r.enrolledAt ? new Date(r.enrolledAt).toLocaleDateString() : '--';
      const lastAccess = r.lastAccessedAt ? new Date(r.lastAccessedAt).toLocaleString() : '--';

      return `
        <tr class="hover:bg-slate-50/80 transition-colors">
          <!-- Student -->
          <td class="py-3 px-5">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-full bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center text-xs shrink-0">
                ${initials}
              </div>
              <div>
                <span class="font-bold text-slate-800 block">${esc(r.studentName)}</span>
                <span class="text-[11px] text-slate-400">${esc(r.studentEmail)}</span>
              </div>
            </div>
          </td>

          <!-- Course -->
          <td class="py-3 px-4">
            <span class="font-semibold text-slate-800 block truncate max-w-xs">${esc(r.courseTitle)}</span>
            <div class="flex items-center gap-1.5 mt-0.5">
              <span class="px-1.5 py-0.2 rounded text-[9px] font-mono font-bold bg-slate-100 text-slate-600">${esc(r.courseCode)}</span>
              <span class="text-[10px] text-slate-400">${esc(r.category)}</span>
            </div>
          </td>

          <!-- Progress -->
          <td class="py-3 px-4">
            <div class="w-28 space-y-1">
              <div class="flex justify-between text-[10px] font-bold ${isCompleted ? 'text-emerald-700' : 'text-slate-700'}">
                <span>${progress}%</span>
              </div>
              <div class="w-full bg-slate-200 rounded-full h-1.5 overflow-hidden">
                <div class="${isCompleted ? 'bg-emerald-500' : 'bg-indigo-600'} h-1.5 rounded-full" style="width: ${progress}%"></div>
              </div>
            </div>
          </td>

          <!-- Status -->
          <td class="py-3 px-4">
            ${statusBadge}
          </td>

          <!-- Certificate -->
          <td class="py-3 px-4">
            ${certBadge}
          </td>

          <!-- Dates -->
          <td class="py-3 px-4 text-slate-500 whitespace-nowrap">
            ${enrollDate}
          </td>
          <td class="py-3 px-4 text-slate-500 whitespace-nowrap">
            ${lastAccess}
          </td>

          <!-- Actions -->
          <td class="py-3 px-5 text-right whitespace-nowrap">
            ${r.certificateId ? `
              <a href="/api/certificates/${esc(r.certificateId)}/pdf" target="_blank" class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-semibold text-indigo-700 bg-indigo-50 hover:bg-indigo-100 border border-indigo-200 transition-colors">
                <i data-lucide="download" class="w-3.5 h-3.5"></i>
                <span>PDF</span>
              </a>
            ` : `
              <span class="text-[11px] text-slate-300">-</span>
            `}
          </td>
        </tr>
      `;
    }).join('');

    tbody.innerHTML = rows;
    if (window.lucide) window.lucide.createIcons();
  },

  async exportEnrollments() {
    const params = new URLSearchParams();
    if (this.filters.courseId) params.append('courseId', this.filters.courseId);
    if (this.filters.category) params.append('category', this.filters.category);
    if (this.filters.status && this.filters.status !== 'all') params.append('status', this.filters.status);
    if (this.filters.search) params.append('search', this.filters.search);
    if (this.filters.startDate) params.append('startDate', this.filters.startDate);
    if (this.filters.endDate) params.append('endDate', this.filters.endDate);

    const url = `/api/admin/reports/enrollments.csv?${params.toString()}`;
    await this.downloadCSV(url, `reporte_matriculas_${new Date().toISOString().slice(0, 10)}.csv`);
  },

  async exportCertificates() {
    const url = `/api/admin/reports/certificates.csv`;
    await this.downloadCSV(url, `reporte_certificados_${new Date().toISOString().slice(0, 10)}.csv`);
  },

  async downloadCSV(url, defaultFilename) {
    try {
      Toast.info('Generando reporte CSV con codificación Excel UTF-8...');
      const res = await fetch(url);
      if (!res.ok) throw new Error('Error descargando reporte');

      const blob = await res.blob();
      const downloadUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = defaultFilename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(downloadUrl);
      Toast.success('¡Reporte generado y descargado correctamente!');
    } catch (err) {
      Toast.error('Fallo en la descarga: ' + err.message);
    }
  },
};
