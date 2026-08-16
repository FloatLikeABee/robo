const API_BASE = import.meta.env.VITE_API_URL ?? '';
const AUTH_TOKEN_KEY = 'tranform_auth_token';
const AUTH_REMEMBER_KEY = 'tranform_auth_remember';
const SHARED_SESSION_COOKIE_KEY = 'userspanel_session_token';
/** Match UsersPanel default JWT expiry (see JWT_EXPIRY_HOURS). */
const SESSION_COOKIE_MAX_AGE_SECONDS = 48 * 3600;

/** Dispatched after session cleared due to expiry or HTTP 401 (navigate UI to `/login`). */
export const AUTH_EXPIRED_EVENT = 'tranform-auth-expired';

const AUTH_NOTICE_KEY = 'tranform_login_notice';

function readCookie(name: string): string {
  if (typeof document === 'undefined') return '';
  const prefix = `${name}=`;
  const row = document.cookie.split(';').map((v) => v.trim()).find((v) => v.startsWith(prefix));
  if (!row) return '';
  return decodeURIComponent(row.slice(prefix.length));
}

function writeCookie(name: string, value: string, maxAgeSeconds: number): void {
  if (typeof document === 'undefined') return;
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`;
}

/** Session-only cookie for “don’t remember me” (expires when browser closes). */
function writeSessionCookie(name: string, value: string): void {
  if (typeof document === 'undefined') return;
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; SameSite=Lax`;
}

export function readLoginNotice(): string {
  try {
    const msg = sessionStorage.getItem(AUTH_NOTICE_KEY);
    if (msg) sessionStorage.removeItem(AUTH_NOTICE_KEY);
    return msg ?? '';
  } catch {
    return '';
  }
}

function rememberMeEnabled(): boolean {
  if (typeof localStorage === 'undefined') return true;
  return localStorage.getItem(AUTH_REMEMBER_KEY) !== '0';
}

export function isJwtExpired(token: string, skewMs = 60_000): boolean {
  const parts = token.split('.');
  if (parts.length !== 3) return false;
  try {
    let b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const pad = b64.length % 4;
    if (pad) b64 += '='.repeat(4 - pad);
    const payload = JSON.parse(atob(b64)) as { exp?: number };
    if (typeof payload.exp !== 'number') return false;
    return Date.now() >= payload.exp * 1000 - skewMs;
  } catch {
    return false;
  }
}

function clearTokenStores(): void {
  if (typeof localStorage !== 'undefined') localStorage.removeItem(AUTH_TOKEN_KEY);
  if (typeof sessionStorage !== 'undefined') sessionStorage.removeItem(AUTH_TOKEN_KEY);
  writeCookie(SHARED_SESSION_COOKIE_KEY, '', 0);
}

function syncSharedCookie(token: string, remember: boolean): void {
  if (!token) return;
  if (remember) {
    writeCookie(SHARED_SESSION_COOKIE_KEY, token, SESSION_COOKIE_MAX_AGE_SECONDS);
  } else {
    writeSessionCookie(SHARED_SESSION_COOKIE_KEY, token);
  }
}

function acceptTokenOrClear(raw: string): string {
  if (!raw) return '';
  if (isJwtExpired(raw)) {
    clearTokenStores();
    return '';
  }
  return raw;
}

export function getAuthToken(): string {
  let shared = acceptTokenOrClear(readCookie(SHARED_SESSION_COOKIE_KEY));
  if (shared) return shared;

  const remember = rememberMeEnabled();
  const fromStore = acceptTokenOrClear(
    remember
      ? (typeof localStorage !== 'undefined' ? localStorage.getItem(AUTH_TOKEN_KEY) ?? '' : '')
      : (typeof sessionStorage !== 'undefined' ? sessionStorage.getItem(AUTH_TOKEN_KEY) ?? '' : '')
  );
  if (!fromStore) return '';
  syncSharedCookie(fromStore, remember);
  return fromStore;
}

export function setAuthToken(token: string, options?: { rememberMe?: boolean }): void {
  if (!token) {
    clearTokenStores();
    return;
  }
  const remember = options?.rememberMe ?? true;
  if (typeof localStorage !== 'undefined') localStorage.setItem(AUTH_REMEMBER_KEY, remember ? '1' : '0');
  if (remember) {
    if (typeof localStorage !== 'undefined') localStorage.setItem(AUTH_TOKEN_KEY, token);
    if (typeof sessionStorage !== 'undefined') sessionStorage.removeItem(AUTH_TOKEN_KEY);
  } else {
    if (typeof sessionStorage !== 'undefined') sessionStorage.setItem(AUTH_TOKEN_KEY, token);
    if (typeof localStorage !== 'undefined') localStorage.removeItem(AUTH_TOKEN_KEY);
  }
  syncSharedCookie(token, remember);
}

function notifySessionExpired(): void {
  try {
    sessionStorage.setItem(AUTH_NOTICE_KEY, 'Session expired. Please sign in again.');
  } catch {
    /* ignore */
  }
  if (typeof window !== 'undefined') window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT));
}

function isLoginPost(path: string, method: string): boolean {
  return method.toUpperCase() === 'POST' && path === '/api/v1/auth/login';
}

function isPublicApiPath(path: string): boolean {
  return path.startsWith('/api/v1/public/');
}

function invalidateServerSession(path: string, method: string): void {
  if (isLoginPost(path, method) || isPublicApiPath(path)) return;
  clearTokenStores();
  notifySessionExpired();
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const method = (options.method ?? 'GET').toString().toUpperCase();
  const sendAuth = !isLoginPost(path, method) && !isPublicApiPath(path);

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(sendAuth && getAuthToken() ? { Authorization: `Bearer ${getAuthToken()}` } : {}),
      ...(options.headers as Record<string, string>),
    },
  });

  if (res.status === 401 && !isLoginPost(path, method) && !isPublicApiPath(path)) {
    invalidateServerSession(path, method);
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error((err as { error?: string }).error || 'Session expired. Please sign in again.');
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error((err as { error?: string }).error || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

/** Limits match the API (reasonable defaults: HD images + short clips). */
export const MAX_QUESTION_PROMPT_IMAGE_BYTES = 10 * 1024 * 1024;
export const MAX_QUESTION_PROMPT_VIDEO_BYTES = 100 * 1024 * 1024;

export function uploadsUrl(relativePath: string): string {
  const trimmed = relativePath.replace(/^\/+/, '').split('/').filter(Boolean).map(encodeURIComponent).join('/');
  if (API_BASE) {
    return `${API_BASE.replace(/\/$/, '')}/uploads/${trimmed}`;
  }
  return `/uploads/${trimmed}`;
}

export interface Form {
  id: number;
  name: string;
  description: string;
  slug: string;
  single_response_only: boolean;
  /** When true, respondent must tap Start before answering; elapsed time recorded on submit. */
  exam_mode?: boolean;
  landing_html?: string;
  created_at: string;
  updated_at: string;
}

export interface FormPage {
  id: number;
  form_id: number;
  name: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface Question {
  id: number;
  form_id: number;
  page_id: number;
  title: string;
  type: string;
  required: boolean;
  sort_order: number;
  config: QuestionConfig;
  created_at: string;
  updated_at: string;
}

export interface QuestionPromptMedia {
  kind: 'image' | 'video';
  relative_path: string;
}

export interface QuestionConfig {
  multiline?: boolean;
  options?: { value: number; label: string }[];
  max_selections?: number;
  min?: number;
  max?: number;
  max_size?: number;
  allowed_mime?: string;
  allowed_extensions?: string[];
  validation_pattern?: string;
  /** Preset for the public form; user may keep or change. */
  default_value?: unknown;
  /** Optional image or video shown with the question (never both). */
  question_prompt_media?: QuestionPromptMedia | null;
}

export interface FormListResponse {
  forms: Form[];
  total: number;
  page: number;
  limit: number;
}

export interface FormResponse {
  id: string;
  form_id: number;
  slug: string;
  respondent_id: string;
  submitted_at: string;
  /** Wall-clock milliseconds from Exam Mode start (client-reported start) until submit; absent if not timed. */
  exam_duration_ms?: number;
  answers: { question_id: number; type: string; value: unknown; filename?: string; size?: number }[];
}

export interface ResponseListResult {
  responses: FormResponse[];
  total: number;
  page: number;
  limit: number;
}

export interface QuestionRule {
  id: number;
  question_id: number;
  depends_on_question_id: number;
  condition: 'answered' | 'not_answered';
  created_at: string;
  updated_at: string;
}

export interface EventInfo {
  id: string;
  title: string;
  detail: string;
  reporter: string;
  time: string;
  created_at: string;
}

export interface EventInfoCollectionInfo {
  submit_page_url: string;
  public_api_url: string;
  sample_curl: string;
  sample_json: { title: string; detail: string; reporter: string; time: string };
}

export interface SurveyBotTemplate {
  id: string;
  slug: string;
  title: string;
  tags?: string[];
  markdown: string;
  summary?: string;
  published?: boolean;
  published_at?: string;
  public_url?: string;
  needs_compile?: boolean;
  created_by?: string;
  created_at?: string;
  updated_at?: string;
}

export interface AISheetUIBlock {
  type?: string;
  widget?: string;
  id: string;
  label?: string;
  options?: Array<{ value: string; label: string }>;
  submit_as?: { field?: string };
}

export interface SurveyFromFileResult {
  markdown: string;
  title: string;
  slug: string;
  source_name: string;
  source_kind: 'pdf' | 'image';
  source_text: string;
  source_truncated?: boolean;
  question_count: number;
  assistant_message?: string;
}

export interface SurveyBotResult {
  id: string;
  template_id: string;
  template_slug?: string;
  title: string;
  answers?: Record<string, string>;
  html?: string;
  session_id?: string;
  created_by?: string;
  created_at?: string;
}

export interface EventInfoListResponse {
  events: EventInfo[];
  total: number;
  page: number;
  limit: number;
}

export interface AssistantMessage {
  role: 'user' | 'assistant';
  content: string;
}

export interface AssistantConversationState {
  intent?: string;
  fields?: Record<string, string>;
}

export interface AssistantChatResponse {
  assistant_message: string;
  intent?: string;
  missing_fields?: string[];
  state?: AssistantConversationState;
  completed?: boolean;
  record?: unknown;
}

export const api = {
  auth: {
    login: async (body: { email: string; password: string }, opts?: { rememberMe?: boolean }) => {
      const out = await request<{ token: string; user: { email: string; username: string; roles: string[] }; permissions: string[] }>(
        '/api/v1/auth/login',
        { method: 'POST', body: JSON.stringify(body) }
      );
      setAuthToken(out.token, { rememberMe: opts?.rememberMe ?? true });
      return out;
    },
    me: () =>
      request<{ user: { email: string; username: string; roles: string[] }; permissions: string[] }>('/api/v1/auth/me'),
    logout: () => setAuthToken(''),
  },
  forms: {
    list: (page = 1, limit = 20, search = '') =>
      request<FormListResponse>(`/api/v1/forms?page=${page}&limit=${limit}&search=${encodeURIComponent(search)}`),
    get: (id: number) => request<Form>(`/api/v1/forms/${id}`),
    getBySlug: (slug: string) => request<Form>(`/api/v1/forms/by-slug/${slug}`),
    create: (body: {
      name: string;
      description?: string;
      slug: string;
      single_response_only?: boolean;
      exam_mode?: boolean;
    }) => request<Form>('/api/v1/forms', { method: 'POST', body: JSON.stringify(body) }),
    update: (
      id: number,
      body: Partial<{ name: string; description: string; slug: string; single_response_only: boolean; exam_mode: boolean; landing_html: string }>
    ) =>
      request<Form>(`/api/v1/forms/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
    delete: (id: number) => request<void>(`/api/v1/forms/${id}`, { method: 'DELETE' }),
  },
  eventsInfo: {
    list: (page = 1, limit = 100) =>
      request<EventInfoListResponse>(`/api/v1/events-info?page=${page}&limit=${limit}`),
    create: (body: { title: string; detail?: string; reporter?: string; time: string }) =>
      request<EventInfo>('/api/v1/events-info', { method: 'POST', body: JSON.stringify(body) }),
    aiDraft: (body: { prompt: string }) =>
      request<{
        title: string;
        detail: string;
        reporter: string;
        time: string;
        assistant_message?: string;
      }>('/api/v1/events-info/ai-draft', { method: 'POST', body: JSON.stringify(body) }),
    aiIngest: async (opts: { files?: File[]; url?: string; paste?: string }) => {
      const path = '/api/v1/events-info/ai-ingest';
      const method = 'POST';
      const fd = new FormData();
      for (const f of opts.files || []) fd.append('files', f);
      if (opts.url?.trim()) fd.append('url', opts.url.trim());
      if (opts.paste?.trim()) fd.append('paste', opts.paste.trim());
      const res = await fetch(`${API_BASE}${path}`, {
        method,
        headers: getAuthToken() ? { Authorization: `Bearer ${getAuthToken()}` } : {},
        body: fd,
      });
      if (res.status === 401) invalidateServerSession(path, method);
      const payload = (await res.json().catch(() => ({ error: res.statusText }))) as {
        drafts?: Array<{ title: string; detail: string; reporter: string; time: string }>;
        total?: number;
        sources?: string[];
        source_truncated?: boolean;
        assistant_message?: string;
        error?: string;
      };
      if (!res.ok) throw new Error(payload.error || res.statusText);
      return payload;
    },
    delete: (id: string) => request<void>(`/api/v1/events-info/${encodeURIComponent(id)}`, { method: 'DELETE' }),
    collectionInfo: () => request<EventInfoCollectionInfo>('/api/v1/events-info/collection-info'),
    shareEmail: (body: { to: string[]; kind: 'page' | 'api'; message?: string }) =>
      request<{ ok: boolean; sent_to: string[] }>('/api/v1/events-info/share/email', {
        method: 'POST',
        body: JSON.stringify(body),
      }),
  },
  surveyBot: {
    listTemplates: (page = 1, limit = 50, search = '') =>
      request<{ templates: SurveyBotTemplate[]; total: number }>(
        `/api/v1/survey-bot/templates?page=${page}&limit=${limit}&search=${encodeURIComponent(search)}`
      ),
    getTemplate: (id: string) => request<SurveyBotTemplate>(`/api/v1/survey-bot/templates/${encodeURIComponent(id)}`),
    createTemplate: (body: {
      slug?: string;
      title?: string;
      tags?: string[];
      markdown: string;
      summary?: string;
    }) =>
      request<SurveyBotTemplate>('/api/v1/survey-bot/templates', {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    updateTemplate: (
      id: string,
      body: Partial<{ slug: string; title: string; tags: string[]; markdown: string; summary: string }>
    ) =>
      request<SurveyBotTemplate>(`/api/v1/survey-bot/templates/${encodeURIComponent(id)}`, {
        method: 'PUT',
        body: JSON.stringify(body),
      }),
    deleteTemplate: (id: string) =>
      request<{ ok: boolean }>(`/api/v1/survey-bot/templates/${encodeURIComponent(id)}`, { method: 'DELETE' }),
    aiDraft: (body: { query?: string; title_hint?: string; description?: string }) =>
      request<{ markdown: string; research_notes?: string; sources?: unknown[]; assistant_message?: string }>(
        '/api/v1/survey-bot/templates/ai-draft',
        { method: 'POST', body: JSON.stringify(body) }
      ),
    fromFile: async (file: File, opts: { title_hint?: string; instructions?: string } = {}) => {
      const path = '/api/v1/survey-bot/templates/from-file';
      const method = 'POST';
      const fd = new FormData();
      fd.append('file', file);
      if (opts.title_hint) fd.append('title_hint', opts.title_hint);
      if (opts.instructions) fd.append('instructions', opts.instructions);
      const res = await fetch(`${API_BASE}${path}`, {
        method,
        headers: getAuthToken() ? { Authorization: `Bearer ${getAuthToken()}` } : {},
        body: fd,
      });
      if (res.status === 401) invalidateServerSession(path, method);
      const payload = (await res.json().catch(() => ({ error: res.statusText }))) as SurveyFromFileResult & {
        error?: string;
      };
      if (!res.ok) {
        const err = new Error(payload.error || res.statusText) as Error & { markdown?: string };
        // A validation failure still returns the rejected markdown to repair.
        if (payload.markdown) err.markdown = payload.markdown;
        throw err;
      }
      return payload;
    },
    compile: (body: { markdown?: string; description?: string; title_hint?: string; use_web_search?: boolean }) =>
      request<{
        markdown: string;
        compiled?: boolean;
        research_notes?: string;
        sources?: unknown[];
        assistant_message?: string;
        needs_compile?: boolean;
      }>('/api/v1/survey-bot/templates/compile', { method: 'POST', body: JSON.stringify(body) }),
    compileAndSave: (id: string) =>
      request<SurveyBotTemplate>(`/api/v1/survey-bot/templates/${encodeURIComponent(id)}/compile`, {
        method: 'POST',
        body: '{}',
      }),
    publish: (id: string) =>
      request<SurveyBotTemplate>(`/api/v1/survey-bot/templates/${encodeURIComponent(id)}/publish`, {
        method: 'POST',
        body: '{}',
      }),
    unpublish: (id: string) =>
      request<SurveyBotTemplate>(`/api/v1/survey-bot/templates/${encodeURIComponent(id)}/unpublish`, {
        method: 'POST',
        body: '{}',
      }),
    listResults: (page = 1, limit = 50, search = '') =>
      request<{ results: SurveyBotResult[]; total: number }>(
        `/api/v1/survey-bot/results?page=${page}&limit=${limit}&search=${encodeURIComponent(search)}`
      ),
    getResult: (id: string) => request<SurveyBotResult>(`/api/v1/survey-bot/results/${encodeURIComponent(id)}`),
    resultHtmlUrl: (id: string) => `${API_BASE}/api/v1/survey-bot/results/${encodeURIComponent(id)}/html`,
    deleteResult: (id: string) =>
      request<{ ok: boolean }>(`/api/v1/survey-bot/results/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  },
  publicAISheet: {
    get: (slug: string) =>
      request<{ slug: string; title: string; summary?: string; instructions?: string; step_count?: number }>(
        `/api/v1/public/ai-sheets/${encodeURIComponent(slug)}`
      ),
    chat: (slug: string, body: { session_id?: string; message: string; state?: unknown }) =>
      request<{
        message: string;
        state?: unknown;
        ui_blocks?: AISheetUIBlock[];
        done?: boolean;
        record?: unknown;
        title?: string;
        slug?: string;
      }>(`/api/v1/public/ai-sheets/${encodeURIComponent(slug)}/chat`, {
        method: 'POST',
        body: JSON.stringify(body),
      }),
  },
  assistant: {
    chat: (body: { messages: AssistantMessage[]; state?: AssistantConversationState }) =>
      request<AssistantChatResponse>('/api/v1/assistant/chat', { method: 'POST', body: JSON.stringify(body) }),
  },
  pages: {
    list: (formId: number) => request<FormPage[]>(`/api/v1/forms/${formId}/pages`),
    create: (formId: number, body: { name?: string; sort_order?: number }) =>
      request<FormPage>(`/api/v1/forms/${formId}/pages`, { method: 'POST', body: JSON.stringify(body) }),
    update: (formId: number, pageId: number, body: Partial<{ name: string; sort_order: number }>) =>
      request<FormPage>(`/api/v1/forms/${formId}/pages/${pageId}`, { method: 'PUT', body: JSON.stringify(body) }),
    delete: (formId: number, pageId: number) =>
      request<void>(`/api/v1/forms/${formId}/pages/${pageId}`, { method: 'DELETE' }),
  },
  questions: {
    list: (formId: number) => request<Question[]>(`/api/v1/forms/${formId}/questions`),
    create: (formId: number, body: { title: string; type: string; required?: boolean; page_id?: number; sort_order?: number; config?: QuestionConfig }) =>
      request<Question>(`/api/v1/forms/${formId}/questions`, { method: 'POST', body: JSON.stringify(body) }),
    update: (formId: number, id: number, body: Partial<{ title: string; type: string; required: boolean; page_id: number; sort_order: number; config: QuestionConfig }>) =>
      request<Question>(`/api/v1/forms/${formId}/questions/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
    delete: (formId: number, id: number) =>
      request<void>(`/api/v1/forms/${formId}/questions/${id}`, { method: 'DELETE' }),
    uploadPromptMedia: async (formId: number, questionId: number, file: File) => {
      const path = `/api/v1/forms/${formId}/questions/${questionId}/prompt-media`;
      const method = 'POST';
      const fd = new FormData();
      fd.append('file', file);
      const res = await fetch(`${API_BASE}${path}`, {
        method,
        headers: getAuthToken() ? { Authorization: `Bearer ${getAuthToken()}` } : {},
        body: fd,
      });
      if (res.status === 401) invalidateServerSession(path, method);
      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: res.statusText }));
        throw new Error((err as { error?: string }).error || res.statusText);
      }
      return res.json() as Promise<Question>;
    },
    deletePromptMedia: async (formId: number, questionId: number) =>
      request<Question>(`/api/v1/forms/${formId}/questions/${questionId}/prompt-media`, { method: 'DELETE' }),
  },
  rules: {
    listByForm: (formId: number) =>
      request<QuestionRule[]>(`/api/v1/forms/${formId}/rules`),
    list: (formId: number, questionId: number) =>
      request<QuestionRule[]>(`/api/v1/forms/${formId}/questions/${questionId}/rules`),
    create: (formId: number, questionId: number, body: { depends_on_question_id: number; condition: 'answered' | 'not_answered' }) =>
      request<QuestionRule>(`/api/v1/forms/${formId}/questions/${questionId}/rules`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (formId: number, questionId: number, ruleId: number) =>
      request<void>(`/api/v1/forms/${formId}/questions/${questionId}/rules/${ruleId}`, { method: 'DELETE' }),
  },
  responses: {
    list: (formId: number, page = 1, limit = 20) =>
      request<ResponseListResult>(`/api/v1/forms/${formId}/responses?page=${page}&limit=${limit}`),
    delete: (formId: number, responseId: string) =>
      request<void>(`/api/v1/forms/${formId}/responses/${responseId}`, { method: 'DELETE' }),
    email: (formId: number, responseId: string, body: { to: string[]; subject?: string }) =>
      request<{ ok: boolean }>(`/api/v1/forms/${formId}/responses/${responseId}/email`, {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    exportUrl: (formId: number) => `${API_BASE}/api/v1/forms/${formId}/responses/export`,
  },
  public: {
    getForm: (slug: string) => request<{ form: Form; pages: FormPage[]; questions: Question[]; rules: QuestionRule[] }>(`/api/v1/public/forms/${slug}`),
    submit: (
      slug: string,
      body: {
        respondent_id?: string;
        exam_started_at?: string;
        answers: { question_id: number; value: unknown }[];
      }
    ) => request<FormResponse>(`/api/v1/public/forms/${slug}/submit`, { method: 'POST', body: JSON.stringify(body) }),
    createEventInfo: (body: { title: string; detail?: string; reporter?: string; time: string }) =>
      request<EventInfo>('/api/v1/public/events-info', { method: 'POST', body: JSON.stringify(body) }),
  },
};
