import React, { useMemo, useState } from 'react';
import AutoAwesomeOutlinedIcon from '@mui/icons-material/AutoAwesomeOutlined';
import { getMorphToken } from '../../auth/morphSession';

const BK_URL = process.env.REACT_APP_BK_URL || 'http://localhost:3000';

/** Build the bk URL carrying the Morph session so the embedded app is authenticated. */
function bkHref(base) {
  const token = getMorphToken();
  if (!token) return base;
  try {
    const url = new URL(base, window.location.origin);
    url.searchParams.set('userspanel_token', token);
    return url.toString();
  } catch {
    return base;
  }
}

/**
 * AI Tools workspace — a large right drawer merged into Morph AI.
 * Embeds the full AI Tools (bk) UI so every module is available in-app:
 * Assistants, RAG, Documents, System.
 */
export default function AiToolsWorkspaceDrawer({ open, onClose, iframeRef, onFrameReady }) {
  const [failed, setFailed] = useState(false);
  const src = useMemo(() => bkHref(BK_URL), []);

  if (!open) return null;

  const overlayClick = (e) => {
    if (e.target === e.currentTarget) onClose?.();
  };

  return (
    <div className="hybrid-drawer-overlay" role="presentation" onMouseDown={overlayClick}>
      <aside
        className="hybrid-drawer ai-tools-workspace"
        aria-labelledby="ai-tools-workspace-title"
        onMouseDown={(e) => e.stopPropagation()}
        style={{ width: 'min(96vw, 1200px)', maxWidth: '100%', alignSelf: 'stretch', minHeight: 0 }}
      >
        <div className="hybrid-drawer-head">
          <div>
            <h2
              id="ai-tools-workspace-title"
              className="hybrid-drawer-title"
              style={{ display: 'flex', alignItems: 'center', gap: 8 }}
            >
              <AutoAwesomeOutlinedIcon style={{ fontSize: 22, opacity: 0.9 }} />
              AI Tools
            </h2>
            <p className="hybrid-drawer-sub">Assistants, RAG, Documents &amp; more — inside Morph AI.</p>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <button type="button" className="hybrid-drawer-close" onClick={onClose} aria-label="Close AI Tools">
              ✕
            </button>
          </div>
        </div>

        <div className="ai-tools-frame-wrap">
          {failed ? (
            <div className="ai-tools-state ai-tools-state--error" role="alert">
              <span>
                Couldn’t load the AI Tools app at <code>{BK_URL}</code>. Make sure the AI Tools UI is running
                (<code>./start-all.sh start bk-ui</code>).
              </span>
            </div>
          ) : (
            <iframe
              ref={iframeRef}
              title="AI Tools"
              src={src}
              className="ai-tools-frame"
              onLoad={() => onFrameReady?.()}
              onError={() => setFailed(true)}
            />
          )}
        </div>
      </aside>
    </div>
  );
}
