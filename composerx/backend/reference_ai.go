package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
	pdf "github.com/ledongthuc/pdf"
)

const (
	chunkRunes     = 900
	chunkOverlap   = 120
	maxRAGChars    = 12000
	maxImageBytes  = 4 << 20 // OpenAI practical limit per image
	topChunksTotal = 12
)

// ReferenceDoc is stored in Badger for AI RAG (text + optional embeddings; images as files).
type ReferenceDoc struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Kind        string     `json:"kind"` // text | pdf | image
	FilePath    string     `json:"file_path,omitempty"`
	MimeType    string     `json:"mime_type,omitempty"`
	TextContent string     `json:"text_content,omitempty"`
	Chunks      []RefChunk `json:"chunks,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// RefChunk is a slice of reference text with an optional embedding vector.
type RefChunk struct {
	Index     int       `json:"index"`
	Text      string    `json:"text"`
	Embedding []float32 `json:"embedding,omitempty"`
}

type referenceDocListRow struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Kind       string    `json:"kind"`
	MimeType   string    `json:"mime_type,omitempty"`
	ChunkCount int       `json:"chunk_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type scoredChunk struct {
	DocName string
	Text    string
	Score   float64
}

func detectRefKind(filename, mime string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "pdf"
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return "image"
	default:
		if strings.HasPrefix(mime, "image/") {
			return "image"
		}
		if mime == "application/pdf" || strings.HasSuffix(filename, ".pdf") {
			return "pdf"
		}
	}
	return "text"
}

func mimeForFile(filename string, reported string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	}
	if reported != "" {
		return reported
	}
	return "application/octet-stream"
}

func normalizeWhitespace(s string) string {
	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func chunkRunesText(s string, size, overlap int) []string {
	runes := []rune(s)
	if len(runes) <= size {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		return []string{string(runes)}
	}
	var out []string
	step := size - overlap
	if step < 1 {
		step = 1
	}
	for i := 0; i < len(runes); i += step {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
		if end == len(runes) {
			break
		}
	}
	return out
}

func cosineSim(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func extractPDFPlainText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	reader, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, reader); err != nil {
		return "", err
	}
	return normalizeWhitespace(buf.String()), nil
}

func embedChunks(ctx context.Context, client *openai.Client, chunks []string) ([][]float32, error) {
	if client == nil || len(chunks) == 0 {
		return nil, nil
	}
	const batch = 32
	var all [][]float32
	for i := 0; i < len(chunks); i += batch {
		end := i + batch
		if end > len(chunks) {
			end = len(chunks)
		}
		batchInput := chunks[i:end]
		resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: openai.SmallEmbedding3,
			Input: batchInput,
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Data) != len(batchInput) {
			return nil, errors.New("embedding size mismatch")
		}
		for _, d := range resp.Data {
			f := make([]float32, len(d.Embedding))
			for j, v := range d.Embedding {
				f[j] = float32(v)
			}
			all = append(all, f)
		}
	}
	return all, nil
}

func ingestReferenceDocument(ctx context.Context, store *ReferenceDocsStore, client *openai.Client, filePath, filename, mime string) (*ReferenceDoc, error) {
	kind := detectRefKind(filename, mime)
	mime = mimeForFile(filename, mime)
	id := uuid.NewString()
	doc := &ReferenceDoc{
		ID:        id,
		Name:      filename,
		Kind:      kind,
		FilePath:  filePath,
		MimeType:  mime,
		CreatedAt: time.Now().UTC(),
	}

	switch kind {
	case "image":
		doc.TextContent = ""
		doc.Chunks = nil
	case "pdf":
		txt, err := extractPDFPlainText(filePath)
		if err != nil {
			return nil, fmt.Errorf("pdf extract: %w", err)
		}
		doc.TextContent = txt
		parts := chunkRunesText(txt, chunkRunes, chunkOverlap)
		doc.Chunks = make([]RefChunk, 0, len(parts))
		vecs, err := embedChunks(ctx, client, parts)
		if err != nil {
			return nil, fmt.Errorf("embed: %w", err)
		}
		for i, p := range parts {
			c := RefChunk{Index: i, Text: p}
			if i < len(vecs) && len(vecs[i]) > 0 {
				c.Embedding = vecs[i]
			}
			doc.Chunks = append(doc.Chunks, c)
		}
	default:
		raw, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		txt := normalizeWhitespace(string(raw))
		doc.TextContent = txt
		parts := chunkRunesText(txt, chunkRunes, chunkOverlap)
		doc.Chunks = make([]RefChunk, 0, len(parts))
		vecs, err := embedChunks(ctx, client, parts)
		if err != nil {
			return nil, fmt.Errorf("embed: %w", err)
		}
		for i, p := range parts {
			c := RefChunk{Index: i, Text: p}
			if i < len(vecs) && len(vecs[i]) > 0 {
				c.Embedding = vecs[i]
			}
			doc.Chunks = append(doc.Chunks, c)
		}
	}

	if err := store.Put(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func loadReferenceDocsByIDs(ctx context.Context, store *ReferenceDocsStore, ids []string) ([]ReferenceDoc, error) {
	return store.GetByIDs(ctx, ids)
}

func retrieveRAG(ctx context.Context, client *openai.Client, query string, docs []ReferenceDoc) (textContext string, imageParts []openai.ChatMessagePart, err error) {
	var embedQuery []float32
	if client != nil && strings.TrimSpace(query) != "" {
		resp, e := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: openai.SmallEmbedding3,
			Input: []string{query},
		})
		if e == nil && len(resp.Data) == 1 {
			embedQuery = make([]float32, len(resp.Data[0].Embedding))
			for i, v := range resp.Data[0].Embedding {
				embedQuery[i] = float32(v)
			}
		}
	}

	var scored []scoredChunk
	var fallback strings.Builder

	for _, d := range docs {
		if d.Kind == "image" && d.FilePath != "" {
			info, statErr := os.Stat(d.FilePath)
			if statErr != nil || info.Size() > maxImageBytes {
				continue
			}
			raw, readErr := os.ReadFile(d.FilePath)
			if readErr != nil {
				continue
			}
			b64 := base64.StdEncoding.EncodeToString(raw)
			mime := d.MimeType
			if mime == "" {
				mime = "image/jpeg"
			}
			dataURL := "data:" + mime + ";base64," + b64
			imageParts = append(imageParts, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL: dataURL,
				},
			})
			continue
		}

		useEmbed := len(embedQuery) > 0
		for _, ch := range d.Chunks {
			if strings.TrimSpace(ch.Text) == "" {
				continue
			}
			if useEmbed && len(ch.Embedding) == len(embedQuery) {
				s := cosineSim(embedQuery, ch.Embedding)
				scored = append(scored, scoredChunk{DocName: d.Name, Text: ch.Text, Score: s})
			} else {
				if fallback.Len() > 0 {
					fallback.WriteString("\n\n")
				}
				fallback.WriteString(ch.Text)
			}
		}
		if !useEmbed && strings.TrimSpace(d.TextContent) != "" && len(d.Chunks) == 0 {
			if fallback.Len() > 0 {
				fallback.WriteString("\n\n")
			}
			fallback.WriteString(d.TextContent)
		}
	}

	if len(scored) > 0 {
		sort.Slice(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
		if len(scored) > topChunksTotal {
			scored = scored[:topChunksTotal]
		}
		var buf strings.Builder
		for _, s := range scored {
			if buf.Len() > 0 {
				buf.WriteString("\n\n---\n\n")
			}
			fmt.Fprintf(&buf, "[Source: %s]\n%s", s.DocName, s.Text)
		}
		textContext = buf.String()
	} else {
		textContext = fallback.String()
	}

	if len(textContext) > maxRAGChars {
		textContext = textContext[:maxRAGChars] + "\n…"
	}
	return textContext, imageParts, nil
}

type composerAIRequest struct {
	Messages             []composerAIMessage `json:"messages"`
	ReferenceDocIDs      []string            `json:"reference_document_ids"`
	CurrentMarkdown      string              `json:"current_markdown"`
	CurrentEmailHTML     string              `json:"current_email_html"` // legacy alias
	UseWebSearch         *bool               `json:"use_web_search"`
	WebSearchQuery       string              `json:"web_search_query"`
}

type composerAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type composerAIResponse struct {
	AssistantMessage    string              `json:"assistant_message"`
	ProposedMarkdown    *string             `json:"proposed_markdown,omitempty"`
	ProposedEmailHTML   *string             `json:"proposed_email_html,omitempty"` // legacy alias
	ResearchNotes       string              `json:"research_notes,omitempty"`
	Sources             []webresearchSource `json:"sources,omitempty"`
}

// composerAIResponseFlexible captures alternate JSON keys some chat models emit.
type composerAIResponseFlexible struct {
	AssistantSnake    string  `json:"assistant_message"`
	AssistantCamel    string  `json:"assistantMessage"`
	ProposedMarkdown  *string `json:"proposed_markdown"`
	ProposedSnake     *string `json:"proposed_email_html"`
	ProposedCamel     *string `json:"proposedEmailHtml"`
}

func mergeComposerAIFields(f composerAIResponseFlexible) (string, *string) {
	msg := strings.TrimSpace(f.AssistantSnake)
	if msg == "" {
		msg = strings.TrimSpace(f.AssistantCamel)
	}
	var prop *string
	if f.ProposedMarkdown != nil && strings.TrimSpace(*f.ProposedMarkdown) != "" {
		prop = f.ProposedMarkdown
	} else if f.ProposedSnake != nil && strings.TrimSpace(*f.ProposedSnake) != "" {
		prop = f.ProposedSnake
	} else if f.ProposedCamel != nil && strings.TrimSpace(*f.ProposedCamel) != "" {
		prop = f.ProposedCamel
	}
	return msg, prop
}

func (a *App) listReferenceDocuments(c *gin.Context) {
	ctx := c.Request.Context()
	docs, err := a.referenceDocs.List(ctx, 500)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reference documents"})
		return
	}
	rows := make([]referenceDocListRow, 0, len(docs))
	for _, d := range docs {
		rows = append(rows, referenceDocListRow{
			ID:         d.ID,
			Name:       d.Name,
			Kind:       d.Kind,
			MimeType:   d.MimeType,
			ChunkCount: len(d.Chunks),
			CreatedAt:  d.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": rows, "total": len(rows)})
}

func (a *App) uploadReferenceDocument(c *gin.Context) {
	ctx := c.Request.Context()
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	name := c.PostForm("name")
	if name == "" {
		name = header.Filename
	}
	mime := c.PostForm("mime_type")
	if mime == "" {
		mime = header.Header.Get("Content-Type")
	}

	dir := filepath.Join(a.cfg.FileStorePath, "reference-docs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare storage"})
		return
	}

	filename := time.Now().UTC().Format("20060102T150405") + "_" + sanitizeUploadBase(header.Filename)
	targetPath := filepath.Join(dir, filename)

	dst, err := os.Create(targetPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		_ = os.Remove(targetPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
		return
	}
	dst.Close()

	client := openAIClientOptional(a.cfg.OpenAIAPIKey)
	doc, err := ingestReferenceDocument(ctx, a.referenceDocs, client, targetPath, name, mime)
	if err != nil {
		_ = os.Remove(targetPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": doc.ID, "name": doc.Name, "kind": doc.Kind})
}

func sanitizeUploadBase(name string) string {
	base := filepath.Base(name)
	if base == "." || base == "/" || base == "" {
		return "upload"
	}
	return strings.Map(func(r rune) rune {
		if r <= 31 || r == '/' || r == '\\' || r == '"' || r == '\'' {
			return '_'
		}
		return r
	}, base)
}

func openAIClientOptional(key string) *openai.Client {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	return openai.NewClient(key)
}

func (a *App) deleteReferenceDocument(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	doc, err := a.referenceDocs.Get(ctx, id)
	if err != nil {
		if errors.Is(err, errReferenceDocNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "reference document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load reference document"})
		return
	}
	if err := a.referenceDocs.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete reference document"})
		return
	}
	if doc.FilePath != "" {
		_ = os.Remove(doc.FilePath)
	}
	c.Status(http.StatusNoContent)
}

func (a *App) downloadReferenceDocument(c *gin.Context) {
	ctx := c.Request.Context()
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	doc, err := a.referenceDocs.Get(ctx, id)
	if err != nil {
		if errors.Is(err, errReferenceDocNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "reference document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load reference document"})
		return
	}
	name := strings.TrimSpace(doc.Name)
	if name == "" {
		name = "reference-" + id
	}
	if strings.TrimSpace(doc.FilePath) != "" {
		c.FileAttachment(doc.FilePath, name)
		return
	}
	if strings.TrimSpace(doc.TextContent) != "" {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(doc.TextContent))
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "file not available"})
}

func (a *App) composerAIChat(c *gin.Context) {
	chatClient, chatModel, ok := a.chatCompletionClient()
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI assistant is not configured — add qwen_api_key to ai.config.json (copy from ai.config.example.json alongside the binary), set TRAN_AI_CONFIG_PATH if needed, or configure TRAN_OPENAI_API_KEY",
		})
		return
	}
	ctx := c.Request.Context()
	var payload composerAIRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if len(payload.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages is required"})
		return
	}

	if len(payload.ReferenceDocIDs) > 3 {
		payload.ReferenceDocIDs = payload.ReferenceDocIDs[:3]
	}

	var lastUser string
	for i := len(payload.Messages) - 1; i >= 0; i-- {
		if payload.Messages[i].Role == "user" {
			lastUser = payload.Messages[i].Content
			break
		}
	}

	var ragText string
	var imageParts []openai.ChatMessagePart
	if len(payload.ReferenceDocIDs) > 0 {
		docs, err := loadReferenceDocsByIDs(ctx, a.referenceDocs, payload.ReferenceDocIDs)
		if err != nil {
			log.Printf("composer-chat: reference docs unavailable, continuing without RAG: %v", err)
		} else {
			// Query embeddings use OpenAI-compatible embedder when TRAN_OPENAI_API_KEY is set (chunk vectors are OpenAI).
			ragEmbedClient := openAIClientOptional(a.cfg.OpenAIAPIKey)
			ragText, imageParts, err = retrieveRAG(ctx, ragEmbedClient, lastUser, docs)
			if err != nil {
				log.Printf("composer-chat: RAG retrieval failed, continuing without excerpts: %v", err)
			}
		}
	}

	system := `You are an expert copywriter and markdown document composer for a content workspace.
This workspace saves markdown documents only — never HTML email templates, table layouts, or inline CSS.

CRITICAL: Output a single JSON object only, with exactly these keys:
- "assistant_message": short conversational reply to the user (plain text, no markdown fences). Summarize what you wrote; do not mention email templates or HTML.
- "proposed_markdown": full markdown document body for the draft, OR null if you are only answering a question and the draft should not change yet.

Use standard markdown (headings, lists, links, emphasis). Keep tone professional unless asked otherwise.
When the user asks for a document, brochure, agenda, invitation, or similar, you MUST put the complete markdown in proposed_markdown.
Never leave proposed_markdown null when generating a new document. Never put the document body only in assistant_message.

Example:
{"assistant_message":"I've drafted a 3-day workshop agenda for your review.","proposed_markdown":"# Workshop\\n\\n..."}`

	currentDraft := strings.TrimSpace(payload.CurrentMarkdown)
	if currentDraft == "" {
		currentDraft = strings.TrimSpace(payload.CurrentEmailHTML)
	}
	if currentDraft != "" {
		system += "\n\nCurrent draft markdown (may be edited or replaced as the user asks):\n" + currentDraft
	}
	if ragText != "" {
		system += "\n\nReference excerpts (ground truth; prefer facts from here when relevant):\n" + ragText
	}

	var researchNotes string
	var sources []webresearchSource
	if webSearchEnabledComposer(payload.UseWebSearch) {
		if q := resolveWebSearchQuery(payload.Messages, payload.WebSearchQuery, lastUser); q != "" {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("composer-chat: web search panic: %v", r)
					}
				}()
				notes, srcs := gatherWebContext(q)
				researchNotes = notes
				for _, s := range srcs {
					sources = append(sources, webresearchSource{Title: s.Title, Type: s.Type, URL: s.URL})
				}
			}()
			if researchNotes != "" {
				system += "\n\nWeb research (ground truth for topic and document content):\n" + researchNotes
			}
		}
	}

	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: system},
	}

	if len(imageParts) > 0 {
		parts := []openai.ChatMessagePart{{Type: openai.ChatMessagePartTypeText, Text: "Reference images the user attached to this project (describe or incorporate faithfully when relevant):"}}
		parts = append(parts, imageParts...)
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:         openai.ChatMessageRoleUser,
			MultiContent: parts,
		})
	}

	for _, m := range payload.Messages {
		switch m.Role {
		case "user":
			msgs = append(msgs, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: m.Content})
		case "assistant":
			msgs = append(msgs, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: m.Content})
		default:
			continue
		}
	}

	resp, err := chatClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    chatModel,
		Messages: msgs,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("AI chat: %v", err)})
		return
	}
	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "empty model response"})
		return
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	msg, prop := safeParseComposerAIResponse(raw)
	if msg == "" && prop != nil {
		msg = "Document draft updated in the editor — review Preview or Markdown on the left."
	}
	if msg == "" {
		msg = raw
	}
	c.JSON(http.StatusOK, composerAIResponse{
		AssistantMessage: msg,
		ProposedMarkdown: prop,
		ResearchNotes:    researchNotes,
		Sources:          sources,
	})
}
