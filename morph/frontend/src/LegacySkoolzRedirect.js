import { Navigate, useLocation } from 'react-router-dom';
import { ADMIN_BASE_PATH, remapLegacyAdminRest } from './adminPaths';

/** Old bookmarks: `/skoolz/*` → `/morphdata/*` */
export default function LegacySkoolzRedirect() {
  const { pathname, search, hash } = useLocation();
  const rest = pathname.replace(/^\/skoolz/, '') || '/';
  const target = `${ADMIN_BASE_PATH}${remapLegacyAdminRest(rest)}`;
  return <Navigate to={`${target}${search}${hash}`} replace />;
}
