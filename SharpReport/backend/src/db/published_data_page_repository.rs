use crate::db::Database;
use chrono::{DateTime, Utc};
use sqlx::FromRow;
use uuid::Uuid;

#[derive(Debug, Clone, FromRow)]
pub struct PublishedDataPageRow {
    pub id: Uuid,
    pub data_table_id: Uuid,
    pub name: String,
    pub slug: String,
    pub theme: String,
    pub html_content: String,
    pub created_by: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug)]
pub struct PublishedDataPageRepository {
    db: Database,
}

pub fn slugify_publish_name(name: &str) -> String {
    let s = name.trim().to_lowercase();
    let mut out = String::new();
    let mut prev_hyphen = false;
    for ch in s.chars() {
        if ch.is_ascii_alphanumeric() {
            out.push(ch);
            prev_hyphen = false;
        } else if !prev_hyphen && !out.is_empty() {
            out.push('-');
            prev_hyphen = true;
        }
    }
    let trimmed = out.trim_matches('-').to_string();
    if trimmed.is_empty() {
        "page".to_string()
    } else {
        trimmed
    }
}

impl PublishedDataPageRepository {
    pub fn new(db: Database) -> Self {
        Self { db }
    }

    pub async fn resolve_unique_slug(&self, name: &str) -> Result<String, sqlx::Error> {
        let base = slugify_publish_name(name);
        self.next_unique_slug(&base).await
    }

    async fn next_unique_slug(&self, base: &str) -> Result<String, sqlx::Error> {
        let base = if base.trim().is_empty() { "page" } else { base };
        let mut candidate = base.to_string();
        for i in 2..=5000 {
            let exists = self.slug_exists(&candidate).await?;
            if !exists {
                return Ok(candidate);
            }
            candidate = format!("{}-{}", base, i);
        }
        Err(sqlx::Error::Configuration(
            "could not find available publish slug".into(),
        ))
    }

    async fn slug_exists(&self, slug: &str) -> Result<bool, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                let row: Option<(i32,)> =
                    sqlx::query_as("SELECT 1 FROM data_page_publishes WHERE slug = $1 LIMIT 1")
                        .bind(slug)
                        .fetch_optional(pool)
                        .await?;
                Ok(row.is_some())
            }
            Database::Sqlite(pool) => {
                let row: Option<(i32,)> =
                    sqlx::query_as("SELECT 1 FROM data_page_publishes WHERE slug = ?1 LIMIT 1")
                        .bind(slug)
                        .fetch_optional(pool)
                        .await?;
                Ok(row.is_some())
            }
        }
    }

    async fn next_unique_name(&self, base: &str) -> Result<String, sqlx::Error> {
        let base = base.trim();
        let base = if base.is_empty() {
            "Untitled page"
        } else {
            base
        };
        let mut candidate = base.to_string();
        for i in 2..=5000 {
            let exists = self.name_exists(&candidate).await?;
            if !exists {
                return Ok(candidate);
            }
            candidate = format!("{} ({})", base, i);
        }
        Err(sqlx::Error::Configuration(
            "could not find available publish name".into(),
        ))
    }

    async fn name_exists(&self, name: &str) -> Result<bool, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                let row: Option<(i32,)> =
                    sqlx::query_as("SELECT 1 FROM data_page_publishes WHERE name = $1 LIMIT 1")
                        .bind(name)
                        .fetch_optional(pool)
                        .await?;
                Ok(row.is_some())
            }
            Database::Sqlite(pool) => {
                let row: Option<(i32,)> =
                    sqlx::query_as("SELECT 1 FROM data_page_publishes WHERE name = ?1 LIMIT 1")
                        .bind(name)
                        .fetch_optional(pool)
                        .await?;
                Ok(row.is_some())
            }
        }
    }

    pub async fn create(
        &self,
        data_table_id: Uuid,
        name: &str,
        theme: &str,
        html_content: &str,
        created_by: Option<&str>,
    ) -> Result<PublishedDataPageRow, sqlx::Error> {
        let name_trim = name.trim();
        let html_trim = html_content.trim();
        if name_trim.is_empty() || html_trim.is_empty() {
            return Err(sqlx::Error::Configuration("name and html required".into()));
        }
        let theme = theme.trim();
        let theme = if theme.is_empty() { "light" } else { theme };

        let unique_name = self.next_unique_name(name_trim).await?;
        let slug = self.resolve_unique_slug(&unique_name).await?;
        let id = Uuid::new_v4();
        let now = Utc::now();

        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, PublishedDataPageRow>(
                    r#"INSERT INTO data_page_publishes (id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at)
                       VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                       RETURNING id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at"#,
                )
                .bind(id)
                .bind(data_table_id)
                .bind(&unique_name)
                .bind(&slug)
                .bind(theme)
                .bind(html_trim)
                .bind(created_by)
                .bind(now)
                .bind(now)
                .fetch_one(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, PublishedDataPageRow>(
                    r#"INSERT INTO data_page_publishes (id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at)
                       VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)
                       RETURNING id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at"#,
                )
                .bind(id)
                .bind(data_table_id)
                .bind(&unique_name)
                .bind(&slug)
                .bind(theme)
                .bind(html_trim)
                .bind(created_by)
                .bind(now)
                .bind(now)
                .fetch_one(pool)
                .await
            }
        }
    }

    pub async fn find_by_slug(
        &self,
        slug: &str,
    ) -> Result<Option<PublishedDataPageRow>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, PublishedDataPageRow>(
                    "SELECT id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at FROM data_page_publishes WHERE slug = $1",
                )
                .bind(slug)
                .fetch_optional(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, PublishedDataPageRow>(
                    "SELECT id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at FROM data_page_publishes WHERE slug = ?1",
                )
                .bind(slug)
                .fetch_optional(pool)
                .await
            }
        }
    }

    pub async fn list_by_table(
        &self,
        data_table_id: Uuid,
        limit: i32,
        offset: i32,
    ) -> Result<Vec<PublishedDataPageRow>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, PublishedDataPageRow>(
                    "SELECT id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at FROM data_page_publishes WHERE data_table_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
                )
                .bind(data_table_id)
                .bind(limit)
                .bind(offset)
                .fetch_all(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, PublishedDataPageRow>(
                    "SELECT id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at FROM data_page_publishes WHERE data_table_id = ?1 ORDER BY created_at DESC LIMIT ?2 OFFSET ?3",
                )
                .bind(data_table_id)
                .bind(limit)
                .bind(offset)
                .fetch_all(pool)
                .await
            }
        }
    }

    pub async fn count_by_table(&self, data_table_id: Uuid) -> Result<i64, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                let row: (i64,) = sqlx::query_as(
                    "SELECT COUNT(*) FROM data_page_publishes WHERE data_table_id = $1",
                )
                .bind(data_table_id)
                .fetch_one(pool)
                .await?;
                Ok(row.0)
            }
            Database::Sqlite(pool) => {
                let row: (i64,) = sqlx::query_as(
                    "SELECT COUNT(*) FROM data_page_publishes WHERE data_table_id = ?1",
                )
                .bind(data_table_id)
                .fetch_one(pool)
                .await?;
                Ok(row.0)
            }
        }
    }

    pub async fn list_recent_all(
        &self,
        limit: i32,
        offset: i32,
    ) -> Result<Vec<PublishedDataPageRow>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, PublishedDataPageRow>(
                    "SELECT id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at FROM data_page_publishes ORDER BY created_at DESC LIMIT $1 OFFSET $2",
                )
                .bind(limit)
                .bind(offset)
                .fetch_all(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, PublishedDataPageRow>(
                    "SELECT id, data_table_id, name, slug, theme, html_content, created_by, created_at, updated_at FROM data_page_publishes ORDER BY created_at DESC LIMIT ?1 OFFSET ?2",
                )
                .bind(limit)
                .bind(offset)
                .fetch_all(pool)
                .await
            }
        }
    }

    pub async fn count_all(&self) -> Result<i64, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                let row: (i64,) =
                    sqlx::query_as("SELECT COUNT(*) FROM data_page_publishes")
                        .fetch_one(pool)
                        .await?;
                Ok(row.0)
            }
            Database::Sqlite(pool) => {
                let row: (i64,) =
                    sqlx::query_as("SELECT COUNT(*) FROM data_page_publishes")
                        .fetch_one(pool)
                        .await?;
                Ok(row.0)
            }
        }
    }
}
