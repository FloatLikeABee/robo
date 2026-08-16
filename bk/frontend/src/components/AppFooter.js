import React from 'react';
import { Box } from '@mui/material';

const AppFooter = () => (
  <Box
    component="footer"
    sx={{
      flexShrink: 0,
      py: 1.25,
      px: 2,
      fontSize: '0.6875rem',
      letterSpacing: '0.08em',
      textTransform: 'uppercase',
      color: 'text.secondary',
      borderTop: 1,
      borderColor: 'divider',
      bgcolor: 'background.paper',
      textAlign: 'center',
      lineHeight: 1.4,
      minHeight: 40,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    }}
  >
    AI tools
  </Box>
);

export default AppFooter;
