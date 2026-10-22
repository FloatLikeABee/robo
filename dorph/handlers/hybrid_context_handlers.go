package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"

	"idongivaflyinfa/hybridcontext"

	"github.com/gin-gonic/gin"
)

const hybridExtractPrompt = "Transcribe all readable text from this document. Preserve structure (headings, lists, tables as plain text). Do not summarize or comment."

const hybridContextSystemPrompt = `You have HybridContext reference material attached to this chat (uploaded files and pasted notes). Use it when the user asks about that material or when their question clearly relates to it. Prefer live API tools when they conflict with static context. Mention source labels when helpful.`

func hybridContextJSON(h *Handlers, userID, sessionID string) gin.H {
	attached := false
	title := ""
	var sources []hybridcontext.SourceSummary
	if h.hybridStore != nil {
		sources = h.hybridStore.Summary(userID, sessionID)
		attached = h.hybridStore.IsAttached(userID, sessionID)
		if attached {
			title = h.hybridStore.AttachmentTitle(userID, sessionID)
		}
	}
	if sources == nil {
		sources = []hybridcontext.SourceSummary{}
	}
	return gin.H{
		"session_id":       sessionID,
		"sources":          sources,
		"attached":         attached,
		"attachment_title": title,
	}
}

func userReferencesHybridContext(userMessage string, sources []hybridcontext.SourceSummary) bool {
	lower := strings.ToLower(strings.TrimSpace(userMessage))
	if lower == "" {
		return false
	}
	triggers := []string{
		"hybrid context", "hybridcontext", "the context", "attached context", "reference material",
		"uploaded file", "uploaded files", "the file", "my files", "data file", "data files",
		"pasted notes", "the paste", "imported schema", "imported data", "from the context",
		"that document", "the document", "use the context", "based on the context",
		"what i uploaded", "what i pasted", "what i imported",
	}
	for _, t := range triggers {
		if strings.Contains(lower, t) {
			return true
		}
	}
	for _, src := range sources {
		label := strings.ToLower(strings.TrimSpace(src.Label))
		if label == "" {
			continue
		}
		if len(label) >= 4 && strings.Contains(lower, label) {
			return true
		}
		if src.Kind == string(hybridcontext.SourceFile) {
			base := strings.ToLower(path.Base(label))
			if len(base) >= 4 && strings.Contains(lower, base) {
				return true
			}
		}
	}
	return false
}

func (h *Handlers) hybridUserSession(c *gin.Context) (userID, sessionID string) {
	userID = strings.TrimSpace(c.GetHeader("X-User-ID"))
	if userID == "" {
		userID = "admin"
	}
	sessionID = resolveSessionID(c.Query("session_id"))
	return userID, sessionID
}

// GetHybridContext returns HybridContext sources for the chat session (?session_id=).
func (h *Handlers) GetHybridContext(c *gin.Context) {
	userID, sessionID := h.hybridUserSession(c)
	c.JSON(http.StatusOK, hybridContextJSON(h, userID, sessionID))
}

func (h *Handlers) DeleteHybridContext(c *gin.Context) {
	userID, sessionID := h.hybridUserSession(c)
	if h.hybridStore != nil {
		h.hybridStore.Clear(userID, sessionID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handlers) DeleteHybridContextSource(c *gin.Context) {
	var body struct {
		SessionID string `json:"session_id"`
		Kind      string `json:"kind"`
		Label     string `json:"label"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if userID == "" {
		userID = "admin"
	}
	sessionID := resolveSessionID(body.SessionID)
	kind := hybridcontext.SourceKind(strings.TrimSpace(body.Kind))
	label := strings.TrimSpace(body.Label)
	if kind == "" || label == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind and label required"})
		return
	}
	if h.hybridStore == nil {
		c.JSON(http.StatusOK, gin.H{"ok": true, "removed": 0, "sources": []any{}})
		return
	}
	removed := h.hybridStore.RemoveSource(userID, sessionID, kind, label)
	if removed == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		return
	}
	log.Printf("[HYBRID] remove source user=%s session=%s kind=%s label=%s chunks=%d", userID, sessionID, kind, label, removed)
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"removed": removed,
		"sources": h.hybridStore.Summary(userID, sessionID),
	})
}

// HybridBringToConversation attaches HybridContext as a chat reference (not a visible user message).
func (h *Handlers) HybridBringToConversation(c *gin.Context) {
	var body struct {
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if userID == "" {
		userID = "admin"
	}
	sessionID := resolveSessionID(body.SessionID)
	if h.hybridStore == nil || !h.hybridStore.HasContent(userID, sessionID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HybridContext is empty — add files or paste notes first."})
		return
	}
	if !h.hybridStore.SetAttached(userID, sessionID, true) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HybridContext is empty"})
		return
	}
	title := h.hybridStore.AttachmentTitle(userID, sessionID)
	log.Printf("[HYBRID] attach to conversation user=%s session=%s title=%q", userID, sessionID, title)
	resp := hybridContextJSON(h, userID, sessionID)
	resp["ok"] = true
	c.JSON(http.StatusOK, resp)
}

func (h *Handlers) HybridDetachFromConversation(c *gin.Context) {
	var body struct {
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if userID == "" {
		userID = "admin"
	}
	sessionID := resolveSessionID(body.SessionID)
	if h.hybridStore != nil {
		h.hybridStore.SetAttached(userID, sessionID, false)
	}
	resp := hybridContextJSON(h, userID, sessionID)
	resp["ok"] = true
	c.JSON(http.StatusOK, resp)
}

type hybridPasteBody struct {
	SessionID string `json:"session_id"`
	Text      string `json:"text"`
}

func (h *Handlers) HybridContextPaste(c *gin.Context) {
	var body hybridPasteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if userID == "" {
		userID = "admin"
	}
	sessionID := resolveSessionID(body.SessionID)
	txt := strings.TrimSpace(body.Text)
	if len([]rune(txt)) < 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text must be at least 50 characters"})
		return
	}
	n := h.hybridStore.AddDocument(userID, sessionID, hybridcontext.SourcePaste, "Pasted notes", txt)
	log.Printf("[HYBRID] user=%s session=%s paste chunks=%d", userID, sessionID, n)
	c.JSON(http.StatusOK, gin.H{"chunks_added": n, "sources": h.hybridStore.Summary(userID, sessionID)})
}

func (h *Handlers) HybridContextUploadFiles(c *gin.Context) {
	userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if userID == "" {
		userID = "admin"
	}
	sessionID := resolveSessionID(c.PostForm("session_id"))
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected multipart form with files"})
		return
	}
	headers := form.File["files"]
	if len(headers) == 0 {
		if _, ok := form.File["file"]; ok {
			headers = form.File["file"]
		}
	}
	if len(headers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files (use field name files)"})
		return
	}
	indexToGraph := parseFormBool(c, "index_to_graph", false)
	var processed []string
	var graphSaved []string
	var firstErr string
	for _, fh := range headers {
		label := fh.Filename
		f, err := fh.Open()
		if err != nil {
			firstErr = err.Error()
			continue
		}
		raw, err := io.ReadAll(io.LimitReader(f, 40<<20))
		f.Close()
		if err != nil {
			firstErr = err.Error()
			continue
		}
		text, err := h.bytesToHybridText(label, raw)
		if err != nil {
			firstErr = err.Error()
			log.Printf("[HYBRID] file %s: %v", label, err)
			continue
		}
		n := h.hybridStore.AddDocument(userID, sessionID, hybridcontext.SourceFile, label, text)
		processed = append(processed, fmt.Sprintf("%s (%d chunks)", label, n))
		if indexToGraph {
			ct := fh.Header.Get("Content-Type")
			if ct == "" {
				ct = "application/octet-stream"
			}
			rec, warn, err := h.saveKnowledgeFromBytes(c, label, ct, label, raw, true)
			if err != nil {
				log.Printf("[HYBRID] graph save %s: %v", label, err)
				continue
			}
			msg := label
			if rec != nil {
				msg = fmt.Sprintf("%s (knowledge #%d)", label, rec.ID)
			}
			if warn != "" {
				msg += " — " + warn
			}
			graphSaved = append(graphSaved, msg)
		}
	}
	if len(processed) == 0 {
		msg := "could not process files"
		if firstErr != "" {
			msg += ": " + firstErr
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"processed":      processed,
		"index_to_graph": indexToGraph,
		"graph_saved":    graphSaved,
		"sources":        h.hybridStore.Summary(userID, sessionID),
	})
}

func (h *Handlers) bytesToHybridText(filename string, raw []byte) (string, error) {
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".txt", ".csv", ".md", ".markdown", ".html", ".htm", ".json":
		s := strings.ToValidUTF8(string(raw), "")
		if ext == ".json" {
			var buf bytes.Buffer
			if json.Valid(raw) {
				_ = json.Indent(&buf, raw, "", "  ")
				if buf.Len() > 0 {
					return buf.String(), nil
				}
			}
		}
		return s, nil
	case ".xlsx", ".xlsm":
		return hybridcontext.XLSXFromBytes(raw)
	case ".pdf":
		r := bytes.NewReader(raw)
		extText, aiText, err := h.ReadPDFAndProcess(r, filename, hybridExtractPrompt)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(extText) != "" {
			return extText, nil
		}
		return aiText, nil
	case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".heic", ".heif":
		r := bytes.NewReader(raw)
		extText, aiText, err := h.ReadImageAndProcess(r, filename, hybridExtractPrompt)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(extText) != "" {
			return extText, nil
		}
		return aiText, nil
	case ".xls":
		return "", fmt.Errorf("legacy .xls is not supported; save as .xlsx or .csv")
	default:
		s := strings.ToValidUTF8(string(raw), "")
		if looksBinary(raw) {
			return "", fmt.Errorf("unsupported or binary file type %s", ext)
		}
		return s, nil
	}
}

func looksBinary(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	n := len(b)
	if n > 8000 {
		n = 8000
	}
	zeros := 0
	for i := 0; i < n; i++ {
		if b[i] == 0 {
			zeros++
		}
	}
	return zeros > 8
}

// hybridAugmentMessage injects HybridContext excerpts when attached and the user references them.
func (h *Handlers) hybridAugmentMessage(userID, sessionID, userMessage string) string {
	if h.hybridStore == nil || !h.hybridStore.IsAttached(userID, sessionID) {
		return userMessage
	}
	sources := h.hybridStore.Summary(userID, sessionID)
	explicit := userReferencesHybridContext(userMessage, sources)
	block := strings.TrimSpace(h.hybridStore.RetrieveIfRelevant(userID, sessionID, userMessage, 14000, explicit))
	if block == "" {
		return userMessage
	}
	return hybridContextSystemPrompt + "\n\n[HybridContext excerpts]\n" + block + "\n\n--- End HybridContext ---\n\nUser message:\n" + userMessage
}
