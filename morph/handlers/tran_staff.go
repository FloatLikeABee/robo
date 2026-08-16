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

const sqlEmployeeTable = "`employee`"

const staffListSelectCols = `e.id, e.last_name, e.first_name, e.middle_name, e.email, e.phone_number, e.active_flag, e.employ_type, e.description, e.facility_id, f.facility_code, f.name`

const staffFromJoin = "`employee` e LEFT JOIN `facility` f ON f.id = e.facility_id"

var staffDetailCandidateCols = []string{
	"id", "last_name", "first_name", "middle_name",
	"staff_guid", "active_flag", "inactive_date", "contractor_id",
	"email", "phone_number", "date_of_birth", "gender", "user_id", "employ_type", "description", "facility_id",
}

var allowedStaffWrite = map[string]struct{}{
	"last_name": {}, "first_name": {}, "middle_name": {}, "staff_guid": {}, "active_flag": {}, "inactive_date": {}, "contractor_id": {},
	"email": {}, "phone_number": {}, "date_of_birth": {}, "gender": {}, "user_id": {},
	"employ_type": {}, "description": {}, "facility_id": {},
}

func staffJSONToCol(lk, origKey string) string {
	switch lk {
	case "last_name", "lastname":
		return "last_name"
	case "first_name":
		return "first_name"
	case "middle_name":
		return "middle_name"
	case "staff_guid":
		return "staff_guid"
	case "active_flag":
		return "active_flag"
	case "inactive_date":
		return "inactive_date"
	case "contractor_id":
		return "contractor_id"
	case "email":
		return "email"
	case "phone_number", "cell_phone":
		return "phone_number"
	case "date_of_birth":
		return "date_of_birth"
	case "gender":
		return "gender"
	case "user_id":
		return "user_id"
	case "employ_type":
		return "employ_type"
	case "description":
		return "description"
	case "facility_id":
		return "facility_id"
	}
	if origKey != "" {
		if _, ok := allowedStaffWrite[origKey]; ok {
			return origKey
		}
	}
	return ""
}

func scanStaffRow(scanner interface {
	Scan(dest ...interface{}) error
}) (models.Staff, error) {
	var s models.Staff
	var fn, mn, em, pn, et, desc, facCode, facName sql.NullString
	var af bool
	var facID sql.NullInt64
	err := scanner.Scan(&s.ID, &s.LastName, &fn, &mn, &em, &pn, &af, &et, &desc, &facID, &facCode, &facName)
	if err != nil {
		return s, err
	}
	if fn.Valid {
		s.FirstName = &fn.String
	}
	if mn.Valid {
		s.MiddleName = &mn.String
	}
	if em.Valid {
		s.Email = &em.String
	}
	if pn.Valid {
		s.PhoneNumber = &pn.String
	}
	s.ActiveFlag = af
	if et.Valid {
		s.EmployType = &et.String
	}
	if desc.Valid {
		s.Description = &desc.String
	}
	if facID.Valid && facID.Int64 > 0 {
		v := int(facID.Int64)
		s.FacilityID = &v
	}
	parts := []string{}
	if facCode.Valid && strings.TrimSpace(facCode.String) != "" {
		parts = append(parts, strings.TrimSpace(facCode.String))
	}
	if facName.Valid && strings.TrimSpace(facName.String) != "" {
		parts = append(parts, strings.TrimSpace(facName.String))
	}
	if len(parts) > 0 {
		d := strings.Join(parts, " — ")
		s.FacilityDisplay = &d
	}
	return s, nil
}

func employeeColumnSet(db *sql.DB) map[string]struct{} {
	out := map[string]struct{}{}
	if db == nil {
		return out
	}
	rows, err := db.Query(`SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'employee'`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err == nil {
			out[strings.ToLower(strings.TrimSpace(c))] = struct{}{}
		}
	}
	return out
}

func employeeHasColumn(cols map[string]struct{}, col string) bool {
	if len(cols) == 0 {
		return false
	}
	_, ok := cols[strings.ToLower(strings.TrimSpace(col))]
	return ok
}

func staffDetailSelectCols(db *sql.DB) string {
	cols := employeeColumnSet(db)
	if len(cols) == 0 {
		// INFORMATION_SCHEMA may be denied or flaky; slim `employee` schemas differ — SELECT * avoids unknown-column SQL errors on /full.
		return "*"
	}
	var list []string
	for _, c := range staffDetailCandidateCols {
		if employeeHasColumn(cols, c) {
			list = append(list, c)
		}
	}
	if employeeHasColumn(cols, "detail") {
		list = append(list, "detail")
	}
	if len(list) == 0 {
		return "*"
	}
	return strings.Join(list, ", ")
}

// ListStaff returns staff.
func (h *Handlers) ListStaff(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	query := "SELECT " + staffListSelectCols + " FROM " + staffFromJoin + " WHERE 1=1 ORDER BY e.last_name, e.first_name LIMIT 500"

	rows, err := h.TranMySQL.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.Staff
	for rows.Next() {
		s, err := scanStaffRow(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, s)
	}
	c.JSON(http.StatusOK, list)
}

// GetStaff returns one staff by ID.
func (h *Handlers) GetStaff(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row := h.TranMySQL.DB.QueryRow("SELECT "+staffListSelectCols+" FROM "+staffFromJoin+" WHERE e.id = ?", id)
	s, err := scanStaffRow(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "staff not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// GetStaffFull returns all employee columns as a map (snake_case keys).
func (h *Handlers) GetStaffFull(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	fullCols := staffDetailSelectCols(h.TranMySQL.DB)
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+fullCols+" FROM "+sqlEmployeeTable+" WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "staff not found"})
		return
	}
	h.attachEntityDetail(c, entityKeyStaff, id, m)
	h.attachEntityAttachmentsToRow(c.Request.Context(), "employees", entityAttachmentEmployee, id, m)
	c.JSON(http.StatusOK, m)
}

// CreateStaff inserts a staff record.
func (h *Handlers) CreateStaff(c *gin.Context) {
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
		lastName, _ = in["LastName"].(string)
	}
	if strings.TrimSpace(lastName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "last_name is required"})
		return
	}

	cols := []string{}
	args := []interface{}{}
	existingCols := employeeColumnSet(h.TranMySQL.DB)

	for k, v := range in {
		if k == "" {
			continue
		}
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		col := staffJSONToCol(lk, k)
		if col == "" {
			if _, ok := allowedStaffWrite[k]; ok {
				col = k
			} else {
				continue
			}
		}
		if _, ok := allowedStaffWrite[col]; !ok {
			continue
		}
		if !employeeHasColumn(existingCols, col) {
			continue
		}
		if col == "last_name" {
			continue
		}

		val := v
		switch col {
		case "employ_type", "description":
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					val = nil
				} else {
					val = t
				}
			}
		case "contractor_id":
			n := intFromAny(in, "contractor_id", "ContractorID")
			val = n
		case "user_id":
			n := intFromAny(in, "user_id", "UserID")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "gender":
			n := intFromAny(in, "gender", "Gender")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "facility_id":
			n := intFromAny(in, "facility_id", "FacilityID")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "date_of_birth":
			val = normalizeStudentDateVal(v)
		case "inactive_date":
			val = normalizeStudentDateVal(v)
		case "active_flag":
			val = boolToTinyint(boolFromAny(v))
		}

		cols = append(cols, col)
		args = append(args, val)
	}

	cols = append(cols, "last_name")
	args = append(args, lastName)

	placeholders := strings.TrimRight(strings.Repeat("?,", len(cols)), ",")
	stmt := "INSERT INTO " + sqlEmployeeTable + " (" + strings.Join(cols, ",") + ") VALUES (" + placeholders + ")"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyStaff, int(id64), detailStr)
	}
	fullCols := staffDetailSelectCols(h.TranMySQL.DB)
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+fullCols+" FROM "+sqlEmployeeTable+" WHERE id = ?", id64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id64})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyStaff, int(id64), m)
	}
	c.JSON(http.StatusOK, m)
}

// UpdateStaff updates a staff record by ID.
func (h *Handlers) UpdateStaff(c *gin.Context) {
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

	var setCols []string
	var args []interface{}
	existingCols := employeeColumnSet(h.TranMySQL.DB)
	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		col := staffJSONToCol(lk, k)
		if col == "" {
			if _, ok := allowedStaffWrite[k]; ok {
				col = k
			} else {
				continue
			}
		}
		if _, ok := allowedStaffWrite[col]; !ok {
			continue
		}
		if !employeeHasColumn(existingCols, col) {
			continue
		}

		val := v
		switch col {
		case "employ_type", "description":
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					val = nil
				} else {
					val = t
				}
			}
		case "contractor_id":
			n := intFromAny(in, "contractor_id", "ContractorID")
			val = n
		case "user_id":
			n := intFromAny(in, "user_id", "UserID")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "gender":
			n := intFromAny(in, "gender", "Gender")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "facility_id":
			n := intFromAny(in, "facility_id", "FacilityID")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "date_of_birth":
			val = normalizeStudentDateVal(v)
		case "inactive_date":
			val = normalizeStudentDateVal(v)
		case "active_flag":
			val = boolToTinyint(boolFromAny(v))
		}

		setCols = append(setCols, col+" = ?")
		args = append(args, val)
	}
	if len(setCols) == 0 {
		fullCols := staffDetailSelectCols(h.TranMySQL.DB)
		m, ok, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+fullCols+" FROM "+sqlEmployeeTable+" WHERE id = ?", id)
		if !ok || m == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "staff not found"})
			return
		}
		if !hasDetail {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		_ = h.savePoppedDetail(c, entityKeyStaff, id, detailStr)
		h.attachEntityDetail(c, entityKeyStaff, id, m)
		c.JSON(http.StatusOK, m)
		return
	}
	args = append(args, id)
	stmt := "UPDATE " + sqlEmployeeTable + " SET " + strings.Join(setCols, ", ") + " WHERE id = ?"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		// MySQL reports 0 affected rows when values are unchanged. Verify existence
		// before treating this as not found so detail-only updates still persist.
		var exists int
		if err := h.TranMySQL.DB.QueryRow("SELECT 1 FROM "+sqlEmployeeTable+" WHERE id = ? LIMIT 1", id).Scan(&exists); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "staff not found"})
			return
		}
	}
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyStaff, id, detailStr)
	}
	fullCols := staffDetailSelectCols(h.TranMySQL.DB)
	m, _, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+fullCols+" FROM "+sqlEmployeeTable+" WHERE id = ?", id)
	if m != nil {
		h.attachEntityDetail(c, entityKeyStaff, id, m)
		c.JSON(http.StatusOK, m)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// DeleteStaff deletes staff by ID.
func (h *Handlers) DeleteStaff(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec("DELETE FROM "+sqlEmployeeTable+" WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "staff not found"})
		return
	}
	h.deleteEntityDetailMongo(context.Background(), entityKeyStaff, id)
	h.purgeEntityAttachments(context.Background(), entityAttachmentEmployee, id)
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
