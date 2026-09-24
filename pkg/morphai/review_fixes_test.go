package morphai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func clearProviderEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"MORPH_AI_API_KEY", "MORPH_AI_MODEL", "MORPH_AI_BASE_URL", "MORPH_AI_API_URL", "MORPH_AI_VISION_MODEL",
		"GEMINI_API_KEY", "GEMINI_MODEL",
		"TRAN_QWEN_API_KEY", "TRAN_QWEN_MODEL", "TRAN_QWEN_BASE_URL", "TRAN_QWEN_VISION_MODEL",
		"OPENAI_API_KEY", "OPENAI_BASE_URL",
		"DASHSCOPE_API_KEY", "DASHSCOPE_BASE_URL",
		"OLLAMA_BASE_URL",
		"ANTHROPIC_API_KEY",
	} {
		t.Setenv(name, "")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type closeNotify struct {
	io.Reader
	closed chan struct{}
	once   sync.Once
}

func (c *closeNotify) Close() error {
	c.once.Do(func() { close(c.closed) })
	return nil
}

func TestLegacyEnvKeyStaysOffOtherProviders(t *testing.T) {
	const legacyKey = "legacy-env-key-value"
	const callerKey = "caller-set-key-value"
	const openaiKey = "openai-env-key-value"

	t.Run("load from env then openai sends nothing", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_API_KEY", legacyKey)
		var hits int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits++
			if strings.Contains(r.Header.Get("Authorization"), legacyKey) {
				t.Error("request carried the legacy key")
			}
			writeJSON(w, http.StatusOK, openAIText("nope"))
		}))
		defer srv.Close()
		cfg := LoadFromEnv()
		cfg.Provider = ProviderOpenAI
		cfg.BaseURL = srv.URL
		c := newHTTPClient(t, cfg)
		_, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("err = %v", err)
		}
		if hits != 0 {
			t.Fatalf("hits = %d", hits)
		}
	})

	t.Run("openai env key is used instead of the legacy key", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_API_KEY", legacyKey)
		t.Setenv("OPENAI_API_KEY", openaiKey)
		var auth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth = r.Header.Get("Authorization")
			writeJSON(w, http.StatusOK, openAIText("ok"))
		}))
		defer srv.Close()
		cfg := LoadFromEnv()
		cfg.Provider = ProviderOpenAI
		cfg.BaseURL = srv.URL
		c := newHTTPClient(t, cfg)
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
		if auth != "Bearer "+openaiKey || strings.Contains(auth, legacyKey) {
			t.Fatal("authorization was not the openai env key")
		}
	})

	t.Run("caller key after load from env is honored", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_API_KEY", legacyKey)
		var auth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth = r.Header.Get("Authorization")
			writeJSON(w, http.StatusOK, openAIText("ok"))
		}))
		defer srv.Close()
		cfg := LoadFromEnv()
		cfg.Provider = ProviderOpenAI
		cfg.APIKey = callerKey
		cfg.BaseURL = srv.URL
		c := newHTTPClient(t, cfg)
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
		if auth != "Bearer "+callerKey || strings.Contains(auth, legacyKey) {
			t.Fatal("authorization was not the caller key")
		}
	})

	t.Run("per-call key is honored", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_API_KEY", legacyKey)
		var auth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth = r.Header.Get("Authorization")
			writeJSON(w, http.StatusOK, openAIText("ok"))
		}))
		defer srv.Close()
		cfg := LoadFromEnv()
		cfg.Provider = ProviderOpenAI
		cfg.BaseURL = srv.URL
		c := newHTTPClient(t, cfg)
		_, err := c.Complete(context.Background(), CompletionRequest{
			APIKey:   callerKey,
			Messages: []Message{{Role: "user", Content: "hi"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if auth != "Bearer "+callerKey || strings.Contains(auth, legacyKey) {
			t.Fatal("authorization was not the per-call key")
		}
	})

	t.Run("ollama does not send the legacy key", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("MORPH_AI_API_KEY", legacyKey)
		var auth string
		var hits int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits++
			auth = r.Header.Get("Authorization")
			writeJSON(w, http.StatusOK, openAIText("ok"))
		}))
		defer srv.Close()
		cfg := LoadFromEnv()
		cfg.Provider = ProviderOllama
		cfg.BaseURL = srv.URL
		c := newHTTPClient(t, cfg)
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
		if hits != 1 {
			t.Fatalf("hits = %d", hits)
		}
		if auth != "" || strings.Contains(auth, legacyKey) {
			t.Fatal("ollama request carried a legacy key")
		}
	})

	t.Run("explicit dashscope does not use the gemini fallback", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("GEMINI_API_KEY", legacyKey)
		var hits int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits++
			if strings.Contains(r.Header.Get("Authorization"), legacyKey) {
				t.Error("request carried the gemini fallback")
			}
			writeJSON(w, http.StatusOK, openAIText("nope"))
		}))
		defer srv.Close()
		cfg := LoadFromEnv()
		cfg.Provider = ProviderDashScope
		cfg.BaseURL = srv.URL
		c := newHTTPClient(t, cfg)
		_, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("err = %v", err)
		}
		if hits != 0 {
			t.Fatalf("hits = %d", hits)
		}
	})

	t.Run("explicit dashscope keeps the tran qwen env key but not a copied gemini key", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("TRAN_QWEN_API_KEY", legacyKey)
		var auth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth = r.Header.Get("Authorization")
			writeJSON(w, http.StatusOK, openAIText("ok"))
		}))
		defer srv.Close()
		cfg := LoadFromEnv()
		cfg.Provider = ProviderDashScope
		cfg.BaseURL = srv.URL
		c := newHTTPClient(t, cfg)
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
		if auth != "Bearer "+legacyKey {
			t.Fatal("authorization was not the dashscope tran qwen env key")
		}
	})
}

func TestCancelledStreamClosesBody(t *testing.T) {
	t.Run("openai", func(t *testing.T) {
		assertStreamCancelReleasesBody(t, ProviderOpenAI, openaiFlood(40))
	})
	t.Run("anthropic", func(t *testing.T) {
		assertStreamCancelReleasesBody(t, ProviderAnthropic, anthropicFlood(40))
	})
}

func assertStreamCancelReleasesBody(t *testing.T, provider ProviderID, payload string) {
	t.Helper()
	closed := make(chan struct{})
	body := &closeNotify{Reader: strings.NewReader(payload), closed: closed}
	c := newHTTPClient(t, Config{Provider: provider, APIKey: "stream-key", Model: "unit-model", BaseURL: "https://stream.example"})
	c.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       body,
			Request:    r,
		}, nil
	})
	before := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := c.Stream(ctx, CompletionRequest{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("response body was not closed after cancel")
	}
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before+2 && time.Now().Before(deadline) {
		runtime.Gosched()
		time.Sleep(10 * time.Millisecond)
	}
	if runtime.NumGoroutine() > before+2 {
		t.Fatalf("stream goroutine still running: before %d now %d", before, runtime.NumGoroutine())
	}
	// Drain so the test process does not leave the channel unread.
	go func() {
		for range ch {
		}
	}()
}

func openaiFlood(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("data: {\"choices\":[{\"delta\":{\"content\":\"x\"}}]}\n\n")
	}
	return b.String()
}

func anthropicFlood(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"x\"}}\n\n")
	}
	return b.String()
}

func TestPrematureStreamEOFIsAnError(t *testing.T) {
	t.Run("openai", func(t *testing.T) {
		c, srv := streamServer(t, ProviderOpenAI, "data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\n")
		defer srv.Close()
		ch, err := c.Stream(context.Background(), CompletionRequest{Messages: []Message{{Role: "user", Content: "hi"}}})
		if err != nil {
			t.Fatal(err)
		}
		_, err = CollectStream(ch)
		if err == nil || !errors.Is(err, errStreamTruncated) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("anthropic", func(t *testing.T) {
		body := "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hi\"}}\n\n"
		c, srv := streamServer(t, ProviderAnthropic, body)
		defer srv.Close()
		ch, err := c.Stream(context.Background(), CompletionRequest{Messages: []Message{{Role: "user", Content: "hi"}}})
		if err != nil {
			t.Fatal(err)
		}
		_, err = CollectStream(ch)
		if err == nil || !errors.Is(err, errStreamTruncated) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestEmptyStreamedToolArgsAreObjects(t *testing.T) {
	t.Run("openai", func(t *testing.T) {
		body := "" +
			"data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call_1\",\"function\":{\"name\":\"lookup\",\"arguments\":\"\"}}]}}]}\n\n" +
			"data: {\"choices\":[{\"finish_reason\":\"tool_calls\"}]}\n\n" +
			"data: [DONE]\n\n"
		c, srv := streamServer(t, ProviderOpenAI, body)
		defer srv.Close()
		ch, err := c.Stream(context.Background(), CompletionRequest{Messages: []Message{{Role: "user", Content: "hi"}}})
		if err != nil {
			t.Fatal(err)
		}
		resp, err := CollectStream(ch)
		if err != nil {
			t.Fatal(err)
		}
		if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Arguments != "{}" {
			t.Fatalf("args = %#v", resp.ToolCalls)
		}
	})
	t.Run("anthropic", func(t *testing.T) {
		body := "" +
			"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"tool_use\",\"id\":\"toolu_1\",\"name\":\"lookup\",\"input\":{}}}\n\n" +
			"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"tool_use\"}}\n\n" +
			"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"
		c, srv := streamServer(t, ProviderAnthropic, body)
		defer srv.Close()
		ch, err := c.Stream(context.Background(), CompletionRequest{Messages: []Message{{Role: "user", Content: "hi"}}})
		if err != nil {
			t.Fatal(err)
		}
		resp, err := CollectStream(ch)
		if err != nil {
			t.Fatal(err)
		}
		if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Arguments != "{}" {
			t.Fatalf("args = %#v", resp.ToolCalls)
		}
	})
}

func TestRateLimitHonorsCancel(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		writeJSON(w, http.StatusOK, openAIText("ok"))
	}))
	defer srv.Close()
	c := newHTTPClient(t, Config{APIKey: "k", Model: "m", BaseURL: srv.URL})
	c.minRequestInterval = 5 * time.Second
	if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, err := c.ChatCompletion(ctx, []Message{{Role: "user", Content: "hi"}})
	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Fatalf("cancelled call waited %s", elapsed)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if hits != 1 {
		t.Fatalf("hits = %d", hits)
	}
}

func TestLegacyBaseURLStaysOffNamedProviders(t *testing.T) {
	const openaiKey = "openai-key-for-base-provenance"

	t.Run("custom morph base is not openai", func(t *testing.T) {
		clearProviderEnv(t)
		var hitsA, hitsB int
		var authA string
		srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsA++
			authA = r.Header.Get("Authorization")
			writeJSON(w, http.StatusOK, openAIText("legacy"))
		}))
		defer srvA.Close()
		srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsB++
			if r.Header.Get("Authorization") != "Bearer "+openaiKey {
				t.Error("server B did not receive the openai key")
			}
			writeJSON(w, http.StatusOK, openAIText("openai"))
		}))
		defer srvB.Close()
		t.Setenv("MORPH_AI_API_KEY", "legacy-key-for-base-provenance")
		t.Setenv("MORPH_AI_BASE_URL", srvA.URL)
		t.Setenv("OPENAI_API_KEY", openaiKey)
		t.Setenv("OPENAI_BASE_URL", srvB.URL)
		cfg := LoadFromEnv()
		cfg.Provider = ProviderOpenAI
		c := newHTTPClient(t, cfg)
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
		if hitsA != 0 || strings.Contains(authA, openaiKey) {
			t.Fatal("legacy base received the openai call")
		}
		if hitsB != 1 {
			t.Fatalf("openai base hits = %d", hitsB)
		}
	})

	t.Run("custom morph base is not ollama", func(t *testing.T) {
		clearProviderEnv(t)
		var hitsA, hitsB int
		var authB string
		srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsA++
			writeJSON(w, http.StatusOK, openAIText("legacy"))
		}))
		defer srvA.Close()
		srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsB++
			authB = r.Header.Get("Authorization")
			writeJSON(w, http.StatusOK, openAIText("ollama"))
		}))
		defer srvB.Close()
		t.Setenv("MORPH_AI_API_KEY", "legacy-key-for-base-provenance")
		t.Setenv("MORPH_AI_BASE_URL", srvA.URL)
		t.Setenv("OLLAMA_BASE_URL", srvB.URL)
		cfg := LoadFromEnv()
		cfg.Provider = ProviderOllama
		c := newHTTPClient(t, cfg)
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
		if hitsA != 0 {
			t.Fatal("legacy base received the ollama call")
		}
		if hitsB != 1 || authB != "" {
			t.Fatal("ollama did not use its own base without a legacy key")
		}
	})

	t.Run("compatible mode morph api url is not openai", func(t *testing.T) {
		clearProviderEnv(t)
		var hitsA, hitsB int
		srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsA++
			if strings.Contains(r.Header.Get("Authorization"), openaiKey) {
				t.Error("openai key reached the compatible-mode legacy url")
			}
			writeJSON(w, http.StatusOK, openAIText("legacy"))
		}))
		defer srvA.Close()
		srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsB++
			writeJSON(w, http.StatusOK, openAIText("openai"))
		}))
		defer srvB.Close()
		t.Setenv("MORPH_AI_API_KEY", "legacy-key-for-base-provenance")
		t.Setenv("MORPH_AI_API_URL", srvA.URL+"/v1")
		t.Setenv("OPENAI_API_KEY", openaiKey)
		t.Setenv("OPENAI_BASE_URL", srvB.URL)
		cfg := LoadFromEnv()
		cfg.Provider = ProviderOpenAI
		c := newHTTPClient(t, cfg)
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
		if hitsA != 0 {
			t.Fatal("compatible-mode legacy url received the openai call")
		}
		if hitsB != 1 {
			t.Fatalf("openai base hits = %d", hitsB)
		}
	})

	t.Run("explicit config base url is honored", func(t *testing.T) {
		clearProviderEnv(t)
		var hitsLegacy, hitsExplicit int
		srvLegacy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsLegacy++
			writeJSON(w, http.StatusOK, openAIText("legacy"))
		}))
		defer srvLegacy.Close()
		srvExplicit := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitsExplicit++
			writeJSON(w, http.StatusOK, openAIText("explicit"))
		}))
		defer srvExplicit.Close()
		t.Setenv("MORPH_AI_API_KEY", "legacy-key-for-base-provenance")
		t.Setenv("MORPH_AI_BASE_URL", srvLegacy.URL)
		t.Setenv("OPENAI_API_KEY", openaiKey)
		t.Setenv("OPENAI_BASE_URL", "https://openai.example/v1")
		cfg := LoadFromEnv()
		cfg.Provider = ProviderOpenAI
		cfg.BaseURL = srvExplicit.URL
		c := newHTTPClient(t, cfg)
		got, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		if err != nil {
			t.Fatal(err)
		}
		if got != "explicit" || hitsExplicit != 1 || hitsLegacy != 0 {
			t.Fatalf("explicit hits %d legacy hits %d reply %q", hitsExplicit, hitsLegacy, got)
		}
	})
}

func TestRotatedEnvKeyIsNotTreatedAsCallerKey(t *testing.T) {
	clearProviderEnv(t)
	const loaded = "loaded-legacy-key-value"
	const rotated = "rotated-legacy-key-value"
	t.Setenv("MORPH_AI_API_KEY", loaded)
	cfg := LoadFromEnv()
	t.Setenv("MORPH_AI_API_KEY", rotated)
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		auth := r.Header.Get("Authorization")
		if strings.Contains(auth, loaded) || strings.Contains(auth, rotated) {
			t.Error("named provider received a legacy env key after rotation")
		}
		writeJSON(w, http.StatusOK, openAIText("nope"))
	}))
	defer srv.Close()
	cfg.Provider = ProviderOpenAI
	cfg.BaseURL = srv.URL
	c := newHTTPClient(t, cfg)
	_, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("err = %v", err)
	}
	if hits != 0 {
		t.Fatalf("hits = %d", hits)
	}
}

func TestGroqHasNoRetiredVisionDefault(t *testing.T) {
	info, ok := LookupProvider(ProviderGroq)
	if !ok {
		t.Fatal("missing groq")
	}
	if info.Capabilities.Vision {
		t.Fatal("groq advertises vision")
	}
	for _, retired := range []string{"llama-3.2-11b-vision-preview", "gemma2-9b-it", "llama-3.3-70b-versatile", "llama-3.1-8b-instant"} {
		for _, model := range info.SuggestedModels {
			if model == retired {
				t.Fatalf("suggested retired model %s", retired)
			}
		}
	}
	c := newHTTPClient(t, Config{Provider: ProviderGroq, APIKey: "k", Model: "openai/gpt-oss-20b"})
	c.httpClient.Transport = denyTransport{t}
	c.httpClientLong.Transport = denyTransport{t}
	_, err := c.ChatCompletionVision(context.Background(), []MultiMessage{UserMultiMessage(TextPart("hi"))}, "")
	if !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("err = %v", err)
	}
}

func TestOpenAIStreamRejectsHugeToolIndex(t *testing.T) {
	body := "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":65,\"id\":\"call_1\",\"function\":{\"name\":\"lookup\",\"arguments\":\"{}\"}}]}}]}\n\n" +
		"data: [DONE]\n\n"
	c, srv := streamServer(t, ProviderOpenAI, body)
	defer srv.Close()
	ch, err := c.Stream(context.Background(), CompletionRequest{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = CollectStream(ch)
	if err == nil || !strings.Contains(err.Error(), "tool call index") {
		t.Fatalf("err = %v", err)
	}
}

func TestOpenAIChatURLKeepsQuery(t *testing.T) {
	got := openAIChatURL("https://example.test/v1?api-version=2024-10-21")
	if got != "https://example.test/v1/chat/completions?api-version=2024-10-21" {
		t.Fatalf("url = %s", got)
	}
}

func streamServer(t *testing.T, provider ProviderID, body string) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, body)
	}))
	c := newHTTPClient(t, Config{Provider: provider, APIKey: "stream-key", Model: "unit-model", BaseURL: srv.URL})
	return c, srv
}
