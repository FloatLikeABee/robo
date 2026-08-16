<script lang="ts">
	import { onMount } from 'svelte';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import {
		DOCS_CONTENT_ACCEPT,
		createDoc,
		deleteDoc,
		isDataTableFile,
		listDocs,
		uploadDoc,
		type DocsDocument
	} from '$lib/docs';

	let docs = $state<DocsDocument[]>([]);
	let loading = $state(true);
	let error = $state('');
	let title = $state('');
	let content = $state('');
	let busy = $state(false);
	let selected = $state<DocsDocument | null>(null);

	async function refresh() {
		loading = true;
		error = '';
		try {
			docs = await listDocs();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load documents';
			docs = [];
		} finally {
			loading = false;
		}
	}

	onMount(() => whenSessionReady(() => void refresh()));

	async function onCreate() {
		if (!title.trim() || !content.trim()) {
			error = 'Title and content are required (or upload a PDF/TXT file).';
			return;
		}
		busy = true;
		error = '';
		try {
			await createDoc({ title: title.trim(), content: content.trim() });
			title = '';
			content = '';
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Create failed';
		} finally {
			busy = false;
		}
	}

	async function onUpload(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		if (isDataTableFile(file.name) && !file.name.toLowerCase().endsWith('.txt')) {
			error = 'CSV/JSON belong in Data tables. Upload PDF or TXT content documents here.';
			return;
		}
		busy = true;
		error = '';
		try {
			await uploadDoc(file, title.trim() || undefined);
			title = '';
			await refresh();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Upload failed';
		} finally {
			busy = false;
		}
	}

	async function onDelete(doc: DocsDocument) {
		if (!confirm(`Delete "${doc.title || doc.id}"?`)) return;
		try {
			await deleteDoc(doc.id);
			if (selected?.id === doc.id) selected = null;
			await refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Delete failed';
		}
	}
</script>

<div class="grid min-h-0 flex-1 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)]">
	<section class="flex min-h-0 flex-col gap-3 overflow-hidden rounded-xl border border-border bg-bg-elevated p-4">
		<div class="flex items-center justify-between gap-2">
			<h2 class="font-semibold text-text-primary">Library</h2>
			<label class="inline-flex cursor-pointer items-center gap-1.5 rounded-lg border border-border px-2.5 py-1 text-xs text-text-secondary hover:bg-bg-tertiary">
				Upload PDF/TXT
				<input type="file" class="hidden" accept={DOCS_CONTENT_ACCEPT} onchange={onUpload} disabled={busy} />
			</label>
		</div>
		<p class="text-xs text-text-tertiary">Content documents only — not CSV/JSON data files.</p>
		{#if error}<p class="text-sm text-danger">{error}</p>{/if}
		{#if loading}
			<p class="text-sm text-text-secondary">Loading…</p>
		{:else if docs.length === 0}
			<p class="text-sm text-text-secondary">No documents yet.</p>
		{:else}
			<ul class="min-h-0 flex-1 space-y-1 overflow-auto">
				{#each docs as doc (doc.id)}
					<li>
						<button
							type="button"
							class="flex w-full items-start justify-between gap-2 rounded-lg px-2 py-2 text-left text-sm hover:bg-bg-tertiary {selected?.id === doc.id
								? 'bg-accent-primary/10'
								: ''}"
							onclick={() => (selected = doc)}
						>
							<span class="min-w-0">
								<span class="block truncate font-medium text-text-primary">{doc.title || 'Untitled'}</span>
								<span class="text-xs text-text-tertiary">{doc.type || 'document'}</span>
							</span>
							<button
								type="button"
								class="shrink-0 text-xs text-danger"
								onclick={(ev) => {
									ev.stopPropagation();
									void onDelete(doc);
								}}>Delete</button
							>
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	<section class="flex min-h-0 flex-col gap-3 overflow-auto rounded-xl border border-border bg-bg-elevated p-4">
		{#if selected}
			<h2 class="font-semibold text-text-primary">{selected.title || 'Document'}</h2>
			<pre class="whitespace-pre-wrap text-sm text-text-secondary">{selected.content || '(no text content — binary upload)'}</pre>
		{:else}
			<h2 class="font-semibold text-text-primary">New note</h2>
			<input class="rounded-lg border border-border bg-bg px-3 py-2 text-sm" placeholder="Title" bind:value={title} />
			<textarea
				class="min-h-[12rem] rounded-lg border border-border bg-bg px-3 py-2 text-sm"
				placeholder="Paste or type document content…"
				bind:value={content}
			></textarea>
			<button
				type="button"
				class="self-start rounded-lg bg-accent-primary px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
				disabled={busy}
				onclick={() => void onCreate()}>Save document</button
			>
		{/if}
	</section>
</div>
