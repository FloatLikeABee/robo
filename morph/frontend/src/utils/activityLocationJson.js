/**
 * Activity (Trip) `location` column: JSON with optional `location` description and `lat` / `lng`.
 * Plain strings from legacy data are treated as description-only.
 */

function pickNum(v) {
  if (v == null || v === '') return null;
  const n = typeof v === 'number' ? v : parseFloat(String(v), 10);
  return Number.isFinite(n) ? n : null;
}

/**
 * @returns {{ description: string, lat: number | null, lng: number | null }}
 */
export function parseActivityLocationJson(raw) {
  if (raw == null || raw === '') return { description: '', lat: null, lng: null };
  if (typeof raw === 'object' && raw !== null && !Array.isArray(raw)) {
    const loc = raw.location != null ? String(raw.location) : '';
    const lat = pickNum(raw.lat ?? raw.latitude ?? raw.y);
    const lng = pickNum(raw.lng ?? raw.lon ?? raw.longitude ?? raw.x);
    return { description: loc, lat, lng };
  }
  const s = String(raw).trim();
  if (!s) return { description: '', lat: null, lng: null };
  try {
    const obj = JSON.parse(s);
    if (obj && typeof obj === 'object' && !Array.isArray(obj)) {
      const loc = obj.location != null ? String(obj.location) : '';
      const lat = pickNum(obj.lat ?? obj.latitude ?? obj.y);
      const lng = pickNum(obj.lng ?? obj.lon ?? obj.longitude ?? obj.x);
      return { description: loc, lat, lng };
    }
  } catch {
    return { description: s, lat: null, lng: null };
  }
  return { description: s, lat: null, lng: null };
}

/** Human list / grid cell: only the optional `location` description string. */
export function activityLocationDescriptionOnly(raw) {
  const { description } = parseActivityLocationJson(raw);
  return description || '';
}

/**
 * Merge editor values into prior JSON (preserves unrelated keys). Returns JSON string or null if empty.
 */
export function mergeActivityLocationJson(prevRaw, { description, lat, lng }) {
  let base = {};
  if (prevRaw != null && prevRaw !== '') {
    if (typeof prevRaw === 'object' && !Array.isArray(prevRaw)) {
      base = { ...prevRaw };
    } else {
      const s = String(prevRaw).trim();
      if (s) {
        try {
          const o = JSON.parse(s);
          if (o && typeof o === 'object' && !Array.isArray(o)) base = { ...o };
        } catch {
          base = {};
        }
      }
    }
  }

  const d = description != null ? String(description).trim() : '';
  if (d) base.location = d;
  else delete base.location;

  const latN = lat != null && lat !== '' ? parseFloat(String(lat), 10) : NaN;
  const lngN = lng != null && lng !== '' ? parseFloat(String(lng), 10) : NaN;
  if (Number.isFinite(latN) && Number.isFinite(lngN)) {
    base.lat = latN;
    base.lng = lngN;
  } else {
    delete base.lat;
    delete base.lng;
  }

  const cleaned = {};
  for (const k of Object.keys(base)) {
    const v = base[k];
    if (v != null && v !== '') cleaned[k] = v;
  }
  if (Object.keys(cleaned).length === 0) return null;
  return JSON.stringify(cleaned);
}
