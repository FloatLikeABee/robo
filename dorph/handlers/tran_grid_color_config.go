package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var defaultGridColorConfig = json.RawMessage(`{"version":1,"rules":[],"groups":[]}`)

// GetGridColorConfig returns JSON color-tag config for a grid (presentation only).
func (h *Handlers) GetGridColorConfig(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	gridKey := strings.TrimSpace(c.Query("grid_key"))
	if gridKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grid_key is required"})
		return
	}
	var raw sql.NullString
	err := h.TranMySQL.DB.QueryRow(
		"SELECT ConfigJSON FROM AdminGridColorConfig WHERE GridKey = ?",
		gridKey,
	).Scan(&raw)
	if err == sql.ErrNoRows || !raw.Valid || strings.TrimSpace(raw.String) == "" {
		c.Data(http.StatusOK, "application/json", defaultGridColorConfig)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !json.Valid([]byte(raw.String)) {
		c.Data(http.StatusOK, "application/json", defaultGridColorConfig)
		return
	}
	c.Data(http.StatusOK, "application/json", []byte(raw.String))
}

// PutGridColorConfig upserts shared color-tag config for a grid.
func (h *Handlers) PutGridColorConfig(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in struct {
		GridKey    string          `json:"grid_key"`
		ConfigJSON json.RawMessage `json:"config_json"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	gridKey := strings.TrimSpace(in.GridKey)
	if gridKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grid_key is required"})
		return
	}
	fj := strings.TrimSpace(string(in.ConfigJSON))
	if fj == "" {
		fj = string(defaultGridColorConfig)
	}
	if !json.Valid([]byte(fj)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "config_json must be valid JSON"})
		return
	}
	_, err := h.TranMySQL.DB.Exec(
		`INSERT INTO AdminGridColorConfig (GridKey, ConfigJSON) VALUES (?, ?)
		 ON DUPLICATE KEY UPDATE ConfigJSON = VALUES(ConfigJSON), UpdatedOn = CURRENT_TIMESTAMP`,
		gridKey, fj,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
