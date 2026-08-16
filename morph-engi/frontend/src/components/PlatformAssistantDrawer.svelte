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
    'Summarize the active project',
    'List site logs for this project',
    'Who are the people on this project?',
    'Create a contractor for the active project',
    'What files are attached to this project?',
  ]

  const chatEndpoint = `${API_BASE.replace(/\/$/, '')}/api/v1/assistant/chat`
</script>

<PlatformChatDrawer
  {open}
  title="Projects AI"
  {chatEndpoint}
  getHeaders={authHeaders}
  {getStateExtra}
  welcomeMessage="Hi! I'm **Projects AI**. I can help with projects, site logs, files, people, and flow log using your live app data."
  progressContext={{ app: 'morph-engi' }}
  {suggestions}
  on:close={() => (open = false)}
/>
