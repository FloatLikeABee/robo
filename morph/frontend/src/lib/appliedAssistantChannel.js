const CHANNEL_NAME = 'morph-ai-applied-assistant';
export const APPLIED_ASSISTANT_STORAGE_KEY = 'morph-ai:applied-assistant';

export const APPLIED_ASSISTANT_MSG = {
  APPLY: 'morph-ai:apply-assistant',
  DISMISS: 'morph-ai:dismiss-assistant',
  REQUEST: 'morph-ai:request-applied-assistant',
  STATE: 'morph-ai:applied-assistant-state',
};

export function bkOrigin() {
  try {
    return new URL(process.env.REACT_APP_BK_URL || 'http://localhost:3000', window.location.origin).origin;
  } catch {
    return 'http://localhost:3000';
  }
}

export function isAllowedBkOrigin(origin) {
  return Boolean(origin) && origin === bkOrigin();
}

export function normalizeAssistant(value) {
  if (!value || typeof value !== 'object') return null;
  const id = String(value.id || '').trim();
  if (!id) return null;
  const name = String(value.name || '').trim() || id;
  return { id, name };
}

export function morphAgentId(assistant) {
  const id = String(assistant?.id || '').trim();
  if (!id) return '';
  return id.startsWith('bk:') ? id : `bk:${id}`;
}

export function parseAppliedAssistantMessage(data) {
  if (!data || typeof data !== 'object') return null;
  const type = data.type;
  if (!Object.values(APPLIED_ASSISTANT_MSG).includes(type)) return null;
  return { type, assistant: normalizeAssistant(data.assistant) };
}

export function readStoredAppliedAssistant() {
  try {
    const raw = sessionStorage.getItem(APPLIED_ASSISTANT_STORAGE_KEY);
    if (!raw) return null;
    return normalizeAssistant(JSON.parse(raw));
  } catch {
    return null;
  }
}

export function writeStoredAppliedAssistant(assistant) {
  try {
    const next = normalizeAssistant(assistant);
    if (!next) {
      sessionStorage.removeItem(APPLIED_ASSISTANT_STORAGE_KEY);
      return;
    }
    sessionStorage.setItem(APPLIED_ASSISTANT_STORAGE_KEY, JSON.stringify(next));
  } catch {
    /* ignore quota / private mode */
  }
}

function getChannel() {
  if (typeof BroadcastChannel === 'undefined') return null;
  try {
    return new BroadcastChannel(CHANNEL_NAME);
  } catch {
    return null;
  }
}

export function postAppliedAssistantChannel(payload) {
  const ch = getChannel();
  if (!ch) return;
  try {
    ch.postMessage(payload);
  } finally {
    ch.close();
  }
}

export function subscribeAppliedAssistantChannel(handler) {
  const ch = getChannel();
  if (!ch) return () => {};
  const onMessage = (event) => handler(event.data);
  ch.addEventListener('message', onMessage);
  return () => {
    ch.removeEventListener('message', onMessage);
    ch.close();
  };
}

export function postStateToIframe(iframe, assistant) {
  const win = iframe?.contentWindow;
  if (!win) return;
  try {
    win.postMessage({ type: APPLIED_ASSISTANT_MSG.STATE, assistant: normalizeAssistant(assistant) }, bkOrigin());
  } catch {
    /* iframe not ready */
  }
}

export function isUnknownAiToolsAssistantError(text) {
  return /unknown or unavailable ai tools assistant/i.test(String(text || ''));
}
