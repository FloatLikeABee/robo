pub mod data_page_build_repository;
pub mod data_table_repository;
pub mod migrations;
pub mod models;
pub mod published_data_page_repository;
pub mod repositories;

use std::fs;
use std::path::Path;

use sqlx::sqlite::SqliteConnectOptions;
use sqlx::{FromRow, Pool, Postgres, Sqlite};
use std::str::FromStr;

/// SQLite only creates the file, not parent directories — ensure they exist.
fn ensure_sqlite_parent_dir(database_url: &str) -> Result<(), sqlx::Error> {
    let lower = database_url.to_lowercase();
    if !lower.starts_with("sqlite") {
        return Ok(());
    }
    if lower.contains(":memory:") || lower.contains("mode=memory") {
        return Ok(());
    }
    let path_str = database_url
        .strip_prefix("sqlite://")
        .or_else(|| database_url.strip_prefix("sqlite:"))
        .unwrap_or(database_url);
    let path_str = path_str.split('?').next().unwrap_or(path_str);
    if path_str.is_empty() {
        return Ok(());
    }
    let path = Path::new(path_str);
    if let Some(parent) = path.parent() {
        if !parent.as_os_str().is_empty() {
            fs::create_dir_all(parent)?;
        }
    }
    Ok(())
}

#[derive(Debug, Clone)]
pub enum Database {
    Postgres(Pool<Postgres>),
    Sqlite(Pool<Sqlite>),
}

impl Database {
    pub async fn new(database_url: &str) -> Result<Self, sqlx::Error> {
        if database_url.starts_with("postgres://") || database_url.starts_with("postgresql://") {
            let pool = Pool::<Postgres>::connect(database_url).await?;
            Ok(Self::Postgres(pool))
        } else {
            ensure_sqlite_parent_dir(database_url)?;
            // Default URL parse uses READWRITE without CREATE — new files fail with SQLite error 14.
            let options = SqliteConnectOptions::from_str(database_url)
                .map_err(|e| sqlx::Error::Configuration(e.into()))?
                .create_if_missing(true);
            let pool = Pool::<Sqlite>::connect_with(options).await?;
            Ok(Self::Sqlite(pool))
        }
    }

    pub async fn execute(&self, query: &str) -> Result<u64, sqlx::Error> {
        match self {
            Database::Postgres(pool) => {
                let result = sqlx::query(query).execute(pool).await?;
                Ok(result.rows_affected())
            }
            Database::Sqlite(pool) => {
                let result = sqlx::query(query).execute(pool).await?;
                Ok(result.rows_affected())
            }
        }
    }

    pub async fn fetch_one<'a, T>(&'a self, query: &'a str) -> Result<T, sqlx::Error>
    where
        T: for<'r> FromRow<'r, sqlx::postgres::PgRow>
            + for<'r> FromRow<'r, sqlx::sqlite::SqliteRow>
            + Unpin
            + Send,
    {
        match self {
            Database::Postgres(pool) => sqlx::query_as::<_, T>(query).fetch_one(pool).await,
            Database::Sqlite(pool) => sqlx::query_as::<_, T>(query).fetch_one(pool).await,
        }
    }

    pub async fn fetch_all<'a, T>(&'a self, query: &'a str) -> Result<Vec<T>, sqlx::Error>
    where
        T: for<'r> FromRow<'r, sqlx::postgres::PgRow>
            + for<'r> FromRow<'r, sqlx::sqlite::SqliteRow>
            + Unpin
            + Send,
    {
        match self {
            Database::Postgres(pool) => sqlx::query_as::<_, T>(query).fetch_all(pool).await,
            Database::Sqlite(pool) => sqlx::query_as::<_, T>(query).fetch_all(pool).await,
        }
    }
}

pub async fn initialize_db(database_url: &str) -> Result<Database, sqlx::Error> {
    let db = Database::new(database_url).await?;

    match &db {
        Database::Postgres(pool) => {
            migrations::run_postgres_migrations(pool).await?;
        }
        Database::Sqlite(pool) => {
            migrations::run_sqlite_migrations(pool).await?;
        }
    }

    Ok(db)
}
