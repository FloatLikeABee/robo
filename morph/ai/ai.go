package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/robo/morphai"
	"idongivaflyinfa/cache"
	"idongivaflyinfa/models"
)

type AIService struct {
	apiKey    string
	modelName string
	cache     *cache.Cache
	llm       *morphai.Client
}

type DashScopeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func New(apiKey string, modelName string, appCache *cache.Cache) (*AIService, error) {
	cfg := morphai.LoadFromEnv()
	if strings.TrimSpace(apiKey) != "" {
		cfg.APIKey = strings.TrimSpace(apiKey)
	}
	if strings.TrimSpace(modelName) != "" {
		cfg.Model = strings.TrimSpace(modelName)
	}
	return &AIService{
		apiKey:    cfg.APIKey,
		modelName: cfg.Model,
		cache:     appCache,
		llm:       morphai.NewClient(cfg),
	}, nil
}

func (a *AIService) Close() error {
	// HTTP client doesn't require explicit closing
	return nil
}

func (a *AIService) callDashScopeAPI(ctx context.Context, messages []DashScopeMessage) (string, error) {
	return a.llm.ChatCompletion(ctx, toMorphMessages(messages))
}

func toMorphMessages(messages []DashScopeMessage) []morphai.Message {
	out := make([]morphai.Message, len(messages))
	for i, m := range messages {
		out[i] = morphai.Message{Role: m.Role, Content: m.Content}
	}
	return out
}

func (a *AIService) callDashScopeAPIWithClient(ctx context.Context, messages []DashScopeMessage, longTimeout bool) (string, error) {
	msgs := toMorphMessages(messages)
	if longTimeout {
		return a.llm.ChatCompletionLong(ctx, msgs)
	}
	return a.llm.ChatCompletion(ctx, msgs)
}

func (a *AIService) GenerateSQL(userPrompt string, sqlFiles []models.SQLFile) (string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("prompt:%s", userPrompt)
	if cached, found := a.cache.Get(cacheKey); found {
		return cached.(string), nil
	}

	ctx := context.Background()

	// Build prompt using helper
	prompt := BuildSQLPrompt(userPrompt, sqlFiles)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	fmt.Println("prompt:", prompt)

	response, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		fmt.Println("error:", err)
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	sql := strings.TrimSpace(response)
	// Remove markdown code blocks if present
	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```SQL")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	sql = strings.TrimSpace(sql)

	// Cache the result
	a.cache.SetDefault(cacheKey, sql)

	return sql, nil
}

// ClassifyDocumentIntent returns "FORM", "RESEARCH", or "SUMMARY" based on user message and document content.
func (a *AIService) ClassifyDocumentIntent(userMessage, extractedText, aiResult string) (string, error) {
	ctx := context.Background()
	prompt := BuildDocumentIntentPrompt(userMessage, extractedText, aiResult)
	messages := []DashScopeMessage{{Role: "user", Content: prompt}}
	reply, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return "SUMMARY", err
	}
	s := strings.TrimSpace(strings.ToUpper(reply))
	if strings.Contains(s, "FORM") {
		return "FORM", nil
	}
	if strings.Contains(s, "RESEARCH") {
		return "RESEARCH", nil
	}
	return "SUMMARY", nil
}

// GenerateFormTemplateFromContent generates a FormTemplate (name, description, user_type, fields) from document content.
func (a *AIService) GenerateFormTemplateFromContent(content string, userContext string) (*models.FormTemplate, error) {
	ctx := context.Background()
	prompt := BuildFormTemplateFromContentPrompt(content, userContext)
	messages := []DashScopeMessage{{Role: "user", Content: prompt}}
	reply, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(reply)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	var t models.FormTemplate
	if err := json.Unmarshal([]byte(raw), &t); err != nil {
		return nil, fmt.Errorf("invalid form template JSON: %w", err)
	}
	if t.UserType == "" {
		t.UserType = "general"
	}
	if t.UserType != "student" && t.UserType != "staff" {
		t.UserType = "general"
	}
	return &t, nil
}

// GenerateSheetTemplateFromUserPrompt builds a FormTemplate from a chat-only request (sheet/form creation).
func (a *AIService) GenerateSheetTemplateFromUserPrompt(userPrompt string) (*models.FormTemplate, error) {
	ctx := context.Background()
	prompt := BuildSheetTemplateFromUserPrompt(userPrompt)
	messages := []DashScopeMessage{{Role: "user", Content: prompt}}
	reply, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(reply)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```JSON")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	var t models.FormTemplate
	if err := json.Unmarshal([]byte(raw), &t); err != nil {
		return nil, fmt.Errorf("invalid sheet template JSON: %w", err)
	}
	if t.UserType == "" {
		t.UserType = "general"
	}
	if t.UserType != "student" && t.UserType != "staff" {
		t.UserType = "general"
	}
	return &t, nil
}

func (a *AIService) GenerateFormHTMLPage(formJSON string) (string, error) {
	// Use context with longer timeout for HTML generation (5 minutes)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	// Parse form JSON to extract form name and description
	var formData map[string]interface{}
	if err := json.Unmarshal([]byte(formJSON), &formData); err != nil {
		return "", fmt.Errorf("failed to parse form JSON: %w", err)
	}

	formName := ""
	formDescription := ""
	if name, ok := formData["Name"].(string); ok {
		formName = name
	}
	if desc, ok := formData["Description"].(string); ok {
		formDescription = desc
	}

	// Build prompt using helper
	prompt := BuildFormHTMLPrompt(formJSON, formName, formDescription)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Use the long timeout client for HTML generation
	response, err := a.callDashScopeAPIWithClient(ctx, messages, true)
	if err != nil {
		return "", fmt.Errorf("failed to generate form HTML: %w", err)
	}

	html := strings.TrimSpace(response)
	// Remove markdown code blocks if present
	html = strings.TrimPrefix(html, "```html")
	html = strings.TrimPrefix(html, "```HTML")
	html = strings.TrimPrefix(html, "```")
	html = strings.TrimSuffix(html, "```")
	html = strings.TrimSpace(html)

	return html, nil
}

// GenerateChatResponse generates a plain chat response for general prompts
func (a *AIService) GenerateChatResponse(userPrompt string) (string, error) {
	return a.GenerateChatResponseWithAgent(userPrompt, "")
}

// GenerateChatResponseWithAgent extends the base assistant with optional agent-specific instructions (no cache when set).
func (a *AIService) GenerateChatResponseWithAgent(userPrompt, agentInstructions string) (string, error) {
	agentInstructions = strings.TrimSpace(agentInstructions)
	useCache := agentInstructions == ""

	if useCache {
		cacheKey := fmt.Sprintf("chat_prompt:%s", userPrompt)
		if cached, found := a.cache.Get(cacheKey); found {
			return cached.(string), nil
		}
	}

	ctx := context.Background()
	persona := "You are a helpful assistant."
	if agentInstructions != "" {
		persona += " " + agentInstructions
	}
	prompt := fmt.Sprintf("%s Please respond to the following user message in a helpful and informative way:\n\n%s", persona, userPrompt)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate chat response: %w", err)
	}

	chatResponse := strings.TrimSpace(response)
	chatResponse = strings.TrimPrefix(chatResponse, "```")
	chatResponse = strings.TrimSuffix(chatResponse, "```")
	chatResponse = strings.TrimSpace(chatResponse)

	if useCache {
		cacheKey := fmt.Sprintf("chat_prompt:%s", userPrompt)
		a.cache.SetDefault(cacheKey, chatResponse)
	}

	return chatResponse, nil
}

// ChatCompletion sends messages to the model without caching. Roles should alternate user/assistant as required by the provider.
func (a *AIService) ChatCompletion(ctx context.Context, messages []DashScopeMessage) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages")
	}
	return a.callDashScopeAPI(ctx, messages)
}

// ChatCompletionLong is like ChatCompletion but uses the longer HTTP timeout.
func (a *AIService) ChatCompletionLong(ctx context.Context, messages []DashScopeMessage) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages")
	}
	return a.callDashScopeAPIWithClient(ctx, messages, true)
}

// GenerateSessionTitle returns a short display title from the user's first message in a chat.
func (a *AIService) GenerateSessionTitle(userFirstMessage string) (string, error) {
	src := strings.TrimSpace(userFirstMessage)
	if len(src) > 800 {
		src = src[:800]
	}
	if src == "" {
		src = "New conversation"
	}
	ctx := context.Background()
	prompt := fmt.Sprintf(`Generate a very short chat session title (maximum 6 words, no quotes, no trailing punctuation, plain text only) that summarizes this user message. Reply with the title only, nothing else.

User message:
%s`, src)
	messages := []DashScopeMessage{{Role: "user", Content: prompt}}
	out, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	out = strings.Trim(out, `"'“”`)
	if i := strings.IndexAny(out, "\n\r"); i >= 0 {
		out = strings.TrimSpace(out[:i])
	}
	if len(out) > 80 {
		out = out[:80]
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "New chat", nil
	}
	return out, nil
}

// CorrectSpelling corrects spelling errors in user input using AI
// It preserves the user's intent while fixing typos and misspellings
func (a *AIService) CorrectSpelling(userInput string) (string, error) {
	// Skip correction for very short inputs or if input seems fine
	if len(userInput) < 3 {
		return userInput, nil
	}

	// Check cache first
	cacheKey := fmt.Sprintf("spell_correct:%s", userInput)
	if cached, found := a.cache.Get(cacheKey); found {
		return cached.(string), nil
	}

	ctx := context.Background()

	// Build prompt for spelling correction
	prompt := fmt.Sprintf(`You are a spelling and grammar correction assistant. Your task is to correct spelling errors and typos in the user's message while preserving their exact meaning and intent. 

IMPORTANT RULES:
1. Only correct actual spelling mistakes and typos
2. Preserve the user's original meaning and intent completely
3. Keep the same tone and style
4. Do NOT change words that are intentionally informal (like "wanna", "gonna", "yeah")
5. Do NOT add or remove words unless they are clearly typos
6. Fix spacing issues (e.g., "iwanna" -> "i wanna")
7. Return ONLY the corrected text, nothing else - no explanations, no markdown, just the corrected message

User's message to correct:
"%s"

Corrected message:`, userInput)

	messages := []DashScopeMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := a.callDashScopeAPI(ctx, messages)
	if err != nil {
		// If AI correction fails, return original input
		return userInput, nil
	}

	// Clean up the response
	corrected := strings.TrimSpace(response)
	// Remove any markdown code blocks if present
	corrected = strings.TrimPrefix(corrected, "```")
	corrected = strings.TrimSuffix(corrected, "```")
	corrected = strings.TrimSpace(corrected)

	// If correction is empty or same as input, return original
	if corrected == "" || corrected == userInput {
		return userInput, nil
	}

	// Cache the result
	a.cache.SetDefault(cacheKey, corrected)

	return corrected, nil
}

// GenerateFromMessages calls the model with the given message list (e.g. system + user + assistant + user).
// Used by registration flow and other custom prompts.
func (a *AIService) GenerateFromMessages(ctx context.Context, messages []DashScopeMessage) (string, error) {
	return a.callDashScopeAPI(ctx, messages)
}

// dashScopeUserPrompt combines instruction + payload into one user turn. DashScope text-generation
// is unreliable with role "system"; the rest of this service uses "user"-only messages.
func dashScopeUserPrompt(instruction, payload string) []DashScopeMessage {
	instruction = strings.TrimSpace(instruction)
	payload = strings.TrimSpace(payload)
	var b strings.Builder
	b.WriteString(instruction)
	if payload != "" {
		b.WriteString("\n\n---\n\n")
		b.WriteString(payload)
	}
	return []DashScopeMessage{{Role: "user", Content: b.String()}}
}

// RegistrationFormSelect asks the model to pick one form name from a list (no IDs). Returns the model reply (form name or NONE).
func (a *AIService) RegistrationFormSelect(ctx context.Context, userMessage, formNamesDescriptions string) (string, error) {
	sys, user := BuildFormSelectionPrompt(userMessage, formNamesDescriptions)
	return a.callDashScopeAPI(ctx, dashScopeUserPrompt(sys, user))
}

// RegistrationFieldGathering asks the model whether we have all required fields or what to ask next. Returns raw reply (JSON string).
func (a *AIService) RegistrationFieldGathering(ctx context.Context, conversationHistory []models.RegConvTurn, formFields []models.FormField, latestUserMessage string) (string, error) {
	sys, conv := BuildFieldGatheringPrompt(conversationHistory, formFields, latestUserMessage)
	return a.callDashScopeAPI(ctx, dashScopeUserPrompt(sys, conv))
}

// RegistrationFieldGatheringWithCurrent merges the user's change request into current answers (confirmation-edit flow).
func (a *AIService) RegistrationFieldGatheringWithCurrent(ctx context.Context, formFields []models.FormField, currentAnswers map[string]interface{}, userMessage string) (string, error) {
	sys, user := BuildFieldGatheringPromptWithCurrent(formFields, currentAnswers, userMessage)
	return a.callDashScopeAPI(ctx, dashScopeUserPrompt(sys, user))
}