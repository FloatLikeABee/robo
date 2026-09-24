#!/bin/sh
# Fails if the Data Access image contract drifts. No daemon.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
repo=$(CDPATH= cd -- "$root/.." && pwd)
fail() {
  echo "contract: $*" >&2
  exit 1
}

df="$root/Dockerfile"
[ -f "$df" ] || fail "missing Dockerfile"
grep -q 'HEALTHCHECK' "$df" || fail "Dockerfile must define HEALTHCHECK"
grep -q 'http://127.0.0.1:${PORT}/health' "$df" || fail "HEALTHCHECK must call http://127.0.0.1:\${PORT}/health"
grep -q 'wget' "$df" || fail "HEALTHCHECK must use wget"
if grep -F '$$' "$df" >/dev/null; then
  fail "Dockerfile must not contain \$\$; shell-form HEALTHCHECK runs under /bin/sh -c, where \$\$ is the PID"
fi
if grep -E '^[[:space:]]*ARG[[:space:]]+.*(JWT|PASSWORD|SECRET|API_KEY|TOKEN)' "$df" >/dev/null; then
  fail "Dockerfile must not take secrets as ARG"
fi
if grep -E '^[[:space:]]*USER[[:space:]]' "$df" >/dev/null; then
  fail "Dockerfile must not set USER; the entrypoint starts as root only to chown /data"
fi
grep -q 'openjdk' "$df" && fail "Dockerfile must not install Java"
grep -q 'SHARPREPORT_UI_DIR' "$df" || fail "Dockerfile must set SHARPREPORT_UI_DIR"

entry="$root/deploy/docker-entrypoint.sh"
[ -f "$entry" ] || fail "missing deploy/docker-entrypoint.sh"
grep -q 'setpriv' "$entry" || fail "entrypoint must exec setpriv"
grep -q 'sharpreport' "$entry" || fail "entrypoint must drop to sharpreport"
grep -q '/data' "$entry" || fail "entrypoint must chown /data"

ignore="$repo/.dockerignore"
[ -f "$ignore" ] || fail "missing .dockerignore"
grep -q 'SharpReport' "$ignore" || fail ".dockerignore missing SharpReport"
if grep -E '^SharpReport$' "$ignore" >/dev/null; then
  fail ".dockerignore must not exclude the whole SharpReport tree"
fi
grep -q '\.env' "$ignore" || fail ".dockerignore missing .env"

example="$root/deploy/.env.production.example"
[ -f "$example" ] || fail "missing deploy/.env.production.example"
grep -q '^JWT_SECRET=$' "$example" || fail "JWT_SECRET must be empty in the example"
grep -q '^USERS_PANEL_BASE_URL=$' "$example" || fail "USERS_PANEL_BASE_URL must be empty in the example"
if grep -E 'change-me|admin123|metabase-secret|secret@' "$example" >/dev/null; then
  fail "example env contains a usable secret"
fi

compose="$root/docker-compose.yml"
[ -f "$compose" ] || fail "missing docker-compose.yml"
grep -q '/data' "$compose" || fail "compose must mount /data"
grep -q 'env_file' "$compose" || fail "compose must use env_file"
grep -q '.env.production' "$compose" || fail "env_file must point at .env.production"
if grep -E 'JWT_SECRET=|PASSWORD=' "$compose" >/dev/null; then
  fail "compose must not set a secret value"
fi
if [ -f "$root/deploy/docker-compose.yml" ]; then
  fail "deploy/docker-compose.yml must not return; it committed a JWT"
fi

readme="$root/README.md"
[ -f "$readme" ] || fail "missing SharpReport README"
grep -q 'docker build' "$readme" || fail "README must show docker build"
grep -q 'USERS_PANEL_BASE_URL' "$readme" || fail "README must document USERS_PANEL_BASE_URL"
grep -q 'GET /health' "$readme" || fail "README must document GET /health"
grep -q 'https://<sharpreport public host>' "$readme" || fail "README must name the public URL placeholder"

yaml="$repo/render.yaml"
deploy_readme="$repo/deploy/README.md"
[ -f "$yaml" ] || fail "missing render.yaml"
[ -f "$deploy_readme" ] || fail "missing deploy/README.md"
python3 - "$yaml" "$deploy_readme" <<'PY' || fail "render.yaml or runbook does not match the Data Access Blueprint"
import sys

yaml_path, deploy_readme = sys.argv[1:]
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
if any("key: VITE_DATAX_URL" in line for line in lines):
    errors.append("must not set VITE_DATAX_URL")

blocks = service_blocks(lines)
names = [service_name(block) for block in blocks]
if "morph" not in names or "morph-utils" not in names:
    errors.append("morph and morph-utils must stay, got " + ", ".join(names))
if names[:2] != ["morph", "morph-utils"]:
    errors.append("morph then morph-utils must stay first, got " + ", ".join(names))
if "sharpreport" not in names:
    errors.append("missing sharpreport")

svc = next((block for block in blocks if service_name(block) == "sharpreport"), None)
if svc is None:
    errors.append("missing sharpreport service")
else:
    body = "\n".join(svc)
    for needle in (
        "type: web",
        "runtime: docker",
        "region: singapore",
        "plan: starter",
        "branch: main",
        "dockerfilePath: ./SharpReport/Dockerfile",
        "dockerContext: .",
        "healthCheckPath: /health",
        "autoDeployTrigger: checksPass",
        "name: sharpreport-data",
        "mountPath: /data",
        "sizeGB: 1",
        "maxShutdownDelaySeconds: 120",
    ):
        if needle not in body:
            errors.append("sharpreport missing " + needle)
    for path in (
        "SharpReport/**",
        "pkg/morphai-rs/**",
        "SharpReport/Dockerfile",
        "SharpReport/deploy/docker-entrypoint.sh",
        ".dockerignore",
        "render.yaml",
    ):
        if path not in body:
            errors.append("sharpreport buildFilter missing " + path)

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

    def value_is(key, expected):
        body = "\n".join(env.get(key, []))
        if key not in env:
            errors.append("missing env " + key)
        elif "value: " + expected not in body and 'value: "' + expected + '"' not in body:
            errors.append(key + " must be " + expected)

    value_is("PORT", "3050")
    value_is("SHARPREPORT_PORT", "3050")
    value_is("SHARPREPORT_DATABASE_URL", "sqlite:///data/datapulse.db")
    value_is("SHARPREPORT_UI_DIR", "/app/ui")

    for key in ("USERS_PANEL_BASE_URL", "JWT_SECRET", "MORPH_AI_API_KEY"):
        body = "\n".join(env.get(key, []))
        if key not in env:
            errors.append("missing " + key)
            continue
        if "sync: false" not in body:
            errors.append(key + " must set sync: false")
        if "value:" in body:
            errors.append(key + " must not have a value")

doc = open(deploy_readme, encoding="utf-8").read()
for needle in (
    "prj-dahc33dbedkc73a1v8n0",
    "https://<morph public host>",
    "https://<sharpreport public host>",
    "USERS_PANEL_BASE_URL",
    "does not call Render",
    "not a secret",
    "VITE_DATAX_URL",
    "sharpreport-data",
    "/data",
    "3050",
):
    if needle not in doc:
        errors.append("deploy/README.md missing " + needle)

if errors:
    print("\n".join(errors), file=sys.stderr)
    sys.exit(1)
PY

echo "data access container contract ok"
