package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Default platform labels (MorphData + generic module names). DB stores overrides only.
func platformUILabelDefaults() map[string]string {
	return map[string]string{
		"product_name":               "MorphData",
		"ai_assistant_name":          "Morph AI",
		"nav_districts_facilities":   "Places",
		"nav_people":                 "People",
		"nav_assets":                 "Assets",
		"nav_activities":             "Activities",
		"nav_generic_data":           "Generic data",
		"nav_big_notes":              "Big notes",
		"nav_user_settings":          "User settings",
		"nav_display_labels":         "Display names",
		"term_facility":              "Place",
		"term_facilities":            "Places",
		"col_facility_code":          "Code",
		"col_facility_name":          "Name",
		"col_facility_type":          "Type",
		"empty_districts_facilities": "No places",
		"col_activity_days":          "Activity days",
		"col_linked_asset_id":        "Asset ID",
	}
}

func mergeUILabels(defaults, overrides map[string]string) map[string]string {
	out := make(map[string]string, len(defaults)+len(overrides))
	for k, v := range defaults {
		out[k] = v
	}
	for k, v := range overrides {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out[k] = v
	}
	return out
}

// Legacy keys: page_* duplicated nav_*; migrate into nav if nav not overridden.
// Also fold old Participants/Employees nav keys into nav_people.
func migrateLegacyPageLabelOverrides(overrides map[string]string) {
	pairs := [][2]string{
		{"page_participants", "nav_people"},
		{"page_employees", "nav_people"},
		{"page_assets", "nav_assets"},
		{"page_activities", "nav_activities"},
		{"nav_participants", "nav_people"},
		{"nav_employees", "nav_people"},
		{"nav_quick_sheets", "nav_big_notes"},
	}
	for _, p := range pairs {
		pv, ok := overrides[p[0]]
		if !ok || strings.TrimSpace(pv) == "" {
			delete(overrides, p[0])
			continue
		}
		if _, hasNav := overrides[p[1]]; !hasNav {
			overrides[p[1]] = strings.TrimSpace(pv)
		}
		delete(overrides, p[0])
	}
}

func sanitizeUILabelOverrides(overrides map[string]string, defaults map[string]string) {
	for k := range overrides {
		if _, ok := defaults[k]; !ok {
			delete(overrides, k)
		}
	}
}

// GetPlatformUiConfig returns merged display labels, merged entity dictionaries, and label overrides for the editor.
func (h *Handlers) GetPlatformUiConfig(c *gin.Context) {
	defaults := platformUILabelDefaults()
	overrides := map[string]string{}
	if h.TranMySQL == nil {
		c.JSON(http.StatusOK, gin.H{
			"labels":       defaults,
			"overrides":    overrides,
			"dictionaries": mergeDictionaries(defaultEntityDictionaries(), nil),
		})
		return
	}
	var raw sql.NullString
	err := h.TranMySQL.DB.QueryRow(`SELECT ConfigJSON FROM PlatformUiConfig WHERE ID = 1`).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var rawStr string
	if err == nil && raw.Valid {
		rawStr = raw.String
	}
	overrides, dictStored := readPlatformConfigFromDB(rawStr)
	migrateLegacyPageLabelOverrides(overrides)
	sanitizeUILabelOverrides(overrides, defaults)
	merged := mergeUILabels(defaults, overrides)
	dicts := mergeDictionaries(defaultEntityDictionaries(), dictStored)
	c.JSON(http.StatusOK, gin.H{
		"labels":         merged,
		"overrides":      overrides,
		"dictionaries":   dicts,
		"dict_overrides": dictStored,
	})
}

// PutPlatformUiConfig saves label overrides and/or entity dictionaries.
func (h *Handlers) PutPlatformUiConfig(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in struct {
		Labels        map[string]string                  `json:"labels"`
		Dictionaries  map[string][]ConfigDictItem         `json:"dictionaries"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	defaults := platformUILabelDefaults()
	var raw sql.NullString
	var rawStr string
	err := h.TranMySQL.DB.QueryRow(`SELECT ConfigJSON FROM PlatformUiConfig WHERE ID = 1`).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err == nil && raw.Valid {
		rawStr = raw.String
	}
	existing, dictStored := readPlatformConfigFromDB(rawStr)
	migrateLegacyPageLabelOverrides(existing)
	if in.Labels != nil {
		for k, v := range in.Labels {
			if !isKnownUILabelKey(k, defaults) {
				continue
			}
			v = strings.TrimSpace(v)
			if v == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "label '" + k + "' cannot be empty"})
				return
			}
			if defaults[k] == v {
				delete(existing, k)
				continue
			}
			existing[k] = v
		}
	}
	if in.Dictionaries != nil {
		dictStored = in.Dictionaries
	}
	sanitizeUILabelOverrides(existing, defaults)
	f := platformFile{Labels: existing, Dictionaries: dictStored}
	blob, err := json.Marshal(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, err = h.TranMySQL.DB.Exec(
		`INSERT INTO PlatformUiConfig (ID, ConfigJSON) VALUES (1, ?)
		 ON DUPLICATE KEY UPDATE ConfigJSON = VALUES(ConfigJSON), UpdatedOn = CURRENT_TIMESTAMP`,
		string(blob),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	merged := mergeUILabels(defaults, existing)
	dicts := mergeDictionaries(defaultEntityDictionaries(), dictStored)
	c.JSON(http.StatusOK, gin.H{
		"labels":         merged,
		"overrides":      existing,
		"dictionaries":   dicts,
		"dict_overrides": dictStored,
	})
}

func isKnownUILabelKey(k string, defaults map[string]string) bool {
	if k == "" {
		return false
	}
	_, ok := defaults[k]
	return ok
}
