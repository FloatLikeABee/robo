import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { getMorphToken } from '../auth/morphSession';
import { safeReturnPath } from '../auth/returnTo';

export default function ProtectedLayout() {
  const token = getMorphToken();
  const location = useLocation();
  if (!token) {
    const returnTo = safeReturnPath(location.pathname + location.search + location.hash) || '/';
    return <Navigate to={`/login?returnTo=${encodeURIComponent(returnTo)}`} replace />;
  }
  return <Outlet />;
}
