package services

import (
	"context"
	"fmt"

	"rsshub/internal/domain"
	"rsshub/internal/ports"
)

type ArticleService struct {
	articleRepo ports.ArticleRepository
}

func NewArticleService(articleRepo ports.ArticleRepository) *ArticleService {
	return &ArticleService{
		articleRepo: articleRepo,
	}
}

func (s *ArticleService) GetArticlesByFeed(ctx context.Context, feedName string, limit int) ([]*domain.Article, error) {
	if feedName == "" {
		return nil, fmt.Errorf("feed name cannot be empty")
	}
	if limit <= 0 {
		limit = 3
	}
	return s.articleRepo.GetByFeedName(ctx, feedName, limit)
}
