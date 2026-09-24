package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

const maxSubAgents = 3

func boolFlag(p *bool, defaultVal bool) bool {
	if p == nil {
		return defaultVal
	}
	return *p
}

// ContextFingerprint identifies an include set so unchanged pins can reuse a cache blob.
func ContextFingerprint(includeFiles, includeNotes, includeKnowledge bool, pinHashes []string) string {
	pins := append([]string{}, pinHashes...)
	sort.Strings(pins)
	parts := []string{
		bool01(includeFiles),
		bool01(includeNotes),
		bool01(includeKnowledge),
		strings.Join(pins, ","),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func bool01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func pinHashesFrom(files []models.AgentPinnedFile) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		h := strings.TrimSpace(f.Hash)
		if h == "" {
			continue
		}
		out = append(out, h)
	}
	return out
}

// AssemblePinnedBlob concatenates pinned file bodies. Missing content yields empty for that file.
func AssemblePinnedBlob(files []models.AgentPinnedFile) string {
	var b strings.Builder
	for _, f := range files {
		body := strings.TrimSpace(f.Content)
		if body == "" {
			continue
		}
		path := strings.TrimSpace(f.Path)
		if path == "" {
			path = "file"
		}
		b.WriteString("### ")
		b.WriteString(path)
		b.WriteString("\n")
		b.WriteString(body)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

// RouteSubAgents picks workers from the user task. Never includes a disk-write worker.
func RouteSubAgents(msg string, hasFiles, hasNotes, hasKnowledge bool) []string {
	text := strings.TrimSpace(msg)
	if text == "" {
		return nil
	}
	if isSimpleChat(text) {
		return nil
	}
	low := strings.ToLower(text)
	var workers []string
	if hasFiles && (strings.Contains(low, "file") || strings.Contains(low, "folder") || strings.Contains(low, "workspace") || looksLikeDocQuestion(low)) {
		workers = append(workers, "files")
	}
	if hasKnowledge && (strings.Contains(low, "knowledge") || strings.Contains(low, "hybrid") || strings.Contains(low, "library") || looksLikeDocQuestion(low)) {
		workers = append(workers, "knowledge")
	}
	if hasNotes && (strings.Contains(low, "note") || strings.Contains(low, "todo") || strings.Contains(low, "task")) {
		workers = append(workers, "notes")
	}
	if looksLikeMorphData(low) {
		workers = append(workers, "morphdata")
	}
	if len(workers) == 0 && (hasFiles || hasKnowledge) && len([]rune(text)) > 40 {
		if hasFiles {
			workers = append(workers, "files")
		}
		if hasKnowledge && len(workers) < maxSubAgents {
			workers = append(workers, "knowledge")
		}
	}
	if len(workers) > maxSubAgents {
		workers = workers[:maxSubAgents]
	}
	return workers
}

func isSimpleChat(text string) bool {
	trimmed := strings.TrimSpace(text)
	if len([]rune(trimmed)) <= 24 && !strings.ContainsAny(trimmed, "?") {
		words := strings.Fields(trimmed)
		if len(words) <= 4 {
			return true
		}
	}
	low := strings.ToLower(trimmed)
	simple := []string{"hi", "hello", "hey", "thanks", "thank you", "ok", "okay", "yes", "no"}
	for _, s := range simple {
		if low == s {
			return true
		}
	}
	return false
}

func looksLikeMorphData(low string) bool {
	keys := []string{"member", "employee", "facility", "district", "case", "asset", "contact", "form", "event", "list my", "show my"}
	for _, k := range keys {
		if strings.Contains(low, k) {
			return true
		}
	}
	return false
}

func looksLikeDocQuestion(low string) bool {
	return strings.Contains(low, "summar") || strings.Contains(low, "what does") || strings.Contains(low, "in this") || strings.Contains(low, "according to")
}

func wrapAgentContext(userMessage, block, label string) string {
	block = strings.TrimSpace(block)
	if block == "" {
		return userMessage
	}
	return "[" + label + "]\n" + block + "\n--- End " + label + " ---\n\nUser message:\n" + userMessage
}

type agentCacheEntry struct {
	fingerprint string
	blob        string
}

var (
	agentCacheMu  sync.Mutex
	agentCacheMem = map[string]agentCacheEntry{}
)

func agentMemKey(userID, sessionID string) string {
	return strings.TrimSpace(userID) + "\x00" + strings.TrimSpace(sessionID)
}

func (h *Handlers) getAgentCache(userID, sessionID, fingerprint string) (string, bool) {
	if h != nil && h.TranMySQL != nil {
		if blob, ok := h.TranMySQL.GetAgentContextCache(userID, sessionID, fingerprint); ok {
			return blob, true
		}
	}
	agentCacheMu.Lock()
	defer agentCacheMu.Unlock()
	ent, ok := agentCacheMem[agentMemKey(userID, sessionID)]
	if !ok || ent.fingerprint != fingerprint || strings.TrimSpace(ent.blob) == "" {
		return "", false
	}
	return ent.blob, true
}

func (h *Handlers) putAgentCache(userID, sessionID, fingerprint, blob string) {
	if strings.TrimSpace(blob) == "" {
		return
	}
	if h != nil && h.TranMySQL != nil {
		_ = h.TranMySQL.PutAgentContextCache(userID, sessionID, fingerprint, blob)
	}
	agentCacheMu.Lock()
	agentCacheMem[agentMemKey(userID, sessionID)] = agentCacheEntry{fingerprint: fingerprint, blob: blob}
	agentCacheMu.Unlock()
}

func (h *Handlers) notesSnippet(userID string) string {
	if h == nil || h.TranMySQL == nil || h.TranMySQL.DB == nil {
		return ""
	}
	rows, err := h.TranMySQL.DB.Query(
		`SELECT ItemType, Title, Body, Completed FROM user_note_todo WHERE UserID = ? ORDER BY LastUpdated DESC LIMIT 40`,
		strings.TrimSpace(userID),
	)
	if err != nil {
		return ""
	}
	defer rows.Close()
	var b strings.Builder
	for rows.Next() {
		var itemType string
		var title, body sql.NullString
		var completed bool
		if err := rows.Scan(&itemType, &title, &body, &completed); err != nil {
			continue
		}
		b.WriteString("- [")
		b.WriteString(itemType)
		if completed {
			b.WriteString(" done")
		}
		b.WriteString("] ")
		if title.Valid {
			b.WriteString(strings.TrimSpace(title.String))
		}
		if body.Valid && strings.TrimSpace(body.String) != "" {
			b.WriteString(" — ")
			b.WriteString(truncateRunes(strings.TrimSpace(body.String), 240))
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

type agentApplyResult struct {
	message          string
	subAgents        []string
	includeKnowledge bool
	hasDocuments     bool
}

func (h *Handlers) applyAgentContext(userID, sessionID string, req *models.ChatRequest) agentApplyResult {
	includeFiles := boolFlag(req.IncludeFiles, true)
	includeNotes := boolFlag(req.IncludeNotes, true)
	includeKnowledge := boolFlag(req.IncludeKnowledge, true)

	hashes := pinHashesFrom(req.PinnedFiles)
	fp := ContextFingerprint(includeFiles, includeNotes, includeKnowledge, hashes)

	hasPinned := false
	for _, f := range req.PinnedFiles {
		if strings.TrimSpace(f.Hash) != "" || strings.TrimSpace(f.Content) != "" {
			hasPinned = true
			break
		}
	}
	workers := RouteSubAgents(req.Message, includeFiles && hasPinned, includeNotes, includeKnowledge)

	var filesBlob, notesBlob string
	var wg sync.WaitGroup
	if includeFiles {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if blob, ok := h.getAgentCache(userID, sessionID, fp); ok {
				filesBlob = blob
				return
			}
			blob := AssemblePinnedBlob(req.PinnedFiles)
			if blob != "" {
				filesBlob = blob
				h.putAgentCache(userID, sessionID, fp, blob)
			}
		}()
	}
	if includeNotes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			notesBlob = h.notesSnippet(userID)
		}()
	}
	wg.Wait()

	msg := req.Message
	if filesBlob != "" {
		msg = wrapAgentContext(msg, filesBlob, "Pinned workspace files")
	}
	if notesBlob != "" {
		msg = wrapAgentContext(msg, notesBlob, "Notes & TODOs")
	}
	if len(workers) > 0 {
		msg = wrapAgentContext(msg, "Collect from these workers concurrently, then write one reply. Never create, modify, or delete files in the user's local workspace folder: "+strings.Join(workers, ", ")+".", "Sub-agents")
	}
	hasDocuments := filesBlob != "" || notesBlob != ""
	if !hasDocuments && h.hybridStore != nil && h.hybridStore.IsAttached(userID, sessionID) {
		hasDocuments = true
	}
	return agentApplyResult{message: msg, subAgents: workers, includeKnowledge: includeKnowledge, hasDocuments: hasDocuments}
}

func parseOptionalBoolForm(c *gin.Context, key string) *bool {
	v := strings.TrimSpace(c.PostForm(key))
	if v == "" {
		return nil
	}
	on := v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes") || strings.EqualFold(v, "on")
	return &on
}

func parsePinnedFilesForm(c *gin.Context) []models.AgentPinnedFile {
	raw := strings.TrimSpace(c.PostForm("pinned_files"))
	if raw == "" {
		return nil
	}
	var files []models.AgentPinnedFile
	if err := json.Unmarshal([]byte(raw), &files); err != nil {
		return nil
	}
	return files
}
