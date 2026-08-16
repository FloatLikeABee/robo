import { useTheme } from '@mui/material/styles';

/**
 * Keep in sync with dark MuiListItemButton selected + MuiDrawer paper in theme.js.
 * Nav highlight: rgba(37, 99, 235, 0.24) + 3px #2563eb accent.
 */
export const NAV_HIGHLIGHT_BG_DARK = 'rgba(37, 99, 235, 0.24)';
export const NAV_HIGHLIGHT_BG_NESTED_DARK = 'rgba(37, 99, 235, 0.16)';
export const NAV_HIGHLIGHT_ACCENT_DARK = '#2563eb';
export const NAV_HIGHLIGHT_BORDER_DARK = '1px solid rgba(37, 99, 235, 0.32)';

/** Detail drawer shell — near-black blue (cards use nav-highlight overlay on top). */
export const DETAIL_DRAWER_BG_DARK = '#020408';
export const DETAIL_PANEL_BG_DARK = '#020408';

export const DETAIL_CARD_BG_DARK = NAV_HIGHLIGHT_BG_DARK;
export const DETAIL_CARD_NESTED_BG_DARK = NAV_HIGHLIGHT_BG_NESTED_DARK;

export function getDetailCardSx(isDark, { nested = false } = {}) {
  if (!isDark) {
    return {
      bgcolor: 'rgba(99, 102, 241, 0.08)',
      border: '1px solid rgba(99, 102, 241, 0.18)',
    };
  }
  return {
    bgcolor: nested ? DETAIL_CARD_NESTED_BG_DARK : DETAIL_CARD_BG_DARK,
    border: NAV_HIGHLIGHT_BORDER_DARK,
    borderLeft: `3px solid ${NAV_HIGHLIGHT_ACCENT_DARK}`,
    boxShadow: 'none',
  };
}

/** Field / attribute cards in the detail drawer (dark). */
export const DETAIL_CARD_SX_DARK = getDetailCardSx(true);

export function useDetailJsonSurfaceSx(options = {}) {
  const theme = useTheme();
  return getDetailCardSx(theme.palette.mode === 'dark', options);
}
