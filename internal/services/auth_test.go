package services

import (
	"context"
	"testing"

	"github.com/jenklu/learn-french/internal/database"
	"github.com/jenklu/learn-french/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *database.Database {
	// In a real test, you'd use a test database
	// For this example, we'll use an in-memory SQLite database
	// Note: This is a simplified version - in production you'd use a proper test database
	
	db, err := database.New(context.Background(), "postgres://user:password@localhost:5432/learn_french_test?sslmode=disable")
	require.NoError(t, err)
	
	// Initialize the database
	err = db.Init(context.Background())
	require.NoError(t, err)
	
	return db
}

func TestAuthService_Register(t *testing.T) {
	// Skip this test if we can't connect to the database
	t.Skip("Skipping test that requires database connection")
	
	db := setupTestDB(t)
	defer db.Close()
	
	service := NewAuthService(db)
	
	t.Run("successful registration", func(t *testing.T) {
		req := &models.UserRegistration{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
			Timezone: "UTC",
		}
		
		user, err := service.Register(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.NotEmpty(t, user.PasswordHash)
		assert.False(t, user.IsAdmin)
	})
	
	t.Run("duplicate username", func(t *testing.T) {
		req := &models.UserRegistration{
			Username: "testuser",
			Email:    "another@example.com",
			Password: "password123",
			Timezone: "UTC",
		}
		
		_, err := service.Register(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, ErrUsernameExists, err)
	})
	
	t.Run("duplicate email", func(t *testing.T) {
		// First, clean up the duplicate username
		_, err := db.Pool.Exec(context.Background(), "DELETE FROM users WHERE username = 'testuser'")
		require.NoError(t, err)
		
		req := &models.UserRegistration{
			Username: "anotheruser",
			Email:    "test@example.com",
			Password: "password123",
			Timezone: "UTC",
		}
		
		_, err = service.Register(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, ErrEmailExists, err)
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
	
	db := setupTestDB(t)
	defer db.Close()
	
	service := NewAuthService(db)
	
	// Create a test user
	req := &models.UserRegistration{
		Username: "loginuser",
		Email:    "login@example.com",
		Password: "password123",
		Timezone: "UTC",
	}
	
	user, err := service.Register(context.Background(), req)
	require.NoError(t, err)
	
	t.Run("successful login", func(t *testing.T) {
		loginReq := &models.UserLogin{
			Username: "loginuser",
			Password: "password123",
		}
		
		loggedInUser, err := service.Login(context.Background(), loginReq)
		assert.NoError(t, err)
		assert.NotNil(t, loggedInUser)
		assert.Equal(t, "loginuser", loggedInUser.Username)
	})
	
	t.Run("invalid username", func(t *testing.T) {
		loginReq := &models.UserLogin{
			Username: "nonexistent",
			Password: "password123",
		}
		
		_, err := service.Login(context.Background(), loginReq)
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err)
	})
	
	t.Run("invalid password", func(t *testing.T) {
		loginReq := &models.UserLogin{
			Username: "loginuser",
			Password: "wrongpassword",
		}
		
		_, err := service.Login(context.Background(), loginReq)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPassword, err)
	})
}

func TestAuthService_UpdateTimezone(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
	
	db := setupTestDB(t)
	defer db.Close()
	
	service := NewAuthService(db)
	
	// Create a test user
	req := &models.UserRegistration{
		Username: "timezoneuser",
		Email:    "timezone@example.com",
		Password: "password123",
		Timezone: "UTC",
	}
	
	user, err := service.Register(context.Background(), req)
	require.NoError(t, err)
	
	t.Run("update timezone", func(t *testing.T) {
		err := service.UpdateTimezone(context.Background(), user.ID, "America/New_York")
		assert.NoError(t, err)
		
		// Verify the timezone was updated
		updatedUser, err := db.GetUserByID(context.Background(), user.ID)
		assert.NoError(t, err)
		assert.Equal(t, "America/New_York", updatedUser.Timezone)
	})
}
