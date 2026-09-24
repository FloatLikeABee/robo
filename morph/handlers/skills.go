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
	"github.com/robo/morphai"
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
		"builtin":       strings.HasPrefix(s.ID, "builtin-"),
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
	lessons := make([]gin.H, 0)
	if rows, err := h.TranMySQL.ListRecentAgentLessons(c.Request.Context(), 8); err == nil {
		for _, l := range rows {
			lessons = append(lessons, gin.H{
				"id":                l.ID,
				"trigger":           l.Trigger,
				"rule":              l.Rule,
				"source_session_id": l.SourceSessionID,
				"created_at":        l.CreatedAt,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"skills": out, "lessons": lessons, "total": len(out)})
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

const (
	builtinResearchID = "builtin-research"
	builtinDesignID   = "builtin-design"
	builtinGraphsID   = "builtin-graphs"
)

const builtinResearchInstructions = `You are Morph AI Research. Analyse the operator's goal, which materials apply (pinned files, notes, knowledge, graph hits), and unknowns. Then:
- For uploaded docs or library content, POST /api/graph/search first; use excerpts already in the prompt.
- Use allowed catalog tools and web search only after that when needed.
- Structure the answer by argument. Diagram first (mermaid) for structure or quantities; do not invent citations.`

const builtinDesignInstructions = `You are Morph AI Design. Analyse the goal, materials, and unknowns first. Name constraints, sketch 2–3 options, recommend one. Show options as a mermaid diagram when it clarifies the trade-off. Do not jump to implementation unless the operator asked to build it.`

const builtinGraphsInstructions = morphai.VisualFirstInstructions

type builtinSkillSeed struct {
	id, name, description, instructions string
}

func builtinSkillSeeds() []builtinSkillSeed {
	return []builtinSkillSeed{
		{
			id:           "builtin-concise-answers",
			name:         "Concise answers",
			description:  "Keep Morph AI replies short, actionable, and free of filler.",
			instructions: "Prefer brief answers. Lead with the result, then one short supporting sentence. Avoid long preambles.",
		},
		{
			id:           "builtin-morph-data-lookup",
			name:         "Morph Data lookup",
			description:  "When summarizing platform records, use /full routes and cover nested detail fields.",
			instructions: "For MorphData records, prefer GET .../:id/full when available. Include every meaningful field from top-level and nested detail JSON. Use readable labels, not raw JSON dumps.",
		},
		{
			id:           "builtin-knowledge-first",
			name:         "Knowledge first",
			description:  "Search the Morph Knowledge Library / graph before guessing about uploaded docs.",
			instructions: "When the question may relate to uploaded knowledge or platform docs, call POST /api/graph/search first and ground the answer in returned hits.",
		},
		{
			id:           builtinResearchID,
			name:         "Research",
			description:  "Search graph and docs first, then tools/web; structure by argument; do not invent citations.",
			instructions: builtinResearchInstructions,
		},
		{
			id:           builtinDesignID,
			name:         "Design",
			description:  "Name constraints, sketch options, recommend; do not jump to implementation.",
			instructions: builtinDesignInstructions,
		},
		{
			id:           builtinGraphsID,
			name:         "Graphs",
			description:  "When a reply can show structure or quantities, include a mermaid diagram or chart.",
			instructions: builtinGraphsInstructions,
		},
	}
}

func defaultAgentSkillBody(id string) (name, instructions string) {
	for _, b := range builtinSkillSeeds() {
		if b.id == id {
			return b.name, b.instructions
		}
	}
	return "", ""
}

// SeedBuiltinSkills inserts missing default skills by id (existing operator skills stay).
func (h *Handlers) SeedBuiltinSkills() {
	if h == nil || h.TranMySQL == nil {
		return
	}
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)
	for _, b := range builtinSkillSeeds() {
		h.ensureBuiltinSkill(ctx, b, now)
	}
}

func (h *Handlers) ensureBuiltinSkill(ctx context.Context, b builtinSkillSeed, now string) {
	existing, err := h.TranMySQL.GetAISkill(ctx, b.id)
	if err != nil || existing != nil {
		return
	}
	s := &db.AISkill{
		ID: b.id, Name: b.name, Description: b.description,
		Enabled: true, OwnerUserID: "system", CreatedAt: now, UpdatedAt: now,
	}
	if err := h.TranMySQL.InsertAISkill(ctx, s); err != nil {
		return
	}
	if h.db != nil {
		_ = h.db.PutAISkillBody(s.ID, db.AISkillBody{Instructions: b.instructions})
	}
	_ = h.TranMySQL.EnqueueNeo4jIngest(ctx, db.Neo4jKindSkill, s.ID, `{"op":"upsert"}`)
}

func (h *Handlers) skillBodyInstructions(id string) string {
	_, fallback := defaultAgentSkillBody(id)
	if h != nil && h.db != nil {
		body, err := h.db.GetAISkillBody(id)
		if err == nil && strings.TrimSpace(body.Instructions) != "" {
			return strings.TrimSpace(body.Instructions)
		}
	}
	return fallback
}

func (h *Handlers) appendAlwaysOnAgentSkills(b *strings.Builder) {
	b.WriteString("\n--- Default agent skills ---\n")
	for _, id := range []string{builtinResearchID, builtinDesignID, builtinGraphsID} {
		name, _ := defaultAgentSkillBody(id)
		instr := h.skillBodyInstructions(id)
		if name == "" || instr == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("### %s\n%s\n\n", name, instr))
	}
}

// buildEnabledSkillsContext appends enabled skill names/descriptions for the assistant system prompt.
// Research and Design instruction bodies are always included. Picker skill_ids add extra bodies.
func (h *Handlers) buildEnabledSkillsContext(skillIDs []string) string {
	var b strings.Builder
	if h != nil && h.TranMySQL != nil {
		ctx := context.Background()
		list, err := h.TranMySQL.ListAISkills(ctx, true)
		if err == nil && len(list) > 0 {
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
		}
	}
	h.appendAlwaysOnAgentSkills(&b)
	return strings.TrimSpace(b.String())
}

func (h *Handlers) agentSkillsAndLessonsContext(skillIDs []string) string {
	a := h.buildEnabledSkillsContext(skillIDs)
	b := h.buildAgentLessonsContext()
	return strings.TrimSpace(strings.TrimSpace(a) + "\n\n" + b)
}
