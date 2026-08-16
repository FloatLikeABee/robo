import React from 'react';
import { Navigate } from 'react-router-dom';
import { ADMIN_BASE_PATH } from '../../adminPaths';

/** Legacy `/forms/*` URLs redirect into MorphData Big notes. */
export default function FormsApp() {
  return <Navigate to={`${ADMIN_BASE_PATH}/big-notes`} replace />;
}
