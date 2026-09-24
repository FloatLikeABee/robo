export const HEADER_APP_ICONS = {
  morphdata: `${process.env.PUBLIC_URL || ''}/icons/morph-data-icon.svg`,
  morphutils: `${process.env.PUBLIC_URL || ''}/icons/morph-utils-icon.svg`,
  bk: `${process.env.PUBLIC_URL || ''}/icons/bk-icon.svg`,
};

// Production builds omit MorphUtils when the URL is missing or loopback.
// CRA inlines nodeEnv and configuredUrl only when the call site reads
// process.env.NODE_ENV and process.env.REACT_APP_MORPH_UTILS_URL directly.
export function morphUtilsBaseURL(nodeEnv, configuredUrl) {
  const configured = String(configuredUrl || '').trim();
  const url = configured || 'http://localhost:3040';
  if (nodeEnv === 'production' && isLoopbackURL(url)) return '';
  return url;
}

function isLoopbackURL(raw) {
  let u;
  try {
    u = new URL(raw, 'https://morph.invalid');
  } catch {
    return true;
  }
  if (u.origin === 'https://morph.invalid') return false;
  const host = u.hostname.toLowerCase().replace(/^\[|\]$/g, '');
  if (!host) return true;
  return host === 'localhost' || host === '::1' || host.startsWith('127.');
}
