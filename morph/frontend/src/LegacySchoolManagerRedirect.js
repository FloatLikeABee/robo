import { Navigate, useLocation } from 'react-router-dom';
import { ADMIN_BASE_PATH, remapLegacyAdminRest } from './adminPaths';

/** Old bookmarks: `/schoolmanager/*` → `/morphdata/*` */
export default function LegacySchoolManagerRedirect() {
  const { pathname, search, hash } = useLocation();
  const rest = pathname.replace(/^\/schoolmanager/, '') || '/';
  const target = `${ADMIN_BASE_PATH}${remapLegacyAdminRest(rest)}`;
  return <Navigate to={`${target}${search}${hash}`} replace />;
}
