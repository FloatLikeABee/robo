package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

func scanUserNoteTodo(row *sql.Row) (models.UserNoteTodo, error) {
	var r models.UserNoteTodo
	var title, body sql.NullString
	var deadlineAt sql.NullTime
	err := row.Scan(&r.ID, &r.UserID, &r.ItemType, &title, &body, &r.Completed, &deadlineAt, &r.CreatedOn, &r.LastUpdated)
	if err != nil {
		return r, err
	}
	if title.Valid {
		t := title.String
		r.Title = &t
	}
	if body.Valid {
		b := body.String
		r.Body = &b
	}
	if deadlineAt.Valid {
		d := deadlineAt.Time
		r.DeadlineAt = &d
	}
	return r, nil
}

func scanUserNoteTodoRows(rows *sql.Rows) ([]models.UserNoteTodo, error) {
	var list []models.UserNoteTodo
	for rows.Next() {
		var r models.UserNoteTodo
		var title, body sql.NullString
		var deadlineAt sql.NullTime
		if err := rows.Scan(&r.ID, &r.UserID, &r.ItemType, &title, &body, &r.Completed, &deadlineAt, &r.CreatedOn, &r.LastUpdated); err != nil {
			return nil, err
		}
		if title.Valid {
			t := title.String
			r.Title = &t
		}
		if body.Valid {
			b := body.String
			r.Body = &b
		}
		if deadlineAt.Valid {
			d := deadlineAt.Time
			r.DeadlineAt = &d
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func parseDeadlineAtRaw(raw json.RawMessage) (bool, *time.Time, error) {
	if len(raw) == 0 {
		return false, nil, nil
	}
	var rawString string
	if err := json.Unmarshal(raw, &rawString); err != nil {
		trimmed := strings.TrimSpace(string(raw))
		if trimmed == "null" {
			return true, nil, nil
		}
		return true, nil, err
	}
	rawString = strings.TrimSpace(rawString)
	if rawString == "" {
		return true, nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, rawString)
	if err != nil {
		return true, nil, err
	}
	utc := t.UTC()
	return true, &utc, nil
}

// ListUserNotesTodos returns notes and TODOs for the current user. Query: type=all|note|todo
func (h *Handlers) ListUserNotesTodos(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	userID := h.tranUserIDFromContext(c)
	filter := strings.ToLower(strings.TrimSpace(c.DefaultQuery("type", "all")))
	var rows *sql.Rows
	var err error
	switch filter {
	case "note":
		rows, err = h.TranMySQL.DB.Query(
			`SELECT ID, UserID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
			 FROM user_note_todo WHERE UserID = ? AND ItemType = 'note' ORDER BY CreatedOn DESC LIMIT 500`,
			userID)
	case "todo":
		rows, err = h.TranMySQL.DB.Query(
			`SELECT ID, UserID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
			 FROM user_note_todo WHERE UserID = ? AND ItemType = 'todo'
			 ORDER BY Completed ASC, (DeadlineAt IS NULL) ASC, DeadlineAt ASC, CreatedOn DESC LIMIT 500`,
			userID)
	default:
		rows, err = h.TranMySQL.DB.Query(
			`SELECT ID, UserID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
			 FROM user_note_todo WHERE UserID = ? ORDER BY ItemType ASC, Completed ASC, (DeadlineAt IS NULL) ASC, DeadlineAt ASC, CreatedOn DESC LIMIT 500`,
			userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	list, err := scanUserNoteTodoRows(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetUserNoteTodo returns one row if it belongs to the current user.
func (h *Handlers) GetUserNoteTodo(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID := h.tranUserIDFromContext(c)
	row := h.TranMySQL.DB.QueryRow(
		`SELECT ID, UserID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
		 FROM user_note_todo WHERE ID = ? AND UserID = ?`, id, userID)
	r, err := scanUserNoteTodo(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

// CreateUserNoteTodo creates a note or TODO for the current user.
func (h *Handlers) CreateUserNoteTodo(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in struct {
		ItemType   string          `json:"item_type"`
		Title      string          `json:"title"`
		Body       string          `json:"body"`
		Completed  *bool           `json:"completed"`
		DeadlineAt json.RawMessage `json:"deadline_at"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	itemType := strings.ToLower(strings.TrimSpace(in.ItemType))
	if itemType != "note" && itemType != "todo" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_type must be note or todo"})
		return
	}
	userID := h.tranUserIDFromContext(c)
	completed := false
	if itemType == "todo" && in.Completed != nil {
		completed = *in.Completed
	}
	deadlineProvided, deadlineAt, err := parseDeadlineAtRaw(in.DeadlineAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deadline_at must be RFC3339 datetime"})
		return
	}
	if itemType != "todo" && deadlineProvided && deadlineAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deadline_at applies to TODO items only"})
		return
	}
	var title interface{}
	if strings.TrimSpace(in.Title) != "" {
		title = strings.TrimSpace(in.Title)
	} else {
		title = nil
	}
	var body interface{}
	if strings.TrimSpace(in.Body) != "" {
		body = strings.TrimSpace(in.Body)
	} else {
		body = nil
	}
	var deadlineArg interface{}
	if deadlineAt != nil {
		deadlineArg = *deadlineAt
	}
	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO user_note_todo (UserID, ItemType, Title, Body, Completed, DeadlineAt) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, itemType, title, body, completed, deadlineArg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id64})
}

// UpdateUserNoteTodo updates title, body, and optionally completed (TODOs).
func (h *Handlers) UpdateUserNoteTodo(c *gin.Context) {
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
		Title      *string         `json:"title"`
		Body       *string         `json:"body"`
		Completed  *bool           `json:"completed"`
		DeadlineAt json.RawMessage `json:"deadline_at"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if in.Title == nil && in.Body == nil && in.Completed == nil && len(in.DeadlineAt) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title, body, completed, or deadline_at required"})
		return
	}
	userID := h.tranUserIDFromContext(c)
	row := h.TranMySQL.DB.QueryRow(
		`SELECT ID, UserID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
		 FROM user_note_todo WHERE ID = ? AND UserID = ?`, id, userID)
	cur, err := scanUserNoteTodo(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if in.Completed != nil && cur.ItemType != "todo" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "completed applies to TODO items only"})
		return
	}
	deadlineProvided, deadlineAt, err := parseDeadlineAtRaw(in.DeadlineAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deadline_at must be RFC3339 datetime"})
		return
	}
	if deadlineProvided && deadlineAt != nil && cur.ItemType != "todo" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deadline_at applies to TODO items only"})
		return
	}
	title := cur.Title
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			title = nil
		} else {
			title = &t
		}
	}
	body := cur.Body
	if in.Body != nil {
		b := strings.TrimSpace(*in.Body)
		if b == "" {
			body = nil
		} else {
			body = &b
		}
	}
	completed := cur.Completed
	if in.Completed != nil {
		completed = *in.Completed
	}
	deadline := cur.DeadlineAt
	if deadlineProvided {
		deadline = deadlineAt
	}
	var titleArg interface{}
	if title != nil {
		titleArg = *title
	}
	var bodyArg interface{}
	if body != nil {
		bodyArg = *body
	}
	var deadlineArg interface{}
	if deadline != nil {
		deadlineArg = *deadline
	}
	_, err = h.TranMySQL.DB.Exec(
		`UPDATE user_note_todo SET Title = ?, Body = ?, Completed = ?, DeadlineAt = ? WHERE ID = ? AND UserID = ?`,
		titleArg, bodyArg, completed, deadlineArg, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DeleteUserNoteTodo deletes a row if it belongs to the current user.
func (h *Handlers) DeleteUserNoteTodo(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID := h.tranUserIDFromContext(c)
	res, err := h.TranMySQL.DB.Exec(`DELETE FROM user_note_todo WHERE ID = ? AND UserID = ?`, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
