package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/academi/backend/internal/config"
	"github.com/academi/backend/internal/database"
	docs "github.com/academi/backend/internal/docs"
	"github.com/academi/backend/internal/research"
)

type Service struct {
	cfg *config.Config
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
}

type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ChatCompletionResponse struct {
	Choices []ChatCompletionChoice `json:"choices"`
}

type SourceReference struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	URL   string `json:"url,omitempty"`
}

type ChatRequest struct {
	Message          string            `json:"message"`
	Context          map[string]string `json:"context"`
	AIProvider       string            `json:"ai_provider"`
	Messages         []ChatMessage     `json:"messages"`
	DocumentMode     bool              `json:"document_mode"`
	DisableResearch  bool              `json:"disable_research"`
	DocIDs           []string          `json:"doc_ids"`
	HelpYouLearn     bool              `json:"help_you_learn"`
}

type SummarizeRequest struct {
	Content string `json:"content" binding:"required"`
}

type GenerateGuideRequest struct {
	Topic string `json:"topic" binding:"required"`
	Steps int    `json:"steps"`
}

type ModerateRequest struct {
	Content string `json:"content" binding:"required"`
}

func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) getProvider(providerName string) (*config.AIProvider, error) {
	if providerName == "" {
		providerName = s.cfg.AI.DefaultProvider
	}

	provider, exists := s.cfg.AI.Providers[providerName]
	if !exists {
		return nil, fmt.Errorf("AI provider '%s' not found", providerName)
	}

	if provider.APIKey == "" {
		return nil, fmt.Errorf("API key not configured for provider '%s'", providerName)
	}

	return &provider, nil
}

func (s *Service) Chat(userMessage string, context map[string]string) (string, error) {
	return s.ChatWithProvider(userMessage, context, "")
}

func (s *Service) ChatWithProvider(userMessage string, context map[string]string, providerName string) (string, error) {
	provider, err := s.getProvider(providerName)
	if err != nil {
		return "", err
	}

	messages := []ChatMessage{
		{Role: "system", Content: systemDefault},
	}

	for key, value := range context {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("Context - %s: %s", key, value),
		})
	}

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	return s.completeChat(provider, messages)
}

func (s *Service) completeChat(provider *config.AIProvider, messages []ChatMessage) (string, error) {
	messages = normalizeChatMessages(messages)
	reqBody := ChatCompletionRequest{
		Model:       provider.Model,
		Messages:    messages,
		MaxTokens:   provider.MaxTokens,
		Temperature: provider.Temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", provider.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("AI Response Status: %d, Body: %s", resp.StatusCode, truncateForLog(string(body), 400))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI provider returned %d: %s", resp.StatusCode, truncateForLog(string(body), 300))
	}

	var completion ChatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return "", fmt.Errorf("failed to parse response: %w (body: %s)", err, truncateForLog(string(body), 200))
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no choices in response, body: %s", string(body))
	}

	return completion.Choices[0].Message.Content, nil
}

func lastUserSnippet(req ChatRequest) string {
	if strings.TrimSpace(req.Message) != "" {
		return strings.TrimSpace(req.Message)
	}
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(req.Messages[i].Role), "user") && strings.TrimSpace(req.Messages[i].Content) != "" {
			return strings.TrimSpace(req.Messages[i].Content)
		}
	}
	return ""
}

func firstUserSnippet(req ChatRequest) string {
	if len(req.Messages) == 0 {
		return ""
	}
	for _, m := range req.Messages {
		if strings.EqualFold(strings.TrimSpace(m.Role), "user") && strings.TrimSpace(m.Content) != "" {
			return strings.TrimSpace(m.Content)
		}
	}
	return ""
}

func researchQueryFromThread(req ChatRequest) string {
	last := lastUserSnippet(req)
	first := firstUserSnippet(req)
	if first != "" && last != "" && first != last && len(first)+len(last) < 500 {
		return first + " | " + last
	}
	return last
}

func clipConversation(msgs []ChatMessage, maxMessages int) []ChatMessage {
	if maxMessages <= 0 || len(msgs) <= maxMessages {
		out := make([]ChatMessage, len(msgs))
		copy(out, msgs)
		return out
	}
	return append([]ChatMessage{}, msgs[len(msgs)-maxMessages:]...)
}

func (s *Service) buildChatPayload(req ChatRequest) ([]ChatMessage, []SourceReference, error) {
	learningGuide := WantsLearningGuide(req)
	sys := systemDefault
	switch {
	case req.DocumentMode:
		sys = systemDocumentAgent
	case len(req.DocIDs) > 0 && !req.HelpYouLearn:
		sys = systemDocumentAnalysis
	case learningGuide:
		sys = systemDefault + "\n\n" + LearningGuidePrompt()
	}

	out := []ChatMessage{{Role: "system", Content: sys}}

	for key, value := range req.Context {
		if strings.TrimSpace(value) == "" {
			continue
		}
		out = append(out, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("Context — %s: %s", key, value),
		})
	}

	docSvc := docs.NewService(s.cfg.Server.UploadDir)
	for _, id := range req.DocIDs {
		d, err := docSvc.GetByID(strings.TrimSpace(id))
		if err != nil {
			continue
		}
		if d.Type == "image" {
			continue
		}
		if d.Content == "" {
			continue
		}
		chunk := d.Content
		if len(chunk) > 120000 {
			chunk = chunk[:120000] + "\n… [truncated]"
		}
		out = append(out, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("Attached document «%s» (type %s):\n%s", d.Title, d.Type, chunk),
		})
	}

	var researchSources []SourceReference
	shouldResearch := !req.DisableResearch && (req.DocumentMode || len(req.DocIDs) > 0 || learningGuide)
	if shouldResearch {
		q := researchQueryFromThread(req)
		if learningGuide {
			q = LearningGuideResearchQuery(lastUserSnippet(req))
		}
		if q != "" {
			if notes, srcs := research.Gather(q); notes != "" {
				label := "Research notes (public sources; verify — may be incomplete):\n\n"
				if learningGuide {
					label = "Web research for learning approaches (synthesize into your Markdown guide; verify — may be incomplete):\n\n"
				}
				out = append(out, ChatMessage{
					Role:    "system",
					Content: label + notes,
				})
				for _, rs := range srcs {
					researchSources = append(researchSources, SourceReference{
						Title: rs.Title,
						Type:  rs.Type,
						URL:   rs.URL,
					})
				}
			}
		}
	}

	if len(req.Messages) > 0 {
		convo := clipConversation(req.Messages, 48)
		for _, m := range convo {
			role := strings.ToLower(strings.TrimSpace(m.Role))
			if role != "user" && role != "assistant" {
				continue
			}
			out = append(out, ChatMessage{Role: role, Content: m.Content})
		}
	}

	if msg := strings.TrimSpace(req.Message); msg != "" {
		if len(out) == 0 || out[len(out)-1].Role != "user" || out[len(out)-1].Content != msg {
			out = append(out, ChatMessage{Role: "user", Content: msg})
		}
	}

	hasUser := false
	for _, m := range out {
		if m.Role == "user" {
			hasUser = true
			break
		}
	}
	if !hasUser {
		return nil, nil, fmt.Errorf("no user messages to process")
	}

	return out, researchSources, nil
}

func (s *Service) ChatWithMessages(req ChatRequest) (string, []SourceReference, error) {
	hasDocs := len(req.DocIDs) > 0 || s.docIDsIncludeImage(req.DocIDs)
	if hasDocs && (req.HelpYouLearn || s.docIDsIncludeImage(req.DocIDs)) {
		user := strings.TrimSpace(req.Message)
		if user == "" {
			user = lastUserSnippet(req)
		}
		if user == "" && s.docIDsIncludeImage(req.DocIDs) {
			user = "Describe and analyze the attached image(s) in an academic context."
		}
		prior := clipConversation(req.Messages, 40)
		return s.runHelpYouLearn(req.DocIDs, user, !req.DisableResearch, req.AIProvider, prior)
	}

	payload, researchSources, err := s.buildChatPayload(req)
	if err != nil {
		return "", nil, err
	}

	provider, err := s.getProvider(req.AIProvider)
	if err != nil {
		return "", nil, err
	}

	reply, err := s.completeChat(provider, payload)
	if err != nil {
		return "", nil, err
	}

	sources := researchSources
	if len(sources) == 0 {
		sources = []SourceReference{{Title: "Academi", Type: "internal"}}
	}
	return reply, sources, nil
}

func (s *Service) ChatHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Context == nil {
		req.Context = make(map[string]string)
	}

	if strings.TrimSpace(req.Message) == "" && len(req.Messages) == 0 && len(req.DocIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message, messages, or doc_ids required"})
		return
	}

	response, sources, err := s.ChatWithMessages(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	display := response
	out := gin.H{
		"response":      display,
		"sources":       sources,
		"document_mode": req.DocumentMode,
	}
	if req.HelpYouLearn || WantsLearningGuide(req) {
		out["offer_save_analysis"] = true
	}

	if len(req.DocIDs) > 0 && !req.HelpYouLearn {
		out["offer_save_analysis"] = true
	}

	if req.DocumentMode && !req.HelpYouLearn {
		if d, title, body, ok := parseAcademiDoc(response); ok {
			display = d
			out["response"] = display
			out["offer_save_analysis"] = true
			out["pending_save"] = gin.H{
				"title":   title,
				"content": body,
				"tags":    []string{"#ai-document", "#rag"},
			}
		}
	}

	c.JSON(http.StatusOK, out)
}

func (s *Service) SummarizeHandler(c *gin.Context) {
	var req SummarizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	summary, err := s.Summarize(req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func (s *Service) GenerateGuideHandler(c *gin.Context) {
	var req GenerateGuideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Steps <= 0 {
		req.Steps = 7
	}

	guide, err := s.GenerateGuide(req.Topic, req.Steps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"guide": guide})
}

func (s *Service) ModerateHandler(c *gin.Context) {
	var req ModerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approved, reason, err := s.ModerateContent(req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"approved": approved,
		"reason":   reason,
	})
}

func (s *Service) Summarize(content string) (string, error) {
	return s.Chat("Summarize the following in 2-3 sentences: "+content, nil)
}

func (s *Service) GenerateGuide(topic string, steps int) (string, error) {
	prompt := fmt.Sprintf("Create a step-by-step learning guide for '%s' with %d steps. Return as a numbered list with clear, actionable steps. Each step should be 1-2 sentences.", topic, steps)
	return s.Chat(prompt, nil)
}

func (s *Service) ModerateContent(content string) (bool, string, error) {
	prompt := fmt.Sprintf("Evaluate the following content for academic appropriateness. Respond with 'APPROVED' if acceptable, or 'FLAGGED: <reason>' if not.\n\nContent: %s", content)

	result, err := s.Chat(prompt, nil)
	if err != nil {
		return true, "", err
	}

	isApproved := strings.HasPrefix(strings.ToUpper(result), "APPROVED")
	return isApproved, result, nil
}

func (s *Service) SearchDocs(query string, tone string, depth string) (string, error) {
	docContext, _ := database.Get([]byte("docs_index"))
	context := map[string]string{
		"tone":  tone,
		"depth": depth,
	}
	if docContext != nil {
		context["available_docs"] = string(docContext)
	}

	return s.Chat("Find relevant information for this query: "+query, context)
}

func (s *Service) ProvidersHandler(c *gin.Context) {
	providers := make([]gin.H, 0)
	for key, provider := range s.cfg.AI.Providers {
		providers = append(providers, gin.H{
			"id":          key,
			"name":        provider.Name,
			"model":       provider.Model,
			"has_api_key": provider.APIKey != "",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"default_provider": s.cfg.AI.DefaultProvider,
		"providers":        providers,
	})
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	ai := r.Group("/ai")
	{
		ai.GET("/providers", s.ProvidersHandler)
		ai.POST("/chat", s.ChatHandler)
		ai.POST("/learn", s.LearnHandler)
		ai.POST("/summarize", s.SummarizeHandler)
		ai.POST("/generate-guide", s.GenerateGuideHandler)
		ai.POST("/moderate", s.ModerateHandler)
	}
}

func truncateForLog(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
