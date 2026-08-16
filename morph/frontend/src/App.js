import React, { useState, useEffect } from 'react';
import { ThemeProvider, CssBaseline } from '@mui/material';
import SkoolAiChat from './SkoolAiChat';
import { getAdminTheme } from './theme';

const CHAT_THEME_KEY = 'skool-ai-chat-theme';

function readInitialChatTheme() {
  try {
    const v = localStorage.getItem(CHAT_THEME_KEY);
    if (v === 'dark' || v === 'light') return v;
  } catch {}
  return 'dark';
}

function App() {
  const [mode, setMode] = useState(readInitialChatTheme);
  useEffect(() => {
    const fn = (e) => {
      const d = e?.detail;
      if (d === 'dark' || d === 'light') setMode(d);
    };
    window.addEventListener('morph-chat-theme', fn);
    return () => window.removeEventListener('morph-chat-theme', fn);
  }, []);
  const theme = getAdminTheme(mode);
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline enableColorScheme />
      <SkoolAiChat variant="page" enableFileUpload />
    </ThemeProvider>
  );
}

export default App;
