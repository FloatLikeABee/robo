<script lang="ts">
  import { get } from 'svelte/store'
  import PlatformChatDrawer from '@robo/platform-chat/svelte'
  import '@robo/platform-chat/chat-drawer.css'
  import { apiOrigin, apiFetch } from '../lib/api'
  import { token } from '../lib/auth'

  const USERS_RELOAD_EVENT = 'users-panel:reload-users'
  const ROLES_RELOAD_EVENT = 'users-panel:reload-roles'

  const SUGGESTIONS = [
    'List roles',
    'List users',
    'Create role role_name: Reviewer description: Read-only access',
    'Create user email: jane@example.com username: jane password: ChangeMe123! last_name: Doe roles: Employee',
  ]

  /** @type {{ open?: boolean }} */
  let { open = $bindable(false) } = $props()

  const chatEndpoint = `${apiOrigin()}/api/assistant/chat`

  function getHeaders() {
    const headers: Record<string, string> = {}
    const t = get(token)
    if (t) headers.Authorization = `Bearer ${t}`
    return headers
  }

  /** Dispatch list reloads after successful create intents. */
  async function assistantFetch(input: RequestInfo | URL, init?: RequestInit) {
    const res = await apiFetch(String(input), init)
    if (res.ok) {
      try {
        const data = (await res.clone().json()) as {
          completed?: boolean
          intent?: string
        }
        if (data.completed && data.intent === 'create_user') {
          window.dispatchEvent(new CustomEvent(USERS_RELOAD_EVENT))
        }
        if (data.completed && data.intent === 'create_role') {
          window.dispatchEvent(new CustomEvent(ROLES_RELOAD_EVENT))
        }
      } catch {
        /* ignore */
      }
    }
    return res
  }
</script>

<PlatformChatDrawer
  {open}
  title="UsersPanel AI"
  {chatEndpoint}
  {getHeaders}
  fetchImpl={assistantFetch}
  welcomeMessage="Hi! I can help with **users**, **roles**, **permissions**, and general questions. Say what you want to do."
  suggestions={SUGGESTIONS}
  progressContext={{ app: 'generic' }}
  on:close={() => (open = false)}
/>
