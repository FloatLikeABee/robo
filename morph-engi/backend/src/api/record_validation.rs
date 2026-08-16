//! Validation shared by the normal create handlers and the AI import path.
//!
//! Both entry points must apply the same rules, so the checks live here rather
//! than inline in each handler.

/// Normalize a project name, rejecting a blank one.
pub fn require_project_name(name: &str) -> Result<String, String> {
    let n = name.trim();
    if n.is_empty() {
        return Err("project name is required".into());
    }
    Ok(n.to_string())
}

/// Return the supplied project code, or derive one from the name when blank.
pub fn project_code_or_generated(code: &str, name: &str) -> String {
    let c = code.trim();
    if !c.is_empty() {
        return c.to_string();
    }
    let slug = name
        .trim()
        .to_uppercase()
        .chars()
        .map(|c| if c.is_ascii_alphanumeric() { c } else { '-' })
        .collect::<String>()
        .trim_matches('-')
        .to_string();
    if slug.is_empty() {
        format!("PRJ-{}", chrono::Utc::now().timestamp_millis())
    } else {
        slug
    }
}

/// Normalize a contractor/person name, rejecting a blank one.
pub fn require_person_name(name: &str) -> Result<String, String> {
    let n = name.trim();
    if n.is_empty() {
        return Err("person name is required".into());
    }
    Ok(n.to_string())
}

/// A validated site log.
pub struct SiteLogFields {
    pub log_date: String,
    pub summary: String,
}

/// Validate a site log: a summary is required, and a blank date defaults to today.
pub fn validate_site_log(log_date: &str, summary: &str) -> Result<SiteLogFields, String> {
    let s = summary.trim();
    if s.is_empty() {
        return Err("log summary is required".into());
    }
    let d = log_date.trim();
    let log_date = if d.is_empty() {
        chrono::Utc::now().format("%Y-%m-%d").to_string()
    } else {
        d.to_string()
    };
    Ok(SiteLogFields {
        log_date,
        summary: s.to_string(),
    })
}

/// Map assorted wordings onto the two stored directions.
pub fn normalize_direction(d: &str) -> Option<&'static str> {
    match d.trim().to_ascii_lowercase().as_str() {
        "income" | "in" | "credit" | "received" | "earn" | "earning" | "revenue" => Some("income"),
        "expense" | "out" | "debit" | "spent" | "spend" | "cost" | "payment" => Some("expense"),
        _ => None,
    }
}

/// A validated flow-log entry.
pub struct FlowEntryFields {
    pub entry_date: String,
    pub direction: &'static str,
    pub amount: f64,
    pub currency: String,
    pub status: String,
    pub title: String,
}

/// Validate a flow-log entry: positive amount, known direction, non-empty date.
pub fn validate_flow_entry(
    entry_date: &str,
    direction: &str,
    amount: f64,
    currency: &str,
    status: &str,
    title: &str,
) -> Result<FlowEntryFields, String> {
    let direction =
        normalize_direction(direction).ok_or_else(|| "direction must be income or expense".to_string())?;
    if !(amount > 0.0) {
        return Err("amount must be positive".into());
    }
    let date = entry_date.trim();
    if date.is_empty() {
        return Err("entry_date is required".into());
    }
    let currency = if currency.trim().is_empty() {
        "USD".to_string()
    } else {
        currency.trim().to_string()
    };
    let status = if status.trim().is_empty() {
        "logged".to_string()
    } else {
        status.trim().to_string()
    };
    let title = if title.trim().is_empty() {
        if direction == "income" {
            "Income".to_string()
        } else {
            "Expense".to_string()
        }
    } else {
        title.trim().to_string()
    };

    Ok(FlowEntryFields {
        entry_date: date.to_string(),
        direction,
        amount,
        currency,
        status,
        title,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn blank_project_name_is_rejected() {
        assert!(require_project_name("   ").is_err());
        assert_eq!(require_project_name(" Bridge ").unwrap(), "Bridge");
    }

    #[test]
    fn missing_code_is_generated_from_name() {
        assert_eq!(
            project_code_or_generated("", "Harbour Bridge"),
            "HARBOUR-BRIDGE"
        );
        assert_eq!(project_code_or_generated(" ABC ", "Harbour Bridge"), "ABC");
    }

    #[test]
    fn uncodeable_name_gets_a_timestamp_code() {
        let code = project_code_or_generated("", "日本語");
        assert!(code.starts_with("PRJ-"), "{code}");
    }

    #[test]
    fn flow_entry_requires_positive_amount() {
        assert!(validate_flow_entry("2026-01-01", "expense", 0.0, "", "", "").is_err());
        assert!(validate_flow_entry("2026-01-01", "expense", -5.0, "", "", "").is_err());
        assert!(validate_flow_entry("2026-01-01", "expense", 5.0, "", "", "").is_ok());
    }

    #[test]
    fn flow_entry_requires_known_direction() {
        assert!(validate_flow_entry("2026-01-01", "sideways", 5.0, "", "", "").is_err());
        assert_eq!(
            validate_flow_entry("2026-01-01", "COST", 5.0, "", "", "")
                .unwrap()
                .direction,
            "expense"
        );
        assert_eq!(
            validate_flow_entry("2026-01-01", "revenue", 5.0, "", "", "")
                .unwrap()
                .direction,
            "income"
        );
    }

    #[test]
    fn flow_entry_defaults_fill_in() {
        let f = validate_flow_entry("2026-01-01", "income", 10.0, "", "", "").unwrap();
        assert_eq!(f.currency, "USD");
        assert_eq!(f.status, "logged");
        assert_eq!(f.title, "Income");
    }

    #[test]
    fn flow_entry_requires_a_date() {
        assert!(validate_flow_entry("  ", "income", 10.0, "", "", "").is_err());
    }

    #[test]
    fn site_log_requires_summary_and_defaults_date() {
        assert!(validate_site_log("2026-01-01", "  ").is_err());
        let f = validate_site_log("", "Poured slab").unwrap();
        assert_eq!(f.summary, "Poured slab");
        assert_eq!(f.log_date.len(), 10, "expected an ISO date, got {}", f.log_date);
    }
}
