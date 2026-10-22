package handlers

import (
	"context"
	"strings"
	"testing"
)

func TestHandleImageGeneratorChatReturnsImageWithoutToolLoop(t *testing.T) {
	orig := generateChatImage
	var gotPrompt string
	generateChatImage = func(ctx context.Context, prompt string) ([]byte, string, error) {
		gotPrompt = prompt
		return []byte{0x89, 0x50, 0x4e}, "image/png", nil
	}
	t.Cleanup(func() { generateChatImage = orig })

	h := &Handlers{}
	resp, err := h.handleImageGeneratorChat(nil, "a red robot")
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil || len(resp.Images) != 1 || resp.Images[0].Base64 == "" {
		t.Fatalf("expected one image, got %#v", resp)
	}
	if !strings.Contains(strings.ToLower(gotPrompt), "pixel") {
		t.Fatalf("image prompt should hint pixel art, got %q", gotPrompt)
	}
}

func TestHandleImageGeneratorChatRejectsEmpty(t *testing.T) {
	h := &Handlers{}
	if _, err := h.handleImageGeneratorChat(nil, "  "); err == nil {
		t.Fatal("empty prompt must fail")
	}
}
