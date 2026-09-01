## 1. Remove MorphNotes File import

- [x] 1.1 Drop `POST /api/tran/generic-data/import` (`ImportGenericData`), `tranEndpoints.genericDataImport`, and the management-chat mention of that path. Keep extract/CRUD. Add or update a handler test that import is not registered / create still works via extract+create.
- [x] 1.2 Remove Settings File import: `DataImport.js`, AppDrawer item, `configuration/file-import` and `configuration/data-import` routes (redirect those URLs to a remaining MorphNotes page). Confirm Generic data page still extracts.

## 2. Task start and end dates

- [x] 2.1 Backend: create/update CaseTask reject missing or unparseable `start_at` / `end_at` (400). Tests for missing start, missing end, and both present.
- [x] 2.2 Frontend: mark Start and End required; new-task draft defaults both to today’s local calendar day (start 00:00 / end 23:59 or date-only). Block submit when empty. Fill today when AI draft omits dates.

## 3. MorphNotes preview scrollbars

- [x] 3.1 Timelines HTML: outer pane `overflow: hidden`; iframe fills remaining height so only the HTML document scrolls; dark `color-scheme` / scrollbar styles on the inner surface.
- [x] 3.2 Big Notes Preview and HTML tabs: same outer-hidden / inner-dark-scroll treatment.

## 4. MorphUtils no logout

- [x] 4.1 Remove MorphUtils shell Sign out (do not clear shared token from chrome). Keep Morph AI Sign out.
- [x] 4.2 Remove Event Logs Logout, Content Maker Out, Data Access Sign out, and Project Sign out.

## 5. Project themed scrollbars

- [x] 5.1 Theme MorphUtils Project Markdown and HTML preview vertical scrollbars dark (`scrollbar-color` / webkit + iframe `color-scheme` / srcDoc scrollbar CSS as needed).

## 6. Verify

- [x] 6.1 Static: File import strings/routes/import POST gone; CaseTask tests pass; logout/Out/Sign out absent from MorphUtils shell and four modules (Morph AI Sign out still present).
- [x] 6.2 Browser: Settings has no File import; new Task dates default today and required; Timelines HTML and Big Notes preview one dark inner scrollbar; MorphUtils modules have no Out/Sign out; Project markdown/HTML scrollbars dark.
