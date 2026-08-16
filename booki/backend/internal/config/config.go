package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	SQLitePath        string
	JWTSecret         string
	JWTAccessExpiry   time.Duration
	JWTRefreshExpiry  time.Duration
	AppEnv            string
	AppPort           string
	CORSOrigin        string
	UsersPanelBaseURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	accessMin, _ := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_MIN", "15"))
	refreshDays, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "7"))

	sec := getEnv("JWT_SECRET", "")
	if sec == "" || sec == "change-me-to-a-long-random-string" {
		if getEnv("APP_ENV", "development") == "production" {
			return nil, fmt.Errorf("JWT_SECRET must be set in production")
		}
		sec = "dev-insecure-secret-change-in-prod"
	}

	return &Config{
		SQLitePath:        getEnv("BOOKI_SQLITE_PATH", "./data/booki.sqlite"),
		JWTSecret:         sec,
		JWTAccessExpiry:   time.Duration(accessMin) * time.Minute,
		JWTRefreshExpiry:  time.Duration(refreshDays) * 24 * time.Hour,
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", "9095"),
		CORSOrigin:        getEnv("CORS_ORIGIN", "http://localhost:5174"),
		UsersPanelBaseURL: strings.TrimRight(getEnv("USERS_PANEL_BASE_URL", "http://127.0.0.1:5001"), "/"),
	}, nil
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
