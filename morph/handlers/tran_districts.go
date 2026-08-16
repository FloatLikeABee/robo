package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

// ListDistricts returns all districts (optionally by DBID).
func (h *Handlers) ListDistricts(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	query := "SELECT ID, DistrictID, District, Name, description FROM District"
	var args []interface{}
	if dbid := c.Query("db_id"); dbid != "" {
		query += " WHERE DBID = ?"
		args = append(args, dbid)
	}
	query += " ORDER BY Name, District"

	rows, err := h.TranMySQL.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.District
	for rows.Next() {
		var d models.District
		var desc sql.NullString
		if err := rows.Scan(&d.ID, &d.DistrictID, &d.District, &d.Name, &desc); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if desc.Valid {
			s := desc.String
			d.Description = &s
		}
		list = append(list, d)
	}
	c.JSON(http.StatusOK, list)
}

// GetDistrict returns one district by ID (with Mongo detail if configured).
func (h *Handlers) GetDistrict(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, `SELECT ID AS id, DistrictID AS district_id, District AS district, Name AS name, description AS description FROM District WHERE ID = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok || m == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "district not found"})
		return
	}
	h.attachEntityDetail(c, entityKeyDistrict, id, m)
	c.JSON(http.StatusOK, m)
}

// CreateDistrict inserts a district. UI does not send id/db_id; DBID defaults to 1; DistrictID auto-generated if missing.
func (h *Handlers) CreateDistrict(c *gin.Context) {
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

	dbid := 1

	// Auto-generate DistrictID if missing/invalid
	var districtID int
	if v, ok := in["district_id"]; ok {
		switch t := v.(type) {
		case float64:
			districtID = int(t)
		case int:
			districtID = t
		}
	}
	if districtID <= 0 {
		_ = h.TranMySQL.DB.QueryRow("SELECT COALESCE(MAX(DistrictID), 0) + 1 FROM District WHERE DBID = ?", dbid).Scan(&districtID)
		if districtID <= 0 {
			districtID = 1
		}
	}

	allowed := map[string]struct{}{
		"District":    {},
		"Name":        {},
		"description": {},
	}

	cols := []string{"DBID", "DistrictID"}
	args := []interface{}{dbid, districtID}

	for k, v := range in {
		if k == "" {
			continue
		}
		lk := strings.ToLower(k)
		if lk == "id" || lk == "db_id" || lk == "districtid" || lk == "district_id" {
			continue
		}
		col := k
		switch lk {
		case "district":
			col = "District"
		case "name":
			col = "Name"
		case "description":
			col = "description"
		}

		if _, ok := allowed[col]; !ok {
			continue
		}
		cols = append(cols, col)
		args = append(args, v)
	}

	if len(cols) == 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to insert"})
		return
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(cols)), ",")
	stmt := "INSERT INTO District (" + strings.Join(cols, ",") + ") VALUES (" + placeholders + ")"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyDistrict, int(id64), detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, `SELECT ID AS id, DistrictID AS district_id, District AS district, Name AS name, description AS description FROM District WHERE ID = ?`, id64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id64})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyDistrict, int(id64), m)
	}
	c.JSON(http.StatusOK, m)
}

// UpdateDistrict updates an existing district.
func (h *Handlers) UpdateDistrict(c *gin.Context) {
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

	allowed := map[string]struct{}{
		"District":    {},
		"Name":        {},
		"description": {},
	}

	var sets []string
	var args []interface{}

	for k, v := range in {
		if k == "" {
			continue
		}
		lk := strings.ToLower(k)
		if lk == "id" || lk == "db_id" || lk == "districtid" || lk == "district_id" {
			continue
		}
		col := k
		switch lk {
		case "district":
			col = "District"
		case "name":
			col = "Name"
		case "description":
			col = "description"
		}

		if _, ok := allowed[col]; !ok {
			continue
		}
		sets = append(sets, col+" = ?")
		args = append(args, v)
	}

	if len(sets) == 0 {
		m, ok, _ := querySingleRowMap(h.TranMySQL.DB, `SELECT ID AS id, DistrictID AS district_id, District AS district, Name AS name, description AS description FROM District WHERE ID = ?`, id)
		if !ok || m == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "district not found"})
			return
		}
		if !hasDetail {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		_ = h.savePoppedDetail(c, entityKeyDistrict, id, detailStr)
		h.attachEntityDetail(c, entityKeyDistrict, id, m)
		c.JSON(http.StatusOK, m)
		return
	}

	args = append(args, id)
	stmt := "UPDATE District SET " + strings.Join(sets, ", ") + " WHERE ID = ?"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "district not found"})
		return
	}
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyDistrict, id, detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, `SELECT ID AS id, DistrictID AS district_id, District AS district, Name AS name, description AS description FROM District WHERE ID = ?`, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyDistrict, id, m)
	}
	c.JSON(http.StatusOK, m)
}

// DeleteDistrict deletes a district by ID.
func (h *Handlers) DeleteDistrict(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := h.TranMySQL.DB.Exec("DELETE FROM District WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "district not found"})
		return
	}
	h.deleteEntityDetailMongo(context.Background(), entityKeyDistrict, id)
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
