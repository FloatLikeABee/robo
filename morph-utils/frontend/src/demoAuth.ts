import { getSharedToken, loginWithMorph } from './auth';

export function netlifyLocalDemoEnabled(): boolean {
  const flag = import.meta.env.VITE_NETLIFY_LOCAL_DEMO;
  return flag === 'true' || flag === '1' || flag === true;
}

/** Same bootstrap admin as repo-root `.env.example` when Netlify demo build is enabled. */
export async function tryNetlifyDemoLogin(): Promise<boolean> {
  if (!netlifyLocalDemoEnabled() || getSharedToken()) return false;
  const username = String(import.meta.env.VITE_DEMO_ADMIN_USERNAME || 'morphadmin').trim();
  const password = String(import.meta.env.VITE_DEMO_ADMIN_PASSWORD || 'admin123');
  if (!username || !password) return false;
  try {
    await loginWithMorph(username, password, true);
    return true;
  } catch {
    return false;
  }
}
