package ai

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/academi/backend/internal/config"
	docs "github.com/academi/backend/internal/docs"
	"github.com/academi/backend/internal/research"
)

const defaultLearnUserPrompt = `Give a thorough learning-focused analysis: key concepts, prerequisites, vocabulary, common pitfalls, how the material fits into broader STEM, and concrete study steps. If the material is ambiguous or incomplete, say what is missing.`

// LearnRequest powers POST /ai/learn (quick "Help you learn" from Docs).
type LearnRequest struct {
	DocID           string   `json:"doc_id"`
	DocIDs          []string `json:"doc_ids"`
	Message         string   `json:"message"`
	DisableResearch bool     `json:"disable_research"`
	AIProvider      string   `json:"ai_provider"`
}

func providerSupportsVision(p *config.AIProvider) bool {
	m := strings.ToLower(p.Model)
	if strings.Contains(m, "gpt-4o") || strings.Contains(m, "gpt-4-turbo") || strings.Contains(m, "vision") {
		return true
	}
	if strings.Contains(m, "claude-3") && (strings.Contains(m, "opus") || strings.Contains(m, "sonnet") || strings.Contains(m, "haiku")) {
		return true
	}
	return strings.Contains(m, "o1")
}

func (s *Service) docIDsIncludeImage(docIDs []string) bool {
	docSvc := docs.NewService(s.cfg.Server.UploadDir)
	for _, id := range docIDs {
		d, err := docSvc.GetByID(id)
		if err != nil {
			continue
		}
		if d.Type == "image" {
			return true
		}
	}
	return false
}

func (s *Service) postChatCompletionRaw(provider *config.AIProvider, messages []map[string]interface{}) (string, error) {
	messages = normalizeRawMessages(messages)
	body := map[string]interface{}{
		"model":       provider.Model,
		"messages":    messages,
		"max_tokens":  provider.MaxTokens,
		"temperature": provider.Temperature,
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	req, err := http.NewRequest("POST", provider.BaseURL+"/chat/completions", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("provider status %d: %s", resp.StatusCode, truncateStr(string(raw), 800))
	}
	var completion struct {
		Choices []struct {
			Message struct {
				Content interface{} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &completion); err != nil {
		return "", fmt.Errorf("parse: %w; body: %s", err, truncateStr(string(raw), 500))
	}
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no choices: %s", truncateStr(string(raw), 500))
	}
	c := completion.Choices[0].Message.Content
	switch v := c.(type) {
	case string:
		return v, nil
	default:
		b, _ := json.Marshal(v)
		return string(b), nil
	}
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// runHelpYouLearn runs multimodal-capable analysis with optional research and prior chat turns.
func (s *Service) runHelpYouLearn(docIDs []string, userMsg string, useResearch bool, providerName string, prior []ChatMessage) (string, []SourceReference, error) {
	provider, err := s.getProvider(providerName)
	if err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(userMsg) == "" {
		userMsg = defaultLearnUserPrompt
	}

	docSvc := docs.NewService(s.cfg.Server.UploadDir)
	messages := []map[string]interface{}{
		{"role": "system", "content": systemHelpYouLearn},
	}

	var sources []SourceReference

	if useResearch {
		rq := userMsg
		for _, id := range docIDs {
			d, err := docSvc.GetByID(id)
			if err != nil || d.Content == "" {
				continue
			}
			if d.Type == "image" {
				continue
			}
			snippet := d.Content
			if len(snippet) > 400 {
				snippet = snippet[:400]
			}
			rq += "\n" + snippet
		}
		if notes, srcs := research.Gather(strings.TrimSpace(truncateStr(rq, 400))); notes != "" {
			messages = append(messages, map[string]interface{}{
				"role":    "system",
				"content": "Research notes (verify — may be incomplete):\n\n" + notes,
			})
			for _, rs := range srcs {
				sources = append(sources, SourceReference{Title: rs.Title, Type: rs.Type, URL: rs.URL})
			}
		}
	}

	var imageDataURLs []string
	for _, id := range docIDs {
		d, err := docSvc.GetByID(id)
		if err != nil {
			continue
		}
		if d.Type == "image" && d.StoredFilename != "" {
			data, err := docSvc.ReadStored(d.StoredFilename)
			if err != nil {
				continue
			}
			mt := d.MimeType
			if mt == "" {
				mt = "image/jpeg"
			}
			b64 := base64.StdEncoding.EncodeToString(data)
			imageDataURLs = append(imageDataURLs, fmt.Sprintf("data:%s;base64,%s", mt, b64))
			continue
		}
		if d.Content != "" {
			chunk := d.Content
			if len(chunk) > 120000 {
				chunk = chunk[:120000] + "\n… [truncated]"
			}
			messages = append(messages, map[string]interface{}{
				"role":    "system",
				"content": fmt.Sprintf("User-provided document «%s» (type %s):\n%s", d.Title, d.Type, chunk),
			})
		} else if d.Type == "pdf" {
			messages = append(messages, map[string]interface{}{
				"role": "system",
				"content": fmt.Sprintf(
					"User attached PDF «%s» but little or no text was extracted — infer what you can from context and the user’s questions; be explicit about uncertainty.",
					d.Title),
			})
		}
	}

	for _, m := range prior {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		messages = append(messages, map[string]interface{}{
			"role":    role,
			"content": m.Content,
		})
	}

	var userContent interface{}
	if len(imageDataURLs) > 0 {
		if !providerSupportsVision(provider) {
			return "", nil, fmt.Errorf("this model may not support images; use a vision-capable model (e.g. gpt-4o)")
		}
		parts := []interface{}{
			map[string]string{"type": "text", "text": userMsg},
		}
		for _, u := range imageDataURLs {
			parts = append(parts, map[string]interface{}{
				"type": "image_url",
				"image_url": map[string]string{
					"url": u,
				},
			})
		}
		userContent = parts
	} else {
		userContent = userMsg
	}
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userContent,
	})

	reply, err := s.postChatCompletionRaw(provider, messages)
	if err != nil {
		return "", nil, err
	}
	if len(sources) == 0 {
		sources = []SourceReference{{Title: "Academi", Type: "internal"}}
	}
	return reply, sources, nil
}

// LearnHandler POST /ai/learn
func (s *Service) LearnHandler(c *gin.Context) {
	var req LearnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ids := append([]string{}, req.DocIDs...)
	if strings.TrimSpace(req.DocID) != "" {
		ids = append(ids, strings.TrimSpace(req.DocID))
	}
	seen := map[string]bool{}
	var uniq []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		uniq = append(uniq, id)
	}
	if len(uniq) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doc_id or doc_ids required"})
		return
	}

	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		msg = defaultLearnUserPrompt
	}

	reply, sources, err := s.runHelpYouLearn(uniq, msg, !req.DisableResearch, req.AIProvider, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"response":            reply,
		"sources":             sources,
		"offer_save_analysis": true,
		"source_document_ids": uniq,
	})
}
