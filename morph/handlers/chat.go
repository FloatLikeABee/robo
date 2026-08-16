package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"idongivaflyinfa/models"
	"idongivaflyinfa/validation"

	"github.com/gin-gonic/gin"
)

// ChatHandler handles chat requests (sheets/forms, registration flows, general assistant).
// @Summary      Chat with Morph AI
// @Description  Send a message; AI may help with sheets, registration, or general questions.
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        request  body      models.ChatRequest  true  "Chat request with message"
// @Header       200      {string}  X-User-ID          "Optional user ID for chat history"
// @Success      200      {object}  models.ChatResponse
// @Failure      400      {object}  map[string]string   "Invalid request"
// @Failure      500      {object}  map[string]string   "Internal server error"
// @Router       /api/chat [post]
func (h *Handlers) ChatHandler(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "admin"
	}

	var req models.ChatRequest
	contentType := c.GetHeader("Content-Type")
	isMultipart := strings.Contains(contentType, "multipart/form-data")
	if isMultipart {
		message := c.PostForm("message")
		req.SessionID = c.PostForm("session_id")
		req.AgentID = c.PostForm("agent_id")
		action := c.PostForm("action")
		file, err := c.FormFile("file")
		if err == nil && file != nil {
			response, err := h.handleChatWithFile(c, userID, message, file, action)
			if err != nil {
				log.Printf("[CHAT HANDLER] File flow error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process file: %v", err)})
				return
			}
			sessionID := resolveSessionID(req.SessionID)
			_ = h.db.EnsureDefaultChatSession(userID)
			persistChatExchange(h, userID, sessionID, message, response)
			c.JSON(http.StatusOK, response)
			return
		}
		req.Message = message
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
	}

	sessionID := resolveSessionID(req.SessionID)
	_ = h.db.EnsureDefaultChatSession(userID)

	agentID := strings.TrimSpace(req.AgentID)
	agentInstructions := ""
	bkAssistantID := ""
	if agentID != "" && !strings.EqualFold(agentID, "general") {
		if bkID, ok := parseBKAssistantID(agentID); ok {
			bkAssistantID = bkID
		} else {
			agent, err := h.db.GetMorphAIAgent(agentID)
			if err != nil || agent == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown assistant"})
				return
			}
			if !agent.Enabled {
				c.JSON(http.StatusBadRequest, gin.H{"error": "That assistant is disabled"})
				return
			}
			if strings.EqualFold(agent.ID, morphAIImageGeneratorAgentID) {
				pendingConfirm := getPendingForm(userID) != nil && isFormConfirmMessage(req.Message)
				if !pendingConfirm {
					correctedMessage, corrErr := h.aiService.CorrectSpelling(req.Message)
					if corrErr != nil {
						correctedMessage = req.Message
					} else if correctedMessage != req.Message {
						req.Message = correctedMessage
					}
					if !validation.IsValidPrompt(req.Message) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "The request appears to be invalid or gibberish. Please provide a meaningful message."})
						return
					}
					imgResp, imgErr := h.handleImageGeneratorChat(c, req.Message)
					if imgErr != nil {
						log.Printf("[CHAT HANDLER] Image generator error: %v", imgErr)
						c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate image: %v", imgErr)})
						return
					}
					persistChatExchange(h, userID, sessionID, req.Message, imgResp)
					c.JSON(http.StatusOK, imgResp)
					return
				}
			}
			agentInstructions = strings.TrimSpace(agent.Instructions)
		}
	}

	if pending := getPendingForm(userID); pending != nil && isFormConfirmMessage(req.Message) {
		response, err := h.savePendingFormAndClear(c, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if response != nil {
			persistChatExchange(h, userID, sessionID, req.Message, response)
			c.JSON(http.StatusOK, response)
			return
		}
	}

	correctedMessage, err := h.aiService.CorrectSpelling(req.Message)
	if err != nil {
		log.Printf("[CHAT HANDLER] Error correcting spelling: %v, using original message", err)
		correctedMessage = req.Message
	} else if correctedMessage != req.Message {
		log.Printf("[CHAT HANDLER] Spelling corrected: '%s' -> '%s'", req.Message, correctedMessage)
		req.Message = correctedMessage
	}

	log.Printf("[CHAT HANDLER] User: %s, Message: %s", userID, req.Message)

	// Sheet/form templates must win over registration and general chat (ticket: sheet requests became normal chat).
	if isSheetOrFormGenerationRequest(req.Message) {
		template, err := h.aiService.GenerateSheetTemplateFromUserPrompt(req.Message)
		if err != nil {
			log.Printf("Error generating sheet template: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate sheet: %v", err)})
			return
		}
		setPendingForm(userID, template)
		response := &models.ChatResponse{
			Response: "I've drafted a **sheet** from your request. Review the columns below, then click **Save sheet** to add it (or reply **yes** to save). Tell me if you want changes.",
			ProposedForm: &models.ProposedFormCard{
				FormTemplate: *template,
			},
		}
		persistChatExchange(h, userID, sessionID, req.Message, response)
		c.JSON(http.StatusOK, response)
		return
	}

	regState, regErr := h.db.GetRegistrationStateByUserID(userID)
	if regErr == nil && regState != nil && regState.Step != "complete" && regState.Step != "" {
		log.Printf("[CHAT HANDLER] User %s has active registration session (form: %s)", userID, regState.FormName)
		response, err := h.handleRegistrationFlow(c, userID, req.Message)
		if err != nil {
			log.Printf("[CHAT HANDLER] Error in registration flow: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process registration: %v", err)})
			return
		}
		if response != nil {
			persistChatExchange(h, userID, sessionID, req.Message, response)
			c.JSON(http.StatusOK, response)
			return
		}
	}

	if isRegisterStudentRequest(req.Message) {
		log.Printf("[CHAT HANDLER] Detected register-student (or similar) request from user %s", userID)
		response, err := h.handleRegistrationFlow(c, userID, req.Message)
		if err != nil {
			log.Printf("[CHAT HANDLER] Error in registration flow: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to process registration: %v", err)})
			return
		}
		if response != nil {
			persistChatExchange(h, userID, sessionID, req.Message, response)
			c.JSON(http.StatusOK, response)
			return
		}
	}

	if !validation.IsValidPrompt(req.Message) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The request appears to be invalid or gibberish. Please provide a meaningful message."})
		return
	}

	if bkAssistantID != "" {
		_, instructions, bkErr := h.buildBKAssistantInstructions(c.Request.Context(), bkAssistantID, req.Message)
		if bkErr != nil {
			log.Printf("[CHAT HANDLER] BK assistant %s: %v", bkAssistantID, bkErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown or unavailable AI tools assistant"})
			return
		}
		agentInstructions = instructions
	}

	var chatResponse string
	var genErr error
	llmMessage := h.hybridAugmentMessage(userID, sessionID, req.Message)
	if h.ginEngine != nil {
		chatResponse, genErr = h.chatWithManagementTools(c, userID, sessionID, llmMessage, agentInstructions, req.SkillIDs)
	} else {
		skillsCtx := h.buildEnabledSkillsContext(req.SkillIDs)
		agentExtra := agentInstructions
		if skillsCtx != "" {
			if agentExtra != "" {
				agentExtra = agentExtra + "\n\n" + skillsCtx
			} else {
				agentExtra = skillsCtx
			}
		}
		chatResponse, genErr = h.aiService.GenerateChatResponseWithAgent(llmMessage, agentExtra)
	}
	if genErr != nil {
		log.Printf("Error generating chat response: %v", genErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate response: %v", genErr)})
		return
	}

	response := models.ChatResponse{
		Response: chatResponse,
		SQL:      "",
	}
	persistChatExchange(h, userID, sessionID, req.Message, &response)
	c.JSON(http.StatusOK, response)
}

// isSheetOrFormGenerationRequest is true when the user wants a new sheet or form template (not free-form chat).
func isSheetOrFormGenerationRequest(prompt string) bool {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	sheetWord := strings.Contains(lower, "sheet") || strings.Contains(lower, "spreadsheet")
	if sheetWord {
		// Action / intent phrases
		if strings.Contains(lower, "create") || strings.Contains(lower, "add") || strings.Contains(lower, "new") ||
			strings.Contains(lower, "generate") || strings.Contains(lower, "make") || strings.Contains(lower, "build") ||
			strings.Contains(lower, "draft") || strings.Contains(lower, "need a") || strings.Contains(lower, "want a") ||
			strings.Contains(lower, "give me") || strings.Contains(lower, "set up") || strings.Contains(lower, "design") ||
			strings.Contains(lower, "prepare") || strings.Contains(lower, "start a") {
			return true
		}
		// Common wording: "sheet to track ...", "sheet for attendance", "create a sheet", "a sheet for ..."
		if strings.Contains(lower, "sheet to ") || strings.Contains(lower, "sheet for ") ||
			strings.Contains(lower, " a sheet") || strings.Contains(lower, " the sheet ") {
			return true
		}
	}
	if (strings.Contains(lower, "create") && strings.Contains(lower, "form")) ||
		strings.Contains(lower, "i want a new form") ||
		strings.Contains(lower, "generate a form") ||
		strings.Contains(lower, "make a form") ||
		strings.Contains(lower, "build a form") ||
		(strings.Contains(lower, "form") && (strings.Contains(lower, "new") || strings.Contains(lower, "create"))) {
		return true
	}
	return false
}

func resolveSessionID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return models.DefaultChatSessionID
	}
	return s
}

func persistChatExchange(h *Handlers, userID, sessionID string, userMessage string, resp *models.ChatResponse) {
	if resp == nil {
		return
	}
	userMsg := &models.StoredChatMessage{Role: "user", Content: userMessage}
	if err := h.db.AppendChatMessage(userID, sessionID, userMsg); err != nil {
		log.Printf("[CHAT] Failed to append user message to session: %v", err)
		return
	}
	assistantMsg := &models.StoredChatMessage{
		Role:             "assistant",
		Content:          resp.Response,
		SQL:              resp.SQL,
		ConfirmationCard: resp.ConfirmationCard,
		ProposedForm:     resp.ProposedForm,
		ResearchContent:  resp.ResearchContent,
		Images:           resp.Images,
	}
	if err := h.db.AppendChatMessage(userID, sessionID, assistantMsg); err != nil {
		log.Printf("[CHAT] Failed to append assistant message to session: %v", err)
	}
	h.maybeAutoTitleSession(userID, sessionID, userMessage)
}

// maybeAutoTitleSession sets an AI-generated title after the first exchange when the session is still named "New chat".
func (h *Handlers) maybeAutoTitleSession(userID, sessionID, userMessage string) {
	if sessionID == models.DefaultChatSessionID {
		return
	}
	sess, err := h.db.GetChatSession(userID, sessionID)
	if err != nil || sess == nil {
		return
	}
	if !strings.EqualFold(strings.TrimSpace(sess.Title), "New chat") {
		return
	}
	msgs, err := h.db.GetChatSessionMessages(userID, sessionID)
	if err != nil || len(msgs) != 2 {
		return
	}
	src := strings.TrimSpace(userMessage)
	if src == "" {
		src = "File or image upload"
	}
	title, err := h.aiService.GenerateSessionTitle(src)
	if err != nil {
		log.Printf("[CHAT] GenerateSessionTitle: %v", err)
		return
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return
	}
	if err := h.db.UpdateChatSessionTitle(userID, sessionID, title); err != nil {
		log.Printf("[CHAT] UpdateChatSessionTitle: %v", err)
	}
}
