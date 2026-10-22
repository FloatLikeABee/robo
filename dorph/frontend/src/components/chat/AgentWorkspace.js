import React, { useEffect } from 'react';
import NotesTodosContent from '../notesTodos/NotesTodosContent';
import HybridContextDrawer from '../../HybridContextDrawer';

export const TABS = [
  { id: 'notes', label: 'Notes & TODOs' },
  { id: 'knowledge', label: 'Context & Knowledge' },
];

const VALID_TABS = new Set(TABS.map((t) => t.id));

export const WORKSPACE_OPEN_KEY = 'dorphai-workspace-open';

export function workspaceTabStorageKey(sessionId) {
  return `dorphai-workspace-tab:${sessionId || 'default'}`;
}

export function workspaceOpenStorageKey() {
  return WORKSPACE_OPEN_KEY;
}

export function readWorkspaceTab(sessionId) {
  try {
    const t = localStorage.getItem(workspaceTabStorageKey(sessionId))
      || localStorage.getItem(`morphai-workspace-tab:${sessionId || 'default'}`);
    if (t === 'files') return 'knowledge';
    if (VALID_TABS.has(t)) return t;
  } catch {
    /* ignore */
  }
  return null;
}

export function writeWorkspaceTab(sessionId, tab) {
  if (!VALID_TABS.has(tab)) return;
  try {
    localStorage.setItem(workspaceTabStorageKey(sessionId), tab);
  } catch {
    /* ignore */
  }
}

export function readWorkspaceOpen() {
  try {
    const v = localStorage.getItem(WORKSPACE_OPEN_KEY) || localStorage.getItem('morphai-workspace-open');
    if (v === '0') return false;
    if (v === '1') return true;
  } catch {
    /* ignore */
  }
  return true;
}

export function readStoredWorkspaceChoice() {
  try {
    const v = localStorage.getItem(WORKSPACE_OPEN_KEY) || localStorage.getItem('morphai-workspace-open');
    if (v === '0' || v === '1') return v;
  } catch {
    /* ignore */
  }
  return null;
}

export function initialWorkspaceOpen({ stored = null, phone = false } = {}) {
  if (stored === '0' || stored === false) return false;
  if (stored === '1' || stored === true) return true;
  return !phone;
}

export function writeWorkspaceOpen(open) {
  try {
    localStorage.setItem(WORKSPACE_OPEN_KEY, open ? '1' : '0');
  } catch {
    /* ignore */
  }
}

export const LAST_SESSION_KEY = 'dorphai-last-session';

export function readLastSessionId() {
  try {
    const v = localStorage.getItem(LAST_SESSION_KEY) || localStorage.getItem('morphai-last-session');
    if (typeof v === 'string' && v.trim()) return v.trim();
  } catch {
    /* ignore */
  }
  return '';
}

export function writeLastSessionId(sessionId) {
  const id = String(sessionId || '').trim();
  if (!id) return;
  try {
    localStorage.setItem(LAST_SESSION_KEY, id);
  } catch {
    /* ignore */
  }
}

export function resolveRestoredSessionId({ lastId, sessionIds } = {}) {
  const ids = Array.isArray(sessionIds) ? sessionIds.filter(Boolean) : [];
  const last = typeof lastId === 'string' ? lastId.trim() : '';
  if (last && ids.includes(last)) return last;
  if (ids.includes('default')) return 'default';
  return ids[0] || 'default';
}

export default function AgentWorkspace({
  sessionId,
  activeTab,
  onTabChange,
  onBringToConversation,
  onAttachmentChange,
}) {
  const tab = TABS.some((item) => item.id === activeTab) ? activeTab : 'knowledge';

  useEffect(() => {
    writeWorkspaceTab(sessionId, tab);
  }, [sessionId, tab]);

  return (
    <aside className="agent-workspace" aria-label="Agent workspace">
      <div className="agent-workspace-tabs" role="tablist">
        {TABS.map((item) => (
          <button
            key={item.id}
            type="button"
            role="tab"
            aria-selected={tab === item.id}
            className={`agent-workspace-tab${tab === item.id ? ' is-active' : ''}`}
            onClick={() => onTabChange?.(item.id)}
          >
            {item.label}
          </button>
        ))}
      </div>
      <div className="agent-workspace-body">
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
