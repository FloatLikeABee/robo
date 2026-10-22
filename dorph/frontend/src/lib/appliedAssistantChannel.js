const CHANNEL_NAME = 'dorph-ai-applied-assistant';
const LEGACY_CHANNEL_NAME = 'morph-ai-applied-assistant';
export const APPLIED_ASSISTANT_STORAGE_KEY = 'dorph-ai:applied-assistant';

export const APPLIED_ASSISTANT_MSG = {
  APPLY: 'dorph-ai:apply-assistant',
  DISMISS: 'dorph-ai:dismiss-assistant',
  REQUEST: 'dorph-ai:request-applied-assistant',
  STATE: 'dorph-ai:applied-assistant-state',
};

const LEGACY_APPLIED_ASSISTANT_MSG = {
  'morph-ai:apply-assistant': APPLIED_ASSISTANT_MSG.APPLY,
  'morph-ai:dismiss-assistant': APPLIED_ASSISTANT_MSG.DISMISS,
  'morph-ai:request-applied-assistant': APPLIED_ASSISTANT_MSG.REQUEST,
  'morph-ai:applied-assistant-state': APPLIED_ASSISTANT_MSG.STATE,
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

export function dorphAgentId(assistant) {
  const id = String(assistant?.id || '').trim();
  if (!id) return '';
  return id.startsWith('bk:') ? id : `bk:${id}`;
}

export function parseAppliedAssistantMessage(data) {
  if (!data || typeof data !== 'object') return null;
  const type = LEGACY_APPLIED_ASSISTANT_MSG[data.type] || data.type;
  if (!Object.values(APPLIED_ASSISTANT_MSG).includes(type)) return null;
  return { type, assistant: normalizeAssistant(data.assistant) };
}

export function readStoredAppliedAssistant() {
  try {
    const raw = sessionStorage.getItem(APPLIED_ASSISTANT_STORAGE_KEY)
      || sessionStorage.getItem('morph-ai:applied-assistant');
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
      sessionStorage.removeItem('morph-ai:applied-assistant');
      return;
    }
    sessionStorage.setItem(APPLIED_ASSISTANT_STORAGE_KEY, JSON.stringify(next));
    sessionStorage.removeItem('morph-ai:applied-assistant');
  } catch {
    /* ignore quota / private mode */
  }
}

function openChannel(name) {
  if (typeof BroadcastChannel === 'undefined') return null;
  try {
    return new BroadcastChannel(name);
  } catch {
    return null;
  }
}

function getChannels() {
  return [CHANNEL_NAME, LEGACY_CHANNEL_NAME].map(openChannel).filter(Boolean);
}

export function postAppliedAssistantChannel(payload) {
  for (const ch of getChannels()) {
    try {
      ch.postMessage(payload);
    } finally {
      ch.close();
    }
  }
}

export function subscribeAppliedAssistantChannel(handler) {
  const channels = getChannels();
  if (!channels.length) return () => {};
  const onMessage = (event) => handler(event.data);
  channels.forEach((ch) => ch.addEventListener('message', onMessage));
  return () => {
    channels.forEach((ch) => {
      ch.removeEventListener('message', onMessage);
      ch.close();
    });
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
  return /unknown or unavailable (ai tools|dorphpools) assistant/i.test(String(text || ''));
}
