package db

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// Outbox job statuses for neo4j_ingest_outbox.
const (
	Neo4jOutboxPending    = "pending"
	Neo4jOutboxProcessing = "processing"
	Neo4jOutboxDone       = "done"
	Neo4jOutboxFailed     = "failed"
	Neo4jOutboxSkipped    = "skipped"
)

// Neo4j outbox kind values.
const (
	Neo4jKindKnowledgeFile = "knowledge_file"
	Neo4jKindSkill         = "skill"
)

// Neo4jIngestJob is one durable Neo4j ingest queue row.
type Neo4jIngestJob struct {
	ID          int64
	Kind        string
	RefID       string
	PayloadJSON string
	Status      string
	Attempts    int
	LastError   string
	CreatedAt   string
	UpdatedAt   string
}

// Neo4jIngestStatus summarizes queue depth for operators.
type Neo4jIngestStatus struct {
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Failed     int64 `json:"failed"`
	Done       int64 `json:"done"`
	Skipped    int64 `json:"skipped"`
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// EnqueueNeo4jIngest inserts a pending Neo4j ingest job after a durable primary write.
func (m *TranSQL) EnqueueNeo4jIngest(ctx context.Context, kind, refID, payloadJSON string) error {
	if m == nil || m.DB == nil {
		return nil
	}
	kind = strings.TrimSpace(kind)
	refID = strings.TrimSpace(refID)
	if kind == "" || refID == "" {
		return nil
	}
	now := nowRFC3339()
	_, err := m.DB.ExecContext(ctx, `
		INSERT INTO neo4j_ingest_outbox (kind, ref_id, payload_json, status, attempts, last_error, created_at, updated_at)
		VALUES (?, ?, ?, ?, 0, NULL, ?, ?)`,
		kind, refID, nullIfEmpty(strings.TrimSpace(payloadJSON)), Neo4jOutboxPending, now, now)
	return err
}

// ListPendingNeo4jIngestJobs returns pending jobs ready for processing (simple attempt backoff).
func (m *TranSQL) ListPendingNeo4jIngestJobs(ctx context.Context, limit int) ([]Neo4jIngestJob, error) {
	if m == nil || m.DB == nil {
		return nil, nil
	}
	if limit < 1 || limit > 200 {
		limit = 25
	}
	rows, err := m.DB.QueryContext(ctx, `
		SELECT id, kind, ref_id, COALESCE(payload_json,''), status, attempts,
		       COALESCE(last_error,''), created_at, updated_at
		FROM neo4j_ingest_outbox
		WHERE status = ?
		ORDER BY id ASC
		LIMIT ?`, Neo4jOutboxPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Neo4jIngestJob
	now := time.Now().UTC()
	for rows.Next() {
		var j Neo4jIngestJob
		if err := rows.Scan(&j.ID, &j.Kind, &j.RefID, &j.PayloadJSON, &j.Status, &j.Attempts,
			&j.LastError, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		if j.Attempts > 0 {
			updated, err := time.Parse(time.RFC3339, j.UpdatedAt)
			if err == nil {
				backoff := time.Duration(j.Attempts) * 30 * time.Second
				if backoff > 15*time.Minute {
					backoff = 15 * time.Minute
				}
				if now.Sub(updated) < backoff {
					continue
				}
			}
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// MarkNeo4jIngestProcessing sets status=processing for a claimed job.
func (m *TranSQL) MarkNeo4jIngestProcessing(ctx context.Context, id int64) error {
	_, err := m.DB.ExecContext(ctx, `
		UPDATE neo4j_ingest_outbox SET status=?, updated_at=? WHERE id=? AND status=?`,
		Neo4jOutboxProcessing, nowRFC3339(), id, Neo4jOutboxPending)
	return err
}

// MarkNeo4jIngestDone marks a job successfully applied (or intentionally skipped).
func (m *TranSQL) MarkNeo4jIngestDone(ctx context.Context, id int64, status string) error {
	if status != Neo4jOutboxDone && status != Neo4jOutboxSkipped {
		status = Neo4jOutboxDone
	}
	_, err := m.DB.ExecContext(ctx, `
		UPDATE neo4j_ingest_outbox SET status=?, last_error=NULL, updated_at=? WHERE id=?`,
		status, nowRFC3339(), id)
	return err
}

// MarkNeo4jIngestFailure increments attempts and leaves pending or marks failed.
func (m *TranSQL) MarkNeo4jIngestFailure(ctx context.Context, id int64, attempts int, errMsg string, maxAttempts int) error {
	if maxAttempts < 1 {
		maxAttempts = 10
	}
	status := Neo4jOutboxPending
	nextAttempts := attempts + 1
	if nextAttempts >= maxAttempts {
		status = Neo4jOutboxFailed
	}
	if len(errMsg) > 500 {
		errMsg = errMsg[:500]
	}
	_, err := m.DB.ExecContext(ctx, `
		UPDATE neo4j_ingest_outbox SET status=?, attempts=?, last_error=?, updated_at=? WHERE id=?`,
		status, nextAttempts, errMsg, nowRFC3339(), id)
	return err
}

// Neo4jIngestQueueStatus returns counts by status.
func (m *TranSQL) Neo4jIngestQueueStatus(ctx context.Context) (Neo4jIngestStatus, error) {
	var st Neo4jIngestStatus
	if m == nil || m.DB == nil {
		return st, nil
	}
	rows, err := m.DB.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM neo4j_ingest_outbox GROUP BY status`)
	if err != nil {
		return st, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var n int64
		if err := rows.Scan(&status, &n); err != nil {
			return st, err
		}
		switch status {
		case Neo4jOutboxPending:
			st.Pending = n
		case Neo4jOutboxProcessing:
			st.Processing = n
		case Neo4jOutboxFailed:
			st.Failed = n
		case Neo4jOutboxDone:
			st.Done = n
		case Neo4jOutboxSkipped:
			st.Skipped = n
		}
	}
	return st, rows.Err()
}

// ResetStuckNeo4jIngestProcessing returns jobs left in processing (e.g. process crash) to pending.
func (m *TranSQL) ResetStuckNeo4jIngestProcessing(ctx context.Context) error {
	if m == nil || m.DB == nil {
		return nil
	}
	_, err := m.DB.ExecContext(ctx, `
		UPDATE neo4j_ingest_outbox SET status=?, updated_at=? WHERE status=?`,
		Neo4jOutboxPending, nowRFC3339(), Neo4jOutboxProcessing)
	return err
}

// GetNeo4jIngestJob is a small helper for tests/debug.
func (m *TranSQL) GetNeo4jIngestJob(ctx context.Context, id int64) (*Neo4jIngestJob, error) {
	var j Neo4jIngestJob
	err := m.DB.QueryRowContext(ctx, `
		SELECT id, kind, ref_id, COALESCE(payload_json,''), status, attempts,
		       COALESCE(last_error,''), created_at, updated_at
		FROM neo4j_ingest_outbox WHERE id=?`, id).
		Scan(&j.ID, &j.Kind, &j.RefID, &j.PayloadJSON, &j.Status, &j.Attempts,
			&j.LastError, &j.CreatedAt, &j.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &j, nil
}
