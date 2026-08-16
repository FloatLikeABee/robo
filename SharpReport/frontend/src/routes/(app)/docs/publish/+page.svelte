<script lang="ts">
	import { onMount } from 'svelte';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import { getDoc, listDocPublishes, listDocs, publishDoc, type DocsDocument } from '$lib/docs';

	let docs = $state<DocsDocument[]>([]);
	let publishes = $state<
		Array<{ id: string; title: string; slug: string; public_path: string; created_at: string }>
	>([]);
	let docId = $state('');
	let title = $state('');
	let content = $state('');
	let analysisPrompt = $state('');
	let error = $state('');
	let info = $state('');
	let busy = $state(false);
	let loading = $state(true);

	async function refresh() {
		loading = true;
		error = '';
		try {
			docs = await listDocs();
			publishes = await listDocPublishes();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load';
		} finally {
			loading = false;
		}
	}

	onMount(() => whenSessionReady(() => void refresh()));

	async function onPickDoc() {
		if (!docId) return;
		try {
			const d = await getDoc(docId);
			title = d.title || '';
			content = d.content || '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load document';
		}
	}

	async function onPublish() {
		if (!title.trim() || !content.trim()) {
			error = 'Title and document content are required.';
			return;
		}
		busy = true;
		error = '';
		info = '';
		try {
			const res = await publishDoc({
				title: title.trim(),
				content: content.trim(),
				doc_id: docId || undefined,
				analysis_prompt: analysisPrompt.trim() || undefined
			});
			info = `Published: ${res.public_path}${res.has_analysis ? ' (with AI analysis)' : ''}`;
			analysisPrompt = '';
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Publish failed';
		} finally {
			busy = false;
		}
	}
</script>

<div class="grid min-h-0 flex-1 gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]">
	<section class="flex min-h-0 flex-col gap-3 overflow-auto rounded-xl border border-border bg-bg-elevated p-4">
		<h2 class="font-semibold text-text-primary">Publish document HTML</h2>
		<p class="text-xs text-text-tertiary">
			Replaces Docs Board. Publishes a shareable HTML page. Optional analysis prompt runs AI over the document.
		</p>
		{#if error}<p class="text-sm text-danger">{error}</p>{/if}
		{#if info}<p class="text-sm text-accent-primary">{info}</p>{/if}

		<label class="text-xs text-text-secondary">
			From library (optional)
			<select
				class="mt-1 w-full rounded-lg border border-border bg-bg px-3 py-2 text-sm"
				bind:value={docId}
				onchange={() => void onPickDoc()}
			>
				<option value="">— Select document —</option>
				{#each docs as d}
					<option value={d.id}>{d.title || d.id}</option>
				{/each}
			</select>
		</label>

		<input class="rounded-lg border border-border bg-bg px-3 py-2 text-sm" placeholder="Title" bind:value={title} />
		<textarea
			class="min-h-[10rem] rounded-lg border border-border bg-bg px-3 py-2 text-sm"
			placeholder="Document content to display…"
			bind:value={content}
		></textarea>
		<textarea
			class="min-h-[4rem] rounded-lg border border-border bg-bg px-3 py-2 text-sm"
			placeholder="Optional AI analysis prompt (leave empty to publish display only)"
			bind:value={analysisPrompt}
		></textarea>
		<button
			type="button"
			class="self-start rounded-lg bg-accent-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
			disabled={busy || loading}
			onclick={() => void onPublish()}>{busy ? 'Publishing…' : 'Publish HTML'}</button
		>
	</section>

	<section class="flex min-h-0 flex-col gap-2 overflow-auto rounded-xl border border-border bg-bg-elevated p-4">
		<h2 class="font-semibold text-text-primary">Published</h2>
		{#if loading}
			<p class="text-sm text-text-secondary">Loading…</p>
		{:else if publishes.length === 0}
			<p class="text-sm text-text-secondary">No published docs yet.</p>
		{:else}
			<ul class="space-y-2">
				{#each publishes as p}
					<li class="rounded-lg border border-border px-3 py-2 text-sm">
						<div class="font-medium text-text-primary">{p.title}</div>
						<a class="text-xs text-accent-primary underline" href={p.public_path} target="_blank" rel="noopener"
							>{p.public_path}</a
						>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
</div>
