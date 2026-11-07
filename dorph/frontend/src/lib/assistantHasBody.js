export function assistantHasBody(msg) {
  if (!msg || msg.type !== 'assistant') return false;
  return Boolean(String(msg.content || '').trim() || String(msg.reasoning || '').trim());
}
