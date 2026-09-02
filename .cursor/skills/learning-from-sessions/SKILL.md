---
name: learning-from-sessions
description: Distills mistakes, user corrections, and redos from this session and past transcripts into durable lessons. Use after the user corrects the agent, undoes work, restates the same request, reports a regression, pastes an error after a claimed fix, or asks to learn from history.
---

# Learning From Sessions

A correction is wasted if it dies in the chat. Distill it into a reusable rule, then follow it.

Store: `.cursor/skills/session-lessons/lessons.md`

## What to capture

| Signal | Treat as |
|---|---|
| User rejects the approach ("wrong", "not that", "I said", "instead") | **Correction** |
| Same intent restated, "try again", "redo", "do it properly" | **Redo** |
| Error pasted after you claimed done | **Mistake** |
| Preference about names, layout, defaults, or verification | **Core experience** |

Capture four facts, then one rule:

- Intent: what they wanted
- Mistake: what you did instead
- Correction: their actual constraint (verbatim if short)
- Rule: the reusable "do X, not Y" that would have prevented it

## Persist

Append to `lessons.md` when the rule will matter in a later session (naming, UX, architecture, which verification to run).

Skip when it is a one-off typo, already in `lessons.md`, a secret, or only true for this file this once.

Newest first. Merge duplicates instead of stacking near-copies. Keep each lesson to 1–4 lines.

```markdown
## YYYY-MM-DD — short title
- Trigger: when this comes up
- Rule: do X, not Y
- Source: [Short Title](conversation-id)
```

## Apply now

Do not wait for a future chat. Change the current approach to match the new rule, then keep working.

Do not announce "I have logged a lesson" unless the user asked to learn or remember.
