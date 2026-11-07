import React, { useState, useEffect } from 'react';
import { ThemeProvider, CssBaseline } from '@mui/material';
import SkoolAiChat from './SkoolAiChat';
import { getAdminTheme } from './theme';
import { chatThemeToAdminMode, resolveChatTheme } from './lib/chatTheme';

function App() {
  const [adminMode, setAdminMode] = useState(() => chatThemeToAdminMode(resolveChatTheme()));
  useEffect(() => {
    const fn = (e) => {
      const d = e?.detail;
      if (d === 'dark' || d === 'light' || d === 'twilight') {
        setAdminMode(chatThemeToAdminMode(d));
      }
    };
    window.addEventListener('dorph-chat-theme', fn);
    return () => window.removeEventListener('dorph-chat-theme', fn);
  }, []);
  const theme = getAdminTheme(adminMode);
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline enableColorScheme />
      <SkoolAiChat variant="page" enableFileUpload />
    </ThemeProvider>
  );
}

export default App;
