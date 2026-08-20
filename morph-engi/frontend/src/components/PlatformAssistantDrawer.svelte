<script lang="ts">
  import PlatformChatDrawer from '@robo/platform-chat/svelte'
  import '@robo/platform-chat/chat-drawer.css'
  import { API_BASE, authHeaders } from '../lib/api'
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

  const chatEndpoint = `${API_BASE.replace(/\/$/, '')}/api/v1/assistant/chat`
</script>

<PlatformChatDrawer
  {open}
  title="Projects AI"
  {chatEndpoint}
  getHeaders={authHeaders}
  {getStateExtra}
  welcomeMessage="Hi! I'm **Projects AI**. I can help with project documents and the files library using your live app data."
  progressContext={{ app: 'morph-engi' }}
  {suggestions}
  on:close={() => (open = false)}
/>
