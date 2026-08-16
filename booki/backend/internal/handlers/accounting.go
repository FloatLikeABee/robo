package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

type API struct {
	DB                  *sql.DB
	AI                  *morphai.Client
	UsersPanelBaseURL   string
}

func (a *API) PatchOrganization(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	var body struct {
		Name             *string  `json:"name"`
		Country          *string  `json:"country"`
		Currency         *string  `json:"currency"`
		FiscalYearStart  *string  `json:"fiscal_year_start"`
		TaxSystem        *string  `json:"tax_system"`
		BaseCurrency     *string  `json:"base_currency"`
		TaxPercent       *float64 `json:"tax_percent"`
		AccountingMethod *string  `json:"accounting_method"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q := `UPDATE organizations SET `
	args := []any{}
	parts := []string{}
	if body.Name != nil {
		parts = append(parts, "name = ?")
		args = append(args, *body.Name)
	}
	if body.Country != nil {
		parts = append(parts, "country = ?")
		args = append(args, *body.Country)
	}
	if body.Currency != nil {
		parts = append(parts, "currency = ?")
		args = append(args, *body.Currency)
	}
	if body.FiscalYearStart != nil {
		parts = append(parts, "fiscal_year_start = ?")
		args = append(args, *body.FiscalYearStart)
	}
	if body.TaxSystem != nil {
		parts = append(parts, "tax_system = ?")
		args = append(args, *body.TaxSystem)
	}
	if body.BaseCurrency != nil {
		parts = append(parts, "base_currency = ?")
		args = append(args, *body.BaseCurrency)
	}
	if body.TaxPercent != nil {
		parts = append(parts, "tax_percent = ?")
		args = append(args, *body.TaxPercent)
	}
	if body.AccountingMethod != nil {
		parts = append(parts, "accounting_method = ?")
		args = append(args, *body.AccountingMethod)
	}
	if len(parts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields"})
		return
	}
	q += strings.Join(parts, ", ") + ` WHERE id = ?`
	args = append(args, orgID)
	if _, err := a.DB.Exec(q, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) GetOrganization(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	var row struct {
		ID               int64
		Name             string
		Country          string
		Currency         string
		Fiscal           sql.NullString
		TaxSystem        string
		BaseCurrency     string
		TaxPercent       float64
		AccountingMethod string
	}
	err := a.DB.QueryRow(`SELECT id, name, country, currency, fiscal_year_start, tax_system, base_currency, tax_percent, accounting_method
		FROM organizations WHERE id=?`, orgID).
		Scan(&row.ID, &row.Name, &row.Country, &row.Currency, &row.Fiscal, &row.TaxSystem, &row.BaseCurrency, &row.TaxPercent, &row.AccountingMethod)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "org not found"})
		return
	}
	fys := ""
	if row.Fiscal.Valid {
		fys = row.Fiscal.String[:10]
	}
	c.JSON(http.StatusOK, gin.H{
		"id": row.ID, "name": row.Name, "country": row.Country, "currency": row.Currency,
		"fiscal_year_start": fys, "tax_system": row.TaxSystem, "base_currency": row.BaseCurrency,
		"tax_percent": row.TaxPercent, "accounting_method": row.AccountingMethod,
	})
}

func (a *API) ListAccounts(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, code, name, type, parent_id, is_system FROM accounts WHERE organization_id=? ORDER BY code`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var out []gin.H
	for rows.Next() {
		var id int64
		var code, name, typ string
		var parent sql.NullInt64
		var sys bool
		_ = rows.Scan(&id, &code, &name, &typ, &parent, &sys)
		h := gin.H{"id": id, "code": code, "name": name, "type": typ, "is_system": sys}
		if parent.Valid {
			h["parent_id"] = parent.Int64
		}
		out = append(out, h)
	}
	c.JSON(http.StatusOK, gin.H{"accounts": out})
}

type journalLineIn struct {
	AccountID int64   `json:"account_id" binding:"required"`
	Debit     float64 `json:"debit"`
	Credit    float64 `json:"credit"`
	Note      string  `json:"note"`
}

func (a *API) CreateJournalEntry(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	uid := middleware.GetUserID(c)
	var body struct {
		Reference   string          `json:"reference"`
		EntryDate   string          `json:"entry_date" binding:"required"`
		Description string          `json:"description"`
		Status      string          `json:"status"`
		Source      string          `json:"source"`
		Lines       []journalLineIn `json:"lines" binding:"required,min=2"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var deb, cred float64
	for _, ln := range body.Lines {
		if ln.Debit < 0 || ln.Credit < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amounts must be non-negative"})
			return
		}
		if ln.Debit > 0 && ln.Credit > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "line cannot be both debit and credit"})
			return
		}
		deb += ln.Debit
		cred += ln.Credit
	}
	if fmt.Sprintf("%.4f", deb) != fmt.Sprintf("%.4f", cred) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "journal must balance (debits = credits)", "debits": deb, "credits": cred})
		return
	}
	status := body.Status
	if status == "" {
		status = "posted"
	}
	src := body.Source
	if src == "" {
		src = "manual"
	}
	tx, err := a.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO journal_entries (organization_id, reference, entry_date, description, status, source, created_by)
		VALUES (?,?,?,?,?,?,?)`, orgID, body.Reference, body.EntryDate, body.Description, status, src, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jeID, _ := res.LastInsertId()
	for _, ln := range body.Lines {
		var cnt int
		_ = tx.QueryRow(`SELECT COUNT(*) FROM accounts WHERE id=? AND organization_id=?`, ln.AccountID, orgID).Scan(&cnt)
		if cnt != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid account %d", ln.AccountID)})
			return
		}
		if _, err := tx.Exec(`INSERT INTO journal_lines (journal_entry_id, account_id, debit, credit, note) VALUES (?,?,?,?,?)`,
			jeID, ln.AccountID, ln.Debit, ln.Credit, ln.Note); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"journal_entry_id": jeID})
}

func (a *API) ListJournalEntries(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, reference, entry_date, description, status, source FROM journal_entries WHERE organization_id=? ORDER BY entry_date DESC, id DESC LIMIT 200`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var ref, desc, stat, src string
		var dt time.Time
		_ = rows.Scan(&id, &ref, &dt, &desc, &stat, &src)
		list = append(list, gin.H{"id": id, "reference": ref, "entry_date": dt.Format("2006-01-02"), "description": desc, "status": stat, "source": src})
	}
	c.JSON(http.StatusOK, gin.H{"entries": list})
}

func (a *API) GetLedger(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	accountID := c.Query("account_id")
	from := c.Query("from")
	to := c.Query("to")
	if accountID == "" || from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id, from, to required (YYYY-MM-DD)"})
		return
	}
	q := `SELECT jl.id, je.entry_date, je.reference, jl.debit, jl.credit, jl.note
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		WHERE je.organization_id = ? AND jl.account_id = ? AND je.entry_date BETWEEN ? AND ?
		ORDER BY je.entry_date, jl.id`
	rows, err := a.DB.Query(q, orgID, accountID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	var run float64
	for rows.Next() {
		var id int64
		var ref, note string
		var dt time.Time
		var d, cr float64
		_ = rows.Scan(&id, &dt, &ref, &d, &cr, &note)
		run += d - cr
		list = append(list, gin.H{"line_id": id, "date": dt.Format("2006-01-02"), "reference": ref, "debit": d, "credit": cr, "balance": run, "note": note})
	}
	c.JSON(http.StatusOK, gin.H{"lines": list})
}

func (a *API) TrialBalance(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	asOf := c.DefaultQuery("as_of", time.Now().Format("2006-01-02"))
	rows, err := a.DB.Query(`
		SELECT a.id, a.code, a.name, a.type,
			COALESCE((
				SELECT SUM(jl.debit - jl.credit)
				FROM journal_lines jl
				INNER JOIN journal_entries je ON je.id = jl.journal_entry_id
				WHERE jl.account_id = a.id AND je.organization_id = a.organization_id AND je.entry_date <= ?
			), 0) AS balance
		FROM accounts a
		WHERE a.organization_id = ?
		ORDER BY a.code`, asOf, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var code, name, typ string
		var bal float64
		_ = rows.Scan(&id, &code, &name, &typ, &bal)
		list = append(list, gin.H{"account_id": id, "code": code, "name": name, "type": typ, "balance": bal})
	}
	c.JSON(http.StatusOK, gin.H{"as_of": asOf, "accounts": list})
}

func (a *API) ProfitLoss(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	from := c.DefaultQuery("from", time.Now().Format("2006-01")+"-01")
	to := c.DefaultQuery("to", time.Now().Format("2006-01-02"))
	rows, err := a.DB.Query(`
		SELECT a.type,
			COALESCE(SUM(jl.credit - jl.debit),0) AS amt
		FROM accounts a
		LEFT JOIN journal_lines jl ON jl.account_id = a.id
		LEFT JOIN journal_entries je ON je.id = jl.journal_entry_id AND je.organization_id = a.organization_id
			AND je.entry_date BETWEEN ? AND ?
		WHERE a.organization_id = ? AND a.type IN ('revenue','expense')
		GROUP BY a.type`, from, to, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var rev, exp float64
	for rows.Next() {
		var typ string
		var amt float64
		_ = rows.Scan(&typ, &amt)
		if typ == "revenue" {
			rev += amt
		} else {
			exp += -amt
		}
	}
	c.JSON(http.StatusOK, gin.H{"from": from, "to": to, "revenue": rev, "expenses": exp, "net_income": rev - exp})
}

func (a *API) OpeningBalances(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	uid := middleware.GetUserID(c)
	var body struct {
		EntryDate   string          `json:"entry_date" binding:"required"`
		Description string          `json:"description"`
		Lines       []journalLineIn `json:"lines" binding:"required,min=2"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Description == "" {
		body.Description = "Opening balances"
	}
	var deb, cred float64
	for _, ln := range body.Lines {
		if ln.Debit < 0 || ln.Credit < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amounts must be non-negative"})
			return
		}
		if ln.Debit > 0 && ln.Credit > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "line cannot be both debit and credit"})
			return
		}
		deb += ln.Debit
		cred += ln.Credit
	}
	if fmt.Sprintf("%.4f", deb) != fmt.Sprintf("%.4f", cred) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "opening entry must balance"})
		return
	}
	tx, err := a.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO journal_entries (organization_id, reference, entry_date, description, status, source, created_by)
		VALUES (?,?,?,?, 'posted', 'opening', ?)`, orgID, "OPENING", body.EntryDate, body.Description, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jeID, _ := res.LastInsertId()
	for _, ln := range body.Lines {
		var cnt int
		_ = tx.QueryRow(`SELECT COUNT(*) FROM accounts WHERE id=? AND organization_id=?`, ln.AccountID, orgID).Scan(&cnt)
		if cnt != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid account %d", ln.AccountID)})
			return
		}
		if _, err := tx.Exec(`INSERT INTO journal_lines (journal_entry_id, account_id, debit, credit, note) VALUES (?,?,?,?,?)`,
			jeID, ln.AccountID, ln.Debit, ln.Credit, ln.Note); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"journal_entry_id": jeID})
}
