## 1. Publish once (API)

- [x] 1.1 Change `published_pages` create so a colliding slug updates the existing row (HTML, theme, timestamps) instead of minting `name-2`
- [x] 1.2 Ensure `POST /publishes` returns the same `slug` / `public_path` on republish of the same name
- [x] 1.3 Add or keep a unique slug constraint so two concurrent first publishes cannot insert two rows

## 2. Unified Published contents table (UI)

- [x] 2.1 Replace the two Published contents tables with one list merged by slugified name (published wins when both exist)
- [x] 2.2 Keep View/Delete for saved-only rows and Open for published rows; one row must not appear twice
- [x] 2.3 Remove the under-title lede; keep heading **Published contents**

## 3. Action alignment

- [x] 3.1 Vertically center Published contents row text and operation buttons (middle-align cells; inline-flex actions)

## 4. Verify

- [x] 4.1 Confirm Publish twice with the same name keeps one public path
- [x] 4.2 Confirm Published contents is one table, no lede, buttons aligned with row text
