import React, { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAdminBasePath } from '../../adminPaths';
import { usePlatformUi } from '../../PlatformUiContext';
import { isMorphAdmin } from '../../auth/isMorphAdmin';
import { getMorphAuthSnapshot } from '../../auth/morphSession';
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
  Collapse,
  useMediaQuery,
} from '@mui/material';
import { useTheme } from '@mui/material/styles';
import {
  ChevronLeft as ChevronLeftIcon,
  AssignmentTurnedIn as CaseTaskIcon,
  Timeline as TimelinesIcon,
  Settings as SettingsIcon,
  ExpandLess,
  ExpandMore,
  NotesOutlined as BigNotesIcon,
  DatasetOutlined as GenericDataIcon,
  Inventory2Outlined as AssetsIcon,
} from '@mui/icons-material';

/** Fixed desktop width — sized for longest labels (Generic data / Configuration / File import). */
export const DRAWER_WIDTH = 200;
const MORPH_DATA_LOGO = `${process.env.PUBLIC_URL || ''}/icons/morph-data-icon.svg`;
const SS_CONFIGURATION = 'morphdata.drawer.configurationOpen';

function readSessionBool(key) {
  try {
    const v = sessionStorage.getItem(key);
    if (v === '1') return true;
    if (v === '0') return false;
  } catch {
    /* ignore */
  }
  return null;
}

function writeSessionBool(key, value) {
  try {
    sessionStorage.setItem(key, value ? '1' : '0');
  } catch {
    /* ignore */
  }
}

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
  const adminNav = isMorphAdmin() || Boolean(getMorphAuthSnapshot()?.user?.is_admin);

  const [configurationOpen, setConfigurationOpen] = useState(() => {
    const saved = readSessionBool(SS_CONFIGURATION);
    if (saved !== null) return saved;
    const path = typeof window !== 'undefined' ? window.location.pathname : '';
    return path.startsWith(`${base}/configuration`);
  });

  useEffect(() => {
    if (location.pathname.startsWith(`${base}/configuration`)) {
      setConfigurationOpen(true);
      writeSessionBool(SS_CONFIGURATION, true);
    }
  }, [location.pathname, base]);

  const isSelected = (path) => location.pathname === path || location.pathname.startsWith(path + '/');

  const go = (to) => {
    if (location.pathname === to) {
      if (isMobile && typeof onMobileClose === 'function') onMobileClose();
      return;
    }
    navigate(to);
    if (isMobile && typeof onMobileClose === 'function') onMobileClose();
  };

  const toggleConfiguration = () => {
    setConfigurationOpen((prev) => {
      const next = !prev;
      writeSessionBool(SS_CONFIGURATION, next);
      return next;
    });
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
    minHeight: 40,
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
            size="small"
            aria-label="Close navigation"
            sx={{ color: 'text.secondary', flexShrink: 0 }}
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
          selected={isSelected(base + '/generic-data')}
          onClick={() => go(`${base}/generic-data`)}
          sx={navButtonSx}
        >
          <ListItemIcon sx={iconSx}>
            <GenericDataIcon sx={{ color: 'secondary.main', fontSize: 20 }} />
          </ListItemIcon>
          <ListItemText primary="Generic data" sx={textPrimarySx} />
        </ListItemButton>
        <ListItemButton
          selected={
            isSelected(base + '/assets') ||
            isSelected(base + '/resources') ||
            isSelected(base + '/people') ||
            isSelected(base + '/members') ||
            isSelected(base + '/employees')
          }
          onClick={() => go(`${base}/assets`)}
          sx={navButtonSx}
        >
          <ListItemIcon sx={iconSx}>
            <AssetsIcon sx={{ color: 'secondary.main', fontSize: 20 }} />
          </ListItemIcon>
          <ListItemText primary="Assets" sx={textPrimarySx} />
        </ListItemButton>
        <Divider sx={{ borderColor: 'grey.700', my: 0.75 }} />
        <ListItemButton onClick={toggleConfiguration} sx={navButtonSx}>
          <ListItemIcon sx={iconSx}>
            <SettingsIcon sx={{ color: 'grey.500', fontSize: 20 }} />
          </ListItemIcon>
          <ListItemText primary="Configuration" sx={textPrimarySx} />
          {configurationOpen ? <ExpandLess fontSize="small" /> : <ExpandMore fontSize="small" />}
        </ListItemButton>
        <Collapse in={configurationOpen} timeout="auto">
          <List component="div" disablePadding>
            {adminNav && (
              <>
                <ListItemButton
                  selected={isSelected(base + '/configuration/users')}
                  onClick={() => go(`${base}/configuration/users`)}
                  sx={{ ...navButtonSx, pl: 5.5 }}
                >
                  <ListItemText primary="Users" sx={textPrimarySx} />
                </ListItemButton>
                <ListItemButton
                  selected={
                    isSelected(base + '/configuration/file-import') ||
                    isSelected(base + '/configuration/data-import')
                  }
                  onClick={() => go(`${base}/configuration/file-import`)}
                  sx={{ ...navButtonSx, pl: 5.5 }}
                >
                  <ListItemText primary="File import" sx={textPrimarySx} />
                </ListItemButton>
              </>
            )}
          </List>
        </Collapse>
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
            width: 'min(220px, 86vw)',
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
