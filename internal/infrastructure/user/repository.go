package user

import (
    "errors"
    domain "go-clean-architecture/internal/domain/user"
)

type userRepository struct {
    users map[int]*domain.User
    nextID int
}

func NewUserRepository() domain.Repository {
    return &userRepository{
        users: map[int]*domain.User{
            1: {ID: 1, Name: "Alice"},
        },
        nextID: 2,
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