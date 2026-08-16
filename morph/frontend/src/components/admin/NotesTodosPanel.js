import React from 'react';
import { Paper, IconButton, Box, Typography } from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import EventNoteOutlinedIcon from '@mui/icons-material/EventNoteOutlined';
import NotesTodosContent from '../notesTodos/NotesTodosContent';
import { adminRightPanelPaperSx, adminRightPanelHeaderSx } from './adminRightPanelStyle';

export default function NotesTodosPanel({ open, onClose }) {
  if (!open) return null;

  return (
    <Paper elevation={10} sx={adminRightPanelPaperSx}>
      <Box sx={adminRightPanelHeaderSx}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
          <EventNoteOutlinedIcon sx={{ color: 'primary.main', fontSize: 26 }} />
          <Typography variant="subtitle1" fontWeight={700}>
            Notes &amp; TODOs
          </Typography>
        </Box>
        <IconButton size="small" onClick={onClose} aria-label="Close">
          <CloseIcon fontSize="small" />
        </IconButton>
      </Box>
      <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <NotesTodosContent variant="admin" open={open} />
      </Box>
    </Paper>
  );
}
