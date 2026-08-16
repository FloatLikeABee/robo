<script>
  import PlatformChatDrawer from '@robo/platform-chat/svelte'
  import '@robo/platform-chat/chat-drawer.css'

  /** @type {{ open?: boolean, apiBase?: string, getAuthHeaders?: () => Record<string, string> }} */
  let { open = $bindable(false), apiBase = '', getAuthHeaders = () => ({}) } = $props()

  const suggestions = [
    'Create template name: Welcome',
    'List templates',
    'List contacts',
    'Help me write a follow-up email',
    'Show my message threads',
  ]

  const chatEndpoint = $derived(`${apiBase.replace(/\/$/, '')}/ai/assistant/chat`)
</script>

<PlatformChatDrawer
  {open}
  title="TranMail AI"
  {chatEndpoint}
  getHeaders={getAuthHeaders}
  welcomeMessage="Hi! I am your **ComposerX assistant**. I can help with saved content, templates, reference docs, and writing support."
  progressContext={{ app: 'composerx' }}
  {suggestions}
  on:close={() => (open = false)}
/>
