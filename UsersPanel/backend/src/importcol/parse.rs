use std::collections::HashMap;

use calamine::{open_workbook_auto_from_rs, Data, Reader};
use csv::ReaderBuilder;
use serde_json::Value;
use std::io::Cursor;

use crate::error::AppError;
use crate::importcol::entities::EntityKind;

#[derive(Debug, Clone)]
pub struct ParsedFile {
    pub headers: Vec<String>,
    pub rows: Vec<HashMap<String, String>>,
    pub format: String,
}

pub fn parse_upload(
    kind: EntityKind,
    filename: &str,
    bytes: &[u8],
) -> Result<ParsedFile, AppError> {
    let lower = filename.to_lowercase();
    if lower.ends_with(".json") {
        return parse_json(bytes);
    }
    if lower.ends_with(".xlsx") || lower.ends_with(".xls") {
        return parse_xlsx(bytes);
    }
    if lower.ends_with(".csv") || lower.ends_with(".txt") {
        return parse_csv(bytes);
    }
    // Sniff JSON
    let trimmed = String::from_utf8_lossy(bytes).trim().to_string();
    if trimmed.starts_with('[') || trimmed.starts_with('{') {
        return parse_json(bytes);
    }
    // Default CSV
    let _ = kind;
    parse_csv(bytes)
}

pub fn parse_csv(bytes: &[u8]) -> Result<ParsedFile, AppError> {
    let mut rdr = ReaderBuilder::new()
        .flexible(true)
        .trim(csv::Trim::All)
        .from_reader(bytes);
    let headers: Vec<String> = rdr
        .headers()
        .map_err(|e| AppError::BadRequest(format!("CSV headers: {e}")))?
        .iter()
        .map(normalize_header)
        .collect();
    if headers.is_empty() || headers.iter().all(|h| h.is_empty()) {
        return Err(AppError::BadRequest("CSV must have a header row".into()));
    }
    let mut rows = Vec::new();
    for result in rdr.records() {
        let record = result.map_err(|e| AppError::BadRequest(format!("CSV row: {e}")))?;
        let mut row = HashMap::new();
        for (i, h) in headers.iter().enumerate() {
            if h.is_empty() {
                continue;
            }
            let val = record.get(i).unwrap_or("").trim().to_string();
            if !val.is_empty() {
                row.insert(h.clone(), val);
            }
        }
        if !row.is_empty() {
            rows.push(row);
        }
    }
    Ok(ParsedFile {
        headers,
        rows,
        format: "csv".into(),
    })
}

pub fn parse_json(bytes: &[u8]) -> Result<ParsedFile, AppError> {
    let v: Value = serde_json::from_slice(bytes)
        .map_err(|e| AppError::BadRequest(format!("Invalid JSON: {e}")))?;
    let arr = match v {
        Value::Array(a) => a,
        Value::Object(o) => {
            if let Some(Value::Array(a)) = o.get("rows").or_else(|| o.get("data")) {
                a.clone()
            } else {
                return Err(AppError::BadRequest(
                    "JSON must be an array of objects or { \"rows\": [...] }".into(),
                ));
            }
        }
        _ => {
            return Err(AppError::BadRequest(
                "JSON must be an array of objects".into(),
            ));
        }
    };
    let mut headers_set = std::collections::BTreeSet::new();
    let mut rows = Vec::new();
    for item in arr {
        let obj = item.as_object().ok_or_else(|| {
            AppError::BadRequest("Each JSON row must be an object".into())
        })?;
        let mut row = HashMap::new();
        for (k, v) in obj {
            let key = normalize_header(k);
            if key.is_empty() {
                continue;
            }
            headers_set.insert(key.clone());
            let s = value_to_string(v);
            if !s.is_empty() {
                row.insert(key, s);
            }
        }
        if !row.is_empty() {
            rows.push(row);
        }
    }
    let headers: Vec<String> = headers_set.into_iter().collect();
    Ok(ParsedFile {
        headers,
        rows,
        format: "json".into(),
    })
}

pub fn parse_xlsx(bytes: &[u8]) -> Result<ParsedFile, AppError> {
    let cursor = Cursor::new(bytes.to_vec());
    let mut workbook =
        open_workbook_auto_from_rs(cursor).map_err(|e| AppError::BadRequest(format!("XLSX: {e}")))?;
    let sheet_names = workbook.sheet_names().to_vec();
    let sheet_name = sheet_names
        .first()
        .ok_or_else(|| AppError::BadRequest("XLSX has no sheets".into()))?
        .clone();
    let range = workbook
        .worksheet_range(&sheet_name)
        .map_err(|e| AppError::BadRequest(format!("XLSX sheet: {e}")))?;

    let mut headers: Vec<String> = Vec::new();
    let mut rows: Vec<HashMap<String, String>> = Vec::new();
    let mut row_idx = 0u32;
    for row in range.rows() {
        row_idx += 1;
        let cells: Vec<String> = row.iter().map(cell_to_string).collect();
        if row_idx == 1 {
            headers = cells.into_iter().map(|h| normalize_header(&h)).collect();
            continue;
        }
        if cells.iter().all(|c| c.is_empty()) {
            continue;
        }
        let mut map = HashMap::new();
        for (i, h) in headers.iter().enumerate() {
            if h.is_empty() {
                continue;
            }
            let val = cells.get(i).cloned().unwrap_or_default().trim().to_string();
            if !val.is_empty() {
                map.insert(h.clone(), val);
            }
        }
        if !map.is_empty() {
            rows.push(map);
        }
    }
    if headers.is_empty() {
        return Err(AppError::BadRequest("XLSX first row must be headers".into()));
    }
    Ok(ParsedFile {
        headers,
        rows,
        format: "xlsx".into(),
    })
}

fn cell_to_string(cell: &Data) -> String {
    match cell {
        Data::Empty => String::new(),
        Data::String(s) => s.trim().to_string(),
        Data::Float(f) => f.to_string(),
        Data::Int(i) => i.to_string(),
        Data::Bool(b) => b.to_string(),
        Data::DateTime(_) | Data::DateTimeIso(_) | Data::DurationIso(_) | Data::Error(_) => {
            String::new()
        }
    }
}

fn value_to_string(v: &Value) -> String {
    match v {
        Value::Null => String::new(),
        Value::String(s) => s.trim().to_string(),
        Value::Number(n) => n.to_string(),
        Value::Bool(b) => b.to_string(),
        other => other.to_string(),
    }
}

pub fn normalize_header(h: &str) -> String {
    let s = h.trim().to_lowercase();
    s.replace([' ', '-'], "_")
        .chars()
        .filter(|c| c.is_ascii_alphanumeric() || *c == '_')
        .collect()
}
