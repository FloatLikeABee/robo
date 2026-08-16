import React, { useState } from 'react';
import { Tabs, Tab, Box } from '@mui/material';
import ModuleShell from '../components/ModuleShell';
import Articles from './Articles';
import GraphicDocumentGenerator from './GraphicDocumentGenerator';
import ScholarForge from './ScholarForge';

const PublishHub = () => {
  const [tab, setTab] = useState(0);
  const subtitleBar = (
    <Tabs value={tab} onChange={(_, v) => setTab(v)} variant="scrollable" scrollButtons="auto">
      <Tab label="Articles" sx={{ fontSize: '0.65rem', minHeight: 36, py: 0 }} />
      <Tab label="Graphic documents" sx={{ fontSize: '0.65rem', minHeight: 36, py: 0 }} />
      <Tab label="ScholarForge" sx={{ fontSize: '0.65rem', minHeight: 36, py: 0 }} />
    </Tabs>
  );
  return (
    <ModuleShell
      title="Documents"
      helpText="Articles: chapter generation. Graphic documents: illustrated markdown. ScholarForge: thesis/article composer with write→review→revise paragraph pipeline, dedicated RAG, dual AI models, and PDF export."
      subtitleBar={subtitleBar}
    >
      <Box sx={{ flex: 1, minHeight: 0, height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {tab === 0 && <Articles suppressModuleHeader />}
        {tab === 1 && <GraphicDocumentGenerator suppressModuleHeader />}
        {tab === 2 && <ScholarForge suppressModuleHeader />}
      </Box>
    </ModuleShell>
  );
};

export default PublishHub;
