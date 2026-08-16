import React, { useState } from 'react';
import { Box, Tab, Tabs, TextField, Typography } from '@mui/material';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

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
  '& a': { color: 'primary.main' },
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
}) {
  const [tab, setTab] = useState('write');
  const body = value || '';

  return (
    <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
      <Tabs
        value={tab}
        onChange={(_, next) => setTab(next)}
        sx={{ minHeight: 36, flexShrink: 0, borderBottom: 1, borderColor: 'divider' }}
      >
        <Tab label="Write" value="write" sx={{ minHeight: 36, py: 0.5, textTransform: 'none' }} />
        <Tab label="Preview" value="preview" sx={{ minHeight: 36, py: 0.5, textTransform: 'none' }} />
      </Tabs>

      {tab === 'write' ? (
        <TextField
          fullWidth
          multiline
          minRows={minRows}
          placeholder={placeholder}
          value={body}
          onChange={(e) => onChange?.(e.target.value)}
          disabled={disabled}
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
            ...markdownPreviewSx,
          }}
        >
          {body.trim() ? (
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{body}</ReactMarkdown>
          ) : (
            <Typography variant="body2" color="text.secondary">
              Nothing to preview yet.
            </Typography>
          )}
        </Box>
      )}

      <Typography variant="caption" color="text.secondary" sx={{ flexShrink: 0 }}>
        Markdown supported: **bold**, lists, ## headings, [links](url)
      </Typography>
    </Box>
  );
}
