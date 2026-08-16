package handlers

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

const morphAIImageGeneratorAgentID = "image-generator"

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

// handleImageGeneratorChat bypasses the tool loop and generates an image from the user prompt.
func (h *Handlers) handleImageGeneratorChat(c *gin.Context, prompt string) (*models.ChatResponse, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, fmt.Errorf("describe the image you want")
	}
	bytes, ctype, err := generateStoryImageBytes(c.Request.Context(), prompt)
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
