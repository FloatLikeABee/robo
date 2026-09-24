import React, { useState } from 'react';
import { Box, Tab, Tabs, TextField, Typography } from '@mui/material';
import VisualMarkdown from '../../lib/VisualMarkdown';

const markdownPreviewSx = {
  '& h1, & h2, & h3, & h4': { mt: 0.75, mb: 0.5, fontWeight: 700, lineHeight: 1.3 },
  '& h1': { fontSize: '1.1rem' },
  '& h2': { fontSize: '1rem' },
  '& h3': { fontSize: '0.925rem' },
  '& p': { my: 0.5 },
  '& ul, & ol': { my: 0.5, pl: 2.25 },
  '& li': { my: 0.25 },
  '& blockquote': {
    my: 0.75,
    pl: 1.5,
    borderLeft: 3,
    borderColor: 'divider',
    color: 'text.secondary',
  },
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
  '& a': { color: '#38bdf8' },
  '& table': { width: '100%', borderCollapse: 'collapse', my: 0.75 },
  '& th, & td': { border: 1, borderColor: 'divider', px: 0.75, py: 0.35, fontSize: '0.875rem' },
  '& hr': { my: 1, borderColor: 'divider' },
};

export default function MarkdownEditor({
  value,
  onChange,
  placeholder = 'Write in Markdown…',
  disabled = false,
  minRows = 10,
  hint = true,
}) {
  const [tab, setTab] = useState('markdown');
  const body = value || '';
  const rawLocked = disabled || typeof onChange !== 'function';

  return (
    <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
      <Tabs
        value={tab}
        onChange={(_, next) => setTab(next)}
        sx={{ minHeight: 36, flexShrink: 0, borderBottom: 1, borderColor: 'divider' }}
      >
        <Tab label="Markdown" value="markdown" sx={{ minHeight: 36, py: 0.5, textTransform: 'none' }} />
        <Tab label="Raw" value="raw" sx={{ minHeight: 36, py: 0.5, textTransform: 'none' }} />
      </Tabs>

      {tab === 'raw' ? (
        <TextField
          fullWidth
          multiline
          minRows={minRows}
          placeholder={placeholder}
          value={body}
          onChange={(e) => onChange?.(e.target.value)}
          disabled={rawLocked}
          InputProps={{ readOnly: rawLocked }}
          sx={{
            flex: 1,
            minHeight: 0,
            '& .MuiOutlinedInput-root': {
              height: '100%',
              alignItems: 'flex-start',
              py: 1,
              boxSizing: 'border-box',
            },
            '& textarea': {
              minHeight: `${minRows * 22}px !important`,
              height: '100% !important',
              maxHeight: '100%',
              overflow: 'auto !important',
              resize: 'none',
              boxSizing: 'border-box',
              fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
              fontSize: '0.8125rem',
              lineHeight: 1.5,
            },
          }}
        />
      ) : (
        <Box
          className="themed-preview-scroll"
          sx={{
            flex: 1,
            minHeight: 0,
            overflow: 'auto',
            px: 1.5,
            py: 1.25,
            border: 1,
            borderColor: 'divider',
            borderRadius: 1,
            bgcolor: 'background.paper',
            color: 'text.primary',
            ...markdownPreviewSx,
          }}
        >
          {body.trim() ? (
            <VisualMarkdown text={body} />
          ) : (
            <Typography variant="body2" color="text.secondary">
              Nothing to preview yet.
            </Typography>
          )}
        </Box>
      )}

      {hint ? (
        <Typography variant="caption" color="text.secondary" sx={{ flexShrink: 0 }}>
          Markdown: **bold**, lists, ## headings, [links](url)
        </Typography>
      ) : null}
    </Box>
  );
}
