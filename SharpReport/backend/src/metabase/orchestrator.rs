use crate::config::MetabaseSettings;
use std::fs;
use std::path::Path;
use std::time::Duration;
use tokio::process::{Child, Command};
use tracing::{error, info, warn};

pub struct Orchestrator {
    settings: MetabaseSettings,
    process: Option<Child>,
    restart_count: u32,
}

impl Orchestrator {
    pub fn max_restarts(&self) -> u32 {
        self.settings.max_restarts
    }

    pub async fn new(settings: MetabaseSettings) -> Result<Self, MetabaseError> {
        let mut orchestrator = Orchestrator {
            settings,
            process: None,
            restart_count: 0,
        };

        if orchestrator.settings.autostart {
            orchestrator.ensure_jar_exists().await?;
            orchestrator.start_metabase().await?;
        } else {
            info!(
                "Metabase autostart is disabled; using external instance at http://{}:{}",
                orchestrator.settings.host, orchestrator.settings.port
            );
        }

        Ok(orchestrator)
    }

    async fn ensure_jar_exists(&self) -> Result<(), MetabaseError> {
        if !Path::new(&self.settings.jar_path).exists() {
            info!("Metabase JAR not found, downloading...");
            self.download_jar().await?;
        }
        Ok(())
    }

    async fn download_jar(&self) -> Result<(), MetabaseError> {
        if let Some(parent) = Path::new(&self.settings.jar_path).parent() {
            fs::create_dir_all(parent)?;
        }

        let response = reqwest::get(&self.settings.download_url).await?;
        if !response.status().is_success() {
            return Err(MetabaseError::DownloadFailed(response.status()));
        }

        let bytes = response.bytes().await?;
        fs::write(&self.settings.jar_path, bytes)?;
        info!("Metabase JAR downloaded successfully");
        Ok(())
    }

    pub async fn start_metabase(&mut self) -> Result<(), MetabaseError> {
        if !self.settings.autostart {
            warn!("Metabase autostart is disabled; ignoring start_metabase");
            return Ok(());
        }

        if self.process.is_some() {
            warn!("Metabase is already running");
            return Ok(());
        }

        info!("Starting Metabase process...");

        let jar_path = &self.settings.jar_path;
        let mut command = Command::new("java");

        let jvm_opts = self.settings.jvm_opts.split_whitespace();
        command.args(jvm_opts);
        command.arg("-jar").arg(jar_path);
        command.env("MB_JETTY_PORT", self.settings.port.to_string());
        command.env("MB_DB_TYPE", self.settings.db_type.clone());

        if let Some(db_url) = &self.settings.db_url {
            command.env("MB_DB_CONNECTION_URI", db_url);
        }

        command.stdin(std::process::Stdio::null());
        command.stdout(std::process::Stdio::piped());
        command.stderr(std::process::Stdio::piped());

        let child = command.spawn()?;

        self.process = Some(child);
        info!("Metabase process started");

        self.wait_for_ready().await?;

        Ok(())
    }

    pub async fn stop_metabase(&mut self) -> Result<(), MetabaseError> {
        if let Some(mut child) = self.process.take() {
            info!("Stopping Metabase process...");
            child.kill().await?;
            child.wait().await?;
            info!("Metabase process stopped");
        }
        Ok(())
    }

    pub async fn restart_metabase(&mut self) -> Result<(), MetabaseError> {
        if !self.settings.autostart {
            warn!("Metabase autostart is disabled; ignoring restart_metabase");
            return Ok(());
        }

        self.stop_metabase().await?;
        self.restart_count += 1;

        if self.restart_count >= self.settings.max_restarts {
            error!(
                "Maximum restarts reached ({}), giving up",
                self.settings.max_restarts
            );
            return Err(MetabaseError::MaxRestartsReached);
        }

        tokio::time::sleep(Duration::from_secs(5)).await;
        self.start_metabase().await
    }

    async fn wait_for_ready(&self) -> Result<(), MetabaseError> {
        let client = reqwest::Client::new();
        let url = format!(
            "http://{}:{}/api/health",
            self.settings.host, self.settings.port
        );
        // Cold start often exceeds 60s; /api/health returns 503 until the app is fully up.
        const MAX_ATTEMPTS: u32 = 90;
        const RETRY_DELAY: Duration = Duration::from_secs(2);

        for attempt in 1..=MAX_ATTEMPTS {
            match client.get(&url).send().await {
                Ok(response) => {
                    let status = response.status();
                    if status.is_success() {
                        info!("Metabase is ready after {} attempts", attempt);
                        return Ok(());
                    }
                    // Metabase serves 503 (sometimes 502) while Jetty is up but the app is still loading.
                    if status == reqwest::StatusCode::SERVICE_UNAVAILABLE
                        || status == reqwest::StatusCode::BAD_GATEWAY
                    {
                        if attempt == 1 || attempt % 10 == 0 {
                            info!(
                                "Metabase still initializing (HTTP {}), waiting... ({}/{})",
                                status.as_u16(),
                                attempt,
                                MAX_ATTEMPTS
                            );
                        }
                    } else {
                        warn!(
                            "Metabase health check unexpected HTTP {} (attempt {}/{})",
                            status.as_u16(),
                            attempt,
                            MAX_ATTEMPTS
                        );
                    }
                }
                Err(e) => {
                    if attempt == 1 || attempt % 10 == 0 {
                        info!(
                            "Metabase not accepting connections yet (attempt {}/{}): {}",
                            attempt, MAX_ATTEMPTS, e
                        );
                    } else {
                        tracing::debug!(error = %e, "Metabase connect retry");
                    }
                }
            }
            tokio::time::sleep(RETRY_DELAY).await;
        }

        Err(MetabaseError::StartupTimeout)
    }

    pub async fn health_check(&self) -> Result<bool, MetabaseError> {
        let client = reqwest::Client::new();
        let url = format!(
            "http://{}:{}/api/health",
            self.settings.host, self.settings.port
        );

        let response = client.get(&url).send().await?;
        Ok(response.status().is_success())
    }

    pub fn get_port(&self) -> u16 {
        self.settings.port
    }

    pub fn get_host(&self) -> &str {
        &self.settings.host
    }
}

#[derive(Debug, thiserror::Error)]
pub enum MetabaseError {
    #[error("Failed to download Metabase JAR: {0}")]
    DownloadFailed(reqwest::StatusCode),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),

    #[error("Metabase startup timeout")]
    StartupTimeout,

    #[error("Maximum restarts reached")]
    MaxRestartsReached,

    #[error("Metabase is not running")]
    NotRunning,
}
