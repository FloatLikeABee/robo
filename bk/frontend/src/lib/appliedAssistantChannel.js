const CHANNEL_NAME = 'dorph-ai-applied-assistant';
const LEGACY_CHANNEL_NAME = 'morph-ai-applied-assistant';

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

export function isFramedInParent() {
  try {
    return Boolean(window.parent && window.parent !== window);
  } catch {
    return true;
  }
}

export function dorphTargetOrigin() {
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
  const type = LEGACY_APPLIED_ASSISTANT_MSG[data.type] || data.type;
  if (!Object.values(APPLIED_ASSISTANT_MSG).includes(type)) return null;
  return { type, assistant: normalizeAssistant(data.assistant) };
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

export function postToDorph(payload) {
  postAppliedAssistantChannel(payload);
  const origin = dorphTargetOrigin();
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

function isStateFromDorph(event) {
  if (event.source === window) return false;
  const origin = dorphTargetOrigin();
  if (origin === '*') return true;
  return event.origin === origin;
}

/** Resolves with the assistant (or null) when Dorph replies with matching state; undefined on timeout. */
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
      if (!isStateFromDorph(event)) return;
      consider(event.data);
    };
    window.addEventListener('message', onWindow);
    const unsub = subscribeAppliedAssistantChannel(consider);
    const timer = setTimeout(() => finish(undefined), timeoutMs);
  });
}
