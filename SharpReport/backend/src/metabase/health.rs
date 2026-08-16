use crate::metabase::orchestrator::Orchestrator;
use std::sync::Arc;
use tokio::sync::Mutex;
use tokio::time::{Duration, interval};
use tracing::{error, info, warn};

#[derive(Clone)]
pub struct HealthMonitor {
    orchestrator: Arc<Mutex<Orchestrator>>,
    failure_count: Arc<Mutex<u32>>,
}

impl HealthMonitor {
    pub fn new(orchestrator: Arc<Mutex<Orchestrator>>) -> Self {
        Self {
            orchestrator,
            failure_count: Arc::new(Mutex::new(0)),
        }
    }

    pub async fn start_monitoring(&self, interval_secs: u64) {
        let mut interval = interval(Duration::from_secs(interval_secs));

        info!(
            "Starting Metabase health monitoring every {} seconds",
            interval_secs
        );

        loop {
            interval.tick().await;
            self.check_health().await;
        }
    }

    async fn check_health(&self) {
        let orchestrator = self.orchestrator.lock().await;

        match orchestrator.health_check().await {
            Ok(true) => {
                // Reset failure count on success
                let mut count = self.failure_count.lock().await;
                if *count > 0 {
                    info!("Metabase health check passed, resetting failure count");
                    *count = 0;
                }
            }
            Ok(false) => {
                error!("Metabase health check failed");
                self.handle_failure().await;
            }
            Err(e) => {
                error!("Metabase health check error: {}", e);
                self.handle_failure().await;
            }
        }
    }

    async fn handle_failure(&self) {
        let mut count = self.failure_count.lock().await;
        *count += 1;

        let max_restarts = {
            let orchestrator = self.orchestrator.lock().await;
            orchestrator.max_restarts()
        };

        if *count >= max_restarts {
            error!(
                "Maximum failures reached ({}), not restarting",
                max_restarts
            );
            // In production, you might want to alert monitoring systems here
        } else {
            warn!("Attempting to restart Metabase (failure count: {})", *count);

            // Release the lock before restarting
            drop(count);

            let mut orchestrator = self.orchestrator.lock().await;
            if let Err(e) = orchestrator.restart_metabase().await {
                error!("Failed to restart Metabase: {}", e);
            }
        }
    }

    pub async fn get_failure_count(&self) -> u32 {
        *self.failure_count.lock().await
    }
}
