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

	handler := cli.NewHandler(feedService, articleService, aggregatorService)

	return runCommands(os.Args, handler, db)
}

func runCommands(args []string, h *cli.Handler, db *postgres.DB) error {
	if len(args) < 2 {
		return h.ShowHelp()
	}

	command := args[1]

	switch command {
	case "migrate-up":
		return db.RunMigrations()
	case "migrate-down":
		return db.DownMigrations()
	case "fetch":
		return h.HandleFetch()
	case "add":
		return h.HandleAdd(args[2:])
	case "set-interval":
		return h.HandleSetInterval(args[2:])
	case "set-workers":
		return h.HandleSetWorkers(args[2:])
	case "list":
		return h.HandleList(args[2:])
	case "delete":
		return h.HandleDelete(args[2:])
	case "articles":
		return h.HandleArticles(args[2:])
	case "--help", "-h", "help":
		return h.ShowHelp()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
