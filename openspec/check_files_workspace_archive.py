#!/usr/bin/env python3
"""Check that removed Morph AI Files-workspace OpenSpec changes stay archived.

The script is the regression check for openspec-files-workspace-archive.
It exits 0 only when the five Files-workspace changes are dated archive
history, their deltas were not merged into live specs, and the drop-files
policy is either still active or archived with its delta synced.
"""

from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[1]
CHANGES = ROOT / "openspec" / "changes"
ARCHIVE = CHANGES / "archive"
SPECS = ROOT / "openspec" / "specs"
DATE = "2026-09-24"

ARCHIVED = (
    "morphai-files-session-workspace",
    "morphai-restore-folder-workspace-session",
    "morphai-files-folder-survives-refresh",
    "morphai-remember-files-workspace-open",
    "morphai-pinned-files-readable-context",
)

FORBIDDEN_LIVE_CAPABILITIES = (
    "morphai-files-session-workspace",
    "morphai-folder-workspace-session",
    "morphai-files-folder-persist",
    "morphai-pinned-files-context",
    "morphai-workspace-open-persist",
)

POLICY_CAPABILITIES = {
    "openspec-files-workspace-archive",
    # Synced drop policy. It names the same surfaces in order to forbid them.
    "morphai-no-files-workspace",
}

LIVE_FILES_TERMS = (
    "Files tab",
    "AgentFilesTab",
    "morphai-files-workspace",
    "Open folder",
    "recent folders",
    "pin-from-folder",
    "filesWorkspaceStore",
)


def archived_drop_dirs():
    if not ARCHIVE.is_dir():
        return []
    return [
        path
        for path in ARCHIVE.iterdir()
        if path.is_dir() and path.name.endswith("-morphai-drop-files-workspace")
    ]


def drop_policy_errors():
    """Accept the drop policy active, or archived with its delta synced."""
    active = (CHANGES / "morphai-drop-files-workspace" / "proposal.md").is_file()
    archived = archived_drop_dirs()
    synced = (SPECS / "morphai-no-files-workspace" / "spec.md").is_file()
    if active:
        return []
    if archived and synced:
        return []
    if archived and not synced:
        return ["morphai-drop-files-workspace is archived without a synced morphai-no-files-workspace spec"]
    return ["morphai-drop-files-workspace is neither active nor archived with its spec synced"]


def main() -> int:
    errors = []

    for name in ARCHIVED:
        active = CHANGES / name
        if active.exists():
            errors.append(f"still active: {name}")
        archived = ARCHIVE / f"{DATE}-{name}"
        if not archived.is_dir():
            errors.append(f"missing archive folder: {archived.relative_to(ROOT)}")
            continue
        proposal = archived / "proposal.md"
        if not proposal.is_file():
            errors.append(f"missing proposal in archive: {name}")
            continue
        text = proposal.read_text(encoding="utf-8")
        if "## Why" not in text:
            errors.append(f"archived proposal lost its Why section: {name}")
        if "Superseded" not in text or "morphai-drop-files-workspace" not in text:
            errors.append(f"missing superseded marker: {name}")
        if "do not sync these delta specs" not in text.lower() and "Do not sync these delta specs" not in text:
            errors.append(f"archived proposal does not say to skip spec sync: {name}")
        for rel in ("design.md", "tasks.md", ".openspec.yaml"):
            if not (archived / rel).is_file():
                errors.append(f"missing {rel} in archive: {name}")
        spec_files = list((archived / "specs").rglob("*.md")) if (archived / "specs").is_dir() else []
        if not spec_files:
            errors.append(f"missing delta specs in archive: {name}")

    remember = ARCHIVE / f"{DATE}-morphai-remember-files-workspace-open" / "proposal.md"
    if remember.is_file():
        remember_text = remember.read_text(encoding="utf-8")
        if "not a reason to restore the Files tab" not in remember_text:
            errors.append("remember-files archive does not keep the pane-open disclaimer")
    restore = ARCHIVE / f"{DATE}-morphai-restore-folder-workspace-session" / "proposal.md"
    if restore.is_file():
        restore_text = restore.read_text(encoding="utf-8")
        if "Last-chat restore without a folder binding" not in restore_text:
            errors.append("restore-folder archive does not keep the last-chat disclaimer")

    for capability in FORBIDDEN_LIVE_CAPABILITIES:
        if (SPECS / capability).exists():
            errors.append(f"live spec exists: {capability}")

    if SPECS.is_dir():
        for spec in SPECS.rglob("*.md"):
            text = spec.read_text(encoding="utf-8")
            mentions_files = any(needle in text for needle in LIVE_FILES_TERMS)
            if not mentions_files:
                continue
            if spec.parent.name not in POLICY_CAPABILITIES:
                errors.append(f"live spec describes the Files tab: {spec.relative_to(ROOT)}")

    errors.extend(drop_policy_errors())

    for name, needle in (
        (
            "morphai-agent-workspace",
            "not permission to restore that tab",
        ),
        (
            "morphai-restore-missing-webpack-modules",
            "Do not restore `AgentFilesTab`",
        ),
        (
            "platform-trim-readme-chat-skills",
            "not permission to restore the IndexedDB Files tab",
        ),
    ):
        folder = CHANGES / name
        if not folder.is_dir():
            errors.append(f"mixed change is not active: {name}")
            continue
        if (ARCHIVE / f"{DATE}-{name}").exists():
            errors.append(f"mixed change was archived: {name}")
        proposal = (folder / "proposal.md").read_text(encoding="utf-8")
        if "Superseded" not in proposal or "morphai-drop-files-workspace" not in proposal:
            errors.append(f"mixed change missing Files disclaimer: {name}")
        if needle not in proposal:
            errors.append(f"mixed change disclaimer does not name {needle}: {name}")

    project = ROOT / "openspec" / "project.md"
    if not project.is_file():
        errors.append("missing openspec/project.md warning")
    else:
        warning = project.read_text(encoding="utf-8")
        for needle in (
            "product owner",
            "IndexedDB",
            "morphai-drop-files-workspace",
            "AgentFilesTab",
        ):
            if needle not in warning:
                errors.append(f"openspec/project.md missing {needle!r}")
        if "yes" not in warning.lower():
            errors.append("openspec/project.md does not require an explicit product-owner yes")

    config = (ROOT / "openspec" / "config.yaml").read_text(encoding="utf-8")
    if "openspec/project.md" not in config or "IndexedDB Files tab" not in config:
        errors.append("openspec/config.yaml context does not point at the Files-tab warning")

    if errors:
        print(f"{len(errors)} archive check failure(s):", file=sys.stderr)
        for err in errors:
            print(f"- {err}", file=sys.stderr)
        return 1
    print("files-workspace archive check passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
