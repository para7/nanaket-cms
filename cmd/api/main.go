package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/para7/nanaket-cms/internal/db"
	"github.com/para7/nanaket-cms/internal/handler"
	"github.com/para7/nanaket-cms/internal/middleware"
	"github.com/para7/nanaket-cms/internal/repository"
	"github.com/para7/nanaket-cms/internal/usecase"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/d1"
)

// setupRoutes configures all application routes
func setupRoutes(mux *http.ServeMux, database *sql.DB) {
	// Health check endpoint
	mux.HandleFunc("GET /health", healthCheckHandler(database))

	// API v1 routes
	mux.HandleFunc("GET /api/v1/status", statusHandler)
	mux.HandleFunc("GET /api/v1/hello", helloHandler)

	// Initialize layers
	queries := db.New(database)

	// Auth handler (no usecase, direct query access for simple temporary implementation)
	authHandler := handler.NewAuthHandler(queries)

	// User layer
	userRepo := repository.NewUserRepository(queries)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewUserHandler(userUsecase)

	// Article layer
	articleRepo := repository.NewArticleRepository(queries)
	articleUsecase := usecase.NewArticleUsecase(articleRepo)
	articleHandler := handler.NewArticleHandler(articleUsecase)

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(queries)

	// Auth endpoints (no authentication required)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	// User CRUD endpoints (no authentication required for now)
	mux.HandleFunc("POST /api/v1/users", userHandler.CreateUser)
	mux.HandleFunc("GET /api/v1/users", userHandler.ListUsers)
	mux.HandleFunc("GET /api/v1/users/{id}", userHandler.GetUser)
	mux.HandleFunc("PUT /api/v1/users/{id}", userHandler.UpdateUser)
	mux.HandleFunc("DELETE /api/v1/users/{id}", userHandler.DeleteUser)

	// Article endpoints
	// Create, Read, List - no authentication required
	mux.HandleFunc("POST /api/v1/articles", articleHandler.CreateArticle)
	mux.HandleFunc("GET /api/v1/articles", articleHandler.ListArticles)
	mux.HandleFunc("GET /api/v1/articles/{id}", articleHandler.GetArticle)
	// Update, Delete - authentication required
	mux.Handle("PUT /api/v1/articles/{id}", authMiddleware(http.HandlerFunc(articleHandler.UpdateArticle)))
	mux.Handle("DELETE /api/v1/articles/{id}", authMiddleware(http.HandlerFunc(articleHandler.DeleteArticle)))
}

// healthCheckHandler returns a handler that checks database connectivity
func healthCheckHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := database.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintf(w, `{"status":"unhealthy","error":"%v"}`, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"healthy","database":"connected"}`)
	}
}

// statusHandler returns API status information
func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"api":"Nanaket CMS","version":"1.0.0","status":"running"}`)
}

// helloHandler is a simple example endpoint
func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"message":"Hello, %s!"}`, name)
}

// loggingMiddleware logs incoming HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware recovers from panics and returns 500 error
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprint(w, `{"error":"Internal server error"}`)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Get D1 database binding from Cloudflare Workers environment
	// The binding name should match what's configured in wrangler.toml
	database, err := d1.NewClient("DB")
	if err != nil {
		log.Fatalf("Failed to create D1 client: %v", err)
	}

	// Initialize router
	mux := http.NewServeMux()

	// Setup routes
	setupRoutes(mux, database)

	// Wrap with middleware
	handler := loggingMiddleware(recoveryMiddleware(mux))

	// Start Cloudflare Workers server
	log.Println("Starting Cloudflare Workers server...")
	workers.Serve(handler)
}
