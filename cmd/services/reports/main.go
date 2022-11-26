package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/database/postgres"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/utils"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/config"
	expCache "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/cache"
	expenseRepository "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/repository/postgres"
	expenseUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/usecase"
	exchangeRatesRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/repository/postgres"
	exrateUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/usecase"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/generated/proto/api"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/kafka"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/providers"
	userRepository "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user/repository/postgres"
	userUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultTimeout = 5 * time.Second
)

var (
	configPath = flag.String("config", "data/config.yaml", "Path to the config in YAML format.")
)

func readConfig(path string) (*config.Service, error) {
	rawYAML, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}
	return config.NewFromReader(bytes.NewReader(rawYAML))
}

func main() {
	// TODO: refactor copy-paste from cmd/bot/main
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := readConfig(*configPath)
	if err != nil {
		log.Fatal("Config init failed: ", err)
	}

	zapLogger, al, err := utils.SetupZapLogger(cfg.Values().DevLogger, cfg.Values().LogLevel)
	if err != nil {
		log.Fatal("Failed to setup zap logger: ", err)
	}
	_ = zap.ReplaceGlobals(zapLogger)
	defer func() {
		_ = zapLogger.Sync()
	}()
	if serviceName := cfg.Values().ServiceNameTracing; serviceName != "" {
		flusher, err := utils.InitTracing(serviceName, zapLogger)
		if err != nil {
			zapLogger.Fatal("Failed to init tracing", zap.Error(err))
		}
		defer func() {
			if err := flusher.Close(); err != nil {
				zapLogger.Error("Failed to flush tracing buffers", zap.Error(err))
			}
		}()
	}

	db, err := sql.Open("pgx", cfg.Values().DBConnectionString)
	if err != nil {
		zapLogger.Fatal("Failed to open db", zap.Error(err))
	}
	if err := db.PingContext(ctx); err != nil {
		zapLogger.Fatal("Failed to ping db", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			zapLogger.Error("Failed to close db", zap.Error(err))
		}
	}()
	if endpoint := cfg.Values().HTTPEndpoint; endpoint != "" {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.Handle("/log/level", al)
		s := &http.Server{Addr: endpoint, Handler: mux, ReadHeaderTimeout: defaultTimeout, ReadTimeout: defaultTimeout}
		endpointField := zap.String("endpoint", endpoint)
		go func() {
			zapLogger.Info("Starting HTTP API", endpointField)
			if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				zapLogger.Fatal("Failed to start HTTP API", zap.Error(err))
			}
			zapLogger.Info("HTTP API successfully stopped", endpointField)
		}()
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
			defer cancel()
			if err := s.Shutdown(ctx); err != nil {
				zapLogger.Error("Failed to shutdown HTTP API", zap.Error(err), endpointField)
			}
		}()
	}

	dbDoer := postgres.NewDBDoer(db)

	userRepo, err := userRepository.New(dbDoer)
	if err != nil {
		zapLogger.Fatal("Failed to create user repository", zap.Error(err))
	}
	userUC, err := userUseCase.New(userRepo)
	if err != nil {
		zapLogger.Fatal("Failed to create user usecase", zap.Error(err))
	}

	exrateRepo, err := exchangeRatesRepo.New(dbDoer)
	if err != nil {
		zapLogger.Fatal("Failed to create exchange rates repository", zap.Error(err))
	}
	exratesProvider, err := providers.NewExchangeRatesWebProvider(cfg.Values().BaseCurrency, cfg.Values().SupportedCurrencies)
	if err != nil {
		zapLogger.Fatal("Failed to create exchange rates web provider", zap.Error(err))
	}
	exrateUC, err := exrateUseCase.New(exrateRepo, exratesProvider)
	if err != nil {
		zapLogger.Fatal("Failed to create exchange rates usecase", zap.Error(err))
	}

	expRepo, err := expenseRepository.New(dbDoer)
	if err != nil {
		zapLogger.Fatal("Failed to create expenses repository", zap.Error(err))
	}
	// we use userUC and exrateUC here to do some interconnected business logic inside expenseUseCase instance
	var reportsCache *expCache.ReportsRedisCache
	if redisCfg := cfg.Values().RedisConfig; redisCfg != nil {
		redisDB := redis.NewClient(&redis.Options{
			Addr:     redisCfg.Address,
			Password: redisCfg.Password,
			DB:       redisCfg.DB,
		})
		if err := redisDB.Ping(ctx).Err(); err != nil {
			zapLogger.Fatal("Failed to ping redis", zap.Error(err))
		}
		defer func() {
			if err := redisDB.Close(); err != nil {
				zapLogger.Error("Failed to close redis", zap.Error(err))
			}
		}()
		reportsCache, err = expCache.NewReportsRedisCache(redisDB)
		if err != nil {
			zapLogger.Fatal("Failed to crete expenses reports cache", zap.Error(err))
		}
	}
	expUC, err := expenseUseCase.NewWithCache(cfg.Values().BaseCurrency, expRepo, userUC, exrateUC, reportsCache)
	if err != nil {
		zapLogger.Fatal("Failed to create expenses usecase", zap.Error(err))
	}

	kafkaConsumer, err := kafka.NewConfig().WithMetrics(
		prometheus.DefaultRegisterer, "kafka-consumer", cfg.Values().KafkaConfig.ConsumerGroup, 1*time.Second,
		func(err error) {
			zapLogger.Error("Failed to update kafka consumer metrics", zap.Error(err))
		},
	).BuildConsumerGroup(cfg.Values().KafkaConfig.Brokers, cfg.Values().KafkaConfig.ConsumerGroup)
	if err != nil {
		zapLogger.Fatal("Failed to build kafka consumer", zap.Error(err))
	}
	defer func() {
		if err := kafkaConsumer.Close(); err != nil {
			zapLogger.Error("Failed to close kafka consumer", zap.Error(err))
		}
	}()

	conn, err := grpc.Dial(cfg.Values().GRPCEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zapLogger.Fatal("Failed to open gRPC connection", zap.Error(err))
	}
	defer func() {
		if err := conn.Close(); err != nil {
			zapLogger.Error("Failed to close gRPC connection", zap.Error(err))
		}
	}()
	reporter := api.NewReportsServiceClient(conn)

	reportsHandler := NewReportsHandler(zapLogger, expUC, reporter)
	topics := []string{cfg.Values().KafkaConfig.ReportsTopic}

	for {
		select {
		default:
		case <-ctx.Done():
			return
		}
		err := kafkaConsumer.Consume(ctx, topics, reportsHandler)
		if err != nil {
			zapLogger.Fatal("Failed to start consuming kafka topic", zap.String("topic", cfg.Values().KafkaConfig.ReportsTopic))
		}
	}
}
