import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { getDorphToken } from '../auth/dorphSession';
import { safeReturnPath } from '../auth/returnTo';

export default function ProtectedLayout() {
  const token = getDorphToken();
  const location = useLocation();
  if (!token) {
    const returnTo = safeReturnPath(location.pathname + location.search + location.hash) || '/';
    return <Navigate to={`/login?returnTo=${encodeURIComponent(returnTo)}`} replace />;
  }
  return <Outlet />;
}
