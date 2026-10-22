#!/bin/sh
# Assembles DorphUtils + embedded module frontends for one Netlify site.
# ponytail: one origin — module APIs must be set via Netlify env (see netlify.toml);
# upgrade path is separate Render services per render.yaml.
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="$ROOT/netlify-dist"
rm -rf "$DIST"
mkdir -p "$DIST"

SITE="${DEPLOY_PRIME_URL:-${URL:-http://127.0.0.1:8888}}"
SITE="${SITE%/}"

DEMO_USER="${NETLIFY_DEMO_ADMIN_USERNAME:-dorphadmin}"
DEMO_PASS="${NETLIFY_DEMO_ADMIN_PASSWORD:-admin123}"
export VITE_NETLIFY_LOCAL_DEMO="${VITE_NETLIFY_LOCAL_DEMO:-1}"
export VITE_DEMO_ADMIN_USERNAME="$DEMO_USER"
export VITE_DEMO_ADMIN_PASSWORD="$DEMO_PASS"

EMBED_SHEETX="$SITE/embed/sheetx"
EMBED_COMPOSERX="$SITE/embed/composerx"
EMBED_PROJECTS="$SITE/embed/projects"
EMBED_DORPH="$SITE/embed/dorph"

export VITE_SHEETX_URL="$EMBED_SHEETX"
export VITE_COMPOSERX_URL="$EMBED_COMPOSERX"
export VITE_PROJECTS_URL="$EMBED_PROJECTS"
export VITE_DORPH_ENGI_URL="$EMBED_PROJECTS"
export VITE_DORPH_AI_URL="$EMBED_DORPH"
# Same-origin /api when DORPH_API_ORIGIN is proxied in _redirects (leave unset in static-only demos).
export VITE_DORPH_API_URL="${VITE_DORPH_API_URL:-}"
export VITE_USERS_PANEL_API_URL="${VITE_USERS_PANEL_API_URL:-}"

if [ -n "${VITE_DATAX_URL:-}" ]; then
  export VITE_DATAX_URL
fi

echo "Netlify bundle site URL: $SITE"

# --- Module embeds (order: dependencies first) ---
(
  cd "$ROOT/dorph-engi/frontend"
  npm ci
  VITE_STORAGE=local npm run build -- --base=/embed/projects/
)
mkdir -p "$DIST/embed/projects"
cp -R "$ROOT/dorph-engi/frontend/dist/." "$DIST/embed/projects/"

(
  cd "$ROOT/formx/frontend"
  npm ci
  VITE_API_URL="${FORMX_API_URL:-}" npm run build -- --base=/embed/sheetx/
)
mkdir -p "$DIST/embed/sheetx"
cp -R "$ROOT/formx/frontend/dist/." "$DIST/embed/sheetx/"

(
  cd "$ROOT/composerx/frontend"
  npm ci
  VITE_API_BASE="${COMPOSERX_API_URL:-}" npm run build -- --base=/embed/composerx/
)
mkdir -p "$DIST/embed/composerx"
cp -R "$ROOT/composerx/frontend/dist/." "$DIST/embed/composerx/"

(
  cd "$ROOT/dorph/frontend"
  npm ci
  # CRA treats CI=true as "fail on lint warnings"; Netlify sets CI globally.
  REACT_APP_NETLIFY_LOCAL_DEMO=true \
  REACT_APP_DEMO_ADMIN_USERNAME="$DEMO_USER" \
  REACT_APP_DEMO_ADMIN_PASSWORD="$DEMO_PASS" \
  PUBLIC_URL=/embed/dorph CI=false npm run build
)
mkdir -p "$DIST/embed/dorph"
cp -R "$ROOT/dorph/frontend/build/." "$DIST/embed/dorph/"

# --- DorphUtils shell at site root ---
(
  cd "$ROOT/dorph-utils/frontend"
  npm ci
  npm run build
)
cp -R "$ROOT/dorph-utils/frontend/dist/." "$DIST/"

# --- Redirects: optional Dorph API proxy + SPA fallbacks ---
REDIRECTS="$DIST/_redirects"
: >"$REDIRECTS"
if [ -n "${DORPH_API_ORIGIN:-}" ]; then
  ORIGIN="${DORPH_API_ORIGIN%/}"
  printf '%s\n' "/api/*  ${ORIGIN}/api/:splat  200!" >>"$REDIRECTS"
fi
printf '%s\n' \
  "/embed/projects/*  /embed/projects/index.html  200" \
  "/embed/sheetx/*  /embed/sheetx/index.html  200" \
  "/embed/composerx/*  /embed/composerx/index.html  200" \
  "/embed/dorph/*  /embed/dorph/index.html  200" \
  "/*  /index.html  200" >>"$REDIRECTS"

echo "Wrote $DIST ($(du -sh "$DIST" | cut -f1))"
