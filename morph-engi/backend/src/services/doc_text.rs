//! Turning uploaded files into AI-readable text.
//!
//! Supported types are PDF, TXT, MD, and CSV. PDF parsing is CPU-bound and can
//! panic on malformed input, so it runs on a blocking thread behind
//! `catch_unwind` and a failure becomes a per-file error rather than aborting
//! the whole request.

use morphai::truncate_chars;
use std::panic::{catch_unwind, AssertUnwindSafe};

/// Cap on the text taken from a single file.
pub const MAX_PER_FILE_CHARS: usize = 4_000;

/// Cap on the combined text across all files in one request.
///
/// `morphai::MAX_CHAT_REQUEST_CHARS` is 12_000 and the shared client enforces it
/// by *deleting messages*, which would silently discard the system prompt that
/// defines our JSON contract. Staying well under it leaves room for the system
/// prompt and the user's instruction.
pub const MAX_COMBINED_CHARS: usize = 9_000;

/// Appended to any text that was cut, so the model and the user can both tell
/// content is missing.
pub const TRUNCATION_MARKER: &str = "\n…[truncated: content exceeded the size limit]";

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum FileKind {
    Pdf,
    Text,
    Unsupported,
}

/// Accepted types for project import, named for error messages.
pub const ACCEPTED_TYPES: &str = "PDF, TXT, CSV, MD";

const TEXT_EXTENSIONS: &[&str] = &[".txt", ".md", ".markdown", ".csv"];

/// Classify an upload. The declared MIME type wins when it carries information;
/// otherwise the filename extension decides.
pub fn classify(name: &str, mime: &str) -> FileKind {
    let m = normalize_mime(mime);
    match m.as_str() {
        "application/pdf" => return FileKind::Pdf,
        "text/plain" | "text/markdown" | "text/csv" | "application/csv" => return FileKind::Text,
        "" | "application/octet-stream" | "binary/octet-stream" => {}
        other => {
            if other.starts_with("image/") {
                return FileKind::Unsupported;
            }
            if other.starts_with("text/") {
                return FileKind::Text;
            }
        }
    }
    classify_by_extension(name)
}

fn classify_by_extension(name: &str) -> FileKind {
    let lower = name.trim().to_lowercase();
    if lower.ends_with(".pdf") {
        return FileKind::Pdf;
    }
    if TEXT_EXTENSIONS.iter().any(|e| lower.ends_with(e)) {
        return FileKind::Text;
    }
    FileKind::Unsupported
}

pub fn mime_from_name(name: &str) -> String {
    let lower = name.to_lowercase();
    if lower.ends_with(".pdf") {
        "application/pdf".into()
    } else if lower.ends_with(".csv") {
        "text/csv".into()
    } else if lower.ends_with(".md") || lower.ends_with(".markdown") {
        "text/markdown".into()
    } else if lower.ends_with(".txt") {
        "text/plain".into()
    } else if lower.ends_with(".png") {
        "image/png".into()
    } else if lower.ends_with(".jpg") || lower.ends_with(".jpeg") {
        "image/jpeg".into()
    } else if lower.ends_with(".json") {
        "application/json".into()
    } else {
        "application/octet-stream".into()
    }
}

fn normalize_mime(mime: &str) -> String {
    let m = mime.trim().to_lowercase();
    match m.split(';').next() {
        Some(head) => head.trim().to_string(),
        None => m,
    }
}

/// Read TXT/MD/CSV bytes as UTF-8, replacing invalid sequences and keeping line
/// structure so CSV header and data rows survive.
pub fn read_text(bytes: &[u8]) -> String {
    let s = String::from_utf8_lossy(bytes);
    normalize_lines(&s)
}

fn normalize_lines(s: &str) -> String {
    let unified = s.replace("\r\n", "\n").replace('\r', "\n");
    let mut out: Vec<&str> = Vec::new();
    let mut blank = 0usize;
    for line in unified.split('\n') {
        let trimmed = line.trim_end_matches([' ', '\t']);
        if trimmed.trim().is_empty() {
            blank += 1;
            if blank > 1 {
                continue;
            }
            out.push("");
        } else {
            blank = 0;
            out.push(trimmed);
        }
    }
    out.join("\n").trim().to_string()
}

/// Extract the text layer of a PDF, off the async runtime and isolated from panics.
pub async fn extract_pdf(bytes: Vec<u8>) -> Result<String, String> {
    let joined = tokio::task::spawn_blocking(move || {
        catch_unwind(AssertUnwindSafe(|| pdf_extract::extract_text_from_mem(&bytes)))
    })
    .await
    .map_err(|e| format!("PDF extraction task failed: {e}"))?;

    let parsed = match joined {
        Ok(result) => result,
        Err(_) => return Err("could not parse the PDF (the file may be malformed)".to_string()),
    };

    let text = parsed.map_err(|e| format!("could not read the PDF: {e}"))?;
    let cleaned = normalize_lines(&text);
    if cleaned.is_empty() {
        return Err(
            "no text could be extracted from the PDF (it may be a scan without a text layer)"
                .to_string(),
        );
    }
    Ok(cleaned)
}

/// Extract text from an upload according to its type.
pub async fn extract(name: &str, mime: &str, bytes: &[u8]) -> Result<String, String> {
    match classify(name, mime) {
        FileKind::Pdf => extract_pdf(bytes.to_vec()).await,
        FileKind::Text => {
            let text = read_text(bytes);
            if text.is_empty() {
                Err("the file is empty".to_string())
            } else {
                Ok(text)
            }
        }
        FileKind::Unsupported => Err(format!(
            "unsupported file type; accepted file types: {ACCEPTED_TYPES}"
        )),
    }
}

/// Cap text at `max_chars`, marking it when content is cut.
pub fn truncate_marked(s: &str, max_chars: usize) -> (String, bool) {
    if s.chars().count() <= max_chars {
        return (s.to_string(), false);
    }
    let kept = truncate_chars(s, max_chars);
    (format!("{}{}", kept.trim_end(), TRUNCATION_MARKER), true)
}

/// Build the combined document section for a prompt, keeping the whole block
/// inside `MAX_COMBINED_CHARS` by sharing the budget across files.
pub fn combine_sections(files: &[(String, String)]) -> String {
    if files.is_empty() {
        return String::new();
    }
    let per_file = (MAX_COMBINED_CHARS / files.len()).min(MAX_PER_FILE_CHARS);

    let mut out = String::new();
    for (name, text) in files {
        let (body, _) = truncate_marked(text, per_file.max(200));
        out.push_str(&format!("\n\n### File: {name}\n---\n{body}\n"));
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn classifies_by_mime_then_extension() {
        assert_eq!(classify("brief.pdf", "application/pdf"), FileKind::Pdf);
        assert_eq!(classify("brief.pdf", ""), FileKind::Pdf);
        assert_eq!(classify("notes.txt", "text/plain"), FileKind::Text);
        assert_eq!(classify("costs.csv", "text/csv"), FileKind::Text);
        assert_eq!(classify("readme.md", ""), FileKind::Text);
        assert_eq!(classify("readme.markdown", ""), FileKind::Text);
        assert_eq!(
            classify("notes.csv", "text/csv; charset=utf-8"),
            FileKind::Text
        );
    }

    #[test]
    fn generic_mime_falls_back_to_extension() {
        assert_eq!(
            classify("brief.pdf", "application/octet-stream"),
            FileKind::Pdf
        );
        assert_eq!(
            classify("plan.dwg", "application/octet-stream"),
            FileKind::Unsupported
        );
    }

    #[test]
    fn images_and_unknown_types_are_unsupported() {
        assert_eq!(classify("form.png", "image/png"), FileKind::Unsupported);
        assert_eq!(classify("form.jpg", ""), FileKind::Unsupported);
        assert_eq!(classify("plan.dwg", "application/acad"), FileKind::Unsupported);
        assert_eq!(classify("noext", ""), FileKind::Unsupported);
    }

    #[tokio::test]
    async fn unsupported_extract_names_accepted_types() {
        let err = extract("form.png", "image/png", b"x").await.unwrap_err();
        assert!(err.contains("PDF"), "{err}");
        assert!(err.contains("CSV"), "{err}");
    }

    #[test]
    fn csv_rows_are_preserved() {
        let raw = b"date,amount,note\r\n2026-01-04,120.50,cement\r\n2026-01-05,80,fuel\r\n";
        let out = read_text(raw);
        let lines: Vec<&str> = out.split('\n').collect();
        assert_eq!(lines.len(), 3, "got {out:?}");
        assert_eq!(lines[0], "date,amount,note");
        assert_eq!(lines[2], "2026-01-05,80,fuel");
        assert!(!out.contains('\r'));
    }

    #[test]
    fn invalid_utf8_is_replaced_not_fatal() {
        let out = read_text(&[b'h', b'i', 0xff, 0xfe, b'!']);
        assert!(out.contains("hi"), "{out}");
        assert!(out.contains('!'), "{out}");
    }

    #[test]
    fn truncation_is_marked() {
        let long = "x".repeat(100);
        let (out, cut) = truncate_marked(&long, 10);
        assert!(cut);
        assert!(out.ends_with(TRUNCATION_MARKER), "{out}");

        let (short, cut) = truncate_marked("short", 100);
        assert!(!cut);
        assert_eq!(short, "short");
    }

    #[test]
    fn combined_sections_stay_within_budget() {
        let files: Vec<(String, String)> = (0..3)
            .map(|i| (format!("f{i}.txt"), "y".repeat(50_000)))
            .collect();
        let out = combine_sections(&files);

        assert!(
            out.chars().count() < morphai::client::MAX_CHAT_REQUEST_CHARS,
            "combined section is {} chars",
            out.chars().count()
        );
        for i in 0..3 {
            assert!(out.contains(&format!("f{i}.txt")), "missing file {i}");
        }
    }

    #[tokio::test]
    async fn malformed_pdf_errors_without_panicking() {
        let err = extract_pdf(b"this is definitely not a pdf".to_vec())
            .await
            .unwrap_err();
        assert!(!err.is_empty());
    }

    #[tokio::test]
    async fn empty_pdf_bytes_error() {
        assert!(extract_pdf(Vec::new()).await.is_err());
    }
}
