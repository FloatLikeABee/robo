#!/bin/sh
# Fails if the Project image or Blueprint drifts. No daemon.
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
grep -q '/health' "$df" || fail "HEALTHCHECK must call /health"
grep -q 'wget' "$df" || fail "HEALTHCHECK must use wget"
grep -q '${PORT}' "$df" || fail "HEALTHCHECK must expand \${PORT}"
if grep -F '$$' "$df" >/dev/null; then
  fail "Dockerfile must not contain \$\$; shell-form HEALTHCHECK runs under /bin/sh -c, where \$\$ is the PID"
fi
if grep -E '^[[:space:]]*USER[[:space:]]' "$df" >/dev/null; then
  fail "Dockerfile must not set USER; the entrypoint starts as root only to chown /data"
fi
if grep -E '^[[:space:]]*ARG[[:space:]]+.*(JWT|PASSWORD|SECRET|API_KEY|TOKEN)' "$df" >/dev/null; then
  fail "Dockerfile must not take secrets as ARG"
fi
grep -q 'VITE_API_BASE_URL=' "$df" || fail "Dockerfile must set VITE_API_BASE_URL empty"
if grep -E '^[[:space:]]*ARG[[:space:]]+VITE_API_BASE_URL' "$df" >/dev/null; then
  fail "VITE_API_BASE_URL must not be a build arg"
fi
if grep -q 'MORPH_ENGI_PORT' "$df"; then
  fail "Dockerfile must not set MORPH_ENGI_PORT; it overrides PORT"
fi

entry="$root/deploy/docker-entrypoint.sh"
[ -f "$entry" ] || fail "missing deploy/docker-entrypoint.sh"
grep -q 'gosu morphengi' "$entry" || fail "entrypoint must exec gosu morphengi"
grep -q '/data/uploads' "$entry" || fail "entrypoint must prepare /data/uploads"

ignore="$root/Dockerfile.dockerignore"
[ -f "$ignore" ] || fail "missing morph-engi/Dockerfile.dockerignore"
grep -q '\.env' "$ignore" || fail "Dockerfile.dockerignore missing .env"
grep -q 'uploads' "$ignore" || fail "Dockerfile.dockerignore missing uploads"
if grep -E '^morph-engi/?$' "$ignore" >/dev/null || grep -E '^pkg/?$' "$ignore" >/dev/null; then
  fail "Dockerfile.dockerignore must keep morph-engi/ and pkg/ in the build context"
fi

root_ignore="$repo/.dockerignore"
if ! grep -E '^morph-engi$' "$root_ignore" >/dev/null; then
  fail "root .dockerignore must keep a morph-engi line so the Morph context excludes Project"
fi

example="$root/deploy/.env.production.example"
[ -f "$example" ] || fail "missing deploy/.env.production.example"
grep -q '^USERS_PANEL_BASE_URL=$' "$example" || fail "USERS_PANEL_BASE_URL must be empty in the example"
grep -q '^JWT_SECRET=$' "$example" || fail "JWT_SECRET must be empty in the example"
grep -q '^MORPH_AI_API_KEY=$' "$example" || fail "MORPH_AI_API_KEY must be empty in the example"
if grep -E 'sk-|admin123|change-me|password=' "$example" >/dev/null; then
  fail "example env contains a secret-like value"
fi

readme="$root/README.md"
deploy_readme="$repo/deploy/README.md"
yaml="$repo/render.yaml"
[ -f "$readme" ] || fail "missing morph-engi README"
[ -f "$deploy_readme" ] || fail "missing deploy README"
[ -f "$yaml" ] || fail "missing render.yaml"

python3 - "$yaml" "$readme" "$deploy_readme" <<'PY' || fail "render.yaml or runbook does not match the Project Blueprint"
import sys

yaml_path, project_readme, deploy_readme = sys.argv[1:]
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

def env_map(block):
    env = {}
    current = None
    for line in block:
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
    return env

if any(line.startswith("projects:") for line in lines):
    errors.append("must not declare projects")
if "numInstances:" in text:
    errors.append("must not set numInstances")
if "generateValue:" in text:
    errors.append("must not generate secret values")

blocks = service_blocks(lines)
names = [service_name(block) for block in blocks]
if names[:2] != ["morph", "morph-utils"] or "morph-engi" not in names:
    errors.append("services must start with morph, morph-utils and include morph-engi, got " + ", ".join(names))

for forbidden in ("VITE_PROJECTS_URL", "VITE_MORPH_ENGI_URL"):
    if any(line.strip() == "- key: " + forbidden for line in lines):
        errors.append("must not set " + forbidden)

by_name = {service_name(block): block for block in blocks}
project = by_name.get("morph-engi")
if project is None:
    errors.append("missing morph-engi service")
else:
    body = "\n".join(project)
    for needle in (
        "type: web",
        "runtime: docker",
        "region: singapore",
        "plan: starter",
        "branch: main",
        "dockerfilePath: ./morph-engi/Dockerfile",
        "dockerContext: .",
        "healthCheckPath: /health",
        "autoDeployTrigger: checksPass",
        "name: morph-engi-data",
        "mountPath: /data",
        "sizeGB: 1",
        "maxShutdownDelaySeconds: 120",
    ):
        if needle not in body:
            errors.append("morph-engi missing " + needle)
    if "MORPH_ENGI_PORT" in body:
        errors.append("morph-engi must not set MORPH_ENGI_PORT")
    for path in (
        "morph-engi/**",
        "pkg/morphai-rs/**",
        "morph-engi/Dockerfile",
        "morph-engi/Dockerfile.dockerignore",
        "morph-engi/deploy/docker-entrypoint.sh",
        "render.yaml",
    ):
        if path not in body:
            errors.append("morph-engi buildFilter missing " + path)
    env = env_map(project)
    port = "\n".join(env.get("PORT", []))
    if 'value: "9096"' not in port and "value: 9096" not in port:
        errors.append("PORT must be 9096")
    for key, value in (
        ("APP_ENV", "production"),
        ("STATIC_DIR", "/app/frontend/dist"),
        ("MORPH_ENGI_DATABASE_URL", "sqlite:///data/morph_engi.db"),
        ("MORPH_ENGI_UPLOAD_DIR", "/data/uploads"),
    ):
        body_env = "\n".join(env.get(key, []))
        if key not in env:
            errors.append("missing env " + key)
        elif "value: " + value not in body_env and 'value: "' + value + '"' not in body_env:
            errors.append(key + " must be " + value)
    for key in ("USERS_PANEL_BASE_URL", "JWT_SECRET", "MORPH_AI_API_KEY"):
        body_env = "\n".join(env.get(key, []))
        if key not in env:
            errors.append("missing prompt " + key)
            continue
        if "sync: false" not in body_env:
            errors.append(key + " must set sync: false")
        if "value:" in body_env:
            errors.append(key + " must not have a value")

morph = by_name.get("morph")
if morph is None:
    errors.append("missing morph service")
else:
    morph_env = env_map(morph)
    port = "\n".join(morph_env.get("PORT", []))
    if 'value: "9090"' not in port:
        errors.append("morph PORT must stay 9090")

utils = by_name.get("morph-utils")
if utils is None:
    errors.append("missing morph-utils service")
else:
    utils_env = env_map(utils)
    port = "\n".join(utils_env.get("PORT", []))
    if 'value: "3040"' not in port:
        errors.append("morph-utils PORT must stay 3040")
    if "disk:" in "\n".join(utils):
        errors.append("morph-utils must not gain a disk")

needles = (
    "prj-dahc33dbedkc73a1v8n0",
    "https://<morph public host>",
    "https://<morph-engi public host>",
    "USERS_PANEL_BASE_URL",
    "not a secret",
    "does not call Render",
    "VITE_PROJECTS_URL",
    "VITE_MORPH_ENGI_URL",
    "MORPH_ENGI_DATABASE_URL",
    "MORPH_ENGI_UPLOAD_DIR",
    "/data",
    "morph-engi-data",
    "single instance",
    "/health",
    "projects",
    "9096",
    "#114",
    "JWT_SECRET",
    "MORPH_AI_API_KEY",
)
for doc in (project_readme, deploy_readme):
    doc_text = open(doc, encoding="utf-8").read()
    for needle in needles:
        if needle not in doc_text:
            errors.append(doc + " missing " + needle)

if errors:
    print("\n".join(errors), file=sys.stderr)
    sys.exit(1)
PY

echo "project container contract ok"
