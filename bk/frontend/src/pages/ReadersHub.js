import React, { useState } from 'react';
import { Tabs, Tab, Box } from '@mui/material';
import ModuleShell from '../components/ModuleShell';
import ImageReader from './ImageReader';
import PDFReader from './PDFReader';

const ReadersHub = () => {
  const [tab, setTab] = useState(0);
  const subtitleBar = (
    <Tabs value={tab} onChange={(_, v) => setTab(v)} variant="scrollable" scrollButtons="auto">
      <Tab label="Image reader" sx={{ fontSize: '0.65rem', minHeight: 36, py: 0 }} />
      <Tab label="PDF reader" sx={{ fontSize: '0.65rem', minHeight: 36, py: 0 }} />
    </Tabs>
  );
  return (
    <ModuleShell
      title="Readers"
      helpText="Extract text from images or PDFs, optionally summarize with AI, and add results to RAG collections."
      subtitleBar={subtitleBar}
    >
      <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden', pb: 1 }}>
        {tab === 0 && <ImageReader suppressModuleHeader />}
        {tab === 1 && <PDFReader suppressModuleHeader />}
      </Box>
    </ModuleShell>
  );
};

export default ReadersHub;
