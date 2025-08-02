package user

import (
	"context"
	"database/sql"
	"errors"
	config "go-clean-architecture/config"
	domain "go-clean-architecture/internal/domain/user"
	"log"
)

type PostgresUserRepository struct {
	DB *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) domain.Repository {
	return &PostgresUserRepository{DB: db}
}

func (r *PostgresUserRepository) GetByUsernameAndPassword(username, password string) (*domain.User, error) {
	row := r.DB.QueryRowContext(context.Background(), "SELECT id, username, password FROM users WHERE username = $1", username)
	var user domain.User
	if err := row.Scan(&user.ID, &user.Name, &user.Password); err != nil {
		log.Printf("Login failed for username=%s", username)
		return nil, errors.New("invalid credentials")
	}
	if !config.VerifyPassword(user.Password, password) {
		log.Printf("Login failed: password mismatch for username=%s", username)
		return nil, errors.New("invalid credentials")
	}
	log.Printf("Login success for username=%s", username)
	return &user, nil
}
