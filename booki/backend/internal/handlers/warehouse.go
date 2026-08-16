package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
)

func warehouseRowToH(id int64, name, code string, loc, detail sql.NullString, rd sql.NullTime) gin.H {
	h := gin.H{"id": id, "name": name, "code": code}
	if loc.Valid {
		h["location"] = loc.String
	}
	if detail.Valid {
		h["detail"] = detail.String
	} else {
		h["detail"] = nil
	}
	if rd.Valid {
		h["record_date"] = rd.Time.Format("2006-01-02")
	} else {
		h["record_date"] = nil
	}
	return h
}

func (a *API) ListWarehouses(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, name, code, location, detail, record_date FROM warehouses WHERE organization_id=? ORDER BY id`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var name, code string
		var loc, detail sql.NullString
		var rd sql.NullTime
		_ = rows.Scan(&id, &name, &code, &loc, &detail, &rd)
		list = append(list, warehouseRowToH(id, name, code, loc, detail, rd))
	}
	c.JSON(http.StatusOK, gin.H{"warehouses": list})
}

func (a *API) GetWarehouse(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	wid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || wid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warehouse id"})
		return
	}
	var name, code string
	var loc, detail sql.NullString
	var rd sql.NullTime
	err = a.DB.QueryRow(`SELECT id, name, code, location, detail, record_date FROM warehouses WHERE id=? AND organization_id=?`, wid, orgID).
		Scan(&wid, &name, &code, &loc, &detail, &rd)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"warehouse": warehouseRowToH(wid, name, code, loc, detail, rd)})
}

func (a *API) CreateWarehouse(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	var body struct {
		Name       string `json:"name" binding:"required"`
		Code       string `json:"code" binding:"required"`
		Location   string `json:"location"`
		Detail     string `json:"detail"`
		RecordDate string `json:"record_date"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if t := ([]byte)(body.Detail); len(t) > 0 && !json.Valid(t) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "detail must be valid JSON"})
		return
	}
	rd := body.RecordDate
	if rd == "" {
		rd = time.Now().Format("2006-01-02")
	}
	var detailArg interface{}
	if body.Detail == "" {
		detailArg = nil
	} else {
		detailArg = body.Detail
	}
	res, err := a.DB.Exec(`INSERT INTO warehouses (organization_id, name, code, location, detail, record_date) VALUES (?,?,?,?,?,?)`,
		orgID, body.Name, body.Code, body.Location, detailArg, rd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *API) UpdateWarehouse(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	wid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || wid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warehouse id"})
		return
	}
	var body struct {
		Name       string `json:"name" binding:"required"`
		Code       string `json:"code" binding:"required"`
		Location   string `json:"location"`
		Detail     string `json:"detail"`
		RecordDate string `json:"record_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if t := ([]byte)(body.Detail); len(t) > 0 && !json.Valid(t) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "detail must be valid JSON"})
		return
	}
	var existingCode string
	err = a.DB.QueryRow(`SELECT code FROM warehouses WHERE id=? AND organization_id=?`, wid, orgID).Scan(&existingCode)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if body.Code != existingCode {
		var dup int
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM warehouses WHERE organization_id=? AND code=? AND id<>?`, orgID, body.Code, wid).Scan(&dup)
		if dup > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "warehouse code already in use"})
			return
		}
	}
	var detailArg interface{}
	if body.Detail == "" {
		detailArg = nil
	} else {
		detailArg = body.Detail
	}
	res, err := a.DB.Exec(`UPDATE warehouses SET name=?, code=?, location=?, detail=?, record_date=? WHERE id=? AND organization_id=?`,
		body.Name, body.Code, body.Location, detailArg, body.RecordDate, wid, orgID)
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

func (a *API) DeleteWarehouse(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	wid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || wid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warehouse id"})
		return
	}
	res, err := a.DB.Exec(`DELETE FROM warehouses WHERE id=? AND organization_id=?`, wid, orgID)
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

func productRowToH(id int64, sku, name, cat, unit, bc string, desc string, cp, sp float64, rt int, detail sql.NullString, rd sql.NullTime) gin.H {
	h := gin.H{"id": id, "sku": sku, "name": name, "description": desc, "category": cat, "unit": unit, "barcode": bc,
		"cost_price": cp, "selling_price": sp, "reorder_threshold": rt}
	if detail.Valid {
		h["detail"] = detail.String
	} else {
		h["detail"] = nil
	}
	if rd.Valid {
		h["record_date"] = rd.Time.Format("2006-01-02")
	} else {
		h["record_date"] = nil
	}
	return h
}

func (a *API) ListProducts(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, sku, name, description, category, unit, barcode, cost_price, selling_price, reorder_threshold, detail, record_date FROM products WHERE organization_id=? ORDER BY id`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int64
		var sku, name, cat, unit, bc, desc string
		var cp, sp float64
		var rt int
		var detail sql.NullString
		var rd sql.NullTime
		_ = rows.Scan(&id, &sku, &name, &desc, &cat, &unit, &bc, &cp, &sp, &rt, &detail, &rd)
		list = append(list, productRowToH(id, sku, name, cat, unit, bc, desc, cp, sp, rt, detail, rd))
	}
	c.JSON(http.StatusOK, gin.H{"products": list})
}

func (a *API) GetProduct(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || pid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var sku, name, cat, unit, bc, desc string
	var cp, sp float64
	var rt int
	var detail sql.NullString
	var rd sql.NullTime
	err = a.DB.QueryRow(`SELECT id, sku, name, description, category, unit, barcode, cost_price, selling_price, reorder_threshold, detail, record_date FROM products WHERE id=? AND organization_id=?`, pid, orgID).
		Scan(&pid, &sku, &name, &desc, &cat, &unit, &bc, &cp, &sp, &rt, &detail, &rd)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": productRowToH(pid, sku, name, cat, unit, bc, desc, cp, sp, rt, detail, rd)})
}

func (a *API) CreateProduct(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	var body struct {
		SKU              string  `json:"sku" binding:"required"`
		Name             string  `json:"name" binding:"required"`
		Description      string  `json:"description"`
		Category         string  `json:"category"`
		Unit             string  `json:"unit"`
		Barcode          string  `json:"barcode"`
		CostPrice        float64 `json:"cost_price"`
		SellingPrice     float64 `json:"selling_price"`
		ReorderThreshold int     `json:"reorder_threshold"`
		Detail           string  `json:"detail"`
		RecordDate       string  `json:"record_date"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if t := ([]byte)(body.Detail); len(t) > 0 && !json.Valid(t) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "detail must be valid JSON"})
		return
	}
	if body.Unit == "" {
		body.Unit = "ea"
	}
	rd := body.RecordDate
	if rd == "" {
		rd = time.Now().Format("2006-01-02")
	}
	var detailArg interface{}
	if body.Detail == "" {
		detailArg = nil
	} else {
		detailArg = body.Detail
	}
	res, err := a.DB.Exec(`INSERT INTO products (organization_id, sku, name, description, category, unit, barcode, cost_price, selling_price, reorder_threshold, detail, record_date)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		orgID, body.SKU, body.Name, body.Description, body.Category, body.Unit, body.Barcode, body.CostPrice, body.SellingPrice, body.ReorderThreshold, detailArg, rd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (a *API) UpdateProduct(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || pid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var body struct {
		SKU              string  `json:"sku" binding:"required"`
		Name             string  `json:"name" binding:"required"`
		Description      string  `json:"description"`
		Category         string  `json:"category"`
		Unit             string  `json:"unit"`
		Barcode          string  `json:"barcode"`
		CostPrice        float64 `json:"cost_price"`
		SellingPrice     float64 `json:"selling_price"`
		ReorderThreshold int     `json:"reorder_threshold"`
		Detail           string  `json:"detail"`
		RecordDate       string  `json:"record_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if t := ([]byte)(body.Detail); len(t) > 0 && !json.Valid(t) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "detail must be valid JSON"})
		return
	}
	var existingSKU string
	err = a.DB.QueryRow(`SELECT sku FROM products WHERE id=? AND organization_id=?`, pid, orgID).Scan(&existingSKU)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if body.SKU != existingSKU {
		var dup int
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE organization_id=? AND sku=? AND id<>?`, orgID, body.SKU, pid).Scan(&dup)
		if dup > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "SKU already in use"})
			return
		}
	}
	if body.Unit == "" {
		body.Unit = "ea"
	}
	var detailArg interface{}
	if body.Detail == "" {
		detailArg = nil
	} else {
		detailArg = body.Detail
	}
	res, err := a.DB.Exec(`UPDATE products SET sku=?, name=?, description=?, category=?, unit=?, barcode=?, cost_price=?, selling_price=?, reorder_threshold=?, detail=?, record_date=?
		WHERE id=? AND organization_id=?`,
		body.SKU, body.Name, body.Description, body.Category, body.Unit, body.Barcode, body.CostPrice, body.SellingPrice, body.ReorderThreshold, detailArg, body.RecordDate, pid, orgID)
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

func (a *API) DeleteProduct(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || pid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	res, err := a.DB.Exec(`DELETE FROM products WHERE id=? AND organization_id=?`, pid, orgID)
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

func (a *API) WarehouseStock(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`
		SELECT ws.warehouse_id, w.name, ws.product_id, p.sku, p.name, ws.quantity,
			date(COALESCE(li.last_in_ts, ws.updated_at)) AS stock_date
		FROM warehouse_stocks ws
		JOIN warehouses w ON w.id = ws.warehouse_id
		JOIN products p ON p.id = ws.product_id
		LEFT JOIN (
			SELECT warehouse_id, product_id, MAX(movement_date) AS last_in_ts
			FROM inventory_movements
			WHERE organization_id = ? AND movement_type = 'stock_in' AND quantity > 0
			GROUP BY warehouse_id, product_id
		) li ON li.warehouse_id = ws.warehouse_id AND li.product_id = ws.product_id
		WHERE ws.organization_id=?`, orgID, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var wid, pid int64
		var wname, sku, pname string
		var qty float64
		var sdate sql.NullString
		_ = rows.Scan(&wid, &wname, &pid, &sku, &pname, &qty, &sdate)
		row := gin.H{"warehouse_id": wid, "warehouse_name": wname, "product_id": pid, "sku": sku, "product_name": pname, "quantity": qty}
		if sdate.Valid {
			row["stock_date"] = sdate.String
		}
		list = append(list, row)
	}
	c.JSON(http.StatusOK, gin.H{"stocks": list})
}

func (a *API) StockIn(c *gin.Context) {
	a.stockMove(c, "stock_in", 1)
}

func (a *API) StockOut(c *gin.Context) {
	a.stockMove(c, "stock_out", -1)
}

func (a *API) stockMove(c *gin.Context, mType string, sign float64) {
	orgID := middleware.GetOrgID(c)
	var body struct {
		WarehouseID int64   `json:"warehouse_id" binding:"required"`
		ProductID   int64   `json:"product_id" binding:"required"`
		Quantity    float64 `json:"quantity" binding:"required"`
		Reference   string  `json:"reference"`
		StockDate   string  `json:"stock_date"` // optional YYYY-MM-DD (movement_date when recording stock)
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be positive"})
		return
	}
	delta := body.Quantity * sign
	var moveMoment interface{}
	if s := strings.TrimSpace(body.StockDate); s != "" {
		t, err := time.ParseInLocation("2006-01-02", s, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stock_date must be YYYY-MM-DD"})
			return
		}
		moveMoment = t
	}

	tx, err := a.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()
	var wcnt, pcnt int
	_ = tx.QueryRow(`SELECT COUNT(*) FROM warehouses WHERE id=? AND organization_id=?`, body.WarehouseID, orgID).Scan(&wcnt)
	_ = tx.QueryRow(`SELECT COUNT(*) FROM products WHERE id=? AND organization_id=?`, body.ProductID, orgID).Scan(&pcnt)
	if wcnt != 1 || pcnt != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warehouse or product"})
		return
	}
	if sign < 0 {
		var q float64
		_ = tx.QueryRow(`SELECT COALESCE(quantity,0) FROM warehouse_stocks WHERE warehouse_id=? AND product_id=?`, body.WarehouseID, body.ProductID).Scan(&q)
		if q+delta < -0.0001 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient stock"})
			return
		}
	}
	mvQty := body.Quantity * sign
	if moveMoment != nil {
		_, err = tx.Exec(`INSERT INTO inventory_movements (organization_id, warehouse_id, product_id, movement_type, quantity, reference, movement_date)
			VALUES (?,?,?,?,?,?,?)`, orgID, body.WarehouseID, body.ProductID, mType, mvQty, body.Reference, moveMoment)
	} else {
		_, err = tx.Exec(`INSERT INTO inventory_movements (organization_id, warehouse_id, product_id, movement_type, quantity, reference)
			VALUES (?,?,?,?,?,?)`, orgID, body.WarehouseID, body.ProductID, mType, mvQty, body.Reference)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = tx.Exec(`
		INSERT INTO warehouse_stocks (organization_id, warehouse_id, product_id, quantity)
		VALUES (?,?,?,?)
		ON CONFLICT(warehouse_id, product_id) DO UPDATE SET quantity = quantity + excluded.quantity`, orgID, body.WarehouseID, body.ProductID, delta)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) StockTransfer(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	var body struct {
		FromWarehouseID int64   `json:"from_warehouse_id" binding:"required"`
		ToWarehouseID   int64   `json:"to_warehouse_id" binding:"required"`
		ProductID       int64   `json:"product_id" binding:"required"`
		Quantity        float64 `json:"quantity" binding:"required"`
		Reference       string  `json:"reference"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.FromWarehouseID == body.ToWarehouseID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "warehouses must differ"})
		return
	}
	if body.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be positive"})
		return
	}
	tx, err := a.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()
	var q float64
	err = tx.QueryRow(`SELECT COALESCE(quantity,0) FROM warehouse_stocks WHERE warehouse_id=? AND product_id=?`, body.FromWarehouseID, body.ProductID).Scan(&q)
	if err == sql.ErrNoRows || q < body.Quantity-0.0001 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient stock at source"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = tx.Exec(`INSERT INTO inventory_movements (organization_id, warehouse_id, product_id, movement_type, quantity, reference, related_warehouse_id)
		VALUES (?,?,?,?,?,?,?)`, orgID, body.FromWarehouseID, body.ProductID, "transfer_out", -body.Quantity, body.Reference, body.ToWarehouseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = tx.Exec(`INSERT INTO inventory_movements (organization_id, warehouse_id, product_id, movement_type, quantity, reference, related_warehouse_id)
		VALUES (?,?,?,?,?,?,?)`, orgID, body.ToWarehouseID, body.ProductID, "transfer_in", body.Quantity, body.Reference, body.FromWarehouseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, _ = tx.Exec(`UPDATE warehouse_stocks SET quantity = quantity - ? WHERE warehouse_id=? AND product_id=?`, body.Quantity, body.FromWarehouseID, body.ProductID)
	_, err = tx.Exec(`
		INSERT INTO warehouse_stocks (organization_id, warehouse_id, product_id, quantity)
		VALUES (?,?,?,?)
		ON CONFLICT(warehouse_id, product_id) DO UPDATE SET quantity = quantity + excluded.quantity`, orgID, body.ToWarehouseID, body.ProductID, body.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *API) ListMovements(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	rows, err := a.DB.Query(`SELECT id, warehouse_id, product_id, movement_type, quantity, reference, movement_date
		FROM inventory_movements WHERE organization_id=? ORDER BY movement_date DESC LIMIT 200`, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id, wid, pid int64
		var mt, ref string
		var qty float64
		var ts time.Time
		_ = rows.Scan(&id, &wid, &pid, &mt, &qty, &ref, &ts)
		list = append(list, gin.H{"id": id, "warehouse_id": wid, "product_id": pid, "movement_type": mt, "quantity": qty, "reference": ref, "movement_date": ts.Format(time.RFC3339)})
	}
	c.JSON(http.StatusOK, gin.H{"movements": list})
}
