function isLoopback(raw) {
  try {
    const host = new URL(raw).hostname.toLowerCase().replace(/^\[|\]$/g, '');
    return host === 'localhost' || host === '::1' || host.startsWith('127.');
  } catch {
    return true;
  }
}

export const MORPH_NOTES_EMBEDS = [
  {
    id: 'event-logs',
    label: 'Event Logs',
    path: '/events-info',
    start: './start-all.sh start formx-ui',
    devOrigin: 'http://localhost:19909',
  },
  {
    id: 'content-maker',
    label: 'Content Maker',
    path: '',
    start: './start-all.sh start composerx-ui',
    devOrigin: 'http://localhost:8044',
  },
  {
    id: 'project',
    label: 'Project',
    path: '',
    start: './start-all.sh start morph-engi-ui',
    devOrigin: 'http://localhost:5179',
  },
];

export function resolveEmbedUrl({ id, nodeEnv, configured }) {
  const spec = MORPH_NOTES_EMBEDS.find((item) => item.id === id);
  if (!spec) return '';
  let origin = String(configured || '').trim().replace(/\/$/, '');
  if (!origin) {
    if (nodeEnv === 'production') return '';
    origin = spec.devOrigin;
  } else if (nodeEnv === 'production' && isLoopback(origin)) {
    return '';
  }
  if (!origin) return '';
  return spec.path ? `${origin}${spec.path}` : origin;
}

export function withSessionToken(url, token) {
  if (!url || !token) return url;
  const next = new URL(url);
  next.searchParams.set('userspanel_token', token);
  return next.toString();
}
