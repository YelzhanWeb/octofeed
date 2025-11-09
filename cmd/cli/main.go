package main

import (
	"fmt"
	"log"
	"os"

	"rsshub/internal/adapters/cli"
	"rsshub/internal/adapters/http"
	"rsshub/internal/adapters/postgres"
	"rsshub/internal/config"
	"rsshub/internal/core/services"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.NewEnvConfig()

	db, err := postgres.NewDB(cfg.GetDSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	feedRepo := postgres.NewFeedRepository(db)
	articleRepo := postgres.NewArticleRepository(db)
	rssFetcher := http.NewRSSFetcher()
	ipcLock := postgres.NewIPCLock(db)

	feedService := services.NewFeedService(feedRepo)
	articleService := services.NewArticleService(articleRepo)

	aggregatorService := services.NewAggregatorService(
		feedRepo,
		articleRepo,
		rssFetcher,
		ipcLock,
		cfg.GetDefaultInterval(),
		cfg.GetDefaultWorkersCount(),
	)

	handler := cli.NewHandler(feedService, articleService, aggregatorService, db)

	if len(os.Args) > 1 && os.Args[1] == "fetch" {
		log.Println("Running migrations...")
		if err := db.Up(); err != nil {
			log.Printf("Warning: migrations failed: %v\n", err)
		}
	}

	return handler.Run(os.Args)
}
