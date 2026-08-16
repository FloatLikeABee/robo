package api

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/academi/booki/internal/auth"
	"github.com/academi/booki/internal/handlers"
	"github.com/academi/booki/internal/middleware"
	"github.com/gin-gonic/gin"
	gocache "github.com/patrickmn/go-cache"
	"github.com/robo/morphai"
)

func NewRouter(db *sql.DB, jwtSecret string, access time.Duration, refresh time.Duration, cors string, appEnv string, usersPanelBaseURL string) *gin.Engine {
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(func(c *gin.Context) {
		allow := cors
		if strings.EqualFold(appEnv, "development") {
			if o := c.GetHeader("Origin"); o != "" {
				allow = o
			}
		}
		if allow == "" {
			allow = "*"
		}
		c.Header("Access-Control-Allow-Origin", allow)
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	cleanup := refresh
	if cleanup < time.Minute {
		cleanup = time.Minute
	}
	authSvc := &auth.Service{
		DB:                db,
		RefreshCache:      gocache.New(refresh, cleanup),
		Secret:            jwtSecret,
		Access:            access,
		RefreshTTL:        refresh,
		UsersPanelBaseURL: strings.TrimRight(usersPanelBaseURL, "/"),
	}
	h := &handlers.API{
		DB:                db,
		AI:                morphai.NewClient(morphai.LoadFromEnv()),
		UsersPanelBaseURL: strings.TrimRight(usersPanelBaseURL, "/"),
	}

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", authSvc.Register)
		v1.POST("/auth/login", authSvc.Login)
		v1.POST("/auth/platform-session", authSvc.PlatformSession)
		if strings.EqualFold(appEnv, "development") {
			v1.POST("/auth/dev-login", authSvc.DevLogin)
		}
		v1.POST("/auth/refresh", authSvc.Refresh)
		v1.POST("/auth/logout", authSvc.Logout)

		authed := v1.Group("")
		authed.Use(middleware.JWTAuth(jwtSecret))
		authed.Use(middleware.RequireBookiLedgerAccess())
		authed.GET("/auth/me", authSvc.Me)
		authed.POST("/assistant/chat", h.AssistantChat)

		authed.GET("/organization", h.GetOrganization)
		authed.PATCH("/organization", h.PatchOrganization)

		authed.GET("/dashboard", h.Dashboard)
		authed.GET("/accounting/glossary", h.AccountingGlossary)

		authed.GET("/accounts", h.ListAccounts)
		authed.POST("/journal-entries", h.CreateJournalEntry)
		authed.GET("/journal-entries", h.ListJournalEntries)
		authed.GET("/ledger", h.GetLedger)
		authed.GET("/reports/trial-balance", h.TrialBalance)
		authed.GET("/reports/profit-loss", h.ProfitLoss)
		authed.POST("/opening-balances", h.OpeningBalances)

		authed.GET("/customers", h.ListCustomers)
		authed.POST("/customers", h.CreateCustomer)

		authed.GET("/bookings", h.ListBookings)
		authed.GET("/bookings/:id", h.GetBooking)
		authed.POST("/bookings", h.CreateBooking)
		authed.PATCH("/bookings/:id", h.UpdateBooking)
		authed.DELETE("/bookings/:id", h.DeleteBooking)
		authed.PATCH("/bookings/:id/status", h.UpdateBookingStatus)
		authed.POST("/bookings/:id/post", h.PostBookingToLedger)

		authed.GET("/warehouses", h.ListWarehouses)
		authed.POST("/warehouses", h.CreateWarehouse)
		authed.GET("/warehouses/:id", h.GetWarehouse)
		authed.PATCH("/warehouses/:id", h.UpdateWarehouse)
		authed.DELETE("/warehouses/:id", h.DeleteWarehouse)

		authed.GET("/products", h.ListProducts)
		authed.POST("/products", h.CreateProduct)
		authed.GET("/products/:id", h.GetProduct)
		authed.PATCH("/products/:id", h.UpdateProduct)
		authed.DELETE("/products/:id", h.DeleteProduct)
		authed.GET("/warehouse/stock", h.WarehouseStock)
		authed.GET("/warehouse/movements", h.ListMovements)
		authed.POST("/warehouse/stock-in", h.StockIn)
		authed.POST("/warehouse/stock-out", h.StockOut)
		authed.POST("/warehouse/transfers", h.StockTransfer)

		authed.GET("/flow-log/entries", h.ListFlowLogEntries)
		authed.POST("/flow-log/entries", h.CreateFlowLogEntry)
		authed.PATCH("/flow-log/entries/:id", h.UpdateFlowLogEntry)
		authed.DELETE("/flow-log/entries/:id", h.DeleteFlowLogEntry)
		authed.GET("/flow-log/summary", h.FlowLogSummary)
		authed.POST("/flow-log/analyze", h.AnalyzeFlowLog)

		authed.GET("/assets", h.ListAssets)
		authed.POST("/assets", h.CreateAsset)
		authed.POST("/assets/sync-morph", h.SyncMorphAssets)

		authed.POST("/imports/csv", h.ImportCSV)
		authed.POST("/imports/json", h.ImportJSON)
		authed.POST("/imports/http", h.ImportHTTP)
		authed.POST("/imports/ledger", h.ImportLedgerFile)
		authed.GET("/imports/logs", h.ImportLogs)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "academi-ledger-api"})
	})

	return r
}

func Listen(addr string, db *sql.DB, cfg struct {
	Secret            string
	Access            time.Duration
	Refresh           time.Duration
	CORS              string
	AppEnv            string
	UsersPanelBaseURL string
}) error {
	r := NewRouter(db, cfg.Secret, cfg.Access, cfg.Refresh, cfg.CORS, cfg.AppEnv, cfg.UsersPanelBaseURL)
	log.Printf("listening %s", addr)
	return r.Run(addr)
}

func NormalizeAddr(port string) string {
	if strings.HasPrefix(port, ":") {
		return "0.0.0.0" + port
	}
	if !strings.Contains(port, ":") {
		return "0.0.0.0:" + port
	}
	return port
}
