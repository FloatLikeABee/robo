import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  TextField,
  Typography,
  Paper,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
  Chip,
  Link,
  Tooltip,
  Stack,
} from '@mui/material';
import {
  Delete as DeleteIcon,
  Edit as EditIcon,
  PlayArrow as RunIcon,
  Add as AddIcon,
  SmartToy as AssistantIcon,
} from '@mui/icons-material';
import { Link as RouterLink } from 'react-router-dom';
import ModuleShell from '../components/ModuleShell';
import api from '../services/api';

const EMPTY_FORM = {
  name: '',
  description: '',
  system_prompt: '',
  llm_provider: 'gemini',
  model_name: '',
  rag_collections: [],
};

const providerLabel = (p) => {
  const v = String(p || '').toLowerCase();
  if (!v) return 'Default';
  return v.charAt(0).toUpperCase() + v.slice(1);
};

const AssistantManager = () => {
  const [assistants, setAssistants] = useState([]);
  const [collections, setCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [form, setForm] = useState(EMPTY_FORM);
  const [runDialogOpen, setRunDialogOpen] = useState(false);
  const [runQuery, setRunQuery] = useState('');
  const [runResult, setRunResult] = useState('');
  const [running, setRunning] = useState(false);

  const fetchAssistants = async () => {
    setLoading(true);
    try {
      const data = await api.getAssistants();
      setAssistants(Array.isArray(data) ? data : data?.assistants || []);
    } catch (e) {
      console.error('Failed to load assistants', e);
    } finally {
      setLoading(false);
    }
  };

  const fetchCollections = async () => {
    try {
      const list = await api.getRAGCollections();
      setCollections(Array.isArray(list) ? list : list?.collections || []);
    } catch (e) {
      console.error('Failed to load RAG collections', e);
      setCollections([]);
    }
  };

  useEffect(() => {
    fetchAssistants();
    fetchCollections();
  }, []);

  const handleSave = async () => {
    const selectedRags = Array.isArray(form.rag_collections) ? form.rag_collections : [];
    const payload = {
      name: form.name,
      description: form.description,
      system_prompt: form.system_prompt,
      llm_provider: form.llm_provider,
      model_name: form.model_name || undefined,
      existing_rag_collections: selectedRags,
    };
    try {
      if (editingId) {
        await api.updateAssistant(editingId, payload);
      } else {
        await api.createAssistant(payload);
      }
      setDialogOpen(false);
      setForm(EMPTY_FORM);
      setEditingId(null);
      await fetchAssistants();
    } catch (e) {
      console.error('Failed to save assistant', e);
    }
  };

  const handleEdit = (a) => {
    setEditingId(a.id);
    setForm({
      name: a.name || '',
      description: a.description || '',
      system_prompt: a.system_prompt || '',
      llm_provider: a.llm_provider || 'gemini',
      model_name: a.model_name || '',
      rag_collections: Array.isArray(a.rag_collections) ? a.rag_collections : [],
    });
    setDialogOpen(true);
    void fetchCollections();
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Delete this assistant?')) return;
    try {
      await api.deleteAssistant(id);
      await fetchAssistants();
    } catch (e) {
      console.error('Failed to delete assistant', e);
    }
  };

  const handleRun = (a) => {
    setEditingId(a.id);
    setRunDialogOpen(true);
    setRunQuery('');
    setRunResult('');
  };

  const submitRun = async () => {
    if (!runQuery.trim()) return;
    setRunning(true);
    try {
      const data = await api.runAssistant(editingId, runQuery);
      setRunResult(data?.response || JSON.stringify(data));
    } catch (e) {
      setRunResult(`Error: ${e.message}`);
    } finally {
      setRunning(false);
    }
  };

  return (
    <ModuleShell
      title="Assistants"
      helpText="Each assistant has a system prompt and can use one or more RAG collections. Collect RAG data under RAG (file upload or API request). Morph AI can select these assistants for chat."
    >
      <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', gap: 2 }}>
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1, flexShrink: 0 }}>
          <Button component={RouterLink} to="/rag" variant="outlined">
            Manage RAG
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => {
              setEditingId(null);
              setForm(EMPTY_FORM);
              setDialogOpen(true);
              void fetchCollections();
            }}
          >
            New Assistant
          </Button>
        </Box>

        <Box sx={{ flex: 1, minHeight: 0, overflow: 'auto', pr: 0.5 }}>
          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
              <CircularProgress size={32} />
            </Box>
          ) : assistants.length === 0 ? (
            <Paper
              sx={{
                p: 5,
                textAlign: 'center',
                border: '1px dashed',
                borderColor: 'rgba(56, 189, 248, 0.28)',
                background:
                  'linear-gradient(160deg, rgba(10,15,26,0.9) 0%, rgba(26,45,74,0.35) 100%)',
              }}
            >
              <AssistantIcon sx={{ fontSize: 48, color: 'primary.light', mb: 1, opacity: 0.85 }} />
              <Typography variant="h6" sx={{ mb: 0.5 }}>
                No assistants yet
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Create an assistant with a system prompt and optional RAG collections.
              </Typography>
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={() => {
                  setEditingId(null);
                  setForm(EMPTY_FORM);
                  setDialogOpen(true);
                  void fetchCollections();
                }}
              >
                New Assistant
              </Button>
            </Paper>
          ) : (
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: {
                  xs: '1fr',
                  sm: 'repeat(2, minmax(0, 1fr))',
                  lg: 'repeat(3, minmax(0, 1fr))',
                },
                gap: 2,
                pb: 1,
              }}
            >
              {assistants.map((a) => {
                const rags = Array.isArray(a.rag_collections) ? a.rag_collections : [];
                return (
                  <Paper
                    key={a.id}
                    elevation={0}
                    sx={{
                      position: 'relative',
                      display: 'flex',
                      flexDirection: 'column',
                      gap: 1.25,
                      p: 2.25,
                      minHeight: 188,
                      borderRadius: 2.5,
                      overflow: 'hidden',
                      border: '1px solid',
                      borderColor: 'rgba(56, 189, 248, 0.18)',
                      background:
                        'linear-gradient(155deg, rgba(6,10,18,0.98) 0%, rgba(26,45,74,0.42) 55%, rgba(10,15,26,0.95) 100%)',
                      transition: 'transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease',
                      '&:hover': {
                        transform: 'translateY(-3px)',
                        borderColor: 'rgba(56, 189, 248, 0.45)',
                        boxShadow: '0 12px 32px rgba(14, 165, 233, 0.12)',
                      },
                      '&::before': {
                        content: '""',
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        right: 0,
                        height: 3,
                        background: 'linear-gradient(90deg, #0ea5e9, #3b82f6, #38bdf8)',
                      },
                    }}
                  >
                    <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1.25 }}>
                      <Box
                        sx={{
                          width: 40,
                          height: 40,
                          borderRadius: 1.5,
                          flexShrink: 0,
                          display: 'grid',
                          placeItems: 'center',
                          background: 'rgba(56, 189, 248, 0.12)',
                          border: '1px solid rgba(56, 189, 248, 0.28)',
                          color: '#38bdf8',
                        }}
                      >
                        <AssistantIcon fontSize="small" />
                      </Box>
                      <Box sx={{ minWidth: 0, flex: 1 }}>
                        <Typography
                          variant="subtitle1"
                          sx={{
                            fontWeight: 700,
                            lineHeight: 1.25,
                            letterSpacing: '0.01em',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                          }}
                          title={a.name}
                        >
                          {a.name}
                        </Typography>
                        <Stack direction="row" spacing={0.75} sx={{ mt: 0.75, flexWrap: 'wrap', gap: 0.5 }}>
                          <Chip
                            size="small"
                            label={providerLabel(a.llm_provider)}
                            sx={{
                              height: 22,
                              bgcolor: 'rgba(59, 130, 246, 0.16)',
                              color: '#93c5fd',
                              border: '1px solid rgba(59, 130, 246, 0.35)',
                            }}
                          />
                          {a.model_name ? (
                            <Chip
                              size="small"
                              label={a.model_name}
                              variant="outlined"
                              sx={{ height: 22, borderColor: 'rgba(148, 163, 184, 0.35)', color: 'text.secondary' }}
                            />
                          ) : null}
                        </Stack>
                      </Box>
                    </Box>

                    <Typography
                      variant="body2"
                      color="text.secondary"
                      sx={{
                        flex: 1,
                        display: '-webkit-box',
                        WebkitLineClamp: 3,
                        WebkitBoxOrient: 'vertical',
                        overflow: 'hidden',
                        lineHeight: 1.45,
                        minHeight: '4.35em',
                      }}
                    >
                      {a.description?.trim() || 'No description'}
                    </Typography>

                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, minHeight: 24 }}>
                      {rags.length === 0 ? (
                        <Typography variant="caption" color="text.secondary">
                          No RAG collections
                        </Typography>
                      ) : (
                        rags.slice(0, 4).map((name) => (
                          <Chip
                            key={name}
                            label={name}
                            size="small"
                            sx={{
                              height: 22,
                              maxWidth: 140,
                              bgcolor: 'rgba(14, 165, 233, 0.1)',
                              border: '1px solid rgba(14, 165, 233, 0.28)',
                              color: '#7dd3fc',
                              '& .MuiChip-label': { overflow: 'hidden', textOverflow: 'ellipsis' },
                            }}
                          />
                        ))
                      )}
                      {rags.length > 4 ? (
                        <Chip size="small" label={`+${rags.length - 4}`} sx={{ height: 22 }} />
                      ) : null}
                    </Box>

                    <Box
                      sx={{
                        display: 'flex',
                        justifyContent: 'flex-end',
                        gap: 0.5,
                        pt: 0.5,
                        borderTop: '1px solid rgba(148, 163, 184, 0.12)',
                      }}
                    >
                      <Tooltip title="Run">
                        <IconButton size="small" onClick={() => handleRun(a)} aria-label={`Run ${a.name}`}>
                          <RunIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Edit">
                        <IconButton size="small" onClick={() => handleEdit(a)} aria-label={`Edit ${a.name}`}>
                          <EditIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Delete">
                        <IconButton size="small" onClick={() => handleDelete(a.id)} aria-label={`Delete ${a.name}`}>
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </Paper>
                );
              })}
            </Box>
          )}
        </Box>
      </Box>

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>{editingId ? 'Edit Assistant' : 'New Assistant'}</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
            <TextField
              label="Name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              fullWidth
            />
            <TextField
              label="Description"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              fullWidth
            />
            <TextField
              label="System Prompt"
              value={form.system_prompt}
              onChange={(e) => setForm({ ...form, system_prompt: e.target.value })}
              multiline
              rows={6}
              fullWidth
            />
            <FormControl fullWidth>
              <InputLabel>Provider</InputLabel>
              <Select
                value={form.llm_provider}
                label="Provider"
                onChange={(e) => setForm({ ...form, llm_provider: e.target.value })}
              >
                <MenuItem value="gemini">Gemini</MenuItem>
                <MenuItem value="qwen">Qwen</MenuItem>
                <MenuItem value="mistral">Mistral</MenuItem>
                <MenuItem value="groq">Groq</MenuItem>
              </Select>
            </FormControl>
            <TextField
              label="Model name (optional)"
              value={form.model_name}
              onChange={(e) => setForm({ ...form, model_name: e.target.value })}
              fullWidth
            />
            <FormControl fullWidth>
              <InputLabel>RAG collections</InputLabel>
              <Select
                multiple
                label="RAG collections"
                value={form.rag_collections}
                onChange={(e) => setForm({ ...form, rag_collections: e.target.value })}
                renderValue={(selected) => (
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                    {selected.map((value) => (
                      <Chip key={value} label={value} size="small" />
                    ))}
                  </Box>
                )}
              >
                {collections.map((collection) => {
                  const name = collection.name || collection;
                  const count = collection.count ?? collection.document_count;
                  return (
                    <MenuItem key={name} value={name}>
                      {name}
                      {count != null ? ` (${count} documents)` : ''}
                    </MenuItem>
                  );
                })}
              </Select>
            </FormControl>
            <Typography variant="caption" color="text.secondary">
              Select one or more RAG collections. Add collections in{' '}
              <Link component={RouterLink} to="/rag" underline="hover">
                RAG
              </Link>{' '}
              via file upload or API request.
            </Typography>
            {collections.length === 0 ? (
              <Typography variant="body2" color="warning.main">
                No RAG collections yet. Open RAG to upload files or pull content from an API request.
              </Typography>
            ) : null}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleSave} variant="contained" disabled={!form.name || !form.system_prompt.trim()}>
            Save
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={runDialogOpen} onClose={() => setRunDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>Run Assistant</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
            <TextField
              label="Query"
              value={runQuery}
              onChange={(e) => setRunQuery(e.target.value)}
              multiline
              rows={3}
              fullWidth
            />
            <Button
              variant="contained"
              onClick={submitRun}
              disabled={running || !runQuery.trim()}
              startIcon={running ? <CircularProgress size={16} /> : <RunIcon />}
            >
              Run
            </Button>
            {runResult && (
              <Paper sx={{ p: 2, bgcolor: 'background.default' }}>
                <Typography variant="subtitle2">Result</Typography>
                <Typography sx={{ whiteSpace: 'pre-wrap' }}>{runResult}</Typography>
              </Paper>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRunDialogOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </ModuleShell>
  );
};

export default AssistantManager;
