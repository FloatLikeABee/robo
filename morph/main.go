package main

import (
	"context"
	"log"
	"strings"

	"idongivaflyinfa/ai"
	"idongivaflyinfa/cache"
	"idongivaflyinfa/config"
	"idongivaflyinfa/db"
	_ "idongivaflyinfa/docs" // Swagger docs
	"idongivaflyinfa/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// corsAllowOrigin returns a value for Access-Control-Allow-Origin.
// Browsers require the echo of the request Origin (not *) when preflight runs cross-origin.
// Local dev often uses different ports (e.g. React 3031, API 9090); reflect localhost/*.
func corsAllowOrigin(requestOrigin string) string {
	if requestOrigin == "" {
		return "*"
	}
	u := strings.ToLower(requestOrigin)
	if strings.HasPrefix(u, "http://localhost:") ||
		strings.HasPrefix(u, "http://127.0.0.1:") ||
		strings.HasPrefix(u, "https://localhost:") ||
		strings.HasPrefix(u, "https://127.0.0.1:") {
		return requestOrigin
	}
	return "*"
}

func main() {
	cfg := config.GetConfig()
	prod, err := config.ParseMorphEnv()
	if err != nil {
		log.Fatalf("refusing to start: %s", err.Error())
	}
	if err = config.ValidateStartup(cfg, prod); err != nil {
		log.Fatalf("refusing to start: %s", err.Error())
	}
	for _, w := range config.DevelopmentWarnings(cfg, prod) {
		log.Printf("warning: %s", w)
	}

	// Initialize Badger app database (forms, chat, etc.)
	database, err := db.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize in-process cache (replaces Redis)
	appCache := cache.New()

	// Initialize AI client
	aiService, err := ai.New(cfg.GeminiAPIKey, cfg.ModelName, appCache)
	if err != nil {
		log.Fatalf("Failed to initialize AI: %v", err)
	}
	defer aiService.Close()

	var tranSQL *db.TranSQL
	var entityDetails db.EntityDetailStore

	// Embedded stores only (SQLite + Badger). Use cmd/migrate_embedded to import from MySQL/Mongo.
	m, err := db.NewTranSQL(cfg.TranSQLitePath)
	if err != nil {
		log.Fatalf("Tran SQLite: %v", err)
	}
	tranSQL = m
	defer tranSQL.Close()
	log.Printf("Tran SQLite ready at %s", cfg.TranSQLitePath)

	if _, err = openSecretsBox(prod, cfg.TranSQLitePath); err != nil {
		log.Fatalf("%s", err.Error())
	}

	details, err := db.NewBadgerEntityDetails(cfg.EntityDetailsBadger)
	if err != nil {
		log.Fatalf("Entity details Badger: %v", err)
	}
	entityDetails = details
	defer details.Close()
	log.Printf("Entity details Badger ready at %s", cfg.EntityDetailsBadger)

	if err := tranSQL.EnsurePlatUsersTable(context.Background()); err != nil {
		if prod {
			log.Fatalf("refusing to start: plat_users schema: %v", err)
		}
		log.Printf("Warning: plat_users schema: %v", err)
	} else {
		if prod {
			rotated, err := tranSQL.GuardStoredDefaultAdminPassword(context.Background(), cfg.AdminEmail, cfg.AdminUsername, config.RotateDefaultAdmin(), cfg.AdminPassword)
			if err != nil {
				log.Fatalf("refusing to start: %s", err.Error())
			}
			if rotated > 0 {
				log.Printf("rotated %d account(s) off the development default admin password", rotated)
			}
		}
		if err := tranSQL.EnsureBootstrapAdmin(context.Background(), cfg.AdminEmail, cfg.AdminUsername, cfg.AdminPassword); err != nil {
			if prod {
				log.Fatalf("refusing to start: bootstrap admin: %v", err)
			}
			log.Printf("Warning: bootstrap admin: %v", err)
		} else {
			log.Printf("Auth ready (admin %s / %s)", cfg.AdminUsername, cfg.AdminEmail)
		}
		if !prod {
			if hit, err := tranSQL.HasStoredDefaultAdminPassword(context.Background(), cfg.AdminEmail, cfg.AdminUsername); err == nil && hit {
				log.Printf("warning: an admin account in the database still uses the development default password. Allowed because MORPH_ENV is not production. See docs/security-hosting-checklist.md")
			}
		}
	}
	if err := tranSQL.EnsurePlatInviteCodesTable(context.Background()); err != nil {
		log.Printf("Warning: plat_invite_codes schema: %v", err)
	}

	// Initialize handlers
	h := handlers.New(database, aiService, cfg.ExternalAPIBase, tranSQL, entityDetails,
		cfg.SharpReportBaseURL, cfg.TranFormBaseURL, cfg.TranMailBaseURL, cfg.BookiBaseURL,
		cfg.TranEntityAttachmentMax, cfg.TranEntityAttachmentDir)

	h.SeedBuiltinSkills()
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	h.StartNeo4jIngestWorker(workerCtx)
	h.StartResearchResume(workerCtx)

	// Setup Gin router
	r := gin.Default()

	// CORS: reflect localhost/127.0.0.1 origins for cross-port dev; otherwise *.
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allow := corsAllowOrigin(origin)
		c.Writer.Header().Set("Access-Control-Allow-Origin", allow)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD, CONNECT, TRACE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	r.Use(h.AuthzMiddleware())

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	handlers.RegisterAPIRoutes(r, h)

	// Serve form management UI
	r.Static("/presentation", "./presentation")
	r.GET("/forms", func(c *gin.Context) {
		c.File("./presentation/forms.html")
	})
	r.GET("/form-answers", func(c *gin.Context) {
		c.File("./presentation/form-answers.html")
	})

	// Serve static files (for React app)
	r.Static("/static", "./frontend/build/static")
	r.StaticFile("/", "./frontend/build/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/build/index.html")
	})

	h.SetGinEngine(r)

	log.Printf("Server starting on port %s (storage=embedded sqlite+badger)", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
