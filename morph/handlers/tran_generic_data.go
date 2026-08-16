package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"idongivaflyinfa/ai"

	"github.com/gin-gonic/gin"
)

const (
	genericDataMaxImportBytes  = 20 << 20
	genericDataMaxCSVRows      = 5000
	genericDataAnalyzeMaxRunes = 12000
)

const genericDataListCols = `id, title, source_type, source_filename, record_count, description, ai_analysis, created_on, last_updated`

const genericDataFullSelectCols = `id, title, source_type, source_filename, record_count, description, ai_analysis, created_on, last_updated`

var allowedGenericDataWrite = map[string]struct{}{
	"title": {}, "description": {},
}

// ListGenericData returns imported generic data records.
func (h *Handlers) ListGenericData(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	rows, err := h.TranMySQL.DB.Query(
		"SELECT " + genericDataListCols + " FROM generic_data ORDER BY last_updated DESC, id DESC LIMIT 500",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		m, ok, err := querySingleRowMapFromRows(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if ok && m != nil {
			list = append(list, m)
		}
	}
	if list == nil {
		list = []gin.H{}
	}
	c.JSON(http.StatusOK, list)
}

// GetGenericData returns one record with Mongo detail.
func (h *Handlers) GetGenericData(c *gin.Context) {
	h.getGenericDataByID(c)
}

// GetGenericDataFull is an alias for full detail (matches other Work data modules).
func (h *Handlers) GetGenericDataFull(c *gin.Context) {
	h.getGenericDataByID(c)
}

func (h *Handlers) getGenericDataByID(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+genericDataFullSelectCols+" FROM generic_data WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok || m == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "generic data not found"})
		return
	}
	h.attachEntityDetail(c, entityKeyGenericData, id, m)
	c.JSON(http.StatusOK, m)
}

// CreateGenericData inserts a record (manual create with optional detail JSON).
func (h *Handlers) CreateGenericData(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in map[string]interface{}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	detailStr, hasDetail, derr := popDetailString(in)
	if derr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": derr.Error()})
		return
	}

	title := strings.TrimSpace(fmt.Sprint(in["title"]))
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	sourceType := strings.ToLower(strings.TrimSpace(fmt.Sprint(in["source_type"])))
	if sourceType == "" {
		sourceType = "json"
	}
	if sourceType != "csv" && sourceType != "json" && sourceType != "pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source_type must be csv, json, or pdf"})
		return
	}

	recordCount := 0
	if rc, ok := in["record_count"].(float64); ok {
		recordCount = int(rc)
	}

	desc := nullableStringFromMap(in, "description")
	filename := nullableStringFromMap(in, "source_filename")

	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO generic_data (title, source_type, source_filename, record_count, description)
		 VALUES (?, ?, ?, ?, ?)`,
		title, sourceType, filename, recordCount, desc,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyGenericData, int(id64), detailStr)
	}
	m, _, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+genericDataFullSelectCols+" FROM generic_data WHERE id = ?", id64)
	if m != nil {
		h.attachEntityDetail(c, entityKeyGenericData, int(id64), m)
	}
	c.JSON(http.StatusOK, m)
}

// UpdateGenericData updates scalar fields and optional detail.
func (h *Handlers) UpdateGenericData(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var in map[string]interface{}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	detailStr, hasDetail, derr := popDetailString(in)
	if derr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": derr.Error()})
		return
	}

	var sets []string
	var args []interface{}
	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		col := genericDataJSONToCol(lk)
		if col == "" {
			continue
		}
		if _, ok := allowedGenericDataWrite[col]; !ok {
			continue
		}
		sets = append(sets, col+" = ?")
		args = append(args, v)
	}
	if len(sets) > 0 {
		args = append(args, id)
		_, err := h.TranMySQL.DB.Exec("UPDATE generic_data SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if hasDetail {
		if err := h.savePoppedDetail(c, entityKeyGenericData, id, detailStr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	m, ok, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+genericDataFullSelectCols+" FROM generic_data WHERE id = ?", id)
	if !ok || m == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "generic data not found"})
		return
	}
	h.attachEntityDetail(c, entityKeyGenericData, id, m)
	c.JSON(http.StatusOK, m)
}

// DeleteGenericData removes a record and Mongo detail.
func (h *Handlers) DeleteGenericData(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec("DELETE FROM generic_data WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "generic data not found"})
		return
	}
	h.deleteEntityDetailMongo(c.Request.Context(), entityKeyGenericData, id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ImportGenericData accepts CSV, JSON, or PDF and stores parsed content in Mongo.
func (h *Handlers) ImportGenericData(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read file"})
		return
	}
	defer f.Close()

	raw, err := io.ReadAll(io.LimitReader(f, genericDataMaxImportBytes+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(raw) > genericDataMaxImportBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 20MB)"})
		return
	}

	filename := fileHeader.Filename
	ext := strings.ToLower(path.Ext(filename))
	sourceType, ok := genericDataExtToType(ext)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type; use .csv, .json, .pdf, or .md"})
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		title = strings.TrimSuffix(filename, path.Ext(filename))
	}
	if title == "" {
		title = "Imported data"
	}
	desc := strings.TrimSpace(c.PostForm("description"))
	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}

	detailObj, recordCount, err := h.parseGenericDataImport(c.Request.Context(), sourceType, filename, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	detailBytes, err := json.Marshal(detailObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO generic_data (title, source_type, source_filename, record_count, description)
		 VALUES (?, ?, ?, ?, ?)`,
		title, sourceType, filename, recordCount, descPtr,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	id := int(id64)
	if err := h.savePoppedDetail(c, entityKeyGenericData, id, string(detailBytes)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "saved row but failed to store detail: " + err.Error()})
		return
	}

	m, _, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+genericDataFullSelectCols+" FROM generic_data WHERE id = ?", id)
	if m != nil {
		h.attachEntityDetail(c, entityKeyGenericData, id, m)
	}
	c.JSON(http.StatusOK, m)
}

// AnalyzeGenericData runs Morph AI analysis on imported material.
func (h *Handlers) AnalyzeGenericData(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+genericDataFullSelectCols+" FROM generic_data WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok || m == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "generic data not found"})
		return
	}
	h.attachEntityDetail(c, entityKeyGenericData, id, m)

	title := fmt.Sprint(m["title"])
	sourceType := fmt.Sprint(m["source_type"])
	filename := fmt.Sprint(m["source_filename"])
	excerpt := genericDataExcerptForAnalysis(m)
	prompt := genericDataAnalysisPrompt(title, sourceType, filename, excerpt)

	ctx := context.Background()
	out, err := h.aiService.ChatCompletion(ctx, []ai.DashScopeMessage{{Role: "user", Content: prompt}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "```markdown")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	out = strings.TrimSpace(out)

	_, err = h.TranMySQL.DB.Exec("UPDATE generic_data SET ai_analysis = ? WHERE id = ?", out, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	m["ai_analysis"] = out
	c.JSON(http.StatusOK, gin.H{"ai_analysis": out, "record": m})
}

func querySingleRowMapFromRows(rows interface {
	Next() bool
	Columns() ([]string, error)
	Scan(dest ...interface{}) error
}) (map[string]interface{}, bool, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, false, err
	}
	if !rows.Next() {
		return nil, false, nil
	}
	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return nil, false, err
	}
	out := make(map[string]interface{}, len(cols))
	for i, c := range cols {
		v := vals[i]
		switch t := v.(type) {
		case nil:
			out[c] = nil
		case []byte:
			out[c] = string(t)
		case time.Time:
			out[c] = t.Format(time.RFC3339)
		default:
			out[c] = v
		}
	}
	return out, true, nil
}

func genericDataExtToType(ext string) (string, bool) {
	switch ext {
	case ".csv":
		return "csv", true
	case ".json":
		return "json", true
	case ".pdf", ".md", ".markdown":
		return "pdf", true
	default:
		return "", false
	}
}

func genericDataJSONToCol(lk string) string {
	switch lk {
	case "title":
		return "title"
	case "description":
		return "description"
	default:
		return ""
	}
}

func nullableStringFromMap(in map[string]interface{}, key string) *string {
	v, ok := in[key]
	if !ok || v == nil {
		return nil
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	if s == "" {
		return nil
	}
	return &s
}

func (h *Handlers) parseGenericDataImport(ctx context.Context, sourceType, filename string, raw []byte) (map[string]interface{}, int, error) {
	meta := map[string]interface{}{
		"source_type": sourceType,
		"filename":      filename,
		"imported_at":   time.Now().UTC().Format(time.RFC3339),
		"size_bytes":    len(raw),
	}
	out := map[string]interface{}{"import_meta": meta}

	switch sourceType {
	case "csv":
		rows, columns, err := parseCSVBytes(raw)
		if err != nil {
			return nil, 0, err
		}
		if len(rows) > genericDataMaxCSVRows {
			rows = rows[:genericDataMaxCSVRows]
			meta["truncated"] = true
			meta["max_rows"] = genericDataMaxCSVRows
		}
		out["columns"] = columns
		out["rows"] = rows
		return out, len(rows), nil
	case "json":
		var parsed interface{}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return nil, 0, fmt.Errorf("invalid JSON: %w", err)
		}
		out["payload"] = parsed
		count := genericDataJSONRecordCount(parsed)
		return out, count, nil
	case "pdf":
		markdown, err := pdfBytesToMarkdown(filename, raw)
		if err != nil {
			return nil, 0, err
		}
		out["content_markdown"] = markdown
		if h1 := extractMarkdownH1(markdown); h1 != "" {
			out["article_title"] = h1
		}
		return out, 1, nil
	default:
		return nil, 0, fmt.Errorf("unsupported source type")
	}
}

func parseCSVBytes(raw []byte) ([]map[string]string, []string, error) {
	text := strings.ToValidUTF8(string(raw), "")
	r := csv.NewReader(strings.NewReader(text))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	all, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("CSV parse error: %w", err)
	}
	if len(all) == 0 {
		return nil, nil, fmt.Errorf("CSV file is empty")
	}
	header := all[0]
	columns := make([]string, len(header))
	for i, h := range header {
		columns[i] = strings.TrimSpace(h)
		if columns[i] == "" {
			columns[i] = fmt.Sprintf("column_%d", i+1)
		}
	}
	var rows []map[string]string
	for _, rec := range all[1:] {
		row := make(map[string]string, len(columns))
		for i, col := range columns {
			val := ""
			if i < len(rec) {
				val = rec[i]
			}
			row[col] = val
		}
		rows = append(rows, row)
	}
	return rows, columns, nil
}

func genericDataJSONRecordCount(parsed interface{}) int {
	switch v := parsed.(type) {
	case []interface{}:
		return len(v)
	case map[string]interface{}:
		if rows, ok := v["rows"].([]interface{}); ok {
			return len(rows)
		}
		if data, ok := v["data"].([]interface{}); ok {
			return len(data)
		}
		return 1
	default:
		return 1
	}
}

func extractMarkdownH1(md string) string {
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func genericDataExcerptForAnalysis(m map[string]interface{}) string {
	detailRaw, ok := m["detail"]
	if !ok || detailRaw == nil {
		return "(no detail payload)"
	}
	var detail map[string]interface{}
	switch v := detailRaw.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &detail); err != nil {
			return truncateRunes(v, genericDataAnalyzeMaxRunes)
		}
	case map[string]interface{}:
		detail = v
	default:
		b, _ := json.Marshal(v)
		return truncateRunes(string(b), genericDataAnalyzeMaxRunes)
	}

	sourceType := fmt.Sprint(m["source_type"])
	switch sourceType {
	case "pdf":
		if md, ok := detail["content_markdown"].(string); ok {
			return truncateRunes(md, genericDataAnalyzeMaxRunes)
		}
	case "csv":
		var b strings.Builder
		if cols, ok := detail["columns"].([]interface{}); ok {
			b.WriteString("Columns: ")
			for i, c := range cols {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(fmt.Sprint(c))
			}
			b.WriteString("\n\nSample rows:\n")
		}
		if rows, ok := detail["rows"].([]interface{}); ok {
			limit := 25
			if len(rows) < limit {
				limit = len(rows)
			}
			if limit > 0 {
				sample, _ := json.MarshalIndent(rows[:limit], "", "  ")
				b.Write(sample)
			}
		}
		return truncateRunes(b.String(), genericDataAnalyzeMaxRunes)
	case "json":
		if payload, ok := detail["payload"]; ok {
			b, _ := json.MarshalIndent(payload, "", "  ")
			return truncateRunes(string(b), genericDataAnalyzeMaxRunes)
		}
	}
	b, _ := json.MarshalIndent(detail, "", "  ")
	return truncateRunes(string(b), genericDataAnalyzeMaxRunes)
}

func genericDataAnalysisPrompt(title, sourceType, filename, excerpt string) string {
	return fmt.Sprintf(`You are Morph AI analyzing imported generic data for a transportation / school operations team.

Source: %s file "%s"
Title: %s

Material excerpt:
%s

Provide a structured analysis in Markdown with these sections:
## Summary
(2-4 sentences describing what this material is and why it matters)

## Key insights
(bullet points — patterns, notable values, themes)

## Data quality notes
(missing fields, inconsistencies, parsing caveats — or "No major issues" if clean)

## Suggested actions
(3-5 actionable next steps for staff)

Write clearly for non-technical readers. Output only the Markdown analysis.`, sourceType, filename, title, excerpt)
}
