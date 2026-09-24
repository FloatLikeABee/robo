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
if grep -F '$$' "$df" >/dev/null; then
  fail "Dockerfile must not contain \$\$; shell-form HEALTHCHECK runs under /bin/sh -c, where \$\$ is the PID"
fi
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

yaml="$root/render.yaml"
[ -f "$yaml" ] || fail "missing render.yaml"
python3 - "$yaml" <<'PY' || fail "render.yaml does not match the Morph Blueprint"
import sys

text = open(sys.argv[1], encoding="utf-8").read()
lines = text.splitlines()
errors = []

if any(line.startswith("projects:") for line in lines):
    errors.append("must not declare projects")
if "numInstances:" in text:
    errors.append("must not set numInstances")
if "generateValue:" in text:
    errors.append("must not generate secret values")

# Service items are "  - ...". Disk name and a later service must not satisfy Morph.
starts = [i for i, line in enumerate(lines) if line.startswith("  - ")]
services = []
for n, i in enumerate(starts):
    j = starts[n + 1] if n + 1 < len(starts) else len(lines)
    services.append(lines[i:j])

def named(block):
    for line in block:
        if line.startswith("    name:"):
            return line.split(":", 1)[1].strip()
    return ""

morph = next((block for block in services if named(block) == "morph"), None)
if morph is None:
    errors.append("missing morph service")
    morph = []
morph_text = "\n".join(morph) + "\n"

def need(needle):
    if needle not in morph_text:
        errors.append("missing " + needle)

for needle in (
    "type: web",
    "runtime: docker",
    "\n    name: morph\n",
    "region: singapore",
    "plan: starter",
    "branch: main",
    "\n    dockerfilePath: ./Dockerfile\n",
    "\n    dockerContext: .\n",
    "healthCheckPath: /health",
    "autoDeployTrigger: checksPass",
    "name: morph-data",
    "mountPath: /data",
    "sizeGB: 1",
    "maxShutdownDelaySeconds: 120",
):
    need(needle)

for path in (
    "morph/**",
    "pkg/**",
    "Dockerfile",
    ".dockerignore",
    "scripts/with-root-env.cjs",
    "deploy/docker-entrypoint.sh",
    "render.yaml",
):
    if "\n        - " + path + "\n" not in morph_text:
        errors.append("buildFilter missing " + path)

# Env entries sit under the morph service: "      - key:" then "        " fields.
blocks = []
current = None
for line in morph:
    if line.startswith("      - key:"):
        if current:
            blocks.append(current)
        current = [line]
    elif current is not None:
        if line.startswith("        "):
            current.append(line)
        else:
            blocks.append(current)
            current = None
if current:
    blocks.append(current)

parsed = {}
for block in blocks:
    key = block[0].split(":", 1)[1].strip()
    parsed[key] = "\n".join(block[1:])

plain = {
    "MORPH_ENV": "production",
    "PORT": "9090",
    "GIN_MODE": "release",
    "MORPH_AI_PROVIDER": "dashscope",
    "ADMIN_USERNAME": "morphadmin",
    "ADMIN_EMAIL": "morphadmin@local.com",
    "DB_PATH": "/data/badger",
    "TRAN_SQLITE_PATH": "/data/tran.sqlite",
    "ENTITY_DETAILS_BADGER": "/data/entity_details",
    "MORPH_KNOWLEDGE_DIR": "/data/knowledge",
    "TRAN_ENTITY_ATTACHMENT_DIR": "/data/uploads/entity_attachments",
}
for key, value in plain.items():
    body = parsed.get(key)
    if body is None:
        errors.append("missing env " + key)
    elif "value: " + value not in body and 'value: "' + value + '"' not in body:
        errors.append(key + " must be " + value)

secrets = (
    "JWT_SECRET",
    "ADMIN_PASSWORD",
    "BOOTSTRAP_ADMIN_PASSWORD",
    "MORPH_AI_API_KEY",
    "GEMINI_API_KEY",
    "TRAN_QWEN_API_KEY",
    "DASHSCOPE_API_KEY",
    "OPENAI_API_KEY",
    "ANTHROPIC_API_KEY",
    "XAI_API_KEY",
    "OPENROUTER_API_KEY",
    "MISTRAL_API_KEY",
    "GROQ_API_KEY",
    "OPENAI_COMPATIBLE_API_KEY",
    "MORPH_IMAGE_API_KEY",
    "POLLINATIONS_API_KEY",
    "SMTP_PASS",
    "TRAN_MYSQL_DSN",
    "TRAN_MONGO_URI",
    "NEO4J_PASSWORD",
    "TRAN_OPENAI_API_KEY",
)
for key in secrets:
    body = parsed.get(key)
    if body is None:
        errors.append("missing secret " + key)
        continue
    if "sync: false" not in body:
        errors.append(key + " must set sync: false")
    if "value:" in body:
        errors.append(key + " must not have a value")

if errors:
    print("\n".join(errors), file=sys.stderr)
    sys.exit(1)
PY

echo "container contract ok"
