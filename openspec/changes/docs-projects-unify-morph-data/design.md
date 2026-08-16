## Context

See proposal.md for motivation. Today Morph Utils still embeds Academi, Booki, and Projects/Engi as separate apps; Booki’s only valued remaining surface is Flow log (`booki` `/flow-log` + `/api/v1/flow-log/*`). Morph Data still lists People, Assets, Activities, and Story Board separately; CaseTasks still enforce assignees; Story Board is feed/post-oriented with comments.

Constraints: Morph Utils iframe embeds; shared UsersPanel auth; prefer relocating Flow log into morph-engi rather than keeping a Booki shell; avoid inventing a second assignee system.

## Goals / Non-Goals

**Goals:**
- Rename Academi → Docs and Engi → Projects in Utils + app chrome.
- Move Flow log into Projects with free-text category; remove Booki from Utils.
- Morph Data: drop Activities; Tasks without assignee; combine People+Assets; Stories as card tiles with notes.

**Non-Goals:**
- Rewriting Academi’s core study/docs AI product beyond branding.
- Keeping Booki AI or Bookings.
- Full schema deletion of legacy Activities/assignee/comment columns in one release (API can soft-deprecate).
- Redesigning SurveyX or ComposerX.

## Decisions

### 1. Flow log ownership: port into morph-engi, proxy Booki API initially
- **Choice:** Add a Flow log section in morph-engi UI; call existing Booki `/api/v1/flow-log/*` via Engi’s backend proxy or direct Booki base URL config until Booki service is retired.
- **Why:** Fastest path to “one module for both” without rewriting persistence first.
- **Alternatives:** Copy tables into Engi DB immediately (more migration risk); keep Booki microservice forever (fails “remove Booki”).

### 2. Category field: plain text input
- **Choice:** Replace closed select / suggestion-only UX with a text input; optional datalist suggestions allowed but not required.
- **Why:** Matches “user can write anything.”

### 3. Combined People + Assets (“Resources”)
- **Choice:** One drawer item opening a unified page with type filter/tabs (People | Assets) reusing existing list/detail components.
- **Why:** Minimal API churn; clear UX merge without merging MySQL tables yet.
- **Alternatives:** True single polymorphic table (deferred).

### 4. Activities removal
- **Choice:** Remove nav/routes/import first; leave backend trip endpoints unused unless trivial to hide.
- **Why:** Matches product intent without blocking on DB teardown.

### 5. Tasks assignee removal
- **Choice:** Drop assignee UI and “at least one assignee” validation; stop sending assignees on create/update; keep reading old assignee data as optional display in detail JSON if present, without editing.
- **Why:** Users put assignment in description/JSON.

### 6. Stories UX
- **Choice:** Card grid → detail panel patterned after CaseTasks tiles; rename comment APIs/UI strings to notes (alias routes if needed for compatibility).
- **Why:** Notes mental model; reuse task interaction pattern.

### 7. Docs / Projects Utils ids
- **Choice:** Prefer stable Utils module ids where possible (`projects` already exists; rename academi id to `docs` and drop `booki`). Update icons/labels/URLs.
- **Why:** Clear IA; accept embed URL/bookmark breaks for Academi→Docs.

## Risks / Trade-offs

- [Booki API still required while Flow log proxied] → Mitigate with Engi env `BOOKI_FLOW_LOG_BASE_URL`; plan follow-up to migrate storage into Engi.
- [Assignee data orphaned] → Mitigate by leaving historical assignees readable in detail JSON; document that new tasks omit them.
- [People/Assets combine confuses power users] → Mitigate with clear People/Assets filter tabs inside the module.
- [Comment→notes rename breaks clients] → Mitigate with dual-label UI and route aliases during transition.

## Migration Plan

1. Ship Utils renames (Docs, Projects) and remove Booki module entry.
2. Ship Projects Flow log UI + free-text category against Booki API.
3. Morph Data: Activities nav/import removal; Tasks assignee UI/validation removal; Resources combined nav; Stories cards + notes labels.
4. Follow-up (optional): migrate flow-log persistence into Engi; archive Booki service; archive prior OpenSpec change.

## Open Questions

- Final Resources nav label: **Resources** vs **People & assets** (default to **Resources** unless product prefers the longer label).
- Whether story “notes” need a new API resource or only UI rename of comments (default: UI + JSON field rename with comment route alias).
