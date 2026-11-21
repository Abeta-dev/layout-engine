package handler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/abeta-dev/go-libs/ginmw"
	"github.com/abeta-dev/go-libs/health"
	"github.com/abeta-dev/go-libs/logger"
	"github.com/abeta-dev/layout-engine/internal/domain"
	"github.com/abeta-dev/layout-engine/internal/usecase"
	"github.com/gin-gonic/gin"
)

// RouterConfig configures dependencies and middlewares for the Layout Engine HTTP server.
type RouterConfig struct {
	ServiceName       string
	Logger            *slog.Logger
	AllowedOrigins    []string
	Metrics           *ginmw.MetricsTracker
	Store             domain.Store
	ResolveService    *usecase.ResolveService
	StudioService     *usecase.StudioService
	ActionService     *usecase.ActionService
	EnableCompression bool
	EnableETag        bool
	EnableRateLimit   bool
	RequireAuth       bool
	JWTSecret         string
}

// NewRouter constructs the Gin HTTP router exposing Runtime SDUI, Studio CMS,
// embedded Visual Layout Studio (/studio), OpenAPI/Scalar (/docs), and observability endpoints.
func NewRouter(cfg RouterConfig) *gin.Engine {
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "abeta-layout-engine"
	}
	log := cfg.Logger
	if log == nil {
		log = logger.Default()
	}
	origins := cfg.AllowedOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	metricsTracker := cfg.Metrics
	if metricsTracker == nil {
		metricsTracker = ginmw.DefaultMetrics
	}

	healthSvc := health.New(
		health.WithStartTime(time.Now()),
		health.WithRuntimeInfo(),
		health.WithChecker("persistence_store", func(ctx context.Context) error {
			if cfg.Store == nil {
				return errors.New("persistence store not configured")
			}
			return cfg.Store.Ping(ctx)
		}),
	)

	engine := gin.New()
	engine.Use(
		ginmw.RequestID(),
		ginmw.Recovery(ginmw.WithRecoveryLogger(log)),
		ginmw.Telemetry(serviceName),
		ginmw.Logger(ginmw.WithLogger(log)),
		ginmw.SecurityHeaders(serviceName),
		ginmw.CORS(origins),
		ginmw.REDMetrics(metricsTracker),
		ginmw.LimitBodyDefault(),
	)
	if cfg.EnableRateLimit {
		engine.Use(ginmw.GlobalRateLimit())
	}
	if cfg.EnableCompression {
		engine.Use(CompressionMiddleware())
	}

	// Observability, Docs & Visual Layout Studio SPA.
	engine.GET("/healthz", gin.WrapF(health.LivenessHandler))
	engine.GET("/readyz", gin.WrapF(healthSvc.Handler))
	engine.GET("/metrics", ginmw.MetricsEndpoint(metricsTracker))
	registerOpenAPIRoutes(engine, serviceName)
	registerStudioUIRoutes(engine)

	withETag := func(h gin.HandlerFunc) []gin.HandlerFunc {
		if cfg.EnableETag {
			return []gin.HandlerFunc{ETagMiddleware(), h}
		}
		return []gin.HandlerFunc{h}
	}

	rt := &runtimeHandler{
		resolve: cfg.ResolveService,
		actions: cfg.ActionService,
	}
	st := &studioHandler{
		studio:  cfg.StudioService,
		resolve: cfg.ResolveService,
	}

	v1 := engine.Group("/api/v1")
	{
		// KMP-compatible Runtime SDUI API (drop-in for abeta-dev/kmp --bff-url)
		v1.GET("/screens", withETag(rt.listScreens)...)
		v1.GET("/screens/:screenId", withETag(rt.getScreen)...)
		v1.GET("/pages/:slug", withETag(rt.getPage)...)
		v1.GET("/config/splash", withETag(rt.getSplash)...)
		v1.GET("/catalog/tokens", withETag(rt.getTokenCatalog)...)
		v1.POST("/actions/execute", rt.executeAction)
		v1.POST("/orders/express", rt.placeExpressOrder)

		// Studio Authoring, Governance & Preview API
		studio := v1.Group("/studio")
		if cfg.RequireAuth {
			if err := ginmw.SetJWTSecret(cfg.JWTSecret); err != nil {
				log.Error("invalid JWT secret configured for studio auth", slog.Any("error", err))
			}
			studio.Use(ginmw.AuthMiddleware())
		}
		{
			studio.GET("/tenants", st.listTenants)
			studio.POST("/tenants", st.upsertTenant)
			studio.GET("/templates", st.listTemplates)

			studio.GET("/widgets", st.listWidgets)
			studio.POST("/widgets", st.upsertWidget)
			studio.DELETE("/widgets/:id", st.deleteWidget)

			studio.GET("/layouts", st.listLayouts)
			studio.GET("/layouts/:slug", st.getLayout)
			studio.PUT("/layouts/:slug", st.upsertLayout)
			studio.POST("/layouts/:slug/validate", st.validateLayout)
			studio.POST("/layouts/:slug/publish", st.publishLayout)
			studio.GET("/layouts/:slug/versions", st.listVersions)
			studio.POST("/layouts/:slug/rollback", st.rollbackLayout)
			studio.POST("/layouts/:slug/fallback", st.promoteFallback)
			studio.POST("/layouts/:slug/preview", st.previewLayout)

			studio.GET("/rules", st.listRules)
			studio.POST("/rules", st.upsertRule)
			studio.DELETE("/rules/:id", st.deleteRule)

			studio.GET("/slotting", st.listSlotting)
			studio.POST("/slotting", st.upsertSlotting)
			studio.DELETE("/slotting/:id", st.deleteSlotting)

			studio.GET("/splash", st.listSplash)
			studio.POST("/splash", st.upsertSplash)
			studio.GET("/tokens", st.listTokens)
			studio.POST("/tokens", st.upsertTokens)
			studio.GET("/changelog", st.listChangeLog)
		}
	}

	return engine
}
