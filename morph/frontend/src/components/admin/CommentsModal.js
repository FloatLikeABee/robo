import React, { useState, useEffect, useRef } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  IconButton,
  Box,
  Typography,
  TextField,
  Button,
  Stack,
} from '@mui/material';
import { useTheme } from '@mui/material/styles';
import CloseIcon from '@mui/icons-material/Close';
import ReplyOutlinedIcon from '@mui/icons-material/ReplyOutlined';
import DeleteOutlineOutlinedIcon from '@mui/icons-material/DeleteOutlineOutlined';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { useConfirm } from '../ConfirmDialog';

/** Strong label on primary actions (indigo fills need near-black contrast text). */
function primaryActionButtonSx(theme) {
  const dark = theme.palette.mode === 'dark';
  return {
    fontWeight: 700,
    px: 2.5,
    py: 0.75,
    minWidth: 96,
    boxShadow: 'none',
    letterSpacing: 0.02,
    color: dark ? '#020617' : '#0f172a',
    '&:hover': {
      boxShadow: 'none',
      color: dark ? '#020617' : '#0f172a',
    },
    '&.Mui-disabled': {
      color: dark ? 'rgba(2, 6, 23, 0.45)' : undefined,
    },
  };
}

export default function CommentsModal({
  open,
  onClose,
  entityType,
  recordId,
  recordLabel,
  onThreadChanged,
  noun = 'comment',
}) {
  const nounPlural = noun === 'note' ? 'notes' : `${noun}s`;
  const nounTitle = noun === 'note' ? 'Notes' : 'Comments';
  const { confirm, alert } = useConfirm();
  const theme = useTheme();
  const [list, setList] = useState([]);
  const [loading, setLoading] = useState(false);
  const [newBody, setNewBody] = useState('');
  const [replyToId, setReplyToId] = useState(null);
  const [replyBody, setReplyBody] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const onThreadChangedRef = useRef(onThreadChanged);

  useEffect(() => {
    onThreadChangedRef.current = onThreadChanged;
  }, [onThreadChanged]);

  useEffect(() => {
    if (open && entityType && recordId) {
      setLoading(true);
      tranApi.get(tranEndpoints.comments, { params: { entity_type: entityType, record_id: recordId } })
        .then((res) => {
          const next = Array.isArray(res.data) ? res.data : [];
          setList(next);
          onThreadChangedRef.current?.(next);
        })
        .catch(() => {
          setList([]);
          onThreadChangedRef.current?.([]);
        })
        .finally(() => setLoading(false));
    } else {
      setList([]);
    }
    setNewBody('');
    setReplyToId(null);
    setReplyBody('');
  }, [open, entityType, recordId]);

  const refresh = () => {
    if (!entityType || !recordId) return;
    tranApi.get(tranEndpoints.comments, { params: { entity_type: entityType, record_id: recordId } })
      .then((res) => {
        const next = Array.isArray(res.data) ? res.data : [];
        setList(next);
        onThreadChangedRef.current?.(next);
      })
      .catch(() => {
        setList([]);
        onThreadChangedRef.current?.([]);
      });
  };

  const handleAdd = async () => {
    const body = newBody.trim();
    if (!body) return;
    setSubmitting(true);
    try {
      await tranApi.post(tranEndpoints.comments, { entity_type: entityType, record_id: recordId, body });
      setNewBody('');
      refresh();
    } catch (e) {
      await alert({ title: 'Error', message: e?.response?.data?.error || `Failed to add ${noun}` });
    } finally {
      setSubmitting(false);
    }
  };

  const handleReply = async () => {
    const body = replyBody.trim();
    if (!body || !replyToId) return;
    setSubmitting(true);
    try {
      await tranApi.post(tranEndpoints.comments, { entity_type: entityType, record_id: recordId, parent_id: replyToId, body });
      setReplyToId(null);
      setReplyBody('');
      refresh();
    } catch (e) {
      await alert({ title: 'Error', message: e?.response?.data?.error || 'Failed to add reply' });
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id) => {
    const ok = await confirm({
      title: `Delete ${noun}`,
      message: `Delete this ${noun}?`,
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    try {
      await tranApi.delete(tranEndpoints.comment(id));
      refresh();
    } catch (e) {
      await alert({ title: 'Error', message: e?.response?.data?.error || 'Delete failed' });
    }
  };

  const renderComment = (c, depth = 0) => (
    <Box
      key={c.id}
      sx={{
        pl: depth * 2,
        py: 0.75,
        borderLeft: depth > 0 ? 2 : 0,
        borderColor: 'primary.main',
        ml: depth > 0 ? 1 : 0,
      }}
    >
      <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 0.5 }}>
        <Box sx={{ flex: 1, minWidth: 0 }}>
          <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>{c.body}</Typography>
          <Typography variant="caption" color="text.secondary">
            {c.created_on ? new Date(c.created_on).toLocaleString() : ''}
            {c.author_user_id ? ` · User #${c.author_user_id}` : ''}
          </Typography>
        </Box>
        <IconButton size="small" onClick={() => setReplyToId(replyToId === c.id ? null : c.id)} title="Reply">
          <ReplyOutlinedIcon fontSize="small" />
        </IconButton>
        <IconButton size="small" onClick={() => handleDelete(c.id)} title="Delete">
          <DeleteOutlineOutlinedIcon fontSize="small" />
        </IconButton>
      </Box>
      {replyToId === c.id && (
        <Stack spacing={1} sx={{ mt: 1, pl: 0.5 }}>
          <TextField
            size="small"
            fullWidth
            placeholder="Write a reply..."
            value={replyBody}
            onChange={(e) => setReplyBody(e.target.value)}
            multiline
            minRows={2}
            maxRows={5}
          />
          <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1, flexWrap: 'wrap' }}>
            <Button size="small" onClick={() => { setReplyToId(null); setReplyBody(''); }}>
              Cancel
            </Button>
            <Button
              size="small"
              variant="contained"
              onClick={handleReply}
              disabled={submitting}
              sx={primaryActionButtonSx(theme)}
            >
              Post
            </Button>
          </Box>
        </Stack>
      )}
      {(c.replies || []).map((r) => renderComment(r, depth + 1))}
    </Box>
  );

  const title = recordLabel ? `${nounTitle} – ${recordLabel}` : `${nounTitle} (${entityType} #${recordId})`;

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          maxHeight: '85vh',
          display: 'flex',
          flexDirection: 'column',
        },
      }}
    >
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderBottom: 1, borderColor: 'divider' }}>
        <Typography variant="h6">{title}</Typography>
        <IconButton size="small" onClick={onClose} aria-label={`Close ${nounPlural}`}>
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent
        dividers
        sx={{
          display: 'flex',
          flexDirection: 'column',
          gap: 0,
          p: 0,
          overflow: 'hidden',
          flex: '1 1 auto',
          minHeight: 0,
          pt: 0,
        }}
      >
        <Box
          sx={{
            flex: '1 1 auto',
            overflow: 'auto',
            px: 2,
            py: 2,
            minHeight: 100,
          }}
        >
          {loading && (
            <Typography color="text.secondary" variant="body2">
              Loading…
            </Typography>
          )}
          {!loading && list.length === 0 && (
            <Typography color="text.secondary" variant="body2">
              No {nounPlural} yet.
            </Typography>
          )}
          {!loading && list.length > 0 && (
            <Stack spacing={0}>{list.map((c) => renderComment(c))}</Stack>
          )}
        </Box>
        <Box
          sx={{
            flexShrink: 0,
            borderTop: 1,
            borderColor: 'divider',
            px: 2,
            py: 1.5,
            bgcolor: (t) => (t.palette.mode === 'dark' ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.02)'),
          }}
        >
          <Stack spacing={1}>
            <TextField
              fullWidth
              size="small"
              placeholder={`Add a ${noun}...`}
              value={newBody}
              onChange={(e) => setNewBody(e.target.value)}
              multiline
              minRows={2}
              maxRows={4}
            />
            <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
              <Button
                variant="contained"
                color="primary"
                onClick={handleAdd}
                disabled={submitting || !newBody.trim()}
                sx={primaryActionButtonSx(theme)}
              >
                Add
              </Button>
            </Box>
          </Stack>
        </Box>
      </DialogContent>
    </Dialog>
  );
}
