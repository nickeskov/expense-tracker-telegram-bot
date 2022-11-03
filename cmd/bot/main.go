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

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/clients/tg"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/database/postgres"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/utils"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/config"
	expenseRepository "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/repository/postgres"
	expenseUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/usecase"
	exchangeRatesRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/repository/postgres"
	exrateUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/usecase"
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
	expUC, err := expenseUseCase.New(cfg.Values().BaseCurrency, expRepo, userUC, exrateUC)
	if err != nil {
		zapLogger.Fatal("Failed to create expenses usecase", zap.Error(err))
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
	go cl.Start(ctx)
	zapLogger.Info("Bot initialized successfully ans started")
	<-ctx.Done()
	zapLogger.Info("Stopping bot...")
	cl.Stop()
	zapLogger.Info("Bot successfully stopped")
}
