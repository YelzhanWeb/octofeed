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

	feedService := services.NewFeedService(feedRepo)
	articleService := services.NewArticleService(articleRepo)

	aggregatorService := services.NewAggregatorService(
		feedRepo,
		articleRepo,
		rssFetcher,
		cfg.GetDefaultInterval(),
		cfg.GetDefaultWorkersCount(),
	)

	handler := cli.NewHandler(feedService, articleService, aggregatorService, db)

	return handler.Run(os.Args)
}
