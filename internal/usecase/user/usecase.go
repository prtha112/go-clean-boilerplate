package user

import (
	domain "go-clean-architecture/internal/domain/user"
)

type useCase struct {
	repo domain.Repository
}

func NewUserUseCase(repo domain.Repository) *useCase {
	return &useCase{repo: repo}
}

func (uc *useCase) Login(user domain.User) (*domain.User, error) {
	return uc.repo.GetByUsernameAndPassword(user.Username, user.Password)
}
