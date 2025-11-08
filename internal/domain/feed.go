package domain

import (
	"time"

	"github.com/google/uuid"
)

type Feed struct {
	ID            uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Name          string
	URL           string
	LastFetchedAt *time.Time
}

func NewFeed(name, url string) *Feed {
	now := time.Now()
	return &Feed{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
		URL:       url,
	}
}

func (f *Feed) MarkAsFetched() {
	now := time.Now()
	f.LastFetchedAt = &now
	f.UpdatedAt = now
}
