package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/formsx/backend/internal/surveybot"
	"github.com/gin-gonic/gin"
	"github.com/robo/docextract"
	"github.com/robo/morphai"
)

const (
	// maxSurveySourceBytes caps a single upload for survey generation.
	maxSurveySourceBytes = 8 << 20 // 8 MiB

	// surveyGenerationAttempts is the initial attempt plus one repair retry.
	surveyGenerationAttempts = 2

	// visionNoContentSentinel is what the vision model must reply when an image
	// holds nothing readable, so an unreadable upload fails instead of yielding
	// an invented survey.
	visionNoContentSentinel = "NO_READABLE_CONTENT"
)

var (
	surveyFrontMatterRe = regexp.MustCompile(`(?s)\A---[ \t]*\n(.*?)\n---[ \t]*\n?(.*)\z`)
	surveyCodeFenceRe   = regexp.MustCompile("(?s)\\A```[a-zA-Z]*\\s*\\n(.*?)\\n?```\\s*\\z")
)

const surveyFromFileSystemPrompt = `You turn source documents into Survey Bot markdown templates.

Output ONLY the markdown template. No commentary, no code fences.

Required structure:
---
slug: kebab-case-slug
title: Human readable title
tags: [survey, survey-bot]
---

# Instructions
One or two sentences telling the bot how to conduct the survey.

## Q1 — Short label
- field: snake_case_name
- collect: text
- required: true
- prompt: The question to ask the respondent.

## Q2 — Short label
- field: another_name
- collect: mcp_html
- widget: select
- options: [First choice, Second choice, Third choice]
- required: true
- prompt: The question to ask the respondent.

Rules:
- One "## Q<n> — <label>" block per question found in the source, numbered from 1 with no gaps.
- A question whose source offers a fixed set of answers MUST use "collect: mcp_html" with "widget: select" (or "widget: multiselect" when several answers apply) and MUST list every choice in "options". A select or multiselect without options is invalid.
- A question with an open-ended or written answer MUST use "collect: text" and MUST NOT have options.
- "field" values are lowercase snake_case, unique within the template.
- "required: true" unless the source marks the question as optional.
- Keep the respondent's wording where the source provides it. Do not invent questions that are not in the source.
- Write at least one question block. Never emit a template with no "## Q" sections.`

type surveyFromFileResult struct {
	Markdown         string `json:"markdown"`
	Title            string `json:"title"`
	Slug             string `json:"slug"`
	SourceName       string `json:"source_name"`
	SourceKind       string `json:"source_kind"`
	SourceText       string `json:"source_text"`
	SourceTruncated  bool   `json:"source_truncated"`
	QuestionCount    int    `json:"question_count"`
	AssistantMessage string `json:"assistant_message"`
}

// DraftSurveyBotTemplateFromFile POST /survey-bot/templates/from-file
//
// Reads one uploaded PDF or image and returns a Survey Bot markdown draft. It
// never persists a template — the caller saves through the normal create route.
func (h *Handler) DraftSurveyBotTemplateFromFile(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected a multipart upload: " + err.Error()})
		return
	}

	files := form.File["file"]
	files = append(files, form.File["files"]...)
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": `at least one file is required (multipart field "file")`})
		return
	}
	if len(files) > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only one file is accepted per request"})
		return
	}

	fh := files[0]
	name := strings.TrimSpace(fh.Filename)
	if name == "" {
		name = "upload"
	}
	mime := fh.Header.Get("Content-Type")

	kind := docextract.Classify(name, mime)
	isPDF := docextract.IsPDF(name, mime)
	if kind == docextract.KindUnsupported || (kind == docextract.KindDocument && !isPDF) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unsupported file type for survey generation; " +
				docextract.AcceptedTypesMessage(docextract.KindImage) +
				", or PDF. Markdown and text files load straight into the editor instead",
		})
		return
	}
	if fh.Size > maxSurveySourceBytes {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("file too large (max %d MiB)", maxSurveySourceBytes/(1024*1024)),
		})
		return
	}

	if !h.AI.Configured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI is not configured — set MORPH_AI_API_KEY to generate surveys from files",
		})
		return
	}

	ctx := c.Request.Context()
	sourceText, sourceKind, err := h.readSurveySource(ctx, c, fh, name, mime, isPDF)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errVisionUnavailable) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	sourceText, truncated := docextract.Truncate(sourceText, docextract.MaxPerFileChars)

	titleHint := strings.TrimSpace(c.PostForm("title_hint"))
	extraInstructions := strings.TrimSpace(c.PostForm("instructions"))

	md, parsed, genErr := h.generateSurveyMarkdown(ctx, sourceText, name, sourceKind, titleHint, extraInstructions)
	if genErr != nil {
		if md != "" {
			// Validation failure: hand back the rejected markdown so it can be repaired by hand.
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":       "generated template failed validation: " + genErr.Error(),
				"markdown":    md,
				"source_text": sourceText,
				"source_name": name,
				"source_kind": sourceKind,
			})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": genErr.Error()})
		return
	}

	c.JSON(http.StatusOK, surveyFromFileResult{
		Markdown:        md,
		Title:           parsed.Title,
		Slug:            parsed.Slug,
		SourceName:      name,
		SourceKind:      sourceKind,
		SourceText:      sourceText,
		SourceTruncated: truncated,
		QuestionCount:   len(parsed.Steps),
		AssistantMessage: fmt.Sprintf(
			"Drafted %d question(s) from %s. Review and save, or ask me to adjust.",
			len(parsed.Steps), name),
	})
}

var errVisionUnavailable = errors.New("vision unavailable")

// readSurveySource extracts text from a PDF or reads an image with a vision model.
func (h *Handler) readSurveySource(
	ctx context.Context,
	c *gin.Context,
	fh *multipart.FileHeader,
	name, mime string,
	isPDF bool,
) (string, string, error) {
	if isPDF {
		tmpDir, err := os.MkdirTemp("", "survey-src-")
		if err != nil {
			return "", "", fmt.Errorf("could not prepare temporary storage")
		}
		defer os.RemoveAll(tmpDir)

		dest := filepath.Join(tmpDir, "upload.pdf")
		if err := c.SaveUploadedFile(fh, dest); err != nil {
			return "", "", fmt.Errorf("could not save the uploaded file")
		}
		text, err := docextract.ExtractPDF(dest)
		if err != nil {
			return "", "", err
		}
		return text, "pdf", nil
	}

	if !h.AI.VisionSupported() {
		return "", "", fmt.Errorf("%w: reading images needs an OpenAI-compatible AI endpoint; "+
			"unset MORPH_AI_API_URL or point it at a /v1 base URL", errVisionUnavailable)
	}

	src, err := fh.Open()
	if err != nil {
		return "", "", fmt.Errorf("could not read the uploaded file")
	}
	defer src.Close()

	raw, err := io.ReadAll(io.LimitReader(src, maxSurveySourceBytes+1))
	if err != nil {
		return "", "", fmt.Errorf("could not read the uploaded file")
	}
	if int64(len(raw)) > maxSurveySourceBytes {
		return "", "", fmt.Errorf("file too large (max %d MiB)", maxSurveySourceBytes/(1024*1024))
	}
	if len(raw) == 0 {
		return "", "", fmt.Errorf("the uploaded image is empty")
	}

	reply, err := h.AI.ChatCompletionVision(ctx, []morphai.MultiMessage{
		morphai.UserMultiMessage(
			morphai.TextPart(
				"Transcribe this image for survey building. List every question exactly as written, "+
					"and under each question list its answer choices if any are shown. "+
					"Mark questions that appear optional. Include any title, section headings, and instructions. "+
					"Report only what is visible; do not invent questions.\n"+
					"If the image is blank, unreadable, or shows nothing that could be a form or questionnaire, "+
					"reply with exactly "+visionNoContentSentinel+" and nothing else."),
			morphai.ImageDataPart(mime, raw),
		),
	}, "")
	if err != nil {
		return "", "", fmt.Errorf("could not read the image: %w", err)
	}
	reply = strings.TrimSpace(reply)
	if reply == "" || strings.Contains(reply, visionNoContentSentinel) {
		// Without readable content the model would invent a survey from nothing.
		return "", "", fmt.Errorf(
			"nothing readable was found in this image — check that it is in focus and shows the questionnaire, or type a description instead")
	}
	return reply, "image", nil
}

// generateSurveyMarkdown asks for a template and validates it, retrying once
// with the parser's complaint fed back in. On failure it returns the rejected
// markdown alongside the error so the caller can surface both.
func (h *Handler) generateSurveyMarkdown(
	ctx context.Context,
	sourceText, sourceName, sourceKind, titleHint, extraInstructions string,
) (string, *surveybot.ParsedTemplate, error) {
	var lastMarkdown string
	var lastErr error

	for attempt := 0; attempt < surveyGenerationAttempts; attempt++ {
		userPrompt := buildSurveyFromFilePrompt(sourceText, sourceName, sourceKind, titleHint, extraInstructions)
		if attempt > 0 && lastErr != nil {
			userPrompt += fmt.Sprintf(
				"\n\nYour previous answer was rejected by the template validator with this error:\n%s\n"+
					"Return corrected markdown that fixes it. Remember: every select or multiselect needs an options list, "+
					"and the template needs at least one \"## Q<n>\" block.",
				lastErr.Error())
		}

		reply, err := h.AI.ChatCompletionLong(ctx, []morphai.Message{
			{Role: "system", Content: surveyFromFileSystemPrompt},
			{Role: "user", Content: userPrompt},
		})
		if err != nil {
			return "", nil, fmt.Errorf("AI request failed: %w", err)
		}

		md := normalizeSurveyMarkdown(reply, titleHint)
		parsed, parseErr := surveybot.ParseMarkdown(md)
		if parseErr == nil && parsed.NeedsCompile {
			parseErr = fmt.Errorf("template has no question blocks — add at least one \"## Q1 — …\" section")
		}
		if parseErr == nil {
			return md, parsed, nil
		}
		lastMarkdown = md
		lastErr = parseErr
	}

	return lastMarkdown, nil, lastErr
}

func buildSurveyFromFilePrompt(sourceText, sourceName, sourceKind, titleHint, extraInstructions string) string {
	var b strings.Builder

	b.WriteString("Build a Survey Bot markdown template from the source below.\n\n")
	if titleHint != "" {
		fmt.Fprintf(&b, "Use this title: %s\n", titleHint)
	}
	if extraInstructions != "" {
		fmt.Fprintf(&b, "Extra instructions from the user: %s\n", extraInstructions)
	}

	label := "document text"
	if sourceKind == "image" {
		label = "image transcription"
	}
	fmt.Fprintf(&b, "\n--- source (%s of %s) ---\n%s\n--- end source ---\n", label, sourceName, sourceText)

	return b.String()
}

// normalizeSurveyMarkdown strips any code fence the model added and rewrites the
// front matter so slug, title, and tags are always present and well formed.
func normalizeSurveyMarkdown(md, titleHint string) string {
	md = strings.TrimSpace(md)
	if m := surveyCodeFenceRe.FindStringSubmatch(md); len(m) == 2 {
		md = strings.TrimSpace(m[1])
	}

	body := md
	fmSlug, fmTitle, fmTags := "", "", ""
	if m := surveyFrontMatterRe.FindStringSubmatch(md); len(m) == 3 {
		fmSlug, fmTitle, fmTags = parseSurveyFrontMatter(m[1])
		body = strings.TrimSpace(m[2])
	}

	title := strings.TrimSpace(titleHint)
	if title == "" {
		title = fmTitle
	}
	if title == "" {
		title = firstMarkdownH1(body)
	}
	if title == "" {
		title = "Survey"
	}

	slug := normalizeSlug(fmSlug)
	if slug == "" {
		slug = normalizeSlug(title)
	}
	if slug == "" {
		slug = "survey"
	}

	tags := strings.TrimSpace(fmTags)
	if tags == "" {
		tags = "[survey, survey-bot]"
	}

	return fmt.Sprintf("---\nslug: %s\ntitle: %s\ntags: %s\n---\n\n%s\n", slug, title, tags, body)
}

func parseSurveyFrontMatter(raw string) (slug, title, tags string) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		switch key {
		case "slug", "id":
			slug = val
		case "title":
			title = val
		case "tags":
			tags = val
		}
	}
	return slug, title, tags
}

func firstMarkdownH1(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			h := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			// "# Instructions" is the template's own section heading, not a title.
			if !strings.EqualFold(h, "instructions") {
				return h
			}
		}
	}
	return ""
}
