#!/usr/bin/env python3
"""Check operator docs against the live MorphNotes nav, Research, and Invite Signup.

Invariants are the ones in openspec/changes/catch-docs-up-product-ia/design.md
(decision 6). The script reads the Invite Signup port from the Vite config and
the launcher, then requires the docs to use that same port.
"""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

DOCS = [
    ROOT / "README.md",
    ROOT / "docs/agents/00-architecture-overview.md",
    ROOT / "docs/agents/03-morph.md",
    ROOT / "docs/agents/12-build-deploy.md",
    ROOT / "morph/README.md",
]

NAV_DOCS = [
    ROOT / "README.md",
    ROOT / "docs/agents/00-architecture-overview.md",
    ROOT / "docs/agents/03-morph.md",
    ROOT / "morph/README.md",
]

NAV_LABELS = ("Tasks", "Timelines", "Big notes", "Research", "Generic data")

MORPH_ROUTES = (
    "/morphdata/case-tasks",
    "/morphdata/timelines",
    "/morphdata/big-notes",
    "/morphdata/research",
    "/morphdata/generic-data",
)

FORBIDDEN = ("Settings → Users", "/morphdata/configuration/users")

LINK_RE = re.compile(r"\[[^\]]*\]\(([^)]+)\)")
TICK_RE = re.compile(r"`([^`\n]+)`")
PORT_RE = re.compile(r"port:\s*(\d+)")
START_CMD_RE = re.compile(r"\./start-all\.sh\s+(?:start|stop|restart|logs)\s+([A-Za-z0-9_-]+)")


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def invite_port() -> str:
    vite = read(ROOT / "invite-signup/frontend/vite.config.ts")
    match = PORT_RE.search(vite)
    if not match:
        raise SystemExit("invite-signup/frontend/vite.config.ts has no server port")
    port = match.group(1)
    launcher = read(ROOT / "start-all.sh")
    if "invite-signup-ui" not in launcher:
        raise SystemExit("start-all.sh does not list invite-signup-ui")
    url = f"http://localhost:{port}"
    if not re.search(rf"invite-signup-ui\)\s+echo \"{re.escape(url)}\"", launcher):
        raise SystemExit(f"start-all.sh service_url for invite-signup-ui is not {url}")
    return port


def service_names() -> set[str]:
    text = read(ROOT / "start-all.sh")
    block = re.search(r"ALL_SERVICES=\((.*?)\)", text, re.S)
    if not block:
        raise SystemExit("start-all.sh has no ALL_SERVICES")
    names = set(re.findall(r"[A-Za-z0-9_-]+", block.group(1)))
    # Aliases from resolve_services case labels (words before |) plus "all".
    resolve = re.search(r"resolve_services\(\) \{.*?\n\}", text, re.S)
    if resolve:
        for label in re.findall(r"^\s+([A-Za-z0-9_|-]+)\)", resolve.group(0), re.M):
            if label in {"all|*", "*"}:
                continue
            for part in label.split("|"):
                if part not in {"*", ""}:
                    names.add(part)
    names.add("all")
    return names


def check_tokens(errors: list[str], port: str) -> None:
    url = f"http://localhost:{port}"
    texts = {path: read(path) for path in DOCS}
    for path, text in texts.items():
        rel = path.relative_to(ROOT)
        for bad in FORBIDDEN:
            if bad in text:
                errors.append(f"{rel} still contains {bad!r}")
    for path in NAV_DOCS:
        text = texts[path]
        rel = path.relative_to(ROOT)
        for label in NAV_LABELS:
            if label not in text:
                errors.append(f"{rel} missing nav label {label!r}")
    morph = texts[ROOT / "docs/agents/03-morph.md"]
    rel = "docs/agents/03-morph.md"
    for token in (
        *MORPH_ROUTES,
        "POST /api/invite/redeem",
        "POST /api/admin/invite-codes",
        "POST /api/tran/research/:id/publish",
        "GET /api/tran/public/research/:slug",
        "Notes & TODOs",
        "Context & Knowledge",
        "A new job runs five verified online rounds",
        "`/redeem`",
        "`/admin`",
        "`/morphdata/settings` and `/morphdata/configuration` redirect to `/morphdata/generic-data`",
        "invite-signup-ui",
        url,
    ):
        if token not in morph:
            errors.append(f"{rel} missing {token!r}")
    if "Files workspace" in morph:
        errors.append(f"{rel} still describes a Files workspace")
    for path in (ROOT / "README.md", ROOT / "docs/agents/12-build-deploy.md"):
        text = texts[path]
        rel = str(path.relative_to(ROOT))
        if "invite-signup-ui" not in text:
            errors.append(f"{rel} missing invite-signup-ui")
        if url not in text:
            errors.append(f"{rel} missing {url}")
    build = texts[ROOT / "docs/agents/12-build-deploy.md"]
    if not build.startswith("# 12 — Build and deploy\n"):
        errors.append("docs/agents/12-build-deploy.md opening heading changed")
    if "invite-signup" not in build.split("## Prerequisites", 1)[0]:
        errors.append("docs/agents/12-build-deploy.md product description omits invite-signup")
    if not table_row_has_empty_api(build):
        errors.append(
            "docs/agents/12-build-deploy.md service table has no invite-signup-ui row with an empty API cell"
        )
    readme = texts[ROOT / "README.md"]
    if not table_row_has_empty_api(readme):
        errors.append("README.md service table has no invite-signup-ui row with an empty API cell")


def table_row_has_empty_api(text: str) -> bool:
    for line in text.splitlines():
        if "invite-signup-ui" not in line or not line.strip().startswith("|"):
            continue
        cells = [c.strip().strip("`") for c in line.strip().strip("|").split("|")]
        if len(cells) < 2:
            continue
        api, ui = cells[0], cells[1]
        if ui == "invite-signup-ui" and api in {"—", "-", "–", ""}:
            return True
    return False


def check_links(errors: list[str]) -> None:
    for path in DOCS:
        text = read(path)
        for raw in LINK_RE.findall(text):
            target = raw.strip().split()[0]
            if not target or target.startswith(("http://", "https://", "mailto:", "#")):
                continue
            target = target.split("#", 1)[0].split("?", 1)[0]
            if not target:
                continue
            resolved = (path.parent / target).resolve()
            if not resolved.exists():
                errors.append(f"{path.relative_to(ROOT)} link {raw!r} does not resolve")


def looks_like_repo_path(token: str) -> bool:
    if token == "start-all.sh":
        return True
    if any(ch in token for ch in " =<>*"):
        return False
    if token.startswith(("/", "./", "../", "http://", "https://")):
        return False
    return "/" in token


def path_excused(text: str, token: str) -> bool:
    """Allow a missing path when the same paragraph says it is not in the repo."""
    for para in re.split(r"\n\s*\n", text):
        if f"`{token}`" not in para:
            continue
        low = para.lower()
        if any(
            phrase in low
            for phrase in (
                "absent",
                "missing",
                "not in this repo",
                "not in the repo",
                "when that tree exists",
            )
        ):
            return True
    return False


def gitignored(token: str) -> bool:
    candidates = [token]
    if not token.endswith("/"):
        candidates.append(token + "/")
    for candidate in candidates:
        result = subprocess.run(
            ["git", "check-ignore", "-q", "--no-index", candidate],
            cwd=ROOT,
            check=False,
        )
        if result.returncode == 0:
            return True
    return False


def check_paths(errors: list[str]) -> None:
    for path in DOCS:
        text = read(path)
        for token in TICK_RE.findall(text):
            token = token.strip()
            if not looks_like_repo_path(token):
                continue
            candidate = ROOT / token
            if candidate.exists():
                continue
            if path_excused(text, token):
                continue
            if gitignored(token):
                parent = candidate.parent
                parent_token = "" if parent == ROOT else str(parent.relative_to(ROOT))
                if parent == ROOT or parent.exists() or (parent_token and gitignored(parent_token)):
                    continue
            errors.append(f"{path.relative_to(ROOT)} path `{token}` does not exist")


def check_commands(errors: list[str], names: set[str]) -> None:
    for path in DOCS:
        text = read(path)
        for name in START_CMD_RE.findall(text):
            if name not in names:
                errors.append(
                    f"{path.relative_to(ROOT)} command uses unknown start-all service {name!r}"
                )


def main() -> int:
    errors: list[str] = []
    port = invite_port()
    names = service_names()
    check_tokens(errors, port)
    check_links(errors)
    check_paths(errors)
    check_commands(errors, names)
    if errors:
        print(f"{len(errors)} operator-doc check(s) failed:")
        for err in errors:
            print(f"  - {err}")
        return 1
    print("operator product docs match the live nav, Research publish paths, and Invite Signup")
    return 0


if __name__ == "__main__":
    sys.exit(main())
