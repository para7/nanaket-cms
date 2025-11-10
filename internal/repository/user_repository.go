package repository

import (
	"context"

	"github.com/para7/nanaket-cms/internal/db"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, email, name string) (db.User, error)
	GetByID(ctx context.Context, id int64) (db.User, error)
	List(ctx context.Context) ([]db.User, error)
	Update(ctx context.Context, id int64, email, name string) (db.User, error)
	Delete(ctx context.Context, id int64) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	querier db.Querier
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(querier db.Querier) UserRepository {
	return &userRepository{
		querier: querier,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, email, name string) (db.User, error) {
	return r.querier.CreateUser(ctx, db.CreateUserParams{
		Email: email,
		Name:  name,
	})
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id int64) (db.User, error) {
	return r.querier.GetUser(ctx, id)
}

// List retrieves all users
func (r *userRepository) List(ctx context.Context) ([]db.User, error) {
	return r.querier.ListUsers(ctx)
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, id int64, email, name string) (db.User, error) {
	return r.querier.UpdateUser(ctx, db.UpdateUserParams{
		ID:    id,
		Email: email,
		Name:  name,
	})
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	return r.querier.DeleteUser(ctx, id)
}
