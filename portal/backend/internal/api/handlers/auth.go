package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"kit.workmate/live-portal/internal/auth"
	"kit.workmate/live-portal/internal/storage"
)

type AuthHandler struct {
	userStore  *storage.UserStore
	jwtService *auth.JWTService
}

func NewAuthHandler(userStore *storage.UserStore, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userStore:  userStore,
		jwtService: jwtService,
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token    string          `json:"token"`
	User     *storage.User   `json:"user"`
	ExpiresIn string         `json:"expires_in"`
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := h.userStore.Authenticate(req.Username, req.Password)
	if err != nil {
		if err == storage.ErrInvalidCredentials {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return token and user info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token:     token,
		User:      user,
		ExpiresIn: "24h",
	})
}

// Logout handles user logout (client-side token removal)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// With JWT, logout is mainly client-side (remove token)
	// We could implement token blacklisting here if needed
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// Verify checks if the current token is valid
func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "No authorization header", http.StatusUnauthorized)
		return
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		return
	}

	token := parts[1]

	// Validate token
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		if err == auth.ErrTokenExpired {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := h.userStore.GetByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Return user info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid": true,
		"user":  user,
	})
}
