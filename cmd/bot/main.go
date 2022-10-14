package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/clients/tg"
	"gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/config"
	expenseRepo "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/repository/inmemory"
	expenseUseCase "gitlab.ozon.dev/mr.eskov1/telegram-bot/internal/expense/usecase"
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
	repo, err := expenseRepo.New()
	if err != nil {
		log.Fatal("Failed to create expenses repository:", err)
	}
	useCase, err := expenseUseCase.New(repo)
	if err != nil {
		log.Fatal("Failed to create expenses usecase:", err)
	}
	cl, err := tg.NewWithOptions(cfg.Token(), useCase, tg.Options{
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
