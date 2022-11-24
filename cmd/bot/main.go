package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/clients/tg"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/database/postgres"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/utils"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/config"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense"
	expCache "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/cache"
	expenseRepository "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/repository/postgres"
	expenseUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/usecase"
	exchangeRatesRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/repository/postgres"
	exrateUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/usecase"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/grpc/reports"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/kafka"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/providers"
	userRepository "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user/repository/postgres"
	userUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user/usecase"
	"go.uber.org/zap"
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

	var expUC expense.UseCase
	regularExpUC, err := expenseUseCase.NewWithCache(cfg.Values().BaseCurrency, expRepo, userUC, exrateUC, reportsCache)
	if err != nil {
		zapLogger.Fatal("Failed to create expenses usecase", zap.Error(err))
	}
	if kafkaCfg := cfg.Values().KafkaConfig; kafkaCfg != nil {
		kafkaAsyncProducer, err := kafka.NewConfig().WithMetrics(
			prometheus.DefaultRegisterer, "kafka-producer", "", 1*time.Second,
			func(err error) {
				zapLogger.Error("Failed to update kafka consumer metrics", zap.Error(err))
			},
		).WithSuccessHandler(func(messages <-chan *sarama.ProducerMessage) {
			for message := range messages {
				zapLogger.Info("Message was sent to kafka",
					zap.String("topic", message.Topic),
					zap.Int64("offset", message.Offset),
					zap.Int32("partition", message.Partition),
				)
			}
		}).BuildAsyncProducer(cfg.Values().KafkaConfig.Brokers, func(producerErrors <-chan *sarama.ProducerError) {
			for err := range producerErrors {
				zapLogger.Error("Failed to produce message to kafka topic",
					zap.Error(err),
					zap.String("topic", err.Msg.Topic),
					zap.Int64("offset", err.Msg.Offset),
					zap.Int32("partition", err.Msg.Partition),
				)
			}
		})
		if err != nil {
			zapLogger.Fatal("Failed to build kafka async producer", zap.Error(err))
		}
		defer func() {
			if err := kafkaAsyncProducer.Close(); err != nil {
				zapLogger.Error("Failed to close kafka async producer", zap.Error(err))
			}
		}()
		extendedExpUC, err := expenseUseCase.NewExtendedUseCase(regularExpUC, cfg.Values().KafkaConfig.ReportsTopic, kafkaAsyncProducer)
		if err != nil {
			zapLogger.Fatal("Failed to create extended expenses usecase", zap.Error(err))
		}
		expUC = extendedExpUC
	} else {
		expUC = regularExpUC
	}

	opts := tg.Options{
		Logger:     zapLogger,
		LogUpdates: cfg.Values().LogUpdates,
		WhiteList:  cfg.Values().WhiteList,
		BlackList:  cfg.Values().BlackList,
		Debug:      cfg.Values().Debug,
	}
	cl, err := tg.NewWithOptions(cfg.Token(), cfg.Values().BaseCurrency, cfg.Values().SupportedCurrencies, expUC, userUC, opts)
	if err != nil {
		zapLogger.Fatal("Failed to init telegram bot", zap.Error(err))
	}
	if interval := cfg.Values().ExchangeRatesUpdateInterval; interval != 0 {
		providerDone, err := exrateUC.RunAutoUpdater(ctx, zapLogger, interval)
		if err != nil {
			zapLogger.Fatal("Failed to run exchange rates auto updater", zap.Error(err))
		}
		defer func() {
			<-providerDone
		}()
	}
	if grpcEndpoint := cfg.Values().GRPCEndpoint; grpcEndpoint != "" {
		reportsService, err := reports.NewService(cl, zapLogger)
		if err != nil {
			zapLogger.Fatal("Failed to create gRPC reports service", zap.Error(err))
		}
		l, err := net.Listen("tcp", grpcEndpoint)
		if err != nil {
			zapLogger.Fatal("Failed to open listener for gRPC", zap.Error(err), zap.String("address", grpcEndpoint))
		}
		defer func() {
			if err := l.Close(); err != nil {
				zapLogger.Error("Failed to close gRPC listener", zap.Error(err))
			}
		}()
		go func() {
			if err := reportsService.Serve(ctx, l); err != nil {
				zapLogger.Fatal("Failed to run serve on gRPC reports service", zap.Error(err))
			}
		}()
	}
	go cl.Start(ctx)
	zapLogger.Info("Bot initialized successfully ans started")
	<-ctx.Done()
	zapLogger.Info("Stopping bot...")
	cl.Stop()
	zapLogger.Info("Bot successfully stopped")
}
