package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

const bookiAssistantInstructions = `You are Booki AI, an expert assistant for small-business accounting, bookings, and ledger operations.

You help staff with:
- Chart of accounts, journal entries, trial balance, profit & loss, account movement
- Customers and bookings (draft invoices)
- Safe arithmetic for margins and ratios

Account types: asset, liability, equity, revenue, expense.

To CREATE via the safe assistant flow, users should say:
- "create customer name: … email: …"
- "create booking customer_id: … date: 2026-05-01"

When you need live Booki data, respond with ONLY one JSON object (no markdown):
{"tool":"<name>","args":{...}}

Tools:
- list_accounts — chart of accounts for the org
- list_customers — customer list
- list_bookings — recent bookings (limit int, default 50)
- profit_loss — args: from (YYYY-MM-DD), to (YYYY-MM-DD)
- trial_balance — args: as_of (YYYY-MM-DD)
- account_movement — args: account (code/name/id), from, to
- calculate — args: expression (e.g. (12500-8300)/12500)
- list_flow_log — args: from, to, direction, status, limit (default 100)
- flow_log_summary — args: from, to
- analyze_flow_log — args: from, to, prompt (detailed markdown report from Flow Log notes)

Flow Log is informal money tracking (income/expense notes) separate from the double-entry ledger. Users set their own status labels., summarize in markdown with **Label:** values. If no tool needed, reply in markdown only.`

const bookiToolMaxRounds = 8
const bookiMaxToolResultRunes = morphai.DefaultToolResultMaxRunes

type bookiToolCall struct {
	Tool string                 `json:"tool"`
	Args map[string]interface{} `json:"args"`
}

func (a *API) chatWithBookiLLM(ctx context.Context, c *gin.Context, req assistantChatRequest, lastUser string) (string, any, error) {
	if a.AI == nil || !a.AI.Configured() {
		return "", nil, fmt.Errorf("MorphAI not configured")
	}

	first := bookiAssistantInstructions
	if hist := formatBookiHistory(req.Messages); hist != "" {
		first += "\n\nRecent conversation:\n" + hist
	}
	if st := formatBookiState(req.State); st != "" {
		first += "\n\nActive assistant state:\n" + st
	}
	first += "\n\nLatest user message:\n" + lastUser

	messages := []morphai.Message{{Role: "user", Content: first}}
	var lastRecord any
	orgID := middleware.GetOrgID(c)

	for round := 0; round < bookiToolMaxRounds; round++ {
		reply, err := a.AI.ChatCompletion(ctx, messages)
		if err != nil {
			return "", nil, err
		}
		reply = strings.TrimSpace(reply)
		if reply == "" {
			return "", nil, fmt.Errorf("empty model response")
		}

		obj, ok := extractJSONObject(reply)
		if !ok {
			return reply, lastRecord, nil
		}

		call, err := parseBookiToolCall(obj)
		if err != nil {
			return reply, lastRecord, nil
		}

		result, record, execErr := a.execBookiTool(ctx, orgID, call)
		if record != nil {
			lastRecord = record
		}
		toolMsg := "TOOL_RESULT"
		if execErr != nil {
			toolMsg += " error=" + execErr.Error()
		} else {
			toolMsg += "\n" + morphai.TruncateRunes(result, bookiMaxToolResultRunes)
		}

		followUp := morphai.ToolFollowUpPrompt(toolMsg)
		messages = append(messages, morphai.Message{Role: "user", Content: followUp})
	}

	return "I hit the tool limit while gathering data. Please narrow your question or try again.", lastRecord, nil
}

func (a *API) respondGeneralAssistant(c *gin.Context, req assistantChatRequest, userMsg string) {
	if a.AI != nil && a.AI.Configured() {
		reply, record, err := a.chatWithBookiLLM(c.Request.Context(), c, req, userMsg)
		if err == nil {
			c.JSON(http.StatusOK, assistantChatResponse{
				AssistantMessage: reply,
				Intent:           "general",
				State:            req.State,
				Completed:        true,
				Record:           record,
			})
			return
		}
	}
	c.JSON(http.StatusOK, assistantChatResponse{
		AssistantMessage: "I can help with **customers**, **bookings**, P&L, trial balance, account movement, and calculations.\n\nTry:\n- `P&L this month`\n- `list customers`\n- `calculate (12500-8300)/12500`\n\nSet `MORPH_AI_API_KEY` to enable the full LLM assistant.",
		Intent:           "general",
		State:            req.State,
		Completed:        true,
	})
}

func formatBookiHistory(messages []assistantMessage) string {
	start := 0
	if len(messages) > morphai.DefaultHistoryMaxMessages {
		start = len(messages) - morphai.DefaultHistoryMaxMessages
	}
	var b strings.Builder
	for _, m := range messages[start:] {
		role := strings.TrimSpace(m.Role)
		if role != "user" && role != "assistant" {
			continue
		}
		content := morphai.TruncateHistoryContent(m.Content, morphai.DefaultHistoryMaxRunes)
		if content == "" {
			continue
		}
		b.WriteString(role)
		b.WriteString(": ")
		b.WriteString(content)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func formatBookiState(st assistantState) string {
	if st.Intent == "" && len(st.Fields) == 0 {
		return ""
	}
	raw, _ := json.Marshal(st)
	return string(raw)
}

func parseBookiToolCall(raw string) (*bookiToolCall, error) {
	var call bookiToolCall
	if err := json.Unmarshal([]byte(raw), &call); err != nil {
		return nil, err
	}
	call.Tool = strings.TrimSpace(call.Tool)
	if call.Tool == "" {
		return nil, fmt.Errorf("missing tool")
	}
	if call.Args == nil {
		call.Args = map[string]interface{}{}
	}
	return &call, nil
}

func (a *API) execBookiTool(ctx context.Context, orgID int64, call *bookiToolCall) (string, any, error) {
	now := time.Now()
	switch call.Tool {
	case "list_accounts":
		rows, err := a.DB.QueryContext(ctx, `SELECT id, code, name, type, parent_id, is_system FROM accounts WHERE organization_id=? ORDER BY code`, orgID)
		if err != nil {
			return "", nil, err
		}
		defer rows.Close()
		var out []gin.H
		for rows.Next() {
			var id int64
			var code, name, typ string
			var parent sql.NullInt64
			var sys bool
			if err := rows.Scan(&id, &code, &name, &typ, &parent, &sys); err != nil {
				return "", nil, err
			}
			h := gin.H{"id": id, "code": code, "name": name, "type": typ, "is_system": sys}
			if parent.Valid {
				h["parent_id"] = parent.Int64
			}
			out = append(out, h)
		}
		payload := gin.H{"accounts": out}
		return mustJSONBooki(payload), payload, rows.Err()

	case "list_customers":
		rows, err := a.DB.QueryContext(ctx, `SELECT id, name, email, phone FROM customers WHERE organization_id=? ORDER BY name`, orgID)
		if err != nil {
			return "", nil, err
		}
		defer rows.Close()
		var list []gin.H
		for rows.Next() {
			var id int64
			var name, email, phone string
			if err := rows.Scan(&id, &name, &email, &phone); err != nil {
				return "", nil, err
			}
			list = append(list, gin.H{"id": id, "name": name, "email": email, "phone": phone})
		}
		payload := gin.H{"customers": list}
		return mustJSONBooki(payload), payload, rows.Err()

	case "list_bookings":
		limit := intArgBooki(call.Args, "limit", 50)
		rows, err := a.DB.QueryContext(ctx, `SELECT id, customer_id, booking_number, status, currency, subtotal, tax, total, booking_date
			FROM bookings WHERE organization_id=? ORDER BY id DESC LIMIT ?`, orgID, limit)
		if err != nil {
			return "", nil, err
		}
		defer rows.Close()
		var list []gin.H
		for rows.Next() {
			var id int64
			var cust sql.NullInt64
			var num, stat, cur string
			var sub, tax, tot float64
			var bd time.Time
			if err := rows.Scan(&id, &cust, &num, &stat, &cur, &sub, &tax, &tot, &bd); err != nil {
				return "", nil, err
			}
			h := gin.H{
				"id": id, "booking_number": num, "status": stat,
				"currency": cur, "subtotal": sub, "tax": tax, "total": tot,
				"booking_date": bd.Format("2006-01-02"),
			}
			if cust.Valid {
				h["customer_id"] = cust.Int64
			}
			list = append(list, h)
		}
		payload := gin.H{"bookings": list}
		return mustJSONBooki(payload), payload, rows.Err()

	case "profit_loss":
		fields := map[string]string{
			"from": stringArgBooki(call.Args, "from"),
			"to":   stringArgBooki(call.Args, "to"),
		}
		from, to := assistantDateRange(fields, now)
		rev, exp, err := a.fetchProfitLoss(orgID, from, to)
		if err != nil {
			return "", nil, err
		}
		net := rev - exp
		margin := 0.0
		if rev != 0 {
			margin = (net / rev) * 100
		}
		payload := gin.H{"from": from, "to": to, "revenue": rev, "expenses": exp, "net_income": net, "margin_percent": margin}
		return mustJSONBooki(payload), payload, nil

	case "trial_balance":
		asOf := stringArgBooki(call.Args, "as_of")
		if asOf == "" {
			asOf = now.Format("2006-01-02")
		}
		dr, cr, top, err := a.fetchTrialBalanceSummary(orgID, asOf)
		if err != nil {
			return "", nil, err
		}
		payload := gin.H{"as_of": asOf, "debit_total": dr, "credit_total": cr, "top_accounts": top}
		return mustJSONBooki(payload), payload, nil

	case "account_movement":
		hint := stringArgBooki(call.Args, "account")
		if hint == "" {
			return "", nil, fmt.Errorf("account_movement requires account")
		}
		fields := map[string]string{
			"from": stringArgBooki(call.Args, "from"),
			"to":   stringArgBooki(call.Args, "to"),
		}
		from, to := assistantDateRange(fields, now)
		accountID, code, name, err := a.resolveAccount(orgID, hint)
		if err != nil {
			return "", nil, err
		}
		opening, debit, credit, movement, closing, count, err := a.fetchAccountMovement(orgID, accountID, from, to)
		if err != nil {
			return "", nil, err
		}
		payload := gin.H{
			"account_id": accountID, "code": code, "name": name,
			"from": from, "to": to, "opening": opening, "debit": debit,
			"credit": credit, "movement": movement, "closing": closing, "line_count": count,
		}
		return mustJSONBooki(payload), payload, nil

	case "calculate":
		expr := stringArgBooki(call.Args, "expression")
		if expr == "" {
			return "", nil, fmt.Errorf("calculate requires expression")
		}
		result, err := evalExpression(expr)
		if err != nil {
			return "", nil, err
		}
		payload := gin.H{"expression": expr, "result": result}
		return mustJSONBooki(payload), payload, nil

	case "list_flow_log":
		from := stringArgBooki(call.Args, "from")
		to := stringArgBooki(call.Args, "to")
		if from == "" || to == "" {
			from, to = assistantDateRange(map[string]string{"from": from, "to": to}, now)
		}
		limit := intArgBooki(call.Args, "limit", 100)
		entries, err := a.fetchFlowLogEntries(orgID, from, to, limit)
		if err != nil {
			return "", nil, err
		}
		payload := gin.H{"from": from, "to": to, "entries": entries}
		return mustJSONBooki(payload), payload, nil

	case "flow_log_summary":
		from := stringArgBooki(call.Args, "from")
		to := stringArgBooki(call.Args, "to")
		if from == "" || to == "" {
			from, to = assistantDateRange(map[string]string{"from": from, "to": to}, now)
		}
		summary, err := a.fetchFlowLogSummary(orgID, from, to)
		if err != nil {
			return "", nil, err
		}
		return mustJSONBooki(summary), summary, nil

	case "analyze_flow_log":
		from := stringArgBooki(call.Args, "from")
		to := stringArgBooki(call.Args, "to")
		if from == "" || to == "" {
			from, to = assistantDateRange(map[string]string{"from": from, "to": to}, now)
		}
		prompt := stringArgBooki(call.Args, "prompt")
		if prompt == "" {
			prompt = "Detailed Flow Log analysis report."
		}
		entries, err := a.fetchFlowLogEntries(orgID, from, to, 500)
		if err != nil {
			return "", nil, err
		}
		summary, err := a.fetchFlowLogSummary(orgID, from, to)
		if err != nil {
			return "", nil, err
		}
		if a.AI == nil || !a.AI.Configured() {
			report := buildFlowLogFallbackReport(from, to, summary, entries)
			payload := gin.H{"from": from, "to": to, "report": report, "ai_enabled": false}
			return mustJSONBooki(payload), payload, nil
		}
		report, err := a.generateFlowLogAIReport(ctx, from, to, prompt, summary, entries)
		if err != nil {
			return "", nil, err
		}
		payload := gin.H{"from": from, "to": to, "report": report, "ai_enabled": true}
		return mustJSONBooki(payload), payload, nil

	default:
		return "", nil, fmt.Errorf("unknown tool %q", call.Tool)
	}
}

func mustJSONBooki(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func intArgBooki(args map[string]interface{}, key string, def int) int {
	v, ok := args[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		if err == nil {
			return i
		}
	}
	return def
}

func stringArgBooki(args map[string]interface{}, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	default:
		return strings.TrimSpace(fmt.Sprint(s))
	}
}

