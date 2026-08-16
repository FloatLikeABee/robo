import React, { useState, useEffect, useCallback } from 'react';
import {
  Typography,
  Box,
  Paper,
  Stack,
  Alert,
  CircularProgress,
  TextField,
  Button,
} from '@mui/material';
import { alpha } from '@mui/material/styles';
import { tranApi, tranEndpoints } from '../../api/tranClient';

export default function UserSettings() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');

  const load = useCallback(() => {
    return tranApi.get(tranEndpoints.userMe).then((res) => {
      const u = res.data || {};
      setFirstName(u.first_name || '');
      setLastName(u.last_name || '');
      setEmail(u.email || '');
      setPhone(u.phone || '');
    });
  }, []);

  useEffect(() => {
    setLoading(true);
    setError(null);
    load()
      .catch((err) =>
        setError(err.response?.data?.error || err.message || 'Failed to load your profile')
      )
      .finally(() => setLoading(false));
  }, [load]);

  const onSave = async (e) => {
    e.preventDefault();
    if (!lastName.trim()) {
      setError('Last name is required');
      return;
    }
    setSaving(true);
    setError(null);
    setSuccess('');
    try {
      await tranApi.put(tranEndpoints.userMe, {
        first_name: firstName.trim() || null,
        last_name: lastName.trim(),
        email: email.trim() || null,
        phone: phone.trim() || null,
      });
      setSuccess('Settings saved');
      await load();
    } catch (err) {
      setError(err.response?.data?.error || err.message || 'Save failed');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Box
      sx={{
        width: '100%',
        maxWidth: 640,
        flex: 1,
        minHeight: 0,
        display: 'flex',
        flexDirection: 'column',
        gap: 2,
      }}
    >
      <Box sx={{ flexShrink: 0 }}>
        <Typography variant="h5" sx={{ color: 'primary.main', fontWeight: 700, mb: 0.5 }}>
          User settings
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Update your Morph profile.
        </Typography>
      </Box>

      {error && (
        <Alert severity="error" onClose={() => setError(null)} sx={{ flexShrink: 0 }}>
          {error}
        </Alert>
      )}
      {success && (
        <Alert severity="success" onClose={() => setSuccess('')} sx={{ flexShrink: 0 }}>
          {success}
        </Alert>
      )}

      <Paper
        elevation={0}
        component="form"
        onSubmit={onSave}
        sx={(theme) => ({
          flex: 1,
          minHeight: 0,
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
          borderRadius: 2,
          border: '1px solid',
          borderColor: 'divider',
          background:
            theme.palette.mode === 'dark'
              ? alpha(theme.palette.primary.main, 0.08)
              : alpha(theme.palette.primary.main, 0.06),
        })}
      >
        <Box sx={{ flex: 1, minHeight: 0, overflow: 'auto', p: { xs: 1.5, sm: 3 } }}>
          {loading ? (
            <CircularProgress size={32} sx={{ display: 'block', mx: 'auto' }} />
          ) : (
            <Stack spacing={2.5}>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                <TextField
                  label="First name"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  fullWidth
                  autoComplete="given-name"
                />
                <TextField
                  label="Last name"
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                  fullWidth
                  required
                  autoComplete="family-name"
                />
              </Stack>
              <TextField
                label="Email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                fullWidth
                autoComplete="email"
              />
              <TextField
                label="Phone"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                fullWidth
                autoComplete="tel"
              />
            </Stack>
          )}
        </Box>

        <Box
          sx={{
            flexShrink: 0,
            px: 3,
            py: 2,
            borderTop: 1,
            borderColor: 'divider',
            display: 'flex',
            justifyContent: 'flex-end',
          }}
        >
          <Button type="submit" variant="contained" disabled={loading || saving}>
            {saving ? 'Saving…' : 'Save'}
          </Button>
        </Box>
      </Paper>
    </Box>
  );
}
