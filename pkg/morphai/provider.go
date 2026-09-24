package morphai

import "strings"

// ProviderID is the stable identifier stored by a future settings UI.
type ProviderID string

const (
	ProviderOpenAI           ProviderID = "openai"
	ProviderAnthropic        ProviderID = "anthropic"
	ProviderXAI              ProviderID = "xai"
	ProviderGemini           ProviderID = "gemini"
	ProviderOpenRouter       ProviderID = "openrouter"
	ProviderOllama           ProviderID = "ollama"
	ProviderDashScope        ProviderID = "dashscope"
	ProviderMistral          ProviderID = "mistral"
	ProviderGroq             ProviderID = "groq"
	ProviderOpenAICompatible ProviderID = "openai-compatible"
)

const (
	capChat   = "chat"
	capStream = "streaming"
	capTools  = "tools"
	capVision = "vision"
	capJSON   = "json_mode"
)

// Capabilities is what a provider's default protocol supports.
// A model behind that protocol may still reject a particular call.
type Capabilities struct {
	Chat     bool `json:"chat"`
	Stream   bool `json:"stream"`
	Tools    bool `json:"tools"`
	Vision   bool `json:"vision"`
	JSONMode bool `json:"json_mode"`
}

// ProviderInfo is the metadata a settings UI needs to render a provider picker.
type ProviderInfo struct {
	ID              ProviderID   `json:"id"`
	DisplayName     string       `json:"display_name"`
	RequiresKey     bool         `json:"requires_key"`
	DefaultBaseURL  string       `json:"default_base_url"`
	SuggestedModels []string     `json:"suggested_models"`
	Capabilities    Capabilities `json:"capabilities"`
	// APIKeyEnv and BaseURLEnv are the per-provider fallbacks used when the
	// matching Config fields are empty. They are not read for the legacy
	// (empty Provider) path.
	APIKeyEnv  string `json:"api_key_env,omitempty"`
	BaseURLEnv string `json:"base_url_env,omitempty"`
}

type adapterKind int

const (
	adapterOpenAI adapterKind = iota
	adapterAnthropic
	adapterDashScopeNative
)

func (k adapterKind) String() string {
	switch k {
	case adapterAnthropic:
		return "anthropic"
	case adapterDashScopeNative:
		return "dashscope-native"
	default:
		return "openai"
	}
}

type providerSpec struct {
	Info          ProviderInfo
	Kind          adapterKind
	DefaultVision string
	// ExtraKeyEnvs and ExtraBaseEnvs are consulted after the primary env var.
	// DashScope also accepts the historical MORPH_AI_* / TRAN_QWEN_* variables
	// because that is the same provider, not a fallback to a different one.
	ExtraKeyEnvs  []string
	ExtraBaseEnvs []string
}

var fullCaps = Capabilities{Chat: true, Stream: true, Tools: true, Vision: true, JSONMode: true}

var anthropicCaps = Capabilities{Chat: true, Stream: true, Tools: true, Vision: true, JSONMode: false}

// providerCatalog is the registry. Order is the settings-UI order.
var providerCatalog = []providerSpec{
	{
		Kind:          adapterOpenAI,
		DefaultVision: "gpt-4o",
		Info: ProviderInfo{
			ID:             ProviderOpenAI,
			DisplayName:    "OpenAI",
			RequiresKey:    true,
			DefaultBaseURL: "https://api.openai.com/v1",
			SuggestedModels: []string{
				"gpt-4.1",
				"gpt-4o",
				"gpt-4o-mini",
			},
			Capabilities: fullCaps,
			APIKeyEnv:    "OPENAI_API_KEY",
			BaseURLEnv:   "OPENAI_BASE_URL",
		},
	},
	{
		Kind:          adapterAnthropic,
		DefaultVision: "claude-sonnet-5",
		Info: ProviderInfo{
			ID:             ProviderAnthropic,
			DisplayName:    "Anthropic",
			RequiresKey:    true,
			DefaultBaseURL: "https://api.anthropic.com",
			SuggestedModels: []string{
				"claude-sonnet-5",
				"claude-opus-5",
				"claude-haiku-4-5",
			},
			Capabilities: anthropicCaps,
			APIKeyEnv:    "ANTHROPIC_API_KEY",
			BaseURLEnv:   "ANTHROPIC_BASE_URL",
		},
	},
	{
		Kind:          adapterOpenAI,
		DefaultVision: "grok-4.3",
		Info: ProviderInfo{
			ID:             ProviderXAI,
			DisplayName:    "xAI",
			RequiresKey:    true,
			DefaultBaseURL: "https://api.x.ai/v1",
			SuggestedModels: []string{
				"grok-4.3",
				"grok-4",
				"grok-3",
			},
			Capabilities: fullCaps,
			APIKeyEnv:    "XAI_API_KEY",
			BaseURLEnv:   "XAI_BASE_URL",
		},
	},
	{
		Kind:          adapterOpenAI,
		DefaultVision: "gemini-2.5-flash",
		Info: ProviderInfo{
			ID:             ProviderGemini,
			DisplayName:    "Google Gemini",
			RequiresKey:    true,
			DefaultBaseURL: "https://generativelanguage.googleapis.com/v1beta/openai",
			SuggestedModels: []string{
				"gemini-2.5-flash",
				"gemini-2.5-pro",
				"gemini-3.8-flash",
			},
			Capabilities: fullCaps,
			APIKeyEnv:    "GEMINI_API_KEY",
			BaseURLEnv:   "GEMINI_BASE_URL",
		},
	},
	{
		Kind:          adapterOpenAI,
		DefaultVision: "openai/gpt-4o",
		Info: ProviderInfo{
			ID:             ProviderOpenRouter,
			DisplayName:    "OpenRouter",
			RequiresKey:    true,
			DefaultBaseURL: "https://openrouter.ai/api/v1",
			SuggestedModels: []string{
				"openai/gpt-4.1",
				"anthropic/claude-sonnet-5",
				"google/gemini-2.5-flash",
			},
			Capabilities: fullCaps,
			APIKeyEnv:    "OPENROUTER_API_KEY",
			BaseURLEnv:   "OPENROUTER_BASE_URL",
		},
	},
	{
		Kind:          adapterOpenAI,
		DefaultVision: "llava",
		Info: ProviderInfo{
			ID:             ProviderOllama,
			DisplayName:    "Ollama",
			RequiresKey:    false,
			DefaultBaseURL: "http://127.0.0.1:11434/v1",
			SuggestedModels: []string{
				"llama3.3",
				"qwen2.5",
				"mistral",
			},
			Capabilities: fullCaps,
			BaseURLEnv:   "OLLAMA_BASE_URL",
		},
	},
	{
		Kind:          adapterOpenAI,
		DefaultVision: DefaultVisionModel,
		ExtraKeyEnvs:  []string{"MORPH_AI_API_KEY", "TRAN_QWEN_API_KEY"},
		ExtraBaseEnvs: []string{"MORPH_AI_BASE_URL", "TRAN_QWEN_BASE_URL"},
		Info: ProviderInfo{
			ID:             ProviderDashScope,
			DisplayName:    "DashScope",
			RequiresKey:    true,
			DefaultBaseURL: DefaultBaseURL,
			SuggestedModels: []string{
				DefaultModel,
				"qwen-plus",
				DefaultVisionModel,
			},
			Capabilities: fullCaps,
			APIKeyEnv:    "DASHSCOPE_API_KEY",
			BaseURLEnv:   "DASHSCOPE_BASE_URL",
		},
	},
	{
		Kind:          adapterOpenAI,
		DefaultVision: "pixtral-large-latest",
		Info: ProviderInfo{
			ID:             ProviderMistral,
			DisplayName:    "Mistral",
			RequiresKey:    true,
			DefaultBaseURL: "https://api.mistral.ai/v1",
			SuggestedModels: []string{
				"mistral-large-latest",
				"mistral-small-latest",
				"codestral-latest",
			},
			Capabilities: fullCaps,
			APIKeyEnv:    "MISTRAL_API_KEY",
			BaseURLEnv:   "MISTRAL_BASE_URL",
		},
	},
	{
		Kind:          adapterOpenAI,
		DefaultVision: "llama-3.2-11b-vision-preview",
		Info: ProviderInfo{
			ID:             ProviderGroq,
			DisplayName:    "Groq",
			RequiresKey:    true,
			DefaultBaseURL: "https://api.groq.com/openai/v1",
			SuggestedModels: []string{
				"llama-3.3-70b-versatile",
				"llama-3.1-8b-instant",
				"gemma2-9b-it",
			},
			Capabilities: fullCaps,
			APIKeyEnv:    "GROQ_API_KEY",
			BaseURLEnv:   "GROQ_BASE_URL",
		},
	},
	{
		Kind: adapterOpenAI,
		Info: ProviderInfo{
			ID:              ProviderOpenAICompatible,
			DisplayName:     "OpenAI-compatible",
			RequiresKey:     true,
			DefaultBaseURL:  "",
			SuggestedModels: []string{},
			Capabilities:    fullCaps,
			APIKeyEnv:       "OPENAI_COMPATIBLE_API_KEY",
			BaseURLEnv:      "OPENAI_COMPATIBLE_BASE_URL",
		},
	},
}

// ListProviders returns the supported providers in UI order.
// Suggested model lists are copies.
func ListProviders() []ProviderInfo {
	out := make([]ProviderInfo, len(providerCatalog))
	for i, spec := range providerCatalog {
		out[i] = cloneInfo(spec.Info)
	}
	return out
}

// LookupProvider returns metadata for id, accepting any case.
func LookupProvider(id ProviderID) (ProviderInfo, bool) {
	spec, ok := providerByID(id)
	if !ok {
		return ProviderInfo{}, false
	}
	return cloneInfo(spec.Info), true
}

func cloneInfo(info ProviderInfo) ProviderInfo {
	if info.SuggestedModels != nil {
		info.SuggestedModels = append([]string(nil), info.SuggestedModels...)
	}
	return info
}

func providerByID(id ProviderID) (providerSpec, bool) {
	want := normalizeProviderID(id)
	for _, spec := range providerCatalog {
		if spec.Info.ID == want {
			return spec, true
		}
	}
	return providerSpec{}, false
}

func normalizeProviderID(id ProviderID) ProviderID {
	return ProviderID(strings.ToLower(strings.TrimSpace(string(id))))
}
