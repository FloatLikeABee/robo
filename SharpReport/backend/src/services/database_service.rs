use crate::db::models::DatabaseConnection;
use crate::db::repositories::DatabaseConnectionRepository;
use uuid::Uuid;

#[derive(Debug)]
pub struct DatabaseService {
    connections: DatabaseConnectionRepository,
}

impl DatabaseService {
    pub fn new(db: crate::db::Database) -> Self {
        Self {
            connections: DatabaseConnectionRepository::new(db),
        }
    }

    pub async fn get_connections(&self) -> Result<Vec<DatabaseConnection>, sqlx::Error> {
        self.connections.find_all().await
    }

    pub async fn get_connection(
        &self,
        id: Uuid,
    ) -> Result<Option<DatabaseConnection>, sqlx::Error> {
        self.connections.find_by_id(id).await
    }

    pub async fn create_connection(
        &self,
        connection: &DatabaseConnection,
    ) -> Result<DatabaseConnection, sqlx::Error> {
        self.connections.insert(connection).await
    }

    pub async fn test_connection(
        &self,
        _connection: &DatabaseConnection,
    ) -> Result<bool, sqlx::Error> {
        // TODO: Implement actual connection testing
        Ok(true)
    }
}
