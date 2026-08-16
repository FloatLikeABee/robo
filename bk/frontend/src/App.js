import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, CssBaseline, Box } from '@mui/material';
import theme from './theme';
import Header from './components/Header';
import AppFooter from './components/AppFooter';
import AssistantManager from './pages/AssistantManager';
import SystemStatus from './pages/SystemStatus';
import ReadersHub from './pages/ReadersHub';
import PublishHub from './pages/PublishHub';
import ImageGenerator from './pages/ImageGenerator';
import VideoStoryGenerator from './pages/VideoStoryGenerator';
import RAGToolsHub from './pages/RAGToolsHub';

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            height: '100vh',
            overflow: 'hidden',
            position: 'relative',
            zIndex: 1,
          }}
        >
          <Header />
          <Box
            component="main"
            sx={{
              flex: 1,
              minHeight: 0,
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
              position: 'relative',
              zIndex: 1,
              width: '100%',
              mx: 0,
              px: 2,
              pt: 1,
            }}
          >
            <Box
              sx={{
                flex: 1,
                minHeight: 0,
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
              }}
            >
              <Routes>
                <Route path="/" element={<Navigate to="/assistants" replace />} />
                <Route path="/assistants" element={<AssistantManager />} />
                <Route path="/rag" element={<RAGToolsHub />} />
                <Route path="/rag-tools" element={<Navigate to="/rag" replace />} />
                <Route path="/images" element={<ImageGenerator />} />
                <Route path="/video-stories" element={<VideoStoryGenerator />} />
                <Route path="/readers" element={<ReadersHub />} />
                <Route path="/documents" element={<PublishHub />} />
                <Route path="/status" element={<SystemStatus />} />
                <Route path="/tools" element={<Navigate to="/assistants" replace />} />
                <Route path="/dialogue" element={<Navigate to="/assistants" replace />} />
                <Route path="/agents" element={<Navigate to="/assistants" replace />} />
                <Route path="/customizations" element={<Navigate to="/assistants" replace />} />
                <Route path="/advisers" element={<Navigate to="/assistants" replace />} />
                <Route path="/image-reader" element={<Navigate to="/readers" replace />} />
                <Route path="/pdf-reader" element={<Navigate to="/readers" replace />} />
                <Route path="/sources" element={<Navigate to="/rag" replace />} />
                <Route path="/crawler" element={<Navigate to="/rag" replace />} />
                <Route path="/gathering" element={<Navigate to="/rag" replace />} />
                <Route path="/flow" element={<Navigate to="/assistants" replace />} />
                <Route path="/db-tools" element={<Navigate to="/assistants" replace />} />
                <Route path="/scholar-forge" element={<Navigate to="/documents" replace />} />
                <Route path="/articles" element={<Navigate to="/documents" replace />} />
                <Route path="/graphic-document" element={<Navigate to="/documents" replace />} />
                <Route path="/dashboard" element={<Navigate to="/status" replace />} />
              </Routes>
            </Box>
          </Box>
          <AppFooter />
        </Box>
      </Router>
    </ThemeProvider>
  );
}

export default App; 