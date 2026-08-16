package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/robo/morphgraph"
)

// KnowledgeFile is a Morph Knowledge Library upload.
type KnowledgeFile struct {
	ID          int64
	Title       string
	Filename    string
	ContentType string
	Kind        string
	StoragePath string
	ByteSize    int64
	TextExcerpt string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// KnowledgeChunk is one searchable slice of a knowledge file.
type KnowledgeChunk struct {
	ID         int64
	FileID     int64
	ChunkIndex int
	Text       string
	Embedding  []float32
}

func (m *TranSQL) EnsureGraphKnowledgeSchema() error {
	if m == nil || m.DB == nil {
		return nil
	}
	if m.isSQLite() {
		// Tables are created in ensureTranSQLiteSchema.
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS graph_sync_outbox (
		  id BIGINT NOT NULL AUTO_INCREMENT,
		  source VARCHAR(32) NOT NULL,
		  entity_type VARCHAR(64) NOT NULL,
		  entity_id VARCHAR(128) NOT NULL,
		  op ENUM('upsert','delete') NOT NULL,
		  payload_json JSON NULL,
		  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
		  available_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
		  attempts INT NOT NULL DEFAULT 0,
		  locked_by VARCHAR(64) NULL,
		  locked_at DATETIME(3) NULL,
		  processed_at DATETIME(3) NULL,
		  last_error TEXT NULL,
		  PRIMARY KEY (id),
		  KEY idx_claim (processed_at, available_at, id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS morph_knowledge_files (
		  id BIGINT NOT NULL AUTO_INCREMENT,
		  title VARCHAR(512) NOT NULL,
		  filename VARCHAR(512) NOT NULL,
		  content_type VARCHAR(128) NOT NULL DEFAULT '',
		  kind VARCHAR(32) NOT NULL,
		  storage_path VARCHAR(1024) NOT NULL,
		  byte_size BIGINT NOT NULL DEFAULT 0,
		  text_excerpt MEDIUMTEXT NULL,
		  created_by VARCHAR(128) NULL,
		  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
		  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
		  PRIMARY KEY (id),
		  KEY idx_updated (updated_at),
		  KEY idx_kind (kind)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS morph_knowledge_chunks (
		  id BIGINT NOT NULL AUTO_INCREMENT,
		  file_id BIGINT NOT NULL,
		  chunk_index INT NOT NULL,
		  text_content MEDIUMTEXT NOT NULL,
		  embedding_json LONGTEXT NULL,
		  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
		  PRIMARY KEY (id),
		  UNIQUE KEY uq_file_chunk (file_id, chunk_index),
		  KEY idx_file (file_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, s := range stmts {
		if _, err := m.DB.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func (m *TranMySQL) InsertKnowledgeFile(ctx context.Context, f *KnowledgeFile) error {
	res, err := m.DB.ExecContext(ctx, `
		INSERT INTO morph_knowledge_files
		  (title, filename, content_type, kind, storage_path, byte_size, text_excerpt, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		f.Title, f.Filename, f.ContentType, f.Kind, f.StoragePath, f.ByteSize, f.TextExcerpt, f.CreatedBy)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	f.ID = id
	return nil
}

func (m *TranMySQL) ListKnowledgeFiles(ctx context.Context, limit int) ([]KnowledgeFile, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	rows, err := m.DB.QueryContext(ctx, `
		SELECT id, title, filename, content_type, kind, storage_path, byte_size,
		       COALESCE(text_excerpt,''), COALESCE(created_by,''), created_at, updated_at
		FROM morph_knowledge_files
		ORDER BY updated_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KnowledgeFile
	for rows.Next() {
		var f KnowledgeFile
		if err := rows.Scan(&f.ID, &f.Title, &f.Filename, &f.ContentType, &f.Kind, &f.StoragePath,
			&f.ByteSize, &f.TextExcerpt, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (m *TranMySQL) GetKnowledgeFile(ctx context.Context, id int64) (*KnowledgeFile, error) {
	var f KnowledgeFile
	err := m.DB.QueryRowContext(ctx, `
		SELECT id, title, filename, content_type, kind, storage_path, byte_size,
		       COALESCE(text_excerpt,''), COALESCE(created_by,''), created_at, updated_at
		FROM morph_knowledge_files WHERE id = ?`, id).
		Scan(&f.ID, &f.Title, &f.Filename, &f.ContentType, &f.Kind, &f.StoragePath,
			&f.ByteSize, &f.TextExcerpt, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (m *TranMySQL) DeleteKnowledgeFile(ctx context.Context, id int64) error {
	_, err := m.DB.ExecContext(ctx, `DELETE FROM morph_knowledge_files WHERE id = ?`, id)
	return err
}

func (m *TranMySQL) ReplaceKnowledgeChunks(ctx context.Context, fileID int64, chunks []KnowledgeChunk) error {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM morph_knowledge_chunks WHERE file_id = ?`, fileID); err != nil {
		return err
	}
	for _, c := range chunks {
		emb := ""
		if len(c.Embedding) > 0 {
			b, _ := json.Marshal(c.Embedding)
			emb = string(b)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO morph_knowledge_chunks (file_id, chunk_index, text_content, embedding_json)
			VALUES (?, ?, ?, ?)`, fileID, c.ChunkIndex, c.Text, nullIfEmpty(emb)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// SearchKnowledgeChunks ranks knowledge chunks by embedding cosine and/or token overlap.
func (m *TranMySQL) SearchKnowledgeChunks(ctx context.Context, query string, queryEmb []float32, limit int) ([]map[string]interface{}, error) {
	if limit < 1 || limit > 50 {
		limit = 12
	}
	rows, err := m.DB.QueryContext(ctx, `
		SELECT c.id, c.file_id, c.chunk_index, c.text_content, COALESCE(c.embedding_json,''),
		       f.title, f.filename, f.kind
		FROM morph_knowledge_chunks c
		JOIN morph_knowledge_files f ON f.id = c.file_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		score float64
		row   map[string]interface{}
	}
	var ranked []scored
	for rows.Next() {
		var id, fileID int64
		var idx int
		var text, embJSON, title, filename, kind string
		if err := rows.Scan(&id, &fileID, &idx, &text, &embJSON, &title, &filename, &kind); err != nil {
			return nil, err
		}
		score := morphgraph.TokenOverlapScore(query, text+" "+title)
		if len(queryEmb) > 0 && embJSON != "" {
			var emb []float32
			if json.Unmarshal([]byte(embJSON), &emb) == nil {
				cos := morphgraph.CosineSim(queryEmb, emb)
				if cos > score {
					score = cos
				} else {
					score = score*0.35 + cos*0.65
				}
			}
		}
		if score <= 0 {
			continue
		}
		ranked = append(ranked, scored{
			score: score,
			row: map[string]interface{}{
				"chunk_id":   id,
				"file_id":    fileID,
				"title":      title,
				"filename":   filename,
				"kind":       kind,
				"chunk_index": idx,
				"text":       morphgraph.TruncateRunes(text, 1200),
				"score":      score,
				"uid":        morphgraph.UID("morph", "knowledge", fmt.Sprintf("%d", fileID)),
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].score > ranked[i].score {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}
	out := make([]map[string]interface{}, 0, limit)
	for i := 0; i < len(ranked) && i < limit; i++ {
		out = append(out, ranked[i].row)
	}
	return out, nil
}

func (m *TranMySQL) EnqueueGraphSync(ctx context.Context, source, entityType, entityID, op string, payloadJSON string) error {
	payload := nullIfEmpty(payloadJSON)
	if m.isSQLite() {
		_, err := m.DB.ExecContext(ctx, `
			INSERT INTO graph_sync_outbox (source, entity_type, entity_id, op, payload_json)
			VALUES (?, ?, ?, ?, ?)`,
			source, entityType, entityID, op, payload)
		return err
	}
	_, err := m.DB.ExecContext(ctx, `
		INSERT INTO graph_sync_outbox (source, entity_type, entity_id, op, payload_json)
		VALUES (?, ?, ?, ?, CAST(? AS JSON))`,
		source, entityType, entityID, op, payload)
	return err
}

func (m *TranMySQL) GraphOutboxDepth(ctx context.Context) (pending int64, err error) {
	err = m.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM graph_sync_outbox WHERE processed_at IS NULL`).Scan(&pending)
	return
}
