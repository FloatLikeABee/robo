import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { getMorphToken } from '../auth/morphSession';

export default function ProtectedLayout() {
  const token = getMorphToken();
  const location = useLocation();
  if (!token) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }
  return <Outlet />;
}
