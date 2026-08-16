import { getMorphAuthSnapshot } from './morphSession';

/** True when the logged-in Morph user has Admin role (can manage users / data import). */
export function isMorphAdmin() {
  const snap = getMorphAuthSnapshot();
  const user = snap?.user;
  if (!user || typeof user !== 'object') return false;
  if (user.is_admin === true) return true;
  const roles = Array.isArray(user.roles) ? user.roles : [];
  return roles.some((r) => String(r).toLowerCase() === 'admin');
}
