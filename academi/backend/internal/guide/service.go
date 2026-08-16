package guide

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/academi/backend/internal/auth"
	"github.com/academi/backend/internal/database"
	"github.com/academi/backend/internal/models"
)

type Service struct {
	seedOnce sync.Once
}

func NewService() *Service {
	return &Service{}
}

func subjectKey(id string) []byte {
	return []byte("guide_subject:" + id)
}

func normalizeGuideSteps(steps []models.GuideStep) []models.GuideStep {
	out := make([]models.GuideStep, 0, len(steps))
	id := 1
	for _, st := range steps {
		t := strings.TrimSpace(st.Title)
		body := strings.TrimSpace(st.Content)
		if t == "" && body == "" {
			continue
		}
		out = append(out, models.GuideStep{ID: id, Title: t, Content: body})
		id++
	}
	return out
}

func maxSubjectSortOrder() int {
	max := 0
	_ = database.Iterate([]byte("guide_subject:"), func(key, value []byte) error {
		var sub models.GuideSubject
		if json.Unmarshal(value, &sub) == nil && sub.SortOrder > max {
			max = sub.SortOrder
		}
		return nil
	})
	return max
}

func countGuidesForSubject(subjectID string) int {
	n := 0
	_ = database.Iterate([]byte("guide:"), func(key, value []byte) error {
		var g models.Guide
		if json.Unmarshal(value, &g) != nil {
			return nil
		}
		if g.SubjectID == subjectID {
			n++
		}
		return nil
	})
	return n
}

func loadSubject(id string) (*models.GuideSubject, error) {
	data, err := database.Get(subjectKey(id))
	if err != nil {
		return nil, err
	}
	var sub models.GuideSubject
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *Service) ensureSeeded() {
	s.seedOnce.Do(func() {
		n := 0
		_ = database.Iterate([]byte("guide_subject:"), func(key, value []byte) error {
			n++
			return nil
		})
		if n > 0 {
			return
		}
		now := time.Now().Unix()
		for _, sub := range mockSubjects(now) {
			b, _ := json.Marshal(sub)
			_ = database.Set(subjectKey(sub.ID), b)
		}
		for _, g := range mockGuides(now) {
			b, _ := json.Marshal(g)
			_ = database.Set([]byte("guide:"+g.ID), b)
		}
	})
}

func mockSubjects(now int64) []models.GuideSubject {
	return []models.GuideSubject{
		{ID: "subj-math", Name: "Mathematics", Description: "Algebra through calculus — problem sets and mock learning paths.", Icon: "📐", SortOrder: 10, CreatedAt: now},
		{ID: "subj-cs", Name: "Computer Science", Description: "Programming, data structures, and how systems fit together.", Icon: "💻", SortOrder: 20, CreatedAt: now},
		{ID: "subj-bio", Name: "Biology", Description: "Cells, genetics, evolution, and ecology (mock catalog).", Icon: "🧬", SortOrder: 30, CreatedAt: now},
		{ID: "subj-hist", Name: "History", Description: "Timelines, sources, and how to read evidence.", Icon: "📜", SortOrder: 40, CreatedAt: now},
		{ID: "subj-write", Name: "Writing", Description: "Structure, clarity, and revising drafts.", Icon: "✍️", SortOrder: 50, CreatedAt: now},
	}
}

func mockGuides(now int64) []models.Guide {
	return []models.Guide{
		{
			ID: "guide-math-calc", SubjectID: "subj-math", Title: "Calculus in one week",
			Description: "A compact mock path: limits, derivatives, and integrals.",
			Category: "Calculus", Icon: "📈", Color: "#5b8cff",
			Steps: []models.GuideStep{
				{ID: 1, Title: "Limits", Content: "Build intuition with graphs and tables before ε–δ."},
				{ID: 2, Title: "Derivatives", Content: "Power rule, product/quotient, and chain rule drills."},
				{ID: 3, Title: "Integrals", Content: "Antiderivatives and interpreting area under curves."},
			},
			CreatedAt: now,
		},
		{
			ID: "guide-math-linalg", SubjectID: "subj-math", Title: "Linear algebra refresh",
			Description: "Vectors, matrices, and eigenvalues — mock overview.",
			Category: "Linear algebra", Icon: "⊡", Color: "#7c6cff",
			Steps: []models.GuideStep{
				{ID: 1, Title: "Vectors & spaces", Content: "Span, basis, and orthogonality in plain language."},
				{ID: 2, Title: "Matrix ops", Content: "Multiplication as composing transformations."},
				{ID: 3, Title: "Eigenstuff", Content: "Why eigenvectors show stable directions."},
			},
			CreatedAt: now,
		},
		{
			ID: "guide-cs-py", SubjectID: "subj-cs", Title: "Python patterns",
			Description: "Readable scripts, modules, and error handling (mock).",
			Category: "Programming", Icon: "🐍", Color: "#3ecf8e",
			Steps: []models.GuideStep{
				{ID: 1, Title: "Project layout", Content: "Packages, `if __name__ == \"__main__\"`, and virtual envs."},
				{ID: 2, Title: "Types & tests", Content: "Gradual typing and a few quick unit tests."},
			},
			CreatedAt: now,
		},
		{
			ID: "guide-cs-git", SubjectID: "subj-cs", Title: "Git for study projects",
			Description: "Branching, commits, and clean history basics.",
			Category: "Tools", Icon: "⎇", Color: "#f0766b",
			Steps: []models.GuideStep{
				{ID: 1, Title: "Commits", Content: "Small, focused commits with clear messages."},
				{ID: 2, Title: "Branches", Content: "Feature branches and merge vs rebase tradeoffs."},
				{ID: 3, Title: "Remotes", Content: "Push, pull, and resolve simple conflicts."},
			},
			CreatedAt: now,
		},
		{
			ID: "guide-bio-cell", SubjectID: "subj-bio", Title: "Cell biology tour",
			Description: "Organelles and how traffic moves inside cells.",
			Category: "Cells", Icon: "🔬", Color: "#5b8cff",
			Steps: []models.GuideStep{
				{ID: 1, Title: "Membrane", Content: "Transport channels and energy cost."},
				{ID: 2, Title: "Energy", Content: "Mitochondria, ATP, and metabolic themes."},
			},
			CreatedAt: now,
		},
		{
			ID: "guide-hist-sources", SubjectID: "subj-hist", Title: "Working with sources",
			Description: "Bias, corroboration, and building an argument.",
			Category: "Methods", Icon: "🕯️", Color: "#c9a45c",
			Steps: []models.GuideStep{
				{ID: 1, Title: "Context", Content: "When/where/who produced the source."},
				{ID: 2, Title: "Compare", Content: "Pair primary accounts and spot gaps."},
			},
			CreatedAt: now,
		},
		{
			ID: "guide-write-essay", SubjectID: "subj-write", Title: "Short essay workflow",
			Description: "Thesis, outline, draft, revise (mock checklist).",
			Category: "Exposition", Icon: "📝", Color: "#8ab4ff",
			Steps: []models.GuideStep{
				{ID: 1, Title: "Claim", Content: "One arguable sentence the whole essay supports."},
				{ID: 2, Title: "Evidence", Content: "Quote or paraphrase with citations."},
				{ID: 3, Title: "Revision", Content: "Cut fluff, tighten topic sentences."},
			},
			CreatedAt: now,
		},
	}
}

// ListSubjects returns persisted subjects (bootstraps mock catalog once if empty).
func (s *Service) ListSubjects(c *gin.Context) {
	s.ensureSeeded()
	var subs []models.GuideSubject
	_ = database.Iterate([]byte("guide_subject:"), func(key, value []byte) error {
		var sub models.GuideSubject
		if json.Unmarshal(value, &sub) != nil {
			return nil
		}
		subs = append(subs, sub)
		return nil
	})
	sort.Slice(subs, func(i, j int) bool {
		if subs[i].SortOrder != subs[j].SortOrder {
			return subs[i].SortOrder < subs[j].SortOrder
		}
		return subs[i].Name < subs[j].Name
	})
	c.JSON(http.StatusOK, subs)
}

func (s *Service) CreateGuide(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req models.CreateGuideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	steps := normalizeGuideSteps(req.Steps)
	if len(steps) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one step is required"})
		return
	}

	guide := models.Guide{
		ID:          uuid.New().String(),
		SubjectID:   strings.TrimSpace(req.SubjectID),
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Category:    strings.TrimSpace(req.Category),
		Icon:        strings.TrimSpace(req.Icon),
		Color:       strings.TrimSpace(req.Color),
		Steps:       steps,
		CreatedAt:   time.Now().Unix(),
		CreatedBy:   userID,
	}

	data, _ := json.Marshal(guide)
	if err := database.Set([]byte("guide:"+guide.ID), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create guide"})
		return
	}

	c.JSON(http.StatusCreated, guide)
}

func (s *Service) UpdateGuide(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	guideID := c.Param("id")
	data, err := database.Get([]byte("guide:" + guideID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guide not found"})
		return
	}
	var existing models.Guide
	if json.Unmarshal(data, &existing) != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guide not found"})
		return
	}
	if existing.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit guides you created"})
		return
	}

	var req models.UpdateGuideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	steps := normalizeGuideSteps(req.Steps)
	if len(steps) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one step is required"})
		return
	}

	existing.SubjectID = strings.TrimSpace(req.SubjectID)
	existing.Title = strings.TrimSpace(req.Title)
	existing.Description = strings.TrimSpace(req.Description)
	existing.Category = strings.TrimSpace(req.Category)
	existing.Icon = strings.TrimSpace(req.Icon)
	existing.Color = strings.TrimSpace(req.Color)
	existing.Steps = steps

	out, _ := json.Marshal(existing)
	if err := database.Set([]byte("guide:"+guideID), out); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update guide"})
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (s *Service) CreateSubject(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req models.CreateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	s.ensureSeeded()
	sub := models.GuideSubject{
		ID:          uuid.New().String(),
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Icon:        strings.TrimSpace(req.Icon),
		SortOrder:   maxSubjectSortOrder() + 10,
		CreatedAt:   time.Now().Unix(),
		CreatedBy:   userID,
	}
	raw, _ := json.Marshal(sub)
	if err := database.Set(subjectKey(sub.ID), raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subject"})
		return
	}
	c.JSON(http.StatusCreated, sub)
}

func (s *Service) UpdateSubject(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	subjectID := c.Param("subjectId")
	sub, err := loadSubject(subjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subject not found"})
		return
	}
	if sub.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit subjects you created"})
		return
	}
	var req models.UpdateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	sub.Name = name
	sub.Description = strings.TrimSpace(req.Description)
	sub.Icon = strings.TrimSpace(req.Icon)
	raw, _ := json.Marshal(sub)
	if err := database.Set(subjectKey(sub.ID), raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subject"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

func (s *Service) DeleteSubject(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	subjectID := c.Param("subjectId")
	sub, err := loadSubject(subjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subject not found"})
		return
	}
	if sub.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete subjects you created"})
		return
	}
	if countGuidesForSubject(subjectID) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Remove or reassign guides under this subject before deleting it"})
		return
	}
	if err := database.Delete(subjectKey(subjectID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subject"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Subject deleted"})
}

func (s *Service) GetGuide(c *gin.Context) {
	s.ensureSeeded()
	guideID := c.Param("id")

	data, err := database.Get([]byte("guide:" + guideID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guide not found"})
		return
	}

	var g models.Guide
	json.Unmarshal(data, &g)

	c.JSON(http.StatusOK, g)
}

// ListGuides lists guides; optional query subject_id filters to one subject.
func (s *Service) ListGuides(c *gin.Context) {
	s.ensureSeeded()
	subjectFilter := c.Query("subject_id")
	var guides []models.Guide
	_ = database.Iterate([]byte("guide:"), func(key, value []byte) error {
		var g models.Guide
		if err := json.Unmarshal(value, &g); err != nil {
			return nil
		}
		if subjectFilter != "" && g.SubjectID != subjectFilter {
			return nil
		}
		guides = append(guides, g)
		return nil
	})
	sort.Slice(guides, func(i, j int) bool {
		return guides[i].Title < guides[j].Title
	})
	c.JSON(http.StatusOK, guides)
}

func (s *Service) DeleteGuide(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	guideID := c.Param("id")
	data, err := database.Get([]byte("guide:" + guideID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guide not found"})
		return
	}
	var g models.Guide
	if json.Unmarshal(data, &g) != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guide not found"})
		return
	}
	if g.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete guides you created"})
		return
	}
	if err := database.Delete([]byte("guide:" + guideID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Guide deleted"})
}

func (s *Service) GetProgress(c *gin.Context) {
	userID := c.Param("userId")
	guideID := c.Param("id")

	data, err := database.Get([]byte("progress:" + userID + ":" + guideID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Progress not found"})
		return
	}

	var progress models.UserGuideProgress
	json.Unmarshal(data, &progress)

	c.JSON(http.StatusOK, progress)
}

func (s *Service) UpdateProgress(c *gin.Context) {
	userID := c.Param("userId")
	guideID := c.Param("id")

	var req struct {
		StepID int  `json:"step_id" binding:"required"`
		Done   bool `json:"done"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var progress models.UserGuideProgress
	data, err := database.Get([]byte("progress:" + userID + ":" + guideID))
	if err != nil {
		progress = models.UserGuideProgress{
			UserID:         userID,
			GuideID:        guideID,
			CompletedSteps: []int{},
			StartedAt:      time.Now().Unix(),
		}
	} else {
		json.Unmarshal(data, &progress)
	}

	found := false
	for i, step := range progress.CompletedSteps {
		if step == req.StepID {
			if !req.Done {
				progress.CompletedSteps = append(progress.CompletedSteps[:i], progress.CompletedSteps[i+1:]...)
			}
			found = true
			break
		}
	}

	if !found && req.Done {
		progress.CompletedSteps = append(progress.CompletedSteps, req.StepID)
	}

	progress.UpdatedAt = time.Now().Unix()
	pdata, _ := json.Marshal(progress)
	database.Set([]byte("progress:"+userID+":"+guideID), pdata)

	c.JSON(http.StatusOK, progress)
}

// RegisterRoutes mounts handlers on a group whose path is already /guides (e.g. /api/v1/guides).
// auth is required for create/update/delete of guides and subjects.
func (s *Service) RegisterRoutes(r *gin.RouterGroup, auth gin.HandlerFunc) {
	r.GET("/subjects", s.ListSubjects)
	r.GET("", s.ListGuides)
	r.GET("/:id/progress/:userId", s.GetProgress)
	r.POST("/:id/progress/:userId", s.UpdateProgress)
	r.GET("/:id", s.GetGuide)

	a := r.Group("")
	a.Use(auth)
	a.POST("/subjects", s.CreateSubject)
	a.PATCH("/subjects/:subjectId", s.UpdateSubject)
	a.DELETE("/subjects/:subjectId", s.DeleteSubject)
	a.POST("", s.CreateGuide)
	a.PUT("/:id", s.UpdateGuide)
	a.DELETE("/:id", s.DeleteGuide)
}
