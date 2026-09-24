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

repo=$(CDPATH= cd -- "$root/.." && pwd)
yaml="$repo/render.yaml"
deploy_readme="$repo/deploy/README.md"
[ -f "$yaml" ] || fail "missing render.yaml"
[ -f "$deploy_readme" ] || fail "missing deploy/README.md"
python3 - "$yaml" "$deploy_readme" "$readme" <<'PY' || fail "render.yaml or runbook does not match the MorphUtils Blueprint"
import sys

yaml_path, deploy_readme, utils_readme = sys.argv[1:]
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

blocks = service_blocks(lines)
names = [service_name(block) for block in blocks]
if names != ["morph", "morph-utils", "formx", "composerx", "sharpreport", "morph-engi"]:
    errors.append(
        "service names must be morph, morph-utils, formx, composerx, sharpreport, morph-engi, got "
        + ", ".join(names)
    )

utils = next((block for block in blocks if service_name(block) == "morph-utils"), None)
if utils is None:
    errors.append("missing morph-utils service")
else:
    body = "\n".join(utils)
    for needle in (
        "type: web",
        "runtime: docker",
        "region: singapore",
        "plan: starter",
        "branch: main",
        "dockerfilePath: ./morph-utils/Dockerfile",
        "dockerContext: ./morph-utils",
        "healthCheckPath: /health",
        "autoDeployTrigger: checksPass",
    ):
        if needle not in body:
            errors.append("morph-utils missing " + needle)
    if any(line.startswith("    disk:") for line in utils):
        errors.append("morph-utils must not declare a disk")
    if "numInstances:" in body:
        errors.append("morph-utils must not set numInstances")
    for path in (
        "morph-utils/frontend/**",
        "morph-utils/Dockerfile",
        "morph-utils/.dockerignore",
        "morph-utils/deploy/docker-entrypoint.sh",
        "morph-utils/deploy/nginx.conf.template",
        "render.yaml",
    ):
        if path not in body:
            errors.append("morph-utils buildFilter missing " + path)

    env = {}
    current = None
    for line in utils:
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
    port = "\n".join(env.get("PORT", []))
    if 'value: "3040"' not in port and "value: 3040" not in port:
        errors.append("PORT must be 3040")
    origin = "\n".join(env.get("VITE_MORPH_API_URL", []))
    if "VITE_MORPH_API_URL" not in env:
        errors.append("missing VITE_MORPH_API_URL")
    else:
        if "sync: false" not in origin:
            errors.append("VITE_MORPH_API_URL must set sync: false")
        if "value:" in origin:
            errors.append("VITE_MORPH_API_URL must not have a value")
    if "REACT_APP_MORPH_UTILS_URL" in env:
        errors.append("must not set REACT_APP_MORPH_UTILS_URL")
    for key in (
        "VITE_SHEETX_URL",
        "VITE_FORMSX_URL",
        "VITE_COMPOSERX_URL",
        "VITE_DATAX_URL",
        "VITE_PROJECTS_URL",
        "VITE_MORPH_ENGI_URL",
        "VITE_MORPH_AI_URL",
    ):
        body_env = "\n".join(env.get(key, []))
        if key not in env:
            errors.append("missing prompt " + key)
        elif "sync: false" not in body_env:
            errors.append(key + " must set sync: false")
        elif "value:" in body_env:
            errors.append(key + " must not have a value")

for doc in (deploy_readme, utils_readme):
    doc_text = open(doc, encoding="utf-8").read()
    for needle in (
        "prj-dahc33dbedkc73a1v8n0",
        "https://<morph public host>",
        "https://<morph-utils public host>",
        "REACT_APP_MORPH_UTILS_URL",
        "does not call Render",
        "not a secret",
        "non-loopback",
        "Authorization",
        "does not edit CORS",
        "VITE_USERS_PANEL_API_URL",
        "VITE_SHEETX_URL",
        "VITE_FORMSX_URL",
        "VITE_COMPOSERX_URL",
        "VITE_DATAX_URL",
        "VITE_PROJECTS_URL",
        "VITE_MORPH_ENGI_URL",
        "VITE_MORPH_AI_URL",
        "PORT",
        "3040",
    ):
        if needle not in doc_text:
            errors.append(doc + " missing " + needle)

if errors:
    print("\n".join(errors), file=sys.stderr)
    sys.exit(1)
PY

echo "morph-utils container contract ok"
