#!/bin/sh
# Fails if the DorphUtils stack create order drifts from the PO runbook.
# ponytail: line-order of two anchors, not a heading parser. Upgrade path is
# numbered-heading parse if this section grows past one screen.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
readme="$root/deploy/README.md"
blueprint="$root/render.yaml"

python3 - "$readme" "$blueprint" <<'PY'
import sys

readme_path, blueprint_path = sys.argv[1], sys.argv[2]
text = open(readme_path, encoding="utf-8").read()
blueprint = open(blueprint_path, encoding="utf-8").read()
errors = []

start = text.find("## DorphUtils stack on Render")
if start < 0:
    errors.append("deploy/README.md missing DorphUtils stack on Render")
    section = ""
else:
    rest = text[start + 1 :]
    nxt = rest.find("\n## ")
    section = text[start:] if nxt < 0 else text[start : start + 1 + nxt]

needles = (
    "Forge must not create services on Render",
    "prj-dahc33dbedkc73a1v8n0",
    "does not call Render",
    "dorph-utils",
    "formx",
    "composerx",
    "sharpreport",
    "dorph-engi",
    "https://<dorph-utils public host>",
    "https://<event-logs public host>",
    "https://<composerx public host>",
    "https://<sharpreport public host>",
    "https://<dorph-engi public host>",
    "https://<dorph public host>",
    "REACT_APP_DORPH_UTILS_URL",
    "VITE_SHEETX_URL",
    "VITE_FORMSX_URL",
    "VITE_COMPOSERX_URL",
    "VITE_DATAX_URL",
    "VITE_PROJECTS_URL",
    "VITE_DORPH_ENGI_URL",
    "USERS_PANEL_BASE_URL",
    "https://dorph-utils.onrender.com",
    "https://formx-vucj.onrender.com",
    "https://composerx.onrender.com",
    "https://sharpreport.onrender.com",
    "https://dorph-engi.onrender.com",
    "https://dorph-gjmb.onrender.com",
    "#114",
    "example, do not recreate",
    "Do not guess an `onrender.com` host from the service name",
    "missing or blank",
    "Dorph image rebuild is required",
    "restart without a rebuild does not set the header link",
    "Copy each public HTTPS origin",
    "rebuild the Dorph image",
)
for needle in needles:
    if needle not in section:
        errors.append("stack section missing " + needle)

record_at = section.find("Copy each public HTTPS origin")
rebuild_at = section.find("rebuild the Dorph image")
if record_at >= 0 and rebuild_at >= 0 and record_at > rebuild_at:
    errors.append("DorphUtils origin record must appear before the Dorph rebuild")

examples = (
    "https://dorph-utils.onrender.com",
    "https://formx-vucj.onrender.com",
    "https://composerx.onrender.com",
    "https://sharpreport.onrender.com",
    "https://dorph-engi.onrender.com",
    "https://dorph-gjmb.onrender.com",
)
for host in examples:
    if host in blueprint:
        errors.append("render.yaml must not contain example host " + host)
dorph_at = blueprint.find("\n    name: dorph\n")
utils_at = blueprint.find("\n    name: dorph-utils\n")
key_at = blueprint.find("key: REACT_APP_DORPH_UTILS_URL")
if dorph_at < 0 or utils_at < 0 or not (dorph_at < key_at < utils_at):
    errors.append("REACT_APP_DORPH_UTILS_URL must be on the dorph service")
elif blueprint.count("key: REACT_APP_DORPH_UTILS_URL") != 1:
    errors.append("REACT_APP_DORPH_UTILS_URL must be listed once")
else:
    body = []
    for line in blueprint[key_at:].splitlines()[1:]:
        if line.startswith("      - key:") or (line and not line.startswith("        ")):
            break
        body.append(line)
    prompt = "\n".join(body)
    if "sync: false" not in prompt:
        errors.append("REACT_APP_DORPH_UTILS_URL must set sync: false")
    if "value:" in prompt:
        errors.append("REACT_APP_DORPH_UTILS_URL must not have a value")

if errors:
    print("\n".join(errors), file=sys.stderr)
    sys.exit(1)
PY

echo "dorphutils stack runbook ok"
