package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

const publishSourceMaxExtract = 48000

type publishSourceMaterial struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Summary string `json:"summary"`
	Error   string `json:"error,omitempty"`
}

type publishSourceProcessResponse struct {
	Items        []publishSourceMaterial `json:"items"`
	CombinedText string                  `json:"combined_text"`
}

func publishSourceKind(name, mime string) string {
	k := detectRefKind(name, mime)
	if k == "image" {
		return "image"
	}
	if k == "pdf" || k == "text" {
		return "document"
	}
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".txt", ".md", ".markdown", ".csv", ".json", ".html", ".htm", ".xml":
		return "document"
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return "image"
	case ".pdf":
		return "document"
	}
	if strings.HasPrefix(strings.ToLower(mime), "image/") {
		return "image"
	}
	if mime == "application/pdf" {
		return "document"
	}
	if strings.HasPrefix(strings.ToLower(mime), "text/") {
		return "document"
	}
	return "unsupported"
}

func extractPublishSourceText(filePath, name, mime string) (string, error) {
	kind := publishSourceKind(name, mime)
	switch kind {
	case "image":
		return "", fmt.Errorf("image files are described visually, not as text")
	case "document":
		if detectRefKind(name, mime) == "pdf" {
			return extractPDFPlainText(filePath)
		}
		raw, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return normalizeWhitespace(string(raw)), nil
	default:
		return "", fmt.Errorf("unsupported file type for text extraction (try PDF, TXT, MD, CSV, or an image)")
	}
}

func describeImageForPublish(ctx context.Context, client *openai.Client, model, filePath, mime, filename string) (string, error) {
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("image not available")
	}
	if info.Size() > maxImageBytes {
		return "", fmt.Errorf("image exceeds size limit")
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	if mime == "" {
		mime = mimeForFile(filename, "")
	}
	if !strings.HasPrefix(mime, "image/") {
		mime = "image/jpeg"
	}
	dataURL := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: "Describe this image in detail for a web page designer. Include subject, layout, colors, text visible in the image, mood, and how it could be used on a landing page. Be factual and concise (under 400 words).",
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: dataURL,
					},
				},
			},
		}},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty vision response")
	}
	out := strings.TrimSpace(resp.Choices[0].Message.Content)
	if out == "" {
		return "", fmt.Errorf("empty vision response")
	}
	return out, nil
}

func summarizeDocumentForPublish(ctx context.Context, client *openai.Client, model, filename, extracted string) (string, error) {
	extracted = strings.TrimSpace(extracted)
	if extracted == "" {
		return "", fmt.Errorf("no text found in document")
	}
	if len(extracted) > publishSourceMaxExtract {
		extracted = extracted[:publishSourceMaxExtract] + "\n...(truncated)"
	}
	prompt := fmt.Sprintf(`Summarize the following document for use when composing a public HTML landing page.
Filename: %s
Return a clear outline: key messages, sections, facts, calls-to-action, and tone. Under 600 words.

--- document ---
%s`, filename, extracted)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You summarize documents for web designers. Be accurate and structured."},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty summary response")
	}
	out := strings.TrimSpace(resp.Choices[0].Message.Content)
	if out == "" {
		return "", fmt.Errorf("empty summary response")
	}
	return out, nil
}

func buildCombinedPublishSourceText(items []publishSourceMaterial) string {
	var b strings.Builder
	for _, it := range items {
		if strings.TrimSpace(it.Summary) == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n\n---\n\n")
		}
		label := strings.TrimSpace(it.Name)
		if label == "" {
			label = "Source"
		}
		kind := strings.TrimSpace(it.Kind)
		if kind == "" {
			kind = "file"
		}
		fmt.Fprintf(&b, "### %s (%s)\n%s", label, kind, strings.TrimSpace(it.Summary))
	}
	return b.String()
}

func (a *App) processPublishSourceFile(ctx context.Context, chatClient *openai.Client, chatModel, filePath, name, mime string) publishSourceMaterial {
	kind := publishSourceKind(name, mime)
	item := publishSourceMaterial{
		Name: name,
		Kind: kind,
	}
	switch kind {
	case "image":
		summary, err := describeImageForPublish(ctx, chatClient, chatModel, filePath, mime, name)
		if err != nil {
			item.Error = err.Error()
		} else {
			item.Summary = summary
		}
	case "document":
		extracted, err := extractPublishSourceText(filePath, name, mime)
		if err != nil {
			item.Error = err.Error()
		} else {
			summary, err := summarizeDocumentForPublish(ctx, chatClient, chatModel, name, extracted)
			if err != nil {
				item.Error = err.Error()
			} else {
				item.Summary = summary
			}
		}
	default:
		item.Kind = "unsupported"
		item.Error = "unsupported file type (use PDF, text, CSV, or images)"
	}
	return item
}

func (a *App) processPublishSources(c *gin.Context) {
	chatClient, chatModel, ok := a.chatCompletionClient()
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI assistant is not configured"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil || len(form.File["files"]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one file is required (multipart field: files)"})
		return
	}
	files := form.File["files"]
	if len(files) > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at most 12 files per request"})
		return
	}

	tempDir, err := os.MkdirTemp("", "publish-src-")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare temp storage"})
		return
	}
	defer os.RemoveAll(tempDir)

	ctx := c.Request.Context()
	items := make([]publishSourceMaterial, 0, len(files))
	for i, fh := range files {
		name := strings.TrimSpace(fh.Filename)
		if name == "" {
			name = fmt.Sprintf("file-%d", i+1)
		}
		mime := fh.Header.Get("Content-Type")
		dest := filepath.Join(tempDir, fmt.Sprintf("%d_%s", i, filepath.Base(name)))
		if err := c.SaveUploadedFile(fh, dest); err != nil {
			items = append(items, publishSourceMaterial{
				Name:  name,
				Kind:  "unsupported",
				Error: "failed to save uploaded file",
			})
			continue
		}
		items = append(items, a.processPublishSourceFile(ctx, chatClient, chatModel, dest, name, mime))
	}

	c.JSON(http.StatusOK, publishSourceProcessResponse{
		Items:        items,
		CombinedText: buildCombinedPublishSourceText(items),
	})
}
