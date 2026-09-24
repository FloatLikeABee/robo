const TOKEN_KEY = 'invite_signup_token';

export function getToken(): string {
  try {
    return localStorage.getItem(TOKEN_KEY) || '';
  } catch {
    return '';
  }
}

export function setToken(token: string): void {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token);
    else localStorage.removeItem(TOKEN_KEY);
  } catch {
    /* ignore */
  }
}

async function parseJson(res: Response) {
  const text = await res.text();
  try {
    return text ? JSON.parse(text) : {};
  } catch {
    return { error: text || `Request failed (${res.status})` };
  }
}

export async function login(email: string, password: string) {
  const res = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  const data = await parseJson(res);
  if (!res.ok || !data.token) {
    throw new Error(data.error || 'Login failed');
  }
  setToken(data.token);
  return data;
}

export async function fetchMe() {
  const token = getToken();
  if (!token) throw new Error('Not signed in');
  const res = await fetch('/api/auth/me', { headers: { Authorization: `Bearer ${token}` } });
  const data = await parseJson(res);
  if (!res.ok) throw new Error(data.error || 'Session expired');
  return data.user as { is_admin?: boolean };
}

export type InviteCodeRow = {
  id: string;
  code: string;
  created_at: string;
  used: boolean;
  redeemed_at?: string;
};

export async function listInviteCodes(): Promise<InviteCodeRow[]> {
  const res = await fetch('/api/admin/invite-codes', {
    headers: { Authorization: `Bearer ${getToken()}` },
  });
  const data = await parseJson(res);
  if (!res.ok) throw new Error(data.error || 'Failed to load codes');
  return Array.isArray(data.codes) ? data.codes : [];
}

export async function createInviteCode(): Promise<string> {
  const res = await fetch('/api/admin/invite-codes', {
    method: 'POST',
    headers: { Authorization: `Bearer ${getToken()}` },
  });
  const data = await parseJson(res);
  if (!res.ok) throw new Error(data.error || 'Failed to create code');
  const row = data.code as { code?: string } | string;
  if (typeof row === 'string') return row;
  return String(row?.code || '');
}

export async function redeemCode(code: string): Promise<{ username: string; password: string }> {
  const res = await fetch('/api/invite/redeem', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  });
  const data = await parseJson(res);
  if (!res.ok) throw new Error(data.error || 'Invalid code');
  return { username: data.username, password: data.password };
}
