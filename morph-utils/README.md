# Morph Utils

Integrated shell for **SheetX**, **ComposerX**, **DataX**, **Booki**, **Projects** (Morph Engi), and **Academi**. The individual app frontends are unchanged; Morph Utils embeds them in a single Morph-themed workspace.

## Dev

```bash
cd morph-utils/frontend
npm install
npm run dev
```

Open http://localhost:3040

Embedded app UIs must be running on their default ports (or set `VITE_*_URL` in `.env`).

Sign in with Morph auth so the shared session cookie is present.

## Environment

| Variable | Default |
|----------|---------|
| `VITE_SHEETX_URL` | `http://localhost:19909` (falls back to `VITE_FORMSX_URL`) |
| `VITE_FORMSX_URL` | `http://localhost:19909` (legacy alias) |
| `VITE_COMPOSERX_URL` | `http://localhost:8044` |
| `VITE_DATAX_URL` | `http://localhost:5178` |
| `VITE_BOOKI_URL` | `http://localhost:5174` |
| `VITE_PROJECTS_URL` | `http://localhost:5179` (falls back to `VITE_MORPH_ENGI_URL`) |
| `VITE_MORPH_ENGI_URL` | `http://localhost:5179` (legacy alias) |
| `VITE_ACADEMI_URL` | `http://localhost:8765` |
| `VITE_MORPH_AI_URL` | `http://localhost:3031` |
