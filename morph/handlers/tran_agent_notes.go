package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

const (
	agentNoteTitleMark = "[morph-mcp]"
	agentNoteBodyMark  = "source: morph-mcp"
)

// agentNoteMarkerSQL keeps a note when either create_note marker is still present.
// instr is 1-based and case-sensitive, so a marker later in the text does not count.
const agentNoteMarkerSQL = `(instr(COALESCE(Title, ''), '` + agentNoteTitleMark + `') = 1 OR instr(COALESCE(Body, ''), '` + agentNoteBodyMark + `') = 1)`

// sessionAgentNoteUser resolves the one active Tran user for the verified session.
// It ignores ?user_id= and X-User-ID. Zero matches and two matches both fail closed.
func (h *Handlers) sessionAgentNoteUser(c *gin.Context) (int, bool) {
	if h == nil || h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return 0, false
	}
	raw, ok := c.Get("auth_user_id")
	subject, _ := raw.(string)
	subject = strings.TrimSpace(subject)
	if !ok || subject == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return 0, false
	}
	rows, err := h.TranMySQL.DB.Query(`
		SELECT u.UserID
		FROM "User" AS u
		INNER JOIN plat_users AS p
			ON p.email IS NOT NULL
			AND u.Email IS NOT NULL
			AND LOWER(TRIM(p.email)) = LOWER(TRIM(u.Email))
		WHERE p.id = ? AND u.Deactivated = 0`, subject)
	if err != nil {
		log.Printf("agent notes owner: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load notes"})
		return 0, false
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Printf("agent notes owner: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load notes"})
			return 0, false
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		log.Printf("agent notes owner: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load notes"})
		return 0, false
	}
	if len(ids) != 1 {
		c.JSON(http.StatusConflict, gin.H{"error": "session does not map to one notes user"})
		return 0, false
	}
	return ids[0], true
}

func (h *Handlers) queryAgentNotes(userID, id int) ([]models.UserNoteTodo, error) {
	q := `SELECT ID, UserID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
		FROM user_note_todo
		WHERE UserID = ? AND ItemType = 'note' AND ` + agentNoteMarkerSQL
	args := []any{userID}
	if id > 0 {
		q += ` AND ID = ?`
		args = append(args, id)
	}
	q += ` ORDER BY CreatedOn DESC LIMIT 200`
	rows, err := h.TranMySQL.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := scanUserNoteTodoRows(rows)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.UserNoteTodo{}
	}
	return list, nil
}

// ListAgentNotes returns the signed-in user's agent-authored notes.
func (h *Handlers) ListAgentNotes(c *gin.Context) {
	userID, ok := h.sessionAgentNoteUser(c)
	if !ok {
		return
	}
	list, err := h.queryAgentNotes(userID, 0)
	if err != nil {
		log.Printf("agent notes list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load notes"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetAgentNote returns one agent note when it belongs to the signed-in user.
// A foreign id and a missing id are both not found, with no note text.
func (h *Handlers) GetAgentNote(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID, ok := h.sessionAgentNoteUser(c)
	if !ok {
		return
	}
	list, err := h.queryAgentNotes(userID, id)
	if err != nil {
		log.Printf("agent notes get: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load notes"})
		return
	}
	if len(list) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, list[0])
}
