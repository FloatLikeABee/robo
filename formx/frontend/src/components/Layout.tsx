import { useCallback, useEffect, useState } from 'react';
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom';
import { PlatformChatDrawer } from '@robo/platform-chat/react';
import '@robo/platform-chat/chat-drawer.css';
import { api, getAuthToken } from '../lib/api';

const API_BASE = import.meta.env.VITE_API_URL ?? '';

const SHEETX_AI_SUGGESTIONS = [
  'What forms do I have?',
  'survey bot staff onboarding',
  'Design a game ideas survey form and create it',
  'Help me build a customer feedback form with multiple pages',
  'Search the web and draft an HTML form template for event registration',
  'List my events',
];

type Theme = 'light' | 'dark';

function readTheme(): Theme {
  const saved = window.localStorage.getItem('sheetx-theme') ?? window.localStorage.getItem('formsx-theme');
  return saved === 'light' ? 'light' : 'dark';
}

const tabClass = (isDark: boolean) => ({ isActive }: { isActive: boolean }) =>
  `shrink-0 inline-flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
    isActive
      ? isDark
        ? 'bg-sky-700/35 text-sky-100 border border-sky-500/40'
        : 'bg-[#d8eef8] text-[#0f4c66] border border-[#b7deee]'
      : isDark
        ? 'text-slate-200 hover:bg-slate-700/70 hover:text-white border border-transparent'
        : 'text-[#31566b] hover:bg-[#e8f5fb] hover:text-[#123f56] border border-transparent'
  }`;

export function Layout() {
  const navigate = useNavigate();
  const [theme, setTheme] = useState<Theme>(() => readTheme());

  useEffect(() => {
    window.localStorage.setItem('sheetx-theme', theme);
  }, [theme]);

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }, [theme]);

  useEffect(() => {
    setAssistantOpen(false);
    document.body.style.removeProperty('overflow');
    document.body.style.removeProperty('padding-right');
  }, []);

  const isDark = theme === 'dark';
  const tc = tabClass(isDark);
  const [assistantOpen, setAssistantOpen] = useState(false);

  const toggleTheme = () => {
    setTheme((prev) => (prev === 'dark' ? 'light' : 'dark'));
  };

  const assistantChatEndpoint = `${API_BASE}/api/v1/assistant/chat`.replace(/([^:]\/)\/+/g, '$1');

  const getAssistantHeaders = useCallback(() => {
    const token = getAuthToken();
    const headers: Record<string, string> = {};
    if (token) headers.Authorization = `Bearer ${token}`;
    return headers;
  }, []);

  return (
    <div
      className={
        isDark
          ? 'min-h-dvh md:h-screen overflow-hidden bg-linear-to-b from-[#14213d] via-[#0f172a] to-[#0b1120] text-slate-100 flex flex-col'
          : 'min-h-dvh md:h-screen overflow-hidden bg-[#f5f8ff] text-slate-900 flex flex-col'
      }
    >
      <header
        className={
          'shrink-0 z-20 border-b ' +
          (isDark ? 'border-slate-700/80 bg-slate-800/95' : 'border-[#c6e3ef] bg-[#eaf6fb]/95')
        }
      >
        <div className="px-3 md:px-4 py-3 flex items-center justify-between gap-2">
          <Link to="/" className="flex items-center gap-2 min-w-0">
            <span className={isDark ? 'text-2xl text-violet-300' : 'text-2xl text-violet-600'}>📄</span>
            <div className="min-w-0">
              <span className={isDark ? 'font-semibold text-white text-lg' : 'font-semibold text-slate-900 text-lg'}>
                Survey Maker
              </span>
            </div>
          </Link>
          <div className="flex items-center gap-1.5 md:gap-3 shrink-0">
            <button
              type="button"
              onClick={() => setAssistantOpen((o) => !o)}
              className={
                'px-2.5 py-2 rounded-lg text-sm font-medium transition-colors sm:px-3 ' +
                (isDark ? 'bg-sky-700/45 text-sky-100 hover:bg-sky-600/50' : 'bg-sky-100 text-sky-800 hover:bg-sky-200')
              }
              aria-expanded={assistantOpen}
              aria-controls="sheetx-ai-assistant-drawer"
            >
              <span className="hidden sm:inline">AI Assistant</span>
              <span className="sm:hidden">AI</span>
            </button>
            <button
              type="button"
              onClick={toggleTheme}
              className={
                'p-2 rounded-lg text-lg transition-colors ' +
                (isDark ? 'text-slate-300 hover:text-white' : 'text-slate-500 hover:text-slate-800')
              }
              aria-label="Toggle theme"
            >
              {isDark ? '☀' : '🌙'}
            </button>
            <button
              type="button"
              onClick={() => {
                api.auth.logout();
                navigate('/login', { replace: true });
              }}
              className={
                'flex items-center gap-2 px-3 py-2 rounded-lg text-sm ' +
                (isDark
                  ? 'bg-[#ea580c] text-slate-50 hover:bg-[#c2410c]'
                  : 'bg-slate-200 text-slate-700 hover:bg-slate-300')
              }
            >
              Logout <span>→</span>
            </button>
          </div>
        </div>

        <nav
          className="flex items-center gap-1 px-3 md:px-4 pb-3 overflow-x-auto"
          aria-label="Sections"
        >
          <NavLink to="/survey-bot" className={tc}>
            <span aria-hidden className="text-base leading-none">🤖</span>
            <span>AI Surveys</span>
          </NavLink>
          <NavLink to="/events-info" className={tc}>
            <span aria-hidden className="text-base leading-none">📌</span>
            <span>Events &amp; Info</span>
          </NavLink>
        </nav>
      </header>

      <main
        className={
          'flex-1 min-w-0 min-h-0 flex flex-col overflow-hidden px-2 sm:px-3 md:px-4 py-2 md:py-4 ' +
          (isDark ? '' : 'bg-[#f6fbfe]')
        }
      >
        <div className="flex-1 min-h-0 flex flex-col overflow-hidden">
          <Outlet />
        </div>
      </main>

      <PlatformChatDrawer
        open={assistantOpen}
        onClose={() => setAssistantOpen(false)}
        title="Survey Maker AI"
        chatEndpoint={assistantChatEndpoint}
        getHeaders={getAssistantHeaders}
        welcomeMessage="Hi! I'm **Survey Maker AI**. Ask in plain language — I can help design AI surveys and manage events."
        suggestions={SHEETX_AI_SUGGESTIONS}
        progressContext={{ app: 'sheetx' }}
      />
    </div>
  );
}
