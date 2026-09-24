## 1. Failing handler tests

- [x] 1.1 Add `morph/handlers/tran_users_test.go` that mounts `AuthzMiddleware` and the user routes. Cover no-session 401, non-admin 403, and admin success for `POST /api/tran/users`, `PUT /api/tran/users/:id`, and `DELETE /api/tran/users/:id`.
- [x] 1.2 Cover self-promotion (`PUT` own id with `administrator` true is 403 and the flag stays false), a spoofed `X-User-Role` plus Tran `Administrator=1` still 403, and a demoted plat user whose old token still says admin is 403.
- [x] 1.3 Cover `PUT /api/tran/users/me`: an email plus last name leaves email unchanged, an email-only body is 400 and writes nothing, and administrator/deactivated on `/me` do not stick.
- [x] 1.4 Cover the takeover sequence (A cannot set B's email or deactivate A's row; A's resolved id stays A's) and ambiguous email/LoginID (409, no new row, notes id is neither match). Keep single-match, deactivated-duplicate, and no-match auto-create.

## 2. Implementation

- [x] 2.1 Call `requireAuthDB` and `requireAdmin` from create, update, and delete. Remove `email` from `allowedTranUserSelfWrite`.
- [x] 2.2 Replace the unordered `LIMIT 1` lookups with a fail-closed active-row match. Map the ambiguous error to 409 on the profile routes and to user id 0 (no fallthrough) in `tranUserIDFromContext`.
- [x] 2.3 Add nullable `User.DeactivatedDate` to the SQLite schema and the startup column pass so admin deactivate succeeds.
- [x] 2.4 Make the profile email field read-only and stop sending `email` from `UserSettings.js`.

## 3. Verification

- [x] 3.1 Run `go test` for the touched Morph packages and the Morph frontend production build. Confirm the new tests failed before the implementation and pass after it.
