package dorphai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNVIDIAChatTemplateThinking(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("DORPH_AI_ORCHESTRATOR_PROVIDER", "nvidia")
	t.Setenv("NVIDIA_API_KEY", "nv-key")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readJSONBody(t, r)
		kw, ok := body["chat_template_kwargs"].(map[string]any)
		if !ok {
			t.Fatalf("chat_template_kwargs missing: %#v", body)
		}
		if kw["enable_thinking"] != true {
			t.Fatalf("enable_thinking = %#v", kw["enable_thinking"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
	}))
	defer srv.Close()
	t.Setenv("NVIDIA_BASE_URL", srv.URL)
	cfg := LoadOrchestratorFromEnv()
	c := newHTTPClient(t, cfg)
	got, err := c.ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatal(err)
	}
	if got != "ok" {
		t.Fatalf("reply = %q", got)
	}
}

func TestNVIDIAStreamReasoningDelta(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("DORPH_AI_ORCHESTRATOR_PROVIDER", "nvidia")
	t.Setenv("NVIDIA_API_KEY", "nv-key")
	body := "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"think\"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"ans\"}}]}\n\ndata: [DONE]\n\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()
	t.Setenv("NVIDIA_BASE_URL", srv.URL)
	c := newHTTPClient(t, LoadOrchestratorFromEnv())
	ch, err := c.Stream(context.Background(), CompletionRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var reasoning, content strings.Builder
	for ev := range ch {
		if ev.Err != nil {
			t.Fatal(ev.Err)
		}
		reasoning.WriteString(ev.ReasoningDelta)
		content.WriteString(ev.Delta)
	}
	if reasoning.String() != "think" {
		t.Fatalf("reasoning = %q", reasoning.String())
	}
	if content.String() != "ans" {
		t.Fatalf("content = %q", content.String())
	}
}
