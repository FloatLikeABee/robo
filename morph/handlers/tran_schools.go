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

// MySQL table `facility` (replaces School).
const sqlFacilityTable = "`facility`"

const schoolFullSelectCols = `id, facility_code, name, district_id, facility_type, description, location`

const schoolScanCols = schoolFullSelectCols

// ListFacilities returns all facilities (MySQL `facility` table), optional district_id query.
func (h *Handlers) ListFacilities(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	query := "SELECT " + schoolScanCols + " FROM " + sqlFacilityTable + " WHERE 1=1"
	var args []interface{}
	if districtID := c.Query("district_id"); districtID != "" {
		query += " AND district_id = ?"
		args = append(args, districtID)
	}
	query += " ORDER BY name, facility_code"

	rows, err := h.TranMySQL.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.School
	for rows.Next() {
		s, err := scanSchoolRow(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, s)
	}
	c.JSON(http.StatusOK, list)
}

func scanSchoolRow(rows *sql.Rows) (models.School, error) {
	var s models.School
	var name sql.NullString
	var districtID sql.NullInt64
	var ftype, desc, loc sql.NullString
	err := rows.Scan(&s.ID, &s.FacilityCode, &name, &districtID, &ftype, &desc, &loc)
	if err != nil {
		return s, err
	}
	if name.Valid {
		s.Name = &name.String
	}
	if districtID.Valid {
		v := int(districtID.Int64)
		s.DistrictID = &v
	}
	if ftype.Valid {
		s.FacilityType = &ftype.String
	}
	if desc.Valid {
		s.Description = &desc.String
	}
	if loc.Valid {
		s.Location = &loc.String
	}
	return s, nil
}

// GetFacility returns one facility (map with detail) by ID.
func (h *Handlers) GetFacility(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+schoolFullSelectCols+" FROM "+sqlFacilityTable+" WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok || m == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "facility not found"})
		return
	}
	h.attachEntityDetail(c, entityKeySchool, id, m)
	h.attachEntityAttachmentsToRow(c.Request.Context(), "facilities", entityAttachmentFacility, id, m)
	c.JSON(http.StatusOK, m)
}

// schoolJSONToCol maps request keys to MySQL column names (snake_case).
func schoolJSONToCol(lk, _ string) string {
	switch lk {
	case "facility_code", "school_code", "facilitycode", "schoolcode":
		return "facility_code"
	case "name":
		return "name"
	case "district_id":
		return "district_id"
	case "facility_type":
		return "facility_type"
	case "description":
		return "description"
	case "location":
		return "location"
	default:
		return ""
	}
}

// CreateFacility inserts a facility.
func (h *Handlers) CreateFacility(c *gin.Context) {
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

	allowed := map[string]struct{}{
		"facility_code": {},
		"name":          {},
		"district_id":   {},
		"facility_type": {},
		"description":   {},
		"location":      {},
	}

	cols := []string{}
	args := []interface{}{}

	for k, v := range in {
		if k == "" {
			continue
		}
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		col := schoolJSONToCol(lk, k)
		if col == "" {
			if _, ok := allowed[k]; ok {
				col = k
			} else {
				continue
			}
		}
		if _, ok := allowed[col]; !ok {
			continue
		}

		val := v
		switch col {
		case "district_id":
			n := intFromAny(in, "district_id", "DistrictID")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "facility_type", "description":
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					val = nil
				} else {
					val = t
				}
			}
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

	if len(cols) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to insert"})
		return
	}

	hasFacilityCode := false
	for i, col := range cols {
		if col != "facility_code" {
			continue
		}
		hasFacilityCode = true
		if s, ok := args[i].(string); !ok || strings.TrimSpace(s) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "facility_code is required"})
			return
		}
		break
	}
	if !hasFacilityCode {
		c.JSON(http.StatusBadRequest, gin.H{"error": "facility_code is required"})
		return
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(cols)), ",")
	stmt := "INSERT INTO " + sqlFacilityTable + " (" + strings.Join(cols, ",") + ") VALUES (" + placeholders + ")"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeySchool, int(id64), detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+schoolFullSelectCols+" FROM "+sqlFacilityTable+" WHERE id = ?", id64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id64})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeySchool, int(id64), m)
	}
	c.JSON(http.StatusOK, m)
}

// UpdateFacility updates an existing facility.
func (h *Handlers) UpdateFacility(c *gin.Context) {
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
		"facility_code": {},
		"name":          {},
		"district_id":   {},
		"facility_type": {},
		"description":   {},
		"location":      {},
	}

	var sets []string
	var args []interface{}

	for k, v := range in {
		if k == "" {
			continue
		}
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		col := schoolJSONToCol(lk, k)
		if col == "" {
			if _, ok := allowed[k]; ok {
				col = k
			} else {
				continue
			}
		}
		if _, ok := allowed[col]; !ok {
			continue
		}

		val := v
		switch col {
		case "district_id":
			n := intFromAny(in, "district_id", "DistrictID")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "facility_type", "description":
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					val = nil
				} else {
					val = t
				}
			}
		case "location":
			norm, e := normalizeLocationValue(v)
			if e != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location json"})
				return
			}
			val = norm
		}

		sets = append(sets, col+" = ?")
		args = append(args, val)
	}

	if len(sets) == 0 {
		m, ok, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+schoolFullSelectCols+" FROM "+sqlFacilityTable+" WHERE id = ?", id)
		if !ok || m == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "facility not found"})
			return
		}
		if !hasDetail {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		_ = h.savePoppedDetail(c, entityKeySchool, id, detailStr)
		h.attachEntityDetail(c, entityKeySchool, id, m)
		c.JSON(http.StatusOK, m)
		return
	}

	args = append(args, id)

	stmt := "UPDATE " + sqlFacilityTable + " SET " + strings.Join(sets, ", ") + " WHERE id = ?"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "facility not found"})
		return
	}
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeySchool, id, detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+schoolFullSelectCols+" FROM "+sqlFacilityTable+" WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeySchool, id, m)
	}
	c.JSON(http.StatusOK, m)
}

// DeleteFacility deletes a facility by ID.
func (h *Handlers) DeleteFacility(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := h.TranMySQL.DB.Exec("DELETE FROM "+sqlFacilityTable+" WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "facility not found"})
		return
	}
	h.deleteEntityDetailMongo(context.Background(), entityKeySchool, id)
	h.purgeEntityAttachments(context.Background(), entityAttachmentFacility, id)
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
