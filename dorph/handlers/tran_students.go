package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"context"
	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

// MySQL table `member` (MEMBER is reserved in 8.0+).
const sqlMemberTable = "`member`"

const studentSelectCols = `id, last_name, first_name, middle_name, dob, entry_date, facility, gender, email, participant_type, description`

const studentSelectColsAliased = studentSelectCols

var allowedStudentWrite = map[string]struct{}{
	"last_name": {}, "first_name": {}, "middle_name": {},
	"dob": {}, "entry_date": {}, "facility": {}, "gender": {}, "email": {},
	"participant_type": {}, "description": {},
}

// canonicalStudentSQLCol maps legacy / alias names to the real MySQL column (never emit "Mi").
func canonicalStudentSQLCol(col string) string {
	switch strings.TrimSpace(col) {
	case "Mi", "MI", "mi":
		return "middle_name"
	default:
		return col
	}
}

func scanStudentRow(scanner interface{ Scan(dest ...interface{}) error }) (models.Student, error) {
	var s models.Student
	var ln, fn, mn sql.NullString
	var dob, entry sql.NullTime
	var gender sql.NullInt64
	var facility, email, ptype, desc sql.NullString
	err := scanner.Scan(
		&s.ID, &ln, &fn, &mn,
		scanDestTime{&dob}, scanDestTime{&entry}, &facility, &gender, &email, &ptype, &desc,
	)
	if err != nil {
		return s, err
	}
	if ln.Valid {
		s.LastName = &ln.String
	}
	if fn.Valid {
		s.FirstName = &fn.String
	}
	if mn.Valid {
		s.MiddleName = &mn.String
	}
	if dob.Valid {
		t := dob.Time
		s.Dob = &t
	}
	if entry.Valid {
		t := entry.Time
		s.EntryDate = &t
	}
	if facility.Valid {
		s.Facility = &facility.String
	}
	if gender.Valid {
		v := int(gender.Int64)
		s.Gender = &v
	}
	if email.Valid {
		s.Email = &email.String
	}
	if ptype.Valid {
		s.ParticipantType = &ptype.String
	}
	if desc.Valid {
		s.Description = &desc.String
	}
	return s, nil
}

// ListStudents returns students (optional school filter).
func (h *Handlers) ListStudents(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	query := "SELECT " + studentSelectCols + " FROM " + sqlMemberTable + " WHERE 1=1"
	var args []interface{}
	facility := c.Query("facility")
	if facility == "" {
		facility = c.Query("school")
	}
	if facility != "" {
		query += " AND facility = ?"
		args = append(args, facility)
	}
	query += " ORDER BY last_name, first_name LIMIT 500"

	rows, err := h.TranMySQL.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []models.Student
	for rows.Next() {
		s, err := scanStudentRow(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, s)
	}
	c.JSON(http.StatusOK, list)
}

// GetStudent returns one student by ID.
func (h *Handlers) GetStudent(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row := h.TranMySQL.DB.QueryRow("SELECT "+studentSelectCols+" FROM "+sqlMemberTable+" WHERE id = ?", id)
	s, err := scanStudentRow(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// GetStudentFull returns student fields as a map (snake_case keys).
func (h *Handlers) GetStudentFull(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+studentSelectColsAliased+" FROM "+sqlMemberTable+" WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}
	h.attachEntityDetail(c, entityKeyStudent, id, m)
	h.attachEntityAttachmentsToRow(c.Request.Context(), "members", entityAttachmentMember, id, m)
	c.JSON(http.StatusOK, m)
}

// CreateStudent inserts a student.
func (h *Handlers) CreateStudent(c *gin.Context) {
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
		col := studentJSONToCol(lk, k)
		if col == "" {
			if _, ok := allowedStudentWrite[k]; ok {
				col = k
			} else {
				continue
			}
		}
		col = canonicalStudentSQLCol(col)
		if _, ok := allowedStudentWrite[col]; !ok {
			continue
		}

		val := v
		switch col {
		case "participant_type", "description":
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					val = nil
				} else {
					val = t
				}
			}
		case "gender":
			n := intFromAny(in, "gender", "Gender")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "dob":
			val = normalizeStudentDateVal(v)
		case "entry_date":
			val = normalizeStudentDateTimeVal(v)
		}

		cols = append(cols, col)
		args = append(args, val)
	}

	if len(cols) == 0 {
		if !hasDetail {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to insert"})
			return
		}
		res, err := h.TranMySQL.DB.Exec("INSERT INTO "+sqlMemberTable+" (last_name) VALUES (?)", "New")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id64, _ := res.LastInsertId()
		_ = h.savePoppedDetail(c, entityKeyStudent, int(id64), detailStr)
		m, _, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+studentSelectColsAliased+" FROM "+sqlMemberTable+" WHERE id = ?", id64)
		if m == nil {
			c.JSON(http.StatusOK, gin.H{"id": id64})
			return
		}
		h.attachEntityDetail(c, entityKeyStudent, int(id64), m)
		c.JSON(http.StatusOK, m)
		return
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(cols)), ",")
	stmt := "INSERT INTO " + sqlMemberTable + " (" + strings.Join(cols, ",") + ") VALUES (" + placeholders + ")"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyStudent, int(id64), detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT "+studentSelectColsAliased+" FROM "+sqlMemberTable+" WHERE id = ?", id64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id64})
		return
	}
	if m != nil {
		h.attachEntityDetail(c, entityKeyStudent, int(id64), m)
	}
	c.JSON(http.StatusOK, m)
}

func normalizeStudentDateVal(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		if len(s) >= 10 {
			s = s[:10]
		}
		if d, err := time.Parse("2006-01-02", s); err == nil {
			return d.Format("2006-01-02")
		}
		return s
	default:
		return v
	}
}

func normalizeStudentDateTimeVal(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		if len(s) == 16 {
			if dt, err := time.Parse("2006-01-02T15:04", s); err == nil {
				return dt
			}
		}
		if len(s) >= 10 {
			if d, err := time.Parse("2006-01-02", s[:10]); err == nil {
				return d
			}
		}
		return s
	default:
		return v
	}
}

// UpdateStudent updates a student by ID.
func (h *Handlers) UpdateStudent(c *gin.Context) {
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
	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "id" {
			continue
		}
		col := studentJSONToCol(lk, k)
		if col == "" {
			if _, ok := allowedStudentWrite[k]; ok {
				col = k
			} else {
				continue
			}
		}
		col = canonicalStudentSQLCol(col)
		if _, ok := allowedStudentWrite[col]; !ok {
			continue
		}

		val := v
		switch col {
		case "participant_type", "description":
			if s, ok := v.(string); ok {
				t := strings.TrimSpace(s)
				if t == "" {
					val = nil
				} else {
					val = t
				}
			}
		case "gender":
			n := intFromAny(in, "gender", "Gender")
			if n <= 0 {
				val = nil
			} else {
				val = n
			}
		case "dob":
			val = normalizeStudentDateVal(v)
		case "entry_date":
			val = normalizeStudentDateTimeVal(v)
		}

		setCols = append(setCols, col+" = ?")
		args = append(args, val)
	}
	if len(setCols) == 0 {
		m, ok, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+studentSelectColsAliased+" FROM "+sqlMemberTable+" WHERE id = ?", id)
		if !ok || m == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
			return
		}
		if !hasDetail {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		_ = h.savePoppedDetail(c, entityKeyStudent, id, detailStr)
		h.attachEntityDetail(c, entityKeyStudent, id, m)
		c.JSON(http.StatusOK, m)
		return
	}
	args = append(args, id)
	stmt := "UPDATE " + sqlMemberTable + " SET " + strings.Join(setCols, ", ") + " WHERE id = ?"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyStudent, id, detailStr)
	}
	m, _, _ := querySingleRowMap(h.TranMySQL.DB, "SELECT "+studentSelectColsAliased+" FROM "+sqlMemberTable+" WHERE id = ?", id)
	if m != nil {
		h.attachEntityDetail(c, entityKeyStudent, id, m)
		c.JSON(http.StatusOK, m)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func studentJSONToCol(lk, k string) string {
	switch lk {
	case "last_name":
		return "last_name"
	case "first_name":
		return "first_name"
	case "middle_name", "mi":
		return "middle_name"
	case "dob":
		return "dob"
	case "entry_date":
		return "entry_date"
	case "school", "facility":
		return "facility"
	case "gender":
		return "gender"
	case "email":
		return "email"
	case "participant_type":
		return "participant_type"
	case "description":
		return "description"
	}
	return ""
}

// DeleteStudent deletes a student by ID (also removes schedules).
func (h *Handlers) DeleteStudent(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := h.TranMySQL.DB.Exec("DELETE FROM "+sqlMemberTable+" WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}
	h.deleteEntityDetailMongo(context.Background(), entityKeyStudent, id)
	h.purgeEntityAttachments(context.Background(), entityAttachmentMember, id)
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
