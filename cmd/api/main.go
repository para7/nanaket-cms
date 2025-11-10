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

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/para7/nanaket-cms/internal/api"
	"github.com/para7/nanaket-cms/internal/db"
	"github.com/para7/nanaket-cms/internal/repository"
	"github.com/para7/nanaket-cms/internal/usecase"
)

// setupAPIServer initializes the OpenAPI server with all dependencies
func setupAPIServer(pool *pgxpool.Pool) http.Handler {
	// Initialize layers
	queries := db.New(pool)
	userRepo := repository.NewUserRepository(queries)
	userUsecase := usecase.NewUserUsecase(userRepo)

	// Create OpenAPI server implementation
	apiServer := api.NewServer(userUsecase)

	// Create chi router
	r := chi.NewRouter()

	// Add OpenAPI spec endpoint
	r.Get("/openapi.yaml", serveOpenAPISpec)
	r.Get("/openapi.json", serveOpenAPISpecJSON)

	// Mount the generated API handler
	return api.HandlerFromMux(apiServer, r)
}

// serveOpenAPISpec serves the OpenAPI specification in YAML format
func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	swagger, err := api.GetSwagger()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load OpenAPI spec: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to YAML
	yamlData, err := swagger.MarshalJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal OpenAPI spec: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(yamlData)
}

// serveOpenAPISpecJSON serves the OpenAPI specification in JSON format
func serveOpenAPISpecJSON(w http.ResponseWriter, r *http.Request) {
	swagger, err := api.GetSwagger()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load OpenAPI spec: %v", err), http.StatusInternalServerError)
		return
	}

	jsonData, err := swagger.MarshalJSON()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal OpenAPI spec: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonData)
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

	// Setup OpenAPI server
	apiHandler := setupAPIServer(pool)

	// Wrap with middleware
	handler := loggingMiddleware(recoveryMiddleware(apiHandler))

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
