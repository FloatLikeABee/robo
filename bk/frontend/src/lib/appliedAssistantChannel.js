const CHANNEL_NAME = 'morph-ai-applied-assistant';

export const APPLIED_ASSISTANT_MSG = {
  APPLY: 'morph-ai:apply-assistant',
  DISMISS: 'morph-ai:dismiss-assistant',
  REQUEST: 'morph-ai:request-applied-assistant',
  STATE: 'morph-ai:applied-assistant-state',
};

function normalizeAssistant(value) {
  if (!value || typeof value !== 'object') return null;
  const id = String(value.id || '').trim();
  if (!id) return null;
  const name = String(value.name || '').trim() || id;
  return { id, name };
}

export function parseAppliedAssistantMessage(data) {
  if (!data || typeof data !== 'object') return null;
  const type = data.type;
  if (!Object.values(APPLIED_ASSISTANT_MSG).includes(type)) return null;
  return { type, assistant: normalizeAssistant(data.assistant) };
}

export function isFramedInParent() {
  try {
    return window.parent && window.parent !== window;
  } catch {
    return false;
  }
}

function parentOrigin() {
  try {
    if (document.referrer) return new URL(document.referrer).origin;
  } catch {
    /* referrer is not a URL */
  }
  return '*';
}

export function postToMorph(payload) {
  if (!isFramedInParent()) return;
  try {
    window.parent.postMessage(payload, parentOrigin());
  } catch {
    /* parent unavailable */
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

export function waitForAppliedState(predicate, timeoutMs = 2000) {
  return new Promise((resolve) => {
    let settled = false;
    const finish = (value) => {
      if (settled) return;
      settled = true;
      window.removeEventListener('message', onWindow);
      unsub();
      clearTimeout(timer);
      resolve(value);
    };
    const consider = (data) => {
      const parsed = parseAppliedAssistantMessage(data);
      if (!parsed || parsed.type !== APPLIED_ASSISTANT_MSG.STATE) return;
      if (predicate(parsed.assistant)) finish(parsed.assistant);
    };
    const onWindow = (event) => consider(event.data);
    window.addEventListener('message', onWindow);
    const unsub = subscribeAppliedAssistantChannel(consider);
    const timer = setTimeout(() => finish(undefined), timeoutMs);
  });
}
