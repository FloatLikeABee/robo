import React, { useState, useEffect, useCallback, useLayoutEffect } from 'react';
import {
  Box,
  Typography,
  Tabs,
  Tab,
  Button,
  TextField,
  List,
  ListItemButton,
  ListItemText,
  FormControlLabel,
  Checkbox,
  CircularProgress,
  Alert,
  IconButton,
  Snackbar,
  Slide,
} from '@mui/material';
import AddOutlinedIcon from '@mui/icons-material/AddOutlined';
import DeleteOutlineOutlinedIcon from '@mui/icons-material/DeleteOutlineOutlined';
import AutoAwesomeOutlinedIcon from '@mui/icons-material/AutoAwesomeOutlined';
import ChevronLeftOutlinedIcon from '@mui/icons-material/ChevronLeftOutlined';
import ChevronRightOutlinedIcon from '@mui/icons-material/ChevronRightOutlined';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import { DateTimePicker } from '@mui/x-date-pickers/DateTimePicker';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { runTextAssistProgress } from '../../lib/aiProgress';
import AiProgressStatus from '../common/AiProgressStatus';
import { useConfirm } from '../ConfirmDialog';

export const MORPH_NOTES_TODOS_CHANGED = 'morph-notes-todos-changed';

const AI_BODY_MIN_FOR_IMPROVE = 20;
const SCHEDULE_TAB = 'todo-schedules';
const WEEKDAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

dayjs.extend(utc);

function toDayKeyLocal(date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

function toDayLabel(isoString) {
  if (!isoString) return '';
  const date = new Date(isoString);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString();
}

function rowSecondaryLabel(row) {
  if (row.item_type === 'todo' && row.deadline_at) {
    return `Due ${toDayLabel(row.deadline_at)}`;
  }
  if (row.created_on) {
    return new Date(row.created_on).toLocaleString();
  }
  return '';
}

function notifyNotesTodosChanged() {
  try {
    window.dispatchEvent(new CustomEvent(MORPH_NOTES_TODOS_CHANGED));
  } catch {}
}

function SlideUp(props) {
  return <Slide {...props} direction="up" />;
}

const SUCCESS_TOAST_MS = 5000;

/**
 * Shared notes & TODO editor — same API as admin panel. Use in AdminLayout panel or Morph AI drawer.
 * @param {{ variant?: 'admin'|'chat', open?: boolean }} props
 */
export default function NotesTodosContent({ variant = 'admin', open = true }) {
  const { confirm } = useConfirm();
  const [tab, setTab] = useState('notes');
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [selected, setSelected] = useState(null);
  const [draftTitle, setDraftTitle] = useState('');
  const [draftBody, setDraftBody] = useState('');
  const [draftCompleted, setDraftCompleted] = useState(false);
  const [draftDeadlineAt, setDraftDeadlineAt] = useState(null);
  const [saving, setSaving] = useState(false);
  const [composerOpen, setComposerOpen] = useState(false);
  const [composerType, setComposerType] = useState('notes');
  const [aiBusy, setAiBusy] = useState(false);
  const [aiProgressStatus, setAiProgressStatus] = useState('');
  const [calendarMonth, setCalendarMonth] = useState(() => dayjs().startOf('month'));
  const [selectedScheduleDay, setSelectedScheduleDay] = useState(() => toDayKeyLocal(new Date()));
  const [successToast, setSuccessToast] = useState({ open: false, message: '' });

  const showSuccessToast = useCallback((message) => {
    setSuccessToast({ open: true, message });
  }, []);

  const closeSuccessToast = useCallback(() => {
    setSuccessToast((prev) => ({ ...prev, open: false }));
  }, []);

  useLayoutEffect(() => {
    if (!open) return;
    setTab('notes');
    setComposerOpen(false);
    setSelected(null);
  }, [open]);

  const load = useCallback(() => {
    if (!open) return;
    const type = tab === 'notes' ? 'note' : 'todo';
    setLoading(true);
    setError(null);
    tranApi
      .get(tranEndpoints.notesTodos, { params: { type } })
      .then((res) => {
        setItems(Array.isArray(res.data) ? res.data : []);
      })
      .catch((err) => {
        setError(err.response?.data?.error || err.message || 'Failed to load');
        setItems([]);
      })
      .finally(() => setLoading(false));
  }, [open, tab]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!open || loading || composerOpen || (tab !== 'notes' && tab !== 'todos')) return;

    const visibleItems = items.filter((row) =>
      tab === 'notes' ? row.item_type === 'note' : row.item_type === 'todo'
    );
    if (visibleItems.length === 0) return;

    const selectedStillValid =
      selected != null && visibleItems.some((row) => row.id === selected.id);
    if (!selectedStillValid) {
      setSelected(visibleItems[0]);
    }
  }, [open, loading, items, tab, composerOpen, selected]);

  useEffect(() => {
    const onExt = () => load();
    window.addEventListener(MORPH_NOTES_TODOS_CHANGED, onExt);
    return () => window.removeEventListener(MORPH_NOTES_TODOS_CHANGED, onExt);
  }, [load]);

  useEffect(() => {
    if (!open) {
      setSelected(null);
      setComposerOpen(false);
      setSuccessToast({ open: false, message: '' });
    }
  }, [open]);

  useEffect(() => {
    if (selected) {
      setDraftTitle(selected.title || '');
      setDraftBody(selected.body || '');
      setDraftCompleted(!!selected.completed);
      setDraftDeadlineAt(selected.deadline_at ? dayjs(selected.deadline_at) : null);
    } else {
      setDraftTitle('');
      setDraftBody('');
      setDraftCompleted(false);
      setDraftDeadlineAt(null);
    }
  }, [selected]);

  const currentItemKind = () => {
    if (composerOpen) return composerType === 'notes' ? 'note' : 'todo';
    if (selected) return selected.item_type;
    return tab === 'notes' ? 'note' : 'todo';
  };

  const runAiAssist = async () => {
    const title = draftTitle.trim();
    if (!title) return;

    const kind = currentItemKind();
    const bodyTrim = draftBody.trim();
    const bodyChars = [...bodyTrim].length;
    const useImprove = bodyChars >= AI_BODY_MIN_FOR_IMPROVE;

    setAiBusy(true);
    setError(null);
    setAiProgressStatus('Reading your draft…');
    const assistMode = useImprove ? 'improve' : kind === 'todo' ? 'generate_todo' : 'generate_note';
    const stopProgress = runTextAssistProgress(
      { mode: assistMode, kind },
      setAiProgressStatus
    );
    try {
      if (useImprove) {
        const { data } = await tranApi.post(tranEndpoints.textAssist, {
          mode: 'improve',
          kind,
          text: bodyTrim,
        });
        const t = typeof data?.text === 'string' ? data.text : '';
        if (t) setDraftBody(t);
      } else {
        const mode = kind === 'todo' ? 'generate_todo' : 'generate_note';
        const { data } = await tranApi.post(tranEndpoints.textAssist, {
          mode,
          kind,
          seed: title,
          text: '',
        });
        const t = typeof data?.text === 'string' ? data.text : '';
        if (t) setDraftBody(t);
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'AI assist failed');
    } finally {
      stopProgress();
      setAiBusy(false);
      setAiProgressStatus('');
    }
  };

  const AiToolbar = (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.75, flexShrink: 0 }}>
      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.75, alignItems: 'center' }}>
        <Button
          size="small"
          variant="outlined"
          startIcon={aiBusy ? <CircularProgress size={14} /> : <AutoAwesomeOutlinedIcon />}
          disabled={aiBusy || !draftTitle.trim()}
          onClick={runAiAssist}
          title={
            draftTitle.trim()
              ? [...draftBody.trim()].length >= AI_BODY_MIN_FOR_IMPROVE
                ? 'Polish existing content (20+ characters)'
                : 'Draft content from the title'
              : 'Add a title first'
          }
        >
          AI assist
        </Button>
      </Box>
      {aiBusy && aiProgressStatus ? <AiProgressStatus status={aiProgressStatus} /> : null}
    </Box>
  );

  const handleSelect = (row) => {
    setSelected(row);
    setComposerOpen(false);
  };

  const handleSave = () => {
    if (!selected) return;
    const nextDeadlineAt =
      selected.item_type === 'todo' && draftDeadlineAt && dayjs(draftDeadlineAt).isValid()
        ? dayjs(draftDeadlineAt).utc().toISOString()
        : null;
    setSaving(true);
    setError(null);
    tranApi
      .put(tranEndpoints.notesTodo(selected.id), {
        title: draftTitle.trim() || '',
        body: draftBody.trim() || '',
        completed: selected.item_type === 'todo' ? draftCompleted : undefined,
        deadline_at: selected.item_type === 'todo' ? nextDeadlineAt : undefined,
      })
      .then(() => load())
      .then(() => {
        notifyNotesTodosChanged();
        showSuccessToast(selected.item_type === 'todo' ? 'To-do saved successfully.' : 'Note saved successfully.');
        setSelected((prev) =>
          prev
            ? {
                ...prev,
                title: draftTitle.trim() || null,
                body: draftBody.trim() || null,
                completed: selected.item_type === 'todo' ? draftCompleted : prev.completed,
                deadline_at: selected.item_type === 'todo' ? nextDeadlineAt : prev.deadline_at,
              }
            : null
        );
      })
      .catch((err) => setError(err.response?.data?.error || err.message || 'Save failed'))
      .finally(() => setSaving(false));
  };

  const handleDelete = async () => {
    if (!selected) return;
    const ok = await confirm({
      title: 'Delete item',
      message: 'Delete this item?',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    tranApi
      .delete(tranEndpoints.notesTodo(selected.id))
      .then(() => {
        setSelected(null);
        load();
        notifyNotesTodosChanged();
      })
      .catch((err) => setError(err.response?.data?.error || err.message || 'Delete failed'));
  };

  const handleCreate = () => {
    const itemType = composerType === 'notes' ? 'note' : 'todo';
    const deadlineAt =
      itemType === 'todo' && draftDeadlineAt && dayjs(draftDeadlineAt).isValid()
        ? dayjs(draftDeadlineAt).utc().toISOString()
        : null;
    setSaving(true);
    setError(null);
    tranApi
      .post(tranEndpoints.notesTodos, {
        item_type: itemType,
        title: draftTitle.trim() || undefined,
        body: draftBody.trim() || undefined,
        completed: itemType === 'todo' ? draftCompleted : false,
        deadline_at: itemType === 'todo' ? deadlineAt : undefined,
      })
      .then(async (res) => {
        const newId = res.data?.id;
        setComposerOpen(false);
        setDraftTitle('');
        setDraftBody('');
        setDraftCompleted(false);
        setDraftDeadlineAt(null);
        const typeParam = tab === 'notes' ? 'note' : 'todo';
        const r = await tranApi.get(tranEndpoints.notesTodos, { params: { type: typeParam } });
        const list = Array.isArray(r.data) ? r.data : [];
        setItems(list);
        notifyNotesTodosChanged();
        showSuccessToast(itemType === 'todo' ? 'To-do created successfully.' : 'Note created successfully.');
        const created = list.find((x) => x.id === Number(newId));
        if (created) setSelected(created);
      })
      .catch((err) => setError(err.response?.data?.error || err.message || 'Create failed'))
      .finally(() => setSaving(false));
  };

  const openComposer = (type) => {
    setComposerType(type);
    setSelected(null);
    setDraftTitle('');
    setDraftBody('');
    setDraftCompleted(false);
    setDraftDeadlineAt(null);
    setComposerOpen(true);
  };

  const tabMatchesItem = (row) =>
    tab === 'notes' ? row.item_type === 'note' : row.item_type === 'todo';

  const scheduledTodos = items
    .filter((row) => row.item_type === 'todo' && row.deadline_at)
    .sort((a, b) => new Date(a.deadline_at).getTime() - new Date(b.deadline_at).getTime());

  const scheduledByDay = scheduledTodos.reduce((acc, row) => {
    const dayKey = toDayKeyLocal(new Date(row.deadline_at));
    if (!acc[dayKey]) acc[dayKey] = [];
    acc[dayKey].push(row);
    return acc;
  }, {});

  const monthStart = calendarMonth.startOf('month');
  const monthEnd = calendarMonth.endOf('month');
  const leadingEmptyDays = monthStart.day();
  const totalDays = monthEnd.date();
  const totalCells = Math.ceil((leadingEmptyDays + totalDays) / 7) * 7;
  const monthCells = Array.from({ length: totalCells }, (_, i) => {
    const dayNumber = i - leadingEmptyDays + 1;
    if (dayNumber < 1 || dayNumber > totalDays) return null;
    const day = monthStart.date(dayNumber);
    return {
      dayNumber,
      dayKey: day.format('YYYY-MM-DD'),
      isToday: day.isSame(dayjs(), 'day'),
    };
  });

  const selectedScheduleItems = scheduledByDay[selectedScheduleDay] || [];

  if (!open) return null;

  const listColWidth = variant === 'chat' ? 168 : 200;

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        flex: 1,
        minHeight: 0,
        height: '100%',
        overflow: 'hidden',
      }}
    >
      {variant === 'chat' && (
        <Typography variant="caption" color="text.secondary" sx={{ px: 2, pt: 1, flexShrink: 0 }}>
          Same list as the header Notes &amp; TODOs panel — edits sync everywhere.
        </Typography>
      )}
      <Tabs
        value={tab}
        onChange={(_, v) => {
          setTab(v);
          setSelected(null);
          setComposerOpen(false);
          if (v === SCHEDULE_TAB) {
            setSelectedScheduleDay(toDayKeyLocal(new Date()));
          }
        }}
        variant="fullWidth"
        sx={{
          borderBottom: 1,
          borderColor: 'divider',
          minHeight: 42,
          flexShrink: 0,
          '& .MuiTab-root': { minHeight: 42, textTransform: 'none', fontWeight: 600 },
        }}
      >
        <Tab label="Notes" value="notes" />
        <Tab label="TODOs" value="todos" />
        <Tab label="TODO Schedules" value={SCHEDULE_TAB} />
      </Tabs>
      {error && (
        <Alert severity="error" onClose={() => setError(null)} sx={{ mx: 1, mt: 1, py: 0, flexShrink: 0 }}>
          {error}
        </Alert>
      )}
      {tab === SCHEDULE_TAB ? (
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            flex: 1,
            minHeight: 0,
            p: 1.5,
            gap: 1,
            overflow: 'hidden',
          }}
        >
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 1, flexShrink: 0 }}>
            <Typography variant="subtitle1" fontWeight={600}>
              TODO Schedules
            </Typography>
            <Button
              startIcon={<AddOutlinedIcon />}
              variant="outlined"
              size="small"
              onClick={() => {
                setTab('todos');
                openComposer('todos');
              }}
            >
              New TODO
            </Button>
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1, flexShrink: 0 }}>
            <IconButton size="small" onClick={() => setCalendarMonth((prev) => prev.subtract(1, 'month'))}>
              <ChevronLeftOutlinedIcon fontSize="small" />
            </IconButton>
            <Typography variant="subtitle2" fontWeight={600}>
              {calendarMonth.format('MMMM YYYY')}
            </Typography>
            <IconButton size="small" onClick={() => setCalendarMonth((prev) => prev.add(1, 'month'))}>
              <ChevronRightOutlinedIcon fontSize="small" />
            </IconButton>
          </Box>
          <Box
            sx={{
              flexShrink: 0,
              display: 'grid',
              gridTemplateColumns: 'repeat(7, minmax(0, 1fr))',
              gap: 0.35,
              gridAutoRows: 'minmax(36px, auto)',
            }}
          >
            {WEEKDAY_LABELS.map((label) => (
              <Typography
                key={label}
                variant="caption"
                color="text.secondary"
                sx={{ textAlign: 'center', py: 0.125, fontSize: '0.65rem', lineHeight: 1.2 }}
              >
                {label}
              </Typography>
            ))}
            {monthCells.map((cell, idx) => {
              if (!cell) {
                return (
                  <Box
                    key={`empty-${idx}`}
                    sx={{ border: 1, borderColor: 'divider', borderRadius: 0.75, minHeight: 36 }}
                  />
                );
              }
              const count = (scheduledByDay[cell.dayKey] || []).length;
              const isSelected = selectedScheduleDay === cell.dayKey;
              return (
                <Box
                  key={cell.dayKey}
                  onClick={() => setSelectedScheduleDay(cell.dayKey)}
                  sx={{
                    border: 1,
                    borderColor: isSelected ? 'primary.main' : 'divider',
                    borderRadius: 0.75,
                    minHeight: 36,
                    px: 0.35,
                    py: 0.25,
                    cursor: 'pointer',
                    bgcolor: cell.isToday ? 'action.hover' : 'background.paper',
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'flex-start',
                    gap: 0,
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.25, lineHeight: 1 }}>
                    <Typography variant="caption" fontWeight={isSelected ? 700 : 600} sx={{ fontSize: '0.75rem' }}>
                      {cell.dayNumber}
                    </Typography>
                    {count > 0 ? (
                      <Typography
                        component="span"
                        variant="caption"
                        sx={{
                          fontSize: '0.65rem',
                          fontWeight: 700,
                          bgcolor: 'primary.main',
                          color: 'primary.contrastText',
                          borderRadius: '10px',
                          px: 0.35,
                          minWidth: 14,
                          textAlign: 'center',
                          lineHeight: 1.35,
                        }}
                      >
                        {count}
                      </Typography>
                    ) : null}
                  </Box>
                </Box>
              );
            })}
          </Box>
          <Box
            sx={{
              flex: 1,
              minHeight: 0,
              overflow: 'auto',
              borderTop: 1,
              borderColor: 'divider',
              pt: 1,
              display: 'flex',
              flexDirection: 'column',
            }}
          >
            <Typography variant="subtitle2" sx={{ mb: 0.5, flexShrink: 0 }}>
              {dayjs(selectedScheduleDay).format('dddd, MMM D')}
            </Typography>
            {selectedScheduleItems.length === 0 ? (
              <Typography variant="body2" color="text.secondary">
                No scheduled TODOs for this day.
              </Typography>
            ) : (
              <List dense disablePadding sx={{ flex: 1, minHeight: 0, py: 0, overflow: 'auto' }}>
                {selectedScheduleItems.map((row) => (
                  <ListItemButton
                    key={row.id}
                    dense
                    onClick={() => {
                      setTab('todos');
                      setComposerOpen(false);
                      setSelected(row);
                    }}
                    sx={{ alignItems: 'flex-start', py: 0.75, borderRadius: 1 }}
                  >
                    <ListItemText
                      primary={row.title || '(TODO)'}
                      secondary={toDayLabel(row.deadline_at)}
                      primaryTypographyProps={{
                        noWrap: true,
                        sx: row.completed ? { textDecoration: 'line-through', opacity: 0.75 } : undefined,
                      }}
                      secondaryTypographyProps={{ sx: { whiteSpace: 'normal' } }}
                    />
                  </ListItemButton>
                ))}
              </List>
            )}
          </Box>
        </Box>
      ) : (
        <Box sx={{ display: 'flex', flex: 1, minHeight: 0, overflow: 'hidden' }}>
          <Box
            sx={{
              width: listColWidth,
              minWidth: variant === 'chat' ? 152 : 180,
              borderRight: 1,
              borderColor: 'divider',
              display: 'flex',
              flexDirection: 'column',
              minHeight: 0,
            }}
          >
            <Box sx={{ px: 1, py: 1, borderBottom: 1, borderColor: 'divider' }}>
              <Button
                startIcon={<AddOutlinedIcon />}
                variant="outlined"
                size="small"
                fullWidth
                sx={{ justifyContent: 'flex-start' }}
                onClick={() => openComposer(tab === 'notes' ? 'notes' : 'todos')}
              >
                New {tab === 'notes' ? 'note' : 'TODO'}
              </Button>
            </Box>
            <List dense sx={{ flex: 1, overflow: 'auto' }}>
              {loading && (
                <ListItemButton disabled>
                  <CircularProgress size={18} sx={{ mr: 1 }} />
                  <ListItemText primary="Loading…" />
                </ListItemButton>
              )}
              {!loading && items.filter(tabMatchesItem).length === 0 && (
                <ListItemButton disabled>
                  <ListItemText primary="Nothing here yet" secondary="Add one above" />
                </ListItemButton>
              )}
              {!loading &&
                items.filter(tabMatchesItem).map((row) => (
                  <ListItemButton key={row.id} selected={selected?.id === row.id} onClick={() => handleSelect(row)}>
                    <ListItemText
                      primary={row.title || (row.item_type === 'todo' ? '(TODO)' : '(Note)')}
                      secondary={rowSecondaryLabel(row)}
                      primaryTypographyProps={{
                        noWrap: true,
                        sx:
                          row.item_type === 'todo' && row.completed
                            ? { textDecoration: 'line-through', opacity: 0.75 }
                            : undefined,
                      }}
                    />
                  </ListItemButton>
                ))}
            </List>
          </Box>
          <Box
            sx={{
              flex: 1,
              minHeight: 0,
              overflow: 'hidden',
              p: 2,
              display: 'flex',
              flexDirection: 'column',
              gap: 0,
            }}
          >
            {composerOpen && (
              <Box
                sx={{
                  display: 'flex',
                  flexDirection: 'column',
                  flex: 1,
                  minHeight: 0,
                  gap: 1.25,
                }}
              >
                <Typography variant="subtitle2" fontWeight={600} sx={{ flexShrink: 0 }}>
                  New {composerType === 'notes' ? 'note' : 'TODO'}
                </Typography>
                {AiToolbar}
                <TextField
                  fullWidth
                  size="small"
                  label="Title"
                  value={draftTitle}
                  onChange={(e) => setDraftTitle(e.target.value)}
                  sx={{ flexShrink: 0 }}
                />
                {composerType === 'todos' && (
                  <DateTimePicker
                    label="Deadline"
                    value={draftDeadlineAt}
                    onChange={(nextValue) => setDraftDeadlineAt(nextValue)}
                    slotProps={{ textField: { size: 'small', fullWidth: true } }}
                  />
                )}
                <Box sx={{ flex: '1 1 0%', minHeight: 0, display: 'flex', flexDirection: 'column' }}>
                  <TextField
                    fullWidth
                    size="small"
                    label="Content"
                    multiline
                    value={draftBody}
                    onChange={(e) => setDraftBody(e.target.value)}
                    sx={{
                      flex: 1,
                      minHeight: 0,
                      '& .MuiOutlinedInput-root': {
                        height: '100%',
                        alignItems: 'flex-start',
                        py: 1,
                        boxSizing: 'border-box',
                      },
                      '& textarea': {
                        minHeight: '88px !important',
                        height: '100% !important',
                        maxHeight: '100%',
                        overflow: 'auto !important',
                        resize: 'none',
                        boxSizing: 'border-box',
                      },
                    }}
                  />
                </Box>
                {composerType === 'todos' && (
                  <FormControlLabel
                    sx={{ flexShrink: 0, m: 0 }}
                    control={<Checkbox checked={draftCompleted} onChange={(e) => setDraftCompleted(e.target.checked)} />}
                    label="Completed"
                  />
                )}
                <Box sx={{ display: 'flex', gap: 1, flexShrink: 0, pt: 0.5 }}>
                  <Button variant="contained" size="small" onClick={handleCreate} disabled={saving}>
                    Create
                  </Button>
                  <Button size="small" onClick={() => setComposerOpen(false)}>
                    Cancel
                  </Button>
                </Box>
              </Box>
            )}
            {!composerOpen && selected && (
              <Box sx={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0, gap: 1.25 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', flexShrink: 0 }}>
                  <Typography variant="subtitle2" fontWeight={600}>
                    Edit
                  </Typography>
                  <IconButton size="small" color="error" onClick={handleDelete} aria-label="Delete">
                    <DeleteOutlineOutlinedIcon fontSize="small" />
                  </IconButton>
                </Box>
                {AiToolbar}
                <TextField
                  fullWidth
                  size="small"
                  label="Title"
                  value={draftTitle}
                  onChange={(e) => setDraftTitle(e.target.value)}
                  sx={{ flexShrink: 0 }}
                />
                {selected.item_type === 'todo' && (
                  <DateTimePicker
                    label="Deadline"
                    value={draftDeadlineAt}
                    onChange={(nextValue) => setDraftDeadlineAt(nextValue)}
                    slotProps={{ textField: { size: 'small', fullWidth: true } }}
                  />
                )}
                <Box sx={{ flex: '1 1 0%', minHeight: 0, display: 'flex', flexDirection: 'column' }}>
                  <TextField
                    fullWidth
                    size="small"
                    label="Content"
                    multiline
                    value={draftBody}
                    onChange={(e) => setDraftBody(e.target.value)}
                    sx={{
                      flex: 1,
                      minHeight: 0,
                      '& .MuiOutlinedInput-root': {
                        height: '100%',
                        alignItems: 'flex-start',
                        py: 1,
                        boxSizing: 'border-box',
                      },
                      '& textarea': {
                        minHeight: '88px !important',
                        height: '100% !important',
                        maxHeight: '100%',
                        overflow: 'auto !important',
                        resize: 'none',
                        boxSizing: 'border-box',
                      },
                    }}
                  />
                </Box>
                {selected.item_type === 'todo' && (
                  <FormControlLabel
                    sx={{ flexShrink: 0, m: 0 }}
                    control={<Checkbox checked={draftCompleted} onChange={(e) => setDraftCompleted(e.target.checked)} />}
                    label="Completed"
                  />
                )}
                <Button
                  variant="contained"
                  size="small"
                  onClick={handleSave}
                  disabled={saving}
                  sx={{ flexShrink: 0, alignSelf: 'flex-start' }}
                >
                  {saving ? 'Saving…' : 'Save'}
                </Button>
              </Box>
            )}
            {!composerOpen && !selected && !loading && (
              <Typography variant="body2" color="text.secondary">
                Select an item or create a new one.
              </Typography>
            )}
          </Box>
        </Box>
      )}
      <Snackbar
        open={successToast.open}
        autoHideDuration={SUCCESS_TOAST_MS}
        onClose={closeSuccessToast}
        TransitionComponent={SlideUp}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
        sx={{
          zIndex: (theme) => theme.zIndex.snackbar + 2,
        }}
      >
        <Alert
          onClose={closeSuccessToast}
          severity="success"
          variant="filled"
          elevation={6}
          sx={{
            minWidth: 280,
            maxWidth: 480,
            boxShadow: 6,
            alignItems: 'center',
          }}
        >
          {successToast.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
