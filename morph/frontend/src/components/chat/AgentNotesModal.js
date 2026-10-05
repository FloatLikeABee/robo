import React, { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { onSheetKeyDown } from '../../lib/phoneSheet';

// ponytail: overlay chrome is copied from the AI tools drawer, with the dialog
// contract from onSheetKeyDown. Ceiling: Skills and AI tools keep their own
// dismiss paths. Upgrade: one shell if a fourth overlay needs this focus contract.

function noteTime(iso) {
  if (!iso) return '';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString();
}

function noteStatus(row) {
  return row.completed ? 'Done' : 'Open';
}

function focusIfConnected(el) {
  if (el && el.isConnected && typeof el.focus === 'function') el.focus();
}

export default function AgentNotesModal({ onClose, restoreFocusRef, fallbackFocusRef }) {
  const [notes, setNotes] = useState([]);
  const [selectedId, setSelectedId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const closeRef = useRef(null);
  const panelRef = useRef(null);

  useEffect(() => {
    let live = true;
    tranApi
      .get(tranEndpoints.agentNotes)
      .then((res) => {
        if (!live) return;
        setNotes(Array.isArray(res.data) ? res.data : []);
      })
      .catch((err) => {
        if (!live) return;
        setError(err.response?.data?.error || err.message || 'Could not load notes');
        setNotes([]);
      })
      .finally(() => {
        if (live) setLoading(false);
      });
    return () => {
      live = false;
    };
  }, []);

  useEffect(() => {
    const restore = restoreFocusRef?.current;
    const fallback = fallbackFocusRef?.current;
    closeRef.current?.focus();
    const onKey = (event) => {
      if (event.key === 'Escape' || event.key === 'Tab') event.stopPropagation();
      if (event.key === 'Escape') {
        event.preventDefault();
        onClose();
        return;
      }
      onSheetKeyDown(event, panelRef.current, onClose);
    };
    window.addEventListener('keydown', onKey, true);
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      window.removeEventListener('keydown', onKey, true);
      document.body.style.overflow = prevOverflow;
      if (restore && restore.isConnected) restore.focus();
      else focusIfConnected(fallback);
    };
  }, [onClose, restoreFocusRef, fallbackFocusRef]);

  const selected = notes.find((row) => row.id === selectedId) || null;

  // The agent shell sets .chat-container to display:contents, so an in-tree
  // overlay becomes a grid item and gets clipped. Portal to body like a real overlay.
  return createPortal(
    <div
      className="hybrid-drawer-overlay"
      role="presentation"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose();
      }}
    >
      <div
        ref={panelRef}
        className="hybrid-drawer agent-notes-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="agent-notes-title"
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="hybrid-drawer-head">
          <h2 id="agent-notes-title" className="hybrid-drawer-title">
            Agent notes
          </h2>
          <button
            ref={closeRef}
            type="button"
            className="hybrid-drawer-close"
            onClick={onClose}
            aria-label="Close agent notes"
          >
            ✕
          </button>
        </div>
        <div className="hybrid-drawer-scroll">
          {error ? (
            <p className="agent-notes-error" role="alert">
              {error}
            </p>
          ) : null}
          <div className={`agent-notes-layout${selected ? ' is-detail' : ''}`}>
            <div className="agent-notes-list">
              {loading ? <p className="agent-notes-empty">Loading notes…</p> : null}
              {!loading && !error && notes.length === 0 ? (
                <p className="agent-notes-empty">
                  Agents post notes for this sign-in with the Morph note tool. Nothing from an agent is here yet.
                </p>
              ) : null}
              {notes.map((row) => (
                <button
                  key={row.id}
                  type="button"
                  className={`agent-notes-row${row.id === selectedId ? ' is-selected' : ''}`}
                  onClick={() => setSelectedId(row.id)}
                >
                  <span className="agent-notes-title">{row.title || '(Note)'}</span>
                  <span className="agent-notes-meta">
                    {noteTime(row.created_on)}
                    {noteTime(row.created_on) ? ' · ' : ''}
                    {noteStatus(row)}
                  </span>
                </button>
              ))}
            </div>
            <div className="agent-notes-detail">
              <button type="button" className="agent-notes-back" onClick={() => setSelectedId(null)}>
                Back
              </button>
              {selected ? (
                <>
                  <h2>{selected.title || '(Note)'}</h2>
                  <p className="agent-notes-meta">
                    {noteTime(selected.created_on)}
                    {noteTime(selected.created_on) ? ' · ' : ''}
                    {noteStatus(selected)}
                  </p>
                  <div className="agent-notes-body">{selected.body || ''}</div>
                </>
              ) : (
                <p className="agent-notes-empty">Select a note to read it.</p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
}
