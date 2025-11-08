package postgres

import (
	"context"
	"fmt"

	"rsshub/internal/domain"

	"github.com/google/uuid"
)

type ArticleRepository struct {
	db *DB
}

func NewArticleRepository(db *DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(ctx context.Context, article *domain.Article) error {
	query := `
		INSERT INTO articles (id, created_at, updated_at, title, link, published_at, description, feed_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (link, feed_id) DO NOTHING
	`
	_, err := r.db.conn.ExecContext(ctx, query,
		article.ID, article.CreatedAt, article.UpdatedAt,
		article.Title, article.Link, article.PublishedAt,
		article.Description, article.FeedID)
	if err != nil {
		return fmt.Errorf("failed to create article: %w", err)
	}
	return nil
}

func (r *ArticleRepository) CreateBatch(ctx context.Context, articles []*domain.Article) error {
	if len(articles) == 0 {
		return nil
	}

	tx, err := r.db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO articles (id, created_at, updated_at, title, link, published_at, description, feed_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (link, feed_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, article := range articles {
		_, err := stmt.ExecContext(ctx,
			article.ID, article.CreatedAt, article.UpdatedAt,
			article.Title, article.Link, article.PublishedAt,
			article.Description, article.FeedID)
		if err != nil {
			return fmt.Errorf("failed to insert article: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *ArticleRepository) GetByFeedName(ctx context.Context, feedName string, limit int) ([]*domain.Article, error) {
	query := `
		SELECT a.id, a.created_at, a.updated_at, a.title, a.link, a.published_at, a.description, a.feed_id
		FROM articles a
		INNER JOIN feeds f ON a.feed_id = f.id
		WHERE f.name = $1
		ORDER BY COALESCE(a.published_at, a.created_at) DESC
		LIMIT $2
	`
	rows, err := r.db.conn.QueryContext(ctx, query, feedName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}
	defer rows.Close()

	var articles []*domain.Article
	for rows.Next() {
		article := &domain.Article{}
		err := rows.Scan(
			&article.ID, &article.CreatedAt, &article.UpdatedAt,
			&article.Title, &article.Link, &article.PublishedAt,
			&article.Description, &article.FeedID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, rows.Err()
}

func (r *ArticleRepository) Exists(ctx context.Context, link string, feedID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM articles WHERE link = $1 AND feed_id = $2)`
	var exists bool
	err := r.db.conn.QueryRowContext(ctx, query, link, feedID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check article existence: %w", err)
	}
	return exists, nil
}
