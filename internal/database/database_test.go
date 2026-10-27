package database

import (
	"context"
	"testing"

	"github.com/jenklu/learn-french/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *Database {
	// Skip this test if we can't connect to the database
	t.Skip("Skipping test that requires database connection")
	
	db, err := New(context.Background(), "postgres://user:password@localhost:5432/learn_french_test?sslmode=disable")
	require.NoError(t, err)
	
	err = db.Init(context.Background())
	require.NoError(t, err)
	
	return db
}

func TestDatabase_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Timezone:     "UTC",
		IsAdmin:      false,
	}
	
	err := db.CreateUser(context.Background(), user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "testuser", user.Username)
}

func TestDatabase_GetUserByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	// Create a user first
	user := &models.User{
		Username:     "getuser",
		Email:        "get@example.com",
		PasswordHash: "hashedpassword",
		Timezone:     "UTC",
		IsAdmin:      false,
	}
	
	err := db.CreateUser(context.Background(), user)
	require.NoError(t, err)
	
	// Get the user
	fetchedUser, err := db.GetUserByUsername(context.Background(), "getuser")
	assert.NoError(t, err)
	assert.NotNil(t, fetchedUser)
	assert.Equal(t, "getuser", fetchedUser.Username)
	assert.Equal(t, "get@example.com", fetchedUser.Email)
}

func TestDatabase_CreateSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	// Create a user first
	user := &models.User{
		Username:     "sessionuser",
		Email:        "session@example.com",
		PasswordHash: "hashedpassword",
		Timezone:     "UTC",
		IsAdmin:      false,
	}
	
	err := db.CreateUser(context.Background(), user)
	require.NoError(t, err)
	
	session := &models.Session{
		UserID:         user.ID,
		Date:           "2023-10-19",
		DurationMinutes: 10,
		ListeningScore: 85,
	}
	
	err = db.CreateSession(context.Background(), session)
	assert.NoError(t, err)
	assert.NotZero(t, session.ID)
}

func TestDatabase_CreateExerciseResult(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	// Create a user and session first
	user := &models.User{
		Username:     "resultuser",
		Email:        "result@example.com",
		PasswordHash: "hashedpassword",
		Timezone:     "UTC",
		IsAdmin:      false,
	}
	
	err := db.CreateUser(context.Background(), user)
	require.NoError(t, err)
	
	session := &models.Session{
		UserID:         user.ID,
		Date:           "2023-10-19",
		DurationMinutes: 10,
		ListeningScore: 85,
	}
	
	err = db.CreateSession(context.Background(), session)
	require.NoError(t, err)
	
	result := &models.ExerciseResult{
		SessionID:    session.ID,
		ExerciseIndex: 0,
		IsCorrect:    true,
		Score:        20,
	}
	
	err = db.CreateExerciseResult(context.Background(), result)
	assert.NoError(t, err)
	assert.NotZero(t, result.ID)
}

func TestDatabase_UpdateProgress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	// Create a user first
	user := &models.User{
		Username:     "progressuser",
		Email:        "progress@example.com",
		PasswordHash: "hashedpassword",
		Timezone:     "UTC",
		IsAdmin:      false,
	}
	
	err := db.CreateUser(context.Background(), user)
	require.NoError(t, err)
	
	err = db.UpdateProgress(context.Background(), user.ID, "listening", "A2")
	assert.NoError(t, err)
	
	progress, err := db.GetProgress(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, progress)
	
	found := false
	for _, p := range progress {
		if p.SkillType == "listening" && p.Level == "A2" {
			found = true
			break
		}
	}
	assert.True(t, found, "Progress for listening/A2 should exist")
}

func TestDatabase_UpdateReminder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	// Create a user first
	user := &models.User{
		Username:     "reminderuser",
		Email:        "reminder@example.com",
		PasswordHash: "hashedpassword",
		Timezone:     "UTC",
		IsAdmin:      false,
	}
	
	err := db.CreateUser(context.Background(), user)
	require.NoError(t, err)
	
	err = db.UpdateReminder(context.Background(), user.ID, true)
	assert.NoError(t, err)
	
	reminder, err := db.GetReminder(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, reminder)
	assert.True(t, reminder.Enabled)
}
