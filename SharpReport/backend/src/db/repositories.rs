use crate::db::Database;
use crate::db::models::*;
use sqlx::FromRow;
use uuid::Uuid;

#[derive(Debug, FromRow)]
pub struct UserRow {
    pub id: Uuid,
    pub email: String,
    pub password_hash: String,
    pub name: String,
    pub role: String,
    pub avatar_url: Option<String>,
    pub metabase_user_id: Option<i32>,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

impl From<UserRow> for User {
    fn from(row: UserRow) -> Self {
        User {
            id: row.id,
            email: row.email,
            password_hash: row.password_hash,
            name: row.name,
            role: row.role,
            avatar_url: row.avatar_url,
            metabase_user_id: row.metabase_user_id,
            created_at: row.created_at,
            updated_at: row.updated_at,
        }
    }
}

#[derive(Debug)]
pub struct UserRepository {
    db: Database,
}

impl UserRepository {
    pub fn new(db: Database) -> Self {
        Self { db }
    }

    pub async fn find_by_email(&self, email: &str) -> Result<Option<User>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                let user = sqlx::query_as::<_, UserRow>(
                    "SELECT id, email, password_hash, name, role, avatar_url, metabase_user_id, created_at, updated_at FROM users WHERE email = $1"
                )
                .bind(email)
                .fetch_optional(pool)
                .await?;
                Ok(user.map(|u| u.into()))
            }
            Database::Sqlite(pool) => {
                let user = sqlx::query_as::<_, UserRow>(
                    "SELECT id, email, password_hash, name, role, avatar_url, metabase_user_id, created_at, updated_at FROM users WHERE email = ?1"
                )
                .bind(email)
                .fetch_optional(pool)
                .await?;
                Ok(user.map(|u| u.into()))
            }
        }
    }

    pub async fn find_by_id(&self, id: Uuid) -> Result<Option<User>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                let user = sqlx::query_as::<_, UserRow>(
                    "SELECT id, email, password_hash, name, role, avatar_url, metabase_user_id, created_at, updated_at FROM users WHERE id = $1"
                )
                .bind(id)
                .fetch_optional(pool)
                .await?;
                Ok(user.map(|u| u.into()))
            }
            Database::Sqlite(pool) => {
                let user = sqlx::query_as::<_, UserRow>(
                    "SELECT id, email, password_hash, name, role, avatar_url, metabase_user_id, created_at, updated_at FROM users WHERE id = ?1"
                )
                .bind(id)
                .fetch_optional(pool)
                .await?;
                Ok(user.map(|u| u.into()))
            }
        }
    }

    pub async fn insert(&self, user: &User) -> Result<(), sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                sqlx::query(
                    r#"INSERT INTO users (id, email, password_hash, name, role, avatar_url, metabase_user_id, created_at, updated_at)
                       VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"#,
                )
                .bind(user.id)
                .bind(&user.email)
                .bind(&user.password_hash)
                .bind(&user.name)
                .bind(&user.role)
                .bind(user.avatar_url.as_deref())
                .bind(user.metabase_user_id)
                .bind(user.created_at)
                .bind(user.updated_at)
                .execute(pool)
                .await?;
            }
            Database::Sqlite(pool) => {
                sqlx::query(
                    r#"INSERT INTO users (id, email, password_hash, name, role, avatar_url, metabase_user_id, created_at, updated_at)
                       VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)"#,
                )
                .bind(user.id)
                .bind(&user.email)
                .bind(&user.password_hash)
                .bind(&user.name)
                .bind(&user.role)
                .bind(user.avatar_url.as_deref())
                .bind(user.metabase_user_id)
                .bind(user.created_at)
                .bind(user.updated_at)
                .execute(pool)
                .await?;
            }
        }
        Ok(())
    }
}

#[derive(Debug, FromRow)]
pub struct DatabaseConnectionRow {
    pub id: Uuid,
    pub name: String,
    pub engine: String,
    pub host: String,
    pub port: i32,
    pub database_name: String,
    pub username: String,
    pub password_encrypted: String,
    pub ssl_enabled: bool,
    pub ssl_config: Option<String>,
    pub additional_config: Option<String>,
    pub metabase_database_id: Option<i32>,
    pub is_active: bool,
    pub created_at: chrono::DateTime<chrono::Utc>,
    pub updated_at: chrono::DateTime<chrono::Utc>,
}

impl From<DatabaseConnectionRow> for DatabaseConnection {
    fn from(row: DatabaseConnectionRow) -> Self {
        DatabaseConnection {
            id: row.id,
            name: row.name,
            engine: row.engine,
            host: row.host,
            port: row.port,
            database_name: row.database_name,
            username: row.username,
            password_encrypted: row.password_encrypted,
            ssl_enabled: row.ssl_enabled,
            ssl_config: row.ssl_config.and_then(|s| serde_json::from_str(&s).ok()),
            additional_config: row
                .additional_config
                .and_then(|s| serde_json::from_str(&s).ok()),
            metabase_database_id: row.metabase_database_id,
            is_active: row.is_active,
            created_at: row.created_at,
            updated_at: row.updated_at,
        }
    }
}

#[derive(Debug)]
pub struct DatabaseConnectionRepository {
    db: Database,
}

impl DatabaseConnectionRepository {
    pub fn new(db: Database) -> Self {
        Self { db }
    }

    pub async fn find_all(&self) -> Result<Vec<DatabaseConnection>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                let rows = sqlx::query_as::<_, DatabaseConnectionRow>(
                    "SELECT id, name, engine, host, port, database_name, username, password_encrypted, ssl_enabled, ssl_config, additional_config, metabase_database_id, is_active, created_at, updated_at FROM database_connections ORDER BY name"
                )
                .fetch_all(pool)
                .await?;
                Ok(rows.into_iter().map(|r| r.into()).collect())
            }
            Database::Sqlite(pool) => {
                let rows = sqlx::query_as::<_, DatabaseConnectionRow>(
                    "SELECT id, name, engine, host, port, database_name, username, password_encrypted, ssl_enabled, ssl_config, additional_config, metabase_database_id, is_active, created_at, updated_at FROM database_connections ORDER BY name"
                )
                .fetch_all(pool)
                .await?;
                Ok(rows.into_iter().map(|r| r.into()).collect())
            }
        }
    }

    pub async fn find_by_id(&self, id: Uuid) -> Result<Option<DatabaseConnection>, sqlx::Error> {
        match &self.db {
            Database::Postgres(pool) => {
                let row = sqlx::query_as::<_, DatabaseConnectionRow>(
                    "SELECT id, name, engine, host, port, database_name, username, password_encrypted, ssl_enabled, ssl_config, additional_config, metabase_database_id, is_active, created_at, updated_at FROM database_connections WHERE id = $1"
                )
                .bind(id)
                .fetch_optional(pool)
                .await?;
                Ok(row.map(|r| r.into()))
            }
            Database::Sqlite(pool) => {
                let row = sqlx::query_as::<_, DatabaseConnectionRow>(
                    "SELECT id, name, engine, host, port, database_name, username, password_encrypted, ssl_enabled, ssl_config, additional_config, metabase_database_id, is_active, created_at, updated_at FROM database_connections WHERE id = ?1"
                )
                .bind(id)
                .fetch_optional(pool)
                .await?;
                Ok(row.map(|r| r.into()))
            }
        }
    }

    pub async fn insert(
        &self,
        connection: &DatabaseConnection,
    ) -> Result<DatabaseConnection, sqlx::Error> {
        let ssl_json = connection.ssl_config.as_ref().map(|v| v.to_string());
        let add_json = connection.additional_config.as_ref().map(|v| v.to_string());

        match &self.db {
            Database::Postgres(pool) => {
                let row = sqlx::query_as::<_, DatabaseConnectionRow>(
                    r#"INSERT INTO database_connections (
                        id, name, engine, host, port, database_name, username,
                        password_encrypted, ssl_enabled, ssl_config, additional_config,
                        metabase_database_id, is_active, created_at, updated_at
                    )
                    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
                    RETURNING id, name, engine, host, port, database_name, username,
                              password_encrypted, ssl_enabled, ssl_config, additional_config,
                              metabase_database_id, is_active, created_at, updated_at"#,
                )
                .bind(connection.id)
                .bind(&connection.name)
                .bind(&connection.engine)
                .bind(&connection.host)
                .bind(connection.port)
                .bind(&connection.database_name)
                .bind(&connection.username)
                .bind(&connection.password_encrypted)
                .bind(connection.ssl_enabled)
                .bind(ssl_json.as_deref())
                .bind(add_json.as_deref())
                .bind(connection.metabase_database_id)
                .bind(connection.is_active)
                .bind(connection.created_at)
                .bind(connection.updated_at)
                .fetch_one(pool)
                .await?;
                Ok(row.into())
            }
            Database::Sqlite(pool) => {
                let row = sqlx::query_as::<_, DatabaseConnectionRow>(
                    r#"INSERT INTO database_connections (
                        id, name, engine, host, port, database_name, username,
                        password_encrypted, ssl_enabled, ssl_config, additional_config,
                        metabase_database_id, is_active, created_at, updated_at
                    )
                    VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11, ?12, ?13, ?14, ?15)
                    RETURNING id, name, engine, host, port, database_name, username,
                              password_encrypted, ssl_enabled, ssl_config, additional_config,
                              metabase_database_id, is_active, created_at, updated_at"#,
                )
                .bind(connection.id)
                .bind(&connection.name)
                .bind(&connection.engine)
                .bind(&connection.host)
                .bind(connection.port)
                .bind(&connection.database_name)
                .bind(&connection.username)
                .bind(&connection.password_encrypted)
                .bind(connection.ssl_enabled)
                .bind(ssl_json.as_deref())
                .bind(add_json.as_deref())
                .bind(connection.metabase_database_id)
                .bind(connection.is_active)
                .bind(connection.created_at)
                .bind(connection.updated_at)
                .fetch_one(pool)
                .await?;
                Ok(row.into())
            }
        }
    }
}
