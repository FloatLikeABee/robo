import React, { useEffect, useState } from 'react';
import {
  Alert,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
} from '@mui/material';
import AutoAwesomeOutlinedIcon from '@mui/icons-material/AutoAwesomeOutlined';
import { tranApi, tranEndpoints } from '../../api/tranClient';

/**
 * Paste plain text, extract a JSON object via Morph AI, then apply (does not save the record).
 */
export default function ExtractJsonFromTextDialog({
  open,
  onClose,
  purpose,
  title = 'Extract JSON from text',
  applyLabel = 'Apply JSON',
  onApply,
}) {
  const [text, setText] = useState('');
  const [jsonText, setJsonText] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!open) {
      setText('');
      setJsonText('');
      setError(null);
      setLoading(false);
    }
  }, [open]);

  const extract = async () => {
    if (!text.trim()) {
      setError('Paste some text to extract');
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const res = await tranApi.post(tranEndpoints.extractJson, {
        text: text.trim(),
        purpose,
      });
      const obj = res.data?.json;
      if (obj == null || typeof obj !== 'object') {
        throw new Error('AI did not return JSON');
      }
      setJsonText(JSON.stringify(obj, null, 2));
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Extract failed');
    } finally {
      setLoading(false);
    }
  };

  const apply = async () => {
    if (!jsonText.trim()) {
      setError('Extract JSON first');
      return;
    }
    let parsed;
    try {
      parsed = JSON.parse(jsonText);
    } catch {
      setError('JSON is not valid — fix it before applying');
      return;
    }
    const result = await onApply(parsed, jsonText);
    if (result !== false) onClose();
  };

  return (
    <Dialog open={open} onClose={loading ? undefined : onClose} maxWidth="sm" fullWidth PaperProps={{ sx: { borderRadius: 2 } }}>
      <DialogTitle sx={{ fontWeight: 700 }}>{title}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 0.5 }}>
          <TextField
            label="Description"
            placeholder="Describe the data in plain language…"
            size="small"
            fullWidth
            multiline
            minRows={5}
            value={text}
            onChange={(e) => setText(e.target.value)}
          />
          <Button
            variant="contained"
            startIcon={loading ? <CircularProgress size={16} color="inherit" /> : <AutoAwesomeOutlinedIcon />}
            onClick={extract}
            disabled={loading}
            sx={{ textTransform: 'none', alignSelf: 'flex-start' }}
          >
            {loading ? 'Extracting…' : 'Extract JSON'}
          </Button>
          {jsonText ? (
            <TextField
              label="JSON (editable)"
              size="small"
              fullWidth
              multiline
              minRows={8}
              value={jsonText}
              onChange={(e) => setJsonText(e.target.value)}
              sx={{ '& .MuiInputBase-input': { fontFamily: 'ui-monospace, monospace', fontSize: 13 } }}
            />
          ) : null}
          {error && <Alert severity="error">{error}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} disabled={loading}>
          Cancel
        </Button>
        <Button variant="contained" onClick={apply} disabled={loading || !jsonText} sx={{ textTransform: 'none' }}>
          {applyLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
