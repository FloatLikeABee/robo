import { getDorphAuthSnapshot } from './dorphSession';

/** True when the logged-in Dorph user has Admin role (can manage users / data import). */
export function isDorphAdmin() {
  const snap = getDorphAuthSnapshot();
  const user = snap?.user;
  if (!user || typeof user !== 'object') return false;
  if (user.is_admin === true) return true;
  const roles = Array.isArray(user.roles) ? user.roles : [];
  return roles.some((r) => String(r).toLowerCase() === 'admin');
}
