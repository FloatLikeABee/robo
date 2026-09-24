export type RuntimeConfig = {
  morphApiUrl?: string;
  usersPanelApiUrl?: string;
  sheetxUrl?: string;
  formsxUrl?: string;
  composerxUrl?: string;
  dataxUrl?: string;
  projectsUrl?: string;
  morphEngiUrl?: string;
  morphAiUrl?: string;
};

declare global {
  interface Window {
    __MORPH_UTILS_CONFIG__?: RuntimeConfig;
  }
}

export function readRuntimeConfig(): RuntimeConfig {
  if (typeof window === 'undefined') return {};
  return window.__MORPH_UTILS_CONFIG__ ?? {};
}

export function sheetxEmbedUrl(origin: string): string {
  if (!origin) return '';
  return `${origin}/events-info`;
}

function nonempty(value: string | undefined): string {
  return (value ?? '').trim().replace(/\/$/, '');
}

export function isLoopbackUrl(raw: string): boolean {
  const trimmed = raw.trim();
  if (!trimmed) return false;
  try {
    const host = new URL(trimmed).hostname.toLowerCase().replace(/^\[|\]$/g, '');
    return host === 'localhost' || host === '::1' || host.startsWith('127.');
  } catch {
    return false;
  }
}

/** Morph sign-in origin. Loopback is kept only for local dev. */
export function morphLoginHref(input: { morphAi?: string; morphApi?: string; dev: boolean }): string {
  for (const candidate of [input.morphAi, input.morphApi]) {
    const trimmed = nonempty(candidate);
    if (!trimmed) continue;
    if (!input.dev && isLoopbackUrl(trimmed)) continue;
    return trimmed;
  }
  return '';
}

export function appendSessionToken(baseUrl: string, token: string): string {
  if (!baseUrl || !token) return baseUrl;
  const url = new URL(baseUrl);
  url.searchParams.set('userspanel_token', token);
  return url.toString();
}

/** Host-only session cookie. No Domain: *.onrender.com is a public suffix. */
export function sessionCookieAttributes(maxAgeSeconds: number): string {
  return `Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`;
}

export function usesLocalStartHint(embedUrl: string): boolean {
  return isLoopbackUrl(embedUrl);
}

/** First non-empty of runtime, runtime alias, build-time, build-time alias. Dev fallback only when all are blank. */
export function resolvePublicUrl(input: {
  runtime?: string;
  runtimeAlias?: string;
  built?: string;
  builtAlias?: string;
  devFallback?: string;
  dev: boolean;
}): string {
  for (const candidate of [input.runtime, input.runtimeAlias, input.built, input.builtAlias]) {
    const trimmed = nonempty(candidate);
    if (trimmed) return trimmed;
  }
  if (input.dev) return nonempty(input.devFallback);
  return '';
}
