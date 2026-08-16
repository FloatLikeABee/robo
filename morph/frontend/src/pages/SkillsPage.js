import React, { useEffect } from 'react';
import { Link } from 'react-router-dom';
import { SkillsPanel } from '../components/SkillsModal';
import '../App.css';

const THEME_KEY = 'skool-ai-chat-theme';

function getStoredChatTheme() {
  try {
    const v = localStorage.getItem(THEME_KEY);
    if (v === 'light' || v === 'dark') return v;
  } catch {}
  return 'dark';
}

/** Bookmark fallback for /skills — same themed panel as the chat modal. */
export default function SkillsPage() {
  useEffect(() => {
    document.documentElement.setAttribute('data-chat-theme', getStoredChatTheme());
  }, []);

  return (
    <div className="skills-panel-page">
      <div className="skills-panel-page-inner">
        <SkillsPanel />
        <div style={{ padding: '0 22px 20px' }}>
          <Link to="/" className="header-app-link" style={{ '--header-app-color': 'var(--chat-accent)' }}>
            <span className="header-app-link-label">← Morph AI</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
