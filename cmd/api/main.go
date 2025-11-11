package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/para7/nanaket-cms/internal/db"
	"github.com/para7/nanaket-cms/internal/handler"
	"github.com/para7/nanaket-cms/internal/middleware"
	"github.com/para7/nanaket-cms/internal/repository"
	"github.com/para7/nanaket-cms/internal/usecase"
)

// setupRoutes configures all application routes
func setupRoutes(mux *http.ServeMux, pool *pgxpool.Pool) {
	// Health check endpoint
	mux.HandleFunc("GET /health", healthCheckHandler(pool))

	// API v1 routes
	mux.HandleFunc("GET /api/v1/status", statusHandler)
	mux.HandleFunc("GET /api/v1/hello", helloHandler)

	// Initialize layers
	queries := db.New(pool)

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
func healthCheckHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
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
		start := time.Now()

		// Create a custom ResponseWriter to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lrw, r)

		log.Printf(
			"%s %s %s %d %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			lrw.statusCode,
			time.Since(start),
		)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
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
	// Database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://nanaket:nanaket@localhost:5432/nanaket_cms?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Test connection
	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	fmt.Println("Successfully connected to database!")

	// Initialize router
	mux := http.NewServeMux()

	// Setup routes
	setupRoutes(mux, pool)

	// Wrap with middleware
	handler := loggingMiddleware(recoveryMiddleware(mux))

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
