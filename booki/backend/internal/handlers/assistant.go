package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/robo/assistmd"
)

type assistantChatRequest struct {
	Messages []assistantMessage `json:"messages"`
	State    assistantState     `json:"state"`
}

type assistantMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type assistantState struct {
	Intent string            `json:"intent"`
	Fields map[string]string `json:"fields"`
}

type assistantChatResponse struct {
	AssistantMessage string         `json:"assistant_message"`
	Intent           string         `json:"intent,omitempty"`
	MissingFields    []string       `json:"missing_fields,omitempty"`
	State            assistantState `json:"state"`
	Completed        bool           `json:"completed"`
	Record           any            `json:"record,omitempty"`
}

// AssistantChat starts a MorphAI-compatible assistant contract for Booki.
// It currently performs conversation + clarification and returns execution-ready payload intent.
func (a *API) AssistantChat(c *gin.Context) {
	var req assistantChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.State.Fields == nil {
		req.State.Fields = map[string]string{}
	}
	userMsg := latestUserMessage(req.Messages)
	if userMsg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages must include a user message"})
		return
	}
	reconcileAssistantIntent(&req.State, userMsg)
	updateFields(req.State.Fields, userMsg)
	applyDateHints(req.State.Fields, userMsg, time.Now())

	switch req.State.Intent {
	case "create_customer":
		missing := missingFields(req.State.Fields, []string{"name"})
		if len(missing) > 0 {
			c.JSON(http.StatusOK, assistantChatResponse{
				AssistantMessage: "I can create the customer. Please provide: " + strings.Join(missing, ", "),
				Intent:           req.State.Intent,
				MissingFields:    missing,
				State:            req.State,
				Completed:        false,
			})
			return
		}
		record, err := a.assistantCreateCustomer(c, req.State.Fields)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: assistmd.Success("Customer created", fmt.Sprint(record["name"])),
			Intent:           req.State.Intent,
			State:            req.State,
			Completed:        true,
			Record:           record,
		})
		return
	case "create_booking":
		missing := missingFields(req.State.Fields, []string{"customer_id", "date"})
		if len(missing) > 0 {
			c.JSON(http.StatusOK, assistantChatResponse{
				AssistantMessage: "I can create the booking. Please provide: " + strings.Join(missing, ", "),
				Intent:           req.State.Intent,
				MissingFields:    missing,
				State:            req.State,
				Completed:        false,
			})
			return
		}
		record, err := a.assistantCreateBooking(c, req.State.Fields)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: assistmd.Success("Draft booking created", fmt.Sprint(record["booking_date"])),
			Intent:           req.State.Intent,
			State:            req.State,
			Completed:        true,
			Record:           record,
		})
		return
	case "list_customers":
		a.handleAssistantListCustomers(c, req.State)
		return
	case "list_bookings":
		a.handleAssistantListBookings(c, req.State)
		return
	case "list_accounts":
		a.handleAssistantListAccounts(c, req.State)
		return
	case "analyze_profit_loss":
		orgID := middleware.GetOrgID(c)
		from, to := assistantDateRange(req.State.Fields, time.Now())
		revenue, expenses, err := a.fetchProfitLoss(orgID, from, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		net := revenue - expenses
		margin := 0.0
		if revenue != 0 {
			margin = (net / revenue) * 100
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: assistmd.KVBlock(
				fmt.Sprintf("**Profit & loss** · `%s` → `%s`", from, to),
				[][2]string{
					{"Revenue", fmt.Sprintf("%.2f", revenue)},
					{"Expenses", fmt.Sprintf("%.2f", expenses)},
					{"Net income", fmt.Sprintf("%.2f (margin %.2f%%)", net, margin)},
				},
			),
			Intent:    req.State.Intent,
			State:     req.State,
			Completed: true,
		})
		return
	case "analyze_trial_balance":
		orgID := middleware.GetOrgID(c)
		asOf := strings.TrimSpace(req.State.Fields["as_of"])
		if asOf == "" {
			asOf = time.Now().Format("2006-01-02")
			req.State.Fields["as_of"] = asOf
		}
		dr, cr, top, err := a.fetchTrialBalanceSummary(orgID, asOf)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		status := "balanced"
		diff := dr - cr
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.01 {
			status = "out of balance"
		}
		topText := "No account activity yet."
		if len(top) > 0 {
			topText = "Top balances: " + strings.Join(top, "; ") + "."
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: assistmd.KVBlock(
				fmt.Sprintf("**Trial balance** · `%s` · _%s_", asOf, status),
				[][2]string{
					{"Debit total", fmt.Sprintf("%.2f", dr)},
					{"Credit total", fmt.Sprintf("%.2f", cr)},
					{"Top accounts", topText},
				},
			),
			Intent:    req.State.Intent,
			State:     req.State,
			Completed: true,
		})
		return
	case "analyze_account":
		orgID := middleware.GetOrgID(c)
		accountHint := strings.TrimSpace(req.State.Fields["account"])
		if accountHint == "" {
			accountHint = extractAccountHint(userMsg)
			if accountHint != "" {
				req.State.Fields["account"] = accountHint
			}
		}
		if accountHint == "" {
			c.JSON(http.StatusOK, assistantChatResponse{
				AssistantMessage: "Tell me which account to analyze (code or name), for example: analyze account 1000 from 2026-01-01 to 2026-03-31.",
				Intent:           req.State.Intent,
				MissingFields:    []string{"account"},
				State:            req.State,
				Completed:        false,
			})
			return
		}
		accountID, accountCode, accountName, err := a.resolveAccount(orgID, accountHint)
		if err != nil {
			c.JSON(http.StatusOK, assistantChatResponse{
				AssistantMessage: fmt.Sprintf("I could not find account %q. Please provide a valid account code or exact name.", accountHint),
				Intent:           req.State.Intent,
				MissingFields:    []string{"account"},
				State:            req.State,
				Completed:        false,
			})
			return
		}
		from, to := assistantDateRange(req.State.Fields, time.Now())
		opening, debit, credit, movement, closing, count, err := a.fetchAccountMovement(orgID, accountID, from, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: assistmd.KVBlock(
				fmt.Sprintf("**Account %s (%s)** · `%s` → `%s`", accountCode, accountName, from, to),
				[][2]string{
					{"Opening", fmt.Sprintf("%.2f", opening)},
					{"Debit", fmt.Sprintf("%.2f", debit)},
					{"Credit", fmt.Sprintf("%.2f", credit)},
					{"Net movement", fmt.Sprintf("%.2f", movement)},
					{"Closing", fmt.Sprintf("%.2f", closing)},
					{"Lines", fmt.Sprintf("%d", count)},
				},
			),
			Intent:    req.State.Intent,
			State:     req.State,
			Completed: true,
		})
		return
	case "analyze_flow_log":
		orgID := middleware.GetOrgID(c)
		from, to := assistantDateRange(req.State.Fields, time.Now())
		prompt := strings.TrimSpace(req.State.Fields["prompt"])
		if prompt == "" {
			prompt = strings.TrimSpace(userMsg)
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
		var report string
		aiEnabled := false
		if a.AI != nil && a.AI.Configured() {
			report, err = a.generateFlowLogAIReport(c.Request.Context(), from, to, prompt, summary, entries)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			aiEnabled = true
		} else {
			report = buildFlowLogFallbackReport(from, to, summary, entries)
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: report,
			Intent:           req.State.Intent,
			State:            assistantState{Intent: "general", Fields: map[string]string{}},
			Completed:        true,
			Record:           gin.H{"from": from, "to": to, "entry_count": len(entries), "ai_enabled": aiEnabled},
		})
		return
	case "calculate":
		expr := strings.TrimSpace(req.State.Fields["expression"])
		if expr == "" {
			expr = extractExpression(userMsg)
			if expr != "" {
				req.State.Fields["expression"] = expr
			}
		}
		if expr == "" {
			c.JSON(http.StatusOK, assistantChatResponse{
				AssistantMessage: "Share the expression to calculate, for example: (12000 - 4500) * 0.1",
				Intent:           req.State.Intent,
				MissingFields:    []string{"expression"},
				State:            req.State,
				Completed:        false,
			})
			return
		}
		result, err := evalExpression(expr)
		if err != nil {
			c.JSON(http.StatusOK, assistantChatResponse{
				AssistantMessage: "I couldn't parse that expression. Please use numbers with +, -, *, /, and parentheses.",
				Intent:           req.State.Intent,
				MissingFields:    []string{"expression"},
				State:            req.State,
				Completed:        false,
			})
			return
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: fmt.Sprintf("**Calculation**\n\n`%s` = **%.2f**", expr, result),
			Intent:           req.State.Intent,
			State:            assistantState{Intent: "general", Fields: map[string]string{}},
			Completed:        true,
		})
		return
	default:
		a.respondGeneralAssistant(c, req, userMsg)
	}
}

func (a *API) assistantCreateCustomer(c *gin.Context, fields map[string]string) (gin.H, error) {
	orgID := middleware.GetOrgID(c)
	name := strings.TrimSpace(fields["name"])
	email := strings.TrimSpace(fields["email"])
	phone := strings.TrimSpace(fields["phone"])
	res, err := a.DB.Exec(`INSERT INTO customers (organization_id, name, email, phone) VALUES (?,?,?,?)`, orgID, name, email, phone)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return gin.H{"id": id, "name": name, "email": email, "phone": phone}, nil
}

func (a *API) assistantCreateBooking(c *gin.Context, fields map[string]string) (gin.H, error) {
	orgID := middleware.GetOrgID(c)
	uid := middleware.GetUserID(c)
	customerID, err := strconv.ParseInt(strings.TrimSpace(fields["customer_id"]), 10, 64)
	if err != nil || customerID <= 0 {
		return nil, fmt.Errorf("invalid customer_id")
	}
	bookingDate := strings.TrimSpace(fields["date"])
	if bookingDate == "" {
		bookingDate = time.Now().Format("2006-01-02")
	}
	bookingNumber := fmt.Sprintf("BK-%d", time.Now().Unix())
	currency := "USD"
	status := "draft"

	tx, err := a.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO bookings (organization_id, customer_id, booking_number, status, currency, subtotal, tax, total, booking_date, due_date, notes, created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, orgID, customerID, bookingNumber, status, currency, 0, 0, 0, bookingDate, nil, "", uid)
	if err != nil {
		return nil, err
	}
	bid, _ := res.LastInsertId()
	if _, err := tx.Exec(`INSERT INTO booking_items (booking_id, product_id, description, quantity, unit_price, line_total) VALUES (?,?,?,?,?,?)`,
		bid, nil, "Draft booking (assistant)", 1, 0, 0); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return gin.H{
		"booking_id":      bid,
		"booking_number":  bookingNumber,
		"customer_id":     customerID,
		"booking_date":    bookingDate,
		"status":          status,
	}, nil
}

func extractJSONObject(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") {
		return "", false
	}
	depth := 0
	for i, ch := range s {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[:i+1], true
			}
		}
	}
	return "", false
}

func (a *API) handleAssistantListCustomers(c *gin.Context, st assistantState) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, name, email, phone FROM customers WHERE organization_id=? ORDER BY name LIMIT 100`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var name, email, phone string
		if err := rows.Scan(&id, &name, &email, &phone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, gin.H{"id": id, "name": name, "email": email, "phone": phone})
	}
	msg := assistmd.Empty("No customers yet.")
	if len(list) > 0 {
		items := make([]string, 0, len(list))
		for _, row := range list {
			line := assistmd.NamedID(fmt.Sprint(row["name"]), row["id"])
			if email := fmt.Sprint(row["email"]); email != "" {
				line += fmt.Sprintf(" — `%s`", email)
			}
			items = append(items, line)
		}
		msg = assistmd.BulletList(fmt.Sprintf("**%d customer(s)**", len(list)), items)
	}
	c.JSON(http.StatusOK, assistantChatResponse{
		AssistantMessage: msg,
		Intent:           "list_customers",
		State:            assistantState{Intent: "general", Fields: map[string]string{}},
		Completed:        true,
		Record:           gin.H{"customers": list},
	})
}

func (a *API) handleAssistantListBookings(c *gin.Context, st assistantState) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, customer_id, booking_number, status, total, booking_date
		FROM bookings WHERE organization_id=? ORDER BY id DESC LIMIT 50`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var custNull sql.NullInt64
		var num, stat string
		var total float64
		var bd time.Time
		if err := rows.Scan(&id, &custNull, &num, &stat, &total, &bd); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h := gin.H{"id": id, "booking_number": num, "status": stat, "total": total, "booking_date": bd.Format("2006-01-02")}
		if custNull.Valid {
			h["customer_id"] = custNull.Int64
		}
		list = append(list, h)
	}
	msg := assistmd.Empty("No bookings yet.")
	if len(list) > 0 {
		items := make([]string, 0, min(len(list), 12))
		for i, row := range list {
			if i >= 12 {
				break
			}
			items = append(items, fmt.Sprintf("**%v** — %v · `%v` · **%.2f**",
				row["booking_number"], row["status"], row["booking_date"], row["total"]))
		}
		intro := fmt.Sprintf("**%d recent booking(s)**", len(list))
		if len(list) > 12 {
			intro += fmt.Sprintf(" _(showing 12 of %d)_", len(list))
		}
		msg = assistmd.BulletList(intro, items)
	}
	c.JSON(http.StatusOK, assistantChatResponse{
		AssistantMessage: msg,
		Intent:           "list_bookings",
		State:            assistantState{Intent: "general", Fields: map[string]string{}},
		Completed:        true,
		Record:           gin.H{"bookings": list},
	})
}

func (a *API) handleAssistantListAccounts(c *gin.Context, st assistantState) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, code, name, type FROM accounts WHERE organization_id=? ORDER BY code LIMIT 200`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var code, name, typ string
		if err := rows.Scan(&id, &code, &name, &typ); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, gin.H{"id": id, "code": code, "name": name, "type": typ})
	}
	msg := assistmd.Empty("No accounts in the chart yet.")
	if len(list) > 0 {
		items := make([]string, 0, min(len(list), 15))
		for i, row := range list {
			if i >= 15 {
				break
			}
			items = append(items, fmt.Sprintf("**%v** %v — _%v_", row["code"], row["name"], row["type"]))
		}
		intro := fmt.Sprintf("**Chart of accounts** (%d)", len(list))
		if len(list) > 15 {
			intro += fmt.Sprintf(" _(showing 15 of %d)_", len(list))
		}
		msg = assistmd.BulletList(intro, items)
	}
	c.JSON(http.StatusOK, assistantChatResponse{
		AssistantMessage: msg,
		Intent:           "list_accounts",
		State:            assistantState{Intent: "general", Fields: map[string]string{}},
		Completed:        true,
		Record:           gin.H{"accounts": list},
	})
}

func latestUserMessage(messages []assistantMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			return strings.TrimSpace(messages[i].Content)
		}
	}
	return ""
}

func detectIntent(message string) string {
	low := strings.ToLower(message)
	switch {
	case strings.Contains(low, "create customer"), strings.Contains(low, "new customer"):
		return "create_customer"
	case strings.Contains(low, "create booking"), strings.Contains(low, "new booking"):
		return "create_booking"
	case strings.Contains(low, "list customers"), strings.Contains(low, "show customers"), strings.Contains(low, "list customer"):
		return "list_customers"
	case strings.Contains(low, "list bookings"), strings.Contains(low, "show bookings"), strings.Contains(low, "list booking"):
		return "list_bookings"
	case strings.Contains(low, "list accounts"), strings.Contains(low, "show accounts"), strings.Contains(low, "chart of accounts"):
		return "list_accounts"
	case strings.Contains(low, "flow log"), strings.Contains(low, "flowlog"), strings.Contains(low, "money log"), strings.Contains(low, "spending log"):
		return "analyze_flow_log"
	case strings.Contains(low, "profit"), strings.Contains(low, "loss"), strings.Contains(low, "p&l"), strings.Contains(low, "net income"):
		return "analyze_profit_loss"
	case strings.Contains(low, "trial balance"):
		return "analyze_trial_balance"
	case strings.Contains(low, "analyze account"), strings.Contains(low, "account balance"), strings.Contains(low, "ledger movement"):
		return "analyze_account"
	case strings.Contains(low, "calculate"), strings.Contains(low, "compute"), looksLikeExpression(message):
		return "calculate"
	default:
		return "general"
	}
}

func reconcileAssistantIntent(st *assistantState, lastUser string) {
	detected := detectIntent(lastUser)
	switch detected {
	case "list_customers", "list_bookings", "list_accounts",
		"analyze_profit_loss", "analyze_trial_balance", "analyze_account", "analyze_flow_log":
		st.Intent = detected
		st.Fields = map[string]string{}
	case "create_customer", "create_booking", "calculate":
		if st.Intent != detected {
			st.Fields = map[string]string{}
		}
		st.Intent = detected
	case "general":
		if st.Intent == "calculate" {
			st.Intent = "general"
			st.Fields = map[string]string{}
		} else if st.Intent == "" {
			st.Intent = "general"
		}
	default:
		if st.Intent == "" {
			st.Intent = detected
		}
	}
}

func updateFields(fields map[string]string, message string) {
	for _, token := range []string{"name", "email", "phone", "customer_id", "date", "note", "from", "to", "as_of", "account", "expression"} {
		if v := extractValue(message, token); v != "" {
			fields[token] = v
		}
	}
}

func extractValue(message, token string) string {
	low := strings.ToLower(message)
	idx := strings.Index(low, token)
	if idx < 0 {
		return ""
	}
	sub := strings.TrimSpace(message[idx+len(token):])
	sub = strings.TrimSpace(strings.TrimLeft(sub, ":=- "))
	if cut := strings.IndexAny(sub, ",;\n"); cut > 0 {
		sub = sub[:cut]
	}
	return strings.TrimSpace(sub)
}

func missingFields(fields map[string]string, required []string) []string {
	out := make([]string, 0)
	for _, field := range required {
		if strings.TrimSpace(fields[field]) == "" {
			out = append(out, field)
		}
	}
	return out
}

func assistantDateRange(fields map[string]string, now time.Time) (string, string) {
	from := strings.TrimSpace(fields["from"])
	to := strings.TrimSpace(fields["to"])
	if from != "" && to != "" {
		return from, to
	}
	startMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if from == "" {
		from = startMonth.Format("2006-01-02")
	}
	if to == "" {
		to = now.Format("2006-01-02")
	}
	fields["from"] = from
	fields["to"] = to
	return from, to
}

func applyDateHints(fields map[string]string, message string, now time.Time) {
	low := strings.ToLower(message)
	if strings.Contains(low, "this month") {
		fields["from"] = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		fields["to"] = now.Format("2006-01-02")
	}
	if strings.Contains(low, "last month") {
		firstThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		lastMonthEnd := firstThisMonth.AddDate(0, 0, -1)
		lastMonthStart := time.Date(lastMonthEnd.Year(), lastMonthEnd.Month(), 1, 0, 0, 0, 0, now.Location())
		fields["from"] = lastMonthStart.Format("2006-01-02")
		fields["to"] = lastMonthEnd.Format("2006-01-02")
	}
	if strings.Contains(low, "this year") {
		fields["from"] = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		fields["to"] = now.Format("2006-01-02")
	}
	if strings.Contains(low, "today") {
		fields["as_of"] = now.Format("2006-01-02")
	}
	dates := regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`).FindAllString(message, -1)
	if len(dates) >= 1 {
		if strings.Contains(low, "as of") || strings.Contains(low, "as_of") {
			fields["as_of"] = dates[0]
		} else if fields["from"] == "" {
			fields["from"] = dates[0]
			if len(dates) >= 2 {
				fields["to"] = dates[1]
			}
		}
	}
}

func (a *API) fetchProfitLoss(orgID int64, from string, to string) (float64, float64, error) {
	rows, err := a.DB.Query(`
		SELECT a.type, COALESCE(SUM(jl.credit - jl.debit),0) AS amt
		FROM accounts a
		LEFT JOIN journal_lines jl ON jl.account_id = a.id
		LEFT JOIN journal_entries je ON je.id = jl.journal_entry_id
			AND je.organization_id = a.organization_id
			AND je.entry_date BETWEEN ? AND ?
		WHERE a.organization_id = ? AND a.type IN ('revenue','expense')
		GROUP BY a.type`, from, to, orgID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	var rev float64
	var exp float64
	for rows.Next() {
		var typ string
		var amt float64
		if err := rows.Scan(&typ, &amt); err != nil {
			return 0, 0, err
		}
		if typ == "revenue" {
			rev += amt
		} else {
			exp += -amt
		}
	}
	return rev, exp, rows.Err()
}

type trialAccountRow struct {
	Code    string
	Name    string
	Type    string
	Balance float64
}

func (a *API) fetchTrialBalanceSummary(orgID int64, asOf string) (float64, float64, []string, error) {
	rows, err := a.DB.Query(`
		SELECT a.code, a.name, a.type,
			COALESCE((
				SELECT SUM(jl.debit - jl.credit)
				FROM journal_lines jl
				INNER JOIN journal_entries je ON je.id = jl.journal_entry_id
				WHERE jl.account_id = a.id AND je.organization_id = a.organization_id AND je.entry_date <= ?
			), 0) AS balance
		FROM accounts a
		WHERE a.organization_id = ?`, asOf, orgID)
	if err != nil {
		return 0, 0, nil, err
	}
	defer rows.Close()
	var debitTotal float64
	var creditTotal float64
	var top []trialAccountRow
	for rows.Next() {
		var row trialAccountRow
		if err := rows.Scan(&row.Code, &row.Name, &row.Type, &row.Balance); err != nil {
			return 0, 0, nil, err
		}
		d, c := toTrialColumns(row.Type, row.Balance)
		debitTotal += d
		creditTotal += c
		if row.Balance != 0 {
			top = append(top, row)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, nil, err
	}
	sort.Slice(top, func(i, j int) bool {
		ai := top[i].Balance
		if ai < 0 {
			ai = -ai
		}
		aj := top[j].Balance
		if aj < 0 {
			aj = -aj
		}
		return ai > aj
	})
	limit := 3
	if len(top) < limit {
		limit = len(top)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, fmt.Sprintf("%s %s %.2f", top[i].Code, top[i].Name, top[i].Balance))
	}
	return debitTotal, creditTotal, out, nil
}

func toTrialColumns(accountType string, balance float64) (float64, float64) {
	switch strings.ToLower(strings.TrimSpace(accountType)) {
	case "asset", "expense":
		if balance >= 0 {
			return balance, 0
		}
		return 0, -balance
	default:
		if balance <= 0 {
			return -balance, 0
		}
		return 0, balance
	}
}

func (a *API) resolveAccount(orgID int64, hint string) (int64, string, string, error) {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return 0, "", "", fmt.Errorf("empty account hint")
	}
	var id int64
	var code, name string
	if n, err := strconv.ParseInt(hint, 10, 64); err == nil && n > 0 {
		if err := a.DB.QueryRow(`SELECT id, code, name FROM accounts WHERE organization_id=? AND id=?`, orgID, n).Scan(&id, &code, &name); err == nil {
			return id, code, name, nil
		}
	}
	if err := a.DB.QueryRow(`SELECT id, code, name FROM accounts WHERE organization_id=? AND code=?`, orgID, hint).Scan(&id, &code, &name); err == nil {
		return id, code, name, nil
	}
	if err := a.DB.QueryRow(`SELECT id, code, name FROM accounts WHERE organization_id=? AND LOWER(name)=LOWER(?)`, orgID, hint).Scan(&id, &code, &name); err == nil {
		return id, code, name, nil
	}
	return 0, "", "", fmt.Errorf("not found")
}

func (a *API) fetchAccountMovement(orgID int64, accountID int64, from string, to string) (float64, float64, float64, float64, float64, int64, error) {
	var opening float64
	err := a.DB.QueryRow(`
		SELECT COALESCE(SUM(jl.debit - jl.credit),0)
		FROM journal_lines jl
		INNER JOIN journal_entries je ON je.id = jl.journal_entry_id
		WHERE je.organization_id=? AND jl.account_id=? AND je.entry_date < ?`,
		orgID, accountID, from).Scan(&opening)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}

	var debit float64
	var credit float64
	var movement float64
	var lineCount int64
	err = a.DB.QueryRow(`
		SELECT COALESCE(SUM(jl.debit),0), COALESCE(SUM(jl.credit),0), COALESCE(SUM(jl.debit - jl.credit),0), COUNT(*)
		FROM journal_lines jl
		INNER JOIN journal_entries je ON je.id = jl.journal_entry_id
		WHERE je.organization_id=? AND jl.account_id=? AND je.entry_date BETWEEN ? AND ?`,
		orgID, accountID, from, to).Scan(&debit, &credit, &movement, &lineCount)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	closing := opening + movement
	return opening, debit, credit, movement, closing, lineCount, nil
}

func extractAccountHint(message string) string {
	re := regexp.MustCompile(`(?i)\baccount\s+([a-z0-9][a-z0-9 \-_]{0,60})`)
	m := re.FindStringSubmatch(message)
	if len(m) >= 2 {
		hint := strings.TrimSpace(strings.TrimRight(m[1], ".,;:!?"))
		parts := regexp.MustCompile(`(?i)\s+(from|to|as|for|last|this)\b`).Split(hint, 2)
		return strings.TrimSpace(parts[0])
	}
	return ""
}

func looksLikeExpression(message string) bool {
	var hasDigit bool
	var hasMath bool
	for _, r := range message {
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if strings.ContainsRune("+-*/()%", r) {
			hasMath = true
		}
	}
	return hasDigit && hasMath
}

func extractExpression(message string) string {
	if !looksLikeExpression(message) {
		return ""
	}
	var b strings.Builder
	for _, r := range message {
		if unicode.IsDigit(r) || unicode.IsSpace(r) || strings.ContainsRune("+-*/().%", r) {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func evalExpression(expr string) (float64, error) {
	rePercent := regexp.MustCompile(`(\d+(?:\.\d+)?)%`)
	expr = rePercent.ReplaceAllString(expr, "($1/100)")
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	rpn, err := toRPN(tokens)
	if err != nil {
		return 0, err
	}
	return evalRPN(rpn)
}

func tokenize(expr string) ([]string, error) {
	out := make([]string, 0, len(expr))
	i := 0
	for i < len(expr) {
		ch := expr[i]
		if ch == ' ' || ch == '\t' || ch == '\n' {
			i++
			continue
		}
		if strings.ContainsRune("+-*/()", rune(ch)) {
			out = append(out, string(ch))
			i++
			continue
		}
		if (ch >= '0' && ch <= '9') || ch == '.' {
			start := i
			i++
			for i < len(expr) && ((expr[i] >= '0' && expr[i] <= '9') || expr[i] == '.') {
				i++
			}
			out = append(out, expr[start:i])
			continue
		}
		return nil, fmt.Errorf("invalid character")
	}
	return out, nil
}

func toRPN(tokens []string) ([]string, error) {
	out := make([]string, 0, len(tokens))
	stack := make([]string, 0)
	prec := func(op string) int {
		if op == "+" || op == "-" {
			return 1
		}
		if op == "*" || op == "/" {
			return 2
		}
		return 0
	}
	for _, t := range tokens {
		if _, err := strconv.ParseFloat(t, 64); err == nil {
			out = append(out, t)
			continue
		}
		switch t {
		case "+", "-", "*", "/":
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top == "(" || prec(top) < prec(t) {
					break
				}
				out = append(out, top)
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, t)
		case "(":
			stack = append(stack, t)
		case ")":
			found := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top == "(" {
					found = true
					break
				}
				out = append(out, top)
			}
			if !found {
				return nil, fmt.Errorf("mismatched parentheses")
			}
		default:
			return nil, fmt.Errorf("invalid token")
		}
	}
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == "(" || stack[i] == ")" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		out = append(out, stack[i])
	}
	return out, nil
}

func evalRPN(tokens []string) (float64, error) {
	stack := make([]float64, 0, len(tokens))
	for _, t := range tokens {
		if v, err := strconv.ParseFloat(t, 64); err == nil {
			stack = append(stack, v)
			continue
		}
		if len(stack) < 2 {
			return 0, fmt.Errorf("invalid expression")
		}
		b := stack[len(stack)-1]
		a := stack[len(stack)-2]
		stack = stack[:len(stack)-2]
		switch t {
		case "+":
			stack = append(stack, a+b)
		case "-":
			stack = append(stack, a-b)
		case "*":
			stack = append(stack, a*b)
		case "/":
			if b == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			stack = append(stack, a/b)
		default:
			return 0, fmt.Errorf("unknown operator")
		}
	}
	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid expression")
	}
	return stack[0], nil
}
