import { tranApi } from '../api/tranClient';
import { getQuickSheetsActor } from '../components/Forms/quickSheetsActor';

// Form Templates API
export const getFormTemplates = async (userType = '') => {
  const response = await tranApi.get('/api/forms/templates', {
    params: userType ? { user_type: userType } : {},
  });
  return response.data;
};

export const getFormTemplate = async (id) => {
  const response = await tranApi.get(`/api/forms/templates/${id}`);
  return response.data;
};

export const createFormTemplate = async (template) => {
  const response = await tranApi.post('/api/forms/templates', template);
  return response.data;
};

export const updateFormTemplate = async (id, template) => {
  const response = await tranApi.put(`/api/forms/templates/${id}`, template);
  return response.data;
};

export const deleteFormTemplate = async (id) => {
  const response = await tranApi.delete(`/api/forms/templates/${id}`);
  return response.data;
};

// Form Answers API
export const getFormAnswers = async (formId = '', userId = '') => {
  const params = {};
  if (formId) params.form_id = formId;
  if (userId) params.user_id = userId;
  const response = await tranApi.get('/api/forms/answers', { params });
  return response.data;
};

export const getFormAnswer = async (id) => {
  const response = await tranApi.get(`/api/forms/answers/${id}`);
  return response.data;
};

export const createFormAnswer = async (answer, actor = null) => {
  void actor;
  const response = await tranApi.post('/api/forms/answers', answer);
  return response.data;
};

export const updateFormAnswer = async (id, answer) => {
  const response = await tranApi.put(`/api/forms/answers/${id}`, answer);
  return response.data;
};

export const deleteFormAnswer = async (id) => {
  const response = await tranApi.delete(`/api/forms/answers/${id}`);
  return response.data;
};

export const getAssignableUsers = async (userType, query = '') => {
  const params = new URLSearchParams();
  params.set('user_type', userType);
  if (query.trim()) params.set('q', query.trim());
  const response = await tranApi.get(`/api/forms/assignees?${params.toString()}`);
  return response.data;
};

export const createFormAssignments = async (payload, actor = null) => {
  void actor;
  const response = await tranApi.post('/api/forms/assignments', payload);
  return response.data;
};

export const getFormAssignments = async (filters = {}) => {
  const params = new URLSearchParams();
  Object.entries(filters || {}).forEach(([key, value]) => {
    if (value === undefined || value === null) return;
    const text = String(value).trim();
    if (text) params.set(key, text);
  });
  const url = params.size ? `/api/forms/assignments?${params.toString()}` : '/api/forms/assignments';
  const response = await tranApi.get(url);
  return response.data;
};

export const getFormNotifications = async (actor = null) => {
  const next = actor || getQuickSheetsActor();
  const params = new URLSearchParams();
  const uid = String(next.user_id ?? '').trim();
  if (uid) params.set('user_id', uid);
  params.set('user_type', String(next.user_type || 'staff'));
  const response = await tranApi.get(`/api/forms/notifications?${params.toString()}`);
  return response.data;
};
