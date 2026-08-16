import { Navigate, useLocation } from 'react-router-dom';
import { ADMIN_BASE_PATH, remapLegacyAdminRest } from './adminPaths';

/** Maps old /transfinderx/* URLs to /morphdata/* (and legacy segments → members, facilities, …). */
export default function LegacyAdminRedirect() {
  const { pathname, search, hash } = useLocation();
  const rest = pathname.replace(/^\/transfinderx/, '') || '/';
  const target = `${ADMIN_BASE_PATH}${remapLegacyAdminRest(rest)}`;
  return <Navigate to={`${target}${search}${hash}`} replace />;
}
