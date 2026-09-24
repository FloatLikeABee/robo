package mcp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	defaultTaskLimit = 50
	maxTaskLimit     = 100
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
	path := strings.TrimSpace(os.Getenv("TRAN_SQLITE_PATH"))
	if path == "" {
		path = "./data/tran.sqlite"
	}
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

// OpenReadOnly opens TRAN_SQLITE_PATH without taking a write lock or running
// migrations. A file: URI is required: modernc treats "path?mode=ro" as a
// filename and would create the file. mode=ro plus _query_only=1 rejects
// writes even when the OS file is writable. journal_mode is left to the API.
func OpenReadOnly(path string) (*sql.DB, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("sqlite path is empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	dsn := (&url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(abs),
		RawQuery: "mode=ro&_query_only=1&_busy_timeout=5000",
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

// ListMyTasks returns the caller's user_note_todo rows. An email with no
// active Tran user yields an empty page, not another user's rows.
func ListMyTasks(ctx context.Context, db *sql.DB, email string, f TaskFilter) (ListResult, error) {
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
	userID, ok, err := tranUserID(ctx, db, email)
	if err != nil || !ok {
		return out, err
	}
	q, args := listQuery(userID, itemType, status, limit)
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

// GetMyTask returns one of the caller's rows. Any other id is ErrNotFound.
func GetMyTask(ctx context.Context, db *sql.DB, email string, id int) (Task, error) {
	if db == nil {
		return Task{}, errors.New("task store is not open")
	}
	if id <= 0 {
		return Task{}, errors.New("invalid id")
	}
	userID, ok, err := tranUserID(ctx, db, email)
	if err != nil {
		return Task{}, err
	}
	if !ok {
		return Task{}, ErrNotFound
	}
	row := db.QueryRowContext(ctx, `
		SELECT ID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
		FROM user_note_todo WHERE ID = ? AND UserID = ?`, id, userID)
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

// tranUserID matches the email branch of handlers.tranUserIDFromContext.
// ponytail: LIMIT 1 if two active User rows share an email; a unique index is the upgrade.
// It does not honor ?user_id=, X-User-ID, or the handler's default of user 1, and it does not insert.
func tranUserID(ctx context.Context, db *sql.DB, email string) (int, bool, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return 0, false, nil
	}
	var id int
	err := db.QueryRowContext(ctx, `
		SELECT UserID FROM "User"
		WHERE Email IS NOT NULL AND LOWER(TRIM(Email)) = LOWER(?) AND Deactivated = 0
		LIMIT 1`, email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	if id <= 0 {
		return 0, false, nil
	}
	return id, true, nil
}

func listQuery(userID int, itemType, status string, limit int) (string, []any) {
	// Order matches handlers.ListUserNotesTodos for the same type filter.
	order := `ItemType ASC, Completed ASC, (DeadlineAt IS NULL) ASC, DeadlineAt ASC, CreatedOn DESC`
	if itemType == "note" {
		order = `CreatedOn DESC`
	} else if itemType == "todo" {
		order = `Completed ASC, (DeadlineAt IS NULL) ASC, DeadlineAt ASC, CreatedOn DESC`
	}
	q := `SELECT ID, ItemType, Title, Body, Completed, DeadlineAt, CreatedOn, LastUpdated
		FROM user_note_todo WHERE UserID = ?`
	args := []any{userID}
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
	return q, args
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
