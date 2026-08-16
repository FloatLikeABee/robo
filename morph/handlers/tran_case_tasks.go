package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phpdave11/gofpdf"
)

const (
	caseTaskListCols = `ID, title, description, start_at, end_at, location, assignee_type, assignee_id, created_on, last_updated`
	caseTaskEntity   = "case_task"
)

var allowedCaseTaskWrite = map[string]struct{}{
	"title": {}, "description": {}, "start_at": {}, "end_at": {}, "location": {},
}

type caseTaskAssigneeOut struct {
	AssigneeKind string `json:"assignee_kind"`
	AssigneeID   int    `json:"assignee_id"`
	Name         string `json:"name"`
	Email        string `json:"email,omitempty"`
}

type caseTaskRow struct {
	ID             int                   `json:"id"`
	Title          string                `json:"title"`
	Description    *string               `json:"description,omitempty"`
	StartAt        *time.Time            `json:"start_at,omitempty"`
	EndAt          *time.Time            `json:"end_at,omitempty"`
	Location       *string               `json:"location,omitempty"`
	AssigneeType   string                `json:"assignee_type,omitempty"`
	AssigneeID     int                   `json:"assignee_id,omitempty"`
	AssigneeName   string                `json:"assignee_name,omitempty"`
	Assignees      []caseTaskAssigneeOut `json:"assignees"`
	AssigneesLabel string                `json:"assignees_label"`
	CreatedOn      *time.Time            `json:"created_on,omitempty"`
	LastUpdated    *time.Time            `json:"last_updated,omitempty"`
}

func parseCaseTaskRow(scanner interface {
	Scan(dest ...interface{}) error
}) (caseTaskRow, error) {
	var r caseTaskRow
	var desc, loc sql.NullString
	var startAt, endAt, createdOn, updatedOn sql.NullTime
	var legacyType sql.NullString
	var legacyID sql.NullInt64
	if err := scanner.Scan(
		&r.ID,
		&r.Title,
		&desc,
		scanDestTime{&startAt},
		scanDestTime{&endAt},
		&loc,
		&legacyType,
		&legacyID,
		scanDestTime{&createdOn},
		scanDestTime{&updatedOn},
	); err != nil {
		return r, err
	}
	if legacyType.Valid {
		r.AssigneeType = legacyType.String
	}
	if legacyID.Valid {
		r.AssigneeID = int(legacyID.Int64)
	}
	if desc.Valid {
		s := desc.String
		r.Description = &s
	}
	if startAt.Valid {
		ts := startAt.Time
		r.StartAt = &ts
	}
	if endAt.Valid {
		ts := endAt.Time
		r.EndAt = &ts
	}
	if loc.Valid {
		s := loc.String
		r.Location = &s
	}
	if createdOn.Valid {
		ts := createdOn.Time
		r.CreatedOn = &ts
	}
	if updatedOn.Valid {
		ts := updatedOn.Time
		r.LastUpdated = &ts
	}
	return r, nil
}

func (h *Handlers) caseTaskAssigneeName(ctx context.Context, assigneeKind string, assigneeID int) string {
	if h == nil || h.TranMySQL == nil || assigneeID <= 0 {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(assigneeKind)) {
	case "member":
		var first, last sql.NullString
		err := h.TranMySQL.DB.QueryRowContext(ctx, "SELECT first_name, last_name FROM `member` WHERE id = ?", assigneeID).Scan(&first, &last)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(strings.TrimSpace(first.String) + " " + strings.TrimSpace(last.String))
	case "employee":
		var first, last sql.NullString
		err := h.TranMySQL.DB.QueryRowContext(ctx, "SELECT first_name, last_name FROM `employee` WHERE id = ?", assigneeID).Scan(&first, &last)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(strings.TrimSpace(first.String) + " " + strings.TrimSpace(last.String))
	case "contact":
		var first, last sql.NullString
		err := h.TranMySQL.DB.QueryRowContext(ctx, "SELECT FirstName, LastName FROM contact WHERE ID = ?", assigneeID).Scan(&first, &last)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(strings.TrimSpace(first.String) + " " + strings.TrimSpace(last.String))
	default:
		return ""
	}
}

func (h *Handlers) caseTaskAssigneeEmail(ctx context.Context, assigneeKind string, assigneeID int) string {
	if h == nil || h.TranMySQL == nil || assigneeID <= 0 {
		return ""
	}
	var email sql.NullString
	switch strings.ToLower(strings.TrimSpace(assigneeKind)) {
	case "member":
		_ = h.TranMySQL.DB.QueryRowContext(ctx, "SELECT email FROM `member` WHERE id = ?", assigneeID).Scan(&email)
	case "employee":
		_ = h.TranMySQL.DB.QueryRowContext(ctx, "SELECT email FROM `employee` WHERE id = ?", assigneeID).Scan(&email)
	case "contact":
		_ = h.TranMySQL.DB.QueryRowContext(ctx, "SELECT Email FROM contact WHERE ID = ?", assigneeID).Scan(&email)
	}
	if !email.Valid {
		return ""
	}
	return strings.TrimSpace(email.String)
}

func normalizeAssigneeKind(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (h *Handlers) validateCaseTaskAssigneeRef(ctx context.Context, kind string, id int) error {
	k := normalizeAssigneeKind(kind)
	if k != "member" && k != "employee" && k != "contact" {
		return errors.New("assignee_kind must be member, employee, or contact")
	}
	if id <= 0 {
		return errors.New("assignee_id must be positive")
	}
	if h == nil || h.TranMySQL == nil {
		return errors.New("Tran SQL store not configured")
	}
	var exists int
	switch k {
	case "member":
		if err := h.TranMySQL.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM `member` WHERE id = ?", id).Scan(&exists); err != nil {
			return err
		}
	case "employee":
		if err := h.TranMySQL.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM `employee` WHERE id = ?", id).Scan(&exists); err != nil {
			return err
		}
	case "contact":
		if err := h.TranMySQL.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM contact WHERE ID = ?", id).Scan(&exists); err != nil {
			return err
		}
	}
	if exists == 0 {
		return fmt.Errorf("%s with id %d not found", k, id)
	}
	return nil
}

type caseTaskAssigneeKey struct {
	Kind string
	ID   int
}

func dedupeAssignees(list []caseTaskAssigneeKey) []caseTaskAssigneeKey {
	seen := map[string]struct{}{}
	out := make([]caseTaskAssigneeKey, 0, len(list))
	for _, a := range list {
		k := normalizeAssigneeKind(a.Kind)
		if k == "" || a.ID <= 0 {
			continue
		}
		sig := k + ":" + strconv.Itoa(a.ID)
		if _, ok := seen[sig]; ok {
			continue
		}
		seen[sig] = struct{}{}
		out = append(out, caseTaskAssigneeKey{Kind: k, ID: a.ID})
	}
	return out
}

func parseAssigneesFromInput(in map[string]interface{}) ([]caseTaskAssigneeKey, bool, error) {
	if in == nil {
		return nil, false, nil
	}
	raw, ok := in["assignees"]
	if !ok || raw == nil {
		return nil, false, nil
	}
	delete(in, "assignees")
	arr, ok := raw.([]interface{})
	if !ok {
		return nil, true, errors.New("assignees must be an array")
	}
	out := make([]caseTaskAssigneeKey, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, true, errors.New("each assignee must be an object")
		}
		kindStr := ""
		if v, ok := m["assignee_kind"].(string); ok && v != "" {
			kindStr = v
		} else if v, ok := m["kind"].(string); ok {
			kindStr = v
		}
		idVal := 0
		switch v := m["assignee_id"].(type) {
		case float64:
			idVal = int(v)
		case int:
			idVal = v
		case string:
			idVal, _ = strconv.Atoi(strings.TrimSpace(v))
		default:
			if v2, ok := m["id"].(float64); ok {
				idVal = int(v2)
			} else if v3, ok := m["id"].(int); ok {
				idVal = v3
			} else if v4, ok := m["id"].(string); ok {
				idVal, _ = strconv.Atoi(strings.TrimSpace(v4))
			}
		}
		k := normalizeAssigneeKind(kindStr)
		if k == "" || idVal <= 0 {
			continue
		}
		out = append(out, caseTaskAssigneeKey{Kind: k, ID: idVal})
	}
	return dedupeAssignees(out), true, nil
}

func legacyAssigneesFromInput(in map[string]interface{}) ([]caseTaskAssigneeKey, error) {
	if in == nil {
		return nil, nil
	}
	typ, _ := in["assignee_type"].(string)
	idVal := 0
	switch v := in["assignee_id"].(type) {
	case float64:
		idVal = int(v)
	case int:
		idVal = v
	case string:
		idVal, _ = strconv.Atoi(strings.TrimSpace(v))
	}
	if normalizeAssigneeKind(typ) == "" || idVal <= 0 {
		return nil, nil
	}
	return []caseTaskAssigneeKey{{Kind: normalizeAssigneeKind(typ), ID: idVal}}, nil
}

func (h *Handlers) assigneesKeysForCaseTask(ctx context.Context, caseTaskID int) ([]caseTaskAssigneeKey, error) {
	rows, err := h.TranMySQL.DB.QueryContext(ctx, `
		SELECT assignee_kind, assignee_id FROM CaseTaskAssignee
		WHERE case_task_id = ? ORDER BY id`, caseTaskID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	list := make([]caseTaskAssigneeKey, 0)
	for rows.Next() {
		var k string
		var id int
		if rows.Scan(&k, &id) != nil {
			continue
		}
		list = append(list, caseTaskAssigneeKey{Kind: normalizeAssigneeKind(k), ID: id})
	}
	if len(list) > 0 {
		return list, nil
	}
	// Fallback: legacy columns
	var typ sql.NullString
	var id sql.NullInt64
	err = h.TranMySQL.DB.QueryRowContext(ctx, `
		SELECT assignee_type, assignee_id FROM CaseTask WHERE ID = ?`, caseTaskID).Scan(&typ, &id)
	if err != nil {
		return nil, err
	}
	if typ.Valid && id.Valid && id.Int64 > 0 {
		return []caseTaskAssigneeKey{{Kind: normalizeAssigneeKind(typ.String), ID: int(id.Int64)}}, nil
	}
	return nil, nil
}

func (h *Handlers) buildAssigneeOutList(ctx context.Context, keys []caseTaskAssigneeKey) []caseTaskAssigneeOut {
	out := make([]caseTaskAssigneeOut, 0, len(keys))
	for _, k := range keys {
		nm := h.caseTaskAssigneeName(ctx, k.Kind, k.ID)
		em := h.caseTaskAssigneeEmail(ctx, k.Kind, k.ID)
		out = append(out, caseTaskAssigneeOut{
			AssigneeKind: k.Kind,
			AssigneeID:   k.ID,
			Name:         nm,
			Email:        em,
		})
	}
	return out
}

func assigneesLabel(list []caseTaskAssigneeOut) string {
	if len(list) == 0 {
		return ""
	}
	names := make([]string, 0, len(list))
	for _, a := range list {
		if strings.TrimSpace(a.Name) != "" {
			names = append(names, a.Name)
		} else {
			names = append(names, fmt.Sprintf("%s #%d", a.AssigneeKind, a.AssigneeID))
		}
	}
	if len(names) <= 2 {
		return strings.Join(names, ", ")
	}
	return names[0] + ", " + names[1] + fmt.Sprintf(" +%d", len(names)-2)
}

func (h *Handlers) replaceCaseTaskAssignees(ctx context.Context, caseTaskID int, keys []caseTaskAssigneeKey) error {
	if _, err := h.TranMySQL.DB.ExecContext(ctx, `DELETE FROM CaseTaskAssignee WHERE case_task_id = ?`, caseTaskID); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return errors.New("CaseTaskAssignee table missing; run migration 035_case_task_multi_assignee.sql")
		}
		return err
	}
	for _, k := range keys {
		if _, err := h.TranMySQL.DB.ExecContext(ctx, `
			INSERT INTO CaseTaskAssignee (case_task_id, assignee_kind, assignee_id) VALUES (?, ?, ?)`,
			caseTaskID, k.Kind, k.ID,
		); err != nil {
			return err
		}
	}
	// Sync legacy columns to first assignee (optional compatibility).
	var lt interface{} = nil
	var lid interface{} = nil
	if len(keys) > 0 {
		lt = keys[0].Kind
		lid = keys[0].ID
	}
	_, err := h.TranMySQL.DB.ExecContext(ctx, `UPDATE CaseTask SET assignee_type = ?, assignee_id = ? WHERE ID = ?`, lt, lid, caseTaskID)
	return err
}

func (h *Handlers) loadAssigneesBatch(ctx context.Context, taskIDs []int) map[int][]caseTaskAssigneeKey {
	out := make(map[int][]caseTaskAssigneeKey)
	if len(taskIDs) == 0 || h.TranMySQL == nil {
		return out
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(taskIDs)), ",")
	args := make([]interface{}, 0, len(taskIDs))
	for _, id := range taskIDs {
		args = append(args, id)
	}
	q := `SELECT case_task_id, assignee_kind, assignee_id FROM CaseTaskAssignee
		WHERE case_task_id IN (` + placeholders + `) ORDER BY case_task_id, id`
	rows, err := h.TranMySQL.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var tid int
		var k string
		var aid int
		if rows.Scan(&tid, &k, &aid) != nil {
			continue
		}
		out[tid] = append(out[tid], caseTaskAssigneeKey{Kind: normalizeAssigneeKind(k), ID: aid})
	}
	return out
}

func (h *Handlers) ListCaseTasks(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	rows, err := h.TranMySQL.DB.Query("SELECT " + caseTaskListCols + " FROM CaseTask ORDER BY start_at DESC, ID DESC LIMIT 500")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	ctx := c.Request.Context()
	tmp := make([]caseTaskRow, 0, 64)
	ids := make([]int, 0, 64)
	for rows.Next() {
		r, err := parseCaseTaskRow(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tmp = append(tmp, r)
		ids = append(ids, r.ID)
	}
	batch := h.loadAssigneesBatch(ctx, ids)
	list := make([]caseTaskRow, 0, len(tmp))
	for _, r := range tmp {
		keys := batch[r.ID]
		if len(keys) == 0 && r.AssigneeType != "" && r.AssigneeID > 0 {
			keys = []caseTaskAssigneeKey{{Kind: normalizeAssigneeKind(r.AssigneeType), ID: r.AssigneeID}}
		}
		r.Assignees = h.buildAssigneeOutList(ctx, keys)
		r.AssigneesLabel = assigneesLabel(r.Assignees)
		if len(r.Assignees) > 0 {
			r.AssigneeName = r.Assignees[0].Name
			r.AssigneeType = r.Assignees[0].AssigneeKind
			r.AssigneeID = r.Assignees[0].AssigneeID
		} else if r.AssigneeType != "" && r.AssigneeID > 0 {
			r.AssigneeName = h.caseTaskAssigneeName(ctx, r.AssigneeType, r.AssigneeID)
		}
		list = append(list, r)
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handlers) GetCaseTask(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row := h.TranMySQL.DB.QueryRow("SELECT "+caseTaskListCols+" FROM CaseTask WHERE ID = ?", id)
	item, err := parseCaseTaskRow(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case/task not found"})
		return
	}
	keys, _ := h.assigneesKeysForCaseTask(c.Request.Context(), item.ID)
	item.Assignees = h.buildAssigneeOutList(c.Request.Context(), keys)
	item.AssigneesLabel = assigneesLabel(item.Assignees)
	if len(item.Assignees) > 0 {
		item.AssigneeName = item.Assignees[0].Name
		item.AssigneeType = item.Assignees[0].AssigneeKind
		item.AssigneeID = item.Assignees[0].AssigneeID
	} else if item.AssigneeType != "" && item.AssigneeID > 0 {
		item.AssigneeName = h.caseTaskAssigneeName(c.Request.Context(), item.AssigneeType, item.AssigneeID)
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handlers) GetCaseTaskFull(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, `
		SELECT
			ID AS id,
			title AS title,
			description AS description,
			start_at AS start_at,
			end_at AS end_at,
			location AS location,
			assignee_type AS assignee_type,
			assignee_id AS assignee_id,
			created_on AS created_on,
			last_updated AS last_updated
		FROM CaseTask
		WHERE ID = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "case/task not found"})
		return
	}
	assigneeType, _ := m["assignee_type"].(string)
	assigneeID, _ := m["assignee_id"].(int64)
	if assigneeID == 0 {
		if fv, okf := m["assignee_id"].(float64); okf {
			assigneeID = int64(fv)
		}
	}
	if assigneeType != "" && assigneeID > 0 {
		m["assignee_name"] = h.caseTaskAssigneeName(c.Request.Context(), assigneeType, int(assigneeID))
	}
	akeys, _ := h.assigneesKeysForCaseTask(c.Request.Context(), id)
	aout := h.buildAssigneeOutList(c.Request.Context(), akeys)
	m["assignees"] = aout
	m["assignees_label"] = assigneesLabel(aout)

	h.attachEntityDetail(c, entityKeyCaseTask, id, m)
	h.attachEntityAttachmentsToRow(c.Request.Context(), "case-tasks", entityAttachmentCaseTask, id, m)
	c.JSON(http.StatusOK, m)
}

func (h *Handlers) CreateCaseTask(c *gin.Context) {
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

	title, _ := in["title"].(string)
	if strings.TrimSpace(title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	assigneeKeys, hadAssignees, perr := parseAssigneesFromInput(in)
	if perr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": perr.Error()})
		return
	}
	if !hadAssignees || len(assigneeKeys) == 0 {
		leg, _ := legacyAssigneesFromInput(in)
		assigneeKeys = leg
	}
	delete(in, "assignee_type")
	delete(in, "assignee_id")
	for _, k := range assigneeKeys {
		if err := h.validateCaseTaskAssigneeRef(c.Request.Context(), k.Kind, k.ID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	cols := make([]string, 0, 8)
	args := make([]interface{}, 0, 8)
	for k, v := range in {
		col := strings.ToLower(strings.TrimSpace(k))
		if col == "id" || col == "db_id" {
			continue
		}
		if _, ok := allowedCaseTaskWrite[col]; !ok {
			continue
		}
		if col == "title" {
			continue
		}
		val := v
		switch col {
		case "description":
			if s, ok := v.(string); ok {
				s = strings.TrimSpace(s)
				if s == "" {
					val = nil
				} else {
					val = s
				}
			}
		case "start_at", "end_at":
			if s, ok := v.(string); ok {
				s = strings.TrimSpace(s)
				if s == "" {
					val = nil
				} else {
					val = s
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

	cols = append(cols, "title")
	args = append(args, strings.TrimSpace(title))

	stmt := "INSERT INTO CaseTask (" + strings.Join(cols, ",") + ") VALUES (" + strings.TrimRight(strings.Repeat("?,", len(cols)), ",") + ")"
	res, err := h.TranMySQL.DB.Exec(stmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	id := int(id64)
	if err := h.replaceCaseTaskAssignees(c.Request.Context(), id, assigneeKeys); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyCaseTask, id, detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT ID AS id, title, description, start_at, end_at, location, assignee_type, assignee_id, created_on, last_updated FROM CaseTask WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id})
		return
	}
	h.attachEntityDetail(c, entityKeyCaseTask, id, m)
	akeys, _ := h.assigneesKeysForCaseTask(c.Request.Context(), id)
	aout := h.buildAssigneeOutList(c.Request.Context(), akeys)
	m["assignees"] = aout
	m["assignees_label"] = assigneesLabel(aout)
	if len(aout) > 0 {
		m["assignee_name"] = aout[0].Name
	}
	h.attachEntityAttachmentsToRow(c.Request.Context(), "case-tasks", entityAttachmentCaseTask, id, m)
	c.JSON(http.StatusOK, m)
}

func (h *Handlers) UpdateCaseTask(c *gin.Context) {
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

	var curTitle string
	if err := h.TranMySQL.DB.QueryRow("SELECT title FROM CaseTask WHERE ID = ?", id).Scan(&curTitle); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case/task not found"})
		return
	}
	nextTitle := curTitle

	var newAssigneeKeys []caseTaskAssigneeKey
	var replaceAssignees bool
	if keys, had, perr := parseAssigneesFromInput(in); perr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": perr.Error()})
		return
	} else if had {
		replaceAssignees = true
		newAssigneeKeys = keys
	}
	delete(in, "assignee_type")
	delete(in, "assignee_id")

	sets := make([]string, 0, len(in))
	args := make([]interface{}, 0, len(in)+1)
	for k, v := range in {
		col := strings.ToLower(strings.TrimSpace(k))
		if col == "id" || col == "db_id" {
			continue
		}
		if _, ok := allowedCaseTaskWrite[col]; !ok {
			continue
		}
		val := v
		switch col {
		case "title":
			if s, ok := v.(string); ok {
				s = strings.TrimSpace(s)
				if s == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "title cannot be empty"})
					return
				}
				val = s
				nextTitle = s
			}
		case "description":
			if s, ok := v.(string); ok {
				s = strings.TrimSpace(s)
				if s == "" {
					val = nil
				} else {
					val = s
				}
			}
		case "start_at", "end_at":
			if s, ok := v.(string); ok {
				s = strings.TrimSpace(s)
				if s == "" {
					val = nil
				} else {
					val = s
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

	if strings.TrimSpace(nextTitle) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if len(sets) > 0 {
		sets = append(sets, "last_updated = CURRENT_TIMESTAMP")
		args = append(args, id)
		stmt := "UPDATE CaseTask SET " + strings.Join(sets, ", ") + " WHERE ID = ?"
		if _, err := h.TranMySQL.DB.Exec(stmt, args...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if replaceAssignees {
		for _, k := range newAssigneeKeys {
			if err := h.validateCaseTaskAssigneeRef(c.Request.Context(), k.Kind, k.ID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
		if err := h.replaceCaseTaskAssignees(c.Request.Context(), id, newAssigneeKeys); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	finalKeys, _ := h.assigneesKeysForCaseTask(c.Request.Context(), id)

	if hasDetail {
		_ = h.savePoppedDetail(c, entityKeyCaseTask, id, detailStr)
	}
	m, _, err := querySingleRowMap(h.TranMySQL.DB, "SELECT ID AS id, title, description, start_at, end_at, location, assignee_type, assignee_id, created_on, last_updated FROM CaseTask WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"updated": id})
		return
	}
	h.attachEntityDetail(c, entityKeyCaseTask, id, m)
	aout := h.buildAssigneeOutList(c.Request.Context(), finalKeys)
	m["assignees"] = aout
	m["assignees_label"] = assigneesLabel(aout)
	if len(aout) > 0 {
		m["assignee_name"] = aout[0].Name
	}
	h.attachEntityAttachmentsToRow(c.Request.Context(), "case-tasks", entityAttachmentCaseTask, id, m)
	c.JSON(http.StatusOK, m)
}

func (h *Handlers) DeleteCaseTask(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec("DELETE FROM CaseTask WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "case/task not found"})
		return
	}
	h.purgeEntityAttachments(context.Background(), entityAttachmentCaseTask, id)
	h.deleteEntityDetailMongo(context.Background(), entityKeyCaseTask, id)
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

func truncatePDFLine(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max-1]) + "…"
}

func parseJSONPretty(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "{}"
	}
	var obj interface{}
	if err := json.Unmarshal([]byte(v), &obj); err != nil {
		return v
	}
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return v
	}
	return string(b)
}

func buildCaseTaskPDF(caseTaskID int, title string, description string, assigneesSummary string, startAt, endAt *time.Time, location string, detailJSON string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 12)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 15)
	pdf.CellFormat(0, 9, "Case & Task Assignment", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(90, 90, 90)
	pdf.CellFormat(0, 6, fmt.Sprintf("Case/Task ID: %d", caseTaskID), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(2)

	writeKV := func(k, v string) {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(48, 6, k, "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 6, v, "", "L", false)
	}
	writeKV("Title", title)
	if strings.TrimSpace(assigneesSummary) != "" {
		writeKV("Assignees", assigneesSummary)
	}
	if startAt != nil {
		writeKV("Start", startAt.Local().Format("2006-01-02 15:04"))
	}
	if endAt != nil {
		writeKV("End", endAt.Local().Format("2006-01-02 15:04"))
	}
	if strings.TrimSpace(description) != "" {
		writeKV("Description", description)
	}
	if strings.TrimSpace(location) != "" {
		writeKV("Location", truncatePDFLine(location, 600))
	}

	pdf.Ln(2)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 7, "Details", "", 1, "L", false, 0, "")
	pdf.SetFont("Courier", "", 8)
	pdf.MultiCell(0, 4, parseJSONPretty(detailJSON), "", "L", false)

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func smtpAddrFromEnv() (string, string, string, string, string) {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := os.Getenv("SMTP_PASS")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if from == "" {
		from = user
	}
	return host, port, user, pass, from
}

func assigneesMultilineForPDF(ctx context.Context, h *Handlers, keys []caseTaskAssigneeKey) string {
	if len(keys) == 0 {
		return ""
	}
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		nm := h.caseTaskAssigneeName(ctx, k.Kind, k.ID)
		if strings.TrimSpace(nm) == "" {
			nm = fmt.Sprintf("%s #%d", k.Kind, k.ID)
		}
		lines = append(lines, fmt.Sprintf("• %s (%s)", nm, k.Kind))
	}
	return strings.Join(lines, "\n")
}

func (h *Handlers) buildCaseTaskPDFBytes(ctx context.Context, id int) ([]byte, error) {
	var title string
	var description, location sql.NullString
	var startAt, endAt sql.NullTime
	err := h.TranMySQL.DB.QueryRow(`
		SELECT title, description, start_at, end_at, location FROM CaseTask WHERE ID = ?`, id).
		Scan(&title, &description, &startAt, &endAt, &location)
	if err != nil {
		return nil, err
	}
	keys, err := h.assigneesKeysForCaseTask(ctx, id)
	if err != nil {
		return nil, err
	}
	summary := assigneesMultilineForPDF(ctx, h, keys)
	detailJSON := "{}"
	if store := h.entityDetailStore(); store != nil {
		if dj, e := store.GetEntityDetailJSON(ctx, entityKeyCaseTask, id); e == nil && strings.TrimSpace(dj) != "" {
			detailJSON = dj
		}
	}
	var startPtr, endPtr *time.Time
	if startAt.Valid {
		t := startAt.Time
		startPtr = &t
	}
	if endAt.Valid {
		t := endAt.Time
		endPtr = &t
	}
	return buildCaseTaskPDF(id, title, description.String, summary, startPtr, endPtr, location.String, detailJSON)
}

var (
	htmlScriptStripper = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	htmlTagStripper    = regexp.MustCompile(`<[^>]+>`)
	htmlSpaceCollapse  = regexp.MustCompile(`\s+`)
)

func htmlToPlainEmail(html string) string {
	s := htmlScriptStripper.ReplaceAllString(html, " ")
	s = htmlTagStripper.ReplaceAllString(s, " ")
	return strings.TrimSpace(htmlSpaceCollapse.ReplaceAllString(s, " "))
}

func sendEmailWithHTML(to, subject, htmlBody string) error {
	host, port, user, pass, from := smtpAddrFromEnv()
	if host == "" || port == "" || from == "" {
		return errors.New("SMTP is not configured: set SMTP_HOST, SMTP_PORT, SMTP_FROM (and SMTP_USER/SMTP_PASS if needed)")
	}
	plain := htmlToPlainEmail(htmlBody)
	if plain == "" {
		plain = "This message is in HTML format."
	}
	boundary := "bnd_" + strconv.FormatInt(time.Now().UnixNano(), 10)

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
	msg.WriteString("\r\n")
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	msg.WriteString(plain + "\r\n\r\n")
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	msg.WriteString(htmlBody + "\r\n\r\n")
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	addr := host + ":" + port
	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg.String()))
}

// DownloadCaseTaskPDF serves the generated case/task PDF (for preview).
func (h *Handlers) DownloadCaseTaskPDF(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var n int
	if err := h.TranMySQL.DB.QueryRow("SELECT COUNT(*) FROM CaseTask WHERE ID = ?", id).Scan(&n); err != nil || n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "case/task not found"})
		return
	}
	pdfBytes, err := h.buildCaseTaskPDFBytes(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF: " + err.Error()})
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="case-task-%d.pdf"`, id))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// SendCaseTaskEmail sends multipart/alternative email (plain + HTML) to selected recipients (member / employee / contact).
func (h *Handlers) SendCaseTaskEmail(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var n int
	if err := h.TranMySQL.DB.QueryRow("SELECT COUNT(*) FROM CaseTask WHERE ID = ?", id).Scan(&n); err != nil || n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "case/task not found"})
		return
	}
	var in struct {
		Subject    string `json:"subject"`
		HTMLBody   string `json:"html_body"`
		Recipients []struct {
			Kind string `json:"kind"`
			ID   int    `json:"id"`
		} `json:"recipients"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(in.Subject) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject is required"})
		return
	}
	if strings.TrimSpace(in.HTMLBody) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "html_body is required"})
		return
	}
	if len(in.Recipients) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one recipient is required"})
		return
	}
	ctx := c.Request.Context()
	sent := 0
	var missing []string
	for _, r := range in.Recipients {
		k := normalizeAssigneeKind(r.Kind)
		if k != "member" && k != "employee" && k != "contact" {
			continue
		}
		if r.ID <= 0 {
			continue
		}
		if err := h.validateCaseTaskAssigneeRef(ctx, k, r.ID); err != nil {
			missing = append(missing, fmt.Sprintf("%s:%d (invalid)", k, r.ID))
			continue
		}
		email := h.caseTaskAssigneeEmail(ctx, k, r.ID)
		if email == "" {
			missing = append(missing, fmt.Sprintf("%s:%d (no email)", k, r.ID))
			continue
		}
		if err := sendEmailWithHTML(email, in.Subject, in.HTMLBody); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email: " + err.Error()})
			return
		}
		sent++
	}
	if sent == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no messages sent; check recipient emails", "skipped": missing})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sent": sent, "case_task_id": id, "skipped": missing})
}
