import { rowMatchesRule } from './gridFilterUtils';

/**
 * @typedef {{ version?: number, rules?: object[], groups?: object[] }} GridColorConfig
 */

export function normalizeColorConfig(raw) {
  if (raw == null || typeof raw !== 'object' || Array.isArray(raw)) {
    return { version: 1, rules: [], groups: [] };
  }
  return {
    version: raw.version || 1,
    rules: Array.isArray(raw.rules) ? raw.rules : [],
    groups: Array.isArray(raw.groups) ? raw.groups : [],
  };
}

export function normalizeHex(c) {
  if (c == null) return null;
  const s = String(c).trim();
  if (/^#[0-9A-Fa-f]{6}$/.test(s)) return s;
  if (/^[0-9A-Fa-f]{6}$/.test(s)) return `#${s}`;
  return null;
}

function hexToRgba(hex, alpha) {
  const h = String(hex).replace('#', '');
  if (h.length !== 6) return `rgba(0,0,0,${alpha})`;
  const r = parseInt(h.slice(0, 2), 16);
  const g = parseInt(h.slice(2, 4), 16);
  const b = parseInt(h.slice(4, 6), 16);
  return `rgba(${r},${g},${b},${alpha})`;
}

/**
 * Priority: rules (by priority asc) → groups (by priority asc).
 * @returns {{ recordId: string, finalColor: string|null, appliedSource: 'rule'|'group'|null, appliedRuleId: string|null, appliedGroupId: string|null }}
 */
export function resolveRowColor(row, config, getValue) {
  const c = normalizeColorConfig(config);
  const id = row?.id;
  const idStr = id != null ? String(id) : '';

  const rulesSorted = [...c.rules].sort((a, b) => (Number(a.priority) || 0) - (Number(b.priority) || 0));
  for (const rule of rulesSorted) {
    if (!rule || !rule.field || rule.color == null) continue;
    const r = {
      field: rule.field,
      op: rule.op || 'equals',
      value: rule.value != null ? String(rule.value) : '',
    };
    if (rowMatchesRule(row, r, getValue)) {
      const hex = normalizeHex(rule.color);
      if (hex) {
        return {
          recordId: idStr,
          finalColor: hex,
          appliedSource: 'rule',
          appliedRuleId: rule.id != null ? String(rule.id) : null,
          appliedGroupId: null,
        };
      }
    }
  }

  const groupsSorted = [...c.groups].sort((a, b) => (Number(a.priority) || 0) - (Number(b.priority) || 0));
  for (const g of groupsSorted) {
    if (!g || !g.field || g.color == null) continue;
    const r = {
      field: g.field,
      op: 'equals',
      value: g.value != null ? String(g.value) : '',
    };
    if (rowMatchesRule(row, r, getValue)) {
      const hex = normalizeHex(g.color);
      if (hex) {
        return {
          recordId: idStr,
          finalColor: hex,
          appliedSource: 'group',
          appliedRuleId: null,
          appliedGroupId: g.id != null ? String(g.id) : null,
        };
      }
    }
  }

  return {
    recordId: idStr,
    finalColor: null,
    appliedSource: null,
    appliedRuleId: null,
    appliedGroupId: null,
  };
}

/**
 * Build stable class tokens per row and unique color styles for injection.
 */
export function buildRowColorTokens(rows, config, getValue) {
  const meta = new Map();
  const tokenForHex = new Map();
  let seq = 0;
  for (const row of rows || []) {
    const res = resolveRowColor(row, config, getValue);
    if (!res.finalColor) continue;
    let tok = tokenForHex.get(res.finalColor);
    if (!tok) {
      tok = `gct-${seq}`;
      seq += 1;
      tokenForHex.set(res.finalColor, tok);
    }
    meta.set(row.id, { ...res, token: tok });
  }
  return { meta, uniqueColors: [...tokenForHex.entries()] };
}

export function buildRowColorCss(uniqueColors, isDark) {
  const bgAlpha = isDark ? 0.14 : 0.09;
  return (uniqueColors || [])
    .map(([hex, token]) => {
      const bg = hexToRgba(hex, bgAlpha);
      return `.MuiDataGrid-row.${token}{background-color:${bg}!important;box-shadow:inset 4px 0 0 ${hex};}`;
    })
    .join('');
}
