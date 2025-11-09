package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"rsshub/internal/core/services"
	"rsshub/internal/ports"
)

type Handler struct {
	feedService    *services.FeedService
	articleService *services.ArticleService
	aggregator     ports.AggregatorPort
}

func NewHandler(
	feedService *services.FeedService,
	articleService *services.ArticleService,
	aggregator ports.AggregatorPort,
) *Handler {
	return &Handler{
		feedService:    feedService,
		articleService: articleService,
		aggregator:     aggregator,
	}
}

func (h *Handler) HandleFetch() error {
	if h.aggregator.IsRunning() {
		fmt.Println("Background process is already running")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.aggregator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start aggregator: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	return h.aggregator.Stop()
}

func (h *Handler) HandleAdd(args []string) error {
	name, url, err := parseAddFlags(args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if err := h.feedService.AddFeed(ctx, name, url); err != nil {
		return fmt.Errorf("failed to add feed: %w", err)
	}

	fmt.Printf("Feed '%s' added successfully\n", name)
	return nil
}

func (h *Handler) HandleSetInterval(args []string) error {
	duration, err := parseSetIntervalFlags(args)
	if err != nil {
		return err
	}

	if err := h.aggregator.SetInterval(duration); err != nil {
		return fmt.Errorf("failed to set interval: %w", err)
	}

	return nil
}

func (h *Handler) HandleSetWorkers(args []string) error {
	count, err := parseSetWorkersFlags(args)
	if err != nil {
		return err
	}

	if err := h.aggregator.Resize(count); err != nil {
		return fmt.Errorf("failed to resize workers: %w", err)
	}

	return nil
}

func (h *Handler) HandleList(args []string) error {
	num, err := parseListFlags(args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	feeds, err := h.feedService.ListFeeds(ctx, num)
	if err != nil {
		return fmt.Errorf("failed to list feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds available")
		return nil
	}

	fmt.Println("# Available RSS Feeds")
	fmt.Println()
	for i, feed := range feeds {
		fmt.Printf("%d. Name: %s\n", i+1, feed.Name)
		fmt.Printf("   URL: %s\n", feed.URL)
		fmt.Printf("   Added: %s\n", feed.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Println()
	}

	return nil
}

func (h *Handler) HandleDelete(args []string) error {
	name, err := parseDeleteFlags(args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if err := h.feedService.DeleteFeed(ctx, name); err != nil {
		return fmt.Errorf("failed to delete feed: %w", err)
	}

	fmt.Printf("Feed '%s' deleted successfully\n", name)
	return nil
}

func (h *Handler) HandleArticles(args []string) error {
	feedName, num, err := parseArticlesFlags(args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	articles, err := h.articleService.GetArticlesByFeed(ctx, feedName, num)
	if err != nil {
		return fmt.Errorf("failed to get articles: %w", err)
	}

	if len(articles) == 0 {
		fmt.Printf("No articles found for feed '%s'\n", feedName)
		return nil
	}

	fmt.Printf("Feed: %s\n\n", feedName)
	for i, article := range articles {
		var date string
		if article.PublishedAt != nil {
			date = article.PublishedAt.Format("2006-01-02")
		} else {
			date = article.CreatedAt.Format("2006-01-02")
		}

		fmt.Printf("%d. [%s] %s\n", i+1, date, article.Title)
		fmt.Printf("   %s\n\n", article.Link)
	}

	return nil
}

func (h *Handler) ShowHelp() error {
	help := `
Usage:
  rsshub COMMAND [OPTIONS]

Common Commands:
  add             add new RSS feed
  set-interval    set RSS fetch interval
  set-workers     set number of workers
  list            list available RSS feeds
  delete          delete RSS feed
  articles        show latest articles
  fetch           starts the background process that periodically fetches and processes RSS feeds using a worker pool

Examples:
  rsshub add --name "tech-crunch" --url "https://techcrunch.com/feed/"
  rsshub set-interval --duration 2m
  rsshub set-workers --count 5
  rsshub list --num 5
  rsshub delete --name "tech-crunch"
  rsshub articles --feed-name "tech-crunch" --num 5
  rsshub fetch
`
	fmt.Println(help)
	return nil
}
