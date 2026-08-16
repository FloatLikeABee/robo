use axum::extract::{Multipart, Path, State};
use axum::Json;
use serde::Serialize;
use uuid::Uuid;

use crate::error::{AppError, AppResult};
use crate::extractors::JwtClaims;
use crate::importcol::entities::{all_specs, spec_for, EntityKind};
use crate::importcol::insert::{insert_record, summarize_results};
use crate::importcol::jobs::{ImportJob, ImportJobStore, JobStatus, RowResult};
use crate::importcol::map::map_all_rows;
use crate::importcol::parse::parse_upload;
use crate::importcol::validate::validate_sample;
use crate::permissions::{require_role, ROLE_ADMIN};
use crate::state::AppState;

#[derive(Serialize)]
pub struct EntitiesResponse {
    pub entities: Vec<crate::importcol::entities::EntitySpec>,
}

pub async fn list_entities(JwtClaims(claims): JwtClaims) -> AppResult<Json<EntitiesResponse>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }
    Ok(Json(EntitiesResponse {
        entities: all_specs(),
    }))
}

pub async fn get_template(
    JwtClaims(claims): JwtClaims,
    Path(entity): Path<String>,
) -> AppResult<Json<serde_json::Value>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }
    let kind = EntityKind::from_str(&entity).ok_or_else(|| {
        AppError::BadRequest(format!("Unknown entity: {entity}"))
    })?;
    let spec = spec_for(kind);
    Ok(Json(serde_json::json!({
        "entity": kind.as_str(),
        "spec": spec,
    })))
}

#[derive(Serialize)]
pub struct ValidateResponse {
    pub report: crate::importcol::validate::ValidationReport,
}

pub async fn validate_upload(
    JwtClaims(claims): JwtClaims,
    mut multipart: Multipart,
) -> AppResult<Json<ValidateResponse>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }
    let (kind, filename, bytes) = read_upload(&mut multipart).await?;
    let parsed = parse_upload(kind, &filename, &bytes)?;
    let report = validate_sample(kind, &parsed);
    Ok(Json(ValidateResponse { report }))
}

#[derive(Serialize)]
pub struct StartJobResponse {
    pub job_id: String,
    pub message: String,
}

pub async fn start_import_job(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    mut multipart: Multipart,
) -> AppResult<Json<StartJobResponse>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }
    let (kind, filename, bytes) = read_upload(&mut multipart).await?;
    let parsed = parse_upload(kind, &filename, &bytes)?;
    let validation = validate_sample(kind, &parsed);
    if !validation.valid {
        return Err(AppError::BadRequest(validation.message));
    }

    let job_id = Uuid::new_v4();
    let format = parsed.format.clone();
    let total = parsed.rows.len();

    let mut job = ImportJob::new(job_id, kind.as_str().to_string(), filename, format);
    job.total_rows = total;
    job.uses_template = validation.uses_template;
    job.status = JobStatus::Pending;
    job.message = "Import queued".into();
    state.import_jobs.insert(job.clone()).await;

    let pool = state.pool.clone();
    let store = state.import_jobs.clone();

    tokio::spawn(async move {
        run_import_job(pool, store, job_id, kind, parsed).await;
    });

    Ok(Json(StartJobResponse {
        job_id: job_id.to_string(),
        message: format!(
            "Import started for {} rows. Poll GET /api/data-collector/jobs/{{id}} for progress.",
            total
        ),
    }))
}

pub async fn get_job(
    JwtClaims(claims): JwtClaims,
    State(state): State<AppState>,
    Path(job_id): Path<String>,
) -> AppResult<Json<ImportJob>> {
    if !require_role(&claims.roles, ROLE_ADMIN) {
        return Err(AppError::Forbidden);
    }
    let id = Uuid::parse_str(&job_id)
        .map_err(|_| AppError::BadRequest("invalid job id".into()))?;
    state
        .import_jobs
        .get(id)
        .await
        .ok_or(AppError::NotFound)
        .map(Json)
}

async fn run_import_job(
    pool: sqlx::MySqlPool,
    store: ImportJobStore,
    job_id: Uuid,
    kind: EntityKind,
    parsed: crate::importcol::parse::ParsedFile,
) {
    let total = parsed.rows.len();

    store
        .update(job_id, |j| {
            j.status = JobStatus::Mapping;
            j.message = "Mapping columns…".into();
            j.percent = 5;
        })
        .await;

    let mapped = match map_all_rows(kind, &parsed).await {
        Ok(m) => m,
        Err(e) => {
            let err_msg = match &e {
                AppError::BadRequest(m) => m.clone(),
                _ => "mapping failed".to_string(),
            };
            store
                .update(job_id, |j| {
                    j.status = JobStatus::Failed;
                    j.message = format!("Mapping failed: {err_msg}");
                    j.finished_at = Some(chrono::Utc::now());
                    j.percent = 100;
                })
                .await;
            return;
        }
    };

    store
        .update(job_id, |j| {
            j.status = JobStatus::Importing;
            j.message = "Importing rows…".into();
        })
        .await;

    let mut results: Vec<RowResult> = Vec::with_capacity(mapped.len());
    const BATCH: usize = 8;
    for (batch_start, chunk) in mapped.chunks(BATCH).enumerate() {
        let mut batch_results = Vec::new();
        for rec in chunk {
            let r = insert_record(&pool, kind, rec).await;
            batch_results.push(r);
        }
        results.extend(batch_results);

        let processed = results.len();
        store
            .update(job_id, |j| {
                j.set_progress(processed, total, JobStatus::Importing, "Importing rows…");
                let (imp, fail) = summarize_results(&results);
                j.imported = imp;
                j.failed = fail;
            })
            .await;

        // Yield between batches for responsiveness
        if batch_start % 4 == 3 {
            tokio::task::yield_now().await;
        }
    }

    let (imported, failed) = summarize_results(&results);
    let summary = build_summary(kind, imported, failed, total, &results);

    store
        .update(job_id, |j| {
            j.status = JobStatus::Completed;
            j.percent = 100;
            j.processed_rows = total;
            j.imported = imported;
            j.failed = failed;
            j.row_results = results.into_iter().take(500).collect();
            j.message = summary;
            j.finished_at = Some(chrono::Utc::now());
        })
        .await;
}

fn build_summary(
    kind: EntityKind,
    imported: usize,
    failed: usize,
    total: usize,
    results: &[RowResult],
) -> String {
    let mut lines = vec![
        format!("**Data Collector** import finished for **{}**.", kind.label()),
        format!("- Total rows processed: {total}"),
        format!("- Successfully imported: {imported}"),
        format!("- Failed: {failed}"),
    ];
    if failed > 0 {
        let samples: Vec<String> = results
            .iter()
            .filter(|r| !r.success)
            .take(5)
            .map(|r| format!("  - {}: {}", r.row_ref, r.message))
            .collect();
        if !samples.is_empty() {
            lines.push("- Sample errors:".into());
            lines.extend(samples);
        }
    }
    if imported > 0 {
        let ok_samples: Vec<String> = results
            .iter()
            .filter(|r| r.success)
            .take(3)
            .map(|r| {
                format!(
                    "  - {} → record id {}",
                    r.row_ref,
                    r.record_id.map(|id| id.to_string()).unwrap_or_else(|| "?".into())
                )
            })
            .collect();
        if !ok_samples.is_empty() {
            lines.push("- Sample successes:".into());
            lines.extend(ok_samples);
        }
    }
    lines.join("\n")
}

async fn read_upload(
    multipart: &mut Multipart,
) -> Result<(EntityKind, String, Vec<u8>), AppError> {
    let mut entity: Option<String> = None;
    let mut filename = String::from("upload.csv");
    let mut bytes: Option<Vec<u8>> = None;

    while let Some(field) = multipart
        .next_field()
        .await
        .map_err(|e| AppError::BadRequest(e.to_string()))?
    {
        let name = field.name().unwrap_or("").to_string();
        if name == "entity" {
            entity = Some(
                field
                    .text()
                    .await
                    .map_err(|e| AppError::BadRequest(e.to_string()))?,
            );
        } else if name == "file" {
            filename = field.file_name().unwrap_or("upload.csv").to_string();
            bytes = Some(
                field
                    .bytes()
                    .await
                    .map_err(|e| AppError::BadRequest(e.to_string()))?
                    .to_vec(),
            );
        }
    }

    let entity = entity.ok_or_else(|| AppError::BadRequest("entity field required".into()))?;
    let kind = EntityKind::from_str(&entity)
        .ok_or_else(|| AppError::BadRequest(format!("Unknown entity: {entity}")))?;
    let bytes = bytes.ok_or_else(|| AppError::BadRequest("file field required".into()))?;
    if bytes.is_empty() {
        return Err(AppError::BadRequest("file is empty".into()));
    }
    Ok((kind, filename, bytes))
}
