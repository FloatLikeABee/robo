## 1. Morph Utils Docs + Projects branding

- [x] 1.1 Rename Utils module Academi → Docs (`config.ts` id/label/shortLabel/description/icon paths as needed)
- [x] 1.2 Update Docs embed chrome brand strings (Academi → Docs) in academi web header
- [x] 1.3 Rename Utils/Projects chrome Engi → Projects where still shown as Engi
- [x] 1.4 Update Morph Utils Settings / login copy that still lists Academi or Booki incorrectly

## 2. Flow log into Projects; remove Booki

- [x] 2.1 Add Flow log section/tab to morph-engi (Projects) UI
- [x] 2.2 Wire Flow log create/list/filter/delete to Booki `/api/v1/flow-log/*` (proxy or base URL config)
- [x] 2.3 Make Flow log category a free-text input (optional suggestions only; any typed value allowed)
- [x] 2.4 Remove Booki module from Morph Utils nav/config and start/embed wiring
- [x] 2.5 Smoke: Projects Flow log round-trip; Utils no longer opens Booki

## 3. Morph Data Tasks (no Activities, no assignee)

- [x] 3.1 Remove Activities from Morph Data drawer, routes, and File import entity options
- [x] 3.2 Remove assignee picker/requirement from CaseTasks create/edit UI
- [x] 3.3 Relax/remove API validation that requires assignees on task create/update
- [x] 3.4 Confirm tasks save with description/JSON only for assignment-like detail

## 4. Morph Data Resources (People + Assets)

- [x] 4.1 Replace People and Assets drawer items with one combined Resources module entry
- [x] 4.2 Build combined list/detail page with People | Assets filter or tabs reusing existing grids
- [x] 4.3 Update File import entity targeting for the combined resources model
- [x] 4.4 Redirect or remove old `/people` and `/assets` top-level nav paths as needed

## 5. Morph Data Stories

- [x] 5.1 Rename Story Board → Stories in drawer and routes/labels
- [x] 5.2 Redesign Stories list as task-like card/tile grid
- [x] 5.3 Open story detail on card click (panel/page patterned after Tasks)
- [x] 5.4 Rename comments → notes in Stories UI (and API field/route aliases if required)

## 6. Verification

- [x] 6.1 Smoke Morph Utils: Docs, Projects (+ Flow log), no Booki/Academi/Engi primary labels
- [x] 6.2 Smoke Morph Data: Resources, Tasks (no assignee), Stories tiles/notes; no Activities
