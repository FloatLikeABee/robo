#!/usr/bin/env python3
"""Compact index of a Cursor agent transcript. Execute; do not read this file as reference."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

USER_QUERY_RE = re.compile(r"<user_query>\s*(.*?)\s*</user_query>", re.S)
WS_RE = re.compile(r"\s+")

CORRECTION_RE = re.compile(
    r"(oh that was wrong|that was wrong|completely wrong|not that|"
    r"\bi said\b|\bi told you\b|\bundo\b|\btry again\b|\bredo\b)",
    re.I,
)
ERROR_RE = re.compile(
    r"\b(error|exception|traceback|failed to|syntaxerror|module build failed|"
    r"cannot find|does the file exist)\b",
    re.I,
)
NOISE_RE = re.compile(
    r"Briefly inform the user about the task result|you have access to MCP",
    re.I,
)


def find_transcript(conv_id: str) -> Path | None:
    conv_id = conv_id.removesuffix(".jsonl").strip()
    root = Path.home() / ".cursor" / "projects"
    if not root.exists():
        return None
    matches = list(root.glob(f"*/agent-transcripts/{conv_id}/{conv_id}.jsonl"))
    if not matches:
        matches = list(root.glob(f"*/agent-transcripts/{conv_id}/*.jsonl"))
    if not matches:
        return None
    matches.sort(key=lambda p: p.stat().st_mtime, reverse=True)
    return matches[0]


def extract_user_text(obj: dict) -> str:
    content = obj.get("message", {}).get("content") or []
    chunks: list[str] = []
    if isinstance(content, str):
        chunks.append(content)
    elif isinstance(content, list):
        for part in content:
            if isinstance(part, dict) and part.get("type") == "text":
                chunks.append(part.get("text") or "")
            elif isinstance(part, str):
                chunks.append(part)
    raw = "\n".join(chunks)
    match = USER_QUERY_RE.search(raw)
    text = match.group(1) if match else raw
    return WS_RE.sub(" ", text).strip()


def classify(text: str, index: int) -> str:
    if NOISE_RE.search(text):
        return "noise"
    if ERROR_RE.search(text) and ("error" in text.lower() or "failed" in text.lower()):
        return "error-report"
    if index > 0 and CORRECTION_RE.search(text):
        return "correction"
    if text.startswith("/"):
        return "command"
    return "initial" if index == 0 else "follow-up"


def load_turns(path: Path) -> list[dict]:
    turns: list[dict] = []
    with path.open(encoding="utf-8") as handle:
        for line_no, line in enumerate(handle, start=1):
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
            except json.JSONDecodeError:
                continue
            if obj.get("role") != "user":
                continue
            text = extract_user_text(obj)
            if not text:
                continue
            kind = classify(text, len(turns))
            if kind == "noise":
                continue
            turns.append({"n": len(turns) + 1, "line": line_no, "kind": kind, "text": text})
    return turns


def main() -> int:
    parser = argparse.ArgumentParser(description="Extract user turns from a Cursor transcript")
    parser.add_argument("conversation_id", help="Conversation UUID (with or without .jsonl)")
    parser.add_argument("--json", action="store_true", help="Print JSON instead of text")
    parser.add_argument("--limit", type=int, default=40, help="Max user turns to print")
    parser.add_argument("--chars", type=int, default=360, help="Max characters per turn")
    args = parser.parse_args()

    path = find_transcript(args.conversation_id)
    if path is None:
        print(f"transcript not found: {args.conversation_id}", file=sys.stderr)
        return 1

    conv_id = path.stem
    turns = load_turns(path)
    shown = turns[: args.limit]
    payload = {
        "id": conv_id,
        "path": str(path),
        "user_turns": len(turns),
        "title": (shown[0]["text"][:80] if shown else conv_id),
        "turns": [
            {
                **turn,
                "text": turn["text"]
                if args.json
                else (turn["text"][: args.chars] + ("…" if len(turn["text"]) > args.chars else "")),
            }
            for turn in shown
        ],
    }

    if args.json:
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return 0

    print(f"id: {conv_id}")
    print(f"user_turns: {len(turns)}" + (f" (showing {len(shown)})" if len(turns) > len(shown) else ""))
    print()
    for turn in payload["turns"]:
        print(f"{turn['n']}. [{turn['kind']}] {turn['text']}")
        print()
    return 0


if __name__ == "__main__":
    sys.exit(main())
