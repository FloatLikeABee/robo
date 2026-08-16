import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { postJson, readAuthTokens } from '../api/client';
import { useAuth } from '../store/auth';

export function RegisterPage() {
  const [organizationName, setOrganizationName] = useState('');
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [err, setErr] = useState('');
  const setTokens = useAuth((s) => s.setTokens);
  const navigate = useNavigate();

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErr('');
    try {
      const raw = await postJson<Record<string, unknown>>('/api/v1/auth/register', {
        organization_name: organizationName,
        name,
        email,
        password,
      });
      const { access_token, refresh_token } = readAuthTokens(raw);
      setTokens(access_token, refresh_token);
      navigate('/', { replace: true });
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Registration failed');
    }
  }

  return (
    <div className="auth-flow-shell min-h-full flex items-center justify-center p-6">
      <div className="auth-flow-panel w-full max-w-md rounded-[20px] p-8">
        <h1 className="text-2xl font-semibold mb-1">Create organization</h1>
        <p className="text-muted text-sm mb-6">Start with chart of accounts seeded for you</p>
        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <label className="block text-xs text-muted mb-1">Organization name</label>
            <input
              className="w-full rounded-xl bg-card border border-white/10 px-4 py-3 text-text outline-none focus:border-violet"
              value={organizationName}
              onChange={(e) => setOrganizationName(e.target.value)}
              required
            />
          </div>
          <div>
            <label className="block text-xs text-muted mb-1">Your name</label>
            <input
              className="w-full rounded-xl bg-card border border-white/10 px-4 py-3 text-text outline-none focus:border-violet"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
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
            <label className="block text-xs text-muted mb-1">Password (min 8)</label>
            <input
              className="w-full rounded-xl bg-card border border-white/10 px-4 py-3 text-text outline-none focus:border-violet"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              type="password"
              minLength={8}
              required
            />
          </div>
          {err && <p className="text-danger text-sm">{err}</p>}
          <button
            type="submit"
            className="w-full rounded-xl bg-yellow py-3 font-medium text-[#1e1e24] hover:opacity-95"
          >
            Create account
          </button>
        </form>
        <p className="text-sm text-muted mt-6 text-center">
          <Link to="/login" className="text-violet-bright">
            Back to login
          </Link>
        </p>
      </div>
    </div>
  );
}
