package ai

import (
	"testing"

	"github.com/robo/dorphai"
)

func TestForChatConfigLeavesSharedClient(t *testing.T) {
	shared, err := New("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	llmBefore := shared.llm
	orchBefore := shared.orchestrator
	next := shared.ForChatConfig(dorphai.Config{
		Provider: dorphai.ProviderOllama,
		Model:    "llama3",
		BaseURL:  "http://127.0.0.1:11434/v1",
	})
	if shared.llm != llmBefore || shared.orchestrator != orchBefore {
		t.Fatal("shared service changed")
	}
	if next.llm == llmBefore || next.orchestrator == orchBefore || next.llm != next.orchestrator {
		t.Fatal("chat copy must use the saved client for both calls")
	}
}
