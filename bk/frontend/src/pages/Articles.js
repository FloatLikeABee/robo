import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Grid,
  IconButton,
  InputAdornment,
  Alert,
  Chip,
  Autocomplete,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  PlayArrow as RunIcon,
  Edit as EditIcon,
  Download as DownloadIcon,
  OpenInFull as OpenInFullIcon,
} from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import api, { getArticleDownloadUrl } from '../services/api';
import SystemPromptInput from '../components/SystemPromptInput';

const Articles = ({ suppressModuleHeader = false }) => {
  const queryClient = useQueryClient();
  const { data: articles = [], isLoading, error } = useQuery('articles', api.getArticles, {
    staleTime: 5 * 60 * 1000,
  });
  const { data: assistants = [] } = useQuery('assistants', api.getAssistants, {
    staleTime: 5 * 60 * 1000,
  });
  const { data: models = [] } = useQuery('models', api.getModels, { staleTime: 5 * 60 * 1000 });
  const { data: providersData } = useQuery('providers', api.getProviders, { staleTime: 5 * 60 * 1000 });
  const { data: collections = [] } = useQuery('collections', api.getRAGCollections, {
    staleTime: 5 * 60 * 1000,
  });

  const [openCreateDialog, setOpenCreateDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [editingArticleId, setEditingArticleId] = useState(null);
  const [selectedArticle, setSelectedArticle] = useState(null);
  const [deleteConfirmDialog, setDeleteConfirmDialog] = useState({ open: false, articleId: null });

  const [createForm, setCreateForm] = useState({
    name: '',
    description: '',
    customization_id: '',
    system_prompt: '',
    rag_collection: '',
    llm_provider: '',
    model_name: '',
    default_chapters: 1,
  });

  const [sourceText, setSourceText] = useState('');
  const [runChapters, setRunChapters] = useState(1);
  const [outputContent, setOutputContent] = useState('');
  const [genError, setGenError] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [lastFileId, setLastFileId] = useState(null);
  const [lastFilename, setLastFilename] = useState(null);
  const [sourceExpandOpen, setSourceExpandOpen] = useState(false);
  const [sourceModalDraft, setSourceModalDraft] = useState('');

  useEffect(() => {
    if (selectedArticle) {
      const dc = selectedArticle.default_chapters && selectedArticle.default_chapters >= 1
        ? selectedArticle.default_chapters
        : 1;
      setRunChapters(dc);
    }
  }, [selectedArticle]);

  const createMutation = useMutation(api.createArticle, {
    onSuccess: () => {
      queryClient.invalidateQueries('articles');
      setOpenCreateDialog(false);
      resetCreateForm();
    },
  });

  const updateMutation = useMutation(({ articleId, payload }) => api.updateArticle(articleId, payload), {
    onSuccess: () => {
      queryClient.invalidateQueries('articles');
      setOpenEditDialog(false);
      setEditingArticleId(null);
      resetCreateForm();
    },
  });

  const deleteMutation = useMutation(api.deleteArticle, {
    onSuccess: () => {
      queryClient.invalidateQueries('articles');
      if (selectedArticle && selectedArticle.id === deleteConfirmDialog.articleId) {
        setSelectedArticle(null);
      }
    },
  });

  const resetCreateForm = () => {
    setCreateForm({
      name: '',
      description: '',
      customization_id: '',
      system_prompt: '',
      rag_collection: '',
      llm_provider: '',
      model_name: '',
      default_chapters: 1,
    });
  };

  const handleCreate = () => {
    const dc = Math.max(1, parseInt(createForm.default_chapters, 10) || 1);
    createMutation.mutate({
      name: createForm.name,
      description: createForm.description || null,
      customization_id: createForm.customization_id,
      system_prompt: createForm.system_prompt,
      rag_collection: createForm.rag_collection || null,
      llm_provider: createForm.llm_provider || null,
      model_name: createForm.model_name || null,
      default_chapters: dc,
    });
  };

  const handleEditArticle = (article) => {
    setEditingArticleId(article.id);
    setCreateForm({
      name: article.name || '',
      description: article.description || '',
      customization_id: article.customization_id || '',
      system_prompt: article.system_prompt || '',
      rag_collection: article.rag_collection || '',
      llm_provider: article.llm_provider || '',
      model_name: article.model_name || '',
      default_chapters: article.default_chapters && article.default_chapters >= 1 ? article.default_chapters : 1,
    });
    setOpenEditDialog(true);
  };

  const handleUpdate = () => {
    const dc = Math.max(1, parseInt(createForm.default_chapters, 10) || 1);
    updateMutation.mutate({
      articleId: editingArticleId,
      payload: {
        name: createForm.name,
        description: createForm.description || null,
        customization_id: createForm.customization_id, // maps to assistant id
        system_prompt: createForm.system_prompt,
        rag_collection: createForm.rag_collection || null,
        llm_provider: createForm.llm_provider || null,
        model_name: createForm.model_name || null,
        default_chapters: dc,
      },
    });
  };

  const handleGenerate = async () => {
    if (!selectedArticle) return;
    const ch = Math.max(1, parseInt(runChapters, 10) || 1);
    setIsGenerating(true);
    setGenError('');
    setOutputContent('');
    setLastFileId(null);
    setLastFilename(null);

    try {
      await api.streamArticleGenerate(
        selectedArticle.id,
        {
          source_text: sourceText.trim() || null,
          chapters: ch,
          n_results: 3,
        },
        (evt) => {
          if (evt.type === 'log') {
            setOutputContent((prev) => prev + (evt.message || '') + '\n');
          } else if (evt.type === 'delta') {
            setOutputContent((prev) => prev + (evt.text || ''));
          } else if (evt.type === 'chapter_start') {
            setOutputContent((prev) => prev + `\n--- Chapter ${evt.chapter} ---\n`);
          } else if (evt.type === 'chapter_done') {
            setOutputContent((prev) => prev + `\n[Chapter ${evt.chapter} complete]\n`);
          } else if (evt.type === 'done') {
            setLastFileId(evt.file_id);
            setLastFilename(evt.filename || null);
            setOutputContent((prev) => prev + `\nSaved: ${evt.filename || evt.file_id}\n`);
          } else if (evt.type === 'error') {
            setGenError(evt.message || 'Generation failed');
          }
        }
      );
    } catch (err) {
      setGenError(err.message || 'Generation failed');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleImportFile = (e) => {
    const f = e.target.files?.[0];
    if (!f) return;
    const reader = new FileReader();
    reader.onload = () => {
      const t = String(reader.result || '');
      setSourceText(t);
      setSourceModalDraft(t);
    };
    reader.readAsText(f);
    e.target.value = '';
  };

  const articleFormFields = (
    <Box sx={{ mt: 1 }}>
      <TextField
        fullWidth
        label="Name"
        value={createForm.name}
        onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
        sx={{ mb: 2 }}
      />
      <TextField
        fullWidth
        label="Description"
        value={createForm.description}
        onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
        multiline
        rows={2}
        sx={{ mb: 2 }}
      />
      <FormControl fullWidth sx={{ mb: 2 }} required>
        <InputLabel id="assistant-label">Assistant</InputLabel>
        <Select
          labelId="assistant-label"
          label="Assistant"
          value={createForm.customization_id}
          onChange={(e) => setCreateForm({ ...createForm, customization_id: e.target.value })}
        >
          {assistants.map((a) => (
            <MenuItem key={a.id} value={a.id}>
              {a.name}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <SystemPromptInput
        label="Article system prompt"
        value={createForm.system_prompt}
        onChange={(e) => setCreateForm({ ...createForm, system_prompt: e.target.value })}
        rows={6}
        placeholder="Instructions for this article or novel (merged with the selected customization)."
        required
      />
      <Autocomplete
        freeSolo
        options={collections.map((col) => col.name)}
        value={createForm.rag_collection}
        onChange={(event, newValue) => {
          setCreateForm({ ...createForm, rag_collection: newValue || '' });
        }}
        renderInput={(params) => (
          <TextField
            {...params}
            label="RAG collection (optional)"
            helperText="Retrieval context per chapter when set"
            sx={{ mb: 2, mt: 2 }}
          />
        )}
      />
      <FormControl fullWidth sx={{ mb: 2 }}>
        <InputLabel>LLM provider</InputLabel>
        <Select
          value={createForm.llm_provider}
          label="LLM provider"
          onChange={(e) => setCreateForm({ ...createForm, llm_provider: e.target.value })}
        >
          <MenuItem value="">
            <em>Use default provider</em>
          </MenuItem>
          {providersData?.providers?.map((provider) => (
            <MenuItem key={provider} value={provider}>
              {provider}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <Autocomplete
        freeSolo
        options={models.map((model) => model.name)}
        value={createForm.model_name}
        onChange={(event, newValue) => {
          setCreateForm({ ...createForm, model_name: newValue || '' });
        }}
        onInputChange={(event, newInputValue) => {
          setCreateForm({ ...createForm, model_name: newInputValue });
        }}
        renderInput={(params) => (
          <TextField
            {...params}
            label="Model name (optional)"
            sx={{ mb: 2 }}
          />
        )}
      />
      <TextField
        fullWidth
        type="number"
        label="Default chapter count"
        value={createForm.default_chapters}
        onChange={(e) => {
          const v = parseInt(e.target.value, 10);
          setCreateForm({ ...createForm, default_chapters: Number.isFinite(v) && v >= 1 ? v : 1 });
        }}
        inputProps={{ min: 1 }}
        helperText="Minimum 1. Used as the default when you run generation."
        sx={{ mb: 2 }}
      />
    </Box>
  );

  return (
    <Box>
      {!suppressModuleHeader ? (
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h4">Articles</Typography>
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={() => setOpenCreateDialog(true)}>
            New article profile
          </Button>
        </Box>
      ) : (
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 1.5 }}>
          <Button
            variant="contained"
            size="small"
            startIcon={<AddIcon />}
            onClick={() => setOpenCreateDialog(true)}
          >
            New article profile
          </Button>
        </Box>
      )}

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          Failed to load articles
        </Alert>
      )}

      <Grid container spacing={3} sx={{ alignItems: 'stretch' }}>
        <Grid item xs={12} md={5}>
          <Box
            sx={{
              maxHeight: 'calc(100dvh - 200px)',
              overflowY: 'auto',
              pr: 0.5,
              pb: 0.5,
            }}
          >
            <Grid container spacing={2}>
            {articles.map((article) => (
              <Grid item xs={12} key={article.id}>
                <Card
                  sx={{
                    boxShadow: 2,
                    cursor: 'pointer',
                    height: 236,
                    display: 'flex',
                    flexDirection: 'column',
                    border:
                      selectedArticle && selectedArticle.id === article.id
                        ? '2px solid #1976d2'
                        : '1px solid rgba(0,0,0,0.12)',
                  }}
                  onClick={() => setSelectedArticle(article)}
                >
                  <CardContent
                    sx={{
                      p: 2,
                      flex: 1,
                      display: 'flex',
                      flexDirection: 'row',
                      alignItems: 'flex-start',
                      gap: 1,
                      minHeight: 0,
                      '&:last-child': { pb: 2 },
                    }}
                  >
                    <Box
                      sx={{
                        flex: 1,
                        minWidth: 0,
                        display: 'flex',
                        flexDirection: 'column',
                        height: '100%',
                        overflow: 'hidden',
                      }}
                    >
                      <Typography
                        variant="subtitle1"
                        component="h3"
                        title={article.name}
                        sx={{
                          fontWeight: 600,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {article.name}
                      </Typography>
                      <Typography
                        variant="body2"
                        color="text.secondary"
                        title={article.description || undefined}
                        sx={{
                          mt: 0.5,
                          flex: '1 1 auto',
                          overflow: 'hidden',
                          display: '-webkit-box',
                          WebkitBoxOrient: 'vertical',
                          WebkitLineClamp: 2,
                          lineHeight: 1.4,
                          minHeight: '2.8em',
                        }}
                      >
                        {article.description || ' '}
                      </Typography>
                      <Box
                        sx={{
                          mt: 'auto',
                          pt: 1,
                          display: 'flex',
                          gap: 0.5,
                          flexWrap: 'wrap',
                          maxHeight: 80,
                          overflow: 'hidden',
                        }}
                      >
                        <Chip size="small" label={`Assistant: ${article.customization_id}`} variant="outlined" />
                        {article.rag_collection && (
                          <Chip size="small" label={`RAG: ${article.rag_collection}`} color="primary" variant="outlined" />
                        )}
                        {article.llm_provider && (
                          <Chip size="small" label={article.llm_provider} variant="outlined" />
                        )}
                        {article.model_name && <Chip size="small" label={article.model_name} variant="outlined" />}
                        <Chip size="small" label={`Default ch.: ${article.default_chapters || 1}`} variant="outlined" />
                      </Box>
                    </Box>
                    <Box sx={{ display: 'flex', gap: 0.5, flexShrink: 0 }}>
                      <IconButton
                        size="small"
                        color="primary"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleEditArticle(article);
                        }}
                      >
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton
                        size="small"
                        color="error"
                        onClick={(e) => {
                          e.stopPropagation();
                          setDeleteConfirmDialog({ open: true, articleId: article.id });
                        }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Box>
                  </CardContent>
                </Card>
              </Grid>
            ))}
            {!isLoading && articles.length === 0 && (
              <Grid item xs={12}>
                <Typography variant="body2" color="text.secondary">
                  No article profiles yet. Create one to generate chapters.
                </Typography>
              </Grid>
            )}
            </Grid>
          </Box>
        </Grid>

        <Grid item xs={12} md={7} sx={{ display: 'flex', minHeight: 0 }}>
          <Card
            sx={{
              boxShadow: 2,
              width: '100%',
              maxHeight: 'calc(100dvh - 200px)',
              display: 'flex',
              flexDirection: 'column',
              minHeight: 0,
            }}
          >
            <CardContent
              sx={{
                p: 3,
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                minHeight: 0,
                overflow: 'hidden',
                '&:last-child': { pb: 3 },
              }}
            >
              <Typography variant="h6" gutterBottom sx={{ flexShrink: 0 }}>
                Generate
              </Typography>
              {selectedArticle ? (
                <Box
                  sx={{
                    flex: 1,
                    minHeight: 0,
                    display: 'flex',
                    flexDirection: 'column',
                    overflow: 'hidden',
                  }}
                >
                  <Typography variant="subtitle1" sx={{ fontWeight: 'medium', mb: 1, flexShrink: 0 }}>
                    {selectedArticle.name}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 1.5, flexShrink: 0 }}>
                    Base customization:{' '}
                    {assistants.find((a) => a.id === selectedArticle.customization_id)?.name ||
                      selectedArticle.customization_id}
                    . Article prompt is merged when generating (see edit).
                  </Typography>

                  <Box sx={{ position: 'relative', width: '100%', mb: 1.5, flexShrink: 0 }}>
                    <TextField
                      fullWidth
                      size="small"
                      label="Optional source text (paste or type)"
                      value={sourceText}
                      onChange={(e) => setSourceText(e.target.value)}
                      multiline
                      rows={2}
                      placeholder="Novel seed, outline, or full text — optional."
                      InputProps={{
                        endAdornment: (
                          <InputAdornment position="end" sx={{ alignSelf: 'flex-start', mt: 0.75 }}>
                            <IconButton
                              size="small"
                              aria-label="Expand source text editor"
                              onMouseDown={(e) => e.preventDefault()}
                              onClick={() => {
                                setSourceModalDraft(sourceText);
                                setSourceExpandOpen(true);
                              }}
                            >
                              <OpenInFullIcon fontSize="small" />
                            </IconButton>
                          </InputAdornment>
                        ),
                      }}
                    />
                  </Box>

                  <Box
                    sx={{
                      display: 'flex',
                      flexDirection: 'row',
                      flexWrap: 'nowrap',
                      alignItems: 'center',
                      gap: 1,
                      mb: 1.5,
                      flexShrink: 0,
                      overflowX: 'auto',
                      pb: 0.25,
                      '&::-webkit-scrollbar': { height: 6 },
                    }}
                  >
                    <Button variant="outlined" component="label" size="small" sx={{ flexShrink: 0, whiteSpace: 'nowrap' }}>
                      Import text file
                      <input type="file" accept=".txt,.md,.text,text/plain" hidden onChange={handleImportFile} />
                    </Button>
                    <TextField
                      type="number"
                      label="Chapters"
                      value={runChapters}
                      size="small"
                      onChange={(e) => {
                        const v = parseInt(e.target.value, 10);
                        setRunChapters(Number.isFinite(v) && v >= 1 ? v : 1);
                      }}
                      inputProps={{ min: 1 }}
                      sx={{ width: 96, flexShrink: 0 }}
                    />
                    <Button
                      variant="contained"
                      startIcon={<RunIcon />}
                      onClick={handleGenerate}
                      disabled={isGenerating}
                      sx={{ flex: '1 1 auto', minWidth: 120, whiteSpace: 'nowrap' }}
                    >
                      {isGenerating ? 'Generating…' : 'Generate'}
                    </Button>
                  </Box>

                  {genError && (
                    <Alert severity="error" sx={{ mb: 1, flexShrink: 0 }}>
                      {genError}
                    </Alert>
                  )}

                  <Typography variant="subtitle2" sx={{ mb: 0.5, flexShrink: 0, mt: 0.5 }}>
                    Output
                  </Typography>
                  <Box
                    sx={{
                      flex: 1,
                      minHeight: 0,
                      whiteSpace: 'pre-wrap',
                      border: '1px solid',
                      borderColor: 'divider',
                      borderRadius: 1,
                      p: 2,
                      overflowY: 'auto',
                      bgcolor: 'action.hover',
                      fontFamily: 'ui-monospace, monospace',
                      fontSize: '0.85rem',
                    }}
                  >
                    {outputContent || (isGenerating ? '…' : 'Generation output will appear here.')}
                  </Box>

                  {lastFileId && (
                    <Button
                      component="a"
                      href={getArticleDownloadUrl(lastFileId)}
                      download={lastFilename || `article_${lastFileId.slice(0, 8)}.txt`}
                      variant="outlined"
                      startIcon={<DownloadIcon />}
                      sx={{ mt: 1.5, flexShrink: 0, alignSelf: 'flex-start' }}
                    >
                      Download saved file
                    </Button>
                  )}
                </Box>
              ) : (
                <Typography variant="body2" color="text.secondary">
                  Select an article profile or create one.
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Dialog
        open={sourceExpandOpen}
        onClose={() => setSourceExpandOpen(false)}
        maxWidth="md"
        fullWidth
        PaperProps={{
          sx: { height: '80vh', maxHeight: '80vh', display: 'flex', flexDirection: 'column' },
        }}
      >
        <DialogTitle>Optional source text</DialogTitle>
        <DialogContent sx={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0, pt: 1 }}>
          <TextField
            fullWidth
            multiline
            value={sourceModalDraft}
            onChange={(e) => setSourceModalDraft(e.target.value)}
            placeholder="Novel seed, outline, or full text — optional."
            sx={{
              flex: 1,
              '& .MuiInputBase-root': {
                height: '100%',
                alignItems: 'flex-start',
                py: 1,
              },
              '& textarea': {
                height: 'calc(80vh - 220px) !important',
                overflow: 'auto !important',
                resize: 'none',
              },
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSourceExpandOpen(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={() => {
              setSourceText(sourceModalDraft);
              setSourceExpandOpen(false);
            }}
          >
            Save
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={openCreateDialog} onClose={() => setOpenCreateDialog(false)} maxWidth="md" fullWidth>
        <DialogTitle>New article profile</DialogTitle>
        <DialogContent>
          {createMutation.isError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Failed to create. Ensure a customization is selected and prompts are filled.
            </Alert>
          )}
          {articleFormFields}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenCreateDialog(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={handleCreate}
            disabled={
              createMutation.isLoading ||
              !createForm.name.trim() ||
              !createForm.system_prompt.trim() ||
              !createForm.customization_id
            }
          >
            {createMutation.isLoading ? 'Creating…' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={openEditDialog}
        onClose={() => {
          setOpenEditDialog(false);
          setEditingArticleId(null);
          resetCreateForm();
        }}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Edit article profile</DialogTitle>
        <DialogContent>
          {updateMutation.isError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Failed to update.
            </Alert>
          )}
          {articleFormFields}
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              setOpenEditDialog(false);
              setEditingArticleId(null);
              resetCreateForm();
            }}
          >
            Cancel
          </Button>
          <Button
            variant="contained"
            onClick={handleUpdate}
            disabled={
              updateMutation.isLoading ||
              !createForm.name.trim() ||
              !createForm.system_prompt.trim() ||
              !createForm.customization_id
            }
          >
            {updateMutation.isLoading ? 'Saving…' : 'Save'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={deleteConfirmDialog.open}
        onClose={() => setDeleteConfirmDialog({ open: false, articleId: null })}
      >
        <DialogTitle>Delete article profile?</DialogTitle>
        <DialogActions>
          <Button onClick={() => setDeleteConfirmDialog({ open: false, articleId: null })}>Cancel</Button>
          <Button
            color="error"
            variant="contained"
            onClick={() => {
              if (deleteConfirmDialog.articleId) {
                deleteMutation.mutate(deleteConfirmDialog.articleId);
              }
              setDeleteConfirmDialog({ open: false, articleId: null });
            }}
          >
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Articles;
