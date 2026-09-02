import { browser } from '$app/environment';

export function analysisDownloadFilename(tableName: string): string {
	const slug =
		tableName
			.trim()
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/^-+|-+$/g, '') || 'table';
	return `${slug}-analysis.md`;
}

export async function renderAnalysisMarkdown(markdown: string): Promise<string> {
	if (!browser) return '';
	const [{ marked }, purifyMod] = await Promise.all([import('marked'), import('dompurify')]);
	const html = marked.parse(markdown, { async: false }) as string;
	return purifyMod.default.sanitize(html);
}
