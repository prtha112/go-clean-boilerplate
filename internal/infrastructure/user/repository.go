package user

import (
	"context"
	"database/sql"
	"errors"
	domain "go-clean-architecture/internal/domain/user"
)

type userRepository struct {
	users  map[int]*domain.User
	nextID int
}

func NewUserRepository() domain.Repository {
	return &userRepository{
		users: map[int]*domain.User{
			1: {ID: 1, Name: "admin", Password: "password"},
			2: {ID: 2, Name: "alice", Password: "alice123"},
		},
		nextID: 3,
	}
}

func (r *userRepository) GetByID(id int) (*domain.User, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (r *userRepository) Create(user *domain.User) error {
	user.ID = r.nextID
	r.users[user.ID] = user
	r.nextID++
	return nil
}

func (r *userRepository) GetByUsernameAndPassword(username, password string) (*domain.User, error) {
	for _, user := range r.users {
		if user.Name == username && user.Password == password {
			return user, nil
		}
	}
	return nil, errors.New("invalid credentials")
}

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
