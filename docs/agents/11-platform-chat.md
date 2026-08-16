# 11 — Platform Chat (`@robo/platform-chat`)

## Overview

`platform-chat/` is a shared, framework-agnostic chat drawer UI component used by all platform apps. It provides a floating AI assistant panel with markdown rendering, UI blocks, progress ticker, and theme support.

- **Package:** `@robo/platform-chat` (private, v0.1.0, ESM-only)
- **Location:** `platform-chat/`
- **Consumed by:** formx, composerx, booki, UsersPanel, SharpReport, morph-engi, morph

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  platform-chat ARCHITECTURE                      │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Framework-agnostic core                                  │   │
│  │  usePlatformChat.ts                                       │   │
│  │  ├── createPlatformChatController(options)                │   │
│  │  │   • Message list state (ChatMessage[])                 │   │
│  │  │   • Assistant state (AssistantState)                   │   │
│  │  │   • send() → POST chatEndpoint                         │   │
│  │  │   • abort() → AbortController                          │   │
│  │  │   • submitUiBlock() → send field=value                 │   │
│  │  │   • toggleTheme() → dark/light (localStorage)          │   │
│  │  │   • navigatePromptHistory() → up/down arrow            │   │
│  │  │   • subscribe(listener) + getSnapshot()                │   │
│  │  ├── renderChatMarkdown(content) → marked (GFM + breaks)  │   │
│  │  └── Types: ChatMessage, AssistantState, UiBlock          │   │
│  │                                                           │   │
│  │  aiProgress.ts                                            │   │
│  │  ├── startProgressTicker(context) → status steps          │   │
│  │  └── runAiProgress(context, onStep) → for non-drawer use  │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│              ┌───────────────┴───────────────┐                   │
│              ▼                               ▼                   │
│  ┌───────────────────────┐   ┌───────────────────────┐          │
│  │  React wrapper         │   │  Svelte wrapper        │          │
│  │  PlatformChatDrawer    │   │  PlatformChatDrawer    │          │
│  │  .react.tsx            │   │  .svelte               │          │
│  │                        │   │                        │          │
│  │  Props:                │   │  Props:                │          │
│  │  • open, onClose       │   │  • open (bindable)     │          │
│  │  • title               │   │  • title               │          │
│  │  • chatEndpoint        │   │  • chatEndpoint        │          │
│  │  • getHeaders          │   │  • getHeaders          │          │
│  │  • welcomeMessage      │   │  • welcomeMessage      │          │
│  │  • suggestions         │   │  • suggestions         │          │
│  │  • quickActions        │   │  • quickActions        │          │
│  │  • enableFileAnalyze   │   │  • enableFileAnalyze   │          │
│  │  • progressContext     │   │  • progressContext     │          │
│  │  • initialMessages     │   │  • getStateExtra       │          │
│  │                        │   │  • attachments         │          │
│  │  Uses:                 │   │  • onRemoveAttachment  │          │
│  │  useSyncExternalStore  │   │                        │          │
│  └───────────────────────┘   └───────────────────────┘          │
│                                                                 │
│  Standalone components:                                         │
│  • MessageBody.react.tsx / MessageBody.svelte                   │
│    — Long message preview + "Read full message" modal           │
│    — Not used by drawer; consumed independently (e.g. ComposerX)│
└─────────────────────────────────────────────────────────────────┘
```

## Package exports

```json
{
  "./chat-tokens.css": "./chat-tokens.css",
  "./chat-drawer.css": "./chat-drawer.css",
  "./message-body.css": "./message-body.css",
  "./usePlatformChat": "./usePlatformChat.ts",
  "./aiProgress": "./aiProgress.ts",
  "./react": "./PlatformChatDrawer.react.tsx",
  "./svelte": "./PlatformChatDrawer.svelte",
  "./MessageBody.react": "./MessageBody.react.tsx",
  "./MessageBody.svelte": "./MessageBody.svelte"
}
```

## How apps consume it

### 1. Link in package.json

```json
{
  "dependencies": {
    "@robo/platform-chat": "file:../../platform-chat"
  }
}
```

### 2. React integration

```tsx
import PlatformChatDrawer from '@robo/platform-chat/react';
import '@robo/platform-chat/chat-drawer.css';
import '@robo/platform-chat/chat-tokens.css';

function App() {
  const [open, setOpen] = useState(false);

  return (
    <PlatformChatDrawer
      open={open}
      onClose={() => setOpen(false)}
      title="AI Assistant"
      chatEndpoint="/api/v1/assistant/chat"
      getHeaders={() => ({
        Authorization: `Bearer ${getToken()}`,
      })}
      welcomeMessage="Hello! How can I help you?"
      suggestions={["Create a form", "List events"]}
      progressContext={{ app: 'booki' }}
    />
  );
}
```

### 3. Svelte integration

```svelte
<script>
  import PlatformChatDrawer from '@robo/platform-chat/svelte';
  import '@robo/platform-chat/chat-drawer.css';
  import '@robo/platform-chat/chat-tokens.css';

  let open = $state(false);
  let chatEndpoint = $derived(`${apiBase}/assistant/chat`);
</script>

<PlatformChatDrawer
  bind:open
  {chatEndpoint}
  getHeaders={() => ({ Authorization: `Bearer ${token}` })}
  title="AI Assistant"
  progressContext={{ app: 'composerx' }}
/>
```

### 4. Using aiProgress standalone

```ts
import { runAiProgress } from '@robo/platform-chat/aiProgress';

// For non-drawer progress indication (e.g., in a publish panel)
runAiProgress(
  { app: 'composerx', userText: 'Generate report' },
  (step) => { statusText = step; }
);
```

## Chat API contract

The controller POSTs to `chatEndpoint`:

```json
// Request
{
  "messages": [
    { "role": "user", "content": "Create a form" }
  ],
  "state": {
    "intent": "",
    "fields": {}
  }
}

// Response
{
  "assistant_message": "I can help with that. What should the form be called?",
  "state": {
    "intent": "create_form",
    "fields": { "name": "" }
  },
  "ui_blocks": []
}
```

This matches `AI_ASSISTANT_MORPHAI_CONTRACT.md`.

## UI blocks

When the backend returns `ui_blocks`, the drawer renders interactive widgets:

```json
{
  "ui_blocks": [
    {
      "type": "mcp_app",
      "widget": "select",
      "id": "gender",
      "label": "Gender",
      "options": [
        { "value": "female", "label": "Female" },
        { "value": "male", "label": "Male" }
      ],
      "submit_as": { "field": "gender" }
    }
  ]
}
```

The user's selection is submitted as `survey_bot_answer:gender=female`.

## Dependencies

- **`marked`** (^15.0.7) — Markdown parsing (GFM + breaks)
- **`xlsx`** (^0.18.5) — Excel file parsing for file-analyze feature
- **Peer:** `react >=18`, `svelte >=4`

## Per-app wrappers

Each app creates a thin wrapper component:

| App | Wrapper file | Framework |
|-----|-------------|-----------|
| Booki | `PlatformAssistantDrawer.tsx` | React |
| ComposerX | `PlatformAssistantDrawer.svelte` | Svelte |
| UsersPanel | `PlatformAssistantDrawer.svelte` | Svelte |
| Morph | `PlatformAssistantDrawer.svelte` | Svelte |

Wrappers typically add: app-specific `progressContext`, auth headers via `getHeaders`, and app-specific `suggestions`.