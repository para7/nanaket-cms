package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/para7/nanaket-cms/internal/db"
	"github.com/para7/nanaket-cms/internal/middleware"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	queries db.Querier
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(queries db.Querier) *AuthHandler {
	return &AuthHandler{
		queries: queries,
	}
}

// LoginRequest represents the request body for login
type LoginRequest struct {
	Token string `json:"token"`
}

// LoginResponse represents the response body for successful login
type LoginResponse struct {
	Message string  `json:"message"`
	User    db.User `json:"user"`
}

// Login handles POST /api/v1/auth/login
// It validates the provided token and sets it as a secure cookie
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Token is required"})
		return
	}

	// Validate token
	user, err := h.queries.GetUserByToken(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid or expired token"})
			return
		}
		log.Printf("Error validating token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
		return
	}

	// Set secure cookie with the token
	cookie := &http.Cookie{
		Name:     middleware.CookieName,
		Value:    req.Token,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7, // 7 days
		HttpOnly: true,              // Prevent JavaScript access (XSS protection)
		Secure:   true,              // Only send over HTTPS
		SameSite: http.SameSiteStrictMode, // CSRF protection
	}
	http.SetCookie(w, cookie)

	// Return success response with user info
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(LoginResponse{
		Message: "Login successful",
		User:    user,
	})
}

// Logout handles POST /api/v1/auth/logout
// It clears the auth cookie
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the cookie by setting MaxAge to -1
	cookie := &http.Cookie{
		Name:     middleware.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}
