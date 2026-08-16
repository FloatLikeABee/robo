import React from 'react';
import EventNoteOutlinedIcon from '@mui/icons-material/EventNoteOutlined';
import NotesTodosContent from '../notesTodos/NotesTodosContent';

/**
 * Right drawer for Morph AI — same notes/TODO data as admin header panel.
 */
export default function ChatNotesTodosDrawer({ open, onClose }) {
  if (!open) return null;

  const overlayClick = (e) => {
    if (e.target === e.currentTarget) onClose?.();
  };

  return (
    <div className="hybrid-drawer-overlay" role="presentation" onMouseDown={overlayClick}>
      <aside
        className="hybrid-drawer hybrid-drawer--notes-todos"
        aria-labelledby="chat-notes-drawer-title"
        onMouseDown={(e) => e.stopPropagation()}
        style={{ width: 'min(94vw, 900px)', maxWidth: '100%', alignSelf: 'stretch', minHeight: 0 }}
      >
        <div className="hybrid-drawer-head">
          <div>
            <h2 id="chat-notes-drawer-title" className="hybrid-drawer-title" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <EventNoteOutlinedIcon style={{ fontSize: 22, opacity: 0.9 }} />
              Notes &amp; TODOs
            </h2>
            <p className="hybrid-drawer-sub">Synced with MorphData — AI assist uses the same items as the admin toolbar.</p>
          </div>
          <button type="button" className="hybrid-drawer-close" onClick={onClose} aria-label="Close notes">
            ✕
          </button>
        </div>
        <div
          className="hybrid-drawer-scroll"
          style={{ padding: 0, flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}
        >
          <NotesTodosContent variant="chat" open={open} />
        </div>
      </aside>
    </div>
  );
}
