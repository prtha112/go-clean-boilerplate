package http

import (
	"database/sql"
	"encoding/json"
	domainUser "go-clean-architecture/internal/domain/user"
	userRepo "go-clean-architecture/internal/infrastructure/user"
	userUsecase "go-clean-architecture/internal/usecase/user"
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
func NewAuthHandler(router *mux.Router, db *sql.DB) {
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		repo := userRepo.NewPostgresUserRepository(db)
		uc := userUsecase.NewUserUseCase(repo)
		userDomain := domainUser.User{Username: req.Username, Password: req.Password}
		user, err := uc.Login(userDomain)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		secret := os.Getenv("JWT_SECRET")
		claims := jwt.MapClaims{
			"sub": user.Username,
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
	}).Methods("POST")
}
