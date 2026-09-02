export async function sha256Hex(text) {
  const data = new TextEncoder().encode(String(text ?? ''));
  const buf = await crypto.subtle.digest('SHA-256', data);
  return Array.from(new Uint8Array(buf))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('');
}

export function contextFingerprint({ includeFiles, includeNotes, includeKnowledge, pinHashes }) {
  const pins = Array.isArray(pinHashes) ? [...pinHashes].sort() : [];
  return [
    includeFiles ? '1' : '0',
    includeNotes ? '1' : '0',
    includeKnowledge ? '1' : '0',
    pins.join(','),
  ].join('|');
}

export function inferSubAgents(text, { hasFiles = false, hasNotes = false, hasKnowledge = false } = {}) {
  const trimmed = String(text || '').trim();
  if (!trimmed) return [];
  const low = trimmed.toLowerCase();
  const simple = ['hi', 'hello', 'hey', 'thanks', 'thank you', 'ok', 'okay', 'yes', 'no'];
  if (simple.includes(low) || ([...trimmed].length <= 24 && !trimmed.includes('?') && trimmed.split(/\s+/).length <= 4)) {
    return [];
  }
  const workers = [];
  const docQ = /summar|what does|in this|according to/.test(low);
  if (hasFiles && (/file|folder|workspace/.test(low) || docQ)) workers.push('files');
  if (hasKnowledge && (/knowledge|hybrid|library/.test(low) || docQ)) workers.push('knowledge');
  if (hasNotes && /note|todo|task/.test(low)) workers.push('notes');
  if (/member|employee|facility|district|case|asset|contact|form|event|list my|show my/.test(low)) {
    workers.push('morphdata');
  }
  return workers.slice(0, 3);
}

export async function readPinnedFileBodies(files, pinnedPaths) {
  const pinned = pinnedPaths instanceof Set ? pinnedPaths : new Set(pinnedPaths || []);
  const out = [];
  for (const f of files || []) {
    if (!f || f.skipped || !pinned.has(f.path)) continue;
    let text = '';
    try {
      if (f.handle && typeof f.handle.getFile === 'function') {
        const file = await f.handle.getFile();
        text = await file.text();
      } else if (f.file && typeof f.file.text === 'function') {
        text = await f.file.text();
      }
    } catch {
      text = '';
    }
    const hash = await sha256Hex(text);
    out.push({ path: f.path, hash, content: text });
  }
  return out;
}
