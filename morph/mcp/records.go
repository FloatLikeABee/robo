package mcp

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// ModuleRecord is one stick note, timeline, story, or research row.
type ModuleRecord struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Body    string `json:"body,omitempty"`
	StartAt string `json:"start_at,omitempty"`
	EndAt   string `json:"end_at,omitempty"`
	Status  string `json:"status,omitempty"`
}

// ModuleList is a page of module records.
type ModuleList struct {
	Records []ModuleRecord `json:"records"`
	Limit   int            `json:"limit"`
}

// CreateStickNote inserts a shared-board stick note. Empty times become the
// current UTC hour and the following hour.
func CreateStickNote(ctx context.Context, db *sql.DB, title, description, startAt, endAt string) (ModuleRecord, error) {
	if db == nil {
		return ModuleRecord{}, errors.New("note store is not open")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return ModuleRecord{}, errors.New("title is required")
	}
	if utf8.RuneCountInString(title) > maxNoteTitleLen {
		return ModuleRecord{}, errors.New("title is too long")
	}
	start, end, err := stickWindow(startAt, endAt)
	if err != nil {
		return ModuleRecord{}, err
	}
	stored := agentTitle(title)
	res, err := db.ExecContext(ctx, `
		INSERT INTO CaseTask (title, description, start_at, end_at)
		VALUES (?, ?, ?, ?)`, stored, strings.TrimSpace(description), start, end)
	if err != nil {
		return ModuleRecord{}, errors.New("could not store the stick note")
	}
	id, _ := res.LastInsertId()
	return GetStickNote(ctx, db, int(id))
}

func stickWindow(startAt, endAt string) (string, string, error) {
	startAt = strings.TrimSpace(startAt)
	endAt = strings.TrimSpace(endAt)
	if startAt == "" && endAt == "" {
		start := time.Now().UTC().Truncate(time.Hour)
		return start.Format(time.RFC3339), start.Add(time.Hour).Format(time.RFC3339), nil
	}
	start, err := time.Parse(time.RFC3339, startAt)
	if err != nil {
		return "", "", errors.New("invalid start_at")
	}
	if endAt == "" {
		return start.UTC().Format(time.RFC3339), start.UTC().Add(time.Hour).Format(time.RFC3339), nil
	}
	end, err := time.Parse(time.RFC3339, endAt)
	if err != nil {
		return "", "", errors.New("invalid end_at")
	}
	return start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339), nil
}

// ListStickNotes returns the shared stick-note board, newest id first.
func ListStickNotes(ctx context.Context, db *sql.DB, limit int) (ModuleList, error) {
	n, err := appliedLimit(limit)
	if err != nil {
		return ModuleList{}, err
	}
	if db == nil {
		return ModuleList{}, errors.New("note store is not open")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT ID, title, COALESCE(description, ''), COALESCE(start_at, ''), COALESCE(end_at, '')
		FROM CaseTask ORDER BY ID DESC LIMIT ?`, n)
	if err != nil {
		return ModuleList{}, errors.New("could not list stick notes")
	}
	defer rows.Close()
	out := ModuleList{Records: []ModuleRecord{}, Limit: n}
	for rows.Next() {
		var rec ModuleRecord
		if err := rows.Scan(&rec.ID, &rec.Title, &rec.Body, &rec.StartAt, &rec.EndAt); err != nil {
			return ModuleList{}, errors.New("could not list stick notes")
		}
		out.Records = append(out.Records, rec)
	}
	return out, rows.Err()
}

// GetStickNote reads one shared stick note. A missing id is not found.
func GetStickNote(ctx context.Context, db *sql.DB, id int) (ModuleRecord, error) {
	if id <= 0 {
		return ModuleRecord{}, errors.New("invalid id")
	}
	if db == nil {
		return ModuleRecord{}, errors.New("note store is not open")
	}
	var rec ModuleRecord
	err := db.QueryRowContext(ctx, `
		SELECT ID, title, COALESCE(description, ''), COALESCE(start_at, ''), COALESCE(end_at, '')
		FROM CaseTask WHERE ID = ?`, id).Scan(&rec.ID, &rec.Title, &rec.Body, &rec.StartAt, &rec.EndAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ModuleRecord{}, ErrNotFound
	}
	if err != nil {
		return ModuleRecord{}, errors.New("could not read the stick note")
	}
	return rec, nil
}

// CreateTimeline stores pasted text for the session user.
func CreateTimeline(ctx context.Context, db *sql.DB, tokenUserID, title, content string) (ModuleRecord, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" || content == "" {
		return ModuleRecord{}, errors.New("title and content are required")
	}
	return insertOwned(ctx, db, tokenUserID, `
		INSERT INTO timeline (user_id, title, source_summary, has_paste, markdown_content)
		SELECT owner.UserID, ?, ?, 1, ?
		FROM (`+ownerUserSelect()+`) AS owner`,
		agentTitle(title), clip(content, 240), content)
}

// CreateStory stores a story for the session user without running the generator.
func CreateStory(ctx context.Context, db *sql.DB, tokenUserID, title, idea, body string) (ModuleRecord, error) {
	title = strings.TrimSpace(title)
	idea = strings.TrimSpace(idea)
	body = strings.TrimSpace(body)
	if idea == "" {
		idea = body
	}
	if title == "" {
		title = idea
	}
	if title == "" || idea == "" {
		return ModuleRecord{}, errors.New("title or idea is required")
	}
	if body == "" {
		body = idea
	}
	return insertOwned(ctx, db, tokenUserID, `
		INSERT INTO big_note (user_id, title, idea, markdown_content, theme)
		SELECT owner.UserID, ?, ?, ?, 'dark'
		FROM (`+ownerUserSelect()+`) AS owner`,
		agentTitle(title), idea, body)
}

// CreateResearch stores a finished research row for the session user.
func CreateResearch(ctx context.Context, db *sql.DB, tokenUserID, prompt, body string) (ModuleRecord, error) {
	prompt = strings.TrimSpace(prompt)
	body = strings.TrimSpace(body)
	if prompt == "" {
		return ModuleRecord{}, errors.New("prompt is required")
	}
	if body == "" {
		body = prompt
	}
	title := prompt
	if i := strings.IndexAny(title, "\r\n"); i >= 0 {
		title = title[:i]
	}
	return insertOwned(ctx, db, tokenUserID, `
		INSERT INTO research (user_id, title, prompt, status, markdown_content, round_target)
		SELECT owner.UserID, ?, ?, 'complete', ?, 0
		FROM (`+ownerUserSelect()+`) AS owner`,
		agentTitle(title), prompt, body)
}

func insertOwned(ctx context.Context, db *sql.DB, tokenUserID, query string, args ...any) (ModuleRecord, error) {
	if db == nil {
		return ModuleRecord{}, errors.New("note store is not open")
	}
	if _, _, ok := ownerClause(tokenUserID); !ok {
		return ModuleRecord{}, errors.New("not a notes user for this session")
	}
	res, err := db.ExecContext(ctx, query, append(args, strings.TrimSpace(tokenUserID))...)
	if err != nil {
		return ModuleRecord{}, errors.New("could not store the record")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ModuleRecord{}, errors.New("not a notes user for this session")
	}
	id, _ := res.LastInsertId()
	return ModuleRecord{ID: int(id), Title: strArg(args, 0), Body: strArg(args, len(args)-1)}, nil
}

func strArg(args []any, i int) string {
	if i < 0 || i >= len(args) {
		return ""
	}
	s, _ := args[i].(string)
	return s
}

func clip(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n])
}

// ListOwned returns the session user's rows from timeline, big_note, or research.
func ListOwned(ctx context.Context, db *sql.DB, tokenUserID, table string, limit int) (ModuleList, error) {
	bodyCol, ok := ownedTable(table)
	if !ok {
		return ModuleList{}, errors.New("invalid module")
	}
	n, err := appliedLimit(limit)
	if err != nil {
		return ModuleList{}, err
	}
	if db == nil {
		return ModuleList{}, errors.New("note store is not open")
	}
	if _, _, ok := ownerClause(tokenUserID); !ok {
		return ModuleList{}, errors.New("not a notes user for this session")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, title, COALESCE(`+bodyCol+`, '')
		FROM `+table+`
		WHERE user_id = (`+ownerUserSelect()+`)
		ORDER BY id DESC LIMIT ?`, strings.TrimSpace(tokenUserID), n)
	if err != nil {
		return ModuleList{}, errors.New("could not list records")
	}
	defer rows.Close()
	out := ModuleList{Records: []ModuleRecord{}, Limit: n}
	for rows.Next() {
		var rec ModuleRecord
		if err := rows.Scan(&rec.ID, &rec.Title, &rec.Body); err != nil {
			return ModuleList{}, errors.New("could not list records")
		}
		out.Records = append(out.Records, rec)
	}
	return out, rows.Err()
}

// GetOwned reads one timeline, story, or research row for the session user.
func GetOwned(ctx context.Context, db *sql.DB, tokenUserID, table string, id int) (ModuleRecord, error) {
	bodyCol, ok := ownedTable(table)
	if !ok {
		return ModuleRecord{}, errors.New("invalid module")
	}
	if id <= 0 {
		return ModuleRecord{}, errors.New("invalid id")
	}
	if db == nil {
		return ModuleRecord{}, errors.New("note store is not open")
	}
	if _, _, ok := ownerClause(tokenUserID); !ok {
		return ModuleRecord{}, ErrNotFound
	}
	var rec ModuleRecord
	err := db.QueryRowContext(ctx, `
		SELECT id, title, COALESCE(`+bodyCol+`, '')
		FROM `+table+`
		WHERE id = ? AND user_id = (`+ownerUserSelect()+`)`, id, strings.TrimSpace(tokenUserID)).
		Scan(&rec.ID, &rec.Title, &rec.Body)
	if errors.Is(err, sql.ErrNoRows) {
		return ModuleRecord{}, ErrNotFound
	}
	if err != nil {
		return ModuleRecord{}, errors.New("could not read the record")
	}
	return rec, nil
}

func ownedTable(table string) (bodyCol string, ok bool) {
	switch table {
	case "timeline", "big_note", "research":
		return "markdown_content", true
	default:
		return "", false
	}
}
