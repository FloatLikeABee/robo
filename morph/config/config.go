package config

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	GeminiAPIKey    string
	ModelName       string
	DBPath          string
	ExternalAPIBase string // Image reader, PDF reader, Gathering (e.g. http://localhost:8000)
	// Embedded Tran stores (default). Legacy MySQL/Mongo only for migrate / STORAGE_BACKEND=legacy.
	TranSQLitePath       string // e.g. ./data/tran.sqlite
	EntityDetailsBadger  string // e.g. ./data/entity_details
	StorageBackend       string // embedded (default) | legacy
	TranMySQLDSN         string // legacy / migrate source
	TranMongoURI         string // legacy / migrate source
	TranMongoDB          string // legacy / migrate source
	// Seed (cmd/seed_tran, full mode): optional; read from env / .env
	SeedTranCap         int  // default 50; max rows per trimmed entity type
	SeedTranSkipPrune bool // when true, skip delete/trim before seed
	// UsersPanelBaseURL is deprecated (Morph hosts auth). Kept for optional legacy messaging proxy.
	UsersPanelBaseURL string
	AdminEmail        string
	AdminUsername     string
	AdminPassword     string
	JWTSecret         string
	// HybridContext: optional cross-app bases (Bearer from Morph is forwarded). Empty disables each integration.
	SharpReportBaseURL string
	TranFormBaseURL    string
	TranMailBaseURL    string
	BookiBaseURL       string
	// Entity record attachments (documents, images, video)
	TranEntityAttachmentMax int    // max files per record; default 10
	TranEntityAttachmentDir string // base directory on disk; default uploads/entity_attachments
}

var dotOnce sync.Once

func loadDotEnvOnce() {
	dotOnce.Do(func() {
		// Does not override variables already set in the process environment.
		_ = godotenv.Load(".env")
	})
}

func GetConfig() Config {
	loadDotEnvOnce()
	return Config{
		Port: getEnv("PORT", "9090"),
		// MorphAI: MORPH_AI_* preferred; GEMINI_* kept for backward compatibility.
		GeminiAPIKey: firstEnv("MORPH_AI_API_KEY", "GEMINI_API_KEY", ""),
		ModelName:    firstEnv("MORPH_AI_MODEL", "GEMINI_MODEL", "qwen3-max"),
		DBPath:               getEnv("DB_PATH", "./data/badger"),
		ExternalAPIBase:      getEnv("EXTERNAL_API_BASE", "http://localhost:8000"),
		TranSQLitePath:       getEnv("TRAN_SQLITE_PATH", "./data/tran.sqlite"),
		EntityDetailsBadger:  getEnv("ENTITY_DETAILS_BADGER", "./data/entity_details"),
		StorageBackend:       strings.ToLower(strings.TrimSpace(getEnv("STORAGE_BACKEND", "embedded"))),
		TranMySQLDSN:         getEnv("TRAN_MYSQL_DSN", ""),
		TranMongoURI:         getEnv("TRAN_MONGO_URI", ""),
		TranMongoDB:          getEnv("TRAN_MONGO_DB", "athena"),
		SeedTranCap:          getEnvInt("SEED_TRAN_CAP", 50),
		SeedTranSkipPrune:    envTruthy("SEED_TRAN_SKIP_PRUNE"),
		UsersPanelBaseURL:    getEnv("USERS_PANEL_BASE_URL", ""),
		AdminEmail:           firstNonEmptyEnv("ADMIN_EMAIL", "BOOTSTRAP_ADMIN_EMAIL", "morphadmin@local.com"),
		AdminUsername:        firstNonEmptyEnv("ADMIN_USERNAME", "BOOTSTRAP_ADMIN_USERNAME", "morphadmin"),
		AdminPassword:        firstNonEmptyEnv("ADMIN_PASSWORD", "BOOTSTRAP_ADMIN_PASSWORD", "admin123"),
		JWTSecret:            getEnv("JWT_SECRET", "morph-dev-jwt-secret-change-me"),
		SharpReportBaseURL:   strings.TrimSuffix(getEnv("SHARPREPORT_BASE_URL", ""), "/"),
		TranFormBaseURL:      strings.TrimSuffix(getEnv("TRANFORM_BASE_URL", ""), "/"),
		TranMailBaseURL:      strings.TrimSuffix(getEnv("TRANMAIL_BASE_URL", ""), "/"),
		BookiBaseURL:         strings.TrimSuffix(getEnv("BOOKI_BASE_URL", ""), "/"),
		TranEntityAttachmentMax: getEnvInt("TRAN_ENTITY_ATTACHMENT_MAX", 10),
		TranEntityAttachmentDir: getEnv("TRAN_ENTITY_ATTACHMENT_DIR", "uploads/entity_attachments"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func firstEnv(keys ...string) string {
	for i := 0; i+1 < len(keys); i += 2 {
		if v := strings.TrimSpace(os.Getenv(keys[i])); v != "" {
			return v
		}
	}
	if len(keys)%2 == 1 {
		return keys[len(keys)-1]
	}
	return ""
}

// firstNonEmptyEnv tries each env key in order; the last arg is a literal default (not an env key).
func firstNonEmptyEnv(keysAndDefault ...string) string {
	if len(keysAndDefault) == 0 {
		return ""
	}
	def := keysAndDefault[len(keysAndDefault)-1]
	for _, k := range keysAndDefault[:len(keysAndDefault)-1] {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return def
}

func getEnvInt(key string, defaultValue int) int {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return defaultValue
	}
	return n
}

func envTruthy(key string) bool {
	s := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}
