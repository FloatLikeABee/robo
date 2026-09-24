#!/bin/sh
# Fails if the MorphUtils stack create order drifts from the PO runbook.
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

start = text.find("## MorphUtils stack on Render")
if start < 0:
    errors.append("deploy/README.md missing MorphUtils stack on Render")
    section = ""
else:
    rest = text[start + 1 :]
    nxt = rest.find("\n## ")
    section = text[start:] if nxt < 0 else text[start : start + 1 + nxt]

needles = (
    "Forge must not create services on Render",
    "prj-dahc33dbedkc73a1v8n0",
    "does not call Render",
    "morph-utils",
    "formx",
    "composerx",
    "sharpreport",
    "morph-engi",
    "https://<morph-utils public host>",
    "https://<event-logs public host>",
    "https://<composerx public host>",
    "https://<sharpreport public host>",
    "https://<morph-engi public host>",
    "https://<morph public host>",
    "REACT_APP_MORPH_UTILS_URL",
    "VITE_SHEETX_URL",
    "VITE_FORMSX_URL",
    "VITE_COMPOSERX_URL",
    "VITE_DATAX_URL",
    "VITE_PROJECTS_URL",
    "VITE_MORPH_ENGI_URL",
    "USERS_PANEL_BASE_URL",
    "https://morph-utils.onrender.com",
    "https://formx-vucj.onrender.com",
    "https://composerx.onrender.com",
    "https://sharpreport.onrender.com",
    "https://morph-engi.onrender.com",
    "https://morph-gjmb.onrender.com",
    "#114",
    "example, do not recreate",
    "Do not guess an `onrender.com` host from the service name",
    "missing or blank",
    "Morph image rebuild is required",
    "restart without a rebuild does not set the header link",
    "Copy each public HTTPS origin",
    "rebuild the Morph image",
)
for needle in needles:
    if needle not in section:
        errors.append("stack section missing " + needle)

record_at = section.find("Copy each public HTTPS origin")
rebuild_at = section.find("rebuild the Morph image")
if record_at >= 0 and rebuild_at >= 0 and record_at > rebuild_at:
    errors.append("MorphUtils origin record must appear before the Morph rebuild")

examples = (
    "https://morph-utils.onrender.com",
    "https://formx-vucj.onrender.com",
    "https://composerx.onrender.com",
    "https://sharpreport.onrender.com",
    "https://morph-engi.onrender.com",
    "https://morph-gjmb.onrender.com",
)
for host in examples:
    if host in blueprint:
        errors.append("render.yaml must not contain example host " + host)
if "key: REACT_APP_MORPH_UTILS_URL" in blueprint:
    errors.append("render.yaml must not list key: REACT_APP_MORPH_UTILS_URL")

if errors:
    print("\n".join(errors), file=sys.stderr)
    sys.exit(1)
PY

echo "morphutils stack runbook ok"
