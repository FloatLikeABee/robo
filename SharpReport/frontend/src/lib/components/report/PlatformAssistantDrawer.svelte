<script lang="ts">
	import PlatformChatDrawer from '@robo/platform-chat/svelte';
	import { inferProgressSteps } from '@robo/platform-chat/aiProgress';
	import '@robo/platform-chat/chat-drawer.css';
	import { apiUrl, authHeaders } from '$lib/api';
	import {
		assistantCtx,
		buildDataTableAnalyzePrompt,
		detachDataTable,
		getAssistantStateExtra
	} from '$lib/stores/assistantContext.svelte';

	/** @type {{ open?: boolean, onClose?: () => void }} */
	let { open = false, onClose = () => {} } = $props();

	const attachments = $derived(
		assistantCtx.attachedDataTables.map((t) => ({ id: t.id, label: t.name }))
	);

	const suggestions = [
		'List my data tables',
		'Analyze the attached data table',
		'Count rows grouped by column',
		'Summarize columns in the attached table',
	];
	const quickActions = [
		{
			label: 'Analyze attached tables',
			prompt:
				'Analyze the attached data table(s) and generate a markdown report with Overview, Key Findings, Risks or Gaps, and Recommended Actions.',
		},
	];

	const chatEndpoint = apiUrl('/api/v1/assistant/chat');
</script>

<PlatformChatDrawer
	{open}
	title="DataX AI"
	{chatEndpoint}
	getHeaders={authHeaders}
	welcomeMessage="Hi! I am your **Data Access assistant**. I can help with data tables and file-based data reports. Add a data table to this conversation from any table page to analyze, search, and aggregate its rows."
	{suggestions}
	{quickActions}
	enableFileAnalyze={true}
	{attachments}
	onRemoveAttachment={detachDataTable}
	buildDataTableAnalyzePrompt={() => buildDataTableAnalyzePrompt(assistantCtx.attachedDataTables)}
	getStateExtra={getAssistantStateExtra}
	progressContext={{ app: 'datax' }}
	getProgressSteps={(userText) =>
		inferProgressSteps({
			app: 'datax',
			userText,
			hasDataTables: assistantCtx.attachedDataTables.length > 0,
		})}
	on:close={onClose}
/>
