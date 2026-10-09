// Lectura de hojas de cálculo (.xlsx y .csv) en el navegador, sin SheetJS.
//
// Reemplaza a la librería SheetJS 0.18.5, que tiene vulnerabilidades conocidas
// (contaminación de prototipos y ReDoS) y ya no se publica en npm. Un .xlsx es un
// ZIP con XML: se descomprime con fflate (solo los archivos necesarios, con tope de
// tamaño contra "zip bombs") y se analiza con DOMParser, el parser nativo del navegador.
// No se admite el formato antiguo .xls (BIFF): se pide guardar como .xlsx o .csv.

import { unzipSync, strFromU8 } from '/vendor/fflate-0.8.2.browser.js';

export const MAX_SPREADSHEET_BYTES = 5 * 1024 * 1024; // archivo subido
const MAX_UNZIPPED_BYTES = 30 * 1024 * 1024; // contenido descomprimido (anti zip-bomb)
const MAX_ROWS = 50000;

/** Devuelve las filas de la primera hoja como arreglos de texto (filas vacías omitidas). */
export async function readSpreadsheetRows(file) {
  if (!file) throw new Error('No se seleccionó ningún archivo.');
  if (file.size > MAX_SPREADSHEET_BYTES) {
    throw new Error(`El archivo supera el máximo de ${MAX_SPREADSHEET_BYTES / 1024 / 1024} MB.`);
  }

  const name = (file.name || '').toLowerCase();
  const bytes = new Uint8Array(await file.arrayBuffer());

  if (name.endsWith('.csv') || name.endsWith('.txt')) {
    return parseCsv(new TextDecoder('utf-8').decode(bytes).replace(/^﻿/, ''));
  }
  if (name.endsWith('.xls')) {
    throw new Error('El formato antiguo .xls no es compatible. Guarda el archivo como .xlsx o .csv e inténtalo de nuevo.');
  }
  if (!name.endsWith('.xlsx') && !isZip(bytes)) {
    throw new Error('Formato no compatible. Usa un archivo .xlsx o .csv.');
  }
  return parseXlsx(bytes);
}

function isZip(bytes) {
  return bytes.length > 4 && bytes[0] === 0x50 && bytes[1] === 0x4b && bytes[2] === 0x03 && bytes[3] === 0x04;
}

// ---------------------------------------------------------------- CSV

/** CSV según RFC 4180 (comillas, comas y saltos de línea dentro de campos). Detecta ; como separador. */
export function parseCsv(text) {
  const firstLine = text.split(/\r?\n/, 1)[0] || '';
  const delimiter = (firstLine.match(/;/g) || []).length > (firstLine.match(/,/g) || []).length ? ';' : ',';

  const rows = [];
  let row = [];
  let field = '';
  let inQuotes = false;

  const endField = () => { row.push(field); field = ''; };
  const endRow = () => {
    endField();
    if (row.some((v) => String(v).trim() !== '')) rows.push(row);
    row = [];
    if (rows.length > MAX_ROWS) throw new Error('El archivo tiene demasiadas filas.');
  };

  for (let i = 0; i < text.length; i++) {
    const ch = text[i];
    if (inQuotes) {
      if (ch === '"') {
        if (text[i + 1] === '"') { field += '"'; i++; } else { inQuotes = false; }
      } else {
        field += ch;
      }
    } else if (ch === '"') {
      inQuotes = true;
    } else if (ch === delimiter) {
      endField();
    } else if (ch === '\n') {
      endRow();
    } else if (ch !== '\r') {
      field += ch;
    }
  }
  if (field !== '' || row.length) endRow();
  return rows;
}

// ---------------------------------------------------------------- XLSX

function xml(bytes) {
  const doc = new DOMParser().parseFromString(strFromU8(bytes), 'application/xml');
  if (doc.getElementsByTagName('parsererror').length) throw new Error('El archivo .xlsx está dañado o no es válido.');
  return doc;
}

const byTag = (node, name) => Array.from(node.getElementsByTagName(name));

function columnIndex(ref) {
  const letters = (ref.match(/^[A-Z]+/i) || [''])[0].toUpperCase();
  let n = 0;
  for (const ch of letters) n = n * 26 + (ch.charCodeAt(0) - 64);
  return n - 1;
}

// Excel guarda números largos en notación científica (p. ej. 5.73E+11): se expanden.
function normalizeNumber(raw) {
  if (/^-?\d+(\.\d+)?[eE][+-]?\d+$/.test(raw)) {
    const n = Number(raw);
    if (Number.isFinite(n)) return n.toLocaleString('fullwide', { useGrouping: false, maximumFractionDigits: 10 });
  }
  return raw;
}

function parseXlsx(bytes) {
  let files;
  try {
    let total = 0;
    files = unzipSync(bytes, {
      // Solo se descomprime lo imprescindible y se corta si el contenido es desproporcionado.
      filter: (f) => {
        total += f.originalSize;
        if (total > MAX_UNZIPPED_BYTES) throw new Error('El archivo descomprimido es demasiado grande.');
        return /^xl\/(workbook\.xml|_rels\/workbook\.xml\.rels|sharedStrings\.xml|worksheets\/[^/]+\.xml)$/.test(f.name);
      },
    });
  } catch (err) {
    if (err.message.includes('demasiado grande')) throw err;
    throw new Error('No se pudo leer el archivo .xlsx. Verifica que no esté dañado.');
  }
  if (!files['xl/workbook.xml']) throw new Error('El archivo no parece un libro de Excel (.xlsx) válido.');

  // Cadenas compartidas
  const shared = [];
  if (files['xl/sharedStrings.xml']) {
    for (const si of byTag(xml(files['xl/sharedStrings.xml']), 'si')) {
      shared.push(byTag(si, 't').map((t) => t.textContent).join(''));
    }
  }

  // Primera hoja según el orden del libro
  const wb = xml(files['xl/workbook.xml']);
  const firstSheet = byTag(wb, 'sheet')[0];
  let sheetPath = 'xl/worksheets/sheet1.xml';
  if (firstSheet && files['xl/_rels/workbook.xml.rels']) {
    const rid = firstSheet.getAttribute('r:id') || firstSheet.getAttributeNS('http://schemas.openxmlformats.org/officeDocument/2006/relationships', 'id');
    const rel = byTag(xml(files['xl/_rels/workbook.xml.rels']), 'Relationship').find((r) => r.getAttribute('Id') === rid);
    if (rel) {
      const target = rel.getAttribute('Target') || '';
      sheetPath = target.startsWith('/') ? target.slice(1) : `xl/${target.replace(/^\.\//, '')}`;
    }
  }
  if (!files[sheetPath]) throw new Error('No se encontró la primera hoja del libro.');

  const rows = [];
  for (const rowEl of byTag(xml(files[sheetPath]), 'row')) {
    const cells = [];
    let next = 0;
    for (const c of byTag(rowEl, 'c')) {
      const ref = c.getAttribute('r');
      const idx = ref ? columnIndex(ref) : next;
      next = idx + 1;

      const type = c.getAttribute('t');
      const v = c.getElementsByTagName('v')[0];
      let value = '';
      if (type === 's') value = shared[parseInt(v?.textContent ?? '', 10)] ?? '';
      else if (type === 'inlineStr') value = byTag(c, 't').map((t) => t.textContent).join('');
      else if (type === 'b') value = v?.textContent === '1' ? 'VERDADERO' : 'FALSO';
      else value = normalizeNumber(v?.textContent ?? '');

      while (cells.length < idx) cells.push('');
      cells[idx] = value;
    }
    if (cells.some((v) => String(v).trim() !== '')) rows.push(cells);
    if (rows.length > MAX_ROWS) throw new Error('El archivo tiene demasiadas filas.');
  }
  return rows;
}
