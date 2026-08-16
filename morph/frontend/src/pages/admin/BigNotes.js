import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Checkbox,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  FormControl,
  FormControlLabel,
  FormGroup,
  FormLabel,
  IconButton,
  List,
  ListItemButton,
  ListItemText,
  MenuItem,
  Radio,
  RadioGroup,
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
import AutoAwesomeIcon from '@mui/icons-material/AutoAwesome';
import RefreshIcon from '@mui/icons-material/Refresh';
import AnalyticsOutlinedIcon from '@mui/icons-material/AnalyticsOutlined';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { usePlatformUi } from '../../PlatformUiContext';
import { useConfirm } from '../../components/ConfirmDialog';

function publicNoteHref(note) {
  const path = String(note?.published_path || '').trim();
  if (path.startsWith('/api/tran/public/big-notes/')) {
    return path; // same-origin (Morph UI proxy or API)
  }
  return note?.published_url || path || '';
}

function formatWhen(v) {
  if (!v) return '';
  try {
    return new Date(v).toLocaleString();
  } catch {
    return String(v);
  }
}

function parseQuestions(raw) {
  if (!raw) return [];
  if (Array.isArray(raw)) return raw;
  if (typeof raw === 'string') {
    try {
      const parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  }
  return [];
}

export default function BigNotes() {
  const { labels } = usePlatformUi();
  const { confirm } = useConfirm();
  const title = labels.nav_big_notes || 'Big notes';

  const [notes, setNotes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');
  const [selectedId, setSelectedId] = useState(null);
  const [selected, setSelected] = useState(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [previewTab, setPreviewTab] = useState(0);

  const [createOpen, setCreateOpen] = useState(false);
  const [idea, setIdea] = useState('');
  const [theme, setTheme] = useState('dark');
  const [noteKind, setNoteKind] = useState('note');
  const [creating, setCreating] = useState(false);

  const [regenPrompt, setRegenPrompt] = useState('');
  const [regenerating, setRegenerating] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const [answers, setAnswers] = useState({});
  const [responses, setResponses] = useState([]);
  const [savingResponse, setSavingResponse] = useState(false);
  const [analyzingId, setAnalyzingId] = useState(null);
  const [analyzingAll, setAnalyzingAll] = useState(false);
  const [allAnalysis, setAllAnalysis] = useState('');

  const questions = useMemo(() => parseQuestions(selected?.questions), [selected]);

  const loadList = useCallback(async () => {
    const res = await tranApi.get(tranEndpoints.bigNotes);
    const list = Array.isArray(res.data) ? res.data : [];
    setNotes(list);
    return list;
  }, []);

  useEffect(() => {
    loadList()
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load notes'))
      .finally(() => setLoading(false));
  }, [loadList]);

  const loadResponses = useCallback(async (id) => {
    if (id == null) {
      setResponses([]);
      return;
    }
    try {
      const res = await tranApi.get(tranEndpoints.bigNoteResponses(id));
      setResponses(Array.isArray(res.data) ? res.data : []);
    } catch {
      setResponses([]);
    }
  }, []);

  const loadDetail = useCallback(
    async (id) => {
      if (id == null) {
        setSelected(null);
        setAnswers({});
        setAllAnalysis('');
        return;
      }
      setDetailLoading(true);
      setError('');
      try {
        const res = await tranApi.get(tranEndpoints.bigNote(id));
        const note = res.data || null;
        setSelected(note);
        const qs = parseQuestions(note?.questions);
        const next = {};
        qs.forEach((q) => {
          next[q.id] = q.type === 'multi' ? [] : '';
        });
        setAnswers(next);
        setAllAnalysis('');
        if (note?.note_kind === 'questionnaire') {
          await loadResponses(id);
        } else {
          setResponses([]);
        }
      } catch (err) {
        setError(err.response?.data?.error || err.message || 'Failed to load note');
        setSelected(null);
      } finally {
        setDetailLoading(false);
      }
    },
    [loadResponses]
  );

  useEffect(() => {
    void loadDetail(selectedId);
  }, [selectedId, loadDetail]);

  const sortedNotes = useMemo(
    () =>
      [...notes].sort((a, b) => String(b.last_updated || '').localeCompare(String(a.last_updated || ''))),
    [notes]
  );

  const applyCreatedNote = (created) => {
    if (!created || created.id == null) return;
    setNotes((prev) => {
      const rest = prev.filter((n) => n.id !== created.id);
      return [created, ...rest];
    });
    setSelected(created);
    setSelectedId(created.id);
    setPreviewTab(0);
  };

  const onCreate = async () => {
    const text = idea.trim();
    if (!text || creating) return;
    setCreating(true);
    setError('');
    setInfo('');
    try {
      const res = await tranApi.post(tranEndpoints.bigNotes, {
        idea: text,
        theme: theme === 'light' ? 'light' : 'dark',
        note_kind: noteKind,
      });
      const created = res.data || {};
      if (created.id == null) {
        throw new Error('Create succeeded but no note id was returned');
      }
      setCreateOpen(false);
      setIdea('');
      setInfo(noteKind === 'questionnaire' ? 'Questionnaire generated.' : 'Note generated.');
      applyCreatedNote(created);
      try {
        await loadList();
      } catch {
        /* keep created note in local list */
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to generate note');
    } finally {
      setCreating(false);
    }
  };

  const onRegenerate = async () => {
    if (!selectedId || !regenPrompt.trim() || regenerating) return;
    setRegenerating(true);
    setError('');
    setInfo('');
    try {
      const res = await tranApi.post(tranEndpoints.bigNoteRegenerate(selectedId), {
        prompt: regenPrompt.trim(),
        note_kind: selected?.note_kind,
      });
      const updated = res.data || null;
      setSelected(updated);
      setRegenPrompt('');
      setInfo('Note regenerated.');
      if (updated?.id != null) {
        setNotes((prev) => prev.map((n) => (n.id === updated.id ? { ...n, ...updated } : n)));
      }
      await loadList().catch(() => {});
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to regenerate');
    } finally {
      setRegenerating(false);
    }
  };

  const onPublish = async () => {
    if (!selectedId || publishing) return;
    setPublishing(true);
    setError('');
    setInfo('');
    try {
      const res = await tranApi.post(tranEndpoints.bigNotePublish(selectedId));
      setSelected(res.data || null);
      const href = publicNoteHref(res.data || {});
      setInfo(href ? `Published: ${href.startsWith('http') ? href : `${window.location.origin}${href}`}` : 'Published.');
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
      title: 'Delete Big note',
      message: 'Delete this Big note? This cannot be undone.',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    setDeleting(true);
    setError('');
    setInfo('');
    try {
      await tranApi.delete(tranEndpoints.bigNote(selectedId));
      setSelectedId(null);
      setSelected(null);
      setInfo('Note deleted.');
      await loadList();
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Delete failed');
    } finally {
      setDeleting(false);
    }
  };

  const onSubmitAnswers = async () => {
    if (!selectedId || savingResponse) return;
    setSavingResponse(true);
    setError('');
    setInfo('');
    try {
      await tranApi.post(tranEndpoints.bigNoteResponses(selectedId), { answers });
      setInfo('Response saved.');
      await loadResponses(selectedId);
      // reset form
      const next = {};
      questions.forEach((q) => {
        next[q.id] = q.type === 'multi' ? [] : '';
      });
      setAnswers(next);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Failed to save response');
    } finally {
      setSavingResponse(false);
    }
  };

  const onAnalyzeResponse = async (responseId) => {
    if (!selectedId || analyzingId) return;
    setAnalyzingId(responseId);
    setError('');
    try {
      await tranApi.post(tranEndpoints.bigNoteResponseAnalyze(selectedId, responseId));
      setInfo('Analysis ready.');
      await loadResponses(selectedId);
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Analysis failed');
    } finally {
      setAnalyzingId(null);
    }
  };

  const onAnalyzeAll = async () => {
    if (!selectedId || analyzingAll) return;
    setAnalyzingAll(true);
    setError('');
    try {
      const res = await tranApi.post(tranEndpoints.bigNoteAnalyzeAll(selectedId));
      setAllAnalysis(res.data?.analysis_markdown || '');
      setInfo('Overall analysis ready.');
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Overall analysis failed');
    } finally {
      setAnalyzingAll(false);
    }
  };

  const isQuestionnaire = selected?.note_kind === 'questionnaire';

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
            Notes or questionnaires — AI writes markdown/HTML; fill answers and analyze.
          </Typography>
        </Box>
        <Stack direction="row" spacing={1} sx={{ flexShrink: 0 }}>
          <Button startIcon={<RefreshIcon />} onClick={() => loadList().catch(() => {})} disabled={loading}>
            Refresh
          </Button>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => setCreateOpen(true)}>
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
          ) : sortedNotes.length === 0 ? (
            <Typography variant="body2" color="text.secondary" sx={{ p: 2 }}>
              No notes yet. Create a note or questionnaire from an idea.
            </Typography>
          ) : (
            <List dense disablePadding>
              {sortedNotes.map((n) => (
                <ListItemButton
                  key={n.id}
                  selected={selectedId === n.id}
                  onClick={() => setSelectedId(n.id)}
                  alignItems="flex-start"
                >
                  <ListItemText
                    primary={n.title || `Note #${n.id}`}
                    secondary={
                      <>
                        {n.note_kind === 'questionnaire' ? 'Questionnaire · ' : 'Note · '}
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
              <Typography color="text.secondary">Select a note or create a new one.</Typography>
            </Box>
          ) : detailLoading || !selected ? (
            <Box sx={{ p: 3, display: 'flex', justifyContent: 'center' }}>
              <CircularProgress size={28} />
            </Box>
          ) : (
            <>
              <Stack
                direction={{ xs: 'column', sm: 'row' }}
                alignItems={{ sm: 'center' }}
                justifyContent="space-between"
                spacing={1}
                sx={{ px: 2, py: 1.5, borderBottom: 1, borderColor: 'divider' }}
              >
                <Box sx={{ minWidth: 0 }}>
                  <Typography variant="subtitle1" sx={{ fontWeight: 700 }} noWrap>
                    {selected.title}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }} noWrap>
                    {isQuestionnaire ? 'Questionnaire' : 'Note'} · Idea: {selected.idea}
                  </Typography>
                </Box>
                <Stack direction="row" spacing={0.5} alignItems="center">
                  {publicNoteHref(selected) ? (
                    <Tooltip title="Open published page">
                      <IconButton
                        size="small"
                        component="a"
                        href={publicNoteHref(selected)}
                        target="_blank"
                        rel="noopener noreferrer"
                        aria-label="Open published page"
                      >
                        <OpenInNewIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  ) : null}
                  <Button
                    size="small"
                    startIcon={<PublishOutlinedIcon />}
                    onClick={onPublish}
                    disabled={publishing}
                  >
                    {publishing ? 'Publishing…' : selected.published_path ? 'Republish' : 'Publish'}
                  </Button>
                  <IconButton size="small" color="error" onClick={onDelete} disabled={deleting} aria-label="Delete note">
                    <DeleteOutlineIcon fontSize="small" />
                  </IconButton>
                </Stack>
              </Stack>

              <Tabs
                value={previewTab}
                onChange={(_e, v) => setPreviewTab(v)}
                sx={{ px: 1, minHeight: 40, borderBottom: 1, borderColor: 'divider' }}
                variant="scrollable"
                allowScrollButtonsMobile
              >
                <Tab label="Preview" sx={{ minHeight: 40, textTransform: 'none' }} />
                <Tab label="Markdown" sx={{ minHeight: 40, textTransform: 'none' }} />
                <Tab label="HTML" sx={{ minHeight: 40, textTransform: 'none' }} />
                {isQuestionnaire ? <Tab label="Fill & analyze" sx={{ minHeight: 40, textTransform: 'none' }} /> : null}
              </Tabs>

              <Box sx={{ flex: 1, minHeight: 0, overflow: 'auto', p: 2 }}>
                {previewTab === 0 ? (
                  selected.html_content ? (
                    <Box
                      component="iframe"
                      title="Big note preview"
                      sandbox="allow-same-origin"
                      srcDoc={selected.html_content}
                      sx={{
                        width: '100%',
                        minHeight: 360,
                        border: 0,
                        borderRadius: 1,
                        bgcolor: '#0b1220',
                      }}
                    />
                  ) : (
                    <Typography color="text.secondary">No HTML content.</Typography>
                  )
                ) : null}
                {previewTab === 1 ? (
                  <Box
                    component="pre"
                    sx={{
                      m: 0,
                      p: 1.5,
                      borderRadius: 1,
                      bgcolor: 'action.hover',
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word',
                      fontSize: 13,
                    }}
                  >
                    {selected.markdown_content || '(empty)'}
                  </Box>
                ) : null}
                {previewTab === 2 ? (
                  <Box
                    component="pre"
                    sx={{
                      m: 0,
                      p: 1.5,
                      borderRadius: 1,
                      bgcolor: 'action.hover',
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word',
                      fontSize: 12,
                    }}
                  >
                    {selected.html_content || '(empty)'}
                  </Box>
                ) : null}
                {previewTab === 3 && isQuestionnaire ? (
                  <Stack spacing={2}>
                    <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                      Fill questionnaire
                    </Typography>
                    {questions.length === 0 ? (
                      <Alert severity="warning">No structured questions on this note. Regenerate as a questionnaire.</Alert>
                    ) : (
                      questions.map((q) => {
                        const type = String(q.type || 'text').toLowerCase();
                        if (type === 'textarea') {
                          return (
                            <TextField
                              key={q.id}
                              label={q.label || q.id}
                              multiline
                              minRows={3}
                              fullWidth
                              value={answers[q.id] || ''}
                              onChange={(e) => setAnswers((prev) => ({ ...prev, [q.id]: e.target.value }))}
                            />
                          );
                        }
                        if (type === 'single') {
                          return (
                            <FormControl key={q.id}>
                              <FormLabel>{q.label || q.id}</FormLabel>
                              <RadioGroup
                                value={answers[q.id] || ''}
                                onChange={(e) => setAnswers((prev) => ({ ...prev, [q.id]: e.target.value }))}
                              >
                                {(q.options || []).map((opt) => (
                                  <FormControlLabel key={opt} value={opt} control={<Radio size="small" />} label={opt} />
                                ))}
                              </RadioGroup>
                            </FormControl>
                          );
                        }
                        if (type === 'multi') {
                          const selectedOpts = Array.isArray(answers[q.id]) ? answers[q.id] : [];
                          return (
                            <FormControl key={q.id}>
                              <FormLabel>{q.label || q.id}</FormLabel>
                              <FormGroup>
                                {(q.options || []).map((opt) => (
                                  <FormControlLabel
                                    key={opt}
                                    control={
                                      <Checkbox
                                        size="small"
                                        checked={selectedOpts.includes(opt)}
                                        onChange={(e) => {
                                          setAnswers((prev) => {
                                            const cur = Array.isArray(prev[q.id]) ? prev[q.id] : [];
                                            const next = e.target.checked
                                              ? [...cur, opt]
                                              : cur.filter((x) => x !== opt);
                                            return { ...prev, [q.id]: next };
                                          });
                                        }}
                                      />
                                    }
                                    label={opt}
                                  />
                                ))}
                              </FormGroup>
                            </FormControl>
                          );
                        }
                        return (
                          <TextField
                            key={q.id}
                            label={q.label || q.id}
                            fullWidth
                            size="small"
                            value={answers[q.id] || ''}
                            onChange={(e) => setAnswers((prev) => ({ ...prev, [q.id]: e.target.value }))}
                          />
                        );
                      })
                    )}
                    <Box>
                      <Button variant="contained" onClick={onSubmitAnswers} disabled={savingResponse || questions.length === 0}>
                        {savingResponse ? 'Saving…' : 'Save response'}
                      </Button>
                    </Box>

                    <Divider />
                    <Stack direction="row" alignItems="center" justifyContent="space-between">
                      <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                        Responses ({responses.length})
                      </Typography>
                      <Button
                        size="small"
                        startIcon={<AnalyticsOutlinedIcon />}
                        onClick={onAnalyzeAll}
                        disabled={analyzingAll || responses.length === 0}
                      >
                        {analyzingAll ? 'Analyzing…' : 'Analyze all'}
                      </Button>
                    </Stack>
                    {allAnalysis ? (
                      <Box
                        component="pre"
                        sx={{
                          m: 0,
                          p: 1.5,
                          borderRadius: 1,
                          bgcolor: 'action.hover',
                          whiteSpace: 'pre-wrap',
                          fontSize: 13,
                        }}
                      >
                        {allAnalysis}
                      </Box>
                    ) : null}
                    {responses.map((r) => (
                      <Box key={r.id} sx={{ border: 1, borderColor: 'divider', borderRadius: 1, p: 1.5 }}>
                        <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={1}>
                          <Typography variant="body2" sx={{ fontWeight: 600 }}>
                            Response #{r.id} · {formatWhen(r.created_on)}
                          </Typography>
                          <Button
                            size="small"
                            startIcon={<AutoAwesomeIcon />}
                            onClick={() => onAnalyzeResponse(r.id)}
                            disabled={analyzingId === r.id}
                          >
                            {analyzingId === r.id ? 'Analyzing…' : 'AI analyze'}
                          </Button>
                        </Stack>
                        <Box
                          component="pre"
                          sx={{ m: 0, mt: 1, fontSize: 12, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}
                        >
                          {JSON.stringify(r.answers || r.answers_json, null, 2)}
                        </Box>
                        {r.analysis_markdown ? (
                          <Box
                            component="pre"
                            sx={{
                              m: 0,
                              mt: 1,
                              p: 1,
                              borderRadius: 1,
                              bgcolor: 'action.hover',
                              whiteSpace: 'pre-wrap',
                              fontSize: 13,
                            }}
                          >
                            {r.analysis_markdown}
                          </Box>
                        ) : null}
                      </Box>
                    ))}
                  </Stack>
                ) : null}
              </Box>

              <Divider />
              <Stack spacing={1} sx={{ p: 2 }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
                  Regenerate with prompt
                </Typography>
                <TextField
                  size="small"
                  multiline
                  minRows={2}
                  placeholder="e.g. Make it shorter, add more questions, more formal tone…"
                  value={regenPrompt}
                  onChange={(e) => setRegenPrompt(e.target.value)}
                  fullWidth
                />
                <Box>
                  <Button
                    variant="outlined"
                    startIcon={<AutoAwesomeIcon />}
                    onClick={onRegenerate}
                    disabled={regenerating || !regenPrompt.trim()}
                  >
                    {regenerating ? 'Regenerating…' : 'Regenerate'}
                  </Button>
                </Box>
              </Stack>
            </>
          )}
        </Box>
      </Box>

      <Dialog open={createOpen} onClose={() => (!creating ? setCreateOpen(false) : null)} fullWidth maxWidth="sm">
        <DialogTitle>New Big note</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 0.5 }}>
            <TextField
              select
              label="Type"
              value={noteKind}
              onChange={(e) => setNoteKind(e.target.value)}
              size="small"
              disabled={creating}
            >
              <MenuItem value="note">Note (article HTML + markdown)</MenuItem>
              <MenuItem value="questionnaire">Questionnaire (fillable + AI analysis)</MenuItem>
            </TextField>
            <TextField
              label="Your idea"
              placeholder={
                noteKind === 'questionnaire'
                  ? 'Describe the questionnaire you want (audience, topics, tone)…'
                  : 'Describe the note you want AI to write…'
              }
              value={idea}
              onChange={(e) => setIdea(e.target.value)}
              multiline
              minRows={4}
              fullWidth
              autoFocus
              disabled={creating}
            />
            <FormControl component="fieldset" disabled={creating}>
              <FormLabel component="legend">Theme</FormLabel>
              <RadioGroup
                row
                value={theme === 'light' ? 'light' : 'dark'}
                onChange={(e) => setTheme(e.target.value === 'light' ? 'light' : 'dark')}
              >
                <FormControlLabel value="dark" control={<Radio size="small" />} label="Dark" />
                <FormControlLabel value="light" control={<Radio size="small" />} label="Light" />
              </RadioGroup>
            </FormControl>
            {creating ? (
              <Alert severity="info" icon={<CircularProgress size={18} />}>
                Generating {noteKind === 'questionnaire' ? 'questionnaire' : 'markdown and HTML'}… this can take a minute.
              </Alert>
            ) : null}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)} disabled={creating}>
            Cancel
          </Button>
          <Button variant="contained" onClick={onCreate} disabled={creating || !idea.trim()}>
            Generate
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
