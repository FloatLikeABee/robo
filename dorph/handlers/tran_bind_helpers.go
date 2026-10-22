package handlers

import (
	"strconv"
	"strings"
)

func intFromAny(in map[string]interface{}, keys ...string) int {
	for _, k := range keys {
		if v, ok := in[k]; ok && v != nil {
			switch t := v.(type) {
			case float64:
				return int(t)
			case int:
				return t
			case string:
				n, _ := strconv.Atoi(strings.TrimSpace(t))
				return n
			}
		}
	}
	return 0
}

func boolFromAny(v interface{}) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case int:
		return t != 0
	case string:
		return strings.TrimSpace(strings.ToLower(t)) == "true" || t == "1" || t == "yes"
	}
	return false
}

func boolToTinyint(b bool) int {
	if b {
		return 1
	}
	return 0
}
