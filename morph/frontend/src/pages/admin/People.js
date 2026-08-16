import React from 'react';
import { Box } from '@mui/material';
import Students from './Students';

/**
 * People — single people list (former members). No employee split.
 */
export default function People() {
  return (
    <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
      <Students titleOverride="People" />
    </Box>
  );
}
