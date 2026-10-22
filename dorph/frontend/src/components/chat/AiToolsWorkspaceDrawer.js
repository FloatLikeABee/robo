import React, { useEffect, useMemo, useState } from 'react';
import AutoAwesomeOutlinedIcon from '@mui/icons-material/AutoAwesomeOutlined';
import { getDorphToken } from '../../auth/dorphSession';
import { bindFrameLocale, getLocale, postLocaleToFrame, withLang } from '../../lib/locale';
import { useT } from '../../lib/localeReact';

const BK_URL = process.env.REACT_APP_BK_URL || 'http://localhost:3000';

/** Build the bk URL carrying the Dorph session so the embedded app is authenticated. */
function bkHref(base) {
  try {
    const url = new URL(base, window.location.origin);
    const token = getDorphToken();
    if (token) url.searchParams.set('userspanel_token', token);
    return withLang(url.toString(), getLocale());
  } catch {
    return base;
  }
}

/**
 * DorphTools workspace — a large right drawer merged into Dorph AI.
 * Embeds the full DorphTools (bk) UI so every module is available in-app:
 * Assistants, RAG, Documents, System.
 */
export default function AiToolsWorkspaceDrawer({ open, onClose, iframeRef, onFrameReady }) {
  const t = useT();
  const [failed, setFailed] = useState(false);
  const [loading, setLoading] = useState(true);
  const [wasOpen, setWasOpen] = useState(open);
  const src = useMemo(() => (open ? bkHref(BK_URL) : ''), [open]);

  useEffect(() => {
    if (!open) return undefined;
    return bindFrameLocale(iframeRef?.current);
  }, [open, iframeRef, src]);

  if (open !== wasOpen) {
    setWasOpen(open);
    if (open) setLoading(true);
  }

  if (!open) return null;

  const overlayClick = (e) => {
    if (e.target === e.currentTarget) onClose?.();
  };

  return (
    <div className="app-shell-modal-overlay" role="presentation" onMouseDown={overlayClick}>
      <aside
        className="hybrid-drawer ai-tools-workspace app-shell-modal app-shell-modal--full"
        aria-labelledby="ai-tools-workspace-title"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="hybrid-drawer-head">
          <div>
            <h2
              id="ai-tools-workspace-title"
              className="hybrid-drawer-title"
              style={{ display: 'flex', alignItems: 'center', gap: 8 }}
            >
              <AutoAwesomeOutlinedIcon style={{ fontSize: 22, opacity: 0.9 }} />
              {t('dorphTools')}
            </h2>
            <p className="hybrid-drawer-sub">{t('dorphToolsSub')}</p>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <button type="button" className="hybrid-drawer-close" onClick={onClose} aria-label={t('closeDorphTools')}>
              ✕
            </button>
          </div>
        </div>

        <div className="ai-tools-frame-wrap">
          {failed ? (
            <div className="ai-tools-state ai-tools-state--error" role="alert">
              <span>
                {t('dorphToolsFail', { url: BK_URL })}
                (<code>./start-all.sh start bk-ui</code>).
              </span>
            </div>
          ) : (
            <iframe
              ref={iframeRef}
              title={t('dorphTools')}
              src={src}
              className="ai-tools-frame"
              onLoad={() => {
                setLoading(false);
                onFrameReady?.();
                postLocaleToFrame(iframeRef?.current, getLocale());
              }}
              onError={() => {
                setLoading(false);
                setFailed(true);
              }}
            />
          )}
          {loading && !failed ? (
            <div className="ai-tools-loading" role="status" aria-live="polite">
              <span className="ai-tools-spinner" aria-hidden />
              {t('loadingDorphTools')}
            </div>
          ) : null}
        </div>
      </aside>
    </div>
  );
}
