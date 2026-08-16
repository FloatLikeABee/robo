package handler

import (
	"github.com/formsx/backend/internal/config"
	"github.com/formsx/backend/internal/mongo"
	"github.com/formsx/backend/internal/mysql"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

type Handler struct {
	Cfg                   *config.Config
	FormRepo              *mysql.FormRepo
	PageRepo              *mysql.PageRepo
	QuestionRepo          *mysql.QuestionRepo
	RuleRepo              *mysql.RuleRepo
	ResponseRepo          *mongo.ResponseRepo
	EventInfoRepo         *mongo.EventInfoRepo
	AIDocRepo             *mongo.AIDocumentRepo
	SurveyBotTemplateRepo *mongo.SurveyBotTemplateRepo
	SurveyBotResultRepo   *mongo.SurveyBotResultRepo
	AI                    *morphai.Client
}

func New(
	cfg *config.Config,
	formRepo *mysql.FormRepo,
	pageRepo *mysql.PageRepo,
	questionRepo *mysql.QuestionRepo,
	ruleRepo *mysql.RuleRepo,
	responseRepo *mongo.ResponseRepo,
	eventInfoRepo *mongo.EventInfoRepo,
	aiDocRepo *mongo.AIDocumentRepo,
	surveyBotTemplateRepo *mongo.SurveyBotTemplateRepo,
	surveyBotResultRepo *mongo.SurveyBotResultRepo,
) *Handler {
	aiClient := morphai.NewClient(morphai.LoadFromEnv())
	return &Handler{
		Cfg:                   cfg,
		FormRepo:              formRepo,
		PageRepo:              pageRepo,
		QuestionRepo:          questionRepo,
		RuleRepo:              ruleRepo,
		ResponseRepo:          responseRepo,
		EventInfoRepo:         eventInfoRepo,
		AIDocRepo:             aiDocRepo,
		SurveyBotTemplateRepo: surveyBotTemplateRepo,
		SurveyBotResultRepo:   surveyBotResultRepo,
		AI:                    aiClient,
	}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set("handler_instance", h)
		c.Next()
	})
	api.POST("/auth/login", h.Login)
	api.GET("/auth/me", h.Me)
	protected := api.Group("")
	protected.Use(requireWorkspaceAccess())
	protected.POST("/assistant/chat", h.AssistantChat)
	protected.GET("/ai/mcp-tools", h.FormsXMorphMiniMCP)
	protected.GET("/ai/app-abilities", h.FormsXAppAbilitiesMCP)
	protected.POST("/ai/web-search", h.WebSearch)
	protected.POST("/ai/form-template-chat", h.FormTemplateAIChat)
	protected.GET("/ai/mongodb-mcp", h.FormsXMongoMCPTools)
	protected.POST("/ai/mongodb-mcp/call", h.FormsXMongoMCPCall)
	// Survey Bot / AI Sheets
	protected.GET("/survey-bot/templates", h.ListSurveyBotTemplates)
	protected.POST("/survey-bot/templates", h.CreateSurveyBotTemplate)
	protected.POST("/survey-bot/templates/ai-draft", h.DraftSurveyBotTemplateAI)
	protected.POST("/survey-bot/templates/from-file", h.DraftSurveyBotTemplateFromFile)
	protected.POST("/survey-bot/templates/compile", h.CompileSurveyBotTemplate)
	protected.GET("/survey-bot/templates/:id", h.GetSurveyBotTemplate)
	protected.PUT("/survey-bot/templates/:id", h.UpdateSurveyBotTemplate)
	protected.DELETE("/survey-bot/templates/:id", h.DeleteSurveyBotTemplate)
	protected.POST("/survey-bot/templates/:id/compile", h.CompileAndSaveSurveyBotTemplate)
	protected.POST("/survey-bot/templates/:id/publish", h.PublishSurveyBotTemplate)
	protected.POST("/survey-bot/templates/:id/unpublish", h.UnpublishSurveyBotTemplate)
	protected.PUT("/survey-bot/templates/:id/published", h.SetSurveyBotTemplatePublished)
	protected.GET("/survey-bot/results", h.ListSurveyBotResults)
	protected.GET("/survey-bot/results/:id", h.GetSurveyBotResult)
	protected.GET("/survey-bot/results/:id/html", h.GetSurveyBotResultHTML)
	protected.DELETE("/survey-bot/results/:id", h.DeleteSurveyBotResult)
	// Forms
	protected.POST("/forms", h.CreateForm)
	protected.GET("/forms", h.ListForms)
	protected.GET("/forms/:id", h.GetForm)
	protected.GET("/forms/by-slug/:slug", h.GetFormBySlug)
	protected.PUT("/forms/:id", h.UpdateForm)
	protected.DELETE("/forms/:id", h.DeleteForm)
	// Events & Info (MongoDB only)
	protected.GET("/events-info", h.ListEventInfo)
	protected.POST("/events-info", h.CreateEventInfo)
	protected.POST("/events-info/ai-draft", h.DraftEventInfoAI)
	protected.POST("/events-info/ai-ingest", h.IngestEventInfoAI)
	protected.GET("/events-info/collection-info", h.GetEventInfoCollectionInfo)
	protected.POST("/events-info/share/email", h.SendEventInfoCollectionEmail)
	protected.GET("/events-info/:id/ai-context", h.GetEventInfoAIContext)
	protected.GET("/events-info/:id", h.GetEventInfo)
	protected.DELETE("/events-info/:id", h.DeleteEventInfo)
	// Pages
	protected.GET("/forms/:id/pages", h.ListPages)
	protected.POST("/forms/:id/pages", h.CreatePage)
	protected.PUT("/forms/:id/pages/:pageId", h.UpdatePage)
	protected.DELETE("/forms/:id/pages/:pageId", h.DeletePage)
	// Questions
	protected.POST("/forms/:id/questions", h.CreateQuestion)
	protected.GET("/forms/:id/questions", h.ListQuestions)
	protected.PUT("/forms/:id/questions/:questionId", h.UpdateQuestion)
	protected.DELETE("/forms/:id/questions/:questionId", h.DeleteQuestion)
	protected.POST("/forms/:id/questions/:questionId/prompt-media", h.UploadQuestionPromptMedia)
	protected.DELETE("/forms/:id/questions/:questionId/prompt-media", h.DeleteQuestionPromptMedia)
	// Rules (visibility per question)
	protected.GET("/forms/:id/rules", h.ListRulesByFormID)
	protected.GET("/forms/:id/questions/:questionId/rules", h.ListRules)
	protected.POST("/forms/:id/questions/:questionId/rules", h.CreateRule)
	protected.DELETE("/forms/:id/questions/:questionId/rules/:ruleId", h.DeleteRule)
	// Responses (owner)
	protected.GET("/forms/:id/responses", h.ListResponses)
	protected.GET("/forms/:id/responses/export", h.ExportResponses)
	protected.DELETE("/forms/:id/responses/:responseId", h.DeleteResponse)
	protected.POST("/forms/:id/responses/:responseId/email", h.EmailResponse)
	// Public
	pub := api.Group("/public")
	pub.GET("/forms/:slug", h.PublicGetForm)
	pub.POST("/forms/:slug/submit", h.SubmitResponse)
	pub.POST("/events-info", h.PublicCreateEventInfo)
	pub.GET("/ai-sheets/:slug", h.PublicGetAISheet)
	pub.POST("/ai-sheets/:slug/chat", h.PublicAISheetChat)
}
