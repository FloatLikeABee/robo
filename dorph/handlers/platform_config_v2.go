package handlers

import (
	"encoding/json"
	"strings"
)

// ConfigDictItem is a single option for type pickers (code stored in MySQL, label from config).
type ConfigDictItem struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// platformFile is the on-disk / DB JSON shape (v2). Legacy: flat string map of label overrides only.
type platformFile struct {
	Labels        map[string]string                 `json:"labels,omitempty"`
	Dictionaries  map[string][]ConfigDictItem        `json:"dictionaries,omitempty"`
}

func defaultEntityDictionaries() map[string][]ConfigDictItem {
	return map[string][]ConfigDictItem{
		"employ_type": {
			{Code: "full_time", Label: "Full time"},
			{Code: "part_time", Label: "Part time"},
			{Code: "contractor", Label: "Contractor"},
		},
		"asset_type": {
			{Code: "bus", Label: "Bus"},
			{Code: "van", Label: "Van"},
			{Code: "car", Label: "Car"},
			{Code: "other", Label: "Other"},
		},
		"activity_type": {
			{Code: "standard", Label: "Standard route"},
			{Code: "field_trip", Label: "Field trip"},
			{Code: "sports", Label: "Sports / events"},
		},
		"facility_type": {
			{Code: "school", Label: "School"},
			{Code: "depot", Label: "Depot / yard"},
			{Code: "office", Label: "Office"},
		},
		"participant_type": {
			{Code: "student", Label: "Student"},
			{Code: "passenger", Label: "Passenger"},
			{Code: "chaperone", Label: "Chaperone"},
		},
	}
}

// mergeDictionaries: defaults first; stored entries replace by key when non-empty.
func mergeDictionaries(def, stored map[string][]ConfigDictItem) map[string][]ConfigDictItem {
	out := make(map[string][]ConfigDictItem, len(def))
	for k, v := range def {
		out[k] = append([]ConfigDictItem(nil), v...)
	}
	for k, v := range stored {
		if k == "" || len(v) == 0 {
			continue
		}
		clean := make([]ConfigDictItem, 0, len(v))
		for _, it := range v {
			c := strings.TrimSpace(it.Code)
			if c == "" {
				continue
			}
			lb := strings.TrimSpace(it.Label)
			if lb == "" {
				lb = c
			}
			clean = append(clean, ConfigDictItem{Code: c, Label: lb})
		}
		if len(clean) > 0 {
			out[k] = clean
		}
	}
	return out
}

// readPlatformConfigFromDB returns raw label overrides and dictionary overrides (before merging with server defaults).
func readPlatformConfigFromDB(raw string) (labels map[string]string, dictStored map[string][]ConfigDictItem) {
	labels = map[string]string{}
	dictStored = map[string][]ConfigDictItem{}
	if strings.TrimSpace(raw) == "" {
		return labels, dictStored
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &root); err == nil {
		if lbl, ok := root["labels"]; ok {
			_ = json.Unmarshal(lbl, &labels)
		}
		if d, ok := root["dictionaries"]; ok {
			_ = json.Unmarshal(d, &dictStored)
		}
		_, hasL := root["labels"]
		_, hasD := root["dictionaries"]
		if hasL || hasD {
			return labels, dictStored
		}
	}
	// Legacy v1: entire JSON is label key -> string
	var flat map[string]string
	if err := json.Unmarshal([]byte(raw), &flat); err == nil {
		for k, v := range flat {
			if v == "" || k == "labels" || k == "dictionaries" {
				continue
			}
			labels[k] = v
		}
	}
	return labels, dictStored
}
