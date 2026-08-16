import { Platform } from 'react-native';

const DEFAULT_API_PORT = '8978';

function defaultApiHost() {
  return Platform.OS === 'android' ? '10.0.2.2' : 'localhost';
}

/** Dev default: emulator loopback. On a physical device, point Metro at your machine or tunnel and set host here. */
export function getAcademiApiBaseUrl(host = defaultApiHost(), port = DEFAULT_API_PORT) {
  return `http://${host}:${port}/api/v1`;
}

let cachedToken = null;

export function getAcademiToken() {
  return cachedToken;
}

export async function ensureAcademiSession(apiBase) {
  if (cachedToken) return cachedToken;
  const res = await fetch(`${apiBase}/auth/mock`, { method: 'POST' });
  if (!res.ok) throw new Error('Could not start demo session');
  const data = await res.json();
  cachedToken = data.token || null;
  return cachedToken;
}

export function authHeaders(token, extra = {}) {
  const h = { Accept: 'application/json', ...extra };
  if (token) h.Authorization = `Bearer ${token}`;
  return h;
}

async function readErrorMessage(res) {
  const text = await res.text();
  try {
    const j = JSON.parse(text);
    if (j && typeof j.error === 'string' && j.error) return j.error;
  } catch (_) {
    /* not JSON */
  }
  const hint = text?.trim() ? text.trim().slice(0, 180) : '';
  return hint || `Request failed (${res.status})`;
}

export async function fetchDocsBrief(apiBase, token) {
  const res = await fetch(`${apiBase}/docs?brief=1`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json();
}

export async function postLearnAnalysis(apiBase, token, docId, { disableResearch }) {
  const res = await fetch(`${apiBase}/ai/learn`, {
    method: 'POST',
    headers: authHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify({
      doc_id: docId,
      disable_research: disableResearch,
    }),
  });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json();
}

export async function createMarkdownDoc(apiBase, token, { title, content, tags }) {
  const res = await fetch(`${apiBase}/docs`, {
    method: 'POST',
    headers: authHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify({
      title,
      type: 'markdown',
      content,
      tags: tags || ['#analysis', '#help-you-learn'],
    }),
  });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json();
}

/** Full document record (includes content and file metadata). */
export async function fetchDocById(apiBase, token, docId) {
  const res = await fetch(`${apiBase}/docs/${encodeURIComponent(docId)}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json();
}

export async function fetchCommunityPosts(apiBase, token) {
  const res = await fetch(`${apiBase}/community/posts`, {
    headers: authHeaders(token),
  });
  if (!res.ok) throw new Error(await readErrorMessage(res));
  return res.json();
}
