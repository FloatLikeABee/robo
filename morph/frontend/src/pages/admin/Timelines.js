import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  List,
  ListItemButton,
  ListItemText,
  Stack,
  Tab,
  Tabs,
  TextField,
  Tooltip,
  Typography,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import PublishOutlinedIcon from '@mui/icons-material/PublishOutlined';
import RefreshIcon from '@mui/icons-material/Refresh';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { useConfirm } from '../../components/ConfirmDialog';

const MAX_FILE_BYTES = 5 * 1024 * 1024;

function publicTimelineHref(item) {
  const path = String(item?.published_path || '').trim();
  if (path.startsWith('/api/tran/public/timelines/')) {
    return path;
  }
  return item?.published_url || path || '';
}

function formatWhen(v) {
  if (!v) return '';
  try {
    return new Date(v).toLocaleString();
  } catch {
    return String(v);
  }
}

function allowedFile(file) {
  if (!file?.name) return false;
  const lower = file.name.toLowerCase();
  return lower.endsWith('.txt') || lower.endsWith('.pdf') || lower.endsWith('.md') || lower.endsWith('.markdown');
}

export default function Timelines() {
  const { confirm } = useConfirm();
  const title = 'Timelines';

  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');
  const [selectedId, setSelectedId] = useState(null);
  const [selected, setSelected] = useState(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [previewTab, setPreviewTab] = useState(0);

  const [createOpen, setCreateOpen] = useState(false);
  const [createTitle, setCreateTitle] = useState('');
  const [paste, setPaste] = useState('');
  const [url, setUrl] = useState('');
  const [file, setFile] = useState(null);
  const [formWarning, setFormWarning] = useState('');
  const [creating, setCreating] = useState(false);

  const [publishing, setPublishing] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const loadList = useCallback(async () => {
    const res = await tranApi.get(tranEndpoints.timelines);
    const list = Array.isArray(res.data) ? res.data : [];
    setItems(list);
    return list;
  }, []);

  useEffect(() => {
    loadList()
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load timelines'))
      .finally(() => setLoading(false));
  }, [loadList]);

  const loadDetail = useCallback(async (id) => {
    if (id == null) {
      setSelected(null);
      return;
    }
    setDetailLoading(true);
    setError('');
    try {
      const res = await tranApi.get(tranEndpoints.timeline(id));
      setSelected(res.data || null);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to load timeline');
      setSelected(null);
    } finally {
      setDetailLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadDetail(selectedId);
  }, [selectedId, loadDetail]);

  const sorted = useMemo(
    () =>
      [...items].sort((a, b) => String(b.last_updated || '').localeCompare(String(a.last_updated || ''))),
    [items]
  );

  const resetCreateForm = () => {
    setCreateTitle('');
    setPaste('');
    setUrl('');
    setFile(null);
    setFormWarning('');
  };

  const applyCreated = (created) => {
    if (!created || created.id == null) return;
    setItems((prev) => {
      const rest = prev.filter((n) => n.id !== created.id);
      return [created, ...rest];
    });
    setSelected(created);
    setSelectedId(created.id);
    setPreviewTab(0);
  };

  const onCreate = async () => {
    const hasPaste = Boolean(paste.trim());
    const hasURL = Boolean(url.trim());
    const hasFile = Boolean(file);
    if (!hasPaste && !hasURL && !hasFile) {
      setFormWarning('Provide at least one source: upload a file, enter a URL, or paste content.');
      return;
    }
    if (file) {
      if (!allowedFile(file)) {
        setFormWarning('Only .txt, .pdf, and .md files are allowed.');
        return;
      }
      if (file.size > MAX_FILE_BYTES) {
        setFormWarning('File must be 5 MB or smaller.');
        return;
      }
    }
    if (creating) return;
    setCreating(true);
    setFormWarning('');
    setError('');
    setInfo('');
    try {
      const form = new FormData();
      if (createTitle.trim()) form.append('title', createTitle.trim());
      if (hasPaste) form.append('paste', paste.trim());
      if (hasURL) form.append('url', url.trim());
      if (hasFile) form.append('file', file);
      const res = await tranApi.post(tranEndpoints.timelines, form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      const created = res.data || {};
      if (created.id == null) {
        throw new Error('Create succeeded but no timeline id was returned');
      }
      setCreateOpen(false);
      resetCreateForm();
      setInfo('Timeline generated.');
      applyCreated(created);
      try {
        await loadList();
      } catch {
        /* keep created in local list */
      }
    } catch (err) {
      setFormWarning(err.response?.data?.error || err.message || 'Failed to generate timeline');
    } finally {
      setCreating(false);
    }
  };

  const onPublish = async () => {
    if (!selectedId || publishing) return;
    setPublishing(true);
    setError('');
    setInfo('');
    try {
      const res = await tranApi.post(tranEndpoints.timelinePublish(selectedId));
      setSelected(res.data || null);
      const href = publicTimelineHref(res.data || {});
      setInfo(
        href
          ? `Published: ${href.startsWith('http') ? href : `${window.location.origin}${href}`}`
          : 'Published.'
      );
      await loadList().catch(() => {});
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Publish failed');
    } finally {
      setPublishing(false);
    }
  };

  const onDelete = async () => {
    if (!selectedId || deleting) return;
    const ok = await confirm({
      title: 'Delete timeline',
      message: 'Delete this timeline? This cannot be undone.',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    setDeleting(true);
    setError('');
    setInfo('');
    try {
      await tranApi.delete(tranEndpoints.timeline(selectedId));
      setSelectedId(null);
      setSelected(null);
      setInfo('Timeline deleted.');
      await loadList();
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Delete failed');
    } finally {
      setDeleting(false);
    }
  };

  const pubHref = publicTimelineHref(selected || {});

  return (
    <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', gap: 1.5 }}>
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        alignItems={{ xs: 'stretch', sm: 'center' }}
        justifyContent="space-between"
        spacing={1}
      >
        <Box>
          <Typography variant="h6" sx={{ fontWeight: 700 }}>
            {title}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ display: { xs: 'none', sm: 'block' } }}>
            Turn a file, URL, or paste into a publishable timeline (markdown + HTML).
          </Typography>
        </Box>
        <Stack direction="row" spacing={1} sx={{ flexShrink: 0 }}>
          <Button startIcon={<RefreshIcon />} onClick={() => loadList().catch(() => {})} disabled={loading}>
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => {
              resetCreateForm();
              setCreateOpen(true);
            }}
          >
            New
          </Button>
        </Stack>
      </Stack>

      {error ? (
        <Alert severity="error" onClose={() => setError('')}>
          {error}
        </Alert>
      ) : null}
      {info ? (
        <Alert severity="success" onClose={() => setInfo('')}>
          {info}
        </Alert>
      ) : null}

      <Box
        sx={{
          flex: 1,
          minHeight: 0,
          display: 'flex',
          flexDirection: { xs: 'column', md: 'row' },
          gap: 1.5,
        }}
      >
        <Box
          sx={{
            width: { xs: '100%', md: 300 },
            maxHeight: { xs: selectedId ? 160 : '40%', md: 'none' },
            flexShrink: 0,
            border: 1,
            borderColor: 'divider',
            borderRadius: 1,
            overflow: 'auto',
            bgcolor: 'background.paper',
          }}
        >
          {loading ? (
            <Box sx={{ p: 2, display: 'flex', justifyContent: 'center' }}>
              <CircularProgress size={28} />
            </Box>
          ) : sorted.length === 0 ? (
            <Typography variant="body2" color="text.secondary" sx={{ p: 2 }}>
              No timelines yet. Create one from a document, URL, or paste. Prior Stories are not imported into
              Timelines.
            </Typography>
          ) : (
            <List dense disablePadding>
              {sorted.map((n) => (
                <ListItemButton
                  key={n.id}
                  selected={selectedId === n.id}
                  onClick={() => setSelectedId(n.id)}
                  alignItems="flex-start"
                >
                  <ListItemText
                    primary={n.title || `Timeline #${n.id}`}
                    secondary={
                      <>
                        {formatWhen(n.last_updated)}
                        {n.published_path ? ' · published' : ''}
                      </>
                    }
                    primaryTypographyProps={{ noWrap: true, fontWeight: 600 }}
                    secondaryTypographyProps={{ noWrap: true }}
                  />
                </ListItemButton>
              ))}
            </List>
          )}
        </Box>

        <Box
          sx={{
            flex: 1,
            minWidth: 0,
            minHeight: 0,
            border: 1,
            borderColor: 'divider',
            borderRadius: 1,
            display: 'flex',
            flexDirection: 'column',
            bgcolor: 'background.paper',
            overflow: 'hidden',
          }}
        >
          {!selectedId ? (
            <Box sx={{ p: 3 }}>
              <Typography color="text.secondary">Select a timeline or create a new one.</Typography>
            </Box>
          ) : detailLoading ? (
            <Box sx={{ p: 3, display: 'flex', justifyContent: 'center' }}>
              <CircularProgress size={32} />
            </Box>
          ) : !selected ? (
            <Box sx={{ p: 3 }}>
              <Typography color="text.secondary">Timeline not found.</Typography>
            </Box>
          ) : (
            <>
              <Stack
                direction="row"
                alignItems="flex-start"
                justifyContent="space-between"
                spacing={1}
                sx={{ p: 1.5, borderBottom: 1, borderColor: 'divider' }}
              >
                <Box sx={{ minWidth: 0 }}>
                  <Typography variant="subtitle1" sx={{ fontWeight: 700 }} noWrap>
                    {selected.title || `Timeline #${selected.id}`}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                    Sources: {selected.source_summary || '—'} · {formatWhen(selected.last_updated)}
                  </Typography>
                </Box>
                <Stack direction="row" spacing={0.5} sx={{ flexShrink: 0 }}>
                  {pubHref ? (
                    <Tooltip title="Open published page">
                      <IconButton
                        size="small"
                        component="a"
                        href={pubHref.startsWith('http') ? pubHref : pubHref}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <OpenInNewIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  ) : null}
                  <Tooltip title="Publish">
                    <span>
                      <IconButton size="small" onClick={onPublish} disabled={publishing}>
                        {publishing ? <CircularProgress size={18} /> : <PublishOutlinedIcon fontSize="small" />}
                      </IconButton>
                    </span>
                  </Tooltip>
                  <Tooltip title="Delete">
                    <span>
                      <IconButton size="small" color="error" onClick={onDelete} disabled={deleting}>
                        <DeleteOutlineIcon fontSize="small" />
                      </IconButton>
                    </span>
                  </Tooltip>
                </Stack>
              </Stack>

              <Tabs
                value={previewTab}
                onChange={(_, v) => setPreviewTab(v)}
                sx={{ borderBottom: 1, borderColor: 'divider', minHeight: 40 }}
              >
                <Tab label="Markdown" sx={{ minHeight: 40 }} />
                <Tab label="HTML" sx={{ minHeight: 40 }} />
              </Tabs>

              <Box sx={{ flex: 1, minHeight: 0, overflow: 'auto', p: 2 }}>
                {previewTab === 0 ? (
                  <Typography
                    component="pre"
                    sx={{
                      m: 0,
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word',
                      fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
                      fontSize: 13,
                    }}
                  >
                    {selected.markdown_content || ''}
                  </Typography>
                ) : (
                  <Box
                    sx={{
                      border: 1,
                      borderColor: 'divider',
                      borderRadius: 1,
                      overflow: 'hidden',
                      bgcolor: '#0b1220',
                      minHeight: 280,
                    }}
                  >
                    <iframe
                      title="Timeline HTML preview"
                      srcDoc={selected.html_content || '<p>No HTML</p>'}
                      sandbox=""
                      style={{ width: '100%', height: 480, border: 0, background: 'transparent' }}
                    />
                  </Box>
                )}
              </Box>
            </>
          )}
        </Box>
      </Box>

      <Dialog open={createOpen} onClose={() => !creating && setCreateOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>New timeline</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 0.5 }}>
            <Typography variant="body2" color="text.secondary">
              Provide at least one source. AI will turn it into a chronological timeline.
            </Typography>
            {formWarning ? <Alert severity="warning">{formWarning}</Alert> : null}
            <TextField
              label="Title (optional)"
              value={createTitle}
              onChange={(e) => setCreateTitle(e.target.value)}
              fullWidth
              size="small"
              disabled={creating}
            />
            <Button variant="outlined" component="label" disabled={creating}>
              {file ? `File: ${file.name}` : 'Upload .txt / .pdf / .md (max 5 MB)'}
              <input
                hidden
                type="file"
                accept=".txt,.pdf,.md,.markdown,text/plain,application/pdf,text/markdown"
                onChange={(e) => {
                  const f = e.target.files?.[0] || null;
                  setFile(f);
                  setFormWarning('');
                }}
              />
            </Button>
            <TextField
              label="URL (optional)"
              value={url}
              onChange={(e) => {
                setUrl(e.target.value);
                setFormWarning('');
              }}
              fullWidth
              size="small"
              placeholder="https://…"
              disabled={creating}
            />
            <TextField
              label="Paste content (optional)"
              value={paste}
              onChange={(e) => {
                setPaste(e.target.value);
                setFormWarning('');
              }}
              fullWidth
              multiline
              minRows={5}
              disabled={creating}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)} disabled={creating}>
            Cancel
          </Button>
          <Button variant="contained" onClick={onCreate} disabled={creating}>
            {creating ? <CircularProgress size={20} /> : 'Generate'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
