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

const tripListCols = `ID, Name, start_date, end_date, location, ActivityType, description`

const tripFullSelectCols = `ID AS id, Name AS name, start_date AS start_date, end_date AS end_date, location AS location, ActivityType AS activity_type, GUID AS guid, description AS description`

var allowedTripWrite = map[string]struct{}{
	"Name": {}, "ActivityType": {}, "start_date": {}, "end_date": {}, "location": {}, "GUID": {}, "description": {},
}

func tripJSONToCol(lk, origKey string) string {
	switch lk {
	case "name":
		return "Name"
	case "activity_type":
		return "ActivityType"
	case "start_date":
		return "start_date"
	case "end_date":
		return "end_date"
	case "location":
		return "location"
	case "guid":
		return "GUID"
	case "description":
		return "description"
	}
	if origKey != "" {
		if _, ok := allowedTripWrite[origKey]; ok {
			return origKey
		}
	}
	return ""
}

func scanTripListRow(scanner interface {
	Scan(dest ...interface{}) error
}) (models.Trip, error) {
	var t models.Trip
	var startDate, endDate sql.NullTime
	var location, atype, desc sql.NullString
	err := scanner.Scan(&t.ID, &t.Name, scanDestTime{&startDate}, scanDestTime{&endDate}, &location, &atype, &desc)
	if err != nil {
		return t, err
	}
	if startDate.Valid {
		ts := startDate.Time
		t.StartDate = &ts
	}
	if endDate.Valid {
		ts := endDate.Time
		t.EndDate = &ts
	}
	if location.Valid {
		s := location.String
		t.Location = &s
	}
	if atype.Valid {
		s := atype.String
		t.ActivityType = &s
	}
	if desc.Valid {
		s := desc.String
		t.Description = &s
	}
	return t, nil
}

// ListTrips returns trips.
func (h *Handlers) ListTrips(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	query := "SELECT " + tripListCols + " FROM Activity WHERE 1=1 ORDER BY Name LIMIT 500"

	rows, err := h.TranMySQL.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.Trip
	for rows.Next() {
		t, err := scanTripListRow(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, t)
	}
	c.JSON(http.StatusOK, list)
}

// GetTrip returns one trip by ID.
func (h *Handlers) GetTrip(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row := h.TranMySQL.DB.QueryRow("SELECT "+tripListCols+" FROM Activity WHERE ID = ?", id)
	t, err := scanTripListRow(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// GetTripFull returns Trip fields used by the admin UI as a map.
func (h *Handlers) GetTripFull(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+tripFullSelectCols+" FROM Activity WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}
	h.attachEntityDetail(c, entityKeyTrip, id, m)
	h.attachEntityAttachmentsToRow(c.Request.Context(), "activities", entityAttachmentActivity, id, m)
	c.JSON(http.StatusOK, m)
}

// CreateTrip inserts a trip.
func (h *Handlers) CreateTrip(c *gin.Context) {
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

	name, _ := in["name"].(string)
	if name == "" {
		name, _ = in["Name"].(string)
	}
	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	cols := []string{}
	args := []interface{}{}

	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "id" || lk == "db_id" {
			continue
		}
		col := tripJSONToCol(lk, k)
		if col == "" {
			if _, ok := allowedTripWrite[k]; ok {
				col = k
			} else {
				continue
			}
		}
		if _, ok := allowedTripWrite[col]; !ok {
			continue
		}
		if col == "Name" {
			continue
		}

		val := v
		switch col {
		case "ActivityType", "description":
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					val = nil
				} else {
					val = t
				}
			}
		case "start_date", "end_date":
			val = normalizeTranDateTimeColVal(v)
		case "location":
			norm, e := normalizeLocationValue(v)
			if e != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location json"})
				return
			}
			val = norm
		}

		cols = append(cols, col)
		args = append(args, val)
	}

	cols = append(cols, "Name")
	args = append(args, name)

	placeholders := strings.TrimRight(strings.Repeat("?,", len(cols)), ",")
	stmt := "INSERT INTO Activity (" + strings.Join(cols, ",") + ") VALUES (" + placeholders + ")"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyTrip, int(id64), detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+tripFullSelectCols+" FROM Activity WHERE ID = ?", id64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id64})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyTrip, int(id64), m)
	}
	c.JSON(http.StatusOK, m)
}

// DeleteTrip deletes a trip by ID.
func (h *Handlers) DeleteTrip(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec("DELETE FROM Activity WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}
	h.deleteEntityDetailMongo(context.Background(), entityKeyTrip, id)
	h.purgeEntityAttachments(context.Background(), entityAttachmentActivity, id)
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
