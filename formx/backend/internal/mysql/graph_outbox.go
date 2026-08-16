package mysql

import (
	"context"
	"fmt"
)

// EnqueueGraphSync writes a best-effort outbox row for Morph GraphRAG sync.
func (r *FormRepo) EnqueueGraphSync(ctx context.Context, source, entityType, entityID, op string) {
	if r == nil || r.db == nil {
		return
	}
	_ = r.db.WithContext(ctx).Exec(`
		INSERT INTO graph_sync_outbox (source, entity_type, entity_id, op)
		VALUES (?, ?, ?, ?)`, source, entityType, entityID, op).Error
}

func (r *FormRepo) EnqueueFormUpsert(ctx context.Context, formID int64) {
	r.EnqueueGraphSync(ctx, "formsx", "form", fmt.Sprintf("%d", formID), "upsert")
}

func (r *FormRepo) EnqueueFormDelete(ctx context.Context, formID int64) {
	r.EnqueueGraphSync(ctx, "formsx", "form", fmt.Sprintf("%d", formID), "delete")
}
