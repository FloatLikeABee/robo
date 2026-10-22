import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  LinearProgress,
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
import SaveOutlinedIcon from '@mui/icons-material/SaveOutlined';
import StopCircleOutlinedIcon from '@mui/icons-material/StopCircleOutlined';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { useConfirm } from '../../components/ConfirmDialog';
import MarkdownEditor from '../../components/admin/MarkdownEditor';
import { darkPreviewIframeSx, withDarkPreviewSrcDoc } from '../../lib/darkPreviewSrcDoc';

const MAX_FILE_BYTES = 8 * 1024 * 1024;
const BUSY = new Set(['ingesting', 'running', 'refining']);

function publicResearchHref(item) {
  const path = String(item?.published_path || '').trim();
  if (path.startsWith('/api/tran/public/research/')) {
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
  return lower.endsWith('.txt') || lower.endsWith('.pdf') || lower.endsWith('.csv') || lower.endsWith('.json');
}

function isBusy(status) {
  return BUSY.has(String(status || ''));
}

export default function Research() {
  const { confirm } = useConfirm();
  const title = 'Research';

  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');
  const [selectedId, setSelectedId] = useState(null);
  const [selected, setSelected] = useState(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [previewTab, setPreviewTab] = useState(0);
  const [draftMd, setDraftMd] = useState('');
  const [savingMd, setSavingMd] = useState(false);

  const [createOpen, setCreateOpen] = useState(false);
  const [prompt, setPrompt] = useState('');
  const [files, setFiles] = useState([]);
  const [formWarning, setFormWarning] = useState('');
  const [creating, setCreating] = useState(false);

  const [publishing, setPublishing] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [cancelling, setCancelling] = useState(false);

  const loadList = useCallback(async () => {
    const res = await tranApi.get(tranEndpoints.research);
    const list = Array.isArray(res.data) ? res.data : [];
    setItems(list);
    return list;
  }, []);

  useEffect(() => {
    loadList()
      .then((list) => {
        setSelectedId((cur) => {
          if (cur != null) return cur;
          return list[0]?.id ?? null;
        });
      })
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load research'))
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
      const res = await tranApi.get(tranEndpoints.researchItem(id));
      setSelected(res.data || null);
      setDraftMd(res.data?.markdown_content || '');
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to load research');
      setSelected(null);
    } finally {
      setDetailLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadDetail(selectedId);
  }, [selectedId, loadDetail]);

  const busy = isBusy(selected?.status);

  useEffect(() => {
    if (!selectedId || !busy) return undefined;
    const tick = async () => {
      try {
        const res = await tranApi.get(tranEndpoints.researchItem(selectedId));
        setSelected(res.data || null);
        setDraftMd(res.data?.markdown_content || '');
      } catch {
        /* keep last snapshot */
      }
      try {
        await loadList();
      } catch {
        /* ignore */
      }
    };
    const t = setInterval(tick, 2000);
    return () => clearInterval(t);
  }, [selectedId, busy, loadList]);

  const sorted = useMemo(() => items, [items]);
  const pieces = Array.isArray(selected?.pieces) ? selected.pieces : [];
  const roundCount = selected?.round_count || 5;
  const currentRound = selected?.current_round || 0;

  const resetCreateForm = () => {
    setPrompt('');
    setFiles([]);
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
    setDraftMd(created.markdown_content || '');
    setPreviewTab(2);
  };

  const onCreate = async () => {
    if (!prompt.trim()) {
      setFormWarning('Prompt is required.');
      return;
    }
    for (const file of files) {
      if (!allowedFile(file)) {
        setFormWarning('Only .txt, .pdf, .csv, and .json files are allowed.');
        return;
      }
      if (file.size > MAX_FILE_BYTES) {
        setFormWarning('Each file must be 8 MB or smaller.');
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
      form.append('prompt', prompt.trim());
      files.forEach((f) => form.append('file', f));
      const res = await tranApi.post(tranEndpoints.research, form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      const created = res.data || {};
      if (created.id == null) {
        throw new Error('Create succeeded but no research id was returned');
      }
      setCreateOpen(false);
      resetCreateForm();
      setInfo('Research started.');
      applyCreated(created);
      try {
        await loadList();
      } catch {
        /* keep created in local list */
      }
    } catch (err) {
      setFormWarning(err.response?.data?.error || err.message || 'Failed to start research');
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
      const res = await tranApi.post(tranEndpoints.researchPublish(selectedId));
      setSelected(res.data || null);
      const href = publicResearchHref(res.data || {});
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

  const onCancel = async () => {
    if (!selectedId || cancelling) return;
    setCancelling(true);
    setError('');
    setInfo('');
    try {
      const res = await tranApi.post(tranEndpoints.researchCancel(selectedId));
      setSelected(res.data || null);
      setInfo('Research cancelled.');
      await loadList().catch(() => {});
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Cancel failed');
    } finally {
      setCancelling(false);
    }
  };

  const onDelete = async () => {
    if (!selectedId || deleting) return;
    const ok = await confirm({
      title: 'Delete research',
      message: 'Delete this research job? This cannot be undone.',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    setDeleting(true);
    setError('');
    setInfo('');
    try {
      await tranApi.delete(tranEndpoints.researchItem(selectedId));
      setInfo('Research deleted.');
      const list = await loadList();
      setSelectedId(list[0]?.id ?? null);
      if (!list[0]) {
        setSelected(null);
        setDraftMd('');
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Delete failed');
    } finally {
      setDeleting(false);
    }
  };

  const onSaveMarkdown = async () => {
    if (!selectedId || savingMd) return;
    setSavingMd(true);
    setError('');
    setInfo('');
    try {
      const res = await tranApi.patch(tranEndpoints.researchItem(selectedId), { markdown_content: draftMd });
      setSelected(res.data || null);
      setDraftMd(res.data?.markdown_content ?? draftMd);
      setInfo('Markdown saved.');
      await loadList().catch(() => {});
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Save failed');
    } finally {
      setSavingMd(false);
    }
  };

  const pubHref = publicResearchHref(selected || {});

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
            Prompt plus optional files; five online rounds with verification, then a publishable conclusion.
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
              No research yet. Start with a prompt and optional .txt / .pdf / .csv / .json files.
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
                    primary={n.title || `Research #${n.id}`}
                    secondary={
                      <>
                        {n.status || ''}
                        {n.current_round ? ` · ${n.current_round}/${n.round_count || 5}` : ''}
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
              <Typography color="text.secondary">Select a research job or create a new one.</Typography>
            </Box>
          ) : detailLoading && !selected ? (
            <Box sx={{ p: 3, display: 'flex', justifyContent: 'center' }}>
              <CircularProgress size={32} />
            </Box>
          ) : !selected ? (
            <Box sx={{ p: 3 }}>
              <Typography color="text.secondary">Research not found.</Typography>
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
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Typography variant="subtitle1" sx={{ fontWeight: 700 }} noWrap>
                      {selected.title || `Research #${selected.id}`}
                    </Typography>
                    <Chip size="small" label={selected.status || '—'} />
                  </Stack>
                  <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                    Round {currentRound}/{roundCount} · {formatWhen(selected.last_updated)}
                  </Typography>
                  {busy ? (
                    <LinearProgress
                      variant="determinate"
                      value={Math.min(100, (currentRound / roundCount) * 100)}
                      sx={{ mt: 1, width: 220 }}
                    />
                  ) : null}
                </Box>
                <Stack direction="row" spacing={0.5} sx={{ flexShrink: 0 }}>
                  {pubHref ? (
                    <Tooltip title="Open published page">
                      <IconButton
                        size="small"
                        component="a"
                        href={pubHref}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <OpenInNewIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  ) : null}
                  {busy ? (
                    <Tooltip title="Cancel">
                      <span>
                        <IconButton size="small" aria-label="Cancel" onClick={onCancel} disabled={cancelling}>
                          {cancelling ? <CircularProgress size={18} /> : <StopCircleOutlinedIcon fontSize="small" />}
                        </IconButton>
                      </span>
                    </Tooltip>
                  ) : (
                    <Tooltip title="Save markdown">
                      <span>
                        <IconButton
                          size="small"
                          aria-label="Save markdown"
                          onClick={onSaveMarkdown}
                          disabled={savingMd || draftMd === (selected.markdown_content || '')}
                        >
                          {savingMd ? <CircularProgress size={18} /> : <SaveOutlinedIcon fontSize="small" />}
                        </IconButton>
                      </span>
                    </Tooltip>
                  )}
                  <Tooltip title="Publish">
                    <span>
                      <IconButton size="small" aria-label="Publish" onClick={onPublish} disabled={publishing || busy}>
                        {publishing ? <CircularProgress size={18} /> : <PublishOutlinedIcon fontSize="small" />}
                      </IconButton>
                    </span>
                  </Tooltip>
                  <Tooltip title="Delete">
                    <span>
                      <IconButton size="small" color="error" aria-label="Delete" onClick={onDelete} disabled={deleting}>
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
                <Tab label="Rounds" sx={{ minHeight: 40 }} />
              </Tabs>

              <Box
                sx={{
                  flex: 1,
                  minHeight: 0,
                  overflow: 'hidden',
                  display: 'flex',
                  flexDirection: 'column',
                  p: previewTab === 1 ? 0 : 1.5,
                }}
              >
                {previewTab === 0 ? (
                  <MarkdownEditor
                    value={draftMd}
                    onChange={busy ? undefined : setDraftMd}
                    minRows={14}
                    hint={false}
                  />
                ) : null}
                {previewTab === 1 ? (
                  <Box sx={{ flex: 1, minHeight: 0, overflow: 'hidden', bgcolor: '#0b1220' }}>
                    <Box
                      component="iframe"
                      title="Research HTML preview"
                      srcDoc={withDarkPreviewSrcDoc(selected.html_content || '<p>No HTML yet</p>')}
                      sandbox=""
                      sx={darkPreviewIframeSx}
                    />
                  </Box>
                ) : null}
                {previewTab === 2 ? (
                  <Box sx={{ flex: 1, minHeight: 0, overflow: 'auto' }}>
                    {pieces.length === 0 ? (
                      <Typography color="text.secondary">
                        {busy ? 'Waiting for the first round…' : 'No round pieces yet.'}
                      </Typography>
                    ) : (
                      <Stack spacing={1.5}>
                        {pieces.map((p) => (
                          <Box key={p.round_index} sx={{ border: 1, borderColor: 'divider', borderRadius: 1, p: 1.25 }}>
                            <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                              Round {p.round_index} · {p.status || 'ok'}
                            </Typography>
                            <MarkdownEditor value={p.markdown || ''} minRows={4} hint={false} />
                            {p.verification ? (
                              <Typography variant="body2" color="text.secondary" sx={{ mt: 1, whiteSpace: 'pre-wrap' }}>
                                {p.verification}
                              </Typography>
                            ) : null}
                          </Box>
                        ))}
                      </Stack>
                    )}
                  </Box>
                ) : null}
              </Box>
            </>
          )}
        </Box>
      </Box>

      <Dialog open={createOpen} onClose={() => !creating && setCreateOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>New research</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 0.5 }}>
            <Typography variant="body2" color="text.secondary">
              Prompt is required. Optional files are ingested before five online rounds.
            </Typography>
            {formWarning ? <Alert severity="warning">{formWarning}</Alert> : null}
            <TextField
              label="Prompt"
              value={prompt}
              onChange={(e) => {
                setPrompt(e.target.value);
                setFormWarning('');
              }}
              fullWidth
              multiline
              minRows={4}
              required
              disabled={creating}
            />
            <Button variant="outlined" component="label" disabled={creating}>
              {files.length ? `${files.length} file(s) selected` : 'Upload .txt / .pdf / .csv / .json (optional)'}
              <input
                hidden
                type="file"
                multiple
                accept=".txt,.pdf,.csv,.json,text/plain,application/pdf,text/csv,application/json"
                onChange={(e) => {
                  setFiles(Array.from(e.target.files || []));
                  setFormWarning('');
                }}
              />
            </Button>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)} disabled={creating}>
            Cancel
          </Button>
          <Button variant="contained" onClick={onCreate} disabled={creating || !prompt.trim()}>
            {creating ? <CircularProgress size={20} /> : 'Start'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
