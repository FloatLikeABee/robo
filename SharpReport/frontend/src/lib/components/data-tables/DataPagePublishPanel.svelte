<script lang="ts">
	import {
		buildDataPage,
		buildDefaultPagePrompt,
		buildSafePreviewHtml,
		chatDataPage,
		createPublish,
		getPageBuild,
		listPageBuilds,
		publicPageUrl,
		resolvePublishPath,
		type PageBuildSummary,
		type PageChatMessage
	} from '$lib/dataPagePublish';
	import { runAiProgress } from '@robo/platform-chat/aiProgress';
	import {
		ArrowLeft,
		Check,
		Copy,
		ExternalLink,
		Globe,
		History,
		Loader2,
		Send,
		Wand2,
		X
	} from 'lucide-svelte';

	let {
		tableId,
		tableName = 'Data table',
		onBack
	}: {
		tableId: string;
		tableName?: string;
		onBack: () => void;
	} = $props();

	let pageHtml = $state('');
	let publishName = $state('');
	let publishPath = $state('');
	let pathLoading = $state(false);
	let buildLoading = $state(false);
	let publishLoading = $state(false);
	let lastPublishedUrl = $state('');
	let error = $state('');
	let success = $state('');

	let aiChatLoading = $state(false);
	let aiLoadingStatus = $state('');
	let aiResponseTimeMs = $state<number | null>(null);
	let aiMessages = $state<Array<{ id: number; role: string; content: string }>>([]);
	let aiInput = $state('');
	let aiMsgSeq = 0;
	let pathResolveSeq = 0;

	let buildHistory = $state<PageBuildSummary[]>([]);
	let activeBuildId = $state<string | null>(null);
	let historyModalOpen = $state(false);
	let historyLoading = $state(false);
	let restoreLoadingId = $state<string | null>(null);

	const previewHtml = $derived(buildSafePreviewHtml(pageHtml, 'dim'));

	function formatBuildTime(iso: string): string {
		try {
			return new Date(iso).toLocaleString(undefined, {
				dateStyle: 'medium',
				timeStyle: 'short'
			});
		} catch {
			return iso;
		}
	}

	function sourceLabel(source: string): string {
		return source === 'build' ? 'Quick rebuild' : 'AI chat';
	}

	function formatResponseTime(ms: number | null): string {
		if (ms == null) return '';
		if (ms < 1000) return `${ms} ms`;
		return `${(ms / 1000).toFixed(ms < 10000 ? 1 : 0)} s`;
	}

	async function refreshBuildHistory() {
		if (!tableId) return [];
		try {
			const data = await listPageBuilds(tableId);
			const items = data.items ?? [];
			buildHistory = items;
			return items;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load saved builds';
			return buildHistory;
		}
	}

	async function loadBuildSnapshot(buildId: string, quiet = false) {
		restoreLoadingId = buildId;
		error = '';
		try {
			const detail = await getPageBuild(tableId, buildId);
			pageHtml = detail.html_content;
			activeBuildId = detail.id;
			if (!quiet) {
				success = `Loaded “${detail.label}”`;
				historyModalOpen = false;
			}
		} catch (e) {
			if (!quiet) {
				error = e instanceof Error ? e.message : 'Failed to load saved build';
			}
		} finally {
			restoreLoadingId = null;
		}
	}

	$effect(() => {
		const id = tableId;
		const name = tableName;
		if (!id) return;

		pageHtml = '';
		activeBuildId = null;
		buildHistory = [];
		aiMessages = [];
		error = '';
		success = '';

		let cancelled = false;
		void (async () => {
			historyLoading = true;
			publishName = `${name} dashboard`;
			aiInput = buildDefaultPagePrompt(name);
			try {
				const data = await listPageBuilds(id);
				if (cancelled) return;
				buildHistory = data.items ?? [];
				if (buildHistory.length > 0) {
					await loadBuildSnapshot(buildHistory[0].id, true);
					if (cancelled) return;
				}
				aiMessages = [
					{
						id: ++aiMsgSeq,
						role: 'assistant',
						content:
							buildHistory.length > 0
								? 'Restored your last saved build in the preview. Refine it here or pick another from Saved builds.'
								: 'Your build prompt is ready below — edit it, then press Send to generate the page. Follow up here to refine layout, colors, or sections.'
					}
				];
			} catch {
				if (!cancelled) {
					aiMessages = [
						{
							id: ++aiMsgSeq,
							role: 'assistant',
							content:
								'Your build prompt is ready below — edit it, then press Send to generate the page. Follow up here to refine layout, colors, or sections.'
						}
					];
				}
			} finally {
				if (!cancelled) historyLoading = false;
			}
		})();

		return () => {
			cancelled = true;
		};
	});

	let buildLoadingStatus = $state('');

	async function runBuild() {
		buildLoading = true;
		buildLoadingStatus = 'Building the data view…';
		error = '';
		const stopProgress = runAiProgress({ app: 'datax', userText: 'quick rebuild dashboard' }, (status) => {
			buildLoadingStatus = status;
		});
		try {
			const res = await buildDataPage(tableId);
			if (res.proposed_page_html) {
				pageHtml = res.proposed_page_html;
			}
			const items = await refreshBuildHistory();
			if (items[0]) activeBuildId = items[0].id;
			aiMessages = [
				...aiMessages,
				{
					id: ++aiMsgSeq,
					role: 'assistant',
					content: res.assistant_message
				}
			];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to build page';
		} finally {
			stopProgress();
			buildLoading = false;
			buildLoadingStatus = '';
		}
	}

	async function resolvePath(name: string, quiet = true) {
		pathLoading = true;
		try {
			const data = await resolvePublishPath(name);
			publishPath = data.public_path || '';
			if (!quiet) success = `Resolved path: ${publishPath}`;
		} catch (e) {
			if (!quiet) error = e instanceof Error ? e.message : 'Failed to resolve path';
		} finally {
			pathLoading = false;
		}
	}

	$effect(() => {
		const name = publishName.trim();
		if (!name) {
			publishPath = '';
			return;
		}
		const seq = ++pathResolveSeq;
		const timer = setTimeout(async () => {
			await resolvePath(name, true);
			if (seq !== pathResolveSeq) return;
		}, 220);
		return () => clearTimeout(timer);
	});

	async function sendAiMessage() {
		const text = aiInput.trim();
		if (!text || aiChatLoading) return;
		aiInput = '';
		aiMessages = [...aiMessages, { id: ++aiMsgSeq, role: 'user', content: text }];
		aiChatLoading = true;
		aiResponseTimeMs = null;
		aiLoadingStatus = 'Reading your question…';
		const startedAt = performance.now();
		const stopProgress = runAiProgress(
			{ app: 'datax', userText: text },
			(status) => {
				aiLoadingStatus = status;
			}
		);
		error = '';
		try {
			const history: PageChatMessage[] = aiMessages.map((m) => ({
				role: m.role === 'assistant' ? 'assistant' : 'user',
				content: m.content
			}));
			const res = await chatDataPage(tableId, {
				messages: history,
				current_html: pageHtml || undefined,
				theme: 'dim'
			});
			aiResponseTimeMs = Math.round(performance.now() - startedAt);
			aiMessages = [
				...aiMessages,
				{ id: ++aiMsgSeq, role: 'assistant', content: res.assistant_message }
			];
			if (res.proposed_page_html) {
				pageHtml = res.proposed_page_html;
			}
			const items = await refreshBuildHistory();
			if (items[0]) activeBuildId = items[0].id;
		} catch (e) {
			aiResponseTimeMs = Math.round(performance.now() - startedAt);
			error = e instanceof Error ? e.message : 'AI chat failed';
		} finally {
			stopProgress();
			aiChatLoading = false;
			aiLoadingStatus = '';
		}
	}

	async function publishPage() {
		const name = publishName.trim();
		const html = pageHtml.trim();
		if (!name || !html) {
			error = 'Page name and HTML content are required';
			return;
		}
		publishLoading = true;
		error = '';
		success = '';
		try {
			const row = await createPublish({
				data_table_id: tableId,
				name,
				theme: 'dim',
				html_content: html
			});
			lastPublishedUrl = publicPageUrl(row.public_path);
			publishPath = row.public_path;
			success = 'Published';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Publish failed';
		} finally {
			publishLoading = false;
		}
	}

	async function copyLink() {
		if (!lastPublishedUrl) return;
		try {
			await navigator.clipboard.writeText(lastPublishedUrl);
			success = 'Link copied';
		} catch {
			error = 'Could not copy link';
		}
	}
</script>

<div class="flex min-h-0 flex-1 flex-col gap-3 overflow-hidden">
	<div class="flex shrink-0 flex-wrap items-start justify-between gap-3">
		<div class="min-w-0">
			<button
				type="button"
				class="inline-flex items-center gap-1 text-sm text-accent-primary hover:underline"
				onclick={onBack}
			>
				<ArrowLeft class="h-4 w-4" />
				Back to table
			</button>
			<h1 class="mt-1 text-2xl font-bold text-text-primary">Build & publish page</h1>
			<p class="text-sm text-text-secondary">
				AI-generated dashboard for <span class="font-medium text-text-primary">{tableName}</span>
			</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<button
				type="button"
				class="inline-flex items-center gap-1.5 rounded-lg border border-border px-3 py-1.5 text-sm text-text-secondary hover:bg-bg-tertiary disabled:opacity-50"
				disabled={historyLoading}
				onclick={async () => {
					historyModalOpen = true;
					await refreshBuildHistory();
				}}
			>
				{#if historyLoading}
					<Loader2 class="h-4 w-4 animate-spin" />
				{:else}
					<History class="h-4 w-4" />
				{/if}
				Saved builds{#if buildHistory.length > 0}&nbsp;({buildHistory.length}){/if}
			</button>
			<button
				type="button"
				class="inline-flex items-center gap-1.5 rounded-lg border border-border px-3 py-1.5 text-sm text-text-secondary hover:bg-bg-tertiary disabled:opacity-50"
				disabled={buildLoading || aiChatLoading}
				onclick={() => runBuild()}
			>
				{#if buildLoading}
					<Loader2 class="h-4 w-4 animate-spin" />
				{:else}
					<Wand2 class="h-4 w-4" />
				{/if}
				Quick rebuild
			</button>
			<button
				type="button"
				class="inline-flex items-center gap-1.5 rounded-lg bg-accent-primary px-3 py-1.5 text-sm font-medium text-white hover:opacity-90 disabled:opacity-50"
				disabled={publishLoading || !pageHtml.trim()}
				onclick={() => publishPage()}
			>
				{#if publishLoading}
					<Loader2 class="h-4 w-4 animate-spin" />
				{:else}
					<Globe class="h-4 w-4" />
				{/if}
				Publish
			</button>
		</div>
	</div>

	<div class="grid shrink-0 gap-3 rounded-xl border border-border bg-bg-elevated p-4 sm:grid-cols-2">
		<label class="block text-sm">
			<span class="font-medium text-text-primary">Publish name</span>
			<input
				class="mt-1 w-full rounded-lg border border-border bg-bg-primary px-3 py-2 text-text-primary"
				bind:value={publishName}
				placeholder="My data dashboard"
			/>
		</label>
		<div class="text-sm">
			<span class="font-medium text-text-primary">Public URL</span>
			<div class="mt-1 flex items-center gap-2">
				<code class="min-w-0 flex-1 truncate rounded-lg border border-border bg-bg-primary px-3 py-2 text-text-secondary">
					{pathLoading ? 'Resolving…' : publishPath || '—'}
				</code>
			</div>
			{#if publishPath}
				<p class="mt-1 text-xs text-text-tertiary">
					Full link: {publicPageUrl(publishPath)}
				</p>
			{/if}
		</div>
	</div>

	{#if error}
		<p class="shrink-0 text-sm text-danger">{error}</p>
	{/if}

	{#if lastPublishedUrl}
		<div
			class="flex shrink-0 flex-wrap items-center gap-2 rounded-lg border border-accent-primary/30 bg-accent-primary/5 px-3 py-2 text-sm"
		>
			<Check class="h-4 w-4 shrink-0 text-accent-primary" />
			<span class="shrink-0 text-text-secondary">{success || 'Published'}</span>
			<a
				href={lastPublishedUrl}
				target="_blank"
				rel="noopener noreferrer"
				class="min-w-0 truncate font-medium text-accent-primary hover:underline"
			>
				{lastPublishedUrl}
			</a>
			<button
				type="button"
				class="inline-flex shrink-0 items-center gap-1 rounded-md border border-border px-2 py-1 text-xs hover:bg-bg-tertiary"
				onclick={() => copyLink()}
			>
				<Copy class="h-3.5 w-3.5" />
				Copy
			</button>
			<a
				href={lastPublishedUrl}
				target="_blank"
				rel="noopener noreferrer"
				class="inline-flex shrink-0 items-center gap-1 rounded-md border border-border px-2 py-1 text-xs hover:bg-bg-tertiary"
			>
				<ExternalLink class="h-3.5 w-3.5" />
				Open
			</a>
		</div>
	{:else if success}
		<p class="shrink-0 text-sm text-accent-primary">{success}</p>
	{/if}

	<div class="grid min-h-0 flex-1 gap-4 overflow-hidden lg:grid-cols-[minmax(0,1fr)_minmax(380px,46%)]">
		<div class="flex min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-bg-elevated">
			<div class="shrink-0 border-b border-border px-4 py-2 text-sm font-medium text-text-primary">
				Preview
				{#if buildLoading || aiChatLoading}
					<span class="ml-2 text-text-tertiary">{buildLoading ? buildLoadingStatus || 'Generating…' : 'Generating…'}</span>
				{/if}
			</div>
			<div class="publish-preview-frame min-h-0 flex-1 overflow-hidden bg-[#13171f]">
				<iframe
					title="Page preview"
					class="h-full w-full border-0 bg-[#13171f]"
					srcdoc={previewHtml}
				></iframe>
			</div>
		</div>

		<div class="flex min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-bg-elevated">
			<div class="shrink-0 border-b border-border px-4 py-2 text-sm font-medium text-text-primary">
				AI assistant
			</div>
			<div class="publish-panel-scroll min-h-0 flex-1 overflow-y-auto overscroll-y-contain p-4 space-y-3">
				{#each aiMessages as msg (msg.id)}
					<div
						class="rounded-lg px-3 py-2.5 text-sm leading-relaxed {msg.role === 'user'
							? 'ml-2 bg-accent-primary/10 text-text-primary'
							: 'mr-2 bg-bg-tertiary text-text-secondary'}"
					>
						{msg.content}
					</div>
				{/each}
				{#if aiChatLoading}
					<div class="flex items-center gap-2 text-sm text-text-tertiary">
						<Loader2 class="h-4 w-4 animate-spin shrink-0" />
						{aiLoadingStatus || 'Working…'}
					</div>
				{/if}
			</div>
			<div class="shrink-0 border-t border-border p-4">
				<form
					class="flex flex-col gap-3"
					onsubmit={(e) => {
						e.preventDefault();
						sendAiMessage();
					}}
				>
					<label class="text-xs font-medium text-text-secondary" for="publish-ai-prompt">
						Prompt
					</label>
					<textarea
						id="publish-ai-prompt"
						class="publish-panel-scroll min-h-[7.5rem] w-full resize-y rounded-lg border border-border bg-bg-primary px-3 py-3 text-sm leading-relaxed text-text-primary placeholder:text-text-tertiary focus:border-accent-primary/50 focus:outline-none focus:ring-1 focus:ring-accent-primary/30"
						placeholder="Describe the dashboard you want…"
						rows="4"
						bind:value={aiInput}
						disabled={aiChatLoading}
						onkeydown={(e) => {
							if (e.key === 'Enter' && !e.shiftKey) {
								e.preventDefault();
								sendAiMessage();
							}
						}}
					></textarea>
					<div class="flex items-center justify-between gap-2">
						<p class="text-xs text-text-tertiary">
							Enter to send · Shift+Enter for new line
							{#if aiResponseTimeMs != null}
								<span class="ml-2 text-accent-primary tabular-nums">Response {formatResponseTime(aiResponseTimeMs)}</span>
							{/if}
						</p>
						<button
							type="submit"
							class="inline-flex items-center gap-2 rounded-lg bg-accent-primary px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
							disabled={aiChatLoading || !aiInput.trim()}
						>
							{#if aiChatLoading}
								<Loader2 class="h-4 w-4 animate-spin" />
							{:else}
								<Send class="h-4 w-4" />
							{/if}
							{pageHtml.trim() ? 'Send' : 'Build page'}
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
</div>

{#if historyModalOpen}
	<div class="fixed inset-0 z-[300] flex items-center justify-center p-3 sm:p-6" role="presentation">
		<button
			type="button"
			class="absolute inset-0 bg-black/65 backdrop-blur-[2px]"
			aria-label="Close saved builds"
			onclick={() => (historyModalOpen = false)}
		></button>
		<div
			class="relative flex max-h-[min(90vh,640px)] w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-bg-elevated shadow-2xl"
			role="dialog"
			aria-modal="true"
			aria-labelledby="saved-builds-title"
		>
			<div class="flex shrink-0 items-center justify-between gap-3 border-b border-border px-4 py-3">
				<div>
					<h2 id="saved-builds-title" class="text-lg font-semibold text-text-primary">Saved builds</h2>
					<p class="mt-0.5 text-xs text-text-tertiary">Up to 5 recent versions per table</p>
				</div>
				<button
					type="button"
					class="rounded-md border border-border p-1.5 text-text-secondary hover:bg-bg-tertiary"
					aria-label="Close"
					onclick={() => (historyModalOpen = false)}
				>
					<X class="h-4 w-4" />
				</button>
			</div>
			<div class="publish-panel-scroll min-h-0 flex-1 overflow-y-auto p-3">
				{#if buildHistory.length === 0}
					<p class="px-2 py-8 text-center text-sm text-text-tertiary">
						No saved builds yet. Use Quick rebuild or the AI assistant to generate one.
					</p>
				{:else}
					<ul class="space-y-2">
						{#each buildHistory as item (item.id)}
							<li
								class="rounded-lg border px-3 py-2.5 {activeBuildId === item.id
									? 'border-accent-primary/40 bg-accent-primary/5'
									: 'border-border bg-bg-primary'}"
							>
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<p class="truncate text-sm font-medium text-text-primary">{item.label}</p>
										<p class="mt-0.5 text-xs text-text-tertiary">
											{sourceLabel(item.source)} · {formatBuildTime(item.created_at)}
										</p>
									</div>
									<button
										type="button"
										class="inline-flex shrink-0 items-center gap-1 rounded-md border border-border px-2.5 py-1 text-xs text-text-secondary hover:bg-bg-tertiary disabled:opacity-50"
										disabled={restoreLoadingId === item.id}
										onclick={() => loadBuildSnapshot(item.id)}
									>
										{#if restoreLoadingId === item.id}
											<Loader2 class="h-3.5 w-3.5 animate-spin" />
										{/if}
										{activeBuildId === item.id ? 'Current' : 'Load'}
									</button>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.publish-panel-scroll {
		scrollbar-width: thin;
		scrollbar-color: color-mix(in srgb, var(--color-text-tertiary) 55%, transparent) transparent;
	}

	.publish-panel-scroll::-webkit-scrollbar {
		width: 7px;
		height: 7px;
	}

	.publish-panel-scroll::-webkit-scrollbar-track {
		background: transparent;
	}

	.publish-panel-scroll::-webkit-scrollbar-thumb {
		background: color-mix(in srgb, var(--color-text-tertiary) 45%, transparent);
		border-radius: 999px;
		border: 2px solid transparent;
		background-clip: padding-box;
	}

	.publish-panel-scroll::-webkit-scrollbar-thumb:hover {
		background: color-mix(in srgb, var(--color-accent-primary) 55%, transparent);
		background-clip: padding-box;
	}

	.publish-preview-frame {
		scrollbar-width: thin;
		scrollbar-color: #3d4658 #13171f;
	}

	.publish-preview-frame::-webkit-scrollbar {
		width: 8px;
	}

	.publish-preview-frame::-webkit-scrollbar-track {
		background: #13171f;
	}

	.publish-preview-frame::-webkit-scrollbar-thumb {
		background: #3d4658;
		border-radius: 999px;
		border: 2px solid #13171f;
	}
</style>
