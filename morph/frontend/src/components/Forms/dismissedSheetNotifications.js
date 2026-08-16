const STORAGE_KEY = 'morph-dismissed-sheet-notifications';

function readSet() {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    const arr = raw ? JSON.parse(raw) : [];
    return new Set(Array.isArray(arr) ? arr.map(String) : []);
  } catch {
    return new Set();
  }
}

function writeSet(set) {
  try {
    const arr = [...set].slice(-200);
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(arr));
  } catch {}
}

export function isSheetNotificationDismissed(assignmentId) {
  if (assignmentId == null || assignmentId === '') return false;
  return readSet().has(String(assignmentId));
}

/** Mark a pending sheet notification as seen; it stays hidden from the bell until the assignment is submitted (or list changes). */
export function dismissSheetNotification(assignmentId) {
  if (assignmentId == null || assignmentId === '') return;
  const s = readSet();
  s.add(String(assignmentId));
  writeSet(s);
}

export function filterUndismissedNotifications(notifications) {
  const dismissed = readSet();
  if (!dismissed.size || !Array.isArray(notifications)) return notifications || [];
  return notifications.filter((n) => {
    const id = String(n.assignment_id || n.id || '');
    return id && !dismissed.has(id);
  });
}
