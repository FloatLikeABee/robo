package mongo

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/formsx/backend/internal/models"
)

const eventKeyPrefix = "event:"

type EventInfoRepo struct {
	store *Store
}

func NewEventInfoRepo(store *Store) *EventInfoRepo {
	return &EventInfoRepo{store: store}
}

func (r *EventInfoRepo) EnsureIndexes(ctx context.Context) error {
	_ = ctx
	return nil
}

func eventKey(id string) []byte {
	return []byte(eventKeyPrefix + id)
}

func (r *EventInfoRepo) Insert(ctx context.Context, ev *models.EventInfo) error {
	_ = ctx
	if ev.ID == "" {
		ev.ID = newID()
	}
	now := time.Now().UTC()
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = now
	}
	return r.store.db.Update(func(txn *badger.Txn) error {
		return putJSON(txn, eventKey(ev.ID), ev)
	})
}

func (r *EventInfoRepo) GetByID(ctx context.Context, id string) (*models.EventInfo, error) {
	_ = ctx
	var out models.EventInfo
	err := r.store.db.View(func(txn *badger.Txn) error {
		return getJSON(txn, eventKey(id), &out)
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *EventInfoRepo) Delete(ctx context.Context, id string) error {
	_ = ctx
	return r.store.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(eventKey(id)); err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNoDocuments
			}
			return err
		}
		return txn.Delete(eventKey(id))
	})
}

func (r *EventInfoRepo) List(ctx context.Context, page, limit int) ([]models.EventInfo, int64, error) {
	_ = ctx
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	var all []models.EventInfo
	err := scanPrefix(r.store.db, []byte(eventKeyPrefix), func(_, val []byte) error {
		var ev models.EventInfo
		if err := json.Unmarshal(val, &ev); err != nil {
			return nil
		}
		all = append(all, ev)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(all, func(i, j int) bool {
		if !all[i].EventTime.Equal(all[j].EventTime) {
			return all[i].EventTime.After(all[j].EventTime)
		}
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	total := int64(len(all))
	start := (page - 1) * limit
	if start >= len(all) {
		return []models.EventInfo{}, total, nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}
