import * as XLSX from 'xlsx';

/** Split CSV/TSV line respecting double quotes. */
function parseDelimitedLine(line: string, delimiter: string): string[] {
	const out: string[] = [];
	let cur = '';
	let inQuote = false;
	for (let i = 0; i < line.length; i++) {
		const c = line[i];
		if (c === '"') {
			if (inQuote && line[i + 1] === '"') {
				cur += '"';
				i++;
			} else {
				inQuote = !inQuote;
			}
			continue;
		}
		if (!inQuote && c === delimiter) {
			out.push(cur);
			cur = '';
			continue;
		}
		if (!inQuote && (c === '\n' || c === '\r')) break;
		cur += c;
	}
	out.push(cur);
	return out;
}

function splitLines(text: string): string[] {
	return text.replace(/\r\n/g, '\n').split('\n').filter((l) => l.length > 0);
}

function detectDelimiter(firstLine: string): string {
	const commas = (firstLine.match(/,/g) ?? []).length;
	const tabs = (firstLine.match(/\t/g) ?? []).length;
	return tabs > commas ? '\t' : ',';
}

export function parseCsv(text: string): Record<string, unknown>[] {
	const lines = splitLines(text);
	if (lines.length === 0) return [];
	const delim = detectDelimiter(lines[0]);
	const headers = parseDelimitedLine(lines[0], delim).map((h) => h.trim().replace(/^"|"$/g, ''));
	const rows: Record<string, unknown>[] = [];
	for (let i = 1; i < lines.length; i++) {
		const raw = parseDelimitedLine(lines[i], delim);
		if (raw.every((c) => c.trim() === '')) continue;
		const row: Record<string, unknown> = {};
		for (let j = 0; j < headers.length; j++) {
			const cell = raw[j] ?? '';
			row[headers[j]] = coerceCell(cell);
		}
		rows.push(row);
	}
	return rows;
}

function coerceCell(s: string): unknown {
	const t = s.trim();
	if (t === '') return null;
	const lower = t.toLowerCase();
	if (lower === 'true') return true;
	if (lower === 'false') return false;
	const n = Number(t);
	if (!Number.isNaN(n) && String(n) === t) return n;
	return s;
}

export function parseJsonArray(text: string): Record<string, unknown>[] {
	const v = JSON.parse(text) as unknown;
	if (Array.isArray(v)) {
		return v.map((x) => (x && typeof x === 'object' ? (x as Record<string, unknown>) : { value: x }));
	}
	if (v && typeof v === 'object') {
		return [v as Record<string, unknown>];
	}
	throw new Error('JSON must be an array of objects or a single object');
}

export function parseExcelBuffer(buf: ArrayBuffer): Record<string, unknown>[] {
	const wb = XLSX.read(buf, { type: 'array' });
	const name = wb.SheetNames[0];
	if (!name) return [];
	const sheet = wb.Sheets[name];
	const rows = XLSX.utils.sheet_to_json(sheet, { defval: null }) as Record<string, unknown>[];
	return rows;
}

export function columnKeys(rows: Record<string, unknown>[]): string[] {
	if (rows.length === 0) return [];
	return Object.keys(rows[0]);
}

export type FileFilter = { column: string; op: string; value: string };

export function applyFilters(
	rows: Record<string, unknown>[],
	filters: FileFilter[]
): Record<string, unknown>[] {
	let out = rows;
	for (const f of filters) {
		if (!f.column) continue;
		out = out.filter((row) => {
			const left = row[f.column];
			const leftStr = String(left ?? '');
			const leftNum = Number(left);
			const v = f.value;
			switch (f.op) {
				case 'eq':
					return leftStr === v;
				case 'ne':
					return leftStr !== v;
				case 'contains':
					return leftStr.toLowerCase().includes(v.toLowerCase());
				case 'starts_with':
					return leftStr.toLowerCase().startsWith(v.toLowerCase());
				case 'gt':
					return Number.isFinite(leftNum) && Number.isFinite(Number(v)) && leftNum > Number(v);
				case 'gte':
					return Number.isFinite(leftNum) && Number.isFinite(Number(v)) && leftNum >= Number(v);
				case 'lt':
					return Number.isFinite(leftNum) && Number.isFinite(Number(v)) && leftNum < Number(v);
				case 'lte':
					return Number.isFinite(leftNum) && Number.isFinite(Number(v)) && leftNum <= Number(v);
				case 'isnotnull':
					return left !== null && left !== undefined;
				default:
					return true;
			}
		});
	}
	return out;
}

export function sortRows(
	rows: Record<string, unknown>[],
	column: string,
	dir: 'asc' | 'desc'
): Record<string, unknown>[] {
	const copy = [...rows];
	copy.sort((a, b) => {
		const va = a[column];
		const vb = b[column];
		let cmp = 0;
		if (typeof va === 'number' && typeof vb === 'number') cmp = va - vb;
		else cmp = String(va ?? '').localeCompare(String(vb ?? ''));
		return dir === 'asc' ? cmp : -cmp;
	});
	return copy;
}

/** Group by one column and sum another (numeric). */
export function groupBySum(
	rows: Record<string, unknown>[],
	groupCol: string,
	sumCol: string
): Record<string, unknown>[] {
	const m = new Map<string, number>();
	for (const r of rows) {
		const k = String(r[groupCol] ?? '');
		const v = Number(r[sumCol]);
		if (!Number.isFinite(v)) continue;
		m.set(k, (m.get(k) ?? 0) + v);
	}
	return Array.from(m.entries()).map(([k, v]) => ({
		[groupCol]: k,
		[`sum_${sumCol}`]: v
	}));
}
