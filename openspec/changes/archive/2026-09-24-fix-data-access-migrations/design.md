## Context

See proposal.md for why Render fails the deploy. The runner in `SharpReport/backend/src/db/migrations.rs` calls `fs::read_dir("migrations")` for both SQLite and Postgres. Nothing changes the process directory. The image sets `WORKDIR /app` and the entrypoint execs `/app/datapulse` without `cd`. The runtime stage copies `/app/config` and does not copy `SharpReport/backend/migrations`. That directory exists in the repo (`0001` through `0005`, `CREATE TABLE IF NOT EXISTS` and `INSERT ... WHERE NOT EXISTS`). `.dockerignore` does not exclude it. The disk mount is `/data` only.

## Goals / Non-Goals

**Goals:**

- The image contains those SQL files at `/app/migrations`.
- The container contract fails if that copy or the source SQL is missing.
- Startup still uses the relative directory the local binary already uses.

**Non-Goals:**

- A new migrations env var, a sqlx embed, or a rewrite of the runner.
- Render API calls, Blueprint edits, or a committed `USERS_PANEL_BASE_URL`.
- Changing other product images.

## Decisions

Copy `SharpReport/backend/migrations` from the build context to `/app/migrations` in the runtime stage, next to the existing config copy. The last `WORKDIR` stays `/app`. The contract requires that exact `COPY` line (not a comment), a last `WORKDIR /app`, at least one `SharpReport/backend/migrations/*.sql` file, and no dockerignore rule that drops that tree or `*.sql`.

Rejected alternatives:

- Copy from the Rust build stage (`COPY --from=build /src/SharpReport/backend/migrations`). The files are there only because the build stage copies the whole backend. A later build-stage trim would remove them while the runtime copy still looked valid until the path moved. A context `COPY` fails the image build as soon as the directory is gone.
- Teach the binary an absolute path or `SHARPREPORT_MIGRATIONS_DIR`. The files still have to be in the image, and an unset env on Render reproduces the panic. Local `cargo run` from `SharpReport/backend` already uses `./migrations`.
- Embed the SQL with `sqlx::migrate!`. That changes the runner for a missing copy. The existing files are already safe to run again.
- Store the SQL on the `/data` disk. A new disk is empty, so the first boot still panics, and a disk snapshot would not carry the image's schema files.

## Risks / Trade-offs

- [Contract matches a comment that mentions the copy] → Require a Dockerfile line that is the `COPY` instruction, and require the last `WORKDIR` to be `/app`.
- [A later `.dockerignore` drops `*.sql`] → The contract fails on an ignore rule for the migrations tree or `*.sql`.
- [Restart re-runs every file] → The committed SQL is idempotent (`IF NOT EXISTS`, `WHERE NOT EXISTS`). Do not change the runner in this hotfix.
- [Root-owned files, process uid 65532] → Default `COPY` mode is world-readable. Do not `chmod` the directory to owner-only.
- [Migrations placed only on `/data`] → Rejected above. Schema SQL stays in the image.
