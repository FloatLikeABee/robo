package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (a *API) Dashboard(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	var bookings, pending, products, assets, wh int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM bookings WHERE organization_id=?`, orgID).Scan(&bookings)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM bookings WHERE organization_id=? AND status IN ('draft','pending')`, orgID).Scan(&pending)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE organization_id=?`, orgID).Scan(&products)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM assets WHERE organization_id=?`, orgID).Scan(&assets)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM warehouses WHERE organization_id=?`, orgID).Scan(&wh)

	var lowStock int
	_ = a.DB.QueryRow(`
		SELECT COUNT(*) FROM (
			SELECT ws.product_id FROM warehouse_stocks ws
			JOIN products p ON p.id = ws.product_id AND p.organization_id = ws.organization_id
			WHERE ws.organization_id = ? AND ws.quantity <= p.reorder_threshold AND p.reorder_threshold > 0
		) t`, orgID).Scan(&lowStock)

	c.JSON(http.StatusOK, gin.H{
		"bookings_total":    bookings,
		"bookings_pending":  pending,
		"products":          products,
		"assets":            assets,
		"warehouses":        wh,
		"low_stock_skus":    lowStock,
	})
}

func (a *API) AccountingGlossary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"terms": []gin.H{
			{"term": "Chart of accounts", "definition": "The complete list of accounts your organization uses, grouped as assets, liabilities, equity, revenue, and expenses."},
			{"term": "Double entry", "definition": "Every transaction has equal debits and credits across accounts; the books must always balance."},
			{"term": "Debit / Credit", "definition": "Debit increases assets and expenses, decreases liabilities, equity, and revenue; credit does the opposite."},
			{"term": "General ledger", "definition": "Running history of all journal entries and their lines per account."},
			{"term": "Journal entry", "definition": "A dated financial record with one or more lines, each with a debit or credit to a specific account."},
			{"term": "Trial balance", "definition": "Snapshot listing each account balance at a date; debits should equal credits across the full set."},
			{"term": "Opening balances", "definition": "Starting balances when you begin using the system—entered as a balanced journal entry."},
			{"term": "Fiscal year", "definition": "Your reporting year and key tax timeline; may not match the calendar year depending on country."},
			{"term": "Inventory valuation", "definition": "How stock is costed (for example FIFO or weighted average) before it hits the books."},
			{"term": "Accrual vs cash", "definition": "Accrual records revenue/expense when earned/incurred; cash records when money moves."},
		},
	})
}

func (a *API) ImportLogs(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, source_type, target_entity, status, imported_rows, failed_rows, message, created_at FROM import_logs WHERE organization_id=? ORDER BY id DESC LIMIT 50`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id, ir, fr int64
		var st, te, stat string
		var msg sql.NullString
		var ct time.Time
		_ = rows.Scan(&id, &st, &te, &stat, &ir, &fr, &msg, &ct)
		h := gin.H{"id": id, "source_type": st, "target_entity": te, "status": stat, "imported_rows": ir, "failed_rows": fr, "created_at": ct.Format(time.RFC3339)}
		if msg.Valid {
			h["message"] = msg.String
		}
		list = append(list, h)
	}
	c.JSON(http.StatusOK, gin.H{"imports": list})
}
