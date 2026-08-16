import { apiJson, apiUrl, getAuthToken } from './api';

const BRIDGE = '/api/v1/docs-bridge';

export type DocsDocument = {
	id: string;
	title?: string;
	type?: string;
	content?: string;
	tags?: string[];
	created_at?: string;
	updated_at?: string;
};

export type DocsSession = {
	id: string;
	title?: string;
	messages?: Array<{ role?: string; content?: string; text?: string; isUser?: boolean }>;
	created_at?: string;
	updated_at?: string;
};

export const DOCS_CONTENT_ACCEPT =
	'.pdf,.txt,.md,.markdown,.png,.jpg,.jpeg,.webp,.gif,.heic,.heif,application/pdf,text/plain,text/markdown,image/*';

export const DATA_TABLE_ACCEPT = '.csv,.tsv,.json,.xlsx,.xls';

export function isDocsContentFile(name: string): boolean {
	const lower = name.toLowerCase();
	return ['.pdf', '.txt', '.md', '.markdown', '.png', '.jpg', '.jpeg', '.webp', '.gif', '.heic', '.heif'].some(
		(ext) => lower.endsWith(ext)
	);
}

export function isDataTableFile(name: string): boolean {
	const lower = name.toLowerCase();
	return ['.csv', '.tsv', '.json', '.xlsx', '.xls'].some((ext) => lower.endsWith(ext));
}

async function bridgeJson<T>(path: string, init: RequestInit = {}): Promise<T> {
	const token = getAuthToken();
	const headers = new Headers(init.headers || {});
	if (token) headers.set('Authorization', `Bearer ${token}`);
	if (init.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}
	const res = await fetch(apiUrl(`${BRIDGE}/${path.replace(/^\//, '')}`), {
		...init,
		headers
	});
	if (!res.ok) {
		const text = await res.text();
		let msg = text;
		try {
			const j = JSON.parse(text);
			msg = j.error || j.message || text;
		} catch {
			/* raw */
		}
		throw new Error(msg || `Docs API ${res.status}`);
	}
	if (res.status === 204) return undefined as T;
	const ct = res.headers.get('content-type') || '';
	if (ct.includes('application/json')) return (await res.json()) as T;
	return (await res.text()) as T;
}

export async function listDocs(): Promise<DocsDocument[]> {
	const data = await bridgeJson<DocsDocument[] | { documents?: DocsDocument[] }>('docs');
	if (Array.isArray(data)) return data;
	return data.documents ?? [];
}

export async function getDoc(id: string): Promise<DocsDocument> {
	return bridgeJson<DocsDocument>(`docs/${encodeURIComponent(id)}`);
}

export async function createDoc(payload: {
	title: string;
	content: string;
	type?: string;
	tags?: string[];
}): Promise<DocsDocument> {
	return bridgeJson<DocsDocument>('docs', {
		method: 'POST',
		body: JSON.stringify({
			title: payload.title,
			content: payload.content,
			type: payload.type || 'markdown',
			tags: payload.tags || ['#notes']
		})
	});
}

export async function uploadDoc(file: File, title?: string): Promise<DocsDocument> {
	if (isDataTableFile(file.name) && !file.name.toLowerCase().endsWith('.txt')) {
		throw new Error('CSV/JSON belong in Data tables. Docs accepts PDF, TXT, Markdown, or images.');
	}
	if (!isDocsContentFile(file.name)) {
		throw new Error('Unsupported Docs content type. Use PDF or TXT (Markdown/images also allowed).');
	}
	const fd = new FormData();
	fd.append('file', file);
	if (title?.trim()) fd.append('title', title.trim());
	return bridgeJson<DocsDocument>('docs/upload', { method: 'POST', body: fd });
}

export async function deleteDoc(id: string): Promise<void> {
	await bridgeJson(`docs/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

export async function listSessions(): Promise<DocsSession[]> {
	const data = await bridgeJson<DocsSession[] | { sessions?: DocsSession[] }>('chat-sessions');
	if (Array.isArray(data)) return data;
	return data.sessions ?? [];
}

export async function createSession(): Promise<DocsSession> {
	return bridgeJson<DocsSession>('chat-sessions', { method: 'POST', body: JSON.stringify({}) });
}

export async function getSession(id: string): Promise<DocsSession> {
	return bridgeJson<DocsSession>(`chat-sessions/${encodeURIComponent(id)}`);
}

export async function deleteSession(id: string): Promise<void> {
	await bridgeJson(`chat-sessions/${encodeURIComponent(id)}`, { method: 'DELETE' });
}

export async function patchSession(
	id: string,
	payload: Record<string, unknown>
): Promise<DocsSession> {
	return bridgeJson<DocsSession>(`chat-sessions/${encodeURIComponent(id)}`, {
		method: 'PATCH',
		body: JSON.stringify(payload)
	});
}

export async function docsAiChat(body: {
	messages: Array<{ role: string; content: string }>;
	doc_ids?: string[];
	document_mode?: boolean;
}): Promise<{ response: string; sources?: unknown[] }> {
	return bridgeJson('ai/chat', {
		method: 'POST',
		body: JSON.stringify({
			messages: body.messages,
			context: {},
			document_mode: body.document_mode ?? true,
			disable_research: true,
			doc_ids: body.doc_ids ?? [],
			help_you_learn: false
		})
	});
}

export type DocsPublishResult = {
	id: string;
	slug: string;
	title: string;
	public_path: string;
	has_analysis: boolean;
};

export async function publishDoc(payload: {
	title: string;
	content: string;
	doc_id?: string;
	analysis_prompt?: string;
}): Promise<DocsPublishResult> {
	return apiJson<DocsPublishResult>('/api/v1/docs-publish', {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export async function listDocPublishes(): Promise<
	Array<{ id: string; title: string; slug: string; public_path: string; created_at: string }>
> {
	const data = await apiJson<{ publishes: Array<{ id: string; title: string; slug: string; public_path: string; created_at: string }> }>(
		'/api/v1/docs-publish'
	);
	return data.publishes ?? [];
}
