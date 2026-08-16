import React, { useCallback, useEffect, useRef, useState } from 'react';
import { tranApi } from './api/tranClient';

/**
 * Right-side drawer: session HybridContext (files + notes) and durable Knowledge Library.
 */
export default function HybridContextDrawer({ open, onClose, sessionId, onBringToConversation, onAttachmentChange }) {
  const [panel, setPanel] = useState('session'); // session | knowledge
  const [chunkCount, setChunkCount] = useState(0);
  const [sources, setSources] = useState([]);
  const [status, setStatus] = useState('');
  const [bringLoading, setBringLoading] = useState(false);
  const [pasteText, setPasteText] = useState('');
  const [knowledgeFiles, setKnowledgeFiles] = useState([]);
  const [knowledgeStatus, setKnowledgeStatus] = useState('');
  const [knowledgeIndexGraph, setKnowledgeIndexGraph] = useState(true);
  const [sessionIndexGraph, setSessionIndexGraph] = useState(false);
  const filesRef = useRef(null);
  const knowledgeRef = useRef(null);

  const refreshMeta = useCallback(async () => {
    try {
      const { data } = await tranApi.get('/api/chat/hybrid-context', { params: { session_id: sessionId } });
      const src = Array.isArray(data.sources) ? data.sources : [];
      const total = src.reduce((acc, s) => acc + (Number(s.chunk_count) || 0), 0);
      setChunkCount(total);
      setSources(src);
      if (typeof onAttachmentChange === 'function') {
        onAttachmentChange();
      }
    } catch {
      setChunkCount(0);
      setSources([]);
      if (typeof onAttachmentChange === 'function') {
        onAttachmentChange();
      }
    }
  }, [sessionId, onAttachmentChange]);

  useEffect(() => {
    if (!open) return;
    refreshMeta();
    setStatus('');
    if (panel === 'knowledge') {
      void refreshKnowledge();
    }
  }, [open, refreshMeta, panel]);

  const refreshKnowledge = useCallback(async () => {
    try {
      const { data } = await tranApi.get('/api/knowledge/files');
      setKnowledgeFiles(Array.isArray(data.files) ? data.files : []);
    } catch (e) {
      setKnowledgeFiles([]);
      setKnowledgeStatus(e.response?.data?.error || e.message || 'Knowledge Library unavailable');
    }
  }, []);

  const onUploadKnowledge = async (ev) => {
    const fs = ev.target.files;
    if (!fs?.length) return;
    setKnowledgeStatus('');
    try {
      for (const f of Array.from(fs)) {
        const fd = new FormData();
        fd.append('file', f);
        fd.append('title', f.name);
        fd.append('index_to_graph', knowledgeIndexGraph ? 'true' : 'false');
        await tranApi.post('/api/knowledge/files', fd, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
      }
      setKnowledgeStatus(
        knowledgeIndexGraph
          ? 'Uploaded to Knowledge Library and queued for Neo4j GraphRAG.'
          : 'Uploaded to Knowledge Library (MySQL chunks only; Neo4j skipped).'
      );
      await refreshKnowledge();
    } catch (e) {
      setKnowledgeStatus(e.response?.data?.error || e.message || 'Upload failed.');
    }
    ev.target.value = '';
  };

  const onDeleteKnowledge = async (id) => {
    setKnowledgeStatus('');
    try {
      await tranApi.delete(`/api/knowledge/files/${id}`);
      await refreshKnowledge();
    } catch (e) {
      setKnowledgeStatus(e.response?.data?.error || e.message || 'Delete failed.');
    }
  };

  const onUploadFiles = async (ev) => {
    const fs = ev.target.files;
    if (!fs?.length) return;
    const fd = new FormData();
    fd.append('session_id', sessionId);
    fd.append('index_to_graph', sessionIndexGraph ? 'true' : 'false');
    Array.from(fs).forEach((f) => fd.append('files', f));
    setStatus('');
    try {
      const { data } = await tranApi.post('/api/chat/hybrid-context/files', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      const graphN = Array.isArray(data?.graph_saved) ? data.graph_saved.length : 0;
      setStatus(
        sessionIndexGraph && graphN > 0
          ? `Files ingested. Also saved ${graphN} to Knowledge Library / Neo4j.`
          : 'Files ingested.'
      );
      refreshMeta();
      if (sessionIndexGraph) void refreshKnowledge();
    } catch (e) {
      setStatus(e.response?.data?.error || e.message || 'Upload failed.');
    }
    ev.target.value = '';
  };

  const fileSources = sources.filter((s) => s.kind === 'file');

  const onRemoveFile = async (source) => {
    setStatus('');
    try {
      await tranApi.delete('/api/chat/hybrid-context/source', {
        data: { session_id: sessionId, kind: source.kind, label: source.label },
      });
      await refreshMeta();
      setStatus(`Removed ${source.label}.`);
    } catch (e) {
      setStatus(e.response?.data?.error || e.message || 'Remove failed.');
    }
  };

  const onApplyPaste = async () => {
    const len = [...pasteText].length;
    if (len < 50) {
      setStatus(`Pasted text must be at least 50 characters (${len}).`);
      return;
    }
    setStatus('');
    try {
      await tranApi.post('/api/chat/hybrid-context/paste', { session_id: sessionId, text: pasteText });
      setStatus('Notes applied to HybridContext.');
      refreshMeta();
    } catch (e) {
      setStatus(e.response?.data?.error || e.message || 'Apply failed.');
    }
  };

  const onClearAll = async () => {
    setStatus('');
    try {
      await tranApi.delete('/api/chat/hybrid-context', { params: { session_id: sessionId } });
      refreshMeta();
      setStatus('HybridContext cleared for this chat.');
    } catch {
      setStatus('Clear failed.');
    }
  };

  const onBringClick = async () => {
    if (typeof onBringToConversation !== 'function') return;
    setBringLoading(true);
    setStatus('');
    try {
      await onBringToConversation();
      onClose?.();
    } catch (e) {
      setStatus(e.response?.data?.error || e.message || 'Could not add to conversation.');
    } finally {
      setBringLoading(false);
    }
  };

  if (!open) return null;

  const overlayClick = (e) => {
    if (e.target === e.currentTarget) onClose();
  };


  return (
    <div className="hybrid-drawer-overlay" role="presentation" onMouseDown={overlayClick}>
      <aside className="hybrid-drawer" aria-labelledby="hybrid-drawer-title" onMouseDown={(e) => e.stopPropagation()}>
        <div className="hybrid-drawer-head">
          <div>
            <h2 id="hybrid-drawer-title" className="hybrid-drawer-title">
              Context &amp; Knowledge
            </h2>
            <p className="hybrid-drawer-sub">
              Session HybridContext is temporary. Knowledge Library is durable GraphRAG for Morph AI.
            </p>
          </div>
          <button type="button" className="hybrid-drawer-close" onClick={onClose} aria-label="Close">
            ✕
          </button>
        </div>

        <div className="hybrid-drawer-toolbar" style={{ gap: 8 }}>
          <button
            type="button"
            className={`hybrid-toolbar-btn ${panel === 'session' ? 'hybrid-toolbar-btn-primary' : ''}`}
            onClick={() => setPanel('session')}
          >
            Session context
          </button>
          <button
            type="button"
            className={`hybrid-toolbar-btn ${panel === 'knowledge' ? 'hybrid-toolbar-btn-primary' : ''}`}
            onClick={() => setPanel('knowledge')}
          >
            Knowledge Library
          </button>
        </div>

        {panel === 'knowledge' ? (
          <div className="hybrid-drawer-scroll">
            {knowledgeStatus && <div className="hybrid-drawer-status">{knowledgeStatus}</div>}
            <section className="hybrid-drawer-section">
              <h3 className="hybrid-h3">Morph Knowledge Library</h3>
              <p className="hybrid-hint">Durable uploads for GraphRAG: md, json, csv, txt, pdf. Morph-only.</p>
              <label className="hybrid-check">
                <input
                  type="checkbox"
                  checked={knowledgeIndexGraph}
                  onChange={(e) => setKnowledgeIndexGraph(e.target.checked)}
                />
                <span>Also save into Neo4j graph (faster GraphRAG search)</span>
              </label>
              <input
                ref={knowledgeRef}
                type="file"
                className="hybrid-hidden"
                multiple
                accept=".md,.markdown,.json,.csv,.txt,.pdf"
                onChange={onUploadKnowledge}
              />
              <button type="button" className="hybrid-primary-btn" onClick={() => knowledgeRef.current?.click()}>
                Upload to Knowledge…
              </button>
              <ul className="hybrid-file-list">
                {knowledgeFiles.length === 0 ? (
                  <li className="hybrid-hint">No knowledge files yet.</li>
                ) : (
                  knowledgeFiles.map((f) => (
                    <li key={f.id} className="hybrid-file-row">
                      <div className="hybrid-file-meta">
                        <span className="hybrid-file-name" title={f.filename}>
                          {f.title || f.filename}
                        </span>
                        <span className="hybrid-file-chunks">
                          {f.kind} · {f.byte_size} bytes
                        </span>
                      </div>
                      <button type="button" className="hybrid-file-remove" onClick={() => onDeleteKnowledge(f.id)}>
                        Remove
                      </button>
                    </li>
                  ))
                )}
              </ul>
            </section>
          </div>
        ) : (
          <>
        <div className="hybrid-drawer-toolbar">
          {typeof onBringToConversation === 'function' && (
            <button
              type="button"
              className="hybrid-toolbar-btn hybrid-toolbar-btn-primary"
              disabled={bringLoading || chunkCount === 0}
              onClick={onBringClick}
              title={chunkCount === 0 ? 'Add HybridContext first' : 'Attach sources as chat reference (shown above input)'}
            >
              {bringLoading ? '…' : 'Bring to conversation'}
            </button>
          )}
          <button type="button" className="hybrid-toolbar-btn hybrid-toolbar-btn-danger" onClick={onClearAll}>
            Clear all
          </button>
        </div>

        <div className="hybrid-drawer-scroll">
          {status && <div className="hybrid-drawer-status">{status}</div>}

          <section className="hybrid-drawer-section">
            <h3 className="hybrid-h3">Data files</h3>
            <p className="hybrid-hint">CSV, JSON, TXT, Markdown, HTML, Excel (.xlsx), PDF, images — session only</p>
            <label className="hybrid-check">
              <input
                type="checkbox"
                checked={sessionIndexGraph}
                onChange={(e) => setSessionIndexGraph(e.target.checked)}
              />
              <span>Also save into Neo4j graph (Knowledge Library + GraphRAG)</span>
            </label>
            <input ref={filesRef} type="file" className="hybrid-hidden" multiple accept=".csv,.json,.txt,.md,.markdown,.html,.htm,.xlsx,.xlsm,.pdf,image/*" onChange={onUploadFiles} />
            <button type="button" className="hybrid-primary-btn" onClick={() => filesRef.current?.click()}>
              Choose files…
            </button>
            {fileSources.length > 0 && (
              <ul className="hybrid-file-list">
                {fileSources.map((source) => (
                  <li key={`${source.kind}:${source.label}`} className="hybrid-file-row">
                    <div className="hybrid-file-meta">
                      <span className="hybrid-file-name" title={source.label}>
                        {source.label}
                      </span>
                      <span className="hybrid-file-chunks">
                        {source.chunk_count} chunk{source.chunk_count === 1 ? '' : 's'}
                      </span>
                    </div>
                    <button
                      type="button"
                      className="hybrid-file-remove"
                      onClick={() => onRemoveFile(source)}
                      aria-label={`Remove ${source.label}`}
                    >
                      Remove
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </section>

          <section className="hybrid-drawer-section">
            <h3 className="hybrid-h3">Paste notes</h3>
            <textarea
              className="hybrid-textarea"
              rows={8}
              placeholder="Minimum 50 characters. Applied as HybridContext excerpts for this chat session."
              value={pasteText}
              onChange={(e) => setPasteText(e.target.value)}
              spellCheck
            />
            <button type="button" className="hybrid-primary-btn" disabled={[...pasteText].length < 50} onClick={onApplyPaste}>
              Apply to HybridContext
            </button>
          </section>
        </div>
          </>
        )}
      </aside>
    </div>
  );
}
