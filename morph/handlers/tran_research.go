package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"idongivaflyinfa/ai"

	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
	"github.com/robo/morphgraph"
	"github.com/robo/webresearch"
)

const (
	researchRoundCount             = 5
	researchLegacyRounds           = 20
	researchMaxFileBytes           = 8 * 1024 * 1024
	researchMaxPromptRunes         = 20000
	researchComposePrefix          = "Compose a thesis-quality research conclusion."
	researchCriticPrefix           = "You are a ruthless editor of a research thesis."
	researchAINotConfiguredMessage = "AI is not configured (set MORPH_AI_API_KEY)"
)

var errResearchAINotConfigured = errors.New(researchAINotConfiguredMessage)

type researchDoc struct {
	ID              int             `json:"id"`
	UserID          int             `json:"user_id"`
	OwnerKey        string          `json:"owner_key"`
	Title           string          `json:"title"`
	Prompt          string          `json:"prompt"`
	Status          string          `json:"status"`
	CurrentRound    int             `json:"current_round"`
	RoundCount      int             `json:"round_count"`
	MarkdownContent string          `json:"markdown_content"`
	HTMLContent     string          `json:"html_content"`
	ErrorText       string          `json:"error_text,omitempty"`
	PublishedSlug   *string         `json:"published_slug,omitempty"`
	PublishedPath   *string         `json:"published_path,omitempty"`
	PublishedURL    string          `json:"published_url,omitempty"`
	Pieces          []researchPiece `json:"pieces,omitempty"`
	CreatedOn       time.Time       `json:"created_on"`
	LastUpdated     time.Time       `json:"last_updated"`
}

type researchPiece struct {
	RoundIndex   int    `json:"round_index"`
	Markdown     string `json:"markdown"`
	Verification string `json:"verification"`
	SourcesJSON  string `json:"sources_json,omitempty"`
	Status       string `json:"status"`
}

var (
	researchMu        sync.Mutex
	researchWebGather = webresearch.Gather
	researchLLMHook   func(ctx context.Context, prompt string) (string, error)
	researchSkipAsync bool
)

func researchFileKind(filename string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt":
		return "txt", true
	case ".pdf":
		return "pdf", true
	case ".csv":
		return "csv", true
	case ".json":
		return "json", true
	default:
		return "", false
	}
}

func extractResearchFile(filename string, raw []byte) (string, error) {
	kind, ok := researchFileKind(filename)
	if !ok {
		return "", fmt.Errorf("unsupported file type (allowed: .txt, .pdf, .csv, .json)")
	}
	if kind == "pdf" {
		text, err := morphgraph.ExtractPDFBytes(raw)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(text), nil
	}
	return strings.TrimSpace(morphgraph.ExtractPlainText(filename, "", string(raw))), nil
}

func researchTitleFromPrompt(prompt string) string {
	p := strings.TrimSpace(prompt)
	if p == "" {
		return "Research"
	}
	if utf8.RuneCountInString(p) > 80 {
		return string([]rune(p)[:80])
	}
	return p
}

func researchErrIsNotConfigured(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, errResearchAINotConfigured) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not configured")
}

func (h *Handlers) researchAIReady() bool {
	if researchLLMHook != nil {
		return true
	}
	return h != nil && h.aiService != nil && h.aiService.Configured()
}

func (h *Handlers) researchLLM(ctx context.Context, prompt string) (string, error) {
	if researchLLMHook != nil {
		out, err := researchLLMHook(ctx, prompt)
		if researchErrIsNotConfigured(err) {
			return "", errResearchAINotConfigured
		}
		return out, err
	}
	if !h.researchAIReady() {
		return "", errResearchAINotConfigured
	}
	out, err := h.aiService.ChatCompletionLong(ctx, []ai.DashScopeMessage{{Role: "user", Content: prompt}})
	if researchErrIsNotConfigured(err) {
		return "", errResearchAINotConfigured
	}
	return out, err
}

func (h *Handlers) researchOwnerKey(c *gin.Context) string {
	return h.bigNoteOwnerKey(c)
}

func scanResearch(scanner interface {
	Scan(dest ...any) error
}) (researchDoc, error) {
	var d researchDoc
	var errText, pubSlug, pubPath sql.NullString
	var createdOn, updatedOn sql.NullTime
	err := scanner.Scan(
		&d.ID, &d.UserID, &d.OwnerKey, &d.Title, &d.Prompt, &d.Status, &d.CurrentRound, &d.RoundCount,
		&d.MarkdownContent, &d.HTMLContent, &errText, &pubSlug, &pubPath,
		scanDestTime{&createdOn}, scanDestTime{&updatedOn},
	)
	if err != nil {
		return d, err
	}
	if d.RoundCount < 1 {
		d.RoundCount = researchRoundCount
	}
	if errText.Valid {
		d.ErrorText = errText.String
	}
	if createdOn.Valid {
		d.CreatedOn = createdOn.Time
	}
	if updatedOn.Valid {
		d.LastUpdated = updatedOn.Time
	}
	if pubSlug.Valid && strings.TrimSpace(pubSlug.String) != "" {
		s := pubSlug.String
		d.PublishedSlug = &s
	}
	if pubPath.Valid && strings.TrimSpace(pubPath.String) != "" {
		s := pubPath.String
		d.PublishedPath = &s
	}
	return d, nil
}

const researchSelectCols = `id, user_id, owner_key, title, prompt, status, current_round, round_target,
	markdown_content, html_content, error_text, published_slug, published_path, created_on, last_updated`

func (h *Handlers) getResearchOwned(c *gin.Context, id int) (researchDoc, error) {
	row := h.TranMySQL.DB.QueryRow(`SELECT `+researchSelectCols+` FROM research WHERE id = ?`, id)
	d, err := scanResearch(row)
	if err != nil {
		return d, err
	}
	owner := h.researchOwnerKey(c)
	if d.OwnerKey != "" && owner != "" && d.OwnerKey != owner {
		return d, sql.ErrNoRows
	}
	return d, nil
}

func loadResearchPieces(db *sql.DB, id int) ([]researchPiece, error) {
	rows, err := db.Query(
		`SELECT round_index, markdown, verification, COALESCE(sources_json,''), status
		 FROM research_piece WHERE research_id = ? ORDER BY round_index ASC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []researchPiece
	for rows.Next() {
		var p researchPiece
		if err := rows.Scan(&p.RoundIndex, &p.Markdown, &p.Verification, &p.SourcesJSON, &p.Status); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (h *Handlers) attachResearchURL(d *researchDoc) {
	if d == nil || d.PublishedPath == nil {
		return
	}
	path := strings.TrimSpace(*d.PublishedPath)
	if strings.HasPrefix(path, "/api/tran/public/research/") {
		d.PublishedURL = path
		return
	}
	d.PublishedURL = path
}

func buildResearchHTML(title, markdown string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Research"
	}
	return strings.Replace(buildTimelineHTML(title, markdown), ">Timeline<", ">Research<", 1)
}

func retrieveResearchChunks(db *sql.DB, researchID int, query string, limit int) []string {
	if limit < 1 {
		limit = 6
	}
	rows, err := db.Query(`SELECT text_content FROM research_chunk WHERE research_id = ?`, researchID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	type scored struct {
		score float64
		text  string
	}
	var ranked []scored
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil
		}
		score := morphgraph.TokenOverlapScore(query, text)
		if score <= 0 {
			continue
		}
		ranked = append(ranked, scored{score: score, text: text})
	}
	if len(ranked) == 0 {
		return nil
	}
	for i := 0; i < len(ranked); i++ {
		best := i
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].score > ranked[best].score {
				best = j
			}
		}
		ranked[i], ranked[best] = ranked[best], ranked[i]
	}
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	out := make([]string, 0, len(ranked))
	for _, s := range ranked {
		out = append(out, s.text)
	}
	return out
}

type researchUpload struct {
	name string
	raw  []byte
}

func ingestResearchFiles(db *sql.DB, researchID int, files []researchUpload) error {
	for _, f := range files {
		text, err := extractResearchFile(f.name, f.raw)
		if err != nil {
			return err
		}
		kind, _ := researchFileKind(f.name)
		res, err := db.Exec(
			`INSERT INTO research_file (research_id, filename, kind, text_excerpt) VALUES (?,?,?,?)`,
			researchID, f.name, kind, morphgraph.TruncateRunes(text, 400),
		)
		if err != nil {
			return err
		}
		fileID, _ := res.LastInsertId()
		chunks := morphgraph.ChunkText(text, morphgraph.DefaultChunkRunes, morphgraph.DefaultChunkOverlap)
		if len(chunks) == 0 && text != "" {
			chunks = []string{text}
		}
		for i, ch := range chunks {
			if _, err := db.Exec(
				`INSERT INTO research_chunk (research_id, file_id, chunk_index, text_content) VALUES (?,?,?,?)`,
				researchID, fileID, i, ch,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func persistResearchPiece(db *sql.DB, researchID, round int, markdown, verification, sourcesJSON, status string) error {
	_, err := db.Exec(
		`INSERT INTO research_piece (research_id, round_index, markdown, verification, sources_json, status)
		 VALUES (?,?,?,?,?,?)
		 ON CONFLICT(research_id, round_index) DO UPDATE SET
		   markdown=excluded.markdown, verification=excluded.verification,
		   sources_json=excluded.sources_json, status=excluded.status`,
		researchID, round, markdown, verification, sourcesJSON, status,
	)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE research SET current_round = ?, last_updated = CURRENT_TIMESTAMP WHERE id = ?`, round, researchID)
	return err
}

func researchRoundSucceeded(status string) bool {
	switch status {
	case "ok", "unverified":
		return true
	default:
		return false
	}
}

func researchPartialFailureText(failed []int, target int, sample string) string {
	parts := make([]string, len(failed))
	for i, n := range failed {
		parts[i] = fmt.Sprintf("%d", n)
	}
	msg := fmt.Sprintf("%d of %d rounds failed (%s)", len(failed), target, strings.Join(parts, ", "))
	if s := strings.TrimSpace(sample); s != "" {
		msg += ": " + s
	}
	return msg
}

func failResearchJob(db *sql.DB, id int, reason string) {
	_, _ = db.Exec(
		`UPDATE research SET status = 'failed', markdown_content = '', html_content = '', error_text = ?, last_updated = CURRENT_TIMESTAMP WHERE id = ? AND status != 'cancelled'`,
		reason, id,
	)
}

func loadResearchRoundTarget(db *sql.DB, id int) int {
	var n sql.NullInt64
	if err := db.QueryRow(`SELECT round_target FROM research WHERE id = ?`, id).Scan(&n); err != nil || !n.Valid || n.Int64 < 1 {
		return researchRoundCount
	}
	return int(n.Int64)
}

func researchSourcePack(prompt string, pieces []researchPiece) string {
	var b strings.Builder
	b.WriteString("Original prompt:\n")
	b.WriteString(strings.TrimSpace(prompt))
	b.WriteString("\n\nWorking notes from research rounds (source material, not the outline of the conclusion):\n")
	for _, p := range pieces {
		fmt.Fprintf(&b, "\n--- Round %d (%s) ---\n%s\nVerification: %s\n",
			p.RoundIndex, p.Status, strings.TrimSpace(p.Markdown), strings.TrimSpace(p.Verification))
	}
	return b.String()
}

func (h *Handlers) synthesizeResearchConclusion(ctx context.Context, prompt string, pieces []researchPiece) (string, error) {
	pack := researchSourcePack(prompt, pieces)
	composePrompt := researchComposePrefix + ` Write one professional document in the manner of a doctoral thesis or expert essay. Structure by argument, not by research-round order. Absorb the best-supported essence of every round; do not concatenate round write-ups; do not use Round 1…Round N as headings. Qualify uncertain or unverified claims. Markdown only.

` + morphai.VisualFirstInstructions + `

` + pack
	draft, err := h.researchLLM(ctx, composePrompt)
	draft = strings.TrimSpace(draft)
	if researchErrIsNotConfigured(err) {
		return "", errResearchAINotConfigured
	}
	criticPrompt := researchCriticPrefix + ` Improve the draft using the source pack. Fix structure, redundancy, holes, and voice so it reads as one finished piece. Do not reintroduce Round N as the outline. Markdown only.

` + morphai.VisualFirstInstructions + `

Source pack:
` + pack + `

Draft:
` + draft
	polished, err2 := h.researchLLM(ctx, criticPrompt)
	polished = strings.TrimSpace(polished)
	if researchErrIsNotConfigured(err2) {
		return "", errResearchAINotConfigured
	}
	if err2 == nil && polished != "" {
		return polished, nil
	}
	if err == nil && draft != "" {
		return draft, nil
	}
	if err2 != nil {
		return "", err2
	}
	if err != nil {
		return "", err
	}
	return "", errors.New("thesis synthesis returned nothing")
}

func researchJobCancelled(db *sql.DB, id int) bool {
	var status string
	if err := db.QueryRow(`SELECT status FROM research WHERE id = ?`, id).Scan(&status); err != nil {
		return false
	}
	return status == "cancelled"
}

func (h *Handlers) runResearchJob(ctx context.Context, id int) {
	db := h.TranMySQL.DB
	var prompt, title string
	if err := db.QueryRow(`SELECT prompt, title FROM research WHERE id = ?`, id).Scan(&prompt, &title); err != nil {
		return
	}
	if researchJobCancelled(db, id) {
		return
	}
	if !h.researchAIReady() {
		failResearchJob(db, id, researchAINotConfiguredMessage)
		return
	}
	_, _ = db.Exec(`UPDATE research SET status = 'running', last_updated = CURRENT_TIMESTAMP WHERE id = ? AND status != 'cancelled'`, id)

	var startRound int
	_ = db.QueryRow(`SELECT COALESCE(MAX(round_index), 0) FROM research_piece WHERE research_id = ?`, id).Scan(&startRound)
	startRound++
	target := loadResearchRoundTarget(db, id)

	configFailed := false
	failSample := ""
	for round := startRound; round <= target; round++ {
		if ctx.Err() != nil || researchJobCancelled(db, id) {
			_, _ = db.Exec(`UPDATE research SET status = 'cancelled', last_updated = CURRENT_TIMESTAMP WHERE id = ?`, id)
			return
		}
		query := strings.TrimSpace(prompt)
		if round > 1 {
			query = fmt.Sprintf("%s\n\nFollow-up angle for research round %d of %d.", prompt, round, target)
		}
		qOut, qErr := h.researchLLM(ctx, "Propose one short web search query (no quotes) for this research prompt:\n"+query)
		if researchErrIsNotConfigured(qErr) {
			_ = persistResearchPiece(db, id, round, fmt.Sprintf("Round %d failed: %s", round, researchAINotConfiguredMessage), "", "[]", "error")
			configFailed = true
			break
		}
		if qErr == nil && strings.TrimSpace(qOut) != "" {
			query = strings.Split(strings.TrimSpace(qOut), "\n")[0]
			if utf8.RuneCountInString(query) > 400 {
				query = string([]rune(query)[:400])
			}
		}
		notes, sources := researchWebGather(query)
		srcJSON, _ := json.Marshal(sources)
		ragHits := retrieveResearchChunks(db, id, query+" "+prompt, 6)
		ragBlock := strings.Join(ragHits, "\n\n")
		writePrompt := fmt.Sprintf(`Write markdown for research round %d of %d.
Topic: %s
Search query: %s
Online notes:
%s
Uploaded file excerpts:
%s
Write 2-6 short paragraphs with headings. Do not invent citations.

%s`, round, target, prompt, query, notes, ragBlock, morphai.VisualFirstInstructions)
		piece, wErr := h.researchLLM(ctx, writePrompt)
		status := "ok"
		verification := ""
		if researchErrIsNotConfigured(wErr) {
			_ = persistResearchPiece(db, id, round, fmt.Sprintf("Round %d failed: %s", round, researchAINotConfiguredMessage), "", string(srcJSON), "error")
			configFailed = true
			break
		}
		if wErr != nil {
			status = "error"
			piece = fmt.Sprintf("Round %d failed: %s", round, wErr.Error())
			if failSample == "" {
				failSample = wErr.Error()
			}
		} else if strings.TrimSpace(piece) == "" {
			piece = notes
			if piece == "" {
				piece = fmt.Sprintf("Round %d gathered no additional notes.", round)
			}
		}
		if status == "ok" {
			verifyPrompt := fmt.Sprintf(`You verify research claims. Sources/notes:\n%s\n\nDraft:\n%s\n\nWrite a short Verification subsection: what is supported, what is uncertain.`, notes, piece)
			var vErr error
			verification, vErr = h.researchLLM(ctx, verifyPrompt)
			if researchErrIsNotConfigured(vErr) {
				_ = persistResearchPiece(db, id, round, strings.TrimSpace(piece), "", string(srcJSON), "ok")
				configFailed = true
				break
			}
			if vErr != nil {
				verification = "Verification unavailable: " + vErr.Error()
				status = "unverified"
			}
		}
		_ = persistResearchPiece(db, id, round, strings.TrimSpace(piece), strings.TrimSpace(verification), string(srcJSON), status)
	}

	if researchJobCancelled(db, id) {
		return
	}
	pieces, _ := loadResearchPieces(db, id)
	var successful []researchPiece
	var failedRounds []int
	for _, p := range pieces {
		if researchRoundSucceeded(p.Status) {
			successful = append(successful, p)
			continue
		}
		if p.Status == "error" {
			failedRounds = append(failedRounds, p.RoundIndex)
		}
	}
	if configFailed {
		failResearchJob(db, id, researchAINotConfiguredMessage)
		return
	}
	if len(successful) == 0 {
		reason := "All research rounds failed"
		if failSample != "" {
			reason += ": " + failSample
		}
		failResearchJob(db, id, reason)
		return
	}
	_, _ = db.Exec(`UPDATE research SET status = 'refining', last_updated = CURRENT_TIMESTAMP WHERE id = ? AND status != 'cancelled'`, id)
	if researchJobCancelled(db, id) {
		return
	}
	synthesized, synErr := h.synthesizeResearchConclusion(ctx, prompt, successful)
	if researchJobCancelled(db, id) {
		return
	}
	if researchErrIsNotConfigured(synErr) {
		failResearchJob(db, id, researchAINotConfiguredMessage)
		return
	}
	synthesized = strings.TrimSpace(synthesized)
	if synErr != nil || synthesized == "" {
		reason := "Thesis could not be synthesized"
		if synErr != nil && strings.TrimSpace(synErr.Error()) != "" {
			reason += ": " + synErr.Error()
		}
		if len(failedRounds) > 0 {
			reason += ". " + researchPartialFailureText(failedRounds, target, failSample)
		}
		failResearchJob(db, id, reason)
		return
	}
	errText := ""
	if len(failedRounds) > 0 {
		errText = researchPartialFailureText(failedRounds, target, failSample)
	}
	htmlOut := buildResearchHTML(title, synthesized)
	_, _ = db.Exec(
		`UPDATE research SET status = 'complete', markdown_content = ?, html_content = ?, error_text = ?, current_round = ?, last_updated = CURRENT_TIMESTAMP WHERE id = ? AND status != 'cancelled'`,
		synthesized, htmlOut, errText, target, id,
	)
}

func (h *Handlers) startResearchJob(id int) {
	if researchSkipAsync {
		return
	}
	go func() {
		researchMu.Lock()
		defer researchMu.Unlock()
		h.runResearchJob(context.Background(), id)
	}()
}

// StartResearchResume continues unfinished research jobs after process restart.
func (h *Handlers) StartResearchResume(ctx context.Context) {
	if h == nil || h.TranMySQL == nil || h.TranMySQL.DB == nil {
		return
	}
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(400 * time.Millisecond):
		}
		rows, err := h.TranMySQL.DB.Query(
			`SELECT id FROM research WHERE status IN ('ingesting','running','refining') ORDER BY id ASC`)
		if err != nil {
			return
		}
		defer rows.Close()
		var ids []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err == nil {
				ids = append(ids, id)
			}
		}
		for _, id := range ids {
			if ctx.Err() != nil {
				return
			}
			h.startResearchJob(id)
		}
	}()
}

// ListResearch GET /api/tran/research
func (h *Handlers) ListResearch(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	owner := h.researchOwnerKey(c)
	rows, err := h.TranMySQL.DB.Query(
		`SELECT `+researchSelectCols+` FROM research WHERE owner_key = ? OR owner_key = '' ORDER BY last_updated DESC LIMIT 200`,
		owner,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := make([]researchDoc, 0)
	for rows.Next() {
		d, err := scanResearch(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.attachResearchURL(&d)
		out = append(out, d)
	}
	c.JSON(http.StatusOK, out)
}

// GetResearch GET /api/tran/research/:id
func (h *Handlers) GetResearch(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	d, err := h.getResearchOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "research not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	d.Pieces, _ = loadResearchPieces(h.TranMySQL.DB, id)
	h.attachResearchURL(&d)
	c.JSON(http.StatusOK, d)
}

// CreateResearch POST /api/tran/research
func (h *Handlers) CreateResearch(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var prompt string
	ct := strings.ToLower(c.ContentType())
	if strings.HasPrefix(ct, "application/json") {
		var in struct {
			Prompt string `json:"prompt"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}
		prompt = strings.TrimSpace(in.Prompt)
	} else {
		_ = c.Request.ParseMultipartForm(researchMaxFileBytes + (1 << 20))
		prompt = strings.TrimSpace(firstNonEmpty(c.PostForm("prompt"), c.Request.FormValue("prompt")))
	}
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	if utf8.RuneCountInString(prompt) > researchMaxPromptRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt too long"})
		return
	}

	var files []researchUpload
	if c.Request.MultipartForm != nil {
		for _, hdrs := range c.Request.MultipartForm.File {
			for _, hdr := range hdrs {
				if _, ok := researchFileKind(hdr.Filename); !ok {
					c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type (allowed: .txt, .pdf, .csv, .json)"})
					return
				}
				if hdr.Size > researchMaxFileBytes {
					c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
					return
				}
				f, err := hdr.Open()
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file"})
					return
				}
				raw, err := io.ReadAll(io.LimitReader(f, researchMaxFileBytes+1))
				_ = f.Close()
				if err != nil || int64(len(raw)) > researchMaxFileBytes {
					c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file"})
					return
				}
				files = append(files, researchUpload{name: filepath.Base(hdr.Filename), raw: raw})
			}
		}
	}

	if !h.researchAIReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": researchAINotConfiguredMessage})
		return
	}

	owner := h.researchOwnerKey(c)
	userID := h.tranUserIDFromContext(c)
	title := researchTitleFromPrompt(prompt)
	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO research (user_id, owner_key, title, prompt, status, current_round, round_target) VALUES (?,?,?,?, 'ingesting', 0, ?)`,
		userID, owner, title, prompt, researchRoundCount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	id := int(id64)
	if err := ingestResearchFiles(h.TranMySQL.DB, id, files); err != nil {
		_, _ = h.TranMySQL.DB.Exec(`DELETE FROM research WHERE id = ?`, id)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, _ = h.TranMySQL.DB.Exec(`UPDATE research SET status = 'running', last_updated = CURRENT_TIMESTAMP WHERE id = ?`, id)
	h.startResearchJob(id)
	d, err := h.getResearchOwned(c, id)
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{"id": id, "title": title, "status": "running", "prompt": prompt})
		return
	}
	h.attachResearchURL(&d)
	c.JSON(http.StatusCreated, d)
}

// PatchResearchMarkdown PATCH /api/tran/research/:id
func (h *Handlers) PatchResearchMarkdown(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	d, err := h.getResearchOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "research not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var in struct {
		MarkdownContent string `json:"markdown_content"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	htmlOut := buildResearchHTML(d.Title, in.MarkdownContent)
	_, err = h.TranMySQL.DB.Exec(
		`UPDATE research SET markdown_content = ?, html_content = ?, last_updated = CURRENT_TIMESTAMP WHERE id = ?`,
		in.MarkdownContent, htmlOut, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	d.MarkdownContent = in.MarkdownContent
	d.HTMLContent = htmlOut
	d.Pieces, _ = loadResearchPieces(h.TranMySQL.DB, id)
	h.attachResearchURL(&d)
	c.JSON(http.StatusOK, d)
}

// CancelResearch POST /api/tran/research/:id/cancel
func (h *Handlers) CancelResearch(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if _, err := h.getResearchOwned(c, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "research not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, _ = h.TranMySQL.DB.Exec(`UPDATE research SET status = 'cancelled', last_updated = CURRENT_TIMESTAMP WHERE id = ?`, id)
	d, _ := h.getResearchOwned(c, id)
	d.Pieces, _ = loadResearchPieces(h.TranMySQL.DB, id)
	c.JSON(http.StatusOK, d)
}

// PublishResearch POST /api/tran/research/:id/publish
func (h *Handlers) PublishResearch(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	d, err := h.getResearchOwned(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "research not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if d.Status == "failed" {
		msg := strings.TrimSpace(d.ErrorText)
		if msg == "" {
			msg = "research failed and cannot be published"
		}
		c.JSON(http.StatusConflict, gin.H{"error": msg})
		return
	}
	if strings.TrimSpace(d.HTMLContent) == "" {
		d.HTMLContent = buildResearchHTML(d.Title, d.MarkdownContent)
	}
	slug := ""
	if d.PublishedSlug != nil {
		slug = strings.TrimSpace(*d.PublishedSlug)
	}
	if slug == "" {
		slug = slugifyTimeline(d.Title)
		if slug == "" {
			slug = fmt.Sprintf("research-%d", id)
		}
		slug = fmt.Sprintf("%s-%d", slug, id)
	}
	path := "/api/tran/public/research/" + slug
	_, err = h.TranMySQL.DB.Exec(
		`UPDATE research SET published_slug = ?, published_path = ?, html_content = ?, last_updated = CURRENT_TIMESTAMP WHERE id = ?`,
		slug, path, d.HTMLContent, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.getResearchOwned(c, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id, "published_path": path})
		return
	}
	updated.Pieces, _ = loadResearchPieces(h.TranMySQL.DB, id)
	h.attachResearchURL(&updated)
	c.JSON(http.StatusOK, updated)
}

// ServePublicResearch GET /api/tran/public/research/:slug
func (h *Handlers) ServePublicResearch(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var title, markdown, htmlContent string
	err := h.TranMySQL.DB.QueryRow(
		`SELECT title, markdown_content, html_content FROM research WHERE published_slug = ? LIMIT 1`, slug,
	).Scan(&title, &markdown, &htmlContent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "published research not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	htmlOut := strings.TrimSpace(htmlContent)
	if htmlOut == "" {
		htmlOut = buildResearchHTML(title, markdown)
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlOut))
}

// DeleteResearch DELETE /api/tran/research/:id
func (h *Handlers) DeleteResearch(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if _, err := h.getResearchOwned(c, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "research not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, _ = h.TranMySQL.DB.Exec(`DELETE FROM research_piece WHERE research_id = ?`, id)
	_, _ = h.TranMySQL.DB.Exec(`DELETE FROM research_chunk WHERE research_id = ?`, id)
	_, _ = h.TranMySQL.DB.Exec(`DELETE FROM research_file WHERE research_id = ?`, id)
	if _, err := h.TranMySQL.DB.Exec(`DELETE FROM research WHERE id = ?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "id": id})
}
