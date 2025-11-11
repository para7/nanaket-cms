package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/para7/nanaket-cms/internal/db"
)

// ArticleRepository defines the interface for article data access
type ArticleRepository interface {
	Create(ctx context.Context, userID int64, title, content string, publishedAt pgtype.Timestamp) (db.Article, error)
	GetByID(ctx context.Context, id int64) (db.Article, error)
	List(ctx context.Context) ([]db.Article, error)
	Update(ctx context.Context, id, userID int64, title, content string, publishedAt pgtype.Timestamp) (db.Article, error)
	Delete(ctx context.Context, id int64) error
}

// articleRepository implements ArticleRepository interface
type articleRepository struct {
	querier db.Querier
}

// NewArticleRepository creates a new instance of ArticleRepository
func NewArticleRepository(querier db.Querier) ArticleRepository {
	return &articleRepository{
		querier: querier,
	}
}

// Create creates a new article
func (r *articleRepository) Create(ctx context.Context, userID int64, title, content string, publishedAt pgtype.Timestamp) (db.Article, error) {
	return r.querier.CreateArticle(ctx, db.CreateArticleParams{
		UserID:      userID,
		Title:       title,
		Content:     content,
		PublishedAt: publishedAt,
	})
}

// GetByID retrieves an article by ID
func (r *articleRepository) GetByID(ctx context.Context, id int64) (db.Article, error) {
	return r.querier.GetArticle(ctx, id)
}

// List retrieves all articles
func (r *articleRepository) List(ctx context.Context) ([]db.Article, error) {
	return r.querier.ListArticles(ctx)
}

// Update updates an article
func (r *articleRepository) Update(ctx context.Context, id, userID int64, title, content string, publishedAt pgtype.Timestamp) (db.Article, error) {
	return r.querier.UpdateArticle(ctx, db.UpdateArticleParams{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Content:     content,
		PublishedAt: publishedAt,
	})
}

// Delete deletes an article
func (r *articleRepository) Delete(ctx context.Context, id int64) error {
	return r.querier.DeleteArticle(ctx, id)
}
