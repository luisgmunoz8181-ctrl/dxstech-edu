import { Toast } from '../components/toast.js';
import { Loading } from '../components/loading.js';

export const LoginView = {
  mode: 'login', // 'login' | 'forgot' | 'reset'
  resetToken: '',

  render() {
    return `
      <div class="min-h-[80vh] flex items-center justify-center py-8 px-4">
        <div class="max-w-md w-full space-y-6">
          <!-- Logo & Header -->
          <div class="text-center space-y-2">
            <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-indigo-500 to-indigo-700 flex items-center justify-center text-white shadow-xl shadow-indigo-100 font-extrabold text-2xl mx-auto">
              D
            </div>
            <h2 class="text-2xl font-extrabold text-slate-900 tracking-tight">DxSTech <span class="text-indigo-600">Edu</span></h2>
            <p class="text-xs text-slate-500">Plataforma de Formación Virtual & LMS Corporativo</p>
          </div>

          <!-- Main Card -->
          <div class="bg-white rounded-3xl p-7 border border-slate-200/80 shadow-xl shadow-slate-100/80 transition-all">
            ${this.renderCardContent()}
          </div>

          <!-- Fast Demo Login Buttons -->
          <div class="bg-slate-50 rounded-2xl p-4 border border-slate-200/60 text-center space-y-3">
            <div class="flex items-center justify-center gap-1.5 text-xs font-semibold text-slate-600">
              <i data-lucide="key-round" class="w-3.5 h-3.5 text-indigo-500"></i>
              <span>Acceso Rápido / Cuentas de Prueba</span>
            </div>
            <p class="text-[11px] text-slate-400">Haz clic en cualquier rol para autocompletar credenciales:</p>
            <div class="grid grid-cols-3 gap-2">
              <button type="button" data-fill-role="superadmin" class="px-2 py-2 rounded-xl text-[11px] font-bold bg-white hover:bg-indigo-50 hover:text-indigo-700 hover:border-indigo-300 transition-all border border-slate-200 shadow-xs flex flex-col items-center gap-1 text-slate-700">
                <span class="w-2 h-2 rounded-full bg-purple-500"></span>
                <span>Superadmin</span>
              </button>
              <button type="button" data-fill-role="admin" class="px-2 py-2 rounded-xl text-[11px] font-bold bg-white hover:bg-indigo-50 hover:text-indigo-700 hover:border-indigo-300 transition-all border border-slate-200 shadow-xs flex flex-col items-center gap-1 text-slate-700">
                <span class="w-2 h-2 rounded-full bg-indigo-500"></span>
                <span>Admin</span>
              </button>
              <button type="button" data-fill-role="student" class="px-2 py-2 rounded-xl text-[11px] font-bold bg-white hover:bg-indigo-50 hover:text-indigo-700 hover:border-indigo-300 transition-all border border-slate-200 shadow-xs flex flex-col items-center gap-1 text-slate-700">
                <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
                <span>Estudiante</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    `;
  },

  renderCardContent() {
    if (this.mode === 'forgot') {
      return `
        <div class="space-y-5">
          <div>
            <h3 class="text-base font-bold text-slate-900">Recuperar Contraseña</h3>
            <p class="text-xs text-slate-500 mt-1">Ingresa el correo asociado a tu cuenta para restablecer el acceso.</p>
          </div>

          <form id="forgot-form" class="space-y-4">
            <div>
              <label class="block text-xs font-semibold text-slate-700 mb-1">Correo Electrónico</label>
              <input type="email" id="forgot-email" required placeholder="tu-correo@ejemplo.com" class="w-full text-xs rounded-xl border border-slate-200 px-3.5 py-2.5 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden transition-all">
            </div>

            <div id="dev-token-alert" class="hidden p-3 rounded-xl bg-amber-50 border border-amber-200 text-xs text-amber-800 space-y-1">
              <p class="font-bold">Token de prueba generado:</p>
              <p id="dev-token-val" class="font-mono text-[11px] break-all"></p>
              <button type="button" id="use-dev-token-btn" class="mt-1 text-[11px] font-bold text-indigo-700 underline">Continuar a cambio de contraseña</button>
            </div>

            <button type="submit" class="w-full py-2.5 px-4 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 transition-all shadow-md shadow-indigo-100 flex items-center justify-center gap-2">
              <i data-lucide="send" class="w-3.5 h-3.5"></i>
              <span>Solicitar Restablecimiento</span>
            </button>

            <button type="button" id="back-to-login-btn" class="w-full text-center text-xs text-slate-500 hover:text-indigo-600 transition-colors font-medium">
              ← Regresar al inicio de sesión
            </button>
          </form>
        </div>
      `;
    }

    if (this.mode === 'reset') {
      return `
        <div class="space-y-5">
          <div>
            <h3 class="text-base font-bold text-slate-900">Definir Nueva Contraseña</h3>
            <p class="text-xs text-slate-500 mt-1">Ingresa el token de seguridad y tu nueva contraseña.</p>
          </div>

          <form id="reset-form" class="space-y-4">
            <div>
              <label class="block text-xs font-semibold text-slate-700 mb-1">Token de Seguridad</label>
              <input type="text" id="reset-token" required value="${this.resetToken}" placeholder="Pega el token aquí" class="w-full text-xs rounded-xl border border-slate-200 px-3.5 py-2.5 font-mono focus:ring-2 focus:ring-indigo-500 focus:outline-hidden transition-all">
            </div>

            <div>
              <label class="block text-xs font-semibold text-slate-700 mb-1">Nueva Contraseña</label>
              <input type="password" id="reset-new-password" required minlength="8" placeholder="Mínimo 8 caracteres (Mayús, minús, número)" class="w-full text-xs rounded-xl border border-slate-200 px-3.5 py-2.5 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden transition-all">
            </div>

            <button type="submit" class="w-full py-2.5 px-4 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-700 transition-all shadow-md shadow-emerald-100 flex items-center justify-center gap-2">
              <i data-lucide="check" class="w-3.5 h-3.5"></i>
              <span>Restablecer Contraseña</span>
            </button>

            <button type="button" id="back-to-login-btn" class="w-full text-center text-xs text-slate-500 hover:text-indigo-600 transition-colors font-medium">
              ← Cancelar y volver al login
            </button>
          </form>
        </div>
      `;
    }

    // Default: Login Form
    return `
      <form id="login-form" class="space-y-4">
        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">Correo Electrónico</label>
          <div class="relative">
            <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
              <i data-lucide="mail" class="w-4 h-4"></i>
            </span>
            <input type="email" id="login-email" required placeholder="nombre@dxstech.edu" class="w-full text-xs rounded-xl border border-slate-200 pl-9 pr-3.5 py-2.5 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden transition-all">
          </div>
        </div>

        <div>
          <div class="flex items-center justify-between mb-1">
            <label class="text-xs font-semibold text-slate-700">Contraseña</label>
            <button type="button" id="forgot-link-btn" class="text-[11px] text-indigo-600 hover:text-indigo-700 font-semibold transition-colors">
              ¿Olvidaste tu contraseña?
            </button>
          </div>
          <div class="relative">
            <span class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
              <i data-lucide="lock" class="w-4 h-4"></i>
            </span>
            <input type="password" id="login-password" required placeholder="••••••••" class="w-full text-xs rounded-xl border border-slate-200 pl-9 pr-10 py-2.5 focus:ring-2 focus:ring-indigo-500 focus:outline-hidden transition-all">
            <button type="button" id="toggle-password-btn" class="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-600">
              <i data-lucide="eye" class="w-4 h-4"></i>
            </button>
          </div>
        </div>

        <div class="flex items-center">
          <input type="checkbox" id="login-remember" class="w-4 h-4 text-indigo-600 border-slate-300 rounded focus:ring-indigo-500 cursor-pointer">
          <label for="login-remember" class="ml-2 text-xs text-slate-600 cursor-pointer">
            Recordar mi sesión en este dispositivo
          </label>
        </div>

        <button type="submit" id="login-submit-btn" class="w-full py-2.5 px-4 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 active:scale-[0.99] transition-all shadow-md shadow-indigo-100 flex items-center justify-center gap-2">
          <i data-lucide="log-in" class="w-4 h-4"></i>
          <span>Iniciar Sesión</span>
        </button>
      </form>
    `;
  },

  mount() {
    this.bindEvents();
    if (window.lucide) window.lucide.createIcons();
  },

  bindEvents() {
    // Quick Demo Buttons
    document.querySelectorAll('[data-fill-role]').forEach(btn => {
      btn.addEventListener('click', () => {
        const role = btn.dataset.fillRole;
        this.fillDemoCredentials(role);
      });
    });

    // Toggle Password Visibility
    const toggleBtn = document.getElementById('toggle-password-btn');
    if (toggleBtn) {
      toggleBtn.addEventListener('click', () => {
        const passInput = document.getElementById('login-password');
        if (passInput) {
          const isPass = passInput.type === 'password';
          passInput.type = isPass ? 'text' : 'password';
          toggleBtn.innerHTML = isPass ? '<i data-lucide="eye-off" class="w-4 h-4"></i>' : '<i data-lucide="eye" class="w-4 h-4"></i>';
          if (window.lucide) window.lucide.createIcons();
        }
      });
    }

    // Toggle to Forgot Password
    const forgotBtn = document.getElementById('forgot-link-btn');
    if (forgotBtn) {
      forgotBtn.addEventListener('click', () => {
        this.mode = 'forgot';
        window.router?.navigate('login');
      });
    }

    // Back to Login
    const backBtn = document.getElementById('back-to-login-btn');
    if (backBtn) {
      backBtn.addEventListener('click', () => {
        this.mode = 'login';
        window.router?.navigate('login');
      });
    }

    // Login Form Submit
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
      loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('login-email').value.trim();
        const password = document.getElementById('login-password').value;
        const rememberMe = document.getElementById('login-remember').checked;

        const submitBtn = document.getElementById('login-submit-btn');
        submitBtn.disabled = true;
        submitBtn.innerHTML = `<span>Validando...</span>`;

        try {
          const res = await fetch('/api/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password, rememberMe }),
          });

          const data = await res.json();
          if (!res.ok) {
            throw new Error(data.error || 'Credenciales incorrectas');
          }

          Toast.success(`¡Bienvenido de nuevo, ${data.user.firstName}!`);
          
          // Trigger global session update
          if (window.router) {
            window.router.currentUser = data.user;
            window.router.updateSessionUI();
            
            // Navigate to courses
            window.location.hash = 'courses';
            window.router.navigate('courses');
          }
        } catch (err) {
          Toast.error(err.message);
          submitBtn.disabled = false;
          submitBtn.innerHTML = `<i data-lucide="log-in" class="w-4 h-4"></i><span>Iniciar Sesión</span>`;
          if (window.lucide) window.lucide.createIcons();
        }
      });
    }

    // Forgot Password Form Submit
    const forgotForm = document.getElementById('forgot-form');
    if (forgotForm) {
      forgotForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('forgot-email').value.trim();

        try {
          const res = await fetch('/api/auth/forgot-password', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email }),
          });

          const data = await res.json();
          Toast.info(data.message);

          if (data.devToken) {
            const alertBox = document.getElementById('dev-token-alert');
            const tokenVal = document.getElementById('dev-token-val');
            const useBtn = document.getElementById('use-dev-token-btn');
            if (alertBox && tokenVal) {
              alertBox.classList.remove('hidden');
              tokenVal.textContent = data.devToken;
              this.resetToken = data.devToken;
              useBtn.onclick = () => {
                this.mode = 'reset';
                window.router?.navigate('login');
              };
            }
          }
        } catch (err) {
          Toast.error(err.message);
        }
      });
    }

    // Reset Password Form Submit
    const resetForm = document.getElementById('reset-form');
    if (resetForm) {
      resetForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const token = document.getElementById('reset-token').value.trim();
        const newPassword = document.getElementById('reset-new-password').value;

        try {
          const res = await fetch('/api/auth/reset-password', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ token, newPassword }),
          });

          const data = await res.json();
          if (!res.ok) throw new Error(data.error);

          Toast.success(data.message);
          this.mode = 'login';
          window.router?.navigate('login');
        } catch (err) {
          Toast.error(err.message);
        }
      });
    }
  },

  fillDemoCredentials(role) {
    if (this.mode !== 'login') {
      this.mode = 'login';
      if (window.router) window.router.navigate('login');
    }

    setTimeout(() => {
      const emailInput = document.getElementById('login-email');
      const passInput = document.getElementById('login-password');
      if (!emailInput || !passInput) return;

      switch (role) {
        case 'superadmin':
          emailInput.value = 'superadmin@dxstech.edu';
          passInput.value = 'Admin1234*';
          break;
        case 'admin':
          emailInput.value = 'admin@dxstech.edu';
          passInput.value = 'Admin1234*';
          break;
        case 'student':
          emailInput.value = 'estudiante@dxstech.edu';
          passInput.value = 'Student1234*';
          break;
      }
      Toast.info(`Credenciales de ${role.toUpperCase()} cargadas.`);
    }, 50);
  }
};
