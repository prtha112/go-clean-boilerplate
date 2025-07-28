package user

import domain "go-clean-architecture/internal/domain/user"

type userUseCase struct {
	repo domain.Repository
}

type UseCase interface {
	GetUser(id int) (*domain.User, error)
	CreateUser(user *domain.User) error
}

func NewUserUseCase(r domain.Repository) UseCase {
	return &userUseCase{repo: r}
}

func (uc *userUseCase) GetUser(id int) (*domain.User, error) {
	return uc.repo.GetByID(id)
}

func (uc *userUseCase) CreateUser(u *domain.User) error {
	return uc.repo.Create(u)
}
