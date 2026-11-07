package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

func writeSSE(w io.Writer, event, data string) {
	fmt.Fprintf(w, "event: %s\n", event)
	for _, line := range strings.Split(data, "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
}

func flushWriter(c *gin.Context) {
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
	}
}

func (h *Handlers) chatHandlerStream(c *gin.Context, userID, sessionID string, userVisible, llmMessage, agentInstructions string, skillIDs []string, memoryBlock string, subAgents []string, hasDocs bool) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	emit := func(reasoningDelta, contentDelta string) error {
		if reasoningDelta != "" {
			writeSSE(c.Writer, "reasoning", reasoningDelta)
			flushWriter(c)
		}
		if contentDelta != "" {
			writeSSE(c.Writer, "content", contentDelta)
			flushWriter(c)
		}
		return nil
	}

	chatResponse, reasoning, toolLog, genErr := h.chatWithManagementTools(
		c, userID, sessionID, llmMessage, agentInstructions, skillIDs, memoryBlock, emit,
	)
	if genErr != nil {
		writeSSE(c.Writer, "error", genErr.Error())
		flushWriter(c)
		return
	}

	response := models.ChatResponse{
		Response:  chatResponse,
		Reasoning: reasoning,
		SubAgents: subAgents,
	}
	persistChatExchange(h, userID, sessionID, userVisible, &response)
	if lessonUserID, ok := h.trustedLessonUserID(c); ok {
		h.maybeHarvestSession(lessonUserID, sessionID, userVisible, countUserTurns(h, lessonUserID, sessionID), len(toolLog), hasDocs)
	}
	payload, _ := json.Marshal(response)
	writeSSE(c.Writer, "done", string(payload))
	flushWriter(c)
}
