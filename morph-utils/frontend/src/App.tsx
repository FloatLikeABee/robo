import { useCallback, useEffect, useMemo, useState, type CSSProperties } from 'react';
import { NavLink, Navigate, Route, Routes, useLocation, useParams } from 'react-router-dom';
import {
  clearSharedToken,
  consumeUrlSessionToken,
  ensureSharedSession,
  getSharedToken,
  withSessionToken,
} from './auth';
import {
  MORPH_AI_URL,
  UTILS_MODULES,
  moduleById,
  normalizeModuleId,
  type UtilsModuleId,
} from './config';
import './App.css';

function ModulePanel({ moduleId }: { moduleId: UtilsModuleId }) {
  const token = getSharedToken();

  return (
    <div className="morph-utils-frame-wrap">
      {UTILS_MODULES.filter((m) => m.embedUrl).map((m) => {
        const src = withSessionToken(m.embedUrl);
        return (
          <iframe
            key={`${m.id}:${token ? 'authed' : 'anon'}:${m.embedUrl}`}
            data-module={m.id}
            title={m.label}
            className={`morph-utils-frame${m.id === moduleId ? '' : ' is-hidden'}`}
            src={src}
            allow="clipboard-read; clipboard-write"
          />
        );
      })}
    </div>
  );
}

function ModuleRoute() {
  const { moduleId } = useParams();
  const normalized = normalizeModuleId(moduleId);
  if (!normalized) return <Navigate to="/sheetx" replace />;
  if (moduleId && moduleId !== normalized) {
    return <Navigate to={`/${normalized}`} replace />;
  }
  return <ModulePanel moduleId={normalized} />;
}

export default function App() {
  const location = useLocation();
  const defaultModule = useMemo(() => UTILS_MODULES[0], []);
  const [authed, setAuthed] = useState(() => Boolean(getSharedToken()));
  const [moreOpen, setMoreOpen] = useState(false);

  const activeModule = useMemo(() => {
    const seg = location.pathname.split('/').filter(Boolean)[0];
    return moduleById(seg);
  }, [location.pathname]);

  const refreshAuth = useCallback(async () => {
    consumeUrlSessionToken();
    const session = await ensureSharedSession();
    setAuthed(session.ok);
  }, []);

  useEffect(() => {
    void refreshAuth();
  }, [refreshAuth]);

  useEffect(() => {
    setMoreOpen(false);
  }, [location.pathname]);

  const onLogout = () => {
    clearSharedToken();
    setAuthed(false);
    setMoreOpen(false);
  };

  const morphAiHref = MORPH_AI_URL || '/';

  return (
    <div className={`morph-utils-shell${moreOpen ? ' is-more-open' : ''}`}>
      <aside className="morph-utils-sidebar" aria-label="Morph Utils navigation">
        <div className="morph-utils-brand" title="Morph Utils">
          <img src="/morph-utils-icon.svg" alt="Morph Utils" />
          <span className="morph-utils-brand-tag">Utils</span>
        </div>

        <nav className="morph-utils-nav">
          {UTILS_MODULES.map((mod) => (
            <NavLink
              key={mod.id}
              to={`/${mod.id}`}
              aria-label={`${mod.label}: ${mod.description}`}
              className={({ isActive }) => `morph-utils-nav-link${isActive ? ' is-active' : ''}`}
              style={{ '--module-accent': mod.accent } as CSSProperties}
            >
              <img src={mod.icon} alt="" aria-hidden />
              <span className="morph-utils-nav-short">{mod.shortLabel}</span>
              <span className="morph-utils-nav-tooltip" role="tooltip">
                <strong>{mod.label}</strong>
                <span>{mod.description}</span>
              </span>
            </NavLink>
          ))}
          <button
            type="button"
            className={`morph-utils-nav-link morph-utils-more-btn${moreOpen ? ' is-active' : ''}`}
            aria-label="More links"
            aria-expanded={moreOpen}
            onClick={() => setMoreOpen((v) => !v)}
          >
            <span className="morph-utils-more-glyph" aria-hidden>
              ···
            </span>
            <span className="morph-utils-nav-short">More</span>
          </button>
        </nav>

        <div className="morph-utils-sidebar-footer">
          {authed ? (
            <button
              type="button"
              className="morph-utils-external-link"
              onClick={onLogout}
              aria-label="Sign out"
            >
              <span aria-hidden>⎋</span>
              <span className="morph-utils-nav-tooltip" role="tooltip">
                <strong>Sign out</strong>
                <span>Clear Morph AI shared session</span>
              </span>
            </button>
          ) : (
            <a
              className="morph-utils-external-link"
              href={morphAiHref}
              aria-label="Sign in on Morph AI"
            >
              <span aria-hidden>⇢</span>
              <span className="morph-utils-nav-tooltip" role="tooltip">
                <strong>Morph AI</strong>
                <span>Sign in once for all Morph apps</span>
              </span>
            </a>
          )}
        </div>
      </aside>

      <div className="morph-utils-workspace">
        <header className="morph-utils-mobile-bar" aria-label="Current app">
          <img src={activeModule.icon} alt="" aria-hidden />
          <div className="morph-utils-mobile-bar-copy">
            <strong>{activeModule.label}</strong>
          </div>
        </header>

        <main className="morph-utils-main">
          <Routes>
            <Route path="/" element={<Navigate to={`/${defaultModule.id}`} replace />} />
            <Route path="/settings" element={<Navigate to={`/${defaultModule.id}`} replace />} />
            <Route path="/formsx" element={<Navigate to="/sheetx" replace />} />
            <Route path="/broadcast" element={<Navigate to="/sheetx" replace />} />
            <Route path="/email-agent" element={<Navigate to="/sheetx" replace />} />
            <Route path="/distiller" element={<Navigate to="/sheetx" replace />} />
            <Route path="/engi" element={<Navigate to="/projects" replace />} />
            <Route path="/morph-engi" element={<Navigate to="/projects" replace />} />
            <Route path="/:moduleId" element={<ModuleRoute />} />
            <Route path="*" element={<Navigate to={`/${defaultModule.id}`} replace />} />
          </Routes>
        </main>
      </div>

      {moreOpen ? (
        <div className="morph-utils-more-sheet" role="dialog" aria-label="More options">
          <button
            type="button"
            className="morph-utils-more-backdrop"
            aria-label="Close more menu"
            onClick={() => setMoreOpen(false)}
          />
          <div className="morph-utils-more-panel">
            <div className="morph-utils-more-head">
              <strong>Morph Utils</strong>
              <button type="button" onClick={() => setMoreOpen(false)} aria-label="Close">
                ✕
              </button>
            </div>
            {authed ? (
              <button type="button" className="morph-utils-more-danger" onClick={onLogout}>
                Sign out
                <span>Clear Morph AI shared session on this device</span>
              </button>
            ) : (
              <a className="morph-utils-more-danger" href={morphAiHref}>
                Sign in on Morph AI
                <span>One login covers Data, Utils, and AI</span>
              </a>
            )}
          </div>
        </div>
      ) : null}
    </div>
  );
}
