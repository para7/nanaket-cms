package usecase

import (
	"context"
	"database/sql"

	"github.com/para7/nanaket-cms/internal/db"
	"github.com/para7/nanaket-cms/internal/repository"
)

// ArticleUsecase defines the interface for article business logic
type ArticleUsecase interface {
	CreateArticle(ctx context.Context, userID int64, title, content string, publishedAt sql.NullString) (db.Article, error)
	GetArticle(ctx context.Context, id int64) (db.Article, error)
	ListArticles(ctx context.Context) ([]db.Article, error)
	UpdateArticle(ctx context.Context, id, userID int64, title, content string, publishedAt sql.NullString) (db.Article, error)
	DeleteArticle(ctx context.Context, id int64) error
}

// articleUsecase implements ArticleUsecase interface
type articleUsecase struct {
	repo repository.ArticleRepository
}

// NewArticleUsecase creates a new instance of ArticleUsecase
func NewArticleUsecase(repo repository.ArticleRepository) ArticleUsecase {
	return &articleUsecase{
		repo: repo,
	}
}

// CreateArticle creates a new article
func (u *articleUsecase) CreateArticle(ctx context.Context, userID int64, title, content string, publishedAt sql.NullString) (db.Article, error) {
	return u.repo.Create(ctx, userID, title, content, publishedAt)
}

// GetArticle retrieves an article by ID
func (u *articleUsecase) GetArticle(ctx context.Context, id int64) (db.Article, error) {
	return u.repo.GetByID(ctx, id)
}

// ListArticles retrieves all articles
func (u *articleUsecase) ListArticles(ctx context.Context) ([]db.Article, error) {
	return u.repo.List(ctx)
}

// UpdateArticle updates an article
func (u *articleUsecase) UpdateArticle(ctx context.Context, id, userID int64, title, content string, publishedAt sql.NullString) (db.Article, error) {
	return u.repo.Update(ctx, id, userID, title, content, publishedAt)
}

// DeleteArticle deletes an article
func (u *articleUsecase) DeleteArticle(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
