package morphai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMorphAIProviderFromEnv(t *testing.T) {
	const xaiKey = "xai-env-key-value"
	const legacyKey = "legacy-env-key-value"

	t.Run("xai uses its own key and base", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("XAI_API_KEY", "")
		t.Setenv("XAI_BASE_URL", "")
		var hitsLegacy, hitsXAI int
		var sawLegacyKey, sawXAIKey bool
		var model string
		legacy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsLegacy++
			auth := r.Header.Get("Authorization")
			sawLegacyKey = strings.Contains(auth, legacyKey)
			sawXAIKey = strings.Contains(auth, xaiKey)
			writeJSON(w, http.StatusOK, openAIText("legacy"))
		}))
		defer legacy.Close()
		xai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsXAI++
			auth := r.Header.Get("Authorization")
			sawXAIKey = auth == "Bearer "+xaiKey
			sawLegacyKey = strings.Contains(auth, legacyKey)
			body := readJSONBody(t, r)
			if s, ok := body["model"].(string); ok {
				model = s
			}
			writeJSON(w, http.StatusOK, openAIText("xai"))
		}))
		defer xai.Close()
		t.Setenv("MORPH_AI_PROVIDER", " XAI ")
		t.Setenv("MORPH_AI_API_KEY", legacyKey)
		t.Setenv("MORPH_AI_MODEL", DefaultModel)
		t.Setenv("MORPH_AI_BASE_URL", legacy.URL)
		t.Setenv("XAI_API_KEY", xaiKey)
		t.Setenv("XAI_BASE_URL", xai.URL)
		cfg := LoadFromEnv()
		if cfg.Provider != ProviderXAI {
			t.Fatalf("provider = %q", cfg.Provider)
		}
		c := newHTTPClient(t, cfg)
		got, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		if err != nil {
			t.Fatal(err)
		}
		if got != "xai" || hitsXAI != 1 || hitsLegacy != 0 || !sawXAIKey || sawLegacyKey {
			t.Fatal("legacy base received the xai call")
		}
		if model != "grok-4.3" {
			t.Fatalf("model = %q", model)
		}
	})

	t.Run("xai without its own key is not configured", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("XAI_API_KEY", "")
		t.Setenv("XAI_BASE_URL", "")
		var hits int
		legacy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits++
			writeJSON(w, http.StatusOK, openAIText("legacy"))
		}))
		defer legacy.Close()
		t.Setenv("MORPH_AI_PROVIDER", "xai")
		t.Setenv("MORPH_AI_API_KEY", legacyKey)
		t.Setenv("MORPH_AI_BASE_URL", legacy.URL)
		c := newHTTPClient(t, LoadFromEnv())
		_, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("err = %v", err)
		}
		if hits != 0 {
			t.Fatalf("hits = %d", hits)
		}
	})

	t.Run("unset provider stays on the legacy client", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("XAI_API_KEY", xaiKey)
		t.Setenv("XAI_BASE_URL", "https://xai.example/v1")
		var hits int
		var model string
		var authOK bool
		legacy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits++
			authOK = r.Header.Get("Authorization") == "Bearer "+legacyKey
			body := readJSONBody(t, r)
			if s, ok := body["model"].(string); ok {
				model = s
			}
			writeJSON(w, http.StatusOK, openAIText("legacy"))
		}))
		defer legacy.Close()
		t.Setenv("MORPH_AI_PROVIDER", "   ")
		t.Setenv("MORPH_AI_API_KEY", legacyKey)
		t.Setenv("MORPH_AI_MODEL", DefaultModel)
		t.Setenv("MORPH_AI_BASE_URL", legacy.URL)
		cfg := LoadFromEnv()
		if cfg.Provider != "" {
			t.Fatalf("provider = %q", cfg.Provider)
		}
		c := newHTTPClient(t, cfg)
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
		if hits != 1 || !authOK || model != DefaultModel {
			t.Fatalf("legacy hits %d model %q", hits, model)
		}
	})

	t.Run("unknown provider errors", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_PROVIDER", " Not-A-Vendor ")
		t.Setenv("MORPH_AI_API_KEY", legacyKey)
		cfg := LoadFromEnv()
		if cfg.Provider != "not-a-vendor" {
			t.Fatalf("provider = %q", cfg.Provider)
		}
		c := newHTTPClient(t, cfg)
		c.httpClient.Transport = denyTransport{t}
		c.httpClientLong.Transport = denyTransport{t}
		_, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		if !errors.Is(err, ErrUnknownProvider) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("explicit model is sent", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_PROVIDER", "xai")
		t.Setenv("MORPH_AI_MODEL", "grok-3")
		t.Setenv("XAI_API_KEY", xaiKey)
		rc, err := ResolveConfig(LoadFromEnv())
		if err != nil {
			t.Fatal(err)
		}
		if rc.Model != "grok-3" || rc.Provider != ProviderXAI {
			t.Fatalf("resolved %s %s", rc.Provider, rc.Model)
		}
	})

	t.Run("missing model uses the provider default", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_PROVIDER", "xai")
		t.Setenv("MORPH_AI_MODEL", "")
		t.Setenv("GEMINI_MODEL", "gemini-2.5-flash")
		t.Setenv("TRAN_QWEN_MODEL", "qwen-plus")
		t.Setenv("XAI_API_KEY", xaiKey)
		rc, err := ResolveConfig(LoadFromEnv())
		if err != nil {
			t.Fatal(err)
		}
		if rc.Model != "grok-4.3" {
			t.Fatalf("model = %q", rc.Model)
		}
	})

	t.Run("legacy default assigned after load is not sent to xai", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_PROVIDER", "xai")
		t.Setenv("MORPH_AI_MODEL", "")
		t.Setenv("XAI_API_KEY", xaiKey)
		cfg := LoadFromEnv()
		cfg.Model = DefaultModel
		rc, err := ResolveConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if rc.Model != "grok-4.3" {
			t.Fatalf("model = %q", rc.Model)
		}
	})

	t.Run("hand built qwen model is honored", func(t *testing.T) {
		rc, err := resolve(Config{Provider: ProviderOpenAI, APIKey: "caller-key", Model: DefaultModel})
		if err != nil {
			t.Fatal(err)
		}
		if rc.Model != DefaultModel {
			t.Fatalf("model = %q", rc.Model)
		}
	})

	t.Run("dashscope keeps the legacy default model", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_PROVIDER", "dashscope")
		t.Setenv("MORPH_AI_MODEL", DefaultModel)
		t.Setenv("DASHSCOPE_API_KEY", "dashscope-env-key")
		rc, err := ResolveConfig(LoadFromEnv())
		if err != nil {
			t.Fatal(err)
		}
		if rc.Provider != ProviderDashScope || rc.Model != DefaultModel {
			t.Fatalf("resolved %s %s", rc.Provider, rc.Model)
		}
	})
}
