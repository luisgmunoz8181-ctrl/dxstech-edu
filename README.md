# DxSTech Edu — Plataforma Educativa Premium

DxSTech Edu es una suite educativa integral de alto rendimiento construida con **Go 1.25+ (Gin)** y una **SPA en Vanilla JavaScript ES Modules**, impulsada por Tailwind CSS, Lucide Icons, SQLite (100% CGO-free con `modernc.org/sqlite`) y la API de Google Gemini bajo el modelo de privacidad BYOK (*Bring Your Own Key*).

---

## 🚀 Inicio Rápido (Desarrollo Local)

Para instalar dependencias y levantar el servidor de desarrollo:

```bash
pnpm install
pnpm dev   # ejecuta el servidor con --dev (APP_ENV=development)
```

> **Modo desarrollo vs. producción:** si no se define `APP_ENV`, el servidor asume **producción** y exige `JWT_SECRET`.
> Solo en desarrollo (`--dev`) se siembran usuarios y cursos de demostración (`superadmin@dxstech.edu` / `Admin1234*`,
> `estudiante@dxstech.edu` / `Student1234*`), se genera un secreto JWT efímero y el gateway de WhatsApp queda sin autenticación.
> **Nunca expongas una instancia `--dev` a internet.**

La aplicación estará lista y accesible en su navegador:

👉 **http://localhost:3000**

---

## 🌟 Características y Módulos Principales

### 1. Generador Profesional de Certificados con Código QR de Verificación
- **Optimización en el Navegador**: Carga mediante selector de archivos, arrastrar y soltar (*drag & drop*) o pegado directo con `Ctrl + V`.
- **Compresión WebP en Canvas**: Convierte instantáneamente las plantillas a formato WebP (calidad ~0.82) antes de cualquier envío, optimizando velocidad y memoria.
- **Editor Visual Interactivo**: Ajuste en vivo de coordenadas X/Y en porcentaje, guías visuales, tamaño de tipografía, selector de color, fuentes estándar y alineación.
- **Carga de Alumnos**:
  - Entrada manual por renglón.
  - Carga masiva de libros de Excel (`.xlsx`, `.xls`) vía **SheetJS**, reconociendo automáticamente columnas como `nombre`, `alumno` o `estudiante`.
- **Código QR de Verificación Digital**:
  - Incorpora un código QR de alta resolución (256px) con posición y tamaño ajustables en la plantilla.
  - Cada certificado emitido recibe un identificador criptográfico único con formato `DXS-YYYY-XXXX`.
  - El código QR enlaza directamente a la URL pública de validación (`/#verify/DXS-YYYY-XXXX`), adaptándose de forma automática al dominio y protocolo HTTPS (`X-Forwarded-Proto`).
  - Registro auditable y persistente en SQLite (`issued_certificates`) con búsqueda reactiva, verificación directa y exportación de la lista completa a CSV.
  - Pantalla pública e interactiva de validación con badge de autenticidad, botón para copiar enlace oficial e impresión de comprobante.
- **Compilación Vectorial PDF + Empaquetado ZIP**: El backend en Go genera los PDFs en milisegundos y los entrega listos para descarga en un archivo ZIP.

### 2. Evaluaciones Inteligentes con Gemini
- Generación automatizada de cuestionarios a partir de cualquier temario, transcripción o material de clase.
- Selector dinámico de 3 a 15 preguntas de opción múltiple.
- Validación estricta en Go: garantiza 4 opciones únicas por reactivo y una clave de respuesta correcta.
- **Simulador Interactivo**: Examen en vivo para estudiantes con control de sonido (activar/silenciar efectos sintetizados con Web Audio API), feedback visual, cálculo de nota (0-100) y aprobación (≥ 60%).
- **Formato Imprimible para Clases Presenciales**: Generación de hojas de examen con campos para alumno, fecha y calificación, listas para imprimir en papel (`window.print()`) o guardar como PDF.
- Búsqueda reactiva, almacenamiento histórico y eliminación permanente en SQLite.

### 3. WhatsApp Gateway & Chatbot IA
- **Doble Modo según Entorno**:
  - `APP_ENV=development`: Acceso libre e inmediato a todas las funciones sin requerir autenticación.
  - `APP_ENV=production`: Protección y control de acceso.
- **Importación Inteligente de Destinatarios**: Carga números de teléfono directamente desde archivos Excel (`.xlsx`, `.xls`) o CSV con detección automática de columnas (`telefono`, `celular`, `phone`, `whatsapp`).
- **Plantillas Rápidas de Notificación**: Botones de un clic para cargar mensajes de entrega de certificados, avisos de evaluaciones y mensajes de bienvenida.
- **Asistente Virtual con Base de Conocimiento**:
  - Switch de activación/pausa instantáneo.
  - Base de datos institucional editable en tiempo real.
  - **Modo Estricto**: Si la consulta no está en los lineamientos institucionales, el bot declina amablemente sin alucinar: *"Lo siento, solo puedo responder preguntas sobre la base de conocimiento proporcionada."*
- **Simulador de Chat en Vivo**: Interfaz tipo WhatsApp Web para probar preguntas y respuestas en tiempo real con indicador de "Escribiendo...".
- **Envíos Masivos con Rate Limit**: Despacho secuencial configurable con retardo controlado entre mensajes, bitácora de auditoría en vivo con exportación a CSV y vaciado seguro.

### 4. Seguridad y Privacidad: Bring Your Own Key (BYOK)
- La aplicación **nunca** almacena una clave API en el servidor ni en archivos del proyecto.
- La Gemini API Key del usuario se conserva exclusivamente en su navegador (`localStorage["dxstech_gemini_api_key"]`).
- Se transmite en memoria únicamente al ejecutar operaciones de IA (vía encabezado `X-Gemini-API-Key`) y nunca se registra en base de datos, logs ni variables de entorno.

---

## 🐳 Despliegue con Docker y Docker Compose

La aplicación cuenta con una compilación multi-etapa optimizada con **Go 1.24+ Alpine** y una imagen final ultraligera de **Alpine 3.21** sin dependencias externas ni compiladores C.

### Opción A: Levantar con Docker Compose (Recomendado)

```bash
docker compose up -d --build
```

El servicio estará disponible en `http://localhost:3000` con persistencia de base de datos en el volumen `dxstech_edu_data`.

Para ver los logs en tiempo real:
```bash
docker compose logs -f
```

Para detener el servicio:
```bash
docker compose down
```

### Opción B: Construcción manual con Docker CLI

```bash
# 1. Construir la imagen
docker build -t dxstech-edu .

# 2. Ejecutar el contenedor con volumen persistente
docker run -d \
  --name dxstech-edu \
  -p 3000:3000 \
  -v dxstech_data:/app/data \
  -e APP_ENV=production \
  dxstech-edu
```

---

## ☁️ Despliegue en Coolify

DxSTech Edu está completamente preparado para desplegarse en **Coolify** en menos de 2 minutos:

### Paso 1: Crear la Aplicación en Coolify
1. Inicia sesión en tu panel de Coolify.
2. Ve a tu **Project** > **Environment** y haz clic en **+ New Resource** > **Application**.
3. Selecciona **Public Repository** o **Private Repository (GitHub / GitLab)** e indica el repositorio de este proyecto.

### Paso 2: Configurar Build Pack
- Coolify detectará automáticamente el archivo `Dockerfile` en la raíz.
- En caso de usar Docker Compose, selecciona **Docker Compose**.

### Paso 3: Configurar Almacenamiento Persistente (Volumen)
Para garantizar que la base de datos SQLite (`dxstech.db`) y los registros de certificados emitidos no se pierdan al actualizar contenedores:
1. En la pestaña **Storages** / **Persistent Storage** de tu aplicación en Coolify, añade un nuevo volumen:
   - **Destination Path**: `/app/data`
   - **Name**: `dxstech-data`

### Paso 4: Variables de Entorno (Environment Variables)
Configura las siguientes variables en la pestaña **Environment Variables**:

| Variable | Valor Recomendado | Descripción |
| :--- | :--- | :--- |
| `APP_ENV` | `production` | Activa el modo de producción y seguridad |
| `PORT` | `3000` | Puerto interno en el que escucha Gin |
| `HOST` | `0.0.0.0` | Permite conexiones externas en el contenedor |
| `DATA_DIR` | `/app/data` | Ruta donde se almacena SQLite persistente |
| `JWT_SECRET` | *(obligatorio)* | Secreto de sesión, mínimo 32 caracteres (`openssl rand -hex 32`). El servidor no arranca en producción sin él |
| `APP_URL` | `https://edu.tuempresa.com` | URL pública (CORS y enlaces de verificación) |
| `CORS_ORIGINS` | *(vacío)* | Orígenes extra permitidos con credenciales, separados por coma |
| `TRUSTED_PROXIES` | rangos privados | IP/CIDR de proxies inversos de confianza (separados por coma, o `none`). Solo de ellos se acepta `X-Forwarded-For`, lo que evita falsear la IP para evadir el límite de intentos de login |
| `BOOTSTRAP_ADMIN_EMAIL` / `BOOTSTRAP_ADMIN_PASSWORD` | opcional | Primer SUPERADMIN. Sin contraseña se genera una aleatoria y se imprime una sola vez en el log; debe cambiarse al ingresar |

### Paso 5: Dominio y SSL
1. Asigna tu dominio en Coolify (por ejemplo `https://edu.tuempresa.com`).
2. Coolify generará y renovará automáticamente los certificados SSL con Let's Encrypt.
3. El generador de certificados detectará automáticamente el dominio HTTPS público a través del encabezado `X-Forwarded-Proto`, generando los códigos QR con la URL definitiva de verificación.

---

## ⚙️ Operación y rendimiento

- **Base de datos (SQLite en modo WAL):** pool de 8 conexiones (varias lecturas en paralelo, escrituras serializadas por SQLite), claves foráneas activas (`ON DELETE CASCADE` se aplica de verdad) e índices en las consultas principales.
- **Migraciones versionadas:** el esquema vive en `internal/database/migrations.go` y se registra en la tabla `schema_migrations`. Para cambiar el esquema, **agrega una migración nueva al final** (nunca edites una ya publicada). Las bases anteriores al versionado se adoptan automáticamente sin perder datos.
- **Healthcheck:** `GET /api/health` responde `200` si la base de datos responde y `503` (`degraded`) si no, de modo que Docker, Coolify o Render puedan reiniciar la instancia.
- **Timeouts:** el servidor usa plazos cortos (10 s de cabeceras, 30 s de lectura/escritura). Las rutas lentas (subida de archivos, generación con IA, certificados masivos y reportes CSV) amplían su plazo solo para ellas.
- **Compresión y caché:** respuestas gzip para JSON/JS/CSS/HTML/CSV; `/vendor` con caché inmutable, `/js` y `/css` con revalidación (304) y `/api` sin caché.
- **Archivos subidos:** al borrar o reemplazar una lección, módulo o curso se eliminan de `/uploads` los archivos que ya nadie usa (los compartidos por cursos duplicados se conservan). Para limpiar huérfanos históricos, un administrador puede llamar a `POST /api/courses/uploads/cleanup` (solo informa) y luego a `POST /api/courses/uploads/cleanup?apply=true` (elimina; ignora archivos de la última hora).

---

## 🎨 Estilos y librerías del frontend

El frontend **no depende de CDNs en tiempo de ejecución**: Tailwind se compila a `web/css/tailwind.css`, y Lucide y SheetJS se sirven desde `web/vendor/` con versión fija.
Si agregas clases de Tailwind nuevas, regenera el CSS y súbelo al repositorio:

```bash
pnpm install
pnpm build:css
```

---

## 🛠️ Estructura del Código

```text
/
├── Dockerfile                  # Multi-stage build optimizado (Alpine + Go CGO-free)
├── docker-compose.yml          # Configuración de servicios, volúmenes y healthchecks
├── .dockerignore               # Filtro de archivos para builds livianos
├── cmd/
│   └── server/main.go          # Entrypoint del servidor Gin, graceful shutdown y rutas
├── internal/
│   ├── config/config.go        # Configuración por variables de entorno (APP_ENV, PORT, DATA_DIR)
│   ├── database/db.go          # SQLite (modernc.org/sqlite) y migraciones automáticas
│   ├── ai/gemini.go            # Cliente HTTP efímero para la API de Gemini (BYOK)
│   ├── certificates/service.go # Generación de PDFs, QR de verificación y empaquetado ZIP
│   ├── quizzes/service.go      # Orquestación de exámenes y persistencia
│   └── whatsapp/service.go     # Gateway desacoplado, cola con rate limit y bitácora
├── web/
│   ├── index.html              # Shell SPA con Tailwind, Lucide y SheetJS
│   └── js/
│       ├── main.js             # Enrutador de vistas, verificación QR pública y estado
│       ├── components/         # Toast, Modal, Loading, Empty State
│       └── views/              # Certificados, Quizzes, WhatsApp, Ajustes
├── data/                       # Almacenamiento SQLite local (dxstech.db)
├── package.json                # Scripts pnpm (dev, build, start)
├── go.mod                      # Módulo Go 1.25+
└── README.md                   # Documentación completa del proyecto
```

---

## 🛡️ Licencia

Distribuido bajo la licencia MIT.
DxSTech Edu — Plataforma Educativa Premium.
