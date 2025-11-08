package ports

import (
	"context"

	"rsshub/internal/domain"
)

type RSSFetcher interface {
	Fetch(ctx context.Context, url string) (*domain.RSSFeed, error)
}
