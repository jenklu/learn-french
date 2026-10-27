package services

import (
	"context"
	"testing"

	"github.com/jenklu/learn-french/internal/database"
	"github.com/jenklu/learn-french/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExerciseService_GetExercises(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
	
	db := setupTestDB(t)
	defer db.Close()
	
	service := NewExerciseService(db)
	
	t.Run("get A2 exercises", func(t *testing.T) {
		exercises, err := service.GetExercises(context.Background(), "A2", 5)
		assert.NoError(t, err)
		assert.NotEmpty(t, exercises)
		assert.Len(t, exercises, 5)
		
		for _, exercise := range exercises {
			assert.Equal(t, "A2", exercise.Level)
			assert.NotEmpty(t, exercise.Question)
			assert.NotEmpty(t, exercise.Options)
			assert.True(t, exercise.CorrectAnswer >= 0 && exercise.CorrectAnswer < len(exercise.Options))
		}
	})
	
	t.Run("get B1 exercises", func(t *testing.T) {
		exercises, err := service.GetExercises(context.Background(), "B1", 5)
		assert.NoError(t, err)
		assert.NotEmpty(t, exercises)
		
		for _, exercise := range exercises {
			assert.Equal(t, "B1", exercise.Level)
		}
	})
}

func TestExerciseService_SubmitExercise(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
	
	db := setupTestDB(t)
	defer db.Close()
	
	service := NewExerciseService(db)
	
	// Create a test user
	authService := NewAuthService(db)
	req := &models.UserRegistration{
		Username: "exerciseuser",
		Email:    "exercise@example.com",
		Password: "password123",
		Timezone: "UTC",
	}
	
	user, err := authService.Register(context.Background(), req)
	require.NoError(t, err)
	
	t.Run("submit correct answer", func(t *testing.T) {
		// Get exercises first
		exercises, err := service.GetExercises(context.Background(), "A2", 1)
		require.NoError(t, err)
		require.NotEmpty(t, exercises)
		
		submission := &models.ExerciseSubmission{
			SessionID:    0, // Will be created automatically
			ExerciseID:   exercises[0].ID,
			ExerciseIndex: 0,
			SelectedOption: exercises[0].CorrectAnswer,
			IsCorrect:    true,
			Score:        20,
			Duration:     30,
		}
		
		result, err := service.SubmitExercise(context.Background(), user.ID, submission)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsCorrect)
		assert.Equal(t, 20, result.Score)
	})
	
	t.Run("submit incorrect answer", func(t *testing.T) {
		// Get exercises first
		exercises, err := service.GetExercises(context.Background(), "A2", 1)
		require.NoError(t, err)
		require.NotEmpty(t, exercises)
		
		// Select wrong option
		wrongOption := exercises[0].CorrectAnswer + 1
		if wrongOption >= len(exercises[0].Options) {
			wrongOption = 0
		}
		
		submission := &models.ExerciseSubmission{
			SessionID:    0,
			ExerciseID:   exercises[0].ID,
			ExerciseIndex: 0,
			SelectedOption: wrongOption,
			IsCorrect:    false,
			Score:        0,
			Duration:     30,
		}
		
		result, err := service.SubmitExercise(context.Background(), user.ID, submission)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsCorrect)
		assert.Equal(t, 0, result.Score)
	})
}

func TestExerciseService_CompleteSession(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
	
	db := setupTestDB(t)
	defer db.Close()
	
	service := NewExerciseService(db)
	
	// Create a test user
	authService := NewAuthService(db)
	req := &models.UserRegistration{
		Username: "sessionuser",
		Email:    "session@example.com",
		Password: "password123",
		Timezone: "UTC",
	}
	
	user, err := authService.Register(context.Background(), req)
	require.NoError(t, err)
	
	t.Run("complete session with high score", func(t *testing.T) {
		completion := &models.SessionCompletion{
			DurationMinutes: 10,
			ListeningScore:  95,
		}
		
		session, err := service.CompleteSession(context.Background(), user.ID, completion)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, 95, session.ListeningScore)
		assert.Equal(t, 10, session.DurationMinutes)
		
		// Check that progress was updated
		progress, err := service.GetUserProgress(context.Background(), user.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, progress)
		
		// Find listening progress
		var listeningProgress *models.Progress
		for _, p := range progress {
			if p.SkillType == "listening" {
				listeningProgress = p
				break
			}
		}
		
		assert.NotNil(t, listeningProgress)
		assert.Equal(t, "B1", listeningProgress.Level)
	})
	
	t.Run("complete session with low score", func(t *testing.T) {
		completion := &models.SessionCompletion{
			DurationMinutes: 5,
			ListeningScore:  45,
		}
		
		session, err := service.CompleteSession(context.Background(), user.ID, completion)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, 45, session.ListeningScore)
	})
}
