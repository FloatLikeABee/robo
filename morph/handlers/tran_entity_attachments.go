package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	entityAttachmentFacility = "facility"
	entityAttachmentMember   = "member"
	entityAttachmentEmployee = "employee"
	entityAttachmentAsset    = "asset"
	entityAttachmentActivity = "activity"
	entityAttachmentCaseTask  = "case_task"
	entityAttachmentStoryPost = "story_post"

	storyPostMaxImageBytes = 20 * 1024 * 1024
	storyPostMaxVideoBytes = 200 * 1024 * 1024
)

type entityAttachmentOut struct {
	ID           int        `json:"id"`
	EntityType   string     `json:"entity_type"`
	RecordID     int        `json:"record_id"`
	OriginalName string     `json:"original_name"`
	MimeType     *string    `json:"mime_type,omitempty"`
	SizeBytes    int64      `json:"size_bytes"`
	CreatedOn    *time.Time `json:"created_on,omitempty"`
	DownloadURL  string     `json:"download_url"`
	Kind         string     `json:"kind"`
}

type entityAttachmentRoute struct {
	entityType   string
	routeSegment string
	table        string
}

var entityAttachmentRoutes = []entityAttachmentRoute{
	{entityAttachmentFacility, "facilities", sqlFacilityTable},
	{entityAttachmentMember, "members", sqlMemberTable},
	{entityAttachmentEmployee, "employees", sqlEmployeeTable},
	{entityAttachmentAsset, "assets", "Asset"},
	{entityAttachmentActivity, "activities", "Activity"},
	{entityAttachmentCaseTask, "case-tasks", "CaseTask"},
	{entityAttachmentStoryPost, "story-posts", "StoryPost"},
}

func entityAttachmentRouteBySegment(segment string) (entityAttachmentRoute, bool) {
	s := strings.Trim(strings.ToLower(strings.TrimSpace(segment)), "/")
	for _, r := range entityAttachmentRoutes {
		if r.routeSegment == s {
			return r, true
		}
	}
	return entityAttachmentRoute{}, false
}

func (h *Handlers) entityAttachmentBaseDir() string {
	if strings.TrimSpace(h.entityAttachmentRootDir) != "" {
		return h.entityAttachmentRootDir
	}
	return filepath.Join("uploads", "entity_attachments")
}

func (h *Handlers) entityAttachmentDir(entityType string, recordID int) string {
	return filepath.Join(h.entityAttachmentBaseDir(), entityType, strconv.Itoa(recordID))
}

func (h *Handlers) entityAttachmentMaxPerRecord() int {
	if h.entityAttachmentMax < 1 {
		return 10
	}
	return h.entityAttachmentMax
}

func (h *Handlers) entityAttachmentMaxForEntity(entityType string) int {
	if entityType == entityAttachmentStoryPost {
		// Allow a few AI-generated images per story (plus optional user media).
		return storyAIMaxImages
	}
	return h.entityAttachmentMaxPerRecord()
}

func validateStoryPostAttachmentFile(fh *multipart.FileHeader) error {
	if fh == nil {
		return fmt.Errorf("invalid file")
	}
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fh.Filename)))
	if !isAllowedImageExt(ext) && !isAllowedVideoExt(ext) {
		return fmt.Errorf("Story Board posts only accept images or videos")
	}
	if isAllowedImageExt(ext) && fh.Size > storyPostMaxImageBytes {
		return fmt.Errorf("images must be under 20 MB")
	}
	if isAllowedVideoExt(ext) && fh.Size > storyPostMaxVideoBytes {
		return fmt.Errorf("videos must be under 200 MB")
	}
	return nil
}

func (h *Handlers) entityRecordExists(ctx context.Context, table string, id int) (bool, error) {
	if h.TranMySQL == nil || id <= 0 || strings.TrimSpace(table) == "" {
		return false, nil
	}
	var n int
	err := h.TranMySQL.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ?", table), id).Scan(&n)
	return n > 0, err
}

func attachmentKind(originalName string, mime *string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	if isAllowedImageExt(ext) {
		return "image"
	}
	if isAllowedVideoExt(ext) {
		return "video"
	}
	if mime != nil {
		m := strings.ToLower(strings.TrimSpace(*mime))
		if strings.HasPrefix(m, "image/") {
			return "image"
		}
		if strings.HasPrefix(m, "video/") {
			return "video"
		}
	}
	return "document"
}

var allowedAttachmentExts = map[string]struct{}{
	".pdf": {}, ".xlsx": {}, ".xls": {}, ".csv": {}, ".json": {}, ".txt": {},
	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {}, ".bmp": {}, ".svg": {}, ".heic": {}, ".heif": {},
	".mp4": {}, ".webm": {}, ".mov": {}, ".avi": {}, ".mkv": {}, ".m4v": {},
}

func isAllowedImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".heic", ".heif":
		return true
	default:
		return false
	}
}

func isAllowedVideoExt(ext string) bool {
	switch ext {
	case ".mp4", ".webm", ".mov", ".avi", ".mkv", ".m4v":
		return true
	default:
		return false
	}
}

func isAllowedAttachmentFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(name)))
	if ext == "" {
		return false
	}
	_, ok := allowedAttachmentExts[ext]
	return ok
}

func (h *Handlers) listEntityAttachments(ctx context.Context, routeSegment string, entityType string, recordID int) ([]entityAttachmentOut, error) {
	if h.TranMySQL == nil {
		return nil, nil
	}
	rows, err := h.TranMySQL.DB.QueryContext(ctx, `
		SELECT id, entity_type, record_id, original_name, mime_type, size_bytes, created_on
		FROM EntityAttachment
		WHERE entity_type = ? AND record_id = ?
		ORDER BY id`, entityType, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]entityAttachmentOut, 0)
	for rows.Next() {
		var a entityAttachmentOut
		var mime sql.NullString
		var created sql.NullTime
		if err := rows.Scan(&a.ID, &a.EntityType, &a.RecordID, &a.OriginalName, &mime, &a.SizeBytes, &created); err != nil {
			return nil, err
		}
		if mime.Valid {
			s := mime.String
			a.MimeType = &s
		}
		if created.Valid {
			ts := created.Time
			a.CreatedOn = &ts
		}
		a.Kind = attachmentKind(a.OriginalName, a.MimeType)
		a.DownloadURL = fmt.Sprintf("/api/tran/%s/%d/attachments/%d/download", routeSegment, recordID, a.ID)
		list = append(list, a)
	}
	return list, nil
}

func (h *Handlers) attachEntityAttachmentsToRow(ctx context.Context, routeSegment, entityType string, recordID int, row map[string]interface{}) {
	if row == nil {
		return
	}
	list, err := h.listEntityAttachments(ctx, routeSegment, entityType, recordID)
	if err != nil {
		row["attachments"] = []entityAttachmentOut{}
		return
	}
	if list == nil {
		list = []entityAttachmentOut{}
	}
	row["attachments"] = list
}

func (h *Handlers) purgeEntityAttachments(ctx context.Context, entityType string, recordID int) {
	if h.TranMySQL == nil {
		return
	}
	rows, err := h.TranMySQL.DB.QueryContext(ctx, `
		SELECT file_path FROM EntityAttachment WHERE entity_type = ? AND record_id = ?`, entityType, recordID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var fp sql.NullString
			if rows.Scan(&fp) == nil && fp.Valid && strings.TrimSpace(fp.String) != "" {
				_ = os.Remove(fp.String)
			}
		}
	}
	_, _ = h.TranMySQL.DB.ExecContext(ctx, `DELETE FROM EntityAttachment WHERE entity_type = ? AND record_id = ?`, entityType, recordID)
	_ = os.RemoveAll(h.entityAttachmentDir(entityType, recordID))
}

func (h *Handlers) GetAttachmentConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"max_per_record": h.entityAttachmentMaxPerRecord(),
		"allowed_extensions": []string{
			"pdf", "xlsx", "xls", "csv", "json", "txt",
			"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg", "heic", "heif",
			"mp4", "webm", "mov", "avi", "mkv", "m4v",
		},
	})
}

func (h *Handlers) UploadEntityAttachments(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	route, ok := entityAttachmentRouteBySegment(c.Param("entity"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported entity type"})
		return
	}
	recordID, err := strconv.Atoi(c.Param("id"))
	if err != nil || recordID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ctx := c.Request.Context()
	exists, err := h.entityRecordExists(ctx, route.table, recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	count := 0
	if err := h.TranMySQL.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM EntityAttachment WHERE entity_type = ? AND record_id = ?`,
		route.entityType, recordID).Scan(&count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart form required"})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files uploaded"})
		return
	}
	max := h.entityAttachmentMaxForEntity(route.entityType)
	if count+len(files) > max {
		if route.entityType == entityAttachmentStoryPost {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("up to %d images or videos are allowed per Story Board post", max)})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("up to %d attachments are allowed per record", max)})
		return
	}

	for _, fh := range files {
		if route.entityType == entityAttachmentStoryPost {
			if err := validateStoryPostAttachmentFile(fh); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			continue
		}
		if !isAllowedAttachmentFile(fh.Filename) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file type not allowed: %s", fh.Filename)})
			return
		}
	}

	dir := h.entityAttachmentDir(route.entityType, recordID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, fh := range files {
		if err := h.saveEntityAttachmentFile(c, route.entityType, recordID, dir, fh); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	attachments, err := h.listEntityAttachments(ctx, route.routeSegment, route.entityType, recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"attachments": attachments})
}

func (h *Handlers) saveEntityAttachmentFile(c *gin.Context, entityType string, recordID int, dir string, fh *multipart.FileHeader) error {
	base := filepath.Base(strings.TrimSpace(fh.Filename))
	if base == "" {
		base = "document"
	}
	stored := uuid.New().String() + "_" + base
	fullPath := filepath.Join(dir, stored)
	if err := c.SaveUploadedFile(fh, fullPath); err != nil {
		return err
	}
	var mime interface{}
	if strings.TrimSpace(fh.Header.Get("Content-Type")) != "" {
		mime = strings.TrimSpace(fh.Header.Get("Content-Type"))
	}
	_, err := h.TranMySQL.DB.Exec(`
		INSERT INTO EntityAttachment(entity_type, record_id, original_name, stored_name, file_path, mime_type, size_bytes)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entityType, recordID, base, stored, fullPath, mime, fh.Size)
	if err != nil {
		_ = os.Remove(fullPath)
	}
	return err
}

func (h *Handlers) DeleteEntityAttachment(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	route, ok := entityAttachmentRouteBySegment(c.Param("entity"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported entity type"})
		return
	}
	recordID, err := strconv.Atoi(c.Param("id"))
	if err != nil || recordID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	attachmentID, err := strconv.Atoi(c.Param("attachmentId"))
	if err != nil || attachmentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attachment id"})
		return
	}
	var filePath string
	err = h.TranMySQL.DB.QueryRow(`
		SELECT file_path FROM EntityAttachment
		WHERE id = ? AND entity_type = ? AND record_id = ?`,
		attachmentID, route.entityType, recordID).Scan(&filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}
	if _, err := h.TranMySQL.DB.Exec(`
		DELETE FROM EntityAttachment WHERE id = ? AND entity_type = ? AND record_id = ?`,
		attachmentID, route.entityType, recordID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(filePath) != "" {
		_ = os.Remove(filePath)
	}
	c.JSON(http.StatusOK, gin.H{"deleted": attachmentID})
}

func (h *Handlers) DownloadEntityAttachment(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	route, ok := entityAttachmentRouteBySegment(c.Param("entity"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported entity type"})
		return
	}
	recordID, err := strconv.Atoi(c.Param("id"))
	if err != nil || recordID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	attachmentID, err := strconv.Atoi(c.Param("attachmentId"))
	if err != nil || attachmentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attachment id"})
		return
	}
	var filePath, originalName string
	err = h.TranMySQL.DB.QueryRow(`
		SELECT file_path, original_name
		FROM EntityAttachment
		WHERE id = ? AND entity_type = ? AND record_id = ?`,
		attachmentID, route.entityType, recordID).Scan(&filePath, &originalName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}
	if _, statErr := os.Stat(filePath); statErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found on server"})
		return
	}
	c.FileAttachment(filePath, originalName)
}
