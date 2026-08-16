package mongo

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/formsx/backend/internal/models"
)

const (
	aiDocKeyPrefix   = "aidoc:"
	aiDocIndexPrefix = "aidoc_idx:"
)

type AIDocumentRepo struct {
	store *Store
}

func NewAIDocumentRepo(store *Store) *AIDocumentRepo {
	return &AIDocumentRepo{store: store}
}

func (r *AIDocumentRepo) EnsureIndexes(ctx context.Context) error {
	_ = ctx
	return nil
}

func aiDocKey(id string) []byte {
	return []byte(aiDocKeyPrefix + id)
}

func aiDocIndexKey(docType, sourceID string) []byte {
	return []byte(aiDocIndexPrefix + docType + ":" + sourceID)
}

func (r *AIDocumentRepo) Upsert(ctx context.Context, doc *models.AIDocument) error {
	_ = ctx
	now := time.Now().UTC()
	return r.store.db.Update(func(txn *badger.Txn) error {
		idx := aiDocIndexKey(doc.DocType, doc.SourceID)
		item, err := txn.Get(idx)
		existingID := ""
		if err == nil {
			_ = item.Value(func(val []byte) error {
				existingID = string(val)
				return nil
			})
		} else if err != badger.ErrKeyNotFound {
			return err
		}

		if existingID != "" {
			var existing models.AIDocument
			if err := getJSON(txn, aiDocKey(existingID), &existing); err == nil {
				doc.ID = existing.ID
				doc.CreatedAt = existing.CreatedAt
			}
		}
		if doc.ID == "" {
			doc.ID = newID()
		}
		if doc.CreatedAt.IsZero() {
			doc.CreatedAt = now
		}
		doc.UpdatedAt = now
		if err := putJSON(txn, aiDocKey(doc.ID), doc); err != nil {
			return err
		}
		return txn.Set(idx, []byte(doc.ID))
	})
}

func (r *AIDocumentRepo) GetByID(ctx context.Context, idHex string) (*models.AIDocument, error) {
	_ = ctx
	var out models.AIDocument
	err := r.store.db.View(func(txn *badger.Txn) error {
		return getJSON(txn, aiDocKey(idHex), &out)
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *AIDocumentRepo) List(ctx context.Context, docType, search string, page, limit int) ([]models.AIDocument, int64, error) {
	_ = ctx
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	var all []models.AIDocument
	err := scanPrefix(r.store.db, []byte(aiDocKeyPrefix), func(_, val []byte) error {
		var doc models.AIDocument
		if err := json.Unmarshal(val, &doc); err != nil {
			return nil
		}
		if docType != "" && doc.DocType != docType {
			return nil
		}
		if search != "" && !containsFold(doc.Title, search) && !containsFold(doc.Summary, search) {
			return nil
		}
		all = append(all, doc)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].UpdatedAt.After(all[j].UpdatedAt)
	})
	total := int64(len(all))
	start := (page - 1) * limit
	if start >= len(all) {
		return []models.AIDocument{}, total, nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}
