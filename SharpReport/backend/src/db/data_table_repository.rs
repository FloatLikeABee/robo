use crate::db::Database;
use chrono::{DateTime, Utc};
use serde_json::Value;
use sqlx::FromRow;
use uuid::Uuid;

#[derive(Debug, Clone, FromRow)]
pub struct DataTableRow {
    pub id: Uuid,
    pub name: String,
    pub source_filename: Option<String>,
    pub source_format: String,
    pub column_schema: String,
    pub row_count: i32,
    pub created_by: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, FromRow)]
pub struct DataTableDataRow {
    pub id: Uuid,
    pub table_id: Uuid,
    pub row_index: i32,
    pub data: String,
}

#[derive(Debug)]
pub struct DataTableRepository {
    db: Database,
}

impl DataTableRepository {
    pub fn new(db: Database) -> Self {
        Self { db }
    }

    pub async fn list_all(&self) -> Result<Vec<DataTableRow>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, DataTableRow>(
                    "SELECT id, name, source_filename, source_format, column_schema, row_count, created_by, created_at, updated_at FROM data_tables ORDER BY updated_at DESC",
                )
                .fetch_all(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, DataTableRow>(
                    "SELECT id, name, source_filename, source_format, column_schema, row_count, created_by, created_at, updated_at FROM data_tables ORDER BY updated_at DESC",
                )
                .fetch_all(pool)
                .await
            }
        }
    }

    pub async fn find_by_id(&self, id: Uuid) -> Result<Option<DataTableRow>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, DataTableRow>(
                    "SELECT id, name, source_filename, source_format, column_schema, row_count, created_by, created_at, updated_at FROM data_tables WHERE id = $1",
                )
                .bind(id)
                .fetch_optional(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, DataTableRow>(
                    "SELECT id, name, source_filename, source_format, column_schema, row_count, created_by, created_at, updated_at FROM data_tables WHERE id = ?1",
                )
                .bind(id)
                .fetch_optional(pool)
                .await
            }
        }
    }

    pub async fn insert_table(
        &self,
        id: Uuid,
        name: &str,
        source_filename: Option<&str>,
        source_format: &str,
        columns: &[String],
        created_by: Option<&str>,
    ) -> Result<DataTableRow, sqlx::Error> {
        let now = Utc::now();
        let schema_json = serde_json::to_string(columns).unwrap_or_else(|_| "[]".to_string());

        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, DataTableRow>(
                    r#"INSERT INTO data_tables (id, name, source_filename, source_format, column_schema, row_count, created_by, created_at, updated_at)
                       VALUES ($1, $2, $3, $4, $5, 0, $6, $7, $8)
                       RETURNING id, name, source_filename, source_format, column_schema, row_count, created_by, created_at, updated_at"#,
                )
                .bind(id)
                .bind(name)
                .bind(source_filename)
                .bind(source_format)
                .bind(&schema_json)
                .bind(created_by)
                .bind(now)
                .bind(now)
                .fetch_one(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query(
                    r#"INSERT INTO data_tables (id, name, source_filename, source_format, column_schema, row_count, created_by, created_at, updated_at)
                       VALUES (?1, ?2, ?3, ?4, ?5, 0, ?6, ?7, ?8)"#,
                )
                .bind(id)
                .bind(name)
                .bind(source_filename)
                .bind(source_format)
                .bind(&schema_json)
                .bind(created_by)
                .bind(now)
                .bind(now)
                .execute(pool)
                .await?;
                self.find_by_id(id)
                    .await?
                    .ok_or_else(|| sqlx::Error::RowNotFound)
            }
        }
    }

    pub async fn insert_rows(&self, table_id: Uuid, rows: &[Value]) -> Result<i32, sqlx::Error> {
        if rows.is_empty() {
            return Ok(0);
        }

        match &self.db {
            Database::Postgres(pool) => {
                let mut tx = pool.begin().await?;
                for (idx, row) in rows.iter().enumerate() {
                    let data_json = row.to_string();
                    sqlx::query(
                        "INSERT INTO data_table_rows (id, table_id, row_index, data) VALUES (gen_random_uuid(), $1, $2, $3)",
                    )
                    .bind(table_id)
                    .bind(idx as i32)
                    .bind(&data_json)
                    .execute(&mut *tx)
                    .await?;
                }
                let count = rows.len() as i32;
                sqlx::query("UPDATE data_tables SET row_count = $1, updated_at = $2 WHERE id = $3")
                    .bind(count)
                    .bind(Utc::now())
                    .bind(table_id)
                    .execute(&mut *tx)
                    .await?;
                tx.commit().await?;
                Ok(count)
            }
            Database::Sqlite(pool) => {
                let mut tx = pool.begin().await?;
                for (idx, row) in rows.iter().enumerate() {
                    let row_id = Uuid::new_v4();
                    let data_json = row.to_string();
                    sqlx::query(
                        "INSERT INTO data_table_rows (id, table_id, row_index, data) VALUES (?1, ?2, ?3, ?4)",
                    )
                    .bind(row_id)
                    .bind(table_id)
                    .bind(idx as i32)
                    .bind(&data_json)
                    .execute(&mut *tx)
                    .await?;
                }
                let count = rows.len() as i32;
                sqlx::query("UPDATE data_tables SET row_count = ?1, updated_at = ?2 WHERE id = ?3")
                    .bind(count)
                    .bind(Utc::now())
                    .bind(table_id)
                    .execute(&mut *tx)
                    .await?;
                tx.commit().await?;
                Ok(count)
            }
        }
    }

    pub async fn fetch_all_rows(
        &self,
        table_id: Uuid,
    ) -> Result<Vec<DataTableDataRow>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, DataTableDataRow>(
                    "SELECT id, table_id, row_index, data FROM data_table_rows WHERE table_id = $1 ORDER BY row_index ASC",
                )
                .bind(table_id)
                .fetch_all(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, DataTableDataRow>(
                    "SELECT id, table_id, row_index, data FROM data_table_rows WHERE table_id = ?1 ORDER BY row_index ASC",
                )
                .bind(table_id)
                .fetch_all(pool)
                .await
            }
        }
    }

    pub async fn fetch_rows(
        &self,
        table_id: Uuid,
        limit: i32,
        offset: i32,
    ) -> Result<Vec<DataTableDataRow>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query_as::<_, DataTableDataRow>(
                    "SELECT id, table_id, row_index, data FROM data_table_rows WHERE table_id = $1 ORDER BY row_index ASC LIMIT $2 OFFSET $3",
                )
                .bind(table_id)
                .bind(limit)
                .bind(offset)
                .fetch_all(pool)
                .await
            }
            Database::Sqlite(pool) => {
                sqlx::query_as::<_, DataTableDataRow>(
                    "SELECT id, table_id, row_index, data FROM data_table_rows WHERE table_id = ?1 ORDER BY row_index ASC LIMIT ?2 OFFSET ?3",
                )
                .bind(table_id)
                .bind(limit)
                .bind(offset)
                .fetch_all(pool)
                .await
            }
        }
    }

    pub async fn update_row_data(
        &self,
        table_id: Uuid,
        row_index: i32,
        data: &Value,
    ) -> Result<bool, sqlx::Error> {
        let data_json = data.to_string();
        let now = Utc::now();
        match &self.db {
            Database::Postgres(pool) => {
                let result = sqlx::query(
                    "UPDATE data_table_rows SET data = $1 WHERE table_id = $2 AND row_index = $3",
                )
                .bind(&data_json)
                .bind(table_id)
                .bind(row_index)
                .execute(pool)
                .await?;
                if result.rows_affected() > 0 {
                    sqlx::query("UPDATE data_tables SET updated_at = $1 WHERE id = $2")
                        .bind(now)
                        .bind(table_id)
                        .execute(pool)
                        .await?;
                }
                Ok(result.rows_affected() > 0)
            }
            Database::Sqlite(pool) => {
                let result = sqlx::query(
                    "UPDATE data_table_rows SET data = ?1 WHERE table_id = ?2 AND row_index = ?3",
                )
                .bind(&data_json)
                .bind(table_id)
                .bind(row_index)
                .execute(pool)
                .await?;
                if result.rows_affected() > 0 {
                    sqlx::query("UPDATE data_tables SET updated_at = ?1 WHERE id = ?2")
                        .bind(now)
                        .bind(table_id)
                        .execute(pool)
                        .await?;
                }
                Ok(result.rows_affected() > 0)
            }
        }
    }

    pub async fn delete_row(&self, table_id: Uuid, row_index: i32) -> Result<bool, sqlx::Error> {
        let now = Utc::now();
        match &self.db {
            Database::Postgres(pool) => {
                let mut tx = pool.begin().await?;
                let result = sqlx::query(
                    "DELETE FROM data_table_rows WHERE table_id = $1 AND row_index = $2",
                )
                .bind(table_id)
                .bind(row_index)
                .execute(&mut *tx)
                .await?;
                if result.rows_affected() == 0 {
                    tx.rollback().await?;
                    return Ok(false);
                }
                sqlx::query(
                    "UPDATE data_table_rows SET row_index = row_index - 1 WHERE table_id = $1 AND row_index > $2",
                )
                .bind(table_id)
                .bind(row_index)
                .execute(&mut *tx)
                .await?;
                sqlx::query(
                    "UPDATE data_tables SET row_count = row_count - 1, updated_at = $1 WHERE id = $2",
                )
                .bind(now)
                .bind(table_id)
                .execute(&mut *tx)
                .await?;
                tx.commit().await?;
                Ok(true)
            }
            Database::Sqlite(pool) => {
                let mut tx = pool.begin().await?;
                let result = sqlx::query(
                    "DELETE FROM data_table_rows WHERE table_id = ?1 AND row_index = ?2",
                )
                .bind(table_id)
                .bind(row_index)
                .execute(&mut *tx)
                .await?;
                if result.rows_affected() == 0 {
                    tx.rollback().await?;
                    return Ok(false);
                }
                sqlx::query(
                    "UPDATE data_table_rows SET row_index = row_index - 1 WHERE table_id = ?1 AND row_index > ?2",
                )
                .bind(table_id)
                .bind(row_index)
                .execute(&mut *tx)
                .await?;
                sqlx::query(
                    "UPDATE data_tables SET row_count = row_count - 1, updated_at = ?1 WHERE id = ?2",
                )
                .bind(now)
                .bind(table_id)
                .execute(&mut *tx)
                .await?;
                tx.commit().await?;
                Ok(true)
            }
        }
    }

    pub async fn delete_table(&self, id: Uuid) -> Result<bool, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                let result = sqlx::query("DELETE FROM data_tables WHERE id = $1")
                    .bind(id)
                    .execute(pool)
                    .await?;
                Ok(result.rows_affected() > 0)
            }
            Database::Sqlite(pool) => {
                let result = sqlx::query("DELETE FROM data_tables WHERE id = ?1")
                    .bind(id)
                    .execute(pool)
                    .await?;
                Ok(result.rows_affected() > 0)
            }
        }
    }
}
