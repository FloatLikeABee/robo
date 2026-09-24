import React, { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAdminBasePath } from '../../adminPaths';
import { usePlatformUi } from '../../PlatformUiContext';
import {
  Drawer,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  IconButton,
  Divider,
  Typography,
  Box,
  useMediaQuery,
} from '@mui/material';
import { useTheme } from '@mui/material/styles';
import {
  ChevronLeft as ChevronLeftIcon,
  AssignmentTurnedIn as CaseTaskIcon,
  Timeline as TimelinesIcon,
  NotesOutlined as BigNotesIcon,
  DatasetOutlined as GenericDataIcon,
  TravelExplore as ResearchIcon,
} from '@mui/icons-material';

/** Fixed desktop width — sized for longest labels (Generic data / Settings). */
export const DRAWER_WIDTH = 200;
const MORPH_DATA_LOGO = `${process.env.PUBLIC_URL || ''}/icons/morph-data-icon.svg`;

export default function AppDrawer({ mobileOpen = false, onMobileClose }) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const location = useLocation();
  const navigate = useNavigate();
  const base = useAdminBasePath();
  const { labels: L } = usePlatformUi();
  const [, setAuthTick] = useState(0);
  useEffect(() => {
    const onAuth = () => setAuthTick((n) => n + 1);
    window.addEventListener('morph-auth-updated', onAuth);
    return () => window.removeEventListener('morph-auth-updated', onAuth);
  }, []);

  const isSelected = (path) => location.pathname === path || location.pathname.startsWith(path + '/');

  const go = (to) => {
    if (location.pathname === to) {
      if (isMobile && typeof onMobileClose === 'function') onMobileClose();
      return;
    }
    navigate(to);
    if (isMobile && typeof onMobileClose === 'function') onMobileClose();
  };

  const drawerPaperSx = {
    width: DRAWER_WIDTH,
    boxSizing: 'border-box',
    overflowX: 'hidden',
    zIndex: (t) => t.zIndex.modal + 10,
    pt: 'env(safe-area-inset-top)',
  };

  const navButtonSx = {
    borderRadius: 0,
    pl: 1.5,
    pr: 1,
    minHeight: 44,
    py: 0.5,
  };
  const iconSx = { minWidth: 32 };
  const textPrimarySx = {
    margin: 0,
    '& .MuiListItemText-primary': { fontSize: '0.875rem', fontWeight: 500, lineHeight: 1.25 },
  };

  const drawerBody = (
    <>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          py: 1.25,
          px: 1.25,
          gap: 1,
          justifyContent: 'space-between',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 0, flex: 1 }}>
          <Box
            component="img"
            src={MORPH_DATA_LOGO}
            alt=""
            sx={{
              width: 28,
              height: 28,
              borderRadius: 1.25,
              flexShrink: 0,
              boxShadow: (t) =>
                t.palette.mode === 'dark'
                  ? '0 0 18px rgba(37, 99, 235, 0.45), 0 0 8px rgba(34, 211, 238, 0.25)'
                  : '0 4px 14px rgba(91, 91, 212, 0.25)',
            }}
          />
          <Typography
            variant="subtitle1"
            noWrap
            sx={{ color: 'primary.main', fontWeight: 700, fontSize: '0.95rem', letterSpacing: '-0.01em' }}
          >
            {L.product_name}
          </Typography>
        </Box>
        {isMobile && (
          <IconButton
            onClick={() => typeof onMobileClose === 'function' && onMobileClose()}
            aria-label="Close navigation"
            sx={{ color: 'text.secondary', flexShrink: 0, width: 44, height: 44 }}
          >
            <ChevronLeftIcon />
          </IconButton>
        )}
      </Box>
      <Divider sx={{ borderColor: 'divider' }} />
      <List
        component="nav"
        sx={{
          pt: 0.75,
          pb: 'max(12px, env(safe-area-inset-bottom))',
        }}
      >
        <ListItemButton
          selected={isSelected(base + '/case-tasks')}
          onClick={() => go(`${base}/case-tasks`)}
          sx={navButtonSx}
        >
          <ListItemIcon sx={iconSx}>
            <CaseTaskIcon color="primary" fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Tasks" sx={textPrimarySx} />
        </ListItemButton>
        <ListItemButton
          selected={isSelected(base + '/timelines') || isSelected(base + '/stories') || isSelected(base + '/story-board')}
          onClick={() => go(`${base}/timelines`)}
          sx={navButtonSx}
        >
          <ListItemIcon sx={iconSx}>
            <TimelinesIcon sx={{ color: 'secondary.main', fontSize: 20 }} />
          </ListItemIcon>
          <ListItemText primary="Timelines" sx={textPrimarySx} />
        </ListItemButton>
        <ListItemButton
          selected={isSelected(base + '/big-notes')}
          onClick={() => go(`${base}/big-notes`)}
          sx={navButtonSx}
        >
          <ListItemIcon sx={iconSx}>
            <BigNotesIcon sx={{ color: 'secondary.main', fontSize: 20 }} />
          </ListItemIcon>
          <ListItemText primary="Big notes" sx={textPrimarySx} />
        </ListItemButton>
        <ListItemButton
          selected={isSelected(base + '/research')}
          onClick={() => go(`${base}/research`)}
          sx={navButtonSx}
        >
          <ListItemIcon sx={iconSx}>
            <ResearchIcon sx={{ color: 'secondary.main', fontSize: 20 }} />
          </ListItemIcon>
          <ListItemText primary="Research" sx={textPrimarySx} />
        </ListItemButton>
        <ListItemButton
          selected={isSelected(base + '/generic-data')}
          onClick={() => go(`${base}/generic-data`)}
          sx={navButtonSx}
        >
          <ListItemIcon sx={iconSx}>
            <GenericDataIcon sx={{ color: 'secondary.main', fontSize: 20 }} />
          </ListItemIcon>
          <ListItemText primary="Generic data" sx={textPrimarySx} />
        </ListItemButton>
      </List>
    </>
  );

  return (
    <>
      <Drawer
        variant="temporary"
        open={mobileOpen}
        onClose={onMobileClose}
        ModalProps={{ keepMounted: true }}
        sx={{
          display: { xs: 'block', md: 'none' },
          zIndex: (t) => t.zIndex.modal + 10,
          '& .MuiDrawer-paper': {
            ...drawerPaperSx,
            width: '100%',
          },
        }}
      >
        {drawerBody}
      </Drawer>
      <Drawer
        variant="permanent"
        open
        sx={{
          display: { xs: 'none', md: 'block' },
          width: DRAWER_WIDTH,
          flexShrink: 0,
          zIndex: (t) => t.zIndex.modal + 10,
          '& .MuiDrawer-paper': drawerPaperSx,
        }}
      >
        {drawerBody}
      </Drawer>
    </>
  );
}
