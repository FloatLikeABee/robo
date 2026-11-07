import React from 'react';
import { Box, Tab, Tabs } from '@mui/material';
import { useSearchParams } from 'react-router-dom';
import EmbeddedModule from './EmbeddedModule';
import GenericData from './GenericData';

export default function ContentsAndData() {
  const [searchParams, setSearchParams] = useSearchParams();
  const tab = searchParams.get('tab') === 'publish' ? 'publish' : 'data';

  return (
    <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
      <Tabs
        value={tab}
        onChange={(_, next) => setSearchParams(next === 'publish' ? { tab: 'publish' } : {})}
        sx={{ minHeight: 44, flexShrink: 0, borderBottom: 1, borderColor: 'divider' }}
      >
        <Tab value="data" label="Data" sx={{ textTransform: 'none', minHeight: 44 }} />
        <Tab value="publish" label="Publish" sx={{ textTransform: 'none', minHeight: 44 }} />
      </Tabs>
      <Box
        sx={{
          flex: 1,
          minHeight: 0,
          display: tab === 'data' ? 'flex' : 'none',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
      >
        <GenericData onPublished={() => setSearchParams({ tab: 'publish' })} />
      </Box>
      <Box
        sx={{
          flex: 1,
          minHeight: 0,
          display: tab === 'publish' ? 'flex' : 'none',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
      >
        <EmbeddedModule id="content-maker" />
      </Box>
    </Box>
  );
}
