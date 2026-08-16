import React from 'react';
import { Drawer, Box, IconButton, Typography, useMediaQuery } from '@mui/material';
import { useTheme } from '@mui/material/styles';
import CloseIcon from '@mui/icons-material/Close';
import SmartToyOutlinedIcon from '@mui/icons-material/SmartToyOutlined';
import SkoolAiChat from '../../SkoolAiChat';
import { usePlatformUi } from '../../PlatformUiContext';

/** Wide enough for chat content; capped on small desktops so the main grid stays visible */
const DRAWER_WIDTH = 720;

/**
 * Side drawer with the same AI chat as the full app, without file upload.
 */
export default function AiPanel({ open, onClose }) {
  const { labels } = usePlatformUi();
  const theme = useTheme();
  const isPhone = useMediaQuery(theme.breakpoints.down('sm'));

  return (
    <Drawer
      anchor="right"
      open={open}
      onClose={onClose}
      PaperProps={{
        sx: isPhone
          ? {
              width: '100%',
              maxWidth: '100vw',
              top: 0,
              height: '100%',
              maxHeight: '100dvh',
              borderRadius: 0,
              display: 'flex',
              flexDirection: 'column',
              overflow: 'hidden',
              boxSizing: 'border-box',
              pt: 'env(safe-area-inset-top)',
              pb: 'env(safe-area-inset-bottom)',
            }
          : {
              width: { xs: '100%', sm: `min(${DRAWER_WIDTH}px, 96vw)` },
              maxWidth: '100vw',
              top: `calc(56px + env(safe-area-inset-top, 0px))`,
              height: `calc(100% - 56px - env(safe-area-inset-top, 0px) - env(safe-area-inset-bottom, 0px))`,
              borderRadius: '14px 0 0 0',
              display: 'flex',
              flexDirection: 'column',
              overflow: 'hidden',
              boxSizing: 'border-box',
              borderLeft: 1,
              borderColor: 'divider',
            },
      }}
      sx={{
        zIndex: (t) => t.zIndex.drawer + 2,
        '& .MuiDrawer-paper': {
          bgcolor: 'background.paper',
        },
      }}
      ModalProps={{
        keepMounted: false,
        disableScrollLock: true,
        BackdropProps: {
          sx: isPhone
            ? {
                backgroundColor: 'rgba(0,0,0,0.45)',
              }
            : {
                top: `calc(56px + env(safe-area-inset-top, 0px))`,
                height: `calc(100% - 56px - env(safe-area-inset-top, 0px))`,
                backgroundColor: 'rgba(0,0,0,0.38)',
              },
        },
      }}
    >
      <Box
        sx={{
          flexShrink: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          px: 1.5,
          py: 0.75,
          minHeight: 48,
          borderBottom: 1,
          borderColor: 'divider',
          background: (t) =>
            t.palette.mode === 'dark'
              ? 'linear-gradient(90deg, rgba(156,39,176,0.15) 0%, rgba(33,150,243,0.08) 100%)'
              : 'linear-gradient(90deg, rgba(156,39,176,0.12) 0%, rgba(33,150,243,0.06) 100%)',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, minWidth: 0 }}>
          <SmartToyOutlinedIcon sx={{ color: 'primary.main', fontSize: 26, flexShrink: 0 }} />
          <Typography variant="subtitle1" fontWeight={700} noWrap>
            {labels.ai_assistant_name}
          </Typography>
        </Box>
        <IconButton size="small" onClick={onClose} aria-label="Close assistant">
          <CloseIcon fontSize="small" />
        </IconButton>
      </Box>
      <Box
        sx={{
          flex: 1,
          minHeight: 0,
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
          bgcolor: (t) =>
            t.palette.mode === 'dark' ? 'rgba(4, 2, 8, 0.92)' : 'rgba(245, 240, 255, 0.97)',
        }}
      >
        <SkoolAiChat variant="embedded" enableFileUpload={false} singleSession />
      </Box>
    </Drawer>
  );
}
