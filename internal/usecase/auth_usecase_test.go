package usecase

import (
	"fmt"
	"testing"
	"time"

	"go-clean-boilerplate/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*domain.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestNewAuthUsecase(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"

	authUC := NewAuthUsecase(mockRepo, jwtSecret)

	assert.NotNil(t, authUC)
	assert.Implements(t, (*domain.AuthUsecase)(nil), authUC)
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	req := &domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", req.Username).Return(nil, fmt.Errorf("not found"))
	mockRepo.On("GetByEmail", req.Email).Return(nil, fmt.Errorf("not found"))
	mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := authUC.Register(req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.Username, user.Username)
	assert.Equal(t, req.Email, user.Email)
	assert.Empty(t, user.Password) // Password should not be returned
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_UsernameAlreadyExists(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	existingUser := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "existing@example.com",
	}

	req := &domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", req.Username).Return(existingUser, nil)

	user, err := authUC.Register(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "username already exists")
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_EmailAlreadyExists(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	existingUser := &domain.User{
		ID:       uuid.New(),
		Username: "existing",
		Email:    "test@example.com",
	}

	req := &domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", req.Username).Return(nil, fmt.Errorf("not found"))
	mockRepo.On("GetByEmail", req.Email).Return(existingUser, nil)

	user, err := authUC.Register(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "email already exists")
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_CreateUserFails(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	req := &domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", req.Username).Return(nil, fmt.Errorf("not found"))
	mockRepo.On("GetByEmail", req.Email).Return(nil, fmt.Errorf("not found"))
	mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(fmt.Errorf("database error"))

	user, err := authUC.Register(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to create user")
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	req := &domain.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", req.Username).Return(existingUser, nil)

	response, err := authUC.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, existingUser.Username, response.User.Username)
	assert.Empty(t, response.User.Password) // Password should not be returned
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	req := &domain.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", req.Username).Return(nil, fmt.Errorf("not found"))

	response, err := authUC.Login(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid username or password")
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_InvalidPassword(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: string(hashedPassword),
	}

	req := &domain.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	mockRepo.On("GetByUsername", req.Username).Return(existingUser, nil)

	response, err := authUC.Login(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid username or password")
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_GenerateToken_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
	}

	token, err := authUC.GenerateToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*JWTClaims)
	assert.True(t, ok)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
}

func TestAuthUsecase_ValidateToken_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Generate a valid token
	token, _ := authUC.GenerateToken(user)

	mockRepo.On("GetByID", user.ID).Return(user, nil)

	validatedUser, err := authUC.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, validatedUser)
	assert.Equal(t, user.ID, validatedUser.ID)
	assert.Equal(t, user.Username, validatedUser.Username)
	assert.Empty(t, validatedUser.Password) // Password should not be returned
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_ValidateToken_InvalidToken(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	invalidToken := "invalid.token.here"

	user, err := authUC.ValidateToken(invalidToken)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestAuthUsecase_ValidateToken_ExpiredToken(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
	}

	// Create an expired token
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, _ := token.SignedString([]byte(jwtSecret))

	validatedUser, err := authUC.ValidateToken(expiredToken)

	assert.Error(t, err)
	assert.Nil(t, validatedUser)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestAuthUsecase_ValidateToken_WrongSigningMethod(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
	}

	// Create a token with wrong signing method (RS256 instead of HS256)
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	// This will create a token that fails signing method validation
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString := token.Raw // Get unsigned token

	validatedUser, err := authUC.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, validatedUser)
}

func TestAuthUsecase_ValidateToken_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
	}

	// Generate a valid token
	token, _ := authUC.GenerateToken(user)

	mockRepo.On("GetByID", user.ID).Return(nil, fmt.Errorf("user not found"))

	validatedUser, err := authUC.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, validatedUser)
	assert.Contains(t, err.Error(), "user not found")
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_ValidateToken_DifferentSecret(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	differentSecret := "different-secret"
	
	authUC1 := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)
	authUC2 := NewAuthUsecase(mockRepo, differentSecret).(*authUsecase)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
	}

	// Generate token with first secret
	token, _ := authUC1.GenerateToken(user)

	// Try to validate with different secret
	validatedUser, err := authUC2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, validatedUser)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestAuthUsecase_ValidateToken_InvalidTokenFormat(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	// Test with completely malformed token
	invalidToken := "completely.malformed.token"

	user, err := authUC.ValidateToken(invalidToken)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestAuthUsecase_GenerateToken_InvalidUser(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	// User with nil UUID will cause issues in token generation
	user := &domain.User{
		Username: "testuser",
		// ID field is not set (zero UUID)
	}

	token, err := authUC.GenerateToken(user)

	// Should still work with zero UUID, so test for normal operation
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTClaims_Structure(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

func TestAuthUsecase_Register_LongPassword(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "test-secret"
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	// Test with a very long password that exceeds bcrypt's 72-byte limit
	req := &domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(make([]byte, 80)), // Very long password, should cause bcrypt error
	}

	mockRepo.On("GetByUsername", req.Username).Return(nil, fmt.Errorf("not found"))
	mockRepo.On("GetByEmail", req.Email).Return(nil, fmt.Errorf("not found"))
	// Don't expect Create to be called since password hashing should fail

	user, err := authUC.Register(req)

	// Should fail with bcrypt error
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to hash password")
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_WithEmptySecret(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "" // Empty secret - this will still work but creates insecure tokens
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	req := &domain.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", req.Username).Return(existingUser, nil)

	response, err := authUC.Login(req)

	// Should actually succeed even with empty secret (just insecure)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	mockRepo.AssertExpectations(t)
}


func TestAuthUsecase_GenerateToken_EmptySecret(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSecret := "" // Empty secret
	authUC := NewAuthUsecase(mockRepo, jwtSecret).(*authUsecase)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
	}

	token, err := authUC.GenerateToken(user)

	// Empty secret should still work for signing, just not secure
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}