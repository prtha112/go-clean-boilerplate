package user

import (
	"go-clean-architecture/internal/domain/user"
)

type UseCase struct {
	repo user.Repository
}

func NewUserUseCase(repo user.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Login(username, password string) (*user.User, error) {
	return uc.repo.GetByUsernameAndPassword(username, password)
}
