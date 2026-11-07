import { API_BASE_URL } from '../apiBase';
import { getDorphToken } from '../auth/dorphSession';

function parseSSEBlock(block) {
  let event = 'message';
  const dataLines = [];
  for (const line of block.split('\n')) {
    if (line.startsWith('event:')) {
      event = line.slice(6).trim();
    } else if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).replace(/^\s/, ''));
    }
  }
  return { event, data: dataLines.join('\n') };
}

/**
 * POST /api/chat with stream:true. Invokes onReasoning/onContent during SSE; resolves with final ChatResponse from the done event.
 */
export async function postChatStream(body, { signal, onReasoning, onContent } = {}) {
  const headers = { 'Content-Type': 'application/json' };
  const token = getDorphToken();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE_URL}/api/chat`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ ...body, stream: true }),
    signal,
  });

  if (!res.ok) {
    let errText = await res.text();
    try {
      const parsed = JSON.parse(errText);
      errText = parsed.error || errText;
    } catch {
      /* keep text */
    }
    throw new Error(errText || `HTTP ${res.status}`);
  }

  const contentType = res.headers.get('content-type') || '';
  if (!contentType.includes('text/event-stream') || !res.body) {
    return res.json();
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  let donePayload = null;

  while (true) {
    const { value, done } = await reader.read();
    if (done) {
      break;
    }
    buffer += decoder.decode(value, { stream: true });
    let sep;
    while ((sep = buffer.indexOf('\n\n')) !== -1) {
      const block = buffer.slice(0, sep);
      buffer = buffer.slice(sep + 2);
      if (!block.trim()) {
        continue;
      }
      const { event, data } = parseSSEBlock(block);
      if (event === 'reasoning' && onReasoning) {
        onReasoning(data);
      } else if (event === 'content' && onContent) {
        onContent(data);
      } else if (event === 'error') {
        throw new Error(data || 'stream error');
      } else if (event === 'done') {
        try {
          donePayload = JSON.parse(data);
        } catch {
          throw new Error('invalid done payload');
        }
      }
    }
  }

  if (!donePayload) {
    throw new Error('stream ended without done event');
  }
  return donePayload;
}
