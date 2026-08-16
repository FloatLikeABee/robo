<script lang="ts">
	import { onMount } from 'svelte';
	import { whenSessionReady } from '$lib/stores/auth.svelte';
	import {
		DOCS_CONTENT_ACCEPT,
		createSession,
		deleteSession,
		docsAiChat,
		getSession,
		isDataTableFile,
		listSessions,
		patchSession,
		uploadDoc,
		type DocsSession
	} from '$lib/docs';

	type ChatMsg = { role: 'user' | 'assistant'; content: string };

	let sessions = $state<DocsSession[]>([]);
	let activeId = $state<string | null>(null);
	let messages = $state<ChatMsg[]>([]);
	let input = $state('');
	let error = $state('');
	let loading = $state(true);
	let sending = $state(false);
	let pendingDocIds = $state<string[]>([]);
	let attachLabel = $state('');

	function sessionTitle(s: DocsSession): string {
		return s.title?.trim() || `Session ${String(s.id).slice(0, 8)}`;
	}

	async function refreshSessions() {
		sessions = await listSessions();
	}

	async function openSession(id: string) {
		activeId = id;
		error = '';
		const s = await getSession(id);
		const raw = s.messages || [];
		messages = raw.map((m) => {
			const role: 'user' | 'assistant' =
				m.role === 'user' || m.isUser ? 'user' : 'assistant';
			const content = String(m.content || m.text || '');
			return { role, content };
		});
	}

	async function onNew() {
		error = '';
		const s = await createSession();
		await refreshSessions();
		await openSession(s.id);
	}

	async function onDelete(id: string) {
		if (!confirm('Delete this session?')) return;
		await deleteSession(id);
		if (activeId === id) {
			activeId = null;
			messages = [];
		}
		await refreshSessions();
	}

	async function onAttach(e: Event) {
		const inputEl = e.target as HTMLInputElement;
		const file = inputEl.files?.[0];
		inputEl.value = '';
		if (!file) return;
		if (isDataTableFile(file.name) && !file.name.toLowerCase().endsWith('.txt')) {
			error = 'CSV/JSON belong in Data tables. Attach PDF or TXT content here.';
			return;
		}
		try {
			const doc = await uploadDoc(file);
			pendingDocIds = [...pendingDocIds, doc.id];
			attachLabel = `${pendingDocIds.length} doc(s) attached`;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Attach failed';
		}
	}

	async function send() {
		const text = input.trim();
		if (!text || sending) return;
		if (!activeId) await onNew();
		const sid = activeId;
		if (!sid) return;

		sending = true;
		error = '';
		input = '';
		const nextMessages: ChatMsg[] = [...messages, { role: 'user', content: text }];
		messages = nextMessages;
		try {
			const res = await docsAiChat({
				messages: nextMessages.map((m) => ({ role: m.role, content: m.content })),
				doc_ids: pendingDocIds,
				document_mode: true
			});
			const reply = res.response || '(empty response)';
			messages = [...nextMessages, { role: 'assistant', content: reply }];
			pendingDocIds = [];
			attachLabel = '';
			try {
				await patchSession(sid, {
					messages: messages.map((m) => ({
						role: m.role,
						content: m.content,
						text: m.content,
						isUser: m.role === 'user'
					})),
					title: messages.find((m) => m.role === 'user')?.content.slice(0, 48) || 'Chat'
				});
			} catch {
				/* persistence best-effort */
			}
			await refreshSessions();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Chat failed';
		} finally {
			sending = false;
		}
	}

	onMount(() =>
		whenSessionReady(async () => {
			try {
				await refreshSessions();
				if (sessions[0]) await openSession(sessions[0].id);
			} catch (e) {
				error = e instanceof Error ? e.message : 'Failed to load sessions';
			} finally {
				loading = false;
			}
		})
	);
</script>

{#if loading}
	<p class="text-sm text-text-secondary">Loading Docs AI…</p>
{:else}
	<div class="grid min-h-0 flex-1 gap-3 overflow-hidden lg:grid-cols-[240px_minmax(0,1fr)]">
		<aside class="flex min-h-0 flex-col gap-2 overflow-hidden rounded-xl border border-border bg-bg-elevated p-3">
			<div class="flex items-center justify-between gap-2">
				<h2 class="text-sm font-semibold text-text-primary">Sessions</h2>
				<button
					type="button"
					class="rounded-lg bg-accent-primary px-2 py-1 text-xs font-medium text-white"
					onclick={() => void onNew()}>New</button
				>
			</div>
			<ul class="min-h-0 flex-1 space-y-1 overflow-auto">
				{#each sessions as s (s.id)}
					<li class="group flex items-center gap-1">
						<button
							type="button"
							class="min-w-0 flex-1 truncate rounded-lg px-2 py-1.5 text-left text-xs {activeId === s.id
								? 'bg-accent-primary/15 text-accent-primary'
								: 'text-text-secondary hover:bg-bg-tertiary'}"
							onclick={() => void openSession(s.id)}>{sessionTitle(s)}</button
						>
						<button
							type="button"
							class="hidden text-xs text-danger group-hover:inline"
							onclick={() => void onDelete(s.id)}>×</button
						>
					</li>
				{:else}
					<li class="px-1 text-xs text-text-tertiary">No sessions yet.</li>
				{/each}
			</ul>
		</aside>

		<section class="flex min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-bg-elevated">
			<div class="min-h-0 flex-1 space-y-3 overflow-auto p-4">
				{#if error}<p class="text-sm text-danger">{error}</p>{/if}
				{#each messages as m, i (i)}
					<div class="rounded-lg px-3 py-2 text-sm {m.role === 'user' ? 'bg-accent-primary/10 ml-8' : 'bg-bg-tertiary mr-8'}">
						<div class="mb-1 text-[10px] uppercase tracking-wide text-text-tertiary">{m.role}</div>
						<div class="whitespace-pre-wrap text-text-primary">{m.content}</div>
					</div>
				{:else}
					<p class="text-sm text-text-secondary">Ask about your PDF/TXT documents. Session history is on the left.</p>
				{/each}
			</div>
			<form
				class="shrink-0 border-t border-border p-3"
				onsubmit={(e) => {
					e.preventDefault();
					void send();
				}}
			>
				{#if attachLabel}
					<p class="mb-1 text-xs text-accent-primary">{attachLabel}</p>
				{/if}
				<div class="flex flex-wrap items-end gap-2">
					<label class="inline-flex cursor-pointer items-center rounded-lg border border-border px-2 py-2 text-xs text-text-secondary hover:bg-bg-tertiary">
						📎 PDF/TXT
						<input type="file" class="hidden" accept={DOCS_CONTENT_ACCEPT} onchange={onAttach} />
					</label>
					<textarea
						class="min-h-[2.5rem] min-w-0 flex-1 rounded-lg border border-border bg-bg px-3 py-2 text-sm"
						placeholder="Message Docs AI…"
						bind:value={input}
						rows="2"
					></textarea>
					<button
						type="submit"
						class="rounded-lg bg-accent-primary px-3 py-2 text-sm font-medium text-white disabled:opacity-50"
						disabled={sending || !input.trim()}>{sending ? '…' : 'Send'}</button
					>
				</div>
			</form>
		</section>
	</div>
{/if}
