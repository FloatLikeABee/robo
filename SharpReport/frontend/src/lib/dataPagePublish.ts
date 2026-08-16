import { apiJson, apiUrl, authHeaders } from '$lib/api';

export type PublishSummary = {
	id: string;
	data_table_id: string;
	name: string;
	slug: string;
	theme: string;
	public_path: string;
	created_at: string;
	updated_at: string;
};

export type PageAIResponse = {
	assistant_message: string;
	proposed_page_html?: string | null;
};

export type PageChatMessage = {
	role: 'user' | 'assistant';
	content: string;
};

export type PageBuildSummary = {
	id: string;
	data_table_id: string;
	label: string;
	source: 'build' | 'chat' | string;
	created_at: string;
};

export type PageBuildDetail = PageBuildSummary & {
	html_content: string;
};

export async function listPageBuilds(dataTableId: string): Promise<{ items: PageBuildSummary[] }> {
	return apiJson(`/api/v1/data-tables/${encodeURIComponent(dataTableId)}/page-builds`);
}

export async function getPageBuild(dataTableId: string, buildId: string): Promise<PageBuildDetail> {
	return apiJson(
		`/api/v1/data-tables/${encodeURIComponent(dataTableId)}/page-builds/${encodeURIComponent(buildId)}`
	);
}

export async function resolvePublishPath(name: string): Promise<{ slug: string; public_path: string }> {
	return apiJson(`/api/v1/publishes/resolve-path?name=${encodeURIComponent(name)}`);
}

export async function listPublishes(
	dataTableId: string,
	limit = 50,
	offset = 0
): Promise<{ items: PublishSummary[]; total: number }> {
	const q = new URLSearchParams({
		data_table_id: dataTableId,
		limit: String(limit),
		offset: String(offset)
	});
	return apiJson(`/api/v1/publishes?${q}`);
}

export async function createPublish(payload: {
	data_table_id: string;
	name: string;
	theme?: string;
	html_content: string;
}): Promise<PublishSummary> {
	return apiJson('/api/v1/publishes', {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export async function buildDataPage(tableId: string): Promise<PageAIResponse> {
	return apiJson(`/api/v1/data-tables/${tableId}/page-build`, { method: 'POST' });
}

export async function chatDataPage(
	tableId: string,
	payload: {
		messages: PageChatMessage[];
		current_html?: string;
		theme?: string;
	}
): Promise<PageAIResponse> {
	return apiJson(`/api/v1/data-tables/${tableId}/page-chat`, {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export function buildDefaultPagePrompt(tableName: string): string {
	const name = tableName.trim() || 'this data table';
	return `Design a polished dim-slate public dashboard for "${name}". Use clear section headings, KPI cards in the summary area, and a readable data table. Keep generous padding and a balanced layout — between light and dark, leaning slightly dark.`;
}

export function buildSafePreviewHtml(rawHtml: string, mode: 'light' | 'dark' | 'dim' = 'dim'): string {
	const html = String(rawHtml || '').trim();
	const baseTag = '<base target="_blank">';
	const dim = mode === 'dim' || mode === 'dark';
	const bg = dim ? '#13171f' : '#ffffff';
	const fg = dim ? '#e9ecf1' : '#0f172a';
	const thumb = dim ? '#3d4658' : '#cbd5e1';
	const track = dim ? '#1a2030' : '#f1f5f9';
	const previewStyle = `<style id="dx-preview-reset">html,body{min-height:100%;scrollbar-width:thin;scrollbar-color:${thumb} ${track};}body{background:${bg};color:${fg};}::-webkit-scrollbar{width:8px;height:8px;}::-webkit-scrollbar-track{background:${track};}::-webkit-scrollbar-thumb{background:${thumb};border-radius:999px;border:2px solid ${track};}::-webkit-scrollbar-thumb:hover{background:${dim ? '#556070' : '#94a3b8'};}</style>`;
	if (!html) {
		return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">${baseTag}<style>html,body{margin:0;padding:0}body{min-height:100vh;background:${bg};color:${fg};font-family:Inter,system-ui,sans-serif}</style></head><body><main style="padding:20px;font-size:12px;opacity:.8">Building page…</main></body></html>`;
	}
	const hasHtml = /<html[\s>]/i.test(html);
	const hasHead = /<head[\s>]/i.test(html);
	if (!hasHtml) {
		return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">${baseTag}<style>html,body{margin:0;padding:0}body{min-height:100vh;background:${bg};color:${fg}}</style></head><body>${html}</body></html>`;
	}
	if (!hasHead) {
		return html.replace(/<html([^>]*)>/i, `<html$1><head>${baseTag}${previewStyle}</head>`);
	}
	return html.replace(/<head([^>]*)>/i, `<head$1>${baseTag}${previewStyle}`);
}

export function publicPageUrl(publicPath: string): string {
	const path = publicPath.startsWith('/') ? publicPath : `/${publicPath}`;
	if (typeof window !== 'undefined') {
		return `${window.location.origin}${path}`;
	}
	return path;
}

export async function copyPublicUrl(publicPath: string): Promise<void> {
	const url = publicPageUrl(publicPath);
	if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
		await navigator.clipboard.writeText(url);
		return;
	}
	throw new Error('Clipboard not available');
}

export async function fetchPublicPageHtml(publicPath: string): Promise<string> {
	const path = publicPath.startsWith('/') ? publicPath : `/${publicPath}`;
	const res = await fetch(apiUrl(path));
	if (!res.ok) throw new Error(`Failed to load page (${res.status})`);
	return res.text();
}

export function authFetchPublic(path: string, init: RequestInit = {}): Promise<Response> {
	return fetch(apiUrl(path), {
		...init,
		headers: {
			'Content-Type': 'application/json',
			...authHeaders(),
			...(init.headers as Record<string, string>)
		}
	});
}
