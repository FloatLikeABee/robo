package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"idongivaflyinfa/ai"
	"idongivaflyinfa/importcol"

	"github.com/gin-gonic/gin"
	"github.com/robo/docextract"
	"github.com/robo/morphai"
)

const (
	caseTaskAIMaxFileBytes = 5 * 1024 * 1024
	caseTaskAIMaxRunes     = 48000
)

type caseTaskAILocation struct {
	Label string      `json:"label,omitempty"`
	Area  [][]float64 `json:"area,omitempty"`
}

type caseTaskAIDraft struct {
	Title       string              `json:"title"`
	Markdown    string              `json:"markdown"`
	Description string              `json:"description"`
	StartAt     string              `json:"start_at"`
	EndAt       string              `json:"end_at"`
	Location    *caseTaskAILocation `json:"location,omitempty"`
	Detail      map[string]any      `json:"detail,omitempty"`
}

// CreateCaseTaskAIDraft POST /api/tran/case-tasks/ai-draft
// Builds a reviewable case/task draft from a prompt and/or uploaded file. Does not insert a row.
func (h *Handlers) CreateCaseTaskAIDraft(c *gin.Context) {
	prompt, fileName, fileBytes, err := readCaseTaskAISources(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var fileText string
	if len(fileBytes) > 0 || fileName != "" {
		fileText, err = extractCaseTaskAIFileText(fileName, fileBytes)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	combined := combineCaseTaskAISources(prompt, fileName, fileText)
	if strings.TrimSpace(combined) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide a prompt, a file, or both"})
		return
	}
	combined = truncateRunes(combined, caseTaskAIMaxRunes)

	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}

	raw, err := h.aiService.ChatCompletionLong(c.Request.Context(), []ai.DashScopeMessage{
		{Role: "user", Content: caseTaskAIDraftPrompt(combined)},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	draft, err := parseCaseTaskAIDraftOutput(raw)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, draft)
}

func readCaseTaskAISources(c *gin.Context) (prompt, fileName string, fileBytes []byte, err error) {
	ct := strings.ToLower(c.ContentType())
	if strings.HasPrefix(ct, "multipart/") {
		if err := c.Request.ParseMultipartForm(caseTaskAIMaxFileBytes + (1 << 20)); err != nil {
			_ = c.Request.ParseForm()
		}
		prompt = strings.TrimSpace(firstNonEmpty(c.PostForm("prompt"), c.Request.FormValue("prompt")))
		file, hdr, fileErr := c.Request.FormFile("file")
		if fileErr == nil && hdr != nil {
			defer file.Close()
			fileName = filepath.Base(hdr.Filename)
			if hdr.Size > caseTaskAIMaxFileBytes {
				return "", "", nil, fmt.Errorf("file exceeds %d byte size limit", caseTaskAIMaxFileBytes)
			}
			limited := io.LimitReader(file, caseTaskAIMaxFileBytes+1)
			raw, readErr := io.ReadAll(limited)
			if readErr != nil {
				return "", "", nil, errors.New("failed to read uploaded file")
			}
			if len(raw) > caseTaskAIMaxFileBytes {
				return "", "", nil, fmt.Errorf("file exceeds %d byte size limit", caseTaskAIMaxFileBytes)
			}
			return prompt, fileName, raw, nil
		}
		if fileErr != nil && !errors.Is(fileErr, http.ErrMissingFile) && c.Request.MultipartForm != nil {
			if _, ok := c.Request.MultipartForm.File["file"]; ok {
				return "", "", nil, errors.New("failed to read uploaded file")
			}
		}
		return prompt, "", nil, nil
	}

	var in struct {
		Prompt string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		return "", "", nil, errors.New("invalid json")
	}
	return strings.TrimSpace(in.Prompt), "", nil, nil
}

func combineCaseTaskAISources(prompt, fileName, fileText string) string {
	var parts []string
	if strings.TrimSpace(prompt) != "" {
		parts = append(parts, "## Prompt\n\n"+strings.TrimSpace(prompt))
	}
	fileText = strings.TrimSpace(fileText)
	if fileText != "" {
		label := strings.TrimSpace(fileName)
		if label == "" {
			label = "upload"
		}
		parts = append(parts, "## Uploaded file: "+label+"\n\n"+fileText)
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func caseTaskAIFileExtOK(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".txt", ".md", ".markdown", ".pdf", ".csv", ".xlsx":
		return true
	default:
		return false
	}
}

func extractCaseTaskAIFileText(filename string, raw []byte) (string, error) {
	name := filepath.Base(strings.TrimSpace(filename))
	if name == "" || name == "." {
		return "", errors.New("file name is required")
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".xls" {
		return "", errors.New("unsupported file type (legacy .xls is not supported; use .xlsx or .csv)")
	}
	if !caseTaskAIFileExtOK(name) {
		return "", errors.New("unsupported file type (allowed: .txt, .md, .pdf, .csv, .xlsx)")
	}
	switch ext {
	case ".pdf":
		text, err := docextract.ExtractPDFBytes(raw)
		if err != nil {
			if errors.Is(err, docextract.ErrNoText) {
				return "", errors.New("PDF has no extractable text")
			}
			return "", fmt.Errorf("PDF extraction failed: %w", err)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			return "", errors.New("PDF has no extractable text")
		}
		return text, nil
	case ".txt", ".md", ".markdown":
		text := strings.TrimSpace(strings.ToValidUTF8(string(raw), ""))
		if text == "" {
			return "", errors.New("file has no extractable text")
		}
		return text, nil
	case ".csv", ".xlsx":
		parsed, err := importcol.ParseUpload(name, raw)
		if err != nil {
			return "", err
		}
		text := tabularFileToPromptText(parsed)
		if text == "" {
			return "", errors.New("file has no extractable text")
		}
		return text, nil
	default:
		return "", errors.New("unsupported file type (allowed: .txt, .md, .pdf, .csv, .xlsx)")
	}
}

func tabularFileToPromptText(parsed *importcol.ParsedFile) string {
	if parsed == nil {
		return ""
	}
	var b strings.Builder
	if len(parsed.Headers) > 0 {
		b.WriteString(strings.Join(parsed.Headers, ", "))
		b.WriteByte('\n')
	}
	for _, row := range parsed.Rows {
		var bits []string
		for _, h := range parsed.Headers {
			v := strings.TrimSpace(row[h])
			if v == "" {
				continue
			}
			bits = append(bits, h+": "+v)
		}
		if len(bits) == 0 {
			continue
		}
		b.WriteString(strings.Join(bits, "; "))
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func caseTaskAIDraftPrompt(source string) string {
	return `You create a MorphNotes case/task draft from source material.
Respond with ONLY one JSON object (no markdown fences around the JSON, no commentary) with keys:
- "title": short task title (required, max ~80 characters)
- "markdown": the task document in markdown. Include a mermaid diagram or chart when the source is structure, process, comparison, or quantities. Use "" if there is nothing to write.
- "start_at": ISO-8601 datetime if the source implies a start time, else ""
- "end_at": ISO-8601 datetime if the source implies an end time, else ""

Do not invent facts that are not in the source. Do not include location or JSON detail objects.

` + morphai.VisualFirstInstructions + `

Source material:
` + source
}

func parseCaseTaskAIDraftOutput(raw string) (caseTaskAIDraft, error) {
	var out caseTaskAIDraft
	obj, ok := extractJSONObjectLoose(raw)
	if !ok {
		if s, found := morphai.ExtractJSONObject(raw); found {
			if err := json.Unmarshal([]byte(s), &obj); err != nil || obj == nil {
				return out, errors.New("AI did not return a valid case/task draft")
			}
			ok = true
		}
	}
	if !ok || obj == nil {
		return out, errors.New("AI did not return a valid case/task draft")
	}
	out.Title = strings.TrimSpace(optionalDraftString(obj["title"]))
	if out.Title == "" {
		return out, errors.New("AI did not return a valid case/task draft")
	}
	if utf8.RuneCountInString(out.Title) > 255 {
		out.Title = string([]rune(out.Title)[:255])
	}
	md := strings.TrimSpace(optionalDraftString(obj["markdown"]))
	if md == "" {
		md = strings.TrimSpace(optionalDraftString(obj["description"]))
	}
	out.Markdown = md
	out.Description = md
	out.StartAt = strings.TrimSpace(optionalDraftString(obj["start_at"]))
	out.EndAt = strings.TrimSpace(optionalDraftString(obj["end_at"]))
	if v, ok := obj["detail"]; ok && v != nil {
		if detail, err := parseCaseTaskAIDetail(v); err == nil {
			out.Detail = detail
		}
	}
	out.Location = sanitizeCaseTaskAILocation(obj["location"])
	return out, nil
}

func optionalDraftString(v any) string {
	if v == nil {
		return ""
	}
	s := strings.TrimSpace(asString(v))
	if s == "" || strings.EqualFold(s, "null") {
		return ""
	}
	return s
}

func parseCaseTaskAIDetail(v any) (map[string]any, error) {
	if v == nil {
		return nil, errors.New("AI did not return a valid case/task draft")
	}
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return nil, errors.New("AI did not return a valid case/task draft")
	}
	if len(m) == 0 {
		return nil, errors.New("AI did not return a valid case/task draft")
	}
	return m, nil
}

func sanitizeCaseTaskAILocation(v any) *caseTaskAILocation {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" || strings.EqualFold(s, "null") {
			return nil
		}
		return &caseTaskAILocation{Label: s}
	case map[string]any:
		label := strings.TrimSpace(optionalDraftString(firstNonNil(t["label"], t["location"])))
		area := parseCaseTaskAIArea(firstNonNil(t["area"], t["points"]))
		if label == "" && len(area) == 0 {
			return nil
		}
		return &caseTaskAILocation{Label: label, Area: area}
	default:
		return nil
	}
}

func firstNonNil(vals ...any) any {
	for _, v := range vals {
		if v != nil {
			return v
		}
	}
	return nil
}

func parseCaseTaskAIArea(v any) [][]float64 {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out [][]float64
	for _, item := range arr {
		lat, lng, ok := parseLatLngPoint(item)
		if !ok {
			continue
		}
		out = append(out, []float64{lat, lng})
	}
	return out
}

func parseLatLngPoint(v any) (lat, lng float64, ok bool) {
	switch t := v.(type) {
	case []any:
		if len(t) < 2 {
			return 0, 0, false
		}
		lat, ok1 := toFiniteFloat(t[0])
		lng, ok2 := toFiniteFloat(t[1])
		if !ok1 || !ok2 {
			return 0, 0, false
		}
		return lat, lng, true
	case []float64:
		if len(t) < 2 {
			return 0, 0, false
		}
		return t[0], t[1], true
	case map[string]any:
		lat, ok1 := toFiniteFloat(firstNonNil(t["lat"], t["latitude"]))
		lng, ok2 := toFiniteFloat(firstNonNil(t["lng"], t["lon"], t["longitude"]))
		if !ok1 || !ok2 {
			return 0, 0, false
		}
		return lat, lng, true
	default:
		return 0, 0, false
	}
}

func toFiniteFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		if t != t { // NaN
			return 0, false
		}
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}
