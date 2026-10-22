import { en, zh } from './messages';

export const LOCALE_STORAGE_KEY = 'dorph-locale';
export const LOCALE_MESSAGE = 'dorph-locale';

const catalogs = { en, zh };
let current = '';
const listeners = new Set();

export function normalizeLocale(value) {
  const v = String(value || '').trim().toLowerCase();
  if (v === 'en' || v === 'zh') return v;
  return '';
}

export function detectLocale(languages) {
  const list = Array.isArray(languages) && languages.length ? languages : ['en'];
  for (const tag of list) {
    if (String(tag || '').toLowerCase().startsWith('zh')) return 'zh';
  }
  return 'en';
}

export function resolveLocale() {
  return 'en';
}

export function documentLang(locale) {
  return locale === 'zh' ? 'zh-Hans' : 'en';
}

export function applyDocumentLang(locale) {
  if (typeof document === 'undefined') return;
  document.documentElement.lang = documentLang(normalizeLocale(locale) || 'en');
}

export function getLocale() {
  if (!current) {
    try {
      localStorage.removeItem(LOCALE_STORAGE_KEY);
      localStorage.removeItem('morph-locale');
    } catch {
      /* private mode */
    }
    current = 'en';
    applyDocumentLang('en');
  }
  return 'en';
}

function publish(locale) {
  current = locale;
  applyDocumentLang(locale);
  listeners.forEach((fn) => fn());
}

export function setLocale() {
  try {
    localStorage.removeItem(LOCALE_STORAGE_KEY);
    localStorage.removeItem('morph-locale');
  } catch {
    /* private mode */
  }
  publish('en');
}

export function subscribeLocale(fn) {
  listeners.add(fn);
  return () => {
    listeners.delete(fn);
  };
}

export function translate(locale, key, vars = undefined) {
  const lang = normalizeLocale(locale) === 'zh' ? 'zh' : 'en';
  let text = catalogs[lang][key] || catalogs.en[key] || key;
  if (vars) {
    for (const [name, value] of Object.entries(vars)) {
      text = text.split(`{${name}}`).join(String(value ?? ''));
    }
  }
  return text;
}

export function installLocaleListener() {
  if (typeof window === 'undefined') return;
  window.addEventListener('message', (event) => {
    const data = event.data;
    if (!data || data.type !== LOCALE_MESSAGE) return;
    setLocale(data.lang);
  });
  window.addEventListener('storage', (event) => {
    if (event.key !== LOCALE_STORAGE_KEY) return;
    const locale = normalizeLocale(event.newValue);
    if (!locale || locale === current) return;
    publish(locale);
  });
}

export function postLocaleToFrame(iframe, locale) {
  const win = iframe?.contentWindow;
  if (!win) return;
  const lang = normalizeLocale(locale) || getLocale();
  try {
    win.postMessage({ type: LOCALE_MESSAGE, lang }, '*');
  } catch {
    /* frame not ready */
  }
}

export function bindFrameLocale(iframe) {
  const send = () => postLocaleToFrame(iframe, getLocale());
  const onChange = () => send();
  window.addEventListener('dorph-locale-changed', onChange);
  send();
  return () => window.removeEventListener('dorph-locale-changed', onChange);
}

export function withLang(url, locale) {
  if (!url) return url;
  try {
    const next = new URL(url, typeof window !== 'undefined' ? window.location.href : 'http://127.0.0.1');
    next.searchParams.set('lang', normalizeLocale(locale) || getLocale());
    return next.toString();
  } catch {
    return url;
  }
}
