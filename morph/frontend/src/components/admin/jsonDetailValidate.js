/** Max object/array nesting depth allowed for entity `detail` JSON (inclusive). */
export const JSON_DETAIL_MAX_NESTING = 5;

/**
 * Longest object/array nesting chain from root. Primitives depth 0.
 * `{}` → 1, `[]` → 1, `{a:1}` → 1, `{a:{b:1}}` → 2
 */
export function jsonDetailNestingDepth(v) {
  if (v === null || typeof v !== 'object') return 0;
  if (Array.isArray(v)) {
    if (!v.length) return 1;
    return 1 + Math.max(0, ...v.map(jsonDetailNestingDepth));
  }
  const vals = Object.values(v);
  if (!vals.length) return 1;
  return 1 + Math.max(...vals.map(jsonDetailNestingDepth));
}

/**
 * Validate detail JSON: parseable + nesting depth cap.
 * @param {string|object|null|undefined} value
 * @returns {{ ok: true, parsed: any } | { ok: false, error: string }}
 */
export function validateJsonDetailStructure(value) {
  if (value == null) return { ok: true, parsed: null };
  let parsed;
  if (typeof value === 'object' && !(value instanceof Date)) {
    parsed = value;
  } else {
    const s = String(value).trim();
    if (!s) return { ok: true, parsed: null };
    try {
      parsed = JSON.parse(s);
    } catch (e) {
      return { ok: false, error: e?.message || 'Invalid JSON' };
    }
  }
  const d = jsonDetailNestingDepth(parsed);
  if (d > JSON_DETAIL_MAX_NESTING) {
    return {
      ok: false,
      error: `JSON nesting is too deep (${d} levels). Maximum allowed is ${JSON_DETAIL_MAX_NESTING} layers.`,
    };
  }
  return { ok: true, parsed };
}
