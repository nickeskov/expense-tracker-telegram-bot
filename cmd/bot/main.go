package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/clients/tg"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/common/database/postgres"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/config"
	expenseRepository "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/repository/postgres"
	expenseUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/usecase"
	exchangeRatesRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/repository/postgres"
	exrateUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/usecase"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/providers"
	userRepository "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user/repository/postgres"
	userUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user/usecase"
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
		log.Fatal("Config init failed:", err)
	}

	db, err := sql.Open("pgx", cfg.Values().DBConnectionString)
	if err != nil {
		log.Fatal("Failed to open db: ", err)
	}
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Failed to ping db: ", err)
	}
	dbDoer := postgres.NewDBDoer(db)

	userRepo, err := userRepository.New(dbDoer)
	if err != nil {
		log.Fatal("Failed to create user repository:", err)
	}
	userUC, err := userUseCase.New(userRepo)
	if err != nil {
		log.Fatal("Failed to create user usecase:", err)
	}

	exrateRepo, err := exchangeRatesRepo.New(dbDoer)
	if err != nil {
		log.Fatal("Failed to create exchange rates repository:", err)
	}
	exratesProvider, err := providers.NewExchangeRatesWebProvider(cfg.Values().BaseCurrency, cfg.Values().SupportedCurrencies)
	if err != nil {
		log.Fatal("Failed to create exchange rates web provider:", err)
	}
	exrateUC, err := exrateUseCase.New(exrateRepo, exratesProvider)
	if err != nil {
		log.Fatal("Failed to create exchange rates usecase:", err)
	}

	expRepo, err := expenseRepository.New(dbDoer)
	if err != nil {
		log.Fatal("Failed to create expenses repository:", err)
	}
	// we use userUC and exrateUC here to do some interconnected business logic inside expenseUseCase instance
	expUC, err := expenseUseCase.New(cfg.Values().BaseCurrency, expRepo, userUC, exrateUC)
	if err != nil {
		log.Fatal("Failed to create expenses usecase:", err)
	}
	opts := tg.Options{
		Logger:     log.Default(),
		LogUpdates: cfg.Values().LogUpdates,
		WhiteList:  cfg.Values().WhiteList,
		BlackList:  cfg.Values().BlackList,
		Debug:      cfg.Values().Debug,
	}
	cl, err := tg.NewWithOptions(cfg.Token(), cfg.Values().BaseCurrency, cfg.Values().SupportedCurrencies, expUC, userUC, opts)
	if err != nil {
		log.Fatal("Failed to init bot:", err)
	}
	if interval := cfg.Values().ExchangeRatesUpdateInterval; interval != 0 {
		providerDone, err := exrateUC.RunAutoUpdater(ctx, log.Default(), interval)
		if err != nil {
			log.Fatal("Failed to run exchange rates auto updater:", err)
		}
		defer func() {
			<-providerDone
		}()
	}
	go cl.Start(ctx)
	log.Println("Bot initialized successfully ans started")
	<-ctx.Done()
	log.Println("Stopping bot...")
	cl.Stop()
	log.Println("Bot successfully stopped")
}
