import { Toast } from '../components/toast.js';
import { catalogMethods } from './courses/catalog.js';
import { classroomMethods } from './courses/classroom.js';
import { adminMethods } from './courses/admin.js';
import { lessonQuizMethods } from './courses/lesson-quiz.js';
import { communityMethods } from './courses/community.js';

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

    // Tutor IA button from classroom header
    const openTutorBtn = document.getElementById('open-tutor-modal-btn');
    if (openTutorBtn) {
      openTutorBtn.addEventListener('click', () => this.openTutorModal());
    }

    // Discussions forum button from classroom header
    const openDiscBtn = document.getElementById('open-discussions-modal-btn');
    if (openDiscBtn) {
      openDiscBtn.addEventListener('click', () => this.openDiscussionsModal());
    }

    // Reviews button from classroom header
    const openReviewsBtn = document.getElementById('open-reviews-modal-btn');
    if (openReviewsBtn) {
      openReviewsBtn.addEventListener('click', () => this.openReviewsModal());
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

    // Check if current lesson is an interactive quiz
    if (this.selectedLesson && this.selectedLesson.contentType === 'quiz') {
      this.loadLessonQuiz(this.selectedLesson);
    }
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
};

// Las vistas grandes se dividen por responsabilidad; los métodos se mezclan en la vista
// y siguen usando `this`, por lo que su comportamiento no cambia.
Object.assign(CoursesView, catalogMethods, classroomMethods, adminMethods, lessonQuizMethods, communityMethods);
