const DB_NAME = 'morphai-files-workspace';
const DB_VERSION = 1;
const RECENTS_STORE = 'recents';
const SESSIONS_STORE = 'sessions';
const RECENTS_CAP = 10;
const MAX_FILE_BYTES = 512 * 1024;
const SKIP_DIRS = new Set(['.git', 'node_modules']);

function shouldSkipDir(name) {
  return SKIP_DIRS.has(name) || (name.startsWith('.') && name !== '.');
}

function shouldSkipFile(name, size) {
  if (!name || name.startsWith('.')) return true;
  if (size > MAX_FILE_BYTES) return true;
  return false;
}

export async function walkDirectoryHandle(handle, prefix = '') {
  const out = [];
  for await (const [name, entry] of handle.entries()) {
    const path = prefix ? `${prefix}/${name}` : name;
    if (entry.kind === 'directory') {
      if (shouldSkipDir(name)) continue;
      out.push(...(await walkDirectoryHandle(entry, path)));
    } else if (entry.kind === 'file') {
      const file = await entry.getFile();
      out.push({
        path,
        size: file.size,
        skipped: shouldSkipFile(name, file.size),
        handle: entry,
      });
    }
  }
  return out;
}

function openDb() {
  return new Promise((resolve, reject) => {
    if (typeof indexedDB === 'undefined') {
      reject(new Error('indexedDB unavailable'));
      return;
    }
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = () => {
      const db = req.result;
      if (!db.objectStoreNames.contains(RECENTS_STORE)) {
        db.createObjectStore(RECENTS_STORE, { keyPath: 'id' });
      }
      if (!db.objectStoreNames.contains(SESSIONS_STORE)) {
        db.createObjectStore(SESSIONS_STORE, { keyPath: 'sessionId' });
      }
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

function idbReq(req) {
  return new Promise((resolve, reject) => {
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

async function withStore(storeName, mode, fn) {
  const db = await openDb();
  try {
    const tx = db.transaction(storeName, mode);
    const store = tx.objectStore(storeName);
    const result = await fn(store);
    await new Promise((resolve, reject) => {
      tx.oncomplete = () => resolve();
      tx.onerror = () => reject(tx.error);
      tx.onabort = () => reject(tx.error);
    });
    return result;
  } finally {
    db.close();
  }
}

export async function queryReadPermission(handle) {
  if (!handle || typeof handle.queryPermission !== 'function') return 'granted';
  try {
    return await handle.queryPermission({ mode: 'read' });
  } catch {
    return 'denied';
  }
}

export async function ensureReadPermission(handle) {
  if (!handle) return false;
  let state = await queryReadPermission(handle);
  if (state === 'granted') return true;
  if (state === 'prompt' && typeof handle.requestPermission === 'function') {
    try {
      state = await handle.requestPermission({ mode: 'read' });
    } catch {
      return false;
    }
  }
  return state === 'granted';
}

export async function listRecentFolders() {
  const rows = await getAllRecents();
  return rows
    .slice()
    .sort((a, b) => (b.openedAt || 0) - (a.openedAt || 0))
    .map((r) => ({ id: r.id, name: r.name || 'Folder' }));
}

export async function getRecentFolder(id) {
  if (!id) return null;
  try {
    return (await withStore(RECENTS_STORE, 'readonly', (store) => idbReq(store.get(id)))) || null;
  } catch {
    return null;
  }
}

async function getAllRecents() {
  try {
    return (await withStore(RECENTS_STORE, 'readonly', (store) => idbReq(store.getAll()))) || [];
  } catch {
    return [];
  }
}

async function putRecent(row) {
  await withStore(RECENTS_STORE, 'readwrite', (store) => idbReq(store.put(row)));
}

async function deleteRecent(id) {
  await withStore(RECENTS_STORE, 'readwrite', (store) => idbReq(store.delete(id)));
}

async function capRecents() {
  const all = await getAllRecents();
  const sorted = all.slice().sort((a, b) => (b.openedAt || 0) - (a.openedAt || 0));
  await Promise.all(sorted.slice(RECENTS_CAP).map((extra) => deleteRecent(extra.id)));
}

export async function rememberDirectory(handle) {
  if (!handle) return null;
  const name = handle.name || 'Folder';
  const openedAt = Date.now();
  try {
    const all = await getAllRecents();
    if (typeof handle.isSameEntry === 'function') {
      for (const row of all) {
        if (!row.handle) continue;
        try {
          if (await handle.isSameEntry(row.handle)) {
            const updated = { ...row, name, handle, openedAt };
            await putRecent(updated);
            await capRecents();
            return updated;
          }
        } catch {
          /* ignore compare errors */
        }
      }
    }
    const created = {
      id: typeof crypto !== 'undefined' && crypto.randomUUID ? crypto.randomUUID() : `f-${openedAt}`,
      name,
      handle,
      openedAt,
    };
    await putRecent(created);
    await capRecents();
    return created;
  } catch {
    return { id: '', name, handle, openedAt };
  }
}

export async function getSessionBinding(sessionId) {
  const sid = sessionId || 'default';
  try {
    return (await withStore(SESSIONS_STORE, 'readonly', (store) => idbReq(store.get(sid)))) || null;
  } catch {
    return null;
  }
}

export async function setSessionBinding(sessionId, { folderId, pinnedPaths }) {
  const sid = sessionId || 'default';
  const row = {
    sessionId: sid,
    folderId: folderId || '',
    pinnedPaths: Array.isArray(pinnedPaths) ? pinnedPaths : [],
  };
  try {
    await withStore(SESSIONS_STORE, 'readwrite', (store) => idbReq(store.put(row)));
  } catch {
    /* ignore */
  }
}

export async function clearSessionBinding(sessionId) {
  const sid = sessionId || 'default';
  try {
    await withStore(SESSIONS_STORE, 'readwrite', (store) => idbReq(store.delete(sid)));
  } catch {
    /* ignore */
  }
}

function emptyWorkspace() {
  return {
    folderName: '',
    files: [],
    pinnedPaths: [],
    folderId: '',
    reconnect: false,
  };
}

export async function loadWorkspaceForSession(sessionId) {
  const bind = await getSessionBinding(sessionId);
  if (!bind?.folderId) return emptyWorkspace();
  const rec = await getRecentFolder(bind.folderId);
  const pinnedPaths = Array.isArray(bind.pinnedPaths) ? bind.pinnedPaths : [];
  if (!rec) return emptyWorkspace();
  if (!rec.handle) {
    return {
      folderName: rec.name || '',
      files: [],
      pinnedPaths,
      folderId: rec.id,
      reconnect: false,
    };
  }
  const perm = await queryReadPermission(rec.handle);
  if (perm === 'granted') {
    try {
      const files = await walkDirectoryHandle(rec.handle);
      const live = new Set(files.map((f) => f.path));
      return {
        folderName: rec.name || rec.handle.name || '',
        files,
        pinnedPaths: pinnedPaths.filter((p) => live.has(p)),
        folderId: rec.id,
        reconnect: false,
      };
    } catch {
      return {
        folderName: rec.name || '',
        files: [],
        pinnedPaths,
        folderId: rec.id,
        reconnect: true,
      };
    }
  }
  return {
    folderName: rec.name || rec.handle.name || '',
    files: [],
    pinnedPaths,
    folderId: rec.id,
    reconnect: true,
  };
}

export async function reconnectRecentFolder(folderId) {
  const rec = await getRecentFolder(folderId);
  if (!rec?.handle) return null;
  const ok = await ensureReadPermission(rec.handle);
  if (!ok) return null;
  const files = await walkDirectoryHandle(rec.handle);
  return {
    folderName: rec.name || rec.handle.name || '',
    files,
    folderId: rec.id,
    handle: rec.handle,
  };
}
