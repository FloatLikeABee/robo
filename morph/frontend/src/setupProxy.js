/**
 * CRA's package.json "proxy" skips requests with Accept: text/html (SPA fallback).
 * That breaks public Big notes opened in a browser tab — they get Morph AI instead
 * of the published HTML. Always proxy /api to the Morph server.
 */
const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function setupProxy(app) {
  const target = process.env.MORPH_API_PROXY || 'http://localhost:9090';
  app.use(
    '/api',
    createProxyMiddleware({
      target,
      changeOrigin: true,
      // Do not bypass HTML navigations — public big-note pages are text/html.
    }),
  );
};
