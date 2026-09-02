---
name: harnessing-history
description: Recalls past Cursor chats and agent transcripts to reuse prior solutions and avoid repeated mistakes. Use when the user mentions history, last time, previous session, continue, as we did, redo, try again, search chats, or when starting work that was already attempted in this repo.
---

# Harnessing History

Past sessions are a first-class data source. Search them on `/harness-history`, when the user asks, and before repeating work.

**REQUIRED:** Apply matching rules from `session-lessons`. After a correction or redo, follow `learning-from-sessions`.

## When

Search history when any of these is true:

- User asks to search history, continue a chat, or do it like last time
- Request looks like work already done in this repo (same product, bug, or feature)
- User restates the same intent (a **redo**)
- About to take an approach that may have failed before

Do not skip because "I know this codebase" or "searching would slow us down".

## Workflow

1. Read `.cursor/skills/session-lessons/lessons.md` and apply matching rules.
2. Run several `SearchConversations` calls in parallel (1–2 keywords each). See [sources.md](sources.md). If that tool is missing, use the grep fallback in sources.md.
3. Pick the best 1–3 hits (prefer `[correction]` / `[error-report]` turns and ids already in `lessons.md`). Extract them:

```bash
python3 .cursor/skills/harnessing-history/scripts/extract-session.py <conversation-id>
```

4. From the extract, recover: original intent, what the agent did, user corrections, what finally worked.
5. Act on that. Do not repeat a failed approach.
6. If you cite a past chat to the user, use `[Short Title](conversation-id)` with at most six words in the title. Never mention transcript folder paths.

## Redo path

A restated request, "try again", "do it properly", or an error pasted after you claimed done is a redo.

1. Recover the original intent from this session first, then from history if needed.
2. Name the failed approach in one line (internal; do not lecture).
3. Take a different path that satisfies the original intent plus the correction.
4. Distill with `learning-from-sessions`.

## Output

Lead with the answer or the action. Mention a prior session only when it changes the work. Do not dump transcripts, snippets, or a tour of old chats.
