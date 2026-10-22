package db

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"idongivaflyinfa/models"
)

const dorphAIAgentPrefix = "dorph_ai_agent:"
const legacyAIAgentPrefix = "morph_ai_agent:"

// DefaultDorphAIAgents are seeded when none exist yet.
func DefaultDorphAIAgents() []models.DorphAIAgent {
	now := time.Now().UTC().Format(time.RFC3339)
	return []models.DorphAIAgent{
		{
			ID:            "general",
			Name:          "General assistant",
			Description:   "Default Dorph AI — registration, sheets, Tran data questions, or general conversation.",
			Instructions:  "",
			SystemDefined: true,
			SortOrder:     0,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "data-report",
			Name:          "Data report generator",
			Description:   "Produce structured summaries and export-friendly reports from Tran data (tables, bullets, CSV-style layouts).",
			Instructions:  "Focus on readable reports from organization data (members, facilities, routes, contacts, etc.). Prefer calling Tran APIs to gather facts, then present results as clear markdown tables or sections (headers, rows, totals). Mention when figures are approximate or truncated. Avoid SQL unless explaining read-only summaries.",
			SystemDefined: true,
			SortOrder:     10,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "data-statistics",
			Name:          "Data statistics maker",
			Description:   "Counts, breakdowns, and simple analytics over lists gathered from Tran APIs.",
			Instructions:  "Prefer GET requests to Tran list endpoints with filters when needed. Produce numeric summaries: totals, percentages, group-by style breakdowns when the payload allows. Explain methodology briefly. Highlight anomalies or zeros worth attention.",
			SystemDefined: true,
			SortOrder:     20,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "route-operations",
			Name:          "Route & operations planner",
			Description:   "Trips, vehicles, facilities, and day-to-day transportation operations.",
			Instructions:  "Optimize answers around activities/trips and assets/vehicles, facilities, districts, contacts, employees. When users ask operational questions, propose concrete checklist-style steps grounded in Tran data you fetch. Mention safety/regulatory disclaimers briefly when relevant (you are advising, not certifying compliance).",
			SystemDefined: true,
			SortOrder:     30,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "case-task-helper",
			Name:          "Cases & tasks specialist",
			Description:   "Structured help for case/task workflows, timelines, assignments, notes.",
			Instructions:  "Orient toward case-task entities: fetch details with full payloads when listing is not enough; summarize statuses, deadlines, owners, attachments. Prefer actionable next-step lists. Respect privacy: summarize sensitive fields neutrally.",
			SystemDefined: true,
			SortOrder:     40,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "forms-sheets-specialist",
			Name:          "Forms & sheets specialist",
			Description:   "Drafting sheets (data collection templates) and interpreting form templates/answers concepts.",
			Instructions:  "Lead with sheet/form usability: sensible column names, data types, and validation hints. If the chat flow offers a proposed sheet UX, steer users to review columns before saving. Connect to DorphData forms/templates when answering process questions.",
			SystemDefined: true,
			SortOrder:     50,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "image-generator",
			Name:          "Image generator",
			Description:   "Create an image from your prompt and show it in the chat.",
			Instructions:  "You generate images from the user's text prompt. Prefer pixel-art style (limited palette, chunky pixels) for characters, items, or scenes. Do not answer with long prose; the product generates a picture for the prompt.",
			SystemDefined: true,
			SortOrder:     60,
			Enabled:       true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
}

func dorphAIAgentKey(id string) []byte {
	return []byte(fmt.Sprintf("%s%s", dorphAIAgentPrefix, id))
}

func legacyAIAgentKey(id string) []byte {
	return []byte(fmt.Sprintf("%s%s", legacyAIAgentPrefix, id))
}

func countAIAgentKeys(txn *badger.Txn) int {
	count := 0
	for _, prefix := range []string{dorphAIAgentPrefix, legacyAIAgentPrefix} {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		it.Close()
	}
	return count
}

// EnsureDorphAIAgentsSeed writes default agents only if none are stored yet.
func (d *DB) EnsureDorphAIAgentsSeed() error {
	var count int
	_ = d.badgerDB.View(func(txn *badger.Txn) error {
		count = countAIAgentKeys(txn)
		return nil
	})
	if count > 0 {
		return nil
	}
	defaultAgents := DefaultDorphAIAgents()
	for i := range defaultAgents {
		if err := d.PutDorphAIAgent(&defaultAgents[i]); err != nil {
			return err
		}
	}
	return nil
}

// EnsureDorphAISystemAgents seeds defaults when empty, then inserts any missing built-in agents by ID
// (so new system agents appear after upgrades without wiping custom ones).
func (d *DB) EnsureDorphAISystemAgents() error {
	if err := d.EnsureDorphAIAgentsSeed(); err != nil {
		return err
	}
	for _, def := range DefaultDorphAIAgents() {
		exists := false
		_ = d.badgerDB.View(func(txn *badger.Txn) error {
			_, err := txn.Get(dorphAIAgentKey(def.ID))
			if err == nil {
				exists = true
				return nil
			}
			_, err = txn.Get(legacyAIAgentKey(def.ID))
			if err == nil {
				exists = true
			}
			return nil
		})
		if exists {
			continue
		}
		a := def
		if err := d.PutDorphAIAgent(&a); err != nil {
			return err
		}
	}
	return nil
}

// ListDorphAIAgents returns agents sorted by SortOrder then Name.
func (d *DB) ListDorphAIAgents() ([]models.DorphAIAgent, error) {
	if err := d.EnsureDorphAISystemAgents(); err != nil {
		return nil, err
	}
	byID := map[string]models.DorphAIAgent{}
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		for _, prefix := range []string{legacyAIAgentPrefix, dorphAIAgentPrefix} {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = []byte(prefix)
			it := txn.NewIterator(opts)
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				var a models.DorphAIAgent
				if err := item.Value(func(val []byte) error {
					return json.Unmarshal(val, &a)
				}); err != nil {
					it.Close()
					return err
				}
				byID[a.ID] = a
			}
			it.Close()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	list := make([]models.DorphAIAgent, 0, len(byID))
	for _, a := range byID {
		list = append(list, a)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].SortOrder != list[j].SortOrder {
			return list[i].SortOrder < list[j].SortOrder
		}
		return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
	})
	return list, nil
}

// GetDorphAIAgent returns one agent by id.
func (d *DB) GetDorphAIAgent(id string) (*models.DorphAIAgent, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("missing id")
	}
	if err := d.EnsureDorphAISystemAgents(); err != nil {
		return nil, err
	}
	var out *models.DorphAIAgent
	err := d.badgerDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(dorphAIAgentKey(id))
		if err != nil {
			item, err = txn.Get(legacyAIAgentKey(id))
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			out = &models.DorphAIAgent{}
			return json.Unmarshal(val, out)
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PutDorphAIAgent creates or updates an agent (full replace).
func (d *DB) PutDorphAIAgent(a *models.DorphAIAgent) error {
	if a == nil {
		return fmt.Errorf("nil agent")
	}
	a.ID = strings.TrimSpace(a.ID)
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if a.CreatedAt == "" {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" {
		return fmt.Errorf("name required")
	}
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		_ = txn.Delete(legacyAIAgentKey(a.ID))
		return txn.Set(dorphAIAgentKey(a.ID), data)
	})
}

// DeleteDorphAIAgent removes a non-system-defined agent.
func (d *DB) DeleteDorphAIAgent(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("missing id")
	}
	a, err := d.GetDorphAIAgent(id)
	if err != nil {
		return err
	}
	if a.SystemDefined {
		return fmt.Errorf("cannot delete built-in agent")
	}
	return d.badgerDB.Update(func(txn *badger.Txn) error {
		_ = txn.Delete(legacyAIAgentKey(id))
		err := txn.Delete(dorphAIAgentKey(id))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		return err
	})
}
