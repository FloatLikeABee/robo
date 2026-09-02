import React, { useCallback, useEffect, useRef, useState } from 'react';
import { listRecentFolders, walkDirectoryHandle } from '../../lib/filesWorkspaceStore';

function filesFromWebkitList(list) {
  return Array.from(list || []).map((file) => {
    const rel = file.webkitRelativePath || file.name;
    const parts = rel.split('/').filter(Boolean);
    const skipDir = parts.slice(0, -1).some((p) => p === '.git' || p === 'node_modules' || (p.startsWith('.') && p !== '.'));
    const name = parts[parts.length - 1] || file.name;
    const skipped =
      skipDir || !name || name.startsWith('.') || file.size > 512 * 1024;
    return {
      path: rel,
      size: file.size,
      skipped,
      file,
    };
  });
}

export default function AgentFilesTab({
  folderName,
  files,
  pinnedPaths,
  reconnectNeeded,
  onFolderOpened,
  onFolderCleared,
  onTogglePin,
  onOpenRecent,
  onReconnectFolder,
}) {
  const webkitRef = useRef(null);
  const [status, setStatus] = useState('');
  const [recents, setRecents] = useState([]);
  const pinned = pinnedPaths instanceof Set ? pinnedPaths : new Set(pinnedPaths || []);

  const refreshRecents = useCallback(async () => {
    try {
      const list = await listRecentFolders();
      setRecents(list);
    } catch {
      setRecents([]);
    }
  }, []);

  useEffect(() => {
    void refreshRecents();
  }, [refreshRecents, folderName]);

  const openWithPicker = useCallback(async () => {
    setStatus('');
    if (typeof window.showDirectoryPicker !== 'function') {
      webkitRef.current?.click();
      return;
    }
    try {
      const handle = await window.showDirectoryPicker({ mode: 'read' });
      const listed = await walkDirectoryHandle(handle);
      onFolderOpened?.({ name: handle.name, files: listed, directoryHandle: handle });
      void refreshRecents();
    } catch (err) {
      if (err && err.name === 'AbortError') {
        setStatus('No folder selected.');
        return;
      }
      setStatus('Could not open a folder here. Try the file-picker fallback.');
      webkitRef.current?.click();
    }
  }, [onFolderOpened, refreshRecents]);

  const onWebkitChange = (ev) => {
    const list = ev.target.files;
    if (!list?.length) {
      setStatus('No folder selected.');
      ev.target.value = '';
      return;
    }
    const listed = filesFromWebkitList(list);
    const top = listed[0]?.path?.split('/')[0] || 'Folder';
    onFolderOpened?.({ name: top, files: listed, directoryHandle: null });
    ev.target.value = '';
  };

  const chooseRecent = async (id) => {
    setStatus('');
    try {
      await onOpenRecent?.(id);
      void refreshRecents();
    } catch {
      setStatus('Could not open that folder.');
    }
  };

  const reconnect = async () => {
    setStatus('');
    const ok = await onReconnectFolder?.();
    if (!ok) setStatus('Allow access to reopen this folder.');
  };

  const recentChoices = recents.filter((r) => r.name);

  return (
    <div className="agent-files-tab">
      <div className="agent-files-actions">
        <button type="button" className="hybrid-primary-btn" onClick={openWithPicker}>
          Open folder
        </button>
        {folderName ? (
          <button type="button" className="hybrid-toolbar-btn" onClick={() => onFolderCleared?.()}>
            Close folder
          </button>
        ) : null}
        {folderName && reconnectNeeded ? (
          <button type="button" className="hybrid-toolbar-btn" onClick={reconnect}>
            Reconnect
          </button>
        ) : null}
        <input
          ref={webkitRef}
          type="file"
          className="hybrid-hidden"
          webkitdirectory=""
          directory=""
          multiple
          onChange={onWebkitChange}
        />
      </div>
      {recentChoices.length ? (
        <div className="agent-files-recents">
          <div className="agent-files-recents-label">Recent</div>
          <ul className="agent-files-recents-list">
            {recentChoices.map((r) => (
              <li key={r.id}>
                <button type="button" className="agent-files-recent-btn" onClick={() => chooseRecent(r.id)}>
                  {r.name}
                </button>
              </li>
            ))}
          </ul>
        </div>
      ) : null}
      {status ? <div className="hybrid-drawer-status">{status}</div> : null}
      {!folderName ? (
        <div className="agent-files-empty">No folder open.</div>
      ) : (
        <>
          <h3 className="hybrid-h3">{folderName}</h3>
          {reconnectNeeded && !(files || []).length ? (
            <div className="agent-files-empty">Reconnect to list files in this folder.</div>
          ) : (
            <ul className="agent-file-tree">
              {(files || []).map((f) => (
                <li key={f.path} className={`agent-file-row${f.skipped ? ' is-skipped' : ''}`}>
                  <span className="agent-file-path" title={f.path}>
                    {f.path}
                  </span>
                  {f.skipped ? (
                    <span className="hybrid-file-chunks">skipped</span>
                  ) : (
                    <button
                      type="button"
                      className={`hybrid-toolbar-btn${pinned.has(f.path) ? ' hybrid-toolbar-btn-primary' : ''}`}
                      onClick={() => onTogglePin?.(f)}
                    >
                      {pinned.has(f.path) ? 'Unpin' : 'Pin'}
                    </button>
                  )}
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </div>
  );
}
