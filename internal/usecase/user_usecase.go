package usecase

import (
	"context"

	"github.com/para7/nanaket-cms/internal/db"
	"github.com/para7/nanaket-cms/internal/repository"
)

// UserUsecase defines the interface for user business logic
type UserUsecase interface {
	CreateUser(ctx context.Context, email, name string) (db.User, error)
	GetUser(ctx context.Context, id int64) (db.User, error)
	ListUsers(ctx context.Context) ([]db.User, error)
	UpdateUser(ctx context.Context, id int64, email, name string) (db.User, error)
	DeleteUser(ctx context.Context, id int64) error
}

// userUsecase implements UserUsecase interface
type userUsecase struct {
	repo repository.UserRepository
}

// NewUserUsecase creates a new instance of UserUsecase
func NewUserUsecase(repo repository.UserRepository) UserUsecase {
	return &userUsecase{
		repo: repo,
	}
}

// CreateUser creates a new user
func (u *userUsecase) CreateUser(ctx context.Context, email, name string) (db.User, error) {
	return u.repo.Create(ctx, email, name)
}

// GetUser retrieves a user by ID
func (u *userUsecase) GetUser(ctx context.Context, id int64) (db.User, error) {
	return u.repo.GetByID(ctx, id)
}

// ListUsers retrieves all users
func (u *userUsecase) ListUsers(ctx context.Context) ([]db.User, error) {
	return u.repo.List(ctx)
}

// UpdateUser updates a user
func (u *userUsecase) UpdateUser(ctx context.Context, id int64, email, name string) (db.User, error) {
	return u.repo.Update(ctx, id, email, name)
}

// DeleteUser deletes a user
func (u *userUsecase) DeleteUser(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
