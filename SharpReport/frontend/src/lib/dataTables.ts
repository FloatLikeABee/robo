import { apiJson } from '$lib/api';
import {
	columnKeys,
	parseCsv,
	parseExcelBuffer,
	parseJsonArray
} from '$lib/report/fileData';

export type DataTableSummary = {
	id: string;
	name: string;
	source_filename: string | null;
	source_format: string;
	columns: string[];
	row_count: number;
	created_at: string;
	updated_at: string;
};

export type DataTableRowsResponse = {
	columns: string[];
	rows: Record<string, unknown>[];
	total: number;
	limit: number;
	offset: number;
};

export type AnalyzeResult = {
	can_be_table: boolean;
	reason: string;
	columns: string[];
	rows: Record<string, unknown>[];
};

const STRUCTURED = new Set(['csv', 'tsv', 'json', 'xlsx', 'xls']);
const TEXT_FORMATS = new Set(['txt', 'md', 'pdf']);

export function fileExtension(name: string): string {
	const i = name.lastIndexOf('.');
	return i >= 0 ? name.slice(i + 1).toLowerCase() : '';
}

export async function listDataTables(): Promise<DataTableSummary[]> {
	return apiJson<DataTableSummary[]>('/api/v1/data-tables');
}

export async function getDataTable(id: string): Promise<DataTableSummary> {
	return apiJson<DataTableSummary>(`/api/v1/data-tables/${id}`);
}

export type DataTableRowsOptions = {
	limit?: number;
	offset?: number;
	search?: string;
	sortBy?: string;
	sortDir?: 'asc' | 'desc';
};

export async function getDataTableRows(
	id: string,
	limitOrOpts: number | DataTableRowsOptions = 50,
	offset = 0
): Promise<DataTableRowsResponse> {
	const opts: DataTableRowsOptions =
		typeof limitOrOpts === 'number'
			? { limit: limitOrOpts, offset }
			: limitOrOpts;
	const q = new URLSearchParams({
		limit: String(opts.limit ?? 50),
		offset: String(opts.offset ?? 0)
	});
	if (opts.search?.trim()) q.set('search', opts.search.trim());
	if (opts.sortBy?.trim()) q.set('sort_by', opts.sortBy.trim());
	if (opts.sortDir) q.set('sort_dir', opts.sortDir);
	return apiJson<DataTableRowsResponse>(`/api/v1/data-tables/${id}/rows?${q}`);
}

export async function queryDataTable(
	id: string,
	payload: {
		search?: string;
		sort_by?: string;
		sort_dir?: 'asc' | 'desc';
		limit?: number;
		offset?: number;
		group_by?: string;
		aggregate_op?: 'count' | 'sum' | 'avg' | 'min' | 'max';
		aggregate_column?: string;
	}
): Promise<{
	table_id: string;
	name: string;
	columns: string[];
	total: number;
	rows: Record<string, unknown>[];
	aggregate: unknown;
}> {
	return apiJson(`/api/v1/data-tables/${id}/query`, {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export async function analyzeTextFile(
	sourceFilename: string,
	sourceFormat: string,
	contentText: string
): Promise<AnalyzeResult> {
	return apiJson<AnalyzeResult>('/api/v1/data-tables/analyze', {
		method: 'POST',
		body: JSON.stringify({
			source_filename: sourceFilename,
			source_format: sourceFormat,
			content_text: contentText
		})
	});
}

export type ImportDataTableResult = DataTableSummary & {
	index_to_graph?: boolean;
	graph_status?: string;
};

export async function importDataTable(payload: {
	name?: string;
	source_filename?: string;
	source_format: string;
	columns: string[];
	rows: Record<string, unknown>[];
	index_to_graph?: boolean;
}): Promise<ImportDataTableResult> {
	return apiJson<ImportDataTableResult>('/api/v1/data-tables/import', {
		method: 'POST',
		body: JSON.stringify(payload)
	});
}

export async function updateDataTableRow(
	tableId: string,
	rowIndex: number,
	data: Record<string, unknown>
): Promise<void> {
	await apiJson(`/api/v1/data-tables/${tableId}/rows/${rowIndex}`, {
		method: 'PUT',
		body: JSON.stringify({ data })
	});
}

export async function deleteDataTableRow(tableId: string, rowIndex: number): Promise<void> {
	await apiJson(`/api/v1/data-tables/${tableId}/rows/${rowIndex}`, {
		method: 'DELETE'
	});
}

export async function deleteDataTable(id: string): Promise<void> {
	await apiJson(`/api/v1/data-tables/${id}`, { method: 'DELETE' });
}

async function readFileAsText(file: File): Promise<string> {
	return file.text();
}

export async function processUploadFile(file: File): Promise<{
	columns: string[];
	rows: Record<string, unknown>[];
	source_format: string;
	reason?: string;
}> {
	const ext = fileExtension(file.name);
	const source_format = ext || 'txt';

	if (ext === 'pdf' || ext === 'md' || ext === 'markdown') {
		throw new Error(
			'PDF/Markdown are Docs content files. Open Data Access → Docs to upload them. Data tables accept CSV, TSV, JSON, or Excel.'
		);
	}

	if (STRUCTURED.has(ext)) {
		if (ext === 'json') {
			const text = await readFileAsText(file);
			const rows = parseJsonArray(text);
			return { columns: columnKeys(rows), rows, source_format: 'json' };
		}
		if (ext === 'csv' || ext === 'tsv') {
			const text = await readFileAsText(file);
			const rows = parseCsv(text);
			return { columns: columnKeys(rows), rows, source_format: ext };
		}
		if (ext === 'xlsx' || ext === 'xls') {
			const buf = await file.arrayBuffer();
			const rows = parseExcelBuffer(buf);
			return { columns: columnKeys(rows), rows, source_format: ext };
		}
	}

	if (TEXT_FORMATS.has(ext) || !STRUCTURED.has(ext)) {
		const text = await readFileAsText(file);
		const analyzed = await analyzeTextFile(file.name, source_format, text);
		if (!analyzed.can_be_table || analyzed.rows.length === 0) {
			throw new Error(analyzed.reason || 'This file could not be converted into a table.');
		}
		return {
			columns: analyzed.columns,
			rows: analyzed.rows as Record<string, unknown>[],
			source_format,
			reason: analyzed.reason
		};
	}

	throw new Error(`Unsupported file type: .${ext}`);
}
