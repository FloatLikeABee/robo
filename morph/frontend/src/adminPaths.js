/** Canonical browser base for the Tran admin SPA. */
export const ADMIN_BASE_PATH = '/morphdata';

/**
 * Remap first path segment for bookmarks that used older admin URLs.
 * @param {string} rest Path after the mount prefix (e.g. `/members`, `/`, ``).
 */
export function remapLegacyAdminRest(rest) {
  if (rest === '/' || rest === '') return '/members';
  const r = rest.startsWith('/') ? rest : `/${rest}`;
  return r
    .replace(/^\/students(\/|$)/, '/members$1')
    .replace(/^\/staff(\/|$)/, '/employees$1')
    .replace(/^\/vehicles(\/|$)/, '/assets$1')
    .replace(/^\/trips(\/|$)/, '/activities$1')
    .replace(/^\/districts-schools(\/|$)/, '/facilities$1');
}

/** Base path for admin routes. */
export function useAdminBasePath() {
  return ADMIN_BASE_PATH;
}
