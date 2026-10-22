package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"idongivaflyinfa/importcol"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type importJobStatus string

const (
	jobPending   importJobStatus = "pending"
	jobRunning   importJobStatus = "running"
	jobCompleted importJobStatus = "completed"
	jobFailed    importJobStatus = "failed"
)

type importJob struct {
	ID           string               `json:"id"`
	Entity       string               `json:"entity"`
	Filename     string               `json:"filename"`
	Format       string               `json:"format"`
	Status       importJobStatus      `json:"status"`
	Message      string               `json:"message"`
	Percent      int                  `json:"percent"`
	TotalRows    int                  `json:"total_rows"`
	Processed    int                  `json:"processed_rows"`
	Imported     int                  `json:"imported"`
	Failed       int                  `json:"failed"`
	UsesTemplate bool                 `json:"uses_template"`
	Results      []importcol.RowResult `json:"results,omitempty"`
	CreatedAt    string               `json:"created_at"`
	UpdatedAt    string               `json:"updated_at"`
}

type importJobStore struct {
	mu   sync.RWMutex
	jobs map[string]*importJob
}

func newImportJobStore() *importJobStore {
	return &importJobStore{jobs: map[string]*importJob{}}
}

func (s *importJobStore) put(j *importJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.ID] = j
}

func (s *importJobStore) get(id string) (*importJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	cp := *j
	return &cp, true
}

func (s *importJobStore) update(id string, fn func(*importJob)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return
	}
	fn(j)
	j.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
}

// ListDataCollectorEntities GET /api/data-collector/entities
func (h *Handlers) ListDataCollectorEntities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"entities": importcol.AllSpecs()})
}

// GetDataCollectorTemplate GET /api/data-collector/templates/:entity
func (h *Handlers) GetDataCollectorTemplate(c *gin.Context) {
	kind, ok := importcol.ParseEntityKind(c.Param("entity"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown entity"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity": string(kind), "spec": importcol.SpecFor(kind)})
}

func (h *Handlers) readCollectorUpload(c *gin.Context) (importcol.EntityKind, string, []byte, error) {
	entity := strings.TrimSpace(c.PostForm("entity"))
	kind, ok := importcol.ParseEntityKind(entity)
	if !ok {
		return "", "", nil, fmt.Errorf("entity is required")
	}
	file, hdr, err := c.Request.FormFile("file")
	if err != nil {
		return "", "", nil, fmt.Errorf("file is required")
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, 40<<20))
	if err != nil {
		return "", "", nil, err
	}
	return kind, hdr.Filename, raw, nil
}

// ValidateDataCollectorUpload POST /api/data-collector/validate
func (h *Handlers) ValidateDataCollectorUpload(c *gin.Context) {
	kind, filename, raw, err := h.readCollectorUpload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parsed, err := importcol.ParseUpload(filename, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": importcol.ValidateSample(kind, parsed)})
}

// StartDataCollectorJob POST /api/data-collector/jobs
func (h *Handlers) StartDataCollectorJob(c *gin.Context) {
	if h.TranMySQL == nil || h.TranMySQL.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL required"})
		return
	}
	kind, filename, raw, err := h.readCollectorUpload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parsed, err := importcol.ParseUpload(filename, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report := importcol.ValidateSample(kind, parsed)
	if !report.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": report.Message})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	job := &importJob{
		ID:           uuid.NewString(),
		Entity:       string(kind),
		Filename:     filename,
		Format:       parsed.Format,
		Status:       jobPending,
		Message:      "Import queued",
		Percent:      0,
		TotalRows:    len(parsed.Rows),
		UsesTemplate: report.UsesTemplate,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	h.importJobs.put(job)

	go h.runImportJob(job.ID, kind, parsed)

	c.JSON(http.StatusOK, gin.H{
		"job_id":  job.ID,
		"message": fmt.Sprintf("Import started for %d rows. Poll GET /api/data-collector/jobs/{id} for progress.", len(parsed.Rows)),
	})
}

// GetDataCollectorJob GET /api/data-collector/jobs/:job_id
func (h *Handlers) GetDataCollectorJob(c *gin.Context) {
	j, ok := h.importJobs.get(c.Param("job_id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, j)
}

func (h *Handlers) runImportJob(jobID string, kind importcol.EntityKind, parsed *importcol.ParsedFile) {
	h.importJobs.update(jobID, func(j *importJob) {
		j.Status = jobRunning
		j.Message = "Importing…"
		j.Percent = 5
	})
	results := make([]importcol.RowResult, 0, len(parsed.Rows))
	total := len(parsed.Rows)
	ctx := context.Background()
	for i, row := range parsed.Rows {
		ref := fmt.Sprintf("row %d", i+1)
		r := importcol.InsertRecord(ctx, h.TranMySQL.DB, kind, row, ref)
		results = append(results, r)
		processed := i + 1
		pct := 5 + (processed * 90 / max(total, 1))
		h.importJobs.update(jobID, func(j *importJob) {
			j.Processed = processed
			j.Percent = pct
			j.Message = fmt.Sprintf("Imported %d / %d", processed, total)
		})
	}
	imported, failed := importcol.Summarize(results)
	// trim results for response size
	outResults := results
	if len(outResults) > 500 {
		outResults = outResults[:500]
	}
	h.importJobs.update(jobID, func(j *importJob) {
		j.Status = jobCompleted
		j.Percent = 100
		j.Imported = imported
		j.Failed = failed
		j.Processed = total
		j.Results = outResults
		j.Message = fmt.Sprintf("Done: %d imported, %d failed", imported, failed)
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
