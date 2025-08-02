package user

import (
	"context"
	"database/sql"
	"errors"
	domain "go-clean-architecture/internal/domain/user"
)

type PostgresUserRepository struct {
	DB *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) domain.Repository {
	return &PostgresUserRepository{DB: db}
}

func (r *PostgresUserRepository) GetByID(id int) (*domain.User, error) {
	row := r.DB.QueryRowContext(context.Background(), "SELECT id, username, password FROM users WHERE id = $1", id)
	var user domain.User
	if err := row.Scan(&user.ID, &user.Name, &user.Password); err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (r *PostgresUserRepository) Create(user *domain.User) error {
	return r.DB.QueryRowContext(context.Background(), "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", user.Name, user.Password).Scan(&user.ID)
}

func (r *PostgresUserRepository) GetByUsernameAndPassword(username, password string) (*domain.User, error) {
	row := r.DB.QueryRowContext(context.Background(), "SELECT id, username, password FROM users WHERE username = $1 AND password = $2", username, password)
	var user domain.User
	if err := row.Scan(&user.ID, &user.Name, &user.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}
