package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

type flowLogEntryRow struct {
	ID          int64
	EntryDate   time.Time
	Direction   string
	Amount      float64
	Currency    string
	Category    string
	Status      string
	Title       string
	Notes       sql.NullString
	TagsRaw     sql.NullString
	CreatedBy   sql.NullInt64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func scanFlowLogEntry(sc interface {
	Scan(dest ...any) error
}) (flowLogEntryRow, error) {
	var row flowLogEntryRow
	err := sc.Scan(
		&row.ID, &row.EntryDate, &row.Direction, &row.Amount, &row.Currency,
		&row.Category, &row.Status, &row.Title, &row.Notes, &row.TagsRaw,
		&row.CreatedBy, &row.CreatedAt, &row.UpdatedAt,
	)
	return row, err
}

func flowLogEntryToJSON(row flowLogEntryRow) gin.H {
	h := gin.H{
		"id":          row.ID,
		"entry_date":  row.EntryDate.Format("2006-01-02"),
		"direction":   row.Direction,
		"amount":      row.Amount,
		"currency":    row.Currency,
		"category":    row.Category,
		"status":      row.Status,
		"title":       row.Title,
		"created_at":  row.CreatedAt.Format(time.RFC3339),
		"updated_at":  row.UpdatedAt.Format(time.RFC3339),
	}
	if row.Notes.Valid {
		h["notes"] = row.Notes.String
	} else {
		h["notes"] = ""
	}
	if row.TagsRaw.Valid && strings.TrimSpace(row.TagsRaw.String) != "" {
		var tags []string
		if err := json.Unmarshal([]byte(row.TagsRaw.String), &tags); err == nil {
			h["tags"] = tags
		} else {
			h["tags"] = []string{}
		}
	} else {
		h["tags"] = []string{}
	}
	if row.CreatedBy.Valid {
		h["created_by"] = row.CreatedBy.Int64
	}
	return h
}

func normalizeFlowDirection(d string) (string, bool) {
	d = strings.ToLower(strings.TrimSpace(d))
	switch d {
	case "income", "in", "credit", "received", "earn", "earning":
		return "income", true
	case "expense", "out", "debit", "spent", "spend", "cost":
		return "expense", true
	default:
		return "", false
	}
}

func encodeFlowTags(tags []string) (sql.NullString, error) {
	clean := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			clean = append(clean, t)
		}
	}
	if len(clean) == 0 {
		return sql.NullString{}, nil
	}
	b, err := json.Marshal(clean)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{String: string(b), Valid: true}, nil
}

func (a *API) ListFlowLogEntries(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))
	direction := strings.TrimSpace(c.Query("direction"))
	status := strings.TrimSpace(c.Query("status"))
	category := strings.TrimSpace(c.Query("category"))
	limit := intArgBooki(map[string]interface{}{"limit": c.Query("limit")}, "limit", 200)

	q := `SELECT id, entry_date, direction, amount, currency, category, status, title, notes, tags, created_by, created_at, updated_at
		FROM flow_log_entries WHERE organization_id=?`
	args := []any{orgID}
	if from != "" {
		q += " AND entry_date >= ?"
		args = append(args, from)
	}
	if to != "" {
		q += " AND entry_date <= ?"
		args = append(args, to)
	}
	if direction != "" {
		if norm, ok := normalizeFlowDirection(direction); ok {
			q += " AND direction = ?"
			args = append(args, norm)
		}
	}
	if status != "" {
		q += " AND status = ?"
		args = append(args, status)
	}
	if category != "" {
		q += " AND category = ?"
		args = append(args, category)
	}
	q += " ORDER BY entry_date DESC, id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := a.DB.Query(q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		row, err := scanFlowLogEntry(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, flowLogEntryToJSON(row))
	}
	if list == nil {
		list = []gin.H{}
	}
	c.JSON(http.StatusOK, gin.H{"entries": list})
}

func (a *API) CreateFlowLogEntry(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	uid := middleware.GetUserID(c)
	var body struct {
		EntryDate string   `json:"entry_date"`
		Direction string   `json:"direction" binding:"required"`
		Amount    float64  `json:"amount" binding:"required"`
		Currency  string   `json:"currency"`
		Category  string   `json:"category"`
		Status    string   `json:"status"`
		Title     string   `json:"title"`
		Notes     string   `json:"notes"`
		Tags      []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dir, ok := normalizeFlowDirection(body.Direction)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be income or expense"})
		return
	}
	if body.Amount < 0 {
		body.Amount = -body.Amount
	}
	entryDate := strings.TrimSpace(body.EntryDate)
	if entryDate == "" {
		entryDate = time.Now().Format("2006-01-02")
	}
	if body.Currency == "" {
		body.Currency = "USD"
	}
	if body.Status == "" {
		body.Status = "logged"
	}
	tagsJSON, err := encodeFlowTags(body.Tags)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var notes sql.NullString
	if strings.TrimSpace(body.Notes) != "" {
		notes = sql.NullString{String: strings.TrimSpace(body.Notes), Valid: true}
	}

	res, err := a.DB.Exec(`INSERT INTO flow_log_entries
		(organization_id, entry_date, direction, amount, currency, category, status, title, notes, tags, created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		orgID, entryDate, dir, body.Amount, strings.ToUpper(body.Currency), strings.TrimSpace(body.Category),
		strings.TrimSpace(body.Status), strings.TrimSpace(body.Title), notes, tagsJSON, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *API) UpdateFlowLogEntry(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")
	var body struct {
		EntryDate *string   `json:"entry_date"`
		Direction *string   `json:"direction"`
		Amount    *float64  `json:"amount"`
		Currency  *string   `json:"currency"`
		Category  *string   `json:"category"`
		Status    *string   `json:"status"`
		Title     *string   `json:"title"`
		Notes     *string   `json:"notes"`
		Tags      *[]string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sets := make([]string, 0, 10)
	args := make([]any, 0, 12)
	if body.EntryDate != nil {
		sets = append(sets, "entry_date=?")
		args = append(args, strings.TrimSpace(*body.EntryDate))
	}
	if body.Direction != nil {
		dir, ok := normalizeFlowDirection(*body.Direction)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "direction must be income or expense"})
			return
		}
		sets = append(sets, "direction=?")
		args = append(args, dir)
	}
	if body.Amount != nil {
		amt := *body.Amount
		if amt < 0 {
			amt = -amt
		}
		sets = append(sets, "amount=?")
		args = append(args, amt)
	}
	if body.Currency != nil {
		sets = append(sets, "currency=?")
		args = append(args, strings.ToUpper(strings.TrimSpace(*body.Currency)))
	}
	if body.Category != nil {
		sets = append(sets, "category=?")
		args = append(args, strings.TrimSpace(*body.Category))
	}
	if body.Status != nil {
		sets = append(sets, "status=?")
		args = append(args, strings.TrimSpace(*body.Status))
	}
	if body.Title != nil {
		sets = append(sets, "title=?")
		args = append(args, strings.TrimSpace(*body.Title))
	}
	if body.Notes != nil {
		sets = append(sets, "notes=?")
		args = append(args, strings.TrimSpace(*body.Notes))
	}
	if body.Tags != nil {
		tagsJSON, err := encodeFlowTags(*body.Tags)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sets = append(sets, "tags=?")
		args = append(args, tagsJSON)
	}
	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	args = append(args, orgID, id)
	q := fmt.Sprintf("UPDATE flow_log_entries SET %s WHERE organization_id=? AND id=?", strings.Join(sets, ", "))
	res, err := a.DB.Exec(q, args...)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) DeleteFlowLogEntry(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")
	res, err := a.DB.Exec(`DELETE FROM flow_log_entries WHERE organization_id=? AND id=?`, orgID, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) FlowLogSummary(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))
	now := time.Now()
	if from == "" {
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	if to == "" {
		to = now.Format("2006-01-02")
	}
	summary, err := a.fetchFlowLogSummary(orgID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (a *API) fetchFlowLogSummary(orgID int64, from, to string) (gin.H, error) {
	rows, err := a.DB.Query(`SELECT direction, amount, category, status
		FROM flow_log_entries WHERE organization_id=? AND entry_date BETWEEN ? AND ?`,
		orgID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var income, expense float64
	var count int64
	byCat := map[string]float64{}
	byStatus := map[string]int64{}
	for rows.Next() {
		var dir, cat, status string
		var amt float64
		if err := rows.Scan(&dir, &amt, &cat, &status); err != nil {
			return nil, err
		}
		count++
		if dir == "income" {
			income += amt
		} else {
			expense += amt
		}
		if cat == "" {
			cat = "(uncategorized)"
		}
		if dir == "expense" {
			byCat[cat] -= amt
		} else {
			byCat[cat] += amt
		}
		if status == "" {
			status = "logged"
		}
		byStatus[status]++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return gin.H{
		"from": from, "to": to,
		"income": income, "expense": expense, "net": income - expense,
		"entry_count": count, "by_category": byCat, "by_status": byStatus,
	}, nil
}

func (a *API) AnalyzeFlowLog(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	var body struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Prompt string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	from := strings.TrimSpace(body.From)
	to := strings.TrimSpace(body.To)
	if from == "" || to == "" {
		from, to = assistantDateRange(map[string]string{"from": from, "to": to}, now)
	}
	prompt := strings.TrimSpace(body.Prompt)
	if prompt == "" {
		prompt = "Produce a detailed spending and income report with trends, category breakdown, status notes, and actionable observations."
	}

	entries, err := a.fetchFlowLogEntries(orgID, from, to, 500)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	summary, err := a.fetchFlowLogSummary(orgID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if a.AI == nil || !a.AI.Configured() {
		report := buildFlowLogFallbackReport(from, to, summary, entries)
		c.JSON(http.StatusOK, gin.H{
			"report":      report,
			"from":        from,
			"to":          to,
			"entry_count": len(entries),
			"ai_enabled":  false,
		})
		return
	}

	report, err := a.generateFlowLogAIReport(c.Request.Context(), from, to, prompt, summary, entries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"report":      report,
		"from":        from,
		"to":          to,
		"entry_count": len(entries),
		"ai_enabled":  true,
	})
}

func (a *API) fetchFlowLogEntries(orgID int64, from, to string, limit int) ([]gin.H, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := a.DB.Query(`SELECT id, entry_date, direction, amount, currency, category, status, title, notes, tags, created_by, created_at, updated_at
		FROM flow_log_entries WHERE organization_id=? AND entry_date BETWEEN ? AND ?
		ORDER BY entry_date DESC, id DESC LIMIT ?`, orgID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		row, err := scanFlowLogEntry(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, flowLogEntryToJSON(row))
	}
	return list, rows.Err()
}

func buildFlowLogFallbackReport(from, to string, summary gin.H, entries []gin.H) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Flow Log Report · %s → %s\n\n", from, to))
	b.WriteString(fmt.Sprintf("**Income:** %.2f · **Expense:** %.2f · **Net:** %.2f · **Entries:** %d\n\n",
		toFloat(summary["income"]), toFloat(summary["expense"]), toFloat(summary["net"]), len(entries)))

	if byCat, ok := summary["by_category"].(map[string]float64); ok && len(byCat) > 0 {
		b.WriteString("## By category\n")
		type kv struct {
			k string
			v float64
		}
		items := make([]kv, 0, len(byCat))
		for k, v := range byCat {
			items = append(items, kv{k, v})
		}
		sort.Slice(items, func(i, j int) bool {
			ai, aj := items[i].v, items[j].v
			if ai < 0 {
				ai = -ai
			}
			if aj < 0 {
				aj = -aj
			}
			return ai > aj
		})
		for _, it := range items {
			b.WriteString(fmt.Sprintf("- **%s:** %.2f\n", it.k, it.v))
		}
		b.WriteByte('\n')
	}

	if len(entries) > 0 {
		b.WriteString("## Recent entries\n")
		max := len(entries)
		if max > 15 {
			max = 15
		}
		for i := 0; i < max; i++ {
			e := entries[i]
			b.WriteString(fmt.Sprintf("- `%v` **%v** %v · %.2f %v · _%v_ · %v\n",
				e["entry_date"], e["direction"], e["title"], e["amount"], e["currency"], e["status"], e["category"]))
		}
	}
	b.WriteString("\n_Set MORPH_AI_API_KEY for AI-generated narrative analysis._")
	return b.String()
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

func (a *API) generateFlowLogAIReport(ctx context.Context, from, to, prompt string, summary gin.H, entries []gin.H) (string, error) {
	entriesJSON, _ := json.Marshal(entries)
	summaryJSON, _ := json.Marshal(summary)
	userPrompt := fmt.Sprintf(`You are analyzing Booki **Flow Log** data — informal money notes that are NOT double-entry ledger entries.
Users choose their own status labels; do not assume accounting balance rules apply.

Date range: %s → %s
User request: %s

Summary JSON:
%s

Entries JSON (%d rows, may be truncated):
%s

Write a detailed markdown report with:
1. Executive summary
2. Income vs expense breakdown
3. Category and status insights
4. Notable patterns or outliers
5. Practical recommendations

Be specific with numbers from the data. If data is sparse, say so clearly.`,
		from, to, prompt, string(summaryJSON), len(entries), string(entriesJSON))

	reply, err := a.AI.ChatCompletion(ctx, []morphai.Message{{Role: "user", Content: userPrompt}})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(reply), nil
}
