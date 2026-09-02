package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/formsx/backend/internal/mongo"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

func TestEventInfoAIDraftRouteIsRegistered(t *testing.T) {
	h := &Handler{AI: morphai.NewClient(morphai.Config{})}
	r := gin.New()
	h.Register(r.Group(""))

	found := false
	ingestFound := false
	recordFound := false
	batchDeleteFound := false
	for _, route := range r.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/v1/events-info/ai-draft" {
			found = true
		}
		if route.Method == http.MethodPost && route.Path == "/api/v1/events-info/ai-ingest" {
			ingestFound = true
		}
		if route.Method == http.MethodPost && route.Path == "/api/v1/survey-bot/results/:id/record-event" {
			recordFound = true
		}
		if route.Method == http.MethodPost && route.Path == "/api/v1/events-info/batch-delete" {
			batchDeleteFound = true
		}
	}
	if !found {
		t.Fatal("POST /api/v1/events-info/ai-draft is not registered")
	}
	if !ingestFound {
		t.Fatal("POST /api/v1/events-info/ai-ingest is not registered")
	}
	if !recordFound {
		t.Fatal("POST /api/v1/survey-bot/results/:id/record-event is not registered")
	}
	if !batchDeleteFound {
		t.Fatal("POST /api/v1/events-info/batch-delete is not registered")
	}
}

func TestDraftEventInfoAIRequiresPrompt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{AI: morphai.NewClient(morphai.Config{APIKey: "k", BaseURL: morphai.DefaultBaseURL})}
	r := gin.New()
	r.POST("/events-info/ai-draft", h.DraftEventInfoAI)

	req := httptest.NewRequest(http.MethodPost, "/events-info/ai-draft", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400: %s", w.Code, w.Body.String())
	}
}

func TestDraftEventInfoAIReportsUnconfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{AI: morphai.NewClient(morphai.Config{})}
	r := gin.New()
	r.POST("/events-info/ai-draft", h.DraftEventInfoAI)

	req := httptest.NewRequest(http.MethodPost, "/events-info/ai-draft", bytes.NewBufferString(`{"prompt":"pump failure at site B"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503: %s", w.Code, w.Body.String())
	}
}

func TestIngestEventInfoAIRequiresSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{AI: morphai.NewClient(morphai.Config{APIKey: "k", BaseURL: morphai.DefaultBaseURL})}
	r := gin.New()
	r.POST("/events-info/ai-ingest", h.IngestEventInfoAI)

	req := httptest.NewRequest(http.MethodPost, "/events-info/ai-ingest", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400: %s", w.Code, w.Body.String())
	}
}

func TestIsAllowedEventIngestFileJSON(t *testing.T) {
	if !isAllowedEventIngestFile("notes.json", "application/json") {
		t.Fatal("json file should be allowed")
	}
	if !isAllowedEventIngestFile("notes.JSON", "") {
		t.Fatal(".json extension should be allowed")
	}
	if isAllowedEventIngestFile("photo.png", "image/png") {
		t.Fatal("png should be rejected")
	}
}

func TestEventIngestTextFromBytesPrettyPrintsJSON(t *testing.T) {
	got := eventIngestTextFromBytes("notes.json", "application/json", []byte(`{"title":"pump","n":1}`))
	if !strings.Contains(got, `"title": "pump"`) || !strings.Contains(got, `"n": 1`) {
		t.Fatalf("expected pretty JSON, got %q", got)
	}
}

func TestIngestEventInfoAIReportsUnconfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{AI: morphai.NewClient(morphai.Config{})}
	r := gin.New()
	r.POST("/events-info/ai-ingest", h.IngestEventInfoAI)

	req := httptest.NewRequest(http.MethodPost, "/events-info/ai-ingest", strings.NewReader("paste=site+notes"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503: %s", w.Code, w.Body.String())
	}
}

func TestIngestEventInfoAIRejectsUnsupportedType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{AI: morphai.NewClient(morphai.Config{})}
	r := gin.New()
	r.POST("/events-info/ai-ingest", h.IngestEventInfoAI)

	body, ct := multipartBody(t, nil, uploadPart{
		field: "file", filename: "photo.png", mime: "image/png", content: []byte("not-a-png"),
	})
	req := httptest.NewRequest(http.MethodPost, "/events-info/ai-ingest", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "unsupported file type") {
		t.Fatalf("expected unsupported type error, got %s", w.Body.String())
	}
}

func TestIngestEventInfoAIJSONReturnsDraftsWithoutPersist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := openTestStore(t)
	h := &Handler{
		AI:            fakeDashScopeAI(t, `[{"title":"Pump leak","detail":"Site B"}]`),
		EventInfoRepo: mongo.NewEventInfoRepo(store),
	}
	r := gin.New()
	r.POST("/events-info/ai-ingest", h.IngestEventInfoAI)

	body, ct := multipartBody(t, nil, uploadPart{
		field:    "file",
		filename: "notes.json",
		mime:     "application/json",
		content:  []byte(`{"incident":"pump leak","site":"B"}`),
	})
	req := httptest.NewRequest(http.MethodPost, "/events-info/ai-ingest", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", w.Code, w.Body.String())
	}
	var out struct {
		Drafts []eventInfoIngestDraft `json:"drafts"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Drafts) != 1 || out.Drafts[0].Title != "Pump leak" {
		t.Fatalf("drafts = %+v", out.Drafts)
	}
	list, total, err := h.EventInfoRepo.List(context.Background(), 1, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(list) != 0 {
		t.Fatalf("ingest must not persist events, got total=%d list=%d", total, len(list))
	}
}
