package mcp

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	shareLicense     = "CC BY 4.0"
	shareLicenseKey  = "CC-BY-4.0"
	reportHideAfter  = 3
	confirmTTL       = 5 * time.Minute
	defaultPoolTopic = "productivity"
)

var (
	ErrConsentRequired = errors.New("consent_required")
	ErrRateLimited     = errors.New("rate_limited")
	ErrShareBlocked    = errors.New("share_blocked")
)

// ShareRecord is one snapshot the owner published.
type ShareRecord struct {
	ID         string   `json:"share_id"`
	Kind       string   `json:"kind"`
	RecordID   string   `json:"record_id,omitempty"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary,omitempty"`
	Content    string   `json:"content,omitempty"`
	Topics     []string `json:"topics,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Visibility string   `json:"visibility"`
	License    string   `json:"license"`
	Alias      string   `json:"author_alias"`
	Status     string   `json:"status"`
	Published  string   `json:"published_at,omitempty"`
	RevokedAt  string   `json:"revoked_at,omitempty"`
	ViewCount  int      `json:"view_count,omitempty"`
}

// SharePreview is the consent snapshot plus findings.
type SharePreview struct {
	ConfirmToken string       `json:"confirm_token"`
	Title        string       `json:"title"`
	Content      string       `json:"content"`
	Findings     []Finding    `json:"findings"`
	Alias        string       `json:"author_alias"`
	Visibility   string       `json:"visibility"`
	Topics       []string     `json:"topics,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
	License      string       `json:"license"`
	Kind         string       `json:"kind"`
	RecordID     string       `json:"record_id"`
}

// PoolHit is one index row. No owner id or email.
type PoolHit struct {
	ID        string   `json:"share_id"`
	Title     string   `json:"title"`
	Summary   string   `json:"summary,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Topics    []string `json:"topics,omitempty"`
	Kind      string   `json:"kind"`
	Alias     string   `json:"author_alias"`
	License   string   `json:"license"`
	Published string   `json:"published_at"`
}

// PoolGet is a reader view of a live share.
type PoolGet struct {
	ID          string `json:"share_id"`
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Alias       string `json:"author_alias"`
	License     string `json:"license"`
	Attribution string `json:"attribution"`
	Content     string `json:"content"`
}

// PoolTopic is a curated slug plus live count.
type PoolTopic struct {
	Slug  string `json:"slug"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type pendingShare struct {
	Owner      string
	Kind       string
	SourceID   int
	RecordID   string
	Title      string
	Content    string
	Summary    string
	Visibility string
	Topics     []string
	Tags       []string
	Alias      string
	Findings   []Finding
	Expires    time.Time
}

var (
	pendingMu sync.Mutex
	pending   = map[string]pendingShare{}
)

func PreviewShare(ctx context.Context, db *sql.DB, owner, recordID, visibility, attribution string, topics, tags []string) (SharePreview, error) {
	if db == nil {
		return SharePreview{}, errors.New("note store is not open")
	}
	rec, err := GetNote(ctx, db, owner, recordID)
	if err != nil {
		return SharePreview{}, err
	}
	vis := normalizeVisibility(visibility)
	title := stripProvenance(rec.Title)
	body := stripProvenance(rec.Body)
	findings := ScanShareText(title + "\n" + body)
	alias, err := resolveAlias(ctx, db, owner, attribution)
	if err != nil {
		return SharePreview{}, err
	}
	tops := normalizeTopics(ctx, db, topics)
	tgs := normalizeTags(tags)
	tok := newShareID()
	pendingMu.Lock()
	pending[tok] = pendingShare{
		Owner: owner, Kind: rec.Kind, SourceID: rec.ID, RecordID: rec.RecordID,
		Title: title, Content: body, Summary: firstSentence(body, title),
		Visibility: vis, Topics: tops, Tags: tgs, Alias: alias,
		Findings: findings, Expires: time.Now().Add(confirmTTL),
	}
	pendingMu.Unlock()
	return SharePreview{
		ConfirmToken: tok, Title: title, Content: body, Findings: findings,
		Alias: alias, Visibility: vis, Topics: tops, Tags: tgs,
		License: shareLicense, Kind: rec.Kind, RecordID: rec.RecordID,
	}, nil
}

func ShareNote(ctx context.Context, db *sql.DB, owner, token string) (ShareRecord, error) {
	if db == nil {
		return ShareRecord{}, errors.New("note store is not open")
	}
	pendingMu.Lock()
	p, ok := pending[token]
	if ok && (p.Expires.Before(time.Now()) || p.Owner != owner) {
		delete(pending, token)
		ok = false
	}
	if ok {
		delete(pending, token)
	}
	pendingMu.Unlock()
	if !ok {
		return ShareRecord{}, ErrConsentRequired
	}
	if shareBlocked(p.Findings) {
		labels := make([]string, 0, len(p.Findings))
		for _, f := range p.Findings {
			labels = append(labels, f.Kind+":"+f.Label)
		}
		return ShareRecord{}, fmt.Errorf("%w: %s", ErrShareBlocked, strings.Join(labels, ", "))
	}
	if !AllowRate(owner, "share_note") {
		return ShareRecord{}, ErrRateLimited
	}
	id := newShareID()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.ExecContext(ctx, `
		INSERT INTO pool_share (id, owner_user_id, source_kind, source_id, visibility, title, summary, content, license, author_alias, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'live', ?)`,
		id, owner, p.Kind, p.SourceID, p.Visibility, p.Title, p.Summary, p.Content, shareLicenseKey, p.Alias, now)
	if err != nil {
		return ShareRecord{}, err
	}
	if err := replaceShareMeta(ctx, db, id, p.Topics, p.Tags); err != nil {
		return ShareRecord{}, err
	}
	upsertPoolFTS(ctx, db, id, p.Title, p.Summary, p.Content, strings.Join(p.Tags, " "))
	return ShareRecord{
		ID: id, Kind: p.Kind, RecordID: p.RecordID, Title: p.Title, Summary: p.Summary,
		Content: p.Content, Topics: p.Topics, Tags: p.Tags, Visibility: p.Visibility,
		License: shareLicense, Alias: p.Alias, Status: "live", Published: now,
	}, nil
}

func UnshareNote(ctx context.Context, db *sql.DB, owner, shareID string) (ShareRecord, error) {
	if db == nil {
		return ShareRecord{}, errors.New("note store is not open")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := db.ExecContext(ctx, `
		UPDATE pool_share SET status = 'revoked', revoked_at = ?
		WHERE id = ? AND owner_user_id = ? AND status = 'live'`, now, shareID, owner)
	if err != nil {
		return ShareRecord{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ShareRecord{}, ErrNotFound
	}
	_, _ = db.ExecContext(ctx, `DELETE FROM pool_fts WHERE share_id = ?`, shareID)
	return ShareRecord{ID: shareID, Status: "revoked", RevokedAt: now}, nil
}

func ListMyShares(ctx context.Context, db *sql.DB, owner string) ([]ShareRecord, error) {
	if db == nil {
		return nil, errors.New("note store is not open")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, source_kind, source_id, visibility, title, summary, license, author_alias, status, created_at,
			COALESCE(revoked_at, ''), view_count
		FROM pool_share WHERE owner_user_id = ? ORDER BY created_at DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ShareRecord
	for rows.Next() {
		var rec ShareRecord
		var sourceID int
		var licenseKey string
		if err := rows.Scan(&rec.ID, &rec.Kind, &sourceID, &rec.Visibility, &rec.Title, &rec.Summary,
			&licenseKey, &rec.Alias, &rec.Status, &rec.Published, &rec.RevokedAt, &rec.ViewCount); err != nil {
			return nil, err
		}
		rec.RecordID = RecordID(rec.Kind, sourceID)
		rec.License = displayLicense(licenseKey)
		out = append(out, rec)
	}
	return out, rows.Err()
}

func ListPoolTopics(ctx context.Context, db *sql.DB) ([]PoolTopic, error) {
	if db == nil {
		return nil, errors.New("note store is not open")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT t.slug, t.label, COUNT(st.share_id)
		FROM pool_topic t
		LEFT JOIN pool_share_topic st ON st.topic = t.slug
		LEFT JOIN pool_share s ON s.id = st.share_id AND s.status = 'live' AND s.visibility = 'pool'
		GROUP BY t.slug, t.label ORDER BY t.slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PoolTopic
	for rows.Next() {
		var t PoolTopic
		if err := rows.Scan(&t.Slug, &t.Label, &t.Count); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func SearchPool(ctx context.Context, db *sql.DB, user, query, kind, topic string, limit int) ([]PoolHit, error) {
	if db == nil {
		return nil, errors.New("note store is not open")
	}
	if !AllowRate(user, "pool_search") {
		return nil, ErrRateLimited
	}
	if limit <= 0 || limit > 25 {
		limit = 25
	}
	kind = strings.TrimSpace(kind)
	topic = slugTag(topic)
	q := strings.Join(ftsToken.FindAllString(query, -1), " AND ")
	var ids []string
	if q == "" {
		rows, err := db.QueryContext(ctx, `
			SELECT s.id FROM pool_share s
			WHERE s.status = 'live' AND s.visibility = 'pool'
			  AND (? = '' OR s.source_kind = ?)
			  AND (? = '' OR EXISTS (SELECT 1 FROM pool_share_topic t WHERE t.share_id = s.id AND t.topic = ?))
			ORDER BY s.created_at DESC LIMIT ?`, kind, kind, topic, topic, limit)
		if err != nil {
			return nil, err
		}
		ids, err = collectIDs(rows)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err := db.QueryContext(ctx, `
			SELECT f.share_id FROM pool_fts f
			JOIN pool_share s ON s.id = f.share_id
			WHERE pool_fts MATCH ? AND s.status = 'live' AND s.visibility = 'pool'
			  AND (? = '' OR s.source_kind = ?)
			  AND (? = '' OR EXISTS (SELECT 1 FROM pool_share_topic t WHERE t.share_id = s.id AND t.topic = ?))
			LIMIT ?`, q, kind, kind, topic, topic, limit)
		if err != nil {
			return nil, err
		}
		ids, err = collectIDs(rows)
		if err != nil {
			return nil, err
		}
	}
	var out []PoolHit
	for _, id := range ids {
		hit, err := loadPoolHit(ctx, db, id)
		if err != nil {
			continue
		}
		out = append(out, hit)
	}
	return out, nil
}

func GetPoolShare(ctx context.Context, db *sql.DB, user, shareID string) (PoolGet, error) {
	if db == nil {
		return PoolGet{}, errors.New("note store is not open")
	}
	if !AllowRate(user, "pool_get") {
		return PoolGet{}, ErrRateLimited
	}
	var rec ShareRecord
	var sourceID int
	var licenseKey, owner string
	err := db.QueryRowContext(ctx, `
		SELECT id, owner_user_id, source_kind, source_id, title, content, license, author_alias, status
		FROM pool_share WHERE id = ?`, shareID).Scan(
		&rec.ID, &owner, &rec.Kind, &sourceID, &rec.Title, &rec.Content, &licenseKey, &rec.Alias, &rec.Status)
	if err != nil {
		return PoolGet{}, ErrNotFound
	}
	if rec.Status != "live" {
		return PoolGet{}, ErrNotFound
	}
	_, _ = db.ExecContext(ctx, `UPDATE pool_share SET view_count = view_count + 1 WHERE id = ?`, shareID)
	attr := AttributionLine(rec.Alias, rec.ID)
	return PoolGet{
		ID: rec.ID, Kind: rec.Kind, Title: rec.Title, Alias: rec.Alias,
		License: displayLicense(licenseKey), Attribution: attr,
		Content: attr + "\n\n<shared_content untrusted=\"true\">\n" + rec.Content + "\n</shared_content>",
	}, nil
}

func SavePoolCopy(ctx context.Context, db *sql.DB, user, shareID, kind string) (NoteRecord, error) {
	got, err := GetPoolShare(ctx, db, user, shareID)
	if err != nil {
		return NoteRecord{}, err
	}
	if kind == "" {
		kind = KindNote
	}
	body := got.Attribution + "\n\nsource share: " + got.ID + "\n\n" + stripEnvelope(got.Content)
	return CreateNote(ctx, db, user, CreateNoteArgs{Kind: kind, Title: got.Title, Body: body})
}

func ReportPoolShare(ctx context.Context, db *sql.DB, user, shareID, reason string) (string, error) {
	if db == nil {
		return "", errors.New("note store is not open")
	}
	if !AllowRate(user, "pool_report") {
		return "", ErrRateLimited
	}
	reason = slugTag(reason)
	if reason == "" {
		reason = "other"
	}
	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM pool_share WHERE id = ?`, shareID).Scan(&status); err != nil {
		return "", ErrNotFound
	}
	if status != "live" {
		return "", ErrNotFound
	}
	id := newShareID()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.ExecContext(ctx, `
		INSERT OR IGNORE INTO pool_report (id, share_id, reporter_user_id, reason, created_at)
		VALUES (?, ?, ?, ?, ?)`, id, shareID, user, reason, now); err != nil {
		return "", err
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pool_report WHERE share_id = ?`, shareID).Scan(&n); err != nil {
		return "", err
	}
	_, _ = db.ExecContext(ctx, `UPDATE pool_share SET report_count = ? WHERE id = ?`, n, shareID)
	if n >= reportHideAfter {
		_, _ = db.ExecContext(ctx, `UPDATE pool_share SET status = 'hidden' WHERE id = ?`, shareID)
		_, _ = db.ExecContext(ctx, `DELETE FROM pool_fts WHERE share_id = ?`, shareID)
	}
	return id, nil
}

// AttributionLine is the public citation. Never includes email.
func AttributionLine(alias, shareID string) string {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		alias = "anonymous"
	}
	return "Shared by @" + alias + " · License " + shareLicense + " · dorph share " + shareID
}

func resolveAlias(ctx context.Context, db *sql.DB, owner, attribution string) (string, error) {
	var username, email, alias string
	err := db.QueryRowContext(ctx, `
		SELECT username, email, COALESCE(pool_alias, '') FROM plat_users WHERE id = ?`, owner).Scan(&username, &email, &alias)
	if err != nil {
		return "anonymous", nil
	}
	_ = email
	if strings.EqualFold(strings.TrimSpace(attribution), "anonymous") {
		return "anonymous", nil
	}
	if alias != "" {
		return alias, nil
	}
	alias = slugTag(username)
	if alias == "" || strings.Contains(username, "@") {
		alias = "anonymous"
	}
	if alias != "anonymous" {
		_, _ = db.ExecContext(ctx, `UPDATE plat_users SET pool_alias = ? WHERE id = ? AND (pool_alias IS NULL OR pool_alias = '')`, alias, owner)
	}
	return alias, nil
}

func normalizeVisibility(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "link") {
		return "link"
	}
	return "pool"
}

func normalizeTopics(ctx context.Context, db *sql.DB, in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range in {
		s := slugTag(raw)
		if s == "" || seen[s] {
			continue
		}
		var ok string
		if err := db.QueryRowContext(ctx, `SELECT slug FROM pool_topic WHERE slug = ?`, s).Scan(&ok); err != nil {
			continue
		}
		seen[s] = true
		out = append(out, s)
		if len(out) == 5 {
			break
		}
	}
	if len(out) == 0 {
		return []string{defaultPoolTopic}
	}
	return out
}

func replaceShareMeta(ctx context.Context, db *sql.DB, id string, topics, tags []string) error {
	_, _ = db.ExecContext(ctx, `DELETE FROM pool_share_topic WHERE share_id = ?`, id)
	_, _ = db.ExecContext(ctx, `DELETE FROM pool_share_tag WHERE share_id = ?`, id)
	for _, t := range topics {
		if _, err := db.ExecContext(ctx, `INSERT INTO pool_share_topic (share_id, topic) VALUES (?, ?)`, id, t); err != nil {
			return err
		}
	}
	for _, t := range tags {
		if _, err := db.ExecContext(ctx, `INSERT INTO pool_share_tag (share_id, tag) VALUES (?, ?)`, id, t); err != nil {
			return err
		}
	}
	return nil
}

func upsertPoolFTS(ctx context.Context, db *sql.DB, id, title, summary, content, tags string) {
	_, _ = db.ExecContext(ctx, `DELETE FROM pool_fts WHERE share_id = ?`, id)
	_, _ = db.ExecContext(ctx, `INSERT INTO pool_fts (title, summary, content, tags, share_id) VALUES (?, ?, ?, ?, ?)`,
		title, summary, content, tags, id)
}

func loadPoolHit(ctx context.Context, db *sql.DB, id string) (PoolHit, error) {
	var hit PoolHit
	var licenseKey string
	err := db.QueryRowContext(ctx, `
		SELECT id, source_kind, title, summary, license, author_alias, created_at
		FROM pool_share WHERE id = ? AND status = 'live' AND visibility = 'pool'`, id).Scan(
		&hit.ID, &hit.Kind, &hit.Title, &hit.Summary, &licenseKey, &hit.Alias, &hit.Published)
	if err != nil {
		return PoolHit{}, err
	}
	hit.License = displayLicense(licenseKey)
	hit.Tags = loadShareTags(ctx, db, id)
	hit.Topics = loadShareTopics(ctx, db, id)
	return hit, nil
}

func loadShareTags(ctx context.Context, db *sql.DB, id string) []string {
	rows, err := db.QueryContext(ctx, `SELECT tag FROM pool_share_tag WHERE share_id = ? ORDER BY tag`, id)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if rows.Scan(&t) == nil {
			out = append(out, t)
		}
	}
	return out
}

func loadShareTopics(ctx context.Context, db *sql.DB, id string) []string {
	rows, err := db.QueryContext(ctx, `SELECT topic FROM pool_share_topic WHERE share_id = ? ORDER BY topic`, id)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if rows.Scan(&t) == nil {
			out = append(out, t)
		}
	}
	return out
}

func collectIDs(rows *sql.Rows) ([]string, error) {
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func displayLicense(key string) string {
	if key == shareLicenseKey || key == "" {
		return shareLicense
	}
	return key
}

func firstSentence(body, title string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return title
	}
	if i := strings.IndexAny(body, ".\n"); i > 0 && i < 160 {
		return strings.TrimSpace(body[:i])
	}
	if len(body) > 160 {
		return strings.TrimSpace(body[:160])
	}
	return body
}

func newShareID() string {
	var raw [8]byte
	_, _ = rand.Read(raw[:])
	return hex.EncodeToString(raw[:])
}

// ResetPoolForTest clears in-memory confirm tokens and rate counters.
func ResetPoolForTest() {
	resetRates()
	pendingMu.Lock()
	pending = map[string]pendingShare{}
	pendingMu.Unlock()
}

// ExpireConfirmTokenForTest marks a preview token expired.
func ExpireConfirmTokenForTest(token string) {
	pendingMu.Lock()
	defer pendingMu.Unlock()
	p, ok := pending[token]
	if !ok {
		return
	}
	p.Expires = time.Now().Add(-time.Second)
	pending[token] = p
}

func stripEnvelope(s string) string {
	s = strings.TrimSpace(s)
	const open = `<shared_content untrusted="true">`
	const close = `</shared_content>`
	if i := strings.Index(s, open); i >= 0 {
		s = s[i+len(open):]
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), close)
	return strings.TrimSpace(s)
}
