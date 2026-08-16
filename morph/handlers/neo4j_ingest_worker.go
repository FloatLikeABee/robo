package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"idongivaflyinfa/db"

	"github.com/robo/morphgraph"
)

const neo4jIngestMaxAttempts = 10
const neo4jIngestPollInterval = 3 * time.Second

// StartNeo4jIngestWorker starts a background goroutine that drains neo4j_ingest_outbox.
// It never blocks HTTP; Neo4j failures leave jobs pending/failed with logs.
func (h *Handlers) StartNeo4jIngestWorker(ctx context.Context) {
	if h == nil || h.TranMySQL == nil {
		log.Printf("[neo4j-ingest] worker not started (no SQLite)")
		return
	}
	_ = h.TranMySQL.ResetStuckNeo4jIngestProcessing(ctx)
	go h.runNeo4jIngestWorker(ctx)
	log.Printf("[neo4j-ingest] worker started (poll %s)", neo4jIngestPollInterval)
}

func (h *Handlers) runNeo4jIngestWorker(ctx context.Context) {
	ticker := time.NewTicker(neo4jIngestPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("[neo4j-ingest] worker stopped")
			return
		case <-ticker.C:
			if err := h.drainNeo4jIngestOnce(ctx); err != nil {
				log.Printf("[neo4j-ingest] drain: %v", err)
			}
		}
	}
}

func (h *Handlers) drainNeo4jIngestOnce(ctx context.Context) error {
	jobs, err := h.TranMySQL.ListPendingNeo4jIngestJobs(ctx, 25)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return nil
	}

	cfg := morphgraph.LoadFromEnv()
	var store *morphgraph.Store
	neoMissing := !cfg.Enabled
	if cfg.Enabled {
		s, openErr := morphgraph.OpenStore(cfg)
		if openErr != nil {
			log.Printf("[neo4j-ingest] neo4j unavailable: %v", openErr)
			for _, j := range jobs {
				_ = h.TranMySQL.MarkNeo4jIngestFailure(ctx, j.ID, j.Attempts, openErr.Error(), neo4jIngestMaxAttempts)
			}
			return nil
		}
		store = s
		if store != nil {
			defer func() { _ = store.Close(ctx) }()
		}
	} else {
		log.Printf("[neo4j-ingest] Neo4j disabled; draining jobs as skipped")
	}

	for _, j := range jobs {
		if err := h.TranMySQL.MarkNeo4jIngestProcessing(ctx, j.ID); err != nil {
			continue
		}
		if neoMissing || store == nil {
			if err := h.TranMySQL.MarkNeo4jIngestDone(ctx, j.ID, db.Neo4jOutboxSkipped); err != nil {
				log.Printf("[neo4j-ingest] mark skipped id=%d: %v", j.ID, err)
			}
			continue
		}
		applyErr := h.applyNeo4jIngestJob(ctx, store, cfg, j)
		if applyErr != nil {
			log.Printf("[neo4j-ingest] job id=%d kind=%s ref=%s failed: %v", j.ID, j.Kind, j.RefID, applyErr)
			_ = h.TranMySQL.MarkNeo4jIngestFailure(ctx, j.ID, j.Attempts, applyErr.Error(), neo4jIngestMaxAttempts)
			continue
		}
		if err := h.TranMySQL.MarkNeo4jIngestDone(ctx, j.ID, db.Neo4jOutboxDone); err != nil {
			log.Printf("[neo4j-ingest] mark done id=%d: %v", j.ID, err)
		}
	}
	return nil
}

func (h *Handlers) applyNeo4jIngestJob(ctx context.Context, store *morphgraph.Store, cfg morphgraph.Config, j db.Neo4jIngestJob) error {
	op := "upsert"
	if j.PayloadJSON != "" {
		var meta map[string]interface{}
		if json.Unmarshal([]byte(j.PayloadJSON), &meta) == nil {
			if v, ok := meta["op"].(string); ok && strings.TrimSpace(v) != "" {
				op = strings.ToLower(strings.TrimSpace(v))
			}
		}
	}
	uid := morphgraph.UID("morph", j.Kind, j.RefID)
	if op == "delete" {
		return store.DeleteEntity(ctx, uid)
	}
	switch j.Kind {
	case db.Neo4jKindKnowledgeFile:
		return h.syncKnowledgeFileToNeo4j(ctx, store, cfg, j.RefID)
	case db.Neo4jKindSkill:
		return h.syncSkillToNeo4j(ctx, store, j.RefID)
	default:
		return store.UpsertEntity(ctx, morphgraph.NodeProps{
			UID: uid, Source: "morph", Type: j.Kind, SourceID: j.RefID,
			Title: j.Kind + " " + j.RefID, UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			Labels: []string{"Entity"},
		})
	}
}

func (h *Handlers) syncKnowledgeFileToNeo4j(ctx context.Context, store *morphgraph.Store, cfg morphgraph.Config, refID string) error {
	id, err := strconv.ParseInt(refID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid knowledge file id: %w", err)
	}
	f, err := h.TranMySQL.GetKnowledgeFile(ctx, id)
	if err != nil {
		return err
	}
	if f == nil {
		return fmt.Errorf("knowledge file %s not found", refID)
	}
	uid := morphgraph.UID("morph", "knowledge_file", refID)
	if err := store.UpsertEntity(ctx, morphgraph.NodeProps{
		UID: uid, Source: "morph", Type: "knowledge_file", SourceID: refID,
		Title: f.Title, Summary: morphgraph.TruncateRunes(f.TextExcerpt, 400),
		Labels: []string{"KnowledgeFile"}, UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		return err
	}
	// Best-effort chunks from SQLite (embeddings already stored if available).
	rows, err := h.TranMySQL.DB.QueryContext(ctx, `
		SELECT chunk_index, text_content, COALESCE(embedding_json,'')
		FROM morph_knowledge_chunks WHERE file_id=? ORDER BY chunk_index`, id)
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var idx int
		var text, embJSON string
		if rows.Scan(&idx, &text, &embJSON) != nil {
			continue
		}
		var emb []float32
		if embJSON != "" {
			_ = json.Unmarshal([]byte(embJSON), &emb)
		}
		cuid := morphgraph.UID("morph", "knowledge_chunk", fmt.Sprintf("%d_%d", id, idx))
		_ = store.UpsertChunk(ctx, cuid, uid, text, emb)
	}
	_ = cfg
	return nil
}

func (h *Handlers) syncSkillToNeo4j(ctx context.Context, store *morphgraph.Store, refID string) error {
	s, err := h.TranMySQL.GetAISkill(ctx, refID)
	if err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("skill %s not found", refID)
	}
	summary := s.Description
	if h.db != nil {
		if body, err := h.db.GetAISkillBody(refID); err == nil && strings.TrimSpace(body.Instructions) != "" {
			summary = morphgraph.TruncateRunes(s.Description+"\n"+body.Instructions, 500)
		}
	}
	return store.UpsertEntity(ctx, morphgraph.NodeProps{
		UID: morphgraph.UID("morph", "skill", refID), Source: "morph", Type: "skill", SourceID: refID,
		Title: s.Name, Summary: summary, Labels: []string{"Skill"},
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	})
}
