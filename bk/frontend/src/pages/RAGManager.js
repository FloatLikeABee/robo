import React, { useState } from 'react';
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
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Grid,
  IconButton,
  List,
  ListItem,
  Alert,
  Collapse,
  Tabs,
  Tab,
  Switch,
  FormControlLabel,
  Paper,
  Divider,
  Chip,

  CircularProgress,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Search as SearchIcon,
  UploadFile as UploadFileIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  Lock as LockIcon,
  Language as WebIcon,
  Api as ApiIcon,
  AutoAwesome as SuggestIcon,
  Save as SaveIcon,
} from '@mui/icons-material';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import api from '../services/api';
import ReactJson from 'react-json-view';
import ReactMarkdown from 'react-markdown';

// ─── Shared helpers ────────────────────────────────────────────────
const parseJsonOrEmpty = (text) => {
  if (!text || !text.trim()) return {};
  try {
    const parsed = JSON.parse(text);
    if (typeof parsed !== 'object' || Array.isArray(parsed)) {
      throw new Error('Must be a JSON object');
    }
    return parsed;
  } catch (e) {
    throw new Error(`Invalid JSON: ${e.message}`);
  }
};

const tryParseJson = (text) => {
  try { return JSON.parse(text); } catch { return text; }
};

// ─── Expandable Result Item ────────────────────────────────────────
const ExpandableResultItem = ({ result, index }) => {
  const [expanded, setExpanded] = useState(false);
  const isLongContent = result.content && result.content.length > 300;
  return (
    <ListItem divider sx={{
      flexDirection: 'column', alignItems: 'flex-start',
      cursor: isLongContent ? 'pointer' : 'default',
      '&:hover': isLongContent ? { bgcolor: 'action.hover' } : {},
    }}
      onClick={() => isLongContent && setExpanded(!expanded)}
    >
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%', mb: 1 }}>
        <Typography variant="subtitle1" sx={{ fontWeight: 'medium' }}>Result {index + 1}</Typography>
        {isLongContent && (
          <IconButton size="small" onClick={(e) => { e.stopPropagation(); setExpanded(!expanded); }}>
            {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
          </IconButton>
        )}
      </Box>
      {!expanded && (
        <Typography variant="body2" sx={{ mb: 2, color: 'text.secondary', whiteSpace: 'pre-wrap' }}>
          {isLongContent ? `${result.content.substring(0, 300)}...` : result.content}
          {isLongContent && <Typography component="span" variant="body2" color="primary" sx={{ ml: 1, fontWeight: 'medium' }}>(Click to expand)</Typography>}
        </Typography>
      )}
      <Collapse in={expanded} timeout="auto" unmountOnExit sx={{ width: '100%' }}>
        <Box sx={{ mb: 2, p: 2, bgcolor: 'background.paper', borderRadius: 1, border: '1px solid', borderColor: 'divider', maxHeight: 400, overflowY: 'auto' }}>
          <Typography variant="body2" sx={{ color: 'text.secondary', whiteSpace: 'pre-wrap' }}>{result.content}</Typography>
        </Box>
      </Collapse>
      {result.metadata && (
        <Box sx={{ width: '100%', mt: 1 }} onClick={(e) => e.stopPropagation()}>
          <ReactJson src={result.metadata} name="Metadata" collapsed={true} displayDataTypes={false}
            theme={{ base00: 'transparent', base01: '#0a0f1a', base02: '#060a12', base03: '#94a3b8', base04: '#3b82f6', base05: '#e2e8f0', base06: '#60a5fa', base07: '#e2e8f0', base08: '#38bdf8', base09: '#38bdf8', base0A: '#60a5fa', base0B: '#00ff88', base0C: '#3b82f6', base0D: '#3b82f6', base0E: '#60a5fa', base0F: '#94a3b8' }}
            style={{ backgroundColor: 'transparent', fontSize: '0.875rem' }}
            iconStyle="circle" enableClipboard={false}
          />
        </Box>
      )}
    </ListItem>
  );
};

// ─── Main RAG Manager ──────────────────────────────────────────────
const RAGManager = ({ suppressModuleHeader = false }) => {
  const PROTECTED_COLLECTIONS = ['system_help'];
  const queryClient = useQueryClient();
  const { data: collections } = useQuery('collections', api.getRAGCollections, { staleTime: 5 * 60 * 1000 });

  // Add Data dialog
  const [openAddDialog, setOpenAddDialog] = useState(false);
  const [addTab, setAddTab] = useState(0); // 0=Manual, 1=File, 2=Crawl, 3=Request, 4=Gather

  // --- Manual / File state ---
  const [formData, setFormData] = useState({ name: '', description: '', format: 'json', content: '', tags: [], metadata: {} });
  const [selectedFile, setSelectedFile] = useState(null);
  const [fileError, setFileError] = useState('');
  const [importMode, setImportMode] = useState('manual');

  // --- Crawl tab state ---
  const [crawlUrl, setCrawlUrl] = useState('');
  const [crawlUseJs, setCrawlUseJs] = useState(false);
  const [crawlCollection, setCrawlCollection] = useState('');
  const [crawlDescription, setCrawlDescription] = useState('');
  const [crawlLlmProvider, setCrawlLlmProvider] = useState('');
  const [crawlModel, setCrawlModel] = useState('');
  const [crawlResult, setCrawlResult] = useState(null);
  const [crawlError, setCrawlError] = useState('');

  // --- Request tab state ---
  const [reqMethod, setReqMethod] = useState('GET');
  const [reqUrl, setReqUrl] = useState('');
  const [reqHeaders, setReqHeaders] = useState('');
  const [reqBody, setReqBody] = useState('');
  const [reqCollection, setReqCollection] = useState('');
  const [reqDescription, setReqDescription] = useState('');
  const [reqResult, setReqResult] = useState(null);
  const [reqError, setReqError] = useState('');

  // --- Gather tab state ---
  const [gatherPrompt, setGatherPrompt] = useState('');
  const [gatherMaxIterations, setGatherMaxIterations] = useState(10);
  const [gatherLlmProvider, setGatherLlmProvider] = useState('');
  const [gatherModel, setGatherModel] = useState('');
  const [gatherResult, setGatherResult] = useState(null);
  const [gatherError, setGatherError] = useState('');
  const [gatherCollection, setGatherCollection] = useState('');
  const [gatherDocTitle, setGatherDocTitle] = useState('');
  const [gatherSuggestingTitle, setGatherSuggestingTitle] = useState(false);

  // Query dialog
  const [openQueryDialog, setOpenQueryDialog] = useState(false);
  const [selectedCollection, setSelectedCollection] = useState('');
  const [queryText, setQueryText] = useState('');
  const [queryResults, setQueryResults] = useState([]);
  const [queryError, setQueryError] = useState('');

  // Delete dialog
  const [deleteConfirmDialog, setDeleteConfirmDialog] = useState({ open: false, collectionName: null });

  const resetAddDialog = () => {
    setAddTab(0);
    setImportMode('manual');
    setSelectedFile(null);
    setFileError('');
    setFormData({ name: '', description: '', format: 'json', content: '', tags: [], metadata: {} });
    setCrawlUrl(''); setCrawlUseJs(false); setCrawlCollection(''); setCrawlDescription(''); setCrawlLlmProvider(''); setCrawlModel('');
    setCrawlResult(null); setCrawlError('');
    setReqMethod('GET'); setReqUrl(''); setReqHeaders(''); setReqBody('');
    setReqCollection(''); setReqDescription(''); setReqResult(null); setReqError('');
    setGatherPrompt(''); setGatherMaxIterations(10); setGatherLlmProvider(''); setGatherModel('');
    setGatherResult(null); setGatherError(''); setGatherCollection(''); setGatherDocTitle('');
  };

  // ── Mutations ──
  const addDataMutation = useMutation(api.addRAGData, {
    onSuccess: () => { queryClient.invalidateQueries('collections'); setOpenAddDialog(false); resetAddDialog(); },
  });
  const deleteCollectionMutation = useMutation(api.deleteRAGCollection, {
    onSuccess: () => { queryClient.invalidateQueries('collections'); setDeleteConfirmDialog({ open: false, collectionName: null }); },
  });
  const gatherMutation = useMutation((payload) => api.gatherData(payload), {
    onSuccess: (data) => { setGatherResult(data); setGatherError(null); },
    onError: (err) => { setGatherError(err.response?.data?.detail || err.message || 'Gathering failed'); setGatherResult(null); },
  });

  // ── Handlers ──
  const handleAddData = () => {
    addDataMutation.mutate({ collection_name: formData.name, data_input: formData });
  };

  const handleQueryCollection = async () => {
    try {
      setQueryError('');
      const results = await api.queryRAGCollection(selectedCollection, queryText);
      setQueryResults(results.results || []);
    } catch (error) {
      setQueryError('Failed to query collection.');
    }
  };

  const handleDeleteCollection = (collectionName) => setDeleteConfirmDialog({ open: true, collectionName });
  const handleDeleteConfirm = () => deleteCollectionMutation.mutate(deleteConfirmDialog.collectionName);

  const handleFileSelect = async (event) => {
    const file = event.target.files[0];
    if (!file) return;
    setSelectedFile(file);
    setFileError('');
    const fileName = file.name.toLowerCase();
    let detectedFormat = 'txt';
    if (fileName.endsWith('.json')) detectedFormat = 'json';
    else if (fileName.endsWith('.csv')) detectedFormat = 'csv';
    if (!formData.name.trim()) {
      const nameWithoutExt = file.name.replace(/\.[^/.]+$/, '').replace(/[^a-zA-Z0-9_-]/g, '_');
      setFormData(prev => ({ ...prev, name: nameWithoutExt }));
    }
    try {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target.result;
        if (detectedFormat === 'json') {
          try { JSON.parse(content); setFormData(prev => ({ ...prev, format: 'json', content })); }
          catch { setFileError('Invalid JSON file.'); setSelectedFile(null); }
        } else if (detectedFormat === 'csv') {
          setFormData(prev => ({ ...prev, format: 'csv', content }));
        } else {
          setFormData(prev => ({ ...prev, format: 'txt', content }));
        }
      };
      reader.onerror = () => { setFileError('Error reading file.'); setSelectedFile(null); };
      reader.readAsText(file);
    } catch { setFileError('Error processing file.'); setSelectedFile(null); }
  };

  const handleClearFile = () => { setSelectedFile(null); setFileError(''); setFormData(prev => ({ ...prev, content: '' })); };

  // ── Crawl handler ──
  const handleCrawl = async () => {
    if (!crawlUrl.trim()) return;
    setCrawlError('');
    setCrawlResult(null);
    try {
      const res = await api.crawlWebsite({ url: crawlUrl.trim(), use_js: crawlUseJs, llm_provider: crawlLlmProvider || null, model: crawlModel || null });
      const collection = res?.collection_name || crawlCollection.trim() || `crawl_${Date.now()}`;
      setCrawlResult({ success: true, collection, content_length: res?.extracted_data ? JSON.stringify(res.extracted_data).length : 0 });
      queryClient.invalidateQueries('collections');
    } catch (err) {
      setCrawlError(err.response?.data?.detail || err.message || 'Crawl failed');
    }
  };

  // ── Request handler ──
  const handleSendRequest = async () => {
    if (!reqUrl.trim()) return;
    setReqError('');
    setReqResult(null);
    try {
      const headers = reqHeaders.trim() ? parseJsonOrEmpty(reqHeaders) : {};
      const body = reqBody.trim() ? tryParseJson(reqBody) : null;
      const res = await api.executeRequestOneShot({ url: reqUrl.trim(), method: reqMethod, headers, body });
      const content = typeof res === 'string' ? res : JSON.stringify(res, null, 2);
      const collection = reqCollection.trim() || `request_${Date.now()}`;
      await api.addRAGData({
        collection_name: collection,
        data_input: { name: `API: ${reqMethod} ${reqUrl}`, description: reqDescription || undefined, format: 'txt', content, tags: ['api-request'] },
      });
      setReqResult({ success: true, collection, content_length: content.length });
      queryClient.invalidateQueries('collections');
    } catch (err) {
      setReqError(err.response?.data?.detail || err.message || 'Request failed');
    }
  };

  // ── Gather handlers ──
  const handleGather = () => {
    if (!gatherPrompt.trim()) { setGatherError('Enter a topic.'); return; }
    setGatherError(null);
    gatherMutation.mutate({ prompt: gatherPrompt.trim(), max_iterations: gatherMaxIterations, llm_provider: gatherLlmProvider || undefined, model_name: gatherModel || undefined });
  };

  const handleSaveGatherToRag = async () => {
    const content = gatherResult?.content || '';
    const title = gatherDocTitle.trim() || `Gathering: ${gatherPrompt.slice(0, 60)}`;
    const collection = gatherCollection.trim() || `gathering_${Date.now()}`;
    if (!content) return;
    try {
      await api.addRAGData({
        collection_name: collection,
        data_input: { name: title, description: `AI gathered research on: ${gatherPrompt}`, format: 'txt', content, tags: ['gathering'] },
      });
      queryClient.invalidateQueries('collections');
      setOpenAddDialog(false);
      resetAddDialog();
    } catch (err) {
      setGatherError(err.response?.data?.detail || err.message || 'Failed to save to RAG');
    }
  };

  const handleSuggestTitle = async () => {
    const content = gatherResult?.content || '';
    if (!content) return;
    setGatherSuggestingTitle(true);
    try { const res = await api.suggestRAGTitle(content); setGatherDocTitle(res?.title || ''); }
    catch { /* ignore */ }
    finally { setGatherSuggestingTitle(false); }
  };

  // ── Renders ──
  const { data: providersData = { providers: [] } } = useQuery('providers', api.getProviders, { retry: false });
  const providers = providersData?.providers || [];
  const { data: modelsData } = useQuery('models', api.getModels, { retry: false });
  const models = Array.isArray(modelsData) ? modelsData : [];

  // ── Manual Entry tab ──
  const renderManualTab = () => (
    <Grid container spacing={3}>
      <Grid item xs={12}><Typography variant="h6" gutterBottom color="primary">Collection Details</Typography></Grid>
      <Grid item xs={12} sm={6}>
        <TextField fullWidth label="Collection Name" value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value.replace(/[^a-zA-Z0-9_-]/g, '_') })}
          required error={!formData.name.trim()}
          helperText={!formData.name.trim() ? 'Collection name is required' : 'Spaces become underscores'} />
      </Grid>
      <Grid item xs={12} sm={6}>
        <FormControl fullWidth>
          <InputLabel>Data Format</InputLabel>
          <Select value={formData.format} onChange={(e) => setFormData({ ...formData, format: e.target.value })}>
            <MenuItem value="json">JSON</MenuItem>
            <MenuItem value="csv">CSV</MenuItem>
            <MenuItem value="txt">Text</MenuItem>
          </Select>
        </FormControl>
      </Grid>
      <Grid item xs={12}>
        <TextField fullWidth label="Description" value={formData.description}
          onChange={(e) => setFormData({ ...formData, description: e.target.value })} multiline rows={2} />
      </Grid>
      <Grid item xs={12}>
        <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
          <Button size="small" variant={importMode === 'manual' ? 'contained' : 'outlined'}
            onClick={() => { setImportMode('manual'); setSelectedFile(null); setFileError(''); }}>Manual Input</Button>
          <Button size="small" variant={importMode === 'file' ? 'contained' : 'outlined'} startIcon={<UploadFileIcon />} component="label">
            Import File
            <input type="file" hidden accept=".txt,.json,.csv" onChange={handleFileSelect} />
          </Button>
        </Box>
      </Grid>
      {importMode === 'file' && (
        <Grid item xs={12}>
          <Box sx={{ mb: 2 }}>
            {selectedFile ? (
              <Box sx={{ p: 2, border: '1px solid', borderColor: 'divider', borderRadius: 1, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Box>
                  <Typography variant="body1" sx={{ fontWeight: 'medium' }}>{selectedFile.name}</Typography>
                  <Typography variant="body2" color="text.secondary">{(selectedFile.size / 1024).toFixed(2)} KB</Typography>
                </Box>
                <Button size="small" onClick={handleClearFile}>Clear</Button>
              </Box>
            ) : (
              <Box sx={{ p: 3, border: '2px dashed', borderColor: 'divider', borderRadius: 1, textAlign: 'center', cursor: 'pointer', '&:hover': { borderColor: 'primary.main', bgcolor: 'action.hover' } }} component="label">
                <input type="file" hidden accept=".txt,.json,.csv" onChange={handleFileSelect} />
                <UploadFileIcon sx={{ fontSize: 48, color: 'text.secondary', mb: 1 }} />
                <Typography variant="body1" sx={{ mb: 1 }}>Click to upload or drag and drop</Typography>
                <Typography variant="body2" color="text.secondary">Supports: TXT, JSON, CSV files</Typography>
              </Box>
            )}
            {fileError && <Alert severity="error" sx={{ mt: 1 }}>{fileError}</Alert>}
          </Box>
        </Grid>
      )}
      <Grid item xs={12}>
        <TextField fullWidth multiline rows={10} label="Content" value={formData.content}
          onChange={(e) => setFormData({ ...formData, content: e.target.value })}
          placeholder={importMode === 'file' ? 'Select a file to import or enter content manually...' : 'Enter your data here...'}
          required error={!formData.content.trim()}
          disabled={importMode === 'file' && !selectedFile && !formData.content.trim()} />
      </Grid>
      <Grid item xs={12}>
        <Typography variant="h6" gutterBottom color="primary">Metadata (Optional)</Typography>
        <TextField fullWidth label="Tags (comma-separated)" value={formData.tags.join(', ')}
          onChange={(e) => setFormData({ ...formData, tags: e.target.value.split(',').map(t => t.trim()).filter(Boolean) })}
          placeholder="e.g. science, research" />
      </Grid>
    </Grid>
  );

  // ── Web Crawl tab ──
  const renderCrawlTab = () => (
    <Box>
      <Typography variant="h6" gutterBottom color="primary">Web Crawl</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        Enter a URL to crawl. Content is extracted and saved directly to a RAG collection.
      </Typography>
      <TextField fullWidth label="URL" value={crawlUrl} onChange={(e) => setCrawlUrl(e.target.value)}
        required placeholder="https://example.com" sx={{ mb: 2 }} />
      <FormControlLabel control={<Switch checked={crawlUseJs} onChange={(e) => setCrawlUseJs(e.target.checked)} />}
        label="Use JavaScript Rendering" sx={{ mb: 2, display: 'block' }} />
      <TextField fullWidth label="Collection Name" value={crawlCollection}
        onChange={(e) => setCrawlCollection(e.target.value)} sx={{ mb: 2 }}
        helperText="Leave empty to auto-generate" />
      <TextField fullWidth label="Description" value={crawlDescription}
        onChange={(e) => setCrawlDescription(e.target.value)} multiline rows={2} sx={{ mb: 2 }} />
      <Divider sx={{ my: 2 }} />
      <Typography variant="subtitle2" color="text.secondary" gutterBottom>LLM Settings (Optional)</Typography>
      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid item xs={6}>
          <FormControl fullWidth size="small">
            <InputLabel>Provider</InputLabel>
            <Select value={crawlLlmProvider} label="Provider" onChange={(e) => setCrawlLlmProvider(e.target.value)}>
              <MenuItem value="">Default</MenuItem>
              {providers.map(p => <MenuItem key={p} value={p}>{p}</MenuItem>)}
            </Select>
          </FormControl>
        </Grid>
        <Grid item xs={6}>
          <TextField fullWidth size="small" label="Model" value={crawlModel} onChange={(e) => setCrawlModel(e.target.value)} />
        </Grid>
      </Grid>
      {crawlError && <Alert severity="error" sx={{ mb: 2 }}>{crawlError}</Alert>}
      {crawlResult && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Crawled and saved to collection "{crawlResult.collection}" ({crawlResult.content_length} chars).
        </Alert>
      )}
      <Button fullWidth variant="contained" onClick={handleCrawl} disabled={!crawlUrl.trim()} startIcon={<WebIcon />}>
        Crawl & Save to RAG
      </Button>
    </Box>
  );

  // ── API Request tab ──
  const renderRequestTab = () => {
    const methodColors = { GET: 'success', POST: 'primary', PUT: 'warning', DELETE: 'error', PATCH: 'info' };
    return (
      <Box>
        <Typography variant="h6" gutterBottom color="primary">API Request</Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Send an HTTP request and save the response to a RAG collection.
        </Typography>
        <Grid container spacing={2} sx={{ mb: 2 }}>
          <Grid item xs={3}>
            <FormControl fullWidth size="small">
              <InputLabel>Method</InputLabel>
              <Select value={reqMethod} label="Method" onChange={(e) => setReqMethod(e.target.value)}>
                {['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map(m => (
                  <MenuItem key={m} value={m}>
                    <Chip label={m} size="small" color={methodColors[m]} />
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={9}>
            <TextField fullWidth size="small" label="URL" value={reqUrl}
              onChange={(e) => setReqUrl(e.target.value)} placeholder="https://api.example.com/data" />
          </Grid>
        </Grid>
        <TextField fullWidth label="Headers (JSON)" value={reqHeaders}
          onChange={(e) => setReqHeaders(e.target.value)} multiline rows={3} sx={{ mb: 2 }}
          placeholder='{"Authorization": "Bearer token"}' />
        {['POST', 'PUT', 'PATCH'].includes(reqMethod) && (
          <TextField fullWidth label="Body" value={reqBody}
            onChange={(e) => setReqBody(e.target.value)} multiline rows={4} sx={{ mb: 2 }}
            placeholder='{"key": "value"}' />
        )}
        <TextField fullWidth label="Collection Name" value={reqCollection}
          onChange={(e) => setReqCollection(e.target.value)} sx={{ mb: 2 }} helperText="Leave empty to auto-generate" />
        <TextField fullWidth label="Description" value={reqDescription}
          onChange={(e) => setReqDescription(e.target.value)} multiline rows={2} sx={{ mb: 2 }} />
        {reqError && <Alert severity="error" sx={{ mb: 2 }}>{reqError}</Alert>}
        {reqResult && (
          <Alert severity="success" sx={{ mb: 2 }}>
            Response saved to collection "{reqResult.collection}" ({reqResult.content_length} chars).
          </Alert>
        )}
        <Button fullWidth variant="contained" onClick={handleSendRequest} disabled={!reqUrl.trim()} startIcon={<ApiIcon />}>
          Send & Save to RAG
        </Button>
      </Box>
    );
  };

  // ── Research Gather tab ──
  const renderGatherTab = () => (
    <Box>
      <Typography variant="h6" gutterBottom color="primary">Research Gather</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        AI-powered research from Wikipedia, Reddit, and web search. Results save to a RAG collection.
      </Typography>
      <TextField fullWidth multiline rows={4} label="Topic or question" value={gatherPrompt}
        onChange={(e) => setGatherPrompt(e.target.value)} placeholder="e.g. Best practices for learning Python"
        sx={{ mb: 2 }} />
      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid item xs={4}>
          <FormControl fullWidth size="small">
            <InputLabel>Max iterations</InputLabel>
            <Select value={gatherMaxIterations} label="Max iterations" onChange={(e) => setGatherMaxIterations(Number(e.target.value))}>
              {[5, 8, 10, 12, 15, 20].map(n => <MenuItem key={n} value={n}>{n}</MenuItem>)}
            </Select>
          </FormControl>
        </Grid>
        <Grid item xs={4}>
          <FormControl fullWidth size="small">
            <InputLabel>Provider</InputLabel>
            <Select value={gatherLlmProvider} label="Provider" onChange={(e) => { setGatherLlmProvider(e.target.value); setGatherModel(''); }}>
              <MenuItem value="">Default</MenuItem>
              {providers.map(p => <MenuItem key={p} value={p}>{p}</MenuItem>)}
            </Select>
          </FormControl>
        </Grid>
        <Grid item xs={4}>
          <FormControl fullWidth size="small">
            <InputLabel>Model</InputLabel>
            <Select value={gatherModel} label="Model" onChange={(e) => setGatherModel(e.target.value)}>
              <MenuItem value="">Default</MenuItem>
              {models.map(m => <MenuItem key={m.name || m} value={m.name || m}>{m.name || m}</MenuItem>)}
            </Select>
          </FormControl>
        </Grid>
      </Grid>
      <Button fullWidth variant="contained" onClick={handleGather} disabled={gatherMutation.isLoading || !gatherPrompt.trim()}
        startIcon={gatherMutation.isLoading ? <CircularProgress size={20} color="inherit" /> : <SearchIcon />} sx={{ mb: 2 }}>
        {gatherMutation.isLoading ? 'Gathering...' : 'Gather'}
      </Button>

      {gatherError && <Alert severity="error" sx={{ mb: 2 }}>{gatherError}</Alert>}

      {gatherResult && gatherResult.success && (
        <>
          <Divider sx={{ my: 2 }} />
          <Typography variant="h6" gutterBottom color="primary">Save to RAG</Typography>
          <Paper variant="outlined" sx={{ p: 2, mb: 2, maxHeight: 300, overflow: 'auto', bgcolor: 'action.hover' }}>
            <Box className="markdown-body" sx={{
              '& h1, & h2, & h3': { mt: 1.5, mb: 0.5, fontWeight: 600 },
              '& p': { mb: 1 }, '& ul, & ol': { pl: 2.5, mb: 1 },
            }}>
              <ReactMarkdown>{gatherResult.content || 'No content.'}</ReactMarkdown>
            </Box>
          </Paper>
          <TextField fullWidth label="Collection Name" value={gatherCollection}
            onChange={(e) => setGatherCollection(e.target.value)} sx={{ mb: 2 }} helperText="Leave empty to auto-generate" />
          <Box sx={{ display: 'flex', gap: 2, alignItems: 'flex-start', mb: 2 }}>
            <TextField fullWidth label="Document title" value={gatherDocTitle}
              onChange={(e) => setGatherDocTitle(e.target.value)} disabled={gatherSuggestingTitle}
              sx={{ flex: 1 }} />
            <Button variant="outlined" size="small" onClick={handleSuggestTitle}
              disabled={gatherSuggestingTitle || !gatherResult?.content}
              sx={{ mt: 1, minWidth: 48 }} title="Suggest title with AI">
              {gatherSuggestingTitle ? <CircularProgress size={24} /> : <SuggestIcon />}
            </Button>
          </Box>
          <Button fullWidth variant="contained" color="success" onClick={handleSaveGatherToRag}
            disabled={!gatherDocTitle.trim() || !gatherCollection.trim()} startIcon={<SaveIcon />}>
            Save to RAG
          </Button>
        </>
      )}
    </Box>
  );

  return (
    <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', width: '100%' }}>
      {!suppressModuleHeader ? (
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1.25, flexShrink: 0 }}>
          <Typography variant="h4">RAG Manager</Typography>
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={() => { resetAddDialog(); setOpenAddDialog(true); }}>
            Add Data
          </Button>
        </Box>
      ) : (
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 0.75, flexShrink: 0 }}>
          <Button variant="contained" size="small" startIcon={<AddIcon />} onClick={() => { resetAddDialog(); setOpenAddDialog(true); }}>
            Add Data
          </Button>
        </Box>
      )}

      {/* Collections Grid */}
      <Box sx={{ flex: 1, minHeight: 0, overflow: 'auto' }}>
        <Grid container spacing={1.25}>
          {collections?.map((collection) => {
            const isProtected = !!collection?.metadata?.protected;
            return (
              <Grid item xs={12} sm={6} md={4} lg={3} key={collection.name}>
                <Card sx={{ height: '100%', minHeight: 152, display: 'flex', flexDirection: 'column', boxShadow: 1 }}>
                  <CardContent sx={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0, py: 1.25 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1, gap: 0.75, minWidth: 0 }}>
                      <Box sx={{ flex: 1, minWidth: 0 }}>
                        <Typography variant="subtitle2" sx={{ mb: 0.5, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', pr: 0.5 }}
                          title={collection.name}>{collection.name}</Typography>
                        {isProtected && (
                          <Typography variant="caption" color="warning.main" sx={{ display: 'flex', alignItems: 'center', gap: 0.35, mb: 0.25, fontSize: '0.625rem' }}>
                            <LockIcon sx={{ fontSize: 12 }} /> Protected collection
                          </Typography>
                        )}
                        <Typography variant="body2" color="text.secondary" sx={{ mb: 0.5, fontSize: '0.7rem' }}>
                          {collection.count} documents
                        </Typography>
                        {collection.metadata?.description && (
                          <Typography variant="body2" sx={{ overflow: 'hidden', textOverflow: 'ellipsis', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', minHeight: '2.6em', maxHeight: '2.6em', lineHeight: 1.35, fontSize: '0.7rem', color: 'text.secondary' }}
                            title={collection.metadata.description}>{collection.metadata.description}</Typography>
                        )}
                      </Box>
                      <IconButton size="small" color="error" onClick={() => handleDeleteCollection(collection.name)}
                        disabled={isProtected} sx={{ bgcolor: 'error.light', '&:hover': { bgcolor: 'error.main', color: 'white' }, flexShrink: 0, ml: 'auto', p: 0.35 }}>
                        <DeleteIcon sx={{ fontSize: 18 }} />
                      </IconButton>
                    </Box>
                    <Button size="small" variant="outlined" startIcon={<SearchIcon />}
                      onClick={() => { setSelectedCollection(collection.name); setOpenQueryDialog(true); }}
                      fullWidth sx={{ mt: 'auto', py: 0.25, fontSize: '0.65rem' }}>
                      Query Collection
                    </Button>
                  </CardContent>
                </Card>
              </Grid>
            );
          })}
        </Grid>
      </Box>

      {/* ── Add Data Dialog ── */}
      <Dialog open={openAddDialog} onClose={() => { setOpenAddDialog(false); resetAddDialog(); }} maxWidth="md" fullWidth>
        <DialogTitle sx={{ pb: 0 }}>
          <Tabs value={addTab} onChange={(_, v) => {
            setAddTab(v);
            if (v === 1) { setImportMode('file'); setSelectedFile(null); setFileError(''); }
            if (v === 0) { setImportMode('manual'); }
          }} variant="scrollable" scrollButtons="auto">
            <Tab label="Manual Entry" />
            <Tab label="File Upload" />
            <Tab label="Web Crawl" icon={<WebIcon />} iconPosition="start" />
            <Tab label="API Request" icon={<ApiIcon />} iconPosition="start" />
            <Tab label="Research" icon={<SearchIcon />} iconPosition="start" />
          </Tabs>
        </DialogTitle>
        <DialogContent sx={{ pt: 2 }}>
          {addDataMutation.isError && (
            <Alert severity="error" sx={{ mb: 2 }}>Failed to add data. Please check your inputs.</Alert>
          )}
          {PROTECTED_COLLECTIONS.includes((formData.name || '').trim()) && (
            <Alert severity="warning" sx={{ mb: 2 }}>Collection "{formData.name}" is protected.</Alert>
          )}

          {addTab === 0 && renderManualTab()}
          {addTab === 1 && renderManualTab()} {/* same as manual, just pre-selects file mode */}
          {addTab === 2 && renderCrawlTab()}
          {addTab === 3 && renderRequestTab()}
          {addTab === 4 && renderGatherTab()}
        </DialogContent>
        {addTab <= 1 && (
          <DialogActions sx={{ p: 3, pt: 1 }}>
            <Button onClick={() => { setOpenAddDialog(false); resetAddDialog(); }}>Cancel</Button>
            <Button onClick={handleAddData} variant="contained"
              disabled={addDataMutation.isLoading || !formData.name.trim() || !formData.content.trim() || PROTECTED_COLLECTIONS.includes((formData.name || '').trim())}>
              {addDataMutation.isLoading ? 'Adding...' : 'Add Data'}
            </Button>
          </DialogActions>
        )}
      </Dialog>

      {/* ── Query Dialog ── */}
      <Dialog open={openQueryDialog} onClose={() => setOpenQueryDialog(false)} maxWidth="lg" fullWidth>
        <DialogTitle sx={{ pb: 1 }}>Query Collection: {selectedCollection}</DialogTitle>
        <DialogContent sx={{ pt: 1 }}>
          {queryError && <Alert severity="error" sx={{ mb: 2 }}>{queryError}</Alert>}
          <Box sx={{ mb: 3 }}>
            <TextField fullWidth label="Search Query" value={queryText} onChange={(e) => setQueryText(e.target.value)}
              multiline rows={3} placeholder="Enter your search query..." sx={{ mb: 2 }} />
            <Button variant="contained" onClick={handleQueryCollection} size="large" fullWidth disabled={!queryText.trim()}>
              Search Collection
            </Button>
          </Box>
          {queryResults.length > 0 && (
            <Box>
              <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <SearchIcon color="primary" /> Results ({queryResults.length})
              </Typography>
              <List sx={{ maxHeight: 500, overflowY: 'auto' }}>
                {queryResults.map((result, index) => (
                  <ExpandableResultItem key={index} result={result} index={index} />
                ))}
              </List>
            </Box>
          )}
        </DialogContent>
        <DialogActions sx={{ p: 3, pt: 1 }}>
          <Button onClick={() => setOpenQueryDialog(false)}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* ── Delete Confirmation ── */}
      <Dialog open={deleteConfirmDialog.open} onClose={() => setDeleteConfirmDialog({ open: false, collectionName: null })}>
        <DialogTitle>Confirm Delete</DialogTitle>
        <DialogContent>
          <Typography>Are you sure you want to delete collection "{deleteConfirmDialog.collectionName}"? This cannot be undone.</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteConfirmDialog({ open: false, collectionName: null })}>Cancel</Button>
          <Button onClick={handleDeleteConfirm} color="error" variant="contained">Delete</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default RAGManager;