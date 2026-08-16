package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

// Entity keys for entity detail documents (Badger / legacy Mongo).
const (
	entityKeyStudent = "student"
	entityKeyStaff   = "staff"
	// staffMongoAlternateEntity matches the MySQL table name (`employee`). Some seeds or ops may only have this key;
	// handlers read/write it alongside entityKeyStaff for the same record_id (= employee.id).
	staffMongoAlternateEntity = "employee"
	entityKeySchool           = "school"
	entityKeyVehicle          = "vehicle"
	entityKeyTrip             = "trip"
	entityKeyContact          = "contact"
	entityKeyDistrict         = "district"
	entityKeyCaseTask         = "case_task"
	entityKeyGenericData      = "generic_data"
)

func isEmptyMongoDetail(js string) bool {
	s := strings.TrimSpace(js)
	return s == "" || s == "{}"
}

func existingDetailFromRow(row map[string]interface{}) string {
	if row == nil {
		return ""
	}
	raw, ok := row["detail"]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(b))
	}
}

func (h *Handlers) entityDetailStore() db.EntityDetailStore {
	if h == nil {
		return nil
	}
	if h.EntityDetails != nil {
		return h.EntityDetails
	}
	return h.TranMongo
}

func (h *Handlers) attachEntityDetail(c *gin.Context, entity string, id int, row map[string]interface{}) {
	existing := existingDetailFromRow(row)
	store := h.entityDetailStore()
	if store == nil {
		if !isEmptyMongoDetail(existing) {
			row["detail"] = existing
		} else {
			row["detail"] = "{}"
		}
		return
	}
	ctx := c.Request.Context()
	js, err := store.GetEntityDetailJSON(ctx, entity, id)
	if err != nil {
		js = "{}"
	}
	if isEmptyMongoDetail(js) && !isEmptyMongoDetail(existing) {
		js = existing
	}
	if isEmptyMongoDetail(js) {
		row["detail"] = "{}"
		return
	}
	row["detail"] = js
}

func popDetailString(in map[string]interface{}) (string, bool, error) {
	if in == nil {
		return "", false, nil
	}
	var raw interface{}
	var ok bool
	if raw, ok = in["detail"]; !ok {
		raw, ok = in["Detail"]
	}
	if !ok || raw == nil {
		return "", false, nil
	}
	s, err := detailPayloadToJSONString(raw)
	if err != nil {
		return "", false, err
	}
	delete(in, "detail")
	delete(in, "Detail")
	return s, true, nil
}

func (h *Handlers) savePoppedDetail(c *gin.Context, entity string, id int, jsonText string) error {
	store := h.entityDetailStore()
	if store == nil {
		return nil
	}
	ctx := c.Request.Context()
	if err := store.SetEntityDetailJSON(ctx, entity, id, jsonText); err != nil {
		return err
	}
	if entity == entityKeyStaff {
		_ = store.SetEntityDetailJSON(ctx, staffMongoAlternateEntity, id, jsonText)
	}
	h.enqueueGraphEntity(ctx, entity, id, "upsert")
	return nil
}

func graphEntityType(entity string) string {
	switch entity {
	case entityKeyStudent:
		return "member"
	case entityKeyStaff, staffMongoAlternateEntity:
		return "employee"
	case entityKeySchool:
		return "facility"
	case entityKeyVehicle:
		return "asset"
	case entityKeyTrip:
		return "activity"
	default:
		return entity
	}
}

func (h *Handlers) enqueueGraphEntity(ctx context.Context, entity string, id int, op string) {
	if h.TranMySQL == nil || id <= 0 {
		return
	}
	_ = h.TranMySQL.EnqueueGraphSync(ctx, "morph", graphEntityType(entity), fmt.Sprintf("%d", id), op, "")
}

func detailPayloadToJSONString(raw interface{}) (string, error) {
	switch v := raw.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "{}", nil
		}
		if !json.Valid([]byte(v)) {
			return "", fmt.Errorf("detail must be valid JSON")
		}
		return v, nil
	case map[string]interface{}:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

func (h *Handlers) deleteEntityDetailMongo(ctx context.Context, entity string, id int) {
	store := h.entityDetailStore()
	if h == nil || store == nil {
		return
	}
	_ = store.DeleteEntityDetail(ctx, entity, id)
	if entity == entityKeyStaff {
		_ = store.DeleteEntityDetail(ctx, staffMongoAlternateEntity, id)
	}
	h.enqueueGraphEntity(ctx, entity, id, "delete")
}
