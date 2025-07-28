package http

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// NewAuthHandler registers the /login endpoint
func NewAuthHandler(router *mux.Router) {
	router.HandleFunc("/login", loginHandler).Methods("POST")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	// Simple hardcoded user/pass for demo (replace with real user validation)
	if req.Username != "admin" || req.Password != "password" {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"sub": req.Username,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "could not sign token", http.StatusInternalServerError)
		return
	}
	resp := loginResponse{Token: tokenStr}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
