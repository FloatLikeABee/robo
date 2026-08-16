package chatsessions

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/academi/backend/internal/database"
	"github.com/academi/backend/internal/models"
)

const (
	maxSessions           = 35
	maxMessagesPerSession = 72
	previewMaxLen         = 120
	titleMaxLen           = 72
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func sessionKey(id string) []byte {
	return []byte("chat_session:" + id)
}

func (s *Service) Create(c *gin.Context) {
	now := time.Now().Unix()
	sess := models.ChatSession{
		ID:        uuid.New().String(),
		Title:     "New chat",
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  nil,
	}
	s.pruneIfNeeded()
	data, _ := json.Marshal(sess)
	if err := database.Set(sessionKey(sess.ID), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save session"})
		return
	}
	c.JSON(http.StatusCreated, sess)
}

func (s *Service) List(c *gin.Context) {
	var briefs []models.ChatSessionBrief
	database.Iterate([]byte("chat_session:"), func(key, value []byte) error {
		var sess models.ChatSession
		if err := json.Unmarshal(value, &sess); err != nil {
			return nil
		}
		briefs = append(briefs, models.ChatSessionBrief{
			ID:        sess.ID,
			Title:     sess.Title,
			Preview:   previewFromMessages(sess.Messages),
			UpdatedAt: sess.UpdatedAt,
		})
		return nil
	})
	sort.Slice(briefs, func(i, j int) bool {
		return briefs[i].UpdatedAt > briefs[j].UpdatedAt
	})
	c.JSON(http.StatusOK, briefs)
}

func previewFromMessages(msgs []models.ChatMessageRow) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		c := strings.TrimSpace(msgs[i].Content)
		if c == "" {
			continue
		}
		if len(c) > previewMaxLen {
			return c[:previewMaxLen] + "…"
		}
		return c
	}
	return ""
}

func (s *Service) Get(c *gin.Context) {
	id := c.Param("id")
	data, err := database.Get(sessionKey(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	var sess models.ChatSession
	if err := json.Unmarshal(data, &sess); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bad data"})
		return
	}
	c.JSON(http.StatusOK, sess)
}

func (s *Service) Patch(c *gin.Context) {
	id := c.Param("id")
	key := sessionKey(id)
	data, err := database.Get(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	var sess models.ChatSession
	if err := json.Unmarshal(data, &sess); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bad data"})
		return
	}

	var req models.PatchChatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msgs := trimMessages(req.Messages, maxMessagesPerSession)
	if strings.TrimSpace(req.Title) != "" {
		t := strings.TrimSpace(req.Title)
		if len(t) > titleMaxLen {
			t = t[:titleMaxLen] + "…"
		}
		sess.Title = t
	} else {
		sess.Title = deriveTitle(sess.Title, msgs)
	}
	sess.Messages = msgs
	sess.UpdatedAt = time.Now().Unix()
	s.pruneIfNeeded()

	out, _ := json.Marshal(sess)
	if err := database.Set(key, out); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save"})
		return
	}
	c.JSON(http.StatusOK, sess)
}

func deriveTitle(current string, msgs []models.ChatMessageRow) string {
	if current != "" && current != "New chat" {
		return current
	}
	for _, m := range msgs {
		if m.Role == "user" {
			c := strings.TrimSpace(m.Content)
			if c == "" {
				continue
			}
			if len(c) > titleMaxLen {
				return c[:titleMaxLen] + "…"
			}
			return c
		}
	}
	return "New chat"
}

func trimMessages(msgs []models.ChatMessageRow, max int) []models.ChatMessageRow {
	if max <= 0 || len(msgs) <= max {
		out := make([]models.ChatMessageRow, len(msgs))
		copy(out, msgs)
		return out
	}
	return append([]models.ChatMessageRow{}, msgs[len(msgs)-max:]...)
}

func (s *Service) pruneIfNeeded() {
	var sessions []models.ChatSession
	database.Iterate([]byte("chat_session:"), func(key, value []byte) error {
		var sess models.ChatSession
		if err := json.Unmarshal(value, &sess); err != nil {
			return nil
		}
		sessions = append(sessions, sess)
		return nil
	})
	if len(sessions) <= maxSessions {
		return
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt > sessions[j].UpdatedAt
	})
	for i := maxSessions; i < len(sessions); i++ {
		_ = database.Delete(sessionKey(sessions[i].ID))
	}
}

func (s *Service) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := database.Delete(sessionKey(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/chat-sessions")
	{
		g.POST("", s.Create)
		g.GET("", s.List)
		g.GET("/:id", s.Get)
		g.PATCH("/:id", s.Patch)
		g.DELETE("/:id", s.Delete)
	}
}
