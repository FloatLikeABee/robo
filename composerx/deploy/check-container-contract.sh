#!/bin/sh
# Fails if the Content Maker image contract drifts. No daemon.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
repo=$(CDPATH= cd -- "$root/.." && pwd)
fail() {
  echo "contract: $*" >&2
  exit 1
}

df="$root/Dockerfile"
[ -f "$df" ] || fail "missing Dockerfile"
grep -q 'CGO_ENABLED=0' "$df" || fail "Dockerfile must set CGO_ENABLED=0"
grep -q 'HEALTHCHECK' "$df" || fail "Dockerfile must define HEALTHCHECK"
grep -q 'http://127.0.0.1:${PORT}/health' "$df" || fail "HEALTHCHECK must call http://127.0.0.1:\${PORT}/health"
if grep -F '$$' "$df" >/dev/null; then
  fail "Dockerfile must not contain \$\$; shell-form HEALTHCHECK runs under /bin/sh -c, where \$\$ is the PID"
fi
if grep -E '^[[:space:]]*ARG[[:space:]]+.*(JWT|PASSWORD|SECRET|API_KEY|TOKEN)' "$df" >/dev/null; then
  fail "Dockerfile must not take secrets as ARG"
fi
grep -q 'VITE_API_BASE=' "$df" || fail "Dockerfile must set VITE_API_BASE empty"
if grep -E '^[[:space:]]*COPY[[:space:]].*(\.env|ai\.config\.json)' "$df" >/dev/null; then
  fail "Dockerfile must not copy .env or ai.config.json"
fi

entry="$root/deploy/docker-entrypoint.sh"
[ -f "$entry" ] || fail "missing deploy/docker-entrypoint.sh"
grep -q 'su-exec composerx' "$entry" || fail "entrypoint must exec su-exec composerx"
grep -q '/data' "$entry" || fail "entrypoint must prepare /data"

ignore="$repo/.dockerignore"
[ -f "$ignore" ] || fail "missing .dockerignore"
grep -q 'composerx' "$ignore" || fail ".dockerignore missing composerx"
grep -q 'composerx/backend/storage' "$ignore" || fail ".dockerignore must exclude composerx/backend/storage"
grep -q 'ai.config.json' "$ignore" || fail ".dockerignore missing ai.config.json"
grep -q '\.env' "$ignore" || fail ".dockerignore missing .env"

example="$root/deploy/.env.production.example"
[ -f "$example" ] || fail "missing deploy/.env.production.example"
grep -q '^USERS_PANEL_BASE_URL=$' "$example" || fail "USERS_PANEL_BASE_URL must be empty in the example"
grep -q '^MORPH_AI_API_KEY=$' "$example" || fail "MORPH_AI_API_KEY must be empty in the example"
grep -q '^TRAN_OPENAI_API_KEY=$' "$example" || fail "TRAN_OPENAI_API_KEY must be empty in the example"
if grep -E 'sk-|api_key.*=.+' "$example" >/dev/null; then
  fail "example env contains a usable key"
fi

yaml="$repo/render.yaml"
backend_readme="$root/backend/README.md"
deploy_readme="$repo/deploy/README.md"
[ -f "$yaml" ] || fail "missing render.yaml"
[ -f "$backend_readme" ] || fail "missing backend README"
[ -f "$deploy_readme" ] || fail "missing deploy/README.md"
python3 - "$yaml" "$backend_readme" "$deploy_readme" <<'PY' || fail "render.yaml or runbook does not match the Content Maker Blueprint"
import sys

yaml_path, backend_readme, deploy_readme = sys.argv[1:]
text = open(yaml_path, encoding="utf-8").read()
lines = text.splitlines()
errors = []

def service_blocks(src):
    idxs = [i for i, line in enumerate(src) if line.startswith("  - ")]
    blocks = []
    for n, i in enumerate(idxs):
        j = idxs[n + 1] if n + 1 < len(idxs) else len(src)
        blocks.append(src[i:j])
    return blocks

def service_name(block):
    for line in block:
        if line.startswith("    name:"):
            return line.split(":", 1)[1].strip()
    return ""

if any(line.startswith("projects:") for line in lines):
    errors.append("must not declare projects")
if "numInstances:" in text:
    errors.append("must not set numInstances")
if "generateValue:" in text:
    errors.append("must not generate secret values")
if "VITE_COMPOSERX_URL" in text:
    errors.append("must not set VITE_COMPOSERX_URL")

blocks = service_blocks(lines)
names = [service_name(block) for block in blocks]
if names != ["morph", "morph-utils", "formx", "composerx", "sharpreport", "morph-engi"]:
    errors.append(
        "service names must be morph, morph-utils, formx, composerx, sharpreport, morph-engi, got "
        + ", ".join(names)
    )

svc = next((block for block in blocks if service_name(block) == "composerx"), None)
if svc is None:
    errors.append("missing composerx service")
else:
    body = "\n".join(svc)
    for needle in (
        "type: web",
        "runtime: docker",
        "region: singapore",
        "plan: starter",
        "branch: main",
        "dockerfilePath: ./composerx/Dockerfile",
        "dockerContext: .",
        "healthCheckPath: /health",
        "autoDeployTrigger: checksPass",
        "maxShutdownDelaySeconds: 120",
        "name: composerx-data",
        "mountPath: /data",
        "sizeGB: 1",
    ):
        if needle not in body:
            errors.append("composerx missing " + needle)
    for path in (
        "composerx/**",
        "pkg/**",
        "composerx/Dockerfile",
        ".dockerignore",
        "composerx/deploy/docker-entrypoint.sh",
        "render.yaml",
    ):
        if path not in body:
            errors.append("composerx buildFilter missing " + path)

    env = {}
    current = None
    for line in svc:
        if line.startswith("      - key:"):
            if current:
                env[current[0]] = current[1]
            current = [line.split(":", 1)[1].strip(), []]
        elif current is not None and line.startswith("        "):
            current[1].append(line)
        elif current is not None:
            env[current[0]] = current[1]
            current = None
    if current:
        env[current[0]] = current[1]

    plain = {
        "PORT": "8043",
        "COMPOSERX_PORT": "8043",
        "GIN_MODE": "release",
        "COMPOSERX_SQLITE_PATH": "/data/composerx.sqlite",
        "COMPOSERX_BADGER_PATH": "/data/composerx_badger",
        "TRAN_FILE_STORAGE_PATH": "/data/storage",
    }
    for key, value in plain.items():
        body_env = "\n".join(env.get(key, []))
        if key not in env:
            errors.append("missing env " + key)
        elif 'value: "' + value + '"' not in body_env and "value: " + value not in body_env:
            errors.append(key + " must be " + value)
    for key in ("USERS_PANEL_BASE_URL", "MORPH_AI_API_KEY", "TRAN_QWEN_API_KEY", "TRAN_OPENAI_API_KEY"):
        body_env = "\n".join(env.get(key, []))
        if key not in env:
            errors.append("missing prompt " + key)
            continue
        if "sync: false" not in body_env:
            errors.append(key + " must set sync: false")
        if "value:" in body_env:
            errors.append(key + " must not have a value")
    if "REACT_APP_MORPH_UTILS_URL" in env:
        errors.append("composerx must not set REACT_APP_MORPH_UTILS_URL")

for doc in (backend_readme, deploy_readme):
    doc_text = open(doc, encoding="utf-8").read()
    for needle in (
        "prj-dahc33dbedkc73a1v8n0",
        "https://<morph public host>",
        "https://<composerx public host>",
        "USERS_PANEL_BASE_URL",
        "not a secret",
        "MORPH_AI_API_KEY",
        "TRAN_OPENAI_API_KEY",
        "does not call Render",
        "VITE_COMPOSERX_URL",
        "GET /health",
        "8043",
    ):
        if needle not in doc_text:
            errors.append(doc + " missing " + needle)

if errors:
    print("\n".join(errors), file=sys.stderr)
    sys.exit(1)
PY

echo "composerx container contract ok"
