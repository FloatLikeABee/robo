import { useCallback, useEffect, useMemo, useState, type CSSProperties } from 'react';
import { NavLink, Navigate, Route, Routes, useLocation, useParams } from 'react-router-dom';
import {
  consumeUrlSessionToken,
  ensureSharedSession,
  getSharedToken,
  withSessionToken,
} from './auth';
import {
  DORPH_LOGIN_HREF,
  UTILS_MODULES,
  moduleById,
  normalizeModuleId,
  type UtilsModuleId,
} from './config';
import { embedStartCommand, probeEmbedOrigin, PROBED_EMBED_IDS } from './embedProbe';
import { usesLocalStartHint } from './publicUrl';
import { tryNetlifyDemoLogin } from './demoAuth';
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
    <div className="dorph-utils-frame-wrap">
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
              className="dorph-utils-embed-down"
              role="status"
            >
              {pending ? (
                <p>Checking {m.label}…</p>
              ) : usesLocalStartHint(m.embedUrl as string) ? (
                <>
                  <p>
                    {m.label} isn’t running. Start it with{' '}
                    <code>{embedStartCommand(m.id)}</code>
                    {m.id === 'datax' ? (
                      <>
                        {' '}
                        (also started by <code>./start-all.sh start dorph-utils</code>).
                      </>
                    ) : null}
                  </p>
                  <button type="button" className="dorph-utils-embed-retry" onClick={() => void recheck()}>
                    Retry
                  </button>
                </>
              ) : (
                <>
                  <p>{m.label} isn’t reachable at its configured origin.</p>
                  <button type="button" className="dorph-utils-embed-retry" onClick={() => void recheck()}>
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
            className={`dorph-utils-frame${visible ? '' : ' is-hidden'}`}
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
    if (!getSharedToken()) await tryNetlifyDemoLogin();
    const session = await ensureSharedSession();
    setAuthed(session.ok);
  }, []);

  useEffect(() => {
    void refreshAuth();
  }, [refreshAuth]);

  useEffect(() => {
    setMoreOpen(false);
  }, [location.pathname]);

  const dorphAiHref = DORPH_LOGIN_HREF;

  return (
    <div className={`dorph-utils-shell${moreOpen ? ' is-more-open' : ''}`}>
      <aside className="dorph-utils-sidebar" aria-label="DorphUtils navigation">
        <div className="dorph-utils-brand" title="DorphUtils">
          <img src="/dorph-utils-icon.svg" alt="DorphUtils" />
          <span className="dorph-utils-brand-tag">Utils</span>
        </div>

        <nav className="dorph-utils-nav">
          {UTILS_MODULES.map((mod) => (
            <NavLink
              key={mod.id}
              to={`/${mod.id}`}
              aria-label={`${mod.label}: ${mod.description}`}
              className={({ isActive }) => `dorph-utils-nav-link${isActive ? ' is-active' : ''}`}
              style={{ '--module-accent': mod.accent } as CSSProperties}
            >
              <img src={mod.icon} alt="" aria-hidden />
              <span className="dorph-utils-nav-short">{mod.shortLabel}</span>
              <span className="dorph-utils-nav-tooltip" role="tooltip">
                <strong>{mod.label}</strong>
                <span>{mod.description}</span>
              </span>
            </NavLink>
          ))}
          <button
            type="button"
            className={`dorph-utils-nav-link dorph-utils-more-btn${moreOpen ? ' is-active' : ''}`}
            aria-label="More links"
            aria-expanded={moreOpen}
            onClick={() => setMoreOpen((v) => !v)}
          >
            <span className="dorph-utils-more-glyph" aria-hidden>
              ···
            </span>
            <span className="dorph-utils-nav-short">More</span>
          </button>
        </nav>

        <div className="dorph-utils-sidebar-footer">
          {authed ? (
            <button
              type="button"
              className="dorph-utils-account-btn"
              aria-label="Your account"
              onClick={() => setProfileOpen(true)}
            >
              <span className="dorph-utils-account-icon" aria-hidden>👤</span>
              <span className="dorph-utils-nav-tooltip" role="tooltip">
                <strong>Account</strong>
                <span>Username & password</span>
              </span>
            </button>
          ) : dorphAiHref ? (
            <a
              className="dorph-utils-external-link"
              href={dorphAiHref}
              target="_top"
              rel="noopener noreferrer"
              aria-label="Sign in on Dorph"
            >
              <span aria-hidden>⇢</span>
              <span className="dorph-utils-nav-tooltip" role="tooltip">
                <strong>Dorph</strong>
                <span>Sign in once for all Dorph apps</span>
              </span>
            </a>
          ) : null}
        </div>
      </aside>

      <div className="dorph-utils-workspace">
        <header className="dorph-utils-mobile-bar" aria-label="Current app">
          <img src={activeModule.icon} alt="" aria-hidden />
          <div className="dorph-utils-mobile-bar-copy">
            <strong>{activeModule.label}</strong>
          </div>
        </header>

        <main className="dorph-utils-main">
          {authed ? null : (
            <div className="dorph-utils-frame-wrap">
              <div className="dorph-utils-embed-down" role="status">
                <p>Sign in on Dorph. One account covers DorphUtils and the apps it embeds.</p>
                {dorphAiHref ? (
                  <a className="dorph-utils-embed-retry" href={dorphAiHref} target="_top" rel="noopener noreferrer">
                    Sign in on Dorph
                  </a>
                ) : (
                  <p>Dorph origin is not set. Set VITE_DORPH_API_URL and restart.</p>
                )}
              </div>
            </div>
          )}
          {authed ? (
          <Routes>
            <Route path="/" element={<Navigate to={`/${defaultModule.id}`} replace />} />
            <Route path="/settings" element={<Navigate to={`/${defaultModule.id}`} replace />} />
            <Route path="/formsx" element={<Navigate to="/sheetx" replace />} />
            <Route path="/broadcast" element={<Navigate to="/sheetx" replace />} />
            <Route path="/email-agent" element={<Navigate to="/sheetx" replace />} />
            <Route path="/distiller" element={<Navigate to="/sheetx" replace />} />
            <Route path="/engi" element={<Navigate to="/projects" replace />} />
            <Route path="/dorph-engi" element={<Navigate to="/projects" replace />} />
            <Route path="/morph-engi" element={<Navigate to="/projects" replace />} />
            <Route path="/:moduleId" element={<ModuleRoute />} />
            <Route path="*" element={<Navigate to={`/${defaultModule.id}`} replace />} />
          </Routes>
          ) : null}
        </main>
      </div>

      {moreOpen ? (
        <div className="dorph-utils-more-sheet" role="dialog" aria-label="More options">
          <button
            type="button"
            className="dorph-utils-more-backdrop"
            aria-label="Close more menu"
            onClick={() => setMoreOpen(false)}
          />
          <div className="dorph-utils-more-panel">
            <div className="dorph-utils-more-head">
              <strong>DorphUtils</strong>
              <button type="button" onClick={() => setMoreOpen(false)} aria-label="Close">
                ✕
              </button>
            </div>
            {authed || !dorphAiHref ? null : (
              <a className="dorph-utils-more-danger" href={dorphAiHref} target="_top" rel="noopener noreferrer">
                Sign in on Dorph
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
