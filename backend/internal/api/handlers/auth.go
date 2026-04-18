package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"backend/internal/api"
	"backend/internal/api/middleware"
)

// AuthHandler provides database-backed authentication methods.
type AuthHandler struct {
	DB *sql.DB
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// LoginHandler processes user authentication against the database and issues JWT tokens.
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var userID, passwordHash, role string
	// Retrieve user credentials from SQLite
	err := h.DB.QueryRow(`SELECT id, password_hash, role FROM users WHERE username = ?`, req.Username).
		Scan(&userID, &passwordHash, &role)
	
	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		} else {
			api.WriteError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Verify password hash using bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		api.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtKey) == 0 {
		jwtKey = []byte("dev_secret_key")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &middleware.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "Could not generate authentication token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: tokenString})
}
