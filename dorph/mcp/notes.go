package mcp

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"idongivaflyinfa/auth"
)

const (
	KindNote     = "note"
	KindTodo     = "todo"
	KindBigNote  = "big_note"
	KindTimeline = "timeline"
	KindResearch = "research"
	KindStick    = "stick"
	KindTask     = "task"

	maxTags    = 10
	maxTagLen  = 32
	visibility = "private"
)

var (
	ErrConflict       = errors.New("conflict")
	ErrForbiddenScope = errors.New("forbidden_scope")
	ErrInvalidInput   = errors.New("invalid_input")
	ftsToken          = regexp.MustCompile(`[A-Za-z0-9_]+`)
)

var noteKinds = []string{KindNote, KindTodo, KindBigNote, KindTimeline, KindResearch, KindStick, KindTask}

// NoteRecord is one owner-scoped note of any kind.
type NoteRecord struct {
	Kind       string   `json:"kind"`
	RecordID   string   `json:"record_id"`
	ID         int      `json:"id,omitempty"`
	Title      string   `json:"title"`
	Body       string   `json:"body,omitempty"`
	Status     string   `json:"status,omitempty"`
	Version    int      `json:"version"`
	Tags       []string `json:"tags,omitempty"`
	UpdatedAt  string   `json:"updated_at,omitempty"`
	Visibility string   `json:"visibility"`
	StartAt    string   `json:"start_at,omitempty"`
	EndAt      string   `json:"end_at,omitempty"`
}

// NoteList is a page of NoteRecord.
type NoteList struct {
	Notes []NoteRecord `json:"notes"`
	Limit int          `json:"limit"`
}

// CreateNoteArgs is one create_note call.
type CreateNoteArgs struct {
	Kind           string
	Title          string
	Body           string
	Tags           []string
	IdempotencyKey string
	StartAt        string
	EndAt          string
	Idea           string
	Prompt         string
}

func RequireScope(id Identity, want string) error {
	if auth.HasScope(id.Scopes, want) {
		return nil
	}
	return ErrForbiddenScope
}

func NormalizeKind(raw string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(raw))
	if k == "" {
		return KindNote, nil
	}
	for _, want := range noteKinds {
		if k == want {
			return k, nil
		}
	}
	return "", errors.New("invalid kind")
}

func RecordID(kind string, id int) string {
	return kind + ":" + strconv.Itoa(id)
}

func ParseRecordID(raw string) (string, int, error) {
	kind, rest, ok := strings.Cut(strings.TrimSpace(raw), ":")
	if !ok {
		return "", 0, ErrNotFound
	}
	kind, err := NormalizeKind(kind)
	if err != nil {
		return "", 0, err
	}
	id, err := strconv.Atoi(rest)
	if err != nil || id <= 0 {
		return "", 0, errors.New("invalid id")
	}
	return kind, id, nil
}

func liveSQL() string {
	return `(archived_at IS NULL OR archived_at = '')`
}

func CreateNote(ctx context.Context, db *sql.DB, owner string, in CreateNoteArgs) (NoteRecord, error) {
	if db == nil {
		return NoteRecord{}, errors.New("note store is not open")
	}
	owner = strings.TrimSpace(owner)
	kind, err := NormalizeKind(in.Kind)
	if err != nil {
		return NoteRecord{}, err
	}
	if key := strings.TrimSpace(in.IdempotencyKey); key != "" {
		if rec, ok := loadIdempotent(ctx, db, owner, key); ok {
			return rec, nil
		}
	}
	body := strings.TrimSpace(in.Body)
	tags := normalizeTags(in.Tags)
	var rec NoteRecord
	switch kind {
	case KindNote:
		task, e := CreateMyNote(ctx, db, owner, in.Title, body)
		if e != nil {
			return NoteRecord{}, e
		}
		rec = noteFromTask(task)
	case KindTodo:
		rec, err = insertNoteTodo(ctx, db, owner, KindTodo, in.Title, body)
	case KindTimeline:
		mod, e := CreateTimeline(ctx, db, owner, in.Title, firstNonEmpty(body, in.Title))
		if e != nil {
			return NoteRecord{}, e
		}
		rec = noteFromModule(KindTimeline, mod)
	case KindBigNote:
		mod, e := CreateStory(ctx, db, owner, in.Title, in.Idea, firstNonEmpty(body, in.Idea))
		if e != nil {
			return NoteRecord{}, e
		}
		rec = noteFromModule(KindBigNote, mod)
	case KindResearch:
		prompt := strings.TrimSpace(in.Prompt)
		if prompt == "" {
			prompt = firstNonEmpty(in.Title, body)
		}
		mod, e := CreateResearch(ctx, db, owner, prompt, body)
		if e != nil {
			return NoteRecord{}, e
		}
		rec = noteFromModule(KindResearch, mod)
	case KindStick, KindTask:
		mod, e := CreateBoardNote(ctx, db, owner, kind, in.Title, body, in.StartAt, in.EndAt)
		if e != nil {
			return NoteRecord{}, e
		}
		rec = noteFromModule(kind, mod)
	}
	if err != nil {
		return NoteRecord{}, err
	}
	if err := replaceTags(ctx, db, owner, rec.Kind, rec.ID, tags); err != nil {
		return NoteRecord{}, err
	}
	rec.Tags = tags
	if err := upsertFTS(ctx, db, owner, rec); err != nil {
		return NoteRecord{}, err
	}
	if key := strings.TrimSpace(in.IdempotencyKey); key != "" {
		_, _ = db.ExecContext(ctx, `
			INSERT OR IGNORE INTO note_idempotency (owner_user_id, idempotency_key, kind, source_id, created_at)
			VALUES (?, ?, ?, ?, ?)`, owner, key, rec.Kind, rec.ID, time.Now().UTC().Format(time.RFC3339))
	}
	return rec, nil
}

func GetNote(ctx context.Context, db *sql.DB, owner, recordID string) (NoteRecord, error) {
	kind, id, err := ParseRecordID(recordID)
	if err != nil {
		return NoteRecord{}, err
	}
	rec, err := loadNote(ctx, db, owner, kind, id)
	if err != nil {
		return NoteRecord{}, err
	}
	rec.Tags = loadTags(ctx, db, owner, kind, id)
	return rec, nil
}

func ListNotes(ctx context.Context, db *sql.DB, owner, kind string, limit int) (NoteList, error) {
	n, err := appliedLimit(limit)
	if err != nil {
		return NoteList{}, err
	}
	kinds := noteKinds
	if strings.TrimSpace(kind) != "" {
		k, err := NormalizeKind(kind)
		if err != nil {
			return NoteList{}, err
		}
		kinds = []string{k}
	}
	var all []NoteRecord
	for _, k := range kinds {
		page, err := listKind(ctx, db, owner, k, n)
		if err != nil {
			return NoteList{}, err
		}
		all = append(all, page...)
	}
	if len(all) > n {
		all = all[:n]
	}
	if all == nil {
		all = []NoteRecord{}
	}
	return NoteList{Notes: all, Limit: n}, nil
}

func UpdateNote(ctx context.Context, db *sql.DB, owner, recordID, title, body string, tags []string, expected int) (NoteRecord, error) {
	kind, id, err := ParseRecordID(recordID)
	if err != nil {
		return NoteRecord{}, err
	}
	cur, err := loadNote(ctx, db, owner, kind, id)
	if err != nil {
		return NoteRecord{}, err
	}
	if expected > 0 && cur.Version != expected {
		return NoteRecord{}, ErrConflict
	}
	if strings.TrimSpace(title) != "" {
		if utf8.RuneCountInString(strings.TrimSpace(title)) > maxNoteTitleLen {
			return NoteRecord{}, errors.New("title is too long")
		}
		cur.Title = agentTitle(strings.TrimSpace(title))
	}
	if body != "" {
		if utf8.RuneCountInString(body) > maxNoteBodyLen {
			return NoteRecord{}, errors.New("body is too long")
		}
		cur.Body = body
	}
	if err := applyNoteWrite(ctx, db, owner, cur, expected); err != nil {
		return NoteRecord{}, err
	}
	if tags != nil {
		norm := normalizeTags(tags)
		if err := replaceTags(ctx, db, owner, kind, id, norm); err != nil {
			return NoteRecord{}, err
		}
		cur.Tags = norm
	} else {
		cur.Tags = loadTags(ctx, db, owner, kind, id)
	}
	cur.Version++
	cur.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := upsertFTS(ctx, db, owner, cur); err != nil {
		return NoteRecord{}, err
	}
	return cur, nil
}

func AppendNote(ctx context.Context, db *sql.DB, owner, recordID, extra string, expected int) (NoteRecord, error) {
	extra = strings.TrimSpace(extra)
	if extra == "" {
		return NoteRecord{}, errors.New("body is required")
	}
	cur, err := GetNote(ctx, db, owner, recordID)
	if err != nil {
		return NoteRecord{}, err
	}
	body := cur.Body
	if body != "" {
		body += "\n\n"
	}
	return UpdateNote(ctx, db, owner, recordID, "", body+extra, nil, expected)
}

func ArchiveNote(ctx context.Context, db *sql.DB, owner, recordID string) error {
	kind, id, err := ParseRecordID(recordID)
	if err != nil {
		return err
	}
	if _, err := loadNote(ctx, db, owner, kind, id); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	var res sql.Result
	switch kind {
	case KindNote, KindTodo:
		clause, args, ok := ownerClause(owner)
		if !ok {
			return ErrNotFound
		}
		res, err = db.ExecContext(ctx, `UPDATE user_note_todo SET archived_at = ? WHERE ID = ? AND `+clause+` AND `+liveSQL(),
			append([]any{now, id}, args...)...)
	case KindBigNote, KindTimeline, KindResearch:
		table := kindTable(kind)
		res, err = db.ExecContext(ctx, `UPDATE `+table+` SET archived_at = ? WHERE id = ? AND user_id = (`+ownerUserSelect()+`) AND `+liveSQL(),
			now, id, strings.TrimSpace(owner))
	case KindStick, KindTask:
		res, err = db.ExecContext(ctx, `UPDATE CaseTask SET archived_at = ? WHERE ID = ? AND owner_user_id = ? AND board_kind = ? AND `+liveSQL(),
			now, id, strings.TrimSpace(owner), kind)
	}
	if err != nil {
		return errors.New("could not archive")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	_, _ = db.ExecContext(ctx, `DELETE FROM my_notes_fts WHERE owner_user_id = ? AND kind = ? AND source_id = ?`,
		strings.TrimSpace(owner), kind, id)
	return nil
}

func SearchMyNotes(ctx context.Context, db *sql.DB, owner, query string, limit int) (NoteList, error) {
	n, err := appliedLimit(limit)
	if err != nil {
		return NoteList{}, err
	}
	out := NoteList{Notes: []NoteRecord{}, Limit: n}
	match := sanitizeFTS(query)
	if match == "" || db == nil {
		return out, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT kind, source_id FROM my_notes_fts
		WHERE my_notes_fts MATCH ? AND owner_user_id = ?
		LIMIT ?`, match, strings.TrimSpace(owner), n)
	if err != nil {
		return out, nil
	}
	type hit struct {
		kind string
		id   int
	}
	var hits []hit
	for rows.Next() {
		var h hit
		if err := rows.Scan(&h.kind, &h.id); err != nil {
			_ = rows.Close()
			return NoteList{}, err
		}
		hits = append(hits, h)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return out, err
	}
	for _, h := range hits {
		rec, err := GetNote(ctx, db, owner, RecordID(h.kind, h.id))
		if err != nil {
			continue
		}
		out.Notes = append(out.Notes, rec)
	}
	return out, nil
}

func sanitizeFTS(q string) string {
	var toks []string
	for _, t := range ftsToken.FindAllString(q, -1) {
		switch strings.ToUpper(t) {
		case "AND", "OR", "NOT", "NEAR":
			continue
		}
		toks = append(toks, t)
	}
	return strings.Join(toks, " AND ")
}

func loadIdempotent(ctx context.Context, db *sql.DB, owner, key string) (NoteRecord, bool) {
	var kind string
	var id int
	err := db.QueryRowContext(ctx, `
		SELECT kind, source_id FROM note_idempotency
		WHERE owner_user_id = ? AND idempotency_key = ?`, owner, key).Scan(&kind, &id)
	if err != nil {
		return NoteRecord{}, false
	}
	rec, err := GetNote(ctx, db, owner, RecordID(kind, id))
	if err != nil {
		return NoteRecord{}, false
	}
	return rec, true
}

func insertNoteTodo(ctx context.Context, db *sql.DB, owner, itemType, title, body string) (NoteRecord, error) {
	if _, _, ok := ownerClause(owner); !ok {
		return NoteRecord{}, errors.New("not a notes user for this session")
	}
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" && body == "" {
		return NoteRecord{}, errors.New("title or body is required")
	}
	if utf8.RuneCountInString(title) > maxNoteTitleLen {
		return NoteRecord{}, errors.New("title is too long")
	}
	if utf8.RuneCountInString(body) > maxNoteBodyLen {
		return NoteRecord{}, errors.New("body is too long")
	}
	storedTitle := agentTitle(title)
	storedBody := agentBody(body)
	res, err := db.ExecContext(ctx, `
		INSERT INTO user_note_todo (UserID, ItemType, Title, Body, Completed, DeadlineAt, version)
		SELECT owner.UserID, ?, ?, ?, 0, NULL, 1
		FROM (`+ownerUserSelect()+`) AS owner`, itemType, storedTitle, storedBody, strings.TrimSpace(owner))
	if err != nil {
		return NoteRecord{}, errors.New("could not store the note")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return NoteRecord{}, errors.New("not a notes user for this session")
	}
	id, _ := res.LastInsertId()
	return loadNote(ctx, db, owner, itemType, int(id))
}

func loadNote(ctx context.Context, db *sql.DB, owner, kind string, id int) (NoteRecord, error) {
	if db == nil || id <= 0 {
		return NoteRecord{}, ErrNotFound
	}
	switch kind {
	case KindNote, KindTodo:
		return loadNoteTodo(ctx, db, owner, kind, id)
	case KindBigNote, KindTimeline, KindResearch:
		return loadOwnedNote(ctx, db, owner, kind, id)
	case KindStick, KindTask:
		return loadBoardNote(ctx, db, owner, kind, id)
	default:
		return NoteRecord{}, errors.New("invalid kind")
	}
}

func loadNoteTodo(ctx context.Context, db *sql.DB, owner, kind string, id int) (NoteRecord, error) {
	clause, args, ok := ownerClause(owner)
	if !ok {
		return NoteRecord{}, ErrNotFound
	}
	var rec NoteRecord
	var archived string
	err := db.QueryRowContext(ctx, `
		SELECT ID, ItemType, COALESCE(Title,''), COALESCE(Body,''), COALESCE(version,1), COALESCE(LastUpdated,''), COALESCE(archived_at,'')
		FROM user_note_todo WHERE ID = ? AND ItemType = ? AND `+clause+` AND `+liveSQL(),
		append([]any{id, kind}, args...)...).
		Scan(&rec.ID, &rec.Kind, &rec.Title, &rec.Body, &rec.Version, &rec.UpdatedAt, &archived)
	if errors.Is(err, sql.ErrNoRows) {
		return NoteRecord{}, ErrNotFound
	}
	if err != nil {
		return NoteRecord{}, err
	}
	return finishNote(rec), nil
}

func loadOwnedNote(ctx context.Context, db *sql.DB, owner, kind string, id int) (NoteRecord, error) {
	if _, _, ok := ownerClause(owner); !ok {
		return NoteRecord{}, ErrNotFound
	}
	table := kindTable(kind)
	var rec NoteRecord
	var archived string
	err := db.QueryRowContext(ctx, `
		SELECT id, COALESCE(title,''), COALESCE(markdown_content,''), COALESCE(version,1), COALESCE(last_updated,''), COALESCE(archived_at,'')
		FROM `+table+`
		WHERE id = ? AND user_id = (`+ownerUserSelect()+`) AND `+liveSQL(),
		id, strings.TrimSpace(owner)).
		Scan(&rec.ID, &rec.Title, &rec.Body, &rec.Version, &rec.UpdatedAt, &archived)
	if errors.Is(err, sql.ErrNoRows) {
		return NoteRecord{}, ErrNotFound
	}
	if err != nil {
		return NoteRecord{}, err
	}
	rec.Kind = kind
	return finishNote(rec), nil
}

func loadBoardNote(ctx context.Context, db *sql.DB, owner, kind string, id int) (NoteRecord, error) {
	var rec NoteRecord
	var archived string
	err := db.QueryRowContext(ctx, `
		SELECT ID, COALESCE(title,''), COALESCE(description,''), COALESCE(start_at,''), COALESCE(end_at,''),
			COALESCE(version,1), COALESCE(last_updated,''), COALESCE(archived_at,'')
		FROM CaseTask
		WHERE ID = ? AND owner_user_id = ? AND board_kind = ? AND `+liveSQL(),
		id, strings.TrimSpace(owner), kind).
		Scan(&rec.ID, &rec.Title, &rec.Body, &rec.StartAt, &rec.EndAt, &rec.Version, &rec.UpdatedAt, &archived)
	if errors.Is(err, sql.ErrNoRows) {
		return NoteRecord{}, ErrNotFound
	}
	if err != nil {
		return NoteRecord{}, err
	}
	rec.Kind = kind
	return finishNote(rec), nil
}

func listKind(ctx context.Context, db *sql.DB, owner, kind string, limit int) ([]NoteRecord, error) {
	switch kind {
	case KindNote, KindTodo:
		return listNoteTodo(ctx, db, owner, kind, limit)
	case KindBigNote, KindTimeline, KindResearch:
		return listOwnedKind(ctx, db, owner, kind, limit)
	case KindStick, KindTask:
		return listBoardKind(ctx, db, owner, kind, limit)
	default:
		return nil, errors.New("invalid kind")
	}
}

func listNoteTodo(ctx context.Context, db *sql.DB, owner, kind string, limit int) ([]NoteRecord, error) {
	clause, args, ok := ownerClause(owner)
	if !ok {
		return []NoteRecord{}, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT ID, ItemType, COALESCE(Title,''), COALESCE(Body,''), COALESCE(version,1), COALESCE(LastUpdated,'')
		FROM user_note_todo WHERE ItemType = ? AND `+clause+` AND `+liveSQL()+`
		ORDER BY ID DESC LIMIT ?`, append([]any{kind}, append(args, limit)...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []NoteRecord
	for rows.Next() {
		var rec NoteRecord
		if err := rows.Scan(&rec.ID, &rec.Kind, &rec.Title, &rec.Body, &rec.Version, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, finishNote(rec))
	}
	return out, rows.Err()
}

func listOwnedKind(ctx context.Context, db *sql.DB, owner, kind string, limit int) ([]NoteRecord, error) {
	if _, _, ok := ownerClause(owner); !ok {
		return []NoteRecord{}, nil
	}
	table := kindTable(kind)
	rows, err := db.QueryContext(ctx, `
		SELECT id, COALESCE(title,''), COALESCE(markdown_content,''), COALESCE(version,1), COALESCE(last_updated,'')
		FROM `+table+`
		WHERE user_id = (`+ownerUserSelect()+`) AND `+liveSQL()+`
		ORDER BY id DESC LIMIT ?`, strings.TrimSpace(owner), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []NoteRecord
	for rows.Next() {
		var rec NoteRecord
		if err := rows.Scan(&rec.ID, &rec.Title, &rec.Body, &rec.Version, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		rec.Kind = kind
		out = append(out, finishNote(rec))
	}
	return out, rows.Err()
}

func listBoardKind(ctx context.Context, db *sql.DB, owner, kind string, limit int) ([]NoteRecord, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT ID, COALESCE(title,''), COALESCE(description,''), COALESCE(start_at,''), COALESCE(end_at,''),
			COALESCE(version,1), COALESCE(last_updated,'')
		FROM CaseTask
		WHERE owner_user_id = ? AND board_kind = ? AND `+liveSQL()+`
		ORDER BY ID DESC LIMIT ?`, strings.TrimSpace(owner), kind, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []NoteRecord
	for rows.Next() {
		var rec NoteRecord
		if err := rows.Scan(&rec.ID, &rec.Title, &rec.Body, &rec.StartAt, &rec.EndAt, &rec.Version, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		rec.Kind = kind
		out = append(out, finishNote(rec))
	}
	return out, rows.Err()
}

func applyNoteWrite(ctx context.Context, db *sql.DB, owner string, rec NoteRecord, expected int) error {
	var (
		res sql.Result
		err error
	)
	switch rec.Kind {
	case KindNote, KindTodo:
		clause, args, ok := ownerClause(owner)
		if !ok {
			return ErrNotFound
		}
		q := `UPDATE user_note_todo SET Title = ?, Body = ?, version = version + 1, LastUpdated = CURRENT_TIMESTAMP
			WHERE ID = ? AND ` + clause + ` AND ` + liveSQL()
		args = append([]any{rec.Title, rec.Body, rec.ID}, args...)
		if expected > 0 {
			q += ` AND version = ?`
			args = append(args, expected)
		}
		res, err = db.ExecContext(ctx, q, args...)
	case KindBigNote, KindTimeline, KindResearch:
		table := kindTable(rec.Kind)
		q := `UPDATE ` + table + ` SET title = ?, markdown_content = ?, version = version + 1, last_updated = CURRENT_TIMESTAMP
			WHERE id = ? AND user_id = (` + ownerUserSelect() + `) AND ` + liveSQL()
		args := []any{rec.Title, rec.Body, rec.ID, strings.TrimSpace(owner)}
		if expected > 0 {
			q += ` AND version = ?`
			args = append(args, expected)
		}
		res, err = db.ExecContext(ctx, q, args...)
	case KindStick, KindTask:
		q := `UPDATE CaseTask SET title = ?, description = ?, version = version + 1, last_updated = CURRENT_TIMESTAMP
			WHERE ID = ? AND owner_user_id = ? AND board_kind = ? AND ` + liveSQL()
		args := []any{rec.Title, rec.Body, rec.ID, strings.TrimSpace(owner), rec.Kind}
		if expected > 0 {
			q += ` AND version = ?`
			args = append(args, expected)
		}
		res, err = db.ExecContext(ctx, q, args...)
	default:
		return errors.New("invalid kind")
	}
	if err != nil {
		return errors.New("could not update")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if expected > 0 {
			if _, look := loadNote(ctx, db, owner, rec.Kind, rec.ID); look == nil {
				return ErrConflict
			}
		}
		return ErrNotFound
	}
	return nil
}

func kindTable(kind string) string {
	switch kind {
	case KindBigNote:
		return "big_note"
	default:
		return kind
	}
}

func finishNote(rec NoteRecord) NoteRecord {
	rec.RecordID = RecordID(rec.Kind, rec.ID)
	rec.Visibility = visibility
	if rec.Version <= 0 {
		rec.Version = 1
	}
	if rec.Tags == nil {
		rec.Tags = []string{}
	}
	return rec
}

func noteFromTask(task Task) NoteRecord {
	return finishNote(NoteRecord{
		Kind:      KindNote,
		ID:        task.ID,
		Title:     task.Title,
		Body:      task.Body,
		Status:    task.Status,
		Version:   1,
		UpdatedAt: task.LastUpdated,
	})
}

func noteFromModule(kind string, rec ModuleRecord) NoteRecord {
	return finishNote(NoteRecord{
		Kind:    kind,
		ID:      rec.ID,
		Title:   rec.Title,
		Body:    rec.Body,
		StartAt: rec.StartAt,
		EndAt:   rec.EndAt,
		Version: 1,
	})
}

func normalizeTags(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range in {
		t := slugTag(raw)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
		if len(out) == maxTags {
			break
		}
	}
	return out
}

func slugTag(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if utf8.RuneCountInString(out) > maxTagLen {
		out = string([]rune(out)[:maxTagLen])
	}
	return out
}

func replaceTags(ctx context.Context, db *sql.DB, owner, kind string, id int, tags []string) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM note_tag WHERE owner_user_id = ? AND source_kind = ? AND source_id = ?`,
		owner, kind, id); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err := db.ExecContext(ctx, `INSERT INTO note_tag (owner_user_id, source_kind, source_id, tag) VALUES (?, ?, ?, ?)`,
			owner, kind, id, tag); err != nil {
			return err
		}
	}
	return nil
}

func loadTags(ctx context.Context, db *sql.DB, owner, kind string, id int) []string {
	rows, err := db.QueryContext(ctx, `SELECT tag FROM note_tag WHERE owner_user_id = ? AND source_kind = ? AND source_id = ? ORDER BY tag`,
		owner, kind, id)
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return out
		}
		out = append(out, tag)
	}
	if out == nil {
		return []string{}
	}
	return out
}

func upsertFTS(ctx context.Context, db *sql.DB, owner string, rec NoteRecord) error {
	_, _ = db.ExecContext(ctx, `DELETE FROM my_notes_fts WHERE owner_user_id = ? AND kind = ? AND source_id = ?`,
		owner, rec.Kind, rec.ID)
	_, err := db.ExecContext(ctx, `
		INSERT INTO my_notes_fts (title, body, tags, owner_user_id, kind, source_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		rec.Title, rec.Body, strings.Join(rec.Tags, " "), owner, rec.Kind, rec.ID)
	return err
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
