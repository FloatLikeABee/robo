use sqlx::{Pool, Postgres, Sqlite};
use std::fs;

pub async fn run_postgres_migrations(pool: &Pool<Postgres>) -> Result<(), sqlx::Error> {
    run_migrations_postgres(pool).await
}

pub async fn run_sqlite_migrations(pool: &Pool<Sqlite>) -> Result<(), sqlx::Error> {
    run_migrations_sqlite(pool).await
}

async fn run_migrations_postgres(pool: &Pool<Postgres>) -> Result<(), sqlx::Error> {
    let migrations_dir = "migrations";
    let entries = fs::read_dir(migrations_dir).map_err(|e| {
        sqlx::Error::Configuration(format!("Failed to read migrations directory: {}", e).into())
    })?;

    let mut migrations: Vec<_> = entries
        .filter_map(|entry| {
            let entry = entry.ok()?;
            let path = entry.path();
            if path.extension()? == "sql" {
                Some(path)
            } else {
                None
            }
        })
        .collect();

    migrations.sort();

    for migration in migrations {
        let content = fs::read_to_string(&migration).map_err(|e| {
            sqlx::Error::Configuration(
                format!(
                    "Failed to read migration file {}: {}",
                    migration.display(),
                    e
                )
                .into(),
            )
        })?;

        sqlx::query(&content).execute(pool).await?;

        println!("Applied migration: {}", migration.display());
    }

    Ok(())
}

async fn run_migrations_sqlite(pool: &Pool<Sqlite>) -> Result<(), sqlx::Error> {
    let migrations_dir = "migrations";
    let entries = fs::read_dir(migrations_dir).map_err(|e| {
        sqlx::Error::Configuration(format!("Failed to read migrations directory: {}", e).into())
    })?;

    let mut migrations: Vec<_> = entries
        .filter_map(|entry| {
            let entry = entry.ok()?;
            let path = entry.path();
            if path.extension()? == "sql" {
                Some(path)
            } else {
                None
            }
        })
        .collect();

    migrations.sort();

    for migration in migrations {
        let content = fs::read_to_string(&migration).map_err(|e| {
            sqlx::Error::Configuration(
                format!(
                    "Failed to read migration file {}: {}",
                    migration.display(),
                    e
                )
                .into(),
            )
        })?;

        // Convert PostgreSQL-specific syntax to SQLite.
        // Order matters: replace `gen_random_uuid()` before `UUID` so we don't mangle the function name.
        let content = content
            .replace("gen_random_uuid()", "(lower(hex(randomblob(16))))")
            .replace("TIMESTAMPTZ", "TEXT")
            .replace("INET", "TEXT")
            .replace("JSONB", "TEXT")
            .replace("UUID", "TEXT")
            .replace("NOW()", "CURRENT_TIMESTAMP");

        sqlx::query(&content).execute(pool).await?;

        println!("Applied migration: {}", migration.display());
    }

    Ok(())
}
