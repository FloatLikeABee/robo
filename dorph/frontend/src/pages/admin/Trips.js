import React, { useState, useEffect } from 'react';
import { Box, CircularProgress, Alert } from '@mui/material';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import AdminDataGrid from '../../components/admin/AdminDataGrid';
import { usePlatformUi } from '../../PlatformUiContext';
import { activityLocationDescriptionOnly } from '../../utils/activityLocationJson';

function mapTripRow(x) {
  return {
    id: x.id,
    name: x.name,
    start_date: x.start_date,
    end_date: x.end_date,
    location: x.location,
    activity_type: x.activity_type,
    description: x.description,
  };
}

function truncDesc(v, n = 80) {
  if (v == null || v === '') return '';
  const s = String(v);
  return s.length > n ? `${s.slice(0, n)}…` : s;
}

export default function Trips() {
  const { labels } = usePlatformUi();
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const load = () =>
    tranApi.get(tranEndpoints.activities).then((res) => {
      const data = res.data || [];
      setRows(data.map(mapTripRow));
    });

  useEffect(() => {
    load()
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load trips'))
      .finally(() => setLoading(false));
  }, []);

  if (error) return <Alert severity="error">{error}</Alert>;

  const columns = [
    { field: 'name', headerName: 'Name', flex: 1.2, minWidth: 220 },
    { field: 'start_date', headerName: 'Start date', width: 170 },
    { field: 'end_date', headerName: 'End date', width: 170 },
    {
      field: 'location',
      headerName: 'Location',
      flex: 1.1,
      minWidth: 220,
      valueGetter: (_v, row) => activityLocationDescriptionOnly(row?.location),
    },
    { field: 'activity_type', headerName: 'Activity type', width: 120 },
    {
      field: 'description',
      headerName: 'Description',
      flex: 1,
      minWidth: 160,
      valueGetter: (_v, row) => truncDesc(row?.description),
    },
  ];

  const getRowClassName = (params) => {
    const name = params.row.name || '';
    if (name.includes('-ES')) return 'theme-elementary';
    if (name.includes('-MS')) return 'theme-middle';
    if (name.includes('-HS')) return 'theme-high';
    return '';
  };

  const onCreate = async (draft) => {
    const res = await tranApi.post(tranEndpoints.activities, draft);
    const t = res.data || {};
    await load();
    return t.ID || t.id;
  };

  const onUpdate = async (id, payload) => {
    await tranApi.put(tranEndpoints.activity(id), payload);
    await load();
  };

  const onDelete = async (row) => {
    await tranApi.delete(tranEndpoints.activity(row.id));
    setRows((prev) => prev.filter((r) => r.id !== row.id));
  };

  return (
    <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
      {loading && <CircularProgress sx={{ mt: 2 }} />}
      <Box sx={{ flex: 1, minHeight: 0, display: 'flex' }}>
        <AdminDataGrid
          title={labels.nav_activities}
          rows={rows}
          columns={columns}
          loading={loading}
          getRowClassName={getRowClassName}
          entityTypeForComments="trip"
          locationJsonFieldKeys={['location']}
          filterGetValue={(row, field) =>
            String(field).toLowerCase() === 'location'
              ? activityLocationDescriptionOnly(row?.location) ?? ''
              : row[field]
          }
          fetchDetail={async (id) => {
            const res = await tranApi.get(`${tranEndpoints.activity(id)}/full`);
            return res.data;
          }}
          createFields={[
            'name',
            'start_date',
            'end_date',
            'location',
            'activity_type',
            'guid',
            'description',
            'detail',
          ]}
          requiredFields={['name']}
          onCreate={onCreate}
          onUpdate={onUpdate}
          onDelete={onDelete}
          entityRouteForAttachments="activities"
        />
      </Box>
    </Box>
  );
}
