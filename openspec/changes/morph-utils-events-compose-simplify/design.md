## Context

See proposal.md for motivation.

**Events & Info today:** FormsX `EventsInfo.tsx` primary secondary CTA is **Share collection** (copy submit URL + email). Manual **Add New Entry** and prompt-based `POST /events-info/ai-draft` already exist (single draft JSON). No multi-file/URL/paste ingest that creates multiple records.

**Content Maker today:** ComposerX nav has Compose & Publish, Publish Records, Compose, Contents, and Reference docs. Compose & Publish and Compose overlap; reference docs are a full library (`/ai/reference-docs*`) selected into the AI composer session.

## Goals / Non-Goals

**Goals:**

- Events & Info ingest UI + API: file (txt/md/pdf) + URL + paste; ≥1 source; AI multi-draft extract; confirm then create.
- ComposerX nav: Compose content (merged), Published contents (rename), in-compose reference upload; remove Reference docs page from nav after wiring upload into compose.

**Non-Goals:**

- Deleting historical published pages or saved contents.
- Removing public Events & Info submit API entirely (out of scope unless it blocks the CTA swap).
- Building a new Morph Utils shell module — work lives in FormsX and ComposerX embeds.
- Guaranteeing perfect extraction quality; confirmation/edit of drafts is the safety net.

## Decisions

### 1. Events ingest API shape
- **Choice:** New protected endpoint (e.g. `POST /api/v1/events-info/ai-ingest`) accepting multipart and/or JSON fields: files[], url, paste/text. Server normalizes to one text corpus (PDF extract via existing `pkg/morphgraph` / repo PDF helpers; URL fetch with size/timeout caps; paste as-is), then MorphAI returns a **JSON array of drafts** `{title, detail, reporter, time}[]`. Client shows review list; create via existing `POST /events-info` per confirmed row (or a batch create if cheap).
- **Rationale:** Keeps create path audited; extends beyond single-prompt `ai-draft`.
- **Alternatives:** Overload `ai-draft` only — rejected; needs files/URL and multi-record output.

### 2. Share collection disposition
- **Choice:** Replace toolbar **Share collection** with **Upload / Import** (or equivalent). Move share URL/email into a secondary control (overflow/menu) or drop from primary chrome if unused; do not block ingest ship on share retention.
- **Rationale:** Matches user request; share is not the ingest job.

### 3. ComposerX merge strategy
- **Choice:** One page id `compose-content` that hosts the stronger of ComposePublishPanel + email-composer capabilities (markdown/HTML compose + publish actions). Default nav lands here. Legacy page ids redirect to `compose-content`. Keep saved **Contents** list if still distinct drafts; rename Publish Records panel label to **Published contents**.
- **Rationale:** User said Compose & Publish and Compose are the same; publish stays in the compose flow.
- **Alternatives:** Keep two panels behind one label — acceptable interim if UX is one nav item with tabs; prefer single canvas.

### 4. Reference docs migration
- **Choice:** Compose content gains an upload control that calls existing upload API (`POST /ai/reference-docs/upload` or a compose-scoped upload). Selected uploads auto-attach to the current AI compose session (`reference_document_ids`). Remove Reference docs nav page; optionally keep list/delete of uploads inside a small drawer on Compose content for this session’s docs only. Do not hard-delete DB files on day one — hide standalone module first.
- **Rationale:** “Upload when making content”; migrate then remove module UX.
- **Alternatives:** Full data wipe of reference library — rejected as non-goal.

### 5. Assistant / MCP copy
- **Choice:** Update ComposerX assistant system prompts to say reference files are attached in Compose content, not a separate Reference docs module.

## Risks / Trade-offs

- [URL fetch SSRF] → Mitigation: allowlist schemes http/https, block private IPs/metadata hosts, size/timeout limits.
- [Large PDFs / many events] → Mitigation: truncate corpus; cap draft count; user confirms subset.
- [Users lose Reference docs library browser] → Mitigation: upload in compose + optional compact recent-uploads list; keep backend for migration.
- [Compose merge drops a niche feature from one panel] → Mitigation: inventory ComposePublishPanel vs ContentMarkdownPanel features before merge; preserve publish + AI chat.

## Migration Plan

1. Ship Events ingest behind new CTA; keep manual add.
2. Ship ComposerX nav rename/merge + compose-time upload; redirect legacy page keys in `LAST_PAGE_KEY` / URL hash.
3. Remove Reference docs nav entry; verify assistant tools still work with uploaded ids.
4. Rollback: restore nav labels and Share collection button; leave new ingest endpoint inert if unused.

## Open Questions

- Whether batch create API is needed vs N× create — defer; N× create is fine for v1.
- Exact label string for Events CTA (“Upload”, “Import”, “Add from file”) — product can pick any clear ingest label.
