package ports

import (
	"context"

	"rsshub/internal/domain"

	"github.com/google/uuid"
)

type FeedRepository interface {
	Create(ctx context.Context, feed *domain.Feed) error
	GetByName(ctx context.Context, name string) (*domain.Feed, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Feed, error)
	List(ctx context.Context, limit int) ([]*domain.Feed, error)
	ListAll(ctx context.Context) ([]*domain.Feed, error)
	Delete(ctx context.Context, name string) error
	Update(ctx context.Context, feed *domain.Feed) error
	GetMostOutdated(ctx context.Context, limit int) ([]*domain.Feed, error)
}
