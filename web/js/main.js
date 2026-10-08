import { LoginView } from './views/login.js';
import { ProfileView } from './views/profile.js';
import { UsersView } from './views/users.js';
import { DashboardView } from './views/dashboard.js';
import { CoursesView } from './views/courses.js';
import { CertificatesView } from './views/certificates.js';
import { QuizzesView } from './views/quizzes.js';
import { WhatsAppView } from './views/whatsapp.js';
import { SettingsView } from './views/settings.js';
import { Modal } from './components/modal.js';
import { Toast } from './components/toast.js';

class AppRouter {
  constructor() {
    this.views = {
      login: LoginView,
      profile: ProfileView,
      dashboard: DashboardView,
      users: UsersView,
      courses: CoursesView,
      certificates: CertificatesView,
      quizzes: QuizzesView,
      whatsapp: WhatsAppView,
      settings: SettingsView,
    };

    this.viewTitles = {
      login: { title: 'Iniciar Sesión', subtitle: 'Acceso a la plataforma LMS DxSTech Edu' },
      profile: { title: 'Mi Perfil & Seguridad', subtitle: 'Datos de la cuenta y actualización de contraseña' },
      dashboard: { title: 'Dashboard Administrativo', subtitle: 'Métricas institucionales, avance de cohortes y reportes LMS' },
      users: { title: 'Gestión de Usuarios', subtitle: 'Administración de roles RBAC y accesos institucionales' },
      courses: { title: 'Cursos & Contenidos', subtitle: 'Catálogo educativo, módulos interactivos y aula virtual' },
      certificates: { title: 'Certificados', subtitle: 'Generación masiva y diseño interactivo en alta fidelidad' },
      quizzes: { title: 'Evaluaciones IA', subtitle: 'Generador inteligente con Gemini y simulador de exámenes' },
      whatsapp: { title: 'WhatsApp + Chatbot IA', subtitle: 'Gateway automatizado con base de conocimiento estricta' },
      settings: { title: 'Configuración & Seguridad', subtitle: 'Gestión local de tu Gemini API Key (BYOK)' },
    };

    this.currentViewId = null;
    this.currentViewInstance = null;
    this.currentUser = null;
  }

  async init() {
    window.router = this;
    this.bindNavigation();
    this.bindKeyStatusIndicator();
    this.updateKeyBadge();

    // Check current session from HttpOnly cookie
    await this.checkSession();

    // Default to Courses catalog
    const initialHash = window.location.hash.replace('#', '') || 'courses';
    if (initialHash.startsWith('verify')) {
      const id = initialHash.replace('verify/', '').replace('verify', '').trim();
      this.renderVerificationScreen(id);
    } else {
      this.navigate(this.views[initialHash] ? initialHash : 'courses');
    }

    window.addEventListener('hashchange', () => {
      const hash = window.location.hash.replace('#', '');
      if (hash.startsWith('verify')) {
        const id = hash.replace('verify/', '').replace('verify', '').trim();
        this.renderVerificationScreen(id);
        return;
      }
      if (this.views[hash] && hash !== this.currentViewId) {
        this.navigate(hash);
      }
    });

    window.addEventListener('gemini-key-changed', () => {
      this.updateKeyBadge();
    });
  }

  bindNavigation() {
    // Desktop & Mobile buttons
    document.querySelectorAll('[data-nav]').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.preventDefault();
        const target = btn.dataset.nav;
        if (target && target !== this.currentViewId) {
          window.location.hash = target;
          this.navigate(target);
        }
      });
    });

    // Header & Sidebar Key buttons
    const headerKeyBtn = document.getElementById('header-key-btn');
    if (headerKeyBtn) {
      headerKeyBtn.addEventListener('click', () => {
        const hasKey = !!localStorage.getItem('dxstech_gemini_api_key');
        if (!hasKey) {
          Modal.promptGeminiKey({
            onSaved: () => {
              this.updateKeyBadge();
            }
          });
        } else {
          this.navigate('settings');
        }
      });
    }

    const sidebarKeyStatus = document.getElementById('sidebar-key-status');
    if (sidebarKeyStatus) {
      sidebarKeyStatus.addEventListener('click', () => {
        this.navigate('settings');
      });
    }
  }

  async checkSession() {
    try {
      const res = await fetch('/api/auth/me');
      if (res.ok) {
        const data = await res.json();
        if (data && data.authenticated && data.user) {
          this.currentUser = data.user;
        } else if (data && data.id) {
          this.currentUser = data;
        } else {
          this.currentUser = null;
        }
      } else {
        this.currentUser = null;
      }
    } catch {
      this.currentUser = null;
    }
    this.updateSessionUI();
  }

  updateSessionUI() {
    const userBtn = document.getElementById('header-user-btn');
    const usersNavBtn = document.querySelector('[data-nav="users"]');
    const mobileUsersBtn = document.querySelector('.mobile-nav-btn[data-nav="users"]');

    if (this.currentUser) {
      const initials = `${(this.currentUser.firstName || 'U').charAt(0)}${(this.currentUser.lastName || '').charAt(0)}`.toUpperCase();
      const fullName = `${this.currentUser.firstName} ${this.currentUser.lastName}`;
      const roleBadge = this.currentUser.role;

      if (userBtn) {
        userBtn.className = 'flex items-center gap-2 text-xs font-semibold px-3 py-1 rounded-full border border-slate-200 bg-white hover:bg-slate-50 transition-all text-slate-700 shadow-2xs cursor-pointer';
        userBtn.innerHTML = `
          <span class="w-6 h-6 rounded-full bg-gradient-to-tr from-indigo-600 to-indigo-800 text-white flex items-center justify-center text-[10px] font-bold tracking-wider">
            ${initials}
          </span>
          <div class="flex flex-col text-left leading-tight">
            <span class="text-xs font-bold text-slate-800">${fullName}</span>
            <span class="text-[9px] font-semibold text-indigo-600 uppercase tracking-wider">${roleBadge}</span>
          </div>
          <i data-lucide="chevron-down" class="w-3.5 h-3.5 text-slate-400"></i>
        `;
        userBtn.onclick = (e) => {
          e.preventDefault();
          this.navigate('profile');
        };
      }

      // RBAC nav button visibility
      const isAdmin = ['SUPERADMIN', 'ADMINISTRADOR'].includes(this.currentUser.role);
      const dashNavBtn = document.querySelector('[data-nav="dashboard"]');
      const mobileDashBtn = document.querySelector('.mobile-nav-btn[data-nav="dashboard"]');
      if (dashNavBtn) dashNavBtn.style.display = isAdmin ? 'flex' : 'none';
      if (mobileDashBtn) mobileDashBtn.style.display = isAdmin ? 'flex' : 'none';
      if (usersNavBtn) usersNavBtn.style.display = isAdmin ? 'flex' : 'none';
      if (mobileUsersBtn) mobileUsersBtn.style.display = isAdmin ? 'flex' : 'none';
    } else {
      if (userBtn) {
        userBtn.className = 'flex items-center gap-2 text-xs font-semibold px-3 py-1.5 rounded-full border border-slate-200 bg-white hover:bg-slate-50 transition-all text-slate-700 shadow-2xs cursor-pointer';
        userBtn.innerHTML = `
          <span class="w-6 h-6 rounded-full bg-slate-100 text-slate-600 flex items-center justify-center text-[10px] font-bold">
            <i data-lucide="log-in" class="w-3.5 h-3.5 text-slate-600"></i>
          </span>
          <span id="header-user-name" class="text-xs font-semibold text-slate-700">Iniciar Sesión</span>
        `;
        userBtn.onclick = (e) => {
          e.preventDefault();
          this.navigate('login');
        };
      }

      const dashNavBtn = document.querySelector('[data-nav="dashboard"]');
      const mobileDashBtn = document.querySelector('.mobile-nav-btn[data-nav="dashboard"]');
      if (dashNavBtn) dashNavBtn.style.display = 'none';
      if (mobileDashBtn) mobileDashBtn.style.display = 'none';
      if (usersNavBtn) usersNavBtn.style.display = 'none';
      if (mobileUsersBtn) mobileUsersBtn.style.display = 'none';
    }

    if (window.lucide) {
      window.lucide.createIcons();
    }
  }

  navigate(viewId) {
    if (!this.views[viewId]) return;

    // RBAC Protection guards
    if (viewId === 'dashboard' || viewId === 'users') {
      if (!this.currentUser) {
        Toast.info('Inicia sesión con credenciales de administrador para acceder a este módulo.');
        this.navigate('login');
        return;
      }
      if (!['SUPERADMIN', 'ADMINISTRADOR'].includes(this.currentUser.role)) {
        Toast.error('Acceso denegado: tu rol de Estudiante no tiene privilegios administrativos.');
        this.navigate('courses');
        return;
      }
    }

    if (viewId === 'profile') {
      if (!this.currentUser) {
        Toast.info('Inicia sesión para acceder a tu perfil.');
        this.navigate('login');
        return;
      }
    }

    // Teardown previous view to clean intervals & listeners
    if (this.currentViewInstance && typeof this.currentViewInstance.destroy === 'function') {
      this.currentViewInstance.destroy();
    }

    this.currentViewId = viewId;
    this.currentViewInstance = this.views[viewId];

    // Update Titles
    const meta = this.viewTitles[viewId] || { title: viewId, subtitle: '' };
    document.getElementById('view-title').textContent = meta.title;
    document.getElementById('view-subtitle').textContent = meta.subtitle;

    // Update Navigation UI active states
    this.updateNavActiveState(viewId);

    // Render View Content
    const mainContainer = document.getElementById('app-content');
    mainContainer.innerHTML = this.currentViewInstance.render();

    // Re-initialize Lucide icons in freshly rendered DOM
    if (window.lucide) {
      window.lucide.createIcons();
    }

    // Mount View
    if (typeof this.currentViewInstance.mount === 'function') {
      this.currentViewInstance.mount();
    }
  }

  updateNavActiveState(activeId) {
    // Desktop Nav
    document.querySelectorAll('.nav-btn').forEach(btn => {
      const isTarget = btn.dataset.nav === activeId;
      if (isTarget) {
        btn.className = 'nav-btn w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl font-bold text-sm transition-all bg-indigo-50 text-indigo-700 shadow-xs';
        const icon = btn.querySelector('i');
        if (icon) icon.className = 'w-5 h-5 text-indigo-600 transition-colors';
      } else {
        btn.className = 'nav-btn w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl font-medium text-sm transition-all text-slate-600 hover:text-slate-900 hover:bg-slate-100';
        const icon = btn.querySelector('i');
        if (icon) icon.className = 'w-5 h-5 text-slate-400 transition-colors';
      }
    });

    // Mobile Nav
    document.querySelectorAll('.mobile-nav-btn').forEach(btn => {
      const isTarget = btn.dataset.nav === activeId;
      if (isTarget) {
        btn.className = 'mobile-nav-btn flex flex-col items-center gap-1 py-1 px-2.5 text-indigo-600 font-bold';
      } else {
        btn.className = 'mobile-nav-btn flex flex-col items-center gap-1 py-1 px-2.5 text-slate-400 font-medium';
      }
    });
  }

  bindKeyStatusIndicator() {
    this.updateKeyBadge();
  }

  updateKeyBadge() {
    const key = localStorage.getItem('dxstech_gemini_api_key');
    const hasKey = !!(key && key.trim().length > 0);

    // Sidebar indicator
    const keyBadge = document.getElementById('key-badge');
    const keyText = document.getElementById('key-status-text');
    if (keyBadge && keyText) {
      if (hasKey) {
        keyBadge.className = 'w-2 h-2 rounded-full bg-emerald-500 ring-4 ring-emerald-100';
        keyText.textContent = 'Activa en navegador';
      } else {
        keyBadge.className = 'w-2 h-2 rounded-full bg-amber-400 ring-4 ring-amber-100';
        keyText.textContent = 'No configurada';
      }
    }

    // Top Header indicator
    const headerDot = document.getElementById('header-key-dot');
    const headerLabel = document.getElementById('header-key-label');
    if (headerDot && headerLabel) {
      if (hasKey) {
        headerDot.className = 'w-2 h-2 rounded-full bg-emerald-500';
        headerLabel.textContent = 'Gemini: Conectado';
      } else {
        headerDot.className = 'w-2 h-2 rounded-full bg-amber-400';
        headerLabel.textContent = 'Gemini: Pendiente';
      }
    }
  }

  async renderVerificationScreen(certId) {
    if (this.currentViewInstance && typeof this.currentViewInstance.destroy === 'function') {
      this.currentViewInstance.destroy();
    }
    this.currentViewId = 'verify';
    this.currentViewInstance = null;

    document.getElementById('view-title').textContent = 'Verificación de Certificado';
    document.getElementById('view-subtitle').textContent = 'Consulta de autenticidad en el registro público de DxSTech Edu';

    // Deselect sidebar buttons
    document.querySelectorAll('.nav-btn, .mobile-nav-btn').forEach(btn => {
      btn.classList.remove('bg-indigo-50', 'text-indigo-700', 'font-bold');
    });

    const mainContainer = document.getElementById('app-content');
    mainContainer.innerHTML = `
      <div class="max-w-xl mx-auto py-8">
        <div class="bg-white rounded-3xl p-8 border border-slate-200/80 shadow-xl text-center space-y-6">
          <div class="w-16 h-16 rounded-2xl bg-indigo-50 text-indigo-600 flex items-center justify-center mx-auto">
            <div class="w-8 h-8 border-3 border-indigo-600/30 border-t-indigo-600 rounded-full animate-spin"></div>
          </div>
          <p class="text-xs font-semibold text-slate-500">Consultando registro de verificación oficial...</p>
        </div>
      </div>
    `;

    try {
      const res = await fetch(`/api/certificates/verify/${certId}`);
      const data = await res.json();

      if (res.ok && data.valid) {
        const c = data.certificate;
        mainContainer.innerHTML = `
          <div class="max-w-xl mx-auto py-6">
            <div class="bg-white rounded-3xl p-8 border border-emerald-200 shadow-xl text-center space-y-6">
              <div class="w-20 h-20 rounded-3xl bg-emerald-100 text-emerald-600 flex items-center justify-center mx-auto shadow-inner">
                <i data-lucide="shield-check" class="w-10 h-10"></i>
              </div>

              <div>
                <span class="inline-block px-3 py-1 rounded-full text-xs font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 mb-2">
                  ✓ Certificado Oficial Auténtico
                </span>
                <h3 class="text-xl font-black text-slate-900 leading-tight">${c.studentName}</h3>
                <p class="text-xs text-slate-500 mt-1">ha acreditado satisfactoriamente los requisitos académicos de:</p>
                <p class="text-sm font-bold text-indigo-700 mt-1">${c.courseTitle}</p>
              </div>

              <div class="bg-slate-50 rounded-2xl p-4 border border-slate-200/80 text-left space-y-2 text-xs">
                <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                  <span class="text-slate-400 font-medium">Código de Registro:</span>
                  <span class="font-mono font-bold text-slate-800 bg-white px-2 py-0.5 rounded border border-slate-200">${c.id}</span>
                </div>
                <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                  <span class="text-slate-400 font-medium">Fecha de Emisión:</span>
                  <span class="font-semibold text-slate-700">${c.issueDate}</span>
                </div>
                ${c.durationHours ? `
                  <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                    <span class="text-slate-400 font-medium">Intensidad Académica:</span>
                    <span class="font-semibold text-slate-700">${c.durationHours} Horas</span>
                  </div>
                ` : ''}
                ${c.instructorName ? `
                  <div class="flex justify-between items-center py-1 border-b border-slate-200/60">
                    <span class="text-slate-400 font-medium">Docente / Director:</span>
                    <span class="font-semibold text-slate-700">${c.instructorName}</span>
                  </div>
                ` : ''}
                <div class="flex justify-between items-center py-1">
                  <span class="text-slate-400 font-medium">Institución Emisora:</span>
                  <span class="font-semibold text-indigo-600">DxSTech Edu — Academy of Technology & AI</span>
                </div>
              </div>

              <div class="pt-2 flex flex-wrap items-center justify-center gap-2.5">
                <a href="/api/certificates/${c.id}/pdf" target="_blank" download="Certificado_${c.id}.pdf" class="px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-colors shadow-md shadow-indigo-100 flex items-center gap-2">
                  <i data-lucide="download" class="w-4 h-4"></i>
                  <span>Descargar Diploma Oficial (PDF)</span>
                </a>
                <button id="copy-verify-url-btn" class="px-4 py-2.5 rounded-xl text-xs font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center gap-1.5 shadow-xs">
                  <i data-lucide="copy" class="w-3.5 h-3.5 text-slate-500"></i>
                  <span>Copiar Enlace</span>
                </button>
                <a href="#courses" class="px-4 py-2.5 rounded-xl text-xs font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors flex items-center gap-1.5 shadow-xs">
                  <i data-lucide="book-open" class="w-3.5 h-3.5 text-slate-500"></i>
                  <span>Explorar Cursos</span>
                </a>
              </div>
            </div>
          </div>
        `;

        document.getElementById('copy-verify-url-btn')?.addEventListener('click', () => {
          navigator.clipboard.writeText(window.location.href).then(() => {
            Toast.success('Enlace oficial de verificación copiado al portapapeles');
          }).catch(() => {
            Toast.info(window.location.href);
          });
        });

        document.getElementById('print-verify-btn')?.addEventListener('click', () => {
          window.print();
        });
      } else {
        mainContainer.innerHTML = `
          <div class="max-w-xl mx-auto py-6">
            <div class="bg-white rounded-3xl p-8 border border-rose-200 shadow-xl text-center space-y-6">
              <div class="w-20 h-20 rounded-3xl bg-rose-100 text-rose-600 flex items-center justify-center mx-auto">
                <i data-lucide="alert-octagon" class="w-10 h-10"></i>
              </div>

              <div>
                <span class="inline-block px-3 py-1 rounded-full text-xs font-bold bg-rose-50 text-rose-700 border border-rose-200 mb-2">
                  Certificado No Encontrado
                </span>
                <h3 class="text-lg font-bold text-slate-900">No se pudo verificar el documento</h3>
                <p class="text-xs text-slate-500 mt-2 max-w-sm mx-auto">
                  El código <code class="font-mono font-bold text-slate-700 bg-slate-100 px-1 py-0.5 rounded">${certId || 'N/A'}</code> no figura en la base de datos oficial de certificaciones emitidas por DxSTech Edu.
                </p>
              </div>

              <div class="pt-2">
                <a href="#certificates" class="px-5 py-2.5 rounded-xl text-xs font-bold text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors inline-flex items-center gap-2">
                  <span>Regresar al inicio</span>
                </a>
              </div>
            </div>
          </div>
        `;
      }
    } catch (e) {
      mainContainer.innerHTML = `<div class="p-6 text-xs text-rose-600 bg-rose-50 rounded-2xl max-w-md mx-auto text-center">Error de red verificando certificado: ${e.message}</div>`;
    }

    if (window.lucide) window.lucide.createIcons();
  }
}

// Initialize App
function startApp() {
  const app = new AppRouter();
  app.init();
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', startApp);
} else {
  startApp();
}
