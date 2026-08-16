package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robo/docextract"
	"github.com/robo/morphai"
)

const (
	maxEventIngestFileBytes = 8 << 20 // 8 MiB
	maxEventIngestFiles     = 5
	maxEventIngestURLBody   = 2 << 20 // 2 MiB
	maxEventIngestDrafts    = 25
)

const eventInfoAIIngestSystem = `You extract operational Events & Info records from mixed source material (files, web pages, pasted notes).

Return ONLY a JSON array (no markdown fences). Each element:
{
  "title": "short concrete title",
  "detail": "markdown body with what happened, context, and next steps if relevant",
  "reporter": "name or role if stated, otherwise empty string",
  "time": "RFC3339 timestamp if a time is implied, otherwise empty string"
}

Rules:
- Emit one object per distinct event or informational item found in the sources.
- If the sources describe a single incident, return an array with one object.
- title is required and must be specific (not "Event" or "Note").
- detail may use Markdown. Do not invent people, sites, or numbers that were not implied.
- Prefer the source wording. Leave time empty when no clock time is given.
- Cap at 25 items; prioritize the most concrete operational items.`

var (
	eventIngestHTMLTagStripper = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>|<[^>]+>`)
	eventIngestHTMLEntityReplacer = strings.NewReplacer(
		"&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'",
	)
)

type eventInfoIngestDraft struct {
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Reporter string `json:"reporter"`
	Time     string `json:"time"`
}

// IngestEventInfoAI POST /events-info/ai-ingest
//
// Accepts multipart fields: file/files (txt/md/pdf), url, paste/text. Requires at
// least one source. Returns draft records only — nothing is persisted.
func (h *Handler) IngestEventInfoAI(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxEventIngestFileBytes * maxEventIngestFiles); err != nil {
		// Allow JSON or form-urlencoded without files.
		_ = c.Request.ParseForm()
	}

	paste := strings.TrimSpace(firstNonEmptyForm(
		c.PostForm("paste"),
		c.PostForm("text"),
		c.PostForm("content"),
	))
	rawURL := strings.TrimSpace(c.PostForm("url"))

	var files []*multipart.FileHeader
	if form := c.Request.MultipartForm; form != nil {
		files = append(files, form.File["file"]...)
		files = append(files, form.File["files"]...)
	}
	if len(files) > maxEventIngestFiles {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("too many files (max %d)", maxEventIngestFiles),
		})
		return
	}

	if paste == "" && rawURL == "" && len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provide at least one source: a txt/md/pdf file, a URL, or pasted text",
		})
		return
	}

	if h.AI == nil || !h.AI.Configured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI is not configured — set MORPH_AI_API_KEY to extract Events & Info from sources",
		})
		return
	}

	ctx := c.Request.Context()
	var parts []string
	var sourceNotes []string
	truncatedAny := false

	for _, fh := range files {
		name := strings.TrimSpace(fh.Filename)
		if name == "" {
			name = "upload"
		}
		mime := fh.Header.Get("Content-Type")
		if !isAllowedEventIngestFile(name, mime) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("unsupported file type %q; accepted: TXT, MD, PDF", name),
			})
			return
		}
		if fh.Size > maxEventIngestFileBytes {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("file %q is too large (max %d MiB)", name, maxEventIngestFileBytes/(1024*1024)),
			})
			return
		}
		text, err := h.readEventIngestFile(c, fh, name, mime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		text, trunc := docextract.Truncate(text, docextract.MaxPerFileChars)
		if trunc {
			truncatedAny = true
		}
		parts = append(parts, fmt.Sprintf("### File: %s\n%s", name, text))
		sourceNotes = append(sourceNotes, "file:"+name)
	}

	if rawURL != "" {
		text, err := fetchEventIngestURLText(ctx, rawURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL: " + err.Error()})
			return
		}
		text, trunc := docextract.Truncate(text, docextract.MaxPerFileChars)
		if trunc {
			truncatedAny = true
		}
		parts = append(parts, fmt.Sprintf("### URL: %s\n%s", rawURL, text))
		sourceNotes = append(sourceNotes, "url")
	}

	if paste != "" {
		text, trunc := docextract.Truncate(paste, docextract.MaxPerFileChars)
		if trunc {
			truncatedAny = true
		}
		parts = append(parts, "### Pasted content\n"+text)
		sourceNotes = append(sourceNotes, "paste")
	}

	corpus := strings.Join(parts, "\n\n")
	corpus, truncAll := docextract.Truncate(corpus, docextract.MaxRequestChars)
	if truncAll {
		truncatedAny = true
	}
	if strings.TrimSpace(corpus) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sources produced no readable text"})
		return
	}

	userPrompt := fmt.Sprintf(
		"Extract Events & Info records from these sources.\nCurrent UTC time for reference: %s\n\n%s",
		time.Now().UTC().Format(time.RFC3339),
		corpus,
	)

	reply, err := h.AI.ChatCompletion(ctx, []morphai.Message{
		{Role: "system", Content: eventInfoAIIngestSystem},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	raw, ok := extractJSONArrayOrObject(reply)
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI reply was not valid JSON", "raw": truncateStr(reply, 800)})
		return
	}

	drafts, err := parseEventIngestDrafts(raw)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "raw": truncateStr(raw, 800)})
		return
	}
	if len(drafts) == 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI returned no usable drafts"})
		return
	}
	if len(drafts) > maxEventIngestDrafts {
		drafts = drafts[:maxEventIngestDrafts]
	}

	c.JSON(http.StatusOK, gin.H{
		"drafts":            drafts,
		"total":             len(drafts),
		"sources":           sourceNotes,
		"source_truncated":  truncatedAny,
		"assistant_message": fmt.Sprintf("Extracted %d draft(s). Review and save the ones you want.", len(drafts)),
	})
}

func firstNonEmptyForm(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func isAllowedEventIngestFile(name, mime string) bool {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(name)))
	switch ext {
	case ".txt", ".md", ".markdown", ".pdf":
		return true
	}
	m := strings.ToLower(strings.TrimSpace(mime))
	if i := strings.Index(m, ";"); i >= 0 {
		m = strings.TrimSpace(m[:i])
	}
	switch m {
	case "application/pdf", "text/plain", "text/markdown":
		return true
	}
	return false
}

func (h *Handler) readEventIngestFile(c *gin.Context, fh *multipart.FileHeader, name, mime string) (string, error) {
	if docextract.IsPDF(name, mime) {
		tmpDir, err := os.MkdirTemp("", "events-ingest-")
		if err != nil {
			return "", fmt.Errorf("could not prepare temporary storage")
		}
		defer os.RemoveAll(tmpDir)
		dest := filepath.Join(tmpDir, "upload.pdf")
		if err := c.SaveUploadedFile(fh, dest); err != nil {
			return "", fmt.Errorf("could not save uploaded file %q", name)
		}
		text, err := docextract.ExtractPDF(dest)
		if err != nil {
			return "", fmt.Errorf("%s: %w", name, err)
		}
		return text, nil
	}

	f, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("could not open %q", name)
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, maxEventIngestFileBytes+1))
	if err != nil {
		return "", fmt.Errorf("could not read %q", name)
	}
	if len(raw) > maxEventIngestFileBytes {
		return "", fmt.Errorf("file %q is too large", name)
	}
	return docextract.ExtractText(raw), nil
}

func fetchEventIngestURLText(ctx context.Context, rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("invalid URL")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", errors.New("URL must be http or https")
	}
	if isBlockedEventIngestHost(u.Hostname()) {
		return "", errors.New("URL host is not allowed")
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if isBlockedEventIngestHost(req.URL.Hostname()) {
				return errors.New("redirect host is not allowed")
			}
			return nil
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "FormsX-EventsInfo/1.0")
	req.Header.Set("Accept", "text/html,text/plain,application/xhtml+xml,application/pdf;q=0.8,*/*;q=0.5")

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("returned HTTP %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxEventIngestURLBody+1))
	if err != nil {
		return "", errors.New("failed to read response body")
	}
	if len(body) > maxEventIngestURLBody {
		return "", errors.New("response too large")
	}

	ct := strings.ToLower(res.Header.Get("Content-Type"))
	pathLower := strings.ToLower(u.Path)
	if strings.Contains(ct, "pdf") || strings.HasSuffix(pathLower, ".pdf") {
		text, err := docextract.ExtractPDFBytes(body)
		if err != nil {
			return "", fmt.Errorf("PDF extraction failed: %w", err)
		}
		return text, nil
	}

	raw := string(body)
	if strings.Contains(ct, "html") || strings.Contains(strings.ToLower(raw[:min(200, len(raw))]), "<html") {
		return htmlToPlainEventIngest(raw), nil
	}
	return docextract.ExtractText(body), nil
}

func htmlToPlainEventIngest(s string) string {
	s = eventIngestHTMLTagStripper.ReplaceAllString(s, " ")
	s = eventIngestHTMLEntityReplacer.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func isBlockedEventIngestHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") || host == "0.0.0.0" {
		return true
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	if ip == nil {
		addrs, err := net.LookupIP(host)
		if err != nil {
			return false
		}
		for _, a := range addrs {
			if isPrivateOrLocalEventIngestIP(a) {
				return true
			}
		}
		return false
	}
	return isPrivateOrLocalEventIngestIP(ip)
}

func isPrivateOrLocalEventIngestIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	}
	return false
}

func extractJSONArrayOrObject(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	if start := strings.Index(s, "["); start >= 0 {
		depth := 0
		for i, ch := range s[start:] {
			switch ch {
			case '[':
				depth++
			case ']':
				depth--
				if depth == 0 {
					return s[start : start+i+1], true
				}
			}
		}
	}
	return morphai.ExtractJSONObject(s)
}

func parseEventIngestDrafts(raw string) ([]eventInfoIngestDraft, error) {
	raw = strings.TrimSpace(raw)
	var list []eventInfoIngestDraft
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &list); err != nil {
			return nil, fmt.Errorf("could not parse AI drafts: %w", err)
		}
	} else {
		var one eventInfoIngestDraft
		if err := json.Unmarshal([]byte(raw), &one); err != nil {
			return nil, fmt.Errorf("could not parse AI draft: %w", err)
		}
		list = []eventInfoIngestDraft{one}
	}

	out := make([]eventInfoIngestDraft, 0, len(list))
	for _, d := range list {
		d.Title = strings.TrimSpace(d.Title)
		d.Detail = strings.TrimSpace(d.Detail)
		d.Reporter = strings.TrimSpace(d.Reporter)
		d.Time = strings.TrimSpace(d.Time)
		if d.Title == "" {
			continue
		}
		if len(d.Title) > eventInfoMaxTitleLen {
			d.Title = d.Title[:eventInfoMaxTitleLen]
		}
		if len(d.Detail) > eventInfoMaxDetailLen {
			d.Detail = d.Detail[:eventInfoMaxDetailLen]
		}
		if len(d.Reporter) > eventInfoMaxReporterLen {
			d.Reporter = d.Reporter[:eventInfoMaxReporterLen]
		}
		if d.Time != "" {
			if t, err := time.Parse(time.RFC3339, d.Time); err == nil {
				d.Time = t.UTC().Format(time.RFC3339)
			} else if t, err := time.Parse(time.RFC3339Nano, d.Time); err == nil {
				d.Time = t.UTC().Format(time.RFC3339)
			} else {
				d.Time = ""
			}
		}
		out = append(out, d)
	}
	return out, nil
}
