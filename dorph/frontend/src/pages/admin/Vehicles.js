import React, { useState, useEffect } from 'react';
import { Box, CircularProgress, Alert } from '@mui/material';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import AdminDataGrid from '../../components/admin/AdminDataGrid';
import { usePlatformUi } from '../../PlatformUiContext';

function mapVehicleRow(v) {
  return {
    id: v.id,
    asset_tag: v.asset_tag,
    description: v.description,
    asset_id: v.asset_id,
    asset_type: v.asset_type,
  };
}

function truncDesc(v, n = 80) {
  if (v == null || v === '') return '';
  const s = String(v);
  return s.length > n ? `${s.slice(0, n)}…` : s;
}

export default function Vehicles() {
  const { labels } = usePlatformUi();
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const load = () =>
    tranApi.get(tranEndpoints.assets).then((res) => {
      const data = res.data || [];
      setRows(data.map(mapVehicleRow));
    });

  useEffect(() => {
    load()
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load vehicles'))
      .finally(() => setLoading(false));
  }, []);

  if (error) return <Alert severity="error">{error}</Alert>;

  const columns = [
    { field: 'id', headerName: 'ID', width: 80 },
    { field: 'asset_tag', headerName: 'Asset tag', flex: 1, minWidth: 160 },
    {
      field: 'description',
      headerName: 'Description',
      flex: 1.2,
      minWidth: 220,
      valueGetter: (_v, row) => truncDesc(row?.description),
    },
    { field: 'asset_id', headerName: 'Asset ID', width: 120 },
    { field: 'asset_type', headerName: 'Asset type', width: 120 },
  ];

  const onCreate = async (draft) => {
    const res = await tranApi.post(tranEndpoints.assets, draft);
    const v = res.data || {};
    await load();
    return v.ID || v.id;
  };

  const onUpdate = async (id, payload) => {
    await tranApi.put(tranEndpoints.asset(id), payload);
    await load();
  };

  const onDelete = async (row) => {
    await tranApi.delete(tranEndpoints.asset(row.id));
    setRows((prev) => prev.filter((r) => r.id !== row.id));
  };

  return (
    <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
      {loading && <CircularProgress sx={{ mt: 2 }} />}
      <Box sx={{ flex: 1, minHeight: 0, display: 'flex' }}>
        <AdminDataGrid
          title={labels.nav_assets}
          rows={rows}
          columns={columns}
          loading={loading}
          entityTypeForComments="vehicle"
          fetchDetail={async (id) => {
            const res = await tranApi.get(`${tranEndpoints.asset(id)}/full`);
            return res.data;
          }}
          createFields={['asset_tag', 'asset_id', 'asset_type', 'contractor_id', 'description', 'detail']}
          onCreate={onCreate}
          onUpdate={onUpdate}
          onDelete={onDelete}
          entityRouteForAttachments="assets"
        />
      </Box>
    </Box>
  );
}
