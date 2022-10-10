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

// TODO: rid of it and fetch data from config
const defaultCurrencyStub = "RUB"

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
	// TODO: initialize value
	var provider providers.ExchangeRatesProvider
	exrateUC, err := exrateUseCase.New(exrateRepo, provider)
	if err != nil {
		log.Fatal("Failed to create exchange rates usecase:", err)
	}
	// TODO call exrateUC.RunAutoUpdater

	expRepo, err := expenseRepository.New()
	if err != nil {
		log.Fatal("Failed to create expenses repository:", err)
	}
	// we use userUC and exrateUC here to do some interconnected business logic inside expenseUseCase instance
	expUC, err := expenseUseCase.New(defaultCurrencyStub, expRepo, userUC, exrateUC)
	if err != nil {
		log.Fatal("Failed to create expenses usecase:", err)
	}
	cl, err := tg.NewWithOptions(cfg.Token(), defaultCurrencyStub, expUC, userUC, tg.Options{
		Logger:     log.Default(),
		LogUpdates: cfg.Values().LogUpdates,
		WhiteList:  cfg.Values().WhiteList,
		BlackList:  cfg.Values().BlackList,
	})
	if err != nil {
		log.Fatal("Failed to init bot:", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	go cl.Start(ctx)
	log.Println("Bot initialized successfully ans started")
	<-ctx.Done()
	log.Println("Stopping bot...")
	cl.Stop()
	log.Println("Bot successfully stopped")
}
