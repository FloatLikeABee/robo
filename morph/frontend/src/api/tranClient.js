import axios from 'axios';
import { API_BASE_URL as API_BASE } from '../apiBase';
import { getMorphToken, clearMorphSession } from '../auth/morphSession';

export const tranApi = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
  /** Morph AI replies can take a long time; 0 hangs forever — cap at 5 min for recoverable UI. */
  timeout: 300000,
});

tranApi.interceptors.request.use((config) => {
  const t = getMorphToken();
  if (t) {
    const h = config.headers || {};
    h.Authorization = `Bearer ${t}`;
    config.headers = h;
  }
  return config;
});

tranApi.interceptors.response.use(
  (res) => res,
  (err) => {
    const status = err.response?.status;
    const url = String(err.config?.url || '');
    const path = typeof window !== 'undefined' ? window.location.pathname : '';
    const onMorphData =
      path.startsWith('/morphdata') || path.startsWith('/forms') || path.startsWith('/transfinderx');
    // Morph Data is usable without login; do not bounce to Morph AI login on API 401.
    if (status === 401 && !url.includes('/api/auth/login') && !onMorphData) {
      clearMorphSession();
      if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
        window.location.assign('/login');
      }
    }
    return Promise.reject(err);
  }
);

export const tranEndpoints = {
  districts: '/api/tran/districts',
  district: (id) => `/api/tran/districts/${id}`,
  facilities: '/api/tran/facilities',
  facility: (id) => `/api/tran/facilities/${id}`,
  members: '/api/tran/members',
  member: (id) => `/api/tran/members/${id}`,
  employees: '/api/tran/employees',
  employee: (id) => `/api/tran/employees/${id}`,
  assets: '/api/tran/assets',
  asset: (id) => `/api/tran/assets/${id}`,
  activities: '/api/tran/activities',
  activity: (id) => `/api/tran/activities/${id}`,
  caseTasks: '/api/tran/case-tasks',
  caseTask: (id) => `/api/tran/case-tasks/${id}`,
  caseTaskFull: (id) => `/api/tran/case-tasks/${id}/full`,
  caseTaskAttachments: (id) => `/api/tran/case-tasks/${id}/attachments`,
  caseTaskAttachment: (id, attachmentId) => `/api/tran/case-tasks/${id}/attachments/${attachmentId}`,
  caseTaskAttachmentDownload: (id, attachmentId) => `/api/tran/case-tasks/${id}/attachments/${attachmentId}/download`,
  caseTaskPdf: (id) => `/api/tran/case-tasks/${id}/pdf`,
  caseTaskSendEmail: (id) => `/api/tran/case-tasks/${id}/send-email`,
  storyPosts: '/api/tran/story-posts',
  storyPost: (id) => `/api/tran/story-posts/${id}`,
  storyPostFull: (id) => `/api/tran/story-posts/${id}/full`,
  storyPostAIGenerate: '/api/tran/story-posts/ai-generate',
  attachmentConfig: '/api/tran/attachment-config',
  entityAttachments: (entityRoute, id) => `/api/tran/${entityRoute}/${id}/attachments`,
  entityAttachment: (entityRoute, id, attachmentId) =>
    `/api/tran/${entityRoute}/${id}/attachments/${attachmentId}`,
  entityAttachmentDownload: (entityRoute, id, attachmentId) =>
    `/api/tran/${entityRoute}/${id}/attachments/${attachmentId}/download`,
  gridSavedFilters: '/api/tran/grid-saved-filters',
  gridColorConfig: '/api/tran/grid-color-config',
  platformUiConfig: '/api/tran/platform-ui-config',
  users: '/api/tran/users',
  user: (id) => `/api/tran/users/${id}`,
  userMe: '/api/tran/users/me',
  toolNotes: '/api/tran/tool-notes',
  toolNote: (id) => `/api/tran/tool-notes/${id}`,
  toolNoteRead: (id) => `/api/tran/tool-notes/${id}/read`,
  comments: '/api/tran/comments',
  comment: (id) => `/api/tran/comments/${id}`,
  /** Personal notes & TODOs (per UserID; default user until auth) */
  notesTodos: '/api/tran/notes-todos',
  notesTodo: (id) => `/api/tran/notes-todos/${id}`,
  bigNotes: '/api/tran/big-notes',
  bigNote: (id) => `/api/tran/big-notes/${id}`,
  bigNoteRegenerate: (id) => `/api/tran/big-notes/${id}/regenerate`,
  bigNotePublish: (id) => `/api/tran/big-notes/${id}/publish`,
  bigNoteResponses: (id) => `/api/tran/big-notes/${id}/responses`,
  bigNoteResponseAnalyze: (id, responseId) => `/api/tran/big-notes/${id}/responses/${responseId}/analyze`,
  bigNoteAnalyzeAll: (id) => `/api/tran/big-notes/${id}/analyze`,
  timelines: '/api/tran/timelines',
  timeline: (id) => `/api/tran/timelines/${id}`,
  timelinePublish: (id) => `/api/tran/timelines/${id}/publish`,
  /** Short AI text helpers (notes, todos, task chain steps) */
  textAssist: '/api/tran/text-assist',
  composerxEmails: '/api/composerx/emails',
  composerxEmail: (id) => `/api/composerx/emails/${id}`,
  genericData: '/api/tran/generic-data',
  genericDataItem: (id) => `/api/tran/generic-data/${id}`,
  genericDataFull: (id) => `/api/tran/generic-data/${id}/full`,
  genericDataImport: '/api/tran/generic-data/import',
  genericDataAnalyze: (id) => `/api/tran/generic-data/${id}/analyze`,
};
