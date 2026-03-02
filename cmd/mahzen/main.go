package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"
	tsclient "github.com/typesense/typesense-go/v3/typesense"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/emirakts0/mahzen/internal/config"
	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/handler"
	"github.com/emirakts0/mahzen/internal/infra/ai"
	"github.com/emirakts0/mahzen/internal/infra/identity"
	"github.com/emirakts0/mahzen/internal/infra/postgres"
	"github.com/emirakts0/mahzen/internal/infra/storage"
	"github.com/emirakts0/mahzen/internal/infra/typesense"
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
	logger.Info("starting mahzen", "grpc_port", cfg.Server.GRPC.Port, "http_port", cfg.Server.HTTP.Port)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	infra := mustInitInfra(ctx, cfg, logger)
	defer infra.close()

	buildServers(ctx, cfg, infra, logger).run(ctx, logger)
}

// ---------------------------------------------------------------------------
// Infrastructure
// ---------------------------------------------------------------------------

// infrastructure groups all external clients and connections. A single
// defer infra.close() in main handles cleanup.
type infrastructure struct {
	pool          *pgxpool.Pool
	tsClient      *tsclient.Client
	objectStorage *storage.ObjectStore
	embedder      domain.Embedder
	summarizer    domain.Summarizer
	kratosClient  *identity.KratosClient
	hydraClient   *identity.HydraClient
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

	tsClient, err := typesense.NewClient(cfg.Typesense)
	if err != nil {
		logger.Error("failed to create typesense client", "error", err)
		os.Exit(1)
	}
	if err := typesense.EnsureCollections(ctx, tsClient); err != nil {
		logger.Error("failed to ensure typesense collections", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to typesense")

	s3Client, err := storage.NewS3Client(ctx, cfg.S3)
	if err != nil {
		logger.Error("failed to create s3 client", "error", err)
		os.Exit(1)
	}
	if err := storage.EnsureBucket(ctx, s3Client, cfg.S3.Bucket); err != nil {
		logger.Error("failed to ensure s3 bucket", "error", err)
		os.Exit(1)
	}
	logger.Info("connected to object storage")

	embedder, summarizer := ai.NewProvider(cfg.OpenAI)
	logger.Info("ai provider initialized")

	return &infrastructure{
		pool:          pool,
		tsClient:      tsClient,
		objectStorage: storage.NewObjectStorage(s3Client, cfg.S3.Bucket),
		embedder:      embedder,
		summarizer:    summarizer,
		kratosClient:  identity.NewKratosClient(cfg.Identity.Kratos),
		hydraClient:   identity.NewHydraClient(cfg.Identity.Hydra),
	}
}

// ---------------------------------------------------------------------------
// Servers
// ---------------------------------------------------------------------------

// servers holds the gRPC and HTTP gateway servers along with the listener.
type servers struct {
	grpc *grpc.Server
	http *http.Server
	lis  net.Listener
}

func buildServers(ctx context.Context, cfg *config.Config, infra *infrastructure, logger *slog.Logger) *servers {
	entryRepo := postgres.NewEntryRepository(infra.pool)
	tagRepo := postgres.NewTagRepository(infra.pool)

	entrySvc := service.NewEntryService(entryRepo, tagRepo, infra.objectStorage, typesense.NewIndexer(infra.tsClient), infra.embedder, cfg.Entry)
	tagSvc := service.NewTagService(tagRepo)
	searchSvc := service.NewSearchService(typesense.NewSearcher(infra.tsClient), infra.embedder)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			handler.LoggingInterceptor(logger),
			handler.RecoveryInterceptor(logger),
		),
	)
	handler.RegisterEntryServer(grpcServer, entrySvc)
	handler.RegisterTagServer(grpcServer, tagSvc)
	handler.RegisterSearchServer(grpcServer, searchSvc)
	reflection.Register(grpcServer)

	grpcAddr := fmt.Sprintf(":%d", cfg.Server.GRPC.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("failed to listen for grpc", "address", grpcAddr, "error", err)
		os.Exit(1)
	}

	gwMux := runtime.NewServeMux()
	if err := handler.RegisterGatewayHandlers(ctx, gwMux, grpcAddr, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}); err != nil {
		logger.Error("failed to register gateway handlers", "error", err)
		os.Exit(1)
	}

	return &servers{
		grpc: grpcServer,
		http: &http.Server{Addr: fmt.Sprintf(":%d", cfg.Server.HTTP.Port), Handler: gwMux},
		lis:  lis,
	}
}

func (s *servers) run(ctx context.Context, logger *slog.Logger) {
	errCh := make(chan error, 2)

	go func() {
		logger.Info("grpc server listening", "address", s.lis.Addr().String())
		errCh <- s.grpc.Serve(s.lis)
	}()
	go func() {
		logger.Info("http gateway listening", "address", s.http.Addr)
		errCh <- s.http.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		logger.Error("server error", "error", err)
	}

	logger.Info("shutting down...")
	s.grpc.GracefulStop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.http.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown error", "error", err)
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
