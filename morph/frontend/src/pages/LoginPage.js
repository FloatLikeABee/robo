import React, { useEffect, useRef, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { API_BASE_URL } from '../apiBase';
import { authKeyboardOverlap, scrollDelta } from '../auth/authKeyboard';
import { loginMorph, setMorphToken, setMorphAuthSnapshot } from '../auth/morphSession';
import { safeReturnPath } from '../auth/returnTo';
import { releaseStuckOverlays } from '../utils/releaseStuckOverlays';
import './LoginPage.css';
import { useT } from '../lib/localeReact';

const NETLIFY_DEMO = process.env.REACT_APP_NETLIFY_LOCAL_DEMO === 'true';
const DEMO_USERNAME = (process.env.REACT_APP_DEMO_ADMIN_USERNAME || 'morphadmin').trim();
const DEMO_PASSWORD = process.env.REACT_APP_DEMO_ADMIN_PASSWORD || 'admin123';

function returnTarget(location) {
  const fromQuery = new URLSearchParams(location.search).get('returnTo') || '';
  const stateFrom = location.state?.from;
  const fromState = stateFrom
    ? `${stateFrom.pathname || ''}${stateFrom.search || ''}${stateFrom.hash || ''}`
    : '';
  const chosen = safeReturnPath(fromQuery) || safeReturnPath(fromState) || '/';
  if (chosen === '/login' || chosen.startsWith('/login?') || chosen.startsWith('/login#')) {
    return '/';
  }
  return chosen;
}

function visibleViewport() {
  const vv = window.visualViewport;
  if (!vv) return null;
  return { height: vv.height, offsetTop: vv.offsetTop };
}

/** UsersPanel-backed login — same session cookie as TranForm / TranMail when on same site. */
export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const from = returnTarget(location);
  const [email, setEmail] = useState(NETLIFY_DEMO ? DEMO_USERNAME : '');
  const [password, setPassword] = useState(NETLIFY_DEMO ? DEMO_PASSWORD : '');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const t = useT();
  const demoLoginAttempted = useRef(false);

  useEffect(() => {
    if (!NETLIFY_DEMO || demoLoginAttempted.current) return undefined;
    demoLoginAttempted.current = true;
    let cancelled = false;
    (async () => {
      setError('');
      setLoading(true);
      try {
        const data = await loginMorph(API_BASE_URL, { email: DEMO_USERNAME, password: DEMO_PASSWORD });
        if (cancelled) return;
        setMorphToken(data.token);
        setMorphAuthSnapshot({ user: data.user, permissions: data.permissions });
        releaseStuckOverlays();
        navigate(from === '/login' ? '/' : from, { replace: true });
      } catch (err) {
        if (!cancelled) setError(err?.message || t('loginFailed'));
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [NETLIFY_DEMO, from, navigate, t]);

  useEffect(() => {
    const shell = document.querySelector('.login-shell');
    if (!shell) return undefined;
    const vv = window.visualViewport;
    let focused = false;

    function reveal() {
      const view = visibleViewport();
      if (!view) return;
      const active = document.activeElement;
      const submit = shell.querySelector('button[type="submit"]');
      const field =
        active instanceof HTMLElement && shell.contains(active) ? active : null;
      const target =
        field && scrollDelta(field.getBoundingClientRect(), view)
          ? field
          : submit && scrollDelta(submit.getBoundingClientRect(), view)
            ? submit
            : null;
      if (!target) return;
      const delta = scrollDelta(target.getBoundingClientRect(), view);
      if (delta) window.scrollBy(0, delta);
    }

    function apply() {
      const overlap = authKeyboardOverlap({
        innerHeight: window.innerHeight,
        visualViewport: vv,
        focused,
      });
      shell.style.setProperty('--auth-keyboard-inset', `${overlap}px`);
      if (focused) requestAnimationFrame(reveal);
    }

    function onFocusIn(event) {
      const field = event.target;
      if (!(field instanceof HTMLElement) || !field.closest('input, textarea')) return;
      focused = true;
      apply();
    }

    function onFocusOut() {
      requestAnimationFrame(() => {
        if (!shell.contains(document.activeElement)) {
          focused = false;
          apply();
        }
      });
    }

    shell.addEventListener('focusin', onFocusIn);
    shell.addEventListener('focusout', onFocusOut);
    vv?.addEventListener('resize', apply);
    vv?.addEventListener('scroll', apply);
    return () => {
      shell.style.setProperty('--auth-keyboard-inset', '0px');
      shell.removeEventListener('focusin', onFocusIn);
      shell.removeEventListener('focusout', onFocusOut);
      vv?.removeEventListener('resize', apply);
      vv?.removeEventListener('scroll', apply);
    };
  }, []);

  async function onSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const data = await loginMorph(API_BASE_URL, { email: email.trim(), password });
      setMorphToken(data.token);
      setMorphAuthSnapshot({ user: data.user, permissions: data.permissions });
      releaseStuckOverlays();
      navigate(from === '/login' ? '/' : from, { replace: true });
    } catch (err) {
      setError(err?.message || t('loginFailed'));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-shell">
      <div className="login-card">
        <div className="login-brand">
          <span style={{ fontSize: 24 }} aria-hidden>
            🤖
          </span>
          <div>
            <h1>{t('morphAI')}</h1>
            <p>{t('signInOnce')}</p>
          </div>
        </div>
        <form className="login-form" onSubmit={onSubmit}>
          <label className="login-field" htmlFor="username">
            {t('username')}
            <input
              id="username"
              name="username"
              type="text"
              autoComplete="username"
              enterKeyHint="next"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </label>
          <label className="login-field" htmlFor="current-password">
            {t('password')}
            <input
              id="current-password"
              name="password"
              type="password"
              autoComplete="current-password"
              enterKeyHint="go"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </label>
          {error ? (
            <div className="login-error" role="alert">
              <span className="login-error-text">{error}</span>
              <button
                type="button"
                className="login-error-dismiss"
                aria-label={t('dismissError')}
                onClick={() => setError('')}
              >
                ×
              </button>
            </div>
          ) : null}
          <button className="login-submit" type="submit" disabled={loading}>
            {loading ? t('signingIn') : t('signIn')}
          </button>
        </form>
      </div>
      <div className="login-keyboard-spacer" aria-hidden="true" />
    </div>
  );
}
