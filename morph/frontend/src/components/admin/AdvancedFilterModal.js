import React, { useState, useEffect, useMemo, useCallback } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TextField,
  IconButton,
  Divider,
  List,
  ListItem,
  ListItemText,
  CircularProgress,
  Alert,
} from '@mui/material';
import AddOutlinedIcon from '@mui/icons-material/AddOutlined';
import DeleteOutlineOutlinedIcon from '@mui/icons-material/DeleteOutlineOutlined';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import {
  newRuleId,
  defaultOperatorsForField,
  operatorsForFieldKind,
  inferFieldKind,
} from './gridFilterUtils';
import { useConfirm } from '../ConfirmDialog';

function parseSavedFilterJson(raw) {
  if (raw == null || String(raw).trim() === '') return { rules: [] };
  try {
    const p = typeof raw === 'string' ? JSON.parse(raw) : raw;
    if (p && Array.isArray(p.rules)) return { rules: p.rules };
    if (Array.isArray(p)) return { rules: p };
    return { rules: [] };
  } catch {
    return { rules: [] };
  }
}

/** API may return a bare array or an object wrapper; always return an array. */
function normalizeSavedListPayload(data) {
  if (Array.isArray(data)) return data;
  if (data && Array.isArray(data.data)) return data.data;
  if (data && Array.isArray(data.items)) return data.items;
  return [];
}

/**
 * Modal: column-based filter rules + load/save shared filters (all users).
 * @param {object} props
 * @param {boolean} props.open
 * @param {function} props.onClose
 * @param {string} props.gridKey — stable key per grid (e.g. students, staff)
 * @param {{ field: string, headerName?: string }[]} props.columns
 * @param {Array<{ id: string, field: string, op: string, value: string }>} props.rules
 * @param {function} props.setRules
 */
export default function AdvancedFilterModal({
  open,
  onClose,
  gridKey,
  columns = [],
  rules,
  setRules,
}) {
  const { confirm } = useConfirm();
  const [savedList, setSavedList] = useState([]);
  const [savedLoading, setSavedLoading] = useState(false);
  const [savedErr, setSavedErr] = useState(null);
  const [saveName, setSaveName] = useState('');
  const [saveBusy, setSaveBusy] = useState(false);
  const [selectedSavedId, setSelectedSavedId] = useState('');
  const [deleteBusy, setDeleteBusy] = useState(false);

  const fieldOptions = useMemo(() => {
    const cols = columns || [];
    return cols
      .filter((c) => c && c.field)
      .map((c) => ({
        field: c.field,
        label: c.headerName != null && c.headerName !== '' ? c.headerName : c.field,
      }));
  }, [columns]);

  const savedRows = Array.isArray(savedList) ? savedList : [];

  const loadSaved = useCallback(() => {
    if (!gridKey) return;
    setSavedLoading(true);
    setSavedErr(null);
    tranApi
      .get(`${tranEndpoints.gridSavedFilters}?grid_key=${encodeURIComponent(gridKey)}`)
      .then((res) => {
        setSavedList(normalizeSavedListPayload(res.data));
      })
      .catch((err) => {
        setSavedErr(err.response?.data?.error || err.message || 'Failed to load saved filters');
        setSavedList([]);
      })
      .finally(() => setSavedLoading(false));
  }, [gridKey]);

  useEffect(() => {
    if (!open) return;
    loadSaved();
  }, [open, loadSaved]);

  const addRule = () => {
    const firstField = fieldOptions[0]?.field || '';
    const ops = defaultOperatorsForField(firstField);
    setRules((prev) => [
      ...prev,
      {
        id: newRuleId(),
        field: firstField,
        op: ops[0]?.value || 'contains',
        value: '',
      },
    ]);
  };

  const updateRule = (id, patch) => {
    setRules((prev) =>
      prev.map((r) => {
        if (r.id !== id) return r;
        const next = { ...r, ...patch };
        if (patch.field != null) {
          const ops = defaultOperatorsForField(patch.field);
          next.op = ops[0]?.value || 'contains';
          next.value = '';
        }
        return next;
      })
    );
  };

  const removeRule = (id) => {
    setRules((prev) => prev.filter((r) => r.id !== id));
  };

  const handleSaveNamed = async () => {
    const name = saveName.trim();
    if (!name) return;
    setSaveBusy(true);
    setSavedErr(null);
    try {
      const filter_json = JSON.stringify({ rules, version: 1 });
      await tranApi.post(tranEndpoints.gridSavedFilters, {
        grid_key: gridKey,
        name,
        filter_json,
      });
      setSaveName('');
      await loadSaved();
    } catch (err) {
      setSavedErr(err.response?.data?.error || err.message || 'Save failed');
    } finally {
      setSaveBusy(false);
    }
  };

  const handleLoadSaved = () => {
    const id = Number(selectedSavedId);
    if (!Number.isFinite(id) || id <= 0) return;
    const row = savedRows.find((s) => s.id === id);
    if (!row) return;
    const { rules: loaded } = parseSavedFilterJson(row.filter_json);
    const normalized = (loaded || []).map((r) => ({
      id: newRuleId(),
      field: r.field || fieldOptions[0]?.field || '',
      op: r.op || 'contains',
      value: r.value != null ? String(r.value) : '',
    }));
    setRules(normalized.length ? normalized : []);
  };

  const handleDeleteSaved = async () => {
    const id = Number(selectedSavedId);
    if (!Number.isFinite(id) || id <= 0) return;
    const ok = await confirm({
      title: 'Delete saved filter',
      message: 'Delete this saved filter for all users?',
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    setDeleteBusy(true);
    setSavedErr(null);
    try {
      await tranApi.delete(`${tranEndpoints.gridSavedFilters}/${id}`);
      setSelectedSavedId('');
      await loadSaved();
    } catch (err) {
      setSavedErr(err.response?.data?.error || err.message || 'Delete failed');
    } finally {
      setDeleteBusy(false);
    }
  };

  const valueHidden = (op) =>
    op === 'empty' || op === 'not_empty' || op === 'bool_true' || op === 'bool_false';

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth scroll="paper">
      <DialogTitle>Filter & search</DialogTitle>
      <DialogContent dividers>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Add one or more conditions (all must match). This works together with the quick search box in the toolbar.
        </Typography>
        {savedErr && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setSavedErr(null)}>
            {savedErr}
          </Alert>
        )}

        <Typography variant="subtitle2" fontWeight={700} sx={{ mb: 1 }}>
          Conditions
        </Typography>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
          {rules.length === 0 && (
            <Typography variant="body2" color="text.secondary">
              No conditions yet. Click &quot;Add condition&quot; or load a saved filter.
            </Typography>
          )}
          {rules.map((rule) => {
            const kind = inferFieldKind(rule.field);
            const ops = operatorsForFieldKind(kind);
            return (
              <Box
                key={rule.id}
                sx={{
                  display: 'grid',
                  gridTemplateColumns: { xs: '1fr', sm: 'minmax(120px,1fr) minmax(120px,1fr) minmax(0,1.2fr) auto' },
                  gap: 1,
                  alignItems: 'center',
                }}
              >
                <FormControl size="small" fullWidth>
                  <InputLabel>Field</InputLabel>
                  <Select
                    label="Field"
                    value={rule.field}
                    onChange={(e) => updateRule(rule.id, { field: e.target.value })}
                  >
                    {fieldOptions.map((o) => (
                      <MenuItem key={o.field} value={o.field}>
                        {o.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <FormControl size="small" fullWidth>
                  <InputLabel>Operator</InputLabel>
                  <Select
                    label="Operator"
                    value={rule.op}
                    onChange={(e) => updateRule(rule.id, { op: e.target.value })}
                  >
                    {ops.map((o) => (
                      <MenuItem key={o.value} value={o.value}>
                        {o.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
                {!valueHidden(rule.op) ? (
                  <TextField
                    size="small"
                    label="Value"
                    value={rule.value}
                    onChange={(e) => updateRule(rule.id, { value: e.target.value })}
                    fullWidth
                  />
                ) : (
                  <Box />
                )}
                <IconButton size="small" color="error" onClick={() => removeRule(rule.id)} aria-label="Remove condition">
                  <DeleteOutlineOutlinedIcon fontSize="small" />
                </IconButton>
              </Box>
            );
          })}
        </Box>
        <Button
          size="small"
          variant="outlined"
          color="primary"
          startIcon={<AddOutlinedIcon />}
          onClick={addRule}
          sx={{ mt: 1, fontWeight: 600 }}
          disabled={!fieldOptions.length}
        >
          Add condition
        </Button>

        <Divider sx={{ my: 2 }} />

        <Typography variant="subtitle2" fontWeight={700} sx={{ mb: 1 }}>
          Saved filters (shared for all users)
        </Typography>
        {savedLoading ? (
          <CircularProgress size={24} />
        ) : (
          <>
            {savedRows.length > 0 && (
              <List dense disablePadding sx={{ mb: 1, maxHeight: 160, overflow: 'auto', border: 1, borderColor: 'divider', borderRadius: 1 }}>
                {savedRows.map((s) => (
                  <ListItem key={s.id} disablePadding>
                    <ListItemText primary={s.name} secondary={s.created_on ? `Updated ${s.created_on}` : null} />
                  </ListItem>
                ))}
              </List>
            )}
            <Box
              sx={{
                display: 'flex',
                flexWrap: 'wrap',
                gap: 1,
                alignItems: 'center',
                mb: 2,
                '& > .MuiFormControl-root': { flex: '0 1 auto' },
              }}
            >
              <FormControl size="small" sx={{ minWidth: 160, maxWidth: 240 }}>
                <InputLabel>Select saved</InputLabel>
                <Select
                  label="Select saved"
                  value={selectedSavedId}
                  onChange={(e) => setSelectedSavedId(e.target.value)}
                >
                  <MenuItem value="">
                    <em>Choose…</em>
                  </MenuItem>
                  {savedRows.map((s) => (
                    <MenuItem key={s.id} value={String(s.id)}>
                      {s.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              <Button size="small" variant="outlined" color="primary" onClick={handleLoadSaved} disabled={!selectedSavedId} sx={{ fontWeight: 600 }}>
                Load
              </Button>
              <Button
                size="small"
                variant="outlined"
                color="error"
                onClick={handleDeleteSaved}
                disabled={!selectedSavedId || deleteBusy}
                sx={{ fontWeight: 600 }}
              >
                Delete
              </Button>
              <TextField
                size="small"
                label="Save current conditions as"
                value={saveName}
                onChange={(e) => setSaveName(e.target.value)}
                sx={{ flex: '1 1 200px', minWidth: 180, maxWidth: 360 }}
              />
              <Button
                size="small"
                variant="contained"
                color="primary"
                onClick={handleSaveNamed}
                disabled={saveBusy || !saveName.trim() || !rules.some((r) => r.field)}
                sx={{
                  fontWeight: 700,
                  flexShrink: 0,
                  color: (theme) => theme.palette.primary.contrastText,
                  '&.Mui-disabled': {
                    color: (theme) => theme.palette.action.disabled,
                  },
                }}
              >
                {saveBusy ? 'Saving…' : 'Save'}
              </Button>
            </Box>
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button
          onClick={() => {
            setRules([]);
            setSaveName('');
          }}
        >
          Clear conditions
        </Button>
        <Button onClick={onClose}>Close</Button>
        <Button
          variant="contained"
          color="primary"
          onClick={onClose}
          sx={{
            fontWeight: 700,
            color: (theme) => theme.palette.primary.contrastText,
          }}
        >
          Apply
        </Button>
      </DialogActions>
    </Dialog>
  );
}
