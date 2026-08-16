## 1. Events & Info AI ingest (FormsX)

- [x] 1.1 Add backend ingest endpoint that accepts file(s) (txt/md/pdf), optional URL, and optional paste; require at least one source; reject unsupported types
- [x] 1.2 Implement source normalization (PDF text extract, safe URL fetch with limits, paste) into one corpus for MorphAI
- [x] 1.3 Return multiple Events & Info draft records from AI (array); clear errors when AI unavailable
- [x] 1.4 Register route + API client method; wire review/confirm → create via existing Events & Info create path
- [x] 1.5 Replace primary **Share collection** CTA with ingest/upload UI (file + URL + paste); demote or relocate share helpers if kept

## 2. ComposerX Compose content merge

- [x] 2.1 Merge Compose & Publish and Compose into one **Compose content** nav/page with publish available in that flow
- [x] 2.2 Rename Publish Records → **Published contents** in nav and panel copy
- [x] 2.3 Add reference file upload inside Compose content and auto-attach to the compose/AI session
- [x] 2.4 Remove standalone Reference docs nav module; redirect legacy page ids to Compose content (or Published contents when appropriate)
- [x] 2.5 Update ComposerX assistant/MCP copy that assumes a separate Reference docs module

## 3. Verify

- [x] 3.1 Smoke-test Events ingest (file-only, URL+paste, empty rejection) and ComposerX nav (Compose content, Published contents, no Reference docs, upload-while-compose)
