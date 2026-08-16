package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"idongivaflyinfa/ai"
)

const (
	storyAIMaxImages     = 4
	storyAIDefaultImages = 1
	storyAIMaxPromptRunes = 2000
)

type storyAIGenerateRequest struct {
	Prompt     string `json:"prompt"`
	ImageCount *int   `json:"image_count"`
}

type storyAIImageOut struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	DataBase64  string `json:"data_base64"`
	Prompt      string `json:"prompt,omitempty"`
}

// GenerateStoryPostAI POST /api/tran/story-posts/ai-generate
// Builds a draft title, content, and optional generated images from a user prompt.
// Does not create the StoryPost — the client may edit then save via CreateStoryPost + attachments.
func (h *Handlers) GenerateStoryPostAI(c *gin.Context) {
	if h.aiService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}
	var in storyAIGenerateRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	if utf8.RuneCountInString(prompt) > storyAIMaxPromptRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("prompt must be at most %d characters", storyAIMaxPromptRunes)})
		return
	}
	imageCount := storyAIDefaultImages
	if in.ImageCount != nil {
		imageCount = *in.ImageCount
	}
	if imageCount < 0 {
		imageCount = 0
	}
	if imageCount > storyAIMaxImages {
		imageCount = storyAIMaxImages
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 4*time.Minute)
	defer cancel()

	title, content, imagePrompts, err := h.generateStoryTextAndImagePrompts(ctx, prompt, imageCount)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	images := make([]storyAIImageOut, 0, len(imagePrompts))
	var imageWarnings []string
	for i, ip := range imagePrompts {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		bytes, ctype, genErr := generateStoryImageBytes(ctx, ip)
		if genErr != nil {
			imageWarnings = append(imageWarnings, fmt.Sprintf("image %d: %v", i+1, genErr))
			continue
		}
		if len(bytes) == 0 {
			imageWarnings = append(imageWarnings, fmt.Sprintf("image %d: empty response", i+1))
			continue
		}
		ext := "png"
		if strings.Contains(ctype, "jpeg") || strings.Contains(ctype, "jpg") {
			ext = "jpg"
		} else if strings.Contains(ctype, "webp") {
			ext = "webp"
		}
		images = append(images, storyAIImageOut{
			Filename:    fmt.Sprintf("story-ai-%d.%s", i+1, ext),
			ContentType: ctype,
			DataBase64:  base64.StdEncoding.EncodeToString(bytes),
			Prompt:      ip,
		})
	}

	out := gin.H{
		"title":   title,
		"content": content,
		"images":  images,
	}
	if len(imageWarnings) > 0 {
		out["image_warnings"] = imageWarnings
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handlers) generateStoryTextAndImagePrompts(ctx context.Context, userPrompt string, imageCount int) (title, content string, imagePrompts []string, err error) {
	sys := `You write Morph Data Story Board posts.
Return ONLY a JSON object with keys:
- "title": short story title (max 80 characters)
- "content": engaging story body in plain text (2-6 short paragraphs, no markdown fences)
- "image_prompts": array of vivid English image-generation prompts that illustrate the story (one distinct scene each)`
	if imageCount <= 0 {
		sys += `
Set "image_prompts" to an empty array.`
	} else {
		sys += fmt.Sprintf(`
Include exactly %d items in "image_prompts".`, imageCount)
	}

	raw, err := h.aiService.ChatCompletionLong(ctx, []ai.DashScopeMessage{
		{Role: "system", Content: sys},
		{Role: "user", Content: "Story request:\n" + userPrompt},
	})
	if err != nil {
		return "", "", nil, fmt.Errorf("story generation failed: %w", err)
	}
	obj, ok := extractJSONObjectLoose(raw)
	if !ok {
		return "", "", nil, fmt.Errorf("AI did not return valid JSON for the story")
	}
	title = strings.TrimSpace(asString(obj["title"]))
	content = strings.TrimSpace(asString(obj["content"]))
	if title == "" {
		title = "Untitled story"
	}
	if utf8.RuneCountInString(title) > 120 {
		title = string([]rune(title)[:120])
	}
	if content == "" {
		return "", "", nil, fmt.Errorf("AI returned an empty story body")
	}
	imagePrompts = asStringSlice(obj["image_prompts"])
	if imageCount <= 0 {
		imagePrompts = nil
	} else if len(imagePrompts) > imageCount {
		imagePrompts = imagePrompts[:imageCount]
	}
	return title, content, imagePrompts, nil
}

func extractJSONObjectLoose(raw string) (map[string]any, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}
	// Strip optional markdown fence.
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSpace(raw)
		if strings.HasPrefix(strings.ToLower(raw), "json") {
			raw = strings.TrimSpace(raw[4:])
		}
		if i := strings.LastIndex(raw, "```"); i >= 0 {
			raw = strings.TrimSpace(raw[:i])
		}
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil, false
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(raw[start:end+1]), &obj); err != nil {
		return nil, false
	}
	return obj, true
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func asStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s := strings.TrimSpace(asString(item))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// generateStoryImageBytes fetches a generated image for prompt.
// Preference: Pollinations when MORPH_IMAGE_API_KEY / POLLINATIONS_API_KEY is set;
// otherwise DashScope wanx via MORPH_AI_API_KEY (same key Morph AI chat uses).
func generateStoryImageBytes(ctx context.Context, prompt string) ([]byte, string, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, "", fmt.Errorf("empty image prompt")
	}
	if pollKey := firstNonEmptyEnv("MORPH_IMAGE_API_KEY", "POLLINATIONS_API_KEY"); pollKey != "" {
		bytes, ctype, err := generateStoryImagePollinations(ctx, prompt, pollKey)
		if err == nil {
			return bytes, ctype, nil
		}
		// Fall through to DashScope when Pollinations fails.
	}
	if morphKey := firstNonEmptyEnv("MORPH_AI_API_KEY", "DASHSCOPE_API_KEY", "TRAN_QWEN_API_KEY", "GEMINI_API_KEY"); morphKey != "" {
		return generateStoryImageDashScope(ctx, prompt, morphKey)
	}
	return nil, "", fmt.Errorf("no image API key configured (set MORPH_IMAGE_API_KEY or MORPH_AI_API_KEY)")
}

func generateStoryImagePollinations(ctx context.Context, prompt, apiKey string) ([]byte, string, error) {
	encoded := url.PathEscape(prompt)
	imageURL := "https://gen.pollinations.ai/image/" + encoded + "?model=flux&nologo=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	return downloadStoryImage(ctx, req)
}

func generateStoryImageDashScope(ctx context.Context, prompt, apiKey string) ([]byte, string, error) {
	model := firstNonEmptyEnv("MORPH_IMAGE_MODEL", "MORPH_AI_IMAGE_MODEL")
	if model == "" {
		model = "wanx-v1"
	}
	payload, _ := json.Marshal(map[string]any{
		"model": model,
		"input": map[string]any{"prompt": prompt},
		"parameters": map[string]any{
			"size": "1024*1024",
			"n":    1,
		},
	})
	createURL := "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, createURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-Async", "enable")

	client := &http.Client{Timeout: 60 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	res.Body.Close()
	if err != nil {
		return nil, "", err
	}
	var created map[string]any
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, "", fmt.Errorf("dashscope image create: %s", truncateRunes(string(body), 200))
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg := asString(created["message"])
		if msg == "" {
			msg = truncateRunes(string(body), 200)
		}
		return nil, "", fmt.Errorf("dashscope image create: %s", msg)
	}
	out, _ := created["output"].(map[string]any)
	taskID := strings.TrimSpace(asString(out["task_id"]))
	if taskID == "" {
		return nil, "", fmt.Errorf("dashscope image create: missing task_id")
	}

	taskURL := "https://dashscope.aliyuncs.com/api/v1/tasks/" + url.PathEscape(taskID)
	deadline := time.Now().Add(3 * time.Minute)
	var imageURL string
	for {
		if err := ctx.Err(); err != nil {
			return nil, "", err
		}
		if time.Now().After(deadline) {
			return nil, "", fmt.Errorf("dashscope image timed out")
		}
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(2 * time.Second):
		}
		tReq, err := http.NewRequestWithContext(ctx, http.MethodGet, taskURL, nil)
		if err != nil {
			return nil, "", err
		}
		tReq.Header.Set("Authorization", "Bearer "+apiKey)
		tRes, err := client.Do(tReq)
		if err != nil {
			return nil, "", err
		}
		tBody, err := io.ReadAll(io.LimitReader(tRes.Body, 1<<20))
		tRes.Body.Close()
		if err != nil {
			return nil, "", err
		}
		var task map[string]any
		if err := json.Unmarshal(tBody, &task); err != nil {
			continue
		}
		tout, _ := task["output"].(map[string]any)
		status := strings.ToUpper(strings.TrimSpace(asString(tout["task_status"])))
		switch status {
		case "SUCCEEDED":
			results, _ := tout["results"].([]any)
			for _, item := range results {
				m, _ := item.(map[string]any)
				u := strings.TrimSpace(asString(m["url"]))
				if u != "" {
					imageURL = u
					break
				}
			}
			if imageURL == "" {
				return nil, "", fmt.Errorf("dashscope image succeeded but no url")
			}
		case "FAILED", "CANCELED", "UNKNOWN":
			msg := asString(tout["message"])
			if msg == "" {
				msg = asString(task["message"])
			}
			if msg == "" {
				msg = status
			}
			return nil, "", fmt.Errorf("dashscope image %s", msg)
		default:
			continue
		}
		break
	}

	dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	return downloadStoryImage(ctx, dlReq)
}

func downloadStoryImage(ctx context.Context, req *http.Request) ([]byte, string, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 20<<20))
	if err != nil {
		return nil, "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = res.Status
		}
		return nil, "", fmt.Errorf("image API %s", truncateRunes(msg, 200))
	}
	ctype := res.Header.Get("Content-Type")
	if ctype == "" || !strings.HasPrefix(ctype, "image/") {
		if len(body) >= 3 && body[0] == 0xff && body[1] == 0xd8 {
			ctype = "image/jpeg"
		} else if len(body) >= 8 && string(body[0:8]) == "\x89PNG\r\n\x1a\n" {
			ctype = "image/png"
		} else if ctype == "" || ctype == "application/octet-stream" {
			ctype = "image/png"
		}
	}
	return body, ctype, nil
}

func firstNonEmptyEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}
