import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import type { ReactNode } from 'react';
import { useEffect } from 'react';
import { ConfirmProvider } from './context/ConfirmContext';
import { Layout } from './components/Layout';
import { PublicForm } from './pages/PublicForm';
import { PublicEventInfoSubmit } from './pages/PublicEventInfoSubmit';
import { EventsInfo } from './pages/EventsInfo';
import { SurveyBot } from './pages/SurveyBot';
import { PublicAISheet } from './pages/PublicAISheet';
import { Login } from './pages/Login';
import { getAuthToken, AUTH_EXPIRED_EVENT } from './lib/api';

function AuthExpiredRedirect() {
  const navigate = useNavigate();
  useEffect(() => {
    const handler = () => navigate('/login', { replace: true });
    window.addEventListener(AUTH_EXPIRED_EVENT, handler);
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, handler);
  }, [navigate]);
  return null;
}

function RequireAuth({ children }: { children: ReactNode }) {
  if (!getAuthToken()) return <Navigate to="/login" replace />;
  return children;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthExpiredRedirect />
      <ConfirmProvider>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<RequireAuth><Layout /></RequireAuth>}>
          <Route index element={<Navigate to="/survey-bot" replace />} />
          <Route path="forms" element={<Navigate to="/survey-bot" replace />} />
          <Route path="events-info" element={<EventsInfo />} />
          <Route path="survey-bot" element={<SurveyBot />} />
          <Route path="settings" element={<Navigate to="/survey-bot" replace />} />
          <Route path="forms/new" element={<Navigate to="/survey-bot" replace />} />
          <Route path="forms/:id/edit" element={<Navigate to="/survey-bot" replace />} />
          <Route path="forms/:id/results" element={<Navigate to="/survey-bot" replace />} />
          <Route path="*" element={<Navigate to="/survey-bot" replace />} />
        </Route>
        <Route path="f/:slug" element={<PublicForm />} />
        <Route path="s/:slug" element={<PublicAISheet />} />
        <Route path="events-info/submit" element={<PublicEventInfoSubmit />} />
      </Routes>
      </ConfirmProvider>
    </BrowserRouter>
  );
}
