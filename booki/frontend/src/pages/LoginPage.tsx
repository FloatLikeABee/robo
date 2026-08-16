import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { postJson, applyAuthResponse } from '../api/client';
import { useAuth } from '../store/auth';

export function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [err, setErr] = useState('');
  const setTokens = useAuth((s) => s.setTokens);
  const navigate = useNavigate();

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErr('');
    try {
      const raw = await postJson<Record<string, unknown>>('/api/v1/auth/login', { email, password });
      const { access_token, refresh_token, platform_token } = applyAuthResponse(raw);
      setTokens(access_token, refresh_token, platform_token);
      navigate('/', { replace: true });
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Login failed');
    }
  }

  return (
    <div className="auth-flow-shell min-h-full flex items-center justify-center p-6">
      <div className="auth-flow-panel w-full max-w-md rounded-[20px] p-8">
        <h1 className="text-2xl font-semibold mb-1">Welcome back</h1>
        <p className="text-muted text-sm mb-6">Sign in with your platform account (UsersPanel)</p>
        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <label className="block text-xs text-muted mb-1">Email</label>
            <input
              className="w-full rounded-xl bg-card border border-white/10 px-4 py-3 text-text outline-none focus:border-violet"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              type="email"
              required
            />
          </div>
          <div>
            <label className="block text-xs text-muted mb-1">Password</label>
            <input
              className="w-full rounded-xl bg-card border border-white/10 px-4 py-3 text-text outline-none focus:border-violet"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              type="password"
              required
            />
          </div>
          {err && <p className="text-danger text-sm">{err}</p>}
          <button
            type="submit"
            className="w-full rounded-xl bg-violet py-3 font-medium text-white shadow-[0_0_24px_rgba(91,63,214,0.35)] hover:opacity-95"
          >
            Sign in
          </button>
        </form>
        <p className="text-sm text-muted mt-6 text-center">
          No account?{' '}
          <Link to="/register" className="text-yellow">
            Register
          </Link>
        </p>
      </div>
    </div>
  );
}
