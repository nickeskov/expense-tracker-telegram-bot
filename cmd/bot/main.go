package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/clients/tg"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/config"
	expenseRepository "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/repository/inmemory"
	expenseUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/usecase"
	exrateInMemRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/repository/inmemory"
	exrateUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/exrate/usecase"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/providers"
	userRepository "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user/repository/inmemory"
	userUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/user/usecase"
)

var (
	configPath = flag.String("config", "data/config.yaml", "Path to the config in YAML format.")
)

func main() {
	flag.Parse()
	cfg, err := config.New(*configPath)
	if err != nil {
		log.Fatal("Config init failed:", err)
	}
	userRepo, err := userRepository.New()
	if err != nil {
		log.Fatal("Failed to create user repository:", err)
	}
	userUC, err := userUseCase.New(userRepo)
	if err != nil {
		log.Fatal("Failed to create user usecase:", err)
	}

	exrateRepo, err := exrateInMemRepo.New()
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

	expRepo, err := expenseRepository.New()
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
	}
	cl, err := tg.NewWithOptions(cfg.Token(), cfg.Values().BaseCurrency, cfg.Values().SupportedCurrencies, expUC, userUC, opts)
	if err != nil {
		log.Fatal("Failed to init bot:", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
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
