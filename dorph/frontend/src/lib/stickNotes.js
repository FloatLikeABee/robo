export const STICKER_PALETTE = [
  '#fde68a',
  '#fdba74',
  '#fca5a5',
  '#f9a8d4',
  '#d8b4fe',
  '#c4b5fd',
  '#a5b4fc',
  '#93c5fd',
  '#67e8f9',
  '#6ee7b7',
  '#bef264',
  '#fcd34d',
];

export const STICKER_TITLE_PALETTE = [
  '#d97706',
  '#ea580c',
  '#dc2626',
  '#db2777',
  '#7c3aed',
  '#6d28d9',
  '#4f46e5',
  '#2563eb',
  '#0891b2',
  '#059669',
  '#65a30d',
  '#ca8a04',
];

function byId(a, b) {
  const an = Number(a.id);
  const bn = Number(b.id);
  if (Number.isFinite(an) && Number.isFinite(bn) && an !== bn) return an - bn;
  return String(a.id).localeCompare(String(b.id));
}

function paletteIndex(rows) {
  const sorted = [...(rows || [])].sort(byId);
  const index = new Map();
  sorted.forEach((row, i) => {
    index.set(row.id, i % STICKER_PALETTE.length);
  });
  return index;
}

/** Face colors from the full list sorted by id. Search must pass this same list, not the visible subset. */
export function stickerColorMap(rows) {
  const map = new Map();
  paletteIndex(rows).forEach((i, id) => {
    map.set(id, STICKER_PALETTE[i]);
  });
  return map;
}

/** Title-bar colors, paired with stickerColorMap by the same id order. */
export function stickerTitleColorMap(rows) {
  const map = new Map();
  paletteIndex(rows).forEach((i, id) => {
    map.set(id, STICKER_TITLE_PALETTE[i]);
  });
  return map;
}

export function filterStickNotes(rows, query) {
  const q = String(query || '').trim().toLowerCase();
  if (!q) return rows || [];
  return (rows || []).filter((row) => {
    const title = String(row.title || '').toLowerCase();
    const body = String(row.description || '').toLowerCase();
    return title.includes(q) || body.includes(q);
  });
}
