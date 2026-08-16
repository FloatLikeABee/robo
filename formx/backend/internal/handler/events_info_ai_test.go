package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

func TestEventInfoAIDraftRouteIsRegistered(t *testing.T) {
	h := &Handler{AI: morphai.NewClient(morphai.Config{})}
	r := gin.New()
	h.Register(r.Group(""))

	found := false
	ingestFound := false
	for _, route := range r.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/v1/events-info/ai-draft" {
			found = true
		}
		if route.Method == http.MethodPost && route.Path == "/api/v1/events-info/ai-ingest" {
			ingestFound = true
		}
	}
	if !found {
		t.Fatal("POST /api/v1/events-info/ai-draft is not registered")
	}
	if !ingestFound {
		t.Fatal("POST /api/v1/events-info/ai-ingest is not registered")
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
