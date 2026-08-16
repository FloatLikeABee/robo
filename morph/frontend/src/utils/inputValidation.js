const BLOCKED_TEXT_CHARS_REGEX = /[<>"'`\\]/g;

export function sanitizeTextInput(value) {
  const src = value == null ? '' : String(value);
  return src.replace(BLOCKED_TEXT_CHARS_REGEX, '');
}

export function sanitizeNumericInput(value, { allowDecimal = false, allowNegative = false } = {}) {
  const src = value == null ? '' : String(value);
  const bodyRegex = allowDecimal ? /[^0-9.]/g : /[^0-9]/g;
  let out = src.replace(bodyRegex, '');
  if (allowDecimal) {
    const firstDot = out.indexOf('.');
    if (firstDot >= 0) {
      out = out.slice(0, firstDot + 1) + out.slice(firstDot + 1).replace(/\./g, '');
    }
  }
  if (allowNegative) {
    out = out.replace(/-/g, '');
    if (src.trim().startsWith('-') && out !== '') {
      out = `-${out}`;
    }
  }
  return out;
}

export function sanitizeMaybeNumber(value, numericOptions) {
  if (value == null) return value;
  if (typeof value !== 'string') return value;
  return sanitizeNumericInput(value, numericOptions);
}
