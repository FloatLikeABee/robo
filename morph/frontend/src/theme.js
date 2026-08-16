import { createTheme } from '@mui/material/styles';

const STORAGE_KEY = 'morphdata-theme-mode';

export function getStoredThemeMode() {
  try {
    const v = localStorage.getItem(STORAGE_KEY);
    if (v === 'light' || v === 'dark') return v;
    const legacy = localStorage.getItem('skoolz-theme-mode');
    if (legacy === 'light' || legacy === 'dark') {
      localStorage.setItem(STORAGE_KEY, legacy);
      return legacy;
    }
  } catch {}
  return 'light';
}

/** DataGrid / table header background. Dark = Morph AI surface2; light = very soft neutral (not saturated blue). */
export function getAdminTableHeaderBg(mode) {
  return mode === 'dark' ? '#0a1220' : '#f1eff7';
}

/** DataGrid / table header text & icon color (pairs with getAdminTableHeaderBg). */
export function getAdminTableHeaderColor(mode) {
  return mode === 'dark' ? '#e8eff7' : 'rgba(26, 20, 35, 0.88)';
}

export function setStoredThemeMode(mode) {
  try {
    localStorage.setItem(STORAGE_KEY, mode);
  } catch {}
}

/**
 * MorphData admin UI: dark mode matches Morph AI chat tokens (App.css data-chat-theme='dark');
 * light mode = soft neutral; primary accent = deep blue.
 */
export function getAdminTheme(mode) {
  const isDark = mode === 'dark';

  return createTheme({
    palette: {
      mode,
      primary: {
        main: isDark ? '#3b82f6' : '#5b5bd4',
        light: isDark ? '#60a5fa' : '#7c7fe0',
        dark: isDark ? '#1d4ed8' : '#4a47b8',
        contrastText: isDark ? '#060a12' : '#fafafc',
      },
      secondary: {
        main: isDark ? '#2563eb' : '#7e6ec9',
        light: isDark ? '#3b82f6' : '#9d8fd9',
        dark: isDark ? '#1e40af' : '#5f51a3',
        contrastText: isDark ? '#f0f7ff' : '#faf8ff',
      },
      text: {
        primary: isDark ? '#e8eff7' : '#1a1423',
        secondary: isDark ? 'rgba(232, 239, 247, 0.74)' : 'rgba(26, 20, 35, 0.65)',
        disabled: isDark ? 'rgba(136, 153, 176, 0.56)' : 'rgba(26, 20, 35, 0.38)',
      },
      divider: isDark ? 'rgba(37, 99, 235, 0.3)' : 'rgba(0, 0, 0, 0.10)',
      action: {
        active: isDark ? 'rgba(37, 99, 235, 0.8)' : 'rgba(79, 70, 229, 0.55)',
        hover: isDark ? 'rgba(37, 99, 235, 0.24)' : 'rgba(99, 102, 241, 0.06)',
        selected: isDark ? 'rgba(37, 99, 235, 0.28)' : 'rgba(99, 102, 241, 0.08)',
        disabled: isDark ? 'rgba(136, 153, 176, 0.45)' : 'rgba(0, 0, 0, 0.26)',
        disabledBackground: isDark ? 'rgba(37, 99, 235, 0.12)' : 'rgba(0, 0, 0, 0.12)',
      },
      background: {
        default: isDark ? '#020408' : '#f9f8fb',
        paper: isDark ? '#060a12' : '#ffffff',
      },
      success: { main: '#2E7D32' },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            touchAction: 'manipulation',
            WebkitTapHighlightColor: 'transparent',
          },
        },
      },
      MuiButton: {
        styleOverrides: {
          root: {
            textTransform: 'none',
            minHeight: 40,
            ...(isDark && {
              '&.MuiButton-textPrimary': {
                color: '#60a5fa',
              },
              '&.MuiButton-textSecondary': {
                color: '#93c5fd',
              },
              '&.MuiButton-outlinedPrimary': {
                borderColor: 'rgba(37, 99, 235, 0.55)',
                color: '#60a5fa',
              },
            }),
          },
          containedPrimary: {
            background: isDark
              ? 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)'
              : 'linear-gradient(135deg, #6d6ad8 0%, #5b58c4 100%)',
            color: isDark ? '#0f172a' : '#fafafc',
            '&:hover': {
              background: isDark
                ? 'linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%)'
                : 'linear-gradient(135deg, #5b58c4 0%, #4f4cb0 100%)',
            },
          },
          containedSecondary: {
            background: isDark
              ? 'linear-gradient(135deg, #1d4ed8 0%, #1e3a8a 100%)'
              : '#8b7fd4',
            color: isDark ? '#f0f7ff' : '#fff',
            '&:hover': {
              background: isDark
                ? 'linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%)'
                : '#7c72c8',
            },
          },
        },
      },
      MuiIconButton: {
        styleOverrides: {
          root: {
            ...(isDark
              ? {
                  '&:hover': { backgroundColor: 'rgba(37, 99, 235, 0.12)' },
                }
              : {}),
          },
          colorInherit: isDark
            ? {
                color: 'rgba(236, 232, 247, 0.88)',
              }
            : {},
        },
      },
      MuiTextField: {
        defaultProps: {
          size: 'small',
        },
        styleOverrides: {
          root: {
            '& .MuiInputBase-input': {
              // Prefer 16px on touch devices to avoid iOS zoom; desktop stays compact via parent size.
              fontSize: '16px',
              [`@media (min-width:${900}px)`]: {
                fontSize: '14px',
              },
            },
          },
        },
      },
      MuiDrawer: {
        styleOverrides: {
          paper: {
            backgroundColor: isDark ? '#040810' : '#f5f3f9',
            borderRight: isDark ? '1px solid rgba(37, 99, 235, 0.3)' : '1px solid rgba(0, 0, 0, 0.08)',
          },
        },
      },
      MuiListItemButton: {
        styleOverrides: {
          root: {
            minHeight: 48,
            '&.Mui-selected': {
              backgroundColor: isDark ? 'rgba(37, 99, 235, 0.24)' : 'rgba(99, 102, 241, 0.08)',
              borderLeft: `3px solid ${isDark ? '#2563eb' : '#6b69c7'}`,
            },
          },
        },
      },
    },
  });
}

export const adminTheme = getAdminTheme('light');
