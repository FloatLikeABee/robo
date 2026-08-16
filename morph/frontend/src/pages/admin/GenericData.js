import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
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
  Drawer,
  IconButton,
  LinearProgress,
  Paper,
  Stack,
  Tab,
  Tabs,
  TextField,
  Tooltip,
  Typography,
  alpha,
  useTheme,
} from '@mui/material';
import { DataGrid } from '@mui/x-data-grid';
import AutoAwesomeOutlinedIcon from '@mui/icons-material/AutoAwesomeOutlined';
import CloseIcon from '@mui/icons-material/Close';
import CloudUploadOutlinedIcon from '@mui/icons-material/CloudUploadOutlined';
import DataObjectOutlinedIcon from '@mui/icons-material/DataObjectOutlined';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import DescriptionOutlinedIcon from '@mui/icons-material/DescriptionOutlined';
import InsertDriveFileOutlinedIcon from '@mui/icons-material/InsertDriveFileOutlined';
import PictureAsPdfOutlinedIcon from '@mui/icons-material/PictureAsPdfOutlined';
import TableChartOutlinedIcon from '@mui/icons-material/TableChartOutlined';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import axios from 'axios';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { API_BASE_URL as API_BASE } from '../../apiBase';
import { getMorphToken } from '../../auth/morphSession';
import { RecordSheetJsonDetail } from '../../components/admin/jsonDetailViews';
import { usePlatformUi } from '../../PlatformUiContext';
import { useConfirm } from '../../components/ConfirmDialog';
import {
  ADMIN_RIGHT_PANEL_TOP,
  adminRightPanelHeightCalc,
} from '../../components/admin/adminRightPanelStyle';

const markdownArticleSx = {
  '& h1, & h2, & h3, & h4': { mt: 1.5, mb: 0.75, fontWeight: 700, lineHeight: 1.25 },
  '& h1': { fontSize: '1.75rem', letterSpacing: '-0.02em' },
  '& h2': { fontSize: '1.25rem' },
  '& h3': { fontSize: '1.05rem' },
  '& p': { my: 0.85, lineHeight: 1.65 },
  '& ul, & ol': { my: 0.85, pl: 2.5 },
  '& li': { my: 0.35 },
  '& blockquote': {
    my: 1.25,
    pl: 2,
    borderLeft: 4,
    borderColor: 'primary.main',
    color: 'text.secondary',
    fontStyle: 'italic',
  },
  '& code': {
    fontFamily: 'ui-monospace, monospace',
    fontSize: '0.9em',
    px: 0.5,
    py: 0.15,
    borderRadius: 0.75,
    bgcolor: 'action.hover',
  },
  '& pre': { my: 1, p: 1.5, borderRadius: 1.5, overflow: 'auto', bgcolor: 'action.hover' },
  '& table': { width: '100%', borderCollapse: 'collapse', my: 1.25 },
  '& th, & td': { border: 1, borderColor: 'divider', px: 1, py: 0.5, fontSize: '0.875rem' },
  '& th': { bgcolor: 'action.hover', fontWeight: 600 },
  '& hr': { my: 2, borderColor: 'divider' },
  '& a': { color: 'primary.main' },
};

function parseDetail(raw) {
  if (raw == null || raw === '') return null;
  if (typeof raw === 'object') return raw;
  try {
    return JSON.parse(String(raw));
  } catch {
    return null;
  }
}

function sourceTypeIcon(type) {
  switch (type) {
    case 'csv':
      return <TableChartOutlinedIcon fontSize="small" />;
    case 'pdf':
      return <PictureAsPdfOutlinedIcon fontSize="small" />;
    case 'json':
      return <DataObjectOutlinedIcon fontSize="small" />;
    default:
      return <InsertDriveFileOutlinedIcon fontSize="small" />;
  }
}

function sourceTypeColor(type) {
  switch (type) {
    case 'csv':
      return 'success';
    case 'pdf':
      return 'error';
    case 'json':
      return 'info';
    default:
      return 'default';
  }
}

function formatWhen(v) {
  if (!v) return '—';
  const d = new Date(v);
  if (Number.isNaN(d.getTime())) return String(v);
  return d.toLocaleString([], { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function GenericDataContentView({ record }) {
  const detail = parseDetail(record?.detail);
  const sourceType = record?.source_type;

  if (!detail) {
    return (
      <Typography variant="body2" color="text.secondary">
        No imported content yet.
      </Typography>
    );
  }

  if (sourceType === 'pdf') {
    const md = detail.content_markdown || '';
    return (
      <Paper
        elevation={0}
        sx={{
          p: { xs: 2, sm: 3 },
          borderRadius: 2,
          border: 1,
          borderColor: 'divider',
          bgcolor: 'background.paper',
          maxWidth: 820,
          mx: 'auto',
        }}
      >
        {detail.article_title && (
          <Typography variant="overline" color="text.secondary" sx={{ letterSpacing: 1.2 }}>
            Article
          </Typography>
        )}
        <Box sx={markdownArticleSx}>
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{md || '_Empty document_'}</ReactMarkdown>
        </Box>
      </Paper>
    );
  }

  if (sourceType === 'csv') {
    const columns = (detail.columns || []).map((col, i) => ({
      field: `c${i}`,
      headerName: String(col),
      flex: 1,
      minWidth: 120,
    }));
    const colKeys = detail.columns || [];
    const rows = (detail.rows || []).map((row, idx) => {
      const out = { id: idx + 1 };
      colKeys.forEach((col, i) => {
        out[`c${i}`] = row?.[col] ?? '';
      });
      return out;
    });
    const truncated = detail.import_meta?.truncated;

    return (
      <Stack spacing={1.5}>
        <Stack direction="row" spacing={1} flexWrap="wrap" alignItems="center">
          <Chip size="small" label={`${rows.length} rows`} color="success" variant="outlined" />
          <Chip size="small" label={`${colKeys.length} columns`} variant="outlined" />
          {truncated && (
            <Chip size="small" color="warning" label={`Truncated at ${detail.import_meta?.max_rows || 'limit'}`} />
          )}
        </Stack>
        <Paper variant="outlined" sx={{ height: 420, width: '100%' }}>
          <DataGrid
            rows={rows}
            columns={columns}
            density="compact"
            disableRowSelectionOnClick
            pageSizeOptions={[25, 50, 100]}
            initialState={{ pagination: { paginationModel: { pageSize: 25 } } }}
            sx={{ border: 0 }}
          />
        </Paper>
        <Typography variant="caption" color="text.secondary">
          Structured data is stored in MongoDB and shown here as a searchable table.
        </Typography>
      </Stack>
    );
  }

  const payload = detail.payload ?? detail;
  return (
    <Stack spacing={1.5} sx={{ width: '100%', minWidth: 0 }}>
      <Chip
        size="small"
        icon={<DataObjectOutlinedIcon />}
        label="JSON structure"
        color="info"
        variant="outlined"
        sx={{ alignSelf: 'flex-start' }}
      />
      <Paper
        variant="outlined"
        sx={{
          p: 2,
          borderRadius: 2,
          width: '100%',
          minWidth: 0,
          boxSizing: 'border-box',
        }}
      >
        <Box sx={{ width: '100%', minWidth: 0 }}>
          <RecordSheetJsonDetail raw={payload} />
        </Box>
      </Paper>
    </Stack>
  );
}

function AiAnalysisPanel({ analysis, loading, onAnalyze, hasRecord }) {
  const theme = useTheme();

  return (
    <Stack spacing={2}>
      <Paper
        elevation={0}
        sx={{
          p: 2,
          borderRadius: 2,
          border: 1,
          borderColor: alpha(theme.palette.primary.main, 0.35),
          background: `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.08)} 0%, ${alpha(
            theme.palette.secondary.main,
            0.06
          )} 100%)`,
        }}
      >
        <Stack direction="row" spacing={1.5} alignItems="flex-start">
          <Box
            sx={{
              width: 40,
              height: 40,
              borderRadius: 2,
              display: 'grid',
              placeItems: 'center',
              bgcolor: alpha(theme.palette.primary.main, 0.15),
              color: 'primary.main',
              flexShrink: 0,
            }}
          >
            <AutoAwesomeOutlinedIcon />
          </Box>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography variant="subtitle1" fontWeight={700}>
              Morph AI analysis
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.25 }}>
              Summarize patterns, flag data quality issues, and suggest next steps from your imported material.
            </Typography>
            <Button
              variant="contained"
              size="small"
              startIcon={loading ? <CircularProgress size={16} color="inherit" /> : <AutoAwesomeOutlinedIcon />}
              onClick={onAnalyze}
              disabled={!hasRecord || loading}
              sx={{ mt: 1.5, textTransform: 'none', borderRadius: 2 }}
            >
              {loading ? 'Analyzing…' : analysis ? 'Re-analyze' : 'Run AI analysis'}
            </Button>
          </Box>
        </Stack>
      </Paper>

      {loading && <LinearProgress sx={{ borderRadius: 1 }} />}

      {analysis ? (
        <Paper variant="outlined" sx={{ p: 2.5, borderRadius: 2 }}>
          <Box sx={markdownArticleSx}>
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{analysis}</ReactMarkdown>
          </Box>
        </Paper>
      ) : (
        !loading && (
          <Typography variant="body2" color="text.secondary" sx={{ fontStyle: 'italic' }}>
            No analysis yet — import a file and run Morph AI to get insights.
          </Typography>
        )
      )}
    </Stack>
  );
}

function ImportDialog({ open, onClose, onImported }) {
  const [file, setFile] = useState(null);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [dragOver, setDragOver] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const inputRef = useRef(null);

  useEffect(() => {
    if (!open) {
      setFile(null);
      setTitle('');
      setDescription('');
      setError(null);
      setDragOver(false);
    }
  }, [open]);

  const pickFile = (f) => {
    if (!f) return;
    const ext = (f.name.split('.').pop() || '').toLowerCase();
    if (!['csv', 'json', 'pdf', 'md', 'markdown'].includes(ext)) {
      setError('Use a .csv, .json, .pdf, or .md file');
      return;
    }
    setFile(f);
    setError(null);
    if (!title) setTitle(f.name.replace(/\.[^.]+$/, ''));
  };

  const submit = async () => {
    if (!file) {
      setError('Choose a file to import');
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const fd = new FormData();
      fd.append('file', file);
      if (title.trim()) fd.append('title', title.trim());
      if (description.trim()) fd.append('description', description.trim());
      const token = getMorphToken();
      const res = await axios.post(`${API_BASE}${tranEndpoints.genericDataImport}`, fd, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        timeout: 300000,
      });
      onImported(res.data);
      onClose();
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Import failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onClose={loading ? undefined : onClose} maxWidth="sm" fullWidth PaperProps={{ sx: { borderRadius: 2 } }}>
      <DialogTitle sx={{ fontWeight: 700 }}>Import generic data</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 0.5 }}>
          <Typography variant="body2" color="text.secondary">
            Upload CSV, JSON, PDF, or Markdown. PDFs are converted to markdown locally (fast, no AI). Content is stored in MongoDB for display and analysis.
          </Typography>
          <Paper
            variant="outlined"
            onDragOver={(e) => {
              e.preventDefault();
              setDragOver(true);
            }}
            onDragLeave={() => setDragOver(false)}
            onDrop={(e) => {
              e.preventDefault();
              setDragOver(false);
              pickFile(e.dataTransfer.files?.[0]);
            }}
            onClick={() => inputRef.current?.click()}
            sx={{
              p: 3,
              textAlign: 'center',
              cursor: 'pointer',
              borderStyle: 'dashed',
              borderWidth: 2,
              borderRadius: 2,
              borderColor: dragOver ? 'primary.main' : 'divider',
              bgcolor: dragOver ? 'action.hover' : 'background.default',
              transition: 'border-color 0.15s, background 0.15s',
            }}
          >
            <input
              ref={inputRef}
              type="file"
              accept=".csv,.json,.pdf,.md,.markdown,text/csv,application/json,application/pdf,text/markdown"
              hidden
              onChange={(e) => pickFile(e.target.files?.[0])}
            />
            <CloudUploadOutlinedIcon sx={{ fontSize: 40, color: 'primary.main', mb: 1 }} />
            <Typography variant="subtitle2">{file ? file.name : 'Drop a file or click to browse'}</Typography>
            <Stack direction="row" spacing={0.75} justifyContent="center" sx={{ mt: 1 }}>
              {['CSV', 'JSON', 'PDF', 'MD'].map((t) => (
                <Chip key={t} size="small" label={t} variant="outlined" />
              ))}
            </Stack>
          </Paper>
          <TextField label="Title" size="small" fullWidth value={title} onChange={(e) => setTitle(e.target.value)} />
          <TextField
            label="Description (optional)"
            size="small"
            fullWidth
            multiline
            minRows={2}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
          {error && <Alert severity="error">{error}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} disabled={loading}>
          Cancel
        </Button>
        <Button variant="contained" onClick={submit} disabled={loading || !file} sx={{ textTransform: 'none' }}>
          {loading ? 'Importing…' : 'Import'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default function GenericData() {
  const { confirm } = useConfirm();
  const theme = useTheme();
  const { labels } = usePlatformUi();
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [importOpen, setImportOpen] = useState(false);
  const [selectedId, setSelectedId] = useState(null);
  const [detail, setDetail] = useState(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailTab, setDetailTab] = useState('content');
  const [analyzeLoading, setAnalyzeLoading] = useState(false);
  const [actionError, setActionError] = useState(null);
  const [rowSelectionModel, setRowSelectionModel] = useState([]);
  const [batchDeleteOpen, setBatchDeleteOpen] = useState(false);
  const [deleteBusy, setDeleteBusy] = useState(false);

  const load = useCallback(() => {
    setLoading(true);
    return tranApi
      .get(tranEndpoints.genericData)
      .then((res) => setRows(res.data || []))
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load'))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const openDetail = useCallback(async (id) => {
    setSelectedId(id);
    setDetailTab('content');
    setActionError(null);
    setDetailLoading(true);
    try {
      const res = await tranApi.get(tranEndpoints.genericDataFull(id));
      setDetail(res.data || null);
    } catch (err) {
      setActionError(err.response?.data?.error || err.message || 'Failed to load record');
      setDetail(null);
    } finally {
      setDetailLoading(false);
    }
  }, []);

  const closeDetail = () => {
    setSelectedId(null);
    setDetail(null);
    setActionError(null);
  };

  const handleDelete = async () => {
    if (!selectedId) return;
    const ok = await confirm({
      title: 'Delete record',
      message: 'Delete this imported data record?',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    try {
      await tranApi.delete(tranEndpoints.genericDataItem(selectedId));
      closeDetail();
      await load();
    } catch (err) {
      setActionError(err.response?.data?.error || err.message || 'Delete failed');
    }
  };

  const handleBatchDelete = async () => {
    if (rowSelectionModel.length === 0) return;
    setDeleteBusy(true);
    setActionError(null);
    try {
      for (const id of rowSelectionModel) {
        await tranApi.delete(tranEndpoints.genericDataItem(id));
      }
      if (selectedId != null && rowSelectionModel.includes(selectedId)) {
        closeDetail();
      }
      setRowSelectionModel([]);
      setBatchDeleteOpen(false);
      await load();
    } catch (err) {
      setActionError(err.response?.data?.error || err.message || 'Delete failed');
    } finally {
      setDeleteBusy(false);
    }
  };

  const handleAnalyze = async () => {
    if (!selectedId) return;
    setAnalyzeLoading(true);
    setActionError(null);
    try {
      const res = await tranApi.post(tranEndpoints.genericDataAnalyze(selectedId));
      const analysis = res.data?.ai_analysis || '';
      setDetail((prev) => (prev ? { ...prev, ai_analysis: analysis } : prev));
      setRows((prev) => prev.map((r) => (r.id === selectedId ? { ...r, ai_analysis: analysis } : r)));
      setDetailTab('analysis');
    } catch (err) {
      setActionError(err.response?.data?.error || err.message || 'Analysis failed');
    } finally {
      setAnalyzeLoading(false);
    }
  };

  const gridColumns = useMemo(
    () => [
      {
        field: 'title',
        headerName: 'Title',
        flex: 1.4,
        minWidth: 200,
      },
      {
        field: 'source_type',
        headerName: 'Type',
        width: 110,
        renderCell: ({ value }) => (
          <Chip
            size="small"
            icon={sourceTypeIcon(value)}
            label={String(value || '').toUpperCase()}
            color={sourceTypeColor(value)}
            variant="outlined"
            sx={{ fontWeight: 600, fontSize: '0.7rem' }}
          />
        ),
      },
      {
        field: 'source_filename',
        headerName: 'File',
        flex: 1,
        minWidth: 160,
        valueGetter: (_v, row) => row?.source_filename || '—',
      },
      {
        field: 'record_count',
        headerName: 'Records',
        width: 100,
        type: 'number',
      },
      {
        field: 'ai_analysis',
        headerName: 'AI',
        width: 72,
        sortable: false,
        renderCell: ({ row }) =>
          row?.ai_analysis ? (
            <Tooltip title="Analysis available">
              <AutoAwesomeOutlinedIcon fontSize="small" color="primary" />
            </Tooltip>
          ) : (
            '—'
          ),
      },
      {
        field: 'last_updated',
        headerName: 'Updated',
        width: 160,
        valueGetter: (_v, row) => formatWhen(row?.last_updated),
      },
    ],
    []
  );

  if (error) return <Alert severity="error">{error}</Alert>;

  return (
    <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', gap: 2 }}>
      <Paper
        elevation={0}
        sx={{
          p: 2,
          borderRadius: 2,
          border: 1,
          borderColor: 'divider',
          background: `linear-gradient(120deg, ${alpha(theme.palette.primary.main, 0.06)} 0%, transparent 55%)`,
        }}
      >
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems={{ sm: 'center' }} justifyContent="space-between">
          <Stack direction="row" spacing={1.5} alignItems="center">
            <Box
              sx={{
                width: 44,
                height: 44,
                borderRadius: 2,
                display: 'grid',
                placeItems: 'center',
                bgcolor: alpha(theme.palette.primary.main, 0.12),
                color: 'primary.main',
              }}
            >
              <DescriptionOutlinedIcon />
            </Box>
            <Box>
              <Typography variant="h6" fontWeight={800} lineHeight={1.2}>
                {labels.nav_generic_data || 'Generic data'}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Import any CSV, JSON, PDF, or Markdown — PDFs become markdown articles instantly.
              </Typography>
            </Box>
          </Stack>
          <Button
            variant="contained"
            startIcon={<CloudUploadOutlinedIcon />}
            onClick={() => setImportOpen(true)}
            sx={{ textTransform: 'none', borderRadius: 2, alignSelf: { xs: 'stretch', sm: 'auto' } }}
          >
            Import file
          </Button>
        </Stack>
      </Paper>

      <Paper variant="outlined" sx={{ flex: 1, minHeight: 360, borderRadius: 2, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
        {rowSelectionModel.length > 0 && (
          <Stack direction="row" alignItems="center" spacing={1} sx={{ px: 1.5, py: 0.75, borderBottom: 1, borderColor: 'divider' }}>
            <Typography variant="body2" color="text.secondary">
              {rowSelectionModel.length} selected
            </Typography>
            <Button
              size="small"
              color="error"
              variant="outlined"
              startIcon={<DeleteOutlineIcon />}
              onClick={() => setBatchDeleteOpen(true)}
              disabled={deleteBusy}
            >
              Delete selected
            </Button>
          </Stack>
        )}
        <DataGrid
          rows={rows}
          columns={gridColumns}
          loading={loading}
          getRowId={(r) => r.id}
          checkboxSelection
          checkboxSelectionVisibleOnly
          rowSelectionModel={rowSelectionModel}
          onRowSelectionModelChange={(model) => setRowSelectionModel(model)}
          onRowClick={(params) => openDetail(params.id)}
          disableRowSelectionOnClick
          pageSizeOptions={[25, 50, 100]}
          initialState={{ pagination: { paginationModel: { pageSize: 25 } } }}
          sx={{
            border: 0,
            flex: 1,
            minHeight: 0,
            '& .MuiDataGrid-row': { cursor: 'pointer' },
          }}
        />
      </Paper>

      <Dialog open={batchDeleteOpen} onClose={() => !deleteBusy && setBatchDeleteOpen(false)}>
        <DialogTitle>Delete {rowSelectionModel.length} records</DialogTitle>
        <DialogContent>
          <Typography>
            Delete {rowSelectionModel.length} imported data record{rowSelectionModel.length === 1 ? '' : 's'}? This cannot be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setBatchDeleteOpen(false)} disabled={deleteBusy}>
            Cancel
          </Button>
          <Button color="error" variant="contained" onClick={handleBatchDelete} disabled={deleteBusy}>
            {deleteBusy ? 'Deleting…' : `Delete ${rowSelectionModel.length}`}
          </Button>
        </DialogActions>
      </Dialog>

      <ImportDialog
        open={importOpen}
        onClose={() => setImportOpen(false)}
        onImported={(record) => {
          load().then(() => {
            if (record?.id) openDetail(record.id);
          });
        }}
      />

      <Drawer
        anchor="right"
        open={Boolean(selectedId)}
        onClose={closeDetail}
        PaperProps={{
          sx: {
            width: { xs: '100%', sm: '66.666vw' },
            maxWidth: '100%',
            top: { xs: 0, sm: `${ADMIN_RIGHT_PANEL_TOP}px` },
            height: { xs: '100%', sm: adminRightPanelHeightCalc },
            maxHeight: { xs: '100%', sm: adminRightPanelHeightCalc },
            boxSizing: 'border-box',
            display: 'flex',
            flexDirection: 'column',
          },
        }}
        ModalProps={{ keepMounted: false, disableScrollLock: true }}
      >
        <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ px: 2, py: 1.5, borderBottom: 1, borderColor: 'divider' }}>
          <Box sx={{ minWidth: 0, pr: 1 }}>
            <Typography variant="subtitle1" fontWeight={700} noWrap>
              {detail?.title || 'Loading…'}
            </Typography>
            {detail && (
              <Stack direction="row" spacing={0.75} sx={{ mt: 0.5 }} flexWrap="wrap">
                <Chip
                  size="small"
                  icon={sourceTypeIcon(detail.source_type)}
                  label={String(detail.source_type || '').toUpperCase()}
                  color={sourceTypeColor(detail.source_type)}
                  variant="outlined"
                />
                {detail.source_filename && (
                  <Chip size="small" label={detail.source_filename} variant="outlined" sx={{ maxWidth: 220 }} />
                )}
              </Stack>
            )}
          </Box>
          <Stack direction="row" spacing={0.5}>
            <Tooltip title="Delete">
              <IconButton size="small" onClick={handleDelete} disabled={!detail || detailLoading}>
                <DeleteOutlineIcon fontSize="small" />
              </IconButton>
            </Tooltip>
            <IconButton size="small" onClick={closeDetail}>
              <CloseIcon />
            </IconButton>
          </Stack>
        </Stack>

        <Tabs
          value={detailTab}
          onChange={(_, v) => setDetailTab(v)}
          sx={{ px: 2, borderBottom: 1, borderColor: 'divider', minHeight: 42 }}
        >
          <Tab value="content" label="Content" sx={{ textTransform: 'none', minHeight: 42 }} />
          <Tab
            value="analysis"
            label={
              <Stack direction="row" spacing={0.5} alignItems="center">
                <span>AI analysis</span>
                {detail?.ai_analysis && <AutoAwesomeOutlinedIcon sx={{ fontSize: 14, color: 'primary.main' }} />}
              </Stack>
            }
            sx={{ textTransform: 'none', minHeight: 42 }}
          />
        </Tabs>

        <Box sx={{ flex: 1, minHeight: 0, overflow: 'auto', p: 2.5 }}>
          {actionError && (
            <Alert severity="error" sx={{ mb: 2 }} onClose={() => setActionError(null)}>
              {actionError}
            </Alert>
          )}
          {detailLoading ? (
            <Box sx={{ display: 'grid', placeItems: 'center', py: 8 }}>
              <CircularProgress />
            </Box>
          ) : detailTab === 'content' ? (
            <Stack spacing={2.5} sx={{ width: '100%', minWidth: 0 }}>
              {detail?.description && (
                <Paper variant="outlined" sx={{ p: 1.5, borderRadius: 2 }}>
                  <Typography variant="overline" color="text.secondary">
                    Description
                  </Typography>
                  <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>
                    {detail.description}
                  </Typography>
                </Paper>
              )}
              <GenericDataContentView record={detail} />
            </Stack>
          ) : (
            <AiAnalysisPanel
              analysis={detail?.ai_analysis}
              loading={analyzeLoading}
              onAnalyze={handleAnalyze}
              hasRecord={Boolean(detail)}
            />
          )}
        </Box>
      </Drawer>
    </Box>
  );
}
