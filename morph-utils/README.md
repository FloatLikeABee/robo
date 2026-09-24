# MorphUtils

Integrated shell for **Event Logs**, **Content Maker**, **Data Access**, and **Project**. Module frontends stay in their own folders; MorphUtils embeds them in one dark Morph-themed workspace.

## Dev

```bash
cd morph-utils/frontend
npm install
npm run dev
```

Open http://localhost:3040

Prefer the repo launcher: `./start-all.sh start morph-utils` starts the MorphUtils shell plus Data Access UI and API (`sharpreport-ui`, `sharpreport-api`).

Event Logs, Content Maker, and Project UIs must still be running on their default ports (or set `VITE_*_URL` in the **repo-root** `.env`). If Data Access is down, MorphUtils shows a start hint instead of a refused iframe.

Sign in with Morph auth so the shared session cookie is present. Chat is **Morph AI**, not a drawer in this shell.

## Environment

| Variable | Default |
|----------|---------|
| `VITE_SHEETX_URL` | `http://localhost:19909` (falls back to `VITE_FORMSX_URL`) |
| `VITE_FORMSX_URL` | `http://localhost:19909` (legacy alias) |
| `VITE_COMPOSERX_URL` | `http://localhost:8044` |
| `VITE_DATAX_URL` | `http://localhost:5178` |
| `VITE_PROJECTS_URL` | `http://localhost:5179` (falls back to `VITE_MORPH_ENGI_URL`) |
| `VITE_MORPH_ENGI_URL` | `http://localhost:5179` (legacy alias) |
| `VITE_MORPH_AI_URL` | `http://localhost:3031` |

Embed ids stay `sheetx`, `composerx`, `datax`, `projects`. Event Logs iframe defaults to `/events-info`.
