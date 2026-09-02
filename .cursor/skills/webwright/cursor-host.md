# Cursor host

This copy of Webwright runs inside Cursor, not Claude Code. Same workspace
contract (`plan.md`, `final_runs/run_<id>/`, instrumented `final_script.py`,
screenshots, action log). Tool names differ.

## Tool map

| Webwright / Claude Code | Cursor |
| --- | --- |
| `Bash` / `bash` | `Shell` |
| `Read` / `read_file` | `Read` (including PNGs) |
| `Write` / `write_file` | `Write` |
| `Edit` | `StrReplace` |

One shell command per step. Observe stdout/stderr before the next command.
Do not wrap commands in JSON.

## Autotesting this repo

Use this skill when the user asks to autotest, acceptance-test, or verify a
web UI flow with reusable Playwright evidence (not a single screenshot).

- `WORKSPACE_DIR` = repo-root `webwright-runs/<task_id>/` (gitignored).
- Launch **Firefox** headless as in `reference/playwright_patterns.md`.
- Python: `python3.12` (Playwright is installed for 3.12). Do not `pip install`
  extra packages during a run.
- Prefer driving the live local app over guessing from source. Start the
  matching frontend if it is not already listening.
- Cursor browser MCP is not a substitute for this skill’s run artifacts.
- MorphNotes `singleSession` chat has no Files pane; do not treat that as a
  product bug while autotesting Morph AI Files.

## Python

```bash
python3.12 -m playwright --version
python3.12 -m playwright install firefox   # one-time, if Firefox is missing
```
