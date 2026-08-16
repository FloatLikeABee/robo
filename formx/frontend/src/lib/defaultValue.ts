import type { QuestionConfig } from './api';

export type QuestionLike = { type: string; config: QuestionConfig };

/** Turn stored JSON into values suitable for public form state. */
export function normalizeDefaultForQuestion(q: QuestionLike): unknown | undefined {
  const raw = q.config?.default_value;
  if (raw === undefined || raw === null) return undefined;

  switch (q.type) {
    case 'text':
    case 'qrcode':
    case 'image':
    case 'document':
      return typeof raw === 'string' ? raw : String(raw);
    case 'integer': {
      if (typeof raw === 'number' && Number.isFinite(raw)) return Math.trunc(raw);
      if (typeof raw === 'string') {
        const n = parseInt(raw, 10);
        return Number.isNaN(n) ? undefined : n;
      }
      return undefined;
    }
    case 'float': {
      if (typeof raw === 'number' && Number.isFinite(raw)) return raw;
      if (typeof raw === 'string') {
        const n = parseFloat(raw);
        return Number.isNaN(n) ? undefined : n;
      }
      return undefined;
    }
    case 'boolean':
      if (typeof raw === 'boolean') return raw;
      if (raw === 'true' || raw === true) return true;
      if (raw === 'false' || raw === false) return false;
      return undefined;
    case 'select': {
      if (typeof raw === 'number' && Number.isFinite(raw)) return Math.trunc(raw);
      if (typeof raw === 'string') {
        const n = parseInt(raw, 10);
        return Number.isNaN(n) ? undefined : n;
      }
      return undefined;
    }
    case 'multiselect': {
      if (!Array.isArray(raw)) return undefined;
      return raw
        .map((x) => (typeof x === 'number' ? Math.trunc(x) : parseInt(String(x), 10)))
        .filter((n) => !Number.isNaN(n));
    }
    case 'date':
    case 'datetime':
      return typeof raw === 'string' ? raw : String(raw);
    default:
      return raw;
  }
}

export function setConfigDefaultValue(cfg: QuestionConfig, def: unknown | undefined): QuestionConfig {
  const next = { ...cfg };
  if (def === undefined) {
    delete next.default_value;
  } else {
    next.default_value = def;
  }
  return next;
}
