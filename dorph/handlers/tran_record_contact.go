package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

var validEntityTypes = map[string]struct{}{
	"student": {}, "school": {}, "trip": {}, "vehicle": {}, "staff": {}, "district": {},
}

// ListRecordContacts returns contacts linked to an entity. Query: entity_type, record_id, optional db_id.
func (h *Handlers) ListRecordContacts(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	entityType := strings.ToLower(strings.TrimSpace(c.Query("entity_type")))
	recordIDStr := c.Query("record_id")
	if entityType == "" || recordIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_type and record_id are required"})
		return
	}
	if _, ok := validEntityTypes[entityType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity_type"})
		return
	}
	recordID, err := strconv.Atoi(recordIDStr)
	if err != nil || recordID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid record_id"})
		return
	}
	dbid := 1
	if d := c.Query("db_id"); d != "" {
		if n, e := strconv.Atoi(d); e == nil {
			dbid = n
		}
	}

	rows, err := h.TranMySQL.DB.Query(
		`SELECT rc.ID, rc.EntityType, rc.RecordID, rc.ContactID, rc.Relationship, rc.IsPrimary
		 FROM record_contact rc
		 WHERE rc.DBID = ? AND rc.EntityType = ? AND rc.RecordID = ?
		 ORDER BY rc.IsPrimary DESC, rc.ID ASC`,
		dbid, entityType, recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.RecordContact
	for rows.Next() {
		var m models.RecordContact
		if err := rows.Scan(&m.ID, &m.EntityType, &m.RecordID, &m.ContactID, &m.Relationship, &m.IsPrimary); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, m)
	}
	c.JSON(http.StatusOK, list)
}

// CreateRecordContact links a contact to an entity. Body: entity_type, record_id, contact_id, relationship?, is_primary?.
func (h *Handlers) CreateRecordContact(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in struct {
		EntityType   string  `json:"entity_type"`
		RecordID     int     `json:"record_id"`
		ContactID    int     `json:"contact_id"`
		Relationship *string `json:"relationship"`
		IsPrimary    *bool   `json:"is_primary"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	entityType := strings.ToLower(strings.TrimSpace(in.EntityType))
	if _, ok := validEntityTypes[entityType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity_type"})
		return
	}
	if in.RecordID <= 0 || in.ContactID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "record_id and contact_id are required and must be positive"})
		return
	}
	dbid := 1
	isPrimary := false
	if in.IsPrimary != nil {
		isPrimary = *in.IsPrimary
	}
	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO record_contact (DBID, EntityType, RecordID, ContactID, Relationship, IsPrimary)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		dbid, entityType, in.RecordID, in.ContactID, in.Relationship, isPrimary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id64, "entity_type": entityType, "record_id": in.RecordID, "contact_id": in.ContactID})
}

// UpdateRecordContact updates relationship and is_primary for a link.
func (h *Handlers) UpdateRecordContact(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var in struct {
		Relationship *string `json:"relationship"`
		IsPrimary    *bool   `json:"is_primary"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var sets []string
	var args []interface{}
	if in.Relationship != nil {
		sets = append(sets, "Relationship = ?")
		args = append(args, *in.Relationship)
	}
	if in.IsPrimary != nil {
		sets = append(sets, "IsPrimary = ?")
		args = append(args, *in.IsPrimary)
	}
	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	args = append(args, id)
	stmt := "UPDATE record_contact SET " + strings.Join(sets, ", ") + " WHERE ID = ?"
	_, err = h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": id})
}

// DeleteRecordContact removes the link between an entity and a contact.
func (h *Handlers) DeleteRecordContact(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec("DELETE FROM record_contact WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "record_contact not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
