import { Toast } from '../../components/toast.js';
import { Loading } from '../../components/loading.js';

// Editor de plantilla: lienzo, imagen, lista de alumnos (Excel) y generación.

export const templateEditorMethods = {
  generateDemoTemplate() {
    const w = 1920;
    const h = 1357; // A4 aspect ratio 1.414
    const canvas = document.createElement('canvas');
    canvas.width = w;
    canvas.height = h;
    const ctx = canvas.getContext('2d');

    // Background gradient
    const bgGrad = ctx.createLinearGradient(0, 0, w, h);
    bgGrad.addColorStop(0, '#f8fafc');
    bgGrad.addColorStop(1, '#f1f5f9');
    ctx.fillStyle = bgGrad;
    ctx.fillRect(0, 0, w, h);

    // Elegant Outer Gold Border
    ctx.strokeStyle = '#d97706';
    ctx.lineWidth = 14;
    ctx.strokeRect(40, 40, w - 80, h - 80);

    // Inner Navy Border
    ctx.strokeStyle = '#1e3a8a';
    ctx.lineWidth = 4;
    ctx.strokeRect(60, 60, w - 120, h - 120);

    // Corner Ornaments
    ctx.fillStyle = '#b45309';
    const cornerSize = 40;
    ctx.fillRect(40, 40, cornerSize, cornerSize);
    ctx.fillRect(w - 40 - cornerSize, 40, cornerSize, cornerSize);
    ctx.fillRect(40, h - 40 - cornerSize, cornerSize, cornerSize);
    ctx.fillRect(w - 40 - cornerSize, h - 40 - cornerSize, cornerSize);

    // Academy Header
    ctx.fillStyle = '#312e81';
    ctx.font = 'bold 54px Helvetica, Arial, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('DxSTech Edu — Academy of Technology & AI', w / 2, 220);

    ctx.fillStyle = '#64748b';
    ctx.font = '500 28px Helvetica, Arial, sans-serif';
    ctx.fillText('OTORGA EL PRESENTE DIPLOMA DE RECONOCIMIENTO A:', w / 2, 330);

    // Course subtitle placeholder
    ctx.fillStyle = '#475569';
    ctx.font = '500 30px Helvetica, Arial, sans-serif';
    ctx.fillText('Por haber culminado con éxito el programa de especialización en:', w / 2, 880);

    ctx.fillStyle = '#1e293b';
    ctx.font = 'bold 44px Helvetica, Arial, sans-serif';
    ctx.fillText('Desarrollo Web Full-Stack, Inteligencia Artificial & Cloud', w / 2, 950);

    // Footer signature lines
    ctx.strokeStyle = '#94a3b8';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(350, 1180);
    ctx.lineTo(750, 1180);
    ctx.moveTo(1170, 1180);
    ctx.lineTo(1570, 1180);
    ctx.stroke();

    ctx.fillStyle = '#64748b';
    ctx.font = '24px Helvetica, Arial, sans-serif';
    ctx.fillText('Dirección Académica', 550, 1220);
    ctx.fillText('Coordinación de Certificación', 1370, 1220);

    // Convert to WebP quality 0.82
    const webpUrl = canvas.toDataURL('image/webp', 0.82);
    this.state.templateDataUrl = webpUrl;
    this.state.imgWidth = w;
    this.state.imgHeight = h;
    this.state.yPercent = 52.0;
    this.state.xPercent = 50.0;
    this.state.fontSize = 38;
    this.state.fontColor = '#1e293b';

    // Show controls
    document.getElementById('drop-zone-prompt').classList.add('hidden');
    document.getElementById('canvas-wrapper').classList.remove('hidden');
    const stats = document.getElementById('image-stats');
    stats.classList.remove('hidden');
    stats.innerHTML = `📐 1920x1357px • Plantilla Demo DxSTech • WebP: <strong>140 KB</strong>`;

    document.getElementById('ctrl-y').value = 52.0;
    document.getElementById('ctrl-x').value = 50.0;
    document.getElementById('ctrl-size').value = 38;

    // Load sample students if empty
    if (this.state.students.length === 0) {
      document.getElementById('students-textarea').value = "Carlos Andrés Mendoza\nMaría Fernanda Gómez\nJuan Sebastián Morales";
      this.parseStudentTextarea();
    }

    this.drawCanvas();
    Toast.success('Plantilla de demostración generada e insertada con éxito.');
  },

  // Image Processing Rule:
  // 1. validate -> 2. canvas -> 3. resize if >2000px -> 4. convert to WebP 0.82 -> 5. stats -> 6. send WebP

  async processImageFile(file) {
    if (!file.type.startsWith('image/')) {
      Toast.error('El archivo seleccionado no es una imagen válida.');
      return;
    }

    this.state.origSizeKb = (file.size / 1024).toFixed(1);

    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        let { width, height } = img;
        const maxDim = 2000;
        if (width > maxDim || height > maxDim) {
          if (width > height) {
            height = Math.round((height * maxDim) / width);
            width = maxDim;
          } else {
            width = Math.round((width * maxDim) / height);
            height = maxDim;
          }
        }

        // Canvas for WebP conversion
        const offCanvas = document.createElement('canvas');
        offCanvas.width = width;
        offCanvas.height = height;
        const ctx = offCanvas.getContext('2d');
        ctx.drawImage(img, 0, 0, width, height);

        // Convert to WebP quality 0.82
        const webpDataUrl = offCanvas.toDataURL('image/webp', 0.82);
        
        // Estimate WebP size
        const head = 'data:image/webp;base64,';
        const rawB64 = webpDataUrl.substring(head.length);
        const webpBytes = Math.round((rawB64.length * 3) / 4);
        this.state.webpSizeKb = (webpBytes / 1024).toFixed(1);
        this.state.templateDataUrl = webpDataUrl;
        this.state.imgWidth = width;
        this.state.imgHeight = height;

        // UI update
        const prompt = document.getElementById('drop-zone-prompt');
        const canvasWrapper = document.getElementById('canvas-wrapper');
        const stats = document.getElementById('image-stats');

        prompt.classList.add('hidden');
        canvasWrapper.classList.remove('hidden');
        stats.classList.remove('hidden');
        stats.innerHTML = `📐 ${width}x${height}px • WebP: <strong>${this.state.webpSizeKb} KB</strong> (original: ${this.state.origSizeKb} KB)`;

        this.drawCanvas();
        Toast.success('Plantilla procesada y optimizada en WebP');
      };
      img.src = e.target.result;
    };
    reader.readAsDataURL(file);
  },

  drawCanvas() {
    if (!this.state.templateDataUrl) return;

    const canvas = document.getElementById('cert-canvas');
    if (!canvas) return;

    const img = new Image();
    img.onload = () => {
      canvas.width = img.width;
      canvas.height = img.height;
      const ctx = canvas.getContext('2d');

      // Draw background template
      ctx.drawImage(img, 0, 0, img.width, img.height);

      // Coordinates
      const x = (this.state.xPercent / 100) * img.width;
      const y = (this.state.yPercent / 100) * img.height;

      // Draw Crosshair Guides
      ctx.save();
      ctx.strokeStyle = 'rgba(99, 102, 241, 0.6)';
      ctx.lineWidth = 1.5;
      ctx.setLineDash([6, 6]);

      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(img.width, y);
      ctx.stroke();

      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, img.height);
      ctx.stroke();
      ctx.restore();

      // Draw Text Preview
      const scaleFactor = img.width / 1000;
      const effectiveFontSize = Math.round(this.state.fontSize * scaleFactor);

      ctx.font = `bold ${effectiveFontSize}px ${this.state.fontFamily}, sans-serif`;
      ctx.fillStyle = this.state.fontColor;
      ctx.textAlign = this.state.textAlign;
      ctx.textBaseline = 'middle';

      // Text shadow for high contrast preview
      ctx.shadowColor = 'rgba(255, 255, 255, 0.8)';
      ctx.shadowBlur = 4;

      ctx.fillText(this.state.sampleName, x, y);

      // Draw indicator target dot
      ctx.beginPath();
      ctx.arc(x, y, 4, 0, Math.PI * 2);
      ctx.fillStyle = '#4f46e5';
      ctx.fill();

      // Draw Verification QR Code Preview
      if (this.state.includeQR) {
        const qrX = (this.state.qrXPercent / 100) * img.width;
        const qrY = (this.state.qrYPercent / 100) * img.height;
        const qrSizePx = (this.state.qrSize / 297) * img.width;

        ctx.save();
        ctx.fillStyle = '#ffffff';
        ctx.strokeStyle = '#4f46e5';
        ctx.lineWidth = 2;
        ctx.fillRect(qrX - qrSizePx / 2, qrY - qrSizePx / 2, qrSizePx, qrSizePx);
        ctx.strokeRect(qrX - qrSizePx / 2, qrY - qrSizePx / 2, qrSizePx, qrSizePx);

        // Simulated QR corners
        ctx.fillStyle = '#1e293b';
        const pSize = Math.max(4, qrSizePx * 0.22);
        ctx.fillRect(qrX - qrSizePx / 2 + 2, qrY - qrSizePx / 2 + 2, pSize, pSize);
        ctx.fillRect(qrX + qrSizePx / 2 - pSize - 2, qrY - qrSizePx / 2 + 2, pSize, pSize);
        ctx.fillRect(qrX - qrSizePx / 2 + 2, qrY + qrSizePx / 2 - pSize - 2, pSize, pSize);

        // QR label
        ctx.fillStyle = '#4f46e5';
        ctx.font = `bold ${Math.max(9, Math.round(qrSizePx * 0.16))}px monospace`;
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText('QR DXS', qrX, qrY);

        // Code label underneath
        ctx.fillStyle = '#64748b';
        ctx.font = `bold ${Math.max(8, Math.round(qrSizePx * 0.12))}px sans-serif`;
        ctx.fillText('VERIFICADO', qrX, qrY + qrSizePx / 2 + 10);
        ctx.restore();
      }
    };
    img.src = this.state.templateDataUrl;
  },

  parseStudentTextarea() {
    const text = document.getElementById('students-textarea').value;
    const lines = text.split('\n').map(l => l.trim()).filter(l => l.length > 0);
    this.state.students = lines;
    this.updateStudentCount();
  },

  processExcelFile(file) {
    if (!window.XLSX) {
      Toast.error('Biblioteca SheetJS no cargada');
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = new Uint8Array(e.target.result);
        const workbook = window.XLSX.read(data, { type: 'array' });
        const firstSheetName = workbook.SheetNames[0];
        const worksheet = workbook.Sheets[firstSheetName];
        const json = window.XLSX.utils.sheet_to_json(worksheet, { header: 1 });

        if (json.length < 2) {
          Toast.warning('El archivo Excel parece estar vacío o no contiene filas de datos.');
          return;
        }

        const headers = json[0].map(h => String(h).trim());
        const select = document.getElementById('excel-column-select');
        select.innerHTML = '';

        // Auto-detect column matching nombre, nombres, alumno, estudiante
        const candidates = ['nombre', 'nombres', 'alumno', 'alumnos', 'estudiante', 'estudiantes', 'participante'];
        let matchedIndex = 0;

        headers.forEach((h, index) => {
          const opt = document.createElement('option');
          opt.value = index;
          opt.textContent = `Columna ${index + 1}: ${h}`;
          select.appendChild(opt);

          const lower = h.toLowerCase();
          if (candidates.some(c => lower.includes(c))) {
            matchedIndex = index;
          }
        });

        select.value = matchedIndex;
        document.getElementById('excel-column-picker').classList.remove('hidden');

        const extractNames = (colIdx) => {
          const names = [];
          for (let i = 1; i < json.length; i++) {
            const row = json[i];
            if (row && row[colIdx]) {
              const val = String(row[colIdx]).trim();
              if (val) names.push(val);
            }
          }
          this.state.students = names;
          this.updateStudentCount();
          Toast.success(`Se importaron ${names.length} estudiantes desde Excel.`);
        };

        select.onchange = () => extractNames(parseInt(select.value));
        extractNames(matchedIndex);
      } catch (err) {
        Toast.error('Error al procesar el archivo Excel: ' + err.message);
      }
    };
    reader.readAsArrayBuffer(file);
  },

  updateStudentCount() {
    const badge = document.getElementById('student-count-badge');
    const count = this.state.students.length;
    badge.textContent = `${count} ${count === 1 ? 'alumno' : 'alumnos'}`;
  },

  async generateCertificates(isSinglePreview = false) {
    if (!this.state.templateDataUrl) {
      Toast.warning('Por favor carga primero la imagen de la plantilla del certificado.');
      return;
    }

    let targetStudents = this.state.students;
    if (isSinglePreview) {
      targetStudents = targetStudents.length > 0 ? [targetStudents[0]] : ['ESTUDIANTE DEMO'];
    }

    if (targetStudents.length === 0) {
      Toast.warning('Debes ingresar al menos un nombre de estudiante.');
      return;
    }

    Loading.overlay(`Generando ${isSinglePreview ? '1 PDF de prueba con QR de verificación' : targetStudents.length + ' certificados con QR en PDF'}...`);

    try {
      const payload = {
        template: this.state.templateDataUrl,
        xPercent: this.state.xPercent,
        yPercent: this.state.yPercent,
        fontSize: this.state.fontSize,
        fontColor: this.state.fontColor,
        fontFamily: this.state.fontFamily,
        textAlign: this.state.textAlign,
        includeQR: this.state.includeQR,
        qrXPercent: this.state.qrXPercent,
        qrYPercent: this.state.qrYPercent,
        qrSize: this.state.qrSize,
        courseTitle: this.state.courseTitle,
        students: targetStudents,
      };

      const response = await fetch('/api/certificates/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errJson = await response.json().catch(() => ({}));
        throw new Error(errJson.error || `Error HTTP ${response.status}`);
      }

      // Download ZIP
      const blob = await response.blob();
      const downloadUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = isSinglePreview ? 'certificado_muestra.zip' : 'certificados_dxstech.zip';
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(downloadUrl);

      Toast.success(isSinglePreview ? 'PDF de prueba generado y descargado.' : `¡Éxito! ${targetStudents.length} certificados descargados en ZIP.`);
      this.loadIssuedCertificates();
    } catch (err) {
      Toast.error('Fallo al generar certificados: ' + err.message);
    } finally {
      Loading.hideOverlay();
    }
  },
};
