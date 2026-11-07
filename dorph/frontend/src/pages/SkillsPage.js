import React, { useEffect } from 'react';
import { Link } from 'react-router-dom';
import { SkillsPanel } from '../components/SkillsModal';
import { resolveChatTheme } from '../lib/chatTheme';
import '../App.css';

/** Bookmark fallback for /skills — same themed panel as the chat modal. */
export default function SkillsPage() {
  useEffect(() => {
    document.documentElement.setAttribute('data-chat-theme', resolveChatTheme());
  }, []);

  return (
    <div className="skills-panel-page">
      <div className="skills-panel-page-inner">
        <SkillsPanel />
        <div style={{ padding: '0 22px 20px' }}>
          <Link to="/" className="header-app-link" style={{ '--header-app-color': 'var(--chat-accent)' }}>
            <span className="header-app-link-label">← Dorph AI</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
