## 1. Data Access health banner

- [x] 1.1 Stop `startBackendHealthMonitor` in SharpReport `+layout.svelte` and remove `BackendStatusBanner` from that layout
- [x] 1.2 Confirm Data Access no longer shows “DataX API offline” / “health check failed”; leave per-request errors in `api.ts`

## 2. Morph Utils title intros

- [x] 2.1 Remove the Events & Info lede (“Operational notes…”) under the h1 in `EventsInfo.tsx`
- [x] 2.2 Remove the Info Sheets lede under the h1 in `SurveyBot.tsx`
- [x] 2.3 Remove Data Access page intros (data-tables, docs layout, and any sibling h1+lede on Data Access routes)
- [x] 2.4 Sweep Content Maker and Project module pages for the same under-title intro pattern and remove those lines

## 3. Events & Info search and delete

- [x] 3.1 Add optional `q` title filter to `GET /events-info` and `EventInfoRepo.List`
- [x] 3.2 Add authenticated batch-delete (`POST /events-info/batch-delete` with ids) reusing per-id delete
- [x] 3.3 Events & Info list: title search field, row delete with confirm, checkboxes + batch delete with confirm
- [x] 3.4 Cover list `q` and batch-delete with handler/repo tests

## 4. MorphNotes product and nav

- [x] 4.1 Default `product_name` and user-visible Morph Data / MorphData chrome to MorphNotes (drawer, document title, Morph AI app link, landing)
- [x] 4.2 Remove Assets from `AppDrawer`; default MorphNotes index to generic-data; redirect former assets/people/resources paths to generic-data
- [x] 4.3 Relabel Configuration → Settings in the drawer
- [x] 4.4 Remove User Settings header button and route (redirect old path); order header Notes, Morph AI, theme

## 5. Softer light theme

- [x] 5.1 Mute MorphNotes light `background.default` / `paper` (and related table/header surfaces) off pure white
- [x] 5.2 Mute Morph AI and platform-chat light tokens (`--chat-page-bg`, surfaces, message/input) off pure white
- [x] 5.3 Mute Event Logs, Data Access, Content Maker, and Project light page canvases the same way

## 6. Verify

- [x] 6.1 Data Access in Morph Utils: no health banner; a normal page load works
- [x] 6.2 Events & Info: no intro lede; search by title; delete one; batch delete
- [x] 6.3 MorphNotes: name, no Assets, Settings label, header Notes–Morph AI–theme, no User Settings; light mode is not stark white
