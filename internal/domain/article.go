package domain

import (
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       string
	Link        string
	PublishedAt *time.Time
	Description string
	FeedID      uuid.UUID
}

func NewArticle(title, link, description string, publishedAt *time.Time, feedID uuid.UUID) *Article {
	now := time.Now()
	return &Article{
		ID:          uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Title:       title,
		Link:        link,
		Description: description,
		PublishedAt: publishedAt,
		FeedID:      feedID,
	}
}
