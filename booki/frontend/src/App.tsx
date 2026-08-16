import type { ReactNode } from 'react';
import { useEffect, useState } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { platformSessionFromCookie } from './api/client';
import { useAuth } from './store/auth';
import { AppLayout } from './components/AppLayout';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { DataAiPage } from './pages/DataAiPage';
import { BookingsPage } from './pages/BookingsPage';
import { FlowLogPage } from './pages/FlowLogPage';

function AuthReady({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(() => useAuth.persist.hasHydrated());

  useEffect(() => {
    if (useAuth.persist.hasHydrated()) {
      setReady(true);
      return;
    }
    return useAuth.persist.onFinishHydration(() => setReady(true));
  }, []);

  if (!ready) {
    return (
      <div className="flex-1 flex items-center justify-center min-h-0 bg-[#1e1e24] text-[#b8b8c7] text-sm">
        Loading…
      </div>
    );
  }
  return <>{children}</>;
}

/** In dev, reuse shared UsersPanel cookie when already signed in elsewhere. */
function DevAutoAuth({ children }: { children: ReactNode }) {
  const setTokens = useAuth((s) => s.setTokens);
  const [done, setDone] = useState(() => {
    if (!import.meta.env.DEV) return true;
    return useAuth.getState().accessToken != null;
  });

  useEffect(() => {
    if (!import.meta.env.DEV) return;
    if (useAuth.getState().accessToken != null) {
      setDone(true);
      return;
    }
    let cancelled = false;
    platformSessionFromCookie()
      .then((t) => {
        if (cancelled || !t) return;
        setTokens(t.access_token, t.refresh_token, t.platform_token ?? undefined);
      })
      .catch(console.error)
      .finally(() => {
        if (!cancelled) setDone(true);
      });
    return () => {
      cancelled = true;
    };
  }, [setTokens]);

  if (import.meta.env.DEV && !done) {
    return (
      <div className="flex-1 flex items-center justify-center min-h-0 bg-[#1e1e24] text-[#b8b8c7] text-sm">
        Dev session…
      </div>
    );
  }

  return <>{children}</>;
}

/** Backfill UsersPanel JWT when Booki session exists but messaging token was never stored. */
function EnsurePlatformSession({ children }: { children: ReactNode }) {
  const accessToken = useAuth((s) => s.accessToken);
  const platformJwt = useAuth((s) => s.platformJwt);
  const setTokens = useAuth((s) => s.setTokens);

  useEffect(() => {
    if (!accessToken || platformJwt?.trim()) return;
    let cancelled = false;
    platformSessionFromCookie()
      .then((t) => {
        if (cancelled || !t?.platform_token) return;
        setTokens(t.access_token, t.refresh_token, t.platform_token);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [accessToken, platformJwt, setTokens]);

  return <>{children}</>;
}

function Protected({ children }: { children: ReactNode }) {
  const token = useAuth((s) => s.accessToken);
  if (!token) return <Navigate to="/login" replace />;
  return <EnsurePlatformSession>{children}</EnsurePlatformSession>;
}

export default function App() {
  return (
    <div className="h-dvh max-h-dvh overflow-hidden flex flex-col bg-bg text-text">
      <AuthReady>
        <DevAutoAuth>
          <Routes>
            <Route
              path="/login"
              element={
                <div className="flex-1 min-h-0 overflow-y-auto">
                  <LoginPage />
                </div>
              }
            />
            <Route
              path="/register"
              element={
                <div className="flex-1 min-h-0 overflow-y-auto">
                  <RegisterPage />
                </div>
              }
            />
            <Route
              path="/"
              element={
                <Protected>
                  <AppLayout />
                </Protected>
              }
            >
              <Route index element={<DataAiPage />} />
              <Route path="bookings" element={<BookingsPage />} />
              <Route path="flow-log" element={<FlowLogPage />} />
              <Route path="accounting" element={<Navigate to="/" replace />} />
              <Route path="warehouse" element={<Navigate to="/" replace />} />
              <Route path="assets" element={<Navigate to="/" replace />} />
              <Route path="settings" element={<Navigate to="/" replace />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </DevAutoAuth>
      </AuthReady>
    </div>
  );
}
