package handlers

import (
	"strings"

	"idongivaflyinfa/ai"
	"idongivaflyinfa/auth"
	"idongivaflyinfa/db"
	"idongivaflyinfa/hybridcontext"

	"github.com/gin-gonic/gin"
)

// @title           MorphData API
// @version         1.0
// @description     Morph AI — AI chat, forms, and MorphData admin APIs (embedded SQLite + Badger).

// @contact.name   API Support

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:9090
// @BasePath  /

// @schemes   http https

// Handlers contains all handler dependencies
type Handlers struct {
	db                *db.DB
	aiService         *ai.AIService
	externalAPIBase   string
	usersPanelBaseURL string // deprecated: messaging stub only; auth is local
	jwtCfg            auth.TokenConfig
	TranMySQL         *db.TranSQL
	EntityDetails     db.EntityDetailStore
	// TranMongo is kept as a deprecated alias field name used by older call sites;
	// prefer EntityDetails. Populated when EntityDetails is a *db.TranMongo.
	TranMongo             *db.TranMongo
	hybridStore           *hybridcontext.Store
	sharpReportBase       string
	tranFormBase          string
	tranMailBase          string
	bookiBase             string
	entityAttachmentMax     int
	entityAttachmentRootDir string
	importJobs              *importJobStore
	// ginEngine is the fully registered router; used to execute /api/tran and /api/forms calls from the AI assistant.
	ginEngine *gin.Engine
}

// New creates a new Handlers instance.
func New(database *db.DB, aiService *ai.AIService, externalAPIBase, usersPanelBaseURL string, tranSQL *db.TranSQL, entityDetails db.EntityDetailStore, sharpReportBaseURL, tranFormBaseURL, tranMailBaseURL, bookiBaseURL string, entityAttachmentMax int, entityAttachmentDir string) *Handlers {
	var legacyMongo *db.TranMongo
	if m, ok := entityDetails.(*db.TranMongo); ok {
		legacyMongo = m
	}
	return &Handlers{
		db:                      database,
		aiService:               aiService,
		externalAPIBase:         externalAPIBase,
		usersPanelBaseURL:       usersPanelBaseURL,
		jwtCfg:                  auth.LoadTokenConfig(),
		TranMySQL:               tranSQL,
		EntityDetails:           entityDetails,
		TranMongo:               legacyMongo,
		hybridStore:             hybridcontext.NewStore(),
		sharpReportBase:         strings.TrimSuffix(strings.TrimSpace(sharpReportBaseURL), "/"),
		tranFormBase:            strings.TrimSuffix(strings.TrimSpace(tranFormBaseURL), "/"),
		tranMailBase:            strings.TrimSuffix(strings.TrimSpace(tranMailBaseURL), "/"),
		bookiBase:               strings.TrimSuffix(strings.TrimSpace(bookiBaseURL), "/"),
		entityAttachmentMax:     entityAttachmentMax,
		entityAttachmentRootDir: strings.TrimSpace(entityAttachmentDir),
		importJobs:              newImportJobStore(),
	}
}
