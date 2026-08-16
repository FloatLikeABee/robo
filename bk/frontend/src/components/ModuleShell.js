import React, { useState } from 'react';
import { Box, IconButton, Typography, Popover } from '@mui/material';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';

/**
 * Standard module chrome: compact title row, optional sticky bar below (e.g. tabs),
 * scrollable workspace, optional footer strip inside shell.
 */
const ModuleShell = ({ title, helpText = '', subtitleBar = null, footerInner = null, children }) => {
  const [helpAnchor, setHelpAnchor] = useState(null);
  const helpOpen = Boolean(helpAnchor);

  return (
    <Box
      sx={{
        height: '100%',
        minHeight: 0,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      <Box
        sx={{
          flexShrink: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 1,
          py: 0.25,
          pr: 0.5,
        }}
      >
        <Typography
          component="h1"
          sx={{
            fontFamily: '"Orbitron", "Roboto", sans-serif',
            fontSize: '0.7rem',
            fontWeight: 600,
            letterSpacing: '0.12em',
            textTransform: 'uppercase',
            color: 'text.primary',
            lineHeight: 1.2,
          }}
        >
          {title}
        </Typography>
        {helpText ? (
          <>
            <IconButton
              size="small"
              aria-label="Module help"
              onClick={(e) => setHelpAnchor(e.currentTarget)}
              sx={{
                p: 0.25,
                color: 'primary.light',
                '& .MuiSvgIcon-root': { fontSize: 16 },
              }}
            >
              <HelpOutlineIcon fontSize="inherit" />
            </IconButton>
            <Popover
              open={helpOpen}
              anchorEl={helpAnchor}
              onClose={() => setHelpAnchor(null)}
              anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
              transformOrigin={{ vertical: 'top', horizontal: 'right' }}
              slotProps={{
                paper: {
                  sx: {
                    maxWidth: 360,
                    p: 1.5,
                    fontSize: '0.75rem',
                    lineHeight: 1.45,
                  },
                },
              }}
            >
              <Typography variant="body2" sx={{ fontSize: '0.75rem', lineHeight: 1.45 }}>
                {helpText}
              </Typography>
            </Popover>
          </>
        ) : null}
      </Box>
      {subtitleBar ? (
        <Box sx={{ flexShrink: 0, mt: 0.25, mb: 0 }}>{subtitleBar}</Box>
      ) : null}
      <Box
        sx={{
          flex: 1,
          minHeight: 0,
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          pt: subtitleBar ? 0.75 : 0.5,
        }}
      >
        {children}
      </Box>
      {footerInner ? (
        <Box
          sx={{
            flexShrink: 0,
            borderTop: 1,
            borderColor: 'divider',
            py: 0.25,
            px: 0.5,
            fontSize: '0.65rem',
            color: 'text.secondary',
          }}
        >
          {footerInner}
        </Box>
      ) : null}
    </Box>
  );
};

export default ModuleShell;
