package morphai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newHTTPClient(t *testing.T, cfg Config) *Client {
	t.Helper()
	c := NewClient(cfg)
	c.minRequestInterval = 0
	c.maxRetries = 0
	c.retryBase = 0
	c.httpClient.Timeout = 5 * time.Second
	c.httpClientLong.Timeout = 5 * time.Second
	return c
}

type denyTransport struct{ t *testing.T }

func (d denyTransport) RoundTrip(*http.Request) (*http.Response, error) {
	d.t.Fatal("unexpected HTTP request")
	return nil, errors.New("unexpected HTTP request")
}

func readJSONBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("json body: %v\n%s", err, raw)
	}
	return body
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func openAIText(text string) map[string]any {
	return map[string]any{
		"choices": []any{
			map[string]any{
				"finish_reason": "stop",
				"message":       map[string]any{"role": "assistant", "content": text},
			},
		},
	}
}

func TestListProviders(t *testing.T) {
	got := ListProviders()
	want := []ProviderID{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderXAI,
		ProviderGemini,
		ProviderOpenRouter,
		ProviderOllama,
		ProviderDashScope,
		ProviderMistral,
		ProviderGroq,
		ProviderOpenAICompatible,
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, id := range want {
		if got[i].ID != id {
			t.Errorf("index %d = %s, want %s", i, got[i].ID, id)
		}
		if got[i].DisplayName == "" {
			t.Errorf("%s missing display name", id)
		}
		info, ok := LookupProvider(ProviderID(strings.ToUpper(string(id))))
		if !ok || info.ID != id {
			t.Errorf("lookup %s failed", id)
		}
	}
	ollama, _ := LookupProvider(ProviderOllama)
	if ollama.RequiresKey {
		t.Fatal("ollama should not require a key")
	}
	if ollama.DefaultBaseURL != "http://127.0.0.1:11434/v1" {
		t.Fatalf("ollama base = %s", ollama.DefaultBaseURL)
	}
	compat, _ := LookupProvider(ProviderOpenAICompatible)
	if compat.DefaultBaseURL != "" || !compat.RequiresKey {
		t.Fatalf("compatible meta = %+v", compat)
	}
	if len(compat.SuggestedModels) != 0 {
		t.Fatalf("compatible models = %v", compat.SuggestedModels)
	}
	mistral, _ := LookupProvider(ProviderMistral)
	groq, _ := LookupProvider(ProviderGroq)
	router, _ := LookupProvider(ProviderOpenRouter)
	if mistral.DefaultBaseURL != "https://api.mistral.ai/v1" || mistral.APIKeyEnv != "MISTRAL_API_KEY" {
		t.Fatalf("mistral meta = %+v", mistral)
	}
	if groq.DefaultBaseURL != "https://api.groq.com/openai/v1" || groq.APIKeyEnv != "GROQ_API_KEY" {
		t.Fatalf("groq meta = %+v", groq)
	}
	if router.ID == mistral.ID || router.DefaultBaseURL == mistral.DefaultBaseURL {
		t.Fatal("openrouter collapsed into mistral")
	}
	for _, info := range []ProviderInfo{mistral, groq, router} {
		if !info.Capabilities.Chat || !info.Capabilities.Stream || !info.Capabilities.Tools {
			t.Fatalf("%s missing protocol caps: %+v", info.ID, info.Capabilities)
		}
	}
	anth, _ := LookupProvider(ProviderAnthropic)
	if anth.Capabilities.JSONMode {
		t.Fatal("anthropic should not advertise json mode")
	}
}

func TestResolveExplicitOrder(t *testing.T) {
	t.Run("config key beats env", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "env-key")
		rc, err := resolve(Config{Provider: ProviderOpenAI, APIKey: "cfg-key", Model: "gpt-custom"})
		if err != nil {
			t.Fatal(err)
		}
		if rc.APIKey != "cfg-key" || rc.Model != "gpt-custom" {
			t.Fatalf("resolved key %q model %q", rc.APIKey, rc.Model)
		}
		if rc.Kind != adapterOpenAI || rc.BaseURL != "https://api.openai.com/v1" {
			t.Fatalf("adapter/base = %s %s", rc.Kind, rc.BaseURL)
		}
	})

	t.Run("env key then default model and base", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "env-key")
		t.Setenv("MORPH_AI_API_KEY", "morph-key")
		t.Setenv("OPENAI_BASE_URL", "")
		rc, err := resolve(Config{Provider: ProviderOpenAI})
		if err != nil {
			t.Fatal(err)
		}
		if rc.APIKey != "env-key" {
			t.Fatalf("key = %q", rc.APIKey)
		}
		if rc.Model != "gpt-4.1" {
			t.Fatalf("model = %q", rc.Model)
		}
	})

	t.Run("base url env then config override", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "env-key")
		t.Setenv("OPENAI_BASE_URL", "https://proxy.example/v1")
		rc, err := resolve(Config{Provider: ProviderOpenAI})
		if err != nil {
			t.Fatal(err)
		}
		if rc.BaseURL != "https://proxy.example/v1" {
			t.Fatalf("base = %s", rc.BaseURL)
		}
		rc, err = resolve(Config{Provider: ProviderOpenAI, BaseURL: "https://call.example/v1"})
		if err != nil {
			t.Fatal(err)
		}
		if rc.BaseURL != "https://call.example/v1" {
			t.Fatalf("override base = %s", rc.BaseURL)
		}
	})

	t.Run("dashscope default base is not an openai override", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "env-key")
		t.Setenv("OPENAI_BASE_URL", "")
		rc, err := resolve(Config{Provider: ProviderOpenAI, BaseURL: DefaultBaseURL, APIKey: "k"})
		if err != nil {
			t.Fatal(err)
		}
		if rc.BaseURL != "https://api.openai.com/v1" {
			t.Fatalf("base = %s", rc.BaseURL)
		}
	})

	t.Run("openai does not use morph key", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "")
		t.Setenv("MORPH_AI_API_KEY", "morph-key")
		_, err := resolve(Config{Provider: ProviderOpenAI, Model: "gpt-4.1"})
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("dashscope uses morph key but not gemini key", func(t *testing.T) {
		t.Setenv("DASHSCOPE_API_KEY", "")
		t.Setenv("MORPH_AI_API_KEY", "morph-key")
		t.Setenv("GEMINI_API_KEY", "gem-key")
		t.Setenv("TRAN_QWEN_API_KEY", "")
		rc, err := resolve(Config{Provider: ProviderDashScope})
		if err != nil {
			t.Fatal(err)
		}
		if rc.APIKey != "morph-key" {
			t.Fatalf("key = %q", rc.APIKey)
		}
		t.Setenv("MORPH_AI_API_KEY", "")
		t.Setenv("TRAN_QWEN_API_KEY", "")
		_, err = resolve(Config{Provider: ProviderDashScope, Model: "qwen3-max"})
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("gemini key was accepted for dashscope: %v", err)
		}
	})

	t.Run("dashscope specific key beats morph key", func(t *testing.T) {
		t.Setenv("DASHSCOPE_API_KEY", "ds-key")
		t.Setenv("MORPH_AI_API_KEY", "morph-key")
		rc, err := resolve(Config{Provider: ProviderDashScope, Model: "qwen-plus"})
		if err != nil {
			t.Fatal(err)
		}
		if rc.APIKey != "ds-key" || rc.Model != "qwen-plus" {
			t.Fatalf("got key %q model %q", rc.APIKey, rc.Model)
		}
	})

	t.Run("gemini uses gemini env", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "gem-key")
		t.Setenv("MORPH_AI_API_KEY", "morph-key")
		rc, err := resolve(Config{Provider: ProviderGemini})
		if err != nil {
			t.Fatal(err)
		}
		if rc.APIKey != "gem-key" {
			t.Fatalf("key = %q", rc.APIKey)
		}
		if rc.BaseURL != "https://generativelanguage.googleapis.com/v1beta/openai" {
			t.Fatalf("base = %s", rc.BaseURL)
		}
		if rc.Kind != adapterOpenAI {
			t.Fatalf("gemini adapter = %s", rc.Kind)
		}
	})

	t.Run("mistral and groq env keys", func(t *testing.T) {
		t.Setenv("MISTRAL_API_KEY", "mi-key")
		t.Setenv("GROQ_API_KEY", "groq-key")
		mi, err := resolve(Config{Provider: ProviderMistral})
		if err != nil {
			t.Fatal(err)
		}
		gr, err := resolve(Config{Provider: ProviderGroq})
		if err != nil {
			t.Fatal(err)
		}
		if mi.APIKey != "mi-key" || mi.BaseURL != "https://api.mistral.ai/v1" || mi.Kind != adapterOpenAI {
			t.Fatalf("mistral %+v", mi.public())
		}
		if gr.APIKey != "groq-key" || gr.BaseURL != "https://api.groq.com/openai/v1" || gr.Kind != adapterOpenAI {
			t.Fatalf("groq %+v", gr.public())
		}
	})

	t.Run("ollama base and no key", func(t *testing.T) {
		t.Setenv("OLLAMA_BASE_URL", "")
		rc, err := resolve(Config{Provider: ProviderOllama, Model: "llama3.3"})
		if err != nil {
			t.Fatal(err)
		}
		if rc.APIKey != "" || rc.BaseURL != "http://127.0.0.1:11434/v1" {
			t.Fatalf("ollama key %q base %s", rc.APIKey, rc.BaseURL)
		}
		t.Setenv("OLLAMA_BASE_URL", "http://localhost:11434")
		rc, err = resolve(Config{Provider: ProviderOllama, Model: "llama3.3"})
		if err != nil {
			t.Fatal(err)
		}
		if rc.BaseURL != "http://localhost:11434/v1" {
			t.Fatalf("normalized base = %s", rc.BaseURL)
		}
		if !(Config{Provider: ProviderOllama}).Configured() {
			t.Fatal("ollama should be configured without a key")
		}
	})

	t.Run("compatible requires key and base", func(t *testing.T) {
		t.Setenv("OPENAI_COMPATIBLE_API_KEY", "")
		t.Setenv("OPENAI_COMPATIBLE_BASE_URL", "")
		_, err := resolve(Config{Provider: ProviderOpenAICompatible, BaseURL: "https://example.test/v1"})
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("missing key: %v", err)
		}
		_, err = resolve(Config{Provider: ProviderOpenAICompatible, APIKey: "k"})
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("missing base: %v", err)
		}
		_, err = resolve(Config{Provider: ProviderOpenAICompatible, APIKey: "k", BaseURL: "https://example.test/v1"})
		if err == nil {
			t.Fatal("model should be required")
		}
		if errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("missing model should not be the not-configured error: %v", err)
		}
	})

	t.Run("unknown provider", func(t *testing.T) {
		_, err := resolve(Config{Provider: "not-a-provider", APIKey: "k"})
		if !errors.Is(err, ErrUnknownProvider) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("per-call switch drops the other provider key", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "")
		cfg := Config{APIKey: "sk-morph", Model: "qwen3-max", BaseURL: DefaultBaseURL}.applyCall(CompletionRequest{
			Provider: ProviderOpenAI,
			Model:    "gpt-4.1",
		})
		if cfg.APIKey != "" {
			t.Fatalf("inherited key %q", cfg.APIKey)
		}
		_, err := resolve(cfg)
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestLegacyIgnoresOtherProviderKeys(t *testing.T) {
	t.Setenv("MORPH_AI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("TRAN_QWEN_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "sk-openai")
	t.Setenv("ANTHROPIC_API_KEY", "sk-anth")
	cfg := LoadFromEnv()
	if cfg.Provider != "" {
		t.Fatalf("provider = %q", cfg.Provider)
	}
	if cfg.Configured() {
		t.Fatal("legacy client became configured from a named provider key")
	}
	c := newHTTPClient(t, cfg)
	c.httpClient.Transport = denyTransport{t}
	_, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil || err.Error() != "MORPH_AI_API_KEY is not configured" {
		t.Fatalf("err = %v", err)
	}
	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatal("legacy missing key should unwrap to ErrProviderNotConfigured")
	}
}

func TestLegacyCompatibleAndNativeShape(t *testing.T) {
	t.Run("compatible", func(t *testing.T) {
		var path, auth string
		var body map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			auth = r.Header.Get("Authorization")
			body = readJSONBody(t, r)
			writeJSON(w, http.StatusOK, openAIText("pong"))
		}))
		defer srv.Close()
		c := newHTTPClient(t, Config{APIKey: "sk-test", Model: "qwen3-max", BaseURL: srv.URL})
		got, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "Hello"}})
		if err != nil {
			t.Fatal(err)
		}
		if got != "pong" || path != "/chat/completions" || auth != "Bearer sk-test" {
			t.Fatalf("reply %q path %s auth %s", got, path, auth)
		}
		if body["enable_thinking"] != false {
			t.Fatalf("enable_thinking = %#v", body["enable_thinking"])
		}
		if body["model"] != "qwen3-max" {
			t.Fatalf("model = %#v", body["model"])
		}
	})

	t.Run("custom legacy base still disables thinking", func(t *testing.T) {
		var body map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body = readJSONBody(t, r)
			writeJSON(w, http.StatusOK, openAIText("ok"))
		}))
		defer srv.Close()
		c := newHTTPClient(t, Config{APIKey: "k", Model: "m", BaseURL: srv.URL})
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
		if body["enable_thinking"] != false {
			t.Fatalf("enable_thinking = %#v", body["enable_thinking"])
		}
	})

	t.Run("native", func(t *testing.T) {
		var path, auth string
		var body map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.RequestURI()
			auth = r.Header.Get("Authorization")
			body = readJSONBody(t, r)
			writeJSON(w, http.StatusOK, map[string]any{
				"output": map[string]any{
					"choices": []any{map[string]any{"message": map[string]any{"content": "native"}}},
				},
			})
		}))
		defer srv.Close()
		c := newHTTPClient(t, Config{
			APIKey: "sk-test", Model: DefaultModel, APIURL: srv.URL, UseNativeAPI: true,
		})
		got, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "Hello"}})
		if err != nil {
			t.Fatal(err)
		}
		if got != "native" || path != "/" || auth != "Bearer sk-test" {
			t.Fatalf("reply %q path %s auth %s", got, path, auth)
		}
		if _, ok := body["enable_thinking"]; ok {
			t.Fatalf("native body included enable_thinking: %#v", body)
		}
		input, _ := body["input"].(map[string]any)
		if input == nil {
			t.Fatalf("body = %#v", body)
		}
	})
}

func TestNamedOpenAIFamilyAdapters(t *testing.T) {
	presets := []struct {
		id       ProviderID
		key      string
		auth     string
		thinking bool
		path     string
	}{
		{ProviderOpenAI, "sk-openai", "Bearer sk-openai", false, "/chat/completions"},
		{ProviderXAI, "sk-xai", "Bearer sk-xai", false, "/chat/completions"},
		{ProviderGemini, "sk-gem", "Bearer sk-gem", false, "/chat/completions"},
		{ProviderOpenRouter, "sk-or", "Bearer sk-or", false, "/chat/completions"},
		{ProviderOllama, "", "", false, "/v1/chat/completions"},
		{ProviderDashScope, "sk-ds", "Bearer sk-ds", true, "/chat/completions"},
		{ProviderMistral, "sk-mi", "Bearer sk-mi", false, "/chat/completions"},
		{ProviderGroq, "sk-groq", "Bearer sk-groq", false, "/chat/completions"},
		{ProviderOpenAICompatible, "sk-compat", "Bearer sk-compat", false, "/chat/completions"},
	}
	for _, preset := range presets {
		preset := preset
		t.Run(string(preset.id), func(t *testing.T) {
			t.Run("chat", func(t *testing.T) {
				var path, auth string
				var body map[string]any
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					path = r.URL.Path
					auth = r.Header.Get("Authorization")
					body = readJSONBody(t, r)
					writeJSON(w, http.StatusOK, openAIText("hello"))
				}))
				defer srv.Close()
				base := srv.URL
				if preset.id == ProviderOllama {
					base = srv.URL // normalized to /v1
				}
				cfg := Config{Provider: preset.id, APIKey: preset.key, Model: "unit-model", BaseURL: base}
				c := newHTTPClient(t, cfg)
				got, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "Hi"}})
				if err != nil {
					t.Fatal(err)
				}
				if got != "hello" || path != preset.path || auth != preset.auth {
					t.Fatalf("reply %q path %s auth %q", got, path, auth)
				}
				_, hasThinking := body["enable_thinking"]
				if hasThinking != preset.thinking {
					t.Fatalf("enable_thinking present = %v, want %v (%#v)", hasThinking, preset.thinking, body["enable_thinking"])
				}
				if preset.thinking && body["enable_thinking"] != false {
					t.Fatalf("thinking = %#v", body["enable_thinking"])
				}
			})

			t.Run("stream", func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body := readJSONBody(t, r)
					if body["stream"] != true {
						t.Errorf("stream flag = %#v", body["stream"])
					}
					w.Header().Set("Content-Type", "text/event-stream")
					_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n")
					_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\" world\"},\"finish_reason\":\"stop\"}]}\n\n")
					_, _ = io.WriteString(w, "data: [DONE]\n\n")
				}))
				defer srv.Close()
				c := newHTTPClient(t, Config{Provider: preset.id, APIKey: preset.key, Model: "unit-model", BaseURL: srv.URL})
				ch, err := c.Stream(context.Background(), CompletionRequest{
					Messages: []Message{{Role: "user", Content: "Hi"}},
				})
				if err != nil {
					t.Fatal(err)
				}
				resp, err := CollectStream(ch)
				if err != nil {
					t.Fatal(err)
				}
				if resp.Content != "Hello world" || resp.FinishReason != "stop" {
					t.Fatalf("stream = %+v", resp)
				}
			})

			t.Run("tools", func(t *testing.T) {
				var calls int
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body := readJSONBody(t, r)
					calls++
					if calls == 1 {
						tools, _ := body["tools"].([]any)
						if len(tools) != 1 {
							t.Errorf("tools = %#v", body["tools"])
						}
						writeJSON(w, http.StatusOK, map[string]any{
							"choices": []any{map[string]any{
								"finish_reason": "tool_calls",
								"message": map[string]any{
									"role":    "assistant",
									"content": nil,
									"tool_calls": []any{map[string]any{
										"id":   "call_1",
										"type": "function",
										"function": map[string]any{
											"name":      "get_weather",
											"arguments": "{\"city\":\"Paris\"}",
										},
									}},
								},
							}},
						})
						return
					}
					msgs, _ := body["messages"].([]any)
					raw, _ := json.Marshal(msgs)
					if !strings.Contains(string(raw), `"role":"tool"`) || !strings.Contains(string(raw), "call_1") {
						t.Errorf("follow-up messages = %s", raw)
					}
					writeJSON(w, http.StatusOK, openAIText("72F"))
				}))
				defer srv.Close()
				c := newHTTPClient(t, Config{Provider: preset.id, APIKey: preset.key, Model: "unit-model", BaseURL: srv.URL})
				first, err := c.Complete(context.Background(), CompletionRequest{
					Messages: []Message{{Role: "user", Content: "weather?"}},
					Tools: []Tool{{
						Name:        "get_weather",
						Description: "weather",
						Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
					}},
				})
				if err != nil {
					t.Fatal(err)
				}
				if len(first.ToolCalls) != 1 || first.ToolCalls[0].Name != "get_weather" || first.ToolCalls[0].Arguments != `{"city":"Paris"}` {
					t.Fatalf("tool call = %+v", first.ToolCalls)
				}
				second, err := c.Complete(context.Background(), CompletionRequest{
					Messages: []Message{
						{Role: "user", Content: "weather?"},
						{Role: "assistant", ToolCalls: first.ToolCalls},
						{Role: "tool", ToolCallID: first.ToolCalls[0].ID, Content: "72F"},
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if second.Content != "72F" || calls != 2 {
					t.Fatalf("second %q calls %d", second.Content, calls)
				}
			})

			t.Run("error", func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					writeJSON(w, http.StatusUnauthorized, map[string]any{
						"error": map[string]any{"message": "bad key", "type": "invalid_request_error", "code": "invalid_api_key"},
					})
				}))
				defer srv.Close()
				c := newHTTPClient(t, Config{Provider: preset.id, APIKey: preset.key, Model: "unit-model", BaseURL: srv.URL})
				_, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
				var api *APIError
				if !errors.As(err, &api) {
					t.Fatalf("err = %v", err)
				}
				if api.StatusCode != 401 || api.Code != "invalid_api_key" || api.Message != "bad key" || api.Provider != preset.id {
					t.Fatalf("api = %+v", api)
				}
				if !strings.Contains(err.Error(), "API error (status 401): invalid_api_key - bad key") {
					t.Fatalf("error text = %v", err)
				}
			})
		})
	}
}

func TestOpenAIJSONModeAndNotConfigured(t *testing.T) {
	var sawFormat bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readJSONBody(t, r)
		fmt, _ := body["response_format"].(map[string]any)
		sawFormat = fmt["type"] == "json_object"
		writeJSON(w, http.StatusOK, openAIText(`{"ok":true}`))
	}))
	defer srv.Close()
	c := newHTTPClient(t, Config{Provider: ProviderOpenAI, APIKey: "k", Model: "gpt-4.1", BaseURL: srv.URL})
	resp, err := c.Complete(context.Background(), CompletionRequest{
		Messages: []Message{{Role: "user", Content: "json"}},
		JSONMode: true,
	})
	if err != nil || !sawFormat || resp.Content != `{"ok":true}` {
		t.Fatalf("resp %q err %v format %v", resp.Content, err, sawFormat)
	}

	bare := newHTTPClient(t, Config{Provider: ProviderOpenAI, Model: "gpt-4.1"})
	bare.httpClient.Transport = denyTransport{t}
	t.Setenv("OPENAI_API_KEY", "")
	_, err = bare.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("err = %v", err)
	}
}

func TestAnthropicAdapter(t *testing.T) {
	t.Run("chat auth and system", func(t *testing.T) {
		var path, key, version string
		var body map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			key = r.Header.Get("x-api-key")
			version = r.Header.Get("anthropic-version")
			if r.Header.Get("Authorization") != "" {
				t.Errorf("bearer set: %s", r.Header.Get("Authorization"))
			}
			body = readJSONBody(t, r)
			writeJSON(w, http.StatusOK, map[string]any{
				"stop_reason": "end_turn",
				"content":     []any{map[string]any{"type": "text", "text": "Hello"}},
			})
		}))
		defer srv.Close()
		c := newHTTPClient(t, Config{Provider: ProviderAnthropic, APIKey: "sk-ant", Model: "claude-sonnet-5", BaseURL: srv.URL})
		got, err := c.ChatCompletion(context.Background(), []Message{
			{Role: "system", Content: "be brief"},
			{Role: "user", Content: "Hi"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got != "Hello" || path != "/v1/messages" || key != "sk-ant" || version != "2023-06-01" {
			t.Fatalf("reply %q path %s key %s version %s", got, path, key, version)
		}
		if body["system"] != "be brief" || body["max_tokens"] == nil {
			t.Fatalf("body = %#v", body)
		}
	})

	t.Run("tools", func(t *testing.T) {
		var calls int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := readJSONBody(t, r)
			calls++
			if calls == 1 {
				tools, _ := body["tools"].([]any)
				if len(tools) != 1 {
					t.Fatalf("tools = %#v", body["tools"])
				}
				tool := tools[0].(map[string]any)
				if tool["name"] != "get_weather" || tool["input_schema"] == nil {
					t.Fatalf("tool = %#v", tool)
				}
				writeJSON(w, http.StatusOK, map[string]any{
					"stop_reason": "tool_use",
					"content": []any{
						map[string]any{"type": "text", "text": "Checking."},
						map[string]any{"type": "tool_use", "id": "toolu_1", "name": "get_weather", "input": map[string]any{"city": "Paris"}},
					},
				})
				return
			}
			raw, _ := json.Marshal(body["messages"])
			if !strings.Contains(string(raw), "tool_result") || !strings.Contains(string(raw), "toolu_1") {
				t.Fatalf("follow-up = %s", raw)
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"stop_reason": "end_turn",
				"content":     []any{map[string]any{"type": "text", "text": "72F"}},
			})
		}))
		defer srv.Close()
		c := newHTTPClient(t, Config{Provider: ProviderAnthropic, APIKey: "sk-ant", Model: "claude-sonnet-5", BaseURL: srv.URL})
		first, err := c.Complete(context.Background(), CompletionRequest{
			Messages: []Message{{Role: "user", Content: "weather?"}},
			Tools:    []Tool{{Name: "get_weather", Description: "weather", Parameters: json.RawMessage(`{"type":"object"}`)}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if first.FinishReason != "tool_calls" || len(first.ToolCalls) != 1 || first.ToolCalls[0].ID != "toolu_1" {
			t.Fatalf("first = %+v", first)
		}
		if !strings.Contains(first.ToolCalls[0].Arguments, "Paris") {
			t.Fatalf("args = %s", first.ToolCalls[0].Arguments)
		}
		second, err := c.Complete(context.Background(), CompletionRequest{
			Messages: []Message{
				{Role: "user", Content: "weather?"},
				{Role: "assistant", Content: first.Content, ToolCalls: first.ToolCalls},
				{Role: "tool", ToolCallID: first.ToolCalls[0].ID, Content: "72F"},
			},
		})
		if err != nil || second.Content != "72F" || calls != 2 {
			t.Fatalf("second %q err %v calls %d", second.Content, err, calls)
		}
	})

	t.Run("stream", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := readJSONBody(t, r)
			if body["stream"] != true {
				t.Errorf("stream = %#v", body["stream"])
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\n")
			_, _ = io.WriteString(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":1,\"content_block\":{\"type\":\"tool_use\",\"id\":\"toolu_9\",\"name\":\"get_weather\",\"input\":{}}}\n\n")
			_, _ = io.WriteString(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":1,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"city\\\":\\\"Paris\\\"}\"}}\n\n")
			_, _ = io.WriteString(w, "event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"tool_use\"}}\n\n")
			_, _ = io.WriteString(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		}))
		defer srv.Close()
		c := newHTTPClient(t, Config{Provider: ProviderAnthropic, APIKey: "sk-ant", Model: "claude-sonnet-5", BaseURL: srv.URL})
		ch, err := c.Stream(context.Background(), CompletionRequest{Messages: []Message{{Role: "user", Content: "Hi"}}})
		if err != nil {
			t.Fatal(err)
		}
		resp, err := CollectStream(ch)
		if err != nil {
			t.Fatal(err)
		}
		if resp.Content != "Hello" || resp.FinishReason != "tool_calls" || len(resp.ToolCalls) != 1 {
			t.Fatalf("stream = %+v", resp)
		}
		if resp.ToolCalls[0].ID != "toolu_9" || resp.ToolCalls[0].Name != "get_weather" || resp.ToolCalls[0].Arguments != `{"city":"Paris"}` {
			t.Fatalf("tool = %+v", resp.ToolCalls[0])
		}
	})

	t.Run("vision", func(t *testing.T) {
		var body map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body = readJSONBody(t, r)
			writeJSON(w, http.StatusOK, map[string]any{
				"stop_reason": "end_turn",
				"content":     []any{map[string]any{"type": "text", "text": "a png"}},
			})
		}))
		defer srv.Close()
		c := newHTTPClient(t, Config{Provider: ProviderAnthropic, APIKey: "sk-ant", Model: "claude-sonnet-5", BaseURL: srv.URL})
		got, err := c.ChatCompletionVision(context.Background(), []MultiMessage{
			UserMultiMessage(TextPart("describe"), ImageDataPart("image/png", []byte{0x89, 0x50, 0x4e, 0x47})),
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "a png" {
			t.Fatal(got)
		}
		raw, _ := json.Marshal(body["messages"])
		if !strings.Contains(string(raw), `"type":"image"`) || !strings.Contains(string(raw), `"type":"base64"`) || !strings.Contains(string(raw), "image/png") {
			t.Fatalf("messages = %s", raw)
		}
	})

	t.Run("json mode and missing key", func(t *testing.T) {
		c := newHTTPClient(t, Config{Provider: ProviderAnthropic, APIKey: "sk-ant", Model: "claude-sonnet-5"})
		c.httpClient.Transport = denyTransport{t}
		_, err := c.Complete(context.Background(), CompletionRequest{
			Messages: []Message{{Role: "user", Content: "json"}},
			JSONMode: true,
		})
		if !errors.Is(err, ErrCapabilityUnsupported) {
			t.Fatalf("err = %v", err)
		}
		t.Setenv("ANTHROPIC_API_KEY", "")
		bare := newHTTPClient(t, Config{Provider: ProviderAnthropic, Model: "claude-sonnet-5"})
		bare.httpClient.Transport = denyTransport{t}
		_, err = bare.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("error mapping", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"type": "error",
				"error": map[string]any{
					"type":    "authentication_error",
					"message": "invalid x-api-key",
				},
			})
		}))
		defer srv.Close()
		c := newHTTPClient(t, Config{Provider: ProviderAnthropic, APIKey: "sk-ant", Model: "claude-sonnet-5", BaseURL: srv.URL})
		_, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		var api *APIError
		if !errors.As(err, &api) || api.Code != "authentication_error" || api.Message != "invalid x-api-key" || api.StatusCode != 401 {
			t.Fatalf("err = %v api %+v", err, api)
		}
	})
}

func TestNativeDashScopeCapabilitiesAndErrors(t *testing.T) {
	c := newHTTPClient(t, Config{APIKey: "k", Model: DefaultModel, APIURL: DefaultAPIURL, UseNativeAPI: true})
	c.httpClient.Transport = denyTransport{t}
	c.httpClientLong.Transport = denyTransport{t}

	_, err := c.ChatCompletionVision(context.Background(), []MultiMessage{UserMultiMessage(TextPart("hi"))}, "")
	if err == nil || !strings.Contains(err.Error(), "MORPH_AI_API_URL") || !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("vision err = %v", err)
	}
	_, err = c.Stream(context.Background(), CompletionRequest{Messages: []Message{{Role: "user", Content: "hi"}}})
	if !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("stream err = %v", err)
	}
	_, err = c.Complete(context.Background(), CompletionRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
		Tools:    []Tool{{Name: "t"}},
	})
	if !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("tools err = %v", err)
	}
	_, err = c.Complete(context.Background(), CompletionRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
		JSONMode: true,
	})
	if !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("json err = %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": "InvalidParameter", "message": "bad model"})
	}))
	defer srv.Close()
	live := newHTTPClient(t, Config{APIKey: "k", Model: DefaultModel, APIURL: srv.URL, UseNativeAPI: true})
	_, err = live.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	var api *APIError
	if !errors.As(err, &api) || api.Code != "InvalidParameter" || api.Message != "bad model" || api.StatusCode != 400 {
		t.Fatalf("err = %v api %+v", err, api)
	}
}

func TestOpenAIVisionRequestShape(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body = readJSONBody(t, r)
		writeJSON(w, http.StatusOK, openAIText("a chart"))
	}))
	defer srv.Close()
	c := newHTTPClient(t, Config{Provider: ProviderOpenAI, APIKey: "k", BaseURL: srv.URL})
	got, err := c.ChatCompletionVision(context.Background(), []MultiMessage{
		UserMultiMessage(TextPart("describe"), ImageDataPart("image/png", []byte{1, 2, 3})),
	}, "gpt-4o")
	if err != nil {
		t.Fatal(err)
	}
	if got != "a chart" || body["model"] != "gpt-4o" {
		t.Fatalf("reply %q model %#v", got, body["model"])
	}
	raw, _ := json.Marshal(body["messages"])
	if !strings.Contains(string(raw), "image_url") || !strings.Contains(string(raw), "data:image/png;base64,") {
		t.Fatalf("messages = %s", raw)
	}
}

func TestStreamingToolDeltasOpenAI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call_1\",\"function\":{\"name\":\"get_weather\",\"arguments\":\"\"}}]}}]}\n\n")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"{\\\"city\\\":\\\"Paris\\\"}\"}}]},\"finish_reason\":\"tool_calls\"}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()
	c := newHTTPClient(t, Config{Provider: ProviderXAI, APIKey: "k", Model: "grok-4.3", BaseURL: srv.URL})
	ch, err := c.Stream(context.Background(), CompletionRequest{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := CollectStream(ch)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Name != "get_weather" || resp.ToolCalls[0].Arguments != `{"city":"Paris"}` || resp.FinishReason != "tool_calls" {
		t.Fatalf("resp = %+v", resp)
	}
}
