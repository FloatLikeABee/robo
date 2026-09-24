import { useCallback, useEffect, useMemo, useState, type CSSProperties } from 'react';
import { NavLink, Navigate, Route, Routes, useLocation, useParams } from 'react-router-dom';
import {
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
import { embedStartCommand, probeEmbedOrigin, PROBED_EMBED_IDS } from './embedProbe';
import UserProfileModal from './UserProfileModal';
import './App.css';

function ModulePanel({ moduleId }: { moduleId: UtilsModuleId }) {
  const token = getSharedToken();
  const [up, setUp] = useState<Partial<Record<UtilsModuleId, boolean>>>({});
  const [checking, setChecking] = useState(true);

  const recheck = useCallback(async () => {
    setChecking(true);
    const next: Partial<Record<UtilsModuleId, boolean>> = {};
    await Promise.all(
      UTILS_MODULES.filter((m) => m.embedUrl && PROBED_EMBED_IDS.includes(m.id)).map(async (m) => {
        next[m.id] = await probeEmbedOrigin(m.embedUrl as string);
      }),
    );
    setUp(next);
    setChecking(false);
  }, []);

  useEffect(() => {
    void recheck();
  }, [recheck]);

  return (
    <div className="morph-utils-frame-wrap">
      {UTILS_MODULES.filter((m) => m.embedUrl).map((m) => {
        const src = withSessionToken(m.embedUrl);
        const visible = m.id === moduleId;
        const needsProbe = PROBED_EMBED_IDS.includes(m.id);
        const ready = !needsProbe || up[m.id] === true;
        const pending = needsProbe && checking && up[m.id] === undefined;

        if (needsProbe && !ready) {
          if (!visible) return null;
          return (
            <div
              key={m.id}
              data-module={m.id}
              className="morph-utils-embed-down"
              role="status"
            >
              {pending ? (
                <p>Checking {m.label}…</p>
              ) : (
                <>
                  <p>
                    {m.label} isn’t running. Start it with{' '}
                    <code>{embedStartCommand(m.id)}</code>
                    {m.id === 'datax' ? (
                      <>
                        {' '}
                        (also started by <code>./start-all.sh start morph-utils</code>).
                      </>
                    ) : null}
                  </p>
                  <button type="button" className="morph-utils-embed-retry" onClick={() => void recheck()}>
                    Retry
                  </button>
                </>
              )}
            </div>
          );
        }

        return (
          <iframe
            key={`${m.id}:${token ? 'authed' : 'anon'}:${m.embedUrl}`}
            data-module={m.id}
            title={m.label}
            className={`morph-utils-frame${visible ? '' : ' is-hidden'}`}
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
  const [profileOpen, setProfileOpen] = useState(false);

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

  const morphAiHref = MORPH_AI_URL || '/';

  return (
    <div className={`morph-utils-shell${moreOpen ? ' is-more-open' : ''}`}>
      <aside className="morph-utils-sidebar" aria-label="MorphUtils navigation">
        <div className="morph-utils-brand" title="MorphUtils">
          <img src="/morph-utils-icon.svg" alt="MorphUtils" />
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
              className="morph-utils-account-btn"
              aria-label="Your account"
              onClick={() => setProfileOpen(true)}
            >
              <span className="morph-utils-account-icon" aria-hidden>👤</span>
              <span className="morph-utils-nav-tooltip" role="tooltip">
                <strong>Account</strong>
                <span>Username & password</span>
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
              <strong>MorphUtils</strong>
              <button type="button" onClick={() => setMoreOpen(false)} aria-label="Close">
                ✕
              </button>
            </div>
            {authed ? null : (
              <a className="morph-utils-more-danger" href={morphAiHref}>
                Sign in on Morph AI
                <span>One login covers Data, Utils, and AI</span>
              </a>
            )}
          </div>
        </div>
      ) : null}

      <UserProfileModal open={profileOpen} onClose={() => setProfileOpen(false)} />
    </div>
  );
}
