import type { UtilsModuleId } from './config';

export const PROBED_EMBED_IDS: UtilsModuleId[] = ['datax', 'projects'];

export function embedStartCommand(id: UtilsModuleId): string {
  if (id === 'datax') return './start-all.sh start sharpreport-ui';
  if (id === 'projects') return './start-all.sh start morph-engi-ui';
  return './start-all.sh start morph-utils';
}

/** True when TCP to the origin succeeds. `no-cors` opaque responses still count as up. */
export async function probeEmbedOrigin(embedUrl: string): Promise<boolean> {
  let origin: string;
  try {
    origin = new URL(embedUrl).origin;
  } catch {
    return false;
  }
  try {
    await fetch(origin, { method: 'GET', mode: 'no-cors', cache: 'no-store' });
    return true;
  } catch {
    return false;
  }
}
