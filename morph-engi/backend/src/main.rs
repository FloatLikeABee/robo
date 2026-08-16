mod api;
mod config;
mod db;
mod middleware;
mod services;

use std::net::SocketAddr;
use std::sync::Arc;
use tracing::{info, Level};
use tracing_subscriber::FmtSubscriber;

#[tokio::main]
async fn main() {
    let subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .finish();
    tracing::subscriber::set_global_default(subscriber).expect("tracing init failed");

    let settings = config::Settings::from_env().expect("config");
    info!("Starting Morph Engi API on port {}", settings.app_port);
    info!("Morph auth: {}", settings.users_panel_base_url);

    let pool = db::connect(&settings.database_url)
        .await
        .expect("database connect");
    let state = Arc::new(services::AppState::new(pool, settings.clone()));
    let app = api::router(state);

    let addr = SocketAddr::from(([0, 0, 0, 0], settings.app_port));
    info!("Listening on http://127.0.0.1:{}", settings.app_port);
    let listener = tokio::net::TcpListener::bind(addr).await.expect("bind");
    axum::serve(listener, app).await.expect("serve");
}
