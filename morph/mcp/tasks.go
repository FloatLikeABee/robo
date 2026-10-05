package mcp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

const (
	defaultTaskLimit = 50
	maxTaskLimit     = 100
	maxNoteTitleLen  = 200
	maxNoteBodyLen   = 32000
	agentTitleMark   = "[morph-mcp]"
	agentBodyMark    = "source: morph-mcp"
)

// ErrNotFound means the id is missing or belongs to someone else.
// The text is the same in both cases so a caller cannot tell them apart.
var ErrNotFound = errors.New("not found")

// Task is one user_note_todo row for the signed-in user.
// Status is "open" or "done". Dates are the SQLite text when it is non-empty.
type Task struct {
	ID          int    `json:"id"`
	ItemType    string `json:"item_type"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Body        string `json:"body"`
	DeadlineAt  string `json:"deadline_at,omitempty"`
	CreatedOn   string `json:"created_on,omitempty"`
	LastUpdated string `json:"last_updated,omitempty"`
}

// TaskFilter selects list_my_tasks rows. Zero values mean all types, all
// statuses, and the default limit.
type TaskFilter struct {
	Type   string
	Status string
	Limit  int
}

// ListResult is a page of the caller's notes and todos.
type ListResult struct {
	Tasks []Task `json:"tasks"`
	Limit int    `json:"limit"`
}

// CheckStartup verifies the JWT secret and token, opens TRAN_SQLITE_PATH
// read-only, and confirms the subject still exists in plat_users.
func CheckStartup(ctx context.Context) (Identity, *sql.DB, error) {
	id, err := IdentityFromEnv()
	if err != nil {
		return Identity{}, nil, err
	}
	path := SQLitePath()
	db, err := OpenReadOnly(path)
	if err != nil {
		return Identity{}, nil, fmt.Errorf("TRAN_SQLITE_PATH: %w", err)
	}
	if err := ConfirmPlatUser(ctx, db, id.UserID); err != nil {
		_ = db.Close()
		return Identity{}, nil, err
	}
	return id, db, nil
}

// SQLitePath is TRAN_SQLITE_PATH, or the API default when that variable is unset.
func SQLitePath() string {
	path := strings.TrimSpace(os.Getenv("TRAN_SQLITE_PATH"))
	if path == "" {
		return "./data/tran.sqlite"
	}
	return path
}

// OpenReadOnly opens TRAN_SQLITE_PATH without taking a write lock or running
// migrations. A file: URI is required: modernc treats "path?mode=ro" as a
// filename and would create the file. mode=ro plus _query_only=1 rejects
// writes even when the OS file is writable. journal_mode is left to the API.
func OpenReadOnly(path string) (*sql.DB, error) {
	return openSQLite(path, "mode=ro&_query_only=1&_busy_timeout=5000", false)
}

// OpenReadWrite opens TRAN_SQLITE_PATH for one create_note insert.
// mode=rw fails when the file is missing, so this does not create a database.
// It does not set journal_mode and does not migrate.
func OpenReadWrite(path string) (*sql.DB, error) {
	return openSQLite(path, "mode=rw&_busy_timeout=5000", true)
}

func openSQLite(path, rawQuery string, mustExist bool) (*sql.DB, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("sqlite path is empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if mustExist {
		if _, err := os.Stat(abs); err != nil {
			return nil, err
		}
	}
	dsn := (&url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(abs),
		RawQuery: rawQuery,
	}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// CreateMyNote inserts one note for the verified JWT subject.
// tokenUserID is plat_users.id. The owner subquery is the same one ownerClause uses.
// An identical stored title and body returns the existing row.
// ponytail: two concurrent identical calls can both insert; a unique index would block the SPA's duplicate titles.
func CreateMyNote(ctx context.Context, db *sql.DB, tokenUserID, title, body string) (Task, error) {
	if db == nil {
		return Task{}, errors.New("note store is not open")
	}
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" && body == "" {
		return Task{}, errors.New("title or body is required")
	}
	if utf8.RuneCountInString(title) > maxNoteTitleLen {
		return Task{}, errors.New("title is too long")
	}
	if utf8.RuneCountInString(body) > maxNoteBodyLen {
		return Task{}, errors.New("body is too long")
	}
	if _, _, ok := ownerClause(tokenUserID); !ok {
		return Task{}, errors.New("not a notes user for this session")
	}
	storedTitle := agentTitle(title)
	storedBody := agentBody(body)
	res, err := db.ExecContext(ctx, `
		INSERT INTO user_note_todo (UserID, ItemType, Title, Body, Completed, DeadlineAt)
		SELECT owner.UserID, 'note', ?, ?, 0, NULL
		FROM (`+ownerUserSelect()+`) AS owner
		WHERE NOT EXISTS (
			SELECT 1 FROM user_note_todo AS existing
			WHERE existing.UserID = owner.UserID
				AND existing.ItemType = 'note'
				AND existing.Title = ?
				AND existing.Body = ?
		)`, storedTitle, storedBody, strings.TrimSpace(tokenUserID), storedTitle, storedBody)
	if err != nil {
		log.Printf("create note: %v", err)
		return Task{}, errors.New("could not store the note")
	}
	if n, _ := res.RowsAffected(); n > 1 {
		log.Printf("create note: inserted %d rows", n)
		return Task{}, errors.New("could not store the note")
	}
	task, err := findMyNote(ctx, db, tokenUserID, storedTitle, storedBody)
	if errors.Is(err, ErrNotFound) {
		return Task{}, errors.New("not a notes user for this session")
	}
	return task, err
}

func agentTitle(title string) string {
	if title == "" {
		return agentTitleMark
	}
	if title == agentTitleMark || strings.HasPrefix(title, agentTitleMark+" ") {
		return title
	}
	return agentTitleMark + " " + title
}

func agentBody(body string) string {
	if body == "" || body == agentBodyMark || strings.HasPrefix(body, agentBodyMark+"\n") {
		if body == "" {
			return agentBodyMark
		}
		return body
	}
	return agentBodyMark + "\n\n" + body
}

func findMyNote(ctx context.Context, db *sql.DB, tokenUserID, title, body string) (Task, error) {
	clause, clauseArgs, ok := ownerClause(tokenUserID)
	if !ok {
		return Task{}, ErrNotFound
	}
	row := db.QueryRowContext(ctx, `
		SELECT ID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
		FROM user_note_todo
		WHERE ItemType = 'note' AND Title = ? AND Body = ? AND `+clause+`
		ORDER BY ID DESC LIMIT 1`, append([]any{title, body}, clauseArgs...)...)
	task, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	return task, err
}

// ListMyTasks returns the caller's user_note_todo rows. tokenUserID is the
// verified JWT subject (plat_users.id). A subject with no active Tran user
// yields an empty page, not another user's rows.
func ListMyTasks(ctx context.Context, db *sql.DB, tokenUserID string, f TaskFilter) (ListResult, error) {
	limit, err := appliedLimit(f.Limit)
	if err != nil {
		return ListResult{}, err
	}
	itemType, err := normalizeType(f.Type)
	if err != nil {
		return ListResult{}, err
	}
	status, err := normalizeStatus(f.Status)
	if err != nil {
		return ListResult{}, err
	}
	out := ListResult{Tasks: []Task{}, Limit: limit}
	if db == nil {
		return ListResult{}, errors.New("task store is not open")
	}
	q, args, ok := listQuery(tokenUserID, itemType, status, limit)
	if !ok {
		return out, nil
	}
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return ListResult{}, err
	}
	defer rows.Close()
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return ListResult{}, err
		}
		out.Tasks = append(out.Tasks, task)
	}
	if err := rows.Err(); err != nil {
		return ListResult{}, err
	}
	return out, nil
}

// GetMyTask returns one of the caller's rows. tokenUserID is the verified
// JWT subject. Any other id is ErrNotFound.
func GetMyTask(ctx context.Context, db *sql.DB, tokenUserID string, id int) (Task, error) {
	if db == nil {
		return Task{}, errors.New("task store is not open")
	}
	if id <= 0 {
		return Task{}, errors.New("invalid id")
	}
	clause, clauseArgs, ok := ownerClause(tokenUserID)
	if !ok {
		return Task{}, ErrNotFound
	}
	row := db.QueryRowContext(ctx, `
		SELECT ID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
		FROM user_note_todo WHERE ID = ? AND `+clause, append([]any{id}, clauseArgs...)...)
	task, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	return task, err
}

// ConfirmPlatUser fails closed when the JWT subject is not a plat_users.id.
func ConfirmPlatUser(ctx context.Context, db *sql.DB, subject string) error {
	subject = strings.TrimSpace(subject)
	if db == nil || subject == "" {
		return errors.New("MORPH_MCP_TOKEN subject is not a Morph user")
	}
	var one int
	err := db.QueryRowContext(ctx, `SELECT 1 FROM plat_users WHERE id = ? LIMIT 1`, subject).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && one != 1) {
		return errors.New("MORPH_MCP_TOKEN subject is not a Morph user")
	}
	if err != nil {
		return fmt.Errorf("plat_users: %w", err)
	}
	return nil
}

func appliedLimit(n int) (int, error) {
	switch {
	case n == 0:
		return defaultTaskLimit, nil
	case n < 0:
		return 0, errors.New("invalid limit")
	case n > maxTaskLimit:
		return maxTaskLimit, nil
	default:
		return n, nil
	}
}

func normalizeType(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "all":
		return "", nil
	case "note", "todo":
		return strings.ToLower(strings.TrimSpace(raw)), nil
	default:
		return "", errors.New("invalid type")
	}
}

func normalizeStatus(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "all":
		return "", nil
	case "open", "done":
		return strings.ToLower(strings.TrimSpace(raw)), nil
	default:
		return "", errors.New("invalid status")
	}
}

// ownerUserSelect is the only copy of the Tran-user join. ownerClause and
// CreateMyNote both use it. Issue #69 will add a shared visibility function
// for REST and MCP. Replace this body with that call.
// ponytail: LIMIT 1 if two active User rows share the plat user's email; a unique email index is the upgrade.
func ownerUserSelect() string {
	return `SELECT u.UserID
		FROM "User" AS u
		INNER JOIN plat_users AS p
			ON p.email IS NOT NULL
			AND u.Email IS NOT NULL
			AND LOWER(TRIM(p.email)) = LOWER(TRIM(u.Email))
		WHERE p.id = ? AND u.Deactivated = 0
		LIMIT 1`
}

// ownerClause is the only owner check in this package. Every user_note_todo
// read ANDs it into the WHERE clause. tokenUserID is the verified JWT subject
// (plat_users.id), bound as p.id, so the statement cannot return another
// user's rows.
func ownerClause(tokenUserID string) (string, []any, bool) {
	tokenUserID = strings.TrimSpace(tokenUserID)
	if tokenUserID == "" {
		return "", nil, false
	}
	return `UserID = (` + ownerUserSelect() + `)`, []any{tokenUserID}, true
}

func listQuery(tokenUserID, itemType, status string, limit int) (string, []any, bool) {
	clause, args, ok := ownerClause(tokenUserID)
	if !ok {
		return "", nil, false
	}
	// Order matches handlers.ListUserNotesTodos for the same type filter.
	order := `ItemType ASC, Completed ASC, (DeadlineAt IS NULL) ASC, DeadlineAt ASC, CreatedOn DESC`
	if itemType == "note" {
		order = `CreatedOn DESC`
	} else if itemType == "todo" {
		order = `Completed ASC, (DeadlineAt IS NULL) ASC, DeadlineAt ASC, CreatedOn DESC`
	}
	q := `SELECT ID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
		FROM user_note_todo WHERE ` + clause
	if itemType != "" {
		q += ` AND ItemType = ?`
		args = append(args, itemType)
	}
	switch status {
	case "open":
		q += ` AND Completed = 0`
	case "done":
		q += ` AND Completed = 1`
	}
	q += ` ORDER BY ` + order + ` LIMIT ?`
	args = append(args, limit)
	return q, args, true
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (Task, error) {
	var task Task
	var title, body, deadline, created, updated sql.NullString
	var completed int
	if err := s.Scan(&task.ID, &task.ItemType, &title, &body, &completed, &deadline, &created, &updated); err != nil {
		return Task{}, err
	}
	task.Title = strings.TrimSpace(title.String)
	task.Body = body.String
	if completed != 0 {
		task.Status = "done"
	} else {
		task.Status = "open"
	}
	task.DeadlineAt = strings.TrimSpace(deadline.String)
	task.CreatedOn = strings.TrimSpace(created.String)
	task.LastUpdated = strings.TrimSpace(updated.String)
	return task, nil
}
