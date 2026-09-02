import React, { useEffect } from 'react';
import NotesTodosContent from '../notesTodos/NotesTodosContent';
import HybridContextDrawer from '../../HybridContextDrawer';
import AgentFilesTab from './AgentFilesTab';

const TABS = [
  { id: 'files', label: 'Files' },
  { id: 'notes', label: 'Notes & TODOs' },
  { id: 'knowledge', label: 'Context & Knowledge' },
];

export function workspaceTabStorageKey(sessionId) {
  return `morphai-workspace-tab:${sessionId || 'default'}`;
}

export default function AgentWorkspace({
  sessionId,
  activeTab,
  onTabChange,
  folderName,
  files,
  pinnedPaths,
  onFolderOpened,
  onFolderCleared,
  onTogglePin,
  onOpenRecent,
  onReconnectFolder,
  reconnectNeeded,
  onBringToConversation,
  onAttachmentChange,
}) {
  const tab = TABS.some((t) => t.id === activeTab) ? activeTab : 'files';

  useEffect(() => {
    try {
      sessionStorage.setItem(workspaceTabStorageKey(sessionId), tab);
    } catch {
      /* ignore */
    }
  }, [sessionId, tab]);

  return (
    <aside className="agent-workspace" aria-label="Agent workspace">
      <div className="agent-workspace-tabs" role="tablist">
        {TABS.map((t) => (
          <button
            key={t.id}
            type="button"
            role="tab"
            aria-selected={tab === t.id}
            className={`agent-workspace-tab${tab === t.id ? ' is-active' : ''}`}
            onClick={() => onTabChange?.(t.id)}
          >
            {t.label}
          </button>
        ))}
      </div>
      <div className="agent-workspace-body">
        {tab === 'files' ? (
          <AgentFilesTab
            folderName={folderName}
            files={files}
            pinnedPaths={pinnedPaths}
            reconnectNeeded={reconnectNeeded}
            onFolderOpened={onFolderOpened}
            onFolderCleared={onFolderCleared}
            onTogglePin={onTogglePin}
            onOpenRecent={onOpenRecent}
            onReconnectFolder={onReconnectFolder}
          />
        ) : null}
        {tab === 'notes' ? <NotesTodosContent variant="chat" open /> : null}
        {tab === 'knowledge' ? (
          <HybridContextDrawer
            variant="panel"
            open
            sessionId={sessionId}
            onBringToConversation={onBringToConversation}
            onAttachmentChange={onAttachmentChange}
          />
        ) : null}
      </div>
    </aside>
  );
}
