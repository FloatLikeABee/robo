import React, { useLayoutEffect } from 'react';
import { createBrowserRouter, Navigate, Outlet, useLocation } from 'react-router-dom';
import { releaseStuckOverlays } from './utils/releaseStuckOverlays';
import App from './App';
import FormsApp from './components/Forms/FormsApp';
import AdminLayout from './AdminLayout';
import LegacyAdminRedirect from './LegacyAdminRedirect';
import LegacySchoolManagerRedirect from './LegacySchoolManagerRedirect';
import LegacySkoolzRedirect from './LegacySkoolzRedirect';
import ProtectedLayout from './components/ProtectedLayout';
import LoginPage from './pages/LoginPage';
import SkillsPage from './pages/SkillsPage';
import { ADMIN_BASE_PATH } from './adminPaths';
import DistrictsSchools from './pages/admin/DistrictsSchools';
import CaseTasks from './pages/admin/CaseTasks';
import Timelines from './pages/admin/Timelines';
import GenericData from './pages/admin/GenericData';
import BigNotes from './pages/admin/BigNotes';
import Research from './pages/admin/Research';
import DocumentBranding from './DocumentBranding';

function RootLayout() {
  const location = useLocation();

  useLayoutEffect(() => {
    releaseStuckOverlays();
  }, [location.pathname]);

  return (
    <>
      <DocumentBranding />
      <Outlet />
    </>
  );
}

/**
 * Login lives only on Morph AI. Morph Data (admin) and Forms are open;
 * they reuse the Morph AI shared session cookie when present.
 */
export const appRouter = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      { path: '/login', element: <LoginPage /> },
      { path: '/forms/*', element: <FormsApp /> },
      { path: '/transfinderx/*', element: <LegacyAdminRedirect /> },
      { path: '/schoolmanager/*', element: <LegacySchoolManagerRedirect /> },
      { path: '/skoolz/*', element: <LegacySkoolzRedirect /> },
      {
        path: ADMIN_BASE_PATH,
        element: <AdminLayout />,
        children: [
          { index: true, element: <Navigate to="generic-data" replace /> },
          { path: 'assets', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'resources', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'people', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'members', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'students', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'employees', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'staff', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'places', element: <DistrictsSchools /> },
          { path: 'facilities', element: <Navigate to={`${ADMIN_BASE_PATH}/places`} replace /> },
          { path: 'districts-schools', element: <Navigate to={`${ADMIN_BASE_PATH}/places`} replace /> },
          { path: 'contacts', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'vehicles', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'activities', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'trips', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
          { path: 'generic-data', element: <GenericData /> },
          { path: 'case-tasks', element: <CaseTasks /> },
          { path: 'timelines', element: <Timelines /> },
          { path: 'stories', element: <Navigate to={`${ADMIN_BASE_PATH}/timelines`} replace /> },
          { path: 'story-board', element: <Navigate to={`${ADMIN_BASE_PATH}/timelines`} replace /> },
          { path: 'big-notes', element: <BigNotes /> },
          { path: 'research', element: <Research /> },
          { path: 'quick-sheets/*', element: <Navigate to={`${ADMIN_BASE_PATH}/big-notes`} replace /> },
          {
            path: 'configuration',
            element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace />,
          },
          {
            path: 'settings',
            element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace />,
          },
          {
            path: 'settings/udfs',
            element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace />,
          },
          {
            path: 'settings/custom-attributes',
            element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace />,
          },
          {
            path: 'settings/disability-codes',
            element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace />,
          },
          {
            path: 'settings/ethnic-codes',
            element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace />,
          },
          {
            path: 'settings/grades',
            element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace />,
          },
          {
            path: 'settings/mailing',
            element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace />,
          },
          { path: '*', element: <Navigate to={`${ADMIN_BASE_PATH}/generic-data`} replace /> },
        ],
      },
      {
        element: <ProtectedLayout />,
        children: [
          { path: 'skills', element: <SkillsPage /> },
          { path: '*', element: <App /> },
        ],
      },
    ],
  },
]);
