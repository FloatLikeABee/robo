package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/formsx/backend/internal/models"
)

const respKeyPrefix = "resp:"

var errStopScan = errors.New("stop scan")

type ResponseRepo struct {
	store *Store
}

func NewResponseRepo(store *Store) *ResponseRepo {
	return &ResponseRepo{store: store}
}

func (r *ResponseRepo) EnsureIndexes(ctx context.Context) error {
	_ = ctx
	return nil
}

func respKey(id string) []byte {
	return []byte(respKeyPrefix + id)
}

func (r *ResponseRepo) Insert(ctx context.Context, resp *models.FormResponse) error {
	_ = ctx
	if resp.ID == "" {
		resp.ID = newID()
	}
	if resp.SubmittedAt.IsZero() {
		resp.SubmittedAt = time.Now().UTC()
	}
	return r.store.db.Update(func(txn *badger.Txn) error {
		return putJSON(txn, respKey(resp.ID), resp)
	})
}

func (r *ResponseRepo) ExistsByFormAndRespondent(ctx context.Context, formID int64, respondentID string) (bool, error) {
	_ = ctx
	if respondentID == "" {
		return false, nil
	}
	found := false
	err := scanPrefix(r.store.db, []byte(respKeyPrefix), func(_, val []byte) error {
		var resp models.FormResponse
		if err := json.Unmarshal(val, &resp); err != nil {
			return nil
		}
		if resp.FormID == formID && resp.RespondentID == respondentID {
			found = true
			return errStopScan
		}
		return nil
	})
	if errors.Is(err, errStopScan) {
		return true, nil
	}
	return found, err
}

func (r *ResponseRepo) List(ctx context.Context, formID int64, page, limit int, since, until *int64) ([]models.FormResponse, int64, error) {
	_ = ctx
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	var all []models.FormResponse
	err := scanPrefix(r.store.db, []byte(respKeyPrefix), func(_, val []byte) error {
		var resp models.FormResponse
		if err := json.Unmarshal(val, &resp); err != nil {
			return nil
		}
		if resp.FormID != formID {
			return nil
		}
		if since != nil && resp.SubmittedAt.Before(time.UnixMilli(*since)) {
			return nil
		}
		if until != nil && resp.SubmittedAt.After(time.UnixMilli(*until)) {
			return nil
		}
		all = append(all, resp)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].SubmittedAt.After(all[j].SubmittedAt)
	})
	total := int64(len(all))
	start := (page - 1) * limit
	if start >= len(all) {
		return []models.FormResponse{}, total, nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

func (r *ResponseRepo) DeleteByID(ctx context.Context, formID int64, responseID string) error {
	_ = ctx
	return r.store.db.Update(func(txn *badger.Txn) error {
		var resp models.FormResponse
		if err := getJSON(txn, respKey(responseID), &resp); err != nil {
			return err
		}
		if resp.FormID != formID {
			return fmt.Errorf("response not found")
		}
		return txn.Delete(respKey(responseID))
	})
}

func (r *ResponseRepo) GetByID(ctx context.Context, formID int64, responseID string) (*models.FormResponse, error) {
	_ = ctx
	var out models.FormResponse
	err := r.store.db.View(func(txn *badger.Txn) error {
		if err := getJSON(txn, respKey(responseID), &out); err != nil {
			return err
		}
		if out.FormID != formID {
			return ErrNoDocuments
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
