#!/bin/sh
# Fails if the Morph image contract drifts: static build, non-root server,
# /data volume, env-only secrets, and the five required CI check names.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
fail() {
  echo "contract: $*" >&2
  exit 1
}

df="$root/Dockerfile"
[ -f "$df" ] || fail "missing Dockerfile"
grep -q 'CGO_ENABLED=0' "$df" || fail "Dockerfile must set CGO_ENABLED=0"
grep -q 'HEALTHCHECK' "$df" || fail "Dockerfile must define HEALTHCHECK"
grep -q '/health' "$df" || fail "HEALTHCHECK must call /health"
entry="$root/deploy/docker-entrypoint.sh"
[ -f "$entry" ] || fail "missing deploy/docker-entrypoint.sh"
grep -q 'su-exec morph' "$entry" || fail "entrypoint must exec su-exec morph"
grep -q 'docker-entrypoint.sh' "$df" || fail "Dockerfile must use deploy/docker-entrypoint.sh"
if grep -E '^[[:space:]]*USER[[:space:]]' "$df" >/dev/null; then
  fail "Dockerfile must not set USER; the entrypoint starts as root only to chown /data"
fi
if grep -E '^[[:space:]]*ARG[[:space:]]+(JWT_SECRET|ADMIN_PASSWORD|MORPH_AI_API_KEY|GEMINI_API_KEY|BOOTSTRAP_ADMIN_PASSWORD)\b' "$df" >/dev/null; then
  fail "Dockerfile must not take secrets as ARG"
fi

ignore="$root/.dockerignore"
[ -f "$ignore" ] || fail "missing .dockerignore"
for needle in node_modules .env data formx composerx SharpReport morph-engi bk morph-utils invite-signup platform-chat; do
  grep -q "$needle" "$ignore" || fail ".dockerignore missing $needle"
done
if grep -E '^pkg/?$' "$ignore" >/dev/null || grep -E '^morph/?$' "$ignore" >/dev/null; then
  fail ".dockerignore must keep morph/ and pkg/ in the build context"
fi

compose="$root/deploy/docker-compose.yml"
[ -f "$compose" ] || fail "missing deploy/docker-compose.yml"
grep -q '/data' "$compose" || fail "compose must mount /data"
grep -q 'morph-data' "$compose" || fail "compose must name the data volume"
grep -q 'env_file' "$compose" || fail "compose must use env_file"
grep -q '.env.production' "$compose" || fail "env_file must point at .env.production"
grep -q 'tls' "$compose" || fail "Caddy must be on the tls profile"
grep -q 'profiles' "$compose" || fail "compose must use a profile for Caddy"
[ -f "$root/deploy/Caddyfile" ] || fail "missing deploy/Caddyfile"

example="$root/deploy/.env.production.example"
[ -f "$example" ] || fail "missing deploy/.env.production.example"
grep -q '^JWT_SECRET=$' "$example" || fail "JWT_SECRET must be empty in the example"
grep -q '^ADMIN_PASSWORD=$' "$example" || fail "ADMIN_PASSWORD must be empty in the example"
if grep -E 'morph-dev-jwt-secret-change-me|^ADMIN_PASSWORD=admin123$|JWT_EXPIRY_HOURS=876000' "$example" >/dev/null; then
  fail "example env contains a production-rejected secret or lifetime"
fi

ci="$root/.github/workflows/ci.yml"
[ -f "$ci" ] || fail "missing ci.yml"
for name in "Morph API" "Event Logs" "Content Maker" "morphai" "Morph frontend"; do
  grep -q "$name" "$ci" || fail "ci.yml missing check name $name"
done
jobs=$(awk '/^jobs:/{p=1; next} p && /^[^ ]/{exit} p && /^  [A-Za-z0-9_-]+:/{print}' "$ci")
expected="  go:
  morph-frontend:"
[ "$jobs" = "$expected" ] || fail "ci.yml jobs changed; image build must not add a required check name
$jobs"

echo "container contract ok"
