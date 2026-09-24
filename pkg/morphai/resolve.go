package morphai

import (
	"fmt"
	"os"
	"strings"
)

// ResolvedConfig is the provider, model, and endpoint a call would use.
// The API key itself is not included.
type ResolvedConfig struct {
	Provider     ProviderID
	Model        string
	VisionModel  string
	BaseURL      string
	APIURL       string
	UseNativeAPI bool
	HasAPIKey    bool
	// Adapter is "openai", "anthropic", or "dashscope-native".
	Adapter string
}

type resolved struct {
	Provider          ProviderID
	Model             string
	VisionModel       string
	BaseURL           string
	APIURL            string
	APIKey            string
	Native            bool
	Legacy            bool
	Kind              adapterKind
	Caps              Capabilities
	EnableThinkingOff bool
}

func (rc resolved) public() ResolvedConfig {
	return ResolvedConfig{
		Provider:     rc.Provider,
		Model:        rc.Model,
		VisionModel:  rc.VisionModel,
		BaseURL:      rc.BaseURL,
		APIURL:       rc.APIURL,
		UseNativeAPI: rc.Native,
		HasAPIKey:    strings.TrimSpace(rc.APIKey) != "",
		Adapter:      rc.Kind.String(),
	}
}

func (rc resolved) supports(capability string) bool {
	switch capability {
	case capChat:
		return rc.Caps.Chat
	case capStream:
		return rc.Caps.Stream
	case capTools:
		return rc.Caps.Tools
	case capVision:
		return rc.Caps.Vision
	case capJSON:
		return rc.Caps.JSONMode
	default:
		return false
	}
}

func (rc resolved) reject(capability string) error {
	if rc.supports(capability) {
		return nil
	}
	if rc.Legacy && rc.Native && capability == capVision {
		return &CapabilityError{
			Provider:   rc.Provider,
			Capability: capability,
			Detail:     nativeVisionDetail,
		}
	}
	return &CapabilityError{Provider: rc.Provider, Capability: capability}
}

// ResolveConfig applies the config resolution rules and returns the settings
// a call would use. It does not contact the provider.
func ResolveConfig(cfg Config) (ResolvedConfig, error) {
	rc, err := resolve(cfg)
	if err != nil {
		return ResolvedConfig{}, err
	}
	return rc.public(), nil
}

func resolve(cfg Config) (resolved, error) {
	cfg.Provider = normalizeProviderID(cfg.Provider)
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Model = strings.TrimSpace(cfg.Model)
	cfg.VisionModel = strings.TrimSpace(cfg.VisionModel)
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.APIURL = strings.TrimRight(strings.TrimSpace(cfg.APIURL), "/")

	if cfg.Provider == "" {
		return resolveLegacy(cfg)
	}
	spec, ok := providerByID(cfg.Provider)
	if !ok {
		return resolved{}, &UnknownProviderError{ID: cfg.Provider}
	}
	return resolveExplicit(cfg, spec)
}

func resolveLegacy(cfg Config) (resolved, error) {
	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}
	base := resolveLegacyBase(cfg)
	api := cfg.APIURL
	if api == "" {
		api = DefaultAPIURL
	}
	vision := cfg.VisionModel
	if vision == "" {
		vision = DefaultVisionModel
	}
	kind := adapterOpenAI
	caps := fullCaps
	provider := ProviderDashScope
	if cfg.UseNativeAPI {
		kind = adapterDashScopeNative
		caps = Capabilities{Chat: true}
	} else if !isDashScopeBase(base) {
		provider = ProviderOpenAICompatible
	}
	key := resolveLegacyKey(cfg)
	rc := resolved{
		Provider:          provider,
		Model:             model,
		VisionModel:       vision,
		BaseURL:           base,
		APIURL:            api,
		APIKey:            key,
		Native:            cfg.UseNativeAPI,
		Legacy:            true,
		Kind:              kind,
		Caps:              caps,
		EnableThinkingOff: !cfg.UseNativeAPI,
	}
	if key == "" {
		return rc, legacyKeyError{}
	}
	return rc, nil
}

func resolveExplicit(cfg Config, spec providerSpec) (resolved, error) {
	key := namedProviderKey(cfg)
	if key == "" {
		key = spec.lookupEnv(spec.Info.APIKeyEnv, spec.ExtraKeyEnvs)
	}

	base := namedProviderBase(cfg)
	if base == "" {
		base = spec.lookupEnv(spec.Info.BaseURLEnv, spec.ExtraBaseEnvs)
	}
	if base == "" {
		base = strings.TrimRight(strings.TrimSpace(spec.Info.DefaultBaseURL), "/")
	}
	if spec.Info.ID == ProviderOllama {
		base = normalizeOllamaBase(base)
	}

	model := cfg.Model
	if cfg.legacyDefaultModel && model == DefaultModel && spec.Info.ID != ProviderDashScope && len(spec.Info.SuggestedModels) > 0 {
		model = ""
	}
	if model == "" && spec.Info.ID == ProviderDashScope {
		model = firstNonEmpty(os.Getenv("MORPH_AI_MODEL"), os.Getenv("TRAN_QWEN_MODEL"))
		if model == "" {
			model = DefaultModel
		}
	}
	if model == "" && len(spec.Info.SuggestedModels) > 0 {
		model = spec.Info.SuggestedModels[0]
	}

	vision := cfg.VisionModel
	if vision == "" {
		vision = spec.DefaultVision
	}
	if vision == "" {
		vision = model
	}

	native := spec.Info.ID == ProviderDashScope && cfg.UseNativeAPI
	kind := spec.Kind
	caps := spec.Info.Capabilities
	apiURL := cfg.APIURL
	if native {
		kind = adapterDashScopeNative
		caps = Capabilities{Chat: true}
		if apiURL == "" {
			apiURL = DefaultAPIURL
		}
	}

	thinkingOff := false
	if kind == adapterOpenAI && (spec.Info.ID == ProviderDashScope || strings.Contains(strings.ToLower(base), "dashscope.aliyuncs.com")) {
		thinkingOff = true
	}

	rc := resolved{
		Provider:          spec.Info.ID,
		Model:             model,
		VisionModel:       vision,
		BaseURL:           base,
		APIURL:            apiURL,
		APIKey:            key,
		Native:            native,
		Legacy:            false,
		Kind:              kind,
		Caps:              caps,
		EnableThinkingOff: thinkingOff,
	}

	if spec.Info.RequiresKey && key == "" {
		envName := spec.Info.APIKeyEnv
		if envName == "" {
			envName = "an API key"
		}
		return rc, &ProviderConfigError{
			Provider: spec.Info.ID,
			Detail:   "set " + envName + " or pass an API key",
		}
	}
	if base == "" {
		envName := spec.Info.BaseURLEnv
		if envName == "" {
			envName = "a base URL"
		}
		return rc, &ProviderConfigError{
			Provider: spec.Info.ID,
			Detail:   "set " + envName + " or pass a base URL",
		}
	}
	if model == "" {
		return rc, fmt.Errorf("provider %s: model is required", spec.Info.ID)
	}
	return rc, nil
}

// namedProviderKey returns a key the caller set on Config. The snapshot
// LoadFromEnv stored is not one of those, even if the env var later changes.
func namedProviderKey(cfg Config) string {
	key := strings.TrimSpace(cfg.APIKey)
	loaded := strings.TrimSpace(cfg.legacyKey)
	if loaded != "" && key == loaded {
		return ""
	}
	return key
}

func resolveLegacyKey(cfg Config) string {
	key := strings.TrimSpace(cfg.APIKey)
	loaded := strings.TrimSpace(cfg.legacyKey)
	if key != loaded {
		return key
	}
	return loaded
}

// namedProviderBase returns a base URL the caller set. A URL LoadFromEnv copied
// from MORPH_AI_BASE_URL, TRAN_QWEN_BASE_URL, or a compatible MORPH_AI_API_URL
// belongs to the empty-provider path only.
func namedProviderBase(cfg Config) string {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	loaded := strings.TrimRight(strings.TrimSpace(cfg.legacyBase), "/")
	if loaded != "" && base == loaded {
		return ""
	}
	return base
}

func resolveLegacyBase(cfg Config) string {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	loaded := strings.TrimRight(strings.TrimSpace(cfg.legacyBase), "/")
	if base != loaded {
		if base != "" {
			return base
		}
		return DefaultBaseURL
	}
	if loaded != "" {
		return loaded
	}
	return DefaultBaseURL
}

func (s providerSpec) lookupEnv(primary string, extra []string) string {
	vals := make([]string, 0, 1+len(extra))
	if primary != "" {
		vals = append(vals, os.Getenv(primary))
	}
	for _, name := range extra {
		if name != "" {
			vals = append(vals, os.Getenv(name))
		}
	}
	return firstNonEmpty(vals...)
}

func isDashScopeBase(u string) bool {
	u = strings.ToLower(strings.TrimSpace(u))
	if u == "" || u == strings.ToLower(DefaultBaseURL) {
		return true
	}
	return strings.Contains(u, "dashscope.aliyuncs.com")
}

func normalizeOllamaBase(u string) string {
	u = strings.TrimRight(strings.TrimSpace(u), "/")
	if u == "" {
		return "http://127.0.0.1:11434/v1"
	}
	if strings.HasSuffix(strings.ToLower(u), "/v1") {
		return u
	}
	return u + "/v1"
}

// applyCall merges a per-call override onto the client config.
// Switching to a different named provider drops the client's key and base URL
// so a DashScope key is never sent to OpenAI (or any other provider).
// Selecting DashScope on a legacy client keeps the client's key and URLs,
// because that client is already DashScope.
func (c Config) applyCall(req CompletionRequest) Config {
	out := c
	reqProvider := normalizeProviderID(req.Provider)
	if reqProvider != "" && reqProvider != normalizeProviderID(c.Provider) {
		if c.Provider == "" && reqProvider == ProviderDashScope {
			out.Provider = ProviderDashScope
		} else {
			out = Config{Provider: reqProvider}
		}
	}
	if reqProvider != "" {
		out.Provider = reqProvider
	}
	if s := strings.TrimSpace(req.APIKey); s != "" {
		out.APIKey = s
		out.legacyKey = ""
	}
	if s := strings.TrimSpace(req.BaseURL); s != "" {
		out.BaseURL = strings.TrimRight(s, "/")
		out.legacyBase = ""
	}
	if s := strings.TrimSpace(req.Model); s != "" {
		out.Model = s
	}
	return out
}
