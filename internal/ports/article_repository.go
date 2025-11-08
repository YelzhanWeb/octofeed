package ports

import (
	"context"

	"rsshub/internal/domain"

	"github.com/google/uuid"
)

type ArticleRepository interface {
	Create(ctx context.Context, article *domain.Article) error
	CreateBatch(ctx context.Context, articles []*domain.Article) error
	GetByFeedName(ctx context.Context, feedName string, limit int) ([]*domain.Article, error)
	Exists(ctx context.Context, link string, feedID uuid.UUID) (bool, error)
}
