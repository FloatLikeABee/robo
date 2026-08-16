package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"idongivaflyinfa/ai"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var bigNoteMarkdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithXHTML()),
)

type bigNote struct {
	ID              int             `json:"id"`
	UserID          int             `json:"user_id"`
	OwnerKey        string          `json:"owner_key"`
	Title           string          `json:"title"`
	Idea            string          `json:"idea"`
	NoteKind        string          `json:"note_kind"`
	MarkdownContent string          `json:"markdown_content"`
	HTMLContent     string          `json:"html_content"`
	Questions       json.RawMessage `json:"questions,omitempty"`
	Theme           string          `json:"theme"`
	PublishedSlug   *string         `json:"published_slug,omitempty"`
	PublishedPath   *string         `json:"published_path,omitempty"`
	PublishedURL    string          `json:"published_url,omitempty"`
	CreatedOn       time.Time       `json:"created_on"`
	LastUpdated     time.Time       `json:"last_updated"`
}

type bigNoteQuestion struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Type    string   `json:"type"` // text | textarea | single | multi
	Options []string `json:"options,omitempty"`
}

type bigNoteAIResult struct {
	Title     string            `json:"title"`
	Markdown  string            `json:"markdown"`
	HTML      string            `json:"html"`
	Questions []bigNoteQuestion `json:"questions"`
}

type bigNoteResponse struct {
	ID                int       `json:"id"`
	BigNoteID         int       `json:"big_note_id"`
	AnswersJSON       string    `json:"answers_json"`
	Answers           any       `json:"answers,omitempty"`
	AnalysisMarkdown  *string   `json:"analysis_markdown,omitempty"`
	CreatedOn         time.Time `json:"created_on"`
	LastUpdated       time.Time `json:"last_updated"`
}

func (h *Handlers) bigNoteOwnerKey(c *gin.Context) string {
	if v, ok := c.Get("auth_email"); ok {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(strings.ToLower(s))
			if s != "" {
				return s
			}
		}
	}
	if v, ok := c.Get("auth_user_id"); ok {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				return "uid:" + s
			}
		}
	}
	return fmt.Sprintf("tran:%d", h.tranUserIDFromContext(c))
}

func (h *Handlers) bigNotePublicURL(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Morph-hosted public notes: keep relative so the browser uses the current Morph origin
	// (frontend :3031 proxy or API :9090). Do NOT use EXTERNAL_API_BASE (document readers on :8000).
	if strings.HasPrefix(path, "/api/tran/public/big-notes/") {
		return path
	}
	base := strings.TrimSpace(h.tranMailBase)
	if base == "" {
		return path
	}
	return strings.TrimSuffix(base, "/") + path
}

var bigNoteSlugCleaner = regexp.MustCompile(`[^a-z0-9]+`)

func slugifyBigNote(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = bigNoteSlugCleaner.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "big-note"
	}
	if utf8.RuneCountInString(s) > 60 {
		runes := []rune(s)
		s = string(runes[:60])
		s = strings.Trim(s, "-")
	}
	return s
}

func (h *Handlers) uniqueBigNoteSlug(base string) (string, error) {
	base = slugifyBigNote(base)
	candidate := base
	for i := 0; i < 50; i++ {
		var existing int
		err := h.TranMySQL.DB.QueryRow(`SELECT id FROM big_note WHERE published_slug = ? LIMIT 1`, candidate).Scan(&existing)
		if errors.Is(err, sql.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		candidate = fmt.Sprintf("%s-%d", base, i+2)
	}
	return fmt.Sprintf("%s-%d", base, time.Now().Unix()%100000), nil
}

func (h *Handlers) attachBigNoteURL(n *bigNote) {
	if n == nil || n.PublishedPath == nil {
		return
	}
	n.PublishedURL = h.bigNotePublicURL(*n.PublishedPath)
}

func extractJSONObjectAware(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	if start < 0 {
		return "", false
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], true
			}
		}
	}
	return "", false
}

func scanBigNote(scanner interface {
	Scan(dest ...any) error
}) (bigNote, error) {
	var n bigNote
	var pubSlug, pubPath, questions sql.NullString
	err := scanner.Scan(
		&n.ID, &n.UserID, &n.OwnerKey, &n.Title, &n.Idea, &n.NoteKind,
		&n.MarkdownContent, &n.HTMLContent, &questions, &n.Theme,
		&pubSlug, &pubPath, &n.CreatedOn, &n.LastUpdated,
	)
	if err != nil {
		return n, err
	}
	if questions.Valid && strings.TrimSpace(questions.String) != "" {
		n.Questions = json.RawMessage(questions.String)
	}
	if pubSlug.Valid {
		s := pubSlug.String
		n.PublishedSlug = &s
	}
	if pubPath.Valid {
		p := pubPath.String
		n.PublishedPath = &p
	}
	if n.NoteKind == "" {
		n.NoteKind = "note"
	}
	return n, nil
}

const bigNoteSelectCols = `id, user_id, owner_key, title, idea, note_kind, markdown_content, html_content,
	questions_json, theme, published_slug, published_path, created_on, last_updated`

func (h *Handlers) getBigNoteOwned(c *gin.Context, id int) (bigNote, error) {
	owner := h.bigNoteOwnerKey(c)
	userID := h.tranUserIDFromContext(c)
	row := h.TranMySQL.DB.QueryRow(
		`SELECT `+bigNoteSelectCols+` FROM big_note
		 WHERE id = ? AND (owner_key = ? OR (owner_key = '' AND user_id = ?) OR user_id = ?)
		 LIMIT 1`,
		id, owner, userID, userID,
	)
	return scanBigNote(row)
}

func normalizeNoteKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "questionnaire", "survey", "form", "quiz":
		return "questionnaire"
	default:
		return "note"
	}
}

func (h *Handlers) generateBigNoteContent(ctx context.Context, idea, theme, kind, prompt, currentMD, currentHTML string) (bigNoteAIResult, error) {
	var out bigNoteAIResult
	if h.aiService == nil {
		return out, errors.New("AI service not configured")
	}
	idea = strings.TrimSpace(idea)
	theme = strings.TrimSpace(theme)
	if theme == "" {
		theme = "dark"
	}
	kind = normalizeNoteKind(kind)
	prompt = strings.TrimSpace(prompt)

	var b strings.Builder
	b.WriteString(`You create MorphData Big Notes.
Respond with ONLY one compact JSON object (no markdown fences) with keys:
- "title": short title (max ~80 chars)
- "markdown": complete but concise markdown document (aim under 800 words)
- "questions": array

Do NOT include an "html" key — HTML is generated by the server.

Each question object:
{"id":"q1","label":"...","type":"text|textarea|single|multi","options":["a","b"]}
`)
	if kind == "questionnaire" {
		b.WriteString(`
Questionnaire rules:
- Produce 4–8 clear questions covering the idea.
- Prefer single/multi with options when useful; use textarea for open feedback.
- Markdown should briefly introduce the questionnaire (no need to re-list every question in detail).
- questions must be non-empty.
`)
	} else {
		b.WriteString(`
Note rules:
- Long-form readable note (article style), concise.
- questions must be [].
`)
	}
	b.WriteString("\nTheme preference: ")
	b.WriteString(theme)
	b.WriteString("\nKind: ")
	b.WriteString(kind)
	b.WriteString("\n\nOriginal idea:\n")
	b.WriteString(idea)
	if prompt != "" {
		b.WriteString("\n\nRegeneration / revision prompt:\n")
		b.WriteString(prompt)
	}
	if strings.TrimSpace(currentMD) != "" {
		b.WriteString("\n\nCurrent markdown (may revise):\n")
		b.WriteString(truncateRunes(currentMD, 8000))
	}
	_ = currentHTML // HTML is rebuilt server-side from markdown/questions.

	genCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	raw, err := h.aiService.ChatCompletionLong(genCtx, []ai.DashScopeMessage{{Role: "user", Content: b.String()}})
	if err != nil {
		return out, err
	}
	jsonStr, ok := extractJSONObjectAware(raw)
	if !ok {
		jsonStr, ok = extractJSONObject(raw)
	}
	if !ok {
		return out, errors.New("AI did not return JSON for the note")
	}
	if err := json.Unmarshal([]byte(jsonStr), &out); err != nil {
		return out, fmt.Errorf("parse AI note JSON: %w", err)
	}
	out.Title = strings.TrimSpace(out.Title)
	out.Markdown = strings.TrimSpace(out.Markdown)
	out.HTML = strings.TrimSpace(out.HTML)
	if out.Title == "" {
		out.Title = "Untitled note"
	}
	if out.Markdown == "" {
		out.Markdown = "# " + out.Title + "\n\n" + idea
	}
	if kind == "questionnaire" && len(out.Questions) == 0 {
		out.Questions = []bigNoteQuestion{{
			ID:    "q1",
			Label: "Your response",
			Type:  "textarea",
		}}
	}
	if kind != "questionnaire" {
		out.Questions = nil
	}
	for i := range out.Questions {
		q := &out.Questions[i]
		q.ID = strings.TrimSpace(q.ID)
		q.Label = strings.TrimSpace(q.Label)
		q.Type = strings.ToLower(strings.TrimSpace(q.Type))
		if q.ID == "" {
			q.ID = fmt.Sprintf("q%d", i+1)
		}
		if q.Label == "" {
			q.Label = q.ID
		}
		switch q.Type {
		case "text", "textarea", "single", "multi":
		default:
			q.Type = "text"
		}
	}
	out.HTML = buildBigNoteHTML(out.Title, out.Markdown, kind, out.Questions, theme)
	if utf8.RuneCountInString(out.Title) > 255 {
		out.Title = string([]rune(out.Title)[:255])
	}
	return out, nil
}

func markdownToHTMLFragment(markdown string) string {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := bigNoteMarkdown.Convert([]byte(markdown), &buf); err != nil {
		return "<p>" + htmlEscapeMinimal(markdown) + "</p>"
	}
	return buf.String()
}

// stripLeadingTitleHeading removes a leading ATX/setext H1 that duplicates the page title.
func stripLeadingTitleHeading(markdown, title string) string {
	md := strings.TrimLeft(markdown, " \t\r\n")
	title = strings.TrimSpace(title)
	if md == "" || title == "" {
		return markdown
	}
	lines := strings.Split(md, "\n")
	first := strings.TrimSpace(lines[0])
	heading := ""
	restStart := 1
	if strings.HasPrefix(first, "# ") && !strings.HasPrefix(first, "##") {
		heading = strings.TrimSpace(strings.TrimPrefix(first, "#"))
	} else if len(lines) >= 2 {
		underline := strings.TrimSpace(lines[1])
		if underline != "" && strings.Trim(underline, "=") == "" {
			heading = first
			restStart = 2
		}
	}
	if heading == "" || !strings.EqualFold(heading, title) {
		return markdown
	}
	return strings.TrimLeft(strings.Join(lines[restStart:], "\n"), " \t\r\n")
}

func questionsFromRaw(raw json.RawMessage) []bigNoteQuestion {
	if len(raw) == 0 {
		return nil
	}
	var qs []bigNoteQuestion
	if err := json.Unmarshal(raw, &qs); err != nil {
		return nil
	}
	return qs
}

func refreshBigNoteHTML(n *bigNote) {
	if n == nil {
		return
	}
	n.HTMLContent = buildBigNoteHTML(n.Title, n.MarkdownContent, n.NoteKind, questionsFromRaw(n.Questions), n.Theme)
}

func buildBigNoteHTML(title, markdown, kind string, questions []bigNoteQuestion, theme string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Big note"
	}
	theme = strings.TrimSpace(theme)
	if theme == "" || strings.EqualFold(theme, "default") {
		theme = "dark"
	}
	dark := strings.Contains(strings.ToLower(theme), "dark")
	proseCSS := `.prose{font-size:1.05rem;color:var(--ink)}
.prose>:first-child{margin-top:0}
.prose h1,.prose h2,.prose h3,.prose h4{line-height:1.25;margin:1.35em 0 .55em;font-weight:700}
.prose h1{font-size:1.55rem}.prose h2{font-size:1.3rem}.prose h3{font-size:1.12rem}
.prose p,.prose ul,.prose ol,.prose blockquote,.prose pre,.prose table{margin:.85em 0}
.prose ul,.prose ol{padding-left:1.4em}
.prose li{margin:.25em 0}
.prose blockquote{padding:.35em 0 .35em 1em;border-left:3px solid var(--accent);color:var(--muted)}
.prose code{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.9em;background:var(--card);padding:.12em .35em;border-radius:4px;border:1px solid var(--line)}
.prose pre{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line)}
.prose pre code{background:transparent;border:0;padding:0}
.prose a{color:var(--accent)}
.prose hr{border:0;border-top:1px solid var(--line);margin:1.5em 0}
.prose strong{font-weight:700}
.prose table{border-collapse:collapse;width:100%;font-size:.95em}
.prose th,.prose td{border:1px solid var(--line);padding:.45em .6em;text-align:left}
.prose th{background:var(--card)}`
	var css string
	if dark {
		css = `:root{--ink:#e8eef7;--muted:#94a3b8;--line:#1e293b;--bg:#0b1220;--card:#111827;--accent:#38bdf8;}
*{box-sizing:border-box}body{margin:0;font-family:Georgia,"Times New Roman",serif;color:var(--ink);background:radial-gradient(1200px 600px at 10% -10%,#1e293b 0%,var(--bg) 55%);line-height:1.55}
.wrap{max-width:720px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}
` + proseCSS + `
.prose h1,.prose h2,.prose h3,.prose h4{color:#f8fafc}
.q{margin:1.25rem 0;padding:1rem 1.1rem;border:1px solid var(--line);border-radius:12px;background:var(--card)}
.q label{display:block;font-family:system-ui,sans-serif;font-weight:600;margin-bottom:.55rem;color:#e2e8f0}
.q input[type=text],.q textarea,.q select{width:100%;padding:.65rem .75rem;border:1px solid #334155;border-radius:8px;font:inherit;background:#0f172a;color:#e2e8f0}
.q textarea{min-height:96px;resize:vertical}
.opts{display:grid;gap:.4rem;font-family:system-ui,sans-serif}
.opts label{font-weight:500;display:flex;gap:.5rem;align-items:flex-start;color:#cbd5e1}
button.primary{margin-top:1.25rem;padding:.7rem 1.1rem;border:0;border-radius:999px;background:var(--accent);color:#0b1220;font:600 14px system-ui,sans-serif;cursor:pointer}
.theme{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}`
	} else {
		css = `:root{--ink:#1c1917;--muted:#57534e;--line:#e7e5e4;--bg:#fafaf9;--card:#fff;--accent:#0f766e;}
*{box-sizing:border-box}body{margin:0;font-family:Georgia,"Times New Roman",serif;color:var(--ink);background:linear-gradient(180deg,#f5f5f4,var(--bg));line-height:1.55}
.wrap{max-width:720px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2}
` + proseCSS + `
.q{margin:1.25rem 0;padding:1rem 1.1rem;border:1px solid var(--line);border-radius:12px;background:var(--card)}
.q label{display:block;font-family:system-ui,sans-serif;font-weight:600;margin-bottom:.55rem}
.q input[type=text],.q textarea,.q select{width:100%;padding:.65rem .75rem;border:1px solid var(--line);border-radius:8px;font:inherit}
.q textarea{min-height:96px;resize:vertical}
.opts{display:grid;gap:.4rem;font-family:system-ui,sans-serif}
.opts label{font-weight:500;display:flex;gap:.5rem;align-items:flex-start}
button.primary{margin-top:1.25rem;padding:.7rem 1.1rem;border:0;border-radius:999px;background:var(--accent);color:#fff;font:600 14px system-ui,sans-serif;cursor:pointer}
.theme{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}`
	}
	mdBody := stripLeadingTitleHeading(markdown, title)
	rendered := markdownToHTMLFragment(mdBody)
	var body strings.Builder
	body.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8"/>`)
	body.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1"/>`)
	body.WriteString(`<title>` + htmlEscapeMinimal(title) + `</title>`)
	body.WriteString(`<style>` + css + `</style></head><body><div class="wrap">`)
	body.WriteString(`<div class="meta">` + htmlEscapeMinimal(kind) + ` · ` + htmlEscapeMinimal(theme) + `</div>`)
	body.WriteString(`<h1 class="page-title">` + htmlEscapeMinimal(title) + `</h1>`)
	body.WriteString(`<div class="prose">` + rendered + `</div>`)
	if kind == "questionnaire" && len(questions) > 0 {
		body.WriteString(`<form id="big-note-form" onsubmit="return false;">`)
		for _, q := range questions {
			qid := htmlEscapeMinimal(q.ID)
			label := htmlEscapeMinimal(q.Label)
			body.WriteString(`<div class="q"><label for="` + qid + `">` + label + `</label>`)
			switch q.Type {
			case "textarea":
				body.WriteString(`<textarea id="` + qid + `" name="` + qid + `"></textarea>`)
			case "single":
				body.WriteString(`<div class="opts">`)
				for i, opt := range q.Options {
					oid := fmt.Sprintf("%s_%d", qid, i)
					o := htmlEscapeMinimal(opt)
					body.WriteString(`<label><input type="radio" name="` + qid + `" value="` + o + `" id="` + oid + `"/> ` + o + `</label>`)
				}
				body.WriteString(`</div>`)
			case "multi":
				body.WriteString(`<div class="opts">`)
				for i, opt := range q.Options {
					oid := fmt.Sprintf("%s_%d", qid, i)
					o := htmlEscapeMinimal(opt)
					body.WriteString(`<label><input type="checkbox" name="` + qid + `" value="` + o + `" id="` + oid + `"/> ` + o + `</label>`)
				}
				body.WriteString(`</div>`)
			default:
				body.WriteString(`<input type="text" id="` + qid + `" name="` + qid + `"/>`)
			}
			body.WriteString(`</div>`)
		}
		body.WriteString(`<button class="primary" type="button">Submit answers</button></form>`)
	}
	body.WriteString(`<p class="theme">Generated Big Note</p></div></body></html>`)
	return body.String()
}

func htmlEscapeMinimal(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

func questionsToJSON(questions []bigNoteQuestion) (sql.NullString, error) {
	if len(questions) == 0 {
		return sql.NullString{}, nil
	}
	b, err := json.Marshal(questions)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{String: string(b), Valid: true}, nil
}

// ListBigNotes GET /api/tran/big-notes
func (h *Handlers) ListBigNotes(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	owner := h.bigNoteOwnerKey(c)
	userID := h.tranUserIDFromContext(c)
	rows, err := h.TranMySQL.DB.Query(
		`SELECT `+bigNoteSelectCols+` FROM big_note
		 WHERE owner_key = ? OR (owner_key = '' AND user_id = ?) OR user_id = ?
		 ORDER BY last_updated DESC LIMIT 500`,
		owner, userID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	list := make([]bigNote, 0)
	for rows.Next() {
		n, err := scanBigNote(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.attachBigNoteURL(&n)
		list = append(list, n)
	}
	c.JSON(http.StatusOK, list)
}

// GetBigNote GET /api/tran/big-notes/:id
func (h *Handlers) GetBigNote(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	n, err := h.getBigNoteOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.attachBigNoteURL(&n)
	refreshBigNoteHTML(&n)
	c.JSON(http.StatusOK, n)
}

// CreateBigNote POST /api/tran/big-notes  { idea, theme?, note_kind? }
func (h *Handlers) CreateBigNote(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in struct {
		Idea     string `json:"idea"`
		Theme    string `json:"theme"`
		NoteKind string `json:"note_kind"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	idea := strings.TrimSpace(in.Idea)
	if idea == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idea is required"})
		return
	}
	if utf8.RuneCountInString(idea) > 8000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idea too long"})
		return
	}
	theme := strings.TrimSpace(in.Theme)
	if theme == "" {
		theme = "dark"
	}
	kind := normalizeNoteKind(in.NoteKind)

	gen, err := h.generateBigNoteContent(c.Request.Context(), idea, theme, kind, "", "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	qJSON, err := questionsToJSON(gen.Questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID := h.tranUserIDFromContext(c)
	owner := h.bigNoteOwnerKey(c)
	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO big_note (user_id, owner_key, title, idea, note_kind, markdown_content, html_content, questions_json, theme)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, owner, gen.Title, idea, kind, gen.Markdown, gen.HTML, qJSON, theme,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	n, err := h.getBigNoteOwned(c, int(id64))
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{
			"id":               id64,
			"title":            gen.Title,
			"idea":             idea,
			"note_kind":        kind,
			"markdown_content": gen.Markdown,
			"html_content":     gen.HTML,
			"questions":        gen.Questions,
			"theme":            theme,
			"owner_key":        owner,
			"user_id":          userID,
		})
		return
	}
	h.attachBigNoteURL(&n)
	c.JSON(http.StatusCreated, n)
}

// RegenerateBigNote POST /api/tran/big-notes/:id/regenerate  { prompt, note_kind? }
func (h *Handlers) RegenerateBigNote(c *gin.Context) {
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
		Prompt   string `json:"prompt"`
		Theme    string `json:"theme"`
		NoteKind string `json:"note_kind"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	if utf8.RuneCountInString(prompt) > 8000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt too long"})
		return
	}

	existing, err := h.getBigNoteOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	theme := strings.TrimSpace(in.Theme)
	if theme == "" {
		theme = existing.Theme
	}
	kind := existing.NoteKind
	if strings.TrimSpace(in.NoteKind) != "" {
		kind = normalizeNoteKind(in.NoteKind)
	}

	gen, err := h.generateBigNoteContent(
		c.Request.Context(),
		existing.Idea,
		theme,
		kind,
		prompt,
		existing.MarkdownContent,
		existing.HTMLContent,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	qJSON, err := questionsToJSON(gen.Questions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = h.TranMySQL.DB.Exec(
		`UPDATE big_note SET title = ?, note_kind = ?, markdown_content = ?, html_content = ?, questions_json = ?, theme = ?, last_updated = NOW()
		 WHERE id = ?`,
		gen.Title, kind, gen.Markdown, gen.HTML, qJSON, theme, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	n, err := h.getBigNoteOwned(c, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id, "title": gen.Title, "note_kind": kind})
		return
	}
	h.attachBigNoteURL(&n)
	c.JSON(http.StatusOK, n)
}

// PublishBigNote POST /api/tran/big-notes/:id/publish
// Publishes HTML via Morph's own public URL (does not require ComposerX).
func (h *Handlers) PublishBigNote(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	n, err := h.getBigNoteOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(n.MarkdownContent) == "" && strings.TrimSpace(n.HTMLContent) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "note has no content to publish"})
		return
	}
	refreshBigNoteHTML(&n)

	slug := ""
	if n.PublishedSlug != nil {
		slug = strings.TrimSpace(*n.PublishedSlug)
	}
	if slug == "" {
		slug, err = h.uniqueBigNoteSlug(n.Title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to allocate publish slug: " + err.Error()})
			return
		}
	}
	path := "/api/tran/public/big-notes/" + slug
	_, err = h.TranMySQL.DB.Exec(
		`UPDATE big_note SET published_slug = ?, published_path = ?, html_content = ?, last_updated = NOW() WHERE id = ?`,
		slug, path, n.HTMLContent, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.getBigNoteOwned(c, id)
	if err != nil {
		pubURL := h.requestAbsoluteURL(c, path)
		c.JSON(http.StatusOK, gin.H{
			"id":             id,
			"published_slug": slug,
			"published_path": path,
			"published_url":  pubURL,
		})
		return
	}
	h.attachBigNoteURL(&updated)
	refreshBigNoteHTML(&updated)
	// Always build an absolute URL from this request's Morph host (not EXTERNAL_API_BASE).
	updated.PublishedURL = h.requestAbsoluteURL(c, path)
	c.JSON(http.StatusOK, updated)
}

func (h *Handlers) requestAbsoluteURL(c *gin.Context, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Prefer Origin / Referer so links open via the Morph frontend proxy when used from the UI.
	if origin := strings.TrimSpace(c.GetHeader("Origin")); origin != "" {
		return strings.TrimSuffix(origin, "/") + path
	}
	if ref := strings.TrimSpace(c.GetHeader("Referer")); ref != "" {
		if u, err := http.NewRequest(http.MethodGet, ref, nil); err == nil && u.URL.Scheme != "" && u.URL.Host != "" {
			return u.URL.Scheme + "://" + u.URL.Host + path
		}
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	}
	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		return path
	}
	return scheme + "://" + host + path
}

// ServePublicBigNote GET /api/tran/public/big-notes/:slug — public HTML (no auth).
func (h *Handlers) ServePublicBigNote(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug required"})
		return
	}
	var title, markdown, kind, theme string
	var questions sql.NullString
	err := h.TranMySQL.DB.QueryRow(
		`SELECT title, markdown_content, note_kind, theme, questions_json FROM big_note WHERE published_slug = ? LIMIT 1`,
		slug,
	).Scan(&title, &markdown, &kind, &theme, &questions)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "published note not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(markdown) == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "published note has no content"})
		return
	}
	var qs []bigNoteQuestion
	if questions.Valid && strings.TrimSpace(questions.String) != "" {
		_ = json.Unmarshal([]byte(questions.String), &qs)
	}
	htmlOut := buildBigNoteHTML(title, markdown, normalizeNoteKind(kind), qs, theme)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlOut))
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// DeleteBigNote DELETE /api/tran/big-notes/:id
func (h *Handlers) DeleteBigNote(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	n, err := h.getBigNoteOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, _ = h.TranMySQL.DB.Exec(`DELETE FROM big_note_response WHERE big_note_id = ?`, n.ID)
	res, err := h.TranMySQL.DB.Exec(`DELETE FROM big_note WHERE id = ?`, n.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListBigNoteResponses GET /api/tran/big-notes/:id/responses
func (h *Handlers) ListBigNoteResponses(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if _, err := h.getBigNoteOwned(c, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := h.TranMySQL.DB.Query(
		`SELECT id, big_note_id, answers_json, analysis_markdown, created_on, last_updated
		 FROM big_note_response WHERE big_note_id = ? ORDER BY created_on DESC LIMIT 200`,
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	list := make([]bigNoteResponse, 0)
	for rows.Next() {
		var r bigNoteResponse
		var analysis sql.NullString
		if err := rows.Scan(&r.ID, &r.BigNoteID, &r.AnswersJSON, &analysis, &r.CreatedOn, &r.LastUpdated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if analysis.Valid {
			s := analysis.String
			r.AnalysisMarkdown = &s
		}
		var answers any
		if json.Unmarshal([]byte(r.AnswersJSON), &answers) == nil {
			r.Answers = answers
		}
		list = append(list, r)
	}
	c.JSON(http.StatusOK, list)
}

// CreateBigNoteResponse POST /api/tran/big-notes/:id/responses  { answers: object }
func (h *Handlers) CreateBigNoteResponse(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	note, err := h.getBigNoteOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if note.NoteKind != "questionnaire" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "responses only apply to questionnaire notes"})
		return
	}
	var in struct {
		Answers any `json:"answers"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if in.Answers == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "answers required"})
		return
	}
	raw, err := json.Marshal(in.Answers)
	if err != nil || len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "answers required"})
		return
	}
	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO big_note_response (big_note_id, answers_json) VALUES (?, ?)`,
		id, string(raw),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rid, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"id":          rid,
		"big_note_id": id,
		"answers":     in.Answers,
	})
}

// AnalyzeBigNoteResponse POST /api/tran/big-notes/:id/responses/:responseId/analyze
func (h *Handlers) AnalyzeBigNoteResponse(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}
	noteID, err := strconv.Atoi(c.Param("id"))
	if err != nil || noteID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	respID, err := strconv.Atoi(c.Param("responseId"))
	if err != nil || respID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid response id"})
		return
	}
	note, err := h.getBigNoteOwned(c, noteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var answersJSON string
	var analysis sql.NullString
	err = h.TranMySQL.DB.QueryRow(
		`SELECT answers_json, analysis_markdown FROM big_note_response WHERE id = ? AND big_note_id = ? LIMIT 1`,
		respID, noteID,
	).Scan(&answersJSON, &analysis)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "response not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prompt := `Analyze questionnaire answers for MorphData Big Notes.
Write clear markdown with:
1) Short summary
2) Key insights / themes
3) Risks or gaps
4) Suggested next actions

Questionnaire idea:
` + note.Idea + `

Title: ` + note.Title + `

Questions JSON:
` + string(note.Questions) + `

Answers JSON:
` + answersJSON

	out, err := h.aiService.ChatCompletionLong(c.Request.Context(), []ai.DashScopeMessage{{Role: "user", Content: prompt}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out = strings.TrimSpace(out)
	_, err = h.TranMySQL.DB.Exec(
		`UPDATE big_note_response SET analysis_markdown = ?, last_updated = NOW() WHERE id = ?`,
		out, respID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":                 respID,
		"big_note_id":        noteID,
		"analysis_markdown":  out,
		"answers_json":       answersJSON,
	})
}

// AnalyzeBigNoteAllResponses POST /api/tran/big-notes/:id/analyze
func (h *Handlers) AnalyzeBigNoteAllResponses(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}
	noteID, err := strconv.Atoi(c.Param("id"))
	if err != nil || noteID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	note, err := h.getBigNoteOwned(c, noteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := h.TranMySQL.DB.Query(
		`SELECT id, answers_json FROM big_note_response WHERE big_note_id = ? ORDER BY created_on ASC LIMIT 100`,
		noteID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var parts []string
	count := 0
	for rows.Next() {
		var id int
		var answers string
		if err := rows.Scan(&id, &answers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		count++
		parts = append(parts, fmt.Sprintf("### Response #%d\n%s", id, answers))
	}
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no responses to analyze"})
		return
	}
	prompt := `Analyze ALL questionnaire responses together for MorphData Big Notes.
Write markdown with overall trends, outliers, and recommended actions.

Questionnaire idea:
` + note.Idea + `

Title: ` + note.Title + `

Questions JSON:
` + string(note.Questions) + `

Responses:
` + strings.Join(parts, "\n\n")

	out, err := h.aiService.ChatCompletionLong(c.Request.Context(), []ai.DashScopeMessage{{Role: "user", Content: prompt}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"big_note_id":       noteID,
		"response_count":    count,
		"analysis_markdown": strings.TrimSpace(out),
	})
}
