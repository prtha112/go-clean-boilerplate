package user_test

import (
	"errors"
	"testing"

	"go-clean-architecture/internal/domain/user"
	usecase "go-clean-architecture/internal/usecase/user"

	"github.com/stretchr/testify/assert"
)

// 🔧 Mock Repository
type mockRepo struct {
	GetByIDFunc func(int) (*user.User, error)
	CreateFunc  func(*user.User) error
	GetByUsernameAndPasswordFunc func(string, string) (*user.User, error)
}

func (m *mockRepo) GetByID(id int) (*user.User, error) {
	return m.GetByIDFunc(id)
}

func (m *mockRepo) Create(u *user.User) error {
	return m.CreateFunc(u)
}

func (m *mockRepo) GetByUsernameAndPassword(username, password string) (*user.User, error) {
	if m.GetByUsernameAndPasswordFunc != nil {
		return m.GetByUsernameAndPasswordFunc(username, password)
	}
	return nil, errors.New("not implemented")
}

func TestGetUser_Success(t *testing.T) {
	mock := &mockRepo{
		GetByIDFunc: func(id int) (*user.User, error) {
			return &user.User{ID: id, Name: "TestUser"}, nil
		},
	}

	uc := usecase.NewUserUseCase(mock)
	u, err := uc.GetUser(1)

	assert.NoError(t, err)
	assert.Equal(t, "TestUser", u.Name)
}

func TestGetUser_NotFound(t *testing.T) {
	mock := &mockRepo{
		GetByIDFunc: func(id int) (*user.User, error) {
			return nil, errors.New("user not found")
		},
	}

	uc := usecase.NewUserUseCase(mock)
	u, err := uc.GetUser(99)

	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestCreateUser_Success(t *testing.T) {
	mock := &mockRepo{
		CreateFunc: func(u *user.User) error {
			u.ID = 99
			return nil
		},
	}

	uc := usecase.NewUserUseCase(mock)
	user := &user.User{Name: "NewUser"}

	err := uc.CreateUser(user)

	assert.NoError(t, err)
	assert.Equal(t, 99, user.ID)
}
