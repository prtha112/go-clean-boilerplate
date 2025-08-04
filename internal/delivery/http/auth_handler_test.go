package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-clean-v2/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock AuthUsecase
type mockAuthUsecase struct {
	mock.Mock
}

func (m *mockAuthUsecase) Register(req *domain.RegisterRequest) (*domain.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockAuthUsecase) Login(req *domain.LoginRequest) (*domain.LoginResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoginResponse), args.Error(1)
}

func (m *mockAuthUsecase) ValidateToken(tokenString string) (*domain.User, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockAuthUsecase) GenerateToken(user *domain.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func setupAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewAuthHandler(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUsecase, handler.authUsecase)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.POST("/register", handler.Register)

	req := domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedUser := &domain.User{
		ID:       uuid.New(),
		Username: req.Username,
		Email:    req.Email,
	}

	mockUsecase.On("Register", &req).Return(expectedUser, nil)

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User registered successfully", response["message"])
	assert.NotNil(t, response["data"])

	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.POST("/register", handler.Register)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")

	mockUsecase.AssertNotCalled(t, "Register")
}

func TestAuthHandler_Register_UsernameExists(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.POST("/register", handler.Register)

	req := domain.RegisterRequest{
		Username: "existinguser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockUsecase.On("Register", &req).Return(nil, errors.New("username already exists"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "username already exists", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Register_EmailExists(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.POST("/register", handler.Register)

	req := domain.RegisterRequest{
		Username: "testuser",
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockUsecase.On("Register", &req).Return(nil, errors.New("email already exists"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "email already exists", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Register_InternalError(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.POST("/register", handler.Register)

	req := domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockUsecase.On("Register", &req).Return(nil, errors.New("database connection failed"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "database connection failed", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.POST("/login", handler.Login)

	req := domain.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	expectedResponse := &domain.LoginResponse{
		Token: "jwt-token-here",
		User: &domain.User{
			ID:       uuid.New(),
			Username: req.Username,
			Email:    "test@example.com",
		},
	}

	mockUsecase.On("Login", &req).Return(expectedResponse, nil)

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Login successful", response["message"])
	assert.NotNil(t, response["data"])

	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.POST("/login", handler.Login)

	req := domain.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	mockUsecase.On("Login", &req).Return(nil, errors.New("invalid username or password"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid username or password", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.POST("/login", handler.Login)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockUsecase.AssertNotCalled(t, "Login")
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.POST("/login", handler.Login)

	req := domain.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockUsecase.On("Login", &req).Return(nil, errors.New("database connection failed"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "database connection failed", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestAuthHandler_GetProfile_Success(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	router := setupAuthTestRouter()
	router.GET("/profile", handler.GetProfile)

	user := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/profile", nil)

	// Create a gin context and set the user
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Set("user", user)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Profile retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
}

func TestAuthHandler_GetProfile_UserNotInContext(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/profile", nil)

	// Create a gin context without setting the user
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	handler.GetProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User not found in context", response["error"])
}

func TestAuthHandler_GetProfile_InvalidUserData(t *testing.T) {
	mockUsecase := new(mockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/profile", nil)

	// Create a gin context and set invalid user data
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq
	c.Set("user", "invalid-user-data")

	handler.GetProfile(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid user data", response["error"])
}
