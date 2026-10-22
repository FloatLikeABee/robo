# Tasks

## 1. Unwrap backticks and never save the envelope

- [x] 1.1 Add a failing fixture from `big2.json` (backtick-wrapped markdown) and tests: parse yields title + inner markdown; generate helper does not store the envelope and uses the idea as title when parse fails. Verify `go test` in `dorph/handlers` fails.
- [x] 1.2 Implement backtick unwrap and envelope fallback. Verify the new tests pass.

## 2. Timeout, H2 count, grill markdown

- [x] 2.1 Add failing tests: generate timeout > 90s and deadline maps to a retry error; a 2-`##` + `###` note is expanded to ≥3 `##`; write prompt has no hex/`html-anything`; sanitizer flattens the tarball hex table and keeps `30k RMB`; compact skill still ≤512 tokens. Verify they fail.
- [x] 2.2 Implement timeout wrap, `countH2`, colour-free write prompt, hex-table sanitize. Verify `go test` in `dorph/handlers`, `dorph/grill`, and `pkg/dorphai` pass.

## Workflow follow-up

- Archive the change after implementation and verification.
- Open a PR that says Closes #155. Do not merge.
