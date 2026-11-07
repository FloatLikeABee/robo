package dorphai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func replyTransport(t *testing.T, wantURL string, check func(*http.Request, map[string]any)) roundTripFunc {
	t.Helper()
	var calls int
	return func(r *http.Request) (*http.Response, error) {
		calls++
		if calls != 1 {
			t.Errorf("request %d to %s", calls, r.URL.String())
		}
		if r.URL.String() != wantURL {
			t.Errorf("url = %s, want %s", r.URL.String(), wantURL)
		}
		body := readJSONBody(t, r)
		if check != nil {
			check(r, body)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)),
			Request:    r,
		}, nil
	}
}

func TestSiliconFlowBesideLocalProfiles(t *testing.T) {
	const siliconKey = "sf-key"
	const dorphKey = "dorph-key"
	const dashURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"

	t.Run("env load uses siliconflow model and default host", func(t *testing.T) {
		clearProviderEnv(t)
		decoy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("dashscope decoy received %s", r.URL.Path)
		}))
		defer decoy.Close()
		t.Setenv("DORPH_AI_PROVIDER", "siliconflow")
		t.Setenv("SILICONFLOW_API_KEY", siliconKey)
		t.Setenv("SILICONFLOW_MODEL", "Qwen/Qwen3.6-27B")
		t.Setenv("DORPH_AI_API_KEY", dorphKey)
		t.Setenv("DORPH_AI_MODEL", "qwen3.7-plus")
		t.Setenv("DORPH_AI_BASE_URL", decoy.URL)
		c := newHTTPClient(t, LoadFromEnv())
		c.httpClient.Transport = replyTransport(t, "https://api.siliconflow.cn/v1/chat/completions", func(r *http.Request, body map[string]any) {
			if r.Header.Get("Authorization") != "Bearer "+siliconKey {
				t.Errorf("auth = %q", r.Header.Get("Authorization"))
			}
			if body["model"] != "Qwen/Qwen3.6-27B" {
				t.Errorf("model = %#v", body["model"])
			}
			if body["enable_thinking"] != false {
				t.Errorf("enable_thinking = %#v", body["enable_thinking"])
			}
		})
		got, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		if err != nil {
			t.Fatal(err)
		}
		if got != "ok" {
			t.Fatalf("reply = %q", got)
		}
	})

	t.Run("openai slot does not authorize siliconflow", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("DORPH_AI_PROVIDER", "siliconflow")
		t.Setenv("OPENAI_API_KEY", "openai-key")
		t.Setenv("OPENAI_BASE_URL", "https://api.siliconflow.cn/v1")
		t.Setenv("DORPH_AI_API_KEY", dorphKey)
		c := newHTTPClient(t, LoadFromEnv())
		c.httpClient.Transport = denyTransport{t: t}
		_, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
		if !errors.Is(err, ErrProviderNotConfigured) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("dashscope still uses the parked dorph profile", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("DORPH_AI_PROVIDER", "dashscope")
		t.Setenv("DORPH_AI_API_KEY", dorphKey)
		t.Setenv("DORPH_AI_MODEL", "qwen3.7-plus")
		t.Setenv("DORPH_AI_BASE_URL", dashURL)
		t.Setenv("SILICONFLOW_API_KEY", siliconKey)
		c := newHTTPClient(t, LoadFromEnv())
		c.httpClient.Transport = replyTransport(t, dashURL+"/chat/completions", func(r *http.Request, body map[string]any) {
			if r.Header.Get("Authorization") != "Bearer "+dorphKey {
				t.Errorf("auth = %q", r.Header.Get("Authorization"))
			}
			if body["model"] != "qwen3.7-plus" {
				t.Errorf("model = %#v", body["model"])
			}
		})
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ollama ignores the siliconflow key", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("DORPH_AI_PROVIDER", "ollama")
		t.Setenv("SILICONFLOW_API_KEY", siliconKey)
		t.Setenv("DORPH_AI_API_KEY", dorphKey)
		t.Setenv("DORPH_AI_BASE_URL", dashURL)
		c := newHTTPClient(t, LoadFromEnv())
		c.httpClient.Transport = replyTransport(t, "http://127.0.0.1:11434/v1/chat/completions", func(r *http.Request, body map[string]any) {
			if r.Header.Get("Authorization") != "" {
				t.Errorf("auth = %q", r.Header.Get("Authorization"))
			}
		})
		if _, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("hand-built model beats SILICONFLOW_MODEL", func(t *testing.T) {
		clearProviderEnv(t)
		t.Setenv("SILICONFLOW_MODEL", "Qwen/Qwen3.6-27B")
		rc, err := resolve(Config{Provider: ProviderSiliconFlow, APIKey: siliconKey, Model: "caller-model"})
		if err != nil {
			t.Fatal(err)
		}
		if rc.Model != "caller-model" {
			t.Fatalf("model = %q", rc.Model)
		}
	})
}
