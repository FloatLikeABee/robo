import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { API_BASE_URL } from '../apiBase';
import { loginMorph, setMorphToken, setMorphAuthSnapshot } from '../auth/morphSession';
import { releaseStuckOverlays } from '../utils/releaseStuckOverlays';

/** UsersPanel-backed login — same session cookie as TranForm / TranMail when on same site. */
export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const from = location.state?.from?.pathname || '/';
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [remember, setRemember] = useState(true);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function onSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const data = await loginMorph(API_BASE_URL, { email: email.trim(), password });
      setMorphToken(data.token, remember);
      setMorphAuthSnapshot({ user: data.user, permissions: data.permissions });
      releaseStuckOverlays();
      navigate(from === '/login' ? '/' : from, { replace: true });
    } catch (err) {
      setError(err?.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  }

  const inputStyle = {
    width: '100%',
    borderRadius: 8,
    border: '1px solid rgba(106, 163, 191, 0.45)',
    background: '#1e3044',
    color: '#f0f7fc',
    padding: '10px 12px',
    fontSize: 16,
    boxSizing: 'border-box',
  };

  const pageBg =
    'radial-gradient(circle at 18% 14%, rgba(56, 189, 248, 0.14), transparent 45%), radial-gradient(circle at 85% 88%, rgba(37, 99, 235, 0.12), transparent 48%), linear-gradient(165deg, #1a2636 0%, #1f354a 52%, #1a2839 100%)';

  return (
    <div
      style={{
        minHeight: '100dvh',
        display: 'grid',
        placeItems: 'center',
        background: pageBg,
        colorScheme: 'dark',
        padding: 'max(16px, env(safe-area-inset-top)) max(16px, env(safe-area-inset-right)) max(16px, env(safe-area-inset-bottom)) max(16px, env(safe-area-inset-left))',
        color: '#eaf4fb',
      }}
    >
      <div
        style={{
          width: '100%',
          maxWidth: 400,
          borderRadius: 16,
          border: '1px solid rgba(106, 174, 204, 0.38)',
          background: 'rgba(28, 42, 58, 0.94)',
          padding: 'clamp(20px, 5vw, 32px)',
          boxShadow: '0 16px 42px rgba(8, 14, 24, 0.42)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 24 }}>
          <span style={{ fontSize: 24 }} aria-hidden>
            🤖
          </span>
          <div>
            <h1 style={{ margin: 0, fontSize: 18, fontWeight: 600, color: '#eaf4fb' }}>Morph AI</h1>
            <p style={{ margin: 0, fontSize: 12, color: '#93adc4' }}>
              Sign in once — Morph Data and Morph Utils use this session
            </p>
          </div>
        </div>
        <form onSubmit={onSubmit} style={{ display: 'grid', gap: 16 }}>
          <label style={{ display: 'grid', gap: 6, fontSize: 12, fontWeight: 500, color: '#b8cfe0' }}>
            Username or email
            <input
              type="text"
              autoComplete="username"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              style={inputStyle}
            />
          </label>
          <label style={{ display: 'grid', gap: 6, fontSize: 12, fontWeight: 500, color: '#b8cfe0' }}>
            Password
            <input
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              style={inputStyle}
            />
          </label>
          <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 14, color: '#b8cfe0', cursor: 'pointer' }}>
            <input type="checkbox" checked={remember} onChange={(e) => setRemember(e.target.checked)} />
            Remember me on this device
          </label>
          {error ? (
            <div
              style={{
                borderRadius: 8,
                border: '1px solid rgba(248, 113, 113, 0.45)',
                background: 'rgba(88, 28, 28, 0.35)',
                padding: '8px 12px',
                fontSize: 14,
                color: '#fecaca',
              }}
            >
              {error}
            </div>
          ) : null}
          <button
            type="submit"
            disabled={loading}
            style={{
              marginTop: 4,
              border: 'none',
              borderRadius: 8,
              padding: '10px 16px',
              fontSize: 14,
              fontWeight: 500,
              color: '#fff',
              background: loading ? '#0e7490' : '#0284c7',
              cursor: loading ? 'not-allowed' : 'pointer',
            }}
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  );
}
