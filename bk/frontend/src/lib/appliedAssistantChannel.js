const CHANNEL_NAME = 'morph-ai-applied-assistant';

export const APPLIED_ASSISTANT_MSG = {
  APPLY: 'morph-ai:apply-assistant',
  DISMISS: 'morph-ai:dismiss-assistant',
  REQUEST: 'morph-ai:request-applied-assistant',
  STATE: 'morph-ai:applied-assistant-state',
};

export function isFramedInParent() {
  try {
    return Boolean(window.parent && window.parent !== window);
  } catch {
    return true;
  }
}

export function morphTargetOrigin() {
  try {
    if (document.referrer) return new URL(document.referrer).origin;
  } catch {
    /* ignore */
  }
  try {
    const ancestor = window.location.ancestorOrigins && window.location.ancestorOrigins[0];
    if (ancestor) return ancestor;
  } catch {
    /* ignore */
  }
  return '*';
}

export function normalizeAssistant(value) {
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

export function postToMorph(payload) {
  postAppliedAssistantChannel(payload);
  const origin = morphTargetOrigin();
  if (isFramedInParent()) {
    try {
      window.parent.postMessage(payload, origin);
    } catch {
      /* ignore */
    }
  }
  if (window.opener && !window.opener.closed) {
    try {
      window.opener.postMessage(payload, origin);
    } catch {
      /* ignore */
    }
  }
}

function isStateFromMorph(event) {
  if (event.source === window) return false;
  const origin = morphTargetOrigin();
  if (origin === '*') return true;
  return event.origin === origin;
}

/** Resolves with the assistant (or null) when Morph replies with matching state; undefined on timeout. */
export function waitForAppliedState(predicate, timeoutMs = 900) {
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
      if (parsed?.type !== APPLIED_ASSISTANT_MSG.STATE) return;
      if (predicate(parsed.assistant)) finish(parsed.assistant);
    };
    const onWindow = (event) => {
      if (!isStateFromMorph(event)) return;
      consider(event.data);
    };
    window.addEventListener('message', onWindow);
    const unsub = subscribeAppliedAssistantChannel(consider);
    const timer = setTimeout(() => finish(undefined), timeoutMs);
  });
}
