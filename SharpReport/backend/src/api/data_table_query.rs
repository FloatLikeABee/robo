use serde_json::Value;
use std::cmp::Ordering;

#[derive(Debug, Clone)]
pub struct RowQueryOptions {
    pub search: Option<String>,
    pub sort_by: Option<String>,
    pub sort_dir: String,
    pub limit: i32,
    pub offset: i32,
}

#[derive(Debug, Clone)]
pub struct ParsedRow {
    pub row_index: i32,
    pub data: Value,
}

#[derive(Clone, Copy)]
enum AggregateOp {
    Count,
    Sum,
    Avg,
    Min,
    Max,
}

impl AggregateOp {
    fn parse(raw: &str) -> Option<Self> {
        match raw.trim().to_lowercase().as_str() {
            "count" => Some(Self::Count),
            "sum" => Some(Self::Sum),
            "avg" => Some(Self::Avg),
            "min" => Some(Self::Min),
            "max" => Some(Self::Max),
            _ => None,
        }
    }

    fn as_key(self) -> &'static str {
        match self {
            Self::Count => "count",
            Self::Sum => "sum",
            Self::Avg => "avg",
            Self::Min => "min",
            Self::Max => "max",
        }
    }
}

pub fn apply_row_query(
    rows: Vec<ParsedRow>,
    columns: &[String],
    opts: &RowQueryOptions,
) -> (Vec<ParsedRow>, i32) {
    let mut filtered = rows;

    if let Some(search) = opts.search.as_ref().map(|s| s.trim().to_lowercase()) {
        if !search.is_empty() {
            filtered = filtered
                .into_iter()
                .filter(|row| row_matches_search(&row.data, columns, &search))
                .collect();
        }
    }

    if let Some(sort_by) = opts.sort_by.as_ref().map(|s| s.trim().to_string()) {
        if !sort_by.is_empty() && columns.iter().any(|c| c == &sort_by) {
            let desc = opts.sort_dir.eq_ignore_ascii_case("desc");
            filtered.sort_by(|a, b| {
                let ord = compare_cell_values(
                    cell_value(&a.data, &sort_by),
                    cell_value(&b.data, &sort_by),
                );
                if desc { ord.reverse() } else { ord }
            });
        }
    }

    let total = filtered.len() as i32;
    let start = opts.offset.max(0) as usize;
    let end = (start + opts.limit.max(1) as usize).min(filtered.len());
    let page = if start >= filtered.len() {
        Vec::new()
    } else {
        filtered[start..end].to_vec()
    };

    (page, total)
}

pub fn aggregate_rows(
    rows: &[ParsedRow],
    group_by: Option<&str>,
    op: &str,
    column: Option<&str>,
) -> Value {
    let Some(op) = AggregateOp::parse(op) else {
        return serde_json::json!({ "error": format!("unknown aggregate op: {op}") });
    };

    match op {
        AggregateOp::Count => {
            if let Some(col) = group_by.filter(|c| !c.is_empty()) {
                let mut groups: std::collections::BTreeMap<String, i64> =
                    std::collections::BTreeMap::new();
                for row in rows {
                    let key = value_to_string(cell_value(&row.data, col));
                    *groups.entry(key).or_insert(0) += 1;
                }
                let out: Vec<Value> = groups
                    .into_iter()
                    .map(|(k, v)| serde_json::json!({ "group": k, "count": v }))
                    .collect();
                Value::Array(out)
            } else {
                serde_json::json!({ "count": rows.len() })
            }
        }
        AggregateOp::Sum | AggregateOp::Avg | AggregateOp::Min | AggregateOp::Max => {
            let col = column.unwrap_or("").trim();
            if col.is_empty() {
                return serde_json::json!({ "error": "aggregate column required" });
            }
            if let Some(gb) = group_by.filter(|c| !c.is_empty()) {
                let mut groups: std::collections::BTreeMap<String, Vec<f64>> =
                    std::collections::BTreeMap::new();
                for row in rows {
                    if let Some(n) = cell_as_f64(cell_value(&row.data, col)) {
                        let key = value_to_string(cell_value(&row.data, gb));
                        groups.entry(key).or_default().push(n);
                    }
                }
                let key = op.as_key();
                let out: Vec<Value> = groups
                    .into_iter()
                    .map(|(k, nums)| {
                        serde_json::json!({
                            "group": k,
                            key: reduce_numbers(&nums, op)
                        })
                    })
                    .collect();
                Value::Array(out)
            } else {
                let nums: Vec<f64> = rows
                    .iter()
                    .filter_map(|r| cell_as_f64(cell_value(&r.data, col)))
                    .collect();
                let key = op.as_key();
                serde_json::json!({ key: reduce_numbers(&nums, op) })
            }
        }
    }
}

fn row_matches_search(data: &Value, columns: &[String], search: &str) -> bool {
    for col in columns {
        let text = value_to_string(cell_value(data, col));
        if text.to_lowercase().contains(search) {
            return true;
        }
    }
    false
}

fn cell_value<'a>(data: &'a Value, column: &str) -> Option<&'a Value> {
    data.as_object()?.get(column)
}

fn value_to_string(v: Option<&Value>) -> String {
    match v {
        None | Some(Value::Null) => String::new(),
        Some(Value::String(s)) => s.clone(),
        Some(Value::Bool(b)) => b.to_string(),
        Some(Value::Number(n)) => n.to_string(),
        Some(other) => other.to_string(),
    }
}

fn cell_as_f64(v: Option<&Value>) -> Option<f64> {
    match v? {
        Value::Number(n) => n.as_f64(),
        Value::String(s) => s.trim().parse().ok(),
        Value::Bool(b) => Some(if *b { 1.0 } else { 0.0 }),
        _ => None,
    }
}

fn compare_cell_values(a: Option<&Value>, b: Option<&Value>) -> Ordering {
    match (cell_as_f64(a), cell_as_f64(b)) {
        (Some(x), Some(y)) => x.partial_cmp(&y).unwrap_or(Ordering::Equal),
        _ => value_to_string(a)
            .to_lowercase()
            .cmp(&value_to_string(b).to_lowercase()),
    }
}

fn reduce_numbers(nums: &[f64], op: AggregateOp) -> Value {
    if nums.is_empty() {
        return Value::Null;
    }
    match op {
        AggregateOp::Sum => Value::from(nums.iter().sum::<f64>()),
        AggregateOp::Avg => Value::from(nums.iter().sum::<f64>() / nums.len() as f64),
        AggregateOp::Min => Value::from(nums.iter().cloned().fold(f64::INFINITY, f64::min)),
        AggregateOp::Max => Value::from(nums.iter().cloned().fold(f64::NEG_INFINITY, f64::max)),
        AggregateOp::Count => Value::from(nums.len()),
    }
}
