package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"rsshub/internal/domain"

	"github.com/google/uuid"
)

type FeedRepository struct {
	db *DB
}

func NewFeedRepository(db *DB) *FeedRepository {
	return &FeedRepository{db: db}
}

func (r *FeedRepository) Create(ctx context.Context, feed *domain.Feed) error {
	query := `
		INSERT INTO feeds (id, created_at, updated_at, name, url)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.conn.ExecContext(ctx, query,
		feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.URL)
	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}
	return nil
}

func (r *FeedRepository) GetByName(ctx context.Context, name string) (*domain.Feed, error) {
	query := `
		SELECT id, created_at, updated_at, name, url, last_fetched_at
		FROM feeds WHERE name = $1
	`
	feed := &domain.Feed{}
	err := r.db.conn.QueryRowContext(ctx, query, name).Scan(
		&feed.ID, &feed.CreatedAt, &feed.UpdatedAt, &feed.Name, &feed.URL, &feed.LastFetchedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feed: %w", err)
	}
	return feed, nil
}

func (r *FeedRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Feed, error) {
	query := `
		SELECT id, created_at, updated_at, name, url, last_fetched_at
		FROM feeds WHERE id = $1
	`
	feed := &domain.Feed{}
	err := r.db.conn.QueryRowContext(ctx, query, id).Scan(
		&feed.ID, &feed.CreatedAt, &feed.UpdatedAt, &feed.Name, &feed.URL, &feed.LastFetchedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed: %w", err)
	}
	return feed, nil
}

func (r *FeedRepository) List(ctx context.Context, limit int) ([]*domain.Feed, error) {
	query := `
		SELECT id, created_at, updated_at, name, url, last_fetched_at
		FROM feeds ORDER BY created_at DESC LIMIT $1
	`
	rows, err := r.db.conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list feeds: %w", err)
	}
	defer rows.Close()

	return r.scanFeeds(rows)
}

func (r *FeedRepository) ListAll(ctx context.Context) ([]*domain.Feed, error) {
	query := `
		SELECT id, created_at, updated_at, name, url, last_fetched_at
		FROM feeds ORDER BY created_at DESC
	`
	rows, err := r.db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all feeds: %w", err)
	}
	defer rows.Close()

	return r.scanFeeds(rows)
}

func (r *FeedRepository) Delete(ctx context.Context, name string) error {
	query := `DELETE FROM feeds WHERE name = $1`
	result, err := r.db.conn.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("failed to delete feed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("feed not found: %s", name)
	}

	return nil
}

func (r *FeedRepository) Update(ctx context.Context, feed *domain.Feed) error {
	query := `
		UPDATE feeds 
		SET updated_at = $1, last_fetched_at = $2
		WHERE id = $3
	`
	_, err := r.db.conn.ExecContext(ctx, query, feed.UpdatedAt, feed.LastFetchedAt, feed.ID)
	if err != nil {
		return fmt.Errorf("failed to update feed: %w", err)
	}
	return nil
}

func (r *FeedRepository) GetMostOutdated(ctx context.Context, limit int) ([]*domain.Feed, error) {
	query := `
		SELECT id, created_at, updated_at, name, url, last_fetched_at
		FROM feeds 
		ORDER BY COALESCE(last_fetched_at, '1970-01-01'::timestamp) ASC
		LIMIT $1
	`
	rows, err := r.db.conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get outdated feeds: %w", err)
	}
	defer rows.Close()

	return r.scanFeeds(rows)
}

func (r *FeedRepository) scanFeeds(rows *sql.Rows) ([]*domain.Feed, error) {
	var feeds []*domain.Feed
	for rows.Next() {
		feed := &domain.Feed{}
		err := rows.Scan(&feed.ID, &feed.CreatedAt, &feed.UpdatedAt, &feed.Name, &feed.URL, &feed.LastFetchedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed: %w", err)
		}
		feeds = append(feeds, feed)
	}
	return feeds, rows.Err()
}
