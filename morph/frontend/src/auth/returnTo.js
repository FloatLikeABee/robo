const SAFE_ORIGIN = 'http://morph.local';

function hasControlChar(value) {
  for (let i = 0; i < value.length; i += 1) {
    if (value.charCodeAt(i) <= 31) return true;
  }
  return false;
}

/** Same-origin relative path, or '' when the value could leave this site. */
export function safeReturnPath(raw) {
  if (typeof raw !== 'string') return '';
  const value = raw.trim();
  if (!value.startsWith('/') || value.startsWith('//') || value.startsWith('/\\')) return '';
  if (value.includes('\\') || value.includes('://') || hasControlChar(value)) return '';
  let url;
  try {
    url = new URL(value, SAFE_ORIGIN);
  } catch {
    return '';
  }
  if (url.origin !== SAFE_ORIGIN) return '';
  const path = url.pathname + url.search + url.hash;
  if (!path.startsWith('/') || path.startsWith('//')) return '';
  return path;
}
