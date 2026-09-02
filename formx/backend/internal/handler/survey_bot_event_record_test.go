package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/formsx/backend/internal/config"
	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/mongo"
	"github.com/formsx/backend/internal/surveybot"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

func openTestStore(t *testing.T) *mongo.Store {
	t.Helper()
	store, err := mongo.NewStore(&config.Config{FormsXBadgerPath: t.TempDir()})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func fakeDashScopeAI(t *testing.T, reply string) *morphai.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"output": map[string]any{
				"choices": []any{
					map[string]any{
						"message": map[string]any{"content": reply},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)
	return morphai.NewClient(morphai.Config{
		APIKey:       "test-key",
		Model:        morphai.DefaultModel,
		APIURL:       srv.URL,
		BaseURL:      srv.URL,
		UseNativeAPI: true,
	})
}

func testEventRecordHandler(t *testing.T, ai *morphai.Client) *Handler {
	t.Helper()
	store := openTestStore(t)
	return &Handler{
		EventInfoRepo:       mongo.NewEventInfoRepo(store),
		SurveyBotResultRepo: mongo.NewSurveyBotResultRepo(store),
		AI:                  ai,
	}
}

func TestFinalizeSurveySavesWhenAIUnset(t *testing.T) {
	h := testEventRecordHandler(t, morphai.NewClient(morphai.Config{}))
	st := assistantConversation{
		SurveyBot: &SurveyBotRuntime{
			Title:        "Pump check",
			TemplateID:   "tpl1",
			TemplateSlug: "pump-check",
			Answers:      map[string]string{"site": "B", "issue": "leak"},
		},
	}
	turn := h.finalizeSurvey(context.Background(), st, &surveybot.ParsedTemplate{Title: "Pump check"})
	if !turn.Done {
		t.Fatal("expected completed turn")
	}
	rec, ok := turn.Record.(map[string]any)
	if !ok {
		t.Fatalf("record type %T", turn.Record)
	}
	id, _ := rec["id"].(string)
	if id == "" {
		t.Fatalf("missing result id in %+v", rec)
	}
	res, err := h.SurveyBotResultRepo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if !res.EventRecorded || res.EventID == "" {
		t.Fatalf("expected fallback Events & Info when AI is unset, recorded=%v id=%q err=%q", res.EventRecorded, res.EventID, res.EventRecordError)
	}
	list, total, err := h.EventInfoRepo.List(context.Background(), 1, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("want 1 fallback event, got total=%d", total)
	}
	if list[0].Source != infoSheetEventSource || list[0].SourceResultID != id {
		t.Fatalf("source fields: %+v", list[0])
	}
	if list[0].Title != "Pump check" {
		t.Fatalf("fallback title = %q", list[0].Title)
	}
	if !strings.Contains(list[0].Detail, "leak") {
		t.Fatalf("fallback detail missing answers: %q", list[0].Detail)
	}
}

func TestInsertInfoSheetEventNoDuplicate(t *testing.T) {
	h := testEventRecordHandler(t, morphai.NewClient(morphai.Config{}))
	res := &models.SurveyBotResult{
		Title:   "Pump check",
		Answers: map[string]string{"site": "B"},
	}
	if err := h.SurveyBotResultRepo.Insert(context.Background(), res); err != nil {
		t.Fatal(err)
	}
	draft := eventInfoIngestDraft{Title: "Pump leak at site B", Detail: "Leak reported", Reporter: "field"}
	first, err := h.insertInfoSheetEvent(context.Background(), res, draft)
	if err != nil {
		t.Fatal(err)
	}
	second, err := h.insertInfoSheetEvent(context.Background(), res, eventInfoIngestDraft{
		Title: "Different title", Detail: "Should not insert",
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("dedupe failed: %s vs %s", first.ID, second.ID)
	}
	_, total, err := h.EventInfoRepo.List(context.Background(), 1, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("want 1 event, got %d", total)
	}
}

func TestRecordInfoSheetResultAsEventOnComplete(t *testing.T) {
	h := testEventRecordHandler(t, fakeDashScopeAI(t, `{"title":"Pump leak at site B","detail":"Leak reported","reporter":"","time":""}`))
	st := assistantConversation{
		SurveyBot: &SurveyBotRuntime{
			Title:        "Pump check",
			TemplateID:   "tpl1",
			TemplateSlug: "pump-check",
			Answers:      map[string]string{"site": "B", "issue": "leak"},
		},
	}
	turn := h.finalizeSurvey(context.Background(), st, &surveybot.ParsedTemplate{Title: "Pump check"})
	rec, _ := turn.Record.(map[string]any)
	id, _ := rec["id"].(string)
	res, err := h.SurveyBotResultRepo.GetByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if !res.EventRecorded || res.EventID == "" {
		t.Fatalf("expected recorded event, got recorded=%v id=%q err=%q", res.EventRecorded, res.EventID, res.EventRecordError)
	}
	list, total, err := h.EventInfoRepo.List(context.Background(), 1, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("want 1 event, got %d", total)
	}
	if list[0].Source != infoSheetEventSource || list[0].SourceResultID != id {
		t.Fatalf("source fields: %+v", list[0])
	}

	h.recordInfoSheetResultAsEvent(context.Background(), res)
	_, total, err = h.EventInfoRepo.List(context.Background(), 1, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("second record must not duplicate, got %d", total)
	}
}

func TestRecordSurveyBotResultAsEventRetry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := testEventRecordHandler(t, fakeDashScopeAI(t, `{"title":"Site notes","detail":"Collected","reporter":"","time":""}`))
	res := &models.SurveyBotResult{
		Title:            "Field notes",
		Answers:          map[string]string{"note": "gate stuck"},
		EventRecordError: "AI is not configured",
	}
	if err := h.SurveyBotResultRepo.Insert(context.Background(), res); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.POST("/survey-bot/results/:id/record-event", h.RecordSurveyBotResultAsEvent)
	req := httptest.NewRequest(http.MethodPost, "/survey-bot/results/"+res.ID+"/record-event", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", w.Code, w.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["event_recorded"] != true {
		t.Fatalf("expected event_recorded, got %+v", out)
	}
	_, total, err := h.EventInfoRepo.List(context.Background(), 1, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("want 1 event after retry, got %d", total)
	}
}
