import React, { useEffect, useMemo, useRef } from 'react';
import { bindFrameLocale, getLocale, postLocaleToFrame, withLang } from '../../lib/locale';
import { useT } from '../../lib/localeReact';

/** DorphNotes over the chat. Same-origin iframe reuses /dorphdata without a second router. */
export default function DorphNotesModal({ open, onClose }) {
  const t = useT();
  const frameRef = useRef(null);
  const src = useMemo(() => (open ? withLang('/dorphdata', getLocale()) : ''), [open]);
  useEffect(() => {
    if (!open) return undefined;
    return bindFrameLocale(frameRef.current);
  }, [open]);

  if (!open) return null;

  const overlayClick = (e) => {
    if (e.target === e.currentTarget) onClose?.();
  };

  return (
    <div className="app-shell-modal-overlay" role="presentation" onMouseDown={overlayClick}>
      <aside
        className="hybrid-drawer app-shell-modal app-shell-modal--full"
        aria-labelledby="dorphnotes-modal-title"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="hybrid-drawer-head">
          <h2 id="dorphnotes-modal-title" className="hybrid-drawer-title">
            {t('dorphNotes')}
          </h2>
          <button type="button" className="hybrid-drawer-close" onClick={onClose} aria-label={t('closeDorphNotes')}>
            ✕
          </button>
        </div>
        <iframe
          ref={frameRef}
          title={t('dorphNotes')}
          src={src}
          className="app-shell-modal-frame"
          onLoad={() => postLocaleToFrame(frameRef.current, getLocale())}
        />
      </aside>
    </div>
  );
}
