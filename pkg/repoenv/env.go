package repoenv

import (
	"os"
	"strings"
)

// Get returns the named environment variable.
//
// DORPH_, REACT_APP_DORPH_, and VITE_DORPH_ keys also read MORPH_,
// REACT_APP_MORPH_, and VITE_MORPH_ when the new name is empty.
// The Dorph name wins when both are set.
func Get(key string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	if alias := legacyEnvKey(key); alias != "" {
		return strings.TrimSpace(os.Getenv(alias))
	}
	return ""
}

// legacyEnvKey maps a DORPH_ key to the MORPH_ key it replaced.
func legacyEnvKey(key string) string {
	switch {
	case strings.HasPrefix(key, "REACT_APP_DORPH_"):
		return "REACT_APP_MORPH_" + strings.TrimPrefix(key, "REACT_APP_DORPH_")
	case strings.HasPrefix(key, "VITE_DORPH_"):
		return "VITE_MORPH_" + strings.TrimPrefix(key, "VITE_DORPH_")
	case strings.HasPrefix(key, "DORPH_"):
		return "MORPH_" + strings.TrimPrefix(key, "DORPH_")
	}
	return ""
}

func dorphKeyFromLegacy(key string) string {
	switch {
	case strings.HasPrefix(key, "REACT_APP_MORPH_"):
		return "REACT_APP_DORPH_" + strings.TrimPrefix(key, "REACT_APP_MORPH_")
	case strings.HasPrefix(key, "VITE_MORPH_"):
		return "VITE_DORPH_" + strings.TrimPrefix(key, "VITE_MORPH_")
	case strings.HasPrefix(key, "MORPH_"):
		return "DORPH_" + strings.TrimPrefix(key, "MORPH_")
	}
	return ""
}

// aliasLegacyEnv copies MORPH_ / REACT_APP_MORPH_ / VITE_MORPH_ names into the
// DORPH_ names when the Dorph name is empty. Call after loading .env so child
// processes that only read DORPH_* still see existing MORPH_* values.
func aliasLegacyEnv() {
	for _, e := range os.Environ() {
		key, val, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		newKey := dorphKeyFromLegacy(key)
		if newKey == "" {
			continue
		}
		if strings.TrimSpace(os.Getenv(newKey)) != "" {
			continue
		}
		_ = os.Setenv(newKey, val)
	}
}
