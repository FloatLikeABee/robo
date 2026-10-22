pub mod migrations;

use sqlx::sqlite::SqlitePool;

pub type DbPool = SqlitePool;

pub async fn connect(url: &str) -> Result<DbPool, sqlx::Error> {
    let sqlite_url = if url.starts_with("mysql://") {
        tracing::warn!("MySQL URL detected; Dorph Engi v1 uses embedded SQLite. Set DATABASE_URL=sqlite://dorph_engi.db");
        "sqlite://dorph_engi.db".to_string()
    } else if url.starts_with("sqlite://") {
        url.to_string()
    } else {
        format!("sqlite://{}", url.trim_start_matches("sqlite:"))
    };

    let file_path = sqlite_url
        .strip_prefix("sqlite://")
        .unwrap_or("dorph_engi.db")
        .split('?')
        .next()
        .unwrap_or("dorph_engi.db");

    let resolved = if file_path.starts_with('/') {
        file_path.to_string()
    } else {
        std::env::current_dir()
            .unwrap_or_else(|_| std::path::PathBuf::from("."))
            .join(file_path)
            .to_string_lossy()
            .into_owned()
    };

    if let Some(parent) = std::path::Path::new(&resolved).parent() {
        let _ = std::fs::create_dir_all(parent);
    }

    let connect_url = if resolved.starts_with('/') {
        format!("sqlite:///{resolved}?mode=rwc")
    } else {
        format!("sqlite://{resolved}?mode=rwc")
    };
    let pool = SqlitePool::connect(&connect_url).await?;
    migrations::run(&pool).await?;
    Ok(pool)
}
