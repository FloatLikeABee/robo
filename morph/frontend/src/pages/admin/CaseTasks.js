import React, { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  Chip,
  CircularProgress,
  Drawer,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Tab,
  Tabs,
  TextField,
  Typography,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import AutoAwesomeOutlinedIcon from '@mui/icons-material/AutoAwesomeOutlined';
import { MapContainer, Polygon, Polyline, TileLayer, useMapEvents } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { JsonDetailEditor, jsonDetailToString } from '../../components/admin/jsonDetailViews';
import { validateJsonDetailStructure } from '../../components/admin/jsonDetailValidate';
import CaseTaskEmailDialog from '../../components/admin/CaseTaskEmailDialog';
import RecordAttachmentsPanel from '../../components/admin/RecordAttachmentsPanel';
import { useConfirm } from '../../components/ConfirmDialog';
import { buildCaseTaskHTML, buildCaseTaskMarkdown } from './caseTaskViewDocs';
import { darkPreviewIframeSx, withDarkPreviewSrcDoc } from '../../lib/darkPreviewSrcDoc';

const MAP_CENTER = [39.8283, -98.5795];

function escapeHtml(s) {
  return String(s || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function buildDefaultEmailHtml(title, assigneesLabel) {
  const t = escapeHtml(title || 'Case/task');
  const al = assigneesLabel
    ? `<p><strong>Assignees:</strong> ${escapeHtml(assigneesLabel)}</p>`
    : '';
  return `<p>Hello,</p><p>Please review the following case/task: <strong>${t}</strong>.</p>${al}<p>Details are in this message below.</p><p>Thank you,</p>`;
}


function MapClickCapture({ onClick }) {
  useMapEvents({
    click(e) {
      onClick(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

function parseLocation(raw) {
  if (!raw) return { label: '', area: [] };
  const parsed = typeof raw === 'string' ? (() => {
    try {
      return JSON.parse(raw);
    } catch {
      return { label: String(raw || ''), area: [] };
    }
  })() : raw;
  const label = typeof parsed?.label === 'string' ? parsed.label : typeof parsed?.location === 'string' ? parsed.location : '';
  const arr = Array.isArray(parsed?.area) ? parsed.area : [];
  const area = arr
    .map((p) => {
      if (!Array.isArray(p) || p.length < 2) return null;
      const lat = Number(p[0]);
      const lng = Number(p[1]);
      if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null;
      return [lat, lng];
    })
    .filter(Boolean);
  return { label, area };
}

function locationSummary(raw) {
  const p = parseLocation(raw);
  if (!p.label && p.area.length === 0) return '';
  if (p.label && p.area.length > 0) return `${p.label} (${p.area.length} points)`;
  if (p.label) return p.label;
  return `Area (${p.area.length} points)`;
}

function toDateTimeLocal(value) {
  if (!value) return '';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return '';
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function dateTimeDisplay(value) {
  if (!value) return '';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
}

function AreaPreviewMap({ raw }) {
  const parsed = parseLocation(raw);
  const points = parsed.area || [];
  const center = points[0] || MAP_CENTER;
  const zoom = points.length > 0 ? 13 : 4;

  return (
    <Box sx={{ height: 220, width: '100%', border: '1px solid', borderColor: 'divider', borderRadius: 1, overflow: 'hidden' }}>
      <MapContainer center={center} zoom={zoom} style={{ height: '100%', width: '100%' }}>
        <TileLayer
          crossOrigin
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        {points.length >= 2 ? <Polyline positions={points} /> : null}
        {points.length >= 3 ? <Polygon positions={points} /> : null}
      </MapContainer>
    </Box>
  );
}

function AreaEditorDialog({ open, value, onCancel, onApply }) {
  const parsed = useMemo(() => parseLocation(value), [value]);
  const [label, setLabel] = useState('');
  const [points, setPoints] = useState([]);

  useEffect(() => {
    if (!open) return;
    setLabel(parsed.label || '');
    setPoints(parsed.area || []);
  }, [open, parsed.area, parsed.label]);

  const apply = () => {
    if (!label.trim() && points.length === 0) {
      onApply(null);
      return;
    }
    onApply(JSON.stringify({ label: label.trim(), area: points }, null, 2));
  };

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth="md">
      <DialogTitle>Select area</DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
          Click map to add polygon points. Keep at least 3 points to represent an area.
        </Typography>
        <Box sx={{ height: 320, width: '100%', borderRadius: 1, overflow: 'hidden', mb: 1.5 }}>
          <MapContainer center={points[0] || MAP_CENTER} zoom={points[0] ? 13 : 4} style={{ height: '100%', width: '100%' }}>
            <TileLayer
              crossOrigin
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />
            <MapClickCapture onClick={(lat, lng) => setPoints((prev) => [...prev, [Number(lat.toFixed(6)), Number(lng.toFixed(6))]])} />
            {points.length >= 2 ? <Polyline positions={points} /> : null}
            {points.length >= 3 ? <Polygon positions={points} /> : null}
          </MapContainer>
        </Box>
        <TextField
          fullWidth
          size="small"
          label="Location label (optional)"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          sx={{ mb: 1.5 }}
        />
        <Stack direction="row" spacing={1} alignItems="center">
          <Typography variant="body2" color="text.secondary">
            Points: {points.length}
          </Typography>
          <Button size="small" onClick={() => setPoints((prev) => prev.slice(0, -1))} disabled={points.length === 0}>
            Undo point
          </Button>
          <Button size="small" color="inherit" onClick={() => setPoints([])} disabled={points.length === 0}>
            Clear area
          </Button>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        <Button variant="contained" onClick={apply}>
          Apply
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function pad2(n) {
  return String(n).padStart(2, '0');
}

function localDayBounds(d = new Date()) {
  const y = d.getFullYear();
  const m = pad2(d.getMonth() + 1);
  const day = pad2(d.getDate());
  return {
    start_at: `${y}-${m}-${day}T00:00`,
    end_at: `${y}-${m}-${day}T23:59`,
  };
}

function emptyTaskDraft() {
  return {
    title: '',
    description: '',
    ...localDayBounds(),
    detail: '',
    assignees: [],
    location: '',
  };
}

const CASE_TASK_AI_ACCEPT = '.txt,.md,.markdown,.pdf,.csv,.xlsx';

function caseTaskDraftLooksFilled(draft) {
  if (String(draft?.title || '').trim()) return true;
  if (String(draft?.description || '').trim()) return true;
  const d = jsonDetailToString(draft?.detail).trim();
  return Boolean(d && d !== '{}' && d !== 'null');
}

function locationFromAiDraft(loc) {
  if (loc == null || loc === '') return '';
  if (typeof loc === 'string') {
    const s = loc.trim();
    if (!s || s === 'null') return '';
    try {
      JSON.parse(s);
      return s;
    } catch {
      return JSON.stringify({ label: s, area: [] }, null, 2);
    }
  }
  if (typeof loc === 'object') {
    const label = typeof loc.label === 'string' ? loc.label.trim() : typeof loc.location === 'string' ? loc.location.trim() : '';
    const area = Array.isArray(loc.area) ? loc.area : [];
    if (!label && area.length === 0) return '';
    return JSON.stringify({ label, area }, null, 2);
  }
  return '';
}

export default function CaseTasks() {
  const { confirm } = useConfirm();
  const [rows, setRows] = useState([]);
  const [members, setMembers] = useState([]);
  const [employees, setEmployees] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState(null);
  const [draft, setDraft] = useState(emptyTaskDraft);
  const [detailError, setDetailError] = useState('');
  const [detailInfo, setDetailInfo] = useState('');
  const [detailEditorMode, setDetailEditorMode] = useState('preview');
  const [submitting, setSubmitting] = useState(false);
  const [emailDialogOpen, setEmailDialogOpen] = useState(false);
  const [locationDialogOpen, setLocationDialogOpen] = useState(false);
  const [attachments, setAttachments] = useState([]);
  const [aiPrompt, setAiPrompt] = useState('');
  const [aiFile, setAiFile] = useState(null);
  const [aiGenerating, setAiGenerating] = useState(false);
  const [aiError, setAiError] = useState('');
  const [detailsTab, setDetailsTab] = useState('details');

  const recipientOptions = useMemo(() => {
    const out = [];
    (employees || []).forEach((x) => {
      out.push({
        kind: 'employee',
        id: x.id,
        label: x.label,
        email: x.email || '',
        group: 'Employees',
      });
    });
    (members || []).forEach((x) => {
      out.push({
        kind: 'member',
        id: x.id,
        label: x.label,
        email: x.email || '',
        group: 'Members',
      });
    });
    return out;
  }, [employees, members]);

  const viewMarkdown = useMemo(() => {
    const assignees =
      (draft.assignees || [])
        .map((a) => a.label || a.name)
        .filter(Boolean)
        .join(', ') || editing?.assignees_label || '';
    return buildCaseTaskMarkdown({
      title: draft.title,
      description: draft.description,
      assignees,
      startAt: draft.start_at,
      endAt: draft.end_at,
      location: draft.location,
      detail: draft.detail,
    });
  }, [draft, editing?.assignees_label]);

  const viewHTML = useMemo(
    () => buildCaseTaskHTML({ title: draft.title, markdown: viewMarkdown, detail: draft.detail }),
    [draft.title, draft.detail, viewMarkdown]
  );

  const load = async () => {
    const [tasksRes, membersRes, employeesRes] = await Promise.all([
      tranApi.get(tranEndpoints.caseTasks),
      tranApi.get(tranEndpoints.members),
      tranApi.get(tranEndpoints.employees),
    ]);
    setRows(tasksRes.data || []);
    setMembers(
      (membersRes.data || []).map((x) => ({
        id: x.id,
        label: [x.first_name, x.last_name].filter(Boolean).join(' ') || `Member #${x.id}`,
        email: x.email || '',
      }))
    );
    setEmployees(
      (employeesRes.data || []).map((x) => ({
        id: x.id,
        label: [x.first_name, x.last_name].filter(Boolean).join(' ') || `Employee #${x.id}`,
        email: x.email || '',
      }))
    );
  };

  useEffect(() => {
    load()
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load case/tasks'))
      .finally(() => setLoading(false));
  }, []);

  const resetAiMaterial = () => {
    setAiPrompt('');
    setAiFile(null);
    setAiGenerating(false);
    setAiError('');
  };

  const openCreate = () => {
    setEditing(null);
    setDraft(emptyTaskDraft());
    setDetailEditorMode('preview');
    setAttachments([]);
    setDetailError('');
    setDetailInfo('');
    setDetailsTab('details');
    resetAiMaterial();
    setDialogOpen(true);
  };

  const openEdit = async (row) => {
    setEditing(row);
    setDetailError('');
    setDetailInfo('');
    setDetailEditorMode('preview');
    setAttachments([]);
    setDetailsTab('details');
    const full = await tranApi.get(tranEndpoints.caseTaskFull(row.id));
    const data = full.data || {};
    const apiAssignees = Array.isArray(data.assignees) ? data.assignees : [];
    const draftAssignees = apiAssignees.map((a) => {
      const kind = a.assignee_kind || a.kind;
      const id = a.assignee_id ?? a.id;
      const found = recipientOptions.find((o) => o.kind === kind && o.id === id);
      if (found) return found;
      return {
        kind,
        id,
        label: a.name || `${kind} #${id}`,
        email: a.email || '',
        group: kind === 'employee' ? 'Employees' : 'Members',
      };
    });
    setEditing({
      ...row,
      assignees: apiAssignees,
      assignees_label: data.assignees_label || row.assignees_label,
    });
    const bounds = localDayBounds();
    setDraft({
      title: data.title || '',
      description: data.description || '',
      start_at: toDateTimeLocal(data.start_at) || bounds.start_at,
      end_at: toDateTimeLocal(data.end_at) || bounds.end_at,
      detail: typeof data.detail === 'string' ? data.detail : JSON.stringify(data.detail || {}, null, 2),
      assignees: draftAssignees,
      location: typeof data.location === 'string' ? data.location : data.location ? JSON.stringify(data.location, null, 2) : '',
    });
    setAttachments(Array.isArray(data.attachments) ? data.attachments : []);
    resetAiMaterial();
    setDialogOpen(true);
  };

  const onCloseDialog = () => {
    if (submitting || aiGenerating) return;
    setDialogOpen(false);
  };

  const generateFromMaterial = async () => {
    const prompt = aiPrompt.trim();
    if (!prompt && !aiFile) {
      setAiError('Provide a prompt, a file, or both.');
      return;
    }
    if (caseTaskDraftLooksFilled(draft)) {
      const ok = await confirm({
        title: 'Replace case/task fields?',
        message: 'This will replace the current title, description, dates, map area, and detail JSON. Save afterward to keep the change.',
        confirmLabel: 'Replace',
      });
      if (!ok) return;
    }
    setAiGenerating(true);
    setAiError('');
    try {
      let res;
      if (aiFile) {
        const form = new FormData();
        if (prompt) form.append('prompt', prompt);
        form.append('file', aiFile);
        res = await tranApi.post(tranEndpoints.caseTaskAiDraft, form, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
      } else {
        res = await tranApi.post(tranEndpoints.caseTaskAiDraft, { prompt });
      }
      const data = res.data || {};
      const title = String(data.title || '').trim();
      const detail = data.detail;
      const detailCheck = validateJsonDetailStructure(detail);
      if (
        !title ||
        detail == null ||
        typeof detail !== 'object' ||
        Array.isArray(detail) ||
        Object.keys(detail).length === 0 ||
        !detailCheck.ok
      ) {
        throw new Error(detailCheck.error || 'AI did not return a valid case/task draft');
      }
      const bounds = localDayBounds();
      setDraft((prev) => ({
        ...prev,
        title,
        description: String(data.description || ''),
        start_at: toDateTimeLocal(data.start_at) || bounds.start_at,
        end_at: toDateTimeLocal(data.end_at) || bounds.end_at,
        location: locationFromAiDraft(data.location),
        detail: JSON.stringify(detail, null, 2),
      }));
      setDetailError('');
      setDetailInfo('Draft filled from material. Review and save when ready.');
    } catch (err) {
      setAiError(err.response?.data?.error || err.message || 'Generate failed');
    } finally {
      setAiGenerating(false);
    }
  };

  const submit = async () => {
    if (!draft.title.trim()) {
      setDetailError('Title is required.');
      return;
    }
    if (!draft.start_at) {
      setDetailError('Start date is required.');
      return;
    }
    if (!draft.end_at) {
      setDetailError('End date is required.');
      return;
    }
    let parsedDetail = null;
    const detailRaw = String(draft.detail || '').trim();
    if (detailRaw) {
      try {
        parsedDetail = JSON.parse(detailRaw);
      } catch {
        setDetailError('Detail must be valid JSON.');
        return;
      }
      const depthCheck = validateJsonDetailStructure(parsedDetail);
      if (!depthCheck.ok) {
        setDetailError(depthCheck.error);
        return;
      }
    }
    let parsedLocation = null;
    const locationRaw = String(draft.location || '').trim();
    if (locationRaw) {
      try {
        parsedLocation = JSON.parse(locationRaw);
      } catch {
        setDetailError('Location JSON is invalid. Re-open the area selector and apply again.');
        return;
      }
    }

    const payload = {
      title: draft.title.trim(),
      description: draft.description?.trim() || null,
      start_at: draft.start_at || null,
      end_at: draft.end_at || null,
      location: parsedLocation,
      detail: parsedDetail ?? {},
    };
    // Assignees are soft-deprecated: omit on write so historical rows stay readable.

    setSubmitting(true);
    setDetailError('');
    setDetailInfo('');
    try {
      let id = editing?.id;
      if (editing) {
        await tranApi.put(tranEndpoints.caseTask(editing.id), payload);
      } else {
        const res = await tranApi.post(tranEndpoints.caseTasks, payload);
        id = res.data?.id || res.data?.ID;
      }
      await load();
      setDialogOpen(false);
    } catch (err) {
      setDetailError(err.response?.data?.error || err.message || 'Failed to save.');
    } finally {
      setSubmitting(false);
    }
  };

  const deleteRow = async (row) => {
    const ok = await confirm({
      title: 'Delete case task',
      message: `Delete "${row.title}"?`,
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    await tranApi.delete(tranEndpoints.caseTask(row.id));
    setRows((prev) => prev.filter((x) => x.id !== row.id));
    setDialogOpen(false);
  };

  const onEmailSent = () => {
    setDetailInfo('Email sent.');
    load().catch(() => {});
  };

  if (error) return <Alert severity="error">{error}</Alert>;

  return (
    <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', gap: 1 }}>
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        justifyContent="space-between"
        alignItems={{ xs: 'stretch', sm: 'center' }}
        spacing={1}
      >
        <Typography variant="h6">Tasks</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate} sx={{ alignSelf: { xs: 'stretch', sm: 'auto' } }}>
          New case/task
        </Button>
      </Stack>

      {loading ? (
        <CircularProgress sx={{ mt: 2 }} />
      ) : (
        <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', pr: 0.5 }}>
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: {
                xs: '1fr',
                sm: 'repeat(2, minmax(0, 1fr))',
                md: 'repeat(3, minmax(0, 1fr))',
              },
              gap: 1.5,
              alignItems: 'start',
            }}
          >
            {rows.map((row) => {
              const startLabel = dateTimeDisplay(row.start_at);
              const endLabel = dateTimeDisplay(row.end_at);
              return (
              <Card key={row.id} variant="outlined" sx={{ minHeight: 96, height: 'auto', display: 'flex' }}>
                <CardActionArea sx={{ height: 1, p: 0 }} onClick={() => openEdit(row)}>
                  <CardContent sx={{ py: 1.1, px: 1.5, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                    <Typography
                      variant="subtitle2"
                      sx={{
                        fontWeight: 600,
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                      }}
                      title={row.title}
                    >
                      {row.title || 'Untitled'}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" noWrap title={row.description || ''}>
                      {row.description || 'No description'}
                    </Typography>
                    {startLabel ? (
                      <Typography variant="caption" color="text.secondary" noWrap title={startLabel}>
                        Start: {startLabel}
                      </Typography>
                    ) : null}
                    {endLabel ? (
                      <Typography variant="caption" color="text.secondary" noWrap title={endLabel}>
                        End: {endLabel}
                      </Typography>
                    ) : null}
                  </CardContent>
                </CardActionArea>
              </Card>
              );
            })}
          </Box>
        </Box>
      )}

      <Drawer
        anchor="right"
        open={dialogOpen}
        onClose={onCloseDialog}
        PaperProps={{
          sx: {
            width: { xs: '100vw', sm: '66.666vw' },
            maxWidth: '100vw',
            height: '100%',
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            pt: { xs: 'env(safe-area-inset-top)', sm: 0 },
            pb: { xs: 'env(safe-area-inset-bottom)', sm: 0 },
          },
        }}
      >
        <Box sx={{ px: 2.5, py: 2, borderBottom: '1px solid', borderColor: 'divider' }}>
          <Typography variant="h6">{editing ? 'Case/task details' : 'Create case/task'}</Typography>
        </Box>
        <Tabs
          value={detailsTab}
          onChange={(_, v) => setDetailsTab(v)}
          sx={{ px: 1, minHeight: 42, borderBottom: 1, borderColor: 'divider' }}
          variant="scrollable"
          allowScrollButtonsMobile
        >
          <Tab value="details" label="Details" sx={{ minHeight: 42, textTransform: 'none' }} />
          <Tab value="markdown" label="Markdown" sx={{ minHeight: 42, textTransform: 'none' }} />
          <Tab value="html" label="HTML" sx={{ minHeight: 42, textTransform: 'none' }} />
        </Tabs>
        <Box
          sx={{
            flex: 1,
            minHeight: 0,
            overflow: detailsTab === 'details' ? 'auto' : 'hidden',
            p: detailsTab === 'details' ? 2.5 : 0,
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          {detailsTab === 'markdown' ? (
            <Box
              component="pre"
              className="themed-preview-scroll"
              sx={{
                m: 0,
                p: 2.5,
                flex: 1,
                minHeight: 0,
                overflow: 'auto',
                borderRadius: 0,
                bgcolor: 'action.hover',
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
                fontSize: 13,
              }}
            >
              {viewMarkdown}
            </Box>
          ) : null}
          {detailsTab === 'html' ? (
            <Box sx={{ flex: 1, minHeight: 0, overflow: 'hidden', bgcolor: '#0b1220' }}>
              <Box
                component="iframe"
                title="Case/task HTML preview"
                srcDoc={withDarkPreviewSrcDoc(viewHTML)}
                sandbox=""
                sx={darkPreviewIframeSx}
              />
            </Box>
          ) : null}
          {detailsTab === 'details' ? (
          <Stack spacing={1.5}>
            <Box sx={{ p: 1.5, border: '1px solid', borderColor: 'divider', borderRadius: 1 }}>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>
                Generate from material
              </Typography>
              <TextField
                size="small"
                label="Prompt"
                placeholder="Describe the case/task, or add notes to go with a file…"
                fullWidth
                multiline
                minRows={2}
                value={aiPrompt}
                onChange={(e) => setAiPrompt(e.target.value)}
                disabled={aiGenerating || submitting}
              />
              <Stack direction="row" spacing={1} alignItems="center" sx={{ mt: 1 }} flexWrap="wrap" useFlexGap>
                <Button component="label" size="small" variant="outlined" disabled={aiGenerating || submitting}>
                  Upload file
                  <input
                    hidden
                    type="file"
                    accept={CASE_TASK_AI_ACCEPT}
                    onChange={(e) => {
                      const f = e.target.files?.[0] || null;
                      setAiFile(f);
                      e.target.value = '';
                    }}
                  />
                </Button>
                {aiFile ? (
                  <Chip size="small" label={aiFile.name} onDelete={aiGenerating ? undefined : () => setAiFile(null)} />
                ) : (
                  <Typography variant="caption" color="text.secondary">
                    .txt, .md, .pdf, .csv, .xlsx
                  </Typography>
                )}
                <Box sx={{ flex: 1 }} />
                <Button
                  size="small"
                  variant="contained"
                  startIcon={aiGenerating ? <CircularProgress size={16} color="inherit" /> : <AutoAwesomeOutlinedIcon />}
                  onClick={generateFromMaterial}
                  disabled={aiGenerating || submitting}
                >
                  {aiGenerating ? 'Generating…' : 'Generate'}
                </Button>
              </Stack>
              {aiError ? (
                <Alert severity="error" sx={{ mt: 1 }}>
                  {aiError}
                </Alert>
              ) : null}
            </Box>
            <TextField
              required
              size="small"
              label="Title"
              value={draft.title}
              onChange={(e) => setDraft((prev) => ({ ...prev, title: e.target.value }))}
            />
            <TextField
              size="small"
              label="Description"
              multiline
              minRows={2}
              value={draft.description}
              onChange={(e) => setDraft((prev) => ({ ...prev, description: e.target.value }))}
              helperText="Put assignment detail in the description or JSON detail below."
            />
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5}>
              <TextField
                required
                size="small"
                type="datetime-local"
                label="Start date/time"
                value={draft.start_at}
                onChange={(e) => setDraft((prev) => ({ ...prev, start_at: e.target.value }))}
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
              <TextField
                required
                size="small"
                type="datetime-local"
                label="End date/time"
                value={draft.end_at}
                onChange={(e) => setDraft((prev) => ({ ...prev, end_at: e.target.value }))}
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
            </Stack>
            {editing?.assignees_label ? (
              <Typography variant="body2" color="text.secondary">
                Historical assignees (read-only): {editing.assignees_label}
              </Typography>
            ) : null}
            <Stack direction="row" spacing={1} alignItems="center">
              <Button variant="outlined" onClick={() => setLocationDialogOpen(true)}>
                Select map area
              </Button>
              {locationSummary(draft.location) ? <Chip size="small" label={locationSummary(draft.location)} /> : null}
              {draft.location ? (
                <Button size="small" color="inherit" onClick={() => setDraft((prev) => ({ ...prev, location: '' }))}>
                  Clear area
                </Button>
              ) : null}
            </Stack>
            {draft.location ? (
              <AreaPreviewMap raw={draft.location} />
            ) : (
              <Typography variant="body2" color="text.secondary">
                No map area selected.
              </Typography>
            )}
            <Box>
              <Typography variant="subtitle2" sx={{ mb: 0.75 }}>
                Details
              </Typography>
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 1, alignItems: 'center' }}>
                {detailEditorMode === 'preview' ? (
                  <Button size="small" variant="contained" color="primary" onClick={() => setDetailEditorMode('tree')}>
                    Edit JSON
                  </Button>
                ) : detailEditorMode === 'tree' ? (
                  <>
                    <Button size="small" variant="outlined" onClick={() => setDetailEditorMode('preview')}>
                      Card view
                    </Button>
                    <Button size="small" variant="outlined" onClick={() => setDetailEditorMode('raw')}>
                      Edit raw JSON
                    </Button>
                  </>
                ) : (
                  <>
                    <Button size="small" variant="outlined" onClick={() => setDetailEditorMode('tree')}>
                      Tree editor
                    </Button>
                    <Button size="small" variant="outlined" onClick={() => setDetailEditorMode('preview')}>
                      Card view
                    </Button>
                  </>
                )}
              </Box>
              <JsonDetailEditor
                value={jsonDetailToString(draft.detail)}
                onChange={(v) => setDraft((prev) => ({ ...prev, detail: v }))}
                mode={detailEditorMode}
                onModeChange={setDetailEditorMode}
                externalEditControls
              />
            </Box>
            <Box sx={{ px: 0.5 }}>
              <RecordAttachmentsPanel
                entityRoute="case-tasks"
                recordId={editing?.id}
                attachments={attachments}
                onChange={setAttachments}
                compact
              />
            </Box>
            {detailInfo ? <Alert severity="success">{detailInfo}</Alert> : null}
            {detailError ? <Alert severity="error">{detailError}</Alert> : null}
          </Stack>
          ) : null}
        </Box>
        <Box sx={{ px: 2.5, py: 1.5, borderTop: '1px solid', borderColor: 'divider' }}>
          {detailInfo && detailsTab !== 'details' ? <Alert severity="success" sx={{ mb: 1 }}>{detailInfo}</Alert> : null}
          {detailError && detailsTab !== 'details' ? <Alert severity="error" sx={{ mb: 1 }}>{detailError}</Alert> : null}
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          {editing ? (
            <Button onClick={() => setEmailDialogOpen(true)} disabled={submitting}>
              Send email…
            </Button>
          ) : null}
          {editing ? (
            <Button color="error" onClick={() => deleteRow(editing)} disabled={submitting}>
              Remove
            </Button>
          ) : null}
          <Box sx={{ flex: 1 }} />
          <Button onClick={onCloseDialog} disabled={submitting || aiGenerating}>
            Cancel
          </Button>
          <Button variant="contained" onClick={submit} disabled={submitting || aiGenerating}>
            {submitting ? 'Saving...' : 'Save'}
          </Button>
          </Box>
        </Box>
      </Drawer>

      <CaseTaskEmailDialog
        key={editing?.id || 0}
        open={emailDialogOpen}
        onClose={() => setEmailDialogOpen(false)}
        caseTaskId={editing?.id}
        taskTitle={editing?.title}
        initialSubject={editing?.title ? `Case/task: ${editing.title}` : 'Case/task'}
        initialHtmlBody={buildDefaultEmailHtml(editing?.title, editing?.assignees_label)}
        recipientOptions={recipientOptions}
        initialRecipients={(editing?.assignees || []).map((a) => ({
          kind: a.assignee_kind,
          id: a.assignee_id,
        }))}
        onSent={onEmailSent}
      />

      <AreaEditorDialog
        open={locationDialogOpen}
        value={draft.location}
        onCancel={() => setLocationDialogOpen(false)}
        onApply={(next) => {
          setDraft((prev) => ({ ...prev, location: next || '' }));
          setLocationDialogOpen(false);
        }}
      />
    </Box>
  );
}
