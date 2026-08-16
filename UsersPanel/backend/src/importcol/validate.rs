use std::collections::HashSet;

use serde::Serialize;

use crate::importcol::entities::{spec_for, EntityKind};
use crate::importcol::map::{header_aliases_for, uses_template_headers};
use crate::importcol::parse::ParsedFile;

#[derive(Debug, Clone, Serialize)]
pub struct ValidationReport {
    pub valid: bool,
    pub entity: String,
    pub row_count: usize,
    pub header_count: usize,
    pub uses_template: bool,
    pub matched_headers: Vec<String>,
    pub extra_headers: Vec<String>,
    pub missing_required: Vec<String>,
    pub warnings: Vec<String>,
    pub sample_issues: Vec<String>,
    pub message: String,
}

pub fn validate_sample(kind: EntityKind, parsed: &ParsedFile) -> ValidationReport {
    let spec = spec_for(kind);
    let template_set: HashSet<String> = spec
        .template_headers
        .iter()
        .map(|h| h.to_lowercase())
        .collect();
    let aliases = header_aliases_for(kind);
    let uses_template = uses_template_headers(kind, &parsed.headers);

    let mut matched = Vec::new();
    let mut extra = Vec::new();
    for h in &parsed.headers {
        let norm = h.to_lowercase();
        if template_set.contains(&norm) || aliases.contains_key(&norm) {
            matched.push(h.clone());
        } else if uses_template {
            extra.push(h.clone());
        } else if known_field(kind, &norm) {
            matched.push(h.clone());
        } else {
            extra.push(h.clone());
        }
    }

    let missing_required: Vec<String> = Vec::new();
    let mut warnings = Vec::new();
    if parsed.rows.is_empty() {
        warnings.push("No data rows found.".into());
    }
    if parsed.rows.len() > 50_000 {
        warnings.push(format!(
            "Large file ({} rows). Import may take a few minutes.",
            parsed.rows.len()
        ));
    }
    if !extra.is_empty() {
        warnings.push(format!(
            "{} column(s) will be stored in JSON Details.",
            extra.len()
        ));
    }

    let sample_issues = Vec::new();
    let valid = !parsed.rows.is_empty()
        && missing_required.is_empty()
        && sample_issues.is_empty();

    let message = if !valid && parsed.rows.is_empty() {
        "No importable data rows found.".to_string()
    } else if !valid {
        "Validation failed.".to_string()
    } else if uses_template {
        "Ready to import. Template columns map to MorphData; extra columns go to JSON Details.".to_string()
    } else {
        "Ready to import. Matching columns map to MorphData; other columns go to JSON Details.".to_string()
    };

    ValidationReport {
        valid,
        entity: kind.as_str().to_string(),
        row_count: parsed.rows.len(),
        header_count: parsed.headers.len(),
        uses_template,
        matched_headers: matched,
        extra_headers: extra,
        missing_required,
        warnings,
        sample_issues,
        message,
    }
}

fn known_field(kind: EntityKind, field: &str) -> bool {
    let spec = spec_for(kind);
    spec.template_headers
        .iter()
        .chain(spec.optional_fields.iter())
        .any(|h| h.eq_ignore_ascii_case(field))
}

impl ParsedFile {
    pub fn row_count(&self) -> usize {
        self.rows.len()
    }
}
