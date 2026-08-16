/**
 * Client-side grid filtering (AND across rules). Optional getValue(row, field) for computed columns.
 */

const TEXT_OPS = [
  { value: 'contains', label: 'Contains' },
  { value: 'not_contains', label: 'Does not contain' },
  { value: 'equals', label: 'Equals' },
  { value: 'not_equals', label: 'Does not equal' },
  { value: 'starts_with', label: 'Starts with' },
  { value: 'ends_with', label: 'Ends with' },
  { value: 'in_list', label: 'In list (comma-separated)' },
  { value: 'not_in_list', label: 'Not in list (comma-separated)' },
  { value: 'empty', label: 'Is empty' },
  { value: 'not_empty', label: 'Is not empty' },
];

const NUM_OPS = [
  { value: 'num_eq', label: '=' },
  { value: 'num_ne', label: '≠' },
  { value: 'num_gt', label: '>' },
  { value: 'num_gte', label: '≥' },
  { value: 'num_lt', label: '<' },
  { value: 'num_lte', label: '≤' },
  { value: 'empty', label: 'Is empty' },
  { value: 'not_empty', label: 'Is not empty' },
];

const BOOL_OPS = [
  { value: 'bool_true', label: 'Is true / yes' },
  { value: 'bool_false', label: 'Is false / no' },
  { value: 'empty', label: 'Is empty' },
  { value: 'not_empty', label: 'Is not empty' },
];

/** Heuristic: pick operator set from field name + sample value. */
export function inferFieldKind(field) {
  const f = String(field || '').toLowerCase();
  if (
    f === 'active_flag' ||
    f === 'private' ||
    f === 'disabled' ||
    f.endsWith('_flag') ||
    f.startsWith('is_')
  ) {
    return 'bool';
  }
  if (
    f.includes('grade') ||
    f.includes('capacity') ||
    f.includes('count') ||
    f === 'distance' ||
    f === 'trip_days' ||
    f === 'vehicle_id' ||
    f === 'trip_id' ||
    f === 'model' ||
    f.endsWith('_id')
  ) {
    return 'number';
  }
  return 'text';
}

export function operatorsForFieldKind(kind) {
  if (kind === 'bool') return BOOL_OPS;
  if (kind === 'number') return NUM_OPS;
  return TEXT_OPS;
}

export function defaultOperatorsForField(field) {
  return operatorsForFieldKind(inferFieldKind(field));
}

function cellStr(row, field, getValue) {
  const raw = getValue ? getValue(row, field) : row[field];
  if (raw == null) return '';
  if (typeof raw === 'boolean') return raw ? 'true' : 'false';
  return String(raw);
}

function cellNum(row, field, getValue) {
  const raw = getValue ? getValue(row, field) : row[field];
  if (raw == null || raw === '') return NaN;
  if (typeof raw === 'number') return raw;
  const n = Number(String(raw).trim());
  return Number.isFinite(n) ? n : NaN;
}

function cellBool(row, field, getValue) {
  const raw = getValue ? getValue(row, field) : row[field];
  if (raw == null) return null;
  if (typeof raw === 'boolean') return raw;
  if (typeof raw === 'number') return raw !== 0;
  const s = String(raw).toLowerCase().trim();
  if (s === 'true' || s === '1' || s === 'yes') return true;
  if (s === 'false' || s === '0' || s === 'no') return false;
  return null;
}

/**
 * @param {object} row
 * @param {{ field: string, op: string, value: string }} rule
 * @param {(row: object, field: string) => any} [getValue]
 */
export function rowMatchesRule(row, rule, getValue) {
  const field = rule.field;
  const op = rule.op || 'contains';
  const v = rule.value == null ? '' : String(rule.value);
  const kind = inferFieldKind(field);

  if (op === 'empty') {
    const s = cellStr(row, field, getValue).trim();
    return s === '';
  }
  if (op === 'not_empty') {
    const s = cellStr(row, field, getValue).trim();
    return s !== '';
  }

  if (op === 'bool_true' || op === 'bool_false') {
    const b = cellBool(row, field, getValue);
    if (op === 'bool_true') return b === true;
    if (op === 'bool_false') return b === false;
  }

  if (kind === 'number') {
    const a = cellNum(row, field, getValue);
    const b = Number(String(v).trim());
    if (op === 'num_eq') return Number.isFinite(a) && Number.isFinite(b) && a === b;
    if (op === 'num_ne') return Number.isFinite(a) && Number.isFinite(b) && a !== b;
    if (op === 'num_gt') return Number.isFinite(a) && Number.isFinite(b) && a > b;
    if (op === 'num_gte') return Number.isFinite(a) && Number.isFinite(b) && a >= b;
    if (op === 'num_lt') return Number.isFinite(a) && Number.isFinite(b) && a < b;
    if (op === 'num_lte') return Number.isFinite(a) && Number.isFinite(b) && a <= b;
    // fallback: compare as strings for date ISO
    const sa = cellStr(row, field, getValue).toLowerCase();
    const sb = v.toLowerCase();
    if (op === 'contains') return sa.includes(sb);
    if (op === 'equals') return sa === sb;
    return false;
  }

  const s = cellStr(row, field, getValue).toLowerCase();
  const t = v.toLowerCase();
  switch (op) {
    case 'contains':
      return s.includes(t);
    case 'not_contains':
      return !s.includes(t);
    case 'equals':
      return s === t;
    case 'not_equals':
      return s !== t;
    case 'starts_with':
      return s.startsWith(t);
    case 'ends_with':
      return s.endsWith(t);
    case 'in_list': {
      const parts = v
        .split(',')
        .map((x) => x.trim().toLowerCase())
        .filter(Boolean);
      return parts.includes(s);
    }
    case 'not_in_list': {
      const parts = v
        .split(',')
        .map((x) => x.trim().toLowerCase())
        .filter(Boolean);
      return !parts.includes(s);
    }
    default:
      return s.includes(t);
  }
}

/**
 * @param {object[]} rows
 * @param {Array<{ field: string, op: string, value: string }>} rules
 * @param {(row: object, field: string) => any} [getValue]
 */
export function applyFilterRules(rows, rules, getValue) {
  if (!rows || !rows.length) return rows || [];
  const active = (rules || []).filter((r) => r && r.field);
  if (!active.length) return rows;
  return rows.filter((row) => active.every((rule) => rowMatchesRule(row, rule, getValue)));
}

/**
 * Same quick-search behavior as AdminDataGrid (any cell substring).
 */
export function applyQuickSearch(rows, searchText) {
  if (!searchText || !String(searchText).trim()) return rows || [];
  const lower = String(searchText).toLowerCase();
  return (rows || []).filter((row) =>
    Object.values(row || {}).some((v) => v != null && String(v).toLowerCase().includes(lower))
  );
}

/**
 * Chain: quick search then advanced rules (matches AdminDataGrid).
 */
export function applyQuickSearchThenRules(rows, searchText, rules, getValue) {
  if (!rows || !rows.length) return rows || [];
  const afterSearch = applyQuickSearch(rows, searchText);
  return applyFilterRules(afterSearch, rules, getValue);
}

export function newRuleId() {
  return `r-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}
