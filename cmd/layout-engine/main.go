// Command layout-engine runs the Abeta Dynamic SDUI Layout Engine & Visual Studio server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/abeta-dev/go-libs/env"
	"github.com/abeta-dev/go-libs/ginmw"
	"github.com/abeta-dev/go-libs/logger"
	"github.com/abeta-dev/go-libs/shutdown"
	"github.com/abeta-dev/layout-engine/internal/engine"
	"github.com/abeta-dev/layout-engine/internal/handler"
	"github.com/abeta-dev/layout-engine/internal/infra/persistence"
	"github.com/abeta-dev/layout-engine/internal/usecase"
	"github.com/abeta-dev/layout-engine/pkg/transformer"
	"github.com/gin-gonic/gin"
)

type appConfig struct {
	ServiceName        string
	Addr               string
	GinMode            string
	DBDialect          string
	DBDSN              string
	AutoMigrate        bool
	SeedOnStart        bool
	AllowedOrigins     []string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
	TLSCertFile        string
	TLSKeyFile         string
	CompressionEnabled bool
	ETagEnabled        bool
	RateLimitEnabled   bool
	RequireAuth        bool
	JWTSecret          string
}

func loadConfig() appConfig {
	originsRaw := env.String("CORS_ALLOWED_ORIGINS", "*")
	parts := strings.Split(originsRaw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	port := env.Int("PORT", 8090)
	addr := env.String("HTTP_ADDR", fmt.Sprintf(":%d", port))

	return appConfig{
		ServiceName:        env.String("SERVICE_NAME", "abeta-layout-engine"),
		Addr:               addr,
		GinMode:            env.String("GIN_MODE", gin.ReleaseMode),
		DBDialect:          strings.ToLower(env.String("DB_DIALECT", string(persistence.DialectSQLite))),
		DBDSN:              env.String("DB_DSN", "file:layout-engine.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"),
		AutoMigrate:        env.Bool("DB_AUTO_MIGRATE", true),
		SeedOnStart:        env.Bool("SEED_DEFAULT_DATA", true),
		AllowedOrigins:     origins,
		ReadTimeout:        env.Duration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:       env.Duration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:        env.Duration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout:    env.Duration("SHUTDOWN_TIMEOUT", 10*time.Second),
		TLSCertFile:        env.String("TLS_CERT_FILE", ""),
		TLSKeyFile:         env.String("TLS_KEY_FILE", ""),
		CompressionEnabled: env.Bool("COMPRESSION_ENABLED", true),
		ETagEnabled:        env.Bool("ETAG_ENABLED", true),
		RateLimitEnabled:   env.Bool("RATE_LIMIT_ENABLED", false),
		RequireAuth:        env.Bool("AUTH_REQUIRED", false),
		JWTSecret:          env.String("JWT_SECRET", ""),
	}
}

func validateConfig(cfg appConfig) error {
	if (cfg.TLSCertFile == "") != (cfg.TLSKeyFile == "") {
		return errors.New("TLS_CERT_FILE and TLS_KEY_FILE must both be set when enabling TLS")
	}
	if cfg.RequireAuth && strings.TrimSpace(cfg.JWTSecret) == "" {
		return errors.New("AUTH_REQUIRED=true requires a non-empty JWT_SECRET")
	}
	switch persistence.Dialect(cfg.DBDialect) {
	case persistence.DialectSQLite, persistence.DialectPostgres:
	default:
		return fmt.Errorf("unsupported DB_DIALECT %q (expected sqlite or postgres)", cfg.DBDialect)
	}
	return nil
}

func run() error {
	cfg := loadConfig()
	if err := validateConfig(cfg); err != nil {
		return err
	}
	log := logger.Default().With(slog.String("service", cfg.ServiceName))
	gin.SetMode(cfg.GinMode)

	initCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	store, err := persistence.Open(initCtx, persistence.Config{
		Dialect:     persistence.Dialect(cfg.DBDialect),
		DSN:         cfg.DBDSN,
		AutoMigrate: cfg.AutoMigrate,
	})
	if err != nil {
		return fmt.Errorf("open persistence store: %w", err)
	}

	if cfg.SeedOnStart {
		if err := persistence.SeedDefaultData(initCtx, store); err != nil {
			_ = store.Close()
			return fmt.Errorf("seed default layout data: %w", err)
		}
	}

	evaluator := engine.NewRuleEvaluator()
	slotting := engine.NewSlottingEngine(evaluator)
	validator := engine.NewLayoutValidator()
	composer := transformer.NewPageComposer(transformer.NewDefaultRegistry())

	resolveSvc := usecase.NewResolveService(store, evaluator, slotting, composer)
	studioSvc := usecase.NewStudioService(store, validator, composer, engine.DiffJSON)
	actionSvc := usecase.NewActionService()

	router := handler.NewRouter(handler.RouterConfig{
		ServiceName:       cfg.ServiceName,
		Logger:            log,
		AllowedOrigins:    cfg.AllowedOrigins,
		Metrics:           &ginmw.MetricsTracker{},
		Store:             store,
		ResolveService:    resolveSvc,
		StudioService:     studioSvc,
		ActionService:     actionSvc,
		EnableCompression: cfg.CompressionEnabled,
		EnableETag:        cfg.ETagEnabled,
		EnableRateLimit:   cfg.RateLimitEnabled,
		RequireAuth:       cfg.RequireAuth,
		JWTSecret:         cfg.JWTSecret,
	})

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	protocols.SetHTTP2(true)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		Protocols:         protocols,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("starting Abeta Dynamic Layout Engine",
			slog.String("addr", cfg.Addr),
			slog.String("studio_url", "http://localhost"+cfg.Addr+"/studio"),
			slog.String("docs_url", "http://localhost"+cfg.Addr+"/docs"),
			slog.String("db_dialect", cfg.DBDialect),
		)
		var serveErr error
		if cfg.TLSCertFile != "" {
			serveErr = srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
		} else {
			serveErr = srv.ListenAndServe()
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
	}()

	mgr := shutdown.New(cfg.ShutdownTimeout)
	mgr.Register("http_server", func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})
	mgr.Register("persistence_store", func(_ context.Context) error {
		return store.Close()
	})

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- mgr.Wait()
	}()

	select {
	case err := <-errCh:
		sdCtx, sdCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer sdCancel()
		_ = mgr.Execute(sdCtx)
		return err
	case err := <-waitCh:
		if err != nil {
			return err
		}
	}

	log.Info("Abeta Layout Engine stopped cleanly")
	return nil
}

func main() {
	if err := run(); err != nil {
		logger.Default().Error("layout-engine exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}
