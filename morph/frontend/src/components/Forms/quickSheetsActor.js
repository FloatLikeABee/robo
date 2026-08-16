import { getMorphAuthSnapshot, getMorphToken, jwtSubjectUnsafe } from '../../auth/morphSession';

/** Mirrors handlers/authz_middleware.abstractRoleFromPanelRoles (UsersPanel roles → app role). */
export function abstractRoleFromPanelRoles(roles) {
  if (!Array.isArray(roles)) return 'employee';
  for (const raw of roles) {
    if (String(raw || '').toLowerCase().trim() === 'admin') return 'admin';
  }
  return 'employee';
}

export function normalizeActor(input = {}) {
  const roleRaw = String(input.role || 'employee').toLowerCase();
  const role =
    roleRaw === 'member' ? 'member' : roleRaw === 'employee' ? 'employee' : roleRaw === 'admin' ? 'admin' : 'employee';
  const userTypeRaw = String(input.user_type || 'staff').toLowerCase();
  const user_type =
    userTypeRaw === 'student' || userTypeRaw === 'member'
      ? 'student'
      : userTypeRaw === 'staff' || userTypeRaw === 'employee'
        ? 'staff'
        : role === 'member'
          ? 'student'
          : 'staff';
  const user_id = String(input.user_id ?? '').trim();
  const name = String(input.name || 'User').trim() || 'User';
  return { role, user_id, user_type, name };
}

/**
 * Current user context for Quick Sheets / forms APIs (JWT sub + cached UsersPanel profile).
 */
export function getQuickSheetsActor() {
  const snapshot = getMorphAuthSnapshot();
  const token = getMorphToken();
  const sub = jwtSubjectUnsafe(token);
  const roles = Array.isArray(snapshot?.user?.roles)
    ? snapshot.user.roles.map((x) => (typeof x === 'string' ? x : String(x)))
    : [];
  const abstract = abstractRoleFromPanelRoles(roles);
  let role = 'employee';
  if (abstract === 'admin') role = 'admin';
  else if (abstract === 'member') role = 'member';
  else if (abstract === 'employee') role = 'employee';

  const user_type = role === 'member' ? 'student' : 'staff';
  const u = snapshot?.user || {};
  const name =
    String(u.email || u.username || '').trim() || String(u.Username || '').trim() || 'User';
  const user_id = sub;

  return normalizeActor({ role, user_type, user_id, name });
}
