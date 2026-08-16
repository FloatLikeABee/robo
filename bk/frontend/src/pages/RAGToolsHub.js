import React from 'react';
import { Box } from '@mui/material';
import ModuleShell from '../components/ModuleShell';
import RAGManager from './RAGManager';

const RAGToolsHub = () => {
  return (
    <ModuleShell
      title="RAG"
      helpText="Build knowledge collections by uploading files (TXT/JSON/CSV) or fetching content via API request. Assistants can select multiple collections."
    >
      <Box sx={{ flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        <RAGManager suppressModuleHeader />
      </Box>
    </ModuleShell>
  );
};

export default RAGToolsHub;
