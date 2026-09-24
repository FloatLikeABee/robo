package morphai

import (
	"os"
	"strings"
)

const (
	DefaultModel       = "qwen3-max"
	DefaultVisionModel = "qwen-vl-max"
	DefaultAPIURL      = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	DefaultBaseURL     = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

// Config holds shared MorphAI model settings used across satellite apps.
//
// Leave Provider empty for the legacy single-endpoint path (today's callers).
// Set Provider, and optionally Model, APIKey, and BaseURL, to select a named
// provider. A later settings store can persist this struct per user or admin
// and pass it to NewClient or CompletionRequest without another code change.
type Config struct {
	Provider     ProviderID // empty = legacy DashScope / MORPH_AI_* behavior
	APIKey       string
	Model        string
	VisionModel  string // multimodal model for image reading
	APIURL       string // native DashScope text-generation endpoint
	BaseURL      string // OpenAI-compatible endpoint (chat + TranMail composer)
	UseNativeAPI bool   // true when MORPH_AI_API_URL is explicitly set
	// legacyKey and legacyBase are the values LoadFromEnv copied from the
	// legacy env vars. resolveLegacy is the only caller that sends them.
	// A named provider ignores a public field that still equals its snapshot.
	legacyKey  string
	legacyBase string
}

// LoadFromEnv reads unified MorphAI settings from the environment.
//
// Primary variables:
//   - MORPH_AI_API_KEY
//   - MORPH_AI_MODEL (default qwen3-max)
//   - MORPH_AI_API_URL (optional native endpoint override)
//   - MORPH_AI_BASE_URL (optional compatible-mode base URL)
//   - MORPH_AI_VISION_MODEL (optional; default qwen-vl-max)
//
// Legacy fallbacks: GEMINI_API_KEY, GEMINI_MODEL, TRAN_QWEN_*.
//
// Provider is left empty. Per-provider variables (OPENAI_API_KEY,
// ANTHROPIC_API_KEY, and the rest) are not consulted here and do not change
// which backend this config talks to.
func LoadFromEnv() Config {
	apiKey := firstNonEmpty(
		os.Getenv("MORPH_AI_API_KEY"),
		os.Getenv("GEMINI_API_KEY"),
		os.Getenv("TRAN_QWEN_API_KEY"),
	)
	model := firstNonEmpty(
		os.Getenv("MORPH_AI_MODEL"),
		os.Getenv("GEMINI_MODEL"),
		os.Getenv("TRAN_QWEN_MODEL"),
	)
	if model == "" {
		model = DefaultModel
	}

	apiURLEnv := strings.TrimSpace(os.Getenv("MORPH_AI_API_URL"))
	baseURLEnv := firstNonEmpty(
		os.Getenv("MORPH_AI_BASE_URL"),
		os.Getenv("TRAN_QWEN_BASE_URL"),
	)

	apiURL := apiURLEnv
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}
	baseURL := baseURLEnv
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	useNative := isNativeDashScopeURL(apiURLEnv)
	if apiURLEnv != "" && !useNative {
		// MORPH_AI_API_URL points at an OpenAI-compatible provider (e.g. SiliconFlow /v1).
		baseURL = apiURLEnv
		useNative = false
	}

	visionModel := firstNonEmpty(
		os.Getenv("MORPH_AI_VISION_MODEL"),
		os.Getenv("TRAN_QWEN_VISION_MODEL"),
	)

	key := strings.TrimSpace(apiKey)
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return Config{
		APIKey:       key,
		Model:        strings.TrimSpace(model),
		VisionModel:  strings.TrimSpace(visionModel),
		APIURL:       strings.TrimRight(strings.TrimSpace(apiURL), "/"),
		BaseURL:      base,
		UseNativeAPI: useNative,
		legacyKey:    key,
		legacyBase:   base,
	}
}

// VisionModelOrDefault returns the configured multimodal model, falling back to
// DefaultVisionModel. The chat model is never used as a fallback: a text-only
// model would reject or silently ignore image content.
func (c Config) VisionModelOrDefault() string {
	if m := strings.TrimSpace(c.VisionModel); m != "" {
		return m
	}
	return DefaultVisionModel
}

// VisionSupported reports whether this configuration can serve image requests.
func (c Config) VisionSupported() bool {
	if strings.TrimSpace(string(c.Provider)) == "" {
		return c.Configured() && !c.UseNativeAPI
	}
	rc, err := resolve(c)
	if err != nil {
		return false
	}
	return rc.supports(capVision)
}

func isNativeDashScopeURL(u string) bool {
	u = strings.ToLower(strings.TrimSpace(u))
	return strings.Contains(u, "text-generation/generation")
}

// Configured reports whether a call can be made.
//
// An empty Provider is configured only when APIKey is set (the legacy rule).
// A named provider is configured when resolve finds a key, or a base URL for
// Ollama. Keys are not borrowed from a different provider.
func (c Config) Configured() bool {
	if strings.TrimSpace(string(c.Provider)) == "" {
		return strings.TrimSpace(c.APIKey) != ""
	}
	_, err := resolve(c)
	return err == nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
