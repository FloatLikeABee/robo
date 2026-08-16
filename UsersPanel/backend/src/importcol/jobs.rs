use std::collections::HashMap;
use std::sync::Arc;

use chrono::{DateTime, Utc};
use serde::Serialize;
use tokio::sync::RwLock;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum JobStatus {
    Pending,
    Validating,
    Mapping,
    Importing,
    Completed,
    Failed,
}

#[derive(Debug, Clone, Serialize)]
pub struct RowResult {
    pub row_ref: String,
    pub success: bool,
    pub message: String,
    pub record_id: Option<i64>,
}

#[derive(Debug, Clone, Serialize)]
pub struct ImportJob {
    pub id: Uuid,
    pub entity: String,
    pub filename: String,
    pub format: String,
    pub status: JobStatus,
    pub percent: u8,
    pub total_rows: usize,
    pub processed_rows: usize,
    pub imported: usize,
    pub failed: usize,
    pub skipped: usize,
    pub uses_template: bool,
    pub message: String,
    pub errors: Vec<String>,
    pub row_results: Vec<RowResult>,
    pub created_at: DateTime<Utc>,
    pub finished_at: Option<DateTime<Utc>>,
}

impl ImportJob {
    pub fn new(id: Uuid, entity: String, filename: String, format: String) -> Self {
        Self {
            id,
            entity,
            filename,
            format,
            status: JobStatus::Pending,
            percent: 0,
            total_rows: 0,
            processed_rows: 0,
            imported: 0,
            failed: 0,
            skipped: 0,
            uses_template: false,
            message: "Queued".into(),
            errors: Vec::new(),
            row_results: Vec::new(),
            created_at: Utc::now(),
            finished_at: None,
        }
    }

    pub fn set_progress(&mut self, processed: usize, total: usize, status: JobStatus, message: &str) {
        self.processed_rows = processed;
        self.total_rows = total;
        self.status = status;
        self.message = message.to_string();
        self.percent = if total == 0 {
            0
        } else {
            ((processed as f64 / total as f64) * 100.0).min(100.0) as u8
        };
    }
}

#[derive(Clone, Default)]
pub struct ImportJobStore {
    inner: Arc<RwLock<HashMap<Uuid, ImportJob>>>,
}

impl ImportJobStore {
    pub fn new() -> Self {
        Self {
            inner: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub async fn insert(&self, job: ImportJob) {
        let mut g = self.inner.write().await;
        g.insert(job.id, job);
    }

    pub async fn get(&self, id: Uuid) -> Option<ImportJob> {
        let g = self.inner.read().await;
        g.get(&id).cloned()
    }

    pub async fn update<F>(&self, id: Uuid, f: F)
    where
        F: FnOnce(&mut ImportJob),
    {
        let mut g = self.inner.write().await;
        if let Some(job) = g.get_mut(&id) {
            f(job);
        }
    }
}
