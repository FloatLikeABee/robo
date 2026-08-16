import React, { useMemo, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import {
  Box,
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Typography,
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';

const WORD_LIMIT = 200;

function countWords(text) {
  const trimmed = String(text || '').trim();
  if (!trimmed) return 0;
  return trimmed.split(/\s+/).length;
}

function truncateWords(text, maxWords) {
  const words = String(text || '').trim().split(/\s+/).filter(Boolean);
  if (words.length <= maxWords) return words.join(' ');
  return words.slice(0, maxWords).join(' ');
}

const markdownSx = {
  '& h1, & h2, & h3': { mt: 0.5, mb: 0.75, fontWeight: 700, lineHeight: 1.3 },
  '& h2': { fontSize: '0.95rem' },
  '& h3': { fontSize: '0.875rem' },
  '& p': { my: 0.5 },
  '& ul, & ol': { my: 0.5, pl: 2.25 },
  '& li': { my: 0.25 },
  '& code': {
    fontFamily: 'ui-monospace, monospace',
    fontSize: '0.92em',
    px: 0.4,
    py: 0.1,
    borderRadius: 0.5,
    bgcolor: 'action.hover',
  },
  '& pre': {
    my: 0.75,
    p: 1,
    borderRadius: 1,
    overflow: 'auto',
    bgcolor: 'action.hover',
  },
  '& pre code': { bgcolor: 'transparent', p: 0 },
};

export const PLATFORM_AI_SENDER_ID = '00000000-0000-4000-8000-00000000a001';

export function resolveSenderLabel(senderId, users = []) {
  if (String(senderId) === PLATFORM_AI_SENDER_ID) return 'Message AI';
  const u = users.find((x) => String(x.id) === String(senderId));
  return u?.username || u?.email || String(senderId);
}

export default function MessageBody({ body, modalTitle = 'Message' }) {
  const [open, setOpen] = useState(false);
  const words = countWords(body);
  const isLong = words > WORD_LIMIT;
  const preview = isLong ? `${truncateWords(body, WORD_LIMIT)}…` : body;

  const previewNode = useMemo(
    () => (
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{preview}</ReactMarkdown>
    ),
    [preview]
  );

  const fullNode = useMemo(
    () => <ReactMarkdown remarkPlugins={[remarkGfm]}>{body}</ReactMarkdown>,
    [body]
  );

  return (
    <>
      <Box sx={markdownSx}>{previewNode}</Box>
      {isLong ? (
        <Button
          size="small"
          variant="text"
          sx={{ mt: 0.5, px: 0, minWidth: 0, textTransform: 'none', fontSize: '0.75rem' }}
          onClick={() => setOpen(true)}
        >
          Read full message ({words} words)
        </Button>
      ) : null}
      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', pr: 1 }}>
          <Typography variant="subtitle1" fontWeight={700}>{modalTitle}</Typography>
          <IconButton size="small" aria-label="Close" onClick={() => setOpen(false)}>
            <CloseIcon fontSize="small" />
          </IconButton>
        </DialogTitle>
        <DialogContent dividers>
          <Box sx={markdownSx}>{fullNode}</Box>
        </DialogContent>
      </Dialog>
    </>
  );
}
