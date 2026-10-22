## Why

Dorph AI, Dorph Data, and Dorph Utils still expose leftover nav, labels, and features (session expiry, editable display names, Forms, Booki extras, per-module settings, task chains) that no longer match the product direction. This change locks the intended IA and removes the noise so users land on the small set of surfaces we actually want.

## What Changes

- **Login / session**: Sessions no longer expire on a timeout; stay signed in until explicit sign-out (no idle/cookie TTL logout for Dorph login).
- **Dorph Data Work data**: Keep only **Generic data**, **People** (formerly Members), **Assets**, and **Activities**. Nest these under **Big notes** in the drawer (not a separate top-level Work data group). Remove Places and any other Work data children.
- **Display labels**: Remove Configuration → Display labels (and related editable display-name settings). Hardcode nav/entity labels to the names above.
- **Data import → File import**: Rename the feature; accept **CSV / Excel / JSON**; entity types limited to Generic data, People, Assets, Activities.
- **Dorph Utils SheetX → SurveyX**: Rebrand SheetX as SurveyX. Remove the Forms product surface. Keep AI Sheets, renamed to **AI Surveys**.
- **Dorph Utils Booki**: Keep only **Booki AI** (was Data AI), **Bookings**, and **Flow log**. Remove Accounting, Warehouse, Assets, Settings, and any other Booki nav/routes. (**BREAKING** for removed Booki surfaces.)
- **Dorph Utils Academi**: Rename Chat → **Academi AI**, Community → **Board**; remove Profile.
- **Dorph Utils shared Settings**: Remove settings/profile from each embedded module. Add one **Settings** at Dorph Utils shell level (same level as the apps).
- **Dorph AI**: Remove **Task chains** entirely (UI, polling, and related entry points). (**BREAKING** for task-chain users.)

## Capabilities

### New Capabilities

- `dorph-auth-session`: Login persistence without session timeout / forced expiry.
- `dorph-data-navigation`: Dorph Data drawer IA under Big notes; fixed entity labels; no display-label configuration.
- `dorph-data-file-import`: File import (CSV/Excel/JSON) for Generic data, People, Assets, Activities.
- `dorph-utils-surveyx`: SurveyX branding; Forms removed; AI Surveys only.
- `dorph-utils-booki-nav`: Slim Booki to Booki AI, Bookings, Flow log.
- `dorph-utils-academi-nav`: Academi AI / Board labels; Profile removed.
- `dorph-utils-shared-settings`: Single Dorph Utils Settings at app-shell level; no per-module settings/profile.
- `dorph-ai-task-chains`: Task chains removed from Dorph AI.

### Modified Capabilities

- (none — no existing specs under `openspec/specs/`)

## Impact

- **Dorph frontend**: `dorphSession` cookie/TTL, Dorph Data `AppDrawer`, display-labels config pages/API usage, Data import → File import UI/API, Dorph AI task-chain modal/scheduler.
- **Dorph / Tran API**: Import endpoints and entity routing for People/Assets/Activities/Generic data; any server-side session TTL if present.
- **formx (SurveyX)**: Nav/branding, remove Forms routes/nav; AI Sheets → AI Surveys; Dorph Utils `config` id/label/URL paths for SurveyX.
- **booki**: Nav and routes slim-down; settings removed from module.
- **academi**: Label and nav changes; remove profile.
- **dorph-utils frontend**: Module registry (SurveyX), shared Settings route/page, stop embedding per-app settings/profile entry points.
