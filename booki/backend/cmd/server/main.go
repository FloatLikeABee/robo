package main

import (
	"log"
	"strings"
	"time"

	"github.com/academi/booki/internal/api"
	"github.com/academi/booki/internal/config"
	"github.com/academi/booki/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Connect(cfg.SQLitePath)
	if err != nil {
		log.Fatal("sqlite: ", err)
	}
	defer db.Close()
	log.Printf("sqlite connected %s", cfg.SQLitePath)

	if strings.EqualFold(cfg.AppEnv, "development") {
		log.Printf("development: sign in with your UsersPanel account (e.g. admin@local.com); USERS_PANEL_BASE_URL=%s", cfg.UsersPanelBaseURL)
	}

	addr := api.NormalizeAddr(cfg.AppPort)
	err = api.Listen(addr, db, struct {
		Secret            string
		Access            time.Duration
		Refresh           time.Duration
		CORS              string
		AppEnv            string
		UsersPanelBaseURL string
	}{cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry, cfg.CORSOrigin, cfg.AppEnv, cfg.UsersPanelBaseURL})
	if err != nil {
		log.Fatal(err)
	}
}
