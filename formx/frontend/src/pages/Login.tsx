import { useState } from 'react';
import type { FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, readLoginNotice } from '../lib/api';

export function Login() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(true);
  const [error, setError] = useState(readLoginNotice);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await api.auth.login({ email, password }, { rememberMe });
      document.body.style.removeProperty('overflow');
      document.body.style.removeProperty('padding-right');
      navigate('/', { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  }

  const fieldClass =
    'w-full rounded-lg border border-cyan-400/35 bg-[#1e3044] px-3 py-2 text-sm text-slate-100 shadow-xs outline-none transition placeholder:text-slate-400/80 focus:border-sky-400 focus:ring-2 focus:ring-sky-400/25';

  return (
    <div
      className="min-h-dvh flex items-center justify-center p-4 text-slate-100"
      style={{
        colorScheme: 'dark',
        background:
          'radial-gradient(circle at 18% 14%, rgba(56, 189, 248, 0.14), transparent 45%), radial-gradient(circle at 85% 88%, rgba(37, 99, 235, 0.12), transparent 48%), linear-gradient(165deg, #1a2636 0%, #1f354a 52%, #1a2839 100%)',
      }}
    >
      <div className="w-full max-w-sm rounded-2xl border border-cyan-400/35 bg-[#1c2a3c]/95 p-8 shadow-[0_16px_42px_rgba(8,14,24,0.42)] backdrop-blur-sm">
        <div className="mb-6 flex items-center gap-2">
          <span className="text-2xl">📄</span>
          <div>
            <h1 className="text-lg font-semibold text-[#eaf4fb]">TranForm</h1>
            <p className="text-xs text-[#93adc4]">Sign in to continue</p>
          </div>
        </div>
        <form onSubmit={onSubmit} className="grid gap-4">
          <div className="grid gap-1.5">
            <label htmlFor="login-email" className="text-xs font-medium text-[#b8cfe0]">
              Email
            </label>
            <input
              id="login-email"
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className={fieldClass}
            />
          </div>
          <div className="grid gap-1.5">
            <label htmlFor="login-password" className="text-xs font-medium text-[#b8cfe0]">
              Password
            </label>
            <input
              id="login-password"
              type="password"
              autoComplete="current-password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className={fieldClass}
            />
          </div>
          <label className="flex cursor-pointer items-center gap-2 text-sm text-[#b8cfe0]">
            <input type="checkbox" checked={rememberMe} onChange={(e) => setRememberMe(e.target.checked)} />
            Remember me on this device
          </label>
          {error ? (
            <div className="rounded-lg border border-red-400/45 bg-red-950/35 px-3 py-2 text-sm text-red-100">{error}</div>
          ) : null}
          <button
            type="submit"
            disabled={loading}
            className="mt-1 w-full rounded-lg bg-sky-600 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition hover:bg-sky-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  );
}
