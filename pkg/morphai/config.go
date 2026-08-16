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
type Config struct {
	APIKey       string
	Model        string
	VisionModel  string // multimodal model for image reading
	APIURL       string // native DashScope text-generation endpoint
	BaseURL      string // OpenAI-compatible endpoint (chat + TranMail composer)
	UseNativeAPI bool   // true when MORPH_AI_API_URL is explicitly set
}

// LoadFromEnv reads unified MorphAI settings from the environment.
//
// Primary variables:
//   - MORPH_AI_API_KEY
//   - MORPH_AI_MODEL (default qwen3-max)
//   - MORPH_AI_API_URL (optional native endpoint override)
//   - MORPH_AI_BASE_URL (optional compatible-mode base URL)
//
// Legacy fallbacks: GEMINI_API_KEY, GEMINI_MODEL, TRAN_QWEN_*.
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

	return Config{
		APIKey:       strings.TrimSpace(apiKey),
		Model:        strings.TrimSpace(model),
		VisionModel:  strings.TrimSpace(visionModel),
		APIURL:       strings.TrimRight(strings.TrimSpace(apiURL), "/"),
		BaseURL:      strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		UseNativeAPI: useNative,
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
	return c.Configured() && !c.UseNativeAPI
}

func isNativeDashScopeURL(u string) bool {
	u = strings.ToLower(strings.TrimSpace(u))
	return strings.Contains(u, "text-generation/generation")
}

func (c Config) Configured() bool {
	return strings.TrimSpace(c.APIKey) != ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
