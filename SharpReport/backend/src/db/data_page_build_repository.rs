use chrono::{DateTime, Utc};
use sqlx::FromRow;
use uuid::Uuid;

use crate::db::Database;

pub const MAX_BUILDS_PER_TABLE: i32 = 5;

#[derive(Debug, Clone, FromRow)]
pub struct DataPageBuildRow {
    pub id: Uuid,
    pub data_table_id: Uuid,
    pub label: String,
    pub source: String,
    pub html_content: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug)]
pub struct DataPageBuildRepository {
    db: Database,
}

impl DataPageBuildRepository {
    pub fn new(db: Database) -> Self {
        Self { db }
    }

    pub async fn save_and_trim(
        &self,
        data_table_id: Uuid,
        html_content: &str,
        source: &str,
        label: &str,
    ) -> Result<DataPageBuildRow, sqlx::Error> {
        let html_trim = html_content.trim();
        if html_trim.is_empty() {
            return Err(sqlx::Error::Configuration("html required".into()));
        }
        let source = source.trim();
        let source = if source.is_empty() { "build" } else { source };
        let label = label.trim();
        let label = if label.is_empty() { "Build" } else { label };

        let id = Uuid::new_v4();
        let now = Utc::now();

        let row = match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, DataPageBuildRow>(
                    r#"INSERT INTO data_page_builds (id, data_table_id, label, source, html_content, created_at)
                       VALUES ($1, $2, $3, $4, $5, $6)
                       RETURNING id, data_table_id, label, source, html_content, created_at"#,
                )
                .bind(id)
                .bind(data_table_id)
                .bind(label)
                .bind(source)
                .bind(html_trim)
                .bind(now)
                .fetch_one(pool)
                .await?
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, DataPageBuildRow>(
                    r#"INSERT INTO data_page_builds (id, data_table_id, label, source, html_content, created_at)
                       VALUES (?1, ?2, ?3, ?4, ?5, ?6)
                       RETURNING id, data_table_id, label, source, html_content, created_at"#,
                )
                .bind(id)
                .bind(data_table_id)
                .bind(label)
                .bind(source)
                .bind(html_trim)
                .bind(now)
                .fetch_one(pool)
                .await?
            }
        };

        self.trim_old(data_table_id).await?;
        Ok(row)
    }

    async fn trim_old(&self, data_table_id: Uuid) -> Result<(), sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query(
                    r#"DELETE FROM data_page_builds
                       WHERE data_table_id = $1
                         AND id NOT IN (
                           SELECT id FROM data_page_builds
                           WHERE data_table_id = $1
                           ORDER BY created_at DESC
                           LIMIT $2
                         )"#,
                )
                .bind(data_table_id)
                .bind(MAX_BUILDS_PER_TABLE)
                .execute(pool)
                .await?;
            }
            Database::Sqlite(pool) => {
                sqlx::query(
                    r#"DELETE FROM data_page_builds
                       WHERE data_table_id = ?1
                         AND id NOT IN (
                           SELECT id FROM data_page_builds
                           WHERE data_table_id = ?1
                           ORDER BY created_at DESC
                           LIMIT ?2
                         )"#,
                )
                .bind(data_table_id)
                .bind(MAX_BUILDS_PER_TABLE)
                .execute(pool)
                .await?;
            }
        }
        Ok(())
    }

    pub async fn list_recent(
        &self,
        data_table_id: Uuid,
        limit: i32,
    ) -> Result<Vec<DataPageBuildRow>, sqlx::Error> {
        let limit = limit.clamp(1, MAX_BUILDS_PER_TABLE);
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, DataPageBuildRow>(
                    r#"SELECT id, data_table_id, label, source, html_content, created_at
                       FROM data_page_builds
                       WHERE data_table_id = $1
                       ORDER BY created_at DESC
                       LIMIT $2"#,
                )
                .bind(data_table_id)
                .bind(limit)
                .fetch_all(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, DataPageBuildRow>(
                    r#"SELECT id, data_table_id, label, source, html_content, created_at
                       FROM data_page_builds
                       WHERE data_table_id = ?1
                       ORDER BY created_at DESC
                       LIMIT ?2"#,
                )
                .bind(data_table_id)
                .bind(limit)
                .fetch_all(pool)
                .await
            }
        }
    }

    pub async fn find_by_id(
        &self,
        data_table_id: Uuid,
        build_id: Uuid,
    ) -> Result<Option<DataPageBuildRow>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, DataPageBuildRow>(
                    r#"SELECT id, data_table_id, label, source, html_content, created_at
                       FROM data_page_builds
                       WHERE data_table_id = $1 AND id = $2"#,
                )
                .bind(data_table_id)
                .bind(build_id)
                .fetch_optional(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, DataPageBuildRow>(
                    r#"SELECT id, data_table_id, label, source, html_content, created_at
                       FROM data_page_builds
                       WHERE data_table_id = ?1 AND id = ?2"#,
                )
                .bind(data_table_id)
                .bind(build_id)
                .fetch_optional(pool)
                .await
            }
        }
    }
}
