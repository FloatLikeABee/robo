package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"idongivaflyinfa/ai"
	dorphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
	"github.com/robo/dorphai"
)

func openProviderSQL(t *testing.T) (*Handlers, *dorphdb.PlatUser) {
	t.Helper()
	sqlDB := openHarnessSQL(t)
	if _, err := sqlDB.Exec(`CREATE TABLE user_ai_provider (
		user_id TEXT NOT NULL PRIMARY KEY,
		provider TEXT NOT NULL,
		base_url TEXT NOT NULL DEFAULT '',
		api_key TEXT NOT NULL DEFAULT '',
		model TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL
	)`); err != nil {
		t.Fatal(err)
	}
	h, user, _ := lessonHandlers(t, sqlDB)
	h.aiService = &ai.AIService{}
	return h, user
}

func providerRouter(h *Handlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/chat/providers", h.ListChatProviders)
	r.GET("/api/chat/provider", h.GetChatProvider)
	r.PUT("/api/chat/provider", h.SaveChatProvider)
	r.POST("/api/chat/provider/scan", h.ScanChatProvider)
	return r
}

func TestGetChatProviderMasksKeyAndEmptyUpdateKeepsIt(t *testing.T) {
	h, userA := openProviderSQL(t)
	if err := h.TranMySQL.SaveUserAIProvider(context.Background(), dorphdb.UserAIProvider{
		UserID: userA.ID, Provider: "openai", BaseURL: "https://api.example/v1", APIKey: "secret-key", Model: "gpt-4.1",
	}); err != nil {
		t.Fatal(err)
	}
	r := providerRouter(h)
	get := httptest.NewRequest(http.MethodGet, "/api/chat/provider", nil)
	setBearer(t, h, get, userA)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, get)
	if rec.Code != http.StatusOK {
		t.Fatalf("get %d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret-key") {
		t.Fatalf("GET returned the key: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"key_set":true`) {
		t.Fatalf("missing key_set: %s", rec.Body.String())
	}
	put := httptest.NewRequest(http.MethodPut, "/api/chat/provider", strings.NewReader(`{"provider":"openai","base_url":"https://api.example/v1","api_key":"","model":"gpt-4.1"}`))
	put.Header.Set("Content-Type", "application/json")
	setBearer(t, h, put, userA)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, put)
	if rec.Code != http.StatusOK {
		t.Fatalf("put %d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret-key") {
		t.Fatalf("update returned the key: %s", rec.Body.String())
	}
	got, ok, err := h.TranMySQL.GetUserAIProvider(context.Background(), userA.ID)
	if err != nil || !ok || got.APIKey != "secret-key" {
		t.Fatalf("empty update dropped the key: %+v ok=%v err=%v", got, ok, err)
	}
}

func TestOllamaSaveWithoutKey(t *testing.T) {
	h, userA := openProviderSQL(t)
	r := providerRouter(h)
	req := httptest.NewRequest(http.MethodPut, "/api/chat/provider", strings.NewReader(`{"provider":"ollama","base_url":"http://127.0.0.1:11434/v1","api_key":"","model":"llama3"}`))
	req.Header.Set("Content-Type", "application/json")
	setBearer(t, h, req, userA)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ollama save %d %s", rec.Code, rec.Body.String())
	}
	got, ok, err := h.TranMySQL.GetUserAIProvider(context.Background(), userA.ID)
	if err != nil || !ok || got.Model != "llama3" || got.APIKey != "" {
		t.Fatalf("saved %+v ok=%v err=%v", got, ok, err)
	}
}

func TestFailedScanDoesNotDeleteSavedRow(t *testing.T) {
	h, userA := openProviderSQL(t)
	if err := h.TranMySQL.SaveUserAIProvider(context.Background(), dorphdb.UserAIProvider{
		UserID: userA.ID, Provider: "openai", BaseURL: "https://api.example/v1", APIKey: "secret-key", Model: "gpt-4.1",
	}); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-key" {
			t.Errorf("scan auth = %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)
	r := providerRouter(h)
	req := httptest.NewRequest(http.MethodPost, "/api/chat/provider/scan", strings.NewReader(`{"provider":"openai","base_url":"`+srv.URL+`/v1","api_key":""}`))
	req.Header.Set("Content-Type", "application/json")
	setBearer(t, h, req, userA)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("scan %d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret-key") {
		t.Fatalf("scan error leaked the key: %s", rec.Body.String())
	}
	got, ok, err := h.TranMySQL.GetUserAIProvider(context.Background(), userA.ID)
	if err != nil || !ok || got.APIKey != "secret-key" || got.Model != "gpt-4.1" {
		t.Fatalf("failed scan changed the row: %+v ok=%v err=%v", got, ok, err)
	}
}

func TestScanOpenAIModelsAndOllamaTags(t *testing.T) {
	openai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"id": "gpt-4.1"}}})
	}))
	t.Cleanup(openai.Close)
	ids, err := scanProviderModels(context.Background(), "openai", openai.URL+"/v1", "k")
	if err != nil || len(ids) != 1 || ids[0] != "gpt-4.1" {
		t.Fatalf("openai scan %v %v", ids, err)
	}

	ollama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"models": []map[string]string{{"name": "llama3"}}})
	}))
	t.Cleanup(ollama.Close)
	ids, err = scanProviderModels(context.Background(), "ollama", ollama.URL+"/v1", "")
	if err != nil || len(ids) != 1 || ids[0] != "llama3" {
		t.Fatalf("ollama scan %v %v", ids, err)
	}
}

func TestNoSavedProviderKeepsEnvironmentClient(t *testing.T) {
	env := dorphai.Config{Provider: dorphai.ProviderOpenAI, Model: "env-model", APIKey: "env-key"}
	if got := resolveChatConfig(nil, env); got != env {
		t.Fatalf("nil row: %+v", got)
	}
	incomplete := &dorphdb.UserAIProvider{Provider: "ollama"}
	if got := resolveChatConfig(incomplete, env); got != env {
		t.Fatalf("incomplete row: %+v", got)
	}

	h, userA := openProviderSQL(t)
	shared := h.aiService
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/chat", nil)
	setBearer(t, h, c.Request, userA)
	if h.aiForChat(c) != shared {
		t.Fatal("user with no row must keep the environment client")
	}
	if err := h.TranMySQL.SaveUserAIProvider(context.Background(), dorphdb.UserAIProvider{
		UserID: userA.ID, Provider: "ollama", BaseURL: "http://127.0.0.1:11434/v1", Model: "llama3",
	}); err != nil {
		t.Fatal(err)
	}
	if h.aiForChat(c) == shared {
		t.Fatal("saved provider must use a request-local client")
	}
}
