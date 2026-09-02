---
name: "/harness-history"
id: "harness-history"
category: "Workflow"
description: "Search past chats, apply session lessons, and distill mistakes, corrections, and redos"
---

Harness history sessions. Follow the project skills `harnessing-history`, `session-lessons`, and `learning-from-sessions`.

**Input**: The argument after `/harness-history` is the topic to recall, a conversation id, `learn` to distill this session, or empty to apply lessons then ask what to search.

**Steps**

1. Read `.cursor/skills/session-lessons/lessons.md` and apply matching rules silently.
2. If the argument is `learn` (or the user asked to learn/remember): follow `learning-from-sessions` on this chat's mistakes, corrections, and redos. Update `lessons.md` only for durable rules.
3. Otherwise search past chats for the topic (or for the current task if the argument is empty and the task is already known):
   - Several `SearchConversations` calls in parallel, 1–2 keywords each
   - Extract the best 1–3 hits:

```bash
python3 .cursor/skills/harnessing-history/scripts/extract-session.py <conversation-id>
```

   - Recover intent, failed approaches, corrections, and what worked
4. Answer or continue the work using that. Cite a past chat as `[Title](conversation-id)` only when it changes the work. Do not dump transcripts.

If the user is in the middle of a redo (restated request, "try again", pasted error after a claimed fix), take a different path than the failed one, then distill.
