import React, { useState, useMemo, useEffect, useLayoutEffect } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import {
  ThemeProvider,
  CssBaseline,
  Box,
  IconButton,
  Typography,
} from '@mui/material';
import { getAdminTheme, setStoredThemeMode } from './theme';
import { PlatformUiProvider, usePlatformUi } from './PlatformUiContext';
import AppDrawer, { DRAWER_WIDTH } from './components/admin/AppDrawer';
import MenuIcon from '@mui/icons-material/Menu';
import { refreshMorphAuthSnapshot } from './auth/morphSession';
import { releaseStuckOverlays } from './utils/releaseStuckOverlays';

function AdminLayoutInner() {
  const location = useLocation();
  const { labels } = usePlatformUi();
  const themeMode = 'dark';
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const theme = useMemo(() => getAdminTheme(themeMode), [themeMode]);

  const isDark = true;

  useEffect(() => {
    setStoredThemeMode('dark');
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
    setMobileNavOpen(false);
    releaseStuckOverlays();
  }, [location.pathname]);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline enableColorScheme />
      <Box
          sx={{
            display: 'flex',
            width: '100%',
            maxWidth: '100%',
            minWidth: 0,
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
              display: { xs: 'flex', md: 'none' },
              boxSizing: 'border-box',
              minHeight: 'calc(56px + env(safe-area-inset-top))',
              pl: 'max(8px, env(safe-area-inset-left))',
              pr: 'max(8px, env(safe-area-inset-right))',
              pt: 'env(safe-area-inset-top)',
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
                flex: 1,
                minWidth: 0,
                color: 'primary.main',
                pl: 0.5,
              }}
            >
              {labels.product_name || 'MorphNotes'}
            </Typography>
          </Box>

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
          {/* Phone: reserve home-indicator space without a tall footer strip.
              Display breakpoint so the spacer is in the first paint. */}
          <Box
            sx={{ display: { xs: 'block', sm: 'none' }, height: 'env(safe-area-inset-bottom)', flexShrink: 0 }}
            aria-hidden
          />
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
