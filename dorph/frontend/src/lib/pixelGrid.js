export const PIXEL_MAX = 64;
const ACCENT = '#38bdf8';
const HEX6 = /^#[0-9a-fA-F]{6}$/;
const HEX3 = /^#[0-9a-fA-F]{3}$/;

function expandHex3(tok) {
  return `#${tok[1]}${tok[1]}${tok[2]}${tok[2]}${tok[3]}${tok[3]}`;
}

function cellColor(tok) {
  if (tok === '.') return null;
  if (tok === '#') return ACCENT;
  if (HEX6.test(tok)) return tok.toLowerCase();
  if (HEX3.test(tok)) return expandHex3(tok).toLowerCase();
  return undefined;
}

/** Parse a ```pixel``` body. Returns null if empty, too large, or malformed. */
export function parsePixelGrid(text) {
  const lines = String(text || '')
    .split(/\n/)
    .map((l) => l.trim())
    .filter((l) => l.length > 0);
  if (lines.length === 0 || lines.length > PIXEL_MAX) return null;
  const rows = [];
  let width = 0;
  for (const line of lines) {
    const cells = line.split(/[\s,]+/).filter(Boolean);
    if (cells.length === 0 || cells.length > PIXEL_MAX) return null;
    const row = [];
    for (const tok of cells) {
      const color = cellColor(tok);
      if (color === undefined) return null;
      row.push(color);
    }
    if (width === 0) width = row.length;
    else if (row.length !== width) return null;
    rows.push(row);
  }
  return { width, height: rows.length, rows };
}

export function drawPixelGrid(canvas, grid, scale = 8) {
  if (!canvas || !grid) return;
  const s = Math.max(1, Math.min(24, scale));
  canvas.width = grid.width * s;
  canvas.height = grid.height * s;
  const ctx = canvas.getContext('2d');
  if (!ctx) return;
  ctx.imageSmoothingEnabled = false;
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  for (let y = 0; y < grid.height; y += 1) {
    for (let x = 0; x < grid.width; x += 1) {
      const c = grid.rows[y][x];
      if (!c) continue;
      ctx.fillStyle = c;
      ctx.fillRect(x * s, y * s, s, s);
    }
  }
}
