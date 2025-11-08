package services

import (
	"context"
	"fmt"

	"rsshub/internal/domain"
	"rsshub/internal/ports"
)

type FeedService struct {
	feedRepo ports.FeedRepository
}

func NewFeedService(feedRepo ports.FeedRepository) *FeedService {
	return &FeedService{
		feedRepo: feedRepo,
	}
}

func (s *FeedService) AddFeed(ctx context.Context, name, url string) error {
	if name == "" {
		return fmt.Errorf("feed name cannot be empty")
	}
	if url == "" {
		return fmt.Errorf("feed url cannot be empty")
	}

	feed := domain.NewFeed(name, url)
	return s.feedRepo.Create(ctx, feed)
}

func (s *FeedService) ListFeeds(ctx context.Context, limit int) ([]*domain.Feed, error) {
	if limit > 0 {
		return s.feedRepo.List(ctx, limit)
	}
	return s.feedRepo.ListAll(ctx)
}

func (s *FeedService) DeleteFeed(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("feed name cannot be empty")
	}
	return s.feedRepo.Delete(ctx, name)
}
