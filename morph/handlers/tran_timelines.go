package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"idongivaflyinfa/ai"

	"github.com/gin-gonic/gin"
	"github.com/robo/morphgraph"
)

const (
	timelineMaxFileBytes   = 5 * 1024 * 1024 // 5 MiB
	timelineMaxURLBody     = 2 * 1024 * 1024 // 2 MiB fetch cap
	timelineMaxSourceRunes = 48000
	timelineMaxPasteRunes  = 100000
)

type timelineDoc struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	OwnerKey        string    `json:"owner_key"`
	Title           string    `json:"title"`
	SourceSummary   string    `json:"source_summary"`
	SourceFileName  *string   `json:"source_file_name,omitempty"`
	SourceURL       *string   `json:"source_url,omitempty"`
	HasPaste        bool      `json:"has_paste"`
	MarkdownContent string    `json:"markdown_content"`
	HTMLContent     string    `json:"html_content"`
	PublishedSlug   *string   `json:"published_slug,omitempty"`
	PublishedPath   *string   `json:"published_path,omitempty"`
	PublishedURL    string    `json:"published_url,omitempty"`
	CreatedOn       time.Time `json:"created_on"`
	LastUpdated     time.Time `json:"last_updated"`
}

type timelineAIResult struct {
	Title    string `json:"title"`
	Markdown string `json:"markdown"`
}

const timelineSelectCols = `id, user_id, owner_key, title, source_summary, source_file_name, source_url, has_paste,
	markdown_content, html_content, published_slug, published_path, created_on, last_updated`

func (h *Handlers) timelineOwnerKey(c *gin.Context) string {
	return h.bigNoteOwnerKey(c)
}

func (h *Handlers) timelinePublicURL(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasPrefix(path, "/api/tran/public/timelines/") {
		return path
	}
	base := strings.TrimSpace(h.tranMailBase)
	if base == "" {
		return path
	}
	return strings.TrimSuffix(base, "/") + path
}

func (h *Handlers) attachTimelineURL(t *timelineDoc) {
	if t == nil || t.PublishedPath == nil {
		return
	}
	t.PublishedURL = h.timelinePublicURL(*t.PublishedPath)
}

func scanTimeline(scanner interface {
	Scan(dest ...any) error
}) (timelineDoc, error) {
	var t timelineDoc
	var fileName, srcURL, pubSlug, pubPath sql.NullString
	var hasPaste int
	err := scanner.Scan(
		&t.ID, &t.UserID, &t.OwnerKey, &t.Title, &t.SourceSummary, &fileName, &srcURL, &hasPaste,
		&t.MarkdownContent, &t.HTMLContent, &pubSlug, &pubPath, &t.CreatedOn, &t.LastUpdated,
	)
	if err != nil {
		return t, err
	}
	t.HasPaste = hasPaste != 0
	if fileName.Valid && strings.TrimSpace(fileName.String) != "" {
		s := fileName.String
		t.SourceFileName = &s
	}
	if srcURL.Valid && strings.TrimSpace(srcURL.String) != "" {
		s := srcURL.String
		t.SourceURL = &s
	}
	if pubSlug.Valid {
		s := pubSlug.String
		t.PublishedSlug = &s
	}
	if pubPath.Valid {
		p := pubPath.String
		t.PublishedPath = &p
	}
	return t, nil
}

func (h *Handlers) getTimelineOwned(c *gin.Context, id int) (timelineDoc, error) {
	owner := h.timelineOwnerKey(c)
	userID := h.tranUserIDFromContext(c)
	row := h.TranMySQL.DB.QueryRow(
		`SELECT `+timelineSelectCols+` FROM timeline
		 WHERE id = ? AND (owner_key = ? OR (owner_key = '' AND user_id = ?) OR user_id = ?)
		 LIMIT 1`,
		id, owner, userID, userID,
	)
	return scanTimeline(row)
}

func slugifyTimeline(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = bigNoteSlugCleaner.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "timeline"
	}
	if utf8.RuneCountInString(s) > 60 {
		runes := []rune(s)
		s = string(runes[:60])
		s = strings.Trim(s, "-")
	}
	return s
}

func (h *Handlers) uniqueTimelineSlug(base string) (string, error) {
	base = slugifyTimeline(base)
	candidate := base
	for i := 0; i < 50; i++ {
		var existing int
		err := h.TranMySQL.DB.QueryRow(`SELECT id FROM timeline WHERE published_slug = ? LIMIT 1`, candidate).Scan(&existing)
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

func timelineFileExtOK(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".txt", ".pdf", ".md", ".markdown":
		return true
	default:
		return false
	}
}

func extractTimelineFileText(filename string, raw []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		text, err := morphgraph.ExtractPDFBytes(raw)
		if err != nil {
			return "", fmt.Errorf("PDF extraction failed: %w", err)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			return "", errors.New("PDF has no extractable text")
		}
		return text, nil
	case ".txt", ".md", ".markdown":
		text := morphgraph.ExtractPlainText(filename, "", string(raw))
		text = strings.TrimSpace(text)
		if text == "" {
			return "", errors.New("file has no extractable text")
		}
		return text, nil
	default:
		return "", fmt.Errorf("unsupported file type %q (allowed: .txt, .pdf, .md)", ext)
	}
}

var timelineHTMLTagStripper = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>|<[^>]+>`)
var timelineHTMLEntityReplacer = strings.NewReplacer(
	"&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'",
)

func htmlToPlainText(s string) string {
	s = timelineHTMLTagStripper.ReplaceAllString(s, " ")
	s = timelineHTMLEntityReplacer.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func isBlockedTimelineHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || host == "0.0.0.0" {
		return true
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	if ip == nil {
		// Resolve and check; if resolve fails, fetch will fail later.
		addrs, err := net.LookupIP(host)
		if err != nil {
			return false
		}
		for _, a := range addrs {
			if isPrivateOrLocalIP(a) {
				return true
			}
		}
		return false
	}
	return isPrivateOrLocalIP(ip)
}

func isPrivateOrLocalIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	// Extra: CGNAT / metadata ranges
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	}
	return false
}

func fetchTimelineURLText(ctx context.Context, rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("invalid URL")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", errors.New("URL must be http or https")
	}
	if isBlockedTimelineHost(u.Hostname()) {
		return "", errors.New("URL host is not allowed")
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if isBlockedTimelineHost(req.URL.Hostname()) {
				return errors.New("redirect host is not allowed")
			}
			return nil
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "MorphData-Timelines/1.0")
	req.Header.Set("Accept", "text/html,text/plain,application/xhtml+xml,application/pdf;q=0.8,*/*;q=0.5")

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("URL returned HTTP %d", res.StatusCode)
	}
	limited := io.LimitReader(res.Body, timelineMaxURLBody+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("failed to read URL body: %w", err)
	}
	if len(body) > timelineMaxURLBody {
		return "", errors.New("URL response too large")
	}

	ct := strings.ToLower(res.Header.Get("Content-Type"))
	pathLower := strings.ToLower(u.Path)
	if strings.Contains(ct, "pdf") || strings.HasSuffix(pathLower, ".pdf") {
		text, err := morphgraph.ExtractPDFBytes(body)
		if err != nil {
			return "", fmt.Errorf("URL PDF extraction failed: %w", err)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			return "", errors.New("URL PDF has no extractable text")
		}
		return text, nil
	}
	raw := string(body)
	var text string
	head := raw
	if len(head) > 200 {
		head = head[:200]
	}
	headLower := strings.ToLower(head)
	if strings.Contains(ct, "html") || strings.Contains(headLower, "<html") || strings.Contains(headLower, "<!doctype") {
		text = htmlToPlainText(raw)
	} else {
		text = morphgraph.ExtractPlainText(filepath.Base(u.Path), ct, raw)
		if strings.TrimSpace(text) == "" {
			text = strings.TrimSpace(raw)
		}
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("URL has no extractable text")
	}
	return text, nil
}

func buildTimelineHTML(title, markdown string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Timeline"
	}
	proseCSS := `.prose{font-size:1.05rem;color:var(--ink)}
.prose>:first-child{margin-top:0}
.prose h1,.prose h2,.prose h3,.prose h4{line-height:1.25;margin:1.35em 0 .55em;font-weight:700;color:#f8fafc}
.prose h1{font-size:1.55rem}.prose h2{font-size:1.3rem}.prose h3{font-size:1.12rem}
.prose p,.prose ul,.prose ol,.prose blockquote,.prose pre,.prose table{margin:.85em 0}
.prose ul,.prose ol{padding-left:1.4em}
.prose li{margin:.35em 0}
.prose blockquote{padding:.35em 0 .35em 1em;border-left:3px solid var(--accent);color:var(--muted)}
.prose code{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.9em;background:var(--card);padding:.12em .35em;border-radius:4px;border:1px solid var(--line)}
.prose pre{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line)}
.prose a{color:var(--accent)}
.prose hr{border:0;border-top:1px solid var(--line);margin:1.5em 0}
.prose strong{font-weight:700}
.prose table{border-collapse:collapse;width:100%;font-size:.95em}
.prose th,.prose td{border:1px solid var(--line);padding:.45em .6em;text-align:left}
.prose th{background:var(--card)}`
	css := `:root{--ink:#e8eef7;--muted:#94a3b8;--line:#1e293b;--bg:#0b1220;--card:#111827;--accent:#38bdf8;}
*{box-sizing:border-box}body{margin:0;font-family:Georgia,"Times New Roman",serif;color:var(--ink);background:radial-gradient(1200px 600px at 10% -10%,#1e293b 0%,var(--bg) 55%);line-height:1.55}
.wrap{max-width:760px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}
` + proseCSS + `
.foot{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}`
	mdBody := stripLeadingTitleHeading(markdown, title)
	rendered := markdownToHTMLFragment(mdBody)
	var body strings.Builder
	body.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8"/>`)
	body.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1"/>`)
	body.WriteString(`<title>` + htmlEscapeMinimal(title) + `</title>`)
	body.WriteString(`<style>` + css + `</style></head><body><div class="wrap">`)
	body.WriteString(`<div class="meta">Timeline</div>`)
	body.WriteString(`<h1 class="page-title">` + htmlEscapeMinimal(title) + `</h1>`)
	body.WriteString(`<div class="prose">` + rendered + `</div>`)
	body.WriteString(`<p class="foot">Generated Timeline</p></div></body></html>`)
	return body.String()
}

func (h *Handlers) generateTimelineContent(ctx context.Context, sourceText, preferredTitle string) (timelineAIResult, string, error) {
	var out timelineAIResult
	if h.aiService == nil {
		return out, "", errors.New("AI service not configured")
	}
	sourceText = strings.TrimSpace(sourceText)
	if sourceText == "" {
		return out, "", errors.New("no source text to generate from")
	}
	sourceText = truncateRunes(sourceText, timelineMaxSourceRunes)

	var b strings.Builder
	b.WriteString(`You create MorphData Timelines from source material.
Respond with ONLY one compact JSON object (no markdown fences) with keys:
- "title": short timeline title (max ~80 chars)
- "markdown": a chronological timeline as markdown

Markdown rules:
- Use clear dated or ordered events (## headings or bullet milestones with dates when known).
- Prefer chronological order; note uncertain dates explicitly.
- Keep it readable and faithful to the source; do not invent major facts.
- Aim under ~1200 words.
`)
	if preferredTitle != "" {
		b.WriteString("\nPreferred title (use unless clearly wrong): ")
		b.WriteString(preferredTitle)
	}
	b.WriteString("\n\nSource material:\n")
	b.WriteString(sourceText)

	genCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	raw, err := h.aiService.ChatCompletionLong(genCtx, []ai.DashScopeMessage{{Role: "user", Content: b.String()}})
	if err != nil {
		return out, "", err
	}
	jsonStr, ok := extractJSONObjectAware(raw)
	if !ok {
		jsonStr, ok = extractJSONObject(raw)
	}
	if !ok {
		return out, "", errors.New("AI did not return JSON for the timeline")
	}
	if err := json.Unmarshal([]byte(jsonStr), &out); err != nil {
		return out, "", fmt.Errorf("parse AI timeline JSON: %w", err)
	}
	out.Title = strings.TrimSpace(out.Title)
	out.Markdown = strings.TrimSpace(out.Markdown)
	if preferredTitle != "" {
		out.Title = preferredTitle
	}
	if out.Title == "" {
		out.Title = "Untitled timeline"
	}
	if out.Markdown == "" {
		return out, "", errors.New("AI returned empty timeline markdown")
	}
	if utf8.RuneCountInString(out.Title) > 255 {
		out.Title = string([]rune(out.Title)[:255])
	}
	htmlOut := buildTimelineHTML(out.Title, out.Markdown)
	return out, htmlOut, nil
}

// ListTimelines GET /api/tran/timelines
func (h *Handlers) ListTimelines(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	owner := h.timelineOwnerKey(c)
	userID := h.tranUserIDFromContext(c)
	rows, err := h.TranMySQL.DB.Query(
		`SELECT `+timelineSelectCols+` FROM timeline
		 WHERE owner_key = ? OR (owner_key = '' AND user_id = ?) OR user_id = ?
		 ORDER BY last_updated DESC LIMIT 500`,
		owner, userID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	list := make([]timelineDoc, 0)
	for rows.Next() {
		t, err := scanTimeline(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.attachTimelineURL(&t)
		list = append(list, t)
	}
	c.JSON(http.StatusOK, list)
}

// GetTimeline GET /api/tran/timelines/:id
func (h *Handlers) GetTimeline(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	t, err := h.getTimelineOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "timeline not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.attachTimelineURL(&t)
	if strings.TrimSpace(t.HTMLContent) == "" && strings.TrimSpace(t.MarkdownContent) != "" {
		t.HTMLContent = buildTimelineHTML(t.Title, t.MarkdownContent)
	}
	c.JSON(http.StatusOK, t)
}

// CreateTimeline POST /api/tran/timelines (multipart: file?, url?, paste|content?, title?)
func (h *Handlers) CreateTimeline(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}

	var titleHint, paste, srcURL string
	ct := strings.ToLower(c.ContentType())
	if strings.HasPrefix(ct, "application/json") {
		var in struct {
			Title   string `json:"title"`
			Paste   string `json:"paste"`
			Content string `json:"content"`
			URL     string `json:"url"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		titleHint = strings.TrimSpace(in.Title)
		paste = strings.TrimSpace(firstNonEmpty(in.Paste, in.Content))
		srcURL = strings.TrimSpace(in.URL)
	} else {
		if err := c.Request.ParseMultipartForm(timelineMaxFileBytes + (1 << 20)); err != nil {
			_ = c.Request.ParseForm()
		}
		titleHint = strings.TrimSpace(firstNonEmpty(c.PostForm("title"), c.Request.FormValue("title")))
		paste = strings.TrimSpace(firstNonEmpty(c.PostForm("paste"), c.PostForm("content"), c.Request.FormValue("paste"), c.Request.FormValue("content")))
		srcURL = strings.TrimSpace(firstNonEmpty(c.PostForm("url"), c.Request.FormValue("url")))
	}

	var fileText string
	var fileName string
	file, hdr, fileErr := c.Request.FormFile("file")
	hasFile := fileErr == nil && hdr != nil
	if hasFile {
		defer file.Close()
		fileName = filepath.Base(hdr.Filename)
		if !timelineFileExtOK(fileName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type (allowed: .txt, .pdf, .md)"})
			return
		}
		if hdr.Size > timelineMaxFileBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file exceeds %d byte size limit", timelineMaxFileBytes)})
			return
		}
		limited := io.LimitReader(file, timelineMaxFileBytes+1)
		raw, err := io.ReadAll(limited)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file"})
			return
		}
		if len(raw) > timelineMaxFileBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file exceeds %d byte size limit", timelineMaxFileBytes)})
			return
		}
		text, err := extractTimelineFileText(fileName, raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fileText = text
	} else if fileErr != nil && !errors.Is(fileErr, http.ErrMissingFile) && c.Request.MultipartForm != nil {
		// Only treat real file field errors as failures when multipart was used with a file attempt.
		if _, ok := c.Request.MultipartForm.File["file"]; ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file"})
			return
		}
	}

	if utf8.RuneCountInString(paste) > timelineMaxPasteRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paste content too long"})
		return
	}

	hasPaste := paste != ""
	hasURL := srcURL != ""
	if !hasFile && !hasPaste && !hasURL {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide at least one source: file, URL, or pasted content"})
		return
	}

	var parts []string
	var summaryBits []string
	if hasFile {
		parts = append(parts, "## Uploaded file: "+fileName+"\n\n"+fileText)
		summaryBits = append(summaryBits, "file:"+fileName)
	}
	if hasURL {
		urlText, err := fetchTimelineURLText(c.Request.Context(), srcURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		parts = append(parts, "## URL: "+srcURL+"\n\n"+urlText)
		summaryBits = append(summaryBits, "url:"+truncateRunes(srcURL, 120))
	}
	if hasPaste {
		parts = append(parts, "## Pasted content\n\n"+paste)
		summaryBits = append(summaryBits, "paste")
	}
	combined := strings.TrimSpace(strings.Join(parts, "\n\n"))
	if combined == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sources produced no usable text"})
		return
	}

	gen, htmlOut, err := h.generateTimelineContent(c.Request.Context(), combined, titleHint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID := h.tranUserIDFromContext(c)
	owner := h.timelineOwnerKey(c)
	summary := strings.Join(summaryBits, ", ")
	var fileNameArg any
	if fileName != "" {
		fileNameArg = fileName
	}
	var urlArg any
	if hasURL {
		urlArg = srcURL
	}
	hasPasteInt := 0
	if hasPaste {
		hasPasteInt = 1
	}

	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO timeline (user_id, owner_key, title, source_summary, source_file_name, source_url, has_paste, markdown_content, html_content)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, owner, gen.Title, summary, fileNameArg, urlArg, hasPasteInt, gen.Markdown, htmlOut,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	t, err := h.getTimelineOwned(c, int(id64))
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{
			"id":               id64,
			"title":            gen.Title,
			"source_summary":   summary,
			"markdown_content": gen.Markdown,
			"html_content":     htmlOut,
			"owner_key":        owner,
			"user_id":          userID,
		})
		return
	}
	h.attachTimelineURL(&t)
	c.JSON(http.StatusCreated, t)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// DeleteTimeline DELETE /api/tran/timelines/:id
func (h *Handlers) DeleteTimeline(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if _, err := h.getTimelineOwned(c, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "timeline not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.TranMySQL.DB.Exec(`DELETE FROM timeline WHERE id = ?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "id": id})
}

// PublishTimeline POST /api/tran/timelines/:id/publish
func (h *Handlers) PublishTimeline(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	t, err := h.getTimelineOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "timeline not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(t.MarkdownContent) == "" && strings.TrimSpace(t.HTMLContent) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timeline has no content to publish"})
		return
	}
	if strings.TrimSpace(t.HTMLContent) == "" {
		t.HTMLContent = buildTimelineHTML(t.Title, t.MarkdownContent)
	}

	slug := ""
	if t.PublishedSlug != nil {
		slug = strings.TrimSpace(*t.PublishedSlug)
	}
	if slug == "" {
		slug, err = h.uniqueTimelineSlug(t.Title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to allocate publish slug: " + err.Error()})
			return
		}
	}
	path := "/api/tran/public/timelines/" + slug
	_, err = h.TranMySQL.DB.Exec(
		`UPDATE timeline SET published_slug = ?, published_path = ?, html_content = ?, last_updated = NOW() WHERE id = ?`,
		slug, path, t.HTMLContent, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.getTimelineOwned(c, id)
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
	h.attachTimelineURL(&updated)
	updated.PublishedURL = h.requestAbsoluteURL(c, path)
	c.JSON(http.StatusOK, updated)
}

// ServePublicTimeline GET /api/tran/public/timelines/:slug
func (h *Handlers) ServePublicTimeline(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug required"})
		return
	}
	var title, markdown, htmlContent string
	err := h.TranMySQL.DB.QueryRow(
		`SELECT title, markdown_content, html_content FROM timeline WHERE published_slug = ? LIMIT 1`,
		slug,
	).Scan(&title, &markdown, &htmlContent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "published timeline not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(markdown) == "" && strings.TrimSpace(htmlContent) == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "published timeline has no content"})
		return
	}
	htmlOut := strings.TrimSpace(htmlContent)
	if htmlOut == "" {
		htmlOut = buildTimelineHTML(title, markdown)
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlOut))
}
