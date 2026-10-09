import { Toast } from '../../components/toast.js';
import { Modal } from '../../components/modal.js';
import { esc, safeUrl } from '../../utils/escape.js';

// Administración: cursos, módulos, lecciones, estudiantes matriculados.

export const adminMethods = {
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
                ${esc((e.studentName || 'U').charAt(0))}
              </div>
              <div>
                <p class="font-bold text-slate-800">${esc(e.studentName)}</p>
                <p class="text-[11px] text-slate-400">${esc(e.studentEmail)}</p>
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
        if (container) container.innerHTML = `<p class="text-rose-500 py-4 text-center">${esc(err.message)}</p>`;
      }
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
              <input type="text" id="course-title" required value="${esc(course?.title || '')}" placeholder="Ej. Arquitectura de Microservicios con Go" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
            </div>
            <div>
              <label class="block font-semibold text-slate-700 mb-1">Código Único *</label>
              <input type="text" id="course-code" required value="${esc(course?.code || '')}" placeholder="DXS-GO-201" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs font-mono">
            </div>
          </div>

          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block font-semibold text-slate-700 mb-1">Categoría</label>
              <input type="text" id="course-category" value="${esc(course?.category || 'Tecnología')}" placeholder="Inteligencia Artificial" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
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
            <input type="text" id="course-instructor" value="${esc(course?.instructorName || 'Equipo DxSTech')}" placeholder="Dr. Alexander Gómez" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
          </div>

          <div>
            <label class="block font-semibold text-slate-700 mb-1">Imagen de Portada (URL o Subir Archivo)</label>
            <div class="flex gap-2">
              <input type="text" id="course-thumb" value="${safeUrl(course?.thumbnailUrl || '')}" placeholder="https://ejemplo.com/imagen.jpg" class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-xs">
              <label class="px-3 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs flex items-center gap-1 cursor-pointer transition-colors shrink-0">
                <i data-lucide="upload" class="w-3.5 h-3.5"></i>
                <span>Subir</span>
                <input type="file" id="course-thumb-file" accept="image/*" class="hidden">
              </label>
            </div>
          </div>

          <div>
            <label class="block font-semibold text-slate-700 mb-1">Descripción Corta</label>
            <input type="text" id="course-short-desc" value="${esc(course?.shortDescription || '')}" placeholder="Resumen conciso en 1-2 líneas" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
          </div>

          <div>
            <label class="block font-semibold text-slate-700 mb-1">Descripción Completa</label>
            <textarea id="course-desc" rows="3" placeholder="Detalles de la capacitación..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">${esc(course?.description || '')}</textarea>
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

          const url = isEdit ? `/api/courses/${esc(course.id)}` : '/api/courses';
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
            <input type="text" id="mod-title" required value="${esc(mod?.title || '')}" placeholder="Ej. Módulo 1: Fundamentos y Conceptos Básicos" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
          </div>
          <div>
            <label class="block font-semibold text-slate-700 mb-1">Descripción Breve</label>
            <textarea id="mod-desc" rows="2" placeholder="Resumen del contenido del módulo..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">${esc(mod?.description || '')}</textarea>
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
          const url = isEdit ? `/api/courses/modules/${esc(mod.id)}` : `/api/courses/${esc(courseId)}/modules`;
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
            <input type="text" id="lesson-title" required value="${esc(lesson?.title || '')}" placeholder="Ej. Introducción y Arquitectura General" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
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
                <option value="quiz" ${lesson?.contentType === 'quiz' ? 'selected' : ''}>📝 Cuestionario / Evaluación IA</option>
              </select>
            </div>
            <div>
              <label class="block font-semibold text-slate-700 mb-1">Duración (Minutos)</label>
              <input type="number" id="lesson-duration" value="${lesson?.durationMinutes || 15}" class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
            </div>
          </div>

          <!-- URL / Upload File Container -->
          <div id="lesson-url-box" class="${lesson?.contentType === 'quiz' ? 'hidden' : ''}">
            <label class="block font-semibold text-slate-700 mb-1">URL o Archivo Adjunto</label>
            <div class="flex gap-2">
              <input type="text" id="lesson-url" value="${safeUrl(lesson?.contentUrl || '')}" placeholder="https://www.youtube.com/watch?v=... o archivo subido" class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-xs">
              <label id="lesson-upload-btn-label" class="px-3 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs flex items-center gap-1 cursor-pointer transition-colors shrink-0">
                <i data-lucide="upload" class="w-3.5 h-3.5"></i>
                <span>Subir</span>
                <input type="file" id="lesson-file-input" class="hidden">
              </label>
            </div>
          </div>

          <!-- Quiz AI Generator Box -->
          <div id="lesson-quiz-box" class="p-3 bg-purple-50 rounded-2xl border border-purple-200 space-y-2 ${lesson?.contentType === 'quiz' ? '' : 'hidden'}">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-purple-900 flex items-center gap-1.5">
                <i data-lucide="sparkles" class="w-4 h-4 text-purple-600"></i> Generación Automática con Gemini
              </span>
              ${isEdit ? `
                <button type="button" id="btn-autogen-lesson-quiz" class="px-3 py-1.5 rounded-xl bg-purple-600 hover:bg-purple-700 text-white font-bold text-[11px] flex items-center gap-1 cursor-pointer shadow-xs transition-colors">
                  <i data-lucide="bot" class="w-3.5 h-3.5"></i>
                  <span>Generar Preguntas con IA</span>
                </button>
              ` : `
                <span class="text-[10px] text-purple-600 bg-white px-2 py-0.5 rounded-md font-semibold border border-purple-200">
                  Guarda la lección para activar el generador IA
                </span>
              `}
            </div>
            <p class="text-[11px] text-purple-700 leading-tight">
              Gemini formulará 5 preguntas pedagógicas de opción múltiple con autocalificación basadas en el título, descripción y notas de esta lección.
            </p>
          </div>

          <!-- Rich Text Body for text types -->
          <div id="lesson-body-box">
            <label class="block font-semibold text-slate-700 mb-1">Contenido de Lectura (Markdown / Texto)</label>
            <textarea id="lesson-body" rows="4" placeholder="Escribe aquí las instrucciones, apuntes o guías de estudio..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">${esc(lesson?.contentBody || '')}</textarea>
          </div>

          <div>
            <label class="block font-semibold text-slate-700 mb-1">Descripción Breve de la Lección</label>
            <input type="text" id="lesson-desc" value="${esc(lesson?.description || '')}" placeholder="Orientaciones para el estudiante..." class="w-full rounded-xl border border-slate-200 px-3 py-2 text-xs">
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
            quizId: lesson?.quizId || null,
          };

          const url = isEdit ? `/api/courses/lessons/${esc(lesson.id)}` : `/api/courses/modules/${esc(moduleId)}/lessons`;
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

    // Event listeners inside Lesson Modal
    setTimeout(() => {
      // Toggle between URL/File box and Quiz box
      document.getElementById('lesson-type')?.addEventListener('change', (e) => {
        const isQuiz = e.target.value === 'quiz';
        const quizBox = document.getElementById('lesson-quiz-box');
        const urlBox = document.getElementById('lesson-url-box');
        if (quizBox) quizBox.classList.toggle('hidden', !isQuiz);
        if (urlBox) urlBox.classList.toggle('hidden', isQuiz);
      });

      // Autogen quiz with Gemini button
      document.getElementById('btn-autogen-lesson-quiz')?.addEventListener('click', async () => {
        if (!lesson || !lesson.id) {
          Toast.warning('Debes guardar la lección primero para asociar las preguntas pedagógicas.');
          return;
        }

        let apiKey = localStorage.getItem('dxstech_gemini_api_key');
        if (!apiKey) {
          Modal.promptGeminiKey({
            onSaved: (key) => this.generateQuizForLesson(lesson.id, key)
          });
          return;
        }
        await this.generateQuizForLesson(lesson.id, apiKey);
      });

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
  },
};
