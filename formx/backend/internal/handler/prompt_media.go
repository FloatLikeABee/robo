package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/formsx/backend/internal/models"
	"github.com/gin-gonic/gin"
)

var questionPromptMediaRelPathRe = regexp.MustCompile(`^question-media/form_\d+/q_\d+\.[a-zA-Z0-9]+$`)

var imagePromptMimes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var videoPromptMimes = map[string]string{
	"video/mp4":  ".mp4",
	"video/webm": ".webm",
}

func normalizeQuestionTitle(title string) (string, error) {
	t := strings.TrimSpace(title)
	if t == "" {
		return "", fmt.Errorf("question title cannot be empty")
	}
	return t, nil
}

func deletePromptMediaOnDisk(uploadDir, relPath string) {
	norm := strings.ReplaceAll(relPath, "\\", "/")
	if norm == "" || !questionPromptMediaRelPathRe.MatchString(norm) {
		return
	}
	full := filepath.Join(uploadDir, filepath.FromSlash(norm))
	_ = os.Remove(full)
}

// UploadQuestionPromptMedia uploads one image or one video shown with the question on the public form.
func (h *Handler) UploadQuestionPromptMedia(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	qID, err := strconv.ParseInt(c.Param("questionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	if _, err := h.FormRepo.GetByID(formID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	q, err := h.QuestionRepo.GetByFormIDAndID(formID, qID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": `missing multipart field "file"`})
		return
	}
	headerMime := strings.ToLower(strings.TrimSpace(strings.Split(fh.Header.Get("Content-Type"), ";")[0]))

	kind, ext := classifyPromptMedia(headerMime, fh.Size)

	if kind == "" || ext == "" {
		srcProbe, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read upload"})
			return
		}
		probeLimit := io.LimitReader(srcProbe, 4096)
		head, _ := io.ReadAll(probeLimit)
		_ = srcProbe.Close()
		kind, ext = classifyPromptMediaFromContent(headerMime, head, fh.Size)
	}
	if kind == "" || ext == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type; use JPEG, PNG, GIF, WebP, MP4, or WebM"})
		return
	}

	maxBytes := int64(models.MaxQuestionPromptImageBytes)
	if kind == models.PromptMediaKindVideo {
		maxBytes = int64(models.MaxQuestionPromptVideoBytes)
	}
	if fh.Size > maxBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file too large (max %d MiB for %s prompts)",
			maxBytes/(1024*1024), kind)})
		return
	}

	relPath := fmt.Sprintf("question-media/form_%d/q_%d%s", formID, qID, ext)
	dir := filepath.Join(h.Cfg.UploadDir, fmt.Sprintf("question-media/form_%d", formID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if q.Config.QuestionPromptMedia != nil && q.Config.QuestionPromptMedia.RelativePath != "" {
		deletePromptMediaOnDisk(h.Cfg.UploadDir, q.Config.QuestionPromptMedia.RelativePath)
	}

	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read upload"})
		return
	}
	defer src.Close()

	fullPath := filepath.Join(h.Cfg.UploadDir, filepath.FromSlash(relPath))
	dst, err := os.Create(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer dst.Close()

	limited := io.LimitReader(src, maxBytes+1) // int64
	written, err := io.Copy(dst, limited)
	if err != nil {
		_ = os.Remove(fullPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if written > maxBytes {
		_ = os.Remove(fullPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file exceeds %d MiB limit", maxBytes/(1024*1024))})
		return
	}

	q.Config.QuestionPromptMedia = &models.QuestionPromptMedia{
		Kind:         kind,
		RelativePath: relPath,
	}
	if err := h.QuestionRepo.Update(q); err != nil {
		_ = os.Remove(fullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

func classifyPromptMedia(headerMime string, declaredSize int64) (kind string, ext string) {
	if e, ok := imagePromptMimes[headerMime]; ok && declaredSize <= models.MaxQuestionPromptImageBytes {
		return models.PromptMediaKindImage, e
	}
	if e, ok := videoPromptMimes[headerMime]; ok && declaredSize <= models.MaxQuestionPromptVideoBytes {
		return models.PromptMediaKindVideo, e
	}
	return "", ""
}

func classifyPromptMediaFromContent(headerMime string, head []byte, declaredSize int64) (kind string, ext string) {
	kind, ext = classifyPromptMedia(headerMime, declaredSize)
	if kind != "" {
		return kind, ext
	}

	detected := ""
	if len(head) > 0 {
		detected = strings.ToLower(http.DetectContentType(head))
	}
	if e, ok := imagePromptMimes[detected]; ok && declaredSize <= models.MaxQuestionPromptImageBytes {
		return models.PromptMediaKindImage, e
	}
	if e, ok := videoPromptMimes[detected]; ok && declaredSize <= models.MaxQuestionPromptVideoBytes {
		return models.PromptMediaKindVideo, e
	}
	if declaredSize <= models.MaxQuestionPromptImageBytes {
		switch {
		case len(head) >= 3 && head[0] == 0xff && head[1] == 0xd8 && head[2] == 0xff:
			return models.PromptMediaKindImage, ".jpg"
		case len(head) >= 8 && bytes.Equal(head[0:8], []byte{0x89, 0x50, 0x4e, 0x47, 0xd, 0xa, 0x1a, 0xa}):
			return models.PromptMediaKindImage, ".png"
		case len(head) >= 6 && (bytes.Equal(head[0:6], []byte("GIF87a")) || bytes.Equal(head[0:6], []byte("GIF89a"))):
			return models.PromptMediaKindImage, ".gif"
		case len(head) >= 12 && bytes.Equal(head[0:4], []byte("RIFF")) && bytes.Equal(head[8:12], []byte("WEBP")):
			return models.PromptMediaKindImage, ".webp"
		}
	}
	if declaredSize <= models.MaxQuestionPromptVideoBytes {
		if seemsMP4(head) {
			return models.PromptMediaKindVideo, ".mp4"
		}
		if len(head) >= 4 && head[0] == 0x1a && head[1] == 0x45 && head[2] == 0xdf && head[3] == 0xa3 {
			return models.PromptMediaKindVideo, ".webm"
		}
	}
	return "", ""
}

func seemsMP4(b []byte) bool {
	if len(b) < 12 {
		return false
	}
	return bytes.Equal(b[4:8], []byte("ftyp"))
}

// DeleteQuestionPromptMedia removes attached prompt media file and clears config.
func (h *Handler) DeleteQuestionPromptMedia(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	qID, err := strconv.ParseInt(c.Param("questionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	q, err := h.QuestionRepo.GetByFormIDAndID(formID, qID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	if q.Config.QuestionPromptMedia != nil {
		deletePromptMediaOnDisk(h.Cfg.UploadDir, q.Config.QuestionPromptMedia.RelativePath)
		q.Config.QuestionPromptMedia = nil
		if err := h.QuestionRepo.Update(q); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, q)
}
