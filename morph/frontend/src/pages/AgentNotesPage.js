import React, { useEffect, useState } from 'react';
import { tranApi, tranEndpoints } from '../api/tranClient';
import { getMorphToken } from '../auth/morphSession';
import '../App.css';

const THEME_KEY = 'skool-ai-chat-theme';

function noteTime(iso) {
  if (!iso) return '';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString();
}

function noteStatus(row) {
  return row.completed ? 'Done' : 'Open';
}

export default function AgentNotesPage() {
  const [notes, setNotes] = useState([]);
  const [selectedId, setSelectedId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    try {
      localStorage.setItem(THEME_KEY, 'dark');
    } catch {
      /* ignore */
    }
    document.documentElement.setAttribute('data-chat-theme', 'dark');
  }, []);

  useEffect(() => {
    if (!getMorphToken()) {
      setLoading(false);
      return;
    }
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

  const selected = notes.find((row) => row.id === selectedId) || null;

  return (
    <div className="agent-notes-page">
      <div className="agent-notes-page-inner">
        <header className="agent-notes-header">
          <h1>Agent notes</h1>
          <a className="agent-notes-home" href="/">
            Morph AI
          </a>
        </header>
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
  );
}
