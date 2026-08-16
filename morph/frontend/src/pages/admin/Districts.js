import React, { useState, useEffect, useMemo } from 'react';
import {
  Typography,
  Box,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  CircularProgress,
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  IconButton,
  Stack,
  Badge,
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import FilterListIcon from '@mui/icons-material/FilterList';
import AddOutlinedIcon from '@mui/icons-material/AddOutlined';
import EditOutlinedIcon from '@mui/icons-material/EditOutlined';
import DeleteOutlineOutlinedIcon from '@mui/icons-material/DeleteOutlineOutlined';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import AdvancedFilterModal from '../../components/admin/AdvancedFilterModal';
import { applyQuickSearchThenRules } from '../../components/admin/gridFilterUtils';
import { sanitizeTextInput } from '../../utils/inputValidation';
import { useConfirm } from '../../components/ConfirmDialog';

const DISTRICTS_GRID_KEY = 'districts';
const DISTRICTS_FILTER_COLUMNS = [
  { field: 'id', headerName: 'ID' },
  { field: 'district', headerName: 'District' },
  { field: 'name', headerName: 'Name' },
];

export default function Districts() {
  const { confirm } = useConfirm();
  const [list, setList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingRow, setEditingRow] = useState(null);
  const [form, setForm] = useState({
    district: '',
    name: '',
    detail: '{}',
  });
  const [saving, setSaving] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [filterRules, setFilterRules] = useState([]);
  const [filterModalOpen, setFilterModalOpen] = useState(false);
  const cleanText = (v) => sanitizeTextInput(v);

  const activeFilterCount = useMemo(
    () => filterRules.filter((r) => r && r.field).length,
    [filterRules],
  );

  const filteredList = useMemo(
    () => applyQuickSearchThenRules(list, searchText, filterRules),
    [list, searchText, filterRules],
  );

  const loadOne = (id) => tranApi.get(tranEndpoints.district(id)).then((res) => res.data);

  useEffect(() => {
    tranApi
      .get(tranEndpoints.districts)
      .then((res) => setList(res.data || []))
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load districts'))
      .finally(() => setLoading(false));
  }, []);

  const refresh = () => tranApi.get(tranEndpoints.districts).then((res) => setList(res.data || []));

  const openAddDialog = () => {
    setEditingRow(null);
    setForm({ district: '', name: '', detail: '{}' });
    setDialogOpen(true);
  };

  const openEditDialog = async (row) => {
    setEditingRow(row);
    setForm({ district: row.district || '', name: row.name || '', detail: '{}' });
    setDialogOpen(true);
    try {
      const full = await loadOne(row.id);
      if (full && full.detail != null) {
        setForm((f) => ({ ...f, detail: typeof full.detail === 'string' ? full.detail : JSON.stringify(full.detail, null, 2) }));
      }
    } catch {
      /* keep {} */
    }
  };

  const handleSave = () => {
    setSaving(true);
    setError(null);
    const payload = {
      district: form.district,
      name: form.name,
      detail: form.detail,
    };
    const req = editingRow
      ? tranApi.put(tranEndpoints.district(editingRow.id), payload)
      : tranApi.post(tranEndpoints.districts, payload);
    req
      .then(() => refresh())
      .then(() => {
        setDialogOpen(false);
        setSaving(false);
      })
      .catch((err) => {
        setSaving(false);
        setError(err.response?.data?.error || err.message || 'Failed to save district');
      });
  };

  const handleDelete = async (row) => {
    const ok = await confirm({
      title: 'Delete district',
      message: `Delete district ${row.district || row.name || row.id}?`,
      confirmLabel: 'Delete',
      danger: true,
    });
    if (!ok) return;
    setError(null);
    tranApi
      .delete(tranEndpoints.district(row.id))
      .then(() => {
        setList((prev) => prev.filter((r) => r.id !== row.id));
      })
      .catch((err) => {
        setError(err.response?.data?.error || err.message || 'Failed to delete district');
      });
  };

  if (loading) return <CircularProgress sx={{ mt: 2 }} />;
  if (error) return <Alert severity="error">{error}</Alert>;

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 2, flexWrap: 'wrap' }}>
        <Typography variant="h5" sx={{ color: 'primary.main', flex: 1 }}>
          Districts
        </Typography>
        <TextField
          size="small"
          placeholder="Search..."
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          InputProps={{ startAdornment: <SearchIcon sx={{ mr: 0.5, color: 'text.secondary' }} /> }}
          sx={{ width: 200 }}
        />
        <Badge badgeContent={activeFilterCount} color="primary" invisible={activeFilterCount === 0}>
          <IconButton
            size="small"
            color={activeFilterCount ? 'primary' : 'default'}
            onClick={() => setFilterModalOpen(true)}
            title="Advanced filters"
          >
            <FilterListIcon fontSize="small" />
          </IconButton>
        </Badge>
        <Button variant="contained" size="small" startIcon={<AddOutlinedIcon />} onClick={openAddDialog}>
          Add District
        </Button>
      </Box>
      <TableContainer component={Paper} sx={{ bgcolor: 'background.paper' }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>ID</TableCell>
              <TableCell>District</TableCell>
              <TableCell>Name</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredList.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} align="center">
                  No districts
                </TableCell>
              </TableRow>
            ) : (
              filteredList.map((row) => (
                <TableRow key={row.id}>
                  <TableCell>{row.id}</TableCell>
                  <TableCell>{row.district}</TableCell>
                  <TableCell>{row.name || '-'}</TableCell>
                  <TableCell align="right">
                    <Stack direction="row" spacing={0.5} justifyContent="flex-end">
                      <IconButton size="small" onClick={() => openEditDialog(row)}>
                        <EditOutlinedIcon fontSize="small" />
                      </IconButton>
                      <IconButton size="small" color="error" onClick={() => handleDelete(row)}>
                        <DeleteOutlineOutlinedIcon fontSize="small" />
                      </IconButton>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={dialogOpen} onClose={() => !saving && setDialogOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>{editingRow ? 'Edit District' : 'Add District'}</DialogTitle>
        <DialogContent dividers>
          <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2, mt: 1 }}>
            <TextField
              label="District"
              value={form.district}
              onChange={(e) => setForm((f) => ({ ...f, district: cleanText(e.target.value) }))}
              required
              size="small"
            />
            <TextField
              label="Name"
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: cleanText(e.target.value) }))}
              size="small"
            />
          </Box>
          <TextField
            label="Detail (JSON)"
            value={form.detail}
            onChange={(e) => setForm((f) => ({ ...f, detail: e.target.value }))}
            fullWidth
            size="small"
            multiline
            rows={8}
            sx={{ mt: 2, '& .MuiInputBase-input': { fontFamily: 'ui-monospace, monospace', fontSize: 13 } }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={handleSave} variant="contained" disabled={saving || !form.district}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
        </DialogActions>
      </Dialog>

      <AdvancedFilterModal
        open={filterModalOpen}
        onClose={() => setFilterModalOpen(false)}
        gridKey={DISTRICTS_GRID_KEY}
        columns={DISTRICTS_FILTER_COLUMNS}
        rules={filterRules}
        setRules={setFilterRules}
      />
    </Box>
  );
}
