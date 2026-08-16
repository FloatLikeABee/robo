import { createTheme } from '@mui/material/styles';

// Sci-Fi Color Palette — dark blue & grey
const colors = {
  darkBlue: '#0a0f1a',
  darkerBlue: '#060a12',
  blue: '#1a2d4a',
  lightBlue: '#243b5c',
  highlight: '#38bdf8',
  highlightDark: '#0ea5e9',
  accent: '#3b82f6',
  accentLight: '#60a5fa',
  accentDark: '#1d4ed8',
  darkGrey: '#1a1f2e',
  black: '#000000',
  lightGrey: '#2d3748',
  midGrey: '#64748b',
  silver: '#94a3b8',
  textPrimary: '#e2e8f0',
  textSecondary: '#94a3b8',
};

const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: colors.accent,
      light: colors.accentLight,
      dark: colors.accentDark,
      contrastText: colors.textPrimary,
    },
    secondary: {
      main: colors.silver,
      light: '#cbd5e1',
      dark: colors.midGrey,
      contrastText: colors.darkerBlue,
    },
    background: {
      default: colors.darkBlue,
      paper: colors.darkerBlue,
    },
    text: {
      primary: colors.textPrimary,
      secondary: colors.textSecondary,
    },
    error: {
      main: '#ff4444',
    },
    warning: {
      main: colors.highlight,
    },
    info: {
      main: colors.accentLight,
    },
    success: {
      main: '#00ff88',
    },
  },
  typography: {
    fontFamily: '"Roboto", "Inter", "Helvetica", "Arial", sans-serif',
    h1: {
      fontFamily: '"Orbitron", "Roboto", sans-serif',
      fontWeight: 700,
      letterSpacing: '0.03em',
      textTransform: 'uppercase',
    },
    h2: {
      fontFamily: '"Orbitron", "Roboto", sans-serif',
      fontWeight: 700,
      letterSpacing: '0.03em',
      textTransform: 'uppercase',
    },
    h3: {
      fontFamily: '"Orbitron", "Roboto", sans-serif',
      fontWeight: 600,
      letterSpacing: '0.02em',
    },
    h4: {
      fontFamily: '"Orbitron", "Roboto", sans-serif',
      fontWeight: 600,
      letterSpacing: '0.02em',
      fontSize: '1rem',
    },
    h5: {
      fontFamily: '"Orbitron", "Roboto", sans-serif',
      fontWeight: 600,
      letterSpacing: '0.01em',
      fontSize: '0.9rem',
    },
    h6: {
      fontFamily: '"Orbitron", "Roboto", sans-serif',
      fontWeight: 600,
      fontSize: '0.75rem',
    },
    subtitle2: {
      fontFamily: '"Orbitron", "Roboto", sans-serif',
      fontWeight: 600,
      fontSize: '0.72rem',
    },
    body1: {
      fontFamily: '"Roboto", "Inter", sans-serif',
      letterSpacing: '0.01em',
      lineHeight: 1.5,
      fontSize: '0.8125rem',
    },
    body2: {
      fontFamily: '"Roboto", "Inter", sans-serif',
      letterSpacing: '0.01em',
      lineHeight: 1.45,
      fontSize: '0.75rem',
    },
    button: {
      fontFamily: '"Orbitron", "Roboto", sans-serif',
      fontWeight: 600,
      letterSpacing: '0.05em',
      textTransform: 'uppercase',
    },
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        html: {
          height: '100%',
          overflow: 'hidden',
        },
        body: {
          height: '100%',
          overflow: 'hidden',
          background: `linear-gradient(135deg, ${colors.darkBlue} 0%, ${colors.darkerBlue} 100%)`,
          backgroundAttachment: 'fixed',
          '&::before': {
            content: '""',
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: `radial-gradient(circle at 20% 50%, ${colors.blue}18 0%, transparent 50%),
                         radial-gradient(circle at 80% 80%, ${colors.highlight}0a 0%, transparent 50%)`,
            pointerEvents: 'none',
            zIndex: 0,
          },
        },
        '#root': {
          height: '100%',
        },
        '.MuiFormControl-root': {
          overflow: 'visible !important',
          '& .MuiInputLabel-root': {
            overflow: 'visible !important',
          },
        },
        '.MuiTextField-root': {
          overflow: 'visible !important',
          '& .MuiInputLabel-root': {
            overflow: 'visible !important',
          },
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          background: `linear-gradient(135deg, ${colors.darkerBlue} 0%, ${colors.darkBlue} 100%)`,
          borderBottom: `2px solid ${colors.accent}40`,
          boxShadow: `0 4px 20px ${colors.accent}20, 0 0 40px ${colors.blue}25`,
          backdropFilter: 'blur(10px)',
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          background: `linear-gradient(135deg, ${colors.darkerBlue} 0%, ${colors.darkGrey} 100%)`,
          border: `1px solid ${colors.accent}30`,
          borderRadius: '8px',
          boxShadow: `0 4px 16px ${colors.black}70, 
                      0 0 12px ${colors.accent}08,
                      inset 0 1px 0 ${colors.accent}15`,
          transition: 'all 0.2s ease',
          '&:hover': {
            borderColor: `${colors.accent}55`,
            boxShadow: `0 6px 20px ${colors.black}85, 
                        0 0 18px ${colors.accent}15,
                        inset 0 1px 0 ${colors.accent}25`,
            transform: 'translateY(-1px)',
          },
        },
      },
    },
    MuiCardContent: {
      styleOverrides: {
        root: {
          padding: '10px 12px',
          '&:last-child': {
            paddingBottom: '10px',
          },
        },
      },
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: '8px',
          padding: '10px 24px',
          fontWeight: 600,
          textTransform: 'uppercase',
          letterSpacing: '0.1em',
          transition: 'all 0.3s ease',
          '&.MuiButton-sizeSmall': {
            padding: '3px 10px',
            fontSize: '0.65rem',
            letterSpacing: '0.06em',
            minHeight: 26,
          },
          '&.MuiButton-containedPrimary': {
            background: `linear-gradient(135deg, ${colors.accent} 0%, ${colors.accentDark} 100%)`,
            boxShadow: `0 4px 15px ${colors.accent}40, 0 0 20px ${colors.accent}20`,
            '&:hover': {
              background: `linear-gradient(135deg, ${colors.accentDark} 0%, ${colors.accent} 100%)`,
              boxShadow: `0 6px 20px ${colors.accent}60, 0 0 30px ${colors.accent}30`,
              transform: 'translateY(-2px)',
            },
          },
          '&.MuiButton-containedSecondary': {
            background: `linear-gradient(135deg, ${colors.midGrey} 0%, ${colors.lightGrey} 100%)`,
            boxShadow: `0 4px 15px ${colors.midGrey}40, 0 0 20px ${colors.silver}15`,
            '&:hover': {
              background: `linear-gradient(135deg, ${colors.lightGrey} 0%, ${colors.silver} 100%)`,
              boxShadow: `0 6px 20px ${colors.midGrey}50, 0 0 30px ${colors.silver}20`,
              transform: 'translateY(-2px)',
            },
          },
          '&.MuiButton-outlined': {
            borderWidth: '2px',
            '&.MuiButton-outlinedPrimary': {
              borderColor: colors.accent,
              color: colors.accent,
              '&:hover': {
                borderColor: colors.accent,
                background: `${colors.accent}15`,
                boxShadow: `0 0 20px ${colors.accent}30`,
              },
            },
            '&.MuiButton-outlinedSecondary': {
              borderColor: colors.silver,
              color: colors.silver,
              '&:hover': {
                borderColor: colors.silver,
                background: `${colors.silver}15`,
                boxShadow: `0 0 20px ${colors.silver}25`,
              },
            },
          },
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          overflow: 'visible',
          '& .MuiInputBase-root': {
            overflow: 'visible',
            '& .MuiInputLabel-root': {
              overflow: 'visible !important',
            },
          },
          '& .MuiOutlinedInput-root': {
            borderRadius: '8px',
            overflow: 'visible',
            '& fieldset': {
              borderColor: `${colors.accent}40`,
              borderWidth: '2px',
            },
            '&:hover fieldset': {
              borderColor: `${colors.accent}60`,
            },
            '&.Mui-focused fieldset': {
              borderColor: colors.accent,
              boxShadow: `0 0 15px ${colors.accent}30`,
            },
            '& input': {
              color: colors.textPrimary,
            },
            '& textarea': {
              color: colors.textPrimary,
            },
          },
          '& .MuiInputLabel-root': {
            color: colors.textSecondary,
            overflow: 'visible !important',
            '&.Mui-focused': {
              color: colors.accent,
            },
          },
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: '4px',
          fontWeight: 600,
          height: 22,
          fontSize: '0.625rem',
          '& .MuiChip-label': { px: 0.75, py: 0 },
          '&.MuiChip-colorPrimary': {
            background: `${colors.accent}20`,
            color: colors.accent,
            border: `1px solid ${colors.accent}40`,
          },
          '&.MuiChip-colorSecondary': {
            background: `${colors.silver}20`,
            color: colors.silver,
            border: `1px solid ${colors.silver}40`,
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          background: `linear-gradient(135deg, ${colors.darkerBlue} 0%, ${colors.darkGrey} 100%)`,
          border: `1px solid ${colors.accent}20`,
        },
      },
    },
    MuiLinearProgress: {
      styleOverrides: {
        root: {
          borderRadius: '4px',
          backgroundColor: `${colors.darkGrey}`,
          '& .MuiLinearProgress-bar': {
            background: `linear-gradient(90deg, ${colors.accent} 0%, ${colors.highlight} 100%)`,
            boxShadow: `0 0 10px ${colors.accent}50`,
          },
        },
      },
    },
    MuiAlert: {
      styleOverrides: {
        root: {
          borderRadius: '8px',
          border: '1px solid',
          '&.MuiAlert-standardInfo': {
            borderColor: `${colors.accent}40`,
            background: `${colors.accent}10`,
            color: colors.accent,
          },
          '&.MuiAlert-standardSuccess': {
            borderColor: '#00ff8840',
            background: '#00ff8810',
            color: '#00ff88',
          },
          '&.MuiAlert-standardWarning': {
            borderColor: `${colors.highlight}40`,
            background: `${colors.highlight}10`,
            color: colors.highlight,
          },
          '&.MuiAlert-standardError': {
            borderColor: '#ff444440',
            background: '#ff444410',
            color: '#ff4444',
          },
        },
      },
    },
    MuiMenuItem: {
      styleOverrides: {
        root: {
          '&:hover': {
            background: `${colors.accent}20`,
          },
          '&.Mui-selected': {
            background: `${colors.accent}30`,
            '&:hover': {
              background: `${colors.accent}40`,
            },
          },
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          background: `linear-gradient(135deg, ${colors.darkerBlue} 0%, ${colors.darkGrey} 100%)`,
          border: `2px solid ${colors.accent}40`,
          boxShadow: `0 20px 60px ${colors.black}90, 0 0 40px ${colors.accent}20`,
        },
      },
    },
    MuiTabs: {
      styleOverrides: {
        root: {
          borderBottom: `2px solid ${colors.accent}30`,
        },
        indicator: {
          background: colors.accent,
          boxShadow: `0 0 10px ${colors.accent}50`,
          height: '3px',
        },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          fontFamily: '"Orbitron", "Roboto", sans-serif',
          fontSize: '0.68rem',
          minHeight: 36,
          color: colors.textSecondary,
          fontWeight: 600,
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          '&.Mui-selected': {
            color: colors.accent,
          },
          '&:hover': {
            color: colors.accent,
          },
        },
      },
    },
    MuiTypography: {
      styleOverrides: {
        root: {
          '&.MuiTypography-body1, &.MuiTypography-body2, &.MuiTypography-caption, &.MuiTypography-overline': {
            fontFamily: '"Roboto", "Inter", sans-serif',
          },
        },
      },
    },
    MuiFormControl: {
      styleOverrides: {
        root: {
          overflow: 'visible !important',
          '& .MuiInputLabel-root': {
            overflow: 'visible !important',
            whiteSpace: 'nowrap',
            textOverflow: 'ellipsis',
          },
        },
      },
    },
    MuiInputBase: {
      styleOverrides: {
        root: {
          '&.MuiInputBase-formControl': {
            overflow: 'visible',
            '& .MuiInputLabel-root': {
              overflow: 'visible !important',
            },
          },
        },
      },
    },
    MuiIconButton: {
      styleOverrides: {
        root: {
          '& .MuiSvgIcon-root': {
            fontSize: '1.5rem',
            fill: 'currentColor',
          },
        },
      },
    },
    MuiInputLabel: {
      styleOverrides: {
        root: {
          overflow: 'visible !important',
          whiteSpace: 'nowrap',
          textOverflow: 'ellipsis',
          zIndex: 1,
          lineHeight: '1.4375em',
        },
        outlined: {
          overflow: 'visible !important',
          lineHeight: '1.4375em',
        },
      },
    },
  },
});

export default theme;
