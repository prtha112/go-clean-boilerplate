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

func (uc *UseCase) Login(user user.User) (*user.User, error) {
	return uc.repo.GetByUsernameAndPassword(user.Username, user.Password)
}
