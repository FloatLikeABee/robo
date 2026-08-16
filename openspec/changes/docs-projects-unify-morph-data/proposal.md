## Why

After the last Morph Utils / Morph Data simplification, product labels and module boundaries still don’t match how teams work: Academi is really a docs workspace, Engi is projects, Booki’s only remaining useful surface is Flow log, and Morph Data still splits People/Assets/Activities and treats Stories like a social feed. This change consolidates naming and modules so Utils and Morph Data match the intended IA.

## What Changes

- **Academi → Docs**: Rename the Morph Utils module and user-facing brand from Academi to **Docs** (nav, shell labels, embed copy). Keep study/docs product behavior; only rename the product surface. (**BREAKING** for bookmarks/URLs that hardcode “Academi” branding.)
- **Booki → Projects (Flow log only)**: Keep only **Flow log** from Booki. Category becomes a free-text input (user can type any category; no forced picker-only list). Move Flow log into **Projects** (current Engi). Remove the whole Booki module from Morph Utils. (**BREAKING** for Booki AI, Bookings, and any remaining Booki routes/embeds.)
- **Engi → Projects**: Rename Morph Engi / Projects branding consistently to **Projects** in Utils and the Engi app chrome.
- **Morph Data — Activities removed**: Remove the Activities module (nav, routes, import entity, related Work-data entry points). Case/work items remain as **Tasks** only.
- **Morph Data — Tasks without assignee**: Remove assignee as a first-class field/UI/requirement on tasks. Assignment and related metadata belong in description or free-form JSON/detail when users need them. (**BREAKING** for APIs/UI that required assignees.)
- **Morph Data — People + Assets combined**: Merge People and Assets into one module (single nav item and unified list/detail experience). Update File import entity options accordingly.
- **Morph Data — Story Board → Stories**: Rename to **Stories**. Present as tile/card grid like Tasks; open a story into a detail view. Treat items as notes (not posts); rename comments to **notes**.

## Capabilities

### New Capabilities

- `morph-utils-docs`: Academi product surface renamed to Docs across Morph Utils and embed chrome.
- `morph-utils-projects-flow-log`: Flow log lives in Projects with free-text category; Booki module removed; Engi branded as Projects.
- `morph-data-tasks`: Activities removed; Tasks without assignee; description/JSON carry optional assignment detail.
- `morph-data-resources`: Combined People + Assets module and import targeting.
- `morph-data-stories`: Stories tile grid + detail; comments renamed to notes.

### Modified Capabilities

- (none — prior Morph Utils/Data specs live only under the previous change and were never archived to `openspec/specs/`; this change introduces superseding capabilities above.)

## Impact

- **morph-utils**: Module list (`academi` → `docs`, remove `booki`, `projects` label/branding), icons, embed URLs, Settings copy.
- **academi web/app**: Brand strings Academi → Docs where user-facing.
- **morph-engi (Projects)**: Brand Engi → Projects; add Flow log UI/API integration (proxy or port from Booki).
- **booki**: Remove from Utils embeds; Flow log capability relocated; module can be retired from start scripts/nav.
- **Morph Data frontend**: `AppDrawer`, routes for activities/people/assets/story-board, CaseTasks assignee UI, StoryBoard → Stories card UX + notes.
- **Morph / Tran API**: Case-task assignee validation; story comments → notes naming; entity routes/import allowlists; optional Activities deprecation.
- **Docs / agent maps**: Port tables and product map updated for Docs / Projects / no Booki.
