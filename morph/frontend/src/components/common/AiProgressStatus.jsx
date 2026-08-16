import React from 'react';
import { Box, Typography } from '@mui/material';

export default function AiProgressStatus({ status, sx = {} }) {
  if (!status) return null;
  return (
    <Box
      role="status"
      aria-live="polite"
      sx={{
        display: 'flex',
        alignItems: 'center',
        gap: 1,
        px: 1,
        py: 0.75,
        borderRadius: 1,
        bgcolor: 'action.hover',
        ...sx,
      }}
    >
      <Box
        component="span"
        aria-hidden="true"
        sx={{
          width: 8,
          height: 8,
          borderRadius: '50%',
          bgcolor: 'primary.main',
          flexShrink: 0,
          animation: 'morphAiProgressPulse 1.1s ease-in-out infinite',
          '@keyframes morphAiProgressPulse': {
            '0%, 100%': { opacity: 0.35, transform: 'scale(0.92)' },
            '50%': { opacity: 1, transform: 'scale(1)' },
          },
        }}
      />
      <Typography variant="caption" color="text.secondary" sx={{ lineHeight: 1.35 }}>
        {status}
      </Typography>
    </Box>
  );
}
