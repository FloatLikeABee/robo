import React, { useState } from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Box,
  IconButton,
  Menu,
  MenuItem,
  Tooltip,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import { alpha } from '@mui/material/styles';
import {
  SmartToy as AssistantIcon,
  Info as StatusIcon,
  Menu as MenuIcon,
  Description as DocumentsIcon,
  Image as ImageIcon,
  ImportContacts as ReadersIcon,
  Movie as VideoStoryIcon,
  Storage as RagIcon,
} from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';
import { useQuery } from 'react-query';
import api from '../services/api';

const Header = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [anchorEl, setAnchorEl] = useState(null);

  useQuery('status', api.getStatus, {
    refetchInterval: 30000,
    retry: 1,
  });

  const navItems = [
    { path: '/assistants', label: 'Assistants', icon: <AssistantIcon /> },
    { path: '/rag', label: 'RAG', icon: <RagIcon /> },
    { path: '/images', label: 'Images', icon: <ImageIcon /> },
    { path: '/video-stories', label: 'Video stories', icon: <VideoStoryIcon /> },
    { path: '/readers', label: 'Readers', icon: <ReadersIcon /> },
    { path: '/documents', label: 'Documents', icon: <DocumentsIcon /> },
    { path: '/status', label: 'System', icon: <StatusIcon /> },
  ];

  const isSelected = (path) => location.pathname === path;

  const handleMenuOpen = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleNavigate = (path) => {
    navigate(path);
    handleMenuClose();
  };

  return (
    <AppBar position="sticky" sx={{ top: 0, zIndex: 1300, flexShrink: 0 }}>
      <Toolbar variant="dense" sx={{ color: 'text.primary', minHeight: 48, px: 2 }}>
        <Typography
          variant="h6"
          component="div"
          sx={{
            flexGrow: 1,
            fontFamily: '"Orbitron", "Roboto", sans-serif',
            fontWeight: 700,
            fontSize: '0.8rem',
            letterSpacing: '0.08em',
            textTransform: 'uppercase',
            background: (t) => `linear-gradient(135deg, ${t.palette.primary.main} 0%, ${t.palette.info.main} 100%)`,
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            backgroundClip: 'text',
            textShadow: (t) => `0 0 30px ${alpha(t.palette.primary.main, 0.5)}`,
          }}
        >
          AI tools
        </Typography>

        {isMobile ? (
          <>
            <IconButton
              color="inherit"
              onClick={handleMenuOpen}
              edge="end"
              sx={{
                color: 'primary.main',
                '& .MuiSvgIcon-root': { fontSize: 22, fill: 'currentColor' },
                '&:hover': {
                  color: 'info.main',
                  backgroundColor: (t) => alpha(t.palette.primary.main, 0.1),
                },
              }}
            >
              <MenuIcon />
            </IconButton>
            <Menu
              anchorEl={anchorEl}
              open={Boolean(anchorEl)}
              onClose={handleMenuClose}
              anchorOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              transformOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              PaperProps={{
                sx: {
                  background: (t) => `linear-gradient(135deg, ${t.palette.background.paper} 0%, ${t.palette.background.default} 100%)`,
                  border: (t) => `1px solid ${alpha(t.palette.primary.main, 0.3)}`,
                  boxShadow: (t) => `0 8px 32px rgba(0, 0, 0, 0.8), 0 0 20px ${alpha(t.palette.primary.main, 0.1)}`,
                },
              }}
            >
              {navItems.map((item) => (
                <MenuItem
                  key={item.path}
                  onClick={() => handleNavigate(item.path)}
                  selected={isSelected(item.path)}
                  sx={{
                    color: isSelected(item.path) ? 'primary.main' : 'text.primary',
                    fontSize: '0.78rem',
                    '&:hover': {
                      backgroundColor: (t) => alpha(t.palette.primary.main, 0.2),
                    },
                    '&.Mui-selected': {
                      backgroundColor: (t) => alpha(t.palette.primary.main, 0.3),
                      '&:hover': {
                        backgroundColor: (t) => alpha(t.palette.primary.main, 0.4),
                      },
                    },
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    {item.icon}
                    {item.label}
                  </Box>
                </MenuItem>
              ))}
            </Menu>
          </>
        ) : (
          <Box sx={{ display: 'flex', gap: 0.35, alignItems: 'center', flexWrap: 'wrap', justifyContent: 'flex-end' }}>
            {navItems.map((item) => (
              <Tooltip key={item.path} title={item.label} arrow placement="bottom">
                <IconButton
                  color="inherit"
                  onClick={() => navigate(item.path)}
                  sx={{
                    color: isSelected(item.path) ? 'primary.main' : 'text.secondary',
                    backgroundColor: isSelected(item.path)
                      ? (t) => alpha(t.palette.primary.main, 0.15)
                      : 'transparent',
                    border: isSelected(item.path)
                      ? (t) => `1px solid ${alpha(t.palette.primary.main, 0.4)}`
                      : '1px solid transparent',
                    borderRadius: '8px',
                    transition: 'all 0.3s ease',
                    padding: '4px',
                    '& .MuiSvgIcon-root': {
                      fontSize: 20,
                      fill: 'currentColor',
                    },
                    '&:hover': {
                      color: 'primary.main',
                      backgroundColor: (t) => alpha(t.palette.primary.main, 0.2),
                      border: (t) => `1px solid ${alpha(t.palette.primary.main, 0.5)}`,
                      boxShadow: (t) => `0 0 15px ${alpha(t.palette.primary.main, 0.3)}`,
                      transform: 'translateY(-2px)',
                    },
                  }}
                >
                  {item.icon}
                </IconButton>
              </Tooltip>
            ))}
          </Box>
        )}
      </Toolbar>
    </AppBar>
  );
};

export default Header;
