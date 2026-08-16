package docs

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/academi/backend/internal/auth"
	"github.com/academi/backend/internal/models"
)

const maxUploadSize = 12 << 20 // 12 MiB

var allowedExt = map[string]struct{}{
	".pdf": {}, ".txt": {}, ".md": {}, ".markdown": {},
	".png": {}, ".jpg": {}, ".jpeg": {}, ".webp": {}, ".gif": {},
	".heic": {}, ".heif": {},
}

// UploadDoc accepts multipart file + optional title. Text is extracted for PDF/txt/md; images stored for vision.
func (s *Service) UploadDoc(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart parse failed"})
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	if fh.Size > maxUploadSize || fh.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file size"})
		return
	}

	ext := strings.ToLower(path.Ext(fh.Filename))
	if _, ok := allowedExt[ext]; ext == "" || !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported type; use pdf, txt, md, or image (png/jpg/webp/gif)"})
		return
	}

	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read failed"})
		return
	}
	defer src.Close()

	blob, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read failed"})
		return
	}
	if int64(len(blob)) > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	detected := http.DetectContentType(blob)
	if len(blob) > 512 {
		detected = http.DetectContentType(blob[:512])
	}

	if err := os.MkdirAll(s.uploadRoot, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "storage init failed"})
		return
	}

	docID := uuid.New().String()
	storedName := docID + ext
	fullPath := filepath.Join(s.uploadRoot, storedName)
	if err := os.WriteFile(fullPath, blob, 0o644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		title = strings.TrimSuffix(fh.Filename, ext)
		if title == "" {
			title = "Uploaded file"
		}
	}

	docType, mimeType := classifyUpload(ext, detected)
	content := ""
	switch docType {
	case "pdf":
		t, err := extractPDFText(fullPath)
		if err != nil {
			content = fmt.Sprintf("(PDF text could not be extracted automatically: %v. The file is saved; use Help you learn with a vision-capable model to analyze.)", err)
		} else {
			content = t
		}
	case "text", "markdown":
		content = string(blob)
		if len(content) > 400_000 {
			content = content[:400_000] + "\n… [truncated]"
		}
	}

	summary := content
	if len(summary) > 320 {
		summary = summary[:320] + "…"
	}
	if docType == "image" {
		summary = "Image upload — use Help you learn or attach in chat (vision model required)."
	}

	tags := []string{"#upload", "#" + docType}
	userID := auth.GetUserID(c)

	doc := &models.Document{
		ID:             docID,
		Title:          title,
		UploaderID:     userID,
		Type:           docType,
		MimeType:       mimeType,
		Size:           fmt.Sprintf("%d", len(blob)),
		Tags:           tags,
		AISummary:      summary,
		Content:        content,
		StoredFilename: storedName,
		SourceFilename: fh.Filename,
		CreatedAt:      time.Now().Unix(),
	}

	if err := s.persistDoc(doc); err != nil {
		_ = os.Remove(fullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "index failed"})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

func classifyUpload(ext, detected string) (docType string, mimeType string) {
	switch ext {
	case ".pdf":
		return "pdf", coalesceMime(detected, "application/pdf")
	case ".txt":
		return "text", coalesceMime(detected, "text/plain")
	case ".md", ".markdown":
		return "markdown", coalesceMime(detected, "text/markdown")
	case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".heic", ".heif":
		mt := coalesceMime(detected, mimeFromExt(ext))
		return "image", mt
	default:
		return "text", detected
	}
}

func mimeFromExt(ext string) string {
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".heic":
		return "image/heic"
	case ".heif":
		return "image/heif"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func coalesceMime(detected, fallback string) string {
	if detected != "" && detected != "application/octet-stream" {
		return detected
	}
	return fallback
}

// ServeDocFile streams the raw uploaded bytes (for development / same-origin web).
func (s *Service) ServeDocFile(c *gin.Context) {
	docID := c.Param("id")
	doc, err := s.GetByID(docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if doc.StoredFilename == "" || strings.Contains(doc.StoredFilename, "..") || strings.Contains(doc.StoredFilename, "/") || strings.Contains(doc.StoredFilename, "\\") {
		c.JSON(http.StatusNotFound, gin.H{"error": "no file"})
		return
	}
	full := filepath.Join(s.uploadRoot, filepath.Base(doc.StoredFilename))
	st, err := os.Stat(full)
	if err != nil || st.IsDir() {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	mt := doc.MimeType
	if mt == "" {
		mt = mime.TypeByExtension(strings.ToLower(path.Ext(doc.StoredFilename)))
	}
	if mt == "" {
		mt = "application/octet-stream"
	}
	c.Header("Content-Type", mt)
	c.File(full)
}
