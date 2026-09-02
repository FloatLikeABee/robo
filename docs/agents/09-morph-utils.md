# 09 — MorphUtils

## Overview

**MorphUtils** (no space) is a frontend-only React Vite shell. It embeds remaining modules in iframes with a shared Morph JWT cookie. Port **3040**. No backend.

Keep the **outer** MorphUtils left nav. Module apps use header tabs, not a second left rail that replaces the shell.

## Architecture

```
 MorphUtils :3040
   ├── Event Logs     id sheetx      → VITE_SHEETX_URL/events-info   (:19909)
   ├── Content Maker  id composerx   → VITE_COMPOSERX_URL            (:8044)
   ├── Data Access    id datax       → VITE_DATAX_URL                (:5178)
   └── Project        id projects    → VITE_PROJECTS_URL             (:5179)
```

Registry: `morph-utils/frontend/src/config.ts`. Auth: `morph-utils/frontend/src/auth.ts`.

## Env

| Variable | Default |
|----------|---------|
| `VITE_SHEETX_URL` | `http://localhost:19909` (fallback `VITE_FORMSX_URL`) |
| `VITE_COMPOSERX_URL` | `http://localhost:8044` |
| `VITE_DATAX_URL` | `http://localhost:5178` |
| `VITE_PROJECTS_URL` | `http://localhost:5179` (fallback `VITE_MORPH_ENGI_URL`) |
| `VITE_MORPH_AI_URL` | `http://localhost:3031` |

## SSO

1. Read `userspanel_session_token` (Morph JWT).
2. Validate Morph `GET /api/auth/user`.
3. Pass `?userspanel_token=` into each iframe.

Chat is Morph AI. Do not add a satellite assistant drawer to the shell.

`normalizeModuleId` may map leftover `academi`/`docs` → `datax` and `booki` → `projects`. Those are compatibility aliases, not live products.

See [`morph-utils/README.md`](../../morph-utils/README.md).
