package repository

import (
	"database/sql"
	"testing"
	"time"

	"go-clean-boilerplate/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUserRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	assert.NotNil(t, repo)
	assert.Implements(t, (*domain.UserRepository)(nil), repo)
}

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), user.Username, user.Email, user.Password, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(user)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), user.Username, user.Email, user.Password, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = repo.Create(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New()
	expectedUser := &domain.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Username, expectedUser.Email, expectedUser.Password, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	user, err := repo.GetByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Username, user.Username)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.Password, user.Password)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByID(userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.GetByID(userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to get user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByUsername_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	username := "testuser"
	expectedUser := &domain.User{
		ID:        uuid.New(),
		Username:  username,
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Username, expectedUser.Email, expectedUser.Password, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE username = \$1`).
		WithArgs(username).
		WillReturnRows(rows)

	user, err := repo.GetByUsername(username)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Username, user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	email := "test@example.com"
	expectedUser := &domain.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     email,
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.Username, expectedUser.Email, expectedUser.Password, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(`SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	user, err := repo.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "updateduser",
		Email:    "updated@example.com",
		Password: "newhashedpassword",
	}

	mock.ExpectExec(`UPDATE users SET username = \$2, email = \$3, password = \$4, updated_at = \$5 WHERE id = \$1`).
		WithArgs(user.ID, user.Username, user.Email, user.Password, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(user)

	assert.NoError(t, err)
	assert.False(t, user.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "updateduser",
		Email:    "updated@example.com",
		Password: "newhashedpassword",
	}

	mock.ExpectExec(`UPDATE users SET username = \$2, email = \$3, password = \$4, updated_at = \$5 WHERE id = \$1`).
		WithArgs(user.ID, user.Username, user.Email, user.Password, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "updateduser",
		Email:    "updated@example.com",
		Password: "newhashedpassword",
	}

	mock.ExpectExec(`UPDATE users SET username = \$2, email = \$3, password = \$4, updated_at = \$5 WHERE id = \$1`).
		WithArgs(user.ID, user.Username, user.Email, user.Password, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = repo.Update(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New()

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New()

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)

	userID := uuid.New()

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	err = repo.Delete(userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user")
	assert.NoError(t, mock.ExpectationsWereMet())
}
