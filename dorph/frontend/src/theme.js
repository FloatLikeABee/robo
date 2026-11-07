import { createTheme } from '@mui/material/styles';

const STORAGE_KEY = 'dorphdata-theme-mode';

export function getStoredThemeMode() {
  try {
    localStorage.setItem(STORAGE_KEY, 'dark');
    localStorage.removeItem('morphdata-theme-mode');
  } catch {}
  return 'dark';
}

/** DataGrid / table header background. Dark = Dorph AI twilight surface; light = soft neutral. */
export function getAdminTableHeaderBg(mode) {
  return mode === 'dark' ? '#3a3252' : '#e4dfd6';
}

/** DataGrid / table header text & icon color (pairs with getAdminTableHeaderBg). */
export function getAdminTableHeaderColor(mode) {
  return mode === 'dark' ? '#f4f0fa' : 'rgba(26, 20, 35, 0.88)';
}

export function setStoredThemeMode(mode) {
  try {
    localStorage.setItem(STORAGE_KEY, mode);
  } catch {}
}

/**
 * DorphData admin UI: MUI `mode: 'dark'` tracks Dorph AI twilight glass chat
 * (`data-chat-theme='twilight'` — purple primary, orange secondary). Light mode = soft neutral.
 */
export function getAdminTheme(mode) {
  const isDark = mode === 'dark';

  return createTheme({
    palette: {
      mode,
      primary: {
        main: isDark ? '#8b5cf6' : '#5b5bd4',
        light: isDark ? '#a78bfa' : '#7c7fe0',
        dark: isDark ? '#7c3aed' : '#4a47b8',
        contrastText: isDark ? '#1a1028' : '#fafafc',
      },
      secondary: {
        main: isDark ? '#f97316' : '#7e6ec9',
        light: isDark ? '#fb923c' : '#9d8fd9',
        dark: isDark ? '#ea580c' : '#5f51a3',
        contrastText: isDark ? '#1a1028' : '#faf8ff',
      },
      text: {
        primary: isDark ? '#f4f0fa' : '#1a1423',
        secondary: isDark ? 'rgba(244, 240, 250, 0.74)' : 'rgba(26, 20, 35, 0.65)',
        disabled: isDark ? 'rgba(136, 153, 176, 0.56)' : 'rgba(26, 20, 35, 0.38)',
      },
      divider: isDark ? 'rgba(167, 139, 250, 0.28)' : 'rgba(0, 0, 0, 0.10)',
      action: {
        active: isDark ? 'rgba(167, 139, 250, 0.85)' : 'rgba(79, 70, 229, 0.55)',
        hover: isDark ? 'rgba(139, 92, 246, 0.22)' : 'rgba(99, 102, 241, 0.06)',
        selected: isDark ? 'rgba(139, 92, 246, 0.26)' : 'rgba(99, 102, 241, 0.08)',
        disabled: isDark ? 'rgba(168, 155, 184, 0.45)' : 'rgba(0, 0, 0, 0.26)',
        disabledBackground: isDark ? 'rgba(139, 92, 246, 0.12)' : 'rgba(0, 0, 0, 0.12)',
      },
      background: {
        default: isDark ? '#221c34' : '#e8e4dc',
        paper: isDark ? '#322a4a' : '#f0ece4',
      },
      success: { main: '#2E7D32' },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            touchAction: 'manipulation',
            WebkitTapHighlightColor: 'transparent',
            '@media (max-width:768px)': {
              fontSize: '16px',
              lineHeight: 1.5,
            },
          },
        },
      },
      MuiTypography: {
        styleOverrides: {
          body2: {
            '@media (max-width:768px)': {
              fontSize: '1rem',
              lineHeight: 1.5,
              overflowWrap: 'anywhere',
            },
          },
        },
      },
      MuiButton: {
        styleOverrides: {
          root: {
            textTransform: 'none',
            minHeight: 40,
            '@media (max-width:768px)': {
              minHeight: 44,
            },
            ...(isDark && {
              '&.MuiButton-textPrimary': {
                color: '#a78bfa',
              },
              '&.MuiButton-textSecondary': {
                color: '#fb923c',
              },
              '&.MuiButton-outlinedPrimary': {
                borderColor: 'rgba(167, 139, 250, 0.55)',
                color: '#a78bfa',
              },
            }),
          },
          containedPrimary: {
            background: isDark
              ? 'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)'
              : 'linear-gradient(135deg, #6d6ad8 0%, #5b58c4 100%)',
            color: isDark ? '#1a1028' : '#fafafc',
            '&:hover': {
              background: isDark
                ? 'linear-gradient(135deg, #a78bfa 0%, #8b5cf6 100%)'
                : 'linear-gradient(135deg, #5b58c4 0%, #4f4cb0 100%)',
            },
          },
          sizeSmall: {
            '@media (max-width:768px)': {
              minHeight: 44,
              minWidth: 44,
              fontSize: '1rem',
            },
          },
          containedSecondary: {
            background: isDark
              ? 'linear-gradient(135deg, #f97316 0%, #ea580c 100%)'
              : '#8b7fd4',
            color: isDark ? '#1a1028' : '#fff',
            '&:hover': {
              background: isDark
                ? 'linear-gradient(135deg, #fb923c 0%, #f97316 100%)'
                : '#7c72c8',
            },
          },
        },
      },
      MuiAlert: {
        styleOverrides: {
          message: {
            '@media (max-width:768px)': {
              fontSize: '1rem',
              lineHeight: 1.5,
              overflowWrap: 'anywhere',
            },
          },
          action: {
            '@media (max-width:768px)': {
              '& .MuiIconButton-root': {
                width: 44,
                height: 44,
                minWidth: 44,
                minHeight: 44,
              },
            },
          },
        },
      },
      MuiSnackbar: {
        styleOverrides: {
          root: {
            '@media (max-width:768px)': {
              left: 8,
              right: 8,
              width: 'auto',
              maxWidth: 'calc(100% - 16px)',
              bottom: 'max(8px, env(safe-area-inset-bottom))',
            },
          },
        },
      },
      MuiIconButton: {
        styleOverrides: {
          root: {
            ...(isDark
              ? {
                  '&:hover': { backgroundColor: 'rgba(139, 92, 246, 0.14)' },
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
            backgroundColor: isDark ? '#352e4a' : '#ece8e1',
            borderRight: isDark ? '1px solid rgba(167, 139, 250, 0.28)' : '1px solid rgba(0, 0, 0, 0.08)',
          },
        },
      },
      MuiListItemButton: {
        styleOverrides: {
          root: {
            minHeight: 48,
            '&.Mui-selected': {
              backgroundColor: isDark ? 'rgba(139, 92, 246, 0.22)' : 'rgba(99, 102, 241, 0.08)',
              borderLeft: `3px solid ${isDark ? '#8b5cf6' : '#6b69c7'}`,
            },
          },
        },
      },
    },
  });
}

export const adminTheme = getAdminTheme('light');
