export const SESSION_SWATCH_COLORS = [
  '#3c2474',
  '#5b34a8',
  '#7c3aed',
  '#8b5cf6',
  '#6d3cc4',
  '#4c2f8a',
];

function hashSessionId(id) {
  const s = String(id || '');
  let h = 0;
  for (let i = 0; i < s.length; i += 1) h = (h * 31 + s.charCodeAt(i)) >>> 0;
  return h;
}

export function sessionColorsFor(ids) {
  const colors = [];
  let previous = null;
  (ids || []).forEach((id) => {
    const n = SESSION_SWATCH_COLORS.length;
    const start = hashSessionId(id) % n;
    let color = SESSION_SWATCH_COLORS[start];
    for (let k = 0; k < n; k += 1) {
      const candidate = SESSION_SWATCH_COLORS[(start + k) % n];
      if (candidate !== previous) {
        color = candidate;
        break;
      }
    }
    colors.push(color);
    previous = color;
  });
  return colors;
}
