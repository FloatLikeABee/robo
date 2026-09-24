package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"idongivaflyinfa/ai"
	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
	"github.com/robo/webresearch"
	_ "modernc.org/sqlite"
)

func openResearchDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:research-"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	stmts := []string{
		`CREATE TABLE research (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			owner_key TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL,
			prompt TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'ingesting',
			current_round INTEGER NOT NULL DEFAULT 0,
			round_target INTEGER NOT NULL DEFAULT 5,
			markdown_content TEXT NOT NULL DEFAULT '',
			html_content TEXT NOT NULL DEFAULT '',
			error_text TEXT NULL,
			published_slug TEXT NULL,
			published_path TEXT NULL,
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE research_file (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			research_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			kind TEXT NOT NULL,
			text_excerpt TEXT NULL
		)`,
		`CREATE TABLE research_chunk (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			research_id INTEGER NOT NULL,
			file_id INTEGER NOT NULL DEFAULT 0,
			chunk_index INTEGER NOT NULL,
			text_content TEXT NOT NULL,
			embedding_json TEXT NULL
		)`,
		`CREATE TABLE research_piece (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			research_id INTEGER NOT NULL,
			round_index INTEGER NOT NULL,
			markdown TEXT NOT NULL DEFAULT '',
			verification TEXT NOT NULL DEFAULT '',
			sources_json TEXT NULL,
			status TEXT NOT NULL DEFAULT 'ok',
			created_on TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(research_id, round_index)
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func researchRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: db}}
	r := gin.New()
	r.POST("/api/tran/research", h.CreateResearch)
	r.GET("/api/tran/research/:id", h.GetResearch)
	r.PATCH("/api/tran/research/:id", h.PatchResearchMarkdown)
	r.POST("/api/tran/research/:id/publish", h.PublishResearch)
	r.GET("/api/tran/public/research/:slug", h.ServePublicResearch)
	return r
}

func TestCreateResearchEmptyPromptRejected(t *testing.T) {
	db := openResearchDB(t)
	r := researchRouter(db)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/research", strings.NewReader(`{"prompt":"  "}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM research`).Scan(&n)
	if n != 0 {
		t.Fatalf("expected no row, got %d", n)
	}
}

func TestCreateResearchRejectsWhenAIUnconfigured(t *testing.T) {
	prev := researchLLMHook
	researchLLMHook = nil
	t.Cleanup(func() { researchLLMHook = prev })
	db := openResearchDB(t)
	r := researchRouter(db)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/research", strings.NewReader(`{"prompt":"What is graphene?"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "MORPH_AI_API_KEY") {
		t.Fatalf("body %s", w.Body.String())
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM research`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected no row, got %d", n)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/tran/research", strings.NewReader(`{"prompt":"  "}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty prompt status %d body %s", w.Code, w.Body.String())
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("prompt", "topic")
	part, err := mw.CreateFormFile("file", "photo.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("not-an-image"))
	_ = mw.Close()
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/tran/research", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad file status %d body %s", w.Code, w.Body.String())
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM research`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected no row after invalid creates, got %d", n)
	}
}

func TestPublishFailedResearchRejected(t *testing.T) {
	db := openResearchDB(t)
	_, err := db.Exec(
		`INSERT INTO research (user_id, owner_key, title, prompt, status, error_text, markdown_content) VALUES (1,'','T','p','failed',?,'')`,
		researchAINotConfiguredMessage,
	)
	if err != nil {
		t.Fatal(err)
	}
	r := researchRouter(db)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/research/1/publish", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var slug sql.NullString
	if err := db.QueryRow(`SELECT published_slug FROM research WHERE id = 1`).Scan(&slug); err != nil {
		t.Fatal(err)
	}
	if slug.Valid && strings.TrimSpace(slug.String) != "" {
		t.Fatalf("published_slug %q", slug.String)
	}
}

func TestCreateResearchPromptOnly(t *testing.T) {
	researchSkipAsync = true
	t.Cleanup(func() { researchSkipAsync = false })
	stubResearchAIReady(t)
	db := openResearchDB(t)
	r := researchRouter(db)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/research", strings.NewReader(`{"prompt":"What is graphene?"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM research`).Scan(&n)
	if n != 1 {
		t.Fatalf("expected 1 row, got %d", n)
	}
	var created researchDoc
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.RoundCount != researchRoundCount {
		t.Fatalf("round_count %d, want %d", created.RoundCount, researchRoundCount)
	}
	var target int
	if err := db.QueryRow(`SELECT round_target FROM research WHERE id = ?`, created.ID).Scan(&target); err != nil {
		t.Fatal(err)
	}
	if target != researchRoundCount {
		t.Fatalf("round_target %d, want %d", target, researchRoundCount)
	}
}

func TestCreateResearchRejectsBadFileType(t *testing.T) {
	db := openResearchDB(t)
	r := researchRouter(db)
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("prompt", "topic")
	part, err := mw.CreateFormFile("file", "photo.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("not-an-image"))
	_ = mw.Close()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/research", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM research`).Scan(&n)
	if n != 0 {
		t.Fatalf("expected no row after bad file, got %d", n)
	}
}

func stubResearchPipeline(t *testing.T) {
	t.Helper()
	prevGather := researchWebGather
	prevLLM := researchLLMHook
	t.Cleanup(func() {
		researchWebGather = prevGather
		researchLLMHook = prevLLM
	})
	researchWebGather = func(query string) (string, []webresearch.Source) {
		return "online note about " + query, []webresearch.Source{{Title: "Wiki", Type: "wiki"}}
	}
	researchLLMHook = func(_ context.Context, prompt string) (string, error) {
		if strings.HasPrefix(prompt, researchComposePrefix) {
			return "# Graphene in energy storage\n\nA unified thesis on conductivity and electrodes.", nil
		}
		if strings.HasPrefix(prompt, researchCriticPrefix) {
			return "# Graphene in energy storage\n\nPolished thesis on conductivity, electrodes, and remaining uncertainty.", nil
		}
		if strings.Contains(prompt, "Propose one short") {
			return "graphene", nil
		}
		if strings.HasPrefix(prompt, "You verify research claims") {
			return "Claims match the notes.", nil
		}
		return "Round draft with **bold**.", nil
	}
}

func TestResearchIngestChunksThenStubbedFiveRounds(t *testing.T) {
	sqlDB := openResearchDB(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	_, err := sqlDB.Exec(`INSERT INTO research (user_id, owner_key, title, prompt, status) VALUES (1,'','T','graphene batteries','running')`)
	if err != nil {
		t.Fatal(err)
	}
	if err := ingestResearchFiles(sqlDB, 1, []researchUpload{{name: "notes.txt", raw: []byte("Graphene has high conductivity.")}}); err != nil {
		t.Fatal(err)
	}
	var chunks int
	_ = sqlDB.QueryRow(`SELECT COUNT(*) FROM research_chunk WHERE research_id = 1`).Scan(&chunks)
	if chunks < 1 {
		t.Fatal("expected RAG chunks")
	}

	stubResearchPipeline(t)
	h.runResearchJob(context.Background(), 1)
	pieces, err := loadResearchPieces(sqlDB, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(pieces) != researchRoundCount {
		t.Fatalf("got %d pieces, want %d", len(pieces), researchRoundCount)
	}
	for i, p := range pieces {
		if p.RoundIndex != i+1 {
			t.Fatalf("piece order %d index %d", i, p.RoundIndex)
		}
		if strings.TrimSpace(p.Verification) == "" {
			t.Fatalf("round %d missing verification", p.RoundIndex)
		}
	}
	var status, md, html string
	var roundCount, currentRound int
	if err := sqlDB.QueryRow(`SELECT status, markdown_content, html_content, round_target, current_round FROM research WHERE id = 1`).Scan(&status, &md, &html, &roundCount, &currentRound); err != nil {
		t.Fatal(err)
	}
	if status != "complete" {
		t.Fatalf("status %s", status)
	}
	if roundCount != researchRoundCount || currentRound != researchRoundCount {
		t.Fatalf("rounds stored %d/%d, want %d/%d", currentRound, roundCount, researchRoundCount, researchRoundCount)
	}
	if strings.Contains(md, "## Round ") {
		t.Fatalf("conclusion used round-order outline: %s", truncateRunes(md, 400))
	}
	if !strings.Contains(md, "Polished thesis") {
		t.Fatalf("conclusion missing critic output: %s", truncateRunes(md, 400))
	}
	if !strings.Contains(html, "Research") {
		t.Fatalf("html missing research chrome")
	}

	gin.SetMode(gin.TestMode)
	r := researchRouter(sqlDB)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tran/research/1", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get %d %s", w.Code, w.Body.String())
	}
	var got researchDoc
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.RoundCount != researchRoundCount {
		t.Fatalf("GET round_count %d, want %d", got.RoundCount, researchRoundCount)
	}
	if len(got.Pieces) != researchRoundCount {
		t.Fatalf("GET pieces %d, want %d", len(got.Pieces), researchRoundCount)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/tran/research/1/publish", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("publish %d %s", w.Code, w.Body.String())
	}
	var pub map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &pub); err != nil {
		t.Fatal(err)
	}
	path, _ := pub["published_path"].(string)
	slug := strings.TrimPrefix(path, "/api/tran/public/research/")
	if slug == "" || slug == path {
		t.Fatalf("published_path %v", pub["published_path"])
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/tran/public/research/"+slug, nil)
	r.ServeHTTP(w, req)
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, "<!doctype html>") {
		t.Fatalf("public get %d %s", w.Code, body)
	}
	if !strings.Contains(body, "Research") {
		t.Fatalf("public html missing research chrome")
	}
	if strings.Contains(body, "## Round ") {
		t.Fatal("published html still has round-order markdown headings")
	}
}

func TestResearchRoundTargetTwentyStillRunsTwenty(t *testing.T) {
	sqlDB := openResearchDB(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	_, err := sqlDB.Exec(
		`INSERT INTO research (user_id, owner_key, title, prompt, status, round_target) VALUES (1,'','T','legacy twenty','running', ?)`,
		researchLegacyRounds,
	)
	if err != nil {
		t.Fatal(err)
	}
	stubResearchPipeline(t)
	h.runResearchJob(context.Background(), 1)
	pieces, err := loadResearchPieces(sqlDB, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(pieces) != researchLegacyRounds {
		t.Fatalf("got %d pieces, want %d", len(pieces), researchLegacyRounds)
	}
	var current, target int
	if err := sqlDB.QueryRow(`SELECT current_round, round_target FROM research WHERE id = 1`).Scan(&current, &target); err != nil {
		t.Fatal(err)
	}
	if current != researchLegacyRounds || target != researchLegacyRounds {
		t.Fatalf("current/target %d/%d, want %d/%d", current, target, researchLegacyRounds, researchLegacyRounds)
	}
}

func TestResearchComposePromptIncludesVisualFirst(t *testing.T) {
	var got string
	prev := researchLLMHook
	t.Cleanup(func() { researchLLMHook = prev })
	researchLLMHook = func(_ context.Context, prompt string) (string, error) {
		if strings.HasPrefix(prompt, researchComposePrefix) {
			got = prompt
		}
		return "# ok", nil
	}
	h := &Handlers{}
	if _, err := h.synthesizeResearchConclusion(context.Background(), "graphene batteries", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Visual-first") {
		t.Fatalf("compose prompt missing Visual-first:\n%s", got)
	}
}

func insertRunningResearch(t *testing.T, db *sql.DB, prompt string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO research (user_id, owner_key, title, prompt, status, round_target) VALUES (1,'','T',?,'running', ?)`,
		prompt, researchRoundCount,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func researchJobRow(t *testing.T, db *sql.DB) (status, markdown, errText string) {
	t.Helper()
	var errNS sql.NullString
	if err := db.QueryRow(`SELECT status, markdown_content, error_text FROM research WHERE id = 1`).Scan(&status, &markdown, &errNS); err != nil {
		t.Fatal(err)
	}
	if errNS.Valid {
		errText = errNS.String
	}
	return status, markdown, errText
}

func stubResearchAIReady(t *testing.T) {
	t.Helper()
	prev := researchLLMHook
	researchLLMHook = func(context.Context, string) (string, error) { return "ok", nil }
	t.Cleanup(func() { researchLLMHook = prev })
}

func silenceResearchGather(t *testing.T) {
	t.Helper()
	prev := researchWebGather
	t.Cleanup(func() { researchWebGather = prev })
	researchWebGather = func(string) (string, []webresearch.Source) {
		return "offline note", nil
	}
}

func TestResearchAllRoundWritesFailEndsFailed(t *testing.T) {
	sqlDB := openResearchDB(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	insertRunningResearch(t, sqlDB, "graphene batteries")
	silenceResearchGather(t)
	prev := researchLLMHook
	t.Cleanup(func() { researchLLMHook = prev })
	researchLLMHook = func(_ context.Context, prompt string) (string, error) {
		if strings.Contains(prompt, "Write markdown for research round") {
			return "", errors.New("model unavailable")
		}
		if strings.Contains(prompt, "Propose one short") {
			return "graphene", nil
		}
		return "", errors.New("model unavailable")
	}
	h.runResearchJob(context.Background(), 1)
	status, markdown, errText := researchJobRow(t, sqlDB)
	if status != "failed" {
		t.Fatalf("status %q, want failed; conclusion %q", status, truncateRunes(markdown, 240))
	}
	if strings.TrimSpace(markdown) != "" {
		t.Fatalf("conclusion should be empty, got %q", truncateRunes(markdown, 240))
	}
	if strings.TrimSpace(errText) == "" || !strings.Contains(strings.ToLower(errText), "failed") {
		t.Fatalf("error_text %q", errText)
	}
	if strings.Contains(markdown, "Round ") {
		t.Fatalf("failure text stored as thesis: %q", markdown)
	}
}

func TestResearchPartialRoundFailureKeepsThesis(t *testing.T) {
	sqlDB := openResearchDB(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	insertRunningResearch(t, sqlDB, "graphene batteries")
	silenceResearchGather(t)
	prev := researchLLMHook
	t.Cleanup(func() { researchLLMHook = prev })
	researchLLMHook = func(_ context.Context, prompt string) (string, error) {
		if strings.HasPrefix(prompt, researchComposePrefix) || strings.HasPrefix(prompt, researchCriticPrefix) {
			return "# Partial thesis\n\nSupported claims only.", nil
		}
		if strings.Contains(prompt, "Write markdown for research round 2 of") || strings.Contains(prompt, "Write markdown for research round 4 of") {
			return "", errors.New("model unavailable")
		}
		if strings.Contains(prompt, "Write markdown for research round") {
			return "Supported finding.", nil
		}
		if strings.Contains(prompt, "Propose one short") {
			return "graphene", nil
		}
		if strings.HasPrefix(prompt, "You verify research claims") {
			return "Claims match the notes.", nil
		}
		return "", errors.New("unexpected prompt")
	}
	h.runResearchJob(context.Background(), 1)
	status, markdown, errText := researchJobRow(t, sqlDB)
	if status != "complete" {
		t.Fatalf("status %q, want complete; error_text %q", status, errText)
	}
	if !strings.Contains(markdown, "Partial thesis") {
		t.Fatalf("conclusion %q", truncateRunes(markdown, 240))
	}
	if strings.Contains(markdown, "Round 2 failed") || strings.Contains(markdown, "## Round ") {
		t.Fatalf("conclusion used failure notes: %q", truncateRunes(markdown, 240))
	}
	if !strings.Contains(errText, "2 of 5") || !strings.Contains(errText, "2") || !strings.Contains(errText, "4") {
		t.Fatalf("error_text %q", errText)
	}
}

func TestResearchSynthesisFailureEndsFailed(t *testing.T) {
	sqlDB := openResearchDB(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	insertRunningResearch(t, sqlDB, "graphene batteries")
	silenceResearchGather(t)
	prev := researchLLMHook
	t.Cleanup(func() { researchLLMHook = prev })
	researchLLMHook = func(_ context.Context, prompt string) (string, error) {
		if strings.HasPrefix(prompt, researchComposePrefix) || strings.HasPrefix(prompt, researchCriticPrefix) {
			return "", nil
		}
		if strings.Contains(prompt, "Write markdown for research round") {
			return "Supported finding.", nil
		}
		if strings.Contains(prompt, "Propose one short") {
			return "graphene", nil
		}
		if strings.HasPrefix(prompt, "You verify research claims") {
			return "Claims match the notes.", nil
		}
		return "", errors.New("unexpected prompt")
	}
	h.runResearchJob(context.Background(), 1)
	status, markdown, errText := researchJobRow(t, sqlDB)
	if status != "failed" {
		t.Fatalf("status %q, want failed; conclusion %q", status, truncateRunes(markdown, 180))
	}
	if strings.TrimSpace(markdown) != "" {
		t.Fatalf("conclusion should be empty, got %q", truncateRunes(markdown, 180))
	}
	if !strings.Contains(strings.ToLower(errText), "could not be synthesized") {
		t.Fatalf("error_text %q", errText)
	}
	pieces, err := loadResearchPieces(sqlDB, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(pieces) != researchRoundCount {
		t.Fatalf("pieces %d, want %d", len(pieces), researchRoundCount)
	}
	for _, p := range pieces {
		if p.Status != "ok" {
			t.Fatalf("round %d status %s", p.RoundIndex, p.Status)
		}
	}
}

func TestResearchNotConfiguredAfterASuccessfulRound(t *testing.T) {
	sqlDB := openResearchDB(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	insertRunningResearch(t, sqlDB, "graphene batteries")
	silenceResearchGather(t)
	prev := researchLLMHook
	t.Cleanup(func() { researchLLMHook = prev })
	researchLLMHook = func(_ context.Context, prompt string) (string, error) {
		if strings.Contains(prompt, "Write markdown for research round 2 of") {
			return "", errResearchAINotConfigured
		}
		if strings.HasPrefix(prompt, researchComposePrefix) || strings.HasPrefix(prompt, researchCriticPrefix) {
			return "", errors.New("synthesis should not run after AI becomes unconfigured")
		}
		if strings.Contains(prompt, "Write markdown for research round") {
			return "Supported finding.", nil
		}
		if strings.Contains(prompt, "Propose one short") {
			return "graphene", nil
		}
		if strings.HasPrefix(prompt, "You verify research claims") {
			return "Claims match the notes.", nil
		}
		return "", errors.New("unexpected prompt")
	}
	h.runResearchJob(context.Background(), 1)
	status, markdown, errText := researchJobRow(t, sqlDB)
	if status != "failed" {
		t.Fatalf("status %q, want failed", status)
	}
	if errText != researchAINotConfiguredMessage {
		t.Fatalf("error_text %q", errText)
	}
	if strings.TrimSpace(markdown) != "" {
		t.Fatalf("conclusion %q", markdown)
	}
	pieces, err := loadResearchPieces(sqlDB, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(pieces) < 1 || pieces[0].Status != "ok" {
		t.Fatalf("expected the successful round to remain, got %+v", pieces)
	}
}

func TestResearchJobFailsWhenAIServiceNil(t *testing.T) {
	sqlDB := openResearchDB(t)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}}
	insertRunningResearch(t, sqlDB, "graphene batteries")
	silenceResearchGather(t)
	prev := researchLLMHook
	researchLLMHook = nil
	t.Cleanup(func() { researchLLMHook = prev })
	h.runResearchJob(context.Background(), 1)
	status, markdown, errText := researchJobRow(t, sqlDB)
	if status != "failed" {
		t.Fatalf("status %q, want failed; conclusion %q", status, truncateRunes(markdown, 240))
	}
	if errText != "AI is not configured (set MORPH_AI_API_KEY)" {
		t.Fatalf("error_text %q", errText)
	}
	if strings.TrimSpace(markdown) != "" {
		t.Fatalf("conclusion should be empty, got %q", markdown)
	}
}

func TestResearchJobFailsWhenAINotConfigured(t *testing.T) {
	t.Setenv("MORPH_AI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("TRAN_QWEN_API_KEY", "")
	sqlDB := openResearchDB(t)
	svc, err := ai.New("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}, aiService: svc}
	insertRunningResearch(t, sqlDB, "graphene batteries")
	silenceResearchGather(t)
	prev := researchLLMHook
	researchLLMHook = nil
	t.Cleanup(func() { researchLLMHook = prev })
	h.runResearchJob(context.Background(), 1)
	status, markdown, errText := researchJobRow(t, sqlDB)
	if status != "failed" {
		t.Fatalf("status %q, want failed; conclusion %q", status, truncateRunes(markdown, 240))
	}
	if errText != "AI is not configured (set MORPH_AI_API_KEY)" {
		t.Fatalf("error_text %q", errText)
	}
	if strings.TrimSpace(markdown) != "" {
		t.Fatalf("conclusion should be empty, got %q", markdown)
	}
}
