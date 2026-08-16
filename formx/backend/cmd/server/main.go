// @title SheetX API
// @version 1.0
// @description Form builder and response collection API
// @BasePath /
package main

import (
	"log"
	"os"

	_ "github.com/formsx/backend/docs"
	"github.com/formsx/backend/internal/config"
	"github.com/formsx/backend/internal/handler"
	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/mongo"
	"github.com/formsx/backend/internal/mysql"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @host localhost:29909
func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Printf("warning: upload dir: %v", err)
	}

	db, err := mysql.NewDB(cfg)
	if err != nil {
		log.Fatalf("sqlite: %v", err)
	}
	store, err := mongo.NewStore(cfg)
	if err != nil {
		log.Fatalf("badger: %v", err)
	}
	defer store.Close()

	responseRepo := mongo.NewResponseRepo(store)
	eventRepo := mongo.NewEventInfoRepo(store)
	aiDocRepo := mongo.NewAIDocumentRepo(store)
	surveyTplRepo := mongo.NewSurveyBotTemplateRepo(store)
	surveyResRepo := mongo.NewSurveyBotResultRepo(store)

	formRepo := mysql.NewFormRepo(db)
	pageRepo := mysql.NewPageRepo(db)
	questionRepo := mysql.NewQuestionRepo(db)
	ruleRepo := mysql.NewRuleRepo(db)
	h := handler.New(cfg, formRepo, pageRepo, questionRepo, ruleRepo, responseRepo, eventRepo, aiDocRepo, surveyTplRepo, surveyResRepo)

	r := gin.Default()
	r.MaxMultipartMemory = models.MaxQuestionPromptVideoBytes + (1 << 20) // accommodate largest prompt video + overhead
	r.Use(corsMiddleware())
	r.Static("/uploads", cfg.UploadDir)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	h.Register(r.Group("/"))

	addr := ":" + cfg.ServerPort
	log.Printf("SheetX listening on %s (storage=embedded sqlite+badger)", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
