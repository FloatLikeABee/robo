package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleChatWithFile processes an uploaded image or PDF: extracts content, classifies intent, then form/research/summary.
func (h *Handlers) handleChatWithFile(c *gin.Context, userID, userMessage string, fileHeader *multipart.FileHeader, action string) (*models.ChatResponse, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open uploaded file: %w", err)
	}
	defer file.Close()

	ext := strings.ToLower(path.Ext(fileHeader.Filename))
	// Always use summarize for extraction; user message is used later for intent and research.
	systemPrompt := "Summarize the following content clearly and concisely."

	var extractedText, aiResult string
	isPDF := ext == ".pdf"
	if isPDF {
		extractedText, aiResult, err = h.ReadPDFAndProcess(file, fileHeader.Filename, systemPrompt)
	} else {
		extractedText, aiResult, err = h.ReadImageAndProcess(file, fileHeader.Filename, systemPrompt)
	}
	if err != nil {
		log.Printf("[CHAT FILE] Extract/process error: %v", err)
		return &models.ChatResponse{
			Response: fmt.Sprintf("Could not process the uploaded file: %v. Make sure the Image Reader / PDF Reader service is running at %s.", err, h.externalAPIBase),
		}, nil
	}

	if strings.EqualFold(strings.TrimSpace(action), "analyze_report") {
		reportPrompt := fmt.Sprintf(`Create a concise markdown analysis report for the uploaded document.

Requirements:
- Use markdown headings and bullet points.
- Include sections: ## Overview, ## Key Findings, ## Risks or Gaps, ## Recommended Actions.
- If the document appears tabular, include a short ## Data Notes section.
- Keep it factual and easy to scan.

Document name: %s

Extracted content:
%s

Existing summary:
%s`, fileHeader.Filename, extractedText, aiResult)
		report, reportErr := h.aiService.GenerateChatResponse(reportPrompt)
		if reportErr != nil {
			log.Printf("[CHAT FILE] Analyze report generation error: %v", reportErr)
			return &models.ChatResponse{
				Response: "## Overview\n\nI extracted your file but could not generate a full report right now.\n\n## Key Findings\n\n- " + strings.TrimSpace(aiResult),
			}, nil
		}
		return &models.ChatResponse{Response: report}, nil
	}

	// Default when user didn't ask for anything specific: just return the summary
	if strings.TrimSpace(userMessage) == "" {
		return &models.ChatResponse{Response: aiResult}, nil
	}

	// Classify intent: FORM, RESEARCH, or SUMMARY
	intent, err := h.aiService.ClassifyDocumentIntent(userMessage, extractedText, aiResult)
	if err != nil {
		log.Printf("[CHAT FILE] Classify intent error: %v, defaulting to SUMMARY", err)
		intent = "SUMMARY"
	}

	switch intent {
	case "FORM":
		template, err := h.aiService.GenerateFormTemplateFromContent(aiResult+"\n\n"+extractedText, userMessage)
		if err != nil {
			log.Printf("[CHAT FILE] Generate form from content error: %v", err)
			return &models.ChatResponse{
				Response: "I extracted the content but couldn't generate a form from it. You can try: \"Create a form from this\" or describe the form you want.",
			}, nil
		}
		setPendingForm(userID, template)
		return &models.ChatResponse{
			Response:     "I've drafted a **sheet** from the document. **Review the columns below** and reply **Yes** or click **Save sheet**, or tell me what to change.",
			ProposedForm: &models.ProposedFormCard{FormTemplate: *template},
		}, nil
	case "RESEARCH":
		gatherPrompt := aiResult
		if userMessage != "" {
			gatherPrompt = userMessage + "\n\nContext from document: " + aiResult
		}
		content, err := h.Gather(gatherPrompt, 10)
		if err != nil {
			log.Printf("[CHAT FILE] Gathering error: %v", err)
			return &models.ChatResponse{
				Response: "I summarized the document but couldn't run the research (Gathering API). " + aiResult + "\n\nError: " + err.Error(),
			}, nil
		}
		return &models.ChatResponse{
			Response:        "Here’s a research summary based on the document and your request:",
			ResearchContent: content,
		}, nil
	default:
		return &models.ChatResponse{Response: aiResult}, nil
	}
}

// isFormConfirmMessage returns true if the user is confirming to save the proposed form.
func isFormConfirmMessage(message string) bool {
	s := strings.TrimSpace(strings.ToLower(message))
	if s == "" {
		return false
	}
	confirmPhrases := []string{"yes", "confirm", "save", "save form", "save sheet", "save it", "looks good", "ok", "okay", "correct", "submit"}
	for _, p := range confirmPhrases {
		if s == p || s == p+"." || strings.HasPrefix(s, p+" ") {
			return true
		}
	}
	return false
}

// savePendingFormAndClear saves the pending form template and clears state. Maps "general" to "student" for API.
func (h *Handlers) savePendingFormAndClear(c *gin.Context, userID string) (*models.ChatResponse, error) {
	template := getPendingForm(userID)
	if template == nil {
		return nil, nil
	}
	clearPendingForm(userID)

	userType := template.UserType
	if userType != "student" && userType != "staff" {
		userType = "student"
	}
	template.UserType = userType
	template.ID = uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	template.CreatedAt = now
	template.UpdatedAt = now
	createdBy := c.GetHeader("X-User-ID")
	if createdBy == "" {
		createdBy = "admin"
	}
	template.CreatedBy = createdBy

	if err := h.db.StoreFormTemplate(template); err != nil {
		log.Printf("[CHAT] Save proposed form error: %v", err)
		return &models.ChatResponse{
			Response: "Failed to save the form: " + err.Error(),
		}, nil
	}
	return &models.ChatResponse{
		Response: fmt.Sprintf("Form **%s** has been saved. You can use it under **Forms** and collect answers under **Form Answers**.", template.Name),
	}, nil
}
