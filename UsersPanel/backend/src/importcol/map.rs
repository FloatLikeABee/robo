use std::collections::{HashMap, HashSet};

use serde::Serialize;
use serde_json::Value;
use uuid::Uuid;

use crate::error::AppError;
use crate::importcol::entities::{spec_for, EntityKind};
use crate::importcol::parse::ParsedFile;

#[derive(Debug, Clone, Serialize)]
pub struct MappedRecord {
    pub mysql_fields: HashMap<String, Value>,
    pub detail: Option<Value>,
    pub row_ref: String,
}

/// File includes every template column (case-insensitive). Extra columns are allowed.
pub fn uses_template_headers(kind: EntityKind, headers: &[String]) -> bool {
    let spec = spec_for(kind);
    let template: HashSet<String> = spec
        .template_headers
        .iter()
        .map(|h| h.to_lowercase())
        .collect();
    let file: HashSet<String> = headers
        .iter()
        .map(|h| h.to_lowercase())
        .filter(|h| !h.is_empty())
        .collect();
    !file.is_empty() && template.is_subset(&file)
}

pub fn header_aliases_for(kind: EntityKind) -> HashMap<String, String> {
    let mut m = HashMap::new();
    let common = [
        ("first", "first_name"),
        ("firstname", "first_name"),
        ("first_name", "first_name"),
        ("last", "last_name"),
        ("lastname", "last_name"),
        ("last_name", "last_name"),
        ("name", "name"),
        ("fullname", "name"),
        ("full_name", "name"),
        ("desc", "description"),
        ("notes", "description"),
        ("note", "description"),
        ("e_mail", "email"),
        ("mail", "email"),
        ("tag", "asset_tag"),
        ("asset", "asset_tag"),
        ("type", "ActivityType"),
        ("activity_type", "ActivityType"),
        ("title", "title"),
        ("record_title", "title"),
        ("headline", "title"),
        ("activity_name", "Name"),
        ("school", "facility"),
        ("facility_name", "name"),
        ("participant", "participant_type"),
        ("login", "login_id"),
        ("user", "login_id"),
        ("phone", "phone_number"),
        ("mobile", "Mobile"),
    ];
    for (k, v) in common {
        m.insert(k.to_string(), v.to_string());
    }
    match kind {
        EntityKind::Contact => {
            m.insert("lastname".into(), "LastName".into());
            m.insert("firstname".into(), "FirstName".into());
        }
        EntityKind::Activity => {
            m.insert("name".into(), "Name".into());
        }
        EntityKind::District => {
            m.insert("district_name".into(), "name".into());
        }
        _ => {}
    }
    m
}

fn known_field_names(kind: EntityKind) -> HashSet<String> {
    let spec = spec_for(kind);
    spec.template_headers
        .iter()
        .chain(spec.optional_fields.iter())
        .map(|s| s.to_lowercase())
        .collect()
}

fn template_column_map(kind: EntityKind, headers: &[String]) -> HashMap<String, String> {
    let spec = spec_for(kind);
    let template_lookup: HashMap<String, String> = spec
        .template_headers
        .iter()
        .map(|h| (h.to_lowercase(), h.clone()))
        .collect();

    headers
        .iter()
        .map(|h| {
            let target = template_lookup
                .get(&h.to_lowercase())
                .cloned()
                .unwrap_or_else(|| h.clone());
            (h.clone(), target)
        })
        .collect()
}

fn heuristic_column_map(kind: EntityKind, headers: &[String]) -> HashMap<String, String> {
    let aliases = header_aliases_for(kind);
    let known = known_field_names(kind);

    headers
        .iter()
        .map(|h| {
            let n = h.to_lowercase();
            let target = if known.contains(&n) {
                h.clone()
            } else if let Some(alias) = aliases.get(&n) {
                if known.contains(&alias.to_lowercase()) {
                    alias.clone()
                } else {
                    h.clone()
                }
            } else {
                h.clone()
            };
            (h.clone(), target)
        })
        .collect()
}

pub fn build_column_map(kind: EntityKind, headers: &[String]) -> HashMap<String, String> {
    if uses_template_headers(kind, headers) {
        template_column_map(kind, headers)
    } else {
        heuristic_column_map(kind, headers)
    }
}

pub async fn map_all_rows(
    kind: EntityKind,
    parsed: &ParsedFile,
) -> Result<Vec<MappedRecord>, AppError> {
    let column_map = build_column_map(kind, &parsed.headers);
    let mut out = Vec::with_capacity(parsed.rows.len());
    for (idx, row) in parsed.rows.iter().enumerate() {
        out.push(map_single_row(kind, row, &column_map, idx + 1)?);
    }
    Ok(out)
}

fn map_single_row(
    kind: EntityKind,
    row: &HashMap<String, String>,
    column_map: &HashMap<String, String>,
    row_num: usize,
) -> Result<MappedRecord, AppError> {
    let mut mysql_fields: HashMap<String, Value> = HashMap::new();
    let mut detail_obj = serde_json::Map::new();

    for (src, val) in row {
        if val.trim().is_empty() {
            continue;
        }
        let target = column_map
            .get(src)
            .cloned()
            .unwrap_or_else(|| src.clone());
        let norm = target.to_lowercase();
        if is_mysql_field(kind, &norm) {
            merge_field(kind, &mut mysql_fields, &target, val);
        } else {
            detail_obj.insert(src.clone(), Value::String(val.clone()));
        }
    }

    apply_name_splits(kind, row, &mut mysql_fields);
    ensure_required_defaults(kind, &mut mysql_fields, row);

    let detail = if detail_obj.is_empty() {
        None
    } else {
        Some(Value::Object(detail_obj))
    };

    Ok(MappedRecord {
        mysql_fields,
        detail,
        row_ref: format!("row-{row_num}"),
    })
}

fn is_mysql_field(kind: EntityKind, field: &str) -> bool {
    let known = known_field_names(kind);
    if known.contains(field) {
        return true;
    }
    match kind {
        EntityKind::Contact => ["lastname", "firstname", "email", "phone", "mobile"]
            .contains(&field),
        EntityKind::Activity => field == "name",
        _ => false,
    }
}

fn merge_field(
    kind: EntityKind,
    fields: &mut HashMap<String, Value>,
    col: &str,
    val: &str,
) {
    let key = canonical_field_key(kind, col);
    fields.insert(key, Value::String(val.trim().to_string()));
}

fn canonical_field_key(kind: EntityKind, col: &str) -> String {
    match kind {
        EntityKind::Contact => match col.to_lowercase().as_str() {
            "lastname" => "LastName".into(),
            "firstname" => "FirstName".into(),
            "email" => "Email".into(),
            "phone" => "Phone".into(),
            "mobile" => "Mobile".into(),
            other => other.to_string(),
        },
        EntityKind::Activity => {
            if col.eq_ignore_ascii_case("name") {
                "Name".into()
            } else {
                col.to_string()
            }
        }
        _ => col.to_string(),
    }
}

fn apply_name_splits(
    kind: EntityKind,
    row: &HashMap<String, String>,
    fields: &mut HashMap<String, Value>,
) {
    let full = row
        .get("name")
        .or_else(|| row.get("full_name"))
        .cloned()
        .filter(|s| !s.trim().is_empty());
    let Some(full) = full else {
        return;
    };
    match kind {
        EntityKind::Member | EntityKind::Employee | EntityKind::User | EntityKind::Contact => {
            if !fields.contains_key("last_name")
                && !fields.contains_key("LastName")
            {
                let parts: Vec<&str> = full.split_whitespace().collect();
                if parts.len() >= 2 {
                    let last = parts[parts.len() - 1].to_string();
                    let first = parts[..parts.len() - 1].join(" ");
                    let (lk, fk) = if matches!(kind, EntityKind::Contact) {
                        ("LastName", "FirstName")
                    } else {
                        ("last_name", "first_name")
                    };
                    fields.entry(lk.into()).or_insert(Value::String(last));
                    fields.entry(fk.into()).or_insert(Value::String(first));
                } else if parts.len() == 1 {
                    let k = if matches!(kind, EntityKind::Contact) {
                        "LastName"
                    } else {
                        "last_name"
                    };
                    fields.entry(k.into()).or_insert(Value::String(parts[0].to_string()));
                }
            }
        }
        EntityKind::District | EntityKind::Facility => {
            fields.entry("name".into()).or_insert(Value::String(full));
        }
        _ => {}
    }
}

fn ensure_required_defaults(
    kind: EntityKind,
    fields: &mut HashMap<String, Value>,
    row: &HashMap<String, String>,
) {
    let title = overall_title(row, fields);
    let title_or_default = |default_text: &str| -> String {
        title
            .clone()
            .filter(|t| !t.trim().is_empty())
            .unwrap_or_else(|| default_text.to_string())
    };

    match kind {
        EntityKind::District | EntityKind::Facility => {
            if !fields.contains_key("name") {
                fields.insert("name".into(), Value::String(title_or_default("Imported record")));
            }
        }
        EntityKind::Member | EntityKind::Employee => {
            if !fields.contains_key("last_name") {
                fields.insert("last_name".into(), Value::String(title_or_default("Imported")));
            }
            if !fields.contains_key("first_name") {
                if let Some(t) = title.clone() {
                    let pieces: Vec<&str> = t.split_whitespace().collect();
                    if pieces.len() > 1 {
                        fields.insert(
                            "first_name".into(),
                            Value::String(pieces[..pieces.len() - 1].join(" ")),
                        );
                    }
                }
            }
        }
        EntityKind::Asset => {
            if !fields.contains_key("asset_tag") {
                let seed = slugify(&title_or_default("asset"));
                let suffix = Uuid::new_v4().to_string()[..8].to_string();
                fields.insert(
                    "asset_tag".into(),
                    Value::String(format!("{}-{}", seed, suffix).to_uppercase()),
                );
            }
        }
        EntityKind::Activity => {
            if !fields.contains_key("Name") {
                fields.insert("Name".into(), Value::String(title_or_default("Imported activity")));
            }
        }
        EntityKind::Contact => {
            if !fields.contains_key("LastName") {
                fields.insert("LastName".into(), Value::String(title_or_default("Imported")));
            }
        }
        EntityKind::User => {
            if !fields.contains_key("first_name") {
                fields.insert("first_name".into(), Value::String(title_or_default("Imported")));
            }
            if !fields.contains_key("last_name") {
                fields.insert("last_name".into(), Value::String("User".to_string()));
            }
            if !fields.contains_key("login_id") {
                let seed = slugify(&title_or_default("user"));
                let suffix = Uuid::new_v4().to_string()[..6].to_string();
                fields.insert("login_id".into(), Value::String(format!("{seed}_{suffix}")));
            }
            if !fields.contains_key("title") {
                if let Some(t) = title.clone().filter(|v| !v.trim().is_empty()) {
                    fields.insert("title".into(), Value::String(t));
                }
            }
        }
    }

    if matches!(kind, EntityKind::Member | EntityKind::Employee)
        && !fields.contains_key("description")
    {
        let summary: Vec<String> = row
            .values()
            .filter(|v| !v.trim().is_empty())
            .take(6)
            .cloned()
            .collect();
        if !summary.is_empty() {
            fields.insert(
                "description".into(),
                Value::String(format!("Imported record: {}", summary.join("; "))),
            );
        }
    }
}

fn overall_title(
    row: &HashMap<String, String>,
    fields: &HashMap<String, Value>,
) -> Option<String> {
    if let Some(Value::String(v)) = fields.get("title") {
        if !v.trim().is_empty() {
            return Some(v.trim().to_string());
        }
    }
    for key in ["title", "name", "full_name", "description"] {
        if let Some(v) = row.get(key) {
            if !v.trim().is_empty() {
                return Some(v.trim().to_string());
            }
        }
    }
    None
}

fn slugify(input: &str) -> String {
    let mut out = String::new();
    for ch in input.chars() {
        if ch.is_ascii_alphanumeric() {
            out.push(ch.to_ascii_lowercase());
        } else if ch.is_whitespace() || ch == '-' || ch == '_' {
            if !out.ends_with('_') {
                out.push('_');
            }
        }
    }
    let trimmed = out.trim_matches('_');
    if trimmed.is_empty() {
        "record".to_string()
    } else {
        trimmed.to_string()
    }
}

pub fn organize_detail(detail: &mut Value) {
    if let Value::Object(map) = detail {
        if map.len() > 1 {
            let sorted: serde_json::Map<String, Value> = map
                .iter()
                .map(|(k, v)| (k.clone(), v.clone()))
                .collect();
            *detail = Value::Object(sorted);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn detects_template_with_extra_columns() {
        let headers = vec![
            "name".into(),
            "district_id".into(),
            "district".into(),
            "description".into(),
            "custom_field".into(),
        ];
        assert!(uses_template_headers(EntityKind::District, &headers));
    }

    #[test]
    fn extra_columns_go_to_detail() {
        let mut row = HashMap::new();
        row.insert("last_name".into(), "Doe".into());
        row.insert("first_name".into(), "Jane".into());
        row.insert("custom_field".into(), "extra".into());
        let headers = vec![
            "last_name".into(),
            "first_name".into(),
            "custom_field".into(),
        ];
        let map = template_column_map(EntityKind::Member, &headers);
        let rec = map_single_row(EntityKind::Member, &row, &map, 1).unwrap();
        assert!(rec.mysql_fields.contains_key("last_name"));
        assert!(rec.detail.is_some());
    }
}
