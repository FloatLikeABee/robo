package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/academi/backend/internal/ai"
	"github.com/academi/backend/internal/auth"
	"github.com/academi/backend/internal/community"
	"github.com/academi/backend/internal/config"
	"github.com/academi/backend/internal/database"
	docs "github.com/academi/backend/internal/docs"
	"github.com/academi/backend/internal/chatsessions"
	notifications "github.com/academi/backend/internal/notifications"
)

func main() {
	cfg := config.Load()

	defer database.CloseDB()

	gin.SetMode(cfg.Server.Mode)

	r := gin.Default()

	// Allow any Origin while echoing it back (required when AllowCredentials is true).
	// Plain "*" + credentials is invalid per CORS and breaks browsers from other hosts/LAN devices.
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:        []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:        []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:       []string{"Content-Length"},
		AllowCredentials:    true,
		AllowPrivateNetwork: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "academi-backend"})
	})

	api := r.Group("/api/v1")
	{
		authSvc := auth.NewService(cfg)
		authSvc.RegisterRoutes(api)

		jwt := auth.JWTMiddleware(cfg)

		aiSvc := ai.NewService(cfg)
		// Public AI routes (no auth required for web app)
		aiSvc.RegisterRoutes(api)

		communitySvc := community.NewService()
		communityRoutes := api.Group("/community")
		communityRoutes.Use(jwt)
		communitySvc.RegisterRoutes(communityRoutes)

		docsSvc := docs.NewService(cfg.Server.UploadDir)
		docsRoutes := api.Group("/docs")
		// Public for local web app (list/create/read); add auth for production if needed
		docsSvc.RegisterRoutes(docsRoutes)

		chatsSvc := chatsessions.NewService()
		chatsSvc.RegisterRoutes(api)

		notifSvc := notifications.NewService()
		notifRoutes := api.Group("/notifications")
		notifRoutes.Use(jwt)
		notifSvc.RegisterRoutes(notifRoutes)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Academi backend starting on :%s", cfg.Server.Port)
		if err := r.Run(":" + cfg.Server.Port); err != nil {
			log.Fatal("Server failed:", err)
		}
	}()

	<-stop
	log.Println("Shutting down...")
}
