package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

const morphAIImageGeneratorAgentID = "image-generator"

const pixelArtImagePrefix = "pixel art, limited palette, chunky pixels: "

// generateChatImage is the image-generator backend. Tests replace this; production uses generateStoryImageBytes.
var generateChatImage = generateStoryImageBytes

func truncateChatImageAlt(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if s == "" || maxRunes <= 0 {
		return "Generated image"
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes]) + "…"
}

func pixelArtImagePrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if strings.Contains(strings.ToLower(prompt), "pixel") {
		return prompt
	}
	return pixelArtImagePrefix + prompt
}

func imageGenContext(c *gin.Context) context.Context {
	if c != nil && c.Request != nil {
		return c.Request.Context()
	}
	return context.Background()
}

// handleImageGeneratorChat bypasses the tool loop and generates an image from the user prompt.
func (h *Handlers) handleImageGeneratorChat(c *gin.Context, prompt string) (*models.ChatResponse, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, fmt.Errorf("describe the image you want")
	}
	bytes, ctype, err := generateChatImage(imageGenContext(c), pixelArtImagePrompt(prompt))
	if err != nil {
		return nil, err
	}
	if ctype == "" {
		ctype = "image/png"
	}
	return &models.ChatResponse{
		Response: "Here’s an image based on your prompt.",
		Images: []models.ChatImage{{
			ContentType: ctype,
			Base64:      base64.StdEncoding.EncodeToString(bytes),
			Alt:         truncateChatImageAlt(prompt, 120),
		}},
	}, nil
}
