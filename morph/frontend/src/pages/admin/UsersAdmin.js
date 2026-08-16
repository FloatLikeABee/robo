import React, { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  IconButton,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import EditOutlinedIcon from '@mui/icons-material/EditOutlined';
import PersonAddAltIcon from '@mui/icons-material/PersonAddAlt';
import { Navigate } from 'react-router-dom';
import { tranApi } from '../../api/tranClient';
import { isMorphAdmin } from '../../auth/isMorphAdmin';
import { useAdminBasePath } from '../../adminPaths';
import { useConfirm } from '../../components/ConfirmDialog';

export default function UsersAdmin() {
  const { confirm } = useConfirm();
  const base = useAdminBasePath();
  const [users, setUsers] = useState([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [editUser, setEditUser] = useState(null);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isAdmin, setIsAdmin] = useState(false);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const { data } = await tranApi.get('/api/admin/users');
      setUsers(Array.isArray(data.users) ? data.users : []);
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Failed to load users');
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  if (!isMorphAdmin()) {
    return <Navigate to={`${base}/configuration/display-labels`} replace />;
  }

  const openCreate = () => {
    setEmail('');
    setPassword('');
    setIsAdmin(false);
    setCreateOpen(true);
  };

  const openEdit = (u) => {
    setEditUser(u);
    setEmail(u.email || '');
    setPassword('');
    setIsAdmin(Boolean(u.is_admin));
  };

  const createUser = async () => {
    setSaving(true);
    setError('');
    try {
      await tranApi.post('/api/admin/users', {
        email: email.trim(),
        password,
        is_admin: isAdmin,
      });
      setCreateOpen(false);
      await load();
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Create failed');
    } finally {
      setSaving(false);
    }
  };

  const saveEdit = async () => {
    if (!editUser) return;
    setSaving(true);
    setError('');
    try {
      const body = { email: email.trim(), is_admin: isAdmin };
      if (password.trim()) body.password = password;
      await tranApi.patch(`/api/admin/users/${editUser.id}`, body);
      setEditUser(null);
      await load();
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Update failed');
    } finally {
      setSaving(false);
    }
  };

  const removeUser = async (u) => {
    const ok = await confirm({
      title: 'Delete user',
      message: `Delete user ${u.email}?`,
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    setError('');
    try {
      await tranApi.delete(`/api/admin/users/${u.id}`);
      await load();
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Delete failed');
    }
  };

  return (
    <Box sx={{ p: 2, maxWidth: 960 }}>
      <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
        <Box>
          <Typography variant="h5" fontWeight={700}>
            Users
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Login accounts (email + password). Every user can use all Morph apps; only admins manage users and data import.
          </Typography>
        </Box>
        <Button variant="contained" startIcon={<PersonAddAltIcon />} onClick={openCreate}>
          Add user
        </Button>
      </Stack>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}
      <Paper variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Email</TableCell>
              <TableCell>Role</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={3}>Loading…</TableCell>
              </TableRow>
            ) : users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={3}>No users yet.</TableCell>
              </TableRow>
            ) : (
              users.map((u) => (
                <TableRow key={u.id}>
                  <TableCell>{u.email}</TableCell>
                  <TableCell>{u.is_admin ? 'Admin' : 'User'}</TableCell>
                  <TableCell align="right">
                    <IconButton size="small" aria-label="Edit" onClick={() => openEdit(u)}>
                      <EditOutlinedIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" aria-label="Delete" onClick={() => removeUser(u)}>
                      <DeleteOutlineIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </Paper>

      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>Add user</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} fullWidth required />
            <TextField label="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} fullWidth required />
            <FormControlLabel control={<Checkbox checked={isAdmin} onChange={(e) => setIsAdmin(e.target.checked)} />} label="Admin (can manage users & data import)" />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)}>Cancel</Button>
          <Button variant="contained" disabled={saving || !email.trim() || !password} onClick={createUser}>
            Create
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={Boolean(editUser)} onClose={() => setEditUser(null)} fullWidth maxWidth="sm">
        <DialogTitle>Edit user</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} fullWidth required />
            <TextField
              label="New password (optional)"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              fullWidth
            />
            <FormControlLabel control={<Checkbox checked={isAdmin} onChange={(e) => setIsAdmin(e.target.checked)} />} label="Admin" />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditUser(null)}>Cancel</Button>
          <Button variant="contained" disabled={saving || !email.trim()} onClick={saveEdit}>
            Save
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
