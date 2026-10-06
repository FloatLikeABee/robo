import React, { useEffect, useMemo, useRef } from 'react';
import { bindFrameLocale, getLocale, postLocaleToFrame, withLang } from '../../lib/locale';
import { useT } from '../../lib/localeReact';

/** MorphNotes over the chat. Same-origin iframe reuses /morphdata without a second router. */
export default function MorphNotesModal({ open, onClose }) {
  const t = useT();
  const frameRef = useRef(null);
  const src = useMemo(() => (open ? withLang('/morphdata', getLocale()) : ''), [open]);
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
        aria-labelledby="morphnotes-modal-title"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="hybrid-drawer-head">
          <h2 id="morphnotes-modal-title" className="hybrid-drawer-title">
            {t('morphNotes')}
          </h2>
          <button type="button" className="hybrid-drawer-close" onClick={onClose} aria-label={t('closeMorphNotes')}>
            ✕
          </button>
        </div>
        <iframe
          ref={frameRef}
          title={t('morphNotes')}
          src={src}
          className="app-shell-modal-frame"
          onLoad={() => postLocaleToFrame(frameRef.current, getLocale())}
        />
      </aside>
    </div>
  );
}
