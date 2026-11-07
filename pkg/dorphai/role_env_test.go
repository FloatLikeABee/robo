package dorphai

import "testing"

func TestLoadOrchestratorFallbackLegacy(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("DORPH_AI_API_KEY", "legacy-key")
	t.Setenv("DORPH_AI_MODEL", "qwen3-max")
	cfg := LoadOrchestratorFromEnv()
	if cfg.Provider != "" {
		t.Fatalf("provider = %q, want empty legacy", cfg.Provider)
	}
	if cfg.APIKey != "legacy-key" {
		t.Fatalf("key = %q", cfg.APIKey)
	}
}

func TestLoadOrchestratorNVIDIA(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("DORPH_AI_ORCHESTRATOR_PROVIDER", "nvidia")
	t.Setenv("NVIDIA_API_KEY", "nv-key")
	cfg := LoadOrchestratorFromEnv()
	if cfg.Provider != ProviderNVIDIA {
		t.Fatalf("provider = %q", cfg.Provider)
	}
	if cfg.Model != defaultNVIDIAModel {
		t.Fatalf("model = %q", cfg.Model)
	}
	rc, err := resolve(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !rc.NVIDIAEnableThinking {
		t.Fatal("expected NVIDIA thinking on by default")
	}
}

func TestLoadCoordinatorSiliconFlow(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("DORPH_AI_COORDINATOR_PROVIDER", "siliconflow")
	t.Setenv("SILICONFLOW_API_KEY", "sf-key")
	t.Setenv("SILICONFLOW_MODEL", "Qwen/Qwen3-8B")
	cfg := LoadCoordinatorFromEnv()
	if cfg.Provider != ProviderSiliconFlow {
		t.Fatalf("provider = %q", cfg.Provider)
	}
	if cfg.Model != "Qwen/Qwen3-8B" {
		t.Fatalf("model = %q", cfg.Model)
	}
}
