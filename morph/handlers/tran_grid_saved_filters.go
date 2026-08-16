package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListSavedGridFilters returns saved filters for a grid key (shared for all users).
func (h *Handlers) ListSavedGridFilters(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	gridKey := strings.TrimSpace(c.Query("grid_key"))
	if gridKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grid_key is required"})
		return
	}
	rows, err := h.TranMySQL.DB.Query(
		"SELECT ID, GridKey, Name, FilterJSON, CreatedOn FROM AdminGridSavedFilter WHERE GridKey = ? ORDER BY Name",
		gridKey,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var id int
		var gk, name, fj string
		var created sql.NullTime
		if err := rows.Scan(&id, &gk, &name, &fj, &created); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		item := gin.H{
			"id":          id,
			"grid_key":    gk,
			"name":        name,
			"filter_json": fj,
		}
		if created.Valid {
			item["created_on"] = created.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		list = append(list, item)
	}
	c.JSON(http.StatusOK, list)
}

// CreateSavedGridFilter inserts a shared saved filter (upsert on duplicate name for same grid).
func (h *Handlers) CreateSavedGridFilter(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in struct {
		GridKey    string          `json:"grid_key"`
		Name       string          `json:"name"`
		FilterJSON json.RawMessage `json:"filter_json"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	gridKey := strings.TrimSpace(in.GridKey)
	name := strings.TrimSpace(in.Name)
	if gridKey == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grid_key and name are required"})
		return
	}
	fj := strings.TrimSpace(string(in.FilterJSON))
	if fj == "" {
		fj = "{}"
	}
	if !json.Valid([]byte(fj)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filter_json must be valid JSON"})
		return
	}

	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO AdminGridSavedFilter (GridKey, Name, FilterJSON) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE FilterJSON = VALUES(FilterJSON), CreatedOn = CURRENT_TIMESTAMP`,
		gridKey, name, fj,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	if id64 == 0 {
		var id int
		_ = h.TranMySQL.DB.QueryRow(
			"SELECT ID FROM AdminGridSavedFilter WHERE GridKey = ? AND Name = ?",
			gridKey, name,
		).Scan(&id)
		c.JSON(http.StatusOK, gin.H{"id": id, "grid_key": gridKey, "name": name})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id64, "grid_key": gridKey, "name": name})
}

// DeleteSavedGridFilter removes a saved filter by ID.
func (h *Handlers) DeleteSavedGridFilter(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec("DELETE FROM AdminGridSavedFilter WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "saved filter not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
