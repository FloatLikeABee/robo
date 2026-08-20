# 09 — Morph Utils (Unified Shell)

## Overview

Morph Utils is a **frontend-only** React shell that embeds all platform apps in iframes with a shared sidebar and single sign-on. It has no backend.

- **Frontend:** `morph-utils/frontend/` — React 19 + Vite 8 + TypeScript 6.0
- **Port:** `3040`
- **Auth:** Shared UsersPanel JWT cookie

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    morph-utils SHELL                             │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Sidebar                                                  │   │
│  │  ┌──────────┐                                             │   │
│  │  │ SurveyX  │──▶ iframe: localhost:19909/survey-bot       │   │
│  │  │ ComposerX│──▶ iframe: localhost:8044                   │   │
│  │  │ DataX    │──▶ iframe: localhost:5178                   │   │
│  │  │ Project  │──▶ iframe: localhost:5179                   │   │
│  │  │ More     │──▶ Links to Morph AI, Morph Data            │   │
│  │  │ Settings │──▶ Shell-level settings page                │   │
│  │  └──────────┘                                             │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Auth flow:                                                     │
│  1. Read userspanel_session_token cookie                        │
│  2. Validate against UsersPanel /api/auth/user                  │
│  3. Pass token to iframes via ?userspanel_token= URL param      │
│  4. Each embedded app consumes token for SSO                    │
└─────────────────────────────────────────────────────────────────┘
```

## Key files

| File | Purpose |
|------|---------|
| `morph-utils/frontend/src/App.tsx` | Main app: sidebar, iframe container, routing |
| `morph-utils/frontend/src/config.ts` | Module registry: IDs, labels, default URLs |
| `morph-utils/frontend/src/auth.ts` | JWT cookie management, login, logout, session validation |

## Module registry (`config.ts`)

```typescript
const modules = [
  { id: 'sheetx',    label: 'Survey Maker', url: 'http://localhost:19909/survey-bot' },
  { id: 'composerx', label: 'Content Maker', url: 'http://localhost:8044' },
  { id: 'datax',     label: 'Data Access', url: 'http://localhost:5178' },
  { id: 'projects',  label: 'Project',  url: 'http://localhost:5179' },
];
```

Module URLs can be overridden via env vars: `VITE_SHEETX_URL`, `VITE_COMPOSERX_URL`, etc.

## Auth flow

1. On load, read `userspanel_session_token` from cookie
2. Validate by calling UsersPanel `GET /api/auth/user`
3. If invalid, show login page (calls UsersPanel `POST /api/auth/login`)
4. On successful login, set cookie and reload
5. Each iframe receives token via `?userspanel_token=` URL parameter
6. Embedded apps consume the token on load for SSO

## How to add a new embedded module

1. Add entry in `morph-utils/frontend/src/config.ts`:
   ```typescript
   { id: 'newapp', label: 'New App', url: 'http://localhost:XXXX' }
   ```
2. Add env var override in `.env`:
   ```
   VITE_NEWAPP_URL=http://localhost:XXXX
   ```
3. The sidebar auto-renders all modules from the registry
4. Ensure the embedded app handles the `?userspanel_token=` URL parameter for SSO

## How to change module branding

1. Update `label` in `config.ts`
2. Update the embedded app's internal nav/labels to match
3. If changing URL paths, update both `config.ts` and the embedded app's routing

## Key env vars

| Variable | Default | Purpose |
|----------|---------|---------|
| `VITE_SHEETX_URL` | `http://localhost:19909/survey-bot` | Survey Maker iframe URL |
| `VITE_COMPOSERX_URL` | `http://localhost:8044` | Content Maker iframe URL |
| `VITE_DATAX_URL` | `http://localhost:5178` | Data Access iframe URL |
| `VITE_PROJECTS_URL` | `http://localhost:5179` | Project iframe URL |
| `VITE_USERS_PANEL_URL` | `http://127.0.0.1:5001` | UsersPanel API URL |