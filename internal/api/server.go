package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/para7/nanaket-cms/internal/usecase"
)

// Server implements the OpenAPI ServerInterface
type Server struct {
	userUsecase usecase.UserUsecase
}

// NewServer creates a new API server instance
func NewServer(userUsecase usecase.UserUsecase) *Server {
	return &Server{
		userUsecase: userUsecase,
	}
}

// HealthCheck implements the health check endpoint
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status:   Healthy,
		Database: stringPtr("connected"),
	})
}

// GetStatus implements the API status endpoint
func (s *Server) GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(StatusResponse{
		Api:     "Nanaket CMS",
		Version: "1.0.0",
		Status:  "running",
	})
}

// SayHello implements the hello endpoint
func (s *Server) SayHello(w http.ResponseWriter, r *http.Request, params SayHelloParams) {
	name := "World"
	if params.Name != nil {
		name = *params.Name
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", name),
	})
}

// CreateUser implements POST /api/v1/users
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Email == "" || req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Email and name are required"})
		return
	}

	user, err := s.userUsecase.CreateUser(r.Context(), string(req.Email), req.Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to create user: %v", err)})
		return
	}

	apiUser := User{
		Id:        user.ID,
		Email:     req.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(apiUser)
}

// GetUser implements GET /api/v1/users/{id}
func (s *Server) GetUser(w http.ResponseWriter, r *http.Request, id int64) {
	user, err := s.userUsecase.GetUser(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found"})
		return
	}

	apiUser := User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiUser)
}

// ListUsers implements GET /api/v1/users
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.userUsecase.ListUsers(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: fmt.Sprintf("Failed to list users: %v", err)})
		return
	}

	apiUsers := make([]User, len(users))
	for i, user := range users {
		apiUsers[i] = User{
			Id:        user.ID,
			Email:     openapi_types.Email(user.Email),
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiUsers)
}

// UpdateUser implements PUT /api/v1/users/{id}
func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request, id int64) {
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Email == "" || req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Email and name are required"})
		return
	}

	user, err := s.userUsecase.UpdateUser(r.Context(), id, string(req.Email), req.Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found"})
		return
	}

	apiUser := User{
		Id:        user.ID,
		Email:     req.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiUser)
}

// DeleteUser implements DELETE /api/v1/users/{id}
func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request, id int64) {
	if err := s.userUsecase.DeleteUser(r.Context(), id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
