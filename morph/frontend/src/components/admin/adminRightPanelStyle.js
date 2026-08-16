/** Matches global admin header + footer in AdminLayout */
export const ADMIN_RIGHT_PANEL_TOP = 56;
export const ADMIN_RIGHT_PANEL_FOOTER = 40;
export const ADMIN_RIGHT_PANEL_WIDTH = 620;

export const adminRightPanelHeightCalc = `calc(100dvh - ${ADMIN_RIGHT_PANEL_TOP}px - ${ADMIN_RIGHT_PANEL_FOOTER}px - env(safe-area-inset-top, 0px) - env(safe-area-inset-bottom, 0px))`;

export const adminRightPanelPaperSx = {
  position: 'fixed',
  top: `calc(${ADMIN_RIGHT_PANEL_TOP}px + env(safe-area-inset-top, 0px))`,
  right: 0,
  width: { xs: '100%', sm: 'min(100vw, 520px)', md: ADMIN_RIGHT_PANEL_WIDTH },
  maxWidth: '100%',
  height: adminRightPanelHeightCalc,
  maxHeight: adminRightPanelHeightCalc,
  zIndex: 1300,
  display: 'flex',
  flexDirection: 'column',
  borderRadius: 0,
  borderLeft: { xs: 0, sm: 1 },
  borderBottom: 1,
  borderColor: 'divider',
  overflow: 'hidden',
  bgcolor: 'background.paper',
};

export const adminRightPanelHeaderSx = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  px: 1.5,
  py: 0.75,
  borderBottom: 1,
  borderColor: 'divider',
  minHeight: 48,
  flexShrink: 0,
  background: (theme) =>
    theme.palette.mode === 'dark'
      ? 'linear-gradient(90deg, rgba(156,39,176,0.15) 0%, rgba(33,150,243,0.08) 100%)'
      : 'linear-gradient(90deg, rgba(156,39,176,0.12) 0%, rgba(33,150,243,0.06) 100%)',
};
