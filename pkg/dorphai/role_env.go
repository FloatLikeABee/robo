package dorphai

import (
	"os"
	"strings"
)

const defaultNVIDIAModel = "nvidia/nemotron-3-ultra-550b-a55b"

// OrchestratorProviderFromEnv returns DORPH_AI_ORCHESTRATOR_PROVIDER when set.
func OrchestratorProviderFromEnv() ProviderID {
	return normalizeProviderID(ProviderID(env("DORPH_AI_ORCHESTRATOR_PROVIDER")))
}

// CoordinatorProviderFromEnv returns DORPH_AI_COORDINATOR_PROVIDER when set.
func CoordinatorProviderFromEnv() ProviderID {
	return normalizeProviderID(ProviderID(env("DORPH_AI_COORDINATOR_PROVIDER")))
}

// LoadOrchestratorFromEnv loads the primary agent model. Empty orchestrator provider keeps LoadFromEnv().
func LoadOrchestratorFromEnv() Config {
	if OrchestratorProviderFromEnv() == "" {
		return LoadFromEnv()
	}
	return loadRoleProviderConfig(OrchestratorProviderFromEnv())
}

// LoadCoordinatorFromEnv loads the coordinator model. Empty coordinator provider falls back to orchestrator.
func LoadCoordinatorFromEnv() Config {
	if CoordinatorProviderFromEnv() == "" {
		return LoadOrchestratorFromEnv()
	}
	return loadRoleProviderConfig(CoordinatorProviderFromEnv())
}

func loadRoleProviderConfig(provider ProviderID) Config {
	spec, ok := providerByID(provider)
	if !ok {
		return Config{Provider: provider}
	}
	cfg := Config{Provider: spec.Info.ID}
	switch provider {
	case ProviderSiliconFlow:
		if model := strings.TrimSpace(os.Getenv("SILICONFLOW_MODEL")); model != "" {
			cfg.Model = model
		}
	case ProviderNVIDIA:
		cfg.Model = firstNonEmpty(env("NVIDIA_MODEL"), defaultNVIDIAModel)
	default:
		if model := env("DORPH_AI_MODEL"); model != "" && model != DefaultModel {
			cfg.Model = model
		}
	}
	if cfg.Model == "" && len(spec.Info.SuggestedModels) > 0 {
		cfg.Model = spec.Info.SuggestedModels[0]
	}
	return cfg
}

// NVIDIAEnableThinkingFromEnv is true unless NVIDIA_ENABLE_THINKING is false/0.
func NVIDIAEnableThinkingFromEnv() bool {
	v := strings.ToLower(strings.TrimSpace(env("NVIDIA_ENABLE_THINKING")))
	return v != "false" && v != "0" && v != "no"
}
