package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jenklu/learn-french/internal/database"
	"github.com/jenklu/learn-french/internal/models"
)

type ExerciseService struct {
	db *database.Database
}

func NewExerciseService(db *database.Database) *ExerciseService {
	return &ExerciseService{db: db}
}

func (s *ExerciseService) GetExercises(ctx context.Context, level string, count int) ([]*models.Exercise, error) {
	if count <= 0 {
		count = 5
	}
	return s.db.GetRandomExercises(ctx, level, count)
}

func (s *ExerciseService) SubmitExercise(ctx context.Context, userID int, submission *models.ExerciseSubmission) (*models.ExerciseResult, error) {
	// Get the current session or create a new one
	session, err := s.getOrCreateSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create exercise result
	result := &models.ExerciseResult{
		SessionID:    session.ID,
		ExerciseIndex: submission.ExerciseIndex,
		IsCorrect:    submission.IsCorrect,
		Score:        submission.Score,
		CreatedAt:    time.Now(),
	}

	if err := s.db.CreateExerciseResult(ctx, result); err != nil {
		return nil, err
	}

	// Update session score (we'll calculate this when completing the session)
	return result, nil
}

func (s *ExerciseService) CompleteSession(ctx context.Context, userID int, completion *models.SessionCompletion) (*models.Session, error) {
	// Get the current session
	session, err := s.getCurrentSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update session
	session.DurationMinutes = completion.DurationMinutes
	session.ListeningScore = completion.ListeningScore
	session.CompletedAt = time.Now()

	if err := s.db.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	// Update progress
	level := "A2"
	if completion.ListeningScore >= 80 {
		level = "B1"
	}
	if err := s.db.UpdateProgress(ctx, userID, "listening", level); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *ExerciseService) getOrCreateSession(ctx context.Context, userID int) (*models.Session, error) {
	// Check for an incomplete session today
	sessions, err := s.db.GetSessionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	for _, session := range sessions {
		if session.Date.Truncate(24*time.Hour).Equal(today) && session.CompletedAt.IsZero() {
			return session, nil
		}
	}

	// Create new session
	session := &models.Session{
		UserID:    userID,
		Date:      today,
		CreatedAt: time.Now(),
	}

	// We don't create it yet, just return a placeholder
	// The actual creation happens when we have results
	return session, nil
}

func (s *ExerciseService) getCurrentSession(ctx context.Context, userID int) (*models.Session, error) {
	sessions, err := s.db.GetSessionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	for _, session := range sessions {
		if session.Date.Truncate(24*time.Hour).Equal(today) && session.CompletedAt.IsZero() {
			return session, nil
		}
	}

	return nil, fmt.Errorf("no active session found")
}

func (s *ExerciseService) GetSessionResults(ctx context.Context, sessionID int) ([]*models.ExerciseResult, error) {
	return s.db.GetExerciseResultsBySession(ctx, sessionID)
}

func (s *ExerciseService) GetUserSessions(ctx context.Context, userID int) ([]*models.Session, error) {
	return s.db.GetSessionsByUser(ctx, userID)
}

func (s *ExerciseService) GetUserProgress(ctx context.Context, userID int) ([]*models.Progress, error) {
	return s.db.GetProgress(ctx, userID)
}
