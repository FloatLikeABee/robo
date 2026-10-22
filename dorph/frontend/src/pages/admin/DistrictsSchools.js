import React, { useState, useEffect } from 'react';
import { Box, CircularProgress, Alert } from '@mui/material';
import { tranApi, tranEndpoints } from '../../api/tranClient';
import AdminDataGrid from '../../components/admin/AdminDataGrid';
import { usePlatformUi } from '../../PlatformUiContext';
import { activityLocationDescriptionOnly } from '../../utils/activityLocationJson';

function formatDistrictLabel(d) {
  const code = String(d?.district ?? '').trim();
  const name = String(d?.name ?? '').trim();
  if (code && name) return `${code} — ${name}`;
  return code || name || (d?.id != null ? `#${d.id}` : '');
}

function mapFacilityRow(x, districtById) {
  const districtId = x.district_id ?? null;
  const district = districtId != null ? districtById.get(Number(districtId)) : null;
  return {
    id: x.id,
    facility_code: x.facility_code,
    name: x.name,
    district_id: districtId,
    district_display: district ? formatDistrictLabel(district) : districtId != null ? `#${districtId}` : '',
    facility_type: x.facility_type,
    location: x.location,
    description: x.description,
  };
}

function truncDesc(v, n = 80) {
  if (v == null || v === '') return '';
  const s = String(v);
  return s.length > n ? `${s.slice(0, n)}…` : s;
}

export default function DistrictsSchools() {
  const { labels, dictionaries } = usePlatformUi();
  const facilityTypeOptions = dictionaries?.facility_type || [];

  const labelFacilityType = (code) => {
    if (code == null || code === '') return '';
    const o = facilityTypeOptions.find((x) => x.code === code);
    return o?.label || String(code);
  };

  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const load = () =>
    Promise.all([tranApi.get(tranEndpoints.districts), tranApi.get(tranEndpoints.facilities)]).then(
      ([dRes, fRes]) => {
        const districtById = new Map((dRes.data || []).map((d) => [Number(d.id), d]));
        setRows((fRes.data || []).map((x) => mapFacilityRow(x, districtById)));
      }
    );

  useEffect(() => {
    load()
      .catch((err) => setError(err.response?.data?.error || err.message || 'Failed to load places'))
      .finally(() => setLoading(false));
  }, []);

  if (error) return <Alert severity="error">{error}</Alert>;

  const columns = [
    {
      field: 'facility_code',
      headerName: labels.col_facility_code || 'Code',
      flex: 1,
      minWidth: 120,
    },
    {
      field: 'name',
      headerName: labels.col_facility_name || 'Name',
      flex: 1.2,
      minWidth: 180,
    },
    {
      field: 'district_display',
      headerName: 'District',
      flex: 1,
      minWidth: 160,
    },
    {
      field: 'facility_type',
      headerName: labels.col_facility_type || 'Type',
      width: 140,
      valueGetter: (_v, row) => labelFacilityType(row?.facility_type),
    },
    {
      field: 'location',
      headerName: 'Location',
      flex: 1.1,
      minWidth: 200,
      valueGetter: (_v, row) => activityLocationDescriptionOnly(row?.location),
    },
    {
      field: 'description',
      headerName: 'Description',
      flex: 1,
      minWidth: 160,
      valueGetter: (_v, row) => truncDesc(row?.description),
    },
  ];

  const getRowClassName = (params) => {
    const code = params.row.facility_code || '';
    if (code.includes('-ES')) return 'theme-elementary';
    if (code.includes('-MS')) return 'theme-middle';
    if (code.includes('-HS')) return 'theme-high';
    return '';
  };

  const onCreate = async (draft) => {
    const payload = { ...draft };
    delete payload.district_display;
    const res = await tranApi.post(tranEndpoints.facilities, payload);
    const t = res.data || {};
    await load();
    return t.ID || t.id;
  };

  const onUpdate = async (id, payload) => {
    const next = { ...payload };
    delete next.district_display;
    await tranApi.put(tranEndpoints.facility(id), next);
    await load();
  };

  const onDelete = async (row) => {
    await tranApi.delete(tranEndpoints.facility(row.id));
    setRows((prev) => prev.filter((r) => r.id !== row.id));
  };

  return (
    <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
      {loading && <CircularProgress sx={{ mt: 2 }} />}
      <Box sx={{ flex: 1, minHeight: 0, display: 'flex' }}>
        <AdminDataGrid
          title={labels.nav_districts_facilities || labels.term_facilities || 'Places'}
          rows={rows}
          columns={columns}
          loading={loading}
          getRowClassName={getRowClassName}
          entityTypeForComments="school"
          districtIdAsSelect
          hiddenFields={['district_display']}
          locationJsonFieldKeys={['location']}
          filterGetValue={(row, field) => {
            const f = String(field).toLowerCase();
            if (f === 'location') return activityLocationDescriptionOnly(row?.location) ?? '';
            if (f === 'district_display') return row.district_display ?? '';
            return row[field];
          }}
          fetchDetail={async (id) => {
            const res = await tranApi.get(tranEndpoints.facility(id));
            return res.data;
          }}
          createFields={[
            'facility_code',
            'name',
            'district_id',
            'facility_type',
            'location',
            'description',
            'detail',
          ]}
          requiredFields={['facility_code']}
          onCreate={onCreate}
          onUpdate={onUpdate}
          onDelete={onDelete}
          entityRouteForAttachments="facilities"
        />
      </Box>
    </Box>
  );
}
