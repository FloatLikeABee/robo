#!/bin/sh
# Fails if the MorphUtils image contract drifts. No daemon.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
fail() {
  echo "contract: $*" >&2
  exit 1
}

df="$root/Dockerfile"
[ -f "$df" ] || fail "missing Dockerfile"
grep -q 'HEALTHCHECK' "$df" || fail "Dockerfile must define HEALTHCHECK"
grep -q '/health' "$df" || fail "HEALTHCHECK must call /health"
grep -q 'curl' "$df" || fail "HEALTHCHECK must use curl"
if grep -F '$$' "$df" >/dev/null; then
  fail "Dockerfile must not contain \$\$; shell-form HEALTHCHECK runs under /bin/sh -c, where \$\$ is the PID"
fi
if grep -E '^[[:space:]]*ARG[[:space:]]+.*(JWT|PASSWORD|SECRET|API_KEY|TOKEN)' "$df" >/dev/null; then
  fail "Dockerfile must not take secrets as ARG"
fi

entry="$root/deploy/docker-entrypoint.sh"
[ -f "$entry" ] || fail "missing deploy/docker-entrypoint.sh"
grep -q 'config.js' "$entry" || fail "entrypoint must write config.js"
grep -q '\${PORT}' "$entry" || fail "entrypoint must substitute only \${PORT}"

template="$root/deploy/nginx.conf.template"
[ -f "$template" ] || fail "missing nginx template"
grep -q '/health' "$template" || fail "nginx template must serve /health"
grep -q '/ready' "$template" || fail "nginx template must serve /ready"
grep -q 'config.js' "$template" || fail "nginx template must serve config.js"

ignore="$root/.dockerignore"
[ -f "$ignore" ] || fail "missing .dockerignore"
grep -q 'node_modules' "$ignore" || fail ".dockerignore missing node_modules"
grep -q '\.env' "$ignore" || fail ".dockerignore missing .env"

readme="$root/README.md"
[ -f "$readme" ] || fail "missing morph-utils README"
grep -q 'docker build' "$readme" || fail "README must show docker build"
grep -q 'VITE_MORPH_API_URL' "$readme" || fail "README must document VITE_MORPH_API_URL"

names="VITE_MORPH_API_URL VITE_USERS_PANEL_API_URL VITE_SHEETX_URL VITE_FORMSX_URL VITE_COMPOSERX_URL VITE_DATAX_URL VITE_PROJECTS_URL VITE_MORPH_ENGI_URL VITE_MORPH_AI_URL"
for name in $names; do
  grep -q "$name" "$df" || fail "Dockerfile missing $name"
  grep -q "$name" "$entry" || fail "entrypoint missing $name"
  grep -q "$name" "$readme" || fail "README missing $name"
done

echo "morph-utils container contract ok"
