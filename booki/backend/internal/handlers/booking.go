package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (a *API) ListBookings(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, customer_id, booking_number, status, currency, subtotal, tax, total, booking_date, due_date
		FROM bookings WHERE organization_id=? ORDER BY id DESC LIMIT 100`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var cust sql.NullInt64
		var num, stat, cur string
		var sub, tax, tot float64
		var bd time.Time
		var due sql.NullTime
		_ = rows.Scan(&id, &cust, &num, &stat, &cur, &sub, &tax, &tot, &bd, &due)
		h := gin.H{"id": id, "booking_number": num, "status": stat, "currency": cur, "subtotal": sub, "tax": tax, "total": tot, "booking_date": bd.Format("2006-01-02")}
		if cust.Valid {
			h["customer_id"] = cust.Int64
		}
		if due.Valid {
			h["due_date"] = due.Time.Format("2006-01-02")
		}
		list = append(list, h)
	}
	c.JSON(http.StatusOK, gin.H{"bookings": list})
}

func (a *API) GetBooking(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")
	var bid int64
	var cust sql.NullInt64
	var num, stat, cur string
	var sub, tax, tot float64
	var bd time.Time
	var due sql.NullTime
	var notes sql.NullString
	err := a.DB.QueryRow(`SELECT id, customer_id, booking_number, status, currency, subtotal, tax, total, booking_date, due_date, notes
		FROM bookings WHERE id=? AND organization_id=?`, id, orgID).
		Scan(&bid, &cust, &num, &stat, &cur, &sub, &tax, &tot, &bd, &due, &notes)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	h := gin.H{
		"id": bid, "booking_number": num, "status": stat, "currency": cur, "subtotal": sub, "tax": tax, "total": tot,
		"booking_date": bd.Format("2006-01-02"),
	}
	if cust.Valid {
		h["customer_id"] = cust.Int64
	}
	if due.Valid {
		h["due_date"] = due.Time.Format("2006-01-02")
	}
	if notes.Valid {
		h["notes"] = notes.String
	}
	rows, err := a.DB.Query(`SELECT id, product_id, description, quantity, unit_price, line_total FROM booking_items WHERE booking_id=? ORDER BY id`, bid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var items []gin.H
	for rows.Next() {
		var iid int64
		var pid sql.NullInt64
		var desc string
		var qty, up, lt float64
		_ = rows.Scan(&iid, &pid, &desc, &qty, &up, &lt)
		line := gin.H{"id": iid, "description": desc, "quantity": qty, "unit_price": up, "line_total": lt}
		if pid.Valid {
			line["product_id"] = pid.Int64
		}
		items = append(items, line)
	}
	h["items"] = items
	c.JSON(http.StatusOK, gin.H{"booking": h})
}

func (a *API) CreateBooking(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	uid := middleware.GetUserID(c)
	var body struct {
		CustomerID    *int64  `json:"customer_id"`
		BookingNumber string  `json:"booking_number"`
		Status        string  `json:"status"`
		Currency      string  `json:"currency"`
		BookingDate   string  `json:"booking_date" binding:"required"`
		DueDate       *string `json:"due_date"`
		Notes         string  `json:"notes"`
		Items         []struct {
			ProductID   *int64  `json:"product_id"`
			Description string  `json:"description"`
			Quantity    float64 `json:"quantity" binding:"required"`
			UnitPrice   float64 `json:"unit_price" binding:"required"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.BookingNumber == "" {
		body.BookingNumber = fmt.Sprintf("BK-%d", time.Now().Unix())
	}
	if body.Currency == "" {
		body.Currency = "USD"
	}
	if body.Status == "" {
		body.Status = "draft"
	}
	var sub, taxPct float64
	_ = a.DB.QueryRow(`SELECT tax_percent FROM organizations WHERE id=?`, orgID).Scan(&taxPct)
	var lineTotals float64
	for _, it := range body.Items {
		lineTotals += it.Quantity * it.UnitPrice
	}
	sub = lineTotals
	tax := sub * (taxPct / 100)
	total := sub + tax

	tx, err := a.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()
	var due interface{}
	if body.DueDate != nil {
		due = *body.DueDate
	} else {
		due = nil
	}
	res, err := tx.Exec(`INSERT INTO bookings (organization_id, customer_id, booking_number, status, currency, subtotal, tax, total, booking_date, due_date, notes, created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, orgID, body.CustomerID, body.BookingNumber, body.Status, body.Currency, sub, tax, total, body.BookingDate, due, body.Notes, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bid, _ := res.LastInsertId()
	for _, it := range body.Items {
		lt := it.Quantity * it.UnitPrice
		if _, err := tx.Exec(`INSERT INTO booking_items (booking_id, product_id, description, quantity, unit_price, line_total) VALUES (?,?,?,?,?,?)`,
			bid, it.ProductID, it.Description, it.Quantity, it.UnitPrice, lt); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"booking_id": bid, "subtotal": sub, "tax": tax, "total": total})
}

func (a *API) UpdateBookingStatus(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := a.DB.Exec(`UPDATE bookings SET status=? WHERE id=? AND organization_id=?`, body.Status, id, orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) UpdateBooking(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")
	bid, err := strconv.ParseInt(id, 10, 64)
	if err != nil || bid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}
	var body struct {
		CustomerID    *int64  `json:"customer_id"`
		BookingNumber string  `json:"booking_number"`
		Currency      string  `json:"currency"`
		BookingDate   string  `json:"booking_date" binding:"required"`
		DueDate       *string `json:"due_date"`
		Notes         string  `json:"notes"`
		Items         []struct {
			ProductID   *int64  `json:"product_id"`
			Description string  `json:"description"`
			Quantity    float64 `json:"quantity" binding:"required"`
			UnitPrice   float64 `json:"unit_price" binding:"required"`
		} `json:"items" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var status, existingNum string
	err = a.DB.QueryRow(`SELECT status, booking_number FROM bookings WHERE id=? AND organization_id=?`, bid, orgID).Scan(&status, &existingNum)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	if status == "posted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot edit a posted booking"})
		return
	}
	num := body.BookingNumber
	if num == "" {
		num = existingNum
	} else if num != existingNum {
		var dup int
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM bookings WHERE organization_id=? AND booking_number=? AND id<>?`, orgID, num, bid).Scan(&dup)
		if dup > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "booking number already in use"})
			return
		}
	}
	cur := body.Currency
	if cur == "" {
		cur = "USD"
	}
	var taxPct float64
	_ = a.DB.QueryRow(`SELECT tax_percent FROM organizations WHERE id=?`, orgID).Scan(&taxPct)
	var lineTotals float64
	for _, it := range body.Items {
		lineTotals += it.Quantity * it.UnitPrice
	}
	sub := lineTotals
	taxAm := sub * (taxPct / 100)
	total := sub + taxAm

	tx, err := a.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()
	var due interface{}
	if body.DueDate != nil {
		due = *body.DueDate
	} else {
		due = nil
	}
	res, err := tx.Exec(`UPDATE bookings SET customer_id=?, booking_number=?, currency=?, subtotal=?, tax=?, total=?, booking_date=?, due_date=?, notes=?
		WHERE id=? AND organization_id=?`,
		body.CustomerID, num, cur, sub, taxAm, total, body.BookingDate, due, body.Notes, bid, orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	if _, err := tx.Exec(`DELETE FROM booking_items WHERE booking_id=?`, bid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, it := range body.Items {
		lt := it.Quantity * it.UnitPrice
		if _, err := tx.Exec(`INSERT INTO booking_items (booking_id, product_id, description, quantity, unit_price, line_total) VALUES (?,?,?,?,?,?)`,
			bid, it.ProductID, it.Description, it.Quantity, it.UnitPrice, lt); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking_id": bid, "subtotal": sub, "tax": taxAm, "total": total})
}

func (a *API) DeleteBooking(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	id := c.Param("id")
	var status string
	err := a.DB.QueryRow(`SELECT status FROM bookings WHERE id=? AND organization_id=?`, id, orgID).Scan(&status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	if status == "posted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete a posted booking"})
		return
	}
	res, err := a.DB.Exec(`DELETE FROM bookings WHERE id=? AND organization_id=?`, id, orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PostBookingToLedger: Dr AR, Cr Sales + Tax Payable (simplified)
func (a *API) PostBookingToLedger(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	uid := middleware.GetUserID(c)
	bid := c.Param("id")
	var sub, tax, total float64
	var status, currency string
	var bookingDate time.Time
	err := a.DB.QueryRow(`SELECT subtotal, tax, total, status, currency, booking_date FROM bookings WHERE id=? AND organization_id=?`, bid, orgID).
		Scan(&sub, &tax, &total, &status, &currency, &bookingDate)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	if status == "posted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "already posted"})
		return
	}
	var arID, salesID, taxID int64
	_ = a.DB.QueryRow(`SELECT id FROM accounts WHERE organization_id=? AND code='1100'`, orgID).Scan(&arID)
	_ = a.DB.QueryRow(`SELECT id FROM accounts WHERE organization_id=? AND code='4000'`, orgID).Scan(&salesID)
	_ = a.DB.QueryRow(`SELECT id FROM accounts WHERE organization_id=? AND code='2200'`, orgID).Scan(&taxID)
	if arID == 0 || salesID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "default accounts missing (1100 AR, 4000 Sales)"})
		return
	}
	if tax > 0.0001 && taxID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tax payable account missing (2200)"})
		return
	}
	tx, err := a.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()
	ref := fmt.Sprintf("BOOK-%s", bid)
	res, err := tx.Exec(`INSERT INTO journal_entries (organization_id, reference, entry_date, description, status, source, booking_id, created_by)
		VALUES (?,?,?,?, 'posted', 'booking', ?, ?)`, orgID, ref, bookingDate.Format("2006-01-02"), "Booking posted", bid, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jeID, _ := res.LastInsertId()
	if _, err := tx.Exec(`INSERT INTO journal_lines (journal_entry_id, account_id, debit, credit, note) VALUES (?,?,?,?,?)`, jeID, arID, total, 0, "Accounts receivable"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := tx.Exec(`INSERT INTO journal_lines (journal_entry_id, account_id, debit, credit, note) VALUES (?,?,?,?,?)`, jeID, salesID, 0, sub, "Sales revenue"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if tax > 0.0001 {
		if _, err := tx.Exec(`INSERT INTO journal_lines (journal_entry_id, account_id, debit, credit, note) VALUES (?,?,?,?,?)`, jeID, taxID, 0, tax, "Tax payable"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if _, err := tx.Exec(`UPDATE bookings SET status='posted' WHERE id=? AND organization_id=?`, bid, orgID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"journal_entry_id": jeID, "booking_id": bid})
}