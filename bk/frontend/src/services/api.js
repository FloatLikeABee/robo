import axios from 'axios';

// Same-origin in dev (CRA proxy → bk-api). Set REACT_APP_API_URL when UI and API are on different hosts.
export const API_BASE_URL = process.env.REACT_APP_API_URL || '';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// System Status
export const getStatus = async () => {
  const response = await api.get('/status');
  return response.data;
};

// RAG Collections
export const getRAGCollections = async () => {
  const response = await api.get('/rag/collections');
  return response.data;
};

export const addRAGData = async (payload) => {
  const { collection_name, data_input } = payload;
  const response = await api.post(`/rag/collections/${collection_name}/data`, data_input);
  return response.data;
};

/** Suggest a short topic title from content using AI (for RAG document name). */
export const suggestRAGTitle = async (content) => {
  const response = await api.post('/rag/suggest-title', { content });
  return response.data;
};

export const validateRAGData = async (data) => {
  const response = await api.post('/rag/validate', data);
  return response.data;
};

export const queryRAGCollection = async (collectionName, query, nResults = 5) => {
  const response = await api.post(`/rag/collections/${collectionName}/query`, {
    query,
    n_results: nResults,
  });
  return response.data;
};

export const deleteRAGCollection = async (collectionName) => {
  const response = await api.delete(`/rag/collections/${collectionName}`);
  return response.data;
};

// Agents
export const getAgents = async () => {
  const response = await api.get('/agents');
  return response.data;
};

export const getAgent = async (agentId) => {
  const response = await api.get(`/agents/${agentId}`);
  return response.data;
};

export const createAgent = async (config) => {
  const response = await api.post('/agents', config);
  return response.data;
};

export const updateAgent = async (agentId, config) => {
  const response = await api.put(`/agents/${agentId}`, config);
  return response.data;
};

export const deleteAgent = async (agentId) => {
  const response = await api.delete(`/agents/${agentId}`);
  return response.data;
};

export const runAgent = async (agentId, query, context = null) => {
  const response = await api.post(`/agents/${agentId}/run`, {
    query,
    context,
  });
  return response.data;
};

// Assistants (unified advisers + customizations)
export const getAssistants = async () => {
  const response = await api.get('/assistants');
  return response.data;
};

export const getAssistant = async (assistantId) => {
  const response = await api.get(`/assistants/${assistantId}`);
  return response.data;
};

export const createAssistant = async (payload) => {
  const response = await api.post('/assistants', payload);
  return response.data;
};

export const updateAssistant = async (assistantId, payload) => {
  const response = await api.put(`/assistants/${assistantId}`, payload);
  return response.data;
};

export const deleteAssistant = async (assistantId) => {
  const response = await api.delete(`/assistants/${assistantId}`);
  return response.data;
};

export const runAssistant = async (assistantId, query, context = null) => {
  const response = await api.post(`/assistants/${assistantId}/run`, {
    query,
    context,
  });
  return response.data;
};

export const migrateAssistants = async () => {
  const response = await api.post('/assistants/migrate');
  return response.data;
};

// Models
export const getModels = async () => {
  const response = await api.get('/models');
  return response.data;
};

export const getProviders = async () => {
  const response = await api.get('/providers');
  return response.data;
};

// System settings
export const getSystemSettings = async () => {
  const response = await api.get('/system/settings');
  return response.data;
};

export const updateSystemSettings = async (payload) => {
  const response = await api.put('/system/settings', payload);
  return response.data;
};

// The Help assistant
export const askHelp = async (payload) => {
  const response = await api.post('/help/ask', payload);
  return response.data;
};

// MCP
export const startMCPServer = async () => {
  const response = await api.post('/mcp/start');
  return response.data;
};

// MCP Hosts
export const getMCPHosts = async () => {
  const response = await api.get('/mcp/hosts');
  return response.data;
};

export const createMCPHost = async (payload) => {
  const response = await api.post('/mcp/hosts', payload);
  return response.data;
};

export const updateMCPHost = async (hostId, payload) => {
  const response = await api.put(`/mcp/hosts/${hostId}`, payload);
  return response.data;
};

export const deleteMCPHost = async (hostId) => {
  const response = await api.delete(`/mcp/hosts/${hostId}`);
  return response.data;
};

// Legacy aliases (to be removed once all callers migrate)
export const getCustomizations = getAssistants;
export const createCustomization = createAssistant;
export const updateCustomization = updateAssistant;
export const deleteCustomization = deleteAssistant;
export const queryCustomization = runAssistant;

// Dialogues
export const getDialogues = async () => {
  const response = await api.get('/dialogues');
  return response.data;
};

export const getDialogue = async (dialogueId) => {
  const response = await api.get(`/dialogues/${dialogueId}`);
  return response.data;
};

export const createDialogue = async (payload) => {
  const response = await api.post('/dialogues', payload);
  return response.data;
};

export const updateDialogue = async (dialogueId, payload) => {
  const response = await api.put(`/dialogues/${dialogueId}`, payload);
  return response.data;
};

export const deleteDialogue = async (dialogueId) => {
  const response = await api.delete(`/dialogues/${dialogueId}`);
  return response.data;
};

export const startDialogue = async (dialogueId, payload) => {
  const response = await api.post(`/dialogues/${dialogueId}/start`, payload);
  return response.data;
};

export const continueDialogue = async (dialogueId, payload) => {
  const response = await api.post(`/dialogues/${dialogueId}/continue`, payload);
  return response.data;
};

// Articles
export const getArticles = async () => {
  const response = await api.get('/articles');
  return response.data;
};

export const getArticle = async (articleId) => {
  const response = await api.get(`/articles/${articleId}`);
  return response.data;
};

export const createArticle = async (payload) => {
  const response = await api.post('/articles', payload);
  return response.data;
};

export const updateArticle = async (articleId, payload) => {
  const response = await api.put(`/articles/${articleId}`, payload);
  return response.data;
};

export const deleteArticle = async (articleId) => {
  const response = await api.delete(`/articles/${articleId}`);
  return response.data;
};

/** NDJSON stream: calls onEvent(obj) for each parsed line. */
export const streamArticleGenerate = async (articleId, payload, onEvent) => {
  const res = await fetch(`${API_BASE_URL}/articles/${articleId}/generate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    let detail = res.statusText;
    try {
      const j = await res.json();
      detail = j.detail || detail;
    } catch (_) {
      try {
        detail = await res.text();
      } catch (__) {
        /* ignore */
      }
    }
    throw new Error(typeof detail === 'string' ? detail : JSON.stringify(detail));
  }
  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  // eslint-disable-next-line no-constant-condition
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    let idx;
    // eslint-disable-next-line no-cond-assign
    while ((idx = buffer.indexOf('\n')) >= 0) {
      const line = buffer.slice(0, idx).trim();
      buffer = buffer.slice(idx + 1);
      if (!line) continue;
      try {
        onEvent(JSON.parse(line));
      } catch (e) {
        console.warn('Articles stream: skip bad line', line);
      }
    }
  }
};

export const getArticleDownloadUrl = (fileId) =>
  `${API_BASE_URL}/articles/downloads/${encodeURIComponent(fileId)}`;

// ScholarForge — academic article & thesis composer
export const getScholarForgeProjects = async () => {
  const response = await api.get('/scholar-forge/projects');
  return response.data;
};

export const getScholarForgeProject = async (projectId) => {
  const response = await api.get(`/scholar-forge/projects/${projectId}`);
  return response.data;
};

export const createScholarForgeProject = async (payload) => {
  const response = await api.post('/scholar-forge/projects', payload);
  return response.data;
};

export const updateScholarForgeProject = async (projectId, payload) => {
  const response = await api.put(`/scholar-forge/projects/${projectId}`, payload);
  return response.data;
};

export const deleteScholarForgeProject = async (projectId) => {
  const response = await api.delete(`/scholar-forge/projects/${projectId}`);
  return response.data;
};

export const uploadScholarForgeImages = async (projectId, formData) => {
  const response = await api.post(`/scholar-forge/projects/${projectId}/upload-images`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return response.data;
};

export const submitScholarForgeClarification = async (projectId, payload) => {
  const response = await api.post(`/scholar-forge/projects/${projectId}/clarify/submit`, payload);
  return response.data;
};

export const updateScholarForgeStructure = async (projectId, payload) => {
  const response = await api.put(`/scholar-forge/projects/${projectId}/structure`, payload);
  return response.data;
};

/** NDJSON stream for ScholarForge pipeline actions: prepare, clarify, structure, confirm-structure, generate */
export const streamScholarForge = async (projectId, action, onEvent, body = {}) => {
  const pathMap = {
    prepare: 'prepare',
    clarify: 'clarify',
    structure: 'structure',
    'confirm-structure': 'confirm-structure',
    generate: 'generate',
  };
  const segment = pathMap[action];
  if (!segment) throw new Error(`Unknown ScholarForge action: ${action}`);

  const res = await fetch(`${API_BASE_URL}/scholar-forge/projects/${projectId}/${segment}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    let detail = res.statusText;
    try {
      const j = await res.json();
      detail = j.detail || detail;
    } catch (_) {
      try {
        detail = await res.text();
      } catch (__) {
        /* ignore */
      }
    }
    throw new Error(typeof detail === 'string' ? detail : JSON.stringify(detail));
  }
  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  // eslint-disable-next-line no-constant-condition
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    let idx;
    // eslint-disable-next-line no-cond-assign
    while ((idx = buffer.indexOf('\n')) >= 0) {
      const line = buffer.slice(0, idx).trim();
      buffer = buffer.slice(idx + 1);
      if (!line) continue;
      try {
        onEvent(JSON.parse(line));
      } catch (e) {
        console.warn('ScholarForge stream: skip bad line', line);
      }
    }
  }
};

export const exportScholarForgePdf = async (projectId) => {
  const response = await api.post(`/scholar-forge/projects/${projectId}/export-pdf`);
  return response.data;
};

export const getScholarForgeDownloadUrl = (fileId, format = 'pdf') =>
  `${API_BASE_URL}/scholar-forge/downloads/${encodeURIComponent(fileId)}?format=${format}`;

// Crawler
export const crawlWebsite = async (payload) => {
  const response = await api.post('/crawler/crawl', payload);
  return response.data;
};

// Crawler Profiles
export const getCrawlerProfiles = async () => {
  const response = await api.get('/crawler/profiles');
  return response.data;
};

export const getCrawlerProfile = async (profileId) => {
  const response = await api.get(`/crawler/profiles/${profileId}`);
  return response.data;
};

export const createCrawlerProfile = async (payload) => {
  const response = await api.post('/crawler/profiles', payload);
  return response.data;
};

export const updateCrawlerProfile = async (profileId, payload) => {
  const response = await api.put(`/crawler/profiles/${profileId}`, payload);
  return response.data;
};

export const deleteCrawlerProfile = async (profileId) => {
  const response = await api.delete(`/crawler/profiles/${profileId}`);
  return response.data;
};

export const executeCrawlerProfile = async (profileId) => {
  const response = await api.post(`/crawler/profiles/${profileId}/execute`);
  return response.data;
};

// Gathering
export const gatherData = async (payload) => {
  const response = await api.post('/gathering/gather', payload);
  return response.data;
};


// Request Tools
export const getRequestTools = async () => {
  const response = await api.get('/request-tools');
  return response.data;
};

export const getRequestTool = async (requestId) => {
  const response = await api.get(`/request-tools/${requestId}`);
  return response.data;
};

export const createRequestTool = async (payload) => {
  const response = await api.post('/request-tools', payload);
  return response.data;
};

export const updateRequestTool = async (requestId, payload) => {
  const response = await api.put(`/request-tools/${requestId}`, payload);
  return response.data;
};

export const deleteRequestTool = async (requestId) => {
  const response = await api.delete(`/request-tools/${requestId}`);
  return response.data;
};

export const executeRequestTool = async (requestId) => {
  const response = await api.post(`/request-tools/${requestId}/execute`);
  return response.data;
};

// One-shot: create temp request, execute, capture response, clean up, return result
export const executeRequestOneShot = async (payload) => {
  // Create temp profile
  const created = await api.post('/request-tools', {
    name: `temp_${Date.now()}`,
    request_type: 'http',
    method: payload.method || 'GET',
    url: payload.url,
    headers: payload.headers || {},
    params: payload.params || {},
    body: payload.body || null,
    timeout: 30,
  });
  const tempId = created.id || created.profile_id;
  if (!tempId) throw new Error('Failed to create temp request');
  try {
    // Execute
    const result = await api.post(`/request-tools/${tempId}/execute`);
    return result.data;
  } finally {
    // Clean up temp profile
    try { await api.delete(`/request-tools/${tempId}`); } catch { /* ignore */ }
  }
};

// Conversations
export const getConversations = async () => {
  const response = await api.get('/conversations');
  return response.data;
};

export const getConversation = async (configId) => {
  const response = await api.get(`/conversations/${configId}`);
  return response.data;
};

export const createConversation = async (payload) => {
  const response = await api.post('/conversations', payload);
  return response.data;
};

export const updateConversation = async (configId, payload) => {
  const response = await api.put(`/conversations/${configId}`, payload);
  return response.data;
};

export const deleteConversation = async (configId) => {
  const response = await api.delete(`/conversations/${configId}`);
  return response.data;
};

export const startConversation = async (payload) => {
  const response = await api.post('/conversations/start', payload);
  return response.data;
};

export const continueConversation = async (payload) => {
  const response = await api.post('/conversations/continue', payload);
  return response.data;
};

export const getConversationHistory = async (sessionId) => {
  const response = await api.get(`/conversations/history/${sessionId}`);
  return response.data;
};

export const listSavedConversations = async () => {
  const response = await api.get('/conversations/saved');
  return response.data;
};

export const getSavedConversationContent = async (filename) => {
  const response = await api.get(`/conversations/saved/${encodeURIComponent(filename)}`);
  return response.data;
};

// Image Generation
export const generateImage = async (payload) => {
  const response = await api.post('/images/generate', payload);
  return response.data;
};

export const polishImagePrompt = async (payload) => {
  const response = await api.post('/images/polish-prompt', payload);
  return response.data;
};

export const getGeneratedImages = async () => {
  const response = await api.get('/images');
  return response.data;
};

export const deleteGeneratedImage = async (filename) => {
  const response = await api.delete(`/images/${encodeURIComponent(filename)}`);
  return response.data;
};

// Graphic Document Generator (LLM + several image downloads; allow long server time)
export const generateGraphicDocument = async (payload) => {
  const response = await api.post('/graphic-document/generate', payload, {
    timeout: 900000, // 15 minutes
  });
  return response.data;
};

// Browser Automation
export const executeBrowserAutomation = async (payload) => {
  const response = await api.post('/browser-automation/execute', payload);
  return response.data;
};

// Image Reader
export const readImage = async (file, prompt = null, minPixels = null, maxPixels = null) => {
  const formData = new FormData();
  formData.append('file', file);
  if (prompt) formData.append('prompt', prompt);
  if (minPixels) formData.append('min_pixels', minPixels);
  if (maxPixels) formData.append('max_pixels', maxPixels);
  
  const response = await api.post('/image-reader/read', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};

export const readMultipleImages = async (files, prompt = null, minPixels = null, maxPixels = null) => {
  const formData = new FormData();
  files.forEach((file) => {
    formData.append('files', file);
  });
  if (prompt) formData.append('prompt', prompt);
  if (minPixels) formData.append('min_pixels', minPixels);
  if (maxPixels) formData.append('max_pixels', maxPixels);
  
  const response = await api.post('/image-reader/read-multiple', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};

/** Read image (OCR) then process with chosen AI model using system prompt. */
export const readImageAndProcess = async (file, systemPrompt, provider, model, ocrPrompt = null) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('system_prompt', systemPrompt);
  formData.append('provider', provider);
  formData.append('model', model);
  if (ocrPrompt) formData.append('ocr_prompt', ocrPrompt);

  const response = await api.post('/image-reader/read-and-process', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};

/** Read multiple images (OCR each), combine text, then process with AI once. */
export const readImageAndProcessMultiple = async (files, systemPrompt, provider, model, ocrPrompt = null) => {
  const formData = new FormData();
  files.forEach((file) => formData.append('files', file));
  formData.append('system_prompt', systemPrompt);
  formData.append('provider', provider);
  formData.append('model', model);
  if (ocrPrompt) formData.append('ocr_prompt', ocrPrompt);

  const response = await api.post('/image-reader/read-and-process-multiple', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};

// PDF Reader
export const readPDF = async (formData) => {
  const response = await api.post('/pdf-reader/read', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};

// Video Story Generator
export const getVideoStories = async () => {
  const response = await api.get('/video-stories');
  return response.data;
};

export const getVideoStory = async (projectId) => {
  const response = await api.get(`/video-stories/${projectId}`);
  return response.data;
};

export const createVideoStory = async (payload) => {
  const response = await api.post('/video-stories', payload);
  return response.data;
};

export const updateVideoStory = async (projectId, payload) => {
  const response = await api.put(`/video-stories/${projectId}`, payload);
  return response.data;
};

export const deleteVideoStory = async (projectId) => {
  const response = await api.delete(`/video-stories/${projectId}`);
  return response.data;
};

export const polishVideoStoryScenes = async (projectId, payload) => {
  const response = await api.post(`/video-stories/${projectId}/polish`, payload, { timeout: 600000 });
  return response.data;
};

export const polishVideoStoryContent = async (projectId, payload) => {
  const response = await api.post(`/video-stories/${projectId}/polish-content`, payload, { timeout: 120000 });
  return response.data;
};

export const generateVideoStoryImages = async (projectId, payload) => {
  const response = await api.post(`/video-stories/${projectId}/generate-images`, payload, { timeout: 900000 });
  return response.data;
};

export const generateVideoStoryVideos = async (projectId, payload) => {
  const response = await api.post(`/video-stories/${projectId}/generate-videos`, payload, { timeout: 900000 });
  return response.data;
};

export const getVideoStoryFileUrl = (projectId, kind, filename, download = false) => {
  const q = download ? '?download=true' : '';
  return `${API_BASE_URL}/video-stories/${projectId}/files/${kind}/${encodeURIComponent(filename)}${q}`;
};

// Export all functions
const apiService = {
  getStatus,
  getRAGCollections,
  addRAGData,
  suggestRAGTitle,
  validateRAGData,
  queryRAGCollection,
  deleteRAGCollection,
  getAgents,
  getAgent,
  createAgent,
  updateAgent,
  deleteAgent,
  runAgent,
  // Assistants
  getAssistants,
  getAssistant,
  createAssistant,
  updateAssistant,
  deleteAssistant,
  runAssistant,
  migrateAssistants,
  getModels,
  getProviders,
  getSystemSettings,
  updateSystemSettings,
  askHelp,
  startMCPServer,
  getMCPHosts,
  createMCPHost,
  updateMCPHost,
  deleteMCPHost,
  getCustomizations,
  createCustomization,
  updateCustomization,
  deleteCustomization,
  queryCustomization,
  getDialogues,
  getDialogue,
  createDialogue,
  updateDialogue,
  deleteDialogue,
  startDialogue,
  continueDialogue,
  getArticles,
  getArticle,
  createArticle,
  updateArticle,
  deleteArticle,
  streamArticleGenerate,
  getArticleDownloadUrl,
  getScholarForgeProjects,
  getScholarForgeProject,
  createScholarForgeProject,
  updateScholarForgeProject,
  deleteScholarForgeProject,
  uploadScholarForgeImages,
  submitScholarForgeClarification,
  updateScholarForgeStructure,
  streamScholarForge,
  exportScholarForgePdf,
  getScholarForgeDownloadUrl,
  getVideoStories,
  getVideoStory,
  createVideoStory,
  updateVideoStory,
  deleteVideoStory,
  polishVideoStoryScenes,
  polishVideoStoryContent,
  generateVideoStoryImages,
  generateVideoStoryVideos,
  getVideoStoryFileUrl,
  crawlWebsite,
  getCrawlerProfiles,
  getCrawlerProfile,
  createCrawlerProfile,
  updateCrawlerProfile,
  deleteCrawlerProfile,
  executeCrawlerProfile,
  executeRequestOneShot,
  gatherData,
  getRequestTools,
  getRequestTool,
  createRequestTool,
  updateRequestTool,
  deleteRequestTool,
  executeRequestTool,
  // Conversations
  getConversations,
  getConversation,
  createConversation,
  updateConversation,
  deleteConversation,
  startConversation,
  continueConversation,
  getConversationHistory,
  listSavedConversations,
  getSavedConversationContent,
  // Image Generation
  generateImage,
  polishImagePrompt,
  getGeneratedImages,
  deleteGeneratedImage,
  // Graphic Document Generator
  generateGraphicDocument,
  // Browser Automation
  executeBrowserAutomation,
  // Image Reader
  readImage,
  readMultipleImages,
  readImageAndProcess,
  readImageAndProcessMultiple,
  // PDF Reader
  readPDF,
};

export default apiService; 