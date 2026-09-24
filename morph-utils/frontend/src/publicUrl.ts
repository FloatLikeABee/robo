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
