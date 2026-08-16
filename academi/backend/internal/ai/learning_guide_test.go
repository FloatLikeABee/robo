package ai

import (
	"strings"
	"testing"
)

func TestWantsLearningGuide(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		req  ChatRequest
		want bool
	}{
		{
			name: "help you learn toggle",
			req: ChatRequest{
				HelpYouLearn: true,
				Messages:     []ChatMessage{{Role: "user", Content: "hello"}},
			},
			want: true,
		},
		{
			name: "intent in message",
			req: ChatRequest{
				Messages: []ChatMessage{{Role: "user", Content: "How do I learn organic chemistry?"}},
			},
			want: true,
		},
		{
			name: "blocked by document mode",
			req: ChatRequest{
				DocumentMode: true,
				HelpYouLearn: true,
				Messages:     []ChatMessage{{Role: "user", Content: "learn math"}},
			},
			want: false,
		},
		{
			name: "blocked by attached docs",
			req: ChatRequest{
				HelpYouLearn: true,
				DocIDs:       []string{"doc-1"},
				Messages:     []ChatMessage{{Role: "user", Content: "learn this"}},
			},
			want: false,
		},
		{
			name: "generic question",
			req: ChatRequest{
				Messages: []ChatMessage{{Role: "user", Content: "What is the capital of France?"}},
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := WantsLearningGuide(tc.req); got != tc.want {
				t.Fatalf("WantsLearningGuide() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLearningGuidePromptLoaded(t *testing.T) {
	t.Parallel()
	p := LearningGuidePrompt()
	if p == "" {
		t.Fatal("expected non-empty learning guide prompt")
	}
	if !strings.Contains(p, "Markdown") {
		t.Fatal("expected markdown guidance in prompt")
	}
}
