# Storage inventory (task 5.x)

| App | Runtime store | Notes |
|-----|---------------|-------|
| Morph | SQLite + Badger + go-cache | Default embedded; `cmd/migrate_embedded` may still import MySQL/Mongo drivers for optional one-time migration from backups |
| ComposerX | SQLite + Badger | Migrated |
| FormsX / SheetX | SQLite + Badger | Migrated |
| Booki | SQLite + go-cache | Migrated (refresh tokens in-memory) |
| Morph Engi | SQLite | Already embedded (`DATABASE_URL=sqlite://…`); MySQL URL warned and remapped |
| SharpReport / DataX | SQLite app DB | Own `DATABASE_URL=sqlite://./data/datapulse.db`; MySQL client code is for querying *user* databases, not platform runtime |
| Academi / BK | N/A / own stacks | No Morph MySQL/Mongo/Redis requirement for Utils shell |

Homebrew MySQL, MongoDB, Redis uninstalled for this machine as task 1.1.
