export type AttachedDataTable = {
	id: string;
	name: string;
};

export const assistantCtx = $state({
	attachedDataTables: [] as AttachedDataTable[],
	openRequest: 0
});

export function requestAssistantOpen() {
	assistantCtx.openRequest += 1;
}

export function getAttachedDataTables(): AttachedDataTable[] {
	return assistantCtx.attachedDataTables;
}

export function attachDataTable(table: AttachedDataTable) {
	if (assistantCtx.attachedDataTables.some((t) => t.id === table.id)) return;
	assistantCtx.attachedDataTables = [...assistantCtx.attachedDataTables, table];
}

export function detachDataTable(id: string) {
	assistantCtx.attachedDataTables = assistantCtx.attachedDataTables.filter((t) => t.id !== id);
}

export function clearAttachedDataTables() {
	assistantCtx.attachedDataTables = [];
}

export function getAssistantStateExtra(): {
	attached_data_tables: AttachedDataTable[];
} {
	return {
		attached_data_tables: assistantCtx.attachedDataTables.map((t) => ({
			id: t.id,
			name: t.name
		}))
	};
}

export function buildDataTableAnalyzePrompt(tables: AttachedDataTable[]): string {
	if (tables.length === 0) {
		return 'Analyze the attached data table using data_table_schema and query_data_table tools.';
	}
	const list = tables.map((t) => `- ${t.name} (table_id: ${t.id})`).join('\n');
	return `Analyze the attached data table(s) below. Use data_table_schema and query_data_table with the given table_id values.

Attached tables:
${list}

Generate a markdown report with Overview, Key Findings, Risks or Gaps, and Recommended Actions.`;
}
