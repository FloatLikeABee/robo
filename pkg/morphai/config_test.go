package morphai

import (
	"os"
	"testing"
)

func TestLoadFromEnvPrimary(t *testing.T) {
	t.Setenv("MORPH_AI_API_KEY", "sk-test")
	t.Setenv("MORPH_AI_MODEL", "qwen3-max")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("TRAN_QWEN_API_KEY", "")

	cfg := LoadFromEnv()
	if cfg.APIKey != "sk-test" {
		t.Fatalf("APIKey = %q, want sk-test", cfg.APIKey)
	}
	if cfg.Model != "qwen3-max" {
		t.Fatalf("Model = %q, want qwen3-max", cfg.Model)
	}
}

func TestLoadFromEnvLegacyFallback(t *testing.T) {
	os.Unsetenv("MORPH_AI_API_KEY")
	os.Unsetenv("MORPH_AI_MODEL")
	t.Setenv("GEMINI_API_KEY", "legacy-key")
	t.Setenv("GEMINI_MODEL", "legacy-model")

	cfg := LoadFromEnv()
	if cfg.APIKey != "legacy-key" {
		t.Fatalf("APIKey = %q, want legacy-key", cfg.APIKey)
	}
	if cfg.Model != "legacy-model" {
		t.Fatalf("Model = %q, want legacy-model", cfg.Model)
	}
}

func TestLoadFromEnvDefaultModel(t *testing.T) {
	os.Unsetenv("MORPH_AI_MODEL")
	os.Unsetenv("GEMINI_MODEL")
	os.Unsetenv("TRAN_QWEN_MODEL")

	cfg := LoadFromEnv()
	if cfg.Model != DefaultModel {
		t.Fatalf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
}

func TestLoadFromEnvUseNativeAPI(t *testing.T) {
	os.Unsetenv("MORPH_AI_API_URL")
	os.Unsetenv("MORPH_AI_BASE_URL")
	cfg := LoadFromEnv()
	if cfg.UseNativeAPI {
		t.Fatal("UseNativeAPI should be false when MORPH_AI_API_URL is unset")
	}

	t.Setenv("MORPH_AI_API_URL", "https://example.com/native/services/aigc/text-generation/generation")
	cfg = LoadFromEnv()
	if !cfg.UseNativeAPI {
		t.Fatal("UseNativeAPI should be true for native DashScope URL")
	}
	if cfg.APIURL != "https://example.com/native/services/aigc/text-generation/generation" {
		t.Fatalf("APIURL = %q", cfg.APIURL)
	}

	os.Unsetenv("MORPH_AI_API_URL")
	t.Setenv("MORPH_AI_API_URL", "https://api.siliconflow.cn/v1")
	cfg = LoadFromEnv()
	if cfg.UseNativeAPI {
		t.Fatal("UseNativeAPI should be false for OpenAI-compatible /v1 URL")
	}
	if cfg.BaseURL != "https://api.siliconflow.cn/v1" {
		t.Fatalf("BaseURL = %q, want siliconflow /v1", cfg.BaseURL)
	}
}
