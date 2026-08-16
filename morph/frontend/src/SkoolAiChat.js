import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import axios from 'axios';
import ChatMarkdown from './ChatMarkdown';
import HybridContextDrawer from './HybridContextDrawer';
import ChatNotesTodosDrawer from './components/chat/ChatNotesTodosDrawer';
import SkillsModal from './components/SkillsModal';
import { tranApi } from './api/tranClient';
import { runAiProgress } from './lib/aiProgress';
import { getMorphToken } from './auth/morphSession';
import { useConfirm } from './components/ConfirmDialog';
import './App.css';
const THEME_KEY = 'skool-ai-chat-theme';
const PROMPT_HISTORY_KEY = 'skool-ai-chat-prompt-history';
const MAX_PROMPTS = 10;
const APP_URLS = {
  morphData: process.env.REACT_APP_MORPHDATA_URL || '/morphdata',
  morphUtils: process.env.REACT_APP_MORPH_UTILS_URL || 'http://localhost:3040',
  bk: process.env.REACT_APP_BK_URL || 'http://localhost:3000',
};

const HEADER_APP_ICONS = {
  morphdata: `${process.env.PUBLIC_URL || ''}/icons/morph-data-icon.svg`,
  morphutils: `${process.env.PUBLIC_URL || ''}/icons/morph-utils-icon.svg`,
  bk: `${process.env.PUBLIC_URL || ''}/icons/bk-icon.svg`,
};

function appHrefWithSession(baseUrl) {
  const token = getMorphToken();
  if (!token) return baseUrl;
  try {
    const url = new URL(baseUrl, window.location.origin);
    url.searchParams.set('userspanel_token', token);
    return url.toString();
  } catch {
    return baseUrl;
  }
}

function getStoredChatTheme() {
  try {
    const v = localStorage.getItem(THEME_KEY);
    if (v === 'light' || v === 'dark') return v;
  } catch {}
  return 'dark';
}

function loadPromptHistory() {
  try {
    const raw = localStorage.getItem(PROMPT_HISTORY_KEY);
    if (!raw) return [];
    const arr = JSON.parse(raw);
    return Array.isArray(arr) ? arr.filter((x) => typeof x === 'string').slice(0, MAX_PROMPTS) : [];
  } catch {
    return [];
  }
}

function savePromptHistory(arr) {
  try {
    localStorage.setItem(PROMPT_HISTORY_KEY, JSON.stringify(arr.slice(0, MAX_PROMPTS)));
  } catch {}
}

function isHybridConversationUserMessage(content) {
  return typeof content === 'string' && content.startsWith('# HybridContext (in this conversation)');
}

function formatResponseTime(ms) {
  if (ms == null) return '';
  if (ms < 1000) return `${ms} ms`;
  return `${(ms / 1000).toFixed(ms < 10000 ? 1 : 0)} s`;
}

/**
 * Skool AI chat — full page or embedded (e.g. admin drawer).
 * @param {{ variant?: 'page' | 'embedded', enableFileUpload?: boolean, singleSession?: boolean }} props
 * When singleSession is true, only the `default` session is used and the session sidebar is hidden.
 */
export default function SkoolAiChat({ variant = 'page', enableFileUpload = true, singleSession = false }) {
  const { confirm } = useConfirm();
  const embedded = variant === 'embedded';
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingStatus, setLoadingStatus] = useState('');
  const [responseTimeMs, setResponseTimeMs] = useState(null);
  const [attachedFile, setAttachedFile] = useState(null);
  const [sessions, setSessions] = useState([]);
  const [currentSessionId, setCurrentSessionId] = useState('default');
  const [editingSessionId, setEditingSessionId] = useState(null);
  const [editTitleValue, setEditTitleValue] = useState('');
  const [theme, setTheme] = useState(getStoredChatTheme);
  const [sidebarNavOpen, setSidebarNavOpen] = useState(false);
  const [hybridDrawerOpen, setHybridDrawerOpen] = useState(false);
  const [hybridAttachment, setHybridAttachment] = useState({ attached: false, title: '', sources: [] });
  const [notesDrawerOpen, setNotesDrawerOpen] = useState(false);
  const [skillsOpen, setSkillsOpen] = useState(false);
  const sessionId = singleSession ? 'default' : currentSessionId || 'default';
  const messagesEndRef = useRef(null);
  const inputRef = useRef(null);
  const fileInputRef = useRef(null);
  const sessionTitleInputRef = useRef(null);
  const promptHistoryRef = useRef(loadPromptHistory());
  const historyNavIndexRef = useRef(-1);
  const draftBeforeNavRef = useRef('');
  const wasLoadingRef = useRef(false);
  const abortControllerRef = useRef(null);

  useEffect(() => {
    document.documentElement.setAttribute('data-chat-theme', theme);
    try {
      localStorage.setItem(THEME_KEY, theme);
    } catch {}
    try {
      window.dispatchEvent(new CustomEvent('morph-chat-theme', { detail: theme }));
    } catch {}
  }, [theme]);

  const storedToMessage = (m) => {
    const type = m.role === 'user' ? 'user' : m.role === 'error' ? 'error' : 'assistant';
    const base = { type, content: m.content || '' };
    if (type === 'assistant') {
      if (m.sql) base.sql = m.sql;
      if (m.confirmation_card) base.confirmationCard = m.confirmation_card;
      if (m.proposed_form) base.proposedForm = m.proposed_form;
      if (m.research_content) base.researchContent = m.research_content;
      if (Array.isArray(m.images) && m.images.length) base.images = m.images;
    }
    return base;
  };


  const loadSessions = async () => {
    try {
      const res = await tranApi.get('/api/chat/sessions');
      setSessions(Array.isArray(res.data) ? res.data : []);
    } catch (e) {
      console.error('Load sessions error:', e);
      setSessions([]);
    }
  };

  const loadSessionMessages = async (sessionId) => {
    try {
      const res = await tranApi.get(`/api/chat/sessions/${sessionId}`);
      const raw = res.data?.messages;
      const list = Array.isArray(raw) ? raw : [];
      setMessages(list.map(storedToMessage));
    } catch (e) {
      console.error('Load session messages error:', e);
      setMessages([]);
    }
  };

  const loadHybridAttachment = useCallback(async () => {
    try {
      const { data } = await tranApi.get('/api/chat/hybrid-context', { params: { session_id: sessionId } });
      setHybridAttachment({
        attached: !!data.attached,
        title: typeof data.attachment_title === 'string' ? data.attachment_title : '',
        sources: Array.isArray(data.sources) ? data.sources : [],
      });
    } catch {
      setHybridAttachment({ attached: false, title: '', sources: [] });
    }
  }, [sessionId]);

  const bringHybridToConversation = useCallback(async () => {
    const { data } = await tranApi.post('/api/chat/hybrid-context/bring-to-conversation', {
      session_id: sessionId,
    });
    setHybridAttachment({
      attached: !!data.attached,
      title: typeof data.attachment_title === 'string' ? data.attachment_title : '',
      sources: Array.isArray(data.sources) ? data.sources : [],
    });
    requestAnimationFrame(() => inputRef.current?.focus());
  }, [sessionId]);

  const detachHybridFromConversation = useCallback(async () => {
    try {
      await tranApi.post('/api/chat/hybrid-context/detach', { session_id: sessionId });
    } catch (e) {
      console.warn('HybridContext detach', e);
    }
    setHybridAttachment((prev) => ({ ...prev, attached: false }));
  }, [sessionId]);

  useEffect(() => {
    if (!singleSession) loadSessions();
  }, [singleSession]);

  const headerAppLinks = useMemo(
    () => [
      {
        id: 'morphdata',
        label: 'Morph Data',
        href: APP_URLS.morphData,
        color: '#3b82f6',
        icon: HEADER_APP_ICONS.morphdata,
      },
      {
        id: 'morphutils',
        label: 'Morph Utils',
        href: appHrefWithSession(APP_URLS.morphUtils),
        color: '#2563eb',
        icon: HEADER_APP_ICONS.morphutils,
      },
      {
        id: 'bk',
        label: 'AI tools',
        href: appHrefWithSession(APP_URLS.bk),
        color: '#059669',
        icon: HEADER_APP_ICONS.bk,
      },
    ],
    []
  );

  useEffect(() => {
    if (!sessionId) return;
    loadSessionMessages(sessionId);
    loadHybridAttachment();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- load on session id change only
  }, [sessionId]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  useEffect(() => {
    if (editingSessionId && sessionTitleInputRef.current) {
      sessionTitleInputRef.current.focus();
      sessionTitleInputRef.current.select();
    }
  }, [editingSessionId]);

  useEffect(() => {
    if (wasLoadingRef.current && !loading) {
      requestAnimationFrame(() => {
        inputRef.current?.focus();
      });
    }
    wasLoadingRef.current = loading;
  }, [loading]);

  const sendMessage = async (messageText, file = null, options = {}) => {
    const trimmed = (messageText || '').trim();
    const effectiveFile = enableFileUpload ? file : null;
    const analyzeFile = options?.analyzeFile === true;
    if (!trimmed && !effectiveFile) return;
    if (loading) return;

    abortControllerRef.current?.abort();
    abortControllerRef.current = new AbortController();
    const signal = abortControllerRef.current.signal;

    const userMessage = trimmed || (effectiveFile ? `Uploaded ${effectiveFile.name}` : '');
    setInput('');
    setAttachedFile(null);
    setMessages((prev) => [
      ...prev,
      {
        type: 'user',
        content: effectiveFile ? `📎 ${effectiveFile.name}${trimmed ? ': ' + trimmed : ''}` : userMessage,
      },
    ]);
    setLoading(true);
    setResponseTimeMs(null);
    setLoadingStatus('Reading your question…');
    const startedAt = performance.now();
    const stopProgress = runAiProgress(
      {
        userText: trimmed || userMessage,
        hasFile: Boolean(effectiveFile),
        analyzeFile,
        hasHybridContext: hybridAttachment.attached,
      },
      setLoadingStatus,
      signal,
    );

    try {
      let response;
      if (effectiveFile) {
        const formData = new FormData();
        formData.append('message', trimmed);
        formData.append('file', effectiveFile);
        formData.append('session_id', sessionId);
        if (analyzeFile) {
          formData.append('action', 'analyze_report');
        }
        response = await tranApi.post('/api/chat', formData, {
          signal,
        });
      } else {
        response = await tranApi.post(
          '/api/chat',
          {
            message: userMessage,
            session_id: sessionId,
          },
          { signal }
        );
      }

      const aiResponse = response.data.response;
      setResponseTimeMs(Math.round(performance.now() - startedAt));
      const sql = response.data.sql;
      const confirmationCard = response.data.confirmation_card || null;
      const proposedForm = response.data.proposed_form || null;
      const researchContent = response.data.research_content || null;
      const images = Array.isArray(response.data.images) ? response.data.images : null;

      setMessages((prev) => [
        ...prev,
        {
          type: 'assistant',
          content: aiResponse,
          sql,
          confirmationCard,
          proposedForm,
          researchContent,
          images,
        },
      ]);
      if (!singleSession) loadSessions();
      historyNavIndexRef.current = -1;
      if (trimmed) {
        const h = [...promptHistoryRef.current];
        if (h[0] !== trimmed) {
          h.unshift(trimmed);
          promptHistoryRef.current = h.slice(0, MAX_PROMPTS);
          savePromptHistory(promptHistoryRef.current);
        }
      }
    } catch (error) {
      const cancelled =
        error?.code === 'ERR_CANCELED' || error?.name === 'CanceledError' || axios.isCancel?.(error);
      if (cancelled) {
        return;
      }
      console.error('Error:', error);
      const timedOut = error.code === 'ECONNABORTED';
      const errText =
        (timedOut && 'Request timed out. Try again or use a shorter message.') ||
        error.response?.data?.error ||
        'Failed to get response. Please try again.';
      setResponseTimeMs(Math.round(performance.now() - startedAt));
      setMessages((prev) => [
        ...prev,
        {
          type: 'error',
          content: errText,
        },
      ]);
    } finally {
      stopProgress();
      setLoadingStatus('');
      setLoading(false);
    }
  };

  const handleInputKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      e.currentTarget.form?.requestSubmit();
      return;
    }
    if (e.key !== 'ArrowUp' && e.key !== 'ArrowDown') return;
    const h = promptHistoryRef.current;
    if (h.length === 0) return;
    e.preventDefault();
    if (e.key === 'ArrowUp') {
      if (historyNavIndexRef.current === -1) {
        draftBeforeNavRef.current = input;
        historyNavIndexRef.current = 0;
        setInput(h[0]);
      } else {
        historyNavIndexRef.current = Math.min(historyNavIndexRef.current + 1, h.length - 1);
        setInput(h[historyNavIndexRef.current]);
      }
    } else {
      if (historyNavIndexRef.current === -1) return;
      if (historyNavIndexRef.current === 0) {
        historyNavIndexRef.current = -1;
        setInput(draftBeforeNavRef.current);
      } else {
        historyNavIndexRef.current -= 1;
        setInput(h[historyNavIndexRef.current]);
      }
    }
  };

  const handleSend = async (e) => {
    e.preventDefault();
    if (loading) return;
    if (enableFileUpload && attachedFile && !input.trim()) {
      await sendMessage('', attachedFile);
      return;
    }
    if (!input.trim() && !(enableFileUpload && attachedFile)) return;
    await sendMessage(input, enableFileUpload ? attachedFile || null : null);
  };

  const handleAnalyzeAttachment = async () => {
    if (!enableFileUpload || !attachedFile || loading) return;
    await sendMessage(
      'Generate a markdown analysis report for this file.',
      attachedFile,
      { analyzeFile: true }
    );
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
  };

  const handleNewChat = async () => {
    try {
      const res = await tranApi.post('/api/chat/sessions', { title: 'New chat' });
      const id = res.data?.id;
      if (id) {
        setCurrentSessionId(id);
        setMessages([]);
        loadSessions();
      }
    } catch (e) {
      console.error('Create session error:', e);
    }
  };

  const handleSelectSession = (selectedId) => {
    if (selectedId === currentSessionId) return;
    setEditingSessionId(null);
    setCurrentSessionId(selectedId);
  };

  const saveSessionTitle = useCallback(
    async (sessionId) => {
      const t = editTitleValue.trim();
      setEditingSessionId(null);
      if (!t) {
        loadSessions();
        return;
      }
      try {
        await tranApi.put(`/api/chat/sessions/${sessionId}`, { title: t });
        loadSessions();
      } catch (e) {
        console.error('Rename session error:', e);
        loadSessions();
      }
    },
    [editTitleValue]
  );

  const handleDeleteSession = async (idToRemove) => {
    const ok = await confirm({
      title: 'Remove chat session',
      message: 'Remove this chat session?',
      confirmLabel: 'Remove',
      danger: true,
    });
    if (!ok) return;
    try {
      await tranApi.delete(`/api/chat/sessions/${idToRemove}`);
      if (idToRemove === currentSessionId) {
        setEditingSessionId(null);
        setCurrentSessionId('default');
        setMessages([]);
        await loadSessionMessages('default');
      }
      loadSessions();
    } catch (e) {
      console.error('Delete session error:', e);
    }
  };

  const handleClearConversation = async () => {
    const ok = await confirm({
      title: 'Clear conversation',
      message: 'Clear all messages in this chat? This cannot be undone.',
      confirmLabel: 'Clear',
      danger: true,
    });
    if (!ok) return;
    try {
      await tranApi.post(`/api/chat/sessions/${sessionId}/clear`, {});
      setMessages([]);
    } catch (e) {
      console.error('Clear conversation error:', e);
    }
  };

  const handleCancelRequest = () => {
    abortControllerRef.current?.abort();
    setLoading(false);
  };

  const handleExportSession = () => {
    const session = singleSession ? null : (sessions || []).find((s) => s.id === currentSessionId);
    const title = singleSession ? 'Assistant' : session?.title || 'Chat';
    const safeTitle = String(title).replace(/[^\w-]+/g, '_').slice(0, 48) || 'chat';
    const exportPayload = {
      exportedAt: new Date().toISOString(),
      app: 'Morph AI',
      sessionId,
      title,
      messages: messages.map((m) => ({ ...m })),
    };
    const blob = new Blob([JSON.stringify(exportPayload, null, 2)], { type: 'application/json;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `morph-ai-${safeTitle}-${sessionId}.json`;
    a.rel = 'noopener';
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  };

  const toggleTheme = () => setTheme((t) => (t === 'dark' ? 'light' : 'dark'));

  const chatInner = (
    <div className={`app${embedded ? ' app--embedded' : ''}${singleSession ? ' app--single-session' : ''}${sidebarNavOpen ? ' app--sidebar-open' : ''}`}>
      {!singleSession && sidebarNavOpen && (
        <button
          type="button"
          className="chat-sidebar-backdrop"
          aria-label="Close sessions menu"
          onClick={() => setSidebarNavOpen(false)}
        />
      )}
      {!singleSession && (
        <aside className={`chat-sidebar${sidebarNavOpen ? ' is-open' : ''}`} aria-hidden={!sidebarNavOpen ? undefined : undefined}>
          <div className="sidebar-split-top">
            <button
              type="button"
              className="sidebar-new-chat"
              onClick={() => {
                handleNewChat();
                setSidebarNavOpen(false);
              }}
            >
              + New chat
            </button>
            <div className="sidebar-sessions">
              {(sessions || []).map((s) => (
                <div key={s.id} className={`sidebar-session-row ${s.id === currentSessionId ? 'active' : ''}`}>
                  <button
                    type="button"
                    className="sidebar-session-main"
                    onClick={() => {
                      if (editingSessionId === s.id) return;
                      handleSelectSession(s.id);
                      setSidebarNavOpen(false);
                    }}
                  >
                    {editingSessionId === s.id ? (
                      <input
                        ref={sessionTitleInputRef}
                        className="sidebar-session-input"
                        value={editTitleValue}
                        onChange={(e) => setEditTitleValue(e.target.value)}
                        onBlur={() => saveSessionTitle(s.id)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            e.preventDefault();
                            saveSessionTitle(s.id);
                          }
                          if (e.key === 'Escape') {
                            setEditingSessionId(null);
                            loadSessions();
                          }
                        }}
                        onClick={(e) => e.stopPropagation()}
                      />
                    ) : (
                      <span
                        className="sidebar-session-title"
                        title="Double-click to rename"
                        onDoubleClick={(e) => {
                          e.stopPropagation();
                          setEditingSessionId(s.id);
                          setEditTitleValue(s.title || 'Chat');
                        }}
                      >
                        {s.title || 'Chat'}
                      </span>
                    )}
                  </button>
                  {s.id !== 'default' && (
                    <button
                      type="button"
                      className="sidebar-session-remove"
                      title="Remove session"
                      aria-label="Remove session"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteSession(s.id);
                      }}
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        aria-hidden
                      >
                        <path d="M3 6h18" />
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                        <line x1="10" y1="11" x2="10" y2="17" />
                        <line x1="14" y1="11" x2="14" y2="17" />
                      </svg>
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>
        </aside>
      )}
      <div className="chat-container">
        <div className="chat-header">
          {!singleSession && (
            <button
              type="button"
              className="chat-nav-toggle"
              aria-label={sidebarNavOpen ? 'Close sessions menu' : 'Open sessions menu'}
              aria-expanded={sidebarNavOpen}
              onClick={() => setSidebarNavOpen((v) => !v)}
            >
              <span aria-hidden>☰</span>
            </button>
          )}
          <div className="chat-header-title-stack">
            <h1 className="app-title">MORPH AI</h1>
            {embedded && singleSession ? (
              <div className="header-embedded-controls">
                <span className="header-embedded-brand">MORPH AI</span>
              </div>
            ) : null}
          </div>
          <div className="header-actions">
            <div className="header-app-links" aria-label="App links">
              <button
                type="button"
                className="header-app-link header-app-link--button"
                style={{ '--header-app-color': '#38bdf8' }}
                title="Skills"
                aria-haspopup="dialog"
                aria-expanded={skillsOpen}
                onClick={() => setSkillsOpen(true)}
              >
                <span className="header-app-link-label">Skills</span>
              </button>
              {headerAppLinks.map((app) => (
                <a
                  key={app.id}
                  href={app.href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="header-app-link"
                  style={{ '--header-app-color': app.color }}
                  title={app.label}
                >
                  <img src={app.icon} alt="" className="header-app-link-icon" aria-hidden />
                  <span className="header-app-link-label">{app.label}</span>
                </a>
              ))}
            </div>
            <button
              type="button"
              className="chat-icon-button"
              onClick={() => setNotesDrawerOpen(true)}
              title="Notes & TODOs — synced with MorphData (AI assist)"
              aria-label="Open notes and todos"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden
              >
                <path d="M8 6h13" />
                <path d="M8 12h13" />
                <path d="M8 18h13" />
                <path d="M3 6h.01" />
                <path d="M3 12h.01" />
                <path d="M3 18h.01" />
              </svg>
            </button>
            <button
              type="button"
              className="chat-icon-button"
              onClick={() => setHybridDrawerOpen(true)}
              title="Context & Knowledge — session HybridContext + durable Knowledge Library"
              aria-label="Open Context and Knowledge"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden
              >
                <path d="M12 3h7a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-7" />
                <path d="M3 10h10" />
                <path d="M3 6h6" />
                <path d="M3 14h8" />
                <path d="M3 18h6" />
              </svg>
            </button>
            <button
              type="button"
              className="chat-icon-button"
              onClick={handleClearConversation}
              title="Clear messages in this chat"
              aria-label="Clear chat"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden
              >
                <path d="M3 6h18" />
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                <line x1="10" y1="11" x2="10" y2="17" />
                <line x1="14" y1="11" x2="14" y2="17" />
              </svg>
            </button>
            <button
              type="button"
              className="chat-icon-button"
              onClick={handleExportSession}
              title="Export session (JSON)"
              aria-label="Export session"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="7 10 12 15 17 10" />
                <line x1="12" y1="15" x2="12" y2="3" />
              </svg>
            </button>
            <button
              type="button"
              className="chat-icon-button"
              onClick={toggleTheme}
              title={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
              aria-label={theme === 'dark' ? 'Light theme' : 'Dark theme'}
            >
              {theme === 'dark' ? (
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  aria-hidden
                >
                  <circle cx="12" cy="12" r="4" />
                  <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
                </svg>
              ) : (
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  aria-hidden
                >
                  <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
                </svg>
              )}
            </button>
          </div>
        </div>

        <div className="messages-container">
          {messages.length === 0 && (
            <div className="welcome-message">
              <div className="morphai-welcome-brand">
                <div className="morphai-neon-mark" aria-hidden>
                  <svg viewBox="0 0 120 120" className="morphai-neon-svg" role="img">
                    <defs>
                      <linearGradient id="morphaiNeonGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                        <stop offset="0%" stopColor="#22d3ee" />
                        <stop offset="45%" stopColor="#818cf8" />
                        <stop offset="100%" stopColor="#e879f9" />
                      </linearGradient>
                    </defs>
                    <polygon
                      className="morphai-neon-ring morphai-neon-ring--outer"
                      points="60,8 104,34 104,86 60,112 16,86 16,34"
                      fill="none"
                      stroke="url(#morphaiNeonGrad)"
                      strokeWidth="1.25"
                    />
                    <polygon
                      className="morphai-neon-ring morphai-neon-ring--inner"
                      points="60,22 92,42 92,78 60,98 28,78 28,42"
                      fill="none"
                      stroke="url(#morphaiNeonGrad)"
                      strokeWidth="1.5"
                      opacity="0.85"
                    />
                    <path
                      className="morphai-neon-core"
                      d="M38 82V38l22 26 22-26v44"
                      fill="none"
                      stroke="url(#morphaiNeonGrad)"
                      strokeWidth="4.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                    <circle className="morphai-neon-node" cx="60" cy="60" r="5" fill="#22d3ee" />
                  </svg>
                </div>
                <h2 className="morphai-welcome-title">MORPHAI</h2>
              </div>
            </div>
          )}

          {messages.map((msg, idx) => (
            <div key={idx} className={`message ${msg.type}`}>
              <div className="message-content">
                {msg.type === 'user' && (
                  <div
                    className={`message-bubble user-bubble${isHybridConversationUserMessage(msg.content) ? ' user-bubble--hybrid' : ''}`}
                  >
                    <ChatMarkdown text={msg.content} />
                  </div>
                )}
                {msg.type === 'assistant' && (
                  <div className="message-bubble assistant-bubble">
                    <ChatMarkdown text={msg.content} />
                    {Array.isArray(msg.images) && msg.images.length > 0 && (
                      <div className="chat-generated-images">
                        {msg.images.map((img, iidx) => {
                          const ctype = img.content_type || img.contentType || 'image/png';
                          const b64 = img.base64 || '';
                          if (!b64) return null;
                          return (
                            <img
                              key={iidx}
                              className="chat-generated-image"
                              src={`data:${ctype};base64,${b64}`}
                              alt={img.alt || 'Generated image'}
                            />
                          );
                        })}
                      </div>
                    )}
                    {msg.proposedForm && msg.proposedForm.form_template && (
                      <div className="proposed-form-card">
                        <h3 className="proposed-form-title">{msg.proposedForm.form_template.name}</h3>
                        {msg.proposedForm.form_template.description && (
                          <p className="proposed-form-desc">{msg.proposedForm.form_template.description}</p>
                        )}
                        <span
                          className={`proposed-form-badge badge-${msg.proposedForm.form_template.user_type || 'general'}`}
                        >
                          {msg.proposedForm.form_template.user_type || 'general'}
                        </span>
                        <div className="proposed-form-fields">
                          {(msg.proposedForm.form_template.fields || []).map((field, fidx) => (
                            <div key={fidx} className="proposed-form-row">
                              <span className="proposed-form-label">{field.label || field.name}</span>
                              <span className="proposed-form-type">
                                ({field.type}
                                {field.required ? ', required' : ''})
                              </span>
                            </div>
                          ))}
                        </div>
                        <div className="proposed-form-actions">
                          <button type="button" className="proposed-form-btn save-btn" onClick={() => sendMessage('yes')}>
                            Save sheet
                          </button>
                          <button
                            type="button"
                            className="proposed-form-btn edit-btn"
                            onClick={() => {
                              sendMessage("I'd like to change something");
                              setTimeout(() => inputRef.current?.focus(), 100);
                            }}
                          >
                            Edit
                          </button>
                        </div>
                      </div>
                    )}
                    {msg.researchContent && (
                      <div className="research-content-block">
                        <pre className="research-content-text">{msg.researchContent}</pre>
                      </div>
                    )}
                    {msg.confirmationCard && (
                      <div className="registration-confirmation-card">
                        <h3 className="confirmation-card-title">{msg.confirmationCard.form_name}</h3>
                        <span
                          className={`confirmation-card-badge badge-${msg.confirmationCard.user_type || 'student'}`}
                        >
                          {msg.confirmationCard.user_type || 'student'}
                        </span>
                        <div className="confirmation-card-fields">
                          {(() => {
                            const answers = msg.confirmationCard.answers || {};
                            const fields =
                              msg.confirmationCard.fields && msg.confirmationCard.fields.length > 0
                                ? msg.confirmationCard.fields
                                : Object.keys(answers).map((name) => ({ name, label: name }));
                            const normalizeKey = (k) => (k || '').toString().toLowerCase().replace(/\s+/g, '_').trim();
                            const getValue = (field) => {
                              const v = answers[field.name] ?? answers[field.label];
                              if (v !== undefined && v !== null && v !== '') return v;
                              const n = normalizeKey(field.name);
                              const labelNorm = normalizeKey(field.label);
                              for (const [key, val] of Object.entries(answers)) {
                                if (val === undefined || val === null || val === '') continue;
                                const keyNorm = normalizeKey(key);
                                if (keyNorm === n || keyNorm === labelNorm) return val;
                              }
                              return undefined;
                            };
                            const rows = fields
                              .map((field, fidx) => {
                                const value = getValue(field);
                                if (value === undefined || value === null || value === '') return null;
                                const label = field.label || field.name;
                                return (
                                  <div key={fidx} className="confirmation-card-row">
                                    <span className="confirmation-card-label">{label}:</span>
                                    <span className="confirmation-card-value">{String(value)}</span>
                                  </div>
                                );
                              })
                              .filter(Boolean);
                            if (rows.length > 0) return rows;
                            return Object.entries(answers)
                              .map(([key, val]) => {
                                if (val === undefined || val === null || val === '') return null;
                                if (typeof val === 'object') return null;
                                const label = typeof key === 'string' && key.length > 0 ? key.replace(/_/g, ' ') : key;
                                return (
                                  <div key={key} className="confirmation-card-row">
                                    <span className="confirmation-card-label">{label}:</span>
                                    <span className="confirmation-card-value">{String(val)}</span>
                                  </div>
                                );
                              })
                              .filter(Boolean);
                          })()}
                        </div>
                        <div className="confirmation-card-actions">
                          <button
                            type="button"
                            className="confirmation-card-btn confirm-btn"
                            onClick={() => sendMessage('confirm')}
                          >
                            Confirm & submit
                          </button>
                          <button
                            type="button"
                            className="confirmation-card-btn edit-btn"
                            onClick={() => {
                              sendMessage("I'd like to change something");
                              setTimeout(() => inputRef.current?.focus(), 100);
                            }}
                          >
                            Edit
                          </button>
                        </div>
                      </div>
                    )}
                    {msg.sql && (
                      <div className="sql-block">
                        <div className="sql-header">
                          <span>SQL Query</span>
                          <button className="copy-button" onClick={() => copyToClipboard(msg.sql)} type="button" title="Copy SQL">
                            📋 Copy
                          </button>
                        </div>
                        <pre>
                          <code>{msg.sql}</code>
                        </pre>
                      </div>
                    )}
                  </div>
                )}
                {msg.type === 'error' && <div className="message-bubble error-bubble">⚠️ {msg.content}</div>}
              </div>
            </div>
          ))}

          {loading && (
            <div className="message assistant">
              <div className="message-bubble assistant-bubble ai-progress-status">
                <span className="ai-progress-dot" aria-hidden="true" />
                {loadingStatus || 'Working…'}
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        <HybridContextDrawer
          open={hybridDrawerOpen}
          onClose={() => setHybridDrawerOpen(false)}
          sessionId={sessionId}
          onBringToConversation={bringHybridToConversation}
          onAttachmentChange={loadHybridAttachment}
        />

        <ChatNotesTodosDrawer open={notesDrawerOpen} onClose={() => setNotesDrawerOpen(false)} />
        <SkillsModal open={skillsOpen} onClose={() => setSkillsOpen(false)} />

        <form className="input-container" onSubmit={handleSend} onClick={(e) => e.stopPropagation()}>
          {hybridAttachment.attached && hybridAttachment.title ? (
            <div className="hybrid-context-attach-bar" role="status" aria-label="Attached HybridContext reference">
              <span className="hybrid-context-attach-icon" aria-hidden>
                📎
              </span>
              <div className="hybrid-context-attach-copy">
                <span className="hybrid-context-attach-label">HybridContext reference</span>
                <span className="hybrid-context-attach-title" title={hybridAttachment.title}>
                  {hybridAttachment.title}
                </span>
              </div>
              <button
                type="button"
                className="hybrid-context-attach-remove"
                onClick={detachHybridFromConversation}
                aria-label="Remove HybridContext reference from chat"
              >
                Remove
              </button>
            </div>
          ) : null}
          {enableFileUpload && attachedFile && (
            <div className="attached-file-tag">
              <span>📎 {attachedFile.name}</span>
              <button type="button" className="attached-file-remove" onClick={() => setAttachedFile(null)} aria-label="Remove file">
                ×
              </button>
            </div>
          )}
          <div className="chat-input-gradient-frame">
            <div className="input-wrapper chat-input-inner">
            {enableFileUpload && (
              <>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*,.pdf,application/pdf"
                  onChange={(e) => {
                    const f = e.target.files?.[0];
                    if (f) setAttachedFile(f);
                    e.target.value = '';
                  }}
                  className="file-input-hidden"
                  aria-label="Upload image or PDF"
                />
                <button
                  type="button"
                  className="chat-icon-button chat-icon-button--attach"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={loading}
                  title="Upload image or PDF"
                  aria-label="Upload image or PDF"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    aria-hidden
                  >
                    <path d="m21.44 11.05-9.19 9.19a6 6 0 0 1-8.49-8.49l8.57-8.57A4 4 0 1 1 18 8.84l-8.59 8.57a2 2 0 0 1-2.83-2.83l8.49-8.48" />
                  </svg>
                </button>
                {attachedFile && (
                  <button
                    type="button"
                    className="chat-icon-button chat-icon-button--attach"
                    onClick={handleAnalyzeAttachment}
                    disabled={loading}
                    title="Analyze uploaded file"
                    aria-label="Analyze uploaded file"
                  >
                    Analyze
                  </button>
                )}
              </>
            )}
            <textarea
              ref={inputRef}
              value={input}
              onChange={(e) => {
                setInput(e.target.value);
                historyNavIndexRef.current = -1;
              }}
              onKeyDown={handleInputKeyDown}
              placeholder={
                enableFileUpload ? 'Type a message or upload image/PDF…' : 'Type a message…'
              }
              className="message-input"
              disabled={loading}
              autoComplete="off"
              rows={1}
            />
            {input.trim() && (
              <button
                type="button"
                className="clear-input-button"
                onClick={() => {
                  setInput('');
                  historyNavIndexRef.current = -1;
                  inputRef.current?.focus();
                }}
                disabled={loading}
                title="Clear text"
                aria-label="Clear message text"
              >
                ✕
              </button>
            )}
            <button
              type="button"
              className="chat-icon-button chat-icon-button--input"
              onClick={handleCancelRequest}
              disabled={!loading}
              title={loading ? 'Cancel request' : 'No request in progress'}
              aria-label="Cancel request"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden
              >
                <circle cx="12" cy="12" r="10" />
                <rect x="9" y="9" width="6" height="6" rx="1" fill="currentColor" stroke="none" />
              </svg>
            </button>
          </div>
          </div>
          <button
            type="submit"
            className="send-button"
            disabled={loading || (!input.trim() && !(enableFileUpload && attachedFile))}
          >
            {loading ? '⏳' : '➤'}
          </button>
        </form>
        {responseTimeMs != null && (
          <div className="chat-response-time" aria-live="polite">
            Response {formatResponseTime(responseTimeMs)}
          </div>
        )}
      </div>
    </div>
  );

  if (embedded) {
    return (
      <div
        className={`skool-ai-chat--embedded${singleSession ? ' skool-ai-chat--single-session' : ''}`}
        style={{ height: '100%', minHeight: 0, display: 'flex', flexDirection: 'column' }}
      >
        {chatInner}
      </div>
    );
  }

  return <div className="app-outer">{chatInner}</div>;
}
