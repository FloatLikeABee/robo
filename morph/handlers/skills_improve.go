package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"idongivaflyinfa/ai"

	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

type skillImproveBody struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Instructions string `json:"instructions"`
}

func parseSkillImproveJSON(raw string) (name, description, instructions string, err error) {
	obj, ok := morphai.ExtractJSONObject(raw)
	if !ok {
		return "", "", "", fmt.Errorf("AI did not return JSON")
	}
	var parsed struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Instructions string `json:"instructions"`
	}
	if err := json.Unmarshal([]byte(obj), &parsed); err != nil {
		return "", "", "", fmt.Errorf("AI JSON was invalid")
	}
	name = strings.TrimSpace(parsed.Name)
	description = strings.TrimSpace(parsed.Description)
	instructions = strings.TrimSpace(parsed.Instructions)
	if name == "" || instructions == "" {
		return "", "", "", fmt.Errorf("AI draft is missing name or instructions")
	}
	return name, description, instructions, nil
}

func skillImprovePrompt(name, description, instructions string) string {
	return strings.TrimSpace(fmt.Sprintf(`Revise this Morph AI skill draft. Keep the operator's intent. Return ONLY a JSON object with keys name, description, and instructions (all strings). Do not save anything.

Current name:
%s

Current description:
%s

Current instructions:
%s
`, name, description, instructions))
}

// ImproveSkill POST /api/skills/improve — drafts name/description/instructions without persisting.
func (h *Handlers) ImproveSkill(c *gin.Context) {
	var body skillImproveBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	name := strings.TrimSpace(body.Name)
	instructions := strings.TrimSpace(body.Instructions)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if instructions == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instructions is required"})
		return
	}
	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}
	raw, err := h.aiService.ChatCompletion(context.Background(), []ai.DashScopeMessage{
		{Role: "user", Content: skillImprovePrompt(name, strings.TrimSpace(body.Description), instructions)},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	outName, outDesc, outInstr, err := parseSkillImproveJSON(raw)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"name":         outName,
		"description":  outDesc,
		"instructions": outInstr,
	})
}
