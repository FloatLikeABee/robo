package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (a *API) ImportCSV(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	entity := c.Query("entity")
	if entity != "products" && entity != "assets" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity must be products or assets"})
		return
	}
	f, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	src, err := f.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r := csv.NewReader(bytes.NewReader(data))
	headers, err := r.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "csv headers: " + err.Error()})
		return
	}
	col := map[string]int{}
	for i, h := range headers {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	imp, fail := 0, 0
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fail++
			continue
		}
		if entity == "products" {
			sku := getCol(row, col, "sku")
			name := getCol(row, col, "name")
			if sku == "" || name == "" {
				fail++
				continue
			}
			cat := getCol(row, col, "category")
			cp := parseFloat(getCol(row, col, "cost_price"))
			sp := parseFloat(getCol(row, col, "selling_price"))
			_, err = a.DB.Exec(`INSERT INTO products (organization_id, sku, name, category, cost_price, selling_price) VALUES (?,?,?,?,?,?)`,
				orgID, sku, name, cat, cp, sp)
			if err != nil {
				fail++
			} else {
				imp++
			}
		} else {
			tag := getCol(row, col, "asset_tag")
			name := getCol(row, col, "name")
			if tag == "" {
				tag = getCol(row, col, "tag")
			}
			if tag == "" || name == "" {
				fail++
				continue
			}
			cat := getCol(row, col, "category")
			pv := parseFloat(getCol(row, col, "purchase_value"))
			st := getCol(row, col, "status")
			if st == "" {
				st = "active"
			}
			_, err = a.DB.Exec(`INSERT INTO assets (organization_id, asset_tag, name, category, purchase_value, current_value, status) VALUES (?,?,?,?,?,?,?)`,
				orgID, tag, name, cat, pv, pv, st)
			if err != nil {
				fail++
			} else {
				imp++
			}
		}
	}
	_, _ = a.DB.Exec(`INSERT INTO import_logs (organization_id, source_type, target_entity, status, imported_rows, failed_rows) VALUES (?,?,?,?,?,?)`,
		orgID, "csv", entity, "done", imp, fail)
	c.JSON(http.StatusOK, gin.H{"imported": imp, "failed": fail})
}

func getCol(row []string, col map[string]int, name string) string {
	i, ok := col[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func (a *API) ImportJSON(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	entity := c.Query("entity")
	if entity != "products" && entity != "assets" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity must be products or assets"})
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	imp, fail := 0, 0
	if entity == "products" {
		var items []map[string]any
		if json.Unmarshal(body, &items) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expected JSON array of objects"})
			return
		}
		for _, it := range items {
			sku, _ := it["sku"].(string)
			name, _ := it["name"].(string)
			if sku == "" || name == "" {
				fail++
				continue
			}
			cat, _ := it["category"].(string)
			cp, _ := toF(it["cost_price"])
			sp, _ := toF(it["selling_price"])
			if _, err := a.DB.Exec(`INSERT INTO products (organization_id, sku, name, category, cost_price, selling_price) VALUES (?,?,?,?,?,?)`,
				orgID, sku, name, cat, cp, sp); err != nil {
				fail++
			} else {
				imp++
			}
		}
	} else {
		var items []map[string]any
		if json.Unmarshal(body, &items) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expected JSON array of objects"})
			return
		}
		for _, it := range items {
			tag, _ := it["asset_tag"].(string)
			if tag == "" {
				tag, _ = it["tag"].(string)
			}
			name, _ := it["name"].(string)
			if tag == "" || name == "" {
				fail++
				continue
			}
			cat, _ := it["category"].(string)
			pv, _ := toF(it["purchase_value"])
			st, _ := it["status"].(string)
			if st == "" {
				st = "active"
			}
			if _, err := a.DB.Exec(`INSERT INTO assets (organization_id, asset_tag, name, category, purchase_value, current_value, status) VALUES (?,?,?,?,?,?,?)`,
				orgID, tag, name, cat, pv, pv, st); err != nil {
				fail++
			} else {
				imp++
			}
		}
	}
	_, _ = a.DB.Exec(`INSERT INTO import_logs (organization_id, source_type, target_entity, status, imported_rows, failed_rows) VALUES (?,?,?,?,?,?)`,
		orgID, "json", entity, "done", imp, fail)
	c.JSON(http.StatusOK, gin.H{"imported": imp, "failed": fail})
}

func toF(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func (a *API) ImportHTTP(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	entity := c.Query("entity")
	if entity != "products" && entity != "assets" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity must be products or assets"})
		return
	}
	var body struct {
		URL     string            `json:"url" binding:"required"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	method := body.Method
	if method == "" {
		method = http.MethodGet
	}
	req, err := http.NewRequest(method, body.URL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for k, v := range body.Headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if resp.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream status", "status": resp.StatusCode, "body": string(raw[:min(500, len(raw))])})
		return
	}
	imp, fail := 0, 0
	if entity == "products" {
		var items []map[string]any
		if json.Unmarshal(raw, &items) != nil {
			var wrap struct {
				Data []map[string]any `json:"data"`
			}
			if json.Unmarshal(raw, &wrap) == nil && len(wrap.Data) > 0 {
				items = wrap.Data
			}
		}
		for _, it := range items {
			sku, _ := it["sku"].(string)
			name, _ := it["name"].(string)
			if sku == "" || name == "" {
				fail++
				continue
			}
			cat, _ := it["category"].(string)
			cp, _ := toF(it["cost_price"])
			sp, _ := toF(it["selling_price"])
			if _, err := a.DB.Exec(`INSERT INTO products (organization_id, sku, name, category, cost_price, selling_price) VALUES (?,?,?,?,?,?)`,
				orgID, sku, name, cat, cp, sp); err != nil {
				fail++
			} else {
				imp++
			}
		}
	} else {
		var items []map[string]any
		if json.Unmarshal(raw, &items) != nil {
			var wrap struct {
				Data []map[string]any `json:"data"`
			}
			if json.Unmarshal(raw, &wrap) == nil {
				items = wrap.Data
			}
		}
		for _, it := range items {
			tag, _ := it["asset_tag"].(string)
			if tag == "" {
				tag, _ = it["tag"].(string)
			}
			name, _ := it["name"].(string)
			if tag == "" || name == "" {
				fail++
				continue
			}
			cat, _ := it["category"].(string)
			pv, _ := toF(it["purchase_value"])
			st, _ := it["status"].(string)
			if st == "" {
				st = "active"
			}
			if _, err := a.DB.Exec(`INSERT INTO assets (organization_id, asset_tag, name, category, purchase_value, current_value, status) VALUES (?,?,?,?,?,?,?)`,
				orgID, tag, name, cat, pv, pv, st); err != nil {
				fail++
			} else {
				imp++
			}
		}
	}
	_, _ = a.DB.Exec(`INSERT INTO import_logs (organization_id, source_type, target_entity, status, imported_rows, failed_rows) VALUES (?,?,?,?,?,?)`,
		orgID, "http", entity, "done", imp, fail)
	c.JSON(http.StatusOK, gin.H{"imported": imp, "failed": fail})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type ledgerLineRecord struct {
	EntryDate   string
	Reference   string
	Description string
	AccountCode string
	Debit       float64
	Credit      float64
	Note        string
}

func (a *API) ImportLedgerFile(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()
	raw, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	var rows []ledgerLineRecord
	switch ext {
	case ".csv":
		rows, err = parseLedgerCSV(raw)
	case ".json":
		rows, err = parseLedgerJSON(raw)
	case ".xlsx", ".xls":
		rows, err = parseLedgerXLSX(raw)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type: use .csv, .xlsx, or .json"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no ledger rows found"})
		return
	}

	accountCodes := map[string]struct{}{}
	for _, row := range rows {
		if row.AccountCode != "" {
			accountCodes[row.AccountCode] = struct{}{}
		}
	}
	accountMap, err := a.loadAccountCodeMap(orgID, accountCodes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type groupedLine struct {
		idx int
		row ledgerLineRecord
	}
	grouped := map[string][]groupedLine{}
	orderedKeys := make([]string, 0, len(rows))
	for i, row := range rows {
		key := fmt.Sprintf("%s|%s|%s", row.EntryDate, row.Reference, row.Description)
		if _, exists := grouped[key]; !exists {
			orderedKeys = append(orderedKeys, key)
		}
		grouped[key] = append(grouped[key], groupedLine{idx: i, row: row})
	}

	tx, err := a.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	entriesImported := 0
	linesImported := 0
	failed := 0
	for _, key := range orderedKeys {
		lines := grouped[key]
		if len(lines) < 2 {
			failed += len(lines)
			continue
		}
		var debit, credit float64
		valid := true
		for _, line := range lines {
			if line.row.AccountCode == "" {
				valid = false
				break
			}
			if _, ok := accountMap[line.row.AccountCode]; !ok {
				valid = false
				break
			}
			if line.row.Debit < 0 || line.row.Credit < 0 || (line.row.Debit == 0 && line.row.Credit == 0) {
				valid = false
				break
			}
			if line.row.Debit > 0 && line.row.Credit > 0 {
				valid = false
				break
			}
			debit += line.row.Debit
			credit += line.row.Credit
		}
		if !valid || mathAbs(debit-credit) > 0.009 {
			failed += len(lines)
			continue
		}
		first := lines[0].row
		res, err := tx.Exec(
			`INSERT INTO journal_entries (organization_id, reference, entry_date, description, status, source, created_by)
			VALUES (?,?,?,?, 'posted', 'import', ?)`,
			orgID,
			first.Reference,
			first.EntryDate,
			first.Description,
			userID,
		)
		if err != nil {
			failed += len(lines)
			continue
		}
		jeID, _ := res.LastInsertId()
		insertFailed := false
		for _, line := range lines {
			accountID := accountMap[line.row.AccountCode]
			if _, err := tx.Exec(
				`INSERT INTO journal_lines (journal_entry_id, account_id, debit, credit, note) VALUES (?,?,?,?,?)`,
				jeID,
				accountID,
				line.row.Debit,
				line.row.Credit,
				line.row.Note,
			); err != nil {
				insertFailed = true
				break
			}
		}
		if insertFailed {
			failed += len(lines)
			if _, delErr := tx.Exec(`DELETE FROM journal_entries WHERE id = ?`, jeID); delErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": delErr.Error()})
				return
			}
			continue
		}
		entriesImported++
		linesImported += len(lines)
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sourceType := strings.TrimPrefix(ext, ".")
	if sourceType == "" {
		sourceType = "file"
	}
	_, _ = a.DB.Exec(
		`INSERT INTO import_logs (organization_id, source_type, target_entity, status, imported_rows, failed_rows)
		 VALUES (?,?,?,?,?,?)`,
		orgID,
		sourceType,
		"ledger",
		"done",
		linesImported,
		failed,
	)
	c.JSON(http.StatusOK, gin.H{
		"entries_imported": entriesImported,
		"lines_imported":   linesImported,
		"failed_rows":      failed,
	})
}

func (a *API) loadAccountCodeMap(orgID int64, codes map[string]struct{}) (map[string]int64, error) {
	out := map[string]int64{}
	if len(codes) == 0 {
		return out, nil
	}
	all := make([]string, 0, len(codes))
	for code := range codes {
		all = append(all, code)
	}
	sort.Strings(all)
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(all)), ",")
	args := make([]any, 0, len(all)+1)
	args = append(args, orgID)
	for _, code := range all {
		args = append(args, code)
	}
	q := fmt.Sprintf(
		`SELECT id, code FROM accounts WHERE organization_id = ? AND code IN (%s)`,
		placeholders,
	)
	rows, err := a.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var code string
		if scanErr := rows.Scan(&id, &code); scanErr != nil {
			return nil, scanErr
		}
		out[strings.TrimSpace(code)] = id
	}
	return out, nil
}

func parseLedgerCSV(data []byte) ([]ledgerLineRecord, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("csv headers: %w", err)
	}
	col := map[string]int{}
	for i, h := range headers {
		col[normalizeLedgerHeader(h)] = i
	}
	required := []string{"entry_date", "account_code", "debit", "credit"}
	for _, k := range required {
		if _, ok := col[k]; !ok {
			return nil, fmt.Errorf("missing required column: %s", k)
		}
	}
	var out []ledgerLineRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rec := ledgerLineRecord{
			EntryDate:   getMappedCol(row, col, "entry_date"),
			Reference:   getMappedCol(row, col, "reference"),
			Description: getMappedCol(row, col, "description"),
			AccountCode: getMappedCol(row, col, "account_code"),
			Debit:       parseFloat(getMappedCol(row, col, "debit")),
			Credit:      parseFloat(getMappedCol(row, col, "credit")),
			Note:        getMappedCol(row, col, "note"),
		}
		if rec.EntryDate == "" || rec.AccountCode == "" {
			continue
		}
		out = append(out, rec)
	}
	return out, nil
}

func parseLedgerJSON(data []byte) ([]ledgerLineRecord, error) {
	var items []map[string]any
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("expected JSON array of objects")
	}
	out := make([]ledgerLineRecord, 0, len(items))
	for _, item := range items {
		rec := ledgerLineRecord{
			EntryDate:   stringAny(item["entry_date"]),
			Reference:   stringAny(item["reference"]),
			Description: stringAny(item["description"]),
			AccountCode: stringAny(item["account_code"]),
			Debit:       floatAny(item["debit"]),
			Credit:      floatAny(item["credit"]),
			Note:        stringAny(item["note"]),
		}
		if rec.EntryDate == "" || rec.AccountCode == "" {
			continue
		}
		out = append(out, rec)
	}
	return out, nil
}

func parseLedgerXLSX(data []byte) ([]ledgerLineRecord, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open excel: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel has no sheets")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("read excel rows: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	col := map[string]int{}
	for i, h := range rows[0] {
		col[normalizeLedgerHeader(h)] = i
	}
	required := []string{"entry_date", "account_code", "debit", "credit"}
	for _, k := range required {
		if _, ok := col[k]; !ok {
			return nil, fmt.Errorf("missing required column: %s", k)
		}
	}
	out := make([]ledgerLineRecord, 0, len(rows)-1)
	for _, row := range rows[1:] {
		rec := ledgerLineRecord{
			EntryDate:   getMappedCol(row, col, "entry_date"),
			Reference:   getMappedCol(row, col, "reference"),
			Description: getMappedCol(row, col, "description"),
			AccountCode: getMappedCol(row, col, "account_code"),
			Debit:       parseFloat(getMappedCol(row, col, "debit")),
			Credit:      parseFloat(getMappedCol(row, col, "credit")),
			Note:        getMappedCol(row, col, "note"),
		}
		if rec.EntryDate == "" || rec.AccountCode == "" {
			continue
		}
		out = append(out, rec)
	}
	return out, nil
}

func normalizeLedgerHeader(s string) string {
	t := strings.TrimSpace(strings.ToLower(s))
	t = strings.ReplaceAll(t, " ", "_")
	switch t {
	case "date":
		return "entry_date"
	case "account":
		return "account_code"
	}
	return t
}

func getMappedCol(row []string, col map[string]int, key string) string {
	i, ok := col[key]
	if !ok || i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func stringAny(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case json.Number:
		return x.String()
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func floatAny(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case json.Number:
		f, _ := x.Float64()
		return f
	case string:
		return parseFloat(x)
	default:
		return 0
	}
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
