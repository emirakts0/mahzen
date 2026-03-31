package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	"github.com/emirakts0/mahzen/internal/config"
	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/handler"
	"github.com/emirakts0/mahzen/internal/infra/ai"
	"github.com/emirakts0/mahzen/internal/infra/auth"
	"github.com/emirakts0/mahzen/internal/infra/meilisearch"
	"github.com/emirakts0/mahzen/internal/infra/postgres"
	"github.com/emirakts0/mahzen/internal/service"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Log)
	slog.SetDefault(logger)
	logger.Info("starting mahzen", "http_port", cfg.Server.HTTP.Port)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	infra := mustInitInfra(ctx, cfg, logger)
	defer infra.close()

	srv, accessTokenSvc := buildServer(cfg, infra, logger)

	// Load access token cache from DB before starting server.
	if err := accessTokenSvc.LoadCacheFromDB(ctx); err != nil {
		logger.Error("failed to load access token cache", "error", err)
		os.Exit(1)
	}

	// Start background expiry worker.
	accessTokenSvc.StartExpiryWorker(ctx)

	srv.run(ctx, logger)
}

// ---------------------------------------------------------------------------
// Infrastructure
// ---------------------------------------------------------------------------

// infrastructure groups all external clients and connections. A single
// defer infra.close() in main handles cleanup.
type infrastructure struct {
	pool             *pgxpool.Pool
	indexer          domain.Indexer
	searcher         domain.Searcher
	searchHealth     func(context.Context, time.Duration) (bool, error)
	searchDocCount   func(context.Context) (int64, error)
	embedder         domain.Embedder
	summarizer       domain.Summarizer
	tokenProvider    *auth.TokenProvider
	hasher           *auth.BcryptHasher
	accessTokenStore *auth.AccessTokenStore
}

func (i *infrastructure) close() {
	i.pool.Close()
}

func mustInitInfra(ctx context.Context, cfg *config.Config, logger *slog.Logger) *infrastructure {
	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to database")

	meilClient, err := meilisearch.NewClient(cfg.Meilisearch)
	if err != nil {
		logger.Error("failed to create meilisearch client", "error", err)
		os.Exit(1)
	}
	if err := meilisearch.EnsureIndex(ctx, meilClient); err != nil {
		logger.Error("failed to ensure meilisearch index", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to meilisearch")

	embedder, summarizer := ai.NewProvider(cfg.OpenAI)
	logger.Info("ai provider initialized")

	tokenProvider := auth.NewTokenProvider(cfg.Auth)
	hasher := auth.NewBcryptHasher()
	accessTokenStore := auth.NewAccessTokenStore()
	logger.Info("auth provider initialized")

	return &infrastructure{
		pool:     pool,
		indexer:  meilisearch.NewIndexer(meilClient),
		searcher: meilisearch.NewSearcher(meilClient),
		searchHealth: func(ctx context.Context, _ time.Duration) (bool, error) {
			health, err := meilClient.HealthWithContext(ctx)
			if err != nil {
				return false, err
			}
			return health.Status == "available", nil
		},
		searchDocCount: func(ctx context.Context) (int64, error) {
			stats, err := meilClient.Index(meilisearch.IndexName).GetStatsWithContext(ctx)
			if err != nil {
				return 0, err
			}
			return stats.NumberOfDocuments, nil
		},
		embedder:         embedder,
		summarizer:       summarizer,
		tokenProvider:    tokenProvider,
		hasher:           hasher,
		accessTokenStore: accessTokenStore,
	}
}

// ---------------------------------------------------------------------------
// Server
// ---------------------------------------------------------------------------

// server holds the HTTP server(s).
type server struct {
	httpServer *http.Server
	h3Server   *http3.Server // nil when TLS is not configured
	tlsEnabled bool
}

func buildServer(cfg *config.Config, infra *infrastructure, logger *slog.Logger) (*server, *service.AccessTokenService) {
	entryRepo := postgres.NewEntryRepository(infra.pool)
	tagRepo := postgres.NewTagRepository(infra.pool)
	userRepo := postgres.NewUserRepository(infra.pool)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(infra.pool)
	accessTokenRepo := postgres.NewAccessTokenRepository(infra.pool)

	entrySvc := service.NewEntryService(entryRepo, tagRepo, infra.indexer, infra.embedder)
	tagSvc := service.NewTagService(tagRepo)
	searchSvc := service.NewSearchService(infra.searcher, infra.embedder)
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, infra.tokenProvider, infra.hasher, infra.tokenProvider.RefreshTokenExpiry())

	defaultTTL := cfg.Auth.AccessTokenDefaultExpiry
	if defaultTTL == 0 {
		defaultTTL = 90 * 24 * time.Hour // 90 days
	}
	accessTokenSvc := service.NewAccessTokenService(accessTokenRepo, infra.tokenProvider, infra.accessTokenStore, defaultTTL)

	router := handler.SetupRouter(handler.RouterDeps{
		Logger:         logger,
		TokenGen:       infra.tokenProvider,
		EntrySvc:       entrySvc,
		TagSvc:         tagSvc,
		SearchSvc:      searchSvc,
		AuthSvc:        authSvc,
		AccessTokenSvc: accessTokenSvc,
		UserRepo:       userRepo,
		DBPing: func(ctx context.Context) error {
			return infra.pool.Ping(ctx)
		},
		DBCount: func(ctx context.Context) (int64, error) {
			return entryRepo.CountAll(ctx)
		},
		SearchEngineHealth:   infra.searchHealth,
		SearchEngineDocCount: infra.searchDocCount,
	})

	// Register SPA handler for non-API routes.
	registerSPARoutes(router)

	addr := fmt.Sprintf(":%d", cfg.Server.HTTP.Port)
	tlsEnabled := cfg.Server.HTTP.TLS.Enabled()

	httpServer := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	var h3Srv *http3.Server
	if tlsEnabled {
		tlsCert, err := tls.LoadX509KeyPair(cfg.Server.HTTP.TLS.CertFile, cfg.Server.HTTP.TLS.KeyFile)
		if err != nil {
			logger.Error("failed to load TLS certificates", "error", err)
			os.Exit(1)
		}

		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		}

		httpServer.TLSConfig = tlsCfg

		h3Srv = &http3.Server{
			Addr:      addr,
			Handler:   router,
			TLSConfig: tlsCfg,
		}

		// Advertise HTTP/3 via Alt-Svc header on HTTP/2 responses.
		origHandler := httpServer.Handler
		httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := h3Srv.SetQUICHeaders(w.Header()); err != nil {
				logger.Debug("failed to set QUIC headers", "error", err)
			}
			origHandler.ServeHTTP(w, r)
		})
	}

	return &server{
		httpServer: httpServer,
		h3Server:   h3Srv,
		tlsEnabled: tlsEnabled,
	}, accessTokenSvc
}

func (s *server) run(ctx context.Context, logger *slog.Logger) {
	errCh := make(chan error, 2)

	if s.tlsEnabled {
		go func() {
			logger.Info("https server listening (HTTP/2)", "address", s.httpServer.Addr)
			errCh <- s.httpServer.ListenAndServeTLS("", "")
		}()
		go func() {
			logger.Info("http/3 (QUIC) server listening", "address", s.h3Server.Addr)
			errCh <- s.h3Server.ListenAndServe()
		}()
	} else {
		go func() {
			logger.Info("http server listening", "address", s.httpServer.Addr)
			errCh <- s.httpServer.ListenAndServe()
		}()
	}

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		logger.Error("server error", "error", err)
	}

	logger.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown error", "error", err)
	}
	if s.h3Server != nil {
		if err := s.h3Server.Shutdown(shutdownCtx); err != nil {
			logger.Error("http/3 server shutdown error", "error", err)
		}
	}

	logger.Info("shutdown complete")
}

// ---------------------------------------------------------------------------
// Logging
// ---------------------------------------------------------------------------

func setupLogger(cfg config.LogConfig) *slog.Logger {
	var zlLevel zerolog.Level
	var slogLevel slog.Level
	switch cfg.Level {
	case "debug":
		zlLevel = zerolog.DebugLevel
		slogLevel = slog.LevelDebug
	case "warn":
		zlLevel = zerolog.WarnLevel
		slogLevel = slog.LevelWarn
	case "error":
		zlLevel = zerolog.ErrorLevel
		slogLevel = slog.LevelError
	default:
		zlLevel = zerolog.InfoLevel
		slogLevel = slog.LevelInfo
	}

	var zl zerolog.Logger
	switch cfg.Format {
	case "text":
		zl = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
			Level(zlLevel).
			With().Timestamp().Logger()
	default:
		zl = zerolog.New(os.Stdout).
			Level(zlLevel).
			With().Timestamp().Logger()
	}

	return slog.New(slogzerolog.Option{
		Level:  slogLevel,
		Logger: &zl,
	}.NewZerologHandler())
}
