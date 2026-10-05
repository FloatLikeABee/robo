## 1. Fail-closed agent-notes API

- [x] 1.1 Add a failing handler test: logged-out list and get are 401 and omit the seeded title; `X-User-ID` alone is 401; user B cannot list or get user A's agent note; `?user_id=` does not switch owner; no Tran user and two active Tran users return 409 and omit user id 1's note; a human note and a todo stay out; a title-only or body-only marker still lists.
- [x] 1.2 Implement `GET /api/tran/agent-notes` and `GET /api/tran/agent-notes/:id` from the session subject only, and register them beside the existing notes routes.

## 2. Agent notes page

- [x] 2.1 Add a failing frontend test: with no token, `/agent-notes` redirects to login and does not call the notes API; with a token, the list shows title, time, and status, the detail shows the body, and the empty state explains how agents post notes.
- [x] 2.2 Add the protected `/agent-notes` page, phone and desktop styles (safe area, wrap, 44px back control), and a same-tab "Agent notes" entry in the Morph AI header menu.

## 3. Verification

- [x] 3.1 Run `go build`, `go vet`, and `go test -race -count=1` in `morph`, and `CI=true npm test -- --watchAll=false` plus `CI=true npm run build` in `morph/frontend`.
