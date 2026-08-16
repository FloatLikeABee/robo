import React, { useState, useEffect, useMemo, useCallback } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  Tabs,
  Tab,
  TextField,
  IconButton,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
  Divider,
} from '@mui/material';
import AddOutlinedIcon from '@mui/icons-material/AddOutlined';
import DeleteOutlineOutlinedIcon from '@mui/icons-material/DeleteOutlineOutlined';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import { normalizeColorConfig, normalizeHex } from './colorTagResolve';
import {
  newRuleId,
  inferFieldKind,
  operatorsForFieldKind,
  defaultOperatorsForField,
} from './gridFilterUtils';

export default function ColorTagsModal({ open, onClose, gridKey, columns = [], config, onSaved }) {
  const [tab, setTab] = useState(0);
  const [rules, setRules] = useState([]);
  const [groups, setGroups] = useState([]);
  const [err, setErr] = useState(null);
  const [saving, setSaving] = useState(false);

  const fieldOptions = useMemo(() => {
    return (columns || [])
      .filter((c) => c && c.field)
      .map((c) => ({
        field: c.field,
        label: c.headerName != null && c.headerName !== '' ? c.headerName : c.field,
      }));
  }, [columns]);

  const resetFromConfig = useCallback(() => {
    const c = normalizeColorConfig(config);
    setRules(
      (c.rules || []).map((r) => ({
        id: r.id || newRuleId(),
        priority: Number(r.priority) || 0,
        field: r.field || fieldOptions[0]?.field || '',
        op: r.op || 'equals',
        value: r.value != null ? String(r.value) : '',
        color: normalizeHex(r.color) || '#3b82f6',
      }))
    );
    setGroups(
      (c.groups || []).map((g) => ({
        id: g.id || newRuleId(),
        priority: Number(g.priority) || 0,
        field: g.field || fieldOptions[0]?.field || '',
        value: g.value != null ? String(g.value) : '',
        color: normalizeHex(g.color) || '#a78bfa',
      }))
    );
  }, [config, fieldOptions]);

  useEffect(() => {
    if (!open) return;
    setTab(0);
    setErr(null);
    resetFromConfig();
  }, [open, resetFromConfig]);

  const handleSave = async () => {
    if (!gridKey) return;
    setSaving(true);
    setErr(null);
    try {
      const payload = {
        version: 1,
        rules: rules.map((r, i) => ({
          id: r.id,
          priority: r.priority != null ? Number(r.priority) : i,
          field: r.field,
          op: r.op || 'equals',
          value: r.value,
          color: normalizeHex(r.color) || '#3b82f6',
        })),
        groups: groups.map((g, i) => ({
          id: g.id,
          priority: g.priority != null ? Number(g.priority) : i,
          field: g.field,
          value: g.value,
          color: normalizeHex(g.color) || '#a78bfa',
        })),
      };
      await tranApi.put(tranEndpoints.gridColorConfig, {
        grid_key: gridKey,
        config_json: JSON.stringify(payload),
      });
      if (onSaved) onSaved(payload);
      onClose();
    } catch (e) {
      setErr(e.response?.data?.error || e.message || 'Save failed');
    } finally {
      setSaving(false);
    }
  };

  const addRule = () => {
    const f = fieldOptions[0]?.field || '';
    const ops = defaultOperatorsForField(f);
    setRules((prev) => [
      ...prev,
      {
        id: newRuleId(),
        priority: prev.length,
        field: f,
        op: ops[0]?.value || 'contains',
        value: '',
        color: '#3b82f6',
      },
    ]);
  };

  const addGroup = () => {
    setGroups((prev) => [
      ...prev,
      {
        id: newRuleId(),
        priority: prev.length,
        field: fieldOptions[0]?.field || '',
        value: '',
        color: '#a78bfa',
      },
    ]);
  };

  const moveRule = (idx, dir) => {
    setRules((prev) => {
      const j = idx + dir;
      if (j < 0 || j >= prev.length) return prev;
      const next = [...prev];
      [next[idx], next[j]] = [next[j], next[idx]];
      return next.map((r, i) => ({ ...r, priority: i }));
    });
  };

  const moveGroup = (idx, dir) => {
    setGroups((prev) => {
      const j = idx + dir;
      if (j < 0 || j >= prev.length) return prev;
      const next = [...prev];
      [next[idx], next[j]] = [next[j], next[idx]];
      return next.map((g, i) => ({ ...g, priority: i }));
    });
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth scroll="paper">
      <DialogTitle>Color tags</DialogTitle>
      <DialogContent dividers>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
          Row colors are display-only (not stored on records). Priority:{' '}
          <strong>rules</strong> (by priority) first, then <strong>groups</strong> (by priority).
        </Typography>
        {err && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setErr(null)}>
            {err}
          </Alert>
        )}
        <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ mb: 2 }}>
          <Tab label="Rules" />
          <Tab label="Groups" />
        </Tabs>

        {tab === 0 && (
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            <Typography variant="caption" color="text.secondary">
              First matching rule wins (lower priority number runs first).
            </Typography>
            {rules.map((rule, idx) => {
              const kind = inferFieldKind(rule.field);
              const ops = operatorsForFieldKind(kind);
              return (
                <Box
                  key={rule.id}
                  sx={{
                    display: 'grid',
                    gridTemplateColumns: { xs: '1fr', sm: '72px repeat(3, minmax(0,1fr)) 100px 36px 36px auto' },
                    gap: 1,
                    alignItems: 'center',
                    p: 1,
                    border: 1,
                    borderColor: 'divider',
                    borderRadius: 1,
                  }}
                >
                  <TextField
                    size="small"
                    label="Priority"
                    type="number"
                    value={rule.priority}
                    onChange={(e) =>
                      setRules((prev) =>
                        prev.map((r) => (r.id === rule.id ? { ...r, priority: Number(e.target.value) } : r))
                      )
                    }
                  />
                  <FormControl size="small" fullWidth>
                    <InputLabel>Field</InputLabel>
                    <Select
                      label="Field"
                      value={rule.field}
                      onChange={(e) => {
                        const f = e.target.value;
                        const o = defaultOperatorsForField(f);
                        setRules((prev) =>
                          prev.map((r) =>
                            r.id === rule.id ? { ...r, field: f, op: o[0]?.value || 'contains', value: '' } : r
                          )
                        );
                      }}
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
                      onChange={(e) =>
                        setRules((prev) =>
                          prev.map((r) => (r.id === rule.id ? { ...r, op: e.target.value } : r))
                        )
                      }
                    >
                      {ops.map((o) => (
                        <MenuItem key={o.value} value={o.value}>
                          {o.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  <TextField
                    size="small"
                    label="Value"
                    value={rule.value}
                    onChange={(e) =>
                      setRules((prev) =>
                        prev.map((r) => (r.id === rule.id ? { ...r, value: e.target.value } : r))
                      )
                    }
                    fullWidth
                  />
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    <TextField
                      size="small"
                      label="Hex"
                      value={rule.color}
                      onChange={(e) =>
                        setRules((prev) =>
                          prev.map((r) => (r.id === rule.id ? { ...r, color: e.target.value } : r))
                        )
                      }
                      sx={{ width: 88 }}
                    />
                    <input
                      type="color"
                      value={normalizeHex(rule.color) || '#3b82f6'}
                      onChange={(e) =>
                        setRules((prev) =>
                          prev.map((r) => (r.id === rule.id ? { ...r, color: e.target.value } : r))
                        )
                      }
                      style={{ width: 32, height: 28, padding: 0, border: 'none' }}
                    />
                  </Box>
                  <IconButton size="small" onClick={() => moveRule(idx, -1)} disabled={idx === 0}>
                    <KeyboardArrowUpIcon fontSize="small" />
                  </IconButton>
                  <IconButton size="small" onClick={() => moveRule(idx, 1)} disabled={idx === rules.length - 1}>
                    <KeyboardArrowDownIcon fontSize="small" />
                  </IconButton>
                  <IconButton size="small" color="error" onClick={() => setRules((prev) => prev.filter((r) => r.id !== rule.id))}>
                    <DeleteOutlineOutlinedIcon fontSize="small" />
                  </IconButton>
                </Box>
              );
            })}
            <Button size="small" startIcon={<AddOutlinedIcon />} onClick={addRule} disabled={!fieldOptions.length}>
              Add rule
            </Button>
          </Box>
        )}

        {tab === 1 && (
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            <Typography variant="caption" color="text.secondary">
              Groups match when field <strong>equals</strong> value (string compare). Evaluated after rules.
            </Typography>
            {groups.map((g, idx) => (
              <Box
                key={g.id}
                sx={{
                  display: 'grid',
                  gridTemplateColumns: { xs: '1fr', sm: '72px repeat(2, minmax(0,1fr)) 100px 36px 36px auto' },
                  gap: 1,
                  alignItems: 'center',
                  p: 1,
                  border: 1,
                  borderColor: 'divider',
                  borderRadius: 1,
                }}
              >
                <TextField
                  size="small"
                  label="Priority"
                  type="number"
                  value={g.priority}
                  onChange={(e) =>
                    setGroups((prev) =>
                      prev.map((x) => (x.id === g.id ? { ...x, priority: Number(e.target.value) } : x))
                    )
                  }
                />
                <FormControl size="small" fullWidth>
                  <InputLabel>Field</InputLabel>
                  <Select
                    label="Field"
                    value={g.field}
                    onChange={(e) =>
                      setGroups((prev) =>
                        prev.map((x) => (x.id === g.id ? { ...x, field: e.target.value } : x))
                      )
                    }
                  >
                    {fieldOptions.map((o) => (
                      <MenuItem key={o.field} value={o.field}>
                        {o.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
                <TextField
                  size="small"
                  label="Value"
                  value={g.value}
                  onChange={(e) =>
                    setGroups((prev) =>
                      prev.map((x) => (x.id === g.id ? { ...x, value: e.target.value } : x))
                    )
                  }
                  fullWidth
                />
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  <TextField
                    size="small"
                    label="Hex"
                    value={g.color}
                    onChange={(e) =>
                      setGroups((prev) =>
                        prev.map((x) => (x.id === g.id ? { ...x, color: e.target.value } : x))
                      )
                    }
                    sx={{ width: 88 }}
                  />
                  <input
                    type="color"
                    value={normalizeHex(g.color) || '#a78bfa'}
                    onChange={(e) =>
                      setGroups((prev) =>
                        prev.map((x) => (x.id === g.id ? { ...x, color: e.target.value } : x))
                      )
                    }
                    style={{ width: 32, height: 28, padding: 0, border: 'none' }}
                  />
                </Box>
                <IconButton size="small" onClick={() => moveGroup(idx, -1)} disabled={idx === 0}>
                  <KeyboardArrowUpIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" onClick={() => moveGroup(idx, 1)} disabled={idx === groups.length - 1}>
                  <KeyboardArrowDownIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" color="error" onClick={() => setGroups((prev) => prev.filter((x) => x.id !== g.id))}>
                  <DeleteOutlineOutlinedIcon fontSize="small" />
                </IconButton>
              </Box>
            ))}
            <Button size="small" startIcon={<AddOutlinedIcon />} onClick={addGroup} disabled={!fieldOptions.length}>
              Add group
            </Button>
          </Box>
        )}

        <Divider sx={{ my: 2 }} />
        <Typography variant="caption" color="text.secondary">
          Tip: use rules for conditions; use groups to color by a single field value (e.g. status).
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" color="primary" onClick={handleSave} disabled={saving || !gridKey} sx={{ fontWeight: 700 }}>
          {saving ? 'Saving…' : 'Save'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
