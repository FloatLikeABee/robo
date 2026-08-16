package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robo/morphgraph"
)

func (h *Handlers) knowledgeRootDir() string {
	root := strings.TrimSpace(os.Getenv("MORPH_KNOWLEDGE_DIR"))
	if root == "" {
		root = filepath.Join("data", "knowledge")
	}
	_ = os.MkdirAll(root, 0o755)
	return root
}

func detectKnowledgeKind(filename, contentType string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".md":
		return "md"
	case ".json":
		return "json"
	case ".csv":
		return "csv"
	case ".pdf":
		return "pdf"
	case ".txt":
		return "txt"
	}
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "json"):
		return "json"
	case strings.Contains(ct, "csv"):
		return "csv"
	case strings.Contains(ct, "pdf"):
		return "pdf"
	case strings.Contains(ct, "markdown"):
		return "md"
	default:
		return "txt"
	}
}

// ListKnowledgeFiles GET /api/knowledge/files
func (h *Handlers) ListKnowledgeFiles(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL required for Knowledge Library"})
		return
	}
	list, err := h.TranMySQL.ListKnowledgeFiles(c.Request.Context(), 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, f := range list {
		out = append(out, gin.H{
			"id": f.ID, "title": f.Title, "filename": f.Filename, "kind": f.Kind,
			"byte_size": f.ByteSize, "created_at": f.CreatedAt, "updated_at": f.UpdatedAt,
			"excerpt": morphgraph.TruncateRunes(f.TextExcerpt, 240),
		})
	}
	c.JSON(http.StatusOK, gin.H{"files": out, "total": len(out)})
}

func parseFormBool(c *gin.Context, key string, defaultVal bool) bool {
	raw := strings.TrimSpace(c.PostForm(key))
	if raw == "" {
		return defaultVal
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultVal
	}
}

// saveKnowledgeFromBytes stores a knowledge file, chunks/embeds it, and optionally
// enqueues Neo4j GraphRAG sync. Returns the record and a non-fatal warn message.
func (h *Handlers) saveKnowledgeFromBytes(
	c *gin.Context,
	filename, contentType, title string,
	raw []byte,
	indexToGraph bool,
) (*db.KnowledgeFile, string, error) {
	if h.TranMySQL == nil {
		return nil, "", fmt.Errorf("MySQL required for Knowledge Library")
	}
	kind := detectKnowledgeKind(filename, contentType)
	var text string
	if kind == "pdf" {
		extracted, err := morphgraph.ExtractPDFBytes(raw)
		if err != nil {
			return nil, "", fmt.Errorf("PDF extract failed: %w", err)
		}
		text = extracted
	} else {
		text = morphgraph.ExtractPlainText(filename, contentType, string(raw))
	}
	if strings.TrimSpace(text) == "" {
		return nil, "", fmt.Errorf("file has no extractable text")
	}
	if strings.TrimSpace(title) == "" {
		title = filename
	}
	storeName := uuid.NewString() + filepath.Ext(filename)
	path := filepath.Join(h.knowledgeRootDir(), storeName)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return nil, "", err
	}
	rec := &db.KnowledgeFile{
		Title:       title,
		Filename:    filename,
		ContentType: contentType,
		Kind:        kind,
		StoragePath: path,
		ByteSize:    int64(len(raw)),
		TextExcerpt: morphgraph.TruncateRunes(text, 2000),
		CreatedBy:   strings.TrimSpace(c.GetHeader("X-User-ID")),
	}
	if err := h.TranMySQL.InsertKnowledgeFile(c.Request.Context(), rec); err != nil {
		return nil, "", err
	}
	warn := ""
	if err := h.indexKnowledgeFile(c, rec, text); err != nil {
		warn = "saved but indexing incomplete: " + err.Error()
		return rec, warn, nil
	}
	if indexToGraph {
		ref := fmt.Sprintf("%d", rec.ID)
		_ = h.TranMySQL.EnqueueNeo4jIngest(c.Request.Context(), db.Neo4jKindKnowledgeFile, ref, `{"op":"upsert"}`)
		// Legacy outbox kept for morphgraph-worker compatibility when still running.
		_ = h.TranMySQL.EnqueueGraphSync(c.Request.Context(), "morph", "knowledge_file", ref, "upsert", "")
	}
	return rec, warn, nil
}

// UploadKnowledgeFile POST /api/knowledge/files
// Optional form field index_to_graph (default true): enqueue Neo4j GraphRAG sync.
func (h *Handlers) UploadKnowledgeFile(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL required for Knowledge Library"})
		return
	}
	file, hdr, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, 12<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	title := strings.TrimSpace(c.PostForm("title"))
	indexToGraph := parseFormBool(c, "index_to_graph", true)
	rec, warn, err := h.saveKnowledgeFromBytes(c, hdr.Filename, hdr.Header.Get("Content-Type"), title, raw, indexToGraph)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payload := gin.H{
		"file": gin.H{
			"id": rec.ID, "title": rec.Title, "filename": rec.Filename,
			"kind": rec.Kind, "byte_size": rec.ByteSize,
		},
		"index_to_graph": indexToGraph,
	}
	if warn != "" {
		payload["warn"] = warn
		c.JSON(http.StatusOK, payload)
		return
	}
	c.JSON(http.StatusCreated, payload)
}

func (h *Handlers) indexKnowledgeFile(c *gin.Context, rec *db.KnowledgeFile, text string) error {
	chunks := morphgraph.ChunkText(text, morphgraph.DefaultChunkRunes, morphgraph.DefaultChunkOverlap)
	cfg := morphgraph.LoadFromEnv()
	emb := morphgraph.NewEmbedder(cfg)
	var vectors [][]float32
	if emb.Configured() {
		var err error
		vectors, err = emb.Embed(c.Request.Context(), chunks)
		if err != nil {
			// keep text chunks even if embed fails
			vectors = nil
			_ = err
		}
	}
	rows := make([]db.KnowledgeChunk, 0, len(chunks))
	for i, ch := range chunks {
		row := db.KnowledgeChunk{FileID: rec.ID, ChunkIndex: i, Text: ch}
		if i < len(vectors) {
			row.Embedding = vectors[i]
		}
		rows = append(rows, row)
	}
	return h.TranMySQL.ReplaceKnowledgeChunks(c.Request.Context(), rec.ID, rows)
}

// DeleteKnowledgeFile DELETE /api/knowledge/files/:id
func (h *Handlers) DeleteKnowledgeFile(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL required"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	f, err := h.TranMySQL.GetKnowledgeFile(c.Request.Context(), id)
	if err != nil || f == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	_ = os.Remove(f.StoragePath)
	if err := h.TranMySQL.DeleteKnowledgeFile(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ref := fmt.Sprintf("%d", id)
	_ = h.TranMySQL.EnqueueNeo4jIngest(c.Request.Context(), db.Neo4jKindKnowledgeFile, ref, `{"op":"delete"}`)
	_ = h.TranMySQL.EnqueueGraphSync(c.Request.Context(), "morph", "knowledge_file", ref, "delete", "")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GraphHealth GET /api/graph/health
func (h *Handlers) GraphHealth(c *gin.Context) {
	cfg := morphgraph.LoadFromEnv()
	pending := int64(-1)
	if h.TranMySQL != nil {
		_ = h.TranMySQL.EnsureGraphKnowledgeSchema()
		if n, err := h.TranMySQL.GraphOutboxDepth(c.Request.Context()); err == nil {
			pending = n
		}
	}
	neoOK := false
	neoErr := ""
	if cfg.Enabled {
		if store, err := morphgraph.OpenStore(cfg); err != nil {
			neoErr = err.Error()
		} else if store != nil {
			if err := store.Ping(c.Request.Context()); err != nil {
				neoErr = err.Error()
			} else {
				neoOK = true
			}
			_ = store.Close(c.Request.Context())
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"enabled":           cfg.Enabled,
		"neo4j_uri":         cfg.Neo4jURI,
		"neo4j_ok":          neoOK,
		"neo4j_error":       neoErr,
		"embeddings":        morphgraph.NewEmbedder(cfg).Configured(),
		"outbox_pending":    pending,
		"knowledge_library": h.TranMySQL != nil,
	})
}

type graphSearchBody struct {
	Query  string   `json:"query"`
	Limit  int      `json:"limit"`
	Sources []string `json:"sources"`
}

// GraphSearch POST /api/graph/search
func (h *Handlers) GraphSearch(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL required"})
		return
	}
	var body graphSearchBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := strings.TrimSpace(body.Query)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}
	cfg := morphgraph.LoadFromEnv()
	var qEmb []float32
	if emb := morphgraph.NewEmbedder(cfg); emb.Configured() {
		if vecs, err := emb.Embed(c.Request.Context(), []string{query}); err == nil && len(vecs) > 0 {
			qEmb = vecs[0]
		}
	}
	hits, err := h.TranMySQL.SearchKnowledgeChunks(c.Request.Context(), query, qEmb, body.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sources := []string{"morph_knowledge"}
	if cfg.Enabled && len(qEmb) > 0 {
		if store, err := morphgraph.OpenStore(cfg); err == nil && store != nil {
			if neoHits, err := store.VectorSearch(c.Request.Context(), qEmb, body.Limit); err == nil && len(neoHits) > 0 {
				for _, nh := range neoHits {
					hits = append(hits, nh)
				}
				sources = append(sources, "neo4j")
			}
			_ = store.Close(c.Request.Context())
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"hits":    hits,
		"total":   len(hits),
		"sources": sources,
	})
}
