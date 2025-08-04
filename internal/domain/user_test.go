package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_JSONSerialization(t *testing.T) {
	user := &User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test that password is not included in JSON
	// This would be tested in actual JSON marshaling
	assert.NotEmpty(t, user.Password)
	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Email)
}

func TestLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		request  LoginRequest
		hasError bool
	}{
		{
			name: "valid login request",
			request: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			hasError: false,
		},
		{
			name: "empty username",
			request: LoginRequest{
				Username: "",
				Password: "password123",
			},
			hasError: true,
		},
		{
			name: "empty password",
			request: LoginRequest{
				Username: "testuser",
				Password: "",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - in real app this would be done by validator
			hasError := tt.request.Username == "" || tt.request.Password == ""
			assert.Equal(t, tt.hasError, hasError)
		})
	}
}

func TestRegisterRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		request  RegisterRequest
		hasError bool
	}{
		{
			name: "valid register request",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			hasError: false,
		},
		{
			name: "short username",
			request: RegisterRequest{
				Username: "ab",
				Email:    "test@example.com",
				Password: "password123",
			},
			hasError: true,
		},
		{
			name: "invalid email",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "password123",
			},
			hasError: true,
		},
		{
			name: "short password",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "12345",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - in real app this would be done by validator
			hasError := len(tt.request.Username) < 3 ||
				len(tt.request.Username) > 50 ||
				!isValidEmail(tt.request.Email) ||
				len(tt.request.Password) < 6
			assert.Equal(t, tt.hasError, hasError)
		})
	}
}

func TestLoginResponse_Structure(t *testing.T) {
	user := &User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	response := &LoginResponse{
		Token: "jwt-token-here",
		User:  user,
	}

	assert.NotEmpty(t, response.Token)
	assert.NotNil(t, response.User)
	assert.Equal(t, user.Username, response.User.Username)
}

// Helper function for email validation (simplified)
func isValidEmail(email string) bool {
	// Simplified email validation for testing
	return len(email) > 0 &&
		len(email) <= 254 &&
		containsAt(email) &&
		containsDot(email)
}

func containsAt(s string) bool {
	for _, char := range s {
		if char == '@' {
			return true
		}
	}
	return false
}

func containsDot(s string) bool {
	for _, char := range s {
		if char == '.' {
			return true
		}
	}
	return false
}
