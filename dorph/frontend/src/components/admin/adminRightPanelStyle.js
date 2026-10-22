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
  borderColor: 'rgba(37, 99, 235, 0.3)',
  overflow: 'hidden',
  bgcolor: '#0c1220',
};

export const adminRightPanelHeaderSx = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  px: 1.5,
  py: 0.75,
  borderBottom: 1,
  borderColor: 'rgba(37, 99, 235, 0.3)',
  minHeight: 48,
  flexShrink: 0,
  background: 'linear-gradient(90deg, rgba(37, 99, 235, 0.22) 0%, rgba(12, 18, 32, 0.55) 100%)',
};
