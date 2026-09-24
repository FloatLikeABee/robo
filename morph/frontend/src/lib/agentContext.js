export function contextFingerprint({ includeNotes, includeKnowledge }) {
  return [
    '0',
    includeNotes ? '1' : '0',
    includeKnowledge ? '1' : '0',
    '',
  ].join('|');
}

export function inferSubAgents(text, { hasNotes = false, hasKnowledge = false } = {}) {
  const trimmed = String(text || '').trim();
  if (!trimmed) return [];
  const low = trimmed.toLowerCase();
  const simple = ['hi', 'hello', 'hey', 'thanks', 'thank you', 'ok', 'okay', 'yes', 'no'];
  if (simple.includes(low) || ([...trimmed].length <= 24 && !trimmed.includes('?') && trimmed.split(/\s+/).length <= 4)) {
    return [];
  }
  const workers = [];
  const docQ = /summar|what does|in this|according to/.test(low);
  if (hasKnowledge && (/knowledge|hybrid|library/.test(low) || docQ)) workers.push('knowledge');
  if (hasNotes && /note|todo|task/.test(low)) workers.push('notes');
  if (/member|employee|facility|district|case|asset|contact|form|event|list my|show my/.test(low)) {
    workers.push('morphdata');
  }
  return workers.slice(0, 3);
}
