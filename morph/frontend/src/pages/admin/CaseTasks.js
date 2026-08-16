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
  TextField,
  Typography,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import { MapContainer, Polygon, Polyline, TileLayer, useMapEvents } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { JsonDetailEditor, jsonDetailToString } from '../../components/admin/jsonDetailViews';
import { validateJsonDetailStructure } from '../../components/admin/jsonDetailValidate';
import CaseTaskEmailDialog from '../../components/admin/CaseTaskEmailDialog';
import RecordAttachmentsPanel from '../../components/admin/RecordAttachmentsPanel';
import { useConfirm } from '../../components/ConfirmDialog';

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
  if (!value) return '—';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleString([], {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
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

const initialDraft = {
  title: '',
  description: '',
  start_at: '',
  end_at: '',
  detail: '',
  assignees: [],
  location: '',
};

export default function CaseTasks() {
  const { confirm } = useConfirm();
  const [rows, setRows] = useState([]);
  const [members, setMembers] = useState([]);
  const [employees, setEmployees] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState(null);
  const [draft, setDraft] = useState(initialDraft);
  const [detailError, setDetailError] = useState('');
  const [detailInfo, setDetailInfo] = useState('');
  const [detailEditorMode, setDetailEditorMode] = useState('preview');
  const [submitting, setSubmitting] = useState(false);
  const [emailDialogOpen, setEmailDialogOpen] = useState(false);
  const [locationDialogOpen, setLocationDialogOpen] = useState(false);
  const [attachments, setAttachments] = useState([]);

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

  const openCreate = () => {
    setEditing(null);
    setDraft(initialDraft);
    setDetailEditorMode('preview');
    setAttachments([]);
    setDetailError('');
    setDetailInfo('');
    setDialogOpen(true);
  };

  const openEdit = async (row) => {
    setEditing(row);
    setDetailError('');
    setDetailInfo('');
    setDetailEditorMode('preview');
    setAttachments([]);
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
    setDraft({
      title: data.title || '',
      description: data.description || '',
      start_at: toDateTimeLocal(data.start_at),
      end_at: toDateTimeLocal(data.end_at),
      detail: typeof data.detail === 'string' ? data.detail : JSON.stringify(data.detail || {}, null, 2),
      assignees: draftAssignees,
      location: typeof data.location === 'string' ? data.location : data.location ? JSON.stringify(data.location, null, 2) : '',
    });
    setAttachments(Array.isArray(data.attachments) ? data.attachments : []);
    setDialogOpen(true);
  };

  const onCloseDialog = () => {
    if (submitting) return;
    setDialogOpen(false);
  };

  const submit = async () => {
    if (!draft.title.trim()) {
      setDetailError('Title is required.');
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
              alignItems: 'stretch',
            }}
          >
            {rows.map((row) => (
              <Card key={row.id} variant="outlined" sx={{ height: 170, display: 'flex' }}>
                <CardActionArea sx={{ height: 1, p: 0 }} onClick={() => openEdit(row)}>
                  <CardContent sx={{ height: 1, display: 'flex', flexDirection: 'column', gap: 1 }}>
                    <Typography
                      variant="subtitle1"
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
                    <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                      <Typography variant="body2" color="text.secondary" noWrap title={row.description || ''}>
                        {row.description || 'No description'}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Start: {dateTimeDisplay(row.start_at)}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        End: {dateTimeDisplay(row.end_at)}
                      </Typography>
                    </Box>
                  </CardContent>
                </CardActionArea>
              </Card>
            ))}
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
        <Box sx={{ flex: 1, minHeight: 0, overflowY: 'auto', p: 2.5 }}>
          <Stack spacing={1.5}>
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
                size="small"
                type="datetime-local"
                label="Start date/time"
                value={draft.start_at}
                onChange={(e) => setDraft((prev) => ({ ...prev, start_at: e.target.value }))}
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
              <TextField
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
        </Box>
        <Box sx={{ px: 2.5, py: 1.5, borderTop: '1px solid', borderColor: 'divider', display: 'flex', alignItems: 'center', gap: 1 }}>
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
          <Button onClick={onCloseDialog} disabled={submitting}>
            Cancel
          </Button>
          <Button variant="contained" onClick={submit} disabled={submitting}>
            {submitting ? 'Saving...' : 'Save'}
          </Button>
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
