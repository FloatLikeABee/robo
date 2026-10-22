import React, { useEffect, useMemo, useState } from 'react';
import {
  Box,
  Button,
  CircularProgress,
  IconButton,
  List,
  ListItem,
  ListItemText,
  Stack,
  Typography,
  Alert,
} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import InsertDriveFileOutlinedIcon from '@mui/icons-material/InsertDriveFileOutlined';
import PictureAsPdfOutlinedIcon from '@mui/icons-material/PictureAsPdfOutlined';
import { tranApi, tranEndpoints } from '../../api/tranClient';

const ACCEPT =
  '.pdf,.xlsx,.xls,.csv,.json,.txt,.jpg,.jpeg,.png,.gif,.webp,.bmp,.svg,.heic,.heif,.mp4,.webm,.mov,.avi,.mkv,.m4v';

function formatSize(bytes) {
  if (!bytes || bytes <= 0) return '';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function fileIcon(att) {
  const name = String(att?.original_name || '').toLowerCase();
  if (name.endsWith('.pdf')) return <PictureAsPdfOutlinedIcon fontSize="small" color="action" />;
  return <InsertDriveFileOutlinedIcon fontSize="small" color="action" />;
}

function MediaPreview({ entityRoute, recordId, attachment }) {
  const [blobUrl, setBlobUrl] = useState('');
  const [busy, setBusy] = useState(true);
  const [err, setErr] = useState('');

  useEffect(() => {
    let alive = true;
    let objectUrl = '';
    async function run() {
      setBusy(true);
      setErr('');
      try {
        const url = tranEndpoints.entityAttachmentDownload(entityRoute, recordId, attachment.id);
        const res = await tranApi.get(url, { responseType: 'blob' });
        if (!alive) return;
        objectUrl = URL.createObjectURL(res.data);
        setBlobUrl(objectUrl);
      } catch (e) {
        if (!alive) return;
        setErr(e.response?.data?.error || e.message || 'Preview unavailable');
      } finally {
        if (alive) setBusy(false);
      }
    }
    run();
    return () => {
      alive = false;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [entityRoute, recordId, attachment.id]);

  if (busy) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
        <CircularProgress size={22} />
      </Box>
    );
  }
  if (err) {
    return (
      <Typography variant="caption" color="text.secondary">
        {err}
      </Typography>
    );
  }
  if (attachment.kind === 'video') {
    return (
      <Box
        component="video"
        src={blobUrl}
        controls
        sx={{ width: '100%', maxHeight: 220, borderRadius: 1, bgcolor: 'black' }}
      />
    );
  }
  return (
    <Box
      component="img"
      src={blobUrl}
      alt={attachment.original_name}
      sx={{
        width: '100%',
        maxHeight: 220,
        objectFit: 'contain',
        borderRadius: 1,
        bgcolor: 'action.hover',
      }}
    />
  );
}

export default function RecordAttachmentsPanel({
  entityRoute,
  recordId,
  attachments: attachmentsProp = [],
  onChange,
  compact = false,
}) {
  const [attachments, setAttachments] = useState(attachmentsProp);
  const [maxPerRecord, setMaxPerRecord] = useState(10);
  const [pickedFiles, setPickedFiles] = useState([]);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    setAttachments(Array.isArray(attachmentsProp) ? attachmentsProp : []);
  }, [attachmentsProp, recordId]);

  useEffect(() => {
    let alive = true;
    tranApi
      .get(tranEndpoints.attachmentConfig)
      .then((res) => {
        if (!alive) return;
        const n = res.data?.max_per_record;
        if (typeof n === 'number' && n > 0) setMaxPerRecord(n);
      })
      .catch(() => {});
    return () => {
      alive = false;
    };
  }, []);

  const remaining = Math.max(0, maxPerRecord - attachments.length);

  const downloadFile = async (att) => {
    try {
      const url = tranEndpoints.entityAttachmentDownload(entityRoute, recordId, att.id);
      const res = await tranApi.get(url, { responseType: 'blob' });
      const blobUrl = URL.createObjectURL(res.data);
      const a = document.createElement('a');
      a.href = blobUrl;
      a.download = att.original_name || 'download';
      a.rel = 'noopener';
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(blobUrl);
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Download failed');
    }
  };

  const uploadFiles = async (files) => {
    if (!recordId || !files?.length) return;
    if (files.length > remaining) {
      setError(`Up to ${maxPerRecord} files per record (${remaining} slot${remaining === 1 ? '' : 's'} left).`);
      return;
    }
    setUploading(true);
    setError('');
    try {
      const form = new FormData();
      files.forEach((f) => form.append('files', f));
      const res = await tranApi.post(tranEndpoints.entityAttachments(entityRoute, recordId), form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      const next = res.data?.attachments || [];
      setAttachments(next);
      setPickedFiles([]);
      onChange?.(next);
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Upload failed');
    } finally {
      setUploading(false);
    }
  };

  const removeAttachment = async (attachmentId) => {
    setError('');
    try {
      await tranApi.delete(tranEndpoints.entityAttachment(entityRoute, recordId, attachmentId));
      const next = attachments.filter((x) => x.id !== attachmentId);
      setAttachments(next);
      onChange?.(next);
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Delete failed');
    }
  };

  const mediaAttachments = useMemo(
    () => attachments.filter((a) => a.kind === 'image' || a.kind === 'video'),
    [attachments]
  );
  const docAttachments = useMemo(
    () => attachments.filter((a) => a.kind !== 'image' && a.kind !== 'video'),
    [attachments]
  );

  if (!recordId) {
    return (
      <Typography variant="body2" color="text.secondary">
        Save the record before attaching files.
      </Typography>
    );
  }

  return (
    <Box>
      <Typography variant="subtitle2" fontWeight={700} sx={{ mb: 1 }}>
        Documents &amp; media
      </Typography>
      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1.25 }}>
        PDF, Excel, CSV, JSON, text, images, and video — up to {maxPerRecord} files per record.
      </Typography>

      {mediaAttachments.length > 0 ? (
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: compact ? '1fr' : 'repeat(auto-fill, minmax(180px, 1fr))',
            gap: 1.5,
            mb: 1.5,
          }}
        >
          {mediaAttachments.map((att) => (
            <Box
              key={att.id}
              sx={{
                border: '1px solid',
                borderColor: 'divider',
                borderRadius: 1,
                p: 1,
                bgcolor: 'background.paper',
              }}
            >
              <MediaPreview entityRoute={entityRoute} recordId={recordId} attachment={att} />
              <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mt: 0.75 }}>
                <Typography variant="caption" noWrap title={att.original_name} sx={{ flex: 1, minWidth: 0, pr: 0.5 }}>
                  {att.original_name}
                </Typography>
                <IconButton size="small" color="error" onClick={() => removeAttachment(att.id)} aria-label="Remove">
                  <DeleteOutlineIcon fontSize="small" />
                </IconButton>
              </Stack>
            </Box>
          ))}
        </Box>
      ) : null}

      {docAttachments.length > 0 ? (
        <List dense sx={{ mb: 1 }}>
          {docAttachments.map((a) => (
            <ListItem
              key={a.id}
              disablePadding
              secondaryAction={
                <Stack direction="row" spacing={0.5}>
                  <IconButton edge="end" onClick={() => downloadFile(a)} aria-label="Download">
                    <DownloadIcon fontSize="small" />
                  </IconButton>
                  <IconButton edge="end" color="error" onClick={() => removeAttachment(a.id)} aria-label="Remove">
                    <DeleteOutlineIcon fontSize="small" />
                  </IconButton>
                </Stack>
              }
            >
              <Stack direction="row" spacing={1} alignItems="center" sx={{ minWidth: 0, pr: 8 }}>
                {fileIcon(a)}
                <ListItemText
                  primary={a.original_name}
                  secondary={formatSize(a.size_bytes)}
                  primaryTypographyProps={{ noWrap: true }}
                />
              </Stack>
            </ListItem>
          ))}
        </List>
      ) : null}

      {attachments.length === 0 ? (
        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
          No files attached yet.
        </Typography>
      ) : null}

      <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
        <Button component="label" variant="outlined" size="small" disabled={uploading || remaining <= 0}>
          {uploading ? 'Uploading…' : 'Choose files'}
          <input
            type="file"
            hidden
            multiple
            accept={ACCEPT}
            onChange={(e) => {
              const files = Array.from(e.target.files || []);
              setPickedFiles(files);
              if (files.length) uploadFiles(files.slice(0, remaining));
              e.target.value = '';
            }}
          />
        </Button>
        {remaining <= 0 ? (
          <Typography variant="caption" color="text.secondary">
            Maximum attachments reached.
          </Typography>
        ) : null}
      </Stack>

      {pickedFiles.length > 0 && uploading ? (
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
          Uploading: {pickedFiles.map((f) => f.name).join(', ')}
        </Typography>
      ) : null}

      {error ? (
        <Alert severity="error" sx={{ mt: 1 }}>
          {error}
        </Alert>
      ) : null}
    </Box>
  );
}
