## 1. Dorph auth — no session timeout

- [x] 1.1 Remove finite Max-Age logout policy from Dorph session cookie/token persistence (`dorphSession.js` and any related login remember flow)
- [x] 1.2 Align UsersPanel / shared session server TTL or refresh so Dorph sessions do not expire solely due to timeout
- [x] 1.3 Verify explicit Sign out still clears Dorph session across Dorph AI / Dorph Data / Dorph Utils

## 2. Dorph Data navigation and labels

- [x] 2.1 Nest Generic data, People, Assets, and Activities under Big notes in `AppDrawer`; remove standalone Work data section and Places item
- [x] 2.2 Hardcode nav/entity labels (Generic data, People, Assets, Activities, Big notes); stop loading editable display-label overrides for those names
- [x] 2.3 Remove Display labels configuration page, route, and nav entry
- [x] 2.4 Redirect or remove obsolete `/configuration/display-labels` deep links

## 3. Dorph Data file import

- [x] 3.1 Rename Data import → File import in nav, page titles, and related copy
- [x] 3.2 Limit File import entity types to Generic data, People, Assets, Activities
- [x] 3.3 Accept CSV, Excel, and JSON uploads; reject other types with a clear error
- [x] 3.4 Wire import handlers for People, Assets, and Activities if only Generic data exists today; keep routes redirecting `data-import` → `file-import` if path changes

## 4. SurveyX (former SheetX)

- [x] 4.1 Rebrand Dorph Utils module label SheetX → SurveyX (keep internal id/env URLs if needed)
- [x] 4.2 Remove Forms from SurveyX/formx primary navigation and default entry points
- [x] 4.3 Rename AI Sheets → AI Surveys in nav and primary UI chrome
- [x] 4.4 Soft-redirect Forms deep links to the AI Surveys experience

## 5. Booki slim nav

- [x] 5.1 Rename Data AI → Booki AI in Booki nav and page chrome
- [x] 5.2 Keep only Booki AI, Bookings, and Flow log in primary nav (desktop and mobile)
- [x] 5.3 Remove Accounting, Warehouse, Assets, and Settings routes/nav from Booki; redirect removed paths to Booki AI home

## 6. Academi nav

- [x] 6.1 Rename Chat → Academi AI and Community → Board in Academi primary UI
- [x] 6.2 Remove Profile from Academi primary navigation and related entry points

## 7. Dorph Utils shared Settings

- [x] 7.1 Add shell-level Settings at Dorph Utils app level (same level as modules / More menu)
- [x] 7.2 Implement a thin Settings page (account/session essentials; embed UsersPanel prefs only if already shared)
- [x] 7.3 Remove Settings/Profile primary nav from SurveyX, Booki, Academi, and other Utils embeds covered by the shared shell Settings

## 8. Dorph AI — remove task chains

- [x] 8.1 Remove Task chains button, modal mount, and due-chain polling from Dorph AI chat
- [x] 8.2 Delete unused task-chain frontend modules/assets that are no longer referenced

## 9. Verification

- [x] 9.1 Smoke Dorph Data drawer: Big notes nesting, File import entity/formats, no Display labels
- [x] 9.2 Smoke Dorph Utils: SurveyX / Booki / Academi labels and removals; shared Settings reachable
- [x] 9.3 Smoke Dorph AI: no Task chains; login persists without timeout and clears on Sign out
