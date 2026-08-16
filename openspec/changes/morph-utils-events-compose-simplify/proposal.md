## Why

Morph Utils surfaces two friction points: Events & Info still centers on “Share collection” instead of ingesting real source material, and Content Maker (ComposerX) duplicates nearly identical Compose flows while keeping Reference docs as a separate module users must manage. Users need one AI ingest path for events from files/URLs/paste, and one Compose content surface with publish plus in-flow reference uploads.

## What Changes

### Events & Info (SheetX / FormsX, under Morph Utils Survey Maker / Forms area)

- Replace the primary **Share collection** action with an **ingest / upload** flow for Events & Info.
- Users MAY supply content via any combination of:
  - **File upload** (`.txt`, `.md`, `.pdf`)
  - **URL** (fetch page/content)
  - **Paste** (free-text / clipboard content)
- At least **one** input method MUST be provided before AI runs.
- AI extracts events and info from the combined inputs and proposes data records; users confirm before persistence (aligned with existing draft-then-save Events & Info patterns where practical).
- Share-collection email/URL helpers MAY remain secondary or elsewhere if still useful; they MUST NOT remain the primary CTA labeled “Share collection” for this ingest job.

### Content Maker (ComposerX)

- **BREAKING (nav):** Merge **Compose & Publish** and **Compose** into a single nav item: **Compose content**, which includes publish functionality in the same flow.
- Rename **Publish Records** → **Published contents**.
- **BREAKING:** Remove the standalone **Reference docs** module from nav/UI.
- Reference material becomes **upload-while-composing** inside Compose content (attach files for that session/draft); migrate existing reference-doc usage into that model, then retire the separate library module UX.
- Keep a list of published pages as **Published contents**.

## Capabilities

### New Capabilities

- `events-info-ai-ingest`: Multi-source (file / URL / paste) AI extraction into Events & Info draft records, requiring at least one source.
- `composerx-compose-content`: Unified Compose content surface with publish, published-contents list, and in-compose reference uploads; Reference docs module removed after migration.

### Modified Capabilities

- (none — main `openspec/specs/` has no archived baselines for these products)

## Impact

- **FormsX**: `formx/frontend/src/pages/EventsInfo.tsx`, `formx/frontend/src/lib/api.ts`, `formx/backend/internal/handler/events_info*.go` — new ingest endpoints or extended AI draft accepting multipart/URL/paste; PDF text extraction reuse from existing repo patterns.
- **ComposerX**: `composerx/frontend/src/App.svelte` nav and panels (`ComposePublishPanel`, email composer, `ComposePublishRecordsPanel`, reference-docs page); backend reference-doc APIs may stay for storage but primary UX is compose-time upload; publish records labeling.
- **Morph Utils shell**: mostly embeds ComposerX/SheetX — verify labels/deep links if any hard-code old nav names.
- **Docs / assistants**: ComposerX platform assistant prompts that assume a Reference docs module need updating.
