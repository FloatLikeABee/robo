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

const contactListCols = `ID, LastName, FirstName, Email, Phone, Mobile, description`

const contactFullSelectCols = `ID AS id, LastName AS last_name, FirstName AS first_name, Email AS email, Phone AS phone, Mobile AS mobile, description AS description`

var allowedContactWrite = map[string]struct{}{
	"LastName": {}, "FirstName": {}, "Email": {}, "Phone": {}, "Mobile": {}, "description": {},
}

func scanContactRow(scanner interface{ Scan(dest ...interface{}) error }) (models.Contact, error) {
	var m models.Contact
	var fn, em, ph, mob, desc sql.NullString
	err := scanner.Scan(&m.ID, &m.LastName, &fn, &em, &ph, &mob, &desc)
	if err != nil {
		return m, err
	}
	if fn.Valid {
		m.FirstName = &fn.String
	}
	if em.Valid {
		m.Email = &em.String
	}
	if ph.Valid {
		m.Phone = &ph.String
	}
	if mob.Valid {
		m.Mobile = &mob.String
	}
	if desc.Valid {
		m.Description = &desc.String
	}
	return m, nil
}

func contactJSONToCol(lk string) string {
	switch lk {
	case "last_name", "lastname":
		return "LastName"
	case "first_name", "firstname":
		return "FirstName"
	case "email":
		return "Email"
	case "phone":
		return "Phone"
	case "mobile":
		return "Mobile"
	case "description":
		return "description"
	}
	return ""
}

// ListContacts returns all contacts.
func (h *Handlers) ListContacts(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	query := "SELECT " + contactListCols + " FROM contact WHERE 1=1 ORDER BY LastName, FirstName LIMIT 500"

	rows, err := h.TranMySQL.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.Contact
	for rows.Next() {
		m, err := scanContactRow(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, m)
	}
	c.JSON(http.StatusOK, list)
}

// GetContact returns one contact with Mongo detail.
func (h *Handlers) GetContact(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+contactFullSelectCols+" FROM contact WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok || m == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
		return
	}
	h.attachEntityDetail(c, entityKeyContact, id, m)
	c.JSON(http.StatusOK, m)
}

// CreateContact inserts a contact. LastName is required.
func (h *Handlers) CreateContact(c *gin.Context) {
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

	lastName, _ := in["last_name"].(string)
	if lastName == "" {
		if v, ok := in["LastName"].(string); ok {
			lastName = v
		}
	}
	if strings.TrimSpace(lastName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "last_name is required"})
		return
	}

	cols := []string{"LastName"}
	args := []interface{}{lastName}

	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		col := contactJSONToCol(lk)
		if col == "" {
			continue
		}
		if _, ok := allowedContactWrite[col]; !ok {
			continue
		}
		if col == "LastName" {
			continue
		}
		cols = append(cols, col)
		args = append(args, v)
	}

	stmt := "INSERT INTO contact (" + strings.Join(cols, ",") + ") VALUES (" + strings.TrimRight(strings.Repeat("?,", len(cols)), ",") + ")"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyContact, int(id64), detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+contactFullSelectCols+" FROM contact WHERE ID = ?", id64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id64})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyContact, int(id64), m)
	}
	c.JSON(http.StatusOK, m)
}

// UpdateContact updates a contact by ID.
func (h *Handlers) UpdateContact(c *gin.Context) {
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

	var sets []string
	var args []interface{}
	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		col := contactJSONToCol(lk)
		if col == "" {
			continue
		}
		if _, ok := allowedContactWrite[col]; !ok {
			continue
		}
		sets = append(sets, col+" = ?")
		args = append(args, v)
	}
	if len(sets) == 0 {
		m, ok, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+contactFullSelectCols+" FROM contact WHERE ID = ?", id)
		if !ok || m == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
			return
		}
		if !hasDetail {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		_ = h.savePoppedDetail(c, entityKeyContact, id, detailStr)
		h.attachEntityDetail(c, entityKeyContact, id, m)
		c.JSON(http.StatusOK, m)
		return
	}
	args = append(args, id)
	stmt := "UPDATE contact SET " + strings.Join(sets, ", ") + " WHERE ID = ?"
	_, err = h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyContact, id, detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+contactFullSelectCols+" FROM contact WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"updated": id})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyContact, id, m)
	}
	c.JSON(http.StatusOK, m)
}

// DeleteContact deletes a contact by ID.
func (h *Handlers) DeleteContact(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec("DELETE FROM contact WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
		return
	}
	h.deleteEntityDetailMongo(context.Background(), entityKeyContact, id)
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
