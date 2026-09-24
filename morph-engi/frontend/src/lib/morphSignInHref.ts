function loopback(raw: string): boolean {
  try {
    const host = new URL(raw).hostname.toLowerCase().replace(/^\[|\]$/g, '');
    return host === 'localhost' || host === '::1' || host.startsWith('127.');
  } catch {
    return true;
  }
}

/** Morph sign-in origin. Production never falls back to localhost. */
export function morphSignInHref(configured: string | undefined, dev: boolean): string {
  const trimmed = (configured ?? '').trim().replace(/\/$/, '');
  if (!trimmed) return dev ? 'http://localhost:3031' : '';
  if (!dev && loopback(trimmed)) return '';
  return trimmed;
}

export function missingMorphSessionReason(dev: boolean): string {
  if (dev) {
    return 'No Morph AI session found. Sign in at http://localhost:3031, then open Projects from MorphUtils or reload.';
  }
  return 'No Morph session found. Sign in on Morph, then open Project from MorphUtils.';
}
