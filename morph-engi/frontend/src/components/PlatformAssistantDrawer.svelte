<script lang="ts">
  import PlatformChatDrawer from '@robo/platform-chat/svelte'
  import '@robo/platform-chat/chat-drawer.css'
  import { api, authHeaders } from '../lib/api'
  import type { AssistantState } from '@robo/platform-chat/usePlatformChat'

  let {
    open = $bindable(false),
    getStateExtra,
  } = $props<{
    open?: boolean
    getStateExtra?: () => AssistantState
  }>()

  const suggestions = [
    'List my project documents',
    'Summarize the selected project',
    'What files are in the library?',
    'Which projects are published?',
  ]

  async function sendChat(payload: { messages: { role: string; content: string }[]; state: unknown }) {
    return api<{ assistant_message?: string; error?: string }>('/api/v1/assistant/chat', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  }
</script>

<PlatformChatDrawer
  {open}
  title="Projects AI"
  {sendChat}
  getHeaders={authHeaders}
  {getStateExtra}
  welcomeMessage="Hi! I'm **Projects AI**. I can help with project documents and the files library using your live app data."
  progressContext={{ app: 'morph-engi' }}
  {suggestions}
  on:close={() => (open = false)}
/>
