<script lang="ts">
  import type { AssistantState } from './usePlatformChat'

  type ChatMsg = { role: string; content: string }

  let {
    open = $bindable(false),
    title = 'AI Assistant',
    chatEndpoint = '',
    getHeaders,
    getStateExtra,
    sendChat,
    welcomeMessage = '',
    suggestions = [],
  } = $props<{
    open?: boolean
    title?: string
    chatEndpoint?: string
    getHeaders?: () => Record<string, string>
    getStateExtra?: () => AssistantState | { fields?: unknown }
    sendChat?: (payload: { messages: ChatMsg[]; state: unknown }) => Promise<{ assistant_message?: string; error?: string }>
    welcomeMessage?: string
    progressContext?: unknown
    suggestions?: string[]
  }>()

  let input = $state('')
  let busy = $state(false)
  let messages = $state<ChatMsg[]>([])
  let error = $state('')

  async function send(text: string) {
    const content = text.trim()
    if (!content || busy) return
    busy = true
    error = ''
    const next = [...messages, { role: 'user', content }]
    messages = next
    input = ''
    try {
      const extra = getStateExtra?.() ?? {}
      const payload = {
        messages: next.map((m) => ({ role: m.role, content: m.content })),
        state: extra,
      }
      let data: { assistant_message?: string; error?: string }
      if (sendChat) {
        data = await sendChat(payload)
      } else {
        const res = await fetch(chatEndpoint, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            ...(getHeaders?.() ?? {}),
          },
          body: JSON.stringify(payload),
        })
        data = (await res.json().catch(() => ({}))) as { assistant_message?: string; error?: string }
        if (!res.ok) throw new Error(data.error || res.statusText)
      }
      const reply = String(data.assistant_message || '').trim()
      if (reply) messages = [...next, { role: 'assistant', content: reply }]
    } catch (e) {
      error = e instanceof Error ? e.message : 'Chat failed'
    } finally {
      busy = false
    }
  }
</script>

{#if open}
  <aside class="platform-chat-drawer">
    <header>
      <strong>{title}</strong>
      <button type="button" onclick={() => (open = false)}>Close</button>
    </header>
    <div class="msgs">
      {#if welcomeMessage && messages.length === 0}
        <div class="bubble assistant">{welcomeMessage}</div>
      {/if}
      {#each messages as m}
        <div class="bubble {m.role}">{m.content}</div>
      {/each}
      {#if error}<div class="bubble assistant">{error}</div>{/if}
    </div>
    <footer>
      {#if suggestions.length}
        <div class="chips">
          {#each suggestions as s}
            <button type="button" disabled={busy} onclick={() => void send(s)}>{s}</button>
          {/each}
        </div>
      {/if}
      <textarea bind:value={input} placeholder="Ask about your projects…" disabled={busy}></textarea>
      <button type="button" style="margin-top:8px" disabled={busy} onclick={() => void send(input)}>
        {busy ? 'Sending…' : 'Send'}
      </button>
    </footer>
  </aside>
{/if}
