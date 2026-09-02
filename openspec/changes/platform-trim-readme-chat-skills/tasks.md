## 1. README and one config

- [x] 1.1 Rewrite root README folders, URLs, env, and AI-provider sections for remaining apps only (Morph AI / MorphNotes, MorphUtils modules, Event Logs, Content Maker, Data Access, Project, AI tools); document single root `.env`
- [x] 1.2 Remove Booki and Academi from `start-all.sh` default services, aliases, install, and list

## 2. Delete platform-chat

- [x] 2.1 Copy `aiProgress` into Event Logs / Content Maker / Data Access as needed; unmount PlatformChatDrawer and drop `@robo/platform-chat` from those apps and Project
- [x] 2.2 Delete the `platform-chat/` package and remaining workspace references

## 3. Remove Video stories

- [x] 3.1 Remove AI tools Video stories nav, page, routes, and frontend API helpers
- [x] 3.2 Remove AI tools backend video-story routes and services (redirect or 404 `/video-stories`)

## 4. Morph AI chrome

- [x] 4.1 Remove Morph AI header Notes & TODOs, Context & Knowledge, and Export session (JSON) buttons
- [x] 4.2 Remove AI tools Open in tab (including error-state new-tab link)

## 5. Skills editor

- [x] 5.1 Skills modal: dark, width ~1100px, left upload / right scrolling catalog, tall instructions field, `.md` file fills name + instructions
- [x] 5.2 Add Improve with AI (`POST /api/skills/improve`) that drafts name, description, and instructions without saving until Upload

## 6. Verify

- [x] 6.1 README/`start-all.sh list` match remaining apps; Event Logs has no platform-chat drawer; AI tools has no Video stories; Morph AI header and Skills layout match the specs
- [x] 6.2 Browser: Morph AI header, Skills split + md + Improve with AI (draft only), AI tools drawer without Open in tab
