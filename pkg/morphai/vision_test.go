package morphai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestMultiMessageSerializationShape(t *testing.T) {
	msg := UserMultiMessage(
		TextPart("describe this"),
		ImageDataPart("image/png", []byte{0x89, 0x50, 0x4e, 0x47}),
	)

	raw, err := json.Marshal([]MultiMessage{msg})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(raw)

	for _, want := range []string{
		`"role":"user"`,
		`"type":"text"`,
		`"text":"describe this"`,
		`"type":"image_url"`,
		`"image_url":{"url":"data:image/png;base64,`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("payload missing %s\ngot: %s", want, got)
		}
	}

	// A text part must not carry an empty image_url key, and vice versa.
	if strings.Contains(got, `"type":"text","image_url"`) {
		t.Errorf("text part leaked an image_url field: %s", got)
	}
	if strings.Contains(got, `"type":"image_url","text"`) {
		t.Errorf("image part leaked a text field: %s", got)
	}
}

func TestDataURLDefaultsToJPEGForNonImageMime(t *testing.T) {
	if got := DataURL("application/octet-stream", []byte{1, 2, 3}); !strings.HasPrefix(got, "data:image/jpeg;base64,") {
		t.Fatalf("got %q", got)
	}
	if got := DataURL("image/webp", []byte{1}); !strings.HasPrefix(got, "data:image/webp;base64,") {
		t.Fatalf("got %q", got)
	}
}

func TestChatCompletionVisionRejectsNativeEndpoint(t *testing.T) {
	c := NewClient(Config{
		APIKey:       "test-key",
		Model:        DefaultModel,
		APIURL:       DefaultAPIURL,
		BaseURL:      DefaultBaseURL,
		UseNativeAPI: true,
	})

	_, err := c.ChatCompletionVision(context.Background(),
		[]MultiMessage{UserMultiMessage(TextPart("hi"))}, "")
	if err == nil {
		t.Fatal("expected an error on the native endpoint")
	}
	if !strings.Contains(err.Error(), "MORPH_AI_API_URL") {
		t.Errorf("error should name the setting to change, got: %v", err)
	}
}

func TestChatCompletionVisionRequiresAPIKey(t *testing.T) {
	c := NewClient(Config{Model: DefaultModel, BaseURL: DefaultBaseURL})

	_, err := c.ChatCompletionVision(context.Background(),
		[]MultiMessage{UserMultiMessage(TextPart("hi"))}, "")
	if err == nil || !strings.Contains(err.Error(), "MORPH_AI_API_KEY") {
		t.Fatalf("expected a missing-key error, got: %v", err)
	}
}

func TestChatCompletionVisionRequiresMessages(t *testing.T) {
	c := NewClient(Config{APIKey: "k", BaseURL: DefaultBaseURL})
	if _, err := c.ChatCompletionVision(context.Background(), nil, ""); err == nil {
		t.Fatal("expected an error for no messages")
	}
}

func TestVisionModelResolution(t *testing.T) {
	if got := (Config{}).VisionModelOrDefault(); got != DefaultVisionModel {
		t.Errorf("empty config: got %q, want %q", got, DefaultVisionModel)
	}
	if got := (Config{VisionModel: " custom-vl "}).VisionModelOrDefault(); got != "custom-vl" {
		t.Errorf("configured model: got %q", got)
	}
	// The text model must never stand in for a vision model.
	if got := (Config{Model: "qwen3-max"}).VisionModelOrDefault(); got != DefaultVisionModel {
		t.Errorf("chat model leaked into vision resolution: %q", got)
	}
}

func TestVisionSupported(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want bool
	}{
		{"configured compatible", Config{APIKey: "k"}, true},
		{"configured native", Config{APIKey: "k", UseNativeAPI: true}, false},
		{"unconfigured", Config{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.VisionSupported(); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadFromEnvReadsVisionModel(t *testing.T) {
	t.Setenv("MORPH_AI_API_KEY", "k")
	t.Setenv("MORPH_AI_VISION_MODEL", "qwen-vl-plus")

	cfg := LoadFromEnv()
	if cfg.VisionModel != "qwen-vl-plus" {
		t.Fatalf("got %q", cfg.VisionModel)
	}
	if cfg.VisionModelOrDefault() != "qwen-vl-plus" {
		t.Fatalf("resolution ignored the env value")
	}
}

func TestLoadFromEnvVisionModelOptional(t *testing.T) {
	t.Setenv("MORPH_AI_API_KEY", "k")
	t.Setenv("MORPH_AI_VISION_MODEL", "")
	t.Setenv("TRAN_QWEN_VISION_MODEL", "")

	cfg := LoadFromEnv()
	if cfg.VisionModel != "" {
		t.Fatalf("expected empty vision model, got %q", cfg.VisionModel)
	}
	if cfg.VisionModelOrDefault() != DefaultVisionModel {
		t.Fatalf("expected the default, got %q", cfg.VisionModelOrDefault())
	}
	if !cfg.VisionSupported() {
		t.Fatal("vision should work with just an API key")
	}
}

func TestNilClientVisionHelpers(t *testing.T) {
	var c *Client
	if c.VisionSupported() {
		t.Error("nil client should not report vision support")
	}
	if c.VisionModel() != "" {
		t.Error("nil client should report no vision model")
	}
}
