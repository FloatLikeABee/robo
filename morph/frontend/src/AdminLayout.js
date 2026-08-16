import React, { useState, useMemo, useEffect, useLayoutEffect } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import {
  ThemeProvider,
  CssBaseline,
  Box,
  IconButton,
  Typography,
  Tooltip,
  useMediaQuery,
} from '@mui/material';
import { getAdminTheme, getStoredThemeMode, setStoredThemeMode } from './theme';
import { PlatformUiProvider, usePlatformUi } from './PlatformUiContext';
import { useAdminBasePath } from './adminPaths';
import AppDrawer, { DRAWER_WIDTH } from './components/admin/AppDrawer';
import PersonOutlinedIcon from '@mui/icons-material/PersonOutlined';
import DarkModeOutlinedIcon from '@mui/icons-material/DarkModeOutlined';
import LightModeOutlinedIcon from '@mui/icons-material/LightModeOutlined';
import EventNoteOutlinedIcon from '@mui/icons-material/EventNoteOutlined';
import SmartToyOutlinedIcon from '@mui/icons-material/SmartToyOutlined';
import MenuIcon from '@mui/icons-material/Menu';
import NotesTodosPanel from './components/admin/NotesTodosPanel';
import AiPanel from './components/admin/AiPanel';
import { refreshMorphAuthSnapshot } from './auth/morphSession';
import { releaseStuckOverlays } from './utils/releaseStuckOverlays';

const headerIconSx = (active, isDark) => ({
  color: active ? 'primary.main' : 'secondary.main',
  bgcolor: active
    ? isDark
      ? 'rgba(156, 39, 176, 0.22)'
      : 'rgba(156, 39, 176, 0.14)'
    : isDark
      ? 'rgba(156, 39, 176, 0.12)'
      : 'rgba(156, 39, 176, 0.08)',
  border: 1,
  borderColor: active ? 'primary.light' : 'transparent',
  width: 42,
  height: 42,
  '&:hover': {
    bgcolor: isDark ? 'rgba(156, 39, 176, 0.28)' : 'rgba(156, 39, 176, 0.18)',
  },
});

function AdminLayoutInner() {
  const location = useLocation();
  const adminBase = useAdminBasePath();
  const { labels } = usePlatformUi();
  const [themeMode, setThemeMode] = useState(getStoredThemeMode);
  const [notesOpen, setNotesOpen] = useState(false);
  const [aiOpen, setAiOpen] = useState(false);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const theme = useMemo(() => getAdminTheme(themeMode), [themeMode]);
  const isPhone = useMediaQuery(theme.breakpoints.down('sm'));

  const toggleTheme = () => {
    const next = themeMode === 'dark' ? 'light' : 'dark';
    setThemeMode(next);
    setStoredThemeMode(next);
  };

  const isDark = themeMode === 'dark';

  useEffect(() => {
    document.body.setAttribute('data-theme', themeMode);
    document.documentElement.classList.add('morph-data-app');
    document.body.classList.add('morph-data-app');
    return () => {
      document.body.removeAttribute('data-theme');
      document.documentElement.classList.remove('morph-data-app');
      document.body.classList.remove('morph-data-app');
    };
  }, [themeMode]);

  useEffect(() => {
    void refreshMorphAuthSnapshot();
  }, []);

  /** Close flyouts on navigation and drop orphaned MUI backdrop/scroll-lock layers. */
  useLayoutEffect(() => {
    setNotesOpen(false);
    setAiOpen(false);
    setMobileNavOpen(false);
    releaseStuckOverlays();
  }, [location.pathname]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline enableColorScheme />
      <Box
        sx={{
          display: 'flex',
          minHeight: '100dvh',
          height: '100dvh',
          maxHeight: '100dvh',
          overflow: 'hidden',
          bgcolor: 'background.default',
        }}
      >
        <AppDrawer mobileOpen={mobileNavOpen} onMobileClose={() => setMobileNavOpen(false)} />
        <Box
          component="main"
          sx={{
            flexGrow: 1,
            display: 'flex',
            flexDirection: 'column',
            height: '100%',
            minHeight: 0,
            overflow: 'hidden',
            width: { xs: '100%', md: `calc(100% - ${DRAWER_WIDTH}px)` },
            minWidth: 0,
          }}
        >
          <Box
            sx={{
              minHeight: 56,
              px: { xs: 1, sm: 2 },
              pt: 'env(safe-area-inset-top)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              bgcolor: 'background.paper',
              borderBottom: isDark ? '1px solid rgba(255,255,255,0.10)' : '1px solid rgba(0,0,0,0.12)',
              flexShrink: 0,
              gap: 0.5,
            }}
          >
            <IconButton
              aria-label="Open navigation"
              onClick={() => setMobileNavOpen(true)}
              sx={{
                display: { xs: 'inline-flex', md: 'none' },
                color: 'text.primary',
                width: 44,
                height: 44,
              }}
            >
              <MenuIcon />
            </IconButton>
            <Typography
              variant="subtitle1"
              fontWeight={700}
              noWrap
              sx={{
                display: { xs: 'block', md: 'none' },
                flex: 1,
                minWidth: 0,
                color: 'primary.main',
                pl: 0.5,
              }}
            >
              {labels.product_name || 'Morph Data'}
            </Typography>
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: { xs: 0.25, sm: 1 },
                ml: { xs: 0, md: 'auto' },
                overflowX: 'auto',
                scrollbarWidth: 'none',
                '&::-webkit-scrollbar': { display: 'none' },
              }}
            >
              <Tooltip title="Notes & TODOs" enterDelay={400}>
                <IconButton
                  aria-label="Notes and TODOs"
                  onClick={() => {
                    setAiOpen(false);
                    setNotesOpen((prev) => !prev);
                  }}
                  sx={headerIconSx(notesOpen, isDark)}
                >
                  <EventNoteOutlinedIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <IconButton
                color="inherit"
                aria-label={isDark ? 'Switch to light theme' : 'Switch to dark theme'}
                onClick={toggleTheme}
                sx={{ width: 42, height: 42 }}
              >
                {isDark ? <LightModeOutlinedIcon /> : <DarkModeOutlinedIcon />}
              </IconButton>
              <Tooltip title={labels.ai_assistant_name} enterDelay={400}>
                <IconButton
                  aria-label="AI assistant"
                  onClick={() => {
                    setNotesOpen(false);
                    setAiOpen((prev) => !prev);
                  }}
                  sx={headerIconSx(aiOpen, isDark)}
                >
                  <SmartToyOutlinedIcon fontSize="small" />
                </IconButton>
              </Tooltip>
              <Tooltip title="User settings" enterDelay={400}>
                <IconButton
                  color="inherit"
                  aria-label="User settings"
                  onClick={() => {
                    const to = `${adminBase}/configuration/user-settings`;
                    if (location.pathname === to) return;
                    window.location.assign(to);
                  }}
                  sx={{ width: 42, height: 42 }}
                >
                  <PersonOutlinedIcon fontSize="small" />
                </IconButton>
              </Tooltip>
            </Box>
          </Box>
          <NotesTodosPanel open={notesOpen} onClose={() => setNotesOpen(false)} />
          <AiPanel open={aiOpen} onClose={() => setAiOpen(false)} />

          <Box
            key={location.pathname}
            sx={{
              flex: 1,
              minHeight: 0,
              px: { xs: 1, sm: 2 },
              py: { xs: 1, sm: 2 },
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
            }}
          >
            <Outlet />
          </Box>

          <Box
            sx={{
              display: { xs: 'none', sm: 'flex' },
              minHeight: 40,
              px: 2,
              pb: 'env(safe-area-inset-bottom)',
              alignItems: 'center',
              justifyContent: 'center',
              bgcolor: 'background.paper',
              borderTop: isDark ? '1px solid rgba(255,255,255,0.10)' : '1px solid rgba(0,0,0,0.12)',
              flexShrink: 0,
            }}
          >
            <Typography variant="caption" color="text.secondary">
              © {new Date().getFullYear()} {labels.product_name}
            </Typography>
          </Box>
          {/* Phone: reserve home-indicator space without a tall footer strip */}
          {isPhone ? <Box sx={{ height: 'env(safe-area-inset-bottom)', flexShrink: 0 }} aria-hidden /> : null}
        </Box>
      </Box>
    </ThemeProvider>
  );
}

export default function AdminLayout() {
  return (
    <PlatformUiProvider>
      <AdminLayoutInner />
    </PlatformUiProvider>
  );
}
