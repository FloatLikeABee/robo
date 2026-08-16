package mongo

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/formsx/backend/internal/models"
)

const (
	surveyTplKeyPrefix  = "survey_tpl:"
	surveyTplSlugPrefix = "survey_tpl_slug:"
	surveyResKeyPrefix  = "survey_res:"
)

type SurveyBotTemplateRepo struct {
	store *Store
}

func NewSurveyBotTemplateRepo(store *Store) *SurveyBotTemplateRepo {
	return &SurveyBotTemplateRepo{store: store}
}

func (r *SurveyBotTemplateRepo) EnsureIndexes(ctx context.Context) error {
	_ = ctx
	return nil
}

func surveyTplKey(id string) []byte {
	return []byte(surveyTplKeyPrefix + id)
}

func surveyTplSlugKey(slug string) []byte {
	return []byte(surveyTplSlugPrefix + slug)
}

func (r *SurveyBotTemplateRepo) Count(ctx context.Context) (int64, error) {
	_ = ctx
	var n int64
	err := scanPrefix(r.store.db, []byte(surveyTplKeyPrefix), func(_, _ []byte) error {
		n++
		return nil
	})
	return n, err
}

func (r *SurveyBotTemplateRepo) Insert(ctx context.Context, t *models.SurveyBotTemplate) error {
	_ = ctx
	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	if t.ID == "" {
		t.ID = newID()
	}
	return r.store.db.Update(func(txn *badger.Txn) error {
		if t.Slug != "" {
			if _, err := txn.Get(surveyTplSlugKey(t.Slug)); err == nil {
				return badger.ErrConflict
			} else if err != badger.ErrKeyNotFound {
				return err
			}
		}
		if err := putJSON(txn, surveyTplKey(t.ID), t); err != nil {
			return err
		}
		if t.Slug != "" {
			return txn.Set(surveyTplSlugKey(t.Slug), []byte(t.ID))
		}
		return nil
	})
}

func (r *SurveyBotTemplateRepo) Update(ctx context.Context, t *models.SurveyBotTemplate) error {
	_ = ctx
	t.UpdatedAt = time.Now().UTC()
	return r.store.db.Update(func(txn *badger.Txn) error {
		var existing models.SurveyBotTemplate
		if err := getJSON(txn, surveyTplKey(t.ID), &existing); err != nil {
			return err
		}
		if existing.Slug != t.Slug {
			if existing.Slug != "" {
				_ = txn.Delete(surveyTplSlugKey(existing.Slug))
			}
			if t.Slug != "" {
				if item, err := txn.Get(surveyTplSlugKey(t.Slug)); err == nil {
					var otherID string
					_ = item.Value(func(val []byte) error {
						otherID = string(val)
						return nil
					})
					if otherID != "" && otherID != t.ID {
						return badger.ErrConflict
					}
				} else if err != badger.ErrKeyNotFound {
					return err
				}
				if err := txn.Set(surveyTplSlugKey(t.Slug), []byte(t.ID)); err != nil {
					return err
				}
			}
		}
		return putJSON(txn, surveyTplKey(t.ID), t)
	})
}

func (r *SurveyBotTemplateRepo) GetPublishedBySlug(ctx context.Context, slug string) (*models.SurveyBotTemplate, error) {
	t, err := r.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !t.Published {
		return nil, ErrNoDocuments
	}
	return t, nil
}

func (r *SurveyBotTemplateRepo) GetByID(ctx context.Context, idHex string) (*models.SurveyBotTemplate, error) {
	_ = ctx
	var out models.SurveyBotTemplate
	err := r.store.db.View(func(txn *badger.Txn) error {
		return getJSON(txn, surveyTplKey(idHex), &out)
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *SurveyBotTemplateRepo) GetBySlug(ctx context.Context, slug string) (*models.SurveyBotTemplate, error) {
	_ = ctx
	var out models.SurveyBotTemplate
	err := r.store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(surveyTplSlugKey(slug))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNoDocuments
			}
			return err
		}
		var id string
		if err := item.Value(func(val []byte) error {
			id = string(val)
			return nil
		}); err != nil {
			return err
		}
		return getJSON(txn, surveyTplKey(id), &out)
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *SurveyBotTemplateRepo) Delete(ctx context.Context, idHex string) error {
	_ = ctx
	return r.store.db.Update(func(txn *badger.Txn) error {
		var existing models.SurveyBotTemplate
		if err := getJSON(txn, surveyTplKey(idHex), &existing); err != nil {
			return err
		}
		if existing.Slug != "" {
			_ = txn.Delete(surveyTplSlugKey(existing.Slug))
		}
		return txn.Delete(surveyTplKey(idHex))
	})
}

func (r *SurveyBotTemplateRepo) List(ctx context.Context, search string, page, limit int) ([]models.SurveyBotTemplate, int64, error) {
	_ = ctx
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	var all []models.SurveyBotTemplate
	err := scanPrefix(r.store.db, []byte(surveyTplKeyPrefix), func(_, val []byte) error {
		var t models.SurveyBotTemplate
		if err := json.Unmarshal(val, &t); err != nil {
			return nil
		}
		if search != "" {
			tags := strings.Join(t.Tags, " ")
			if !containsFold(t.Title, search) && !containsFold(t.Summary, search) &&
				!containsFold(t.Slug, search) && !containsFold(tags, search) {
				return nil
			}
		}
		all = append(all, t)
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
		return []models.SurveyBotTemplate{}, total, nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

func (r *SurveyBotTemplateRepo) SearchRanked(ctx context.Context, query string, limit int) ([]models.SurveyBotTemplate, error) {
	list, _, err := r.List(ctx, "", 1, 200)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(query) == "" {
		if limit > 0 && len(list) > limit {
			return list[:limit], nil
		}
		return list, nil
	}
	type scored struct {
		t models.SurveyBotTemplate
		s int
	}
	var ranked []scored
	for _, t := range list {
		hay := strings.ToLower(t.Title + " " + t.Summary + " " + strings.Join(t.Tags, " ") + " " + t.Markdown)
		score := 0
		for _, tok := range strings.Fields(strings.ToLower(query)) {
			tok = strings.Trim(tok, ".,!?")
			if len(tok) < 2 {
				continue
			}
			if strings.Contains(hay, tok) {
				score++
			}
		}
		if score > 0 {
			ranked = append(ranked, scored{t: t, s: score})
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].s > ranked[j].s
	})
	out := make([]models.SurveyBotTemplate, 0, len(ranked))
	for _, rnk := range ranked {
		out = append(out, rnk.t)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

type SurveyBotResultRepo struct {
	store *Store
}

func NewSurveyBotResultRepo(store *Store) *SurveyBotResultRepo {
	return &SurveyBotResultRepo{store: store}
}

func (r *SurveyBotResultRepo) EnsureIndexes(ctx context.Context) error {
	_ = ctx
	return nil
}

func surveyResKey(id string) []byte {
	return []byte(surveyResKeyPrefix + id)
}

func (r *SurveyBotResultRepo) Insert(ctx context.Context, res *models.SurveyBotResult) error {
	_ = ctx
	if res.CreatedAt.IsZero() {
		res.CreatedAt = time.Now().UTC()
	}
	if res.ID == "" {
		res.ID = newID()
	}
	return r.store.db.Update(func(txn *badger.Txn) error {
		return putJSON(txn, surveyResKey(res.ID), res)
	})
}

func (r *SurveyBotResultRepo) GetByID(ctx context.Context, idHex string) (*models.SurveyBotResult, error) {
	_ = ctx
	var out models.SurveyBotResult
	err := r.store.db.View(func(txn *badger.Txn) error {
		return getJSON(txn, surveyResKey(idHex), &out)
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *SurveyBotResultRepo) List(ctx context.Context, search string, page, limit int) ([]models.SurveyBotResult, int64, error) {
	_ = ctx
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	var all []models.SurveyBotResult
	err := scanPrefix(r.store.db, []byte(surveyResKeyPrefix), func(_, val []byte) error {
		var res models.SurveyBotResult
		if err := json.Unmarshal(val, &res); err != nil {
			return nil
		}
		if search != "" && !containsFold(res.Title, search) && !containsFold(res.TemplateSlug, search) {
			return nil
		}
		// Omit HTML in list (former projection).
		res.HTML = ""
		all = append(all, res)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	total := int64(len(all))
	start := (page - 1) * limit
	if start >= len(all) {
		return []models.SurveyBotResult{}, total, nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

func (r *SurveyBotResultRepo) Delete(ctx context.Context, idHex string) error {
	_ = ctx
	return r.store.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(surveyResKey(idHex)); err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNoDocuments
			}
			return err
		}
		return txn.Delete(surveyResKey(idHex))
	})
}
