package importcol

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"idongivaflyinfa/hybridcontext"
)

type ParsedFile struct {
	Headers []string
	Rows    []map[string]string
	Format  string
}

func ParseUpload(filename string, raw []byte) (*ParsedFile, error) {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".json"):
		return parseJSON(raw)
	case strings.HasSuffix(lower, ".xlsx"), strings.HasSuffix(lower, ".xls"):
		return parseExcel(filename, raw)
	case strings.HasSuffix(lower, ".csv"), strings.HasSuffix(lower, ".txt"):
		return parseCSV(raw)
	default:
		trim := strings.TrimSpace(string(raw))
		if strings.HasPrefix(trim, "[") || strings.HasPrefix(trim, "{") {
			return parseJSON(raw)
		}
		return parseCSV(raw)
	}
}

func parseExcel(filename string, raw []byte) (*ParsedFile, error) {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".xls") && !strings.HasSuffix(lower, ".xlsx") {
		return nil, fmt.Errorf("legacy .xls is not supported; save as .xlsx, CSV, or JSON")
	}
	text, err := hybridcontext.XLSXFromBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("Excel: %w", err)
	}
	// XLSXFromBytes returns tab-separated rows; convert to CSV for the shared parser.
	lines := strings.Split(text, "\n")
	var b strings.Builder
	for i, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		for j, c := range cols {
			if j > 0 {
				b.WriteByte(',')
			}
			b.WriteString(csvEscape(c))
		}
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	parsed, err := parseCSV([]byte(b.String()))
	if err != nil {
		return nil, err
	}
	parsed.Format = "excel"
	return parsed, nil
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func normalizeHeader(h string) string {
	return strings.TrimSpace(h)
}

func parseCSV(raw []byte) (*ParsedFile, error) {
	r := csv.NewReader(bytes.NewReader(raw))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV must have a header row")
	}
	headers := make([]string, len(records[0]))
	for i, h := range records[0] {
		headers[i] = normalizeHeader(h)
	}
	if len(headers) == 0 {
		return nil, fmt.Errorf("CSV must have a header row")
	}
	var rows []map[string]string
	for _, rec := range records[1:] {
		row := map[string]string{}
		empty := true
		for i, h := range headers {
			if h == "" || i >= len(rec) {
				continue
			}
			val := strings.TrimSpace(rec[i])
			if val != "" {
				row[h] = val
				empty = false
			}
		}
		if !empty {
			rows = append(rows, row)
		}
	}
	return &ParsedFile{Headers: headers, Rows: rows, Format: "csv"}, nil
}

func parseJSON(raw []byte) (*ParsedFile, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("Invalid JSON: %w", err)
	}
	var arr []any
	switch t := v.(type) {
	case []any:
		arr = t
	case map[string]any:
		if rows, ok := t["rows"].([]any); ok {
			arr = rows
		} else if rows, ok := t["data"].([]any); ok {
			arr = rows
		} else {
			return nil, fmt.Errorf("JSON must be an array of objects or { \"rows\": [...] }")
		}
	default:
		return nil, fmt.Errorf("JSON must be an array of objects")
	}
	headerSet := map[string]struct{}{}
	var headerOrder []string
	var rows []map[string]string
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		row := map[string]string{}
		for k, val := range obj {
			h := normalizeHeader(k)
			if h == "" {
				continue
			}
			if _, seen := headerSet[h]; !seen {
				headerSet[h] = struct{}{}
				headerOrder = append(headerOrder, h)
			}
			switch x := val.(type) {
			case nil:
			case string:
				if strings.TrimSpace(x) != "" {
					row[h] = strings.TrimSpace(x)
				}
			case float64:
				row[h] = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", x), "0"), ".")
			case bool:
				row[h] = fmt.Sprintf("%v", x)
			default:
				b, _ := json.Marshal(x)
				row[h] = string(b)
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}
	return &ParsedFile{Headers: headerOrder, Rows: rows, Format: "json"}, nil
}

type ValidationReport struct {
	Valid        bool     `json:"valid"`
	Message      string   `json:"message"`
	UsesTemplate bool     `json:"uses_template"`
	RowCount     int      `json:"row_count"`
	Headers      []string `json:"headers"`
	Errors       []string `json:"errors,omitempty"`
}

func ValidateSample(kind EntityKind, parsed *ParsedFile) ValidationReport {
	spec := SpecFor(kind)
	if parsed == nil || len(parsed.Rows) == 0 {
		return ValidationReport{Valid: false, Message: "No data rows found", Headers: nil}
	}
	uses := headersMatchTemplate(parsed.Headers, spec.TemplateHeaders)
	return ValidationReport{
		Valid:        true,
		Message:      fmt.Sprintf("OK: %d rows ready to import", len(parsed.Rows)),
		UsesTemplate: uses,
		RowCount:     len(parsed.Rows),
		Headers:      parsed.Headers,
	}
}

func headersMatchTemplate(got, want []string) bool {
	if len(got) < len(want) {
		return false
	}
	for i, h := range want {
		if !strings.EqualFold(strings.TrimSpace(got[i]), strings.TrimSpace(h)) {
			return false
		}
	}
	return true
}
