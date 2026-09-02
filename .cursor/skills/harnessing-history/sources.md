# History sources

Load this when searching or opening past sessions.

## SearchConversations

Index of local chats plus cached cloud agent chats. Returns ids, titles, and short snippets — not full transcripts. Some cloud chats are not searchable.

Query rules:

- Every unquoted keyword must match the **same** conversation. Long queries return nothing.
- Use **1–2 keywords** per call. Run several searches in parallel.
- Quotes for a short exact phrase only: `"Survey Maker"`.
- Prefer product names, error tokens, and distinctive phrases over generic words (`wrong`, `error`, `fix`).

Recipes (run the column in parallel):

| Intent | Searches |
|---|---|
| Prior work on a product | product name; distinctive feature word |
| User correction | `"that was wrong"`; product + `tabs` (or the UI word they used) |
| Redo / retry | `redo`; `"try again"`; distinctive error token |
| Naming / UX preference | `"name change"`; visible label |
| Continue a named chat | distinctive title words from the user |

If zero hits, split the query. Do not add more keywords.

If `SearchConversations` is not available (subagents, other runtimes), grep user prompts only:

```bash
rg -l "<user_query>" ~/.cursor/projects/*/agent-transcripts/*/*.jsonl | head
rg -l "that was wrong" ~/.cursor/projects/*/agent-transcripts/*/*.jsonl
```

Then extract by id. Do not grep assistant/tool dumps. Do not tell the user these paths.

## Transcript files

After you have an id, extract — do not read the whole JSONL first.

```bash
python3 .cursor/skills/harnessing-history/scripts/extract-session.py <conversation-id>
python3 .cursor/skills/harnessing-history/scripts/extract-session.py --json <conversation-id>
```

The script finds `~/.cursor/projects/*/agent-transcripts/<id>/<id>.jsonl`.

JSONL shape: one JSON object per line, `role` is `user` or `assistant`. User text often wraps the real prompt in `<user_query>…</user_query>`. Assistant lines can be huge; the extractor skips tool dumps.

If you must open the file directly: `Grep` for `<user_query>` or Read with a line offset. Never load a full multi-hundred-line transcript into context.

Subagent logs live under `…/agent-transcripts/<id>/subagents/`. Ignore them unless the parent extract is not enough.

## Citation

To the user: `[Title](uuid)` — title ≤ 6 words, uuid without `.jsonl`.

Do not mention `agent-transcripts`, `.jsonl`, or the projects folder path.
