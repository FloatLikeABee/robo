import { useCallback, useEffect, useState } from 'react';
import { Link, Navigate, Route, Routes, useNavigate } from 'react-router-dom';
import {
  createInviteCode,
  fetchMe,
  getToken,
  listInviteCodes,
  login,
  redeemCode,
  setToken,
  type InviteCodeRow,
} from './api';

function Home() {
  return (
    <div className="shell">
      <div className="card">
        <h1>Invite Signup</h1>
        <p className="sub">Get Morph login credentials with an invitation code.</p>
        <div className="nav">
          <Link to="/redeem">Redeem a code</Link>
          <Link to="/admin">Admin</Link>
        </div>
      </div>
    </div>
  );
}

function RedeemPage() {
  const [code, setCode] = useState('');
  const [error, setError] = useState('');
  const [creds, setCreds] = useState<{ username: string; password: string } | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      const out = await redeemCode(code.trim());
      setCreds(out);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Redeem failed');
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="shell">
      <div className="card">
        <h1>Redeem code</h1>
        <p className="sub">Enter the invitation code you received.</p>
        {error ? <p className="error">{error}</p> : null}
        {creds ? (
          <div className="creds">
            <p className="success">Account created — copy these now. They are shown once.</p>
            <span>Username</span>
            <strong>{creds.username}</strong>
            <span style={{ marginTop: '0.75rem' }}>Password</span>
            <strong>{creds.password}</strong>
            <p className="sub" style={{ marginTop: '1rem' }}>
              Sign in on <a href="http://localhost:3031">Morph AI</a> with these credentials.
            </p>
          </div>
        ) : (
          <form onSubmit={(e) => void onSubmit(e)}>
            <label htmlFor="code">Invitation code</label>
            <input id="code" value={code} onChange={(e) => setCode(e.target.value)} autoComplete="off" required />
            <button type="submit" disabled={busy}>{busy ? 'Working…' : 'Get username & password'}</button>
          </form>
        )}
        <div className="nav">
          <Link to="/">Home</Link>
        </div>
      </div>
    </div>
  );
}

function AdminPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [authed, setAuthed] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  const [newCode, setNewCode] = useState('');
  const [codes, setCodes] = useState<InviteCodeRow[]>([]);

  const refresh = useCallback(async () => {
    if (!getToken()) {
      setAuthed(false);
      return;
    }
    try {
      const me = await fetchMe();
      if (!me.is_admin) {
        setAuthed(true);
        setIsAdmin(false);
        return;
      }
      setAuthed(true);
      setIsAdmin(true);
      setCodes(await listInviteCodes());
    } catch {
      setToken('');
      setAuthed(false);
      setIsAdmin(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onLogin(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      await login(email.trim(), password);
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setBusy(false);
    }
  }

  async function onCreate() {
    setBusy(true);
    setError('');
    setNewCode('');
    try {
      const code = await createInviteCode();
      setNewCode(code);
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Create failed');
    } finally {
      setBusy(false);
    }
  }

  if (!authed) {
    return (
      <div className="shell">
        <div className="card">
          <h1>Admin</h1>
          <p className="sub">Sign in with your Morph admin account to create invitation codes.</p>
          {error ? <p className="error">{error}</p> : null}
          <form onSubmit={(e) => void onLogin(e)}>
            <label htmlFor="email">Email or username</label>
            <input id="email" value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="username" required />
            <label htmlFor="password">Password</label>
            <input id="password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password" required />
            <button type="submit" disabled={busy}>{busy ? 'Signing in…' : 'Sign in'}</button>
          </form>
          <div className="nav">
            <Link to="/">Home</Link>
          </div>
        </div>
      </div>
    );
  }

  if (!isAdmin) {
    return (
      <div className="shell">
        <div className="card">
          <h1>Admin</h1>
          <p className="error">Admin access required to create invitation codes.</p>
          <button type="button" className="secondary" onClick={() => { setToken(''); navigate('/admin'); }}>
            Sign out
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="shell">
      <div className="card">
        <h1>Invitation codes</h1>
        <p className="sub">Create a code and share it with someone who needs a Morph login.</p>
        {error ? <p className="error">{error}</p> : null}
        {newCode ? (
          <div className="creds">
            <p className="success">New code — copy and send it now.</p>
            <strong>{newCode}</strong>
          </div>
        ) : null}
        <button type="button" onClick={() => void onCreate()} disabled={busy}>
          {busy ? 'Creating…' : 'Create invitation code'}
        </button>
        <button type="button" className="secondary" style={{ marginLeft: '0.5rem' }} onClick={() => { setToken(''); setAuthed(false); }}>
          Sign out
        </button>
        {codes.length > 0 ? (
          <ul className="codes">
            {codes.map((c) => (
              <li key={c.id}>
                <code>{c.code}</code> — {c.used ? `used ${c.redeemed_at || ''}` : 'unused'} ({c.created_at})
              </li>
            ))}
          </ul>
        ) : null}
        <div className="nav">
          <Link to="/">Home</Link>
        </div>
      </div>
    </div>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/redeem" element={<RedeemPage />} />
      <Route path="/admin" element={<AdminPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
