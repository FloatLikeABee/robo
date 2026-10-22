import React, { useState, useEffect } from 'react';
import {
  Typography,
  Box,
  Paper,
  Button,
  TextField,
  Stack,
  Alert,
  CircularProgress,
  Grid,
} from '@mui/material';
import SaveOutlinedIcon from '@mui/icons-material/SaveOutlined';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import {
  defaultPlatformUILabels,
  platformUILabelHints,
  displayLabelSettingsGroups,
} from '../../platformUiDefaults';

function LabelFields({ keys, labels, overrides, fieldErrors, onChange, onResetKey }) {
  return (
    <Stack spacing={1.25}>
      {keys.map((key) => (
        <Box key={key}>
          <TextField
            label={key.replace(/_/g, ' ')}
            value={labels[key] ?? ''}
            onChange={(e) => onChange(key, e.target.value)}
            fullWidth
            size="small"
            required
            error={Boolean(fieldErrors[key])}
            helperText={
              (fieldErrors[key] ? `${fieldErrors[key]} ` : '') +
              (platformUILabelHints[key] || '') +
              (overrides[key] ? ' — customized' : '') +
              (labels[key] !== defaultPlatformUILabels[key]
                ? ` (default: ${defaultPlatformUILabels[key]})`
                : '')
            }
            FormHelperTextProps={{ sx: { mt: 0.5, lineHeight: 1.35 } }}
          />
          {labels[key] !== defaultPlatformUILabels[key] && (
            <Button size="small" onClick={() => onResetKey(key)} sx={{ mt: 0.25, py: 0, minHeight: 0 }}>
              Reset to default
            </Button>
          )}
        </Box>
      ))}
    </Stack>
  );
}

export default function DisplayLabelsSettings() {
  const [labels, setLabels] = useState(defaultPlatformUILabels);
  const [overrides, setOverrides] = useState({});
  const [fieldErrors, setFieldErrors] = useState({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [savedMsg, setSavedMsg] = useState(null);

  const load = () => {
    setError(null);
    setSavedMsg(null);
    tranApi
      .get(tranEndpoints.platformUiConfig)
      .then((res) => {
        const merged = res.data?.labels && typeof res.data.labels === 'object' ? res.data.labels : {};
        const ovr = res.data?.overrides && typeof res.data.overrides === 'object' ? res.data.overrides : {};
        setLabels({ ...defaultPlatformUILabels, ...merged });
        setOverrides(ovr);
      })
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load();
  }, []);

  const handleChange = (key, value) => {
    setLabels((prev) => ({ ...prev, [key]: value }));
    setFieldErrors((prev) => {
      if (!prev[key]) return prev;
      const next = { ...prev };
      delete next[key];
      return next;
    });
  };

  const handleResetKey = (key) => {
    setLabels((prev) => ({ ...prev, [key]: defaultPlatformUILabels[key] }));
  };

  const save = () => {
    setError(null);
    setSavedMsg(null);
    setFieldErrors({});
    const payload = { labels: {} };
    const nextFieldErrors = {};
    for (const key of Object.keys(defaultPlatformUILabels)) {
      const v = (labels[key] ?? '').trim();
      if (v === '') {
        nextFieldErrors[key] = 'Cannot be empty.';
      }
      payload.labels[key] = v;
    }
    if (Object.keys(nextFieldErrors).length > 0) {
      setFieldErrors(nextFieldErrors);
      setError('Display names cannot be empty.');
      return;
    }
    setSaving(true);
    tranApi
      .put(tranEndpoints.platformUiConfig, payload)
      .then((res) => {
        const merged = res.data?.labels && typeof res.data.labels === 'object' ? res.data.labels : {};
        const ovr = res.data?.overrides && typeof res.data.overrides === 'object' ? res.data.overrides : {};
        setLabels({ ...defaultPlatformUILabels, ...merged });
        setOverrides(ovr);
        setSavedMsg('Saved. Navigation and page titles will use these names.');
      })
      .catch((err) => setError(err.response?.data?.error || err.message || 'Save failed'))
      .finally(() => setSaving(false));
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box
      sx={{
        flex: 1,
        minHeight: 0,
        display: 'flex',
        flexDirection: 'column',
        width: '100%',
        maxWidth: 1100,
        mx: 'auto',
      }}
    >
      <Box
        sx={{
          flex: 1,
          minHeight: 0,
          overflowY: 'auto',
          overflowX: 'hidden',
          pr: 0.5,
        }}
      >
        <Typography variant="h5" sx={{ mb: 1 }}>
          Display names
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Defaults use generic terms (Facility, Participant, …). Override any label to match your organization (e.g.
          Places → “Sites”, People → “Team”). Saving requires the admin database connection.
        </Typography>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
            {error}
          </Alert>
        )}
        {savedMsg && (
          <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSavedMsg(null)}>
            {savedMsg}
          </Alert>
        )}

        <Grid container spacing={2} alignItems="flex-start">
          <Grid item xs={12} md={6}>
            <Paper
              sx={{
                p: 2,
                height: '100%',
              }}
            >
              <Typography variant="subtitle1" fontWeight={700} sx={{ mb: 1.5, color: 'text.primary' }}>
                Branding & navigation
              </Typography>
              <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1.5 }}>
                App name, assistant name, and menu labels (each main-module label is also the grid page title)
              </Typography>
              <LabelFields
                keys={displayLabelSettingsGroups.brandingAndNavigation}
                labels={labels}
                overrides={overrides}
                fieldErrors={fieldErrors}
                onChange={handleChange}
                onResetKey={handleResetKey}
              />
            </Paper>
          </Grid>
          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 2, height: '100%' }}>
              <Typography variant="subtitle1" fontWeight={700} sx={{ mb: 1.5, color: 'text.primary' }}>
                Terms & table columns
              </Typography>
              <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1.5 }}>
                Facility wording, empty-state text, and trips-related column headers
              </Typography>
              <LabelFields
                keys={displayLabelSettingsGroups.termsAndTableColumns}
                labels={labels}
                overrides={overrides}
                fieldErrors={fieldErrors}
                onChange={handleChange}
                onResetKey={handleResetKey}
              />
            </Paper>
          </Grid>
        </Grid>
      </Box>

      <Box
        sx={{
          flexShrink: 0,
          pt: 2,
          pb: 0.5,
          borderTop: 1,
          borderColor: 'divider',
          bgcolor: 'background.default',
        }}
      >
        <Button
          variant="contained"
          startIcon={saving ? <CircularProgress size={18} color="inherit" /> : <SaveOutlinedIcon />}
          onClick={save}
          disabled={saving}
        >
          Save display names
        </Button>
      </Box>
    </Box>
  );
}
