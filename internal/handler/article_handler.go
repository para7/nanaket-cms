package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/para7/nanaket-cms/internal/usecase"
)

// ArticleHandler handles HTTP requests for article operations
type ArticleHandler struct {
	usecase usecase.ArticleUsecase
}

// NewArticleHandler creates a new instance of ArticleHandler
func NewArticleHandler(usecase usecase.ArticleUsecase) *ArticleHandler {
	return &ArticleHandler{
		usecase: usecase,
	}
}

// CreateArticleRequest represents the request body for creating an article
type CreateArticleRequest struct {
	UserID      int64  `json:"user_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	PublishedAt *int64 `json:"published_at,omitempty"` // Unix timestamp (nullable)
}

// UpdateArticleRequest represents the request body for updating an article
type UpdateArticleRequest struct {
	UserID      int64  `json:"user_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	PublishedAt *int64 `json:"published_at,omitempty"` // Unix timestamp (nullable)
}

// CreateArticle handles POST /api/v1/articles
func (h *ArticleHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var req CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.UserID == 0 || req.Title == "" || req.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "UserID, title, and content are required"})
		return
	}

	// Convert publishedAt to pgtype.Timestamp
	var publishedAt pgtype.Timestamp
	if req.PublishedAt != nil {
		publishedAt = pgtype.Timestamp{
			Time:  time.Unix(*req.PublishedAt, 0),
			Valid: true,
		}
	} else {
		publishedAt = pgtype.Timestamp{
			Valid: false,
		}
	}

	article, err := h.usecase.CreateArticle(r.Context(), req.UserID, req.Title, req.Content, publishedAt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to create article: %v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(article)
}

// GetArticle handles GET /api/v1/articles/{id}
func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid article ID"})
		return
	}

	article, err := h.usecase.GetArticle(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Article not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(article)
}

// ListArticles handles GET /api/v1/articles
func (h *ArticleHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := h.usecase.ListArticles(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to list articles: %v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(articles)
}

// UpdateArticle handles PUT /api/v1/articles/{id}
func (h *ArticleHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var req UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.UserID == 0 || req.Title == "" || req.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "UserID, title, and content are required"})
		return
	}

	// Convert publishedAt to pgtype.Timestamp
	var publishedAt pgtype.Timestamp
	if req.PublishedAt != nil {
		publishedAt = pgtype.Timestamp{
			Time:  time.Unix(*req.PublishedAt, 0),
			Valid: true,
		}
	} else {
		publishedAt = pgtype.Timestamp{
			Valid: false,
		}
	}

	article, err := h.usecase.UpdateArticle(r.Context(), id, req.UserID, req.Title, req.Content, publishedAt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to update article: %v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(article)
}

// DeleteArticle handles DELETE /api/v1/articles/{id}
func (h *ArticleHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid article ID"})
		return
	}

	if err := h.usecase.DeleteArticle(r.Context(), id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Article not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
