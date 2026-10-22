import React, { useState, useEffect } from 'react';
import { Box, CircularProgress, Alert } from '@mui/material';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import AdminDataGrid from '../../components/admin/AdminDataGrid';
import { usePlatformUi } from '../../PlatformUiContext';

function buildName(lastName, firstName, middleName) {
  const parts = [firstName, middleName, lastName]
    .map((v) => String(v ?? '').trim())
    .filter(Boolean);
  return parts.join(' ');
}

function withLegacyNameFields(payload) {
  const next = { ...payload };
  if (!Object.prototype.hasOwnProperty.call(next, 'name')) return next;
  const fullName = String(next.name ?? '').trim();
  next.last_name = fullName || null;
  next.first_name = null;
  next.middle_name = null;
  delete next.name;
  return next;
}

export default function Students({ titleOverride } = {}) {
  const { labels } = usePlatformUi();
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    load()
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load students'))
      .finally(() => setLoading(false));
  }, []);

  if (error) return <Alert severity="error">{error}</Alert>;

  const truncDesc = (v, n = 80) => {
    if (v == null || v === '') return '';
    const s = String(v);
    return s.length > n ? `${s.slice(0, n)}…` : s;
  };

  const columns = [
    { field: 'name', headerName: 'Name', flex: 1.2, minWidth: 220 },
    { field: 'facility', headerName: labels.term_facility || 'Place', flex: 1, minWidth: 160 },
    { field: 'email', headerName: 'Email', flex: 1.2, minWidth: 220 },
    { field: 'participant_type', headerName: 'Type', width: 120 },
    {
      field: 'description',
      headerName: 'Description',
      flex: 1,
      minWidth: 160,
      valueGetter: (_v, row) => truncDesc(row?.description),
    },
  ];

  const load = () =>
    tranApi.get(tranEndpoints.members).then((res) => {
      const data = res.data || [];
      setRows(
        data.map((x) => ({
          id: x.id,
          name: buildName(x.last_name, x.first_name, x.middle_name),
          dob: x.dob,
          facility: x.facility,
          gender: x.gender,
          email: x.email,
          participant_type: x.participant_type,
          description: x.description,
        }))
      );
    });

  const onCreate = async (draft) => {
    const res = await tranApi.post(tranEndpoints.members, withLegacyNameFields(draft));
    const s = res.data || {};
    await load();
    return s.ID ?? s.id;
  };

  const onUpdate = async (id, payload) => {
    await tranApi.put(tranEndpoints.member(id), withLegacyNameFields(payload));
    await load();
  };

  const onDelete = async (row) => {
    await tranApi.delete(tranEndpoints.member(row.id));
    setRows((prev) => prev.filter((r) => r.id !== row.id));
  };

  return (
    <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
      {loading && <CircularProgress sx={{ mt: 2 }} />}
      <Box sx={{ flex: 1, minHeight: 0, display: 'flex' }}>
        <AdminDataGrid
          title={titleOverride || labels.nav_people || labels.nav_participants || 'People'}
          rows={rows}
          columns={columns}
          loading={loading}
          facilityAsSelect
          entityTypeForComments="student"
          fetchDetail={async (id) => {
            const res = await tranApi.get(`${tranEndpoints.member(id)}/full`);
            const row = res.data || {};
            return {
              ...row,
              name: buildName(row.last_name, row.first_name, row.middle_name),
            };
          }}
          createFields={[
            'name',
            'email',
            'dob',
            'entry_date',
            'facility',
            'gender',
            'participant_type',
            'description',
            'detail',
          ]}
          hiddenFields={['first_name', 'middle_name', 'last_name']}
          onCreate={onCreate}
          onUpdate={onUpdate}
          onDelete={onDelete}
          entityRouteForAttachments="members"
        />
      </Box>
    </Box>
  );
}
