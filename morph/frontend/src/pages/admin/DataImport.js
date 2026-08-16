import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  LinearProgress,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Typography,
} from '@mui/material';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import { Navigate } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../../apiBase';
import { getMorphToken } from '../../auth/morphSession';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { isMorphAdmin } from '../../auth/isMorphAdmin';
import { useAdminBasePath } from '../../adminPaths';

const FILE_IMPORT_ENTITIES = [
  {
    kind: 'generic_data',
    label: 'Generic data',
    description: 'Import a CSV, Excel, or JSON file as a Generic data bag for Morph AI analysis.',
    mode: 'generic',
  },
  {
    kind: 'asset',
    label: 'Assets',
    description: 'Import assets (vehicles, equipment, people-as-assets) into Morph Data Assets.',
    mode: 'collector',
  },
];

const ALLOWED_EXTS = ['csv', 'json', 'xlsx', 'xls'];

function fileExt(name) {
  const parts = String(name || '').split('.');
  return parts.length > 1 ? parts.pop().toLowerCase() : '';
}

export default function DataImport() {
  const base = useAdminBasePath();
  const fileRef = useRef(null);
  const [entity, setEntity] = useState(FILE_IMPORT_ENTITIES[0].kind);
  const [error, setError] = useState('');
  const [status, setStatus] = useState('');
  const [report, setReport] = useState(null);
  const [job, setJob] = useState(null);
  const [busy, setBusy] = useState(false);
  const [collectorSpecs, setCollectorSpecs] = useState({});

  const loadCollectorSpecs = useCallback(async () => {
    try {
      const { data } = await tranApi.get('/api/data-collector/entities');
      const list = Array.isArray(data.entities) ? data.entities : [];
      const map = {};
      list.forEach((e) => {
        map[e.kind] = e;
      });
      setCollectorSpecs(map);
    } catch {
      /* examples optional */
    }
  }, []);

  useEffect(() => {
    void loadCollectorSpecs();
  }, [loadCollectorSpecs]);

  useEffect(() => {
    if (!job?.id || job.status === 'completed' || job.status === 'failed') return undefined;
    const t = setInterval(async () => {
      try {
        const { data } = await tranApi.get(`/api/data-collector/jobs/${job.id}`);
        setJob(data);
      } catch {
        /* ignore poll errors */
      }
    }, 800);
    return () => clearInterval(t);
  }, [job]);

  if (!isMorphAdmin()) {
    return <Navigate to={`${base}/configuration/users`} replace />;
  }

  const selected = FILE_IMPORT_ENTITIES.find((e) => e.kind === entity);
  const collectorSpec = collectorSpecs[entity];

  const downloadExample = (kind) => {
    if (!collectorSpec) return;
    const text = kind === 'json' ? collectorSpec.json_example : collectorSpec.csv_example;
    const blob = new Blob([text || ''], { type: kind === 'json' ? 'application/json' : 'text/csv' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = `${entity}-example.${kind === 'json' ? 'json' : 'csv'}`;
    a.click();
    URL.revokeObjectURL(a.href);
  };

  const assertAllowedFile = (file) => {
    const ext = fileExt(file.name);
    if (!ALLOWED_EXTS.includes(ext)) {
      throw new Error('Use a .csv, .xlsx/.xls, or .json file');
    }
  };

  const withFile = async (fn) => {
    const file = fileRef.current?.files?.[0];
    if (!file) {
      setError('Choose a file first');
      return;
    }
    if (!entity) {
      setError('Choose an entity');
      return;
    }
    setBusy(true);
    setError('');
    setStatus('');
    try {
      assertAllowedFile(file);
      await fn(file);
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Request failed');
    } finally {
      setBusy(false);
    }
  };

  const validate = () =>
    withFile(async (file) => {
      if (selected?.mode === 'generic') {
        setReport(null);
        setStatus(`Ready to import ${file.name} as Generic data`);
        return;
      }
      const fd = new FormData();
      fd.append('entity', entity);
      fd.append('file', file);
      const { data } = await tranApi.post('/api/data-collector/validate', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      setReport(data.report || null);
      setStatus(data.report?.message || 'Validated');
    });

  const startImport = () =>
    withFile(async (file) => {
      if (selected?.mode === 'generic') {
        const fd = new FormData();
        fd.append('file', file);
        fd.append('title', file.name.replace(/\.[^.]+$/, ''));
        const token = getMorphToken();
        await axios.post(`${API_BASE_URL}${tranEndpoints.genericDataImport}`, fd, {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
          timeout: 300000,
        });
        setJob(null);
        setReport(null);
        setStatus('Generic data import completed');
        return;
      }
      const fd = new FormData();
      fd.append('entity', entity);
      fd.append('file', file);
      const { data } = await tranApi.post('/api/data-collector/jobs', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      setStatus(data.message || 'Import started');
      if (data.job_id) {
        const { data: j } = await tranApi.get(`/api/data-collector/jobs/${data.job_id}`);
        setJob(j);
      }
    });

  return (
    <Box sx={{ p: 2, maxWidth: 720 }}>
      <Typography variant="h5" fontWeight={700} sx={{ mb: 0.5 }}>
        File import
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        Bulk-import Morph Data records from CSV, Excel, or JSON.
      </Typography>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}
      {status && (
        <Alert severity="info" sx={{ mb: 2 }} onClose={() => setStatus('')}>
          {status}
        </Alert>
      )}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack spacing={2}>
          <TextField select label="Data type" value={entity} onChange={(e) => setEntity(e.target.value)} fullWidth size="small">
            {FILE_IMPORT_ENTITIES.map((e) => (
              <MenuItem key={e.kind} value={e.kind}>
                {e.label}
              </MenuItem>
            ))}
          </TextField>
          {selected && (
            <Typography variant="body2" color="text.secondary">
              {selected.description}
            </Typography>
          )}
          {selected?.mode === 'collector' && (
            <Stack direction="row" spacing={1} flexWrap="wrap">
              <Button size="small" onClick={() => downloadExample('csv')} disabled={!collectorSpec}>
                Download CSV example
              </Button>
              <Button size="small" onClick={() => downloadExample('json')} disabled={!collectorSpec}>
                Download JSON example
              </Button>
            </Stack>
          )}
          <Button variant="outlined" component="label" startIcon={<UploadFileIcon />} disabled={busy}>
            Choose file…
            <input ref={fileRef} type="file" accept=".csv,.json,.xlsx,.xls,.txt" hidden />
          </Button>
          <Stack direction="row" spacing={1}>
            <Button variant="outlined" disabled={busy} onClick={validate}>
              Validate
            </Button>
            <Button variant="contained" disabled={busy} onClick={startImport}>
              Start import
            </Button>
          </Stack>
          {report && (
            <Typography variant="body2">
              Rows: {report.row_count} · Template match: {report.uses_template ? 'yes' : 'no'}
            </Typography>
          )}
          {job && (
            <Box>
              <Typography variant="body2" sx={{ mb: 0.5 }}>
                Job {job.status}: {job.message} ({job.processed_rows || 0}/{job.total_rows || 0})
              </Typography>
              <LinearProgress variant="determinate" value={job.percent || 0} />
            </Box>
          )}
        </Stack>
      </Paper>
    </Box>
  );
}
