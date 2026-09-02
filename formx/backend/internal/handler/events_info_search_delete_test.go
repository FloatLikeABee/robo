package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/formsx/backend/internal/config"
	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/mongo"
	"github.com/gin-gonic/gin"
)

func testEventsInfoHandler(t *testing.T) *Handler {
	t.Helper()
	store, err := mongo.NewStore(&config.Config{FormsXBadgerPath: t.TempDir()})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return &Handler{EventInfoRepo: mongo.NewEventInfoRepo(store)}
}

func TestListEventInfoQFiltersTitle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := testEventsInfoHandler(t)
	now := time.Now().UTC()
	if err := h.EventInfoRepo.Insert(context.Background(), &models.EventInfo{
		Title: "Pump leak at site B", EventTime: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := h.EventInfoRepo.Insert(context.Background(), &models.EventInfo{
		Title: "Quarterly review", EventTime: now,
	}); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.GET("/events-info", h.ListEventInfo)

	req := httptest.NewRequest(http.MethodGet, "/events-info?q=pump", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Events []models.EventInfoResponse `json:"events"`
		Total  int64                      `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Total != 1 || len(body.Events) != 1 {
		t.Fatalf("want 1 match, got total=%d n=%d", body.Total, len(body.Events))
	}
	if body.Events[0].Title != "Pump leak at site B" {
		t.Fatalf("title=%q", body.Events[0].Title)
	}

	reqAll := httptest.NewRequest(http.MethodGet, "/events-info", nil)
	wAll := httptest.NewRecorder()
	r.ServeHTTP(wAll, reqAll)
	if err := json.Unmarshal(wAll.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Total != 2 {
		t.Fatalf("empty q should list all, total=%d", body.Total)
	}
}

func TestBatchDeleteEventInfoSkipsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := testEventsInfoHandler(t)
	now := time.Now().UTC()
	keep := &models.EventInfo{Title: "Keep me", EventTime: now}
	drop := &models.EventInfo{Title: "Drop me", EventTime: now}
	if err := h.EventInfoRepo.Insert(context.Background(), keep); err != nil {
		t.Fatal(err)
	}
	if err := h.EventInfoRepo.Insert(context.Background(), drop); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.POST("/events-info/batch-delete", h.BatchDeleteEventInfo)
	payload, _ := json.Marshal(map[string]any{"ids": []string{drop.ID, "missing-id"}})
	req := httptest.NewRequest(http.MethodPost, "/events-info/batch-delete", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", w.Code, w.Body.String())
	}
	var out struct {
		Deleted int `json:"deleted"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Deleted != 1 {
		t.Fatalf("deleted=%d", out.Deleted)
	}
	list, total, err := h.EventInfoRepo.List(context.Background(), 1, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || list[0].ID != keep.ID {
		t.Fatalf("remaining total=%d id=%s", total, list[0].ID)
	}
}
