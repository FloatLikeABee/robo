package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type skillCreateBody struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Instructions string                 `json:"instructions"`
	Body         string                 `json:"body"`
	Enabled      *bool                  `json:"enabled"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type skillPatchBody struct {
	Name         *string                `json:"name"`
	Description  *string                `json:"description"`
	Instructions *string                `json:"instructions"`
	Body         *string                `json:"body"`
	Enabled      *bool                  `json:"enabled"`
	Metadata     map[string]interface{} `json:"metadata"`
}

func skillInstructionsFrom(create skillCreateBody) string {
	s := strings.TrimSpace(create.Instructions)
	if s == "" {
		s = strings.TrimSpace(create.Body)
	}
	return s
}

func (h *Handlers) skillJSON(c *gin.Context, s *db.AISkill, includeBody bool) gin.H {
	out := gin.H{
		"id":            s.ID,
		"name":          s.Name,
		"description":   s.Description,
		"enabled":       s.Enabled,
		"owner_user_id": s.OwnerUserID,
		"created_at":    s.CreatedAt,
		"updated_at":    s.UpdatedAt,
	}
	if includeBody && h.db != nil {
		body, err := h.db.GetAISkillBody(s.ID)
		if err == nil {
			out["instructions"] = body.Instructions
			if body.Metadata != nil {
				out["metadata"] = body.Metadata
			}
		}
	}
	return out
}

// ListSkills GET /api/skills
func (h *Handlers) ListSkills(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sqlite not available"})
		return
	}
	enabledOnly := strings.EqualFold(c.Query("enabled"), "true") || c.Query("enabled") == "1"
	list, err := h.TranMySQL.ListAISkills(c.Request.Context(), enabledOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(list))
	for i := range list {
		out = append(out, h.skillJSON(c, &list[i], false))
	}
	c.JSON(http.StatusOK, gin.H{"skills": out, "total": len(out)})
}

// GetSkill GET /api/skills/:id
func (h *Handlers) GetSkill(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sqlite not available"})
		return
	}
	s, err := h.TranMySQL.GetAISkill(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}
	c.JSON(http.StatusOK, h.skillJSON(c, s, true))
}

// CreateSkill POST /api/skills
func (h *Handlers) CreateSkill(c *gin.Context) {
	if h.TranMySQL == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "skills store not available"})
		return
	}
	var body skillCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	name := strings.TrimSpace(body.Name)
	instructions := skillInstructionsFrom(body)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if instructions == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body/instructions is required"})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	owner := strings.TrimSpace(c.GetHeader("X-User-ID"))
	s := &db.AISkill{
		ID:          uuid.NewString(),
		Name:        name,
		Description: strings.TrimSpace(body.Description),
		Enabled:     enabled,
		OwnerUserID: owner,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.TranMySQL.InsertAISkill(c.Request.Context(), s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.PutAISkillBody(s.ID, db.AISkillBody{
		Instructions: instructions,
		Metadata:     body.Metadata,
	}); err != nil {
		_ = h.TranMySQL.DeleteAISkill(c.Request.Context(), s.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.TranMySQL.EnqueueNeo4jIngest(c.Request.Context(), db.Neo4jKindSkill, s.ID, `{"op":"upsert"}`)
	c.JSON(http.StatusCreated, h.skillJSON(c, s, true))
}

// PatchSkill PATCH /api/skills/:id
func (h *Handlers) PatchSkill(c *gin.Context) {
	if h.TranMySQL == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "skills store not available"})
		return
	}
	id := c.Param("id")
	s, err := h.TranMySQL.GetAISkill(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}
	var body skillPatchBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if body.Name != nil {
		name := strings.TrimSpace(*body.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		s.Name = name
	}
	if body.Description != nil {
		s.Description = strings.TrimSpace(*body.Description)
	}
	if body.Enabled != nil {
		s.Enabled = *body.Enabled
	}
	existingBody, _ := h.db.GetAISkillBody(s.ID)
	instructions := existingBody.Instructions
	if body.Instructions != nil {
		instructions = strings.TrimSpace(*body.Instructions)
	} else if body.Body != nil {
		instructions = strings.TrimSpace(*body.Body)
	}
	if body.Instructions != nil || body.Body != nil {
		if instructions == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "body/instructions is required"})
			return
		}
	}
	meta := existingBody.Metadata
	if body.Metadata != nil {
		meta = body.Metadata
	}
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := h.TranMySQL.UpdateAISkill(c.Request.Context(), s); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.PutAISkillBody(s.ID, db.AISkillBody{Instructions: instructions, Metadata: meta}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.TranMySQL.EnqueueNeo4jIngest(c.Request.Context(), db.Neo4jKindSkill, s.ID, `{"op":"upsert"}`)
	c.JSON(http.StatusOK, h.skillJSON(c, s, true))
}

// DeleteSkill DELETE /api/skills/:id
func (h *Handlers) DeleteSkill(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sqlite not available"})
		return
	}
	id := c.Param("id")
	if err := h.TranMySQL.DeleteAISkill(c.Request.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.db != nil {
		_ = h.db.DeleteAISkillBody(id)
	}
	_ = h.TranMySQL.EnqueueNeo4jIngest(c.Request.Context(), db.Neo4jKindSkill, id, `{"op":"delete"}`)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// SeedBuiltinSkills inserts default skills when the store is empty.
func (h *Handlers) SeedBuiltinSkills() {
	if h.TranMySQL == nil || h.db == nil {
		return
	}
	ctx := context.Background()
	n, err := h.TranMySQL.CountAISkills(ctx)
	if err != nil || n > 0 {
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	builtins := []struct {
		id, name, description, instructions string
	}{
		{
			id:          "builtin-concise-answers",
			name:        "Concise answers",
			description: "Keep Morph AI replies short, actionable, and free of filler.",
			instructions: "Prefer brief answers. Lead with the result, then one short supporting sentence. Avoid long preambles.",
		},
		{
			id:          "builtin-morph-data-lookup",
			name:        "Morph Data lookup",
			description: "When summarizing platform records, use /full routes and cover nested detail fields.",
			instructions: "For MorphData records, prefer GET .../:id/full when available. Include every meaningful field from top-level and nested detail JSON. Use readable labels, not raw JSON dumps.",
		},
		{
			id:          "builtin-knowledge-first",
			name:        "Knowledge first",
			description: "Search the Morph Knowledge Library / graph before guessing about uploaded docs.",
			instructions: "When the question may relate to uploaded knowledge or platform docs, call POST /api/graph/search first and ground the answer in returned hits.",
		},
	}
	for _, b := range builtins {
		s := &db.AISkill{
			ID: b.id, Name: b.name, Description: b.description,
			Enabled: true, OwnerUserID: "system", CreatedAt: now, UpdatedAt: now,
		}
		if err := h.TranMySQL.InsertAISkill(ctx, s); err != nil {
			continue
		}
		_ = h.db.PutAISkillBody(s.ID, db.AISkillBody{Instructions: b.instructions})
		_ = h.TranMySQL.EnqueueNeo4jIngest(ctx, db.Neo4jKindSkill, s.ID, `{"op":"upsert"}`)
	}
}

// buildEnabledSkillsContext appends enabled skill names/descriptions for the assistant system prompt.
// When skillIDs is non-empty, also loads those skills' instruction bodies (if enabled).
func (h *Handlers) buildEnabledSkillsContext(skillIDs []string) string {
	if h == nil || h.TranMySQL == nil {
		return ""
	}
	ctx := context.Background()
	list, err := h.TranMySQL.ListAISkills(ctx, true)
	if err != nil || len(list) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("--- Enabled AI skills (catalog) ---\n")
	b.WriteString("Use these skills when relevant. Catalog:\n")
	for _, s := range list {
		b.WriteString(fmt.Sprintf("- [%s] %s: %s\n", s.ID, s.Name, s.Description))
	}
	if len(skillIDs) > 0 && h.db != nil {
		want := map[string]struct{}{}
		for _, id := range skillIDs {
			id = strings.TrimSpace(id)
			if id != "" {
				want[id] = struct{}{}
			}
		}
		b.WriteString("\n--- Selected skill instructions ---\n")
		for _, s := range list {
			if _, ok := want[s.ID]; !ok {
				continue
			}
			body, err := h.db.GetAISkillBody(s.ID)
			if err != nil || strings.TrimSpace(body.Instructions) == "" {
				continue
			}
			b.WriteString(fmt.Sprintf("### %s\n%s\n\n", s.Name, strings.TrimSpace(body.Instructions)))
		}
	}
	return strings.TrimSpace(b.String())
}
