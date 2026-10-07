import type { Config, Context } from '@netlify/edge-functions';

/** Runtime proxy so MORPH_API_ORIGIN works without a rebuild (build-time _redirects is optional). */
export default async (request: Request, _context: Context) => {
  const origin = (Netlify.env.get('MORPH_API_ORIGIN') || '').trim().replace(/\/$/, '');
  if (!origin) {
    return new Response(
      JSON.stringify({
        error:
          'Morph API is not configured. Set MORPH_API_ORIGIN in Netlify (Site configuration → Environment variables) to your Morph API public URL, then retry.',
      }),
      { status: 503, headers: { 'Content-Type': 'application/json' } },
    );
  }

  const incoming = new URL(request.url);
  const target = `${origin}${incoming.pathname}${incoming.search}`;

  const headers = new Headers(request.headers);
  headers.delete('host');

  const init: RequestInit = {
    method: request.method,
    headers,
    redirect: 'manual',
  };
  if (request.method !== 'GET' && request.method !== 'HEAD') {
    init.body = request.body;
  }

  const upstream = await fetch(target, init);
  const outHeaders = new Headers(upstream.headers);
  outHeaders.delete('content-encoding');
  return new Response(upstream.body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers: outHeaders,
  });
};

export const config: Config = {
  path: '/api/*',
};
